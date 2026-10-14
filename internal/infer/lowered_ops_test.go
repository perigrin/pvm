// ABOUTME: Tests for AnalyzeLowered's operand and argument rules where perl 5.42 -w is silent and PSC was not.
// ABOUTME: Each pairs the silent cases with a control perl warns about or rejects, which must still be reported.

package infer_test

import (
	"strings"
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

// A Boolean in a Str slot stringifies to "1" or "" and perl 5.42 -w says
// nothing: `"" . $s =~ /^threads=/` is "1", `"[" . (1 == 2) . "]"` is
// "[]", and `length(1 == 1)` is 1. An undef there is still reported, as
// perl warns "Use of uninitialized value in concatenation (.) or string"
// on `"" . undef`, and so is a Boolean that may be undef.
func TestBooleanInStrSlotIsSilent(t *testing.T) {
	for _, src := range []string{
		"my $s = \"threads=1\";\nprint \"\" . $s =~ /^threads=/, \"\\n\";\n",
		"my $b = \"[\" . (1 == 2) . \"]\";\n",
		"my $n = length(1 == 1);\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := "my $s = \"\" . undef;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
	for _, c := range []struct{ actual, want types.Type }{
		{types.Boolean, types.Str},
		{types.Boolean | types.Undef, types.Str | types.Undef},
		{types.Scalar, types.Scalar},
		{types.Any, types.Any},
	} {
		if got := infer.Numified(c.actual, types.Str); got != c.want {
			t.Errorf("Numified(%v, Str) = %v, want %v", c.actual, got, c.want)
		}
	}
}

// lossyDiags analyses src on the lowered path with the lossy-coercion
// warning on and returns its diagnostics.
func lossyDiags(t *testing.T, src string) []infer.Diagnostic {
	t.Helper()
	_, diags, _ := analyzeSourceWithOptions(t, []byte(src), infer.Options{WarnLossyCoercion: true})
	return diags
}

// A reference in a Str slot is a valid coercion: perl 5.42 -w is silent on
// `print $h` and `print STDOUT $h` for a hashref, which print HASH(0x...),
// on `$r . "x"` for an arrayref, and on `"" . A->new`. perigrin ruled
// 2026-10-13 that it is no diagnostic by default. A ref in a numeric slot,
// `$h + 1`, is kept: he ruled that on the same day. So is a ref where no
// Str is taken: perl dies "Not a GLOB reference" on `close({})`.
func TestRefInStrSlotIsSilent(t *testing.T) {
	for _, src := range []string{
		"my $h = {};\nprint STDOUT $h;\nprint $h;\n",
		"my $r = [];\nmy $s = $r . \"x\";\n",
		"package A; sub new { bless {}, shift }\npackage main;\nmy $s = \"\" . A->new;\n",
		"my $c = sub { 1 };\nmy $n = length($c);\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	for _, src := range []string{
		"my $h = {};\nmy $x = ($h) + 1;\n",
		"close({});\n",
	} {
		if diags := loweredDiags(src); len(diags) != 1 || diags[0].Code == infer.CodeLossyCoercion {
			t.Errorf("%q: diagnostics %v, want one that is no lossy coercion", src, diags)
		}
	}
}

// Under WarnLossyCoercion a ref in a Str slot is a lossy-coercion warning
// naming the ref type: the string, HASH(0x...), does not give the referent
// back. A ref in a numeric slot is reported as it is by default.
func TestLossyCoercionWarningIsOptIn(t *testing.T) {
	for src, ref := range map[string]string{
		"my $h = {};\nprint $h;\n":           "HashRef",
		"my $r = [];\nmy $s = $r . \"x\";\n": "ArrayRef",
	} {
		diags := lossyDiags(t, src)
		if len(diags) != 1 {
			t.Errorf("%q: diagnostics %v, want one", src, diags)
			continue
		}
		d := diags[0]
		if d.Code != infer.CodeLossyCoercion || d.Severity != infer.Warning || !strings.Contains(d.Message, ref) {
			t.Errorf("%q: %s %v %q, want a lossy-coercion warning naming %s", src, d.Code, d.Severity, d.Message, ref)
		}
	}

	src := "my $h = {};\nmy $x = ($h) + 1;\n"
	if diags := lossyDiags(t, src); len(diags) != 1 || diags[0].Code != infer.CodeCoercionMismatch {
		t.Errorf("%q: diagnostics %v, want one coercion mismatch", src, diags)
	}
}

// A Str in a numeric slot whose content PSC does not know is silent by
// default, and a lossy coercion under WarnLossyCoercion: perigrin ruled so
// on 2026-10-14 for `$1 + 0`, `"$]" < 5.016` and `substr($s, 0, 1) % 2`.
// A literal known to be non-numeric is still reported by default, as perl
// 5.42 -w warns `Argument "abc" isn't numeric in addition (+)` on
// `"abc" + 1`, and on `abs("abc")`, `"" + 1` and `"1abc" + 1`. A literal
// that looks like a number is silent, as perl is on `"3" + 1`,
// `" 3.5e2 " * 1`, `"inf" + 0` and `"0 but true" + 1`.
func TestStrInNumericSlotIsOptIn(t *testing.T) {
	unknown := []string{
		"my $s = \"11\";\nif ($s =~ /(.+)/) { my $n = $1 + 0; }\n",
		"my $old = \"$]\" < 5.016;\n",
		"my $s = \"13\";\nmy $odd = substr($s, 0, 1) % 2;\n",
		"my $s = \"abc\";\nmy $n = $s * 2;\n",
	}
	for _, src := range unknown {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}
	for _, src := range []string{
		"my $n = \"3\" + 1;\n",
		"my $m = \" 3.5e2 \" * 1;\n",
		"my $i = \"inf\" + 0;\n",
		"my $t = \"0 but true\" + 1;\n",
	} {
		for _, d := range lossyDiags(t, src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := "my $s = \"abc\";\nmy $n = $s * 2;\n"
	if diags := lossyDiags(t, src); len(diags) != 1 || diags[0].Code != infer.CodeLossyCoercion {
		t.Errorf("%q: diagnostics %v under the opt-in, want one lossy coercion", src, diags)
	}

	for _, src := range []string{
		"my $n = \"abc\" + 1;\n",
		"my $n = abs(\"abc\");\n",
		"my $n = \"\" + 1;\n",
		"my $n = \"1abc\" + 1;\n",
	} {
		if diags := loweredDiags(src); len(diags) != 1 || diags[0].Code == infer.CodeLossyCoercion {
			t.Errorf("%q: diagnostics %v, want one that is no lossy coercion", src, diags)
		}
	}
}
