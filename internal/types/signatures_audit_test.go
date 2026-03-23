// ABOUTME: Tests validating builtin signatures against perldoc.
// ABOUTME: Covers MinArity fixes, ArgType corrections, and print/say Str acceptance.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/types"
)

// --- MinArity corrections: builtins that default to $_ when called with no args ---

func TestAuditMinArityDefaults(t *testing.T) {
	// These builtins can be called with zero arguments (defaulting to $_ or @_).
	zeroArity := []string{"pop", "shift", "chr", "defined", "ref"}
	for _, name := range zeroArity {
		t.Run(name, func(t *testing.T) {
			sig, ok := types.GetBuiltin(name)
			require.True(t, ok, "%s should be a known builtin", name)
			assert.Equal(t, 0, sig.MinArity,
				"%s can be called with zero args (defaults to $_ or @_)", name)
		})
	}
}

// --- print and say accept Str, not Any ---
// perldoc: print LIST — prints a string representation of each element.
// Arguments are stringified, so the expected type is Str (which includes
// Num and Int via subtyping). Passing a Ref produces "ARRAY(0x...)" which
// is almost never intended.

func TestAuditPrintSayAcceptStr(t *testing.T) {
	for _, name := range []string{"print", "say"} {
		t.Run(name, func(t *testing.T) {
			sig, ok := types.GetBuiltin(name)
			require.True(t, ok, "%s should be a known builtin", name)
			require.Len(t, sig.ArgTypes, 1, "%s should have 1 arg type (variadic Str)", name)
			assert.Equal(t, types.Str, sig.ArgTypes[0],
				"%s should accept Str (not Any) — args are stringified", name)
		})
	}
}

// --- splice uses Int offsets, not Num ---
// perldoc: splice ARRAY, OFFSET, LENGTH, LIST
// OFFSET and LENGTH are integer indices/counts.

func TestAuditSpliceIntOffsets(t *testing.T) {
	sig, ok := types.GetBuiltin("splice")
	require.True(t, ok, "splice should be a known builtin")
	require.True(t, len(sig.ArgTypes) >= 3, "splice should have at least 3 arg types")
	assert.Equal(t, types.Int, sig.ArgTypes[1],
		"splice OFFSET should be Int, not Num")
	assert.Equal(t, types.Int, sig.ArgTypes[2],
		"splice LENGTH should be Int, not Num")
}

// --- keys/values/each accept Hash|Array (since 5.12) ---
// perldoc: keys HASH or keys ARRAY

func TestAuditKeysValuesEachAcceptArray(t *testing.T) {
	hashOrArray := types.Hash | types.Array
	for _, name := range []string{"keys", "values", "each"} {
		t.Run(name, func(t *testing.T) {
			sig, ok := types.GetBuiltin(name)
			require.True(t, ok, "%s should be a known builtin", name)
			require.Len(t, sig.ArgTypes, 1, "%s should have 1 arg type", name)
			assert.Equal(t, hashOrArray, sig.ArgTypes[0],
				"%s should accept Hash|Array (since Perl 5.12)", name)
		})
	}
}

// --- split: first arg is Regex, not Scalar ---
// perldoc: split /PATTERN/, EXPR, LIMIT

func TestAuditSplitPatternType(t *testing.T) {
	sig, ok := types.GetBuiltin("split")
	require.True(t, ok, "split should be a known builtin")
	require.True(t, len(sig.ArgTypes) >= 1, "split should have at least 1 arg type")
	assert.Equal(t, types.Regex, sig.ArgTypes[0],
		"split first arg should be Regex (the /PATTERN/), not Scalar")
}

// --- join: second arg accepts Str (the list elements to join) ---
// perldoc: join EXPR, LIST — joins string representations.
// The variadic list elements should be Str, not Any.

func TestAuditJoinListElements(t *testing.T) {
	sig, ok := types.GetBuiltin("join")
	require.True(t, ok, "join should be a known builtin")
	require.True(t, len(sig.ArgTypes) >= 2, "join should have at least 2 arg types")
	assert.Equal(t, types.Str, sig.ArgTypes[1],
		"join list elements should be Str (not Any) — elements are stringified")
}

// --- chomp/chop accept Str, not Any ---
// perldoc: chomp VARIABLE — removes trailing newline from a string.
// chop VARIABLE — removes last character from a string.

func TestAuditChompChopAcceptStr(t *testing.T) {
	for _, name := range []string{"chomp", "chop"} {
		t.Run(name, func(t *testing.T) {
			sig, ok := types.GetBuiltin(name)
			require.True(t, ok, "%s should be a known builtin", name)
			require.Len(t, sig.ArgTypes, 1, "%s should have 1 arg type", name)
			assert.Equal(t, types.Str, sig.ArgTypes[0],
				"%s should accept Str (not Any) — operates on strings", name)
		})
	}
}

// --- die/warn accept Str, not Any ---
// perldoc: die LIST — raises an exception with the given message.
// warn LIST — prints a warning message.
// Arguments are stringified for the message.

func TestAuditDieWarnAcceptStr(t *testing.T) {
	for _, name := range []string{"die", "warn"} {
		t.Run(name, func(t *testing.T) {
			sig, ok := types.GetBuiltin(name)
			require.True(t, ok, "%s should be a known builtin", name)
			require.Len(t, sig.ArgTypes, 1, "%s should have 1 arg type", name)
			assert.Equal(t, types.Str, sig.ArgTypes[0],
				"%s should accept Str (not Any) — message is stringified", name)
		})
	}
}
