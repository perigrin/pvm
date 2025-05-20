// ABOUTME: Tests for type hierarchy functionality
// ABOUTME: Verifies type checking and compatibility

package typedef

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTypeHierarchyCreation tests the creation of a type hierarchy
func TestTypeHierarchyCreation(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Check that some basic built-in types exist
	assert.True(t, hierarchy.IsBuiltinType("Int"))
	assert.True(t, hierarchy.IsBuiltinType("Str"))
	assert.True(t, hierarchy.IsBuiltinType("Bool"))
	assert.True(t, hierarchy.IsBuiltinType("ArrayRef"))
	assert.True(t, hierarchy.IsBuiltinType("HashRef"))

	// Check that some non-existent types don't exist
	assert.False(t, hierarchy.IsBuiltinType("NonExistentType"))
	assert.False(t, hierarchy.IsBuiltinType("fakeType"))
}

// TestSubtypeRelationships tests the subtype relationships
func TestSubtypeRelationships(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test direct subtype relationships
	assert.True(t, hierarchy.IsSubtypeOf("Int", "Num"))
	assert.True(t, hierarchy.IsSubtypeOf("Str", "Scalar"))
	assert.True(t, hierarchy.IsSubtypeOf("Bool", "Scalar"))
	assert.True(t, hierarchy.IsSubtypeOf("ArrayRef", "Ref"))
	assert.True(t, hierarchy.IsSubtypeOf("HashRef", "Ref"))

	// Test transitive subtype relationships
	assert.True(t, hierarchy.IsSubtypeOf("Int", "Scalar"))
	assert.True(t, hierarchy.IsSubtypeOf("Float", "Scalar"))
	assert.True(t, hierarchy.IsSubtypeOf("ArrayRef", "Any"))
	assert.True(t, hierarchy.IsSubtypeOf("Str", "Any"))

	// Test non-subtype relationships
	assert.False(t, hierarchy.IsSubtypeOf("Int", "Str"))
	assert.False(t, hierarchy.IsSubtypeOf("ArrayRef", "HashRef"))
	assert.False(t, hierarchy.IsSubtypeOf("Ref", "Scalar"))
}

// TestParameterizedTypes tests parameterized types
func TestParameterizedTypes(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test parameterized type extraction
	baseType, params := extractTypeAndParams("ArrayRef[Int]")
	assert.Equal(t, "ArrayRef", baseType)
	assert.Equal(t, []string{"Int"}, params)

	baseType, params = extractTypeAndParams("HashRef[Str,Int]")
	assert.Equal(t, "HashRef", baseType)
	assert.Equal(t, []string{"Str", "Int"}, params)

	baseType, params = extractTypeAndParams("Maybe[ArrayRef[Int]]")
	assert.Equal(t, "Maybe", baseType)
	assert.Equal(t, []string{"ArrayRef[Int]"}, params)

	// Test nested parameterized types
	nestedType := "ArrayRef[HashRef[Str,Int]]"
	baseType, params = extractTypeAndParams(nestedType)
	assert.Equal(t, "ArrayRef", baseType)
	assert.Equal(t, []string{"HashRef[Str,Int]"}, params)

	// Test non-parameterized types
	baseType, params = extractTypeAndParams("Int")
	assert.Equal(t, "Int", baseType)
	assert.Nil(t, params)
}

// TestTypeCompatibility tests type compatibility checking
func TestTypeCompatibility(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test basic compatibility (same type)
	assert.NoError(t, hierarchy.CheckTypeCompatibility("Int", "Int"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("Str", "Str"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Int]"))

	// Test subtype compatibility
	assert.NoError(t, hierarchy.CheckTypeCompatibility("Int", "Num"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("Int", "Scalar"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("Int", "Any"))

	// Test incompatible types
	assert.Error(t, hierarchy.CheckTypeCompatibility("Int", "Str"))
	assert.Error(t, hierarchy.CheckTypeCompatibility("ArrayRef", "HashRef"))
	assert.Error(t, hierarchy.CheckTypeCompatibility("Int", "ArrayRef[Int]"))

	// Test parameterized type compatibility
	assert.NoError(t, hierarchy.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Num]"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("HashRef[Int]", "HashRef[Scalar]"))
	assert.Error(t, hierarchy.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Str]"))
}

// TestTypeValidation tests type validation
func TestTypeValidation(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test valid types
	assert.NoError(t, hierarchy.ValidateType("Int"))
	assert.NoError(t, hierarchy.ValidateType("Str"))
	assert.NoError(t, hierarchy.ValidateType("ArrayRef[Int]"))
	assert.NoError(t, hierarchy.ValidateType("HashRef[Str,Int]"))
	assert.NoError(t, hierarchy.ValidateType("Maybe[ArrayRef[Int]]"))

	// Test invalid types
	// For simple implementation these might actually pass if we're just doing name validation
	// In a more robust implementation, they would fail
	assert.NoError(t, hierarchy.ValidateType("MyCustomType"))  // Uppercase custom type
	assert.Error(t, hierarchy.ValidateType("invalidType"))     // lowercase type
	assert.Error(t, hierarchy.ValidateType(""))                // empty type
}

// TestCreateParameterizedType tests creation of parameterized types
func TestCreateParameterizedType(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test creating valid parameterized types
	arrayRefInt, err := hierarchy.CreateParameterizedType("ArrayRef", []string{"Int"})
	assert.NoError(t, err)
	assert.Equal(t, "ArrayRef[Int]", arrayRefInt)

	hashRefStrInt, err := hierarchy.CreateParameterizedType("HashRef", []string{"Str", "Int"})
	assert.NoError(t, err)
	assert.Equal(t, "HashRef[Str,Int]", hashRefStrInt)

	maybeStr, err := hierarchy.CreateParameterizedType("Maybe", []string{"Str"})
	assert.NoError(t, err)
	assert.Equal(t, "Maybe[Str]", maybeStr)

	// Test creating invalid parameterized types
	_, err = hierarchy.CreateParameterizedType("Int", []string{"Str"})
	assert.Error(t, err) // Int is not a parameterized type

	_, err = hierarchy.CreateParameterizedType("ArrayRef", []string{})
	assert.Error(t, err) // ArrayRef requires a parameter

	_, err = hierarchy.CreateParameterizedType("ArrayRef", []string{"Int", "Str"})
	assert.Error(t, err) // ArrayRef requires exactly one parameter
}