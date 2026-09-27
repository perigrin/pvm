// ABOUTME: A bareword hash key spelled like a quote-like operator is still a key.
// ABOUTME: `$h{m}` autoquotes; the lexer read it as a match delimited by `}`.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// quoteOpKeyNames is perl's whole quote-like set. Every one of them
// autoquotes as a lone bareword subscript -- measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my %h; my $a=$h{m}; my $b=$h{q}; my $c=$h{qq};
//	                       my $d=$h{qw}; my $e=$h{qx}; my $f=$h{qr};
//	                       my $g=$h{s}; my $i=$h{tr}; my $j=$h{y};'
//	  ->  $h{'m'} $h{'q'} $h{'qq'} $h{'qw'} $h{'qx'} $h{'qr'}
//	      $h{'s'} $h{'tr'} $h{'y'}
//
// The list is the table in lexer/quote.go, so a name added there without a
// case here is a name this test does not hold.
var quoteOpKeyNames = []string{"q", "qq", "qw", "qx", "m", "qr", "s", "tr", "y"}

// TestBarewordSubscriptKey: a subscript holding one bareword that happens to
// name a quote-like operator is the STRING, not the operator.
//
// `$h{m}` lexed as a match whose delimiter was `}`, so the body ran to the
// NEXT `}` and the emission gained a spurious `};` while the enclosing
// construct lost its closer. `$h{m}{q}` was worse: it swallowed BOTH
// subscripts into one Quote and still scored Unknown=0, so the count alone
// could not see it. The canon is the check.
//
// A `}` never delimits one of these operators in a program perl will
// COMPILE, in any context and not just a subscript -- measured on 5.42.0,
// where each of these is fatal:
//
//	$ perl -e 'sub f { m }'
//	Search pattern not terminated at -e line 1.
//	$ perl -e 'my @a = map { qw } 1;'
//	Can't find string terminator "}" anywhere before EOF at -e line 1.
//
// so the decline needs no bracket-stack knowledge. The subscript is the only
// spelling that compiles at all, and the other spellings were already
// refusals; declining changes which refusal they get, not whether they are
// one.
func TestBarewordSubscriptKey(t *testing.T) {
	for _, name := range quoteOpKeyNames {
		for _, tc := range []struct{ src, want string }{
			{"my $v = $h{" + name + "};", "my $v = $h{" + name + "};"},
			{"my $v = $t->{" + name + "};", "my $v = $t->{" + name + "};"},
			{"$t->{" + name + "} = 1;", "$t->{" + name + "} = 1;"},
			{"my @v = @h{" + name + "};", "my @v = @h{" + name + "};"},
		} {
			b := []byte(tc.src)
			root := parse.Parse(b)
			if got := strings.TrimSpace(parse.Canon(root, b)); got != tc.want {
				t.Errorf("%q\n  canon got  %q\n        want %q",
					tc.src, got, tc.want)
			}
			if n := countUnknown(root); n != 0 {
				t.Errorf("%q: %d Unknown nodes; the key is a string", tc.src, n)
			}
			for _, tok := range lexer.Tokenize(b) {
				switch tok.Kind {
				case lexer.Quote, lexer.UnknownRest, lexer.Error:
					t.Errorf("%q: lexed %v %q; the key is a bareword",
						tc.src, tok.Kind, tc.src[tok.Start:tok.End])
				}
			}
		}
	}

	// A CHAIN is the case the Unknown count cannot see: `$h{m}{q}` scored
	// zero while one Quote token held `m}{q}`.
	b := []byte(`my $v = $h{m}{q};`)
	if got := strings.TrimSpace(parse.Canon(parse.Parse(b), b)); got != `my $v = $h{m}{q};` {
		t.Errorf("a subscript chain: canon got %q", got)
	}

	// Whitespace inside the braces changes nothing. perl deparses
	// `$h{ m }` to `$h{'m'}`.
	b = []byte(`my $v = $h{ m };`)
	if got := strings.TrimSpace(parse.Canon(parse.Parse(b), b)); !strings.Contains(got, "{m") {
		t.Errorf("`$h{ m }`: canon got %q", got)
	}
}

// TestQuoteOpStillLexesWithBraceDelimiter is the mutation guard: a REAL
// quote operator whose delimiter happens to be `{` is untouched, because
// what follows the name there is `{` and not `}`.
//
//	$ perl -MO=Deparse -e 'my %h; my $z = $h{ m{a} };'
//	my $z = $h{/a/};
//
// A bareword key and a brace-delimited match sit in the same subscript and
// perl tells them apart by the byte after the name. So does this.
func TestQuoteOpStillLexesWithBraceDelimiter(t *testing.T) {
	for _, src := range []string{
		`my $z = $h{ m{a} };`,
		`$x =~ m{abc};`,
		`$x =~ s{a}{b};`,
		`my @w = qw{a b};`,
		`my $q = q{a};`,
		`$x =~ tr{a}{b};`,
	} {
		var sawQuote bool
		for _, tok := range lexer.Tokenize([]byte(src)) {
			switch tok.Kind {
			case lexer.Quote:
				sawQuote = true
			case lexer.UnknownRest, lexer.Error:
				t.Errorf("%q: lexed %v %q", src, tok.Kind, src[tok.Start:tok.End])
			}
		}
		if !sawQuote {
			t.Errorf("%q: no Quote token; a brace-delimited operator is still "+
				"an operator", src)
		}
	}
}
