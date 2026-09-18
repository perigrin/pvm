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
			// perl takes ANY non-word character as m's delimiter and reports
			// match for every one of them. Measured on 5.42.0:
			//	$ perl -MO=Concise,-exec -e 'my $a = m(x); my $b = m[x];
			//	                             my $c = m!x!; my $d = m,x,;'
			//	 -- four match ops.
			// comp/opsubs.t:119 is `isnt( m('unqualified'), ... )`, which
			// scored WRONG while only `m/` and `m{` counted.
			`my $r = m(foo);`,
			`my $r = m[foo];`,
			`my $r = m!foo!;`,
			`my $r = m'foo';`,
			`isnt( m('x'), "y", "z" );`,
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

// TestBindingDecidesWhateverTheOperandShape: `=~` is the match, and the
// operand's SHAPE does not change that.
//
// `isMatchOperand` required a `Term` on the right, so a binding against
// anything built -- a ternary of two `qr//`, a parenthesised expression, a
// concatenation -- decided nothing and scored WRONG wherever perl found the
// op. Measured on perl 5.42.0:
//
//	$ perl -MO=Concise,-exec -e 'my $x; print $x =~ ((1 & 1) ? qr/^$/ : qr/o/);'
//	  one match, one qr, one regcomp
//
// perl reports ONE match for the binding. The `qr//` compile a pattern and
// are separately excluded from deciding match on their own (they are not a
// match until something binds them), but that exclusion is about the qr
// LEAF, not about what the binding does.
//
// t/cmd/for.t:79 is this shape and was the last non-M4 match failure.
func TestBindingDecidesWhateverTheOperandShape(t *testing.T) {
	for _, src := range []string{
		`print $x =~ ((f() & 1) ? qr/^$/ : qr/other/);`,
		`my $r = $x =~ ($cond ? $a : $b);`,
		`my $r = $x =~ ("a" . "b");`,
		`my $r = $x =~ $h->{pat};`,
	} {
		if got := decidedKinds(src); !has(got, parseoracle.SiteKindMatch) {
			t.Errorf("%q: `=~` is the match whatever the operand's shape; got %v",
				src, got)
		}
	}
}

// TestBindingStillExcludesSubstAndTrans is the negative half. Widening the
// operand test must not let s/// or tr/// through: perl reports those with
// subst and trans, which are not the match marker.
func TestBindingStillExcludesSubstAndTrans(t *testing.T) {
	for _, src := range []string{
		`$x =~ s/a/b/;`,
		`$x =~ tr/a/b/;`,
		`$x =~ y/a/b/;`,
		`$x =~ s(a)(b);`,
	} {
		if got := decidedKinds(src); has(got, parseoracle.SiteKindMatch) {
			t.Errorf("%q: must NOT decide match, got %v", src, got)
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
			// The same delimiters that make `m(...)` a match leave these
			// three what they already were. Measured:
			//	$ perl -MO=Concise,-exec -e 'my $x; $x =~ s(a)(b);
			//	                             $x =~ tr(a)(b); my $q = qr(a);'
			//	subst, trans, qr -- no match op among them.
			`$x =~ s(a)(b);`,
			`$x =~ tr(a)(b);`,
			`my $q = qr(a);`,
			// `qw(...)` is a list, and its leading `q` must not read as a
			// match with `w` for a delimiter.
			`my @w = qw(a b);`,
			// A NAME beginning with a quote-op keyword is a name. These are
			// what the delimiter's word-character test protects: without it
			// `sort` reads as `s` delimited by `o`, `my(...)` as `m`
			// delimited by `y`, and `tr` as `t` -- every one a match or an
			// exclusion invented from a name.
			`my @s = sort @a;`,
			`my $string = 1;`,
			`mkdir("d");`,
			`$yes = 1;`,
			`trim($x);`,
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

// TestOpaqueDerefBlockHedges: a `${...}` or `@{...}` whose contents the
// lexer swallowed must HEDGE, not stay silent.
//
// The lexer brace-matches a braced name to its closer and emits one Variable
// token (`internal/lexer/scan.go:77-94`, whose comment says the full rule
// "belongs to a later issue" -- 01a0ad52). So `${[{a=>214}]}` is one opaque
// Term and the anonymous hash inside it is invisible to a walk over nodes.
//
// Silence there is not a refusal. It is the claim that nothing is inside,
// and it is wrong wherever perl found something -- the exact failure this
// file's doc comment is about, arriving through a node that is not marked
// Unknown. Measured on perl 5.42.0:
//
//	$ perl -MO=Concise,-exec -e 'my $x = ${[{a=>214}]}[0];'
//	  one anonhash
//	$ perl -MO=Concise,-exec -e 'our @x; my @k=("x"); my @g = @{$::{$k[0]}};'
//	  two hash ops
//
// t/comp/parser.t:574 and t/comp/retainedlines.t:70 are those two lines, and
// both scored WRONG while this reported nothing.
//
// This hedges rather than decides on purpose: the parser genuinely cannot
// see the contents, so `wider` is the honest verdict. When 01a0ad52 gives
// the block real children, these become decided and the hedge stops firing.
func TestOpaqueDerefBlockHedges(t *testing.T) {
	for _, src := range []string{
		`my $x = ${[{a=>214}]}[0];`,
		`my @g = @{$::{$keys[0]}};`,
		`my $r = ${$h->{k}};`,
		`my @a = @{[ 1, 2 ]};`,
	} {
		hedged := hedgedKinds(src)
		for _, kind := range markerKindsForTest {
			if !has(hedged, kind) {
				t.Errorf("%q: an opaque deref block must hedge %s; got %v",
					src, kind, hedged)
			}
		}
	}
}

// TestPlainDerefDoesNotHedge is the negative scenario. A dereference with
// nothing hidden inside it -- `${$x}`, `$$x` -- has no contents to be blind
// to, so hedging there would be noise that scores wider where the subject
// could have been silent and correct.
func TestPlainDerefDoesNotHedge(t *testing.T) {
	for _, src := range []string{
		`my $r = ${$x};`,
		`my $r = $$x;`,
		`my @a = @$r;`,
		`my $n = 1;`,
		// A subscript carries the same bracket characters and must not hedge:
		// hedging it would trade exact for wider on every hash and array
		// access in the corpus.
		//
		// These pass whatever `hidesStructure` does, and that is worth saying
		// rather than implying otherwise. `$h{k}` parses to an Index whose
		// Term child is `$h` -- the braces belong to the Index node, so the
		// leaf text `hidesStructure` sees is never `"$h{k}"`. Its
		// `text[1] != '{'` guard is therefore DEFENSIVE, not load-bearing:
		// removing it leaves the whole suite green, because no parsed input
		// reaches it. Kept because the function takes a string and a future
		// caller may not have an Index between it and the source.
		`my $v = $h{k};`,
		`my $v = $a[0];`,
		`$h{k} = 1;`,
	} {
		if got := hedgedKinds(src); len(got) != 0 {
			t.Errorf("%q: nothing is hidden here, want no hedge, got %v",
				src, got)
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

// TestNestedUnknownHedgesToo: an Unknown inside a block must speak its
// refusal, exactly as a top-level one does.
//
// Sites hedged only the Unknowns that were direct children of the root, and
// decidedMarkers declined to descend into a nested one on the stated
// grounds that it was "already hedged at statement level". It was not. A
// `BEGIN { $::{u} = \undef }` parsed to a Phaser holding a Block holding
// two Unknowns, and the subject reported NOTHING for that statement --
// neither a decided site nor a hedge.
//
// That is the silent wrong answer this file's doc comment is about, reached
// from the one direction the original rule did not cover: perl took a
// reference inside those bytes, innermost found no statement offering
// anything at that line, and the verdict was WRONG rather than wider.
//
// comp/fold.t:176 is this shape exactly.
func TestNestedUnknownHedgesToo(t *testing.T) {
	for _, src := range []string{
		`BEGIN { $::{u} = \undef }`,
		"sub f {\n    $::{u} = \\undef;\n}\n",
		"{\n    $::{u} = \\undef;\n}\n",
	} {
		root := parse.Parse([]byte(src))
		if !containsKind(root, parse.Unknown) {
			t.Skipf("%q now parses; this test needs a construct the parser declines", src)
		}
		hedged := hedgedKinds(src)
		for _, kind := range markerKindsForTest {
			if !has(hedged, kind) {
				t.Errorf("%q: a nested Unknown must hedge %s; got %v",
					src, kind, hedged)
			}
		}
	}
}

// markerKindsForTest is the five kinds the oracle scores, mirroring the
// unexported markerKinds so a test in the _test package can name them.
var markerKindsForTest = []string{
	parseoracle.SiteKindReference, parseoracle.SiteKindHash,
	parseoracle.SiteKindMatch, parseoracle.SiteKindReadline,
	parseoracle.SiteKindAnonhash,
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
