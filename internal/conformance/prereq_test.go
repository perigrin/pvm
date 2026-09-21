// ABOUTME: Pins each tier's declared prerequisite against the spec's edge.
// ABOUTME: Ops cannot prove a declaration; measured, all three op rules fail.
package conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// specPrerequisites is the spec's dependency edge for each tier.
//
// Written here rather than read from the corpus, for the reason
// `specTiers` above it is written here: a list read from the thing it
// checks cannot report an absence, and this one has the sharper version
// of that problem. The corpus states each edge in its tier README's
// `## DEPENDS ON` block, which is exactly the value under test. A check
// that read the edge from there would compare the README with itself.
//
// THE OP PAIRING CANNOT DO THIS JOB, and that is measured rather than
// assumed. Each tier's adjacency test asks whether `00_adjacency.t`
// emits an op the declared prerequisite introduces; an adjacency file
// composes the tier's constructs, so it legitimately touches several
// tiers' ops and the question is satisfied by almost any plausible
// declaration. Three tiers proved it independently by mutation -- 12
// repointed to 11_oo, 11 to 07_subroutines, 04 to 02_variables -- and
// each stayed GREEN.
//
// The three stronger op rules were measured against all fourteen tiers
// before this list was written, and all three are dead:
//
//   - "an op no OTHER tier introduces" is vacuous. 103 ops across the
//     fourteen INTRODUCES blocks, none claimed twice, so every op
//     already names exactly one tier and the rule accepts what the
//     current check accepts.
//   - "an op absent from the tier's own construct files" fails 10 of
//     14. Only 02, 04 and 05 have one; 07 and 10 emit no op of their
//     prerequisite at all.
//   - "the declared prerequisite is the LATEST tier reached" fails 4,
//     and two of those contradict a README's stated argument. 09_regex
//     reaches tier 04 through a single `not`, which perl emits for the
//     `!~` in `$s !~ m{zzz}`. 12_packages reaches tier 11 through
//     `method_named`, emitted for `Greet->import("tag")` -- a plain sub
//     call written with an arrow -- while that file's comment argues
//     that pairing with tier 11 "would assert nothing at all".
//
// So this is the corpus's own lesson applied to the declaration: ops
// LINT a declared tier and cannot derive one. See `lintOps`. What the
// ops still do is the pairing check in each tier's own test, which says
// the adjacency file HONOURS the declaration; this says the declaration
// is the one the spec makes. Neither implies the other.
//
// `nothing` is a value, not an absence: tier 01 has no earlier tier, and
// leaving it out would make a deleted entry indistinguishable from a
// tier that legitimately depends on nothing.
var specPrerequisites = map[string]string{
	"01_literals":    "nothing",
	"02_variables":   "01_literals",
	"03_context":     "02_variables",
	"04_operators":   "03_context",
	"05_scoping":     "04_operators",
	"06_control":     "05_scoping",
	"07_subroutines": "06_control",
	"08_references":  "07_subroutines",
	"09_regex":       "01_literals",
	"10_io":          "03_context",
	"11_oo":          "08_references",
	"12_packages":    "07_subroutines",
	"13_opaque":      "10_io",
	"14_recursive":   "09_regex",
}

// checkDeclaredPrerequisites compares what the READMEs declare against
// what the spec says, in both directions.
//
// Both directions, because one alone is satisfiable without the other. A
// tier whose README declares an edge the spec does not have is drift in
// the corpus; a tier the spec has an edge for and the corpus does not
// declare is a missing README block, and reporting only the first would
// pass over the second in silence.
func checkDeclaredPrerequisites(declared, spec map[string]string) error {
	var problems []string
	for tier, want := range spec {
		got, ok := declared[tier]
		if !ok {
			problems = append(problems, fmt.Sprintf(
				"%s declares no prerequisite; the spec says %s", tier, want))
			continue
		}
		if got != want {
			problems = append(problems, fmt.Sprintf(
				"%s declares DEPENDS ON %s; the spec says %s", tier, got, want))
		}
	}
	for tier, got := range declared {
		if _, ok := spec[tier]; !ok {
			problems = append(problems, fmt.Sprintf(
				"%s declares DEPENDS ON %s, and the spec has no such tier", tier, got))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("%s", strings.Join(problems, "; "))
}

// TestSpecPrerequisitesCoverEveryTier pins the list against specTiers.
//
// An edge missing from the map would make its tier unchecked, and
// unchecked is the state the whole issue is about. Driven by specTiers
// so that adding a tier to the spec without an edge fails here.
func TestSpecPrerequisitesCoverEveryTier(t *testing.T) {
	for _, tier := range specTiers {
		if _, ok := specPrerequisites[tier]; !ok {
			t.Errorf("%s has no entry in specPrerequisites, so its declaration is unchecked", tier)
		}
	}
	for tier := range specPrerequisites {
		var found bool
		for _, s := range specTiers {
			if s == tier {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("specPrerequisites names %s, which is not a tier in specTiers", tier)
		}
	}
}

// TestTierPrerequisitesMatchTheSpec is the gate over the real corpus.
//
// This is the assertion the per-tier op pairing could not make. It says
// the fourteen declarations are the ones the spec makes, and a repointed
// README fails here whether or not its adjacency file happens to emit an
// op of the tier it was repointed to.
func TestTierPrerequisitesMatchTheSpec(t *testing.T) {
	declared, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier prerequisites: %v", err)
	}
	if len(declared) == 0 {
		t.Fatal("no tier prerequisites found; the check would have nothing to verify")
	}
	if err := checkDeclaredPrerequisites(declared, specPrerequisites); err != nil {
		t.Errorf("the corpus disagrees with the spec about which tier depends on which: %v\n"+
			"\tA change to a DEPENDS ON block is a change to the corpus's "+
			"ordering claim. If the spec is what moved, this list moves with it, "+
			"deliberately.", err)
	}
}

// TestTierPrerequisiteMutationsAreCaught proves the check by mutation.
//
// The three cases are not invented. Each is a mutation that was applied
// to the real corpus while the op pairing was the only check, and each
// left the suite GREEN -- which is the whole finding. They run here
// against README fixtures rather than against `conformance/`, so the
// proof does not depend on editing the tree it is proving.
//
// The fourth case is the direction the op pairing could never see at
// all: a tier that declares a prerequisite the spec never gave it.
func TestTierPrerequisiteMutationsAreCaught(t *testing.T) {
	for _, tc := range []struct {
		name  string
		tier  string
		wrong string
		want  string
	}{{
		name:  "tier 12 repointed to 11_oo",
		tier:  "12_packages",
		wrong: "11_oo",
		want:  "12_packages declares DEPENDS ON 11_oo; the spec says 07_subroutines",
	}, {
		name:  "tier 11 repointed to 07_subroutines",
		tier:  "11_oo",
		wrong: "07_subroutines",
		want:  "11_oo declares DEPENDS ON 07_subroutines; the spec says 08_references",
	}, {
		name:  "tier 04 repointed to 02_variables",
		tier:  "04_operators",
		wrong: "02_variables",
		want:  "04_operators declares DEPENDS ON 02_variables; the spec says 03_context",
	}, {
		name:  "a tier repointed to nothing",
		tier:  "09_regex",
		wrong: "nothing",
		want:  "09_regex declares DEPENDS ON nothing; the spec says 01_literals",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			declared := writeMutatedCorpus(t, tc.tier, tc.wrong)

			err := checkDeclaredPrerequisites(declared, specPrerequisites)
			if err == nil {
				t.Fatalf("repointing %s at %s left the check GREEN, which is "+
					"the bug this test exists for", tc.tier, tc.wrong)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not say %q", err, tc.want)
			}
		})
	}
}

// writeMutatedCorpus builds a README tree from the spec's edges with one
// tier repointed, and reads it back through readTierDeps.
//
// Through readTierDeps rather than by building the map directly, so the
// mutation travels the path a real edit would: a changed DEPENDS ON
// block, parsed by the same reader the gate uses. A test that mutated
// the map would prove the comparison and not the reading.
func writeMutatedCorpus(t *testing.T, tier, wrong string) map[string]string {
	t.Helper()

	dir := t.TempDir()
	for name, dep := range specPrerequisites {
		if name == tier {
			dep = wrong
		}
		sub := filepath.Join(dir, name)
		if err := os.MkdirAll(sub, 0o750); err != nil {
			t.Fatalf("building the fixture: %v", err)
		}
		readme := fmt.Sprintf("# %s\n\n## INTRODUCES\n\n    nextstate\n\n## DEPENDS ON\n\n    %s\n", name, dep)
		if err := os.WriteFile(filepath.Join(sub, "README.md"), []byte(readme), 0o600); err != nil {
			t.Fatalf("building the fixture: %v", err)
		}
	}

	declared, err := readTierDeps(dir)
	if err != nil {
		t.Fatalf("reading the fixture back: %v", err)
	}
	return declared
}

// TestTierPrerequisiteCheckAcceptsTheUnmutatedFixture is the control.
//
// Without it, a `checkDeclaredPrerequisites` that returned an error for
// every input would pass every mutation case above and prove nothing.
func TestTierPrerequisiteCheckAcceptsTheUnmutatedFixture(t *testing.T) {
	declared := writeMutatedCorpus(t, "", "")
	if err := checkDeclaredPrerequisites(declared, specPrerequisites); err != nil {
		t.Errorf("the check rejected an unmutated corpus: %v", err)
	}
}
