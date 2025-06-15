// ABOUTME: Tests for method signature validation functionality
// ABOUTME: Tests validation of method signatures, parameter types, and signature compatibility

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
)

// TestMethodSignatureValidator tests the method signature validation functionality
func TestMethodSignatureValidator(t *testing.T) {
	tc := &TypeChecker{} // Simplified for testing
	validator := NewMethodSignatureValidator(tc)

	// Create test annotations similar to what the parser would generate
	annotations := []*ast.TypeAnnotation{
		{
			AnnotatedItem:  "add",
			TypeExpression: &ast.TypeExpression{Name: "Int"},
			Kind:           ast.MethodReturnAnnotation,
			Pos:            ast.Position{Line: 1, Column: 1},
		},
		{
			AnnotatedItem:  "$a",
			TypeExpression: &ast.TypeExpression{Name: "Int"},
			Kind:           ast.MethodParamAnnotation,
			Pos:            ast.Position{Line: 1, Column: 10},
		},
		{
			AnnotatedItem:  "$b",
			TypeExpression: &ast.TypeExpression{Name: "Int"},
			Kind:           ast.MethodParamAnnotation,
			Pos:            ast.Position{Line: 1, Column: 20},
		},
	}

	// Validate the method signatures
	errors := validator.ValidateMethodSignatures(annotations)
	if len(errors) > 0 {
		for _, err := range errors {
			t.Logf("Validation error: %v", err)
		}
		t.Errorf("Expected no validation errors, got %d", len(errors))
	}

	// Test method signature retrieval
	addSig, exists := validator.GetMethodSignature("add")
	if !exists {
		t.Error("Expected to find 'add' method signature")
	} else {
		if addSig.ReturnType == nil || addSig.ReturnType.Name != "Int" {
			t.Errorf("Expected add method to return Int, got %v", addSig.ReturnType)
		}
		t.Logf("Successfully found method signature: %s returns %s", addSig.Name, addSig.ReturnType.Name)
	}
}

// TestTypeExpressionValidation tests validation of individual type expressions
func TestTypeExpressionValidation(t *testing.T) {
	tc := &TypeChecker{}
	validator := NewMethodSignatureValidator(tc)

	testCases := []struct {
		name        string
		typeExpr    *ast.TypeExpression
		expectError bool
		errorType   string
	}{
		{
			name:        "valid_builtin_type",
			typeExpr:    &ast.TypeExpression{Name: "Int"},
			expectError: false,
		},
		{
			name:        "valid_class_type",
			typeExpr:    &ast.TypeExpression{Name: "MyClass"},
			expectError: false,
		},
		{
			name: "valid_parameterized_type",
			typeExpr: &ast.TypeExpression{
				Name: "ArrayRef[Int]",
				Parameters: []*ast.TypeExpression{
					{Name: "Int"},
				},
			},
			expectError: false,
		},
		{
			name:        "empty_type_name",
			typeExpr:    &ast.TypeExpression{Name: ""},
			expectError: true,
			errorType:   "empty type name",
		},
		{
			name:        "unknown_lowercase_type",
			typeExpr:    &ast.TypeExpression{Name: "unknowntype"},
			expectError: true,
			errorType:   "unknown type",
		},
		{
			name: "invalid_parameterization",
			typeExpr: &ast.TypeExpression{
				Name: "Int[Str]", // Int cannot be parameterized
				Parameters: []*ast.TypeExpression{
					{Name: "Str"},
				},
			},
			expectError: true,
			errorType:   "invalid parameterization",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.validateTypeExpression(tc.typeExpr, tc.name)

			if tc.expectError && err == nil {
				t.Errorf("Expected validation error for %s but got none", tc.name)
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected validation error for %s: %v", tc.name, err)
			} else if err != nil {
				t.Logf("Expected validation error for %s: %v", tc.name, err)
			} else {
				t.Logf("Successfully validated type expression: %s", tc.typeExpr.Name)
			}
		})
	}
}

// TestMethodCallValidation tests validation of method calls against signatures
func TestMethodCallValidation(t *testing.T) {
	tc := &TypeChecker{}
	validator := NewMethodSignatureValidator(tc)

	// Setup a method signature
	annotations := []*ast.TypeAnnotation{
		{
			AnnotatedItem:  "divide",
			TypeExpression: &ast.TypeExpression{Name: "Num"},
			Kind:           ast.MethodReturnAnnotation,
		},
		{
			AnnotatedItem:  "$dividend",
			TypeExpression: &ast.TypeExpression{Name: "Num"},
			Kind:           ast.MethodParamAnnotation,
		},
		{
			AnnotatedItem:  "$divisor",
			TypeExpression: &ast.TypeExpression{Name: "Num"},
			Kind:           ast.MethodParamAnnotation,
		},
	}

	validator.ValidateMethodSignatures(annotations)

	testCases := []struct {
		name        string
		methodName  string
		argTypes    []*ast.TypeExpression
		expectError bool
		errorType   string
	}{
		{
			name:       "valid_call",
			methodName: "divide",
			argTypes: []*ast.TypeExpression{
				{Name: "Num"},
				{Name: "Num"},
			},
			expectError: false,
		},
		{
			name:        "too_few_args",
			methodName:  "divide",
			argTypes:    []*ast.TypeExpression{{Name: "Num"}},
			expectError: true,
			errorType:   "insufficient arguments",
		},
		{
			name:       "too_many_args",
			methodName: "divide",
			argTypes: []*ast.TypeExpression{
				{Name: "Num"},
				{Name: "Num"},
				{Name: "Num"},
			},
			expectError: true,
			errorType:   "too many arguments",
		},
		{
			name:        "unknown_method",
			methodName:  "unknown_method",
			argTypes:    []*ast.TypeExpression{{Name: "Str"}},
			expectError: true,
			errorType:   "unknown method",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateMethodCall(tc.methodName, tc.argTypes, ast.Position{Line: 1, Column: 1})

			if tc.expectError && err == nil {
				t.Errorf("Expected validation error for %s but got none", tc.name)
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected validation error for %s: %v", tc.name, err)
			} else if err != nil {
				t.Logf("Expected validation error for %s: %v", tc.name, err)
			} else {
				t.Logf("Successfully validated method call: %s", tc.methodName)
			}
		})
	}
}

// TestGenericMethodValidation tests validation of generic method signatures
func TestGenericMethodValidation(t *testing.T) {
	tc := &TypeChecker{}
	validator := NewMethodSignatureValidator(tc)

	// Test generic type parameter validation
	signature := &MethodSignature{
		Name: "generic_method",
		Parameters: []MethodParameter{
			{
				Name: "$input",
				Type: &ast.TypeExpression{Name: "T"},
			},
		},
		ReturnType: &ast.TypeExpression{Name: "T"},
	}

	genericParams := []string{"T"}
	err := validator.ValidateGenericMethodSignature(signature, genericParams)

	if err != nil {
		t.Errorf("Unexpected error validating generic method signature: %v", err)
	} else {
		t.Log("Successfully validated generic method signature")
	}

	// Test unused generic parameter
	signature2 := &MethodSignature{
		Name: "unused_generic",
		Parameters: []MethodParameter{
			{
				Name: "$input",
				Type: &ast.TypeExpression{Name: "Str"},
			},
		},
		ReturnType: &ast.TypeExpression{Name: "Int"},
	}

	unusedParams := []string{"T", "U"}
	err = validator.ValidateGenericMethodSignature(signature2, unusedParams)

	// This should not error, but unused parameters could generate warnings
	if err != nil {
		t.Errorf("Unexpected error with unused generic parameters: %v", err)
	} else {
		t.Log("Successfully handled unused generic parameters")
	}
}
