// ABOUTME: Tests validating the type system can represent Perl type gotchas.
// ABOUTME: Covers operator confusion, undef propagation, DualVar, NaN/Inf, refs, and booleans.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/types"
)

// --- Operator type confusion ---
// Perl's == forces numeric context on strings: "foo" == "bar" is true (both become 0).
// Perl's eq forces string context on numbers: 1.0 eq "1" may be false.
// The type system must distinguish Str from Num to detect these.

func TestGotchaOperatorTypeConfusion(t *testing.T) {
	// Str is NOT a subtype of Num: a Str value in numeric context is a type mismatch.
	// Covers: "hello" + 1 silently producing 1, "foo" == "bar" being true.
	assert.False(t, types.IsSubtype(types.Str, types.Num),
		"Str is NOT a subtype of Num — string in numeric context is a mismatch")

	// Num is NOT a subtype of Int: a float where an integer is required is a mismatch.
	assert.False(t, types.IsSubtype(types.Num, types.Int),
		"Num is NOT a subtype of Int — 3.14 is not an integer")

	// Int IS a subtype of Str: integers have lossless string representations.
	// This means using eq on an Int is not a type error (but may be surprising).
	assert.True(t, types.IsSubtype(types.Int, types.Str),
		"Int IS a subtype of Str — integers stringify losslessly")
}

// --- Undef propagation ---
// Perl's undef silently becomes 0 in numeric context and "" in string context.
// The type system must keep Undef distinct from Int, Num, and Str to detect this.

func TestGotchaUndefPropagation(t *testing.T) {
	// Undef is NOT a subtype of Int, Num, or Str
	assert.False(t, types.IsSubtype(types.Undef, types.Int),
		"Undef is NOT Int — undef + 1 silently producing 1 is a type error")
	assert.False(t, types.IsSubtype(types.Undef, types.Num),
		"Undef is NOT Num — undef in arithmetic is a type error")
	assert.False(t, types.IsSubtype(types.Undef, types.Str),
		"Undef is NOT Str — undef in string context is a type error")

	// Undef IS a subtype of Scalar: it can appear in scalar variables
	assert.True(t, types.IsSubtype(types.Undef, types.Scalar),
		"Undef IS Scalar — undef is a valid scalar value")

	// Undef vs None: they are distinct concepts
	assert.NotEqual(t, types.Undef, types.None,
		"Undef is a concrete value; None is the empty/bottom type")
}

// --- Context narrowing for refs in scalar context ---
// References in scalar context pass through unchanged (they don't become a count).
// This is unlike Array/Hash which narrow to Int in scalar context.

func TestGotchaRefInScalarContext(t *testing.T) {
	// Ref types in scalar context should pass through unchanged
	narrowed, valid := types.NarrowByContext(types.Ref, types.ScalarCtx)
	assert.True(t, valid, "Ref in scalar context is valid")
	assert.Equal(t, types.Ref, narrowed,
		"Ref in scalar context passes through unchanged (not a count)")

	narrowed, valid = types.NarrowByContext(types.HashRef, types.ScalarCtx)
	assert.True(t, valid, "HashRef in scalar context is valid")
	assert.Equal(t, types.HashRef, narrowed,
		"HashRef in scalar context passes through unchanged")

	narrowed, valid = types.NarrowByContext(types.Object, types.ScalarCtx)
	assert.True(t, valid, "Object in scalar context is valid")
	assert.Equal(t, types.Object, narrowed,
		"Object in scalar context passes through unchanged")
}

// --- DualVar semantics ---
// DualVar represents values where string and numeric interpretations diverge.
// The canonical example is $! (errno): numerically 2, stringifies to
// "No such file or directory". DualVar sits in Scalar but outside Str/Num.

func TestGotchaDualVarSemantics(t *testing.T) {
	// DualVar is Scalar but NOT Str or Num
	assert.True(t, types.IsSubtype(types.DualVar, types.Scalar),
		"DualVar is Scalar — $! is a valid scalar")
	assert.False(t, types.IsSubtype(types.DualVar, types.Str),
		"DualVar is NOT Str — string and numeric representations diverge")
	assert.False(t, types.IsSubtype(types.DualVar, types.Num),
		"DualVar is NOT Num — string and numeric representations diverge")

	// DualVar does not satisfy Int or Num requirements
	assert.False(t, types.TypeSatisfies(types.DualVar, types.Int),
		"DualVar does not satisfy Int — cannot rely on numeric value alone")
	assert.False(t, types.TypeSatisfies(types.DualVar, types.Num),
		"DualVar does not satisfy Num")
}

// --- NaN and Inf semantics ---
// NaN and Inf are IEEE 754 special values that sit in Scalar but outside
// both Str and Num. The type system must distinguish them for diagnostics.

func TestGotchaNaNSemantics(t *testing.T) {
	// NaN is Scalar but not Num or Str
	assert.True(t, types.IsSubtype(types.NaN, types.Scalar),
		"NaN is Scalar")
	assert.False(t, types.IsSubtype(types.NaN, types.Num),
		"NaN is NOT Num — NaN != NaN violates reflexivity")
	assert.False(t, types.IsSubtype(types.NaN, types.Str),
		"NaN is NOT Str — 'NaN' is a representational artifact")

	// NaN does not satisfy Num or Int requirements
	assert.False(t, types.TypeSatisfies(types.NaN, types.Num),
		"NaN does not satisfy Num — arithmetic on NaN is undefined")
	assert.False(t, types.TypeSatisfies(types.NaN, types.Int),
		"NaN does not satisfy Int")
}

func TestGotchaInfSemantics(t *testing.T) {
	// Inf is Scalar but not Num or Str
	assert.True(t, types.IsSubtype(types.Inf, types.Scalar),
		"Inf is Scalar")
	assert.False(t, types.IsSubtype(types.Inf, types.Num),
		"Inf is NOT Num — Inf - Inf = NaN violates subtraction identity")
	assert.False(t, types.IsSubtype(types.Inf, types.Str),
		"Inf is NOT Str — 'Inf' is a representational artifact")

	// Inf does not satisfy Num or Int requirements
	assert.False(t, types.TypeSatisfies(types.Inf, types.Num),
		"Inf does not satisfy Num")
	assert.False(t, types.TypeSatisfies(types.Inf, types.Int),
		"Inf does not satisfy Int")
}

// --- Reference type safety ---
// References are NOT subtypes of Str. Stringifying a reference produces
// "ARRAY(0x...)" which is almost never intended.

func TestGotchaRefTypeStringification(t *testing.T) {
	refTypes := []types.Type{
		types.ScalarRef, types.ArrayRef, types.HashRef,
		types.CodeRef, types.GlobRef, types.Object, types.Ref,
	}

	for _, rt := range refTypes {
		assert.False(t, types.IsSubtype(rt, types.Str),
			"%s is NOT a subtype of Str — stringification produces 'TYPE(0x...)'", rt)
	}
}

// --- Boolean edge cases ---
// Bool is a primitive type (builtin::true/builtin::false via is_bool()).
// It is NOT the same as numeric 0/1 or truthy/falsy.

func TestGotchaBoolPrimitive(t *testing.T) {
	// Bool is NOT a subtype of Int
	assert.False(t, types.IsSubtype(types.Bool, types.Int),
		"Bool is NOT Int — is_bool() primitives are distinct from 0/1")

	// Bool is NOT a subtype of Num
	assert.False(t, types.IsSubtype(types.Bool, types.Num),
		"Bool is NOT Num — primitive booleans are not numbers")

	// Bool IS a subtype of Scalar
	assert.True(t, types.IsSubtype(types.Bool, types.Scalar),
		"Bool IS Scalar — booleans are valid scalar values")

	// GuardBool narrows Int to None (an Int cannot be a primitive boolean)
	guard := types.GuardPattern{Kind: types.GuardBool}
	narrowed, ok := types.NarrowByGuard(types.Int, guard)
	assert.True(t, ok, "Int narrows under bool guard")
	assert.Equal(t, types.None, narrowed,
		"Int under bool guard is unreachable (None) — Int has no Bool bit")
}

// --- Undef vs None distinction ---
// Undef is a concrete value that exists. None is the bottom type (empty set).

func TestGotchaUndefVsNone(t *testing.T) {
	// They are distinct
	assert.NotEqual(t, types.Undef, types.None)

	// Undef is a subtype of Scalar
	assert.True(t, types.IsSubtype(types.Undef, types.Scalar))

	// None is a subtype of everything (bottom type)
	assert.True(t, types.IsSubtype(types.None, types.Scalar))
	assert.True(t, types.IsSubtype(types.None, types.Int))
	assert.True(t, types.IsSubtype(types.None, types.Undef))

	// Undef is NOT a subtype of None
	assert.False(t, types.IsSubtype(types.Undef, types.None),
		"Undef is NOT None — undef is a concrete value, not an empty set")
}

// --- Sort gotcha: numeric array with string comparison ---
// sort without a comparison function uses string comparison.
// The type system distinguishes Array from Str to allow detecting this.

func TestGotchaSortContext(t *testing.T) {
	// Array is NOT a subtype of Str — arrays cannot be treated as strings
	assert.False(t, types.IsSubtype(types.Array, types.Str),
		"Array is NOT Str — sort's default string comparison on numeric arrays is a gotcha")

	// Array is NOT a subtype of Scalar
	assert.False(t, types.IsSubtype(types.Array, types.Scalar),
		"Array is NOT Scalar — passing an array where a scalar is expected is a type error")
}
