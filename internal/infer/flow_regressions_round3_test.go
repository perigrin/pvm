// ABOUTME: Tests for the round-3 review's flow findings (V2, V5, V6 and the loop-deref false positive):
// ABOUTME: a ternary's arms joining back, frame bookkeeping staying linear, element shadows ending at every write, a loop deref's always-run parts.
package infer_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
)

// TestTernaryThenArmJoinsBack (V2): whatever a ternary's then arm writes
// is one of the two values a variable may hold past the ternary, whether
// or not the else arm touches the variable at all. perl 5.42 dies "Can't
// bless non-reference value" on each of these when $c is true; pu and
// round 2 report every one.
func TestTernaryThenArmJoinsBack(t *testing.T) {
	for _, src := range []string{
		"my $c = shift; my $y = {}; my $r = $c ? ($y = undef) : 1; bless $y, 'Foo';\n",
		"my $c = shift; my $y = {}; $c ? ($y = undef) : 0; bless $y, 'Foo';\n",
		"my $c = shift; my $y = {}; my $r = $c ? ($y = undef) : ($y = []); bless $y, 'Foo';\n",
		"my $y = {}; my $r = ref $y ? ($y = undef) : 0; bless $y, 'Foo';\n",
		"my $c = shift; my $y = {}; my $r = $c ? do { $y = undef; 1 } : 0; bless $y, 'Foo';\n",
		"my $c = shift; my $y = {}; my $w = 1; my $r = $c ? ($y = undef) : ($w = 2); bless $y, 'Foo';\n",
		"my $c = shift; my $y = {}; my @l = ($c ? ($y = undef) : 0); bless $y, 'Foo';\n",
		"my ($x, $c) = @ARGV; my $y = {}; if ($x // 1) { my $r = $c ? ($y = undef) : 0 } bless $y, 'Foo';\n",
		"my ($x, $c) = @ARGV; my $y = {}; my $r = $c ? ($x ? ($y = undef) : 1) : 0; bless $y, 'Foo';\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) == 0 {
			t.Errorf("%q: no diagnostics, want bless's possibly-Undef argument reported", src)
		}
	}
}

// TestTernaryGuardNoFalsePositive (V2, negative): past a ternary each
// variable holds what one arm or the other left it -- not also what it
// held before, since one arm always runs. Where an arm writes the
// condition's own variable, the other arm holds it as the condition
// narrowed it; an arm that always exits adds nothing. perl 5.42 -w is
// silent on each; pu is too.
func TestTernaryGuardNoFalsePositive(t *testing.T) {
	for _, src := range []string{
		"my $x = undef; !defined $x ? ($x = 'd') : 0; my $s = ($x eq 'a');\n",
		"my $x = undef; defined $x ? 0 : ($x = 'd'); my $s = ($x eq 'a');\n",
		"my $c = shift; my $y = undef; $c ? ($y = {}) : ($y = []); bless $y, 'Foo';\n",
		"my $c = shift; my $x = undef; $c ? ($x = 'a') : ($x = 'b'); my $s = lc($x);\n",
		"my $c = shift; my $y = undef; my $n = $c ? ($y = {})->{k} : ($y = {})->{j}; bless $y, 'Foo';\n",
		"my $c = shift; my $y = {}; my $r = !$c ? do { $y = undef; die 'no' } : 0; bless $y, 'Foo';\n",
		"my $c = shift; my $v = 's'; my $r = $c ? ($v = undef) : ($v eq 'x');\n",
		"my $c = shift; my @a = ('x', undef); if (defined $a[0]) { my $q = $c ? ($a[0] = undef) : ($a[0] eq 'x') }\n",
		"my @a = (1, undef); my @x = map { defined $_ ? ($_ = 1) : 0 } @a;\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) != 0 {
			t.Errorf("%q: diagnostics %v, want none", src, diags)
		}
	}
}

// TestFrameBookkeepingIsLinear (V5): recording a touch, finding a
// variable, or joining a ternary's arms must not cost the depth of the
// scopes and frames open around it, or 600 nested ifs, a 600-deep
// ternary chain whose arms write, and a 900-deep chain of element guards
// take tens of seconds or minutes where pu takes about one, and a
// 2000-arm elsif chain takes 20 to 40 where pu takes about five. The bound is
// generous for a loaded machine; the cost in depth this guards against
// is many times over it.
func TestFrameBookkeepingIsLinear(t *testing.T) {
	var ifs strings.Builder
	const depth = 600
	for i := 0; i < depth; i++ {
		fmt.Fprintf(&ifs, "my $v%d = shift;\n", i)
	}
	for i := 0; i < depth; i++ {
		fmt.Fprintf(&ifs, "if ($v%d) {\n", i)
	}
	ifs.WriteString("my $z = 1;\n")
	ifs.WriteString(strings.Repeat("}\n", depth))

	var tern strings.Builder
	const chain = 600
	for i := 0; i < chain; i++ {
		fmt.Fprintf(&tern, "my $v%d = {};\n", i)
	}
	tern.WriteString("my $y = ")
	for i := 0; i < chain; i++ {
		fmt.Fprintf(&tern, "ref $v%d ? ($v%d = 1) : ", i, i)
	}
	tern.WriteString("0;\n")

	var elsifs strings.Builder
	const arms = 2000
	for i := 0; i < arms; i++ {
		fmt.Fprintf(&elsifs, "my $v%d = shift;\n", i)
	}
	elsifs.WriteString("if (ref $v0) { $v0 = 0 }\n")
	for i := 1; i < arms; i++ {
		fmt.Fprintf(&elsifs, "elsif (ref $v%d) { $v%d = %d }\n", i, i, i)
	}

	var elems strings.Builder
	const elemChain = 900
	fmt.Fprintf(&elems, "my @a = ({}) x %d;\nmy $y = ", elemChain)
	for i := 0; i < elemChain; i++ {
		fmt.Fprintf(&elems, "ref $a[%d] ? ($a[%d] = undef) : ", i, i)
	}
	elems.WriteString("0;\n")

	for name, src := range map[string]string{
		"nested ifs":            ifs.String(),
		"writing ternary chain": tern.String(),
		"element guard chain":   elems.String(),
		"elsif chain":           elsifs.String(),
	} {
		// The elsif chain is quadratic on pu too -- every arm restores
		// each variable the chain names -- and takes about 5s there.
		bound := 8 * time.Second
		if name == "elsif chain" {
			bound = 15 * time.Second
		}
		start := time.Now()
		diagsWithDefaults(src)
		if elapsed := time.Since(start); elapsed > bound {
			t.Errorf("analysing %s took %s, want well under %s", name, elapsed, bound)
		}
		t.Logf("%s: %s", name, time.Since(start))
	}
}

// TestElementShadowEndsAtIncrementAndUndef (V6): inside `if (ref $r[0])`
// each of these may replace what sits at index 0 of @r -- an increment,
// an undef, a $_ write in map or grep, a foreach over an element, a slice
// or a list naming @r, or a write through a reference or a sub's @_ -- so
// the guard's Ref shadow of $r[0] must not survive it. perl 5.42 dies
// "Can't bless non-reference value" on each; pu reports each.
func TestElementShadowEndsAtIncrementAndUndef(t *testing.T) {
	for _, body := range []string{
		`$r[0]++;`,
		`--$r[0];`,
		`undef $r[0];`,
		`for my $e (@r) { $e++ }`,
		`for my $e (@r) { undef $e }`,
		`for my $e ($r[0]) { $e = undef }`,
		`for my $e (@r[0,1]) { $e = undef }`,
		`for my $e (reverse @r) { $e = undef }`,
		`$_ = undef for grep { 1 } @r;`,
		`map { $_ = undef } @r;`,
		`grep { $_ = undef } @r;`,
		`for my $e (@r) { my $f = \$e; $$f = undef }`,
		`my $rr = \@r; $rr->[0] = undef;`,
		`sub z { $_[0] = undef } z($r[0]);`,
		`sub { $_[0] = undef }->($r[0]);`,
	} {
		src := `my @r = ({}, undef); if (ref $r[0]) { ` + body + ` bless $r[0], "Foo" }` + "\n"
		if diags := diagsWithDefaults(src); len(diags) == 0 {
			t.Errorf("%q: no diagnostics, want the stale $r[0] guard shadow gone after the write", src)
		}
	}
}

// TestLoopDerefNoFalsePositive (the loop-deref false positive): a loop's
// condition, list or initialiser runs at least once, and so does a bare
// block or a do-while body, so a deref there forces its base to a ref as
// a straight-line one does; and inside a body, what follows a deref in
// the same pass sees the ref. Only past the loop does the zero-pass path
// join back in. perl 5.42 -w is silent on each; r2 and cf were too.
func TestLoopDerefNoFalsePositive(t *testing.T) {
	for _, src := range []string{
		"my $y = undef; while ($y->{k}) { last } bless $y, 'Foo';\n",
		"my $y = undef; my @x = grep { 1 } $y->{k}; bless $y, 'Foo';\n",
		"my $y = undef; for my $e ($y->{k}) { } bless $y, 'Foo';\n",
		"my $y = undef; do { my $q = $y->{k} } while (0); bless $y, 'Foo';\n",
		"my $y = undef; for (my $i = $y->{k}; 0; ) { } bless $y, 'Foo';\n",
		"my $y = undef; { my $q = $y->{k}; } bless $y, 'Foo';\n",
		"my $y = undef; my @s = sort { $a <=> $b } keys %{$y->{h}}; bless $y, 'Foo';\n",
		"for my $i (1..2) { my $h = undef; $h->{k} = $i; bless $h, 'Foo'; }\n",
		"my @out; for my $i (1..2) { my $h = undef; my $v = $h->{k}; push @out, bless($h, 'Foo'); }\n",
		"my @r = map { my $h = undef; $h->{k} = $_; bless $h, 'Foo' } 1..3;\n",
		"my $c = @ARGV; my $y = undef; for my $i (1..$c) { my $q = $y->{k}; bless $y, 'Foo' }\n",
	} {
		if diags := diagsWithDefaults(src); len(diags) != 0 {
			t.Errorf("%q: diagnostics %v, want none", src, diags)
		}
	}
}

// TestElementShadowEndsAtArgumentWrite (round-4 review): a sub or a
// builtin that writes its argument writes through an alias the same way
// an assignment does -- a user sub's @_ aliases each argument, and read,
// sysread, chop and chomp write theirs -- so passing the foreach alias,
// map's or a modifier for's $_, or the element itself, ends the guard's
// Ref shadow of $r[0]. perl 5.42 dies "Can't bless non-reference value"
// on each; pu reports each.
func TestElementShadowEndsAtArgumentWrite(t *testing.T) {
	for _, body := range []string{
		`for my $e (@r) { z($e) }`,
		`map { z($_) } @r;`,
		`z($_) for @r;`,
		`for my $e (@r) { $o->m($e) }`,
		`read(STDIN, $r[0], 0);`,
		`sysread(STDIN, $r[0], 0);`,
		`for my $e (@r) { read(STDIN, $e, 0) }`,
	} {
		src := `sub z { $_[0] = undef } my $o = shift; my @r = ({}, undef); if (ref $r[0]) { ` + body + ` bless $r[0], "Foo" }` + "\n"
		if diags := diagsWithDefaults(src); len(diags) == 0 {
			t.Errorf("%q: no diagnostics, want the stale $r[0] guard shadow gone after the write", src)
		}
	}
}

// TestLexicalArrayHidesElementShadow (round-4 review): a `my @r` inside
// the guarded block is a different array, so the guard's shadow of the
// outer $r[0] says nothing of the inner one's element. perl 5.42 dies
// "Can't bless non-reference value"; pu reports.
func TestLexicalArrayHidesElementShadow(t *testing.T) {
	const src = `my @r = ({}, undef); if (ref $r[0]) { my @r = (undef); bless $r[0], "Foo" }` + "\n"
	if diags := diagsWithDefaults(src); len(diags) == 0 {
		t.Errorf("%q: no diagnostics, want the inner @r's element unshadowed", src)
	}
}

// TestNestedFrameKeepsFirstTouch (round-4 review): a variable's value
// before a logical operator's right side is its value at the first touch
// in that whole side, not at the first touch inside a nested operator
// within it -- `$c && do { $y = {}; $d && ($y = []); 1 }` may skip the
// do entirely, leaving $y undef. perl 5.42 dies "Can't bless
// non-reference value" with $c false; pu reports.
func TestNestedFrameKeepsFirstTouch(t *testing.T) {
	const src = `my ($c, $d) = (0, 1); my $y = undef; $c && do { $y = {}; $d && ($y = []); 1 }; bless $y, "Foo";` + "\n"
	if diags := diagsWithDefaults(src); len(diags) == 0 {
		t.Errorf("%q: no diagnostics, want bless's possibly-Undef argument reported", src)
	}
}

// TestSetFlowSkipsUnchanged (V5): restoring or joining a variable to the
// value it already holds touches nothing, so the restore every arm of an
// elsif chain makes of every variable the chain names records only the
// few its arms changed.
func TestSetFlowSkipsUnchanged(t *testing.T) {
	st := infer.NewSymbolTable()
	st.Define(infer.Symbol{Name: "$x", Type: types.Int, Kind: infer.SymVariable})
	if n := infer.SetFlowTouches(st, "$x", types.Int); n != 0 {
		t.Errorf("setting $x to the Int it holds touched %d name(s), want 0", n)
	}
	if n := infer.SetFlowTouches(st, "$x", types.Str); n != 1 {
		t.Errorf("setting $x from Int to Str touched %d name(s), want 1", n)
	}
}
