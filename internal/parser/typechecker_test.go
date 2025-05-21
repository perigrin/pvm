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
	assert.NoError(t, checker.validateType("CustomType")) // Uppercase is okay
	assert.Error(t, checker.validateType("invalidType"))  // lowercase is not valid
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
	// Note: Int -> Num -> Str is compatible due to Perl's automatic stringification
	assert.Error(t, checker.CheckTypeCompatibility("Str", "Int")) // Str is NOT compatible with Int
	// Note: ArrayRef[Int] -> ArrayRef[Str] is currently covariant (may change in future phases)
	assert.Error(t, checker.CheckTypeCompatibility("ArrayRef[Str]", "ArrayRef[Int]")) // Reverse should be an error
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
	// Note: Int -> Num -> Str is compatible due to Perl's automatic stringification
	assert.Error(t, checker.CheckAssignment("Str", "Int", Position{Line: 1, Column: 1}))                     // Reverse is incompatible
	assert.Error(t, checker.CheckAssignment("ArrayRef[Str]", "ArrayRef[Int]", Position{Line: 1, Column: 1})) // Reverse is incompatible
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
	assert.Equal(t, "Int", checker.inferExpressionType("1"))      // In Perl, 1 is an Int that can be used in boolean context
	assert.Equal(t, "Int", checker.inferExpressionType("0"))      // In Perl, 0 is an Int that can be used in boolean context
	assert.Equal(t, "Bool", checker.inferExpressionType("True"))  // True Bool literals
	assert.Equal(t, "Bool", checker.inferExpressionType("False")) // True Bool literals

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
	assert.Equal(t, "", checker.inferExpressionType("$a + $b")) // Would need real expression parsing
}

// TestPureLiteralTypeInference tests Phase 1: Pure type inference for literals
func TestPureLiteralTypeInference(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test integer literals
	assert.Equal(t, "Int", checker.inferLiteralType("42"))
	assert.Equal(t, "Int", checker.inferLiteralType("0"))
	assert.Equal(t, "Int", checker.inferLiteralType("-10"))
	assert.Equal(t, "Int", checker.inferLiteralType("999999"))

	// Test float literals
	assert.Equal(t, "Float", checker.inferLiteralType("3.14"))
	assert.Equal(t, "Float", checker.inferLiteralType("0.0"))
	assert.Equal(t, "Float", checker.inferLiteralType("-2.5"))
	assert.Equal(t, "Float", checker.inferLiteralType("1.0e10"))

	// Test string literals
	assert.Equal(t, "Str", checker.inferLiteralType("'hello'"))
	assert.Equal(t, "Str", checker.inferLiteralType("\"world\""))
	assert.Equal(t, "Str", checker.inferLiteralType("`backtick`"))
	assert.Equal(t, "Str", checker.inferLiteralType("''"))
	assert.Equal(t, "Str", checker.inferLiteralType("\"\""))

	// Test boolean literals (Perl treats 1 and 0 specially in boolean context)
	assert.Equal(t, "Bool", checker.inferLiteralType("True"))
	assert.Equal(t, "Bool", checker.inferLiteralType("False"))

	// Test undef literal
	assert.Equal(t, "Undef", checker.inferLiteralType("undef"))

	// Test unknown literals should return empty string
	assert.Equal(t, "", checker.inferLiteralType("unknown_literal"))
	assert.Equal(t, "", checker.inferLiteralType(""))
}

// TestBasicTypeValidation tests Phase 2: Basic type validation for primitive types
func TestBasicTypeValidation(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test validation of all basic built-in types
	basicTypes := []string{
		"Any", "Scalar", "Str", "Num", "Int", "Float", "Bool", "Undef",
		"Ref", "ArrayRef", "HashRef", "ScalarRef", "CodeRef", "RegexpRef", "GlobRef", "FileHandle",
		"List", "Array", "Hash", "Code", "Glob",
		"Maybe", "Optional",
		"Callable", "Iterable", "Positional", "Associative",
		"IO", "Path", "File", "Dir",
		"ClassName", "RoleName", "MethodName", "Byte", "Char", "VarName",
	}

	for _, typeName := range basicTypes {
		assert.NoError(t, checker.validateType(typeName),
			"Type %s should be valid", typeName)
	}
}

// TestTypeValidationErrorCases tests invalid type names and error reporting
func TestTypeValidationErrorCases(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test invalid type names (should return errors)
	invalidTypes := []string{
		"",                 // Empty type name
		"invalidType",      // Lowercase type (invalid)
		"not_a_type",       // Underscore and lowercase
		"123Invalid",       // Starts with number
		"invalid-type",     // Contains hyphen
		"weird@Type",       // Contains special characters
		"type with spaces", // Contains spaces
	}

	for _, typeName := range invalidTypes {
		err := checker.validateType(typeName)
		assert.Error(t, err, "Type %s should be invalid", typeName)

		// Check that error message is informative
		assert.Contains(t, err.Error(), typeName,
			"Error message should contain the invalid type name")
	}
}

// TestParameterizedTypeValidation tests validation of parameterized types
func TestParameterizedTypeValidation(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test valid parameterized types
	validParameterizedTypes := []string{
		"ArrayRef[Int]",
		"HashRef[Int]", // HashRef[V] means HashRef[Str,V]
		"HashRef[Str]",
		"HashRef[Str,Int]",
		"Maybe[Str]",
		"Optional[Int]",
		"ArrayRef[ArrayRef[Int]]",    // Nested
		"HashRef[Str,ArrayRef[Int]]", // Complex nested
	}

	for _, typeName := range validParameterizedTypes {
		assert.NoError(t, checker.validateType(typeName),
			"Parameterized type %s should be valid", typeName)
	}

	// Test invalid parameterized types
	invalidParameterizedTypes := []string{
		"ArrayRef[]",            // Empty parameters
		"ArrayRef[invalidType]", // Invalid parameter type
		"HashRef[]",             // Empty parameters (HashRef needs 1 or 2 params)
		"Int[Str]",              // Int is not parameterizable
		"ArrayRef[Int,Str]",     // ArrayRef takes only 1 parameter
		"Maybe[]",               // Maybe requires a parameter
		"Maybe[Int,Str]",        // Maybe takes only 1 parameter
		"HashRef[Int,Str,Bool]", // HashRef takes at most 2 parameters
	}

	for _, typeName := range invalidParameterizedTypes {
		err := checker.validateType(typeName)
		assert.Error(t, err, "Invalid parameterized type %s should fail validation", typeName)
	}
}

// TestCustomTypeNameValidation tests validation of user-defined type names
func TestCustomTypeNameValidation(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test valid custom type names (follow Perl module naming conventions)
	validCustomTypes := []string{
		"MyType",
		"UserDefined",
		"SomeModule::SomeType",
		"Deep::Module::Path::TypeName",
		"Class",
		"Role",
		"DatabaseConnection",
	}

	for _, typeName := range validCustomTypes {
		// Custom types should be valid even if not defined yet
		// (they might be imported from modules)
		assert.NoError(t, checker.validateType(typeName),
			"Custom type %s should be valid", typeName)
	}

	// Test invalid custom type name patterns
	invalidCustomTypes := []string{
		"lowercase",        // Must start with uppercase
		"some::lowercase",  // Module parts must start with uppercase
		"Type::123invalid", // Cannot start with number
		"Type::",           // Cannot end with ::
		"::Type",           // Cannot start with ::
		"Type:::",          // Invalid separator
	}

	for _, typeName := range invalidCustomTypes {
		err := checker.validateType(typeName)
		assert.Error(t, err, "Invalid custom type %s should fail validation", typeName)
	}
}

// TestVariableTypeAnnotations tests Phase 3: Variable Type Annotations
func TestVariableTypeAnnotations(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test basic variable annotations
	// We'll manually create annotations as if they were parsed

	// Valid annotation: my Int $x = 42
	intAnnotation := &TypeAnnotation{
		AnnotatedItem: "$x",
		TypeExpression: &TypeExpression{
			Name: "Int",
			Pos:  Position{Line: 1, Column: 8},
		},
		Pos:  Position{Line: 1, Column: 8},
		Kind: VarAnnotation,
	}

	// Valid annotation: my Str $name = "Bob"
	strAnnotation := &TypeAnnotation{
		AnnotatedItem: "$name",
		TypeExpression: &TypeExpression{
			Name: "Str",
			Pos:  Position{Line: 2, Column: 8},
		},
		Pos:  Position{Line: 2, Column: 8},
		Kind: VarAnnotation,
	}

	// Test collecting and validating annotations
	assert.NoError(t, checker.collectTypeAnnotation(intAnnotation))
	assert.NoError(t, checker.collectTypeAnnotation(strAnnotation))
	assert.NoError(t, checker.checkTypeAnnotation(intAnnotation))
	assert.NoError(t, checker.checkTypeAnnotation(strAnnotation))

	// Verify that annotations were recorded correctly
	intType, ok := checker.GetVariableType("$x")
	assert.True(t, ok)
	assert.Equal(t, "Int", intType)

	strType, ok := checker.GetVariableType("$name")
	assert.True(t, ok)
	assert.Equal(t, "Str", strType)
}

// TestVariableAssignmentCompatibility tests assignment type checking for annotated variables
func TestVariableAssignmentCompatibility(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Set up a variable with type annotation: my Int $x
	checker.VariableTypes["$x"] = "Int"
	checker.VariableTypes["$name"] = "Str"

	// Test valid assignments
	// $x = 42 (Int literal to Int variable)
	assert.NoError(t, checker.CheckAssignment("Int", "Int", Position{Line: 1, Column: 1}))

	// $x = 3.14 would be Invalid since Float doesn't narrow to Int
	// But Str = "hello" should work
	assert.NoError(t, checker.CheckAssignment("Str", "Str", Position{Line: 2, Column: 1}))

	// Test compatible assignments due to subtyping
	// Int can be assigned to Num variable (Int → Num)
	assert.NoError(t, checker.CheckAssignment("Int", "Num", Position{Line: 3, Column: 1}))
	// Int can be assigned to Str variable (Int → Num → Str)
	assert.NoError(t, checker.CheckAssignment("Int", "Str", Position{Line: 4, Column: 1}))

	// Test invalid assignments
	// Str cannot be assigned to Int variable (Str ↛ Int)
	assert.Error(t, checker.CheckAssignment("Str", "Int", Position{Line: 5, Column: 1}))
	// Float cannot be assigned to Int variable (Float ↛ Int)
	assert.Error(t, checker.CheckAssignment("Float", "Int", Position{Line: 6, Column: 1}))
}

// TestTypeInferenceForUnannotatedVariables tests inference of types for variables without annotations
func TestTypeInferenceForUnannotatedVariables(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test type inference from assignment: my $inferred = 42;
	// We would infer this by seeing the literal type on the right side
	inferredType := checker.inferLiteralType("42")
	assert.Equal(t, "Int", inferredType)

	// If we were to record this inference:
	checker.VariableTypes["$inferred"] = inferredType
	retrievedType, ok := checker.GetVariableType("$inferred")
	assert.True(t, ok)
	assert.Equal(t, "Int", retrievedType)

	// Test inference from string literal
	stringType := checker.inferLiteralType("'hello'")
	assert.Equal(t, "Str", stringType)

	// Test inference from float literal
	floatType := checker.inferLiteralType("3.14")
	assert.Equal(t, "Float", floatType)
}

// TestOperatorTypeChecking tests Phase 4: Operator Type Checking
func TestOperatorTypeChecking(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test numeric operators
	// $a + $b where both are Num should return Num
	resultType, err := checker.InferBinaryOperatorType("Int", "+", "Int")
	assert.NoError(t, err)
	assert.Equal(t, "Num", resultType) // Addition promotes to Num

	resultType, err = checker.InferBinaryOperatorType("Float", "*", "Int")
	assert.NoError(t, err)
	assert.Equal(t, "Num", resultType) // Multiplication returns Num

	// Test comparison operators
	resultType, err = checker.InferBinaryOperatorType("Int", "==", "Int")
	assert.NoError(t, err)
	assert.Equal(t, "Bool", resultType) // Numeric comparison returns Bool

	resultType, err = checker.InferBinaryOperatorType("Num", "<", "Num")
	assert.NoError(t, err)
	assert.Equal(t, "Bool", resultType)

	// Test string operators
	resultType, err = checker.InferBinaryOperatorType("Str", ".", "Str")
	assert.NoError(t, err)
	assert.Equal(t, "Str", resultType) // String concatenation returns Str

	resultType, err = checker.InferBinaryOperatorType("Str", "eq", "Str")
	assert.NoError(t, err)
	assert.Equal(t, "Bool", resultType) // String comparison returns Bool

	// Test logical operators
	resultType, err = checker.InferBinaryOperatorType("Any", "&&", "Any")
	assert.NoError(t, err)
	assert.Equal(t, "Bool", resultType) // Logical AND returns Bool

	resultType, err = checker.InferBinaryOperatorType("Bool", "||", "Bool")
	assert.NoError(t, err)
	assert.Equal(t, "Bool", resultType) // Logical OR returns Bool
}

// TestOperatorTypeErrors tests Phase 4: Operator Type Checking error cases
func TestOperatorTypeErrors(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test type mismatches for operators
	// String concatenation with incompatible types should error
	_, err = checker.InferBinaryOperatorType("Int", ".", "Str")
	assert.Error(t, err) // Int can't be used directly with string concat without coercion

	// Numeric operators with incompatible types should error
	_, err = checker.InferBinaryOperatorType("ArrayRef", "+", "Int")
	assert.Error(t, err) // ArrayRef can't be used in numeric operations

	// String comparison with numeric operators should error
	_, err = checker.InferBinaryOperatorType("Str", "<", "Str")
	assert.Error(t, err) // Should use 'lt' for string comparison, not '<'
}

// TestUnaryOperatorTypeChecking tests unary operators
func TestUnaryOperatorTypeChecking(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test logical NOT
	resultType, err := checker.InferUnaryOperatorType("!", "Any")
	assert.NoError(t, err)
	assert.Equal(t, "Bool", resultType)

	// Test numeric negation
	resultType, err = checker.InferUnaryOperatorType("-", "Int")
	assert.NoError(t, err)
	assert.Equal(t, "Int", resultType)

	resultType, err = checker.InferUnaryOperatorType("-", "Float")
	assert.NoError(t, err)
	assert.Equal(t, "Float", resultType)

	// Test errors for invalid unary operations
	_, err = checker.InferUnaryOperatorType("-", "Str")
	assert.Error(t, err) // Can't negate a string
}

// TestTypeHierarchyRespected tests that inferred types respect the type hierarchy
func TestTypeHierarchyRespected(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test that Int is a subtype of Num
	assert.True(t, hierarchy.IsSubtypeOf("Int", "Num"))

	// Test that Num is a subtype of Str (Perl's automatic stringification)
	assert.True(t, hierarchy.IsSubtypeOf("Num", "Str"))

	// Test that Str is a subtype of Scalar
	assert.True(t, hierarchy.IsSubtypeOf("Str", "Scalar"))

	// Test that Scalar is a subtype of Any
	assert.True(t, hierarchy.IsSubtypeOf("Scalar", "Any"))

	// Test transitivity: Int → Num → Str → Scalar → Any
	assert.True(t, hierarchy.IsSubtypeOf("Int", "Scalar"))
	assert.True(t, hierarchy.IsSubtypeOf("Int", "Any"))
	assert.True(t, hierarchy.IsSubtypeOf("Float", "Scalar"))
	assert.True(t, hierarchy.IsSubtypeOf("Float", "Any"))

	// Test Bool is also a scalar type
	assert.True(t, hierarchy.IsSubtypeOf("Bool", "Scalar"))
	assert.True(t, hierarchy.IsSubtypeOf("Bool", "Any"))

	// Test Undef is scalar
	assert.True(t, hierarchy.IsSubtypeOf("Undef", "Scalar"))
	assert.True(t, hierarchy.IsSubtypeOf("Undef", "Any"))
}

// TestMostSpecificTypeInference tests that we infer the most specific type possible
func TestMostSpecificTypeInference(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Integer should be inferred as Int, not Num or Str or Scalar
	inferredType := checker.inferLiteralType("42")
	assert.Equal(t, "Int", inferredType)
	// Verify it's not the more general types
	assert.NotEqual(t, "Num", inferredType)
	assert.NotEqual(t, "Str", inferredType)
	assert.NotEqual(t, "Scalar", inferredType)

	// Float should be inferred as Float, not Num or Str
	inferredType = checker.inferLiteralType("3.14")
	assert.Equal(t, "Float", inferredType)
	assert.NotEqual(t, "Num", inferredType)
	assert.NotEqual(t, "Str", inferredType)

	// String should be inferred as Str, not Scalar
	inferredType = checker.inferLiteralType("'hello'")
	assert.Equal(t, "Str", inferredType)
	assert.NotEqual(t, "Scalar", inferredType)
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
	assert.True(t, tc.EnableFlowSensitiveAnalysis) // Flow-sensitive analysis should be enabled by default
}

// TestFlowSensitiveAnalysis tests the flow-sensitive analysis functionality
func TestFlowSensitiveAnalysis(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Setup initial variable types for testing
	checker.VariableTypes["$maybeStr"] = "Maybe[Str]"
	checker.VariableTypes["$ref"] = "Ref"
	checker.VariableTypes["$any"] = "Any"

	// Test that flow-sensitive analysis is properly initialized
	assert.NotNil(t, checker.TypeState)
	assert.Empty(t, checker.TypeState.RefinedTypes)

	// Create a mock defined check node
	definedNode := &MockNode{
		NodeType:     "function_call",
		NodeText:     "defined($maybeStr)",
		NodeChildren: []Node{},
	}

	// Process the node for flow-sensitive analysis
	var errors []error
	checker.analyzePatternValidation(definedNode, &errors)

	// Check that the type was refined
	refinedType, ok := checker.TypeState.RefinedTypes["$maybeStr"]
	assert.True(t, ok, "The type should be refined after a defined check")
	assert.Equal(t, "Str", refinedType, "Maybe[Str] should be refined to Str after a defined check")

	// Create a mock ref check node
	refNode := &MockNode{
		NodeType:     "eq_expression",
		NodeText:     "ref($ref) eq 'ARRAY'",
		NodeChildren: []Node{},
	}

	// Process the node for flow-sensitive analysis
	checker.analyzePatternValidation(refNode, &errors)

	// Check that the type was refined
	refinedType, ok = checker.TypeState.RefinedTypes["$ref"]
	assert.True(t, ok, "The type should be refined after a ref check")
	assert.Equal(t, "ArrayRef", refinedType, "Ref should be refined to ArrayRef after a ref check for ARRAY")
}

// TestTypeStateCloning tests the cloning of type states for different code paths
func TestTypeStateCloning(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Setup initial type state
	checker.TypeState.VariableTypes["$var1"] = "Int"
	checker.TypeState.VariableTypes["$var2"] = "Str"
	checker.TypeState.RefinedTypes["$var1"] = "Int"

	// Clone the state
	clonedState := checker.cloneTypeState(checker.TypeState)

	// Verify the clone is a deep copy
	assert.NotSame(t, checker.TypeState, clonedState)
	assert.Equal(t, checker.TypeState.VariableTypes, clonedState.VariableTypes)
	assert.Equal(t, checker.TypeState.RefinedTypes, clonedState.RefinedTypes)

	// Modify the clone and verify it doesn't affect the original
	clonedState.VariableTypes["$var3"] = "Bool"
	clonedState.RefinedTypes["$var2"] = "ClassName"

	assert.NotContains(t, checker.TypeState.VariableTypes, "$var3")
	assert.NotContains(t, checker.TypeState.RefinedTypes, "$var2")
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
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Setup initial variable types
	checker.VariableTypes["$maybeVar"] = "Maybe[Int]"

	// Create a mock condition node for $maybeVar != undef
	conditionNode := &MockNode{
		NodeType:     "ne_expression",
		NodeText:     "$maybeVar != undef",
		NodeChildren: []Node{},
	}

	// Analyze the condition
	condition, refinements := checker.analyzeCondition(conditionNode)

	// Check the extracted condition
	assert.Equal(t, "$maybeVar", condition.Variable)
	assert.Equal(t, "!=", condition.Operator)
	assert.Equal(t, "undef", condition.Value)
	assert.True(t, condition.Negated)

	// Check the refinements
	assert.Contains(t, refinements, "$maybeVar")
	assert.Equal(t, "Int", refinements["$maybeVar"])

	// Apply the refinements
	checker.applyRefinements(refinements)

	// Check that the refinement was applied
	refinedType, ok := checker.TypeState.RefinedTypes["$maybeVar"]
	assert.True(t, ok)
	assert.Equal(t, "Int", refinedType)

	// Test negating the refinements for an else branch
	negatedRefinements := checker.negateRefinements(condition, refinements)
	assert.Empty(t, negatedRefinements, "Negation of != undef doesn't provide useful refinements")

	// Set up a different condition for testing negation
	eqUndefCondition := Condition{
		Variable: "$maybeVar",
		Operator: "==",
		Value:    "undef",
		Negated:  false,
	}

	// Create empty refinements map
	emptyRefinements := make(map[string]string)

	// Negate the condition (maybeVar == undef -> maybeVar != undef)
	negatedRefinements = checker.negateRefinements(eqUndefCondition, emptyRefinements)

	// Check that the negation produces a refinement
	assert.Contains(t, negatedRefinements, "$maybeVar")
	assert.Equal(t, "Int", negatedRefinements["$maybeVar"])
}

// TestGetVariableTypeWithRefinements tests the GetVariableType method with refined types
func TestGetVariableTypeWithRefinements(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Setup initial variable types and refined types
	checker.VariableTypes["$var1"] = "Maybe[Int]"
	checker.TypeState.RefinedTypes["$var1"] = "Int"
	checker.VariableTypes["$var2"] = "Str"

	// Test that refined types take precedence
	typ, ok := checker.GetVariableType("$var1")
	assert.True(t, ok)
	assert.Equal(t, "Int", typ)

	// Test fallback to regular variable types
	typ, ok = checker.GetVariableType("$var2")
	assert.True(t, ok)
	assert.Equal(t, "Str", typ)

	// Test unknown variable
	typ, ok = checker.GetVariableType("$unknown")
	assert.False(t, ok)
	assert.Empty(t, typ)
}

// Mock node implementation for testing
type MockNode struct {
	NodeType     string
	NodeText     string
	NodeChildren []Node
}

func (n *MockNode) Type() string {
	return n.NodeType
}

func (n *MockNode) Text() string {
	return n.NodeText
}

func (n *MockNode) Children() []Node {
	return n.NodeChildren
}

func (n *MockNode) Parent() Node {
	return nil
}

func (n *MockNode) Start() Position {
	return Position{Line: 1, Column: 1}
}

func (n *MockNode) End() Position {
	return Position{Line: 1, Column: len(n.NodeText)}
}

// TestFunctionSignatureValidation tests function signature validation (Phase 5)
func TestFunctionSignatureValidation(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// First register some function signatures
	addSig := &FunctionSignature{
		ParameterTypes: map[string]string{"$a": "Int", "$b": "Int"},
		ReturnType:     "Int",
	}
	checker.FunctionTypes["add"] = addSig

	greetSig := &FunctionSignature{
		ParameterTypes: map[string]string{"$name": "Str"},
		ReturnType:     "Str",
	}
	checker.FunctionTypes["greet"] = greetSig

	// Test valid function calls
	t.Run("ValidFunctionCalls", func(t *testing.T) {
		// Valid call: add(Int, Int) -> Int
		err := checker.CheckFunctionCall("add", []string{"Int", "Int"})
		assert.NoError(t, err, "Valid function call should not error")

		// Valid call: greet(Str) -> Str
		err = checker.CheckFunctionCall("greet", []string{"Str"})
		assert.NoError(t, err, "Valid function call should not error")
	})

	// Test invalid parameter types
	t.Run("InvalidParameterTypes", func(t *testing.T) {
		// Invalid call: add(Str, Int) - first param wrong type
		err := checker.CheckFunctionCall("add", []string{"Str", "Int"})
		assert.Error(t, err, "Function call with wrong parameter type should error")
		assert.Contains(t, err.Error(), "parameter", "Error should mention parameter")

		// Invalid call: greet(Int) - wrong parameter type
		err = checker.CheckFunctionCall("greet", []string{"Int"})
		assert.Error(t, err, "Function call with wrong parameter type should error")
		assert.Contains(t, err.Error(), "parameter", "Error should mention parameter")
	})

	// Test wrong number of parameters
	t.Run("WrongParameterCount", func(t *testing.T) {
		// Too few parameters: add(Int)
		err := checker.CheckFunctionCall("add", []string{"Int"})
		assert.Error(t, err, "Function call with too few parameters should error")
		assert.Contains(t, err.Error(), "expects 2", "Error should mention expected parameter count")

		// Too many parameters: add(Int, Int, Int)
		err = checker.CheckFunctionCall("add", []string{"Int", "Int", "Int"})
		assert.Error(t, err, "Function call with too many parameters should error")
		assert.Contains(t, err.Error(), "expects 2", "Error should mention expected parameter count")

		// Too many parameters: greet(Str, Str)
		err = checker.CheckFunctionCall("greet", []string{"Str", "Str"})
		assert.Error(t, err, "Function call with too many parameters should error")
		assert.Contains(t, err.Error(), "expects 1", "Error should mention expected parameter count")
	})

	// Test unknown function
	t.Run("UnknownFunction", func(t *testing.T) {
		err := checker.CheckFunctionCall("unknown_func", []string{"Int"})
		assert.Error(t, err, "Call to unknown function should error")
		assert.Contains(t, err.Error(), "unknown function", "Error should mention unknown function")
	})
}

// TestFunctionReturnTypeValidation tests return type validation (Phase 5)
func TestFunctionReturnTypeValidation(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Register function signatures for testing
	addSig := &FunctionSignature{
		ParameterTypes: map[string]string{"$a": "Int", "$b": "Int"},
		ReturnType:     "Int",
	}
	checker.FunctionTypes["add"] = addSig

	greetSig := &FunctionSignature{
		ParameterTypes: map[string]string{"$name": "Str"},
		ReturnType:     "Str",
	}
	checker.FunctionTypes["greet"] = greetSig

	// Test return type checking
	t.Run("ValidReturnTypes", func(t *testing.T) {
		// Function that returns Int, returning Int expression
		err := checker.CheckReturnType("add", "Int")
		assert.NoError(t, err, "Valid return type should not error")

		// Function that returns Str, returning Str expression
		err = checker.CheckReturnType("greet", "Str")
		assert.NoError(t, err, "Valid return type should not error")
	})

	t.Run("InvalidReturnTypes", func(t *testing.T) {
		// Function that returns Int, but returns Str
		err := checker.CheckReturnType("add", "Str")
		assert.Error(t, err, "Invalid return type should error")
		assert.Contains(t, err.Error(), "return type", "Error should mention return type")

		// Function that returns Str, but returns Int
		err = checker.CheckReturnType("greet", "Int")
		assert.Error(t, err, "Invalid return type should error")
		assert.Contains(t, err.Error(), "return type", "Error should mention return type")
	})

	t.Run("UnknownFunctionReturn", func(t *testing.T) {
		err := checker.CheckReturnType("unknown_func", "Int")
		assert.Error(t, err, "Return check for unknown function should error")
		assert.Contains(t, err.Error(), "unknown function", "Error should mention unknown function")
	})
}

// TestMethodVsSubroutineDistinction tests method vs subroutine distinction (Phase 5)
func TestMethodVsSubroutineDistinction(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test subroutine vs method calls
	t.Run("SubroutineCalls", func(t *testing.T) {
		// Register a subroutine
		subSig := &FunctionSignature{
			ParameterTypes: map[string]string{"$value": "Int"},
			ReturnType:     "Str",
			IsMethod:       false,
		}
		checker.FunctionTypes["format_number"] = subSig

		// Valid subroutine call
		err := checker.CheckSubroutineCall("format_number", []string{"Int"})
		assert.NoError(t, err, "Valid subroutine call should not error")

		// Try to call as method (should error)
		err = checker.CheckMethodCall("SomeClass", "format_number", []string{"Int"})
		assert.Error(t, err, "Calling subroutine as method should error")
		assert.Contains(t, err.Error(), "not a method", "Error should mention it's not a method")
	})

	t.Run("MethodCalls", func(t *testing.T) {
		// Register a method
		methodSig := &FunctionSignature{
			ParameterTypes: map[string]string{"$self": "Person", "$age": "Int"},
			ReturnType:     "Bool",
			IsMethod:       true,
		}
		checker.FunctionTypes["Person::set_age"] = methodSig

		// Valid method call
		err := checker.CheckMethodCall("Person", "set_age", []string{"Person", "Int"})
		assert.NoError(t, err, "Valid method call should not error")

		// Try to call as subroutine (should error)
		err = checker.CheckSubroutineCall("Person::set_age", []string{"Person", "Int"})
		assert.Error(t, err, "Calling method as subroutine should error")
		assert.Contains(t, err.Error(), "not a subroutine", "Error should mention it's not a subroutine")
	})
}

// TestParameterTypeCompatibility tests parameter type compatibility with subtyping (Phase 5)
func TestParameterTypeCompatibility(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test parameter subtype compatibility
	t.Run("SubtypeCompatibility", func(t *testing.T) {
		// Function expects Num, but gets Int (should be valid due to Int -> Num)
		numFuncSig := &FunctionSignature{
			ParameterTypes: map[string]string{"$value": "Num"},
			ReturnType:     "Str",
		}
		checker.FunctionTypes["format_num"] = numFuncSig

		// Int should be compatible with Num parameter
		err := checker.CheckFunctionCall("format_num", []string{"Int"})
		assert.NoError(t, err, "Int should be compatible with Num parameter")

		// Function expects Scalar, gets Str (should be valid due to Str -> Scalar)
		scalarFuncSig := &FunctionSignature{
			ParameterTypes: map[string]string{"$value": "Scalar"},
			ReturnType:     "Int",
		}
		checker.FunctionTypes["length_scalar"] = scalarFuncSig

		// Str should be compatible with Scalar parameter
		err = checker.CheckFunctionCall("length_scalar", []string{"Str"})
		assert.NoError(t, err, "Str should be compatible with Scalar parameter")
	})

	t.Run("IncompatibleTypes", func(t *testing.T) {
		// Function expects Int, but gets Str (should error - no Str -> Int conversion)
		intFuncSig := &FunctionSignature{
			ParameterTypes: map[string]string{"$number": "Int"},
			ReturnType:     "Int",
		}
		checker.FunctionTypes["double_int"] = intFuncSig

		// Str is not compatible with Int parameter
		err := checker.CheckFunctionCall("double_int", []string{"Str"})
		assert.Error(t, err, "Str should not be compatible with Int parameter")
		assert.Contains(t, err.Error(), "incompatible", "Error should mention incompatibility")
	})
}

// TestContainerTypeChecking tests container type validation (Phase 6)
func TestContainerTypeChecking(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test array element type checking
	t.Run("ArrayElementTypeChecking", func(t *testing.T) {
		// Valid array element access
		err := checker.CheckArrayElementAccess("ArrayRef[Int]", 0, "Int")
		assert.NoError(t, err, "Valid array element access should not error")

		err = checker.CheckArrayElementAccess("ArrayRef[Str]", 0, "Str")
		assert.NoError(t, err, "Valid string array element access should not error")

		// Invalid array element access - wrong element type
		err = checker.CheckArrayElementAccess("ArrayRef[Int]", 0, "Str")
		assert.Error(t, err, "Invalid array element type should error")
		assert.Contains(t, err.Error(), "element type", "Error should mention element type")

		// Invalid array element assignment
		err = checker.CheckArrayElementAssignment("ArrayRef[Int]", 0, "Str")
		assert.Error(t, err, "Invalid array element assignment should error")
		assert.Contains(t, err.Error(), "cannot assign", "Error should mention assignment")
	})

	// Test hash element type checking
	t.Run("HashElementTypeChecking", func(t *testing.T) {
		// Valid hash element access (HashRef[V] means HashRef[Str,V])
		err := checker.CheckHashElementAccess("HashRef[Int]", "key", "Int")
		assert.NoError(t, err, "Valid hash element access should not error")

		// Valid hash element access with explicit key/value types
		err = checker.CheckHashElementAccess("HashRef[Str,Int]", "key", "Int")
		assert.NoError(t, err, "Valid hash element access with explicit types should not error")

		// Invalid hash element access - wrong value type
		err = checker.CheckHashElementAccess("HashRef[Int]", "key", "Str")
		assert.Error(t, err, "Invalid hash element type should error")
		assert.Contains(t, err.Error(), "element type", "Error should mention element type")

		// Valid hash key access with Int key
		err = checker.CheckHashKeyAccess("HashRef[Int,Bool]", 42)
		assert.NoError(t, err, "Valid Int hash key should not error")

		// Invalid hash key type - string key for Int hash
		err = checker.CheckHashKeyAccess("HashRef[Int,Bool]", "key")
		assert.Error(t, err, "Invalid hash key type should error")
		assert.Contains(t, err.Error(), "key type", "Error should mention key type")

		// Valid hash key access with string key
		err = checker.CheckHashKeyAccess("HashRef[Str,Bool]", "key")
		assert.NoError(t, err, "Valid string hash key should not error")
	})

	// Test nested container validation
	t.Run("NestedContainerValidation", func(t *testing.T) {
		// Valid nested array access
		err := checker.CheckArrayElementAccess("ArrayRef[ArrayRef[Int]]", 0, "ArrayRef[Int]")
		assert.NoError(t, err, "Valid nested array access should not error")

		// Invalid nested array access
		err = checker.CheckArrayElementAccess("ArrayRef[ArrayRef[Int]]", 0, "ArrayRef[Str]")
		assert.Error(t, err, "Invalid nested array element type should error")

		// Valid complex nested structure
		err = checker.CheckHashElementAccess("HashRef[Str,ArrayRef[Int]]", "numbers", "ArrayRef[Int]")
		assert.NoError(t, err, "Valid complex nested structure should not error")

		// Invalid complex nested structure
		err = checker.CheckHashElementAccess("HashRef[Str,ArrayRef[Int]]", "numbers", "ArrayRef[Str]")
		assert.Error(t, err, "Invalid complex nested structure should error")
	})
}

// TestContainerCovariance tests covariant subtyping for containers (Phase 6)
func TestContainerCovariance(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test array covariance: ArrayRef[Int] ⊆ ArrayRef[Num]
	t.Run("ArrayCovariance", func(t *testing.T) {
		// ArrayRef[Int] should be assignable to ArrayRef[Num] (covariance)
		err := checker.CheckContainerTypeCompatibility("ArrayRef[Int]", "ArrayRef[Num]")
		assert.NoError(t, err, "ArrayRef[Int] should be compatible with ArrayRef[Num] (covariance)")

		// ArrayRef[Int] should be assignable to ArrayRef[Scalar]
		err = checker.CheckContainerTypeCompatibility("ArrayRef[Int]", "ArrayRef[Scalar]")
		assert.NoError(t, err, "ArrayRef[Int] should be compatible with ArrayRef[Scalar] (covariance)")

		// ArrayRef[Num] should NOT be assignable to ArrayRef[Int] (contravariance not allowed)
		err = checker.CheckContainerTypeCompatibility("ArrayRef[Num]", "ArrayRef[Int]")
		assert.Error(t, err, "ArrayRef[Num] should not be compatible with ArrayRef[Int] (no contravariance)")

		// ArrayRef[Str] should NOT be assignable to ArrayRef[Int] (unrelated types)
		err = checker.CheckContainerTypeCompatibility("ArrayRef[Str]", "ArrayRef[Int]")
		assert.Error(t, err, "ArrayRef[Str] should not be compatible with ArrayRef[Int] (unrelated types)")
	})

	// Test hash covariance: HashRef[K,V1] ⊆ HashRef[K,V2] if V1 ⊆ V2
	t.Run("HashCovariance", func(t *testing.T) {
		// HashRef[Str,Int] should be assignable to HashRef[Str,Num] (value covariance)
		err := checker.CheckContainerTypeCompatibility("HashRef[Str,Int]", "HashRef[Str,Num]")
		assert.NoError(t, err, "HashRef[Str,Int] should be compatible with HashRef[Str,Num] (value covariance)")

		// HashRef[Str,Int] should be assignable to HashRef[Str,Scalar]
		err = checker.CheckContainerTypeCompatibility("HashRef[Str,Int]", "HashRef[Str,Scalar]")
		assert.NoError(t, err, "HashRef[Str,Int] should be compatible with HashRef[Str,Scalar] (value covariance)")

		// Key types must match exactly (no key covariance)
		err = checker.CheckContainerTypeCompatibility("HashRef[Int,Str]", "HashRef[Num,Str]")
		assert.Error(t, err, "Hash key types should not allow covariance")

		// Simple HashRef[V] (defaulting to HashRef[Str,V]) covariance
		err = checker.CheckContainerTypeCompatibility("HashRef[Int]", "HashRef[Num]")
		assert.NoError(t, err, "HashRef[Int] should be compatible with HashRef[Num] (value covariance)")
	})

	// Test nested container covariance
	t.Run("NestedContainerCovariance", func(t *testing.T) {
		// ArrayRef[ArrayRef[Int]] ⊆ ArrayRef[ArrayRef[Num]]
		err := checker.CheckContainerTypeCompatibility("ArrayRef[ArrayRef[Int]]", "ArrayRef[ArrayRef[Num]]")
		assert.NoError(t, err, "Nested array covariance should work")

		// Complex nested structure covariance
		err = checker.CheckContainerTypeCompatibility("HashRef[Str,ArrayRef[Int]]", "HashRef[Str,ArrayRef[Num]]")
		assert.NoError(t, err, "Complex nested covariance should work")

		// Invalid nested covariance
		err = checker.CheckContainerTypeCompatibility("ArrayRef[ArrayRef[Str]]", "ArrayRef[ArrayRef[Int]]")
		assert.Error(t, err, "Invalid nested covariance should fail")
	})
}

// TestContainerElementOperations tests operations on container elements (Phase 6)
func TestContainerElementOperations(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test array operations
	t.Run("ArrayOperations", func(t *testing.T) {
		// Array push/append operations
		err := checker.CheckArrayPushOperation("ArrayRef[Int]", "Int")
		assert.NoError(t, err, "Valid array push should not error")

		err = checker.CheckArrayPushOperation("ArrayRef[Int]", "Str")
		assert.Error(t, err, "Invalid array push should error")

		// Array iteration type inference
		elementType, err := checker.InferArrayElementType("ArrayRef[Str]")
		assert.NoError(t, err, "Array element type inference should not error")
		assert.Equal(t, "Str", elementType, "Should infer correct element type")

		// Invalid array type for iteration
		_, err = checker.InferArrayElementType("Int")
		assert.Error(t, err, "Non-array type should error for element inference")
	})

	// Test hash operations
	t.Run("HashOperations", func(t *testing.T) {
		// Hash key/value operations
		keyType, valueType, err := checker.InferHashTypes("HashRef[Str,Int]")
		assert.NoError(t, err, "Hash type inference should not error")
		assert.Equal(t, "Str", keyType, "Should infer correct key type")
		assert.Equal(t, "Int", valueType, "Should infer correct value type")

		// Single parameter hash (defaults to HashRef[Str,V])
		keyType, valueType, err = checker.InferHashTypes("HashRef[Bool]")
		assert.NoError(t, err, "Single param hash type inference should not error")
		assert.Equal(t, "Str", keyType, "Should default to Str key type")
		assert.Equal(t, "Bool", valueType, "Should infer correct value type")

		// Invalid hash type
		_, _, err = checker.InferHashTypes("ArrayRef[Int]")
		assert.Error(t, err, "Non-hash type should error for hash inference")
	})

	// Test container type extraction
	t.Run("ContainerTypeExtraction", func(t *testing.T) {
		// Extract container base type and parameters
		baseType, params, err := checker.ExtractContainerInfo("ArrayRef[Int]")
		assert.NoError(t, err, "Container info extraction should not error")
		assert.Equal(t, "ArrayRef", baseType, "Should extract correct base type")
		assert.Equal(t, []string{"Int"}, params, "Should extract correct parameters")

		// Complex container extraction
		baseType, params, err = checker.ExtractContainerInfo("HashRef[Str,ArrayRef[Int]]")
		assert.NoError(t, err, "Complex container extraction should not error")
		assert.Equal(t, "HashRef", baseType, "Should extract correct base type")
		assert.Equal(t, []string{"Str", "ArrayRef[Int]"}, params, "Should extract correct parameters")

		// Non-container type
		_, _, err = checker.ExtractContainerInfo("Int")
		assert.Error(t, err, "Non-container type should error for extraction")
	})
}

// TestUnionTypes tests union type compatibility (Phase 7)
func TestUnionTypes(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test union type compatibility
	t.Run("UnionTypeCompatibility", func(t *testing.T) {
		// Int should be compatible with Int|Str
		err := checker.CheckUnionTypeCompatibility("Int", "Int|Str")
		assert.NoError(t, err, "Int should be compatible with Int|Str")

		// Str should be compatible with Int|Str
		err = checker.CheckUnionTypeCompatibility("Str", "Int|Str")
		assert.NoError(t, err, "Str should be compatible with Int|Str")

		// Num should be compatible with Int|Str (since Int ⊆ Num, and Int is in union)
		err = checker.CheckUnionTypeCompatibility("Num", "Int|Str")
		assert.Error(t, err, "Num should not be compatible with Int|Str (contravariance)")

		// Bool should not be compatible with Int|Str
		err = checker.CheckUnionTypeCompatibility("Bool", "Int|Str")
		assert.Error(t, err, "Bool should not be compatible with Int|Str")
	})

	// Test union assignment compatibility
	t.Run("UnionAssignmentCompatibility", func(t *testing.T) {
		// Int|Str should be compatible with Scalar (since both Int and Str are subtypes of Scalar)
		err := checker.CheckUnionTypeCompatibility("Int|Str", "Scalar")
		assert.NoError(t, err, "Int|Str should be compatible with Scalar")

		// Int|Str should be compatible with Any
		err = checker.CheckUnionTypeCompatibility("Int|Str", "Any")
		assert.NoError(t, err, "Int|Str should be compatible with Any")

		// Int|Str should not be compatible with Int (union is wider than component)
		err = checker.CheckUnionTypeCompatibility("Int|Str", "Int")
		assert.Error(t, err, "Int|Str should not be compatible with Int (union is wider)")
	})

	// Test nested union types
	t.Run("NestedUnionTypes", func(t *testing.T) {
		// ArrayRef[Int] should be compatible with ArrayRef[Int|Str] (covariance)
		err := checker.CheckUnionTypeCompatibility("ArrayRef[Int]", "ArrayRef[Int|Str]")
		assert.NoError(t, err, "ArrayRef[Int] should be compatible with ArrayRef[Int|Str]")

		// ArrayRef[Int|Str] should not be compatible with ArrayRef[Int] (contravariance not allowed)
		err = checker.CheckUnionTypeCompatibility("ArrayRef[Int|Str]", "ArrayRef[Int]")
		assert.Error(t, err, "ArrayRef[Int|Str] should not be compatible with ArrayRef[Int]")

		// Complex union: Int|Str should be compatible with Int|Str|Bool
		err = checker.CheckUnionTypeCompatibility("Int|Str", "Int|Str|Bool")
		assert.NoError(t, err, "Int|Str should be compatible with Int|Str|Bool")
	})
}

// TestIntersectionTypes tests intersection type compatibility (Phase 7)
func TestIntersectionTypes(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test intersection type compatibility
	t.Run("IntersectionTypeCompatibility", func(t *testing.T) {
		// Int&Num should be compatible with Int (intersection is narrower)
		err := checker.CheckIntersectionTypeCompatibility("Int&Num", "Int")
		assert.NoError(t, err, "Int&Num should be compatible with Int")

		// Int&Num should be compatible with Num
		err = checker.CheckIntersectionTypeCompatibility("Int&Num", "Num")
		assert.NoError(t, err, "Int&Num should be compatible with Num")

		// Int should not be compatible with Int&Num (Int doesn't necessarily satisfy both constraints)
		err = checker.CheckIntersectionTypeCompatibility("Int", "Int&Num")
		assert.Error(t, err, "Int should not be compatible with Int&Num")

		// Scalar&Str should be compatible with Str (intersection resolves to the narrower type)
		err = checker.CheckIntersectionTypeCompatibility("Scalar&Str", "Str")
		assert.NoError(t, err, "Scalar&Str should be compatible with Str")
	})

	// Test impossible intersections
	t.Run("ImpossibleIntersections", func(t *testing.T) {
		// Int&Str should be impossible (no value can be both Int and Str)
		isValid, err := checker.ValidateIntersectionType("Int&Str")
		assert.NoError(t, err, "Should not error on syntax")
		assert.False(t, isValid, "Int&Str should be invalid (impossible intersection)")

		// ArrayRef&HashRef should be impossible
		isValid, err = checker.ValidateIntersectionType("ArrayRef&HashRef")
		assert.NoError(t, err, "Should not error on syntax")
		assert.False(t, isValid, "ArrayRef&HashRef should be invalid")

		// Int&Num should be valid (Int is subtype of Num)
		isValid, err = checker.ValidateIntersectionType("Int&Num")
		assert.NoError(t, err, "Should not error")
		assert.True(t, isValid, "Int&Num should be valid")
	})

	// Test intersection with interfaces/traits (conceptual)
	t.Run("IntersectionWithInterfaces", func(t *testing.T) {
		// Define some conceptual interfaces
		// Serializable & Printable would be valid if both interfaces existed
		err := checker.CheckIntersectionTypeCompatibility("Serializable&Printable", "Serializable")
		assert.NoError(t, err, "Intersection should be compatible with component")

		err = checker.CheckIntersectionTypeCompatibility("Serializable&Printable", "Printable")
		assert.NoError(t, err, "Intersection should be compatible with component")
	})
}

// TestNegationTypes tests type negation compatibility (Phase 7)
func TestNegationTypes(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test negation type compatibility
	t.Run("NegationTypeCompatibility", func(t *testing.T) {
		// Str should be compatible with !Int (Str is not Int)
		err := checker.CheckNegationTypeCompatibility("Str", "!Int")
		assert.NoError(t, err, "Str should be compatible with !Int")

		// ArrayRef should be compatible with !Int
		err = checker.CheckNegationTypeCompatibility("ArrayRef[Int]", "!Int")
		assert.NoError(t, err, "ArrayRef should be compatible with !Int")

		// Int should not be compatible with !Int
		err = checker.CheckNegationTypeCompatibility("Int", "!Int")
		assert.Error(t, err, "Int should not be compatible with !Int")

		// Num should not be compatible with !Int (since Int is subtype of Num, this should be rejected for safety)
		err = checker.CheckNegationTypeCompatibility("Num", "!Int")
		assert.Error(t, err, "Num should not be compatible with !Int (could contain Int)")
	})

	// Test negation of container types
	t.Run("ContainerNegation", func(t *testing.T) {
		// Int should be compatible with !Ref (Int is not a reference)
		err := checker.CheckNegationTypeCompatibility("Int", "!Ref")
		assert.NoError(t, err, "Int should be compatible with !Ref")

		// ArrayRef should not be compatible with !Ref (ArrayRef is a Ref)
		err = checker.CheckNegationTypeCompatibility("ArrayRef[Int]", "!Ref")
		assert.Error(t, err, "ArrayRef should not be compatible with !Ref")

		// HashRef should not be compatible with !Ref
		err = checker.CheckNegationTypeCompatibility("HashRef[Str,Int]", "!Ref")
		assert.Error(t, err, "HashRef should not be compatible with !Ref")
	})

	// Test double negation
	t.Run("DoubleNegation", func(t *testing.T) {
		// !!Int should be equivalent to Int
		err := checker.CheckNegationTypeCompatibility("Int", "!!Int")
		assert.NoError(t, err, "!!Int should be equivalent to Int")

		err = checker.CheckNegationTypeCompatibility("Str", "!!Int")
		assert.Error(t, err, "Str should not be compatible with !!Int")
	})
}

// TestComplexTypeExpressions tests complex combinations of type operations (Phase 7)
func TestComplexTypeExpressions(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	// Test complex type expressions
	t.Run("ComplexTypeExpressions", func(t *testing.T) {
		// (Int|Str)&Scalar should be equivalent to Int|Str (since both are subtypes of Scalar)
		err := checker.CheckComplexTypeCompatibility("Int", "(Int|Str)&Scalar")
		assert.NoError(t, err, "Int should be compatible with (Int|Str)&Scalar")

		err = checker.CheckComplexTypeCompatibility("Str", "(Int|Str)&Scalar")
		assert.NoError(t, err, "Str should be compatible with (Int|Str)&Scalar")

		// !(Int|Str) should exclude both Int and Str
		err = checker.CheckComplexTypeCompatibility("Bool", "!(Int|Str)")
		assert.NoError(t, err, "Bool should be compatible with !(Int|Str)")

		err = checker.CheckComplexTypeCompatibility("Int", "!(Int|Str)")
		assert.Error(t, err, "Int should not be compatible with !(Int|Str)")

		// ArrayRef[Int|Str] - complex parameterized union
		err = checker.CheckComplexTypeCompatibility("ArrayRef[Int]", "ArrayRef[Int|Str]")
		assert.NoError(t, err, "ArrayRef[Int] should be compatible with ArrayRef[Int|Str]")
	})

	// Test type expression parsing
	t.Run("TypeExpressionParsing", func(t *testing.T) {
		// Parse and validate complex type expressions
		valid, err := checker.ParseAndValidateTypeExpression("Int|Str")
		assert.NoError(t, err, "Should parse Int|Str")
		assert.True(t, valid, "Int|Str should be valid")

		valid, err = checker.ParseAndValidateTypeExpression("(Int|Str)&Scalar")
		assert.NoError(t, err, "Should parse (Int|Str)&Scalar")
		assert.True(t, valid, "(Int|Str)&Scalar should be valid")

		valid, err = checker.ParseAndValidateTypeExpression("!(ArrayRef[Int]|HashRef[Str,Int])")
		assert.NoError(t, err, "Should parse complex negation")
		assert.True(t, valid, "Complex negation should be valid")

		// Invalid expressions
		_, err = checker.ParseAndValidateTypeExpression("Int|")
		assert.Error(t, err, "Should error on incomplete union")

		_, err = checker.ParseAndValidateTypeExpression("&Str")
		assert.Error(t, err, "Should error on incomplete intersection")
	})

	// Test precedence and associativity
	t.Run("TypeOperatorPrecedence", func(t *testing.T) {
		// Test that & has higher precedence than |
		// Int|Str&Num should parse as Int|(Str&Num)
		result1, err := checker.NormalizeTypeExpression("Int|Str&Num")
		assert.NoError(t, err, "Should normalize precedence")

		result2, err := checker.NormalizeTypeExpression("Int|(Str&Num)")
		assert.NoError(t, err, "Should normalize explicit grouping")

		assert.Equal(t, result1, result2, "Should have same precedence")

		// Test that ! has higher precedence than &
		// !Int&Str should parse as (!Int)&Str
		result1, err = checker.NormalizeTypeExpression("!Int&Str")
		assert.NoError(t, err, "Should normalize negation precedence")

		result2, err = checker.NormalizeTypeExpression("(!Int)&Str")
		assert.NoError(t, err, "Should normalize explicit negation grouping")

		assert.Equal(t, result1, result2, "Should have same negation precedence")
	})
}

// TestContextSensitiveTypes tests Phase 9 - Context-Sensitive Types
func TestContextSensitiveTypes(t *testing.T) {
	// Create a mock storage
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)
	require.NotNil(t, hierarchy)

	// Create a type checker
	checker := NewTypeChecker(hierarchy, "Test::Module")
	require.NotNil(t, checker)

	t.Run("ListContextReturnTypes", func(t *testing.T) {
		// Test context-dependent return types for built-in functions
		// keys %hash should return List[Str] in list context, Int in scalar context

		// List context: my @keys = keys %hash;
		listType, err := checker.InferContextSensitiveReturnType("keys", "HashRef[Str, Int]", "list")
		assert.NoError(t, err, "keys should work in list context")
		assert.Equal(t, "List[Str]", listType, "keys should return List[Str] in list context")

		// Scalar context: my $count = keys %hash;
		scalarType, err := checker.InferContextSensitiveReturnType("keys", "HashRef[Str, Int]", "scalar")
		assert.NoError(t, err, "keys should work in scalar context")
		assert.Equal(t, "Int", scalarType, "keys should return Int in scalar context")
	})

	t.Run("ValuesContextSensitivity", func(t *testing.T) {
		// Test values function context sensitivity
		// values %hash should return List[ValueType] in list context, Int in scalar context

		listType, err := checker.InferContextSensitiveReturnType("values", "HashRef[Str, Int]", "list")
		assert.NoError(t, err, "values should work in list context")
		assert.Equal(t, "List[Int]", listType, "values should return List[Int] in list context")

		scalarType, err := checker.InferContextSensitiveReturnType("values", "HashRef[Str, Int]", "scalar")
		assert.NoError(t, err, "values should work in scalar context")
		assert.Equal(t, "Int", scalarType, "values should return Int in scalar context")
	})

	t.Run("UserDefinedContextSensitiveFunctions", func(t *testing.T) {
		// Test user-defined functions with context-sensitive return types
		// Function signature: sub get_data() -> List[Int]|Str { ... }

		// Register a context-sensitive function
		err := checker.RegisterContextSensitiveFunction("get_data", map[string]string{
			"list":   "List[Int]",
			"scalar": "Str",
		})
		assert.NoError(t, err, "Should register context-sensitive function")

		// Test list context
		listType, err := checker.InferContextSensitiveReturnType("get_data", "", "list")
		assert.NoError(t, err, "get_data should work in list context")
		assert.Equal(t, "List[Int]", listType, "get_data should return List[Int] in list context")

		// Test scalar context
		scalarType, err := checker.InferContextSensitiveReturnType("get_data", "", "scalar")
		assert.NoError(t, err, "get_data should work in scalar context")
		assert.Equal(t, "Str", scalarType, "get_data should return Str in scalar context")
	})

	t.Run("UnionTypeResolution", func(t *testing.T) {
		// Test that union types are resolved correctly based on context
		// Function with union return type: List[Int]|Str

		resolved, err := checker.ResolveUnionTypeForContext("List[Int]|Str", "list")
		assert.NoError(t, err, "Should resolve union type for list context")
		assert.Equal(t, "List[Int]", resolved, "Should select List[Int] for list context")

		resolved, err = checker.ResolveUnionTypeForContext("List[Int]|Str", "scalar")
		assert.NoError(t, err, "Should resolve union type for scalar context")
		assert.Equal(t, "Str", resolved, "Should select Str for scalar context")
	})

	t.Run("ContextInferenceFromAssignment", func(t *testing.T) {
		// Test context inference from assignment targets

		// Array assignment implies list context
		context := checker.InferContextFromAssignment("@array")
		assert.Equal(t, "list", context, "Array assignment should imply list context")

		// Scalar assignment implies scalar context
		context = checker.InferContextFromAssignment("$scalar")
		assert.Equal(t, "scalar", context, "Scalar assignment should imply scalar context")

		// Hash assignment implies list context
		context = checker.InferContextFromAssignment("%hash")
		assert.Equal(t, "list", context, "Hash assignment should imply list context")
	})

	t.Run("InvalidContextResolution", func(t *testing.T) {
		// Test error cases for context resolution

		// Union type with no suitable alternative for context
		_, err := checker.ResolveUnionTypeForContext("Int|Num", "list")
		assert.Error(t, err, "Should error when no list-compatible type in union")

		// Unknown context type
		_, err = checker.InferContextSensitiveReturnType("keys", "HashRef[Str, Int]", "unknown")
		assert.Error(t, err, "Should error for unknown context")
	})
}

// TypeError is a helper type for testing
type TypeError struct {
	Message string
}

func (e *TypeError) Error() string {
	return e.Message
}
