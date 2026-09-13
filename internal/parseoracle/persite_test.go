// ABOUTME: Pins per-statement attribution: a reference perl took is explained only by a hedge at that statement.
// ABOUTME: Whole-file totals were quantity-blind, and let a surplus at one statement mask a deficit at another.

package parseoracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func took(line int) SubjectCallSite      { return SubjectCallSite{Line: line, TookReference: true} }
func hedged(line int) SubjectCallSite    { return SubjectCallSite{Line: line, Unresolved: true} }
func committed(line int) SubjectCallSite { return SubjectCallSite{Line: line} }

func subjectWith(sites ...SubjectCallSite) SubjectFacts {
	return SubjectFacts{OK: true, KnowsCallSites: true, CallSites: sites}
}

// TestFoldedConstRefAloneIsExact: `my $b = \1;` is a backslash the subject
// reports and an srefgen perl folded into a constant. Under whole-file totals
// that surplus made the file unanswerable. It masks nothing -- there is no
// other statement for it to mask -- so the honest verdict is agreement.
func TestFoldedConstRefAloneIsExact(t *testing.T) {
	v := CompareFacts(withRefsAt(), subjectWith(took(3)))
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
}

// TestFoldedConstRefDoesNotMaskAPrototype is F2 as the reviewer stated it: a
// folded `\1` at line 3 and a prototype-driven call at line 5 summed to equal
// totals, and the file scored exact. Per statement, line 5 is a reference the
// subject did not take, explained only by its hedge there.
func TestFoldedConstRefDoesNotMaskAPrototype(t *testing.T) {
	v := CompareFacts(withRefsAt(5), subjectWith(took(3), hedged(5)))
	assert.Equal(t, BucketWider, v.Bucket, v.Detail)
}

// TestSurplusCannotMaskAcrossStatements is the property the old surplus rule
// protected by refusing to answer. Two backslashes perl represented some other
// way at lines 3 and 4 say nothing about the committed call at line 5 where
// perl took a reference the subject did not.
func TestSurplusCannotMaskAcrossStatements(t *testing.T) {
	v := CompareFacts(withRefsAt(5), subjectWith(took(3), took(4), committed(5)))
	assert.Equal(t, BucketWrong, v.Bucket, v.Detail)
	assert.Contains(t, v.Detail, "line 5", "the detail must name the statement so it can be triaged")
}

// TestDeficitIsExplainedOnlyByAHedgeAtThatStatement is the reviewer's
// quantity-blind case: srefgen=5, hedged=1, committed=4 scored wider because
// one hedge anywhere excused every deficit. One hedge explains one statement.
func TestDeficitIsExplainedOnlyByAHedgeAtThatStatement(t *testing.T) {
	oracle := withRefsAt(1, 2, 3, 4, 5)

	oneHedge := subjectWith(hedged(1), committed(2), committed(3), committed(4), committed(5))
	v := CompareFacts(oracle, oneHedge)
	assert.Equal(t, BucketWrong, v.Bucket,
		"a hedge at line 1 does not explain the references at lines 2-5: %s", v.Detail)

	allHedged := subjectWith(hedged(1), hedged(2), hedged(3), hedged(4), hedged(5))
	assert.Equal(t, BucketWider, CompareFacts(oracle, allHedged).Bucket)
}

// TestHedgeAtTheStatementIsWider keeps the wider bucket reachable: a hedge on
// the statement where perl took the reference is exactly the honest refusal
// the bucket exists for, and a committed call elsewhere does not change that.
func TestHedgeAtTheStatementIsWider(t *testing.T) {
	v := CompareFacts(withRefsAt(7), subjectWith(hedged(7), committed(8)))
	assert.Equal(t, BucketWider, v.Bucket, v.Detail)
	assert.Equal(t, MarkerSrefgen, v.Marker)
}

// TestCleanAgreementIsStillExact guards against overshooting: a subject that
// took the reference where perl did has agreed, whatever else is on the line.
func TestCleanAgreementIsStillExact(t *testing.T) {
	v := CompareFacts(withRefsAt(1), subjectWith(took(1), committed(1)))
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
	assert.Equal(t, MarkerSrefgen, v.Marker)
}

// TestNoReferencesEitherSideIsExact is the commonest corpus shape.
func TestNoReferencesEitherSideIsExact(t *testing.T) {
	v := CompareFacts(withRefsAt(), subjectWith(committed(1), committed(2)))
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
	assert.Equal(t, MarkerNone, v.Marker)
}

// TestSurplusIsNeverWrong: a backslash perl represented as something other
// than srefgen is a counting difference the harness owns, not a subject error.
func TestSurplusIsNeverWrong(t *testing.T) {
	v := CompareFacts(withRefsAt(), subjectWith(took(1), took(2)))
	assert.NotEqual(t, BucketWrong, v.Bucket, v.Detail)
}

// TestSiteWithoutALineExplainsNothing pins that lines are part of the
// contract. A hedge that names no statement cannot be matched to the
// statement where perl took the reference, so it does not excuse it.
func TestSiteWithoutALineExplainsNothing(t *testing.T) {
	v := CompareFacts(withRefsAt(4), subjectWith(SubjectCallSite{Unresolved: true}))
	assert.Equal(t, BucketWrong, v.Bucket, v.Detail)
}

// TestUnwalkedOracleIsNoAnswer: the walk is the measurement. If perl compiled
// the file but the probe that walks its CVs never ran, the reference facts
// are absent rather than zero, and a zero here is the defect this milestone
// removed, wearing a different hat.
func TestUnwalkedOracleIsNoAnswer(t *testing.T) {
	oracle := compiledFacts()
	oracle.Walked = false
	v := CompareFacts(oracle, subjectWith(committed(1)))
	assert.Equal(t, BucketNoAnswer, v.Bucket, v.Detail)
	assert.Contains(t, v.Detail, "walk")
}
