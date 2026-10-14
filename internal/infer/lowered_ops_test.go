// ABOUTME: Tests for AnalyzeLowered's operand and argument rules where perl 5.42 -w is silent and PSC was not.
// ABOUTME: Each pairs the silent cases with a control perl warns about or rejects, which must still be reported.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
)

// A Scalar is some scalar, so it may be any member of a union slot that
// has scalar members: perl 5.42 -w is silent on `chdir $dir` and
// `prototype($dir)` for `my $dir = shift // "."`. Any and a Hole's Any
// satisfy it too. A slot with no scalar member still rejects a Scalar:
// perl 5.42 dies "Experimental push on scalar is now forbidden" on
// `push $s, 1`.
func TestUntypedScalarSatisfiesUnionParameter(t *testing.T) {
	for _, src := range []string{
		"my $dir = shift // \".\";\nchdir $dir or die;\nprint defined prototype($dir) ? \"p\\n\" : \"none\\n\";\n",
		"open my $fh, \"+>\", \"t.out\" or die;\ntruncate $fh, 0;\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	handleOrStr := types.Int | types.Num | types.Str | types.GlobRef | types.Glob | types.NaN | types.Inf | types.IO
	for _, actual := range []types.Type{types.Scalar, types.Any} {
		if !types.TypeSatisfies(actual, handleOrStr) {
			t.Errorf("TypeSatisfies(%v, %v) = false, want true", actual, handleOrStr)
		}
	}
	for _, required := range []types.Type{types.Array, types.Code | types.Glob | types.IO} {
		if types.TypeSatisfies(types.Scalar, required) {
			t.Errorf("TypeSatisfies(Scalar, %v) = true, want false", required)
		}
	}

	src := "my $s = shift // \".\";\npush $s, 1;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
}

// perl compiles a string on the right of `=~` or `!~` as a pattern: perl
// 5.42 -w is silent on `$a[0] =~ PKG_PATTERN` for a constant sub returning
// q#\A[a-zA-Z_]\w*\z#, and on `$x !~ $p` for a Str $p. An undef there is
// still reported: perl warns "Use of uninitialized value in regexp
// compilation" on `$x =~ undef`.
func TestStringPatternRightOfBindIsSilent(t *testing.T) {
	for _, src := range []string{
		"sub PKG_PATTERN () { q#\\A[a-zA-Z_]\\w*\\z# }\nmy @a = (\"Foo\");\nprint \"m\\n\" if defined $a[0] and $a[0] =~ PKG_PATTERN;\n",
		"my $x = \"a\";\nmy $p = \"a\";\nprint \"m\\n\" if $x !~ $p;\n",
		"my $x = \"a\";\nmy $p = \"a\";\nmy $m = $x =~ $p;\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := "my $x = \"a\";\nmy $m = $x =~ undef;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
}

// Unary plus is the identity, so `+@d` is @d as its slot takes it: perl
// 5.42 -w is silent on `splice @t, 0, +@d`, which removes one element per
// element of @d, and `my $n = +@e` is the count. It does not hide what
// perl warns about without it: `splice @t, 0, +undef` warns "Use of
// uninitialized value in splice". `length(@a)`, which warns "length()
// used on @a", is still reported; `length(+@a)`, on which perl warns the
// same, is counted, and is not.
func TestUnaryPlusArrayIsItsCount(t *testing.T) {
	for _, src := range []string{
		"sub f { my @d = @{$_[0]}; my @t = @{$_[1]}; return join('/', @d) eq join('/', splice @t, 0, +@d) }\nprint f([1], [1, 2]) ? \"same\\n\" : \"differ\\n\";\n",
		"my @e = (1, 2, 3);\nmy $n = +@e;\nmy $m = $n + 1;\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	for _, src := range []string{
		"my @a = (1, 2);\nmy $l = length(@a);\n",
		"my @t = (1, 2, 3);\nsplice @t, 0, +undef;\n",
	} {
		if diags := loweredDiags(src); len(diags) != 1 {
			t.Errorf("%q: diagnostics %v, want one", src, diags)
		}
	}
}

// `<=>` is undef only when an operand is NaN, so a comparison of operands
// not known to be NaN is an Int: perl 5.42 -w is silent on
// `$sgn * ($re1 <=> $re2)` for (1, 2, 3), which is -1. An operand known to
// be NaN keeps the Undef: perl warns "Use of uninitialized value in
// multiplication (*)" on `1 * ($nan <=> 1)`. No expression is typed exactly
// NaN yet, so the test holds that half at SpaceshipResult itself.
func TestSpaceshipInArithmeticIsSilent(t *testing.T) {
	for _, src := range []string{
		"my ($sgn, $re1, $re2) = (1, 2, 3);\nmy $c = $sgn * ($re1 <=> $re2);\n",
		"my $x = 2.5;\nmy $c = 1 * ($x <=> 3);\n",
		"my $x = 2 ** 0.5;\nmy $c = 1 * ($x <=> 3);\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	ordered := types.Int | types.Undef
	for _, c := range []struct{ l, r, want types.Type }{
		{types.Int, types.Num, types.Int},
		{types.Unknown, types.Any, types.Int},
		{types.Num | types.NaN, types.Int, types.Int},
		{types.NaN, types.Int, ordered},
		{types.Int, types.NaN | types.Inf, ordered},
	} {
		if got := infer.SpaceshipResult(c.l, c.r, ordered); got != c.want {
			t.Errorf("SpaceshipResult(%v, %v) = %v, want %v", c.l, c.r, got, c.want)
		}
	}
}
