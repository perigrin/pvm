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
// COMPILE, and outside a subscript perl refuses -- measured on 5.42.0,
// where each of these is fatal:
//
//	$ perl -e 'sub f { m }'
//	Search pattern not terminated at -e line 1.
//	$ perl -e 'my @a = map { qw } 1;'
//	Can't find string terminator "}" anywhere before EOF at -e line 1.
//
// so the decline is for a subscript's brace only, and the other spellings
// lex as the unterminated operator perl refuses.
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

	// The INTERPOLATED spelling was already right when the bare one was not.
	// perl deparses `"x $t->{m} y"` to `"x $t->{'m'} y"`.
	b = []byte(`my $s = "x $t->{m} y";`)
	root := parse.Parse(b)
	if got := strings.TrimSpace(parse.Canon(root, b)); got != `my $s = "x $t->{m} y";` {
		t.Errorf("an interpolated subscript: canon got %q", got)
	}
	if n := countUnknown(root); n != 0 {
		t.Errorf("an interpolated subscript: %d Unknown nodes, want 0", n)
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

// TestNotSubscriptKey: `not` filling a subscript is the key "not", as every
// lone word there is, and builds no missing_operand (R-g′). perl 5.44.0's
// own B/Op_private.pm holds `$bits{not}{0} = $bf[0];`, and -MO=Deparse
// prints `$h{not}` as `$h{'not'}`. Outside a subscript, and before anything
// but the `}`, it is still the operator.
func TestNotSubscriptKey(t *testing.T) {
	// Each canons as the same source with another word for the key does.
	for _, src := range []string{"my %h; $h{not}", "$bits{not}{0} = $bf[0];", "my %h; $h{ not } = 1;", "my $r; $r->{not}"} {
		want := strings.ReplaceAll(clean(t, strings.ReplaceAll(src, "not", "and")), "and", "not")
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
	if got := clean(t, "my %h; $h{not 1}"); !strings.Contains(got, "not 1") {
		t.Errorf("`$h{not 1}`: canon %q; `not` before a term is the operator", got)
	}
	if got := clean(t, "my %h; $h{-not}"); !strings.Contains(got, "-not}") {
		t.Errorf("`$h{-not}`: canon %q", got)
	}
	clean(t, "my %h; $h{- not}")
	// Only a `not` first in the subscript, or after a lone `-`, is the key:
	// perl 5.44.0 has each of these "syntax error near "not }"".
	for _, src := range []string{"my $x = not;", "my %h; $h{1, not}", "my %h; my $x; $h{ $x ? 1 : not }", "my %h; my $x; $h{ $x // not }", "my %h; $h{(not)}"} {
		if root := parse.Parse([]byte(src)); !root.HasError && len(collectUnknown(root)) == 0 {
			t.Errorf("%q: perl refuses this, but it parses", src)
		}
	}
}
