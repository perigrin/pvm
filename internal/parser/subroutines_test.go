// ABOUTME: Test subroutine parsing for comprehensive coverage of Perl subroutine patterns
// ABOUTME: Validates subroutine definitions, calls, methods, prototypes, and references

package parser

import (
	"path/filepath"
	"testing"
)

func TestSubroutineParsing(t *testing.T) {
	framework := NewParserTestFramework("testdata")

	// Set up parser - using the existing parser
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Run all subroutine-related tests
	metrics := framework.RunTestsByCategory(t, UntypedPerl)

	// Print summary for visibility
	framework.PrintMetricsSummary(t, metrics)

	// Save metrics for tracking
	metricsPath := filepath.Join("testdata", "reports", "subroutines_metrics.json")
	err = framework.SaveMetricsReport(metrics, metricsPath)
	if err != nil {
		t.Logf("Warning: Failed to save metrics report: %v", err)
	}

	// Require reasonable success rate
	if metrics.TotalTests > 0 {
		accuracy := float64(metrics.PassedTests) / float64(metrics.TotalTests) * 100
		if accuracy < 80.0 {
			t.Errorf("Subroutine parsing accuracy too low: %.1f%% (expected >= 80%%)", accuracy)
		}
	}
}

func TestBasicSubroutineDefinitions(t *testing.T) {
	framework := NewParserTestFramework("testdata")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	testCase := framework.GenerateTestCase(
		"basic_subroutine_definition",
		"sub hello { return 'world'; }",
		"Basic subroutine definition with return statement",
		UntypedPerl,
		[]string{"subroutines", "basic", "definitions"},
	)

	success := framework.RunTestCase(t, testCase)
	if !success {
		t.Errorf("Basic subroutine definition test failed")
	}
}

func TestAnonymousSubroutines(t *testing.T) {
	framework := NewParserTestFramework("testdata")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	testCase := framework.GenerateTestCase(
		"anonymous_subroutine",
		"my $sub = sub { my $x = shift; return $x * 2; };",
		"Anonymous subroutine assigned to variable",
		UntypedPerl,
		[]string{"subroutines", "anonymous", "code_references"},
	)

	success := framework.RunTestCase(t, testCase)
	if !success {
		t.Errorf("Anonymous subroutine test failed")
	}
}

func TestSubroutineCalls(t *testing.T) {
	framework := NewParserTestFramework("testdata")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	testCase := framework.GenerateTestCase(
		"subroutine_calls",
		"my $result = function_name(42, 'hello'); my $no_parens = other_function;",
		"Subroutine calls with and without parentheses",
		UntypedPerl,
		[]string{"subroutines", "calls", "arguments"},
	)

	success := framework.RunTestCase(t, testCase)
	if !success {
		t.Errorf("Subroutine calls test failed")
	}
}

func TestMethodCalls(t *testing.T) {
	framework := NewParserTestFramework("testdata")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	testCase := framework.GenerateTestCase(
		"method_calls",
		"my $obj = Package->new(); $obj->method(1, 2); my $result = $obj->chain()->result();",
		"Object method calls and method chaining",
		UntypedPerl,
		[]string{"subroutines", "methods", "arrow_notation", "objects"},
	)

	success := framework.RunTestCase(t, testCase)
	if !success {
		t.Errorf("Method calls test failed")
	}
}

func TestSubroutineReferences(t *testing.T) {
	framework := NewParserTestFramework("testdata")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	testCase := framework.GenerateTestCase(
		"subroutine_references",
		"my $ref = \\&function; my $result = $ref->(42); my $anon = sub { 'test' };",
		"Subroutine references and code reference calls",
		UntypedPerl,
		[]string{"subroutines", "references", "code_references", "dereferencing"},
	)

	success := framework.RunTestCase(t, testCase)
	if !success {
		t.Errorf("Subroutine references test failed")
	}
}
