// ABOUTME: Tests for the round-2 review's flow findings against pu (I3 residual, I4 residual, I5 residual, I10, S5):
// ABOUTME: a guard shadow dropping the possible-undef mark, a loop's deref narrowing escaping a zero-iteration path, an element shadow surviving several write forms, a ternary's arms sharing state, and a cubic nested ternary.
package infer_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestGuardShadowKeepsPossibleUndef (I3 residual): `my $b = $c ? blessed($x)
// : []; if (!ref $b) { $b eq "C" }` narrows $b to a shadow that still
// carries Undef (blessed's own Str|Undef, forgiven by the ternary's
// possible-undef mark); the guard's shadow Define must carry that mark
// forward too, not just the narrowed type, or the read inside the guard
// reports by default where pu and perl 5.42 are silent.
func TestGuardShadowKeepsPossibleUndef(t *testing.T) {
	const src = `use Scalar::Util qw(blessed reftype);
my ($x, $c) = @ARGV;
my $b = $c ? blessed($x) : [];
if (!ref $b) { my $t = ($b eq "C"); }
`
	if diags := diagsWithDefaults(src); len(diags) != 0 {
		t.Errorf("%q: diagnostics %v, want none: the guard shadow should keep $b's possible-undef mark", src, diags)
	}
}

// TestLoopBodyJoinsZeroIterations (I4 residual, the deref-narrowing
// regression): `for my $i (1..$c) { my $q = $y->{k} }` only narrows $y to a
// HashRef on the path where the loop body actually runs; with $c = 0 that
// path never runs, so $y past the loop must still include its pre-loop
// Undef -- pu and perl agree bless($y, 'Foo') then dies.
func TestLoopBodyJoinsZeroIterations(t *testing.T) {
	for _, src := range []string{
		"my $c = 0; my $y = undef; for my $i (1..$c) { my $q = $y->{k} } bless $y, 'Foo';\n",
		"my $c = 0; my $y = undef; while ($c) { my $q = $y->{k}; last } bless $y, 'Foo';\n",
		"my @a = (); my $y = undef; for my $i (@a) { my $q = $y->{k} } bless $y, 'Foo';\n",
		"my @a = (); my $y = undef; my @m = map { $y->{k} } @a; bless $y, 'Foo';\n",
		"my @a = (); my $y = undef; my @m = grep { $y->{$_} } @a; bless $y, 'Foo';\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) == 0 {
			t.Errorf("%q: no diagnostics, want bless's Undef argument reported", src)
		}
	}
}

// TestElementShadowEndsAtEveryWrite (I5 residual): inside `if (ref
// $r[0])`, each of these writes replaces what sits at index 0 of @r --
// through a foreach alias, a slice, a computed index, or a list-assign --
// so the guard's stale Ref shadow for $r[0] must not survive any of them;
// bless($r[0], ...) past the write must still see the Undef pu and perl
// see.
func TestElementShadowEndsAtEveryWrite(t *testing.T) {
	for _, src := range []string{
		`my @r = ({}, undef); if (ref $r[0]) { for my $e (@r) { $e = undef } bless $r[0], "Foo" }` + "\n",
		`my @r = ({}, undef); if (ref $r[0]) { for (@r) { $_ = undef } bless $r[0], "Foo" }` + "\n",
		`my @r = ({}, undef); if (ref $r[0]) { $_ = undef for @r; bless $r[0], "Foo" }` + "\n",
		`my @r = ({}, undef); if (ref $r[0]) { @r[0,1] = (undef, undef); bless $r[0], "Foo" }` + "\n",
		`my @r = ({}, undef); my $i = 0; if (ref $r[0]) { $r[$i] = undef; bless $r[0], "Foo" }` + "\n",
		`my @r = ({}, undef); if (ref $r[0]) { (my $q, @r) = (1, undef); bless $r[0], "Foo" }` + "\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) == 0 {
			t.Errorf("%q: no diagnostics, want the stale $r[0] guard shadow gone after the write", src)
		}
	}
}

// TestElementShadowEndsAtSubstitutionThroughAlias (I5, the foreach-alias
// residual): a substitution, a transliteration or an explicit `=~`
// binding writing through a foreach's alias invalidates the array's
// element guard shadow the same way a plain assignment through it does
// (TestElementShadowEndsAtEveryWrite), even though none of the three
// changes the alias's own type -- the write-back guard that decides
// whether to clear is keyed to a type change, and none of them produce
// one. perl 5.42 dies "Can't bless non-reference value" for each; pu
// reports; this branch, head and cf were all silent on all three.
func TestElementShadowEndsAtSubstitutionThroughAlias(t *testing.T) {
	for _, src := range []string{
		`my @r = ({}, undef); if (ref $r[0]) { for (@r) { tr/x/y/; } bless $r[0], "Foo" }` + "\n",
		`my @r = ({}, undef); if (ref $r[0]) { s/x/y/ for @r; bless $r[0], "Foo" }` + "\n",
		`my @r = ({}, undef); if (ref $r[0]) { for my $e (@r) { $e =~ s/x/y/ } bless $r[0], "Foo" }` + "\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) == 0 {
			t.Errorf("%q: no diagnostics, want the stale $r[0] guard shadow gone after the write", src)
		}
	}
}

// TestShadowClearStaysInItsBranch (I5, R6/S4's overreach, negative): a
// write to $a[0] inside one arm of an if/else must not delete the guard
// shadow from the scope the sibling arm shares -- `if ($c) { $a[0] = undef
// } else { $a[0] eq "x" } ` must stay silent in the arm that never wrote,
// as cf and perl are.
func TestShadowClearStaysInItsBranch(t *testing.T) {
	const src = `my @a = (undef, "s");
my ($x, $c) = @ARGV;
$a[0] = $x if $x;
if (defined $a[0]) {
    if ($c) { $a[0] = undef } else { my $t = ($a[0] eq "x") }
}
`
	if diags := diagsWithDefaults(src); len(diags) != 0 {
		t.Errorf("%q: diagnostics %v, want none: the else arm never wrote $a[0]", src, diags)
	}
}

// TestTernaryArmsStartFromTheCondition (S5): a ternary's else arm must
// start from the state before the ternary, not from whatever the then arm's
// own processing left -- `my $v = "s"; my $r = $c ? ($v = undef) : ($v eq
// "x");` must stay silent, as pu, cf and perl 5.42 should (perl never
// raises a warning here): the else arm's `$v eq "x"` reads $v as it was
// before the ternary ran, since the then arm never ran on that path.
func TestTernaryArmsStartFromTheCondition(t *testing.T) {
	const src = `my ($x, $c) = @ARGV;
my $v = "s";
my $r = $c ? ($v = undef) : ($v eq "x");
`
	if diags := diagsWithDefaults(src); len(diags) != 0 {
		t.Errorf("%q: diagnostics %v, want none: the else arm should read $v as it was before the ternary", src, diags)
	}
}

// TestNestedTernaryIsNotCubic (I10): a 1000-deep right-nested ternary over
// distinct variables must analyse in time roughly linear in its depth, not
// cubic -- cubic turns a ~40KB file into a CPU-pinning, roughly
// minute-long analysis (55s measured at n=1000 before the fix); linear
// finishes in a small number of seconds even generously bounded for a
// shared, possibly loaded CI machine.
func TestNestedTernaryIsNotCubic(t *testing.T) {
	var b strings.Builder
	b.WriteString("my $v0 = shift;\n")
	const n = 1000
	for i := 1; i < n; i++ {
		fmt.Fprintf(&b, "my $v%d = shift;\n", i)
	}
	b.WriteString("my $y = ")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "ref $v%d ? $v%d : ", i, i)
	}
	b.WriteString("0;\n")
	src := b.String()

	start := time.Now()
	diagsWithDefaults(src)
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Errorf("analysing a %d-deep nested ternary took %s, want it roughly linear in depth", n, elapsed)
	}
}
