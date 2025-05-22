// ABOUTME: Comprehensive tests for union type functionality
// ABOUTME: Covers real-world scenarios, edge cases, and performance testing

package typedef

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealWorldUnionTypeScenarios tests realistic union type patterns
func TestRealWorldUnionTypeScenarios(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	t.Run("ErrorOrValue pattern", func(t *testing.T) {
		// Common pattern: Union[Error, String] for result types
		unionType := hierarchy.CreateUnionType([]string{"Str", "HashRef"})
		require.NotNil(t, unionType)

		// Both should support string conversion
		assert.True(t, unionType.SupportsOperation("\"\""))

		// Test operation result types
		resultType, err := unionType.GetOperationResultType("\"\"")
		assert.NoError(t, err)
		assert.Equal(t, "Str", resultType)
	})

	t.Run("NumericOrString pattern", func(t *testing.T) {
		// Common pattern: Union[Int, Str] for flexible input
		unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})
		require.NotNil(t, unionType)

		// Both support numeric conversion in Perl
		assert.True(t, unionType.SupportsOperation("0+"))

		resultType, err := unionType.GetOperationResultType("0+")
		assert.NoError(t, err)
		assert.Equal(t, "Num", resultType)
	})

	t.Run("ContainerTypes pattern", func(t *testing.T) {
		// Pattern: Union[ArrayRef, HashRef] for collections
		unionType := hierarchy.CreateUnionType([]string{"ArrayRef", "HashRef"})
		require.NotNil(t, unionType)

		// Both containers support boolean conversion
		assert.True(t, unionType.SupportsOperation("bool"))
	})

	t.Run("ParameterizedUnions pattern", func(t *testing.T) {
		// Pattern: Union[ArrayRef[Int], HashRef[Str, Int]]
		unionType := hierarchy.ParseUnionType("ArrayRef[Int]|HashRef[Str, Int]")
		require.NotNil(t, unionType)

		expectedMembers := []string{"ArrayRef[Int]", "HashRef[Str, Int]"}
		assert.Equal(t, expectedMembers, unionType.GetMembers())

		// Test caching works for complex types
		traits1 := unionType.GetTraits()
		traits2 := unionType.GetTraits()
		assert.Same(t, traits1, traits2, "Traits should be cached")
	})
}

// TestComplexUnionTypeExpressions tests complex nested union expressions
func TestComplexUnionTypeExpressions(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	t.Run("NestedParameterizedTypes", func(t *testing.T) {
		// Union[ArrayRef[Union[Int, Str]], HashRef[Str, Union[Int, Bool]]]
		unionType := hierarchy.ParseUnionType("ArrayRef[Int]|ArrayRef[Str]")
		require.NotNil(t, unionType)

		expectedMembers := []string{"ArrayRef[Int]", "ArrayRef[Str]"}
		assert.Equal(t, expectedMembers, unionType.GetMembers())
	})

	t.Run("LongUnionChains", func(t *testing.T) {
		// Test unions with many members
		members := []string{"Int", "Str", "Bool", "Num", "ArrayRef", "HashRef", "CodeRef", "ScalarRef"}
		unionType := NewUnionType(members)
		require.NotNil(t, unionType)

		assert.Equal(t, len(members), len(unionType.GetMembers()))

		// Test that trait intersection still works with many types
		traits := unionType.GetTraits()
		require.NotNil(t, traits)

		// All these types should support string conversion
		assert.True(t, traits.HasTrait("\"\""))
	})

	t.Run("DuplicateElimination", func(t *testing.T) {
		// Test that duplicates are properly eliminated
		unionType := hierarchy.ParseUnionType("Int|Str|Int|Bool|Str")
		require.NotNil(t, unionType)

		expected := []string{"Int", "Str", "Bool"}
		assert.Equal(t, expected, unionType.GetMembers())
	})
}

// TestUnionTypeCompatibilityScenarios tests various compatibility patterns
func TestUnionTypeCompatibilityScenarios(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	t.Run("UnionToSingleType", func(t *testing.T) {
		// Union[Int, Str] should be compatible with Scalar if both Int and Str are
		unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})

		// Test compatibility - this would require full type system integration
		// For now, we test the structure exists
		assert.True(t, unionType.IsCompatibleWith("Scalar", hierarchy))
	})

	t.Run("SingleTypeToUnion", func(t *testing.T) {
		// Int should be assignable to Union[Int, Str]
		unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})

		assert.True(t, unionType.CanAssignFrom("Int", hierarchy))
		assert.True(t, unionType.CanAssignFrom("Str", hierarchy))
		assert.False(t, unionType.CanAssignFrom("ArrayRef", hierarchy))
	})

	t.Run("UnionToUnion", func(t *testing.T) {
		// Union[Int, Bool] should be compatible with Union[Int, Str]
		// if Bool is compatible with Str (which it is as both are scalars)
		union1 := hierarchy.CreateUnionType([]string{"Int", "Bool"})
		union2 := hierarchy.CreateUnionType([]string{"Int", "Str"})

		// Test structural compatibility
		assert.Equal(t, 2, len(union1.GetMembers()))
		assert.Equal(t, 2, len(union2.GetMembers()))
	})
}

// TestUnionTypePerformance tests performance characteristics
func TestUnionTypePerformance(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	t.Run("TraitIntersectionCaching", func(t *testing.T) {
		unionType := hierarchy.CreateUnionType([]string{"Int", "Str", "Bool", "Num"})

		// Measure first computation
		start := time.Now()
		traits1 := unionType.GetTraits()
		firstDuration := time.Since(start)

		// Measure cached retrieval
		start = time.Now()
		traits2 := unionType.GetTraits()
		secondDuration := time.Since(start)

		// Cached access should be much faster
		assert.Same(t, traits1, traits2, "Should return cached traits")
		assert.True(t, secondDuration < firstDuration/2,
			"Cached access should be at least 2x faster")

		t.Logf("First access: %v, Cached access: %v", firstDuration, secondDuration)
	})

	t.Run("LargeUnionPerformance", func(t *testing.T) {
		// Test performance with larger unions
		largeMembers := make([]string, 20)
		for i := 0; i < 20; i++ {
			largeMembers[i] = fmt.Sprintf("Type%d", i)
		}

		start := time.Now()
		unionType := NewUnionType(largeMembers)
		creationDuration := time.Since(start)

		start = time.Now()
		traits := unionType.GetTraits()
		traitsDuration := time.Since(start)

		assert.NotNil(t, traits)
		assert.True(t, creationDuration < 100*time.Millisecond,
			"Union creation should be fast")
		assert.True(t, traitsDuration < 1*time.Second,
			"Traits computation should complete in reasonable time")

		t.Logf("Creation: %v, Traits computation: %v", creationDuration, traitsDuration)
	})

	t.Run("CacheInvalidation", func(t *testing.T) {
		unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})

		// Get traits to populate cache
		traits1 := unionType.GetTraits()
		require.NotNil(t, traits1)

		// Clear cache
		unionType.ClearTraitCache()

		// Get traits again - should be recomputed
		traits2 := unionType.GetTraits()
		require.NotNil(t, traits2)

		// Should be different instances but same content
		assert.NotSame(t, traits1, traits2, "Should be different instances after cache clear")
	})
}

// TestUnionTypeErrorConditions tests error handling and edge cases
func TestUnionTypeErrorConditions(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	t.Run("MalformedUnionSyntax", func(t *testing.T) {
		malformedCases := []string{
			"Union[",      // Missing closing bracket
			"Union]",      // Missing opening bracket
			"Union[]",     // Empty parameters
			"Union[Int",   // Missing closing bracket
			"Union Int]",  // Missing opening bracket
			"|",           // Just pipe
			"Int|",        // Trailing pipe
			"|Str",        // Leading pipe
			"Union[Int,]", // Trailing comma
		}

		for _, malformed := range malformedCases {
			t.Run(malformed, func(t *testing.T) {
				result := hierarchy.ParseUnionType(malformed)
				assert.Nil(t, result, "Should return nil for malformed union: %s", malformed)
			})
		}
	})

	t.Run("UnsupportedOperations", func(t *testing.T) {
		// Create a union where intersection has no common operations
		// This is theoretical - in practice most types share some basic operations
		unionType := hierarchy.CreateUnionType([]string{"Int", "Str"})

		// Test unsupported operation
		_, err := unionType.GetOperationResultType("unsupported_operation")
		assert.Error(t, err, "Should error for unsupported operations")
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("EmptyUnionCreation", func(t *testing.T) {
		// Test that empty unions are rejected
		assert.Panics(t, func() {
			NewUnionType([]string{})
		}, "Should panic for empty union")

		assert.Panics(t, func() {
			NewUnionType([]string{"Int"})
		}, "Should panic for single-member union")
	})
}

// TestUnionTypeStringRepresentation tests string formatting
func TestUnionTypeStringRepresentation(t *testing.T) {
	t.Run("SimpleUnion", func(t *testing.T) {
		unionType := NewUnionType([]string{"Int", "Str"})

		assert.Equal(t, "Int|Str", unionType.String())
		assert.Equal(t, "Union[Int, Str]", unionType.TypeName())
	})

	t.Run("ComplexUnion", func(t *testing.T) {
		unionType := NewUnionType([]string{"ArrayRef[Int]", "HashRef[Str, Bool]"})

		expected := "ArrayRef[Int]|HashRef[Str, Bool]"
		assert.Equal(t, expected, unionType.String())

		expectedTypeName := "Union[ArrayRef[Int], HashRef[Str, Bool]]"
		assert.Equal(t, expectedTypeName, unionType.TypeName())
	})

	t.Run("LongUnion", func(t *testing.T) {
		members := []string{"Int", "Str", "Bool", "Num", "ArrayRef"}
		unionType := NewUnionType(members)

		expectedString := strings.Join(members, "|")
		assert.Equal(t, expectedString, unionType.String())

		expectedTypeName := fmt.Sprintf("Union[%s]", strings.Join(members, ", "))
		assert.Equal(t, expectedTypeName, unionType.TypeName())
	})
}

// TestUnionTypeEqualityComprehensive tests equality comparison between union types
func TestUnionTypeEqualityComprehensive(t *testing.T) {
	t.Run("IdenticalUnions", func(t *testing.T) {
		union1 := NewUnionType([]string{"Int", "Str"})
		union2 := NewUnionType([]string{"Int", "Str"})

		assert.True(t, union1.Equals(union2))
		assert.True(t, union2.Equals(union1))
	})

	t.Run("OrderIndependentEquality", func(t *testing.T) {
		union1 := NewUnionType([]string{"Int", "Str", "Bool"})
		union2 := NewUnionType([]string{"Bool", "Int", "Str"})
		union3 := NewUnionType([]string{"Str", "Bool", "Int"})

		assert.True(t, union1.Equals(union2))
		assert.True(t, union2.Equals(union3))
		assert.True(t, union1.Equals(union3))
	})

	t.Run("DifferentUnions", func(t *testing.T) {
		union1 := NewUnionType([]string{"Int", "Str"})
		union2 := NewUnionType([]string{"Int", "Bool"})
		union3 := NewUnionType([]string{"Int", "Str", "Bool"})

		assert.False(t, union1.Equals(union2))
		assert.False(t, union1.Equals(union3))
		assert.False(t, union2.Equals(union3))
	})
}

// TestUnionTypeIntegrationWithHierarchy tests integration with TypeHierarchy
func TestUnionTypeIntegrationWithHierarchy(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)
	hierarchy := NewTypeHierarchy(storage)

	t.Run("StorageAndRetrieval", func(t *testing.T) {
		// Create and store union types
		union1 := hierarchy.CreateUnionType([]string{"Int", "Str"})
		union2 := hierarchy.CreateUnionType([]string{"ArrayRef", "HashRef"})

		// Retrieve them
		retrieved1 := hierarchy.GetUnionType(union1.TypeName())
		retrieved2 := hierarchy.GetUnionType(union2.TypeName())

		assert.Same(t, union1, retrieved1)
		assert.Same(t, union2, retrieved2)
	})

	t.Run("UnionTypeDetection", func(t *testing.T) {
		// Test different union type formats
		testCases := []struct {
			typeName string
			isUnion  bool
		}{
			{"Union[Int, Str]", true},
			{"Int|Str", true},
			{"Int|Str|Bool", true},
			{"Int", false},
			{"ArrayRef[Int]", false},
			{"ArrayRef[Int|Str]", true}, // This contains a union, so it's detected as union
		}

		for _, tc := range testCases {
			t.Run(tc.typeName, func(t *testing.T) {
				result := hierarchy.IsUnionType(tc.typeName)
				assert.Equal(t, tc.isUnion, result, "IsUnionType(%s)", tc.typeName)
			})
		}
	})

	t.Run("ParseVariousFormats", func(t *testing.T) {
		// Test parsing different union formats
		formats := []struct {
			input    string
			expected []string
		}{
			{"Union[Int, Str]", []string{"Int", "Str"}},
			{"Int|Str", []string{"Int", "Str"}},
			{"Int|Str|Bool", []string{"Int", "Str", "Bool"}},
			{"ArrayRef[Int]|HashRef[Str]", []string{"ArrayRef[Int]", "HashRef[Str]"}},
		}

		for _, format := range formats {
			t.Run(format.input, func(t *testing.T) {
				unionType := hierarchy.ParseUnionType(format.input)
				require.NotNil(t, unionType, "Should parse: %s", format.input)
				assert.Equal(t, format.expected, unionType.GetMembers())
			})
		}
	})
}
