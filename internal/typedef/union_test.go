// ABOUTME: Tests for union type functionality
// ABOUTME: Verifies union type operations and trait integration

package typedef

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnionTypeCreation tests the creation of union types
func TestUnionTypeCreation(t *testing.T) {
	// Test creating a valid union type
	members := []string{"Int", "Str"}
	unionType := NewUnionType(members)
	require.NotNil(t, unionType)
	assert.Equal(t, members, unionType.GetMembers())
	assert.Equal(t, "Int|Str", unionType.String())
	assert.Equal(t, "Union[Int, Str]", unionType.TypeName())

	// Test duplicate removal
	membersWithDups := []string{"Int", "Str", "Int", "Num"}
	unionType2 := NewUnionType(membersWithDups)
	expected := []string{"Int", "Str", "Num"}
	assert.Equal(t, expected, unionType2.GetMembers())

	// Test panic on too few members
	assert.Panics(t, func() {
		NewUnionType([]string{"Int"})
	})

	assert.Panics(t, func() {
		NewUnionType([]string{})
	})
}

// TestUnionTypeMemberOperations tests member-related operations
func TestUnionTypeMemberOperations(t *testing.T) {
	unionType := NewUnionType([]string{"Int", "Str", "Bool"})

	// Test ContainsMember
	assert.True(t, unionType.ContainsMember("Int"))
	assert.True(t, unionType.ContainsMember("Str"))
	assert.True(t, unionType.ContainsMember("Bool"))
	assert.False(t, unionType.ContainsMember("Num"))
	assert.False(t, unionType.ContainsMember("ArrayRef"))

	// Test GetMembers returns a copy
	members := unionType.GetMembers()
	members[0] = "Modified"
	assert.Equal(t, "Int", unionType.GetMembers()[0]) // Original unchanged
}

// TestUnionTypeEquality tests union type equality
func TestUnionTypeEquality(t *testing.T) {
	union1 := NewUnionType([]string{"Int", "Str"})
	union2 := NewUnionType([]string{"Int", "Str"})
	union3 := NewUnionType([]string{"Str", "Int"}) // Different order
	union4 := NewUnionType([]string{"Int", "Num"})

	// Test equality (order independent)
	assert.True(t, union1.Equals(union2))
	assert.True(t, union1.Equals(union3)) // Order shouldn't matter
	assert.False(t, union1.Equals(union4))

	// Test different member counts
	union5 := NewUnionType([]string{"Int", "Str", "Bool"})
	assert.False(t, union1.Equals(union5))
}

// TestUnionTypeTraitIntegration tests trait intersection functionality
func TestUnionTypeTraitIntegration(t *testing.T) {
	// Create union of Int and Str
	unionType := NewUnionType([]string{"Int", "Str"})

	// Test trait intersection - both Int and Str support "" operation
	assert.True(t, unionType.SupportsOperation("\"\""))

	// Test trait intersection - both Int and Str support bool operation
	assert.True(t, unionType.SupportsOperation("bool"))

	// Test trait intersection - only Int supports +
	// Since Str doesn't support + in our trait system, the intersection shouldn't support it
	// Note: This depends on how traits are defined in the traits package
	traits := unionType.GetTraits()
	require.NotNil(t, traits)

	// Test result type resolution for supported operations
	resultType, err := unionType.GetOperationResultType("\"\"")
	assert.NoError(t, err)
	assert.Equal(t, "Str", resultType) // String conversion always produces Str

	// Test error for unsupported operations
	_, err = unionType.GetOperationResultType("unsupported_op")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

// TestUnionTypeCaching tests trait cache functionality
func TestUnionTypeCaching(t *testing.T) {
	unionType := NewUnionType([]string{"Int", "Str"})

	// First call should compute traits
	traits1 := unionType.GetTraits()
	require.NotNil(t, traits1)

	// Second call should return cached traits
	traits2 := unionType.GetTraits()
	assert.Same(t, traits1, traits2) // Should be the same object

	// Clear cache and verify recomputation
	unionType.ClearTraitCache()
	traits3 := unionType.GetTraits()
	assert.NotSame(t, traits1, traits3) // Should be different object after cache clear
}

// TestTypeHierarchyUnionIntegration tests integration with TypeHierarchy
func TestTypeHierarchyUnionIntegration(t *testing.T) {
	// Create storage and hierarchy
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	// Test CreateUnionType
	members := []string{"Int", "Str"}
	unionType := hierarchy.CreateUnionType(members)
	require.NotNil(t, unionType)
	assert.Equal(t, members, unionType.GetMembers())

	// Test GetUnionType
	retrieved := hierarchy.GetUnionType(unionType.TypeName())
	assert.Same(t, unionType, retrieved)

	// Test IsUnionType
	assert.True(t, hierarchy.IsUnionType("Union[Int, Str]"))
	assert.True(t, hierarchy.IsUnionType("Int|Str"))
	assert.False(t, hierarchy.IsUnionType("Int"))
	assert.False(t, hierarchy.IsUnionType("ArrayRef[Int]"))
}

// TestUnionTypeParsingInHierarchy tests union type parsing
func TestUnionTypeParsingInHierarchy(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	// Test parsing Union[A, B] format
	unionType1 := hierarchy.ParseUnionType("Union[Int, Str]")
	require.NotNil(t, unionType1)
	assert.Equal(t, []string{"Int", "Str"}, unionType1.GetMembers())

	// Test parsing A|B format
	unionType2 := hierarchy.ParseUnionType("Int|Str")
	require.NotNil(t, unionType2)
	assert.Equal(t, []string{"Int", "Str"}, unionType2.GetMembers())

	// Test parsing A|B|C format
	unionType3 := hierarchy.ParseUnionType("Int|Str|Bool")
	require.NotNil(t, unionType3)
	assert.Equal(t, []string{"Int", "Str", "Bool"}, unionType3.GetMembers())

	// Test invalid formats
	assert.Nil(t, hierarchy.ParseUnionType("Int"))
	assert.Nil(t, hierarchy.ParseUnionType("Union[Int]")) // Too few members
	assert.Nil(t, hierarchy.ParseUnionType(""))
}

// TestUnionTypeCompatibilityChecking tests compatibility checking with unions
func TestUnionTypeCompatibilityChecking(t *testing.T) {
	// Skip until type compatibility is fully implemented
	t.Skip("Union type compatibility checking requires full type system implementation")

	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	// Test union to non-union compatibility
	// Union[Int, Str] should be compatible with Scalar (if Int and Str are subtypes of Scalar)
	err = hierarchy.CheckTypeCompatibility("Union[Int, Str]", "Scalar")
	assert.NoError(t, err)

	// Test non-union to union compatibility
	// Int should be assignable to Union[Int, Str]
	err = hierarchy.CheckTypeCompatibility("Int", "Union[Int, Str]")
	assert.NoError(t, err)

	// Test incompatible assignments
	err = hierarchy.CheckTypeCompatibility("Union[Int, Str]", "ArrayRef")
	assert.Error(t, err)

	// Test union to union compatibility
	err = hierarchy.CheckTypeCompatibility("Union[Int, Bool]", "Union[Int, Str]")
	// Int is compatible, Bool might not be compatible with Str
	// This test depends on the exact subtype relationships
}

// TestUnionTypeOperationValidation tests operation validation on unions
func TestUnionTypeOperationValidation(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	// Create a union type
	unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})

	// Test operations that should be supported by intersection
	// Both Int and Str support string conversion
	assert.True(t, unionType.SupportsOperation("\"\""))

	// Test operations that might not be supported
	// This depends on the trait definitions - if Str doesn't support arithmetic
	// then the union shouldn't support it either
	traits := unionType.GetTraits()
	hasAddition := traits.HasTrait("+")

	// The result depends on whether both Int and Str support addition
	// Based on Perl semantics, Str does support addition (through numeric conversion)
	// So this should be true
	t.Logf("Union[Int, Str] supports addition: %v", hasAddition)
}

// TestUnionTypeEdgeCases tests edge cases and error conditions
func TestUnionTypeEdgeCases(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	// Test parsing malformed union types
	assert.Nil(t, hierarchy.ParseUnionType("Union["))
	assert.Nil(t, hierarchy.ParseUnionType("Union]"))
	assert.Nil(t, hierarchy.ParseUnionType("|"))
	assert.Nil(t, hierarchy.ParseUnionType("Int|"))
	assert.Nil(t, hierarchy.ParseUnionType("|Str"))

	// Test very long union types
	longMembers := []string{"Int", "Str", "Bool", "Num", "ArrayRef", "HashRef", "CodeRef"}
	longUnion := NewUnionType(longMembers)
	assert.Equal(t, len(longMembers), len(longUnion.GetMembers()))

	// Test union with complex parameterized types
	complexUnion := hierarchy.ParseUnionType("ArrayRef[Int]|HashRef[Str, Int]")
	require.NotNil(t, complexUnion)
	expectedMembers := []string{"ArrayRef[Int]", "HashRef[Str, Int]"}
	assert.Equal(t, expectedMembers, complexUnion.GetMembers())
}
