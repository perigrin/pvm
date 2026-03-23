// ABOUTME: Tests for builtin function, binary operator, and unary operator signatures.
// ABOUTME: Covers struct construction, lookup functions, and completeness of all signature maps.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/types"
)

// TestSignatureStructs verifies that the three signature struct types can be
// constructed and their fields hold the assigned values.
func TestSignatureStructs(t *testing.T) {
	t.Run("BuiltinSig", func(t *testing.T) {
		sig := types.BuiltinSig{
			MinArity:   2,
			ArgTypes:   []types.Type{types.Array, types.Any},
			ReturnType: types.Int,
		}
		assert.Equal(t, 2, sig.MinArity)
		assert.Equal(t, []types.Type{types.Array, types.Any}, sig.ArgTypes)
		assert.Equal(t, types.Int, sig.ReturnType)
	})

	t.Run("BinaryOpSig", func(t *testing.T) {
		sig := types.BinaryOpSig{
			Left:   types.Num,
			Right:  types.Num,
			Result: types.Num,
		}
		assert.Equal(t, types.Num, sig.Left)
		assert.Equal(t, types.Num, sig.Right)
		assert.Equal(t, types.Num, sig.Result)
	})

	t.Run("UnaryOpSig", func(t *testing.T) {
		sig := types.UnaryOpSig{
			Operand: types.Num,
			Result:  types.Num,
		}
		assert.Equal(t, types.Num, sig.Operand)
		assert.Equal(t, types.Num, sig.Result)
	})
}

// TestGetBuiltinPush verifies the signature for the push builtin.
func TestGetBuiltinPush(t *testing.T) {
	sig, ok := types.GetBuiltin("push")
	require.True(t, ok, "push should be a known builtin")
	assert.Equal(t, 2, sig.MinArity)
	assert.Equal(t, []types.Type{types.Array, types.Any}, sig.ArgTypes)
	assert.Equal(t, types.Int, sig.ReturnType)
}

// TestGetBuiltinSplit verifies the signature for the split builtin, which has
// a variadic trailing argument.
func TestGetBuiltinSplit(t *testing.T) {
	sig, ok := types.GetBuiltin("split")
	require.True(t, ok, "split should be a known builtin")
	assert.Equal(t, 0, sig.MinArity)
	assert.Equal(t, []types.Type{types.Regex, types.Str, types.Int}, sig.ArgTypes)
	assert.Equal(t, types.List, sig.ReturnType)
}

// TestGetBuiltinDefined verifies the signature for the defined builtin.
func TestGetBuiltinDefined(t *testing.T) {
	sig, ok := types.GetBuiltin("defined")
	require.True(t, ok, "defined should be a known builtin")
	assert.Equal(t, 0, sig.MinArity)
	assert.Equal(t, []types.Type{types.Scalar}, sig.ArgTypes)
	assert.Equal(t, types.Bool, sig.ReturnType)
}

// TestGetBuiltinUnknown verifies that an unknown builtin name returns false.
func TestGetBuiltinUnknown(t *testing.T) {
	_, ok := types.GetBuiltin("nonexistent_builtin_xyz")
	assert.False(t, ok, "unknown builtin should return false")
}

// TestHasBuiltin verifies the convenience predicate for checking builtin existence.
func TestHasBuiltin(t *testing.T) {
	assert.True(t, types.HasBuiltin("print"), "print should be a known builtin")
	assert.True(t, types.HasBuiltin("map"), "map should be a known builtin")
	assert.True(t, types.HasBuiltin("sort"), "sort should be a known builtin")
	assert.False(t, types.HasBuiltin("notabuiltin"), "notabuiltin should not be known")
}

// TestGetBinaryOpArithmetic verifies the signatures for arithmetic operators.
func TestGetBinaryOpArithmetic(t *testing.T) {
	ops := []string{"+", "-", "*", "/", "%", "**"}
	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			sig, ok := types.GetBinaryOp(op)
			require.True(t, ok, "%q should be a known binary op", op)
			assert.Equal(t, types.Num, sig.Left)
			assert.Equal(t, types.Num, sig.Right)
			assert.Equal(t, types.Num, sig.Result)
		})
	}
}

// TestGetBinaryOpConcat verifies the string concatenation and repeat operators.
func TestGetBinaryOpConcat(t *testing.T) {
	t.Run(".", func(t *testing.T) {
		sig, ok := types.GetBinaryOp(".")
		require.True(t, ok, ". should be a known binary op")
		assert.Equal(t, types.Str, sig.Left)
		assert.Equal(t, types.Str, sig.Right)
		assert.Equal(t, types.Str, sig.Result)
	})

	t.Run("x", func(t *testing.T) {
		sig, ok := types.GetBinaryOp("x")
		require.True(t, ok, "x should be a known binary op")
		assert.Equal(t, types.Str, sig.Left)
		assert.Equal(t, types.Int, sig.Right)
		assert.Equal(t, types.Str, sig.Result)
	})
}

// TestGetBinaryOpComparison verifies numeric comparison operators.
func TestGetBinaryOpComparison(t *testing.T) {
	ops := []string{"==", "!=", "<", ">", "<=", ">="}
	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			sig, ok := types.GetBinaryOp(op)
			require.True(t, ok, "%q should be a known binary op", op)
			assert.Equal(t, types.Num, sig.Left)
			assert.Equal(t, types.Num, sig.Right)
			assert.Equal(t, types.Bool, sig.Result)
		})
	}
}

// TestGetBinaryOpSpaceship verifies the <=> operator returns Int (not Bool),
// since it produces -1, 0, or 1.
func TestGetBinaryOpSpaceship(t *testing.T) {
	sig, ok := types.GetBinaryOp("<=>")
	require.True(t, ok, "\"<=>\" should be a known binary op")
	assert.Equal(t, types.Num, sig.Left)
	assert.Equal(t, types.Num, sig.Right)
	assert.Equal(t, types.Int, sig.Result)
}

// TestGetBinaryOpStringCmp verifies string comparison operators.
func TestGetBinaryOpStringCmp(t *testing.T) {
	ops := []string{"eq", "ne", "lt", "gt", "le", "ge"}
	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			sig, ok := types.GetBinaryOp(op)
			require.True(t, ok, "%q should be a known binary op", op)
			assert.Equal(t, types.Str, sig.Left)
			assert.Equal(t, types.Str, sig.Right)
			assert.Equal(t, types.Bool, sig.Result)
		})
	}
}

// TestGetBinaryOpCmp verifies the cmp operator returns Int (not Bool),
// since it produces -1, 0, or 1.
func TestGetBinaryOpCmp(t *testing.T) {
	sig, ok := types.GetBinaryOp("cmp")
	require.True(t, ok, "\"cmp\" should be a known binary op")
	assert.Equal(t, types.Str, sig.Left)
	assert.Equal(t, types.Str, sig.Right)
	assert.Equal(t, types.Int, sig.Result)
}

// TestGetBinaryOpUnknown verifies that an unknown operator name returns false.
func TestGetBinaryOpUnknown(t *testing.T) {
	_, ok := types.GetBinaryOp("@@@@unknown@@@@")
	assert.False(t, ok, "unknown operator should return false")
}

// TestGetUnaryOp verifies all six unary operator signatures.
func TestGetUnaryOp(t *testing.T) {
	cases := []struct {
		op      string
		operand types.Type
		result  types.Type
	}{
		{"-", types.Num, types.Num},
		{"+", types.Num, types.Num},
		{"!", types.Any, types.Bool},
		{"not", types.Any, types.Bool},
		{"~", types.Int, types.Int},
		{`\`, types.Any, types.Ref},
	}

	for _, tc := range cases {
		t.Run(tc.op, func(t *testing.T) {
			sig, ok := types.GetUnaryOp(tc.op)
			require.True(t, ok, "%q should be a known unary op", tc.op)
			assert.Equal(t, tc.operand, sig.Operand, "wrong Operand for %q", tc.op)
			assert.Equal(t, tc.result, sig.Result, "wrong Result for %q", tc.op)
		})
	}
}

// TestAllBuiltinsHaveReturnType verifies that every registered builtin has a
// ReturnType that is not Unknown. This guards against accidental zero-value entries.
func TestAllBuiltinsHaveReturnType(t *testing.T) {
	builtinNames := []string{
		"push", "pop", "shift", "unshift", "splice",
		"keys", "values", "delete", "exists", "each",
		"length", "chomp", "chop", "chr", "ord",
		"join", "split", "sprintf", "substr",
		"defined", "ref", "scalar",
		"die", "warn", "bless",
		"print", "say", "return",
		"map", "grep", "sort",
	}

	for _, name := range builtinNames {
		t.Run(name, func(t *testing.T) {
			sig, ok := types.GetBuiltin(name)
			require.True(t, ok, "%q should be a registered builtin", name)
			assert.NotEqual(t, types.Unknown, sig.ReturnType,
				"builtin %q must have a non-Unknown ReturnType", name)
		})
	}
}
