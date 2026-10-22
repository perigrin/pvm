// ABOUTME: Tests for the round-5 review's flow findings (A1 and the I4 residual) and their review:
// ABOUTME: a guarded deref in a loop joins the zero-pass path, and code past an exit narrows nothing but a named sub.
package infer_test

import "testing"

// TestLoopGuardedDerefJoinsZeroIterations (A1): a deref in a loop body
// whose base a guard shadows -- an if or else body, a modifier, a
// ternary arm, a map block -- still runs only on a pass that happens at
// all, so past the loop the base also holds its type from before it. A
// guard around the loop, still open past it, joins the same way. perl
// 5.42 dies "Can't bless non-reference value" on each; pu reports each.
func TestLoopGuardedDerefJoinsZeroIterations(t *testing.T) {
	const pre = "my $c = 0; my $n = 0; my $y = $c ? {} : undef;\n"
	for _, src := range []string{
		pre + "for my $i (1..$n) { if (defined $y) { } else { $y->{k} = 1 } }\nbless $y, 'Foo';\n",
		pre + "for my $i (1..$n) { $y->{k} = 1 unless defined $y }\nbless $y, 'Foo';\n",
		pre + "for my $i (1..$n) { defined $y ? 1 : $y->{k} }\nbless $y, 'Foo';\n",
		pre + "for my $i (1..$n) { $y->{k} = 1 if !defined $y }\nbless $y, 'Foo';\n",
		pre + "for my $i (1..$n) { if (!defined $y) { $y->{k} = 1 } else { $y->{j} = 1 } }\nbless $y, 'Foo';\n",
		pre + "while ($n) { if (ref $y) { $y->{k} = 1 } else { $y->{j} = 1 } $n = 0 }\nbless $y, 'Foo';\n",
		pre + "my @l = (); my @m = map { defined $y ? 1 : $y->{k} } @l;\nbless $y, 'Foo';\n",
		"my $c = 0; my $y = undef; if (!defined $y) { for my $i (1..$c) { $y->{k} = 1 } bless $y, 'Foo' }\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) == 0 {
			t.Errorf("%q: no diagnostics, want bless's possibly-Undef argument reported", src)
		}
	}
}

// TestDerefAfterDieNarrowsNothing (I4 residual): a deref, or an
// assignment, after a statement that always exits never runs, so it
// narrows nothing on the path past the eval or sub around it. perl 5.42
// dies "Can't bless non-reference value" on each; pu reports each deref.
func TestDerefAfterDieNarrowsNothing(t *testing.T) {
	for _, src := range []string{
		"my $y = undef; eval { die 'x'; my $q = $y->{k} }; bless $y, 'Foo';\n",
		"my $y = undef; eval { die 'x'; { my $q = $y->{k} } }; bless $y, 'Foo';\n",
		"my $y = undef; do { eval { return 1; my $q = $y->{k} } }; bless $y, 'Foo';\n",
		"my $y = undef; eval { die 'x'; $y = {} }; bless $y, 'Foo';\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) == 0 {
			t.Errorf("%q: no diagnostics, want bless's possibly-Undef argument reported", src)
		}
	}
}

// TestDerefBeforeOrPastConditionalExitNarrows: a deref before the exit,
// or past one that runs only on some path, is reached, so it forces its
// base to a ref as a straight-line one does; so is one in a sub's body
// past an exit, which runs when the sub is called. Code past an exit
// sees what it narrows itself. perl 5.42 -w is silent on each.
func TestDerefBeforeOrPastConditionalExitNarrows(t *testing.T) {
	for _, src := range []string{
		"my $y = undef; eval { $y->{k} = 1; die 'x' }; bless $y, 'Foo';\n",
		"my $c = 0; my $y = undef; die 'x' if $c; my $q = $y->{k}; bless $y, 'Foo';\n",
		"my $c = 0; my $y = undef; eval { if ($c) { die 'x' } my $q = $y->{k} }; bless $y, 'Foo';\n",
		"my $y = undef; exit main(); sub main { my $q = $y->{k}; bless $y, 'Foo'; return 0 }\n",
		"sub f { die 'x'; my $h = undef; my $q = $h->{k}; bless $h, 'Foo' } 1;\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) != 0 {
			t.Errorf("%q: diagnostics %v, want none", src, diags)
		}
	}
}

// TestNamedSubAfterExitKeepsWrites (round-5 fix review, F1): perl
// defines a named sub at compile time, so one written past an exit still
// widens what it writes, as one anywhere else does. perl 5.42 prints ok;
// pu, r2, h3 and h4 are silent.
func TestNamedSubAfterExitKeepsWrites(t *testing.T) {
	const src = `my $p = undef; eval { die "x"; sub w { $p = {} } }; w(); bless $p, "Foo";` + "\n"
	if diags := diagsWithDefaults(src); len(diags) != 0 {
		t.Errorf("%q: diagnostics %v, want none", src, diags)
	}
}

// TestLoopDerefJoinsFromAfterAlwaysRunParts (round-5 fix review, F2):
// when a loop runs no pass, the variable past it holds what the
// condition, list or initialiser, which always run, left it, not its
// value from before the loop; and a `my $y` inside the loop is not the
// $y past it. perl 5.42 prints ok on each; pu, r2 and h3 are silent.
func TestLoopDerefJoinsFromAfterAlwaysRunParts(t *testing.T) {
	const pre = "my $y = undef; my $n = 0; "
	for _, src := range []string{
		pre + `for ($y = {}, my $i = 0; $i < $n; $i++) { my $y = $n ? {} : undef; $y->{k} = 1 unless defined $y } bless $y, "Foo";` + "\n",
		pre + `while (($y = {}) && $n) { my $y = $n ? {} : undef; $y->{k} = 1 unless defined $y; $n = 0 } bless $y, "Foo";` + "\n",
		pre + `for my $i (($y = {}) ? (1..$n) : ()) { my $y = $n ? {} : undef; $y->{k} = 1 unless defined $y } bless $y, "Foo";` + "\n",
		pre + `for ($y = {}; $n; ) { $y->{k} = 1 } bless $y, "Foo";` + "\n",
		pre + `for ($y = {}; $n; ) { $y->{k} = 1 unless defined $y } bless $y, "Foo";` + "\n",
		pre + `while (($y = {}) && $n) { $y->{k} = 1 } bless $y, "Foo";` + "\n",
		pre + `while (($y = {}) && $n) { $y->{k} = 1 unless defined $y } bless $y, "Foo";` + "\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) != 0 {
			t.Errorf("%q: diagnostics %v, want none", src, diags)
		}
	}
}
