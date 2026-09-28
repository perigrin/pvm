// ABOUTME: The repeat operator touching its count: `"ab"x4` is `x` then `4` in operator position.
// ABOUTME: toke.c's `case 'x'` splits it before keyword lookup; in term position `x4` stays a word.

package lexer

import "testing"

// TestRepeatTouchingItsCount holds toke.c's rule for an `x` followed by a
// digit:
//
//	case 'x':
//	    if (isDIGIT(s[1]) && PL_expect == XOPERATOR) {
//	        s++;
//	        Mop(OP_REPEAT);
//	    }
//
// It runs BEFORE keyword lookup, so no declaration changes it -- measured on
// 5.42.0, `sub x4 {9} my $x = "ab" x4;` deparses to `my $x = 'abababab';`.
// Scanned as one word, `"ab" x4` was a string followed by a call to x4, at
// Unknown=0.
func TestRepeatTouchingItsCount(t *testing.T) {
	for _, src := range []string{`"ab"x4`, `"ab" x4`, `$s x10`} {
		toks := significant(Tokenize([]byte(src)))
		if len(toks) != 3 {
			t.Errorf("%q: %d significant tokens, want 3 (term, x, count)", src, len(toks))
			continue
		}
		if op := src[toks[1].Start:toks[1].End]; op != "x" {
			t.Errorf("%q: operator is %q, want \"x\"", src, op)
		}
		if toks[2].Kind != Number {
			t.Errorf("%q: count is %v, want Number", src, toks[2].Kind)
		}
	}

	// In TERM position there is no operator to form: `x4` is a word, a call
	// or a bareword. perl's rule is gated on XOPERATOR for exactly this.
	src := "x4;"
	toks := significant(Tokenize([]byte(src)))
	if len(toks) == 0 || toks[0].Kind != Word || src[toks[0].Start:toks[0].End] != "x4" {
		t.Errorf("%q: statement-start `x4` must stay one Word, got %v", src, toks)
	}
}
