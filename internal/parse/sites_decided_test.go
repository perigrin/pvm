// ABOUTME: The subject must DECIDE markers, not only hedge them: exact=0 is a vacuous WRONG=0.
// ABOUTME: Each test asserts the site kind the parser reports for a construct perl reports too.

package parse_test

import (
	"fmt"
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
		// perl: rv2hv. DEREFERENCES only: a named `%h` hedges now, because
		// whether perl emits the op for it depends on how it was declared
		// and the text cannot carry that (01a0bae1-b76f, measured):
		//
		//	my %h;  keys %h    0 rv2hv
		//	our %h; keys %h    1 rv2hv
		//
		// TestDecidableHashStillCommits owns the positive rule.
		{parseoracle.SiteKindHash, []string{
			`my $n = keys %$r;`,
			`my $n = keys %{$r};`,
			`my @k = values %$r;`,
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

// TestSiteSpansReachTheirStatement: a site on a continuation line must span
// back to the line its STATEMENT begins on.
//
// perl attributes a site to the enclosing nextstate, which records the
// statement's FIRST line. Measured on perl 5.42.0:
//
//	my $r = f(          # line 1
//	    "a",            # line 2
//	    { s => 1 },     # line 3
//	);
//	-> nextstate(main 1 ml.pl:1), anonhash
//
// The hash is on line 3 and perl reports line 1. CompareFacts then looks for
// a subject statement whose [start,end] CONTAINS line 1
// (`compare_facts.go:349-359`), so a site reported as the point 3-3 is never
// found and the marker scores WRONG.
//
// Reporting the node's own line alone was the fix for sub bodies -- a
// backslash inside `sub foo {` must not pool under the sub's whole span --
// and it broke this. Both hold if the span runs from the statement's line to
// the node's: it contains perl's line either way.
//
// t/comp/parser_run.t:17 is the surviving case, a four-line call whose hash
// sits on line 19.
func TestSiteSpansReachTheirStatement(t *testing.T) {
	const src = "my $r = f(\n    \"a\",\n    { k => 1 },\n);\n"

	root := parse.Parse([]byte(src))
	var found bool
	for _, s := range parse.Sites(root, []byte(src)) {
		if s.Kind != parseoracle.SiteKindAnonhash || s.Unresolved {
			continue
		}
		found = true
		if s.Line > 1 {
			t.Errorf("anonhash site is %d-%d; it must reach line 1, where "+
				"the statement begins and where perl reports it",
				s.Line, s.EndLine)
		}
		if s.EndLine < 3 {
			t.Errorf("anonhash site is %d-%d; it must reach line 3, where "+
				"the hash actually is", s.Line, s.EndLine)
		}
	}
	if !found {
		t.Fatal("no decided anonhash site at all")
	}
}

// TestSiteSpansStayInsideTheirStatement is the other half, and the reason
// the span cannot simply be the whole statement.
//
// A `\` inside a sub body must NOT pool under `sub foo {`'s span: perl walks
// the CVs and reports it at the body's own line, and a site covering the
// whole sub is never reached from there. The span must START at the
// statement but END at the node.
func TestSiteSpansStayInsideTheirStatement(t *testing.T) {
	const src = "sub foo {\n    my $r = \\$x;\n}\n"

	root := parse.Parse([]byte(src))
	for _, s := range parse.Sites(root, []byte(src)) {
		if s.Kind != parseoracle.SiteKindReference || s.Unresolved {
			continue
		}
		if s.EndLine != 2 {
			t.Errorf("reference site is %d-%d; it must END at line 2, where "+
				"the backslash is, not at the sub's closing brace",
				s.Line, s.EndLine)
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

// TestDerefBlockContentsAreDecided: a `${...}` or `@{...}` reports what is
// actually inside it.
//
// This test used to assert the opposite, and its own comment said why: the
// lexer brace-matched a deref to its closer and emitted ONE Variable token,
// so `${[{a=>214}]}` was an opaque Term and the anonymous hash inside it was
// invisible to a walk over nodes. Silence there is not a refusal -- it is
// the claim that nothing is inside -- so every marker was hedged, and the
// comment ended: "When 01a0ad52 gives the block real children, these become
// decided and the hedge stops firing."
//
// 01a0ad52 landed. They are decided, and this asserts the count rather than
// the hedge. Measured on perl 5.42.0:
//
//	$ perl -MO=Concise,-exec -e 'my $x = ${[{a=>214}]}[0];'
//	  one anonhash
//
// The last two cases hold the other half: a deref containing NO marker now
// reports nothing, and nothing is the right answer. Under the hedge they
// reported all five, which is `wider` -- honest, but five claims of "there
// might be something here" about a statement with nothing in it.
func TestDerefBlockContentsAreDecided(t *testing.T) {
	for _, c := range []struct {
		src   string
		count int
	}{
		// The anonymous hash inside the deref, which perl reports once.
		{`my $x = ${[{a=>214}]}[0];`, 1},
		// A deref of a plain variable holds no marker at all.
		{`my $r = ${$h->{k}};`, 0},
		{`my @a = @{[ 1, 2 ]};`, 0},
	} {
		var n int
		for _, k := range decidedKinds(c.src) {
			if k == parseoracle.SiteKindAnonhash {
				n++
			}
		}
		if n != c.count {
			t.Errorf("%q: %d decided anonhash site(s), want %d",
				c.src, n, c.count)
		}
		if hedged := hedgedKinds(c.src); len(hedged) > 0 {
			t.Errorf("%q: nothing is hidden here any more, yet it hedges %v",
				c.src, hedged)
		}
	}

	// `$::{...}` USED to be opaque and is not any more.
	//
	// It failed for a reason that was never the dereference: it lexed as
	//
	//	DerefSigil "@"  Operator "{"  Variable "$:"  Operator ":"  ...
	//
	// because `$:` is a real punctuation variable (the format line-break set)
	// and `leadingPackageSeparator` required a word byte after `::` -- it
	// found `{`, declined, and the name ended at the first colon. So the
	// statement fell to Unknown and every marker was hedged from there.
	//
	// This comment named that as "its own gap, not a leftover of this one",
	// and issue 01a0de97-77fe closed it: two colons are now enough, so
	// `$::{n}`, `%::` and `@::` all lex as one Variable. The deref's contents
	// are decided and nothing hedges.
	//
	// Measured: the optree is one rv2hv.
	const wasOpaque = `my @g = @{$::{$keys[0]}};`
	if hedged := hedgedKinds(wasOpaque); len(hedged) > 0 {
		t.Errorf("%q: the symbol-table deref is decided now, yet it hedges %v",
			wasOpaque, hedged)
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
		// An ARRAY subscript carries the same bracket characters and must
		// not hedge on that account.
		//
		// This passes whatever `hidesStructure` does, and that is worth
		// saying rather than implying otherwise. `$a[0]` parses to an Index
		// whose Term child is `$a` -- the brackets belong to the Index node,
		// so the leaf text `hidesStructure` sees is never `"$a[0]"`. Its
		// `text[1] != '{'` guard is therefore DEFENSIVE, not load-bearing:
		// removing it leaves the whole suite green, because no parsed input
		// reaches it. Kept because the function takes a string and a future
		// caller may not have an Index between it and the source.
		//
		// A HASH subscript is no longer here, and the reason is a different
		// question from this test's. `$h{k}` hedges now because whether perl
		// emits rv2hv for it depends on a declaration the text cannot carry
		// (01a0bae1-b76f), not because anything is hidden inside it.
		// TestUndecidableHashHedges owns that rule; this one owns opacity.
		`my $v = $a[0];`,
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
// So a site's span must CONTAIN the line perl reports, and which line that
// is depends on where the site sits:
//
//   - Inside a sub body, perl's CV walk reports the body's own line.
//     Measured: `perl -MO=Concise,-exec,foo` on `sub foo {\n my $r = \$x;\n}`
//     gives `nextstate(main 2 sb.pl:2)` and the srefgen after it.
//   - In the continuation of a multi-line statement, perl reports the
//     statement's FIRST line. Measured: `-exec,-main` on
//     `f(\n    \$x,\n);` gives `nextstate(main 1 mc.pl:1)`.
//
// An earlier version of this test asserted the node's line exactly and put
// the second case at line 2. perl says 1, so that expectation was wrong; a
// span from the statement's line to the node's satisfies both, which is what
// `Sites` now produces.
func TestSitesReportTheirOwnLine(t *testing.T) {
	for _, c := range []struct {
		src  string
		kind string
		// mustContain is the line perl attributes the site to.
		mustContain int
		// mustReach is the node's own line, which the span must also cover
		// so a site deeper in a body is not pooled under a whole sub.
		mustReach int
	}{
		// perl's CV walk reports the backslash at line 2; `sub foo {` is 1.
		{"sub foo {\n    my $r = \\$x;\n}\n",
			parseoracle.SiteKindReference, 2, 2},
		{"if (1) {\n    my $r = {a=>1};\n}\n",
			parseoracle.SiteKindAnonhash, 2, 2},
		// A DEREFERENCE, not `keys %h`: a named hash hedges now, because its
		// op depends on how it was declared (01a0bae1-b76f), and this case
		// needs a site the subject decides.
		{"while (1) {\n\n    my @k = keys %$r;\n}\n",
			parseoracle.SiteKindHash, 3, 3},
		// A multi-line statement: perl reports line 1, the node is on 2.
		{"f(\n    \\$x,\n);\n",
			parseoracle.SiteKindReference, 1, 2},
	} {
		var got []string
		var ok bool
		for _, s := range decidedSites(c.src) {
			if s.Kind != c.kind {
				continue
			}
			got = append(got, fmt.Sprintf("%d-%d", s.Line, s.EndLine))
			if s.Line <= c.mustContain && c.mustContain <= s.EndLine &&
				s.Line <= c.mustReach && c.mustReach <= s.EndLine {
				ok = true
			}
		}
		if !ok {
			t.Errorf("%q: want one decided %s whose span contains line %d "+
				"(where perl reports it) and line %d (the node's own); "+
				"got spans %v", c.src, c.kind, c.mustContain, c.mustReach, got)
		}
	}
}
