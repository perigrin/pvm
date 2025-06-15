// ABOUTME: Tests for core variable declarations (untyped Perl) - Step 2 of parser improvements
// ABOUTME: Validates scalar, array, and hash variable declarations with comprehensive coverage

package parser

import (
	"testing"
)

func TestUntypedVariableDeclarations(t *testing.T) {
	// This test now covered by markdown tests
	t.Skip("Untyped variable declarations now tested via TestRunMarkdownTestsByCategory - JSON files removed")
}

func TestScalarDeclarations(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		tags  []string
	}{
		{
			name:  "simple_scalar",
			input: "my $var;",
			tags:  []string{"variables", "scalars", "simple"},
		},
		{
			name:  "scalar_with_assignment",
			input: "my $var = 42;",
			tags:  []string{"variables", "scalars", "assignment"},
		},
		{
			name:  "our_scalar",
			input: "our $global;",
			tags:  []string{"variables", "scalars", "our", "global"},
		},
		{
			name:  "state_scalar",
			input: "state $persistent = \"value\";",
			tags:  []string{"variables", "scalars", "state", "persistent"},
		},
		{
			name:  "local_scalar",
			input: "local $localized;",
			tags:  []string{"variables", "scalars", "local"},
		},
		{
			name:  "package_qualified",
			input: "$Package::var = \"test\";",
			tags:  []string{"variables", "scalars", "package-qualified"},
		},
	}

	framework := NewParserTestFramework("../../test/corpus/parser")
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCase := framework.GenerateTestCase(tc.name, tc.input,
				"Test scalar variable declaration: "+tc.name, UntypedPerl, tc.tags)

			success := framework.RunTestCase(t, testCase)
			if !success {
				t.Errorf("Scalar declaration test failed: %s", tc.name)
			} else {
				t.Logf("Scalar declaration test passed: %s", tc.name)
			}
		})
	}
}

func TestArrayDeclarations(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		tags  []string
	}{
		{
			name:  "simple_array",
			input: "my @array;",
			tags:  []string{"variables", "arrays", "simple"},
		},
		{
			name:  "array_with_assignment",
			input: "my @array = (1, 2, 3);",
			tags:  []string{"variables", "arrays", "assignment"},
		},
		{
			name:  "our_array",
			input: "our @global_array;",
			tags:  []string{"variables", "arrays", "our", "global"},
		},
		{
			name:  "qw_assignment",
			input: "my @words = qw(one two three);",
			tags:  []string{"variables", "arrays", "qw", "assignment"},
		},
		{
			name:  "package_qualified_array",
			input: "@Package::array = (\"a\", \"b\");",
			tags:  []string{"variables", "arrays", "package-qualified"},
		},
	}

	framework := NewParserTestFramework("../../test/corpus/parser")
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCase := framework.GenerateTestCase(tc.name, tc.input,
				"Test array variable declaration: "+tc.name, UntypedPerl, tc.tags)

			success := framework.RunTestCase(t, testCase)
			if !success {
				t.Errorf("Array declaration test failed: %s", tc.name)
			}
		})
	}
}

func TestHashDeclarations(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		tags  []string
	}{
		{
			name:  "simple_hash",
			input: "my %hash;",
			tags:  []string{"variables", "hashes", "simple"},
		},
		{
			name:  "hash_with_assignment",
			input: "my %hash = (key => 'value');",
			tags:  []string{"variables", "hashes", "assignment"},
		},
		{
			name:  "our_hash",
			input: "our %global_hash;",
			tags:  []string{"variables", "hashes", "our", "global"},
		},
		{
			name:  "multi_pair_hash",
			input: "my %config = (host => 'localhost', port => 8080);",
			tags:  []string{"variables", "hashes", "multi-pair", "assignment"},
		},
		{
			name:  "package_qualified_hash",
			input: "%Package::config = (debug => 1);",
			tags:  []string{"variables", "hashes", "package-qualified"},
		},
	}

	framework := NewParserTestFramework("../../test/corpus/parser")
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCase := framework.GenerateTestCase(tc.name, tc.input,
				"Test hash variable declaration: "+tc.name, UntypedPerl, tc.tags)

			success := framework.RunTestCase(t, testCase)
			if !success {
				t.Errorf("Hash declaration test failed: %s", tc.name)
			}
		})
	}
}
