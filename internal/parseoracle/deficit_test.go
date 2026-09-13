// ABOUTME: Tests the deficit arm of the reference rule: perl took a reference the subject did not account for.
// ABOUTME: WRONG needs a commitment to be wrong about; a subject with nothing to say is scored no-answer.

package parseoracle

import (
	"context"
	"testing"
)

// TestDeficitWithNothingCommittedIsNoAnswer pins F3: WRONG fires only when the
// subject committed to something. `s/x/\@a/e` gives perl an srefgen and a
// static parser no call site at all; the old rule scored that WRONG with a
// detail claiming the subject "committed to its calls", when it reported none.
func TestDeficitWithNothingCommittedIsNoAnswer(t *testing.T) {
	subject := SubjectFacts{OK: true, KnowsCallSites: true, CallSites: []SubjectCallSite{}}

	v := CompareFacts(withRefsAt(1), subject)
	if v.Bucket != BucketNoAnswer {
		t.Errorf("Bucket: got %v (%s), want %v -- a subject that committed to nothing cannot be wrong",
			v.Bucket, v.Detail, BucketNoAnswer)
	}
}

// TestDeficitWithACommittedCallIsStillWrong guards the other direction: the
// WRONG bucket must stay reachable, or the metric cannot fail. A subject that
// committed to a plain call where perl took a reference is wrong about it.
func TestDeficitWithACommittedCallIsStillWrong(t *testing.T) {
	subject := SubjectFacts{OK: true, KnowsCallSites: true, CallSites: []SubjectCallSite{{Line: 1, Name: "f"}}}

	v := CompareFacts(withRefsAt(1), subject)
	if v.Bucket != BucketWrong {
		t.Errorf("Bucket: got %v (%s), want %v", v.Bucket, v.Detail, BucketWrong)
	}
}

// TestReferenceSiteIsNotACall pins F4: an honest subject can account for
// `my $r = \@a;` -- which has no call -- without inventing one. The contract
// names a site kind for it, so the adapter's pseudo call site becomes a
// documented answer rather than a trick only our own adapter knows.
func TestReferenceSiteIsNotACall(t *testing.T) {
	subject := writeFakeSubject(t,
		`{"ok": true, "call_sites": [{"line": 1, "kind": "reference", "took_reference": true}]}`)

	facts, err := Subject{Command: []string{subject}}.Parse(context.Background(), "irrelevant.pl")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := facts.CallSites[0].Kind; got != SiteKindReference {
		t.Errorf("Kind: got %q, want %q", got, SiteKindReference)
	}

	v := CompareFacts(withRefsAt(1), facts)
	if v.Bucket != BucketExact {
		t.Errorf("Bucket: got %v (%s), want %v -- the reference is accounted for", v.Bucket, v.Detail, BucketExact)
	}
}
