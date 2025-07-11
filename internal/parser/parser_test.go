// ABOUTME: Consolidated parser test suite that auto-discovers and runs all test categories
// ABOUTME: Provides parallel execution at both category and individual test case levels

package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestParserByCategory runs all parser tests organized by category with full parallelization
func TestParserByCategory(t *testing.T) {
	framework := NewParserTestFramework("../../testdata/corpus/parser")

	// Initialize parser
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser
	framework.Verbose = testing.Verbose()

	// Auto-discover test categories based on directory structure
	categories := []struct {
		name     string
		category TestCategory
		path     string
	}{
		{"typed-perl", TypedPerl, "typed-perl"},
		{"untyped-perl", UntypedPerl, "untyped-perl"},
		{"error-cases", ErrorCases, "error-cases"},
	}

	for _, cat := range categories {
		cat := cat // capture loop variable
		t.Run(cat.name, func(t *testing.T) {
			t.Parallel() // Parallelize categories
			framework.RunTestsByCategory(t, cat.category)
		})
	}
}

// TestSpecificFeatures runs targeted tests for specific parsing features
func TestSpecificFeatures(t *testing.T) {
	framework := NewParserTestFramework("../../testdata/corpus/parser")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Test specific features that need individual validation
	featureTests := []struct {
		name        string
		input       string
		shouldError bool
		category    TestCategory
	}{
		{
			name:        "basic_variable_declaration",
			input:       "my $var = 42;",
			shouldError: false,
			category:    UntypedPerl,
		},
		{
			name:        "typed_variable_declaration",
			input:       "my Int $count = 42;",
			shouldError: false,
			category:    TypedPerl,
		},
		{
			name:        "simple_subroutine",
			input:       "sub hello { print \"Hello, World!\\n\"; }",
			shouldError: false,
			category:    UntypedPerl,
		},
		{
			name:        "typed_method",
			input:       "method Int process(Str $input) { return length($input); }",
			shouldError: false,
			category:    TypedPerl,
		},
		{
			name:        "malformed_type",
			input:       "my ArrayRef[Int $broken;",
			shouldError: false, // Parser now recovers from missing closing bracket
			category:    TypedPerl,
		},
	}

	for _, test := range featureTests {
		test := test // capture loop variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel() // Parallelize individual feature tests

			testCase := framework.GenerateTestCase(
				test.name,
				test.input,
				"Feature test: "+test.name,
				test.category,
				[]string{"feature-test"},
			)
			testCase.ShouldError = test.shouldError

			framework.RunTestCase(t, testCase)
		})
	}
}

// TestTreeSitterIntegration validates tree-sitter parser integration
func TestTreeSitterIntegration(t *testing.T) {
	integrationTests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple_perl_script",
			input: "#!/usr/bin/perl\nuse strict;\nprint \"Hello, World!\\n\";",
		},
		{
			name:  "typed_perl_script",
			input: "my Int $number = 42;\nmy Str $text = \"hello\";",
		},
		{
			name:  "complex_type_expression",
			input: "my ArrayRef[HashRef[Str, Int]] $data = [];",
		},
	}

	for _, test := range integrationTests {
		test := test // capture loop variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel() // Parallelize integration tests

			// Use parser pool for thread safety in parallel tests
			parser := GlobalParserPool.Get()
			if parser == nil {
				// Fallback: create new parser if pool is empty
				var err error
				parser, err = NewParser()
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}
			}
			defer GlobalParserPool.Put(parser)

			ast, err := parser.ParseString(test.input)
			if err != nil {
				t.Errorf("Failed to parse %s: %v", test.name, err)
				return
			}

			if ast == nil {
				t.Errorf("Parser returned nil AST for %s", test.name)
				return
			}

			if ast.Source != test.input {
				t.Errorf("AST source mismatch for %s", test.name)
			}

			t.Logf("Successfully parsed %s: AST has %d type annotations",
				test.name, len(ast.TypeAnnotations))
		})
	}
}

// TestErrorRecovery validates parser error handling and recovery
func TestErrorRecovery(t *testing.T) {
	errorTests := []struct {
		name        string
		input       string
		expectError bool
		errorType   string
	}{
		{
			name:        "missing_semicolon",
			input:       "my $var = 42",
			expectError: false, // Perl doesn't always require semicolons
		},
		{
			name:        "malformed_type_bracket",
			input:       "my ArrayRef[Int $broken;",
			expectError: false, // Parser now recovers from missing closing bracket
		},
		{
			name:        "invalid_union_syntax",
			input:       "my Int||Str $bad_union;",
			expectError: true,
			errorType:   "error[TSP003]",
		},
		{
			name:        "incomplete_type_assertion",
			input:       "my $val = $input as ;",
			expectError: true,
			errorType:   "error[TSP004]",
		},
	}

	for _, test := range errorTests {
		test := test // capture loop variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel() // Parallelize error tests

			// Use parser pool for thread safety in parallel tests
			parser := GlobalParserPool.Get()
			if parser == nil {
				// Fallback: create new parser if pool is empty
				var err error
				parser, err = NewParser()
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}
			}
			defer GlobalParserPool.Put(parser)

			ast, err := parser.ParseString(test.input)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but parsing succeeded", test.name)
					return
				}
				if test.errorType != "" && !strings.Contains(err.Error(), test.errorType) {
					t.Errorf("Expected error type '%s' for %s but got: %v",
						test.errorType, test.name, err)
				}
				t.Logf("Successfully caught expected error for %s: %v", test.name, err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", test.name, err)
					return
				}
				if ast == nil {
					t.Errorf("Parser returned nil AST for %s", test.name)
				}
			}
		})
	}
}

// TestMarkdownTestDiscovery validates the test discovery mechanism
func TestMarkdownTestDiscovery(t *testing.T) {
	framework := NewParserTestFramework("../../testdata/corpus/parser")

	t.Run("discovery_statistics", func(t *testing.T) {
		testCases, err := framework.DiscoverTestCases()
		if err != nil {
			t.Fatalf("Failed to discover test cases: %v", err)
		}

		if len(testCases) == 0 {
			t.Fatal("No test cases discovered")
		}

		// Count by category
		categoryCount := make(map[TestCategory]int)
		for _, tc := range testCases {
			categoryCount[tc.Category]++
		}

		t.Logf("Discovered %d total test cases", len(testCases))
		for category, count := range categoryCount {
			t.Logf("  %s: %d test cases", category, count)
		}

		// Ensure we have a reasonable distribution
		if categoryCount[TypedPerl] == 0 {
			t.Error("No TypedPerl test cases discovered")
		}
		if categoryCount[UntypedPerl] == 0 {
			t.Error("No UntypedPerl test cases discovered")
		}
	})

	t.Run("markdown_loading", func(t *testing.T) {
		// Test loading a specific markdown file
		testFile := filepath.Join("../../testdata/corpus/parser", "typed-perl", "simple-annotations", "basic-typed-variables.md")
		testCases, err := framework.LoadMarkdownTestCases(testFile)
		if err != nil {
			t.Fatalf("Failed to load markdown test cases: %v", err)
		}

		if len(testCases) == 0 {
			t.Fatal("No test cases loaded from markdown file")
		}

		// Validate first test case structure
		firstCase := testCases[0]
		if firstCase.Category != TypedPerl {
			t.Errorf("Expected category %s, got %s", TypedPerl, firstCase.Category)
		}
		if firstCase.Input == "" {
			t.Error("Test case has empty input")
		}
		if firstCase.Name == "" {
			t.Error("Test case has empty name")
		}

		t.Logf("Successfully loaded %d test cases from markdown", len(testCases))
	})
}
