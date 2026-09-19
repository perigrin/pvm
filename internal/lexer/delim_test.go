// ABOUTME: A quote delimiter is a CHARACTER — a multi-byte one must not close on its first byte.
// ABOUTME: Only the non-paired forms; perl itself rejects a paired multi-byte delimiter.

package lexer_test

import (
	"testing"
	"unicode/utf8"

	"tamarou.com/pvm/internal/lexer"
)

// TestMultiByteDelimiterClosesOnTheCharacter: a delimiter is compared whole.
//
// `ϟ` is 0xCF 0x9F. Matching a byte at a time stops at the 0xCF of the
// CLOSING delimiter and leaves 0x9F behind as an Error token, so the Quote
// ends mid-sequence -- a span that cannot be re-lexed or shown to a user.
//
// Measured on perl 5.42.0; only the non-paired forms are accepted, and each
// of these was run:
//
//	qq ϟ hello ϟ        works
//	q ϟ raw ϟ           works
//	s ϟ b ϟ X ϟ         works, three parts
//	q«paired»           "Use of '«' is deprecated as a string delimiter"
//	                    then "Can't find string terminator"
func TestMultiByteDelimiterClosesOnTheCharacter(t *testing.T) {
	for _, c := range []struct{ name, src string }{
		{"qq", "qq ϟ a ϟ"},
		{"q", "q ϟ a ϟ"},
		{"m", "m ϟ a ϟ"},
		{"three-part s", "s ϟ a ϟ b ϟ"},
		{"three-part tr", "tr ϟ a ϟ b ϟ"},
	} {
		t.Run(c.name, func(t *testing.T) {
			toks := lexer.Tokenize([]byte(c.src))

			var quotes, errors int
			for _, tok := range toks {
				switch tok.Kind {
				case lexer.Quote:
					quotes++
				case lexer.Error, lexer.UnknownRest:
					errors++
				}
			}
			if quotes != 1 || errors != 0 {
				t.Errorf("%q: %d Quote and %d Error tokens, want 1 and 0",
					c.src, quotes, errors)
			}

			// And the Quote must span the WHOLE construct.
			for _, tok := range toks {
				if tok.Kind != lexer.Quote {
					continue
				}
				if got := c.src[tok.Start:tok.End]; got != c.src {
					t.Errorf("Quote spans %q, want the whole %q", got, c.src)
				}
			}
		})
	}
}

// TestNoTokenEndsMidSequence: every token boundary falls on a UTF-8 character
// boundary.
//
// The property rather than the instance: a boundary inside a sequence is
// unusable whatever produced it, and this catches the next such bug without
// naming it.
func TestNoTokenEndsMidSequence(t *testing.T) {
	for _, src := range []string{
		"qq ϟ a ϟ",
		"q ϟ raw ϟ",
		"s ϟ a ϟ b ϟ",
		"my $x = qq ϟ a ϟ;",
		"${\n#line 57\nqq ϟϟ }\n",
		`my $s = "héllo wörld";`,
		"m/ünicode/",
		"qw(ä ö ü)",
	} {
		b := []byte(src)
		for _, tok := range lexer.Tokenize(b) {
			for _, at := range []struct {
				where string
				off   int
			}{{"Start", tok.Start}, {"End", tok.End}} {
				if at.off < 0 || at.off > len(b) {
					t.Errorf("%q: %s %d out of range", src, at.where, at.off)
					continue
				}
				if at.off < len(b) && !utf8.RuneStart(b[at.off]) {
					t.Errorf("%q: token %v %s=%d falls inside a UTF-8 sequence",
						src, tok.Kind, at.where, at.off)
				}
			}
		}
	}
}

// TestQuoteKeywordFollowedByAWideLetterIsAName: a quote keyword followed by a
// MULTI-BYTE word character is part of an identifier, not an operator.
//
// t/uni/gv.t:488 is the case, and it is why this rule has its own test:
//
//	my $rv = \*sምḲ;
//
// Under `use utf8` perl widens the identifier class, so `sምḲ` is ONE name and
// the leading `s` is not a substitution. Verified by running it:
//
//	use utf8; sub sምḲ { 42 } print sምḲ();   ->  42
//
// Testing only the next BYTE reads `ም` as a delimiter, and the substitution
// then runs to EOF -- 10,124 bytes of gv.t swallowed by a single token. That
// regression scored BETTER on the node ratchet (5 -> 1) while being far
// worse, which is the blind spot the byte measurement exists to catch.
func TestQuoteKeywordFollowedByAWideLetterIsAName(t *testing.T) {
	for _, c := range []struct {
		name, src string
		wantQuote bool
	}{
		{"utf8 name beginning with s", "use utf8;\nmy $rv = \\*sምḲ;\n", false},
		{"utf8 name beginning with q", "use utf8;\nmy $x = qምḲ;\n", false},
		{"utf8 name beginning with tr", "use utf8;\nmy $x = trምḲ;\n", false},
		// Without the pragma the wide bytes are not identifier characters,
		// so the keyword IS an operator and the character delimits.
		{"no pragma, s delimits", "my $x = s ϟ a ϟ b ϟ;\n", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			var quotes int
			for _, tok := range lexer.Tokenize([]byte(c.src)) {
				if tok.Kind == lexer.Quote {
					quotes++
				}
			}
			if got := quotes > 0; got != c.wantQuote {
				t.Errorf("%q: %d Quote tokens, want any = %v",
					c.src, quotes, c.wantQuote)
			}
		})
	}
}

// TestAsciiDelimiterStillWorks: the single-byte path is unchanged.
func TestAsciiDelimiterStillWorks(t *testing.T) {
	for _, src := range []string{
		"qq ! a !",
		"q/raw/",
		"s/a/b/",
		"s{a}{b}",
		// GLUED, which is the form where `#` delimits. Spaced, `qq #a#` is a
		// comment and the delimiter is whatever follows on a later line --
		// measured, and the reason this row is not `qq #a#`:
		//
		//	$ perl -e 'print q #comment
		//	xfoox'
		//	foo
		"qq#a#",
		`s\a\b\`,
		"tr/a/b/",
		"m,a,",
	} {
		toks := lexer.Tokenize([]byte(src))
		var quotes int
		for _, tok := range toks {
			if tok.Kind == lexer.Quote {
				quotes++
				if got := src[tok.Start:tok.End]; got != src {
					t.Errorf("%q: Quote spans %q, want all of it", src, got)
				}
			}
		}
		if quotes != 1 {
			t.Errorf("%q: %d Quote tokens, want 1", src, quotes)
		}
	}
}
