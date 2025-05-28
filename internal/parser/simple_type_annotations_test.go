// ABOUTME: Tests for simple type annotations parsing (Step 7 implementation)
// ABOUTME: Validates parser correctly handles basic type annotations on variable declarations

package parser

import (
	"path/filepath"
	"testing"
)

// TestSimpleTypeAnnotations tests the comprehensive simple type annotation coverage from Step 7
func TestSimpleTypeAnnotations(t *testing.T) {
	// Set up test framework
	testDataDir := filepath.Join("testdata", "typed-perl", "simple-annotations")
	framework := NewParserTestFramework(testDataDir)
	framework.Verbose = testing.Verbose()

	// Create parser
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Run all simple type annotation tests
	metrics := framework.RunTestsByCategory(t, TypedPerl)

	// Print summary
	framework.PrintMetricsSummary(t, metrics)

	// Save metrics report
	reportsDir := filepath.Join("testdata", "reports")
	metricsPath := filepath.Join(reportsDir, "simple_annotations_metrics.json")
	err = framework.SaveMetricsReport(metrics, metricsPath)
	if err != nil {
		t.Logf("Warning: Failed to save metrics report: %v", err)
	}

	// Validate we have reasonable accuracy
	expectedMinAccuracy := 80.0 // Start with 80% as we're implementing step by step
	overallAccuracy := float64(metrics.PassedTests) / float64(metrics.TotalTests) * 100

	if overallAccuracy < expectedMinAccuracy {
		t.Errorf("Simple type annotation accuracy %.1f%% is below expected minimum %.1f%%", 
			overallAccuracy, expectedMinAccuracy)
	}

	// Ensure we actually ran some tests
	if metrics.TotalTests == 0 {
		t.Error("No simple type annotation tests were discovered")
	}
}

// TestSpecificSimpleTypeAnnotations tests specific patterns individually
func TestSpecificSimpleTypeAnnotations(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name     string
		input    string
		expectTypeAnnotations int
		expectedTypes []string
		expectedVars  []string
	}{
		{
			name:     "basic_built_in_types",
			input:    "my Int $count = 42; my Str $name = \"test\";",
			expectTypeAnnotations: 2,
			expectedTypes: []string{"Int", "Str"},
			expectedVars:  []string{"$count", "$name"},
		},
		{
			name:     "parameterized_types",
			input:    "my ArrayRef[Int] @numbers = (1, 2, 3);",
			expectTypeAnnotations: 1,
			expectedTypes: []string{"ArrayRef[Int]"},
			expectedVars:  []string{"@numbers"},
		},
		{
			name:     "custom_types", 
			input:    "my MyType $custom; my Package::Type $qualified;",
			expectTypeAnnotations: 2,
			expectedTypes: []string{"MyType", "Package::Type"},
			expectedVars:  []string{"$custom", "$qualified"},
		},
		{
			name:     "scoping_keywords",
			input:    "our Int $global; state Str $persistent;",
			expectTypeAnnotations: 2,
			expectedTypes: []string{"Int", "Str"},
			expectedVars:  []string{"$global", "$persistent"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := parser.ParseString(tc.input)
			if err != nil {
				t.Fatalf("Parsing failed: %v", err)
			}

			if ast == nil {
				t.Fatal("AST is nil")
			}

			// Check type annotation count
			if len(ast.TypeAnnotations) != tc.expectTypeAnnotations {
				t.Errorf("Expected %d type annotations, got %d", 
					tc.expectTypeAnnotations, len(ast.TypeAnnotations))
			}

			// Check that we found the expected types and variables
			foundTypes := make(map[string]bool)
			foundVars := make(map[string]bool)

			for _, annotation := range ast.TypeAnnotations {
				if annotation.TypeExpression != nil {
					foundTypes[annotation.TypeExpression.String()] = true
				}
				foundVars[annotation.AnnotatedItem] = true
			}

			// Validate expected types are found
			for _, expectedType := range tc.expectedTypes {
				if !foundTypes[expectedType] {
					t.Errorf("Expected type '%s' not found in annotations", expectedType)
				}
			}

			// Validate expected variables are found
			for _, expectedVar := range tc.expectedVars {
				if !foundVars[expectedVar] {
					t.Errorf("Expected variable '%s' not found in annotations", expectedVar)
				}
			}

			// Ensure AST has proper structure
			if ast.TypeAnnotations != nil && len(ast.TypeAnnotations) > 0 {
				for i, annotation := range ast.TypeAnnotations {
					if annotation.Kind != VarAnnotation {
						t.Errorf("Annotation %d: Expected VarAnnotation, got %v", i, annotation.Kind)
					}
					if annotation.TypeExpression == nil {
						t.Errorf("Annotation %d: TypeExpression is nil", i)
					}
					if annotation.AnnotatedItem == "" {
						t.Errorf("Annotation %d: AnnotatedItem is empty", i)
					}
				}
			}
		})
	}
}

// TestBackwardCompatibilityWithUntypedCode ensures that adding type annotations doesn't break untyped code
func TestBackwardCompatibilityWithUntypedCode(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test that pure untyped code still works
	untypedCode := `
		my $var = "value";
		my @array = (1, 2, 3);
		my %hash = (key => 'value');
		sub function { return 42; }
	`

	ast, err := parser.ParseString(untypedCode)
	if err != nil {
		t.Fatalf("Untyped code failed to parse: %v", err)
	}

	if ast == nil {
		t.Fatal("AST is nil for untyped code")
	}

	// Should have no type annotations for pure untyped code
	if len(ast.TypeAnnotations) != 0 {
		t.Errorf("Untyped code produced %d type annotations, expected 0", len(ast.TypeAnnotations))
	}

	// Test mixed typed and untyped code
	mixedCode := `
		my Int $typed = 42;
		my $untyped = "hello";
		my ArrayRef[Str] @typed_array = ("a", "b");
		my @untyped_array = (1, 2, 3);
	`

	ast, err = parser.ParseString(mixedCode)
	if err != nil {
		t.Fatalf("Mixed code failed to parse: %v", err)
	}

	if ast == nil {
		t.Fatal("AST is nil for mixed code")
	}

	// Should have exactly 2 type annotations (for the typed variables)
	expectedAnnotations := 2
	if len(ast.TypeAnnotations) != expectedAnnotations {
		t.Errorf("Mixed code produced %d type annotations, expected %d", 
			len(ast.TypeAnnotations), expectedAnnotations)
	}
}