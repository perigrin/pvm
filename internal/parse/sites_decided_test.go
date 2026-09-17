// ABOUTME: The subject must DECIDE markers, not only hedge them: exact=0 is a vacuous WRONG=0.
// ABOUTME: Each test asserts the site kind the parser reports for a construct perl reports too.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// decidedKinds returns the kinds reported WITHOUT Unresolved, which is what
// scores exact. Hedged sites score wider and are counted separately.
func decidedKinds(src string) []string {
	root := parse.Parse([]byte(src))
	var out []string
	for _, s := range parse.Sites(root, []byte(src)) {
		if !s.Unresolved {
			out = append(out, s.Kind)
		}
	}
	return out
}

func hedgedKinds(src string) []string {
	root := parse.Parse([]byte(src))
	var out []string
	for _, s := range parse.Sites(root, []byte(src)) {
		if s.Unresolved {
			out = append(out, s.Kind)
		}
	}
	return out
}

func has(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

// TestSitesDecideMarkers is the AC behind "exact is at least 40% of markers
// attempted, so WRONG=0 is not vacuous".
//
// Before this, Sites emitted ONLY hedges -- an Unresolved site of every kind
// for each Unknown statement, and nothing at all for a statement it parsed.
// That scores wider where perl found something and no-answer where it did
// not, so WRONG stays zero and exact stays zero with it. A subject that
// never commits cannot be wrong and cannot be right.
//
// Each case below is a construct perl reports with the marker op named in
// the comment, measured with -MO=Concise,-exec.
func TestSitesDecideMarkers(t *testing.T) {
	for _, c := range []struct {
		kind string
		srcs []string
	}{
		// perl: srefgen
		{parseoracle.SiteKindReference, []string{
			`my $r = \@a;`,
			`my $r = \$x;`,
			`my $r = \%h;`,
			`f(\@a);`,
		}},
		// perl: rv2hv
		{parseoracle.SiteKindHash, []string{
			`my %h = (a => 1);`,
			`my $n = keys %h;`,
			`my @k = values %h;`,
		}},
		// perl: match
		{parseoracle.SiteKindMatch, []string{
			`$x =~ /foo/;`,
			`$x =~ m/foo/;`,
			`if (/foo/) { 1 }`,
		}},
		// perl: readline
		{parseoracle.SiteKindReadline, []string{
			`my $l = <STDIN>;`,
			`my $l = <$fh>;`,
			`my @all = <$fh>;`,
		}},
		// perl: anonhash
		{parseoracle.SiteKindAnonhash, []string{
			`my $r = {a => 1};`,
			`my $r = {};`,
			`f({a => 1});`,
		}},
	} {
		for _, src := range c.srcs {
			got := decidedKinds(src)
			if !has(got, c.kind) {
				t.Errorf("%q: want a decided %s site, got %v",
					src, c.kind, got)
			}
		}
	}
}

// TestSitesDoNotDecideWhatTheyCannotSee is the negative scenario, and it is
// the one that keeps the metric honest.
//
// A subject that reported every marker on every statement would score exact
// everywhere perl agreed and WRONG everywhere it did not. These are
// statements where the marker's bytes appear but the construct does NOT --
// perl reports no such op, so neither may this parser.
func TestSitesDoNotDecideWhatTheyCannotSee(t *testing.T) {
	for _, c := range []struct {
		kind string
		srcs []string
	}{
		{parseoracle.SiteKindHash, []string{
			`my $n = $a % $b;`, // modulus, not a hash
		}},
		{parseoracle.SiteKindMatch, []string{
			`my $n = $a / $b;`, // division, not a pattern
		}},
		{parseoracle.SiteKindReadline, []string{
			`my $b = $a < $c;`, // comparison, not a readline
		}},
		{parseoracle.SiteKindAnonhash, []string{
			`if ($x) { 1 }`, // a block, not a constructor
			`sub f { 1 }`,   // a body, not a constructor
		}},
		{parseoracle.SiteKindReference, []string{
			`my $x = 1;`, // no reference taken
		}},
	} {
		for _, src := range c.srcs {
			if got := decidedKinds(src); has(got, c.kind) {
				t.Errorf("%q: must NOT decide %s, got %v", src, c.kind, got)
			}
		}
	}
}

// TestUnknownStillHedgesEveryMarker guards the rule Sites was written for:
// an Unknown statement must speak its refusal, because silence is scored as
// claiming nothing is there.
//
// This is the existing contract and it must survive deciding. The fixture is
// a construct this parser declines; if it ever starts parsing, replace it
// rather than deleting the test.
func TestUnknownStillHedgesEveryMarker(t *testing.T) {
	const src = "goto &other;\n" // Unknown at 3c997a6b; issue 01a0afc0

	root := parse.Parse([]byte(src))
	if !containsKind(root, parse.Unknown) {
		t.Skipf("%q now parses; this test needs a construct the parser declines", src)
	}

	hedged := hedgedKinds(src)
	for _, kind := range []string{
		parseoracle.SiteKindReference, parseoracle.SiteKindHash,
		parseoracle.SiteKindMatch, parseoracle.SiteKindReadline,
		parseoracle.SiteKindAnonhash,
	} {
		if !has(hedged, kind) {
			t.Errorf("an Unknown statement must hedge %s; got %v",
				kind, strings.Join(hedged, " "))
		}
	}
}
