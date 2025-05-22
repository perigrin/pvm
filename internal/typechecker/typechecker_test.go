// ABOUTME: Tests for type checking functionality
// ABOUTME: Verifies type annotation checking in Perl code

package typechecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/parser"
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
	checker := NewTypeChecker(hierarchy, "Test::Module")
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
	checker := NewTypeChecker(hierarchy, "Test::Module")
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
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Create a valid type annotation for a variable
	validVarAnnotation := &parser.TypeAnnotation{
		AnnotatedItem: "$var",
		TypeExpression: &parser.TypeExpression{
			Name: "Int",
			Pos:  parser.Position{Line: 1, Column: 5},
		},
		Pos:  parser.Position{Line: 1, Column: 5},
		Kind: parser.VarAnnotation,
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
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	pos := parser.Position{Line: 1, Column: 1}

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
	checker := NewTypeChecker(hierarchy, "Test::Module")
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
