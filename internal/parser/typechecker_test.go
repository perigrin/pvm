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

// TypeError is a helper type for testing
type TypeError struct {
	Message string
}

func (e *TypeError) Error() string {
	return e.Message
}
