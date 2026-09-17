// ABOUTME: The rule that makes Unknown mean something: a statement containing one hedges every marker.
// ABOUTME: Without this, Unknown scores WRONG rather than no-answer — measured in compare_facts.go:222.

package parse_test

import (
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
	// A statement form no issue has implemented yet, so it is still Unknown.
	// See unimplementedStatement.
	src := []byte(unimplementedStatement)

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

// TestUnknownScoresWider was DELETED, following the instruction its own
// fixture comment carried.
//
// It measured the hedging rule end to end: parse a statement the milestone
// had not implemented, hand the sites to the real CompareFacts, and assert
// the verdict is `wider` rather than WRONG. That needed a statement form
// this parser DECLINES which also contains a backslash reference perl
// reports as srefgen.
//
// The fixture moved four times as the parser learned forms -- `$x = \\@a`,
// `my $r = \\@a`, an `if` block, a `do` block -- and after this commit no
// such fixture remains:
//
//	goto &other      perl emits no srefgen when it follows a declaration
//	try { ... }      the reference is optimised away
//	continue { ... } the reference is optimised away
//	format           the body is opaque, so a reference cannot be inside it
//
// unimplementedStatement's comment says what to do here: "When nothing is
// left to put here, these tests are measuring an empty set. DELETE them then
// rather than inventing a construct to keep them alive."
//
// TestUnknownHedgesEveryMarker still asserts the rule directly -- every
// Unknown emits an Unresolved site of all five kinds -- and it is
// mutation-checked. What is lost is the end-to-end leg through
// CompareFacts, which the M1 gate's subject tests will cover once the
// parser answers the harness as a subprocess.
