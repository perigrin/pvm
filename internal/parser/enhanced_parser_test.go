// ABOUTME: Tests for enhanced parser with error recovery and position tracking
// ABOUTME: Validates comprehensive error handling and recovery capabilities

package parser

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestEnhancedParserBasicFunctionality(t *testing.T) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Test basic valid code parsing
	validCode := `my Int $count = 42;`
	ast, err := parser.ParseString(validCode)
	if err != nil {
		t.Errorf("Failed to parse valid code: %v", err)
	}
	if ast == nil {
		t.Error("Expected non-nil AST for valid code")
	}
}

func TestEnhancedParserErrorRecovery(t *testing.T) {
	parser, err := NewEnhancedParser(true) // Enable debug mode
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	testCases := []struct {
		name        string
		input       string
		expectAST   bool
		expectError bool
	}{
		{
			name:        "missing_closing_bracket_auto_fix",
			input:       "my ArrayRef[Int $var;",
			expectAST:   true,  // Should recover and produce AST
			expectError: false, // Should auto-fix
		},
		{
			name:        "double_union_operator_auto_fix",
			input:       "my Int||Str $union;",
			expectAST:   true,  // Should recover
			expectError: false, // Should auto-fix
		},
		{
			name:        "double_intersection_auto_fix",
			input:       "my Object&&Serializable $obj;",
			expectAST:   true,  // Should recover
			expectError: false, // Should auto-fix
		},
		{
			name:        "complex_error_partial_recovery",
			input:       "my ArrayRef[Int $var; my Str $valid = \"test\";",
			expectAST:   true,  // Should get partial AST
			expectError: false, // Should recover with auto-fix
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := parser.ParseString(tc.input)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.expectAST && ast == nil {
				t.Error("Expected AST but got nil")
			}
			if !tc.expectAST && ast != nil {
				t.Error("Expected nil AST but got result")
			}

			// Check if AST contains error recovery information
			if ast != nil && len(ast.Errors) > 0 {
				t.Logf("AST contains %d recovery errors/warnings", len(ast.Errors))
				for i, astErr := range ast.Errors {
					t.Logf("  Error %d: %v", i, astErr)
				}
			}
		})
	}
}

func TestEnhancedParserPartialParsing(t *testing.T) {
	parser, err := NewEnhancedParser(true)
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Code with mix of valid and invalid statements
	mixedCode := `
my Int $valid = 42;
my ArrayRef[Int $broken;  // Missing closing bracket
my Str $another = "test";
method test() -> Int { return 42; }
my Invalid||Type $bad;   // Double union operator
`

	ast, err := parser.ParseString(mixedCode)

	// Should not fail completely - should recover
	if err != nil {
		t.Logf("Parse completed with recovery: %v", err)
	}

	if ast == nil {
		t.Error("Expected partial AST from mixed valid/invalid code")
	}

	if ast != nil {
		t.Logf("Parsed AST with %d type annotations", len(ast.TypeAnnotations))
		if len(ast.Errors) > 0 {
			t.Logf("AST contains %d recovery errors", len(ast.Errors))
		}
	}
}

func TestEnhancedParserPositionTracking(t *testing.T) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	code := `my Int $first = 1;
my Str $second = "test";
my ArrayRef[Bool] @flags;`

	ast, err := parser.ParseString(code)
	if err != nil {
		t.Errorf("Failed to parse code: %v", err)
	}

	if ast == nil {
		t.Fatal("Expected non-nil AST")
	}

	// Verify position tracking for type annotations
	if len(ast.TypeAnnotations) > 0 {
		for i, ta := range ast.TypeAnnotations {
			pos := ta.TypeExpression.Start()
			t.Logf("Type annotation %d at line %d, column %d", i, pos.Line, pos.Column)

			// Validate position is reasonable
			if pos.Line < 1 || pos.Line > 3 {
				t.Errorf("Invalid line number %d for annotation %d", pos.Line, i)
			}
			if pos.Column < 1 {
				t.Errorf("Invalid column number %d for annotation %d", pos.Column, i)
			}
		}
	}
}

func TestEnhancedParserCommonErrorFixes(t *testing.T) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	fixTests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "missing_bracket_fix",
			input:    "my ArrayRef[Int $var;",
			expected: "ArrayRef[Int]", // Should be fixed to include closing bracket
		},
		{
			name:     "double_union_fix",
			input:    "my Int||Str $union;",
			expected: "Int|Str", // Should be fixed to single union operator
		},
		{
			name:     "double_intersection_fix",
			input:    "my Object&&Serializable $obj;",
			expected: "Object&Serializable", // Should be fixed to single intersection
		},
	}

	for _, tt := range fixTests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the fix logic directly
			fixedContent, fixes := parser.fixCommonTypeErrors(tt.input)

			if len(fixes) == 0 {
				t.Error("Expected at least one fix to be applied")
			}

			if !strings.Contains(fixedContent, tt.expected) {
				t.Errorf("Expected fixed content to contain %q, got %q", tt.expected, fixedContent)
			}

			t.Logf("Applied %d fixes: %v", len(fixes), fixes)
		})
	}
}

func TestEnhancedParserSegmentSplitting(t *testing.T) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	testCases := []struct {
		name             string
		input            string
		expectedSegments int
	}{
		{
			name:             "simple_statements",
			input:            "my Int $a; my Str $b; my Bool $c;",
			expectedSegments: 3,
		},
		{
			name: "method_with_block",
			input: `method test() {
				my Int $local = 42;
				return $local;
			}`,
			expectedSegments: 1, // Should be one block
		},
		{
			name: "mixed_statements_and_blocks",
			input: `my Int $global = 1;
			method test() { return 42; }
			my Str $another = "test";`,
			expectedSegments: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			segments := parser.splitIntoSegments(tc.input)

			if len(segments) != tc.expectedSegments {
				t.Errorf("Expected %d segments, got %d", tc.expectedSegments, len(segments))
				for i, seg := range segments {
					t.Logf("Segment %d: %q", i, seg)
				}
			}
		})
	}
}

func TestEnhancedParserValidationErrors(t *testing.T) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Test type expression validation
	testExpr := &ast.TypeExpression{
		BaseNode:   ast.NewBaseNode("type_expr", ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 1}),
		IsUnion:    true,
		UnionTypes: []*ast.TypeExpression{}, // Invalid: empty union
	}

	errors := parser.errorRecovery.ValidateTypeExpression(testExpr, "test source")
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid union type")
	}

	for _, err := range errors {
		t.Logf("Validation error: %v", err)
	}
}

func TestEnhancedParserDeepNestingHandling(t *testing.T) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Create deeply nested type expression
	deepType := "ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[HashRef[ArrayRef[Int]]]]]]]]]]]]]]]]]]]]]"
	code := "my " + deepType + " $deep;"

	ast, err := parser.ParseString(code)

	// Should handle gracefully (either parse or provide meaningful error)
	if err != nil {
		t.Logf("Deep nesting handled with error: %v", err)
	}

	if ast != nil && len(ast.Errors) > 0 {
		t.Logf("Deep nesting produced %d warnings/errors", len(ast.Errors))
		for _, astErr := range ast.Errors {
			if typeErr, ok := astErr.(*TypeError); ok && typeErr.ErrorCode == DeepNestingError {
				t.Logf("Correctly detected deep nesting: %v", typeErr)
			}
		}
	}
}

func TestEnhancedParserPositionAccuracy(t *testing.T) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	code := `my Int $count = 42;
my Str $name = "test";
my ArrayRef[Bool] @flags = [];`

	ast, err := parser.ParseString(code)
	if err != nil {
		t.Errorf("Failed to parse test code: %v", err)
	}

	if ast == nil {
		t.Fatal("Expected non-nil AST")
	}

	// Test position accuracy by checking if positions correspond to actual content
	lines := strings.Split(code, "\n")

	for _, ta := range ast.TypeAnnotations {
		if ta.TypeExpression == nil {
			continue
		}

		pos := ta.TypeExpression.Start()
		if pos.Line < 1 || pos.Line > len(lines) {
			t.Errorf("Position line %d out of bounds (1-%d)", pos.Line, len(lines))
			continue
		}

		line := lines[pos.Line-1]
		if pos.Column < 1 || pos.Column > len(line) {
			t.Errorf("Position column %d out of bounds (1-%d) on line %d", pos.Column, len(line), pos.Line)
			continue
		}

		// Check if the position actually points to something type-related
		remaining := line[pos.Column-1:]
		if ta.TypeExpression.Name != "" && !strings.HasPrefix(remaining, ta.TypeExpression.Name) {
			t.Logf("Position accuracy note: Expected %q at %d:%d, found %q",
				ta.TypeExpression.Name, pos.Line, pos.Column, remaining[:min(10, len(remaining))])
		}
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func BenchmarkEnhancedParserValidCode(b *testing.B) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		b.Fatalf("Failed to create enhanced parser: %v", err)
	}

	code := `my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] @numbers = (1, 2, 3);
method calculate(Int $a, Int $b) -> Int {
	return $a + $b;
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseString(code)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

func BenchmarkEnhancedParserErrorRecovery(b *testing.B) {
	parser, err := NewEnhancedParser(false)
	if err != nil {
		b.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Code with errors that need recovery
	code := `my ArrayRef[Int $broken;
my Int||Str $union;
my Object&&Serializable $intersection;`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseString(code) // Ignore errors for benchmark
	}
}
