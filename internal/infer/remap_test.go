// ABOUTME: Internal tests for the remapGuardVar helper function.
// ABOUTME: Uses package infer (not infer_test) to access unexported guardResult.

package infer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/types"
)

func TestRemapGuardVarSimple(t *testing.T) {
	guard := &guardResult{
		VarName: "$val",
		Guard:   types.GuardPattern{Kind: types.GuardRef},
	}
	remapped := remapGuardVar(guard, "$val", "$x")
	require.NotNil(t, remapped)
	assert.Equal(t, "$x", remapped.VarName)
	assert.Equal(t, types.GuardRef, remapped.Guard.Kind)
}

func TestRemapGuardVarNoMatch(t *testing.T) {
	guard := &guardResult{
		VarName: "$other",
		Guard:   types.GuardPattern{Kind: types.GuardDefined},
	}
	remapped := remapGuardVar(guard, "$val", "$x")
	require.NotNil(t, remapped)
	assert.Equal(t, "$other", remapped.VarName, "should not remap non-matching var")
}

func TestRemapGuardVarCompound(t *testing.T) {
	guard := &guardResult{
		Compound: &compoundGuard{
			Op: "&&",
			Left: &guardResult{
				VarName: "$_[0]",
				Guard:   types.GuardPattern{Kind: types.GuardDefined},
			},
			Right: &guardResult{
				VarName: "$_[0]",
				Guard:   types.GuardPattern{Kind: types.GuardRef},
			},
		},
	}
	remapped := remapGuardVar(guard, "$_[0]", "$x")
	require.NotNil(t, remapped)
	require.NotNil(t, remapped.Compound)
	assert.Equal(t, "$x", remapped.Compound.Left.VarName)
	assert.Equal(t, "$x", remapped.Compound.Right.VarName)
}

func TestRemapGuardVarNil(t *testing.T) {
	assert.Nil(t, remapGuardVar(nil, "$val", "$x"))
}
