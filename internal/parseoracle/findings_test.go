// ABOUTME: A verdict carries WHERE it went wrong, not only a bucket and a sentence about it.
// ABOUTME: The attribution already happens per statement; keeping it is what names the work.

package parseoracle

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVerdictCarriesFindings: the per-statement attribution survives as data.
//
// decideMarker already walks every one of perl's sites, finds the innermost
// statement covering each, and sorts them into wider and wrong. It then
// formats those slices into prose and returns one bucket, so nothing
// downstream can aggregate them -- and a file verdict names no work.
//
// Measured over the 620-file corpus at 461ea788: 72 files WRONG, and the
// line lists inside their Detail strings hold 249 sites concentrated in a
// few constructs. The top ten files hold 145 of them. That ranking is what
// turns "72 files are wrong" into a list of things to fix, and it is
// recoverable today only by regex over a sentence.
func TestVerdictCarriesFindings(t *testing.T) {
	// Two of perl's references, one hedged by the subject and one committed
	// to with nothing there: one wider finding and one wrong finding.
	hedge := SubjectCallSite{Line: 10, EndLine: 10, Unresolved: true}
	commit := SubjectCallSite{Line: 20, EndLine: 20}
	v := CompareFacts(withRefsAt(10, 20), subjectWith(hedge, commit))

	assert.Equal(t, BucketWrong, v.Bucket, v.Detail)
	assert.Len(t, v.Findings, 2, "both sites should be reported: %v", v.Findings)

	byLine := map[int]Finding{}
	for _, f := range v.Findings {
		byLine[f.Line] = f
	}

	assert.Equal(t, BucketWider, byLine[10].Bucket, "line 10 was hedged")
	assert.Equal(t, MarkerSrefgen, byLine[10].Marker)
	assert.Equal(t, BucketWrong, byLine[20].Bucket, "line 20 was committed to with no site")
	assert.Equal(t, MarkerSrefgen, byLine[20].Marker)
}

// TestFindingsSpanEveryMarker: findings accumulate across markers, not only
// the worst one.
//
// compareMarkers keeps the WORST verdict and discards the rest, so a file
// wrong on srefgen and also wrong on rv2hv would report only one of them.
// That is exactly the concentration the corpus shows -- re/pat.t has 35
// wrong sites and they are not all one marker -- so dropping the others
// would understate the work by most of it.
func TestFindingsSpanEveryMarker(t *testing.T) {
	f := compiledFacts()
	f.Srefgen = 1
	f.Sites = map[Marker][]int{
		MarkerSrefgen: {10},
		MarkerRv2hv:   {20},
	}

	// The subject answers both marker questions somewhere in the file, and
	// commits to both statements with no site at either.
	subject := subjectWith(
		SubjectCallSite{Line: 1, EndLine: 1, TookReference: true},
		SubjectCallSite{Line: 2, EndLine: 2, Kind: SiteKindHash},
		SubjectCallSite{Line: 10, EndLine: 10},
		SubjectCallSite{Line: 20, EndLine: 20},
	)
	v := CompareFacts(f, subject)

	assert.Equal(t, BucketWrong, v.Bucket, v.Detail)

	markers := map[Marker]bool{}
	for _, fnd := range v.Findings {
		markers[fnd.Marker] = true
	}
	assert.True(t, markers[MarkerSrefgen], "srefgen finding missing: %v", v.Findings)
	assert.True(t, markers[MarkerRv2hv], "rv2hv finding missing: %v", v.Findings)
}

// TestFindingsAgreeWithDetail: the structured findings say what the prose
// says.
//
// The Detail sentence is what every existing test and every report reads
// today. If the two disagree, one of them is lying, and the findings are the
// half nothing has been checking.
func TestFindingsAgreeWithDetail(t *testing.T) {
	for _, tc := range []struct {
		name    string
		oracle  Facts
		subject SubjectFacts
	}{
		{
			"one wrong",
			withRefsAt(20),
			subjectWith(SubjectCallSite{Line: 1, EndLine: 1, TookReference: true},
				SubjectCallSite{Line: 20, EndLine: 20}),
		},
		{
			"one wider",
			withRefsAt(10),
			subjectWith(SubjectCallSite{Line: 10, EndLine: 10, Unresolved: true}),
		},
		{
			"wider and wrong together",
			withRefsAt(10, 20),
			subjectWith(SubjectCallSite{Line: 10, EndLine: 10, Unresolved: true},
				SubjectCallSite{Line: 20, EndLine: 20}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := CompareFacts(tc.oracle, tc.subject)

			// A finding of the verdict's OWN bucket is named in the prose.
			// The others are not, and that is the gap: a WRONG verdict's
			// sentence lists the lines it committed to and says only a
			// COUNT for the wider ones, because one sentence cannot carry
			// two line lists without becoming unreadable. The findings can.
			for _, f := range v.Findings {
				if f.Bucket != v.Bucket {
					continue
				}
				assert.Contains(t, v.Detail, strconv.Itoa(f.Line),
					"finding at line %d is not in the Detail text", f.Line)
			}
			// The prose's COUNT covers every finding, which is the half it
			// does carry: "perl took a reference at N site(s) the subject
			// did not" is wider plus wrong together.
			if v.Bucket == BucketWrong {
				assert.Contains(t, v.Detail, strconv.Itoa(len(v.Findings))+" site(s)",
					"the count in the prose should be every finding")
			}

			// And the worst finding's bucket is the verdict's bucket.
			worst := BucketExact
			for _, f := range v.Findings {
				if f.Bucket == BucketWrong {
					worst = BucketWrong
				} else if f.Bucket == BucketWider && worst != BucketWrong {
					worst = BucketWider
				}
			}
			assert.Equal(t, v.Bucket, worst,
				"the verdict says %v, the findings say %v", v.Bucket, worst)
		})
	}
}

// TestFindingsDoNotMoveBuckets: adding findings changes no verdict.
//
// The gate reads Verdict.Bucket, and TestGoParserWrongIsZero fails a build
// on it. A refactor that moved the numbers would be a bug, not a refactor --
// compare_facts.go says so at the top and it holds here too.
func TestFindingsDoNotMoveBuckets(t *testing.T) {
	for _, tc := range []struct {
		name string
		want Bucket
		f    Facts
		s    SubjectFacts
	}{
		{"exact", BucketExact, withRefsAt(10),
			subjectWith(SubjectCallSite{Line: 10, EndLine: 10, TookReference: true})},
		{"wider", BucketWider, withRefsAt(10),
			subjectWith(SubjectCallSite{Line: 10, EndLine: 10, Unresolved: true})},
		{"wrong", BucketWrong, withRefsAt(20),
			subjectWith(SubjectCallSite{Line: 1, EndLine: 1, TookReference: true},
				SubjectCallSite{Line: 20, EndLine: 20})},
		{"no-answer", BucketNoAnswer, withRefsAt(10),
			SubjectFacts{OK: true, KnowsCallSites: true, CallSites: []SubjectCallSite{}}},
		{"declined", BucketNoAnswer, withRefsAt(10),
			SubjectFacts{Declined: true, DeclinedReason: "nope"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, CompareFacts(tc.f, tc.s).Bucket)
		})
	}

	// A verdict with no deficit carries no findings: exact means there is
	// nothing to name.
	v := CompareFacts(withRefsAt(10),
		subjectWith(SubjectCallSite{Line: 10, EndLine: 10, TookReference: true}))
	assert.Empty(t, v.Findings, "an exact verdict has no work to name")
}

// TestSweepReportsWhereWrongComesFrom: a report ranks the work, not just the
// failures.
//
// "72 files are WRONG" is 72 units of undifferentiated work. The same run's
// findings say 249 sites across five markers, with the ten worst files
// holding 145 of them -- 58% of the total in 1.6% of the corpus. That is not
// a corpus-wide problem, it is a handful of constructs, and the ranking is
// what says which to fix first.
//
// Measured over the 620-file corpus at 461ea788:
//
//	srefgen 88   match 61   rv2hv 55   anonhash 26   readline 19
//	re/pat.t 35  comp/proto.t 18  op/ref.t 15  op/local.t 14
func TestSweepReportsWhereWrongComesFrom(t *testing.T) {
	report := Report{Files: []Result{
		{Path: "re/pat.t", Verdict: Verdict{Bucket: BucketWrong, Marker: MarkerMatch,
			Findings: []Finding{
				{Line: 10, Marker: MarkerMatch, Bucket: BucketWrong},
				{Line: 20, Marker: MarkerMatch, Bucket: BucketWrong},
				{Line: 30, Marker: MarkerSrefgen, Bucket: BucketWrong},
			}}},
		{Path: "op/ref.t", Verdict: Verdict{Bucket: BucketWrong, Marker: MarkerSrefgen,
			Findings: []Finding{
				{Line: 5, Marker: MarkerSrefgen, Bucket: BucketWrong},
				// A wider finding is real work and must not be counted as
				// wrong: the subject hedged there, which is correct behaviour.
				{Line: 6, Marker: MarkerSrefgen, Bucket: BucketWider},
			}}},
		// A wider FILE contributes no wrong sites at all.
		{Path: "op/ok.t", Verdict: Verdict{Bucket: BucketWider, Marker: MarkerSrefgen,
			Findings: []Finding{{Line: 1, Marker: MarkerSrefgen, Bucket: BucketWider}}}},
	}}

	byMarker, byFile := report.WrongSites()

	assert.Equal(t, 2, byMarker[MarkerMatch], "re/pat.t has two wrong match sites")
	assert.Equal(t, 2, byMarker[MarkerSrefgen], "one in each wrong file; the wider one does not count")
	assert.Equal(t, 3, byFile["re/pat.t"])
	assert.Equal(t, 1, byFile["op/ref.t"], "the hedged site is not wrong")
	assert.NotContains(t, byFile, "op/ok.t", "a wider file contributes no wrong sites")

	// The rendering is ranked, worst first, so a reader sees the work in the
	// order it is worth doing.
	table := report.WhereWrongComesFrom()
	assert.Contains(t, table, "re/pat.t")
	assert.Less(t, strings.Index(table, "re/pat.t"), strings.Index(table, "op/ref.t"),
		"files should be ranked by wrong sites, worst first:\n%s", table)
	assert.NotContains(t, table, "op/ok.t", "a wider file is not work this table names")
}
