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

// --- NaN and Inf are Str, not Num (paper §"Example 3", line 2102) ---
// The paper places NaN and Inf in Str: both pass syntactic preservation
// ("NaN" -> NaN -> "NaN" round-trips) and are excluded from Num by the
// SEMANTIC component alone. Example 3 states it directly: "NaN" ∈ Str
// because it satisfies both the syntactic and semantic requirements for
// string membership. The contract table distinguishes them — NaN fails
// Contract_==, Contract_- and Contract_+; Inf fails Contract_- alone.

func TestPaperNaNInfAreStrings(t *testing.T) {
	assert.True(t, types.IsSubtype(types.NaN, types.Str),
		"NaN IS a subtype of Str — stable string representation, correct under string operations")
	assert.True(t, types.IsSubtype(types.Inf, types.Str),
		"Inf IS a subtype of Str — stable string representation, correct under string operations")

	assert.False(t, types.IsSubtype(types.NaN, types.Num),
		"NaN is NOT a subtype of Num — fails Contract_== (NaN != NaN) and Contract_-")
	assert.False(t, types.IsSubtype(types.Inf, types.Num),
		"Inf is NOT a subtype of Num — passes Contract_== but fails Contract_- (Inf - Inf = NaN)")
}

// --- Arity ordering: Scalar <: List (paper §"Complete Type Hierarchy") ---
// Scalar, Void and List are distinguished by how many values they denote, and
// the subtype relation between them is subset inclusion on those arities:
// Scalar {1} ⊆ List {0,1,2,...} gives Scalar <: List. This states Perl's
// list-flattening rule as a subtype fact: a scalar satisfies a list position
// BECAUSE one value is one of the arities a list admits.

func TestPaperArityOrdering(t *testing.T) {
	assert.True(t, types.IsSubtype(types.Scalar, types.List),
		"Scalar <: List — {1} is one of the arities a list admits (list flattening)")
	assert.True(t, types.IsSubtype(types.Array, types.List),
		"Array <: List")
	assert.True(t, types.IsSubtype(types.Hash, types.List),
		"Hash <: List")
	assert.True(t, types.IsSubtype(types.Int, types.List),
		"Int <: List — transitive through Num, Str, Scalar")

	// List is NOT a subtype of Scalar: {0,1,2,...} is not contained in {1}.
	assert.False(t, types.IsSubtype(types.List, types.Scalar),
		"List is NOT a subtype of Scalar — a list may denote zero or many values")
}

// --- Regex <: Object <: Ref (paper §"Regex") ---
// Regex := {v ∈ Object | ref(v) eq 'Regexp'}. A compiled pattern is a blessed
// reference — ref is 'Regexp', blessed is 'Regexp', reftype is 'REGEXP' — which
// is this paper's definition of Object. So Regex is a subtype of Object rather
// than a sibling of Ref, and it participates in method dispatch.

func TestPaperRegexIsObject(t *testing.T) {
	assert.True(t, types.IsSubtype(types.Regex, types.Object),
		"Regex <: Object — a compiled pattern is a blessed reference")
	assert.True(t, types.IsSubtype(types.Regex, types.Ref),
		"Regex <: Ref — transitive through Object")
	assert.True(t, types.IsSubtype(types.Object, types.Ref),
		"Object <: Ref")

	// Regex is NOT a Str: syntactic preservation fails because qr// applied to
	// a stringified pattern NESTS it rather than reconstructing the original.
	assert.False(t, types.IsSubtype(types.Regex, types.Str),
		"Regex is NOT a subtype of Str — restringifying a pattern wraps it, losing the value")
}
