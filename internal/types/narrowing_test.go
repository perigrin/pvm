// ABOUTME: Tests for context-based and guard-based type narrowing in the PSC type inference engine.
// ABOUTME: Covers Context enum strings, NarrowByContext rules, GuardPattern structs, and NarrowByGuard rules.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/types"
)

// TestContextString verifies that all Context values produce the correct string names.
func TestContextString(t *testing.T) {
	cases := []struct {
		ctx      types.Context
		expected string
	}{
		{types.UnknownCtx, "Unknown"},
		{types.ScalarCtx, "Scalar"},
		{types.ListCtx, "List"},
		{types.VoidCtx, "Void"},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.ctx.String())
		})
	}
}

// TestNarrowByContextScalar verifies that scalar context narrows aggregate types to counts
// and passes scalar types through unchanged.
func TestNarrowByContextScalar(t *testing.T) {
	// Array in scalar context yields its count (Int)
	narrowed, valid := types.NarrowByContext(types.Array, types.ScalarCtx)
	assert.True(t, valid, "Array in scalar context is valid")
	assert.Equal(t, types.Int, narrowed, "Array in scalar context narrows to Int (count)")

	// Hash in scalar context yields its count (Int)
	narrowed, valid = types.NarrowByContext(types.Hash, types.ScalarCtx)
	assert.True(t, valid, "Hash in scalar context is valid")
	assert.Equal(t, types.Int, narrowed, "Hash in scalar context narrows to Int (count)")

	// List (Array|Hash) in scalar context: both aggregate bits become Int
	narrowed, valid = types.NarrowByContext(types.List, types.ScalarCtx)
	assert.True(t, valid, "List in scalar context is valid")
	assert.Equal(t, types.Int, narrowed, "List in scalar context narrows to Int (both Array and Hash become count)")

	// Str in scalar context passes through unchanged
	narrowed, valid = types.NarrowByContext(types.Str, types.ScalarCtx)
	assert.True(t, valid, "Str in scalar context is valid")
	assert.Equal(t, types.Str, narrowed, "Str in scalar context is unchanged")

	// Int in scalar context passes through unchanged
	narrowed, valid = types.NarrowByContext(types.Int, types.ScalarCtx)
	assert.True(t, valid, "Int in scalar context is valid")
	assert.Equal(t, types.Int, narrowed, "Int in scalar context is unchanged")
}

// TestNarrowByContextList verifies that list context passes all types through unchanged.
func TestNarrowByContextList(t *testing.T) {
	typesToCheck := []types.Type{
		types.Unknown,
		types.Any,
		types.Scalar,
		types.Str,
		types.Int,
		types.Num,
		types.List,
		types.Array,
		types.Hash,
		types.Ref,
		types.HashRef,
		types.Object,
	}

	for _, typ := range typesToCheck {
		t.Run(typ.String(), func(t *testing.T) {
			narrowed, valid := types.NarrowByContext(typ, types.ListCtx)
			assert.True(t, valid, "%s in list context is valid", typ)
			assert.Equal(t, typ, narrowed, "%s in list context is unchanged", typ)
		})
	}
}

// TestNarrowByContextVoid verifies that void context discards the type (returns Unknown, false).
func TestNarrowByContextVoid(t *testing.T) {
	typesToCheck := []types.Type{
		types.Unknown,
		types.Any,
		types.Scalar,
		types.Str,
		types.Int,
		types.List,
		types.Array,
		types.Hash,
	}

	for _, typ := range typesToCheck {
		t.Run(typ.String(), func(t *testing.T) {
			narrowed, valid := types.NarrowByContext(typ, types.VoidCtx)
			assert.False(t, valid, "%s in void context is discarded (valid=false)", typ)
			assert.Equal(t, types.Unknown, narrowed, "%s in void context returns Unknown", typ)
		})
	}
}

// TestNarrowByContextUnknown verifies that unknown context passes all types through unchanged.
func TestNarrowByContextUnknown(t *testing.T) {
	typesToCheck := []types.Type{
		types.Unknown,
		types.Any,
		types.Scalar,
		types.Str,
		types.Int,
		types.List,
		types.Array,
		types.Hash,
	}

	for _, typ := range typesToCheck {
		t.Run(typ.String(), func(t *testing.T) {
			narrowed, valid := types.NarrowByContext(typ, types.UnknownCtx)
			assert.True(t, valid, "%s in unknown context is valid", typ)
			assert.Equal(t, typ, narrowed, "%s in unknown context is unchanged", typ)
		})
	}
}

// TestNarrowByGuardDefined verifies that a defined() guard removes the Undef bit
// from the type using bitset operations. Types with no Undef bit are unchanged.
func TestNarrowByGuardDefined(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardDefined}

	// Scalar &^ Undef: removes the Undef bit, narrowing occurs
	narrowed, narrowed_ok := types.NarrowByGuard(types.Scalar, guard)
	assert.True(t, narrowed_ok, "Scalar narrows under defined guard (has Undef bit)")
	assert.Equal(t, types.Scalar&^types.Undef, narrowed, "Scalar under defined guard removes Undef bit")

	// Undef &^ Undef = 0 → (None, true): the defined-branch of an Undef value is unreachable
	narrowed, narrowed_ok = types.NarrowByGuard(types.Undef, guard)
	assert.True(t, narrowed_ok, "Undef narrowing under defined guard is significant (branch unreachable)")
	assert.Equal(t, types.None, narrowed, "Undef under defined guard is unreachable (None)")

	// Any &^ Undef: removes only the Undef bit from all concrete types
	narrowed, narrowed_ok = types.NarrowByGuard(types.Any, guard)
	assert.True(t, narrowed_ok, "Any narrows under defined guard (has Undef bit)")
	assert.Equal(t, types.Any&^types.Undef, narrowed, "Any under defined guard removes Undef bit")

	// Int &^ Undef = Int: Int has no Undef bit, passes through unchanged
	narrowed, narrowed_ok = types.NarrowByGuard(types.Int, guard)
	assert.False(t, narrowed_ok, "Int does not narrow under defined guard (no Undef bit)")
	assert.Equal(t, types.Int, narrowed, "Int under defined guard is unchanged")

	// Str &^ Undef = Str: Str has no Undef bit, passes through unchanged
	narrowed, narrowed_ok = types.NarrowByGuard(types.Str, guard)
	assert.False(t, narrowed_ok, "Str does not narrow under defined guard (no Undef bit)")
	assert.Equal(t, types.Str, narrowed, "Str under defined guard is unchanged")

	// Unknown is treated as Any for narrowing: Any &^ Undef
	narrowed, narrowed_ok = types.NarrowByGuard(types.Unknown, guard)
	assert.True(t, narrowed_ok, "Unknown treated as Any under defined guard")
	assert.Equal(t, types.Any&^types.Undef, narrowed, "Unknown under defined guard removes Undef bit from Any")
}

// TestNarrowByGuardRef verifies that a plain ref() guard intersects with the Ref mask.
// Scalar & Ref = Ref (narrows); Ref & Ref = Ref = typ (no change); Unknown treated as Any.
func TestNarrowByGuardRef(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef}

	// Scalar & Ref = Ref ≠ Scalar → (Ref, true)
	narrowed, narrowed_ok := types.NarrowByGuard(types.Scalar, guard)
	assert.True(t, narrowed_ok, "Scalar narrows under ref guard")
	assert.Equal(t, types.Ref, narrowed, "Scalar & Ref = Ref")

	// Any & Ref = Ref ≠ Any → (Ref, true)
	narrowed, narrowed_ok = types.NarrowByGuard(types.Any, guard)
	assert.True(t, narrowed_ok, "Any narrows under ref guard")
	assert.Equal(t, types.Ref, narrowed, "Any & Ref = Ref")

	// Unknown treated as Any: Any & Ref = Ref → (Ref, true)
	narrowed, narrowed_ok = types.NarrowByGuard(types.Unknown, guard)
	assert.True(t, narrowed_ok, "Unknown (treated as Any) narrows under ref guard")
	assert.Equal(t, types.Ref, narrowed, "Unknown & Ref = Ref")
}

// TestNarrowByGuardIsa verifies that an isa guard narrows any type to Object.
func TestNarrowByGuardIsa(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardIsa}

	narrowed, narrowed_ok := types.NarrowByGuard(types.Scalar, guard)
	assert.True(t, narrowed_ok, "Scalar narrows under isa guard")
	assert.Equal(t, types.Object, narrowed, "Scalar under isa guard narrows to Object")

	narrowed, narrowed_ok = types.NarrowByGuard(types.Any, guard)
	assert.True(t, narrowed_ok, "Any narrows under isa guard")
	assert.Equal(t, types.Object, narrowed, "Any under isa guard narrows to Object")

	narrowed, narrowed_ok = types.NarrowByGuard(types.Ref, guard)
	assert.True(t, narrowed_ok, "Ref narrows under isa guard")
	assert.Equal(t, types.Object, narrowed, "Ref under isa guard narrows to Object")
}

// TestNegateGuardDefined verifies that negating a defined() guard keeps only the Undef bit
// (typ & Undef). Types with no Undef bit produce (None, true) — the else-branch is unreachable.
func TestNegateGuardDefined(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardDefined}

	// Scalar & Undef = Undef ≠ Scalar → (Undef, true)
	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar negates under defined guard (Undef bit kept)")
	assert.Equal(t, types.Undef, narrowed)

	// Any & Undef = Undef ≠ Any → (Undef, true)
	narrowed, ok = types.NegateGuard(types.Any, guard)
	assert.True(t, ok, "Any negates under defined guard (Undef bit kept)")
	assert.Equal(t, types.Undef, narrowed)

	// Undef & Undef = Undef = typ → (Undef, false): no change (already only Undef)
	narrowed, ok = types.NegateGuard(types.Undef, guard)
	assert.False(t, ok, "Undef negated defined: result equals input, no useful narrowing")
	assert.Equal(t, types.Undef, narrowed)

	// Int & Undef = 0 → (None, true): Int can never be Undef, else-branch unreachable
	narrowed, ok = types.NegateGuard(types.Int, guard)
	assert.True(t, ok, "Int negated defined: narrowing occurred (unreachable branch)")
	assert.Equal(t, types.None, narrowed)
}

// TestNegateGuardRef verifies that negating ref() removes all Ref bits (typ &^ Ref).
// Types with no Ref bits are unchanged.
func TestNegateGuardRef(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef}

	// Scalar &^ Ref removes the Ref sub-mask from Scalar (keeps non-ref scalar bits)
	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar negates under ref guard (Ref bits removed)")
	assert.Equal(t, types.Scalar&^types.Ref, narrowed)

	// Any &^ Ref removes Ref bits from Any
	narrowed, ok = types.NegateGuard(types.Any, guard)
	assert.True(t, ok, "Any negates under ref guard (Ref bits removed)")
	assert.Equal(t, types.Any&^types.Ref, narrowed)

	// Undef &^ Ref = Undef = typ → (Undef, false): Undef has no Ref bits
	narrowed, ok = types.NegateGuard(types.Undef, guard)
	assert.False(t, ok, "Undef negated ref: no change (no Ref bits to remove)")
	assert.Equal(t, types.Undef, narrowed)
}

func TestNegateGuardRefWithType(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef, RefType: "HASH"}

	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.False(t, ok, "ref eq TYPE negation is not useful")
	assert.Equal(t, types.Scalar, narrowed)
}

func TestNegateGuardIsa(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardIsa}

	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.False(t, ok, "isa negation is not useful")
	assert.Equal(t, types.Scalar, narrowed)
}

// TestNarrowByGuardRefPreservesSubtypes verifies that a plain ref() guard does
// not widen types that are already subtypes of Ref.
func TestNarrowByGuardRefPreservesSubtypes(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef}

	// HashRef is a subtype of Ref — ref($x) should not widen it to Ref.
	narrowed, ok := types.NarrowByGuard(types.HashRef, guard)
	assert.False(t, ok, "HashRef should not narrow under plain ref guard (already a Ref subtype)")
	assert.Equal(t, types.HashRef, narrowed, "HashRef should be preserved, not widened to Ref")

	// ArrayRef is a subtype of Ref.
	narrowed, ok = types.NarrowByGuard(types.ArrayRef, guard)
	assert.False(t, ok, "ArrayRef should not narrow under plain ref guard")
	assert.Equal(t, types.ArrayRef, narrowed, "ArrayRef should be preserved")

	// CodeRef is a subtype of Ref.
	narrowed, ok = types.NarrowByGuard(types.CodeRef, guard)
	assert.False(t, ok, "CodeRef should not narrow under plain ref guard")
	assert.Equal(t, types.CodeRef, narrowed, "CodeRef should be preserved")

	// Object is a subtype of Ref.
	narrowed, ok = types.NarrowByGuard(types.Object, guard)
	assert.False(t, ok, "Object should not narrow under plain ref guard")
	assert.Equal(t, types.Object, narrowed, "Object should be preserved")

	// Ref & Ref = Ref = typ → (Ref, false): no change because result equals input.
	narrowed, ok = types.NarrowByGuard(types.Ref, guard)
	assert.False(t, ok, "Ref under ref guard: result equals input, no narrowing")
	assert.Equal(t, types.Ref, narrowed, "Ref under ref guard is unchanged (already Ref)")
}

// TestNarrowByGuardBitsetDefined verifies that defined() guard uses bit operations:
// removing the Undef bit from the type, returning None when the result is empty.
func TestNarrowByGuardBitsetDefined(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardDefined}

	// Scalar &^ Undef = all scalar bits minus Undef bit
	narrowed, ok := types.NarrowByGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar narrows under defined guard (has Undef bit)")
	assert.Equal(t, types.Scalar&^types.Undef, narrowed, "Scalar under defined guard removes Undef bit")

	// Undef &^ Undef = 0 → empty set → (None, true) unreachable branch
	narrowed, ok = types.NarrowByGuard(types.Undef, guard)
	assert.True(t, ok, "Undef narrowing is significant (branch unreachable)")
	assert.Equal(t, types.None, narrowed, "Undef under defined guard is unreachable (None)")

	// Any &^ Undef = all concrete bits minus Undef
	narrowed, ok = types.NarrowByGuard(types.Any, guard)
	assert.True(t, ok, "Any narrows under defined guard")
	assert.Equal(t, types.Any&^types.Undef, narrowed, "Any under defined guard removes Undef bit")

	// Int has no Undef bit, so result == typ → no narrowing
	narrowed, ok = types.NarrowByGuard(types.Int, guard)
	assert.False(t, ok, "Int has no Undef bit, no narrowing")
	assert.Equal(t, types.Int, narrowed, "Int under defined guard is unchanged")

	// Unknown treated as Any → Any &^ Undef
	narrowed, ok = types.NarrowByGuard(types.Unknown, guard)
	assert.True(t, ok, "Unknown treated as Any under defined guard")
	assert.Equal(t, types.Any&^types.Undef, narrowed, "Unknown under defined guard removes Undef bit from Any")
}

// TestNegateGuardBitsetDefined verifies that negating defined() keeps only the Undef bit,
// and returns None for types that have no Undef bit (no useful narrowing — unreachable).
func TestNegateGuardBitsetDefined(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardDefined}

	// Scalar & Undef = Undef → (Undef, true)
	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar negates under defined guard (has Undef bit)")
	assert.Equal(t, types.Undef, narrowed, "Scalar negated defined = Undef (only Undef bit)")

	// Undef & Undef = Undef = typ → (Undef, false) — no change
	narrowed, ok = types.NegateGuard(types.Undef, guard)
	assert.False(t, ok, "Undef negated defined: result equals input, no narrowing")
	assert.Equal(t, types.Undef, narrowed, "Undef negated defined = Undef (unchanged)")

	// Int & Undef = 0 → empty set → (None, true) — Int can never be Undef, branch unreachable
	narrowed, ok = types.NegateGuard(types.Int, guard)
	assert.True(t, ok, "Int negated defined: narrowing occurred (None — unreachable)")
	assert.Equal(t, types.None, narrowed, "Int negated defined is unreachable (None)")
}

// TestNegateGuardBitsetRef verifies that negating ref() removes all Ref bits.
func TestNegateGuardBitsetRef(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef}

	// Scalar &^ Ref removes all reference bits from Scalar
	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar negated ref: narrowing occurred (Ref bits removed)")
	assert.Equal(t, types.Scalar&^types.Ref, narrowed, "Scalar negated ref removes Ref bits")

	// Any &^ Ref removes all reference bits from Any
	narrowed, ok = types.NegateGuard(types.Any, guard)
	assert.True(t, ok, "Any negated ref: narrowing occurred")
	assert.Equal(t, types.Any&^types.Ref, narrowed, "Any negated ref removes Ref bits")

	// Undef &^ Ref = Undef = typ → no change
	narrowed, ok = types.NegateGuard(types.Undef, guard)
	assert.False(t, ok, "Undef negated ref: no change (Undef has no Ref bits)")
	assert.Equal(t, types.Undef, narrowed, "Undef negated ref is unchanged")

	// Ref &^ Ref = 0 → empty set → (None, true) — Ref can never pass negated ref guard
	narrowed, ok = types.NegateGuard(types.Ref, guard)
	assert.True(t, ok, "Ref negated ref: narrowing occurred (None — unreachable)")
	assert.Equal(t, types.None, narrowed, "Ref negated ref is unreachable (None)")
}

// TestNarrowByGuardBitsetRef verifies that a plain ref() guard uses bit intersection:
// typ & Ref. Ref itself returns (Ref, false) because result == typ.
func TestNarrowByGuardBitsetRef(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef}

	// Ref & Ref = Ref = typ → (Ref, false) — no change, already the right type
	narrowed, ok := types.NarrowByGuard(types.Ref, guard)
	assert.False(t, ok, "Ref under ref guard: result equals input, no narrowing")
	assert.Equal(t, types.Ref, narrowed, "Ref under ref guard is unchanged (already Ref)")

	// Scalar & Ref = Ref → (Ref, true)
	narrowed, ok = types.NarrowByGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar narrows under ref guard")
	assert.Equal(t, types.Ref, narrowed, "Scalar & Ref = Ref")

	// Int & Ref = 0 → empty set → (None, true) — Int cannot be a ref
	narrowed, ok = types.NarrowByGuard(types.Int, guard)
	assert.True(t, ok, "Int under ref guard: narrowing occurred (None — unreachable)")
	assert.Equal(t, types.None, narrowed, "Int under ref guard is unreachable (None)")
}

// TestNarrowByContextBitsetUnion verifies that NarrowByContext handles union types correctly.
// When a union contains Array or Hash bits in scalar context, those bits become Int.
func TestNarrowByContextBitsetUnion(t *testing.T) {
	// Array|Str in scalar context: Array→Int, Str→Str, result = Int|Str
	union := types.Array | types.Str
	narrowed, ok := types.NarrowByContext(union, types.ScalarCtx)
	assert.True(t, ok, "Array|Str in scalar context is valid")
	assert.Equal(t, types.Int|types.Str, narrowed, "Array|Str in scalar context narrows Array bit to Int")

	// Array|Hash in scalar context: both become Int, result = Int
	union2 := types.Array | types.Hash
	narrowed2, ok2 := types.NarrowByContext(union2, types.ScalarCtx)
	assert.True(t, ok2, "Array|Hash in scalar context is valid")
	assert.Equal(t, types.Int, narrowed2, "Array|Hash in scalar context both become Int")

	// Scalar in scalar context: passes through unchanged (Scalar has no Array/Hash bits directly)
	narrowed3, ok3 := types.NarrowByContext(types.Scalar, types.ScalarCtx)
	assert.True(t, ok3, "Scalar in scalar context is valid")
	assert.Equal(t, types.Scalar, narrowed3, "Scalar in scalar context is unchanged")
}

// TestNarrowByGuardRefEq verifies that a ref() eq 'TYPE' guard narrows to the
// appropriate specific reference type.
func TestNarrowByGuardRefEq(t *testing.T) {
	cases := []struct {
		refType  string
		expected types.Type
	}{
		{"HASH", types.HashRef},
		{"ARRAY", types.ArrayRef},
		{"SCALAR", types.ScalarRef},
		{"CODE", types.CodeRef},
		{"GLOB", types.GlobRef},
		{"REF", types.Ref},
		{"Foo", types.Object},     // unknown ref type string → blessed reference
		{"MyClass", types.Object}, // arbitrary class name → Object
	}

	for _, tc := range cases {
		t.Run(tc.refType, func(t *testing.T) {
			guard := types.GuardPattern{Kind: types.GuardRef, RefType: tc.refType}
			narrowed, narrowed_ok := types.NarrowByGuard(types.Scalar, guard)
			assert.True(t, narrowed_ok, "Scalar narrows under ref eq %q guard", tc.refType)
			assert.Equal(t, tc.expected, narrowed, "Scalar under ref eq %q narrows to %s", tc.refType, tc.expected)
		})
	}
}
