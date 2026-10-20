// ABOUTME: Tests for Options.WarnPossibleUndef: a declared call's Undef is silent by default unless its own arguments prove it.
// ABOUTME: Covers a single declared candidate (blessed), a multi's catch-all (refaddr), and a proven arity (sum()).

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

// TestWarnPossibleUndefSingleCandidate: Scalar::Util::blessed has one
// declared candidate, `(Any $ref) Str|Undef` -- no sibling, no multi at
// all -- yet perigrin's ruling (2026-10-20) covers it exactly the same
// as a multi's catch-all: the call's own argument never proves whether
// $x is a blessed reference, so blessed($x) is one of two outcomes the
// declaration itself states, not one the call's arguments proved, and
// its Undef is silent by default, reported only under the flag.
func TestWarnPossibleUndefSingleCandidate(t *testing.T) {
	const src = `use Scalar::Util qw(blessed);
my ($x) = @_;
my $same = (blessed($x) eq "Some::Class");
`
	assert.Empty(t, diagsWithOptions(src, infer.Options{}))
	diags := diagsWithOptions(src, infer.Options{WarnPossibleUndef: true})
	if assert.Len(t, diags, 1) {
		assert.Equal(t, infer.CodeCoercionMismatch, diags[0].Code)
	}
}

// TestWarnPossibleUndefCatchAllCandidate: Scalar::Util::refaddr is a
// multi, and its widest, catch-all candidate, `(Any $ref) Int|Undef`,
// selected over a Scalar argument never proven a Ref, is the same
// shape: the declaration states two outcomes, and the call's own
// argument does not rule either out.
func TestWarnPossibleUndefCatchAllCandidate(t *testing.T) {
	const src = `use Scalar::Util qw(refaddr);
my ($x, $y) = @_;
my $same = (refaddr($x) eq refaddr($y));
`
	assert.Empty(t, diagsWithOptions(src, infer.Options{}))
	diags := diagsWithOptions(src, infer.Options{WarnPossibleUndef: true})
	if assert.Len(t, diags, 2) {
		assert.Equal(t, infer.CodeCoercionMismatch, diags[0].Code)
		assert.Equal(t, infer.CodeCoercionMismatch, diags[1].Code)
	}
}

// TestWarnPossibleUndefProvenSelectionReports: List::Util::sum's own
// `()` candidate, the only one zero arguments admit, answers nothing
// but Undef -- perigrin's own example of arguments proving the
// Undef-returning candidate selected, since sum()'s resolved return has
// no other bit left to be silent about. It reports unconditionally,
// flag or no flag.
func TestWarnPossibleUndefProvenSelectionReports(t *testing.T) {
	const src = `use List::Util qw(sum);
my $same = (sum() eq "0");
`
	assert.Len(t, diagsWithOptions(src, infer.Options{}), 1)
	assert.Len(t, diagsWithOptions(src, infer.Options{WarnPossibleUndef: true}), 1)
}
