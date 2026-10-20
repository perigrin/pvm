// ABOUTME: Tests for Options.WarnPossibleUndef: a multi call Select left genuinely Undecided is silent by default.
// ABOUTME: A Selected call's own declared Undef, or an Undecided reached only by coercion retry, still reports.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// diagsWithOptions is resolvedDiags (flow_test.go) with the given Options,
// for WarnPossibleUndef, which the shared helper has no parameter for.
func diagsWithOptions(src string, opts infer.Options) []infer.Diagnostic {
	root := parse.ParseWithLoader([]byte(src), func(string) ([]byte, bool) { return nil, false })
	f := lower.Lower(root)
	_, d, _ := infer.Analyze(f, lower.NewPositions(root, f), nil, opts)
	return d
}

// undecidedRefaddrEq calls refaddr directly on an unresolved sub's own
// return, which carries no information at all -- Unknown, not some
// concrete type -- so Scalar::Util::refaddr's multi cannot decide
// between its Ref and Any candidates and joins them to Int|Undef (a
// plain `my ($x, $y) = @_`, Math::BigInt's own shape, does not do this:
// such a read is "Scalar", a concrete type that selects the Any
// candidate directly, Selected rather than Undecided --
// TestNarrowKeepsRealLibSites's reftype_unguarded is exactly this
// shape, and must keep reporting, so WarnPossibleUndef cannot reach it
// either). Reaching an "eq" through the genuine Undecided join is
// silent by default.
const undecidedRefaddrEq = `use Scalar::Util qw(refaddr);
my $same = (refaddr(Unresolved::opaque_x()) eq refaddr(Unresolved::opaque_y()));
`

// selectedRefaddrEq calls refaddr with a provably non-Ref argument, which
// Select resolves outright (Selected, not Undecided) to its declared
// Int|Undef: this Undef is not from a join of candidates, and reports
// whether or not WarnPossibleUndef is set.
const selectedRefaddrEq = `use Scalar::Util qw(refaddr);
my $same = (refaddr("plain string") eq "123");
`

// TestWarnPossibleUndefDefaultSilent: by default, an Undecided call's
// Undef reaches "eq" without a diagnostic.
func TestWarnPossibleUndefDefaultSilent(t *testing.T) {
	assert.Empty(t, diagsWithOptions(undecidedRefaddrEq, infer.Options{}))
}

// TestWarnPossibleUndefFlagReports: --warn-possible-undef reports it.
func TestWarnPossibleUndefFlagReports(t *testing.T) {
	diags := diagsWithOptions(undecidedRefaddrEq, infer.Options{WarnPossibleUndef: true})
	if assert.Len(t, diags, 2) {
		assert.Equal(t, infer.CodeCoercionMismatch, diags[0].Code)
		assert.Equal(t, infer.CodeCoercionMismatch, diags[1].Code)
	}
}

// TestWarnPossibleUndefSelectedStillReports: a Selected call's own
// declared Undef reports either way -- perigrin's ruling applies only to
// an Undecided join, never to what a single resolved candidate declares.
func TestWarnPossibleUndefSelectedStillReports(t *testing.T) {
	assert.Len(t, diagsWithOptions(selectedRefaddrEq, infer.Options{}), 1)
	assert.Len(t, diagsWithOptions(selectedRefaddrEq, infer.Options{WarnPossibleUndef: true}), 1)
}
