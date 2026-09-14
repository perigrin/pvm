// ABOUTME: Pins the five markers end to end: perl reports each per statement, the subject answers each, the verdict decides each.
// ABOUTME: Every op shape asserted here was measured on perl 5.42.0 with B::Concise before it was written down.

package parseoracle

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

// --- the oracle side: perl reports each marker by statement ----------------

// TestOracleReportsEachMarkerByStatement: parse_facts.pl reports every
// marker op with the line of the statement it belongs to, inside a sub body
// as well as in the main program, through the same CV walk that finds
// srefgen.
//
// Measured with `perl -MO=Concise,-exec` on 5.42.0:
//
//	keys %$r        rv2hv        (a lexical %h is padhv, $h{a} is multideref)
//	$_ =~ /b/       match        ($a / $b is divide; split /,/ has no match op)
//	<STDIN>         readline     (<*.c> is glob)
//	{ a => 1 }      anonhash
//	{}              emptyavhv with OPpEMPTYAVHV_IS_HV set
//	$p / $q, $p % $q  divide, modulo: not markers, must not be reported
func TestOracleReportsEachMarkerByStatement(t *testing.T) {
	src := "sub f {\n" + // 1
		"    my $r = shift;\n" + // 2
		"    my @k = keys %$r;\n" + // 3
		"    my $m = $_ =~ /b/;\n" + // 4
		"    my $l = <STDIN>;\n" + // 5
		"    return { a => 1 };\n" + // 6
		"}\n" + // 7
		"my $e = {};\n" + // 8
		"my ($p, $q) = @ARGV;\n" + // 9
		"my $d = $p / $q;\n" + // 10
		"my $mo = $p % $q;\n" // 11

	facts, err := Ask(context.Background(), []byte(src), Options{})
	require.NoError(t, err)
	require.True(t, facts.OK, "stderr: %s", facts.Stderr)
	require.True(t, facts.Walked)

	assert.Empty(t, facts.Sites[MarkerSrefgen], "no reference is taken anywhere")
	assert.Equal(t, []int{3}, facts.Sites[MarkerRv2hv])
	assert.Equal(t, []int{4}, facts.Sites[MarkerMatch])
	assert.Equal(t, []int{5}, facts.Sites[MarkerReadline])
	assert.Equal(t, []int{6, 8}, facts.Sites[MarkerAnonhash],
		"{} is emptyavhv rather than anonhash, and is still an anonymous hash")
}

// TestOracleReadlineSurvivesTheOptimiser: `$x .= <FH>` is rewritten to a
// single rcatline op (measured), so a readline marker keyed on the op name
// alone would lose it.
func TestOracleReadlineSurvivesTheOptimiser(t *testing.T) {
	src := "my $x;\n" +
		"$x .= <STDIN>;\n" // 2
	facts, err := Ask(context.Background(), []byte(src), Options{})
	require.NoError(t, err)
	require.True(t, facts.OK, "stderr: %s", facts.Stderr)
	assert.Equal(t, []int{2}, facts.Sites[MarkerReadline])
}

// TestOracleAnonhashIsNotAnEmptyArray: emptyavhv serves both `[]` and `{}`;
// only the one flagged as a hash is an anonhash site.
func TestOracleAnonhashIsNotAnEmptyArray(t *testing.T) {
	src := "my $a = [];\n" +
		"my $h = {};\n" // 2
	facts, err := Ask(context.Background(), []byte(src), Options{})
	require.NoError(t, err)
	require.True(t, facts.OK, "stderr: %s", facts.Stderr)
	assert.Equal(t, []int{2}, facts.Sites[MarkerAnonhash])
}

// --- the verdict side: one decision per marker per statement -------------

func siteOf(kind string, line int) SubjectCallSite {
	return SubjectCallSite{Kind: kind, Line: line}
}

func hedgedSiteOf(kind string, line int) SubjectCallSite {
	return SubjectCallSite{Kind: kind, Line: line, Unresolved: true}
}

func withSites(sites map[Marker][]int) Facts {
	f := compiledFacts()
	f.Sites = sites
	f.Srefgen = len(sites[MarkerSrefgen])
	return f
}

// markerKinds pairs each new marker with the site kind a subject reports
// for it. srefgen keeps its own shape (took_reference on a call or a
// reference site) and is covered by the existing per-site tests.
var markerKinds = []struct {
	marker Marker
	kind   string
}{
	{MarkerRv2hv, SiteKindHash},
	{MarkerMatch, SiteKindMatch},
	{MarkerReadline, SiteKindReadline},
	{MarkerAnonhash, SiteKindAnonhash},
}

// TestEachMarkerIsDecidedPerStatement applies the srefgen rules, unchanged,
// to each of the other four markers: a site of the marker's kind in the same
// statement is agreement; a hedge there is wider; a subject that described
// the file but not this site is WRONG; a subject that never answered this
// marker's question at all has a gap, not a wrong answer.
func TestEachMarkerIsDecidedPerStatement(t *testing.T) {
	for _, mk := range markerKinds {
		t.Run(string(mk.marker), func(t *testing.T) {
			perl := withSites(map[Marker][]int{mk.marker: {3}})

			v := CompareFacts(perl, subjectWith(siteOf(mk.kind, 3)))
			assert.Equal(t, BucketExact, v.Bucket, v.Detail)
			assert.Equal(t, mk.marker, v.Marker)

			v = CompareFacts(perl, subjectWith(hedgedSiteOf(mk.kind, 3)))
			assert.Equal(t, BucketWider, v.Bucket, v.Detail)
			assert.Equal(t, mk.marker, v.Marker)

			// Answered the question elsewhere in the file, so silence at
			// line 3 is a claim about line 3.
			v = CompareFacts(perl, subjectWith(siteOf(mk.kind, 5)))
			assert.Equal(t, BucketWrong, v.Bucket, v.Detail)
			assert.Equal(t, mk.marker, v.Marker)
			assert.Contains(t, v.Detail, "line 3")

			// Reported calls, but never a site of this kind: the marker's
			// question was not answered, which is a gap and not a lie.
			v = CompareFacts(perl, subjectWith(committed(3)))
			assert.Equal(t, BucketNoAnswer, v.Bucket, v.Detail)
			assert.Equal(t, mk.marker, v.Marker)

			// Reported nothing whatsoever: the same gap.
			v = CompareFacts(perl, subjectWith())
			assert.Equal(t, BucketNoAnswer, v.Bucket, v.Detail)
		})
	}
}

// TestAHedgeExplainsOnlyItsOwnMarker: an unresolved call at a statement says
// the subject could not settle a prototype there. It says nothing about
// whether `/x/` in the same statement is a match, so it cannot turn a
// missing match site into wider.
func TestAHedgeExplainsOnlyItsOwnMarker(t *testing.T) {
	perl := withSites(map[Marker][]int{MarkerMatch: {3}})
	v := CompareFacts(perl, subjectWith(hedged(3), siteOf(SiteKindMatch, 9)))
	assert.Equal(t, BucketWrong, v.Bucket, v.Detail)
	assert.Equal(t, MarkerMatch, v.Marker)
}

// TestTheWorstMarkerDecidesTheFile: a file is WRONG if any marker is, wider
// if any is, and the verdict names the marker that made it so.
func TestTheWorstMarkerDecidesTheFile(t *testing.T) {
	perl := withSites(map[Marker][]int{
		MarkerSrefgen:  {2},
		MarkerRv2hv:    {3},
		MarkerMatch:    {4},
		MarkerAnonhash: {5},
	})
	subject := subjectWith(
		took(2),                       // srefgen: exact
		hedgedSiteOf(SiteKindHash, 3), // rv2hv: wider
		siteOf(SiteKindMatch, 6),      // match: answered, but not at 4
		siteOf(SiteKindAnonhash, 5),   // anonhash: exact
	)
	v := CompareFacts(perl, subject)
	assert.Equal(t, BucketWrong, v.Bucket, v.Detail)
	assert.Equal(t, MarkerMatch, v.Marker)

	subject.CallSites[2] = siteOf(SiteKindMatch, 4)
	v = CompareFacts(perl, subject)
	assert.Equal(t, BucketWider, v.Bucket, v.Detail)
	assert.Equal(t, MarkerRv2hv, v.Marker)

	subject.CallSites[1] = siteOf(SiteKindHash, 3)
	v = CompareFacts(perl, subject)
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
}

// TestExactNamesAMarkerOnlyWhenOneWasInPlay: the verified surface is the
// exact verdicts that exercised a marker. An exact reached with nothing in
// play must be distinguishable from one that was actually checked, or the
// exact rate overstates what was verified (findings s0.11: 332 of 411
// exact files had no reference in play).
func TestExactNamesAMarkerOnlyWhenOneWasInPlay(t *testing.T) {
	v := CompareFacts(withSites(nil), subjectWith(siteOf(SiteKindMatch, 4)))
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
	assert.Equal(t, MarkerNone, v.Marker)

	v = CompareFacts(withSites(map[Marker][]int{MarkerMatch: {4}}),
		subjectWith(siteOf(SiteKindMatch, 4)))
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
	assert.Equal(t, MarkerMatch, v.Marker)
}

// --- the adapter side: our tree-sitter parser answering each question ----

// markerSample is one source line per construct, with the site kinds the
// adapter must report for it. Every node kind the predicates read was
// found by parsing these lines with `psc parse --format sexpr`, not
// guessed; every "nothing" is a construct perl compiles to no marker op
// (measured), which a subject must therefore not claim either.
var markerSample = []struct {
	src   string
	kinds []string
}{
	{"my %h = (a => 1);", []string{SiteKindHash}},
	{"my $r = \\%h;", []string{SiteKindHash, SiteKindReference}},
	{"my @k = keys %$r;", []string{SiteKindHash}},
	{"my @k2 = keys %{$r};", []string{SiteKindHash}},
	{"my $d = $r->%*;", []string{SiteKindHash}},
	{"my $e = $h{a};", []string{SiteKindHash}},
	{"my $e2 = $$r{a};", []string{SiteKindHash}},
	{"my $e3 = $r->{a};", []string{SiteKindHash}},
	{"my @s = @$r{qw(a b)};", []string{SiteKindHash}},
	{"my @s2 = $r->@{qw(a b)};", []string{SiteKindHash}},
	{"my %kv = %$r{qw(a)};", []string{SiteKindHash, SiteKindHash}}, // %kv and the slice
	{"my $str = \"$h{a}\";", []string{SiteKindHash}},
	{"my @as = @a[1,2];", nil},
	{"my %kva = %a[0,1];", []string{SiteKindHash}}, // %kva only: the [ slice is an array
	{"my $mo = $p % $q;", nil},
	{"my $m = $_ =~ /b/;", []string{SiteKindMatch}},
	{"my $m2 = /b/;", []string{SiteKindMatch}},
	{"my $m3 = $_ !~ m/b/;", []string{SiteKindMatch}},
	{"my $m4 = $_ =~ $re;", []string{SiteKindMatch}},
	{"my @g = grep { /x/ } @l;", []string{SiteKindMatch}},
	{"my $s = $_ =~ s/a/b/;", nil},
	{"my $t = $_ =~ tr/a/b/;", nil},
	{"my $qr = qr/x/;", nil},
	{"my $dv = $p / $q;", nil},
	{"my $l = <STDIN>;", []string{SiteKindReadline}},
	{"my $l2 = <$fh>;", []string{SiteKindReadline}},
	{"my $l3 = <>;", []string{SiteKindReadline}},
	{"my $l4 = readline($fh);", []string{SiteKindReadline}},
	{"my $l5 = readline FH;", []string{SiteKindReadline}},
	{"my $g = <*.c>;", nil},
	{"my $g2 = glob(\"*.c\");", nil},
	{"my $a = { a => 1 };", []string{SiteKindAnonhash}},
	{"my $a2 = {};", []string{SiteKindAnonhash}},
	{"my $a3 = +{ a => 1 };", []string{SiteKindAnonhash}},
	{"my @m = map { +{ a => 1 } } @l;", []string{SiteKindAnonhash}},
	{"my @m2 = map { $_ => 1 } @l;", nil},
	{"my $ar = [];", nil},
}

// TestTreeSitterSubjectReportsEachMarker: the adapter answers all four new
// questions from node kinds, and stays silent where perl builds no marker.
func TestTreeSitterSubjectReportsEachMarker(t *testing.T) {
	var src strings.Builder
	for _, s := range markerSample {
		src.WriteString(s.src)
		src.WriteByte('\n')
	}
	tree, err := parser.New().Parse([]byte(src.String()))
	require.NoError(t, err)
	subject := TreeSitterSubject(tree)
	require.True(t, subject.OK, subject.DeclinedReason)

	byLine := map[int][]string{}
	for _, c := range subject.CallSites {
		if c.Kind == "" {
			continue // calls are the srefgen question, measured elsewhere
		}
		byLine[c.Line] = append(byLine[c.Line], c.Kind)
	}
	for i, s := range markerSample {
		got := byLine[i+1]
		sort.Strings(got)
		want := append([]string(nil), s.kinds...)
		sort.Strings(want)
		assert.Equal(t, want, got, "line %d: %s", i+1, s.src)
	}
}

// TestOurParserAgreesWithPerlOnEveryMarker is the round trip: perl reports
// the sites, our parser reports the sites, and the verdict is exact with a
// marker in play -- not the vacuous exact of a file with nothing to check.
func TestOurParserAgreesWithPerlOnEveryMarker(t *testing.T) {
	src := "my %h = (a => 1);\n" +
		"my $r = \\%h;\n" +
		"sub f {\n" +
		"    my @k = keys %$r;\n" +
		"    my $m = $_ =~ /b/;\n" +
		"    my $l = <STDIN>;\n" +
		"    return { a => 1 };\n" +
		"}\n" +
		"my $e = {};\n" +
		"my @s = @$r{qw(a)};\n"
	facts, err := Ask(context.Background(), []byte(src), Options{})
	require.NoError(t, err)
	require.True(t, facts.OK, facts.Stderr)
	for _, m := range Markers {
		assert.NotEmpty(t, facts.Sites[m], "perl must report %s for this source", m)
	}

	tree, err := parser.New().Parse([]byte(src))
	require.NoError(t, err)
	v := Compare(facts, tree, []byte(src))
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
	assert.NotEqual(t, MarkerNone, v.Marker, "a marker was in play, and the verdict must say so")
}

// TestPerlSubjectSpeaksEveryKind: the reference subject -- perl itself
// through the contract -- reports the four new kinds, so the contract has
// a working non-Go producer for each.
func TestPerlSubjectSpeaksEveryKind(t *testing.T) {
	dir := t.TempDir()
	src := "my %h = (a => 1);\n" +
		"my @k = keys %ENV;\n" + // 2: rv2hv (a global hash; a lexical is padhv)
		"my $m = $_ =~ /b/;\n" + // 3
		"my $l = <STDIN>;\n" + // 4
		"my $a = { a => 1 };\n" + // 5
		"my $e = {};\n" // 6
	require.NoError(t, os.WriteFile(filepath.Join(dir, "kinds.pl"), []byte(src), 0o644))

	script, err := filepath.Abs(filepath.Join("testdata", "perl_subject.pl"))
	require.NoError(t, err)
	subject := Subject{Command: []string{"perl", script}, Dir: dir}
	facts, err := subject.Parse(context.Background(), "kinds.pl")
	require.NoError(t, err)
	require.True(t, facts.OK)

	byLine := map[int][]string{}
	for _, c := range facts.CallSites {
		byLine[c.Line] = append(byLine[c.Line], c.Kind)
	}
	assert.Equal(t, []string{SiteKindHash}, byLine[2])
	assert.Equal(t, []string{SiteKindMatch}, byLine[3])
	assert.Equal(t, []string{SiteKindReadline}, byLine[4])
	assert.Equal(t, []string{SiteKindAnonhash}, byLine[5])
	assert.Equal(t, []string{SiteKindAnonhash}, byLine[6], "{} is emptyavhv, and still an anonymous hash")
}

// --- the report side: the verified surface, not just the exact rate ------

func exactWith(path string, sites map[Marker][]int) Result {
	return Result{Path: path, Facts: withSites(sites),
		Verdict: Verdict{Bucket: BucketExact}}
}

// TestReportSeparatesVerifiedFromVacuousExact: an exact verdict on a file
// with no marker in play verified nothing. The report has to say how many
// exact verdicts actually exercised a marker, because that number -- not
// the exact rate -- is what a new parser would be measured against.
func TestReportSeparatesVerifiedFromVacuousExact(t *testing.T) {
	report := Report{Files: []Result{
		exactWith("a.t", map[Marker][]int{MarkerRv2hv: {3}}),
		exactWith("b.t", nil),
		{Path: "c.t", Facts: withSites(map[Marker][]int{MarkerMatch: {1}, MarkerRv2hv: {2}}),
			Verdict: Verdict{Bucket: BucketWider, Marker: MarkerMatch}},
		{Path: "d.t", Facts: withSites(map[Marker][]int{MarkerMatch: {1}}), Excluded: true},
	}}
	report.Totals[BucketExact] = 2
	report.Totals[BucketWider] = 1
	report.Environmental = 1

	assert.Equal(t, 1, report.VerifiedExact())
	assert.Equal(t, 1, report.VacuousExact())
	assert.Equal(t, map[Marker]int{MarkerRv2hv: 2, MarkerMatch: 1}, report.InPlay(),
		"an excluded file is not measured, so its markers are not in play")

	rendered := report.String()
	assert.Contains(t, rendered, "exact with a marker in play")
	assert.Contains(t, rendered, "exact with nothing in play")
	assert.Contains(t, rendered, "rv2hv 2")
	assert.Contains(t, rendered, "match 1")
}

// TestBaselineMetricCountsEveryMarker: the baseline's metric column is the
// number of marker sites perl reported, across all five markers, so a row
// with metric 0 is an exact that verified nothing.
func TestBaselineMetricCountsEveryMarker(t *testing.T) {
	report := Report{Files: []Result{
		exactWith("a.t", map[Marker][]int{MarkerSrefgen: {1}, MarkerMatch: {2, 3}, MarkerAnonhash: {4}}),
	}}
	b := NewBaseline(Pin{Interpreter: "5.042000", Revision: "x"}, report, "")
	require.Len(t, b.Rows, 1)
	assert.Equal(t, 4, b.Rows[0].Metric)
}
