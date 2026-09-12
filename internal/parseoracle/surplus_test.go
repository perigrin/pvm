// ABOUTME: Pins the surplus-reference defect: the oracle and the subject count different populations.
// ABOUTME: A surplus of subject references must not mask a prototype-driven one perl resolved.

package parseoracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSurplusIsNotScored is the defect, stated as the corpus shows it.
//
// The oracle counts only srefgen. The adapter counts every source `\` as a
// TookReference. Measured, those are not the same population:
//
//	\@a  \%h  \&f  \$x  \(&f)   -> srefgen
//	\(@a)  \(@a,@b)  \($x,$y)   -> refgen
//
// so a list-form `\( ... )` contributes to the subject's total and not to
// perl's. The old rule was `oracle.Srefgen <= accounted -> exact`, which read
// that difference as agreement. It is not agreement; it is headroom, and
// headroom big enough to swallow a prototype-driven reference whole.
//
// op/aassign.t is the real instance: srefgen=8, accounted=27, surplus=19.
func TestSurplusIsNotScored(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites: []SubjectCallSite{
			{TookReference: true},
			{TookReference: true},
			{Unresolved: true},
		},
	}
	// Perl took one reference; the subject reports two. The counts are of
	// different populations, so there is nothing to conclude from them.
	oracle := Facts{OK: true, Srefgen: 1}

	v := CompareFacts(oracle, subject)
	assert.Equal(t, BucketNoAnswer, v.Bucket,
		"a surplus means the populations disagree, so the comparison is not meaningful: %s", v.Detail)
}

// TestSurplusIsNotWrongEither guards the other direction. A surplus is an
// unanswerable question, not a subject error: scoring it WRONG would blame the
// subject for a counting mismatch the harness created.
func TestSurplusIsNotWrongEither(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{TookReference: true}, {TookReference: true}},
	}
	v := CompareFacts(Facts{OK: true, Srefgen: 0}, subject)
	assert.NotEqual(t, BucketWrong, v.Bucket, v.Detail)
	assert.Equal(t, BucketNoAnswer, v.Bucket, v.Detail)
}

// TestHedgeStillSeparatesWiderFromWrong keeps the hedge load-bearing where it
// is actually evidence: when perl took a reference the subject did NOT.
//
// The hedge is deliberately not consulted on a matching total, and that is a
// measured decision rather than an oversight. Our grammar marks every
// unqualified `f(...)` call unresolved -- op/aassign.t hedges 181 times -- so
// `hedged > 0` is near-universal and says nothing on its own. A hedge is
// evidence only when there is an unexplained reference for it to explain.
func TestHedgeStillSeparatesWiderFromWrong(t *testing.T) {
	hedging := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{Unresolved: true}},
	}
	committed := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{}},
	}
	oracle := Facts{OK: true, Srefgen: 1}

	assert.Equal(t, BucketWider, CompareFacts(oracle, hedging).Bucket)
	assert.Equal(t, BucketWrong, CompareFacts(oracle, committed).Bucket)
}

// TestCleanAgreementIsStillExact guards the fix from overshooting. A subject
// that resolved every call and matches perl's count has genuinely agreed, and
// must keep scoring exact -- otherwise the bucket becomes unreachable and the
// metric is broken in the other direction.
func TestCleanAgreementIsStillExact(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites: []SubjectCallSite{
			{TookReference: true},
			{}, // a committed call with no reference
		},
	}
	oracle := Facts{OK: true, Srefgen: 1}

	v := CompareFacts(oracle, subject)
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
}

// TestNoReferencesEitherSideIsExact is the commonest corpus shape: a file with
// calls but no references at all. Nothing disagrees, so nothing is at stake.
func TestNoReferencesEitherSideIsExact(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{}, {}},
	}
	oracle := Facts{OK: true, Srefgen: 0}

	v := CompareFacts(oracle, subject)
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
}
