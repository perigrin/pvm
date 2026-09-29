// ABOUTME: `$$` is the process id unless a name, digit, `$`, `{` or `::` follows it, so
// ABOUTME: `f($$, 1)` keeps its comma and `kill 9, $$;` its terminator.

package lexer

import "testing"

// TestPidVariableEndsAtItsSecondSigil holds toke.c's scan_ident rule for `$$`:
//
//	if (*s == '$' && s[1]
//	    && (   isIDFIRST_lazy_if_safe(s+1, PL_bufend, is_utf8)
//	        || isDIGIT_A((U8)s[1])
//	        || s[1] == '$'
//	        || s[1] == '{'
//	        || memBEGINs(s+1, (STRLEN) (PL_bufend - (s+1)), "::")) )
//	    /* Dereferencing a value in a scalar variable. */
//
// scanVarName treated every second `$` as a dereference and recursed, and the
// recursion then took the next byte as a one-byte punctuation name: `$$,`,
// `$$;` and `$$)` each lexed as ONE variable, swallowing the comma, the
// terminator or the closer. Measured on 5.42.0, `my @a = ($$, 1)` has two
// elements.
func TestPidVariableEndsAtItsSecondSigil(t *testing.T) {
	for _, src := range []string{"$$, 1", "$$;", "$$)", "$$ . 1"} {
		toks := significant(Tokenize([]byte(src)))
		if len(toks) == 0 || toks[0].Kind != Variable || toks[0].End != 2 {
			t.Errorf("%q: want the Variable `$$` alone, got %v", src, toks)
		}
	}
	// A dereference still lexes as one: a sigil, then the name it derefs.
	for _, src := range []string{"$$ref", "$$1", "$$::x"} {
		toks := significant(Tokenize([]byte(src + ";")))
		if len(toks) == 0 || toks[len(toks)-2].End != len(src) {
			t.Errorf("%q: the dereference must reach the end of the name, got %v", src, toks)
		}
	}
}
