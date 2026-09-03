// ABOUTME: Tests for IsCoercible, the coercion relation the paper defines separately from subtyping.
// ABOUTME: Coercibility is directional and lossy; subtyping is neither.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/types"
)

// The paper's "Coercion Rules" section defines what Perl will convert to what.
// That relation is NOT subtyping, and conflating the two is how a checker ends
// up either silent or wrong:
//
//   - Every value coerces to Str, including references ("HASH(0x...)"), but a
//     reference is not a Str — the string is metadata and does not round-trip.
//   - "hello" coerces to Num (0, with a warning) but is not a Num.
//
// Subtyping says a value ALREADY IS a member of the supertype; coercion says
// Perl will convert it, possibly losing the value. IsSubtype answers the first
// question and IsCoercible the second.

func TestCoercionIsNotSubtyping(t *testing.T) {
	// Everything stringifies, including references — but they are not Strs.
	for _, from := range []types.Type{
		types.Undef, types.Bool, types.Int, types.Num,
		types.HashRef, types.ArrayRef, types.CodeRef, types.Object,
	} {
		assert.True(t, types.IsCoercible(from, types.Str),
			"%s coerces to Str — Perl stringifies everything", from)
	}
	assert.False(t, types.IsSubtype(types.HashRef, types.Str),
		"HashRef is NOT a Str — 'HASH(0x...)' is metadata that does not round-trip")
	assert.False(t, types.IsSubtype(types.Undef, types.Str),
		"Undef is NOT a Str")

	// Str coerces to Num (non-numeric strings give 0, with a warning).
	assert.True(t, types.IsCoercible(types.Str, types.Num),
		"Str coerces to Num — non-numeric strings numify to 0")
	assert.False(t, types.IsSubtype(types.Str, types.Num),
		"Str is NOT a Num — 'hello' numifies to 0, losing the value")

	// Everything booleanises: the paper's "To Boolean" rules take all values.
	assert.True(t, types.IsCoercible(types.HashRef, types.Bool),
		"HashRef coerces to Bool — every value has a truth value")
	assert.False(t, types.IsSubtype(types.HashRef, types.Bool),
		"HashRef is NOT a Bool")
}

// Subtyping implies coercibility: if a value already IS a member of the
// supertype, no conversion is needed and the identity coercion applies.

func TestSubtypingImpliesCoercibility(t *testing.T) {
	pairs := []struct{ child, parent types.Type }{
		{types.Int, types.Num},
		{types.Num, types.Str},
		{types.Int, types.Str},
		{types.Str, types.Scalar},
		{types.Regex, types.Object},
		{types.Object, types.Ref},
		{types.Array, types.List},
		{types.Scalar, types.List},
	}
	for _, p := range pairs {
		assert.True(t, types.IsSubtype(p.child, p.parent),
			"precondition: %s <: %s", p.child, p.parent)
		assert.True(t, types.IsCoercible(p.child, p.parent),
			"%s <: %s implies %s coerces to %s", p.child, p.parent, p.child, p.parent)
	}
}

// Coercion is DIRECTIONAL where subtyping is not symmetric either, but the
// asymmetries differ: Str coerces to Num and Num coerces to Str, while only
// Num <: Str holds. A relation that got this wrong would accept a Str
// wherever an Int is required.

func TestCoercionDirectionality(t *testing.T) {
	assert.True(t, types.IsCoercible(types.Str, types.Num), "Str -> Num (numify)")
	assert.True(t, types.IsCoercible(types.Num, types.Str), "Num -> Str (stringify)")
	assert.True(t, types.IsSubtype(types.Num, types.Str), "Num <: Str")
	assert.False(t, types.IsSubtype(types.Str, types.Num), "Str is NOT <: Num")
}

// Aggregates and code do not coerce to scalars by conversion — an array in
// scalar context yields its COUNT, which is a context rule rather than a
// value-preserving coercion, and PSC models that in NarrowByContext.

func TestCoercionExcludesAggregates(t *testing.T) {
	assert.False(t, types.IsCoercible(types.Array, types.Num),
		"Array does not COERCE to Num — scalar context yields a count, which is NarrowByContext's job")
	assert.False(t, types.IsCoercible(types.Hash, types.Str),
		"Hash does not coerce to Str")
	assert.False(t, types.IsCoercible(types.Code, types.Str),
		"Code (a CV) does not coerce to Str")
}

// Nothing coerces INTO a reference type: Perl will not turn a string into a
// hashref. (Symbolic references are a separate mechanism, off under strict.)

func TestCoercionIntoRefsIsRejected(t *testing.T) {
	for _, to := range []types.Type{
		types.HashRef, types.ArrayRef, types.CodeRef, types.ScalarRef,
		types.GlobRef, types.Object, types.Regex,
	} {
		assert.False(t, types.IsCoercible(types.Str, to),
			"Str does not coerce to %s — Perl will not fabricate a reference", to)
		assert.False(t, types.IsCoercible(types.Int, to),
			"Int does not coerce to %s", to)
	}
}

// Unknown is epistemic, not a value: it means inference has not determined a
// type. It must not participate in coercion as though it were a real type.

func TestCoercionUnknown(t *testing.T) {
	assert.False(t, types.IsCoercible(types.Unknown, types.Str),
		"Unknown does not coerce — it is the absence of a claim, not a value")
	assert.False(t, types.IsCoercible(types.Str, types.Unknown),
		"nothing coerces to Unknown")
}

// CoercionMismatch names the gap between the two relations: code that runs
// but probably does not mean what it says.

func TestCoercionMismatch(t *testing.T) {
	// undef in arithmetic: runs (undef numifies to 0), almost always a bug.
	assert.True(t, types.CoercionMismatch(types.Undef, types.Num),
		"undef in a Num position is a coercion mismatch — it numifies to 0")

	// A float where an integer is wanted: runs, truncates.
	assert.True(t, types.CoercionMismatch(types.Num, types.Int),
		"Num in an Int position is a coercion mismatch — truncation loses the fraction")

	// NaN where a number is wanted: stringifies fine, arithmetic is nonsense.
	assert.True(t, types.CoercionMismatch(types.NaN, types.Num),
		"NaN in a Num position is a coercion mismatch")

	// A reference in a string position: runs, yields "HASH(0x...)".
	assert.True(t, types.CoercionMismatch(types.HashRef, types.Str),
		"HashRef in a Str position is a coercion mismatch — yields an address")

	// Not a mismatch when the value already IS a member: no conversion.
	assert.False(t, types.CoercionMismatch(types.Int, types.Num),
		"Int in a Num position is not a mismatch — Int <: Num, nothing is lost")
	assert.False(t, types.CoercionMismatch(types.Int, types.Str),
		"Int in a Str position is not a mismatch — Int <: Str")

	// Not a mismatch when nothing can happen at all: that is a harder error.
	assert.False(t, types.CoercionMismatch(types.Str, types.HashRef),
		"Str where a HashRef is wanted is not a COERCION mismatch — Perl cannot fabricate a reference")

	// Unknown is epistemic and never reported.
	assert.False(t, types.CoercionMismatch(types.Unknown, types.Num),
		"Unknown never produces a diagnostic — inference has not determined a type")
}

// --- Strict mode: Unknown must not silently satisfy ---
//
// PSC's Unknown has behaved as TypeScript's `any`: it satisfies every
// requirement, so an un-inferred value passes every check and the checker goes
// quiet exactly where it knows least. TypeScript separates these — `any`
// disables checking, `unknown` must be narrowed before use — and Unknown is
// the second thing, not the first.
//
// TypeSatisfiesStrict is the `unknown` reading: an un-inferred value satisfies
// nothing until inference or a guard establishes what it is.

func TestStrictUnknownSatisfiesNothing(t *testing.T) {
	for _, required := range []types.Type{
		types.Int, types.Num, types.Str, types.Bool, types.Scalar,
		types.HashRef, types.ArrayRef, types.Object, types.List,
	} {
		assert.False(t, types.TypeSatisfiesStrict(types.Unknown, required),
			"strict: Unknown must NOT satisfy %s — narrow it first", required)
		assert.True(t, types.TypeSatisfies(types.Unknown, required),
			"permissive: Unknown satisfies %s (the default, unchanged)", required)
	}
}

// Any is the escape hatch and keeps `any` semantics in BOTH modes: a required
// type of Any means the position accepts anything by construction.

func TestStrictAnyStillAccepts(t *testing.T) {
	assert.True(t, types.TypeSatisfiesStrict(types.Unknown, types.Any),
		"strict: a required type of Any accepts even Unknown — that is what Any means")
	assert.True(t, types.TypeSatisfiesStrict(types.Str, types.Any),
		"strict: Any accepts Str")
}

// Strict mode changes ONLY the Unknown case. Every other judgement is
// identical, so turning it on cannot alter an existing verdict about a value
// whose type IS known.

func TestStrictDiffersOnlyOnUnknown(t *testing.T) {
	known := []types.Type{
		types.Undef, types.Bool, types.Int, types.Num, types.Str,
		types.NaN, types.Inf, types.DualVar, types.Regex,
		types.ScalarRef, types.ArrayRef, types.HashRef, types.CodeRef,
		types.GlobRef, types.Object, types.Ref, types.Scalar,
		types.Array, types.Hash, types.List, types.Code, types.Glob,
	}
	for _, a := range known {
		for _, r := range known {
			assert.Equal(t, types.TypeSatisfies(a, r), types.TypeSatisfiesStrict(a, r),
				"strict and permissive must agree on %s vs %s — only Unknown differs", a, r)
		}
	}
}

// None is the bottom type and remains a subtype of everything in both modes:
// an unreachable branch is not an un-inferred value.

func TestStrictNoneUnaffected(t *testing.T) {
	assert.True(t, types.TypeSatisfiesStrict(types.None, types.Int),
		"strict: None satisfies Int — bottom is a subtype of everything")
}
