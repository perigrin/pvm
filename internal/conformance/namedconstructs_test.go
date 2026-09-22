// ABOUTME: The 35 parsing-distinctive constructs T1 uses must each appear in some corpus file's SOURCE.
// ABOUTME: Measured against T1 rather than chosen, so the list is a coverage debt, not a wish.
package conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// namedConstructs is the list issue 01a0c730 measured: every perl
// construct T1 uses whose PARSE is distinctive and which no corpus file
// named at the time of measurement.
//
// The triage that produced it discarded 47 ordinary named operators
// (`unlink`, `chdir`, `mkdir`, `fileno` ...) because a file for those
// asserts nothing a hundred other list-operator calls do not. These 35
// are the ones where a parser can be wrong in a way nothing else catches:
// an argument that is a PATTERN rather than an expression, an operator
// that is also an lvalue, a keyword gated behind a pragma, a form whose
// meaning changes with its arity.
//
// The list is FROZEN. It is the measurement taken at a3a0ed0c, not a
// running inventory -- adding to it later would make the issue's claim
// unfalsifiable, since any gap could be defined away. A construct that
// turns out to be uninteresting gets removed with a recorded reason, in
// the same way the 47 were.
var namedConstructs = []string{
	"caller", "can", "chr", "defined", "die", "do", "each", "elsif",
	"eval", "exit", "grep", "index", "isa", "lock", "map", "ord",
	"pack", "pos", "prototype", "push", "qq", "say", "select",
	"split", "sprintf", "substr", "tie", "tied", "time", "tr",
	"undef", "unpack", "unshift", "values", "warn",
}

// TestEveryNamedConstructAppears is the gate on issue 01a0c730.
//
// It reads the SOURCE section only. A construct named in a comment is not
// a claim about parsing -- the corpus documents heavily, and a header that
// explains why `sprintf` folds would otherwise satisfy a check for
// `sprintf` without any file ever running it.
//
// The word-boundary match is deliberately loose about CONTEXT: `time`
// matches `time` in `my $t = time` and would also match a hypothetical
// `$time`. Tightening it to reject sigil-prefixed uses costs more than it
// buys, because the failure it would prevent -- a construct "covered" only
// by a variable that happens to share its name -- is caught by the tier
// lint, which requires the tier's claimed ops to be genuinely emitted.
func TestEveryNamedConstructAppears(t *testing.T) {
	sources, err := allSources(corpusDir)
	if err != nil {
		t.Fatalf("reading corpus sources: %v", err)
	}

	var missing []string
	for _, c := range namedConstructs {
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(c) + `\b`)
		found := false
		for _, src := range sources {
			if re.MatchString(src) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, c)
		}
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d of %d constructs appear in no corpus file's source:\n\t%s",
			len(missing), len(namedConstructs), strings.Join(missing, " "))
	}
}

// allSources returns the `--- source` body of every corpus file.
//
// Reading through ParseFile rather than slicing the text means the same
// section boundaries the RUNNER uses decide what counts as source, so a
// construct hiding in a header cannot satisfy the check.
//
// One tier directory deep, matching every other check in the package,
// rather than a recursive walk. A walk also reaches the package's own
// `.t` FIXTURES -- deliberately malformed files that exist to prove the
// parser rejects them -- and a corpus check that trips over a fixture
// designed to be broken is reporting on the wrong tree.
func allSources(corpus string) ([]string, error) {
	tiers, err := os.ReadDir(corpus)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, tier := range tiers {
		if !tier.IsDir() {
			continue
		}
		paths, err := filepath.Glob(filepath.Join(corpus, tier.Name(), "*.t"))
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			raw, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			f, err := ParseFile(string(raw))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			out = append(out, f.Source)
		}
	}
	return out, nil
}
