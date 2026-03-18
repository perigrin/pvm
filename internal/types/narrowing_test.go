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

	// List in scalar context narrows to Scalar
	narrowed, valid = types.NarrowByContext(types.List, types.ScalarCtx)
	assert.True(t, valid, "List in scalar context is valid")
	assert.Equal(t, types.Scalar, narrowed, "List in scalar context narrows to Scalar")

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

// TestNarrowByGuardDefined verifies that a defined() guard narrows possibly-undef
// types to Scalar, while types that cannot be undef are left unchanged.
func TestNarrowByGuardDefined(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardDefined}

	// Scalar could be Undef, so it narrows to Scalar (non-undef Scalar)
	narrowed, narrowed_ok := types.NarrowByGuard(types.Scalar, guard)
	assert.True(t, narrowed_ok, "Scalar narrows under defined guard")
	assert.Equal(t, types.Scalar, narrowed, "Scalar under defined guard stays Scalar (known non-undef)")

	// Undef narrows to Scalar (it becomes defined — i.e., the else branch)
	narrowed, narrowed_ok = types.NarrowByGuard(types.Undef, guard)
	assert.True(t, narrowed_ok, "Undef narrows under defined guard")
	assert.Equal(t, types.Scalar, narrowed, "Undef under defined guard narrows to Scalar")

	// Any could be Undef, so it narrows to Scalar
	narrowed, narrowed_ok = types.NarrowByGuard(types.Any, guard)
	assert.True(t, narrowed_ok, "Any narrows under defined guard")
	assert.Equal(t, types.Scalar, narrowed, "Any under defined guard narrows to Scalar")

	// Int cannot be Undef (it is already a concrete non-undef type), so it passes through unchanged
	narrowed, narrowed_ok = types.NarrowByGuard(types.Int, guard)
	assert.False(t, narrowed_ok, "Int does not narrow under defined guard (already non-undef)")
	assert.Equal(t, types.Int, narrowed, "Int under defined guard is unchanged")

	// Str cannot be Undef, so it passes through unchanged
	narrowed, narrowed_ok = types.NarrowByGuard(types.Str, guard)
	assert.False(t, narrowed_ok, "Str does not narrow under defined guard (already non-undef)")
	assert.Equal(t, types.Str, narrowed, "Str under defined guard is unchanged")

	// Unknown is a top type (type not yet determined) — defined() narrows it to Scalar.
	// This is consistent with GuardRef which narrows Unknown to Ref.
	narrowed, narrowed_ok = types.NarrowByGuard(types.Unknown, guard)
	assert.True(t, narrowed_ok, "Unknown narrows under defined guard")
	assert.Equal(t, types.Scalar, narrowed, "Unknown under defined guard narrows to Scalar")
}

// TestNarrowByGuardRef verifies that a plain ref() guard narrows any type to Ref.
func TestNarrowByGuardRef(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef}

	narrowed, narrowed_ok := types.NarrowByGuard(types.Scalar, guard)
	assert.True(t, narrowed_ok, "Scalar narrows under ref guard")
	assert.Equal(t, types.Ref, narrowed, "Scalar under ref guard narrows to Ref")

	narrowed, narrowed_ok = types.NarrowByGuard(types.Any, guard)
	assert.True(t, narrowed_ok, "Any narrows under ref guard")
	assert.Equal(t, types.Ref, narrowed, "Any under ref guard narrows to Ref")

	narrowed, narrowed_ok = types.NarrowByGuard(types.Unknown, guard)
	assert.True(t, narrowed_ok, "Unknown narrows under ref guard")
	assert.Equal(t, types.Ref, narrowed, "Unknown under ref guard narrows to Ref")
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

// TestNegateGuardDefined verifies that negating a defined() guard narrows to Undef.
func TestNegateGuardDefined(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardDefined}

	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar negates under defined guard")
	assert.Equal(t, types.Undef, narrowed)

	narrowed, ok = types.NegateGuard(types.Any, guard)
	assert.True(t, ok, "Any negates under defined guard")
	assert.Equal(t, types.Undef, narrowed)

	narrowed, ok = types.NegateGuard(types.Undef, guard)
	assert.True(t, ok, "Undef negates under defined guard")
	assert.Equal(t, types.Undef, narrowed)

	narrowed, ok = types.NegateGuard(types.Int, guard)
	assert.False(t, ok, "Int does not negate meaningfully under defined guard")
	assert.Equal(t, types.Int, narrowed)
}

func TestNegateGuardRef(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef}

	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar negates under ref guard")
	assert.Equal(t, types.Scalar, narrowed)

	narrowed, ok = types.NegateGuard(types.Any, guard)
	assert.True(t, ok, "Any negates under ref guard")
	assert.Equal(t, types.Scalar, narrowed)

	narrowed, ok = types.NegateGuard(types.Undef, guard)
	assert.False(t, ok, "Undef does not negate meaningfully under ref guard")
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

	// Ref itself should still narrow (identity — it's exactly Ref, not more specific).
	narrowed, ok = types.NarrowByGuard(types.Ref, guard)
	assert.True(t, ok, "Ref narrows under ref guard (it is Ref, not a subtype)")
	assert.Equal(t, types.Ref, narrowed, "Ref under ref guard stays Ref")
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
