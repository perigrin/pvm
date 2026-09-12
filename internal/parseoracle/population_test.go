// ABOUTME: Guards that the population rule is load-bearing in both directions and that no optree means no verdict.
// ABOUTME: Each test here is one that a mutation of the guard it names must kill; a guard nobody watched fail is not a guard.

package parseoracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSurplusGuardIsLoadBearing is the B6 test: it must die when the equality
// guard is loosened back to `oracle.Srefgen <= accounted`, which is round 1's
// exact defect.
//
// The round-1 fix could not be tested as it was written. `accounted > Srefgen`
// returned no-answer BEFORE the equality check, so the only inputs that ever
// reached `Srefgen <= accounted` were ones where `accounted <= Srefgen` held
// too -- that is, equality. The `<` half was unreachable, so `<=` and `==`
// were indistinguishable to every possible input and the whole suite stayed
// green with the defect restored.
//
// The fix is to state the rule once rather than twice: the populations either
// match or they do not, and this test pins the surplus direction at the single
// place that now decides it.
func TestSurplusGuardIsLoadBearing(t *testing.T) {
	// A surplus: the subject accounts for more references than perl took.
	// Under a `<=` rule this reads as agreement, which is the defect.
	surplus := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{TookReference: true}, {TookReference: true}},
	}
	v := CompareFacts(withSrefgen(1), surplus)
	assert.Equal(t, BucketNoAnswer, v.Bucket,
		"a subject surplus is an unanswerable question, not agreement: %s", v.Detail)
	assert.NotEqual(t, BucketExact, v.Bucket,
		"scoring a surplus exact is round 1's defect: %s", v.Detail)

	// And the guard must be specific rather than a blanket refusal: an equal
	// count still agrees, or the exact bucket becomes unreachable and the
	// metric is broken in the other direction.
	equal := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{TookReference: true}},
	}
	assert.Equal(t, BucketExact, CompareFacts(withSrefgen(1), equal).Bucket,
		"an equal count must still be exact")
}

// TestNoOptreeIsNoAnswer is B5: perl reporting success having produced no
// optree is not a parse to compare against.
//
// A corpus file that calls skip_all inside BEGIN exits during compilation.
// Perl reports ok:1 with op_count:0 and an empty op list -- it never finished
// parsing the file. Our parser meanwhile produces a whole tree, finds no
// references in it, and the two zero totals match trivially, so the file
// scores exact. The harness compared against a parse that never happened.
//
// Ten corpus files report this shape; six of them were scored exact.
func TestNoOptreeIsNoAnswer(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{}, {}},
	}
	// ok, but nothing was compiled: the file exited during compilation.
	oracle := Facts{OK: true, OpCount: 0, Ops: nil, Srefgen: 0}

	v := CompareFacts(oracle, subject)
	assert.Equal(t, BucketNoAnswer, v.Bucket,
		"perl produced no optree, so there is no parse to compare against: %s", v.Detail)
}

// TestOptreePresentStillScores keeps the B5 fix from overshooting into a
// blanket refusal. A file perl really did compile must still reach a verdict,
// or the fix has made every bucket but no-answer unreachable.
func TestOptreePresentStillScores(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{TookReference: true}},
	}
	oracle := Facts{OK: true, OpCount: 12, Ops: []string{"leave", "enter"}, Srefgen: 1}

	assert.Equal(t, BucketExact, CompareFacts(oracle, subject).Bucket)
}

// TestNoOptreeNeedsBothSignals pins the conjunction in noOptree.
//
// perl reports the op list and its length as two fields, and "no optree" means
// both are empty. Deciding on OpCount alone would let a decoding bug -- a
// dropped count beside a populated op list -- read as a file that never
// compiled, and silently move a file out of the denominator. The richer signal
// is the one to trust when the two disagree.
//
// Without this test the `len(oracle.Ops) == 0` half is unreachable by every
// other case in the suite: dropping it leaves the whole suite green, which is
// the same "guard nobody watched fail" that TestSurplusGuardIsLoadBearing
// above exists for.
func TestNoOptreeNeedsBothSignals(t *testing.T) {
	subject := SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      []SubjectCallSite{{}},
	}
	// An op list is present; only the count says zero. There IS an optree.
	oracle := Facts{OK: true, OpCount: 0, Ops: []string{"enter", "leave"}}

	v := CompareFacts(oracle, subject)
	assert.NotEqual(t, BucketNoAnswer, v.Bucket,
		"an op list is an optree even when the count disagrees: %s", v.Detail)
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)

	// And the genuinely empty case still declines, or the conjunction has been
	// deleted rather than made precise.
	assert.Equal(t, BucketNoAnswer,
		CompareFacts(Facts{OK: true, OpCount: 0}, subject).Bucket)
}

// compiledFacts is the minimum perl reports for a file it really compiled: an
// optree with something in it.
//
// Tests that synthesise perl's side need this because `Facts{OK: true}` alone
// describes an interpreter that succeeded without building an optree, which is
// how a file that exits during compilation looks -- a real corpus shape, and
// one that now correctly declines. A test meaning "perl compiled this" must
// say so, or it is asserting against a parse that did not happen.
func compiledFacts() Facts {
	return Facts{OK: true, OpCount: 4, Ops: []string{"enter", "nextstate", "padsv", "leave"}}
}

// withSrefgen is compiledFacts carrying n reference-taking ops.
func withSrefgen(n int) Facts {
	f := compiledFacts()
	f.Srefgen = n
	return f
}
