// ABOUTME: A bare /.../ matches against $_, and a trailing flag does not stop it being a match.
// ABOUTME: Requiring the text to END in `/` lost every pattern that carried g, i, o or x.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// TestBarePatternIsAMatchSite: a pattern with flags is still a match.
//
// isBarePattern's last case required `text[len(text)-1] == '/'`, so an
// unflagged `/abc/` was reported and a flagged `/abc/g` was not -- the text
// ends in the flag letter. Every pattern in re/pat.t carries a flag, which
// is why 35 of its sites scored WRONG.
//
// The flags perl accepts here are msixpodualngcer; only the ones that change
// whether it MATCHES matter to this marker, and none of them do -- a flagged
// match is still a match:
//
//	$ cat p.pl
//	$_ = "abc";
//	my @w = /(\w+)/g;
//	$ perl -MO=Concise,-exec p.pl | grep -c match
//	1
func TestBarePatternIsAMatchSite(t *testing.T) {
	for _, src := range []string{
		// Unflagged, which already worked.
		`my @words = /(\w+)/;`,

		// Flagged, which did not.
		`my @words = /(\w+)/g;`,
		`while (/\w+/g) { 1; }`,
		`$t1++ if /$pat1/o;`,
		`my @x = /abc/gi;`,
		`/\Gc/gc;`,
		`my @out = /(?<!foo)bar./g;`,

		// The m// spelling, flagged and not.
		`my @w = m/abc/;`,
		`my @w = m{abc}g;`,
	} {
		if !hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindMatch) {
			t.Errorf("%s\n  reports no match site; perl emits the match op", src)
		}
	}

	// What must NOT be claimed. perl reports s/// with subst and tr/// with
	// trans, and qr// compiles a pattern without matching it -- three
	// different ops, none of them this marker. The flags must not let any
	// of them through.
	for _, src := range []string{
		`s/a/b/;`,
		`s/a/b/g;`,
		`tr/a/b/;`,
		`tr/a/b/r;`,
		`y/a/b/;`,
		`my $re = qr/abc/;`,
		`my $re = qr/abc/i;`,
	} {
		if hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindMatch) {
			t.Errorf("%s\n  reports a match site; perl emits a different op", src)
		}
	}

	// A DIVISION is not a pattern. The lexer decides this by position and
	// the leaf's text is its answer, but a rule loosened to accept a
	// trailing flag must not start reading arithmetic as a regex.
	for _, src := range []string{
		`my $x = $a / $b;`,
		`my $x = 6/2/3;`,
	} {
		if hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindMatch) {
			t.Errorf("%s\n  reports a match site; this is division", src)
		}
	}
}
