// ABOUTME: A union prints as the largest families it holds, then any leaves left:
// ABOUTME: Undef|Num, not Undef|Int|Num -- and a printed family reads back as itself.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/types"
)

func TestUnionPrintsItsFamilies(t *testing.T) {
	cases := []struct {
		typ  types.Type
		want string
	}{
		{types.Undef | types.Num, "Undef|Num"},
		{types.Undef | types.Str, "Undef|Str"},
		{types.Num | types.Inf, "Num|Inf"},
		// a whole blessed family is BlessedRef, not Regex|BlessedRef|Object
		{types.Undef | types.BlessedRef, "Undef|BlessedRef"},
		{types.HashRef | types.BlessedRef, "HashRef|BlessedRef"},
		{types.Scalar &^ types.Undef, "Boolean|Str|DualVar|Ref"},
		{types.Undef | types.Ref | types.Array, "Undef|Ref|Array"},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			assert.Equal(t, c.want, c.typ.String())
			back, err := types.FromName(c.typ.String())
			require.NoError(t, err)
			assert.Equal(t, c.typ, back, "%s reads back as itself", c.want)
		})
	}
}
