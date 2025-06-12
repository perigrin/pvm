// ABOUTME: Tests for union type expressions parsing (Step 10 implementation)
// ABOUTME: Validates parser correctly handles union type syntax (Type1|Type2) in various contexts

package parser

import (
	"path/filepath"
	"testing"
)

// TestUnionTypes tests the comprehensive union type coverage from Step 10
func TestUnionTypes(t *testing.T) {
	t.Skip("Union type tests now covered by TestRunMarkdownTestsByCategory - JSON files removed to avoid duplication")
	// Set up test framework
	testDataDir := filepath.Join("testdata", "typed-perl", "union-types")
	framework := NewParserTestFramework(testDataDir)
	framework.Verbose = testing.Verbose()

	// Create parser
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Run all union type tests
	metrics := framework.RunTestsByCategory(t, TypedPerl)

	// Print summary
	framework.PrintMetricsSummary(t, metrics)

	// Save metrics report
	reportsDir := filepath.Join("testdata", "reports")
	metricsPath := filepath.Join(reportsDir, "union_types_metrics.json")
	err = framework.SaveMetricsReport(metrics, metricsPath)
	if err != nil {
		t.Logf("Warning: Failed to save metrics report: %v", err)
	}

	// Validate we have reasonable accuracy
	expectedMinAccuracy := 70.0 // Start with 70% as we're implementing new features
	overallAccuracy := float64(metrics.PassedTests) / float64(metrics.TotalTests) * 100

	if overallAccuracy < expectedMinAccuracy {
		t.Errorf("Union type accuracy %.1f%% is below expected minimum %.1f%%",
			overallAccuracy, expectedMinAccuracy)
	}

	// Ensure we actually ran some tests
	if metrics.TotalTests == 0 {
		t.Error("No union type tests were discovered")
	}

	// Log individual test results for debugging
	if testing.Verbose() {
		t.Logf("Union types test completed: %d total, %d passed, %d failed",
			metrics.TotalTests, metrics.PassedTests, metrics.FailedTests)
	}
}

// TestUnionTypeVariations tests specific union type parsing scenarios
func TestUnionTypeVariations(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name        string
		input       string
		shouldError bool
		description string
	}{
		{
			name:        "simple_two_type_union",
			input:       "my Int|Str $var;",
			shouldError: false,
			description: "Basic union of two built-in types",
		},
		{
			name:        "three_type_union",
			input:       "my Int|Str|Bool $var;",
			shouldError: false,
			description: "Union of three built-in types",
		},
		{
			name:        "custom_type_union",
			input:       "my MyClass|OtherClass $var;",
			shouldError: false,
			description: "Union of custom type names",
		},
		{
			name:        "package_qualified_union",
			input:       "my Package::Type1|Package::Type2 $var;",
			shouldError: false,
			description: "Union with package-qualified types",
		},
		{
			name:        "whitespace_variations",
			input:       "my Int | Str | Bool $var;",
			shouldError: false,
			description: "Union with spaces around pipe operators",
		},
		{
			name:        "method_parameter_union",
			input:       "method test(Int|Str $param) { }",
			shouldError: false,
			description: "Union type in method parameter",
		},
		{
			name:        "method_return_union",
			input:       "method test() returns Int|Str { }",
			shouldError: false,
			description: "Union type in method return type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := parser.ParseString(tc.input)

			if tc.shouldError && err == nil {
				t.Errorf("Expected error for input %q but parsing succeeded", tc.input)
			} else if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error for input %q: %v", tc.input, err)
			} else if !tc.shouldError && ast == nil {
				t.Errorf("Expected AST for input %q but got nil", tc.input)
			}

			if testing.Verbose() && !tc.shouldError && err == nil {
				t.Logf("✓ %s: %s", tc.description, tc.input)
			}
		})
	}
}
