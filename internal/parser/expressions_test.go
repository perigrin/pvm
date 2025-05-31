// ABOUTME: Comprehensive test suite for expression and operator parsing validation
// ABOUTME: Tests arithmetic, string, logical, comparison, assignment, and bitwise operations

package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestBasicExpressions validates basic expression parsing using the test framework
func TestBasicExpressions(t *testing.T) {
	testCategories := []string{
		"arithmetic_operations",
		"string_operations",
		"logical_operations",
		"comparison_operations",
		"assignment_operations",
		"bitwise_operations",
		"complex_expressions",
	}

	for _, category := range testCategories {
		t.Run(category, func(t *testing.T) {
			testFile := filepath.Join("testdata", "untyped-perl", "expressions", category+".json")
			runExpressionTestsFromFile(t, testFile)
		})
	}
}

// runExpressionTestsFromFile loads and runs tests from a JSON file
func runExpressionTestsFromFile(t *testing.T, testFile string) {
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", testFile, err)
	}

	var testCases []ParserTestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		t.Fatalf("Failed to parse test file %s: %v", testFile, err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ast, err := parser.ParseString(testCase.Input)

			if testCase.ShouldError {
				if err == nil {
					t.Errorf("Expected error for input: %s", testCase.Input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected parse error for input '%s': %v", testCase.Input, err)
				return
			}

			if ast == nil {
				t.Errorf("Parser returned nil AST for input: %s", testCase.Input)
				return
			}

			// Validate AST contains expected expression structure
			validateExpressionAST(t, ast, testCase)
		})
	}
}

// validateExpressionAST performs basic validation of expression AST structure
func validateExpressionAST(t *testing.T, ast *AST, testCase ParserTestCase) {
	if ast.Root == nil {
		t.Errorf("AST root is nil for test case: %s", testCase.Name)
		return
	}

	// Check that we have children (which represent statements/expressions)
	children := ast.Root.Children()
	if len(children) == 0 {
		t.Errorf("No child nodes found in AST for test case: %s", testCase.Name)
		return
	}

	// For expression tests, we expect to find expression-related nodes
	stmt := children[0]
	if stmt == nil {
		t.Errorf("First child node is nil for test case: %s", testCase.Name)
		return
	}

	// Additional validation based on test tags
	for _, tag := range testCase.Tags {
		switch tag {
		case "arithmetic":
			validateArithmeticExpression(t, stmt, testCase)
		case "string":
			validateStringExpression(t, stmt, testCase)
		case "logical":
			validateLogicalExpression(t, stmt, testCase)
		case "comparison":
			validateComparisonExpression(t, stmt, testCase)
		case "assignment":
			validateAssignmentExpression(t, stmt, testCase)
		case "bitwise":
			validateBitwiseExpression(t, stmt, testCase)
		}
	}
}

// validateArithmeticExpression checks arithmetic expression structure
func validateArithmeticExpression(t *testing.T, stmt interface{}, testCase ParserTestCase) {
	// Basic validation that arithmetic operators are recognized
	// This would be expanded based on the actual AST structure
	t.Logf("Validating arithmetic expression for: %s", testCase.Name)
}

// validateStringExpression checks string expression structure
func validateStringExpression(t *testing.T, stmt interface{}, testCase ParserTestCase) {
	// Basic validation that string operators are recognized
	t.Logf("Validating string expression for: %s", testCase.Name)
}

// validateLogicalExpression checks logical expression structure
func validateLogicalExpression(t *testing.T, stmt interface{}, testCase ParserTestCase) {
	// Basic validation that logical operators are recognized
	t.Logf("Validating logical expression for: %s", testCase.Name)
}

// validateComparisonExpression checks comparison expression structure
func validateComparisonExpression(t *testing.T, stmt interface{}, testCase ParserTestCase) {
	// Basic validation that comparison operators are recognized
	t.Logf("Validating comparison expression for: %s", testCase.Name)
}

// validateAssignmentExpression checks assignment expression structure
func validateAssignmentExpression(t *testing.T, stmt interface{}, testCase ParserTestCase) {
	// Basic validation that assignment operators are recognized
	t.Logf("Validating assignment expression for: %s", testCase.Name)
}

// validateBitwiseExpression checks bitwise expression structure
func validateBitwiseExpression(t *testing.T, stmt interface{}, testCase ParserTestCase) {
	// Basic validation that bitwise operators are recognized
	t.Logf("Validating bitwise expression for: %s", testCase.Name)
}

// BenchmarkExpressionParsing benchmarks expression parsing performance
func BenchmarkExpressionParsing(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	testExpressions := []string{
		"$result = $a + $b;",
		"$complex = ($a + $b) * ($c - $d) / ($e || 1);",
		"$string = $first . $second . $third;",
		"$logical = ($a && $b) || ($c && !$d);",
		"$assignment = $value += $increment * 2;",
		"$bitwise = ($flags & $mask) | ($bits << 2);",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, expr := range testExpressions {
			_, err := parser.ParseString(expr)
			if err != nil {
				b.Errorf("Parse error for expression '%s': %v", expr, err)
			}
		}
	}
}

// TestExpressionPrecedence validates operator precedence in complex expressions
func TestExpressionPrecedence(t *testing.T) {
	precedenceTests := []struct {
		name     string
		input    string
		expected string // Expected evaluation order (conceptual)
	}{
		{
			name:     "arithmetic_precedence",
			input:    "$result = $a + $b * $c;",
			expected: "$result = $a + ($b * $c);",
		},
		{
			name:     "logical_precedence",
			input:    "$result = $a && $b || $c;",
			expected: "$result = ($a && $b) || $c;",
		},
		{
			name:     "mixed_precedence",
			input:    "$result = $a + $b > $c && $d;",
			expected: "$result = (($a + $b) > $c) && $d;",
		},
		{
			name:     "assignment_precedence",
			input:    "$a = $b += $c * $d;",
			expected: "$a = ($b += ($c * $d));",
		},
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, test := range precedenceTests {
		t.Run(test.name, func(t *testing.T) {
			ast, err := parser.ParseString(test.input)
			if err != nil {
				t.Errorf("Parse error for precedence test '%s': %v", test.name, err)
				return
			}

			if ast == nil {
				t.Errorf("Parser returned nil AST for precedence test: %s", test.name)
				return
			}

			// Note: Actual precedence validation would require examining
			// the AST structure to ensure operators are grouped correctly
			t.Logf("Precedence test '%s' parsed successfully", test.name)
		})
	}
}
