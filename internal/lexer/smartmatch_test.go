// ABOUTME: `~~` is the smartmatch operator in operator position and two bitwise nots in term
// ABOUTME: position, as toke.c decides it from PL_expect.

package lexer

import "testing"

// TestSmartmatchIsOneOperator holds toke.c's rule for `~~`:
//
//	if (FEATURE_SMARTMATCH_IS_ENABLED &&
//	    s[1] == '~' && (PL_expect == XOPERATOR || PL_expect == XTERMORDORDOR))
//	    ... NCEop(OP_SMARTMATCH);
//
// The lexer emitted two `~` tokens, so `$x ~~ @a` split into two statements at
// Unknown=0. In TERM position `~~$x` is still two bitwise nots -- measured on
// 5.42.0, `my $n = ~~$x;` deparses unchanged -- and that idiom must survive.
func TestSmartmatchIsOneOperator(t *testing.T) {
	toks := significant(Tokenize([]byte("$x ~~ @a")))
	if len(toks) != 3 || string("$x ~~ @a"[toks[1].Start:toks[1].End]) != "~~" {
		t.Errorf("`$x ~~ @a`: want one `~~` operator between the operands, got %d tokens", len(toks))
	}
	src := "my $n = ~~$x;"
	toks = significant(Tokenize([]byte(src)))
	if src[toks[3].Start:toks[3].End] != "~" || src[toks[4].Start:toks[4].End] != "~" {
		t.Errorf("%q: in term position `~~` is two bitwise nots", src)
	}
}
