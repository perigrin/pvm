// ABOUTME: Tests for tree-sitter parser implementation
// ABOUTME: Verifies the functionality of the tree-sitter parser

package treesitter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTypeExpression(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectBaseType string
		expectParams   int
		expectUnion    bool
		expectError    bool
	}{
		{
			name:           "Simple type",
			input:          "Int",
			expectBaseType: "Int",
			expectParams:   0,
			expectUnion:    false,
			expectError:    false,
		},
		{
			name:           "Parameterized type",
			input:          "ArrayRef[Int]",
			expectBaseType: "ArrayRef",
			expectParams:   1,
			expectUnion:    false,
			expectError:    false,
		},
		{
			name:           "Complex parameterized type",
			input:          "HashRef[Str, ArrayRef[Int]]",
			expectBaseType: "HashRef",
			expectParams:   2,
			expectUnion:    false,
			expectError:    false,
		},
		{
			name:           "Union type",
			input:          "Int|Str",
			expectBaseType: "",
			expectParams:   0,
			expectUnion:    true,
			expectError:    false,
		},
		{
			name:           "Intersection type",
			input:          "Foo&Bar",
			expectBaseType: "",
			expectParams:   0,
			expectUnion:    false,
			expectError:    false,
		},
		{
			name:           "Negation type",
			input:          "!Int",
			expectBaseType: "",
			expectParams:   0,
			expectUnion:    false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := Position{Line: 1, Column: 1}
			expr, err := ParseTypeExpression(tt.input, pos)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, expr)

			if tt.expectUnion {
				assert.True(t, expr.IsUnion, "Expected union type")
				assert.Greater(t, len(expr.UnionTypes), 0, "Expected union types")
			} else {
				assert.Equal(t, tt.expectBaseType, expr.BaseType, "Base type mismatch")
				assert.Equal(t, tt.expectParams, len(expr.Parameters), "Parameter count mismatch")
			}

			// Check that the expression string matches the input
			assert.Equal(t, tt.input, expr.String(), "String representation mismatch")
		})
	}
}

func TestParseWithMockData(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "parser-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a sample Perl file with type annotations
	sampleFile := filepath.Join(tempDir, "sample.pl")
	sampleContent := `use v5.36;

my Int $count = 0;
my Str $name = "Example";
my ArrayRef[Int] $numbers = [1, 2, 3];

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

method greet(Str $name) returns Str {
    return "Hello, $name!";
}

field Bool $flag;

type ID = Int;
type Names = ArrayRef[Str];
`

	err = os.WriteFile(sampleFile, []byte(sampleContent), 0644)
	require.NoError(t, err)

	// Test the tree-sitter-perl integration
	// Let's see if it works now!

	// Create a parser
	parser, err := NewParser(true)
	require.NoError(t, err)
	defer parser.Close()

	// Parse the file
	ast, err := parser.ParseFile(sampleFile)
	require.NoError(t, err)
	require.NotNil(t, ast)

	// Check that we found at least some type annotations
	assert.Greater(t, len(ast.TypeAnnotations), 0, "Expected type annotations to be found")
}

// TestSplitParams tests the splitParams function
func TestSplitParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple parameters",
			input:    "Int, Str, Bool",
			expected: []string{"Int", " Str", " Bool"},
		},
		{
			name:     "Nested parameters",
			input:    "Int, ArrayRef[Int], HashRef[Str, Int]",
			expected: []string{"Int", " ArrayRef[Int]", " HashRef[Str, Int]"},
		},
		{
			name:     "Complex nested parameters",
			input:    "Int, HashRef[Str, ArrayRef[Int | Num]]",
			expected: []string{"Int", " HashRef[Str, ArrayRef[Int | Num]]"},
		},
		{
			name:     "Single parameter",
			input:    "Int",
			expected: []string{"Int"},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitParams(tt.input)
			assert.Equal(t, tt.expected, result, "Unexpected split result")
		})
	}
}
