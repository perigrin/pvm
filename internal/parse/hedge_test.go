// ABOUTME: The rule that makes Unknown mean something: a statement containing one hedges every marker.
// ABOUTME: Without this, Unknown scores WRONG rather than no-answer — measured in compare_facts.go:222.

package parse_test

import (
	"context"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// markerKinds is every kind the oracle scores a statement on. A statement
// containing an Unknown must hedge all of them: the Unknown could be hiding
// any construct, so reporting "no hash here" is a claim the parser cannot
// support.
var markerKinds = []string{
	parseoracle.SiteKindReference,
	parseoracle.SiteKindHash,
	parseoracle.SiteKindMatch,
	parseoracle.SiteKindReadline,
	parseoracle.SiteKindAnonhash,
}

// TestUnknownHedgesEveryMarker: the rule itself.
//
// An Unknown node emits no sites of its own, and CompareFacts scores a perl
// site with no owning statement as WRONG (compare_facts.go:222-229) -- not
// as no-answer. So a parser that emits Unknown and stays silent is not
// declining to answer; it is answering "nothing here", wrongly.
//
// The fix is for the statement to say "something here I could not settle",
// which is an Unresolved site of each kind.
func TestUnknownHedgesEveryMarker(t *testing.T) {
	src := []byte("$x = \\@a;\n")

	sites := parse.Sites(parse.Parse(src), src)
	if len(sites) == 0 {
		t.Fatal("no sites emitted for a statement containing an Unknown")
	}

	seen := map[string]bool{}
	for _, s := range sites {
		if !s.Unresolved {
			t.Errorf("site of kind %q is committed, but its statement is Unknown", s.Kind)
		}
		seen[s.Kind] = true
	}
	for _, k := range markerKinds {
		if !seen[k] {
			t.Errorf("no hedged site of kind %q; an Unknown must hedge every marker", k)
		}
	}
}

// TestUnknownScoresWider: the rule, measured through the harness rather than
// asserted about it.
//
// This is the test that would have caught the claim an earlier draft of this
// issue made -- that Unknown scores no-answer on its own. It runs the real
// CompareFacts against real perl output, so it cannot pass by agreeing with
// my reading of the scoring code.
func TestUnknownScoresWider(t *testing.T) {
	// A backslash reference: perl reports srefgen, so the oracle has a site
	// to score the statement against.
	src := []byte("my @a = (1);\nmy $r = \\@a;\n")

	facts, err := parseoracle.Ask(context.Background(), src, parseoracle.Options{})
	if err != nil {
		t.Skipf("perl unavailable: %v", err)
	}
	if !facts.OK {
		t.Fatalf("perl declined the fixture: %s", facts.Stderr)
	}
	if facts.Srefgen == 0 {
		t.Fatalf("perl reported no srefgen for %q; the fixture is wrong", src)
	}

	subject := parseoracle.SubjectFacts{
		OK:             true,
		KnowsCallSites: true,
		CallSites:      parse.Sites(parse.Parse(src), src),
	}

	v := parseoracle.CompareFacts(facts, subject)
	if v.Bucket == parseoracle.BucketWrong {
		t.Errorf("Unknown scored WRONG: %s", v.Detail)
	}
	if v.Bucket != parseoracle.BucketWider {
		t.Errorf("bucket = %v, want wider; the hedge did not reach the comparison: %s",
			v.Bucket, v.Detail)
	}
}
