// ABOUTME: Integration tests for Markdown and JSON test case discovery and execution
// ABOUTME: Validates that the framework can run tests from both formats together

package parser

import (
	"testing"
)

func TestMixedFormatDiscovery(t *testing.T) {
	framework := NewParserTestFramework("../../test/corpus/parser")

	testCases, err := framework.DiscoverTestCases()
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	if len(testCases) == 0 {
		t.Fatal("No test cases discovered")
	}

	// Count cases by format
	var markdownCases, jsonCases int
	var typedPerlCases, errorCases int

	for _, tc := range testCases {
		// Determine source format by checking for markdown-style names
		if tc.Name == "simple-annotations_basic_typed_variables" ||
			tc.Name == "malformed-types_missing_closing_bracket" {
			markdownCases++
		} else {
			jsonCases++
		}

		// Count by category
		switch tc.Category {
		case TypedPerl:
			typedPerlCases++
		case ErrorCases:
			errorCases++
		}
	}

	t.Logf("Discovered %d total test cases", len(testCases))
	t.Logf("  Markdown-sourced: %d", markdownCases)
	t.Logf("  JSON-sourced: %d", jsonCases)
	t.Logf("  Typed Perl: %d", typedPerlCases)
	t.Logf("  Error Cases: %d", errorCases)

	if markdownCases == 0 {
		t.Error("Expected to find some Markdown-sourced test cases")
	}

	if jsonCases == 0 {
		t.Error("Expected to find some JSON-sourced test cases")
	}
}

func TestRunMarkdownTestsByCategory(t *testing.T) {
	framework := NewParserTestFramework("../../test/corpus/parser")

	// Run tests from our markdown files specifically
	metrics := framework.RunTestsByCategory(t, TypedPerl)
	if metrics == nil {
		t.Fatal("No metrics returned")
	}

	t.Logf("Typed Perl test metrics:")
	t.Logf("  Total: %d", metrics.TotalTests)
	t.Logf("  Passed: %d", metrics.PassedTests)
	t.Logf("  Failed: %d", metrics.FailedTests)

	// We should have some tests from our new markdown files
	if metrics.TotalTests == 0 {
		t.Error("Expected to find some typed Perl tests")
	}
}

func TestRunUntypedPerlMarkdownTests(t *testing.T) {
	framework := NewParserTestFramework("../../test/corpus/parser")

	// Run tests from untyped perl markdown files
	metrics := framework.RunTestsByCategory(t, UntypedPerl)
	if metrics == nil {
		t.Fatal("No metrics returned")
	}

	t.Logf("Untyped Perl test metrics:")
	t.Logf("  Total: %d", metrics.TotalTests)
	t.Logf("  Passed: %d", metrics.PassedTests)
	t.Logf("  Failed: %d", metrics.FailedTests)

	// We should have tests from expressions.md and control-flow.md
	if metrics.TotalTests == 0 {
		t.Error("Expected to find some untyped Perl tests")
	}
}
