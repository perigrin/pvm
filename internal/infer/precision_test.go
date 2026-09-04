// ABOUTME: Measures PSC inference precision against types observed in a real perl run.
// ABOUTME: Perl is the oracle, so no hand-written annotations are needed.

package infer_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/types"
)

// RECALL AND PRECISION ARE DIFFERENT MEASUREMENTS, and PSC only had the first.
//
// "96.8% of value nodes typed" says how OFTEN inference answers. It says
// nothing about whether the answer is RIGHT, and the two move independently:
// a change that turns Unknowns into confident-but-wrong types raises coverage
// and makes the checker worse. TypeEvalPy names these recall and precision and
// reports them separately, for exactly that reason.
//
// The hard part is ground truth, since perl has no annotations to harvest. So
// perl itself is the oracle: testdata/precision_corpus.pl runs, records what
// each variable ACTUALLY held, and this test compares that against what PSC
// predicted statically.
//
// The limit is real and worth stating plainly: only executed paths are
// covered. A branch never taken contributes nothing. That is a KNOWN limit,
// where a hand-written expectation table has an unknown one.

type observation struct {
	Line int    `json:"line"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// observedToLattice maps an observed type name to the PSC type it should be.
var observedToLattice = map[string]types.Type{
	"Undef":     types.Undef,
	"Bool":      types.Bool,
	"Int":       types.Int,
	"Num":       types.Num,
	"Str":       types.Str,
	"Regex":     types.Regex,
	"ScalarRef": types.ScalarRef,
	"ArrayRef":  types.ArrayRef,
	"HashRef":   types.HashRef,
	"CodeRef":   types.CodeRef,
	"GlobRef":   types.GlobRef,
	"Object":    types.Object,
}

func runObserver(t *testing.T) []observation {
	t.Helper()
	perl, err := exec.LookPath("perl")
	if err != nil {
		t.Skip("perl not found in PATH — the precision oracle runs a real interpreter")
	}

	cmd := exec.Command(perl, "internal/infer/testdata/precision_corpus.pl")
	cmd.Dir = repoRoot(t)
	out, err := cmd.Output()
	require.NoError(t, err, "precision corpus failed to run")

	var line string
	for _, l := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "[") {
			line = strings.TrimSpace(l)
		}
	}
	require.NotEmpty(t, line, "observer produced no JSON")

	var obs []observation
	require.NoError(t, json.Unmarshal([]byte(line), &obs), "observer emitted invalid JSON")
	require.NotEmpty(t, obs, "observer recorded nothing")
	return obs
}

// repoRoot returns the module root, since the corpus requires the observer by
// a repo-relative path.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	// tests run in internal/infer
	return strings.TrimSuffix(wd, "/internal/infer")
}

// TestInferencePrecision reports how often PSC's inferred type agrees with the
// type observed at runtime, and prints every disagreement.
func TestInferencePrecision(t *testing.T) {
	obs := runObserver(t)

	corpus := repoRoot(t) + "/internal/infer/testdata/precision_corpus.pl"
	src, err := os.ReadFile(corpus)
	require.NoError(t, err)

	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	ann, _, st := infer.Analyze(tree, src, nil)
	_ = ann

	var exact, compatible, wrong, unknown int
	var disagreements []string
	var widened []string

	for _, o := range obs {
		want, ok := observedToLattice[o.Type]
		if !ok {
			t.Fatalf("observed type %q has no lattice mapping", o.Type)
		}

		sym, found := st.Lookup(o.Name)
		if !found || sym.Type == types.Unknown {
			unknown++
			continue
		}
		got := sym.Type

		switch {
		case got == want:
			exact++
		case types.IsSubtype(want, got):
			// PSC was wider than the observed value but not wrong: an Int
			// really is a Num. Correct, less precise.
			compatible++
			widened = append(widened,
				fmt.Sprintf("  line %d %-9s observed %-9s PSC said %s",
					o.Line, o.Name, o.Type, got))
		default:
			wrong++
			disagreements = append(disagreements,
				fmt.Sprintf("  line %d %-9s observed %-9s PSC said %s",
					o.Line, o.Name, o.Type, got))
		}
	}

	total := exact + compatible + wrong + unknown
	pct := func(n int) float64 { return 100 * float64(n) / float64(total) }

	fmt.Printf("\nPRECISION against %d observed values:\n", total)
	fmt.Printf("  exact:      %3d (%.1f%%)\n", exact, pct(exact))
	fmt.Printf("  wider:      %3d (%.1f%%)  correct but less precise\n", compatible, pct(compatible))
	fmt.Printf("  WRONG:      %3d (%.1f%%)\n", wrong, pct(wrong))
	fmt.Printf("  no answer:  %3d (%.1f%%)\n", unknown, pct(unknown))
	if len(widened) > 0 {
		sort.Strings(widened)
		fmt.Printf("\nwider than observed (correct, less precise):\n%s\n", strings.Join(widened, "\n"))
	}
	if len(disagreements) > 0 {
		sort.Strings(disagreements)
		fmt.Printf("\ndisagreements:\n%s\n", strings.Join(disagreements, "\n"))
	}

	// A wrong answer is the failure that matters: it is worse than saying
	// nothing, because a consumer acts on it.
	assert.Zero(t, wrong,
		"PSC inferred a type the value does not have; see the disagreements above")
}
