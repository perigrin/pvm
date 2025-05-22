// ABOUTME: Tests for the operation resolution system
// ABOUTME: Tests trait-based operation validation and result type resolution

package traits

import (
	"testing"
)

func TestOperationResolver(t *testing.T) {
	resolver := NewOperationResolver()

	// Test basic operation resolution
	result, err := resolver.ResolveOperation("Int", "+", "Int")
	if err != nil {
		t.Errorf("Expected no error for Int + Int, got %v", err)
	}
	if result != "Num" {
		t.Errorf("Expected Int + Int to result in Num, got %s", result)
	}

	// Test unsupported operation
	_, err = resolver.ResolveOperation("ArrayRef", "+", "Int")
	if err == nil {
		t.Error("Expected error for ArrayRef + Int operation")
	}
}

func TestOperationValidation(t *testing.T) {
	resolver := NewOperationResolver()

	tests := []struct {
		typeName  string
		operation string
		expected  bool
	}{
		// Int supports arithmetic
		{"Int", "+", true},
		{"Int", "-", true},
		{"Int", "*", true},
		{"Int", "/", true},
		{"Int", "%", true},
		{"Int", "**", true},
		{"Int", "<=>", true},
		{"Int", "&", true},
		{"Int", "|", true},
		{"Int", "^", true},
		{"Int", "~", true},
		{"Int", "<<", true},
		{"Int", ">>", true},

		// Int supports conversions
		{"Int", "\"\"", true},
		{"Int", "0+", true},
		{"Int", "bool", true},

		// Int does not support string operations
		{"Int", "cmp", false},
		{"Int", "eq", false},
		{"Int", "ne", false},
		{"Int", "lt", false},
		{"Int", "le", false},
		{"Int", "gt", false},
		{"Int", "ge", false},

		// Str supports string operations
		{"Str", "cmp", true},
		{"Str", "eq", true},
		{"Str", "ne", true},
		{"Str", "lt", true},
		{"Str", "le", true},
		{"Str", "gt", true},
		{"Str", "ge", true},

		// Str supports conversions
		{"Str", "\"\"", true},
		{"Str", "0+", true},
		{"Str", "bool", true},

		// Str does not support arithmetic
		{"Str", "+", false},
		{"Str", "-", false},
		{"Str", "*", false},
		{"Str", "/", false},
		{"Str", "%", false},
		{"Str", "**", false},
		{"Str", "<=>", false},

		// ArrayRef only supports conversions
		{"ArrayRef", "\"\"", true},
		{"ArrayRef", "0+", true},
		{"ArrayRef", "bool", true},
		{"ArrayRef", "+", false},
		{"ArrayRef", "cmp", false},

		// HashRef only supports conversions
		{"HashRef", "\"\"", true},
		{"HashRef", "0+", true},
		{"HashRef", "bool", true},
		{"HashRef", "+", false},
		{"HashRef", "cmp", false},
	}

	for _, test := range tests {
		t.Run(test.typeName+"_"+test.operation, func(t *testing.T) {
			supported := resolver.IsOperationSupported(test.typeName, test.operation)
			if supported != test.expected {
				t.Errorf("Expected %s %s support to be %v, got %v",
					test.typeName, test.operation, test.expected, supported)
			}
		})
	}
}

func TestResultTypeResolution(t *testing.T) {
	resolver := NewOperationResolver()

	tests := []struct {
		typeName     string
		operation    string
		expectedType string
	}{
		// Conversion operations
		{"Int", "\"\"", "Str"},
		{"Int", "0+", "Num"},
		{"Int", "bool", "Bool"},
		{"Str", "\"\"", "Str"},
		{"Str", "0+", "Num"},
		{"Str", "bool", "Bool"},
		{"ArrayRef", "\"\"", "Str"},
		{"ArrayRef", "0+", "Num"},
		{"ArrayRef", "bool", "Bool"},

		// Arithmetic operations
		{"Int", "+", "Num"},
		{"Int", "-", "Num"},
		{"Int", "*", "Num"},
		{"Int", "/", "Num"},
		{"Int", "%", "Num"},
		{"Int", "**", "Num"},
		{"Num", "+", "Num"},
		{"Num", "-", "Num"},
		{"Num", "*", "Num"},

		// Comparison operations
		{"Int", "<=>", "Int"},
		{"Str", "cmp", "Int"},
		{"Str", "eq", "Bool"},
		{"Str", "ne", "Bool"},
		{"Str", "lt", "Bool"},
		{"Str", "le", "Bool"},
		{"Str", "gt", "Bool"},
		{"Str", "ge", "Bool"},

		// Bitwise operations
		{"Int", "&", "Int"},
		{"Int", "|", "Int"},
		{"Int", "^", "Int"},
		{"Int", "~", "Int"},
		{"Int", "<<", "Int"},
		{"Int", ">>", "Int"},
		{"Bool", "&", "Int"},
		{"Bool", "|", "Int"},
		{"Bool", "^", "Int"},
		{"Bool", "~", "Int"},
	}

	for _, test := range tests {
		t.Run(test.typeName+"_"+test.operation, func(t *testing.T) {
			resultType := resolver.GetResultType(test.typeName, test.operation)
			if resultType != test.expectedType {
				t.Errorf("Expected %s %s to result in %s, got %s",
					test.typeName, test.operation, test.expectedType, resultType)
			}
		})
	}
}

func TestBinaryOperationResolution(t *testing.T) {
	resolver := NewOperationResolver()

	tests := []struct {
		leftType     string
		operation    string
		rightType    string
		expectedType string
		shouldError  bool
	}{
		// Valid binary operations
		{"Int", "+", "Int", "Num", false},
		{"Int", "+", "Num", "Num", false},
		{"Num", "+", "Int", "Num", false},
		{"Num", "+", "Num", "Num", false},
		{"Str", "eq", "Str", "Bool", false},
		{"Int", "<=>", "Int", "Int", false},
		{"Int", "&", "Int", "Int", false},

		// Invalid binary operations - left type doesn't support operation
		{"ArrayRef", "+", "Int", "", true},
		{"HashRef", "*", "Num", "", true},
		{"Str", "/", "Int", "", true},

		// Note: We're focusing on left operand support for now
		// In a more sophisticated system, we might check both operands
	}

	for _, test := range tests {
		t.Run(test.leftType+"_"+test.operation+"_"+test.rightType, func(t *testing.T) {
			resultType, err := resolver.ResolveOperation(test.leftType, test.operation, test.rightType)

			if test.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s %s %s, but got none",
						test.leftType, test.operation, test.rightType)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s %s %s, but got %v",
						test.leftType, test.operation, test.rightType, err)
				}
				if resultType != test.expectedType {
					t.Errorf("Expected %s %s %s to result in %s, got %s",
						test.leftType, test.operation, test.rightType, test.expectedType, resultType)
				}
			}
		})
	}
}

func TestUnsupportedOperationErrors(t *testing.T) {
	resolver := NewOperationResolver()

	// Test error messages for unsupported operations
	_, err := resolver.ResolveOperation("ArrayRef", "+", "Int")
	if err == nil {
		t.Error("Expected error for unsupported operation")
	}

	expectedMsg := "operation '+' not supported on type 'ArrayRef'"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestUnknownTypeHandling(t *testing.T) {
	resolver := NewOperationResolver()

	// Test with unknown types
	supported := resolver.IsOperationSupported("UnknownType", "+")
	if supported {
		t.Error("Unknown types should not support arithmetic operations")
	}

	// Unknown types should support basic conversions
	supported = resolver.IsOperationSupported("UnknownType", "\"\"")
	if !supported {
		t.Error("Unknown types should support string conversion")
	}

	supported = resolver.IsOperationSupported("UnknownType", "0+")
	if !supported {
		t.Error("Unknown types should support numeric conversion")
	}

	supported = resolver.IsOperationSupported("UnknownType", "bool")
	if !supported {
		t.Error("Unknown types should support boolean conversion")
	}
}

func TestOperationResolverWithCustomTraits(t *testing.T) {
	resolver := NewOperationResolver()

	// Test adding custom traits to the resolver
	customTraits := NewTraitSet()
	customTraits.AddTrait(Trait{Operation: "+", ResultType: "CustomNum"})
	customTraits.AddTrait(Trait{Operation: "\"\"", ResultType: "Str"})

	resolver.SetTraitsForType("CustomType", customTraits)

	// Test that custom type now supports the operations
	if !resolver.IsOperationSupported("CustomType", "+") {
		t.Error("CustomType should support '+' operation")
	}

	if !resolver.IsOperationSupported("CustomType", "\"\"") {
		t.Error("CustomType should support '\"\"' operation")
	}

	if resolver.IsOperationSupported("CustomType", "-") {
		t.Error("CustomType should not support '-' operation")
	}

	// Test result type
	resultType := resolver.GetResultType("CustomType", "+")
	if resultType != "CustomNum" {
		t.Errorf("Expected CustomType + to result in CustomNum, got %s", resultType)
	}
}
