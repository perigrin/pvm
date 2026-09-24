// ABOUTME: The 35 parsing-distinctive constructs T1 uses must each appear in some corpus case's SOURCE.
// ABOUTME: Measured against T1 rather than chosen, so the list is a coverage debt, not a wish.
package conformance

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

// namedConstructs is the list issue 01a0c730 measured: every perl
// construct T1 uses whose PARSE is distinctive and which no corpus file
// covered at the time of measurement.
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
// It reads the ```perl block only. A construct named in a comment is not
// a claim about parsing -- the corpus documents heavily, and a header that
// explains why `sprintf` folds would otherwise satisfy a check for
// `sprintf` without any case ever running it.
//
// The word-boundary match is deliberately loose about CONTEXT: `time`
// matches `time` in `my $t = time` and would also match a hypothetical
// `$time`. Tightening it to reject sigil-prefixed uses costs more than it
// buys, because the failure it would prevent -- a construct "covered" only
// by a variable that happens to share its name -- is caught by the tier
// lint, which requires the tier's claimed ops to be genuinely emitted.
//
// AllCases reads the topics under conformance/mdtest/ and nothing else,
// which keeps this check off the package's own `.t` FIXTURES --
// deliberately malformed files that exist to prove the parser rejects
// them. A corpus check that tripped over a fixture designed to be broken
// would be reporting on the wrong tree.
func TestEveryNamedConstructAppears(t *testing.T) {
	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatalf("reading corpus: %v", err)
	}

	var missing []string
	for _, c := range namedConstructs {
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(c) + `\b`)
		found := false
		for _, cc := range cases {
			if re.MatchString(cc.Source) {
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
		t.Errorf("%d of %d constructs appear in no corpus case's source:\n\t%s",
			len(missing), len(namedConstructs), strings.Join(missing, " "))
	}
}
