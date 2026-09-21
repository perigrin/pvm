// ABOUTME: The dependency lint: a file's ops must be introduced at or before its tier.
// ABOUTME: Ops come from perl's optree, so the check needs no parser of our own.
package conformance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// tierOps maps a tier directory name to the ops that tier introduces.
//
// Derived from the tier READMEs rather than maintained here: a second
// hand-written list would have to agree with the first, and two lists that
// must agree are how drift starts. See readTierOps.
type tierOps map[string][]string

// reIntroduces matches a tier README's machine-read block:
//
//	## INTRODUCES
//
//	    const nextstate padsv_store
//
// Indented so it renders as a code block, which keeps the one part a
// machine reads from looking like prose a human may reword freely.
var reIntroduces = regexp.MustCompile(`(?ms)^## INTRODUCES\s*\n(.*?)(?:\n## |\z)`)

// readTierOps builds the op-to-tier map from the tier READMEs.
//
// The alternative is a Go table beside this file, and then the tier's
// contents and its declared contents are two lists that must agree. This
// milestone has already watched a hand-copied op list lose an entry on
// its first day, so the set lives in the directory it describes.
func readTierOps(corpus string) (tierOps, error) {
	entries, err := os.ReadDir(corpus)
	if err != nil {
		return nil, fmt.Errorf("reading the corpus root: %w", err)
	}

	out := tierOps{}
	for _, e := range entries {
		if !e.IsDir() || !isTierDir(e.Name()) {
			continue
		}
		path := filepath.Join(corpus, e.Name(), "README.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("tier %s has no README: %w", e.Name(), err)
		}

		m := reIntroduces.FindSubmatch(raw)
		if m == nil {
			// Loudly, rather than as an empty set: an empty set would
			// make every file in the tier report unclaimed ops, which
			// reads as a corpus-wide failure rather than one bad README.
			return nil, fmt.Errorf("%s: no `## INTRODUCES` block", path)
		}
		out[e.Name()] = strings.Fields(string(m[1]))
	}
	return out, nil
}

// isTierDir reports whether a directory name is a numbered tier.
func isTierDir(name string) bool {
	return len(name) > 3 && name[2] == '_' &&
		name[0] >= '0' && name[0] <= '9' &&
		name[1] >= '0' && name[1] <= '9'
}

// lintOps reports whether a file uses an op no tier at or before its own
// introduces.
//
// This is check 1, and the ONLY thing ops can prove. They cannot derive a
// tier: `my $x = 1+2` arrives as `const[IV 3] s/FOLD` with no add op, so
// the optimiser can erase the very construct a tier is about. A file
// therefore DECLARES its tier and the ops lint that declaration --
// necessary, not sufficient, and parser-independent, which is what lets
// the corpus check itself while the parser it measures is incomplete.
func lintOps(t *testing.T, source, tier string, tiers tierOps) error {
	t.Helper()

	ops, err := opsOf(t, source)
	if err != nil {
		return err
	}

	allowed, claimedBy := reachable(tier, tiers)

	var misplaced, unclaimed []string
	seen := map[string]bool{}
	for _, op := range ops {
		if allowed[op] || seen[op] {
			continue
		}
		seen[op] = true
		if owner, ok := claimedBy[op]; ok {
			misplaced = append(misplaced, fmt.Sprintf("%s (introduced by %s)", op, owner))
		} else {
			unclaimed = append(unclaimed, op)
		}
	}

	// Reported separately because they call for different fixes. A
	// misplaced file moves to a later tier; an unclaimed op means the
	// tier READMEs do not yet describe the language, and defaulting
	// either way hides that -- allowed lets the union grow by accident,
	// denied makes every new construct look like a violation.
	var problems []string
	if len(misplaced) > 0 {
		sort.Strings(misplaced)
		problems = append(problems, fmt.Sprintf(
			"uses ops introduced after %s: %s", tier, strings.Join(misplaced, ", ")))
	}
	if len(unclaimed) > 0 {
		sort.Strings(unclaimed)
		problems = append(problems, fmt.Sprintf(
			"uses ops no tier claims: %s", strings.Join(unclaimed, ", ")))
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(problems, "; "))
}

// reachable returns the ops available at a tier -- its own and every
// earlier one's -- and which tier introduces each op anywhere.
//
// "Earlier" is by directory name, which sorts correctly because the names
// are numerically prefixed. That is the same property the filesystem
// ordering relies on, so a renumbering moves both together.
func reachable(tier string, tiers tierOps) (allowed map[string]bool, claimedBy map[string]string) {
	allowed = map[string]bool{}
	claimedBy = map[string]string{}

	for name, ops := range tiers {
		for _, op := range ops {
			claimedBy[op] = name
			if name <= tier {
				allowed[op] = true
			}
		}
	}
	return allowed, claimedBy
}

// reConciseOp matches one op in `B::Concise,-exec` output:
//
//	5  <0> pushmark s
//
// The op name is the third field. Taken from the oracle's own extraction
// in testdata/parse_facts.pl, which reads the same format.
var reConciseOp = regexp.MustCompile(`^\s*\S+\s+<[^>]*>\s+(\w+)`)

// opsOf returns the ops perl compiles a source into, in execution order.
//
// AFTER the peephole optimiser, which is the point: the lint must see
// what perl actually built, not what the source appears to say.
func opsOf(t *testing.T, source string) ([]string, error) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "case.pl")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		return nil, fmt.Errorf("writing the case: %w", err)
	}

	out, err := exec.Command("perl", "-MO=Concise,-exec", path).Output()
	if err != nil {
		return nil, fmt.Errorf("perl -MO=Concise refused the source: %w", err)
	}

	var ops []string
	for _, line := range strings.Split(string(out), "\n") {
		if m := reConciseOp.FindStringSubmatch(line); m != nil {
			ops = append(ops, m[1])
		}
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("no ops in B::Concise output, which cannot be right")
	}
	return ops, nil
}
