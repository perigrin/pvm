// ABOUTME: Tests for Analyze's operand and argument rules where perl 5.42 -w is silent and PSC was not.
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
// 2026-10-13 that it is no diagnostic by default. A ref where no Str is
// taken is still reported: perl dies "Not a GLOB reference" on
// `close({})`.
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
		"close({});\n",
	} {
		if diags := loweredDiags(src); len(diags) != 1 || diags[0].Code == infer.CodeLossyCoercion {
			t.Errorf("%q: diagnostics %v, want one that is no lossy coercion", src, diags)
		}
	}
}

// Under WarnLossyCoercion a ref in a Str slot is a lossy-coercion warning
// naming the ref type: the string, HASH(0x...), does not give the referent
// back.
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
}

// A heredoc's body is not carried by the lowering, so its content is
// unknown, never the empty string. perl 5.42 -w is silent on
// `my $n = <<E + 1;` over a body of "3" and prints 4; an empty string
// would be reported as non-numeric by default.
func TestHeredocContentIsUnknown(t *testing.T) {
	src := "my $n = <<E + 1;\n3\nE\n"
	for _, d := range loweredDiags(src) {
		t.Errorf("%q: diagnostic %q, want none", src, d.Message)
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

// An object or a reference in a numeric slot is a lossy coercion, as
// perigrin ruled on 2026-10-14: perl numifies it through its overloading
// or to its address, and PSC cannot see most overloads (version's come
// from XS). perl 5.42 -w is silent on each repro, overloaded or not: a
// Math::BigInt `%` is 1, a version `<=` is "le", N's overloaded `+` is 7,
// and `($h) + 1` and `Foo->new + 1` add the address. Under
// WarnLossyCoercion each is a lossy-coercion warning. What perl warns
// about in a numeric slot is still reported by default: `undef() + 1`
// warns "Use of uninitialized value in addition (+)" and `"abc" + 1`
// "Argument "abc" isn't numeric".
func TestOverloadedObjectOperandIsSilent(t *testing.T) {
	cases := []string{
		"use Math::BigInt;\nmy $r = Math::BigInt->new(10) % Math::BigInt->new(3);\nprint \"$r\\n\";\n",
		"use version;\nprint version->new(\"1.2\") <= version->new(\"1.3\") ? \"le\\n\" : \"gt\\n\";\n",
		"package N; use overload '+' => sub { 7 }, fallback => 1; sub new { bless {}, shift }\npackage main;\nmy $value = N->new; print 0 + $value, \"\\n\";\n",
		"package A; use overload '\"\"' => sub { \"a\" }; sub new { bless {}, shift }\npackage main;\nprint \"\" . A->new eq 'a' ? \"ok\\n\" : \"no\\n\";\n",
		"my $h = {};\nmy $x = ($h) + 1;\n",
		"my $obj = Foo->new();\nmy $n = $obj + 1;\n",
	}
	for _, src := range cases {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
		diags := lossyDiags(t, src)
		if len(diags) == 0 {
			t.Errorf("%q: no diagnostic under the opt-in, want lossy coercions", src)
		}
		for _, d := range diags {
			if d.Code != infer.CodeLossyCoercion || d.Severity != infer.Warning {
				t.Errorf("%q: %s %v %q under the opt-in, want a lossy-coercion warning", src, d.Code, d.Severity, d.Message)
			}
		}
	}

	for _, src := range []string{
		"my $x = undef() + 1;\n",
		"my $n = \"abc\" + 1;\n",
	} {
		if diags := loweredDiags(src); len(diags) != 1 || diags[0].Code == infer.CodeLossyCoercion {
			t.Errorf("%q: diagnostics %v, want one that is no lossy coercion", src, diags)
		}
	}
}

// `//` and `||` replace an undefined left operand with the right, so their
// value never carries the left's Undef: perl 5.42 -w is silent on `my $v =
// $x // 0; $v + 1` and `my $l = $o || 0; length($l)` with $x and $o undef,
// and prints 1. `&&` yields a false left operand as it is, so it keeps the
// left's Undef: perl warns "Use of uninitialized value $c in addition (+)"
// on `my $c = $x && 0; $c + 1`.
func TestDefinedOrDropsLeftUndef(t *testing.T) {
	for _, src := range []string{
		"my $x = undef;\nmy $v = $x // 0;\nprint $v + 1;\n",
		"my $x = undef;\nmy $w = $x || 0;\nprint $w + 1;\n",
		"my $o = undef;\nmy $l = $o || 0;\nprint length($l);\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := "my $x = undef;\nmy $c = $x && 0;\nprint $c + 1;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
}

// A compound assignment `op=` leaves its target holding what op gives,
// CORE.pmt's return type for it, not its right side's type: perl 5.42,
// `my $s = "a"; $s .= 1` is "a1", a Str, and `my $t = "ab"; $t x= 3` is
// "ababab". `//=` assigns only an undefined target, so a defined one keeps
// its type, and `||=` replaces only a false one: perl 5.42 -w is silent on
// `my $k = 4; $k //= undef; $k + 1`, and on `my $s; $s .= <STDIN>; $s .
// "x"`. `&&=` assigns only a true target, so an undefined one stays undef:
// perl warns "Use of uninitialized value $u in addition (+)" on `my $u =
// undef; $u &&= 5; $u + 1`.
func TestCompoundAssignNarrowsToOperatorResult(t *testing.T) {
	for _, c := range []struct {
		src, name string
		want      types.Type
	}{
		{"my $s = \"a\";\n$s .= 1;\n", "$s", types.Str},
		{"my $t = \"ab\";\n$t x= 3;\n", "$t", types.Str},
		{"my $k = 4;\n$k //= undef;\n", "$k", types.Int},
		{"my $n = 2;\n$n ||= \"x\";\n", "$n", types.Int | types.Str},
		{"my $u = undef;\n$u &&= 5;\n", "$u", types.Undef | types.Int},
	} {
		if got, _ := loweredSymbolType(t, c.src, c.name); got != c.want {
			t.Errorf("%q: %s is %v, want %v", c.src, c.name, got, c.want)
		}
	}

	for _, src := range []string{
		"my $k = 4;\n$k //= undef;\nprint $k + 1;\n",
		"my $s;\n$s .= <STDIN>;\nprint $s . \"x\";\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := "my $u = undef;\n$u &&= 5;\nprint $u + 1;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
}

// A parenthesised declaration takes its initialiser in list context, so
// its first scalar is the list's first element, not the array's count:
// perl 5.42 -w is silent on `my ($fh) = @ARGV; open $fh, "<", $0; flock
// $fh, 1`, and `my @a = (1, 2); my ($x) = @a` makes $x 1. A hash's first
// element is a key, a Str. The element keeps what perl warns on: `my @u =
// (undef, undef); my ($v) = @u; $v + 1` warns "Use of uninitialized value
// $v in addition (+)".
func TestListAssignFromArrayIsElement(t *testing.T) {
	src := "my ($fh) = @ARGV;\nopen $fh, \"<\", $0 or die;\nflock $fh, 1;\n"
	for _, d := range loweredDiags(src) {
		t.Errorf("%q: diagnostic %q, want none", src, d.Message)
	}

	for _, c := range []struct {
		src, name string
		want      types.Type
	}{
		{"my @a = (1, 2);\nmy ($x) = @a;\n", "$x", types.Int},
		{"my %g = (k => 1);\nmy ($k) = %g;\n", "$k", types.Str},
		{"my ($y) = @ARGV;\n", "$y", types.Scalar},
	} {
		if got, _ := loweredSymbolType(t, c.src, c.name); got != c.want {
			t.Errorf("%q: %s is %v, want %v", c.src, c.name, got, c.want)
		}
	}

	src = "my @u = (undef, undef);\nmy ($v) = @u;\nprint $v + 1;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
}
