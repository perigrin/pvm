// ABOUTME: Checks that each op is INTRODUCED by exactly one tier, not two.
// ABOUTME: The lint's claimedBy map takes the last writer, so a collision is otherwise invisible.
package conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// opCollisions names every op more than one tier claims to introduce.
//
// `reachable` builds its op-to-tier map by assigning `claimedBy[op] =
// name` as it walks the tiers, so a second claimant silently overwrites
// the first and the collision is invisible BY CONSTRUCTION -- both tiers
// pass, and which one the lint reports as owner depends on Go's map
// iteration order.
//
// Every collision at once, not the first. Writing the fourteen tiers
// produced twelve ops claimed twice or more (issue 01a0c547); a check
// that named one per run would have made that fix twelve runs long.
func opCollisions(tiers tierOps) []string {
	claimants := map[string][]string{}
	for tier, ops := range tiers {
		for _, op := range ops {
			claimants[op] = append(claimants[op], tier)
		}
	}

	var problems []string
	for op, owners := range claimants {
		if len(owners) < 2 {
			continue
		}
		// Sorted so the message is stable: map iteration order is not,
		// and a test whose failure text shuffles between runs is one
		// nobody can diff.
		sort.Strings(owners)
		problems = append(problems, fmt.Sprintf("%s: %s", op, strings.Join(owners, " and ")))
	}
	sort.Strings(problems)
	return problems
}

// TestOpClaimedByExactlyOneTier enforces the rule the twelve were
// resolved by hand under.
//
// The rule is "the tier that emits it first owns it": a later tier that
// REACHES an op explains in its README why it does not INTRODUCE it.
// Nothing enforced that, so a fifteenth tier or an edited README could
// re-claim an op and the suite would stay green -- `TestCorpusLints`
// checks that a claimed op is emitted and that no file outruns its tier,
// neither of which a duplicate claim violates.
//
// A duplicate claim is not cosmetic. It makes the lint's "introduced by
// X" attribution a coin flip, and it widens the earlier tier: an op
// claimed by both 02 and 08 is allowed in every tier from 02 on, so a
// file that genuinely needs 08 passes its lint at 02.
func TestOpClaimedByExactlyOneTier(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	if problems := opCollisions(tiers); len(problems) > 0 {
		t.Errorf("ops claimed by more than one tier:\n\t%s\n"+
			"\tthe tier that emits an op FIRST owns it; a later tier that "+
			"reaches it says so in prose rather than claiming it again",
			strings.Join(problems, "\n\t"))
	}
}

// TestOpCollisionsNamesBothTiers is the mutation test, run in a fixture
// rather than by editing a real README.
//
// It pins the two properties that make the check usable: it names BOTH
// claimants and the op, and it reports every collision in one run. A
// check that fails on the first collision would have turned the twelve
// into twelve runs.
func TestOpCollisionsNamesBothTiers(t *testing.T) {
	dir := t.TempDir()
	for tier, body := range map[string]string{
		"02_variables":  "    padsv gv shift\n",
		"08_references": "    srefgen refgen\n",
		"11_oo":         "    bless shift\n",
	} {
		path := filepath.Join(dir, tier)
		if err := os.MkdirAll(path, 0o750); err != nil {
			t.Fatalf("creating %s: %v", path, err)
		}
		readme := "# " + tier + "\n\n## INTRODUCES\n\n" + body
		if err := os.WriteFile(filepath.Join(path, "README.md"), []byte(readme), 0o600); err != nil {
			t.Fatalf("writing %s/README.md: %v", tier, err)
		}
	}

	tiers, err := readTierOps(dir)
	if err != nil {
		t.Fatalf("reading the fixture READMEs: %v", err)
	}

	got := opCollisions(tiers)
	want := []string{"shift: 02_variables and 11_oo"}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("opCollisions() = %v, want %v", got, want)
	}
}

// TestOpCollisionsReportsAllAtOnce is the twelve-at-a-time case.
//
// Two ops colliding across three tiers, which is the shape the real
// corpus produced: `enterloop`/`leaveloop` were claimed by 05, 06 AND 11.
func TestOpCollisionsReportsAllAtOnce(t *testing.T) {
	tiers := tierOps{
		"05_scoping": {"enterloop", "leaveloop", "padsv"},
		"06_control": {"enterloop", "leaveloop", "and"},
		"11_oo":      {"enterloop", "leaveloop", "bless"},
	}

	got := opCollisions(tiers)
	want := []string{
		"enterloop: 05_scoping and 06_control and 11_oo",
		"leaveloop: 05_scoping and 06_control and 11_oo",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("opCollisions() = %v,\nwant %v", got, want)
	}
}
