// ABOUTME: Tests for type annotation spacing normalization in typed Perl compiler
// ABOUTME: Verifies that input without space between type and sigil outputs with proper space

package compiler

import (
	"testing"
)

func TestTypeAnnotationSpacingNormalization(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple type without space should be normalized",
			input:    "sub Int add (Int$a, Int$b) { return $a + $b; }",
			expected: "sub Int add (Int $a, Int $b) { return $a + $b; }",
		},
		{
			name:     "Complex parameterized type without space should be normalized",
			input:    "sub Result[Bool] process (ArrayRef[HashRef[Str]]$data) { return 1; }",
			expected: "sub Result[Bool] process (ArrayRef[HashRef[Str]] $data) { return 1; }",
		},
		{
			name:     "Union type without space should be normalized",
			input:    "sub Str complex (Union[Str, Union[Int, Bool]]$param) { return \"result\"; }",
			expected: "sub Str complex (Union[Str, Union[Int, Bool]] $param) { return \"result\"; }",
		},
		{
			name:     "Intersection and negation types without space should be normalized",
			input:    "sub Bool validate (Object&Serializable$obj, !Undef$config) { return 1; }",
			expected: "sub Bool validate (Object&Serializable $obj, !Undef $config) { return 1; }",
		},
		{
			name:     "Already properly spaced input should remain unchanged",
			input:    "sub Int add (Int $a, Int $b) { return $a + $b; }",
			expected: "sub Int add (Int $a, Int $b) { return $a + $b; }",
		},
		{
			name:     "Variable declarations without space should be normalized",
			input:    "my ComplexTypes$self = bless {}; my ArrayRef[HashRef[Str|Int]]$users = [];",
			expected: "my ComplexTypes $self = bless {}; my ArrayRef[HashRef[Str|Int]] $users = [];",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create typed Perl compiler
			compiler := NewTypedPerlCompilerUnified()

			// Compile the input
			result, err := compiler.CompileString(tc.input)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			// Verify the spacing normalization
			if result != tc.expected {
				t.Errorf("Spacing normalization failed:\nInput:    %s\nExpected: %s\nActual:   %s", tc.input, tc.expected, result)
			}
		})
	}
}

func TestSpaceNormalizationEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Multiple spaces should be normalized to single space",
			input:    "my ArrayRef[Str]  $arr = [];",
			expected: "my ArrayRef[Str] $arr = [];",
		},
		{
			name:     "Mixed spacing scenarios in same line",
			input:    "sub process(Int$a, Str $b, HashRef[Int]$c) { }",
			expected: "sub process(Int $a, Str $b, HashRef[Int] $c) { }",
		},
		{
			name:     "Array and hash sigils should also be normalized",
			input:    "my ArrayRef[Int]@arr = (); my HashRef[Str]%hash = ();",
			expected: "my ArrayRef[Int] @arr = (); my HashRef[Str] %hash = ();",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create typed Perl compiler
			compiler := NewTypedPerlCompilerUnified()

			// Compile the input
			result, err := compiler.CompileString(tc.input)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			// Verify the spacing normalization
			if result != tc.expected {
				t.Errorf("Edge case spacing normalization failed:\nInput:    %s\nExpected: %s\nActual:   %s", tc.input, tc.expected, result)
			}
		})
	}
}
