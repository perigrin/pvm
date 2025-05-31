// ABOUTME: Tests for type error recovery and position tracking
// ABOUTME: Validates graceful error handling for malformed type expressions

package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

// ErrorTestCase represents a test case for error recovery
type ErrorTestCase struct {
	Name               string `json:"name"`
	Input              string `json:"input"`
	ExpectedError      string `json:"expected_error"`
	ExpectedSuggestion string `json:"expected_suggestion"`
	Context            string `json:"context"`
}

// ErrorTestSuite represents a collection of error test cases
type ErrorTestSuite struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	TestCases   []ErrorTestCase `json:"test_cases"`
}

func TestTypeErrorRecovery(t *testing.T) {
	// Test basic error recovery functionality
	recovery := NewTypeErrorRecovery()

	testCases := []struct {
		name         string
		source       string
		position     ast.Position
		context      string
		expectedCode TypeErrorCode
		expectedMsg  string
	}{
		{
			name:         "missing_closing_bracket",
			source:       "my ArrayRef[Int $var;",
			position:     ast.Position{Line: 1, Column: 16},
			context:      "variable declaration",
			expectedCode: MissingClosingBracketError,
			expectedMsg:  "Missing closing bracket in parameterized type",
		},
		{
			name:         "invalid_union_syntax",
			source:       "my Int||Str $union;",
			position:     ast.Position{Line: 1, Column: 7},
			context:      "union type",
			expectedCode: InvalidUnionSyntaxError,
			expectedMsg:  "Invalid union type syntax - use single '|' between types",
		},
		{
			name:         "incomplete_type_assertion",
			source:       "my $val = $input as ;",
			position:     ast.Position{Line: 1, Column: 20},
			context:      "type assertion",
			expectedCode: IncompleteTypeAssertionError,
			expectedMsg:  "Incomplete type assertion - missing target type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := recovery.RecoverFromTypeError(tc.source, tc.position, tc.context)

			if err.ErrorCode != tc.expectedCode {
				t.Errorf("Expected error code %v, got %v", tc.expectedCode, err.ErrorCode)
			}

			if err.Message != tc.expectedMsg {
				t.Errorf("Expected message %q, got %q", tc.expectedMsg, err.Message)
			}

			if err.Context != tc.context {
				t.Errorf("Expected context %q, got %q", tc.context, err.Context)
			}
		})
	}
}

func TestTypeErrorRecoveryFromFile(t *testing.T) {
	// Load test cases from JSON file
	testFile := filepath.Join("testdata", "error-cases", "malformed_types.json")
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	var testSuite ErrorTestSuite
	if err := json.Unmarshal(data, &testSuite); err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	recovery := NewTypeErrorRecovery()

	for _, tc := range testSuite.TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			position := ast.Position{Line: 1, Column: 1}
			err := recovery.RecoverFromTypeError(tc.Input, position, tc.Context)

			// Map error codes to string names for comparison
			expectedCode := getErrorCodeByName(tc.ExpectedError)
			if err.ErrorCode != expectedCode {
				t.Errorf("Expected error code %s (%v), got %v", tc.ExpectedError, expectedCode, err.ErrorCode)
			}

			if err.Suggestion != tc.ExpectedSuggestion {
				t.Errorf("Expected suggestion %q, got %q", tc.ExpectedSuggestion, err.Suggestion)
			}
		})
	}
}

func TestNestingDepthValidation(t *testing.T) {
	recovery := NewTypeErrorRecovery()
	position := ast.Position{Line: 1, Column: 1}

	// Test acceptable nesting depth
	err := recovery.CheckNestingDepth(10, position)
	if err != nil {
		t.Errorf("Expected no error for depth 10, got: %v", err)
	}

	// Test excessive nesting depth
	err = recovery.CheckNestingDepth(25, position)
	if err == nil {
		t.Error("Expected error for depth 25, got nil")
		return
	}

	if err.ErrorCode != DeepNestingError {
		t.Errorf("Expected DeepNestingError, got %v", err.ErrorCode)
	}
}

func TestSynchronizationPointFinding(t *testing.T) {
	recovery := NewTypeErrorRecovery()

	testCases := []struct {
		name     string
		source   string
		position ast.Position
		expected ast.Position
	}{
		{
			name:     "find_semicolon",
			source:   "my ArrayRef[Int $var; my Str $other;",
			position: ast.Position{Line: 1, Column: 16},
			expected: ast.Position{Line: 1, Column: 22}, // After semicolon
		},
		{
			name:     "find_block_start",
			source:   "my ArrayRef[Int $var { block }",
			position: ast.Position{Line: 1, Column: 16},
			expected: ast.Position{Line: 1, Column: 23}, // After opening brace
		},
		{
			name:     "find_variable_declaration",
			source:   "my ArrayRef[Int $var my Str $other;",
			position: ast.Position{Line: 1, Column: 16},
			expected: ast.Position{Line: 1, Column: 24}, // After 'my'
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			syncPoint := recovery.FindSynchronizationPoint(tc.source, tc.position)

			if syncPoint.Column != tc.expected.Column {
				t.Errorf("Expected column %d, got %d", tc.expected.Column, syncPoint.Column)
			}
		})
	}
}

func TestTypeExpressionValidation(t *testing.T) {
	recovery := NewTypeErrorRecovery()

	testCases := []struct {
		name         string
		expr         *ast.TypeExpression
		expectedErrs int
	}{
		{
			name: "valid_union_type",
			expr: &ast.TypeExpression{
				BaseNode: ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 10}),
				IsUnion:  true,
				UnionTypes: []*ast.TypeExpression{
					{BaseNode: ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 3}), Name: "Int"},
					{BaseNode: ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 5}, ast.Position{Line: 1, Column: 7}), Name: "Str"},
				},
			},
			expectedErrs: 0,
		},
		{
			name: "invalid_union_type_single",
			expr: &ast.TypeExpression{
				BaseNode:   ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 3}),
				IsUnion:    true,
				UnionTypes: []*ast.TypeExpression{{BaseNode: ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 3}), Name: "Int"}},
			},
			expectedErrs: 1,
		},
		{
			name: "valid_parameterized_type",
			expr: &ast.TypeExpression{
				BaseNode: ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 12}),
				Name:     "ArrayRef",
				Parameters: []*ast.TypeExpression{
					{BaseNode: ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 10}, ast.Position{Line: 1, Column: 12}), Name: "Int"},
				},
			},
			expectedErrs: 0,
		},
		{
			name: "invalid_parameterized_type_no_name",
			expr: &ast.TypeExpression{
				BaseNode: ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 5}),
				Name:     "",
				Parameters: []*ast.TypeExpression{
					{BaseNode: ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 2}, ast.Position{Line: 1, Column: 4}), Name: "Int"},
				},
			},
			expectedErrs: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errors := recovery.ValidateTypeExpression(tc.expr, "test source")

			if len(errors) != tc.expectedErrs {
				t.Errorf("Expected %d errors, got %d", tc.expectedErrs, len(errors))
				for i, err := range errors {
					t.Logf("Error %d: %v", i, err)
				}
			}
		})
	}
}

func TestErrorStringFormatting(t *testing.T) {
	err := &TypeError{
		Message:    "Test error message",
		Position:   ast.Position{Line: 5, Column: 10},
		Suggestion: "Try this fix",
		Context:    "test context",
		ErrorCode:  InvalidUnionSyntaxError,
		Source:     "source code",
	}

	expected := "5:10: Test error message (in test context) - Try this fix"
	actual := err.Error()

	if actual != expected {
		t.Errorf("Expected error string %q, got %q", expected, actual)
	}
}

func TestPositionTracking(t *testing.T) {
	// Test that position information is accurately maintained
	testCases := []struct {
		name     string
		source   string
		line     int
		column   int
		expected ast.Position
	}{
		{
			name:     "simple_position",
			source:   "my Int $var;",
			line:     1,
			column:   4,
			expected: ast.Position{Line: 1, Column: 4, Offset: 3},
		},
		{
			name:     "multiline_position",
			source:   "my Int $var;\nmy Str $other;",
			line:     2,
			column:   4,
			expected: ast.Position{Line: 2, Column: 4, Offset: 16},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate offset for given line/column
			lines := []string{}
			currentLine := ""
			for i, char := range tc.source {
				if char == '\n' {
					lines = append(lines, currentLine)
					currentLine = ""
				} else {
					currentLine += string(char)
				}

				// If we're at the target position, check offset calculation
				lineNum := len(lines) + 1
				colNum := len(currentLine)

				if lineNum == tc.line && colNum == tc.column {
					if i != tc.expected.Offset {
						t.Errorf("Expected offset %d, got %d", tc.expected.Offset, i)
					}
				}
			}
		})
	}
}

// Helper function to map error code names to enum values
func getErrorCodeByName(name string) TypeErrorCode {
	switch name {
	case "MissingClosingBracketError":
		return MissingClosingBracketError
	case "InvalidUnionSyntaxError":
		return InvalidUnionSyntaxError
	case "IncompleteTypeAssertionError":
		return IncompleteTypeAssertionError
	case "InvalidParameterizedTypeError":
		return InvalidParameterizedTypeError
	case "MissingTypeNameError":
		return MissingTypeNameError
	case "InvalidWhereClauseError":
		return InvalidWhereClauseError
	case "InvalidIntersectionSyntaxError":
		return InvalidIntersectionSyntaxError
	case "InvalidNegationSyntaxError":
		return InvalidNegationSyntaxError
	case "DeepNestingError":
		return DeepNestingError
	default:
		return UnknownTypeError
	}
}

func BenchmarkErrorRecovery(b *testing.B) {
	recovery := NewTypeErrorRecovery()
	source := "my ArrayRef[Int $var;"
	position := ast.Position{Line: 1, Column: 16}
	context := "variable declaration"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recovery.RecoverFromTypeError(source, position, context)
	}
}

func BenchmarkNestingDepthCheck(b *testing.B) {
	recovery := NewTypeErrorRecovery()
	position := ast.Position{Line: 1, Column: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recovery.CheckNestingDepth(15, position)
	}
}
