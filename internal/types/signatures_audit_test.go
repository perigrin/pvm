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

// --- split: the pattern is a Regex OR a Str ---
// perldoc writes it `split /PATTERN/, EXPR, LIMIT`, but perl compiles a plain
// string into a pattern, so all of these are ordinary correct Perl:
//
//	split /:/, $x      split ":", $x      split $sep, $x      split qr/:/, $x
//
// Measured on 5.42: all four yield ("a","b","c") for "a:b:c". Requiring Regex
// alone made `split ":", $x` a type error, which was 51 of the diagnostics on
// perl5/lib — the single largest false-positive class.
//
// Regex|Str rather than Scalar: the position must still reject a reference,
// and it does.

func TestAuditSplitPatternType(t *testing.T) {
	sig, ok := types.GetBuiltin("split")
	require.True(t, ok, "split should be a known builtin")
	require.True(t, len(sig.ArgTypes) >= 1, "split should have at least 1 arg type")

	assert.True(t, types.TypeSatisfies(types.Regex, sig.ArgTypes[0]),
		"split /:/, $x — a compiled pattern is accepted")
	assert.True(t, types.TypeSatisfies(types.Str, sig.ArgTypes[0]),
		"split \":\", $x — a string pattern is accepted, perl compiles it")
	assert.False(t, types.TypeSatisfies(types.HashRef, sig.ArgTypes[0]),
		"split $hashref, $x — a reference is still rejected")
	assert.NotEqual(t, types.Any, sig.ArgTypes[0],
		"the position must still constrain something")
}

// --- join: separator then a LIST ---
// perldoc: join EXPR, LIST. The second position is the list itself, not one
// string: `join ":", @foo` passes an array that flattens into LIST, and the
// variadic tail repeats this entry, so Str here rejected every array — 15 of
// the arity/type diagnostics on perl5/lib were exactly that.
//
// List rather than Any: the point of the audit is that the position must
// still constrain something, and List excludes Code and Glob, which cannot
// appear in a list to be joined.

func TestAuditJoinListElements(t *testing.T) {
	sig, ok := types.GetBuiltin("join")
	require.True(t, ok, "join should be a known builtin")
	require.True(t, len(sig.ArgTypes) >= 2, "join should have at least 2 arg types")
	assert.Equal(t, types.List, sig.ArgTypes[1],
		"join's second position is LIST — an array flattens into it")
	assert.NotEqual(t, types.Any, sig.ArgTypes[1],
		"but not Any: the position must still constrain something")

	// A Str is still acceptable there, since Scalar <: List.
	assert.True(t, types.TypeSatisfies(types.Str, sig.ArgTypes[1]),
		"join \":\", \"a\", \"b\" — strings satisfy the list position")
	assert.True(t, types.TypeSatisfies(types.Array, sig.ArgTypes[1]),
		"join \":\", @foo — an array satisfies the list position")
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
