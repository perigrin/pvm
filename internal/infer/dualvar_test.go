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
// number gets the DualVar as it is, and numified leaves it for lossyCoercion to judge.
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

	// numified leaves a DualVar as it is: its one-leg reading is a lossy
	// coercion, which lossyCoercion excuses, not a numification.
	for _, expected := range []types.Type{types.Num, types.Int, types.Str, types.ArrayRef, types.Str | types.Glob} {
		if got := infer.Numified(types.DualVar, expected); got != types.DualVar {
			t.Errorf("Numified(DualVar, %v) = %v, want DualVar", expected, got)
		}
	}
}

// TestDualvarInUnionSlot: a slot whose type is a union holding Str --
// split's pattern (Str|Regex), stat's and lstat's file (Str|Glob),
// chdir's and truncate's (Str|FileHandle), prototype's (Str|Code) --
// reads a dualvar's string, and perl 5.42 -w is silent on each:
// `split dualvar(1, ","), "a,b"` is ("a", "b"), `stat dualvar(1,
// "/tmp")` is 13 fields, `chdir dualvar(1, "/tmp")` is true and
// `prototype dualvar(1, "CORE::push")` is '\@@'.
func TestDualvarInUnionSlot(t *testing.T) {
	for _, src := range []string{
		`use Scalar::Util qw(dualvar); my @f = split dualvar(1, ","), "a,b";`,
		`use Scalar::Util qw(dualvar); my @s = stat dualvar(1, "/tmp");`,
		`use Scalar::Util qw(dualvar); my @s = lstat dualvar(1, "/tmp");`,
		`use Scalar::Util qw(dualvar); chdir dualvar(1, "/tmp");`,
		`use Scalar::Util qw(dualvar); truncate dualvar(1, "/tmp/x"), 0;`,
		`use Scalar::Util qw(dualvar); my $p = prototype dualvar(1, "CORE::push");`,
	} {
		wantSilentResolved(t, src)
	}
}

// TestDualvarLossyUnderFlag: reading one leg of a dualvar loses the
// other -- `0 + "$d"` is 0 for `dualvar(5, "x")`, not 5 (the paper's
// Example 4) -- so under --warn-lossy-coercion each such slot is a
// lossy coercion, in a plain slot and in a union one alike. A Boolean's
// coercion loses only its is_bool mark, and stays silent under the flag.
func TestDualvarLossyUnderFlag(t *testing.T) {
	lossy := infer.Options{WarnLossyCoercion: true}
	for _, src := range []string{
		`use Scalar::Util qw(dualvar); my $x = dualvar(5, "x"); my $n = $x + 1;`,
		`use Scalar::Util qw(dualvar); my $x = dualvar(5, "x"); my $s = lc $x;`,
		`use Scalar::Util qw(dualvar); my $x = dualvar(5, "x"); my $e = ($x eq "x");`,
		`use Scalar::Util qw(dualvar); my $r = "-" x dualvar(2, "two");`,
		`use Scalar::Util qw(dualvar); my @f = split dualvar(1, ","), "a,b";`,
		`use Scalar::Util qw(dualvar); my @s = stat dualvar(1, "/tmp");`,
	} {
		if diags := diagsWithOptions(src, lossy); len(diags) != 1 || diags[0].Code != infer.CodeLossyCoercion {
			t.Errorf("%q: diagnostics %v, want one lossy coercion", src, diags)
		}
	}
	for _, src := range []string{
		"my $f = 7 < 3;\nmy $y = $f + 1;\n",
		"my @g = split((3 > 2), \"a1b\");\n",
	} {
		for _, d := range diagsWithOptions(src, lossy) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}
}

// TestBooleanInUnionSlot: a Boolean in a union slot holding Str reads
// as the string it stringifies to, as it does in a plain Str slot:
// perl 5.42 -w is silent on `split((3 > 2), "a1b")`, which is ("a", "b").
func TestBooleanInUnionSlot(t *testing.T) {
	wantSilentResolved(t, "my @g = split((3 > 2), \"a1b\");\n")
	for _, c := range []struct{ expected, want types.Type }{
		{types.Str | types.Regex, types.Str},
		{types.Str | types.Glob, types.Str},
		{types.Str | types.Code, types.Str},
	} {
		if got := infer.Numified(types.Boolean, c.expected); got != c.want {
			t.Errorf("Numified(Boolean, %v) = %v, want %v", c.expected, got, c.want)
		}
	}
}
