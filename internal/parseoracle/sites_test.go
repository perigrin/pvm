// ABOUTME: Pins how the oracle counts verdicts per marker site rather than per file.
// ABOUTME: Every site of a graded file lands in exactly one bucket; a whole-file no-answer puts them all in no-answer.

package parseoracle

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withSites is compiledFacts carrying the given marker sites, one entry per
// op the way parse_facts.pl pushes them: two ops on one line are two entries.
func withSites(sites map[Marker][]int) Facts {
	f := compiledFacts()
	f.Sites = sites
	return f
}

// TestNoAnswerFindingsArePerSite: perl matched three times, twice on line 5,
// and the subject reported no match site anywhere, so it never answered the
// match question. Each of perl's three ops is a site the subject left
// unanswered, and the count of findings is the count of sites -- one per
// line would lose the second match on line 5.
func TestNoAnswerFindingsArePerSite(t *testing.T) {
	v := CompareFacts(withSites(map[Marker][]int{MarkerMatch: {5, 5, 9}}), subjectWith(committed(1)))
	require.Equal(t, BucketNoAnswer, v.Bucket, v.Detail)
	assert.Equal(t, []Finding{
		{Line: 5, Marker: MarkerMatch, Bucket: BucketNoAnswer},
		{Line: 5, Marker: MarkerMatch, Bucket: BucketNoAnswer},
		{Line: 9, Marker: MarkerMatch, Bucket: BucketNoAnswer},
	}, v.Findings)
}

// buckets is a Totals literal for one marker's sites: exact, wider, WRONG,
// no-answer, in Bucket order.
func buckets(exact, wider, wrong, noAnswer int) Totals {
	return Totals{BucketExact: exact, BucketWider: wider, BucketWrong: wrong, BucketNoAnswer: noAnswer}
}

// graded is a measured Result for facts and the verdict CompareFacts reaches
// against subject, which is what the runner records for a graded file.
func graded(facts Facts, subject SubjectFacts) Result {
	return Result{Path: "fixture.t", Facts: facts, Verdict: CompareFacts(facts, subject)}
}

// siteBuckets is Result.SiteBuckets that fails the test on an error.
func siteBuckets(t *testing.T, r Result) map[Marker]Totals {
	t.Helper()
	counts, err := r.SiteBuckets()
	require.NoError(t, err)
	return counts
}

// TestWholeFileNoAnswerCountsEverySite: a verdict naming no marker is a
// whole-file no-answer, reached before any marker was compared, so it
// carries no findings to subtract. Every site perl reported for the file
// is a site nobody graded, across every marker.
func TestWholeFileNoAnswerCountsEverySite(t *testing.T) {
	r := Result{Path: "fixture.t",
		Facts:   withSites(map[Marker][]int{MarkerSrefgen: {3}, MarkerMatch: {4, 4}}),
		Verdict: Verdict{Bucket: BucketNoAnswer, Marker: MarkerNone, Detail: "declined"}}
	assert.Equal(t, map[Marker]Totals{
		MarkerSrefgen: buckets(0, 0, 0, 1),
		MarkerMatch:   buckets(0, 0, 0, 2),
	}, siteBuckets(t, r))
}

// TestSiteBucketsSumToFactSites grades one file whose markers land in every
// bucket, and checks each marker's buckets add back up to perl's sites:
//
//	srefgen 1,2,3   took at 1 (exact), hedged at 2 (wider), committed at 3 (WRONG)
//	rv2hv   5,6     a decided hash at 5 (exact), nothing at 6 (WRONG)
//	match   7,7     no match site anywhere (no-answer, both)
func TestSiteBucketsSumToFactSites(t *testing.T) {
	facts := withSites(map[Marker][]int{
		MarkerSrefgen: {1, 2, 3},
		MarkerRv2hv:   {5, 6},
		MarkerMatch:   {7, 7},
	})
	subject := subjectWith(took(1), hedged(2), committed(3),
		SubjectCallSite{Line: 5, Kind: SiteKindHash}, committed(6))
	counts := siteBuckets(t, graded(facts, subject))

	assert.Equal(t, map[Marker]Totals{
		MarkerSrefgen: buckets(1, 1, 1, 0),
		MarkerRv2hv:   buckets(1, 0, 1, 0),
		MarkerMatch:   buckets(0, 0, 0, 2),
	}, counts)
	for m, lines := range facts.Sites {
		assert.Equal(t, len(lines), counts[m].Sum(), "%s: buckets must sum to perl's sites", m)
	}
}

// TestSubjectDeclineCountsEverySiteNoAnswer: perl built an optree and walked
// it, so there is ground truth; the subject is the side that declined. A
// rule keyed only on perl having produced nothing would grade these sites,
// as exact, against a subject that said nothing about them.
func TestSubjectDeclineCountsEverySiteNoAnswer(t *testing.T) {
	facts := withSites(map[Marker][]int{MarkerSrefgen: {2}, MarkerRv2hv: {3, 8}})
	want := map[Marker]Totals{MarkerSrefgen: buckets(0, 0, 0, 1), MarkerRv2hv: buckets(0, 0, 0, 2)}

	declined := SubjectFacts{Declined: true, DeclinedReason: "an error node at line 2"}
	assert.Equal(t, want, siteBuckets(t, graded(facts, declined)), "subject.Declined")

	rejected := SubjectFacts{OK: false, KnowsCallSites: true}
	assert.Equal(t, want, siteBuckets(t, graded(facts, rejected)), "!subject.OK")
}

// TestPerlDeclinedFileWithSitesCountsNoAnswer is the op/magic.t shape: the
// oracle reports perl declined while still carrying the sites its walk
// found. The file stays in the denominator with every site no-answer;
// dropping it is how 46 sites went missing from a per-site measurement.
func TestPerlDeclinedFileWithSitesCountsNoAnswer(t *testing.T) {
	facts := withSites(map[Marker][]int{MarkerSrefgen: {10, 11}, MarkerRv2hv: {12}})
	facts.OK = false
	r := graded(facts, subjectWith(took(10), took(11), SubjectCallSite{Line: 12, Kind: SiteKindHash}))
	require.Equal(t, BucketNoAnswer, r.Verdict.Bucket, r.Verdict.Detail)

	report := Report{Files: []Result{r}}
	counts, err := report.SiteBuckets()
	require.NoError(t, err)
	assert.Equal(t, map[Marker]Totals{
		MarkerSrefgen: buckets(0, 0, 0, 2),
		MarkerRv2hv:   buckets(0, 0, 0, 1),
	}, counts)
}

// TestWholeFileNoAnswerNeverCountsExact: a whole-file no-answer has no
// findings, and "sites minus findings" would score every one of its sites
// exact. Nothing was compared, so nothing agreed.
func TestWholeFileNoAnswerNeverCountsExact(t *testing.T) {
	facts := withSites(map[Marker][]int{MarkerAnonhash: {1, 2, 3}})
	facts.OpCount, facts.Ops = 0, nil // perl reported success with no optree
	r := graded(facts, subjectWith(committed(1)))
	require.Empty(t, r.Verdict.Findings, "premise: a whole-file no-answer carries no findings")

	counts := siteBuckets(t, r)
	assert.Zero(t, counts[MarkerAnonhash][BucketExact])
	assert.Equal(t, buckets(0, 0, 0, 3), counts[MarkerAnonhash])
}

// TestWholeFileNoAnswerWithNoSitesAddsNothing: a whole-file no-answer is
// counted by its sites, not as one unit, so a file with no sites adds no
// phantom site anywhere.
func TestWholeFileNoAnswerWithNoSitesAddsNothing(t *testing.T) {
	r := graded(withSites(nil), SubjectFacts{Declined: true})
	require.Equal(t, BucketNoAnswer, r.Verdict.Bucket)

	counts, err := Report{Files: []Result{r}}.SiteBuckets()
	require.NoError(t, err)
	for m, totals := range counts {
		assert.Zero(t, totals.Sum(), "%s gained a phantom site", m)
	}
}

// TestMarkerNoAnswerDoesNotSpreadToOtherMarkers: the file's verdict is
// no-answer because the match question went unanswered, but it names that
// marker, so it is not a whole-file no-answer. The srefgen site the subject
// accounted for is still exact.
func TestMarkerNoAnswerDoesNotSpreadToOtherMarkers(t *testing.T) {
	facts := withSites(map[Marker][]int{MarkerSrefgen: {1, 4}, MarkerMatch: {2}})
	r := graded(facts, subjectWith(took(1), hedged(4)))
	require.Equal(t, BucketNoAnswer, r.Verdict.Bucket)
	require.Equal(t, MarkerMatch, r.Verdict.Marker)

	assert.Equal(t, map[Marker]Totals{
		MarkerSrefgen: buckets(1, 1, 0, 0),
		MarkerMatch:   buckets(0, 0, 0, 1),
	}, siteBuckets(t, r))
}

// TestNoAnswerWithoutSitesEmitsNoFindings: the subject never answers the
// match question, but perl built no match op, so there is no site to leave
// unanswered and no finding to subtract.
func TestNoAnswerWithoutSitesEmitsNoFindings(t *testing.T) {
	assert.Empty(t, decideMarker(MarkerMatch, nil, nil, false).Findings)

	r := graded(withSites(map[Marker][]int{MarkerSrefgen: {1}}), subjectWith(took(1)))
	assert.Equal(t, BucketExact, r.Verdict.Bucket, r.Verdict.Detail)
	assert.Equal(t, map[Marker]Totals{MarkerSrefgen: buckets(1, 0, 0, 0)}, siteBuckets(t, r))
}

// TestFindingsExceedingSitesIsAnError: a finding is a site perl reported, so
// more findings than sites means the verdict and the facts describe
// different files. The helper says so instead of counting exact below zero.
func TestFindingsExceedingSitesIsAnError(t *testing.T) {
	r := Result{Path: "fixture.t",
		Facts: withSites(map[Marker][]int{MarkerSrefgen: {1}}),
		Verdict: Verdict{Bucket: BucketWrong, Marker: MarkerSrefgen, Findings: []Finding{
			{Line: 1, Marker: MarkerSrefgen, Bucket: BucketWrong},
			{Line: 1, Marker: MarkerSrefgen, Bucket: BucketWrong},
		}}}
	_, err := r.SiteBuckets()
	assert.Error(t, err)

	// A finding for a marker perl reported no site for at all.
	r.Verdict.Findings = []Finding{{Line: 1, Marker: MarkerRv2hv, Bucket: BucketWrong}}
	_, err = r.SiteBuckets()
	assert.Error(t, err)

	_, err = Report{Files: []Result{r}}.SiteBuckets()
	assert.ErrorContains(t, err, "fixture.t", "the error names the file")
}

// TestSiteCountedInExactlyOneBucket: two srefgen ops on line 4 against one
// hedge there are two wider findings and two wider sites; two match ops on
// line 4 with no match site anywhere are two no-answer sites and never
// wider, because an unanswered marker is no-answer as a whole.
func TestSiteCountedInExactlyOneBucket(t *testing.T) {
	facts := withSites(map[Marker][]int{MarkerSrefgen: {4, 4}, MarkerMatch: {4, 4}})
	r := graded(facts, subjectWith(hedged(4)))
	assert.Len(t, r.Verdict.Findings, 4)

	assert.Equal(t, map[Marker]Totals{
		MarkerSrefgen: buckets(0, 2, 0, 0),
		MarkerMatch:   buckets(0, 0, 0, 2),
	}, siteBuckets(t, r))
}

// TestExcludedAndErrorFilesContributeNoSites: an environmental exclusion
// and an oracle-side runner error never reached a verdict. Their zero-value
// Verdict reads as exact, so counting their sites would pad the exact
// count with files nobody graded.
func TestExcludedAndErrorFilesContributeNoSites(t *testing.T) {
	facts := withSites(map[Marker][]int{MarkerSrefgen: {1, 2}, MarkerRv2hv: {3}})
	report := Report{Files: []Result{
		{Path: "excluded.t", Facts: facts, Excluded: true, Failure: Environmental},
		{Path: "error.t", Facts: facts, Err: "oracle on error.t: signal: killed"},
	}}
	counts, err := report.SiteBuckets()
	require.NoError(t, err)
	for m, totals := range counts {
		assert.Zero(t, totals.Sum(), "%s counted sites from a file that was never graded", m)
	}
}

// TestSubjectErrorIsWholeFileNoAnswer splits a run's failures by side. Perl
// answered for ref.t and the subject did not -- it crashed on one run and
// timed out on another -- so the file has ground truth and no answer: a
// whole-file no-answer, every perl site counted. When the ORACLE fails
// there is no ground truth, and the file stays a runner error, out of every
// bucket.
func TestSubjectErrorIsWholeFileNoAnswer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the perl oracle is POSIX-only: parse_facts.pl runs perl through sh -c")
	}
	dir := t.TempDir()
	write := func(name, src string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644))
	}
	write("ref.t", "sub f (\\@) {}\nmy @a;\nf(@a);\n")
	// A compile-time fork that outlives its parent wedges the oracle, which
	// is the shape test.pl's watchdog() creates; see TestRunBoundsAWedgedFile.
	write("wedged.t", "BEGIN { my $pid = fork(); if (!$pid) { sleep 60; exit } }\nmy $x = 1;\n")

	crashed := writeSubjectScript(t, "#!/bin/sh\necho 'psc: segmentation fault' >&2\nexit 139\n")
	hung := writeSubjectScript(t, "#!/bin/sh\nexec sleep 30\n")
	for name, subject := range map[string]*Subject{
		"crashed":   {Command: []string{crashed}},
		"timed out": {Command: []string{hung}, Timeout: 200 * time.Millisecond},
	} {
		t.Run(name, func(t *testing.T) {
			report, err := Run(context.Background(), []string{"ref.t"}, RunOptions{Dir: dir, Subject: subject})
			require.NoError(t, err)
			r := report.Files[0]
			require.Empty(t, r.Err, "a subject failure is not a runner error")
			require.NotEmpty(t, r.SubjectErr, "the subject's failure is kept for triage")
			assert.Equal(t, BucketNoAnswer, r.Verdict.Bucket, r.Verdict.Detail)
			assert.Equal(t, MarkerNone, r.Verdict.Marker)
			assert.Equal(t, 0, report.Errors)
			assert.Equal(t, 1, report.Totals[BucketNoAnswer])
			assert.Contains(t, report.String(), "subject error files:\n  ref.t: "+r.SubjectErr,
				"a crash graded no-answer must still be named in the report")

			counts, err := report.SiteBuckets()
			require.NoError(t, err)
			assert.Equal(t, map[Marker]Totals{MarkerSrefgen: buckets(0, 0, 0, 1)}, counts)
		})
	}

	t.Run("oracle failed", func(t *testing.T) {
		report, err := Run(context.Background(), []string{"wedged.t"},
			RunOptions{Dir: dir, Subject: &Subject{Command: []string{crashed}}, Timeout: 2 * time.Second})
		require.NoError(t, err)
		r := report.Files[0]
		require.NotEmpty(t, r.Err, "the oracle timed out: no ground truth")
		assert.Empty(t, r.SubjectErr)
		assert.Equal(t, 1, report.Errors)
		assert.Zero(t, report.Measured())

		counts, err := report.SiteBuckets()
		require.NoError(t, err)
		assert.Empty(t, counts)
	})
}

// TestRatchetDistinguishesSubjectErrorFromRefusal freezes a subject crash as
// its own status. Both are graded no-answer, but a baseline row that said
// only "no-answer" could not tell a psc crash from an honest refusal, and a
// crash that turned into a refusal would move nothing the ratchet checks.
func TestRatchetDistinguishesSubjectErrorFromRefusal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the perl oracle is POSIX-only: parse_facts.pl runs perl through sh -c")
	}
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ref.t"), []byte("sub f (\\@) {}\nmy @a;\nf(@a);\n"), 0o644))

	run := func(subject *Subject) Report {
		t.Helper()
		report, err := Run(context.Background(), []string{"ref.t"}, RunOptions{Dir: dir, Subject: subject})
		require.NoError(t, err)
		return report
	}
	crashed := run(&Subject{Command: []string{writeSubjectScript(t, "#!/bin/sh\nexit 139\n")}})
	refused := run(&Subject{Command: []string{writeFakeSubject(t, `{"ok": false}`)}})
	require.Equal(t, BucketNoAnswer, crashed.Files[0].Verdict.Bucket)
	require.Equal(t, BucketNoAnswer, refused.Files[0].Verdict.Bucket, refused.Files[0].Verdict.Detail)

	pin := Pin{Interpreter: "perl", Revision: "r"}
	crashRow := NewBaseline(pin, crashed, dir).Rows[0]
	refusalRow := NewBaseline(pin, refused, dir).Rows[0]
	assert.Equal(t, "subject-error", crashRow.Status)
	assert.Equal(t, "no-answer", refusalRow.Status)

	// The status survives a round trip through the committed form.
	parsed, err := ParseBaseline([]byte(NewBaseline(pin, crashed, dir).Render()))
	require.NoError(t, err)
	assert.Equal(t, "subject-error", parsed.Rows[0].Status)

	// A crash that becomes a refusal, or the reverse, is a move a human
	// must look at -- neither is an improvement.
	var diff *RatchetError
	require.ErrorAs(t, NewBaseline(pin, crashed, dir).Check(refused), &diff)
	assert.Len(t, diff.Regressions, 1)
	require.ErrorAs(t, NewBaseline(pin, refused, dir).Check(crashed), &diff)
	assert.Len(t, diff.Regressions, 1)
}

// TestSubjectBuildFailureIsRunnerError keeps the harness's own failures out
// of the subject's grade. A default subject that would not build, or a
// Subject with no command, never asked any parser anything, so the file is a
// runner error -- not a no-answer charged to the parser.
func TestSubjectBuildFailureIsRunnerError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the perl oracle is POSIX-only: parse_facts.pl runs perl through sh -c")
	}
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ref.t"), []byte("my $x = 1;\n"), 0o644))

	check := func(t *testing.T, subject *Subject, want string) {
		t.Helper()
		report, err := Run(context.Background(), []string{"ref.t"}, RunOptions{Dir: dir, Subject: subject})
		require.NoError(t, err)
		r := report.Files[0]
		assert.Contains(t, r.Err, want)
		assert.Empty(t, r.SubjectErr, "the harness failed, not the subject")
		assert.Equal(t, 1, report.Errors)
		assert.Zero(t, report.Measured())
	}

	t.Run("no command", func(t *testing.T) {
		check(t, &Subject{}, "no command configured")
	})

	t.Run("default subject does not build", func(t *testing.T) {
		// Settle the once first, then stand a build failure in for its
		// result; the real result is restored for the tests that follow.
		// This patches a package global, so the test must never be made
		// parallel.
		_, _ = defaultSubject()
		saved := internalParse.err
		internalParse.err = errors.New("building the internal/parse subject: exit status 1")
		t.Cleanup(func() { internalParse.err = saved })
		check(t, nil, "building the internal/parse subject")
	})
}
