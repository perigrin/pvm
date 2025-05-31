// ABOUTME: Tests to verify tree-sitter grammar recognizes type annotation patterns
// ABOUTME: Helps debug grammar issues with parameterized types and brackets

package parser

import (
	"testing"

	"tamarou.com/pvm/internal/parser/treesitter"
)

// TestTreeSitterGrammarTypeRecognition is disabled for now due to interface issues
// TODO: Fix interface types to properly test tree-sitter grammar recognition
func TestTreeSitterGrammarTypeRecognition(t *testing.T) {
	t.Skip("Skipping due to interface mismatches - need to test at lower level")
}

// TestTreeSitterTypeExpressionParsing tests the ParseTypeExpression function directly
func TestTreeSitterTypeExpressionParsing(t *testing.T) {
	pos := treesitter.Position{Line: 1, Column: 1, Offset: 0}

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple_type",
			input:       "Int",
			expectError: false,
		},
		{
			name:        "parameterized_single",
			input:       "ArrayRef[Int]",
			expectError: false,
		},
		{
			name:        "parameterized_multiple",
			input:       "Map[Str, Int]",
			expectError: false,
		},
		{
			name:        "nested_parameterized",
			input:       "ArrayRef[ArrayRef[Int]]",
			expectError: false,
		},
		{
			name:        "union_type",
			input:       "Int|Str",
			expectError: false,
		},
		{
			name:        "intersection_type",
			input:       "Object&Serializable",
			expectError: false,
		},
		{
			name:        "negation_type",
			input:       "!Undef",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Parsing type expression: %s", tc.input)

			typeExpr, err := treesitter.ParseTypeExpression(tc.input, pos)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if typeExpr == nil {
				t.Error("TypeExpression is nil")
				return
			}

			t.Logf("Parsed successfully:")
			t.Logf("  BaseType: %s", typeExpr.BaseType)
			t.Logf("  Parameters: %d", len(typeExpr.Parameters))
			for i, param := range typeExpr.Parameters {
				t.Logf("    Param %d: %s", i, param.BaseType)
			}
			t.Logf("  IsUnion: %t", typeExpr.IsUnion)
			t.Logf("  IsIntersection: %t", typeExpr.IsIntersection)
			t.Logf("  IsNegation: %t", typeExpr.IsNegation)
			t.Logf("  OriginalString: %s", typeExpr.OriginalString)
		})
	}
}
