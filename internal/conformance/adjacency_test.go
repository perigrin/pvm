// ABOUTME: The adjacency property, checked once for every tier: its adjacency topic emits every op the tier introduces.
// ABOUTME: Also that each tier has such a topic and that every case's main program is multi-statement.
package conformance

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// adjacencyTopic is the topic file holding a tier's constructs beside
// one another. Named by tier because an adjacency case is tier-specific
// by nature; see TestRatchetKeysCarryNoTier.
func adjacencyTopic(tier string) string {
	return "adjacency-" + tier + ".md"
}

// TestEveryTierHasAnAdjacencyTopic walks the spec's tier list, not the
// directory, so a tier that ships without an adjacency topic fails here
// even if nothing else looks for it.
//
// A one-construct-per-case corpus CANNOT reach an adjacency bug: a parser
// that handles each construct alone and mis-handles a pair goes green
// over every case that holds only one.
func TestEveryTierHasAnAdjacencyTopic(t *testing.T) {
	for _, tier := range specTiers {
		t.Run(tier, func(t *testing.T) {
			tierPending(t, tier)
			path := filepath.Join(corpusDir, "mdtest", adjacencyTopic(tier))
			if _, err := os.Stat(path); err != nil {
				t.Errorf("%s has no adjacency topic: %v", tier, err)
			}
		})
	}
}

// TestAdjacencyCoversItsTier checks what every tier README and adjacency
// topic claims: the topic holds every construct its tier introduces.
//
// THE CONSTRUCTS ARE THE TIER'S INTRODUCES OPS, read by readTierOps, so
// there is one list and it lives in the README. The topic's ops are the
// union over its cases: a topic may add a second case for a pair the
// first does not reach, and that case is as much the tier's adjacency as
// the first.
//
// One further claim is checked because it is cheap and ops can prove it:
// EVERY CASE IS MULTI-STATEMENT. Adjacency needs two constructs to be
// adjacent; a one-statement case cannot satisfy it whatever its prose
// says. `nextstate` is one per statement, counted in the MAIN PROGRAM
// only: a statement inside a sub body composes nothing with the main
// program unless something calls the sub. See
// TestAdjacencyCountsMainStatementsOnly.
//
// The tier's declared prerequisite is not checked here: ops cannot tell
// one plausible declaration from another, as prereq_test.go measures, so
// TestTierPrerequisitesMatchTheSpec pins the declaration instead.
func TestAdjacencyCoversItsTier(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus: %v", err)
	}

	for _, tier := range specTiers {
		t.Run(tier, func(t *testing.T) {
			tierPending(t, tier)
			topic := adjacencyTopic(tier)

			var used []string
			n := 0
			for _, c := range cases {
				if c.Topic != topic {
					continue
				}
				n++
				ops, k, err := adjacencyCaseOps(t, c.Source)
				if err != nil {
					t.Fatalf("%s: %v", c.Key, err)
				}
				if k < 2 {
					t.Errorf("%s has %d statement(s) in its main program; an adjacency case needs at least 2", c.Key, k)
				}
				used = append(used, ops...)
			}
			if n == 0 {
				t.Fatalf("%s holds no cases", topic)
			}

			var missing []string
			for _, op := range tiers[tier] {
				if !slices.Contains(used, op) {
					missing = append(missing, op)
				}
			}
			if len(missing) > 0 {
				t.Errorf("%s emits none of %s, which %s/README.md introduces.\n"+
					"\tThe adjacency topic must hold every construct its tier "+
					"introduces, or the pairing it exists to reach is not "+
					"reachable for that construct.",
					topic, strings.Join(missing, " "), tier)
			}
		})
	}
}

// adjacencyCaseOps returns every op a case emits and the number of
// statements in its MAIN PROGRAM, which is where adjacency lives.
func adjacencyCaseOps(t *testing.T, source string) ([]string, int, error) {
	ops, main, err := fileOps(t, source)
	if err != nil {
		return nil, 0, err
	}
	return ops, countOp(main, "nextstate"), nil
}

// TestAdjacencyCountsMainStatementsOnly pins that a statement inside a
// sub body does not count toward an adjacency case's two. A sub nobody
// calls composes nothing with the main program, so a case whose main
// program is one statement is still a one-statement case.
func TestAdjacencyCountsMainStatementsOnly(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{`print "hello"; sub unused { my $a=1; my $b=2 }`, 1},
		{`print "a"; print "b";`, 2},
	} {
		_, statements, err := adjacencyCaseOps(t, tc.source)
		if err != nil {
			t.Fatalf("%s: %v", tc.source, err)
		}
		if statements != tc.want {
			t.Errorf("%s: counted %d statement(s), want %d", tc.source, statements, tc.want)
		}
	}
}

// TestConciseOpsSplitsCRLFOutput holds the main-program split to perl's
// output on Windows, where B::Concise ends each line with "\r\n". A
// header compared with its "\r" still attached never matches, and the
// main program then counts no statements at all.
func TestConciseOpsSplitsCRLFOutput(t *testing.T) {
	out := "main::f:\r\n" +
		"1  <;> nextstate(main 2 case.pl:1) v ->2\r\n" +
		"main program:\r\n" +
		"2  <;> nextstate(main 1 case.pl:1) v:{ ->3\r\n" +
		"3  <;> nextstate(main 1 case.pl:1) v:{ ->4\r\n"
	all, main := splitConciseOps(out)
	if got := countOp(all, "nextstate"); got != 3 {
		t.Errorf("all ops: counted %d nextstate, want 3", got)
	}
	if got := countOp(main, "nextstate"); got != 2 {
		t.Errorf("main program: counted %d nextstate, want 2", got)
	}
}
