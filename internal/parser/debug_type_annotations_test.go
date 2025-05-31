// ABOUTME: Debug tests to understand current type annotation detection capabilities
// ABOUTME: Helps identify what the parser currently extracts vs what we expect

package parser

import (
	"testing"
)

// TestDebugTypeAnnotations helps understand what the parser currently detects
func TestDebugTypeAnnotations(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "simple_scalar",
			input: "my Int $count = 42;",
		},
		{
			name:  "parameterized_array",
			input: "my ArrayRef[Int] @numbers = (1, 2, 3);",
		},
		{
			name:  "package_qualified",
			input: "my Package::Type $qualified;",
		},
		{
			name:  "hash_ref",
			input: "my HashRef[Str] %config = (key => 'value');",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := parser.ParseString(tc.input)
			if err != nil {
				t.Fatalf("Parsing failed: %v", err)
			}

			t.Logf("Input: %s", tc.input)
			t.Logf("Type annotations count: %d", len(ast.TypeAnnotations))

			for i, annotation := range ast.TypeAnnotations {
				t.Logf("  Annotation %d:", i)
				t.Logf("    Kind: %v", annotation.Kind)
				t.Logf("    AnnotatedItem: '%s'", annotation.AnnotatedItem)
				if annotation.TypeExpression != nil {
					t.Logf("    TypeExpression: '%s'", annotation.TypeExpression.String())
					t.Logf("    TypeExpression.Name: '%s'", annotation.TypeExpression.Name)
					t.Logf("    TypeExpression.Parameters count: %d", len(annotation.TypeExpression.Parameters))
					for j, param := range annotation.TypeExpression.Parameters {
						t.Logf("      Param %d: '%s'", j, param.String())
					}
				} else {
					t.Logf("    TypeExpression: nil")
				}
				t.Logf("    Position: %d:%d", annotation.Pos.Line, annotation.Pos.Column)
			}
			t.Logf("")
		})
	}
}
