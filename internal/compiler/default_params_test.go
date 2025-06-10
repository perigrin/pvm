package compiler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/parser/treesitter"
)

func TestCompilerStripDefaultParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "preserve numeric default",
			input:    `sub new(Str $class, Num $initial_value = 0) { return 1; }`,
			expected: `sub new($class, $initial_value = 0) { return 1; }`,
		},
		{
			name:     "preserve string default",
			input:    `sub greet(Str $name = "world") { print "Hello, $name!"; }`,
			expected: `sub greet($name = "world") { print "Hello, $name!"; }`,
		},
		{
			name:     "preserve multiple defaults",
			input:    `sub process(Int $a, Str $b = "default", Bool $flag = 1) { }`,
			expected: `sub process($a, $b = "default", $flag = 1) { }`,
		},
		{
			name:     "preserve float default",
			input:    `sub calculate(Num $value = 3.14) { return $value * 2; }`,
			expected: `sub calculate($value = 3.14) { return $value * 2; }`,
		},
		{
			name:     "strip return type with defaults",
			input:    `sub add(Int $a, Int $b = 0) -> Int { return $a + $b; }`,
			expected: `sub add($a, $b = 0) { return $a + $b; }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := treesitter.NewParser(false)
			require.NoError(t, err)

			astResult, err := parser.ParseString(tt.input)
			require.NoError(t, err)

			// Use the compiler to strip types
			astCompiler := NewASTCompiler()
			result, err := astCompiler.Compile(astResult)
			require.NoError(t, err)

			// Normalize whitespace for comparison
			expected := strings.TrimSpace(tt.expected)
			actual := strings.TrimSpace(result)

			// Remove the "use v5.36;" pragma if present for comparison
			if strings.HasPrefix(actual, "use v5.36;") {
				actual = strings.TrimSpace(strings.TrimPrefix(actual, "use v5.36;"))
			}

			assert.Equal(t, expected, actual)
		})
	}
}
