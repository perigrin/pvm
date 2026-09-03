// ABOUTME: Requires the PSC lattice to agree with edge verdicts measured from real Perl.
// ABOUTME: Verdicts come from testdata/lattice_oracle.pl, which computes rather than declares them.

package types_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/types"
)

// WHY THIS FILE EXISTS ALONGSIDE paper_alignment_test.go.
//
// paper_alignment_test.go asserts the lattice against a table transcribed by
// hand from the paper. That catches an edit nobody reflected in the tests, but
// it cannot catch a table that was wrong when transcribed, because the table is
// its own oracle. Two instances of exactly that turned up while aligning to
// paper commit 04c8107: chalk's type table stayed green throughout the time its
// Boolean edge was wrong, and the paper's own appendix verifier reported PASS
// over all 729 pairs while asserting the claim the paper had just abandoned.
//
// Here Perl is the oracle. testdata/lattice_oracle.pl measures both membership
// factors for each edge and COMPUTES holds = factor1 && factor2; this test
// requires the Go lattice to agree with whatever came back. Nothing in either
// file writes the expected verdict down, so the test can disagree with its
// author.

// perlEdge is one measured verdict from the oracle script.
type perlEdge struct {
	Child      string   `json:"child"`
	Parent     string   `json:"parent"`
	Holds      int      `json:"holds"`
	Factor1    int      `json:"factor1"`
	Factor2    int      `json:"factor2"`
	F1Fail     []string `json:"f1_fail"`
	F1Untested []string `json:"f1_untested"`
	F2Fail     []string `json:"f2_fail"`
}

// oracleTypes maps the oracle's type names to PSC's Type values. A name the
// oracle emits that is missing here fails the test rather than being skipped:
// a silently ignored edge is the failure mode this whole file exists to avoid.
var oracleTypes = map[string]types.Type{
	"Undef":     types.Undef,
	"Bool":      types.Bool,
	"Int":       types.Int,
	"Num":       types.Num,
	"Str":       types.Str,
	"NaN":       types.NaN,
	"Inf":       types.Inf,
	"Regex":     types.Regex,
	"ScalarRef": types.ScalarRef,
	"ArrayRef":  types.ArrayRef,
	"HashRef":   types.HashRef,
	"CodeRef":   types.CodeRef,
	"GlobRef":   types.GlobRef,
	"Object":    types.Object,
	"Ref":       types.Ref,
	"Scalar":    types.Scalar,
	"List":      types.List,
}

// runLatticeOracle executes the Perl oracle and returns its measured edges.
// The test is skipped when no perl is available, since the oracle measures a
// real interpreter and cannot be faked.
func runLatticeOracle(t *testing.T) []perlEdge {
	t.Helper()

	perlPath, err := exec.LookPath("perl")
	if err != nil {
		t.Skip("perl not found in PATH — the lattice oracle measures a real interpreter")
	}

	out, err := exec.Command(perlPath, "testdata/lattice_oracle.pl").Output()
	require.NoError(t, err, "lattice oracle failed to run")

	// The oracle prints one JSON array on the last non-empty line; shell
	// startup noise on some systems can precede it.
	var line string
	for _, l := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "[") {
			line = strings.TrimSpace(l)
		}
	}
	require.NotEmpty(t, line, "lattice oracle produced no JSON output")

	var edges []perlEdge
	require.NoError(t, json.Unmarshal([]byte(line), &edges), "lattice oracle emitted invalid JSON")
	require.NotEmpty(t, edges, "lattice oracle measured no edges")

	return edges
}

// TestLatticeAgreesWithPerl requires IsSubtype to match the verdict measured
// from Perl for every edge the oracle covers.

func TestLatticeAgreesWithPerl(t *testing.T) {
	edges := runLatticeOracle(t)

	for _, e := range edges {
		child, ok := oracleTypes[e.Child]
		require.True(t, ok, "oracle measured type %q with no PSC mapping", e.Child)
		parent, ok := oracleTypes[e.Parent]
		require.True(t, ok, "oracle measured type %q with no PSC mapping", e.Parent)

		measured := e.Holds == 1
		actual := types.IsSubtype(child, parent)

		assert.Equal(t, measured, actual,
			"%s <: %s — perl measured %v (factor1=%d factor2=%d f1_fail=%v f2_fail=%v), lattice says %v",
			e.Child, e.Parent, measured, e.Factor1, e.Factor2, e.F1Fail, e.F2Fail, actual)
	}
}

// TestLatticeOracleMeasuredBothFactors guards the oracle itself. An edge with
// no membership oracle reports untested, and a run where nothing was untested
// but also nothing ever failed would mean the machinery is not discriminating
// — every predicate returning true measures nothing.

func TestLatticeOracleMeasuredBothFactors(t *testing.T) {
	edges := runLatticeOracle(t)

	for _, e := range edges {
		assert.Empty(t, e.F1Untested,
			"%s <: %s has witnesses with no membership oracle: %v — add one or drop the edge",
			e.Child, e.Parent, e.F1Untested)
	}

	// At least one edge must FAIL, or the oracle is not discriminating. Bool
	// <: Str is that edge: factor 2 passes and factor 1 fails on `false`
	// alone, which is the asymmetry the paper's 04c8107 turns on.
	var sawFailure, sawBoolStr bool
	for _, e := range edges {
		if e.Holds == 0 {
			sawFailure = true
		}
		if e.Child == "Bool" && e.Parent == "Str" {
			sawBoolStr = true
			assert.Equal(t, 0, e.Holds, "Bool <: Str must not hold")
			assert.Equal(t, 1, e.Factor2,
				"Bool <: Str factor 2 (substitutability) SHOULD pass — string ops work on booleans uncoerced")
			assert.Equal(t, 0, e.Factor1,
				"Bool <: Str factor 1 (membership) should fail")
			assert.Equal(t, []string{"false"}, e.F1Fail,
				"only `false` should fail — `true` passes, which is why single-value testing got this wrong")
		}
	}
	assert.True(t, sawFailure,
		"no edge failed — the oracle is not discriminating, every predicate may be returning true")
	assert.True(t, sawBoolStr,
		"the Bool <: Str edge is missing — it is the worked example both factors exist for")
}
