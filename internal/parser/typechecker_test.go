// ABOUTME: Tests for type checking functionality
// ABOUTME: Verifies type annotation checking in Perl code

package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NoError(t, checker.validateType("ArrayRef[Int]"))
	assert.NoError(t, checker.validateType("HashRef[Str,Int]"))
	assert.NoError(t, checker.validateType("Maybe[Str]"))

	// Test validation of union and intersection types
	assert.NoError(t, checker.validateType("Int|Str"))
	assert.NoError(t, checker.validateType("Int&Num"))

	// Test validation of negation types
	assert.NoError(t, checker.validateType("!Int"))
	assert.NoError(t, checker.validateType("!ArrayRef[Int]"))

	// Test error cases - with our current implementation, custom types might pass validation
	// but in a more robust implementation they would fail if the type is not defined
	assert.NoError(t, checker.validateType("CustomType"))  // Uppercase is okay
	assert.Error(t, checker.validateType("invalidType"))   // lowercase is not valid
}

// TestTypeCompatibilityChecking tests checking of type compatibility
func TestTypeCompatibilityChecking(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test compatible types
	assert.NoError(t, checker.CheckTypeCompatibility("Int", "Int"))
	assert.NoError(t, checker.CheckTypeCompatibility("Int", "Num"))
	assert.NoError(t, checker.CheckTypeCompatibility("Int", "Scalar"))
	assert.NoError(t, checker.CheckTypeCompatibility("Int", "Any"))
	assert.NoError(t, checker.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Num]"))

	// Test incompatible types
	assert.Error(t, checker.CheckTypeCompatibility("Int", "Str"))
	assert.Error(t, checker.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Str]"))
	assert.Error(t, checker.CheckTypeCompatibility("Int", "ArrayRef[Int]"))
}

// TestCheckAnnotation tests checking of individual type annotations
func TestCheckAnnotation(t *testing.T) {
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
	validVarAnnotation := &TypeAnnotation{
		AnnotatedItem: "$var",
		TypeExpression: &TypeExpression{
			Name: "Int",
			Pos:  Position{Line: 1, Column: 5},
		},
		Pos:  Position{Line: 1, Column: 5},
		Kind: VarAnnotation,
	}

	// Check the valid annotation
	assert.NoError(t, checker.collectTypeAnnotation(validVarAnnotation))
	assert.NoError(t, checker.checkTypeAnnotation(validVarAnnotation))

	// Verify that the type annotation was recorded
	typeStr, ok := checker.GetAnnotatedType("$var")
	assert.True(t, ok)
	assert.Equal(t, "Int", typeStr)

	// Create an invalid type annotation
	invalidTypeAnnotation := &TypeAnnotation{
		AnnotatedItem: "$invalid",
		TypeExpression: &TypeExpression{
			Name: "invalidType", // lowercase type name is invalid
			Pos:  Position{Line: 2, Column: 5},
		},
		Pos:  Position{Line: 2, Column: 5},
		Kind: VarAnnotation,
	}

	// Check the invalid annotation
	assert.NoError(t, checker.collectTypeAnnotation(invalidTypeAnnotation))
	assert.Error(t, checker.checkTypeAnnotation(invalidTypeAnnotation))

	// Verify that checking for a type annotation with a nil type expression fails
	nilTypeAnnotation := &TypeAnnotation{
		AnnotatedItem:  "$nil",
		TypeExpression: nil,
		Pos:            Position{Line: 3, Column: 5},
		Kind:           VarAnnotation,
	}

	assert.Error(t, checker.collectTypeAnnotation(nilTypeAnnotation))
}

// TestCheckAssignment tests checking of type compatibilities for assignments
func TestCheckAssignment(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test compatible assignments
	assert.NoError(t, checker.CheckAssignment("Int", "Int", Position{Line: 1, Column: 1}))
	assert.NoError(t, checker.CheckAssignment("Int", "Num", Position{Line: 1, Column: 1}))
	assert.NoError(t, checker.CheckAssignment("ArrayRef[Int]", "ArrayRef[Num]", Position{Line: 1, Column: 1}))

	// Test incompatible assignments
	assert.Error(t, checker.CheckAssignment("Int", "Str", Position{Line: 1, Column: 1}))
	assert.Error(t, checker.CheckAssignment("ArrayRef[Int]", "ArrayRef[Str]", Position{Line: 1, Column: 1}))
}

// TestExpressionTypeInference tests inference of expression types
func TestExpressionTypeInference(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Add some known variable types
	checker.VariableTypes["$intVar"] = "Int"
	checker.VariableTypes["$strVar"] = "Str"
	checker.VariableTypes["@array"] = "Array"
	checker.VariableTypes["%hash"] = "Hash"

	// Test numeric literals
	assert.Equal(t, "Int", checker.inferExpressionType("42"))
	assert.Equal(t, "Float", checker.inferExpressionType("3.14"))
	assert.Equal(t, "Int", checker.inferExpressionType("-10"))

	// Test string literals
	assert.Equal(t, "Str", checker.inferExpressionType("'hello'"))
	assert.Equal(t, "Str", checker.inferExpressionType("\"world\""))
	assert.Equal(t, "Str", checker.inferExpressionType("`backtick`"))

	// Test boolean literals
	assert.Equal(t, "Bool", checker.inferExpressionType("1"))  // Special case for Perl
	assert.Equal(t, "Bool", checker.inferExpressionType("0"))  // Special case for Perl
	assert.Equal(t, "Bool", checker.inferExpressionType("True"))
	assert.Equal(t, "Bool", checker.inferExpressionType("False"))

	// Test undef
	assert.Equal(t, "Undef", checker.inferExpressionType("undef"))

	// Test variables
	assert.Equal(t, "Int", checker.inferExpressionType("$intVar"))
	assert.Equal(t, "Str", checker.inferExpressionType("$strVar"))
	assert.Equal(t, "Array", checker.inferExpressionType("@array"))
	assert.Equal(t, "Hash", checker.inferExpressionType("%hash"))

	// Test unknown variable
	assert.Equal(t, "", checker.inferExpressionType("$unknownVar"))

	// Test array references
	assert.Equal(t, "ArrayRef", checker.inferExpressionType("[1, 2, 3]"))
	assert.Equal(t, "ArrayRef", checker.inferExpressionType("[]"))

	// Test hash references
	assert.Equal(t, "HashRef", checker.inferExpressionType("{key => 'value'}"))
	assert.Equal(t, "HashRef", checker.inferExpressionType("{}"))

	// Test complex expressions
	assert.Equal(t, "", checker.inferExpressionType("$a + $b"))  // Would need real expression parsing
}

// TestCollectTypeAnnotations tests collecting and accessing type annotations
func TestCollectTypeAnnotations(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Create annotations for variables, parameters, and returns
	varAnnotation := &TypeAnnotation{
		AnnotatedItem: "$var",
		TypeExpression: &TypeExpression{
			Name: "Int",
			Pos:  Position{Line: 1, Column: 5},
		},
		Pos:  Position{Line: 1, Column: 5},
		Kind: VarAnnotation,
	}

	attrAnnotation := &TypeAnnotation{
		AnnotatedItem: "$attr",
		TypeExpression: &TypeExpression{
			Name: "Str",
			Pos:  Position{Line: 2, Column: 5},
		},
		Pos:  Position{Line: 2, Column: 5},
		Kind: AttrAnnotation,
	}

	// Collect the annotations
	assert.NoError(t, checker.collectTypeAnnotation(varAnnotation))
	assert.NoError(t, checker.collectTypeAnnotation(attrAnnotation))

	// Verify that the annotations were recorded
	typeStr, ok := checker.GetAnnotatedType("$var")
	assert.True(t, ok)
	assert.Equal(t, "Int", typeStr)

	typeStr, ok = checker.GetVariableType("$var")
	assert.True(t, ok)
	assert.Equal(t, "Int", typeStr)

	typeStr, ok = checker.GetAnnotatedType("$attr")
	assert.True(t, ok)
	assert.Equal(t, "Str", typeStr)

	typeStr, ok = checker.GetVariableType("$attr")
	assert.True(t, ok)
	assert.Equal(t, "Str", typeStr)
}

// TestFormatTypeError tests formatting of type errors
func TestFormatTypeError(t *testing.T) {
	err := &TypeError{
		Message: "Type error message",
	}
	pos := Position{Line: 10, Column: 5}
	path := "/path/to/file.pl"

	formatted := FormatTypeError(err, pos, path)
	assert.Equal(t, "/path/to/file.pl:10:5: Type error message", formatted)
}

// TestTypeCheckIntegration tests the integration of the TypeCheck with the TypeChecker
func TestTypeCheckIntegration(t *testing.T) {
	// Create a TypeCheck instance
	tc, err := NewTypeCheck()
	require.NoError(t, err)
	require.NotNil(t, tc)

	// Verify it was created properly
	assert.NotNil(t, tc.Parser)
	assert.NotNil(t, tc.TypeStore)
	assert.NotNil(t, tc.TypeHierarchy)
}

// TypeError is a helper type for testing
type TypeError struct {
	Message string
}

func (e *TypeError) Error() string {
	return e.Message
}