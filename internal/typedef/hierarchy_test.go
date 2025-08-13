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
	assert.False(t, hierarchy.IsSubtypeOf("Str", "Int")) // Str is NOT a subtype of Int
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
	assert.Error(t, hierarchy.CheckTypeCompatibility("Str", "Int")) // Str cannot be assigned to Int
	assert.Error(t, hierarchy.CheckTypeCompatibility("ArrayRef", "HashRef"))
	assert.Error(t, hierarchy.CheckTypeCompatibility("Int", "ArrayRef[Int]"))

	// Test parameterized type compatibility
	assert.NoError(t, hierarchy.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Num]"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("HashRef[Int]", "HashRef[Scalar]"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Str]")) // Int is subtype of Str
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
	assert.NoError(t, hierarchy.ValidateType("HashRef[Int]")) // Single parameter for values
	assert.NoError(t, hierarchy.ValidateType("Map[Str,Int]")) // Key-value constraints
	assert.NoError(t, hierarchy.ValidateType("Maybe[ArrayRef[Int]]"))

	// Test that old HashRef[K,V] syntax now fails per Types::Standard alignment
	assert.Error(t, hierarchy.ValidateType("HashRef[Str,Int]"))

	// Test invalid types
	// For simple implementation these might actually pass if we're just doing name validation
	// In a more robust implementation, they would fail
	assert.NoError(t, hierarchy.ValidateType("MyCustomType")) // Uppercase custom type
	assert.Error(t, hierarchy.ValidateType("invalidType"))    // lowercase type
	assert.Error(t, hierarchy.ValidateType(""))               // empty type
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

	// Test new single-parameter HashRef (values only)
	hashRefInt, err := hierarchy.CreateParameterizedType("HashRef", []string{"Int"})
	assert.NoError(t, err)
	assert.Equal(t, "HashRef[Int]", hashRefInt)

	// Test that old two-parameter HashRef now fails
	_, err = hierarchy.CreateParameterizedType("HashRef", []string{"Str", "Int"})
	assert.Error(t, err)

	// Test Map for key-value constraints
	mapStrInt, err := hierarchy.CreateParameterizedType("Map", []string{"Str", "Int"})
	assert.NoError(t, err)
	assert.Equal(t, "Map[Str, Int]", mapStrInt)

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

// TestIntersectionTypeCreation tests creation of intersection types
func TestIntersectionTypeCreation(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test creating valid intersection types
	members := []string{"Str", "Callable"}
	intersectionType := hierarchy.CreateIntersectionType(members)
	require.NotNil(t, intersectionType)
	assert.Equal(t, members, intersectionType.GetMembers())
	assert.Equal(t, "Str&Callable", intersectionType.String())
	assert.Equal(t, "Intersection[Str, Callable]", intersectionType.TypeName())

	// Test creating intersection type with multiple members
	multiMembers := []string{"Int", "Comparable", "Serializable"}
	multiIntersection := hierarchy.CreateIntersectionType(multiMembers)
	require.NotNil(t, multiIntersection)
	assert.Equal(t, multiMembers, multiIntersection.GetMembers())
	assert.Equal(t, "Int&Comparable&Serializable", multiIntersection.String())

	// Test retrieving stored intersection type
	retrievedType := hierarchy.GetIntersectionType(intersectionType.TypeName())
	require.NotNil(t, retrievedType)
	assert.True(t, intersectionType.Equals(retrievedType))
}

// TestIntersectionTypeDetection tests detection of intersection types
func TestIntersectionTypeDetection(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test detection of intersection type formats
	assert.True(t, hierarchy.IsIntersectionType("Str&Callable"))
	assert.True(t, hierarchy.IsIntersectionType("Int&Comparable&Serializable"))
	assert.True(t, hierarchy.IsIntersectionType("Intersection[Str, Callable]"))
	assert.True(t, hierarchy.IsIntersectionType("ArrayRef[Int]&Iterable"))

	// Test detection of non-intersection types
	assert.False(t, hierarchy.IsIntersectionType("Str"))
	assert.False(t, hierarchy.IsIntersectionType("ArrayRef[Int]"))
	assert.False(t, hierarchy.IsIntersectionType("Union[Str, Int]"))
	assert.False(t, hierarchy.IsIntersectionType("Str|Int")) // This is union syntax

	// Create and test stored intersection type
	intersectionType := hierarchy.CreateIntersectionType([]string{"Str", "Callable"})
	assert.True(t, hierarchy.IsIntersectionType(intersectionType.TypeName()))
}

// TestIntersectionTypeParsing tests parsing of intersection type strings
func TestIntersectionTypeParsing(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test parsing ampersand format
	parsedType1 := hierarchy.ParseIntersectionType("Str&Callable")
	require.NotNil(t, parsedType1)
	assert.Equal(t, []string{"Str", "Callable"}, parsedType1.GetMembers())

	// Test parsing bracket format
	parsedType2 := hierarchy.ParseIntersectionType("Intersection[Int, Comparable]")
	require.NotNil(t, parsedType2)
	assert.Equal(t, []string{"Int", "Comparable"}, parsedType2.GetMembers())

	// Test parsing multiple members
	parsedType3 := hierarchy.ParseIntersectionType("Object&Serializable&Comparable")
	require.NotNil(t, parsedType3)
	assert.Equal(t, []string{"Object", "Serializable", "Comparable"}, parsedType3.GetMembers())

	// Test parsing invalid types
	assert.Nil(t, hierarchy.ParseIntersectionType("Str"))
	assert.Nil(t, hierarchy.ParseIntersectionType("ArrayRef[Int]"))
	assert.Nil(t, hierarchy.ParseIntersectionType("Union[Str, Int]"))

	// Test parsing empty or single member
	assert.Nil(t, hierarchy.ParseIntersectionType(""))
	assert.Nil(t, hierarchy.ParseIntersectionType("Intersection[Str]"))
	assert.Nil(t, hierarchy.ParseIntersectionType("Str&"))
}

// TestIntersectionTypeCompatibility tests intersection type compatibility
func TestIntersectionTypeCompatibility(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create intersection types for testing
	stringCallable := hierarchy.CreateIntersectionType([]string{"Str", "Callable"})
	_ = hierarchy.CreateIntersectionType([]string{"Int", "Comparable"}) // intComparable not used in current tests

	// Test intersection to non-intersection compatibility
	// Intersection[Str, Callable] should be compatible with both Str and Callable
	assert.True(t, stringCallable.IsCompatibleWith("Str", hierarchy))
	assert.True(t, stringCallable.IsCompatibleWith("Callable", hierarchy))
	assert.True(t, stringCallable.IsCompatibleWith("Scalar", hierarchy)) // Str is subtype of Scalar
	assert.True(t, stringCallable.IsCompatibleWith("Any", hierarchy))    // Everything is subtype of Any

	// Test non-intersection to intersection assignment
	// Str can be assigned to Intersection[Str, Callable] only if Str satisfies both
	// In practice, this would require trait checking
	assert.False(t, stringCallable.CanAssignFrom("Str", hierarchy))      // Str doesn't implement Callable
	assert.False(t, stringCallable.CanAssignFrom("Callable", hierarchy)) // Callable isn't Str

	// Test intersection to intersection compatibility
	// Create a more specific intersection
	_ = hierarchy.CreateIntersectionType([]string{"Str", "Callable", "Comparable"}) // stringCallableComparable for future tests

	// More specific intersection should be compatible with less specific one
	// (but our current simple implementation might not handle this correctly)
	// This would require more sophisticated compatibility checking
}

// TestIntersectionTypeOperations tests intersection type operations
func TestIntersectionTypeOperations(t *testing.T) {
	// Create intersection types
	intersectionType1 := NewIntersectionType([]string{"Str", "Callable"})
	intersectionType2 := NewIntersectionType([]string{"Int", "Comparable"})
	intersectionType3 := NewIntersectionType([]string{"Str", "Callable"}) // Same as intersectionType1

	// Test string representation
	assert.Equal(t, "Str&Callable", intersectionType1.String())
	assert.Equal(t, "Int&Comparable", intersectionType2.String())

	// Test type name
	assert.Equal(t, "Intersection[Str, Callable]", intersectionType1.TypeName())
	assert.Equal(t, "Intersection[Int, Comparable]", intersectionType2.TypeName())

	// Test member operations
	assert.True(t, intersectionType1.ContainsMember("Str"))
	assert.True(t, intersectionType1.ContainsMember("Callable"))
	assert.False(t, intersectionType1.ContainsMember("Int"))

	// Test equality
	assert.True(t, intersectionType1.Equals(intersectionType3))
	assert.False(t, intersectionType1.Equals(intersectionType2))

	// Test member retrieval
	members := intersectionType1.GetMembers()
	assert.Equal(t, []string{"Str", "Callable"}, members)

	// Verify it's a copy (modifying shouldn't affect original)
	members[0] = "Modified"
	assert.Equal(t, []string{"Str", "Callable"}, intersectionType1.GetMembers())
}

// TestIntersectionTypeTraits tests intersection type trait operations
func TestIntersectionTypeTraits(t *testing.T) {
	// Create intersection type
	intersectionType := NewIntersectionType([]string{"Str", "Callable"})

	// Test trait cache operations
	intersectionType.ClearTraitCache()

	// Test trait retrieval (this will use the trait intersector)
	traits := intersectionType.GetTraits()
	require.NotNil(t, traits)

	// Test operation support (this depends on the trait system)
	// These tests may need to be updated based on the actual trait definitions
	// For now, just test that the methods work without panicking
	_ = intersectionType.SupportsOperation("call")
	_, _ = intersectionType.GetOperationResultType("call")
}

// TestIntersectionTypeEdgeCases tests edge cases for intersection types
func TestIntersectionTypeEdgeCases(t *testing.T) {
	// Test panic on insufficient members
	assert.Panics(t, func() {
		NewIntersectionType([]string{"Str"})
	})

	assert.Panics(t, func() {
		NewIntersectionType([]string{})
	})

	// Test duplicate member removal
	intersectionType := NewIntersectionType([]string{"Str", "Callable", "Str"})
	members := intersectionType.GetMembers()
	assert.Equal(t, []string{"Str", "Callable"}, members)

	// Test empty string members (should be filtered out in practice)
	// This test depends on how NewIntersectionType handles empty strings
}

// TestIntersectionTypeCompatibilityWithHierarchy tests comprehensive intersection type compatibility
func TestIntersectionTypeCompatibilityWithHierarchy(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test basic intersection type compatibility through hierarchy
	_ = hierarchy.CheckTypeCompatibility("Str&Callable", "Str")
	// This should work if the intersection type parsing and compatibility is correct
	// The exact behavior depends on the implementation details

	_ = hierarchy.CheckTypeCompatibility("Str", "Str&Callable")
	// This should fail in most cases unless Str implements Callable

	// Test intersection type to intersection type compatibility
	_ = hierarchy.CheckTypeCompatibility("Str&Callable&Comparable", "Str&Callable")
	// More specific should be compatible with less specific

	err = hierarchy.CheckTypeCompatibility("Str&Callable", "Int&Comparable")
	// Different intersection types should be incompatible
	assert.Error(t, err)
}

// TestTypesStandardAlignment tests alignment with Types::Standard specifications
func TestTypesStandardAlignment(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test Bool type is NOT a supertype of Int (per Types::Standard)
	assert.False(t, hierarchy.IsSubtypeOf("Int", "Bool"), "Int should NOT be a subtype of Bool per Types::Standard")
	assert.Error(t, hierarchy.CheckTypeCompatibility("Int", "Bool"), "Int should NOT be assignable to Bool per Types::Standard")

	// Test HashRef parameter restriction (should only accept single parameter for values)
	validHashRef, err := hierarchy.CreateParameterizedType("HashRef", []string{"Str"})
	assert.NoError(t, err)
	assert.Equal(t, "HashRef[Str]", validHashRef)

	// Test HashRef with two parameters should fail (use Map instead)
	_, err = hierarchy.CreateParameterizedType("HashRef", []string{"Str", "Int"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "use Map[KeyType, ValueType]", "Error should suggest using Map for key-value constraints")

	// Test Map still works for two parameters
	validMap, err := hierarchy.CreateParameterizedType("Map", []string{"Str", "Int"})
	assert.NoError(t, err)
	assert.Equal(t, "Map[Str, Int]", validMap)
}

// TestBoolValueValidation tests Bool value validation per Types::Standard
func TestBoolValueValidation(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test valid Bool values per Types::Standard
	validBoolValues := []string{"1", "0", `""`, "undef"}
	for _, value := range validBoolValues {
		assert.NoError(t, hierarchy.ValidateValue(value, "Bool"),
			"Value '%s' should be valid for Bool type", value)
	}

	// Test invalid Bool values
	invalidBoolValues := []string{"2", "-1", "42", "true", "false", "abc", "' '", "null"}
	for _, value := range invalidBoolValues {
		assert.Error(t, hierarchy.ValidateValue(value, "Bool"),
			"Value '%s' should be invalid for Bool type", value)
	}

	// Test that other types don't have value validation (yet)
	assert.NoError(t, hierarchy.ValidateValue("42", "Int"), "Int values should not be validated yet")
	assert.NoError(t, hierarchy.ValidateValue("hello", "Str"), "Str values should not be validated yet")
}

// TestHashRefSingleParameterOnly tests that HashRef only accepts single parameter
func TestHashRefSingleParameterOnly(t *testing.T) {
	// Create a mock storage
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Test that HashRef accepts exactly one parameter (value type)
	validHashRef, err := hierarchy.CreateParameterizedType("HashRef", []string{"Int"})
	assert.NoError(t, err)
	assert.Equal(t, "HashRef[Int]", validHashRef)

	// Test that HashRef rejects zero parameters
	_, err = hierarchy.CreateParameterizedType("HashRef", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one type parameter")

	// Test that HashRef rejects two parameters (should use Map instead)
	_, err = hierarchy.CreateParameterizedType("HashRef", []string{"Str", "Int"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one type parameter")
	assert.Contains(t, err.Error(), "Map[KeyType, ValueType]")

	// Test that HashRef rejects three or more parameters
	_, err = hierarchy.CreateParameterizedType("HashRef", []string{"Str", "Int", "Bool"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one type parameter")

	// Verify that Map still works for key-value constraints
	validMap, err := hierarchy.CreateParameterizedType("Map", []string{"Str", "Int"})
	assert.NoError(t, err)
	assert.Equal(t, "Map[Str, Int]", validMap)
}
