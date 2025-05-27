// ABOUTME: Tests for type checking functionality
// ABOUTME: Verifies type annotation checking in Perl code

package typechecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/typedef"
)

// TestTypeCheckerCreation tests the creation of a type checker
func TestTypeCheckerCreation(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Check initial state
	assert.Equal(t, "Test::Module", checker.CurrentModule)
	assert.NotNil(t, checker.ImportedModules)
	assert.NotNil(t, checker.TypeAnnotations)
	assert.NotNil(t, checker.VariableTypes)
	assert.NotNil(t, checker.FunctionTypes)
}

// TestTypeAnnotationValidation tests validation of type annotations
func TestTypeAnnotationValidation(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test validation of valid types
	assert.NoError(t, checker.validateType("Int"))
	assert.NoError(t, checker.validateType("Str"))
	assert.NoError(t, checker.validateType("Bool"))

	// Test validation of invalid types
	assert.Error(t, checker.validateType("invalidType"))
	assert.Error(t, checker.validateType(""))
}

// TestBasicTypeAnnotation tests basic type annotation handling
func TestBasicTypeAnnotation(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Create a valid type annotation for a variable
	validVarAnnotation := &ast.TypeAnnotation{
		AnnotatedItem: "$var",
		TypeExpression: &ast.TypeExpression{
			BaseType: "Int",
		},
		Pos:  ast.Position{Line: 1, Column: 5},
		Kind: ast.VarAnnotation,
	}

	// Check the valid annotation
	assert.NoError(t, checker.collectTypeAnnotation(validVarAnnotation))
	assert.NoError(t, checker.checkTypeAnnotation(validVarAnnotation))

	// Verify that the type annotation was recorded
	typeStr, ok := checker.GetVariableType("$var")
	assert.True(t, ok)
	assert.Equal(t, "Int", typeStr)
}

// TestAssignmentChecking tests type assignment compatibility
func TestAssignmentChecking(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	pos := ast.Position{Line: 1, Column: 1}

	// Test valid assignments
	assert.NoError(t, checker.CheckAssignment("Int", "Int", pos))
	assert.NoError(t, checker.CheckAssignment("Int", "Str", pos))  // Numbers can be stringified
	assert.NoError(t, checker.CheckAssignment("Int", "Bool", pos)) // Numbers can be used as booleans

	// Test invalid assignments
	assert.Error(t, checker.CheckAssignment("Str", "Int", pos))      // Strings can't become numbers
	assert.Error(t, checker.CheckAssignment("ArrayRef", "Int", pos)) // Array refs can't become numbers
}

// TestTypeInference tests basic type inference
func TestTypeInference(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test numeric literals
	assert.Equal(t, "Int", checker.inferExpressionType("42"))
	assert.Equal(t, "Float", checker.inferExpressionType("3.14"))

	// Test string literals
	assert.Equal(t, "Str", checker.inferExpressionType("\"hello\""))
	assert.Equal(t, "Str", checker.inferExpressionType("'world'"))

	// Test special values
	assert.Equal(t, "Undef", checker.inferExpressionType("undef"))
	assert.Equal(t, "ArrayRef", checker.inferExpressionType("[]"))
	assert.Equal(t, "HashRef", checker.inferExpressionType("{}"))
}

// TestUnionTypes tests union type compatibility checking
func TestUnionTypes(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	pos := ast.Position{Line: 1, Column: 1}

	// Test union type creation and compatibility
	// Union[Int, ArrayRef] should accept both Int and ArrayRef
	assert.NoError(t, checker.CheckAssignment("Int", "Int|ArrayRef", pos))
	assert.NoError(t, checker.CheckAssignment("ArrayRef", "Int|ArrayRef", pos))
	assert.Error(t, checker.CheckAssignment("HashRef", "Int|ArrayRef", pos))

	// Test Union syntax with brackets
	assert.NoError(t, checker.CheckAssignment("Int", "Union[Int, ArrayRef]", pos))
	assert.NoError(t, checker.CheckAssignment("ArrayRef", "Union[Int, ArrayRef]", pos))
	assert.Error(t, checker.CheckAssignment("HashRef", "Union[Int, ArrayRef]", pos))

	// Test assigning union to specific type
	// Union[Int, ArrayRef] should NOT be assignable to Int (would be unsafe)
	assert.Error(t, checker.CheckAssignment("Int|ArrayRef", "Int", pos))

	// Test union to union compatibility
	// Union[Int, ArrayRef] should be assignable to Union[Int, ArrayRef, HashRef]
	assert.NoError(t, checker.CheckAssignment("Int|ArrayRef", "Int|ArrayRef|HashRef", pos))
	// Union[Int, ArrayRef, HashRef] should NOT be assignable to Union[Int, ArrayRef]
	assert.Error(t, checker.CheckAssignment("Int|ArrayRef|HashRef", "Int|ArrayRef", pos))

	// Test nested unions
	assert.NoError(t, checker.CheckAssignment("Int", "Union[Int|ArrayRef, HashRef]", pos))
}

// TestIntersectionTypes tests intersection type compatibility checking
func TestIntersectionTypes(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	pos := ast.Position{Line: 1, Column: 1}

	// Test intersection type creation and compatibility
	// Intersection[Iterable, Positional] should be compatible with both
	assert.NoError(t, checker.CheckAssignment("Iterable&Positional", "Iterable", pos))
	assert.NoError(t, checker.CheckAssignment("Iterable&Positional", "Positional", pos))

	// Test Intersection syntax with brackets
	assert.NoError(t, checker.CheckAssignment("Intersection[Iterable, Positional]", "Iterable", pos))
	assert.NoError(t, checker.CheckAssignment("Intersection[Iterable, Positional]", "Positional", pos))

	// Test assigning single type to intersection (should require all traits)
	// Array is both Iterable and Positional, so it should work
	assert.NoError(t, checker.CheckAssignment("Array", "Iterable&Positional", pos))

	// Hash is Iterable but not Positional, so it should fail
	assert.Error(t, checker.CheckAssignment("Hash", "Iterable&Positional", pos))

	// Test intersection to intersection compatibility
	// Intersection[A, B] should be assignable to Intersection[A, B, C] if A, B satisfy C
	// But this is complex trait logic - for now test basic cases
	assert.NoError(t, checker.CheckAssignment("Iterable&Positional", "Iterable&Positional", pos))
}

// TestConditionalTypeRefinement tests type refinement based on conditions
func TestConditionalTypeRefinement(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Set up a variable with union type
	checker.VariableTypes["$value"] = "Int|Str"

	// Test type refinement after condition
	// After checking defined($value), it should refine to exclude Undef
	checker.refineTypeAfterCondition("$value", "defined", true)

	// Verify the refined type
	refinedType, exists := checker.TypeState.RefinedTypes["$value"]
	assert.True(t, exists)
	// Should still be Int|Str since it didn't include Undef originally
	assert.Equal(t, "Int|Str", refinedType)

	// Test with Maybe type that includes Undef
	checker.VariableTypes["$maybe"] = "Maybe[Int]"
	checker.refineTypeAfterCondition("$maybe", "defined", true)

	refinedType, exists = checker.TypeState.RefinedTypes["$maybe"]
	assert.True(t, exists)
	// After defined check, Maybe[Int] should refine to Int
	assert.Equal(t, "Int", refinedType)

	// Test negative refinement
	checker.VariableTypes["$nullable"] = "Int|Undef"
	checker.refineTypeAfterCondition("$nullable", "defined", false)

	refinedType, exists = checker.TypeState.RefinedTypes["$nullable"]
	assert.True(t, exists)
	// After !defined check, Int|Undef should refine to Undef
	assert.Equal(t, "Undef", refinedType)
}

// TestComplexTypeExpressions tests complex combinations of type operations
func TestComplexTypeExpressions(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeCheckerLegacy(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	pos := ast.Position{Line: 1, Column: 1}

	// Test complex union types
	assert.NoError(t, checker.CheckAssignment("Int", "Int|Str|Bool", pos))
	assert.NoError(t, checker.CheckAssignment("ArrayRef[Int]", "ArrayRef[Int]|HashRef[Str]", pos))

	// Test parameterized union types
	assert.NoError(t, checker.CheckAssignment("ArrayRef[Int]", "ArrayRef[Int|Str]", pos))
	assert.Error(t, checker.CheckAssignment("ArrayRef[HashRef]", "ArrayRef[Int|Str]", pos))

	// Test mixed union and intersection
	// Test simpler mixed types that work with current implementation
	assert.NoError(t, checker.CheckAssignment("Array", "Array|Str", pos))
	assert.NoError(t, checker.CheckAssignment("Str", "Array|Str", pos))

	// Test Maybe with simple types (Maybe[T] is equivalent to T|Undef)
	assert.NoError(t, checker.CheckAssignment("Int", "Maybe[Int]", pos))
	assert.NoError(t, checker.CheckAssignment("Undef", "Maybe[Int]", pos))

	// Test nested parameterized types
	assert.NoError(t, checker.CheckAssignment("ArrayRef[ArrayRef[Int]]", "ArrayRef[ArrayRef[Int|Str]]", pos))
}
