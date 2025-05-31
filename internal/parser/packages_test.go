// ABOUTME: Test package and module construct parsing for untyped Perl
// ABOUTME: Validates comprehensive coverage of package declarations, use/require statements, and module organization

package parser

import (
	"testing"
)

func TestPackageAndModuleConstructs(t *testing.T) {
	framework := NewParserTestFramework("testdata")

	// Initialize the parser
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Run package-specific tests
	metrics := framework.RunTestsByCategory(t, UntypedPerl)

	// Print summary for debugging
	framework.PrintMetricsSummary(t, metrics)

	// Validate we have reasonable coverage
	if metrics.TotalTests == 0 {
		t.Error("No package tests found")
	}

	// Check that we have high accuracy for package parsing
	overallAccuracy := float64(metrics.PassedTests) / float64(metrics.TotalTests) * 100
	if overallAccuracy < 90.0 {
		t.Errorf("Package parsing accuracy too low: %.1f%% (expected >= 90%%)", overallAccuracy)
	}
}

func TestSpecificPackageFeatures(t *testing.T) {
	framework := NewParserTestFramework("testdata")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Test specific package-related features
	testCases := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:  "simple_package_declaration",
			input: "package MyPackage;",
		},
		{
			name:  "package_with_version",
			input: "package MyPackage 1.23;",
		},
		{
			name:  "nested_package_declaration",
			input: "package My::Nested::Package;",
		},
		{
			name:  "basic_use_statement",
			input: "use strict;",
		},
		{
			name:  "use_with_import_list",
			input: "use Data::Dumper qw(Dumper);",
		},
		{
			name:  "use_with_version",
			input: "use MyModule 1.5;",
		},
		{
			name:  "require_statement",
			input: "require MyModule;",
		},
		{
			name:  "package_qualified_variable",
			input: "$Package::variable = 'value';",
		},
		{
			name:  "package_qualified_function_call",
			input: "Package::function();",
		},
		{
			name:  "perl_version_requirement",
			input: "use 5.010;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCase := framework.GenerateTestCase(
				tc.name,
				tc.input,
				"Test "+tc.name,
				UntypedPerl,
				[]string{"packages"},
			)
			testCase.ShouldError = tc.shouldError

			success := framework.RunTestCase(t, testCase)
			if !success {
				t.Errorf("Test case %s failed", tc.name)
			}
		})
	}
}
