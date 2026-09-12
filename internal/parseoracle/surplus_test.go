// ABOUTME: Pins the surplus-reference defect: the oracle and the subject count different populations.
// ABOUTME: A surplus of subject references must not mask a prototype-driven one perl resolved.

package parseoracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSurplusDoesNotMaskAProtoReference is the reviewer's reproduction.
//
// The oracle counts only srefgen. The adapter counts every source `\` as a
// TookReference. Those are not the same population: a list-form `\(@a)` emits
// refgen, not srefgen, so it contributes to the subject's side and not to
// perl's. That difference is headroom, and headroom absorbs exactly the thing
// this harness exists to detect.
//
// Here two subject references cover an oracle Srefgen of 2 -- but only one of
// perl's two is the explicit one. The other was resolved by a prototype at the
// call the subject marked unresolved. Scoring this exact is the mask.
func TestSurplusDoesNotMaskAProtoReference(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites: []SubjectCallSite{
			{TookReference: true},
			{TookReference: true},
			{Unresolved: true},
		},
	}
	oracle := Facts{OK: true, Srefgen: 2}

	v := CompareFacts(oracle, subject)
	assert.NotEqual(t, BucketExact, v.Bucket,
		"a surplus of subject references must not score exact while a call is unresolved: %s", v.Detail)
}

// TestHedgeIsConsultedWhenTotalsMatch is the short-circuit, stated on its own.
//
// `oracle.Srefgen <= accounted` returns before the hedge is ever looked at, so
// a file carrying hundreds of unresolved calls scores exact on a matching
// total. A file that admits it could not resolve 300 calls has not agreed with
// perl about them; it has declined to answer.
func TestHedgeIsConsultedWhenTotalsMatch(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites: []SubjectCallSite{
			{TookReference: true},
			{Unresolved: true},
		},
	}
	oracle := Facts{OK: true, Srefgen: 1}

	v := CompareFacts(oracle, subject)
	assert.NotEqual(t, BucketExact, v.Bucket,
		"an unresolved call must be consulted even when the totals match: %s", v.Detail)
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
