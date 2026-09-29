// ABOUTME: `$#+` and `$#-` are the last index of @+ and @-, one Variable each, as toke.c
// ABOUTME: reads `$#` before any of `{ $ : + - @` or an identifier.

package lexer

import "testing"

// TestLastIndexOfPunctuationArray holds toke.c's rule for `$#`:
//
//	if (   s[1] == '#'
//	    && (   isIDFIRST_lazy_if_safe(s+2, PL_bufend, UTF)
//	        || memCHRs("{$:+-@", s[2])))
//
// The lexer took the last-index form only before a word byte, `{` or `$`, so
// `$#+` lexed as the variable `$#` followed by a `+` operator. Measured on
// 5.42.0, after `"ab" =~ /(a)(b)/`, `print $#+, " ", $#-` prints `2 2`.
func TestLastIndexOfPunctuationArray(t *testing.T) {
	for _, src := range []string{"$#+", "$#-", "$#@", "$#::x"} {
		toks := significant(Tokenize([]byte(src + ";")))
		if len(toks) == 0 || toks[0].Kind != Variable || toks[0].End != len(src) {
			t.Errorf("%q: want one Variable spanning it, got %v", src, toks)
		}
	}
	// `$#` alone before anything else is still the variable named `#`.
	toks := significant(Tokenize([]byte("$# = 1;")))
	if toks[0].End != 2 {
		t.Errorf("`$# = 1`: `$#` is its own variable")
	}
}
