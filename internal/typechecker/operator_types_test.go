// ABOUTME: Tests for Perl binary operator type registry
// ABOUTME: Validates operator type definition parsing and result type lookup

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

func TestOperatorTypeRegistry_BasicOperators(t *testing.T) {
	registry, err := NewOperatorTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create operator type registry: %v", err)
	}

	testCases := []struct {
		operator  string
		leftType  string
		rightType string
		expected  string
		desc      string
	}{
		{"+", "Int", "Int", "Int", "Int + Int should return Int"},
		{"+", "Num", "Num", "Num", "Num + Num should return Num"},
		{"+", "Int", "Num", "Num", "Int + Num should return Num"},
		{".", "Str", "Str", "Str", "Str . Str should return Str"},
		{"==", "Int", "Int", "Bool", "Int == Int should return Bool"},
		{"eq", "Str", "Str", "Bool", "Str eq Str should return Bool"},
		{"=~", "Str", "RegexRef", "Bool", "Str =~ RegexRef should return Bool"},
		{"//", "Any", "Any", "Any", "Any // Any should return Any"},
	}

	for _, tc := range testCases {
		t.Run(tc.operator, func(t *testing.T) {
			actual := registry.GetOperatorType(tc.operator, tc.leftType, tc.rightType)
			if actual != tc.expected {
				t.Errorf("%s: expected %s, got %s", tc.desc, tc.expected, actual)
			}
		})
	}
}

func TestOperatorTypeRegistry_IsKnownOperator(t *testing.T) {
	registry, err := NewOperatorTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create operator type registry: %v", err)
	}

	testCases := []struct {
		operator string
		expected bool
		desc     string
	}{
		{"+", true, "+ should be recognized as known operator"},
		{"-", true, "- should be recognized as known operator"},
		{".", true, ". should be recognized as known operator"},
		{"==", true, "== should be recognized as known operator"},
		{"eq", true, "eq should be recognized as known operator"},
		{"unknown_op", false, "unknown operator should not be recognized"},
		{"custom_op", false, "custom operator should not be recognized"},
	}

	for _, tc := range testCases {
		t.Run(tc.operator, func(t *testing.T) {
			actual := registry.IsKnownOperator(tc.operator)
			if actual != tc.expected {
				t.Errorf("%s: expected %v, got %v", tc.desc, tc.expected, actual)
			}
		})
	}
}

func TestOperatorTypeRegistry_TypeCoercion(t *testing.T) {
	registry, err := NewOperatorTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create operator type registry: %v", err)
	}

	// Test that operators work with Any types (Perl's type coercion)
	testCases := []struct {
		operator  string
		leftType  string
		rightType string
		expected  string
		desc      string
	}{
		{"+", "Any", "Any", "Num", "Any + Any should coerce to Num"},
		{".", "Any", "Any", "Str", "Any . Any should coerce to Str"},
		{"==", "Any", "Any", "Bool", "Any == Any should return Bool"},
		{"+", "Int", "Any", "Num", "Int + Any should work with coercion"},
		{".", "Str", "Any", "Str", "Str . Any should work with coercion"},
	}

	for _, tc := range testCases {
		t.Run(tc.operator+"_coercion", func(t *testing.T) {
			actual := registry.GetOperatorType(tc.operator, tc.leftType, tc.rightType)
			if actual != tc.expected {
				t.Errorf("%s: expected %s, got %s", tc.desc, tc.expected, actual)
			}
		})
	}
}

func TestOperatorTypeRegistry_GetSignatures(t *testing.T) {
	registry, err := NewOperatorTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create operator type registry: %v", err)
	}

	// Test that we can get all signatures for an operator
	signatures := registry.GetOperatorSignatures("+")
	if len(signatures) == 0 {
		t.Error("Expected to find signatures for +, but got none")
	}

	// Verify signature string formatting works
	for _, sig := range signatures {
		sigStr := sig.GetTypeSignatureString()
		if sigStr == "" {
			t.Error("Expected non-empty signature string")
		}
		t.Logf("+ signature: %s", sigStr)
	}
}

func TestOperatorTypeRegistry_ListAllOperators(t *testing.T) {
	registry, err := NewOperatorTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create operator type registry: %v", err)
	}

	operators := registry.ListAllOperators()
	if len(operators) == 0 {
		t.Error("Expected to find operators, but got none")
	}

	// Should include some core Perl operators
	expectedOperators := []string{"+", "-", ".", "==", "eq", "=~", "&&", "||"}
	for _, expected := range expectedOperators {
		found := false
		for _, op := range operators {
			if op == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find %s in operator list, but didn't", expected)
		}
	}

	t.Logf("Found %d operators", len(operators))
}

func TestFlowAnalyzer_WithOperatorRegistry(t *testing.T) {
	// Test that FlowAnalyzer properly uses the operator registry
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")

	analyzer := NewFlowAnalyzer(tc)

	if analyzer.OperatorTypes == nil {
		t.Error("Expected FlowAnalyzer to have OperatorTypes initialized")
	}

	// Test operator type inference using mock binary expressions
	// Note: This tests the integration without requiring full AST parsing
	if analyzer.OperatorTypes != nil {
		resultType := analyzer.OperatorTypes.GetOperatorType("+", "Int", "Int")
		if resultType != "Int" {
			t.Errorf("Expected Int + Int to return Int, got %s", resultType)
		}

		resultType = analyzer.OperatorTypes.GetOperatorType(".", "Str", "Str")
		if resultType != "Str" {
			t.Errorf("Expected Str . Str to return Str, got %s", resultType)
		}

		resultType = analyzer.OperatorTypes.GetOperatorType("==", "Any", "Any")
		if resultType != "Bool" {
			t.Errorf("Expected Any == Any to return Bool, got %s", resultType)
		}
	}
}

func TestOperatorSignature_TypeSignatureString(t *testing.T) {
	// Test binary operator signature
	binarySig := &OperatorSignature{
		Operator:   "+",
		LeftType:   "Int",
		RightType:  "Int",
		ResultType: "Int",
		Context:    "scalar",
	}

	sigStr := binarySig.GetTypeSignatureString()
	expected := "Int + Int -> Int"
	if sigStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, sigStr)
	}

	// Test unary operator signature
	unarySig := &OperatorSignature{
		Operator:   "!",
		LeftType:   "Any",
		RightType:  "", // Empty for unary
		ResultType: "Bool",
		Context:    "scalar",
	}

	sigStr = unarySig.GetTypeSignatureString()
	expected = "Any ! -> Bool"
	if sigStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, sigStr)
	}
}
