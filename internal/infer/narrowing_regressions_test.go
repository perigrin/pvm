// ABOUTME: Tests for the three narrowing regressions the crochet:review round found against pu (I1, I3, I6):
// ABOUTME: a deref overwriting a blessed object's type, the possible-undef mark not joining across branches, and a stale element guard shadow.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// TestDerefKeepsBlessedObject (I1a): `$o->{k}` reads a blessed object
// through an arrow, which perl autovivifies nothing and leaves $o exactly
// what it was -- an Object, Foo->new's own type (infer.go's methodCall) --
// not a plain HashRef: Encode's `my $enc = find_encoding("UTF-8"); my $name
// = $enc->{Name}; decode($enc, "abc")` must not lose $enc's Object to
// report decode's Object parameter a coercion mismatch.
func TestDerefKeepsBlessedObject(t *testing.T) {
	const src = `package Foo;
sub new { my $c = shift; return bless {}, $c }
package main;
my $o = Foo->new;
my $v = $o->{k};
`
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	_, _, st := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
	sym, ok := st.Lookup("$o")
	if !ok {
		t.Fatalf("$o not found")
	}
	if sym.Type&types.Object == 0 {
		t.Errorf("$o is %v after $o->{k}, want Object kept", sym.Type)
	}
}

// TestDerefNarrowingStaysInItsArm (I1b): `$c && $y->{k}` dereferences $y
// only where $c is true; where $c is false $y is never touched and keeps
// whatever it held before, undef here, so the type after the statement
// must still include Undef -- narrowDerefBase must not narrow $y on the
// outer, unconditional scope the && short-circuit's right operand runs
// inside of.
func TestDerefNarrowingStaysInItsArm(t *testing.T) {
	const src = `sub f {
    my ($c) = @_;
    my $y = undef;
    $c && $y->{k};
}
`
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	_, _, st := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
	sym, ok := st.LookupAll("$y")
	if !ok {
		t.Fatalf("$y not found")
	}
	if sym.Type&types.Undef == 0 {
		t.Errorf("$y is %v after `$c && $y->{k}`, want Undef still possible", sym.Type)
	}
}

// diagsWithDefaults is loweredDiags with Options{} spelled out, for cases
// that read better naming the default explicitly alongside a case that
// passes WarnPossibleUndef or WarnLossyCoercion.
func diagsWithDefaults(src string) []infer.Diagnostic {
	return diagsWithOptions(src, infer.Options{})
}

// TestDefiniteUndefInElseIsReported (I3, the leak): `if (@ARGV) { $x =
// blessed($r) } else { print lc($x) }` reads $x inside the else arm, which
// the if never runs -- there $x is exactly the plain `undef` it was
// declared, never blessed()'s forgiven one, so it must report by default,
// as pu does. Walking the primary arm must not leave its own
// possible-undef mark on the symbol for the else arm to inherit.
func TestDefiniteUndefInElseIsReported(t *testing.T) {
	const src = `use Scalar::Util qw(blessed);
my $r = {};
my $x = undef;
if (@ARGV) { $x = blessed($r); } else { print lc($x), "\n"; }
`
	if diags := diagsWithDefaults(src); len(diags) == 0 {
		t.Errorf("%q: no diagnostics, want the else arm's definite Undef reported", src)
	}
}

// TestPossibleUndefJoinsAcrossBranches (I3, the join): `if (@ARGV) { $x =
// blessed($r) }` with no else leaves $x, where @ARGV is empty, exactly the
// plain `undef` it was declared -- a real, unforgiven path past the
// statement -- so the join must still report, even though the branch that
// did run left a forgiven mark: the mark joins only when every path that
// can still carry Undef past the conditional is itself forgiven.
func TestPossibleUndefJoinsAcrossBranches(t *testing.T) {
	const src = `use Scalar::Util qw(blessed);
my $r = {};
my $x = undef;
if (@ARGV) { $x = blessed($r); }
print lc($x), "\n";
`
	if diags := diagsWithDefaults(src); len(diags) == 0 {
		t.Errorf("%q: no diagnostics, want the join to still report", src)
	}
}

// TestPossibleUndefFollowsCopies (I3, the copy): a plain copy, `my $j =
// $k;`, carries $k's possible-undef mark to $j, the same way narrow
// already carries it from a declared call's own result: Test2::Mock-shaped
// code reassigns $k afterward, which must not retroactively un-mark the
// $j it already copied from.
func TestPossibleUndefFollowsCopies(t *testing.T) {
	const src = `use Scalar::Util qw(blessed);
my ($x) = @_;
my $k = blessed($x);
my $j = $k;
$k = 1;
my $same = ($j eq "Foo");
`
	if diags := diagsWithDefaults(src); len(diags) != 0 {
		t.Errorf("%q: diagnostics %v, want none: $j should still carry $k's mark", src, diags)
	}
}

// TestElementGuardEndsAtAssignment (I6): inside `if (ref $r[0])`,
// assigning `$r[0] = ...` replaces whatever the guard shadow said $r[0]
// was; a read after it must fall back to @r's own recorded element type,
// not the stale Ref the guard shadowed it to before the assignment ran.
func TestElementGuardEndsAtAssignment(t *testing.T) {
	const src = `my @r = (undef, 2);
if (ref $r[0]) {
    $r[0] = "abc";
    my $after = $r[0];
}
`
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	_, _, st := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
	sym, ok := st.LookupAll("$after")
	if !ok {
		t.Fatalf("$after not found")
	}
	if sym.Type == types.Ref {
		t.Errorf("$after is %v, want the stale $r[0] guard shadow gone after the assignment", sym.Type)
	}
}

// TestElementGuardEndsAtShift (I6): a shift renumbers every element of the
// array, so a guard's shadow for one index -- what the guard found true of
// whatever sat there before the shift -- answers for nothing after it.
func TestElementGuardEndsAtShift(t *testing.T) {
	const src = `my @r = (undef, 2);
if (ref $r[0]) {
    shift @r;
    my $after = $r[0];
}
`
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	_, _, st := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
	sym, ok := st.LookupAll("$after")
	if !ok {
		t.Fatalf("$after not found")
	}
	if sym.Type == types.Ref {
		t.Errorf("$after is %v, want the stale $r[0] guard shadow gone after the shift", sym.Type)
	}
}
