// ABOUTME: Tests for the round-2 false positives cutover 1 adds on the lowered path: Int slots, fileno's handle, a sub's returned result, readline's value.
// ABOUTME: Each silent case is one perl 5.42 -w runs silently, paired with a control perl warns about or dies on.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
)

// perl truncates any finite number to an integer silently where it takes
// one, and arithmetic that only might overflow is checked as the finite
// number it almost always is. perl 5.42 -w is silent on each case below:
// splice's length, a shift count, a bitwise operand, a constant sub's
// `2**1`, which perl folds to 2.0, and a broad Scalar or Any.
func TestMaybeInfInIntSlotIsSilent(t *testing.T) {
	for _, src := range []string{
		"my @h = (1 .. 5); my $histsize = 2;\nsplice @h, 0, @h - $histsize if @h > $histsize;\nmy $x = 1 << (4 * 7);\n",
		"use constant { PUBLIC => 1 << 0, PRIVATE => 1 << 1 };\nmy $fattr = 3;\nprint \"private\\n\" if $fattr & PRIVATE;\n",
		"sub PRIVATE () { 2**1 }\nmy $battr = 3;\nprint \"private\\n\" if $battr & PRIVATE;\n",
		"my @h = (1 .. 5);\nsplice @h, 0, 2.5;\nmy $q = 20 / 8;\nmy $x = 1 << $q;\n",
		"my @h = (1 .. 5);\nmy $n = shift;\nsplice @h, 0, $n;\nmy $x = 1 << $n;\n",
		"my @h = (1 .. 5);\nmy $n = eval '2.5';\nsplice @h, 0, $n;\nmy $x = 1 << $n;\n",
	} {
		wantSilent(t, src)
	}

	// perl 5.42 -w warns "Use of uninitialized value in splice" and "in
	// left bitshift (<<)", and "Argument "abc" isn't numeric".
	wantReported(t, "my @h = (1 .. 5);\nsplice @h, 0, undef;\n", infer.CodeTypeMismatch)
	wantReported(t, "my $x = 1 << undef;\n", infer.CodeCoercionMismatch)
	wantReported(t, "my $x = 1 << \"abc\";\n", infer.CodeCoercionMismatch)
}

// A user sub whose every returned value is a builtin's result hands its
// caller that result, and a slot judges it as it judges the builtin's own:
// without the Undef the builtin gives only when it fails. perl 5.42 -w is
// silent on each case below, explicit return or implicit, and on a broad
// Scalar or a Hole's Any returned the same way.
func TestUserSubReturningBuiltinResult(t *testing.T) {
	for _, src := range []string{
		"sub xfileno { return fileno $_[0] }\nprint \"same\\n\" if xfileno(\\*STDOUT) == fileno \\*STDOUT;\n",
		"sub xfileno { fileno $_[0] }\nprint \"same\\n\" if xfileno(\\*STDOUT) == 1;\n",
		"sub xs { my $v = shift; return $v }\nprint \"x\\n\" if xs(1) == 1;\n",
		"sub xa { return eval '1' }\nprint \"x\\n\" if xa() == 1;\n",
	} {
		wantSilent(t, src)
	}

	// perl 5.42 -w warns "Use of uninitialized value in numeric eq (==)"
	// on each: the sub's Undef is its own, not a failing builtin's.
	for _, src := range []string{
		"sub xf { return undef }\nprint \"x\\n\" if xf() == 1;\n",
		"sub xf { return fileno($_[0]) if @_; return }\nprint \"x\\n\" if xf() == 1;\n",
	} {
		wantReported(t, src, infer.CodeCoercionMismatch)
	}
}

// readline in print's list hands print the lines it reads, each a Str,
// and none at all at end of file: the empty list is no argument, so it is
// never Void there. perl 5.42 -w is silent on each case below, and on a
// broad Scalar or a Hole's Any printed the same way.
func TestPrintOfReadlineIsNotVoid(t *testing.T) {
	for _, src := range []string{
		"my $laststash = \"main\";\nprint readline(*{$laststash.\"::DATA\"});\n__DATA__\nline\n",
		"print readline(*DATA);\n__DATA__\nline\n",
		"my $s = shift;\nprint $s, readline(*DATA);\n__DATA__\nline\n",
		"print eval 'x', readline(*DATA);\n__DATA__\nline\n",
	} {
		wantSilent(t, src)
	}

	// perl 5.42 -w dies "Not a GLOB reference" on a hash reference as
	// readline's handle.
	wantCode(t, "print readline({});\n", infer.CodeTypeMismatch)
	// perl 5.42 -w warns "Use of uninitialized value in print" on an
	// array of undef flattened into print's list.
	wantCode(t, "my @a = (undef);\nprint @a;\n", infer.CodeTypeMismatch)
}
