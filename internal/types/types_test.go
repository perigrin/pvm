// ABOUTME: Tests for the PSC type system used in type inference.
// ABOUTME: Covers type enums, string names, subtype hierarchy, and type satisfaction.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/types"
)

func TestTypeString(t *testing.T) {
	cases := []struct {
		typ      types.Type
		expected string
	}{
		{types.Unknown, "Unknown"},
		{types.Any, "Any"},
		{types.Scalar, "Scalar"},
		{types.Undef, "Undef"},
		{types.Bool, "Bool"},
		{types.Str, "Str"},
		{types.Num, "Num"},
		{types.Int, "Int"},
		{types.DualVar, "DualVar"},
		{types.Regex, "Regex"},
		{types.Ref, "Ref"},
		{types.ScalarRef, "ScalarRef"},
		{types.ArrayRef, "ArrayRef"},
		{types.HashRef, "HashRef"},
		{types.CodeRef, "CodeRef"},
		{types.GlobRef, "GlobRef"},
		{types.Object, "Object"},
		{types.List, "List"},
		{types.Array, "Array"},
		{types.Hash, "Hash"},
		{types.Code, "Code"},
		{types.Glob, "Glob"},
		{types.None, "None"},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.typ.String())
		})
	}
}

func TestTypeStringArbitraryUnion(t *testing.T) {
	// Arbitrary union types that have no named mask decompose to "A|B" notation.
	// Type(999) = 0b1111100111 = Undef|Bool|Int|DualVar|Regex|ScalarRef|ArrayRef|HashRef.
	// The zero value with no bits set is Unknown.
	assert.Equal(t, "Unknown", types.Type(0).String())
	// A union of two named leaf types that has no named mask is displayed as A|B.
	assert.Equal(t, "Undef|Bool", types.Type(types.Undef|types.Bool).String())
}

func TestIsSubtypeDirect(t *testing.T) {
	// Direct parent relationships from the hierarchy
	cases := []struct {
		child  types.Type
		parent types.Type
	}{
		{types.Scalar, types.Any},
		{types.Undef, types.Scalar},
		{types.Bool, types.Scalar},
		{types.Str, types.Scalar},
		{types.Num, types.Str},
		{types.Int, types.Num},
		{types.DualVar, types.Scalar},
		{types.Regex, types.Scalar},
		{types.Ref, types.Scalar},
		{types.ScalarRef, types.Ref},
		{types.ArrayRef, types.Ref},
		{types.HashRef, types.Ref},
		{types.CodeRef, types.Ref},
		{types.GlobRef, types.Ref},
		{types.Object, types.Ref},
		{types.List, types.Any},
		{types.Array, types.List},
		{types.Hash, types.List},
		{types.Code, types.Any},
		{types.Glob, types.Any},
	}

	for _, tc := range cases {
		t.Run(tc.child.String()+"->"+tc.parent.String(), func(t *testing.T) {
			assert.True(t, types.IsSubtype(tc.child, tc.parent),
				"%s should be a subtype of %s", tc.child, tc.parent)
		})
	}
}

func TestIsSubtypeTransitive(t *testing.T) {
	// Transitive relationships that skip levels
	assert.True(t, types.IsSubtype(types.Int, types.Any), "Int should be transitive subtype of Any")
	assert.True(t, types.IsSubtype(types.Int, types.Scalar), "Int should be transitive subtype of Scalar")
	assert.True(t, types.IsSubtype(types.Int, types.Str), "Int should be transitive subtype of Str (Num->Str->Scalar)")
	assert.True(t, types.IsSubtype(types.HashRef, types.Scalar), "HashRef should be transitive subtype of Scalar")
	assert.True(t, types.IsSubtype(types.Array, types.Any), "Array should be transitive subtype of Any")
	assert.True(t, types.IsSubtype(types.Object, types.Scalar), "Object should be transitive subtype of Scalar")
	assert.True(t, types.IsSubtype(types.Object, types.Any), "Object should be transitive subtype of Any")
}

func TestIsSubtypeNegative(t *testing.T) {
	// Types that are NOT in a subtype relationship
	assert.False(t, types.IsSubtype(types.Str, types.Int), "Str is NOT subtype of Int")
	assert.False(t, types.IsSubtype(types.Array, types.Scalar), "Array is NOT subtype of Scalar")
	assert.False(t, types.IsSubtype(types.Num, types.Int), "Num is NOT subtype of Int (wrong direction)")
	assert.False(t, types.IsSubtype(types.Scalar, types.Str), "Scalar is NOT subtype of Str")
	assert.False(t, types.IsSubtype(types.Code, types.List), "Code is NOT subtype of List")
	assert.False(t, types.IsSubtype(types.Hash, types.Array), "Hash is NOT subtype of Array (siblings)")
}

func TestIsSubtypeEdges(t *testing.T) {
	// Identity: a type is a subtype of itself
	assert.True(t, types.IsSubtype(types.Int, types.Int), "Int is subtype of itself")
	assert.True(t, types.IsSubtype(types.Any, types.Any), "Any is subtype of itself")
	assert.True(t, types.IsSubtype(types.None, types.None), "None is subtype of itself")

	// None is subtype of everything
	assert.True(t, types.IsSubtype(types.None, types.Any), "None is subtype of Any")
	assert.True(t, types.IsSubtype(types.None, types.Int), "None is subtype of Int")
	assert.True(t, types.IsSubtype(types.None, types.Scalar), "None is subtype of Scalar")

	// Any is NOT subtype of its descendants
	assert.False(t, types.IsSubtype(types.Any, types.Int), "Any is NOT subtype of Int")
	assert.False(t, types.IsSubtype(types.Any, types.Scalar), "Any is NOT subtype of Scalar")
}

func TestTypeSatisfiesBasic(t *testing.T) {
	// Int satisfies Num (subtype relationship)
	assert.True(t, types.TypeSatisfies(types.Int, types.Num), "Int satisfies Num")
	// Int satisfies Str (transitive subtype)
	assert.True(t, types.TypeSatisfies(types.Int, types.Str), "Int satisfies Str (transitive)")
	// Int satisfies Any (required is Any — accepts everything)
	assert.True(t, types.TypeSatisfies(types.Int, types.Any), "Int satisfies Any")
	// Str does NOT satisfy Int
	assert.False(t, types.TypeSatisfies(types.Str, types.Int), "Str does not satisfy Int")
	// Array does NOT satisfy Scalar
	assert.False(t, types.TypeSatisfies(types.Array, types.Scalar), "Array does not satisfy Scalar")
}

func TestTypeSatisfiesPolymorphic(t *testing.T) {
	// Scalar satisfies Str: Scalar is a polymorphic type and Str is a subtype of Scalar,
	// so at runtime the Scalar variable could hold a Str value.
	assert.True(t, types.TypeSatisfies(types.Scalar, types.Str),
		"Scalar satisfies Str (polymorphic: Scalar could hold any scalar subtype)")

	// Any satisfies Int: Any is polymorphic, could hold an Int.
	assert.True(t, types.TypeSatisfies(types.Any, types.Int),
		"Any satisfies Int (polymorphic)")

	// List satisfies Array: List is polymorphic.
	assert.True(t, types.TypeSatisfies(types.List, types.Array),
		"List satisfies Array (polymorphic)")

	// Str does NOT satisfy Int: Str is not polymorphic, and Str is not a subtype of Int.
	assert.False(t, types.TypeSatisfies(types.Str, types.Int),
		"Str does not satisfy Int (Str is not polymorphic w.r.t. Int)")
}

func TestTypeSatisfiesUnknown(t *testing.T) {
	// Zero value (unknown type) passes permissively against any required type
	var unknown types.Type
	assert.True(t, types.TypeSatisfies(unknown, types.Int), "unknown type satisfies Int permissively")
	assert.True(t, types.TypeSatisfies(unknown, types.Str), "unknown type satisfies Str permissively")
	assert.True(t, types.TypeSatisfies(unknown, types.Any), "unknown type satisfies Any permissively")
}
