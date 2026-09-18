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
		// perl: match. The variable form is a match too -- perl compiles
		// $expected as a pattern and reports the same op, measured:
		//	$ perl -MO=Concise,-exec -e 'my ($g,$e); my $r = $g =~ $e;'
		//	7  </> match()[$g:1,3] sK
		// t/comp/use.t:24 is this shape.
		{parseoracle.SiteKindMatch, []string{
			`$x =~ /foo/;`,
			`$x =~ m/foo/;`,
			`if (/foo/) { 1 }`,
			`$got =~ $expected;`,
			`$got !~ $expected;`,
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
			// perl reports these with subst and trans, not match, so a
			// binding operator alone is not enough to decide.
			`$x =~ s/a/b/;`,
			`$x =~ tr/a/b/;`,
			`$x =~ y/a/b/;`,
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

// TestSitesCountEveryOccurrence: one statement with two anonymous hashes
// must report TWO sites, not one.
//
// CompareFacts consumes a decided site per marker perl found
// (`compare_facts.go:224`, `owner.decided[m]--`), so a statement that reports
// one site for two of perl's scores the first exact and the second WRONG.
// Deduplicating per statement was the cause of 11 of the 20 WRONG files in
// T2's first subject measurement.
//
// Measured: `perl -MO=Concise,-exec` on the accessor.t line below reports
// two anonhash ops.
func TestSitesCountEveryOccurrence(t *testing.T) {
	for _, c := range []struct {
		src  string
		kind string
		want int
	}{
		// t/class/accessor.t:31, which scored WRONG on exactly this.
		{`ok(eq_hash({$o->h}, {qw( the hash )}), 'x');`,
			parseoracle.SiteKindAnonhash, 2},
		{`my @r = (\$a, \$b);`, parseoracle.SiteKindReference, 2},
		{`my @r = (\$a, \$b, \$c);`, parseoracle.SiteKindReference, 3},
		{`f({a=>1}, {b=>2});`, parseoracle.SiteKindAnonhash, 2},
		// One occurrence stays one.
		{`my $r = {a=>1};`, parseoracle.SiteKindAnonhash, 1},
	} {
		n := 0
		for _, k := range decidedKinds(c.src) {
			if k == c.kind {
				n++
			}
		}
		if n != c.want {
			t.Errorf("%q: want %d decided %s sites, got %d",
				c.src, c.want, c.kind, n)
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

// TestReferenceSitesTakeTheReference is the other half of a reference
// site's contract, and without it a reference site accounts for nothing.
//
// SubjectCallSite.TookReference is the question the srefgen marker asks
// (subject.go:119): did this site pass a reference rather than a flattened
// list? groupByStatement reads it as the decision --
//
//	case m == MarkerSrefgen && !c.TookReference:
//		// A committed call that passed no reference: an answer, but
//		// not one that accounts for a reference perl took.
//
// -- so a `reference` site left at the zero value lands in neither decided
// nor hedged. It answers the srefgen question, which lifts the file out of
// no-answer, and then explains none of perl's srefgens, which scores every
// one of them WRONG.
//
// A `\` this parser read IS a reference taken, so the field states that.
// The other four markers are decided by their kind alone and do not read
// this field.
func TestReferenceSitesTakeTheReference(t *testing.T) {
	for _, src := range []string{
		`my $r = \$x;`,
		`f(\@a);`,
		"sub f {\n    g(\\$destroyed, 1);\n}\n",
		`*pi = \undef;`, // opbasic/concat.t:100
	} {
		n := 0
		for _, s := range decidedSites(src) {
			if s.Kind != parseoracle.SiteKindReference {
				continue
			}
			n++
			if !s.TookReference {
				t.Errorf("%q: a reference site must set TookReference, "+
					"or the harness counts it as neither decided nor hedged", src)
			}
		}
		if n == 0 {
			t.Errorf("%q: want a decided reference site, got none", src)
		}
	}
}

// decidedSites returns the decided sites with their spans, which is what
// CompareFacts pools by.
func decidedSites(src string) []parseoracle.SubjectCallSite {
	root := parse.Parse([]byte(src))
	var out []parseoracle.SubjectCallSite
	for _, s := range parse.Sites(root, []byte(src)) {
		if !s.Unresolved {
			out = append(out, s)
		}
	}
	return out
}

// TestSitesReportTheirOwnLine is the line half of the contract, and it is
// what decides whether a decided site can ever be spent.
//
// CompareFacts attributes each of perl's sites to the INNERMOST subject
// statement covering its line (`compare_facts.go:349`) and spends a decided
// site there. perl's probe walks the CVs, so a backslash in a sub body is
// reported at the body's line, not at the `sub` keyword's. A subject that
// pools every site under the enclosing top-level statement's whole span
// offers nothing at the inner line for that site to be spent against, and
// the verdict is WRONG -- for a marker the parser read correctly.
//
// So a site's line is the line of the NODE that decided it.
func TestSitesReportTheirOwnLine(t *testing.T) {
	for _, c := range []struct {
		src  string
		kind string
		line int
	}{
		// The backslash is on line 2; `sub foo {` is line 1.
		{"sub foo {\n    my $r = \\$x;\n}\n",
			parseoracle.SiteKindReference, 2},
		{"if (1) {\n    my $r = {a=>1};\n}\n",
			parseoracle.SiteKindAnonhash, 2},
		{"while (1) {\n\n    my @k = keys %h;\n}\n",
			parseoracle.SiteKindHash, 3},
		// A site in the continuation of a multi-line statement belongs to
		// its own line too, not to the line the statement opened on.
		{"f(\n    \\$x,\n);\n",
			parseoracle.SiteKindReference, 2},
	} {
		var got []int
		for _, s := range decidedSites(c.src) {
			if s.Kind == c.kind {
				got = append(got, s.Line)
			}
		}
		if len(got) != 1 || got[0] != c.line {
			t.Errorf("%q: want one decided %s at line %d, got lines %v",
				c.src, c.kind, c.line, got)
		}
	}
}
