// ABOUTME: A quote-op keyword before a fat comma is a string, not an operator.
// ABOUTME: `{ s => 1 }` is a hash key; reading `s` as a substitution swallows the rest.

package lexer_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// TestQuoteOpBeforeFatCommaIsAName: `=>` quotes the word to its left, and
// that includes the quote-op keywords.
//
// Measured on perl 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $h = { s => 1, y => 2, q => 3, tr => 4, m => 5 };'
//	my $h = {'s', 1, 'y', 2, 'q', 3, 'tr', 4, 'm', 5};
//
// All five autoquote. Without this, `s` took `=` as its delimiter and ran to
// the next `=`, turning `{ s => 1 }` into a substitution and everything
// after it into one opaque token. t/comp/parser_run.t:17 carries
// `{ stderr => 1 }` -- which is fine, `stderr` is not a keyword -- but the
// same file's other hashes were not, and the whole statement went Unknown.
//
// §4.5.4 already states the rule for barewords generally; this is the
// keyword case of it.
func TestQuoteOpBeforeFatCommaIsAName(t *testing.T) {
	for _, src := range []string{
		`my $h = { s => 1 };`,
		`my $h = { y => 1 };`,
		`my $h = { q => 1 };`,
		`my $h = { tr => 1 };`,
		`my $h = { m => 1 };`,
		`my $h = { qq => 1 };`,
		`f(s => 1);`,
		`my %h = (s => 1, m => 2);`,
		// Spacing must not matter: perl skips whitespace before `=>`.
		`my $h = { s    => 1 };`,
	} {
		for _, tok := range lexer.Tokenize([]byte(src)) {
			if tok.Kind == lexer.UnknownRest || tok.Kind == lexer.Error {
				t.Errorf("%q: %v %q -- a quote-op keyword before `=>` is a name",
					src, tok.Kind, src[tok.Start:tok.End])
			}
		}
	}
}

// TestQuoteOpStillOperatesWhereItShould is the negative scenario. Declining
// before `=>` must not decline anywhere else: these are real quote operators
// and must keep lexing as one token each.
func TestQuoteOpStillOperatesWhereItShould(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{`$x =~ s/a/b/;`, "s/a/b/"},
		{`$x =~ tr/a/b/;`, "tr/a/b/"},
		{`my $s = q(text);`, "q(text)"},
		{`my $m = m/pat/;`, "m/pat/"},
		{`my @w = qw(a b);`, "qw(a b)"},
		// `=` alone is not `=>`, so this stays a substitution.
		{`$x =~ s=a=b=;`, "s=a=b="},
	} {
		var found bool
		for _, tok := range lexer.Tokenize([]byte(c.src)) {
			if strings.Contains(c.src[tok.Start:tok.End], c.want) {
				found = true
			}
		}
		if !found {
			t.Errorf("%q: want a token containing %q; the quote operator "+
				"must still lex where nothing quotes it", c.src, c.want)
		}
	}
}
