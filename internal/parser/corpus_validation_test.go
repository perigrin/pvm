// ABOUTME: Comprehensive corpus validation tests that verify expected compilation outputs
// ABOUTME: Tests load markdown corpus files and validate parsing + compilation against expected results

package parser

import (
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/compiler"
)

// TestCorpusValidation runs comprehensive validation of all corpus files
// This test validates that corpus files with expected outputs actually produce those outputs
func TestCorpusValidation(t *testing.T) {
	framework := NewParserTestFramework("../../test/corpus/parser")

	// Discover all test cases from corpus
	testCases, err := framework.DiscoverTestCases()
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	if len(testCases) == 0 {
		t.Fatal("No test cases discovered from corpus")
	}

	t.Logf("Found %d corpus test cases to validate", len(testCases))

	// Group test cases by category for organized testing
	categories := make(map[TestCategory][]*ParserTestCase)
	for _, tc := range testCases {
		categories[tc.Category] = append(categories[tc.Category], tc)
	}

	// Run validation for each category
	for category, cases := range categories {
		category := category // capture loop variable
		cases := cases       // capture loop variable

		t.Run(string(category), func(t *testing.T) {
			t.Parallel()

			for _, testCase := range cases {
				testCase := testCase // capture loop variable

				t.Run(testCase.Name, func(t *testing.T) {
					t.Parallel()

					validateCorpusTestCase(t, testCase)
				})
			}
		})
	}
}

// TestSpecificCorpusFiles tests specific important corpus files individually
func TestSpecificCorpusFiles(t *testing.T) {
	framework := NewParserTestFramework("../../test/corpus/parser")

	// Critical corpus files that must work correctly
	criticalFiles := []string{
		"typed-perl/simple-annotations/basic-typed-variables.md",
		"typed-perl/simple-annotations/whitespace-variations.md",
		"typed-perl/union-types/simple-union-types.md",
		"typed-perl/parameterized-types/basic-parameterized-types.md",
	}

	for _, file := range criticalFiles {
		file := file // capture loop variable

		t.Run(filepath.Base(file), func(t *testing.T) {
			t.Parallel()

			fullPath := filepath.Join("../../test/corpus/parser", file)
			testCases, err := framework.LoadMarkdownTestCases(fullPath)
			if err != nil {
				t.Fatalf("Failed to load critical corpus file %s: %v", file, err)
			}

			if len(testCases) == 0 {
				t.Fatalf("No test cases found in critical corpus file %s", file)
			}

			for _, testCase := range testCases {
				testCase := testCase // capture loop variable

				t.Run(testCase.Name+"_validation", func(t *testing.T) {
					validateCorpusTestCase(t, testCase)
				})
			}
		})
	}
}

// validateCorpusTestCase validates a single corpus test case including compilation outputs
func validateCorpusTestCase(t *testing.T, testCase *ParserTestCase) {
	t.Helper()

	// Skip if test case doesn't have expected compilation outputs
	if !hasExpectedOutputs(testCase) {
		t.Skipf("Test case %s has no expected compilation outputs to validate", testCase.Name)
		return
	}

	// Get parser from pool for thread safety
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

	// Parse the input
	ast, err := parser.ParseString(testCase.Input)
	if testCase.ShouldError {
		if err == nil {
			t.Errorf("Expected parsing error for %s, but parsing succeeded", testCase.Name)
		}
		return // If we expected an error, don't continue with compilation validation
	}

	if err != nil {
		t.Fatalf("Failed to parse input for %s: %v", testCase.Name, err)
	}

	if ast == nil {
		t.Fatalf("Parser returned nil AST for %s", testCase.Name)
	}

	// Validate expected AST structure if specified
	validateASTStructure(t, testCase, ast)

	// Validate expected compilation outputs if they exist
	validateCompilationOutputs(t, testCase, ast)
}

// validateASTStructure validates the AST structure and type annotations
func validateASTStructure(t *testing.T, testCase *ParserTestCase, actualAST *ast.AST) {
	t.Helper()

	// Validate type annotations from the corpus expectations
	validateTypeAnnotations(t, testCase, actualAST)

	// Validate AST structure if expected AST strings are provided
	if testCase.ExpectedASTBeforeInfer != "" {
		t.Run("ast_before_inference", func(t *testing.T) {
			validateASTString(t, testCase.Name, actualAST, testCase.ExpectedASTBeforeInfer, "before inference")
		})
	}

	if testCase.ExpectedASTAfterInfer != "" {
		t.Run("ast_after_inference", func(t *testing.T) {
			validateASTString(t, testCase.Name, actualAST, testCase.ExpectedASTAfterInfer, "after inference")
		})
	}

	// If ExpectedAST is provided, do detailed comparison
	if testCase.ExpectedAST != nil {
		t.Run("detailed_ast_comparison", func(t *testing.T) {
			validateDetailedAST(t, testCase, actualAST)
		})
	}
}

// validateTypeAnnotations validates that type annotations match expectations from corpus
func validateTypeAnnotations(t *testing.T, testCase *ParserTestCase, actualAST *ast.AST) {
	t.Helper()

	// For now, we can extract expected type annotations from the AST strings in the corpus
	// This is a simplified validation - we could make it more sophisticated later
	actualAnnotations := actualAST.TypeAnnotations

	// Log what we found for debugging
	t.Logf("Found %d type annotations in %s:", len(actualAnnotations), testCase.Name)
	for _, ann := range actualAnnotations {
		typeName := ""
		if ann.TypeExpression != nil {
			typeName = ann.TypeExpression.Name
		}
		t.Logf("  %s :: %s (kind: %d)", ann.AnnotatedItem, typeName, ann.Kind)
	}

	// Basic validation: ensure we have type annotations for typed Perl tests
	if testCase.Category == TypedPerl && len(actualAnnotations) == 0 {
		t.Errorf("Expected type annotations for typed Perl test %s, but found none", testCase.Name)
	}
}

// validateASTString validates AST structure against expected string representation
func validateASTString(t *testing.T, testName string, actualAST *ast.AST, expectedASTString, phase string) {
	t.Helper()

	// Parse expected annotations from the corpus AST string
	expectedAnnotations := parseExpectedAnnotations(expectedASTString)
	actualAnnotations := actualAST.TypeAnnotations

	// Compare annotation counts
	if len(actualAnnotations) != len(expectedAnnotations) {
		t.Errorf("AST %s annotation count mismatch for %s: expected %d, got %d",
			phase, testName, len(expectedAnnotations), len(actualAnnotations))
		return
	}

	// Compare each annotation
	for i, expectedAnn := range expectedAnnotations {
		if i >= len(actualAnnotations) {
			break
		}

		actualAnn := actualAnnotations[i]
		actualTypeName := ""
		if actualAnn.TypeExpression != nil {
			actualTypeName = actualAnn.TypeExpression.Name
		}

		if actualAnn.AnnotatedItem != expectedAnn.Variable || actualTypeName != expectedAnn.TypeName {
			t.Errorf("AST %s annotation mismatch for %s at index %d:", phase, testName, i)
			t.Errorf("  Expected: %s :: %s", expectedAnn.Variable, expectedAnn.TypeName)
			t.Errorf("  Actual:   %s :: %s", actualAnn.AnnotatedItem, actualTypeName)
		}
	}
}

// validateDetailedAST does detailed AST comparison if ExpectedAST is provided
func validateDetailedAST(t *testing.T, testCase *ParserTestCase, actualAST *ast.AST) {
	t.Helper()

	expectedAST := testCase.ExpectedAST

	// Compare source lengths
	if len(actualAST.Source) != len(expectedAST.Source) {
		t.Errorf("AST source length mismatch for %s: expected %d, got %d",
			testCase.Name, len(expectedAST.Source), len(actualAST.Source))
	}

	// Compare type annotation counts
	if len(actualAST.TypeAnnotations) != len(expectedAST.TypeAnnotations) {
		t.Errorf("AST type annotation count mismatch for %s: expected %d, got %d",
			testCase.Name, len(expectedAST.TypeAnnotations), len(actualAST.TypeAnnotations))
	}

	// Compare root node types
	if actualAST.Root != nil && expectedAST.Root != nil {
		if actualAST.Root.Type() != expectedAST.Root.Type() {
			t.Errorf("AST root node type mismatch for %s: expected %s, got %s",
				testCase.Name, expectedAST.Root.Type(), actualAST.Root.Type())
		}
	}
}

// ExpectedAnnotation represents an expected type annotation parsed from corpus
type ExpectedAnnotation struct {
	Variable string
	TypeName string
	Position string
}

// parseExpectedAnnotations parses expected type annotations from corpus AST string
func parseExpectedAnnotations(astString string) []ExpectedAnnotation {
	var annotations []ExpectedAnnotation

	lines := strings.Split(astString, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for lines like "VarAnnotation: $count :: Int at 1:1"
		if strings.HasPrefix(line, "VarAnnotation:") {
			// Parse: "VarAnnotation: $count :: Int at 1:1"
			parts := strings.Split(line, "::")
			if len(parts) >= 2 {
				// Extract variable name from first part
				varPart := strings.TrimSpace(parts[0])
				varPart = strings.TrimPrefix(varPart, "VarAnnotation:")
				varPart = strings.TrimSpace(varPart)

				// Extract type name from second part (before "at")
				typePart := strings.TrimSpace(parts[1])
				if atIndex := strings.Index(typePart, " at "); atIndex != -1 {
					typePart = typePart[:atIndex]
				}
				typePart = strings.TrimSpace(typePart)

				annotations = append(annotations, ExpectedAnnotation{
					Variable: varPart,
					TypeName: typePart,
				})
			}
		}
	}

	return annotations
}

// hasExpectedOutputs checks if a test case has expected compilation outputs
func hasExpectedOutputs(testCase *ParserTestCase) bool {
	if testCase.ExpectedCompilationOutcomes == nil {
		return false
	}

	return testCase.ExpectedCompilationOutcomes.ExpectedCleanPerl != "" ||
		testCase.ExpectedCompilationOutcomes.ExpectedTypedPerl != "" ||
		testCase.ExpectedCompilationOutcomes.ExpectedInferredPerl != ""
}

// validateCompilationOutputs validates all expected compilation outputs for a test case
func validateCompilationOutputs(t *testing.T, testCase *ParserTestCase, ast *ast.AST) {
	t.Helper()

	outcomes := testCase.ExpectedCompilationOutcomes
	if outcomes == nil {
		return
	}

	// Validate Clean Perl output (most important for CleanPerlCompiler)
	if outcomes.ExpectedCleanPerl != "" {
		t.Run("clean_perl_output", func(t *testing.T) {
			validateCleanPerlOutput(t, testCase, ast)
		})
	}

	// Validate Typed Perl output
	if outcomes.ExpectedTypedPerl != "" {
		t.Run("typed_perl_output", func(t *testing.T) {
			validateTypedPerlOutput(t, testCase, ast)
		})
	}

	// Validate Inferred Perl output (if implemented)
	if outcomes.ExpectedInferredPerl != "" && outcomes.ExpectedInferredPerl != "# Type inference not yet fully implemented" {
		t.Run("inferred_perl_output", func(t *testing.T) {
			validateInferredPerlOutput(t, testCase, ast)
		})
	}
}

// validateCleanPerlOutput validates the Clean Perl compilation output
func validateCleanPerlOutput(t *testing.T, testCase *ParserTestCase, ast *ast.AST) {
	t.Helper()

	// Create CleanPerlCompiler (use unified compiler for better type handling)
	cleanCompiler := compiler.NewCleanPerlCompilerUnified()

	// Compile to clean Perl
	actualOutput, err := cleanCompiler.Compile(ast)
	if err != nil {
		t.Fatalf("CleanPerlCompiler failed for %s: %v", testCase.Name, err)
	}

	// Normalize outputs for comparison (remove extra whitespace, version pragmas, etc.)
	expectedNormalized := normalizeOutput(testCase.ExpectedCompilationOutcomes.ExpectedCleanPerl)
	actualNormalized := normalizeOutput(actualOutput)

	// Compare outputs
	if actualNormalized != expectedNormalized {
		t.Errorf("Clean Perl output mismatch for %s", testCase.Name)
		t.Errorf("Input:\n%s", testCase.Input)
		t.Errorf("Expected Clean Perl:\n%s", testCase.ExpectedCompilationOutcomes.ExpectedCleanPerl)
		t.Errorf("Actual Clean Perl:\n%s", actualOutput)
		t.Errorf("Expected (normalized):\n%s", expectedNormalized)
		t.Errorf("Actual (normalized):\n%s", actualNormalized)

		// Additional debugging: show what the parser actually extracted
		debugParsingResult(t, testCase, ast)
	}
}

// validateTypedPerlOutput validates the Typed Perl compilation output
func validateTypedPerlOutput(t *testing.T, testCase *ParserTestCase, ast *ast.AST) {
	t.Helper()

	// Create TypedPerlCompiler
	typedCompiler := compiler.NewTypedPerlCompiler()

	// Compile to typed Perl
	actualOutput, err := typedCompiler.Compile(ast)
	if err != nil {
		t.Fatalf("TypedPerlCompiler failed for %s: %v", testCase.Name, err)
	}

	// Normalize outputs for comparison
	expectedNormalized := normalizeOutput(testCase.ExpectedCompilationOutcomes.ExpectedTypedPerl)
	actualNormalized := normalizeOutput(actualOutput)

	// Compare outputs
	if actualNormalized != expectedNormalized {
		t.Errorf("Typed Perl output mismatch for %s", testCase.Name)
		t.Errorf("Input:\n%s", testCase.Input)
		t.Errorf("Expected Typed Perl:\n%s", testCase.ExpectedCompilationOutcomes.ExpectedTypedPerl)
		t.Errorf("Actual Typed Perl:\n%s", actualOutput)
	}
}

// validateInferredPerlOutput validates the Inferred Perl compilation output (future)
func validateInferredPerlOutput(t *testing.T, testCase *ParserTestCase, ast *ast.AST) {
	t.Helper()

	// For now, just log that inferred output validation is not yet implemented
	t.Logf("Inferred Perl output validation not yet implemented for %s", testCase.Name)
	t.Logf("Expected inferred output: %s", testCase.ExpectedCompilationOutcomes.ExpectedInferredPerl)
}

// normalizeOutput normalizes compilation output for comparison
func normalizeOutput(output string) string {
	lines := strings.Split(output, "\n")
	var normalized []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip version pragmas for comparison (they can vary)
		if strings.HasPrefix(line, "use v") {
			continue
		}

		normalized = append(normalized, line)
	}

	return strings.Join(normalized, "\n")
}

// debugParsingResult provides detailed debugging information about parsing results
func debugParsingResult(t *testing.T, testCase *ParserTestCase, ast *ast.AST) {
	t.Helper()

	// Get source content
	source, err := ast.GetContent()
	if err != nil {
		t.Logf("Failed to get source content: %v", err)
		return
	}

	// Get root node for inspection
	rootNode, err := ast.GetRootNode()
	if err != nil {
		t.Logf("Failed to get root node: %v", err)
		return
	}

	t.Logf("=== PARSING DEBUG INFO for %s ===", testCase.Name)
	t.Logf("Source: %q", source)
	t.Logf("Root node type: %s", rootNode.Type())

	// Show AST structure (limited depth to avoid spam)
	debugNodeStructure(t, rootNode, "", 3)

	// Show type annotations if any
	annotations := ast.TypeAnnotations
	if len(annotations) > 0 {
		t.Logf("Type annotations found:")
		for _, ann := range annotations {
			typeName := ""
			if ann.TypeExpression != nil {
				typeName = ann.TypeExpression.Name
			}
			t.Logf("  %s :: %s (kind: %d)", ann.AnnotatedItem, typeName, ann.Kind)
		}
	} else {
		t.Logf("No type annotations found")
	}
}

// debugNodeStructure recursively prints AST node structure for debugging
func debugNodeStructure(t *testing.T, node ast.Node, indent string, maxDepth int) {
	if node == nil || maxDepth <= 0 {
		return
	}

	t.Logf("%s%T - Type: %s", indent, node, node.Type())

	children := node.Children()
	for i, child := range children {
		t.Logf("%s  [%d]", indent, i)
		debugNodeStructure(t, child, indent+"    ", maxDepth-1)
	}
}
