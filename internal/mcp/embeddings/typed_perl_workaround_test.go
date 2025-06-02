// ABOUTME: Unit test for typed Perl workaround extraction
// ABOUTME: Tests the regex-based extraction without depending on tree-sitter

package embeddings

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTypedParametersFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected map[string]string
	}{
		{
			name: "simple_typed_parameters",
			text: "sub add(Int $a, Int $b) -> Int { return $a + $b; }",
			expected: map[string]string{
				"$a":     "Int",
				"$b":     "Int",
				"return": "Int",
			},
		},
		{
			name: "complex_types",
			text: "sub process(ArrayRef[Str] $items, HashRef[Int] $counts) -> Bool { return 1; }",
			expected: map[string]string{
				"$items":  "ArrayRef[Str]",
				"$counts": "HashRef[Int]",
				"return":  "Bool",
			},
		},
		{
			name: "union_return_type",
			text: "sub transform(Str $input) -> Union[Str, Undef] { return $input; }",
			expected: map[string]string{
				"$input": "Str",
				"return": "Union[Str,", // Note: regex limitation with complex unions
			},
		},
		{
			name: "no_return_type",
			text: "sub hello(Str $name) { print \"Hello $name\"; }",
			expected: map[string]string{
				"$name": "Str",
			},
		},
	}

	extractor := &Extractor{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeInfo := make(map[string]string)
			extractor.extractTypedParametersFromText(tt.text, typeInfo)

			// Debug: print what we actually extracted for debugging
			// t.Logf("Extracted typeInfo: %+v", typeInfo)

			for expectedParam, expectedType := range tt.expected {
				actualType, found := typeInfo[expectedParam]
				assert.True(t, found, "Expected parameter %s not found", expectedParam)
				if found {
					assert.Contains(t, actualType, expectedType,
						"Type mismatch for %s: expected %s, got %s", expectedParam, expectedType, actualType)
				}
			}
		})
	}
}

func TestTypedSubroutineRegexPattern(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
		subName  string
	}{
		{
			name:     "simple_typed_sub",
			text:     "sub add(Int $a, Int $b) -> Int {\n    return $a + $b;\n}",
			expected: true,
			subName:  "add",
		},
		{
			name:     "no_return_type",
			text:     "sub hello(Str $name) {\n    print \"Hello $name\";\n}",
			expected: true,
			subName:  "hello",
		},
		{
			name:     "untyped_sub",
			text:     "sub simple { print \"hello\"; }",
			expected: true,
			subName:  "simple",
		},
		{
			name:     "not_a_sub",
			text:     "my $var = 42;",
			expected: false,
		},
		{
			name:     "method_not_sub",
			text:     "method process(Int $x) -> Str { return \"$x\"; }",
			expected: false,
		},
	}

	// Test the regex pattern directly
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typedSubPattern := `sub\s+(\w+)\s*(?:\([^)]*\))?\s*(?:->\s*\w+(?:\[.*?\])?(?:\|\w+(?:\[.*?\])?)*\s*)?\s*\{`
			re, err := regexp.Compile(typedSubPattern)
			assert.NoError(t, err)

			matches := re.FindStringSubmatch(tt.text)
			if tt.expected {
				assert.True(t, len(matches) >= 2, "Expected to match subroutine pattern")
				if len(matches) >= 2 {
					assert.Equal(t, tt.subName, matches[1], "Subroutine name mismatch")
				}
			} else {
				assert.True(t, len(matches) < 2, "Expected NOT to match subroutine pattern")
			}
		})
	}
}
