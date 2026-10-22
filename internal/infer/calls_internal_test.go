// ABOUTME: Internal tests for placedCandidate, the candidate a call with a flattening argument checks its leading arguments against.
// ABOUTME: Uses package infer (not infer_test) to reach the unexported helper with hand-built signatures.

package infer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"tamarou.com/pvm/internal/types"
)

func scalarParams(ts ...types.Type) []types.Param {
	ps := make([]types.Param, len(ts))
	for i, t := range ts {
		ps[i] = types.Param{Name: "p", Sigil: '$', Type: t, Required: true}
	}
	return ps
}

// TestPlacedCandidateRefusesWhenCountsDisagree: a multi with `(Str $a)`
// beside `(Int $a, Int $b)` selects the first at one argument and the
// second at two, so a call whose flattening argument may make either
// count has no one candidate, and its leading arguments are checked
// against none. So does one whose one-argument call cannot decide
// between `(Str $a)` and `(Num $a)`, though two arguments select `(Int
// $a, Int $b)`. One candidate every count selects is the control.
func TestPlacedCandidateRefusesWhenCountsDisagree(t *testing.T) {
	multi := []types.Signature{
		{Params: scalarParams(types.Str), Returns: types.Str},
		{Params: scalarParams(types.Int, types.Int), Returns: types.Int},
	}
	for _, placed := range [][]types.Type{nil, {types.Str}} {
		c, ok := placedCandidate(multi, placed, len(placed), types.ScalarCtx)
		assert.False(t, ok, "placed %v selected candidate %d", placed, c)
	}

	undecided := []types.Signature{
		{Params: scalarParams(types.Str), Returns: types.Str},
		{Params: scalarParams(types.Num), Returns: types.Num},
		{Params: scalarParams(types.Int, types.Int), Returns: types.Int},
	}
	c, ok := placedCandidate(undecided, nil, 0, types.ScalarCtx)
	assert.False(t, ok, "undecided count selected candidate %d", c)

	single := []types.Signature{{Params: scalarParams(types.Num, types.Num), Returns: types.Num}}
	c, ok = placedCandidate(single, []types.Type{types.Undef}, 1, types.ScalarCtx)
	assert.True(t, ok)
	assert.Equal(t, 0, c)
}
