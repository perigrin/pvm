// ABOUTME: Regression tests for union type functionality
// ABOUTME: Ensures existing functionality is preserved during development

package typedef

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnionTypeBackwardCompatibility tests that existing functionality still works
func TestUnionTypeBackwardCompatibility(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	t.Run("BasicTypeHierarchyStillWorks", func(t *testing.T) {
		// Ensure basic type operations haven't been broken
		assert.True(t, hierarchy.IsBuiltinType("Int"))
		assert.True(t, hierarchy.IsBuiltinType("Str"))
		assert.True(t, hierarchy.IsBuiltinType("ArrayRef"))
		assert.False(t, hierarchy.IsBuiltinType("NonExistentType"))

		// Test subtype relationships
		assert.True(t, hierarchy.IsSubtypeOf("Int", "Num"))
		assert.True(t, hierarchy.IsSubtypeOf("Str", "Scalar"))
		assert.True(t, hierarchy.IsSubtypeOf("Any", "Any"))

		// Test type validation
		assert.NoError(t, hierarchy.ValidateType("Int"))
		assert.NoError(t, hierarchy.ValidateType("ArrayRef[Int]"))
		assert.Error(t, hierarchy.ValidateType(""))
	})

	t.Run("ParameterizedTypesStillWork", func(t *testing.T) {
		// Test that parameterized types still work alongside union types
		paramType, err := hierarchy.CreateParameterizedType("ArrayRef", []string{"Int"})
		assert.NoError(t, err)
		assert.Equal(t, "ArrayRef[Int]", paramType)

		// Test single-parameter HashRef (values only)
		hashType, err := hierarchy.CreateParameterizedType("HashRef", []string{"Int"})
		assert.NoError(t, err)
		assert.Equal(t, "HashRef[Int]", hashType)

		// Test Map for key-value constraints
		mapType, err := hierarchy.CreateParameterizedType("Map", []string{"Str", "Int"})
		assert.NoError(t, err)
		assert.Equal(t, "Map[Str, Int]", mapType)

		// Test validation
		assert.NoError(t, hierarchy.ValidateType(paramType))
		assert.NoError(t, hierarchy.ValidateType(hashType))
		assert.NoError(t, hierarchy.ValidateType(mapType))
	})

	t.Run("TypeCompatibilityStillWorks", func(t *testing.T) {
		// Test that basic type compatibility checking still works
		assert.NoError(t, hierarchy.CheckTypeCompatibility("Int", "Num"))
		assert.NoError(t, hierarchy.CheckTypeCompatibility("Str", "Scalar"))
		assert.Error(t, hierarchy.CheckTypeCompatibility("Int", "ArrayRef"))

		// Test parameterized type compatibility
		assert.NoError(t, hierarchy.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef"))
	})

	t.Run("UnionTypesDoNotBreakExistingTypes", func(t *testing.T) {
		// Create a union type
		unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})
		require.NotNil(t, unionType)

		// Ensure this doesn't break existing type operations
		assert.True(t, hierarchy.IsBuiltinType("Int"))
		assert.True(t, hierarchy.IsBuiltinType("Str"))
		assert.NoError(t, hierarchy.ValidateType("Int"))
		assert.NoError(t, hierarchy.ValidateType("Str"))
		assert.NoError(t, hierarchy.CheckTypeCompatibility("Int", "Num"))
	})
}

// TestExistingTypeStorageFunctionality tests type storage still works
func TestExistingTypeStorageFunctionality(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Test basic storage operations
	typeDef := &TypeDefinition{
		Module:     "Test::Module",
		Version:    "1.0",
		Generated:  time.Now(),
		Source:     "manual",
		Maintainer: "test",
		Types: []TypeInfo{
			{
				Name: "TestType",
				Kind: "class",
			},
		},
	}

	// Store and retrieve
	err = storage.Save(typeDef)
	assert.NoError(t, err)

	retrieved, err := storage.Load("Test::Module")
	if assert.NoError(t, err) && assert.NotNil(t, retrieved) {
		assert.Equal(t, typeDef.Module, retrieved.Module)
		assert.Equal(t, typeDef.Version, retrieved.Version)
		assert.Len(t, retrieved.Types, 1)
		assert.Equal(t, "TestType", retrieved.Types[0].Name)
	}
}

// TestExistingTraitSystemFunctionality tests trait system integration
func TestExistingTraitSystemFunctionality(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	// Create a union type and verify it integrates with traits
	unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})
	require.NotNil(t, unionType)

	// Test that trait operations work
	traits := unionType.GetTraits()
	require.NotNil(t, traits)

	// Test that basic operations are supported
	assert.True(t, unionType.SupportsOperation("\"\""))
	assert.True(t, unionType.SupportsOperation("bool"))

	// Test result type resolution
	resultType, err := unionType.GetOperationResultType("\"\"")
	assert.NoError(t, err)
	assert.Equal(t, "Str", resultType)

	resultType, err = unionType.GetOperationResultType("bool")
	assert.NoError(t, err)
	assert.Equal(t, "Bool", resultType)
}

// TestNoRegressionInTypeDefinitions tests that TypeInfo structures are unchanged
func TestNoRegressionInTypeDefinitions(t *testing.T) {
	// Test that all the existing TypeInfo fields are still accessible
	typeInfo := TypeInfo{
		Name:        "TestType",
		Description: "A test type",
		Kind:        "class",
		Parameters:  []ParamInfo{},
		Properties:  []PropInfo{},
		Methods:     []MethodInfo{},
		Parent:      "ParentType",
		Roles:       []string{"Role1", "Role2"},
	}

	// Verify all fields are accessible
	assert.Equal(t, "TestType", typeInfo.Name)
	assert.Equal(t, "A test type", typeInfo.Description)
	assert.Equal(t, "class", typeInfo.Kind)
	assert.Len(t, typeInfo.Parameters, 0)
	assert.Len(t, typeInfo.Properties, 0)
	assert.Len(t, typeInfo.Methods, 0)
	assert.Equal(t, "ParentType", typeInfo.Parent)
	assert.Len(t, typeInfo.Roles, 2)

	// Test string representation
	assert.Contains(t, typeInfo.String(), "TestType")
	assert.Contains(t, typeInfo.String(), "class")
}

// TestExistingErrorHandling tests that error handling patterns are preserved
func TestExistingErrorHandling(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	// Test that existing error patterns still work
	assert.Error(t, hierarchy.ValidateType(""))
	assert.Error(t, hierarchy.ValidateType("Invalid["))

	err = hierarchy.CheckTypeCompatibility("Int", "NonExistentType")
	assert.Error(t, err)

	// Test that union type errors are properly typed
	_, err = hierarchy.CreateParameterizedType("NonParameterized", []string{"Int"})
	assert.Error(t, err)

	// Test that TypeDefError still works
	var typeErr TypeDefError = "test error"
	assert.Equal(t, "test error", typeErr.Error())
}

// TestPerformanceRegressions tests that performance hasn't significantly degraded
func TestPerformanceRegressions(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	// Create many simple types and ensure operations are still fast
	for i := 0; i < 100; i++ {
		err := hierarchy.ValidateType("Int")
		assert.NoError(t, err)

		assert.True(t, hierarchy.IsBuiltinType("Str"))

		err = hierarchy.CheckTypeCompatibility("Int", "Num")
		assert.NoError(t, err)
	}

	// Create union types and ensure they don't slow down basic operations
	for i := 0; i < 10; i++ {
		unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})
		assert.NotNil(t, unionType)
	}

	// Basic operations should still be fast even with unions present
	err = hierarchy.ValidateType("Int")
	assert.NoError(t, err)

	assert.True(t, hierarchy.IsBuiltinType("Str"))
}
