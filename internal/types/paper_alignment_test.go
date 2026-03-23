// ABOUTME: Tests validating the PSC type lattice against the formal paper definition.
// ABOUTME: Covers subtype chain, exclusion rules, blessed ref unions, and bottom/top properties.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/types"
)

// --- Subtype chain: Int <: Num <: Str <: Scalar ---
// The paper defines type membership via two-component testing:
//   1. Syntactic preservation: round-trip through the type's domain preserves value
//   2. Semantic fulfillment: all operations in Operations(T) satisfy their contracts
// Int <: Num because every integer is a valid number (both components hold).
// Num <: Str because every number has a lossless string representation.
// Str <: Scalar because strings maintain identity through scalar operations.

func TestPaperSubtypeChain(t *testing.T) {
	assert.True(t, types.IsSubtype(types.Int, types.Num),
		"Int <: Num — every integer is a valid number")
	assert.True(t, types.IsSubtype(types.Num, types.Str),
		"Num <: Str — every number has a lossless string representation")
	assert.True(t, types.IsSubtype(types.Str, types.Scalar),
		"Str <: Scalar — strings maintain identity through scalar operations")

	// Transitive: Int <: Str (through Num)
	assert.True(t, types.IsSubtype(types.Int, types.Str),
		"Int <: Str — transitive through Num")
	// Transitive: Int <: Scalar (through Num, Str)
	assert.True(t, types.IsSubtype(types.Int, types.Scalar),
		"Int <: Scalar — transitive through Num and Str")
}

// --- DualVar exclusion ---
// The paper proves DualVar sits in Scalar but outside the Str/Num branches.
// Example: Scalar::Util::dualvar(42, "hello") — numerically 42 but stringifies
// to "hello". Fails Num test (stringified detour != direct path) and fails Str
// test (numified detour != direct path). Passes Scalar test.

func TestPaperDualVarExclusion(t *testing.T) {
	assert.False(t, types.IsSubtype(types.DualVar, types.Str),
		"DualVar is NOT a subtype of Str — string and numeric representations diverge")
	assert.False(t, types.IsSubtype(types.DualVar, types.Num),
		"DualVar is NOT a subtype of Num — string and numeric representations diverge")
	assert.True(t, types.IsSubtype(types.DualVar, types.Scalar),
		"DualVar IS a subtype of Scalar — maintains identity through scalar operations")
}

// --- NaN exclusion ---
// The paper proves NaN sits in Scalar but outside Str and Num.
// "NaN" passes syntactic preservation for Num ("NaN" -> NaN -> "NaN" round-trips)
// but fails semantic fulfillment (NaN != NaN violates reflexivity).
// NaN is excluded from Str because its string representation is a representational
// artifact, not a meaningful string identity.

func TestPaperNaNExclusion(t *testing.T) {
	assert.False(t, types.IsSubtype(types.NaN, types.Num),
		"NaN is NOT a subtype of Num — fails semantic fulfillment (NaN != NaN)")
	assert.False(t, types.IsSubtype(types.NaN, types.Str),
		"NaN is NOT a subtype of Str — 'NaN' is a representational artifact")
	assert.True(t, types.IsSubtype(types.NaN, types.Scalar),
		"NaN IS a subtype of Scalar")
}

// --- Inf exclusion ---
// Inf sits in Scalar but outside Str and Num, same lattice position as NaN.
// Inf passes syntactic preservation but fails semantic fulfillment
// (Inf - Inf = NaN violates subtraction identity).

func TestPaperInfExclusion(t *testing.T) {
	assert.False(t, types.IsSubtype(types.Inf, types.Num),
		"Inf is NOT a subtype of Num — Inf - Inf = NaN violates subtraction identity")
	assert.False(t, types.IsSubtype(types.Inf, types.Str),
		"Inf is NOT a subtype of Str — 'Inf' is a representational artifact")
	assert.True(t, types.IsSubtype(types.Inf, types.Scalar),
		"Inf IS a subtype of Scalar")
}

// --- NaN and Inf are distinct ---

func TestPaperNaNInfDistinct(t *testing.T) {
	assert.False(t, types.IsSubtype(types.NaN, types.Inf),
		"NaN is NOT a subtype of Inf — distinct IEEE 754 special values")
	assert.False(t, types.IsSubtype(types.Inf, types.NaN),
		"Inf is NOT a subtype of NaN — distinct IEEE 754 special values")
	assert.NotEqual(t, types.NaN, types.Inf,
		"NaN and Inf occupy different bit positions")
}

// --- Blessed reference unions ---
// The paper says a blessed hashref satisfies both Object AND HashRef.
// In the bitset, Object|HashRef is a union type representing this.

func TestPaperBlessedRefUnion(t *testing.T) {
	blessedHashRef := types.Object | types.HashRef

	// A blessed hashref is a subtype of Ref
	assert.True(t, types.IsSubtype(blessedHashRef, types.Ref),
		"Object|HashRef is a subtype of Ref")

	// A blessed hashref is a subtype of Scalar
	assert.True(t, types.IsSubtype(blessedHashRef, types.Scalar),
		"Object|HashRef is a subtype of Scalar")

	// A blessed hashref satisfies Object requirements
	assert.True(t, types.TypeSatisfies(blessedHashRef, types.Object),
		"Object|HashRef satisfies Object — can call methods")

	// A blessed hashref satisfies HashRef requirements
	assert.True(t, types.TypeSatisfies(blessedHashRef, types.HashRef),
		"Object|HashRef satisfies HashRef — can dereference as hash")

	// A blessed hashref satisfies Ref requirements
	assert.True(t, types.TypeSatisfies(blessedHashRef, types.Ref),
		"Object|HashRef satisfies Ref")
}

// --- Bottom and top type properties ---

func TestPaperBottomTopTypes(t *testing.T) {
	// None (bottom) is a subtype of every type
	allTypes := []types.Type{
		types.Undef, types.Bool, types.Int, types.Num, types.Str,
		types.DualVar, types.NaN, types.Inf, types.Regex,
		types.ScalarRef, types.ArrayRef, types.HashRef, types.CodeRef,
		types.GlobRef, types.Object, types.Ref,
		types.Array, types.Hash, types.List,
		types.Code, types.Glob,
		types.Scalar, types.Any,
	}
	for _, typ := range allTypes {
		assert.True(t, types.IsSubtype(types.None, typ),
			"None should be a subtype of %s", typ)
	}

	// Unknown (zero) is only a subtype of Unknown
	assert.True(t, types.IsSubtype(types.Unknown, types.Unknown),
		"Unknown is a subtype of itself")
	assert.False(t, types.IsSubtype(types.Unknown, types.Any),
		"Unknown is NOT a subtype of Any")

	// Any contains all concrete leaf bits
	assert.True(t, types.Any&types.NaN == types.NaN,
		"Any should contain NaN")
	assert.True(t, types.Any&types.Inf == types.Inf,
		"Any should contain Inf")
}
