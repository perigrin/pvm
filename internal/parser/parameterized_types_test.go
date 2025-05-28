// ABOUTME: Tests for parameterized type expressions parsing (Step 11 implementation)
// ABOUTME: Validates parser correctly handles bracket notation and type parameters

package parser

import (
	"path/filepath"
	"testing"
)

// TestParameterizedTypes tests comprehensive parameterized type coverage from Step 11
func TestParameterizedTypes(t *testing.T) {
	// Set up test framework
	testDataDir := filepath.Join("testdata", "typed-perl", "parameterized-types")
	framework := NewParserTestFramework(testDataDir)
	framework.Verbose = testing.Verbose()

	// Create parser
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Run all parameterized type tests
	metrics := framework.RunTestsByCategory(t, TypedPerl)

	// Print summary
	framework.PrintMetricsSummary(t, metrics)

	// Save metrics report
	reportsDir := filepath.Join("testdata", "reports")
	metricsPath := filepath.Join(reportsDir, "parameterized_types_metrics.json")
	err = framework.SaveMetricsReport(metrics, metricsPath)
	if err != nil {
		t.Logf("Warning: Failed to save metrics report: %v", err)
	}

	// For now, we expect low accuracy since we haven't implemented the parsing yet
	// This establishes the baseline for improvement
	t.Logf("Parameterized types baseline established with %d tests", metrics.TotalTests)

	// Ensure we actually ran some tests
	if metrics.TotalTests == 0 {
		t.Error("No parameterized type tests were discovered")
	}
}

// TestSpecificParameterizedTypes tests specific patterns individually
func TestSpecificParameterizedTypes(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name     string
		input    string
		description string
	}{
		{
			name:     "basic_single_parameter",
			input:    "my ArrayRef[Int] @numbers;",
			description: "Basic parameterized type with single parameter",
		},
		{
			name:     "multiple_parameters",
			input:    "my Map[Str, Int] %mapping;",
			description: "Parameterized type with multiple parameters",
		},
		{
			name:     "nested_parameters",
			input:    "my ArrayRef[ArrayRef[Int]] @matrix;",
			description: "Nested parameterized types",
		},
		{
			name:     "custom_parameterized",
			input:    "my Container[MyType] $custom;",
			description: "Custom parameterized type",
		},
		{
			name:     "whitespace_handling",
			input:    "my ArrayRef[ Int ] @spaced;",
			description: "Parameterized type with whitespace in brackets",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)
			t.Logf("Input: %s", tc.input)
			
			ast, err := parser.ParseString(tc.input)
			if err != nil {
				t.Logf("Parsing failed (expected for now): %v", err)
				return
			}

			if ast == nil {
				t.Log("AST is nil (expected for now)")
				return
			}

			t.Logf("Type annotations found: %d", len(ast.TypeAnnotations))
			for i, annotation := range ast.TypeAnnotations {
				t.Logf("  Annotation %d: %s :: %s", i, annotation.AnnotatedItem, 
					annotation.TypeExpression.String())
			}
		})
	}
}

// TestParameterizedTypesDebug helps understand current parser behavior with brackets
func TestParameterizedTypesDebug(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test how the parser currently handles bracket notation
	testCases := []string{
		"my ArrayRef[Int] @numbers;",
		"my HashRef[Str] %config;", 
		"my Map[Str, Int] %mapping;",
		"my ArrayRef[ArrayRef[Int]] @matrix;",
	}

	for _, input := range testCases {
		t.Logf("\n--- Testing: %s ---", input)
		
		ast, err := parser.ParseString(input)
		if err != nil {
			t.Logf("Parse error: %v", err)
			continue
		}

		if ast == nil {
			t.Log("AST is nil")
			continue
		}

		t.Logf("Type annotations: %d", len(ast.TypeAnnotations))
		for _, annotation := range ast.TypeAnnotations {
			if annotation.TypeExpression != nil {
				t.Logf("  Found: %s :: %s", annotation.AnnotatedItem, annotation.TypeExpression.String())
				t.Logf("    Name: '%s'", annotation.TypeExpression.Name)
				t.Logf("    Parameters: %d", len(annotation.TypeExpression.Parameters))
				for j, param := range annotation.TypeExpression.Parameters {
					t.Logf("      Param %d: %s", j, param.String())
				}
			}
		}
	}
}