// ABOUTME: A lexical sub named like a quote operator shadows it in its block:
// ABOUTME: after `my sub s`, `s(1)` is a call, and after the block `s///` is back.

package lexer

import "testing"

// TestLexicalSubShadowsQuoteOp: measured on 5.42.0,
//
//	$ perl -e '{ my sub s { 42 } print s(1), "\n"; } $_ = "a"; s/a/b/;
//	      print "$_\n"; { my sub y { 43 } print y, "\n"; }'
//	42
//	b
//	43
//
// perl.git t/op/lvref.t:653 (`my sub s ($arg)`) and t/op/lexsub.t:869
// (`my sub y :prototype() {$x}; is y, 43`).
func TestLexicalSubShadowsQuoteOp(t *testing.T) {
	src := `{ my sub s { 42 } print s(1); } s/a/b/; { my sub y { 43 } is y, 43; }`
	var words, quotes []string
	for _, tok := range Tokenize([]byte(src)) {
		text := src[tok.Start:tok.End]
		switch tok.Kind {
		case Word:
			if text == "s" || text == "y" {
				words = append(words, text)
			}
		case Quote:
			quotes = append(quotes, text)
		}
	}
	// `sub s`, the call `s`, `sub y` and the call `y` are words; `s/a/b/`
	// outside the first block is the one substitution.
	if len(words) != 4 {
		t.Errorf("want s, s, y, y as words, got %q", words)
	}
	if len(quotes) != 1 || quotes[0] != "s/a/b/" {
		t.Errorf("want the one substitution s/a/b/ outside the block, got %q", quotes)
	}
}
