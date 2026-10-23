// ABOUTME: The adjacency property, checked once for every tier: its adjacency topic emits every op the tier introduces.
// ABOUTME: Also that each tier has such a topic, that it is multi-statement, and that it pairs with its prerequisite.
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
// Two further claims are checked because they are cheap and ops can
// prove them:
//
//   - EVERY CASE IS MULTI-STATEMENT. Adjacency needs two constructs to be
//     adjacent; a one-statement case cannot satisfy it whatever its prose
//     says. `nextstate` is one per statement.
//   - THE TOPIC USES AN OP OF ITS DECLARED PREREQUISITE. Weak where the
//     prerequisite is 01_literals -- measured, any program emits most of
//     tier 01's ops, because `print` and a statement are tier 01's -- and
//     real elsewhere. A source-spelling pairing check, which tier 01
//     would need, took a bespoke helper per tier and is not restored.
func TestAdjacencyCoversItsTier(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	deps, err := readTierDeps(corpusDir)
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
				ops, err := opsOf(t, c.Source)
				if err != nil {
					t.Fatalf("%s: %v", c.Key, err)
				}
				if k := countOp(ops, "nextstate"); k < 2 {
					t.Errorf("%s has %d statement(s); an adjacency case needs at least 2", c.Key, k)
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

			if dep := deps[tier]; dep != "nothing" &&
				!slices.ContainsFunc(tiers[dep], func(op string) bool { return slices.Contains(used, op) }) {
				t.Errorf("%s emits no op of its prerequisite %s, so it pairs "+
					"with nothing it claims to need", topic, dep)
			}
		})
	}
}
