// ABOUTME: Tests that a dualvar's result is checked as the string or number perl reads from it,
// ABOUTME: while DualVar stays its own lattice member, outside both Str and Num.
package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
)

// TestDualvarInStringAndNumericOps: a dualvar reads its number where a
// number is wanted and its string where a string is, and perl 5.42 -w
// says nothing: with `my $x = dualvar(5, "x")`, `$x + 1` is 6, `lc $x`
// is "x", `$x eq "x"` is true, `length dualvar(3, "abc")` is 3,
// `substr(dualvar(1, "abc"), 1)` is "bc", `"-" x dualvar(2, "two")` is
// "--", `chr(dualvar(65, "A"))` is "A" and `dualvar(1, "a") . "b"` is
// "ab". The paper's [Coerce-DualVar] states the same two legs: d ⇓^Str s,
// d ⇓^Num n.
func TestDualvarInStringAndNumericOps(t *testing.T) {
	for _, src := range []string{
		`use Scalar::Util qw(dualvar); my $x = dualvar(5, "x"); my $n = $x + 1;`,
		`use Scalar::Util qw(dualvar); my $x = dualvar(5, "x"); my $s = lc $x;`,
		`use Scalar::Util qw(dualvar); my $x = dualvar(5, "x"); my $e = ($x eq "x");`,
		`use Scalar::Util qw(dualvar); my $l = length dualvar(3, "abc");`,
		`use Scalar::Util qw(dualvar); my $s = substr(dualvar(1, "abc"), 1);`,
		`use Scalar::Util qw(dualvar); my $r = "-" x dualvar(2, "two");`,
		`use Scalar::Util qw(dualvar); my $c = chr(dualvar(65, "A"));`,
		`use Scalar::Util qw(dualvar); my $k = dualvar(1, "a") . "b";`,
	} {
		wantSilentResolved(t, src)
	}
}

// TestDualvarStillDistinct: reading one leg of a dualvar does not make
// it a member of that leg's type. The paper's Example 4 places DualVar
// in Scalar but outside Str and Num -- `0 + "$d"` is 0 for
// `dualvar(5, "x")`, not 5 -- so the lattice keeps it apart, dualvar's
// call is still a DualVar, a slot that wants neither a string nor a
// number gets the DualVar as it is, and an Undef beside it is kept.
func TestDualvarStillDistinct(t *testing.T) {
	for _, required := range []types.Type{types.Str, types.Num, types.Int} {
		if types.TypeSatisfies(types.DualVar, required) {
			t.Errorf("TypeSatisfies(DualVar, %v) = true, want false", required)
		}
	}

	src := `use Scalar::Util qw(dualvar); my $x = dualvar(5, "x");`
	if got, _ := declaredCallType(t, src, `dualvar(5, "x")`); got != types.DualVar {
		t.Errorf("dualvar(5, \"x\") is %v, want DualVar", got)
	}

	for _, c := range []struct{ actual, expected, want types.Type }{
		{types.DualVar, types.Num, types.Num},
		{types.DualVar, types.Int, types.Int},
		{types.DualVar, types.Str, types.Str},
		{types.DualVar | types.Undef, types.Num, types.Num | types.Undef},
		{types.DualVar, types.ArrayRef, types.DualVar},
		{types.DualVar, types.HashRef | types.Undef, types.DualVar},
	} {
		if got := infer.Numified(c.actual, c.expected); got != c.want {
			t.Errorf("Numified(%v, %v) = %v, want %v", c.actual, c.expected, got, c.want)
		}
	}
}
