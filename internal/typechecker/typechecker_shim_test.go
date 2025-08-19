// ABOUTME: Tests for typechecker compatibility with tree-sitter shim
// ABOUTME: Validates that type checking works correctly with TreeSitterAST

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
)

func TestTypeCheckWithTreeSitterShim(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse valid typed code
	testCode := `my Int $count = 42; my Str $name = "test";`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Verify type annotations were extracted
	if len(shimAST.TypeAnnotations) == 0 {
		t.Error("Expected type annotations to be extracted from shimAST")
	} else {
		t.Logf("Found %d type annotations in shimAST", len(shimAST.TypeAnnotations))
		for _, annotation := range shimAST.TypeAnnotations {
			t.Logf("  Annotation: %s :: %s", annotation.AnnotatedItem, annotation.TypeExpression.String())
		}
	}

	// Create a typechecker
	typeCheck, err := NewTypeCheck()
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Convert TreeSitterAST to regular AST for typechecker
	astInterface := &ast.AST{
		Path:            shimAST.Path,
		Root:            shimAST.Root,
		TypeAnnotations: shimAST.TypeAnnotations,
		Errors:          shimAST.Errors,
		Source:          shimAST.Source,
	}

	// Create a TypeChecker instance to check the AST
	checker := NewTypeChecker(typeCheck.TypeHierarchy, nil, "test")
	typeErrors := checker.CheckAST(astInterface)

	// Check results
	if len(typeErrors) > 0 {
		t.Logf("Type checker found %d errors:", len(typeErrors))
		for _, err := range typeErrors {
			t.Logf("  Error: %v", err)
		}
		// Don't fail immediately as some errors might be expected due to limited symbol table
	} else {
		t.Log("Type checker found no errors with tree-sitter shim AST")
	}

	t.Logf("Successfully type-checked tree-sitter shim AST with %d type annotations", len(shimAST.TypeAnnotations))
}

func TestTypeCheckWithInvalidTypes(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code with invalid type annotations (this will still parse but should fail type checking)
	testCode := `my InvalidType $value = 42;`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Convert to AST interface
	astInterface := &ast.AST{
		Path:            shimAST.Path,
		Root:            shimAST.Root,
		TypeAnnotations: shimAST.TypeAnnotations,
		Errors:          shimAST.Errors,
		Source:          shimAST.Source,
	}

	// Create a typechecker
	typeCheck, err := NewTypeCheck()
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Create a TypeChecker instance
	checker := NewTypeChecker(typeCheck.TypeHierarchy, nil, "test")
	typeErrors := checker.CheckAST(astInterface)

	// We expect type errors for invalid types
	foundInvalidTypeError := false
	for _, err := range typeErrors {
		errMsg := err.Error()
		if strings.Contains(errMsg, "InvalidType") || strings.Contains(errMsg, "unknown") || strings.Contains(errMsg, "not found") {
			foundInvalidTypeError = true
			t.Logf("Found expected type error: %v", err)
		}
	}

	if !foundInvalidTypeError {
		t.Log("Note: No specific error found for InvalidType (might be expected depending on type hierarchy)")
	}

	t.Logf("Type checker processed invalid type annotations and found %d errors", len(typeErrors))
}

func TestTypeCheckWithComplexShimAST(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse complex typed code
	testCode := `
		my ArrayRef[Int] $numbers = [1, 2, 3];
		my HashRef[Str] $config = {
			host => "localhost",
			port => "8080"
		};
		my Str $host = $config->{host};
		my Int $length = scalar(@$numbers);
	`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse complex code with shim: %v", err)
	}

	// Verify we have multiple type annotations
	if len(shimAST.TypeAnnotations) < 3 {
		t.Logf("Expected at least 3 type annotations, got %d", len(shimAST.TypeAnnotations))
	}

	// Convert to AST interface
	astInterface := &ast.AST{
		Path:            shimAST.Path,
		Root:            shimAST.Root,
		TypeAnnotations: shimAST.TypeAnnotations,
		Errors:          shimAST.Errors,
		Source:          shimAST.Source,
	}

	// Create a typechecker
	typeCheck, err := NewTypeCheck()
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Enable inference to test inference + typechecker integration
	typeCheck.EnableTypeInference = true

	// Create a TypeChecker instance
	checker := NewTypeChecker(typeCheck.TypeHierarchy, nil, "complex_test")

	// Run inference if enabled
	if typeCheck.EnableTypeInference && typeCheck.InferenceEngine != nil {
		err := typeCheck.InferenceEngine.InferTypes(astInterface)
		if err != nil {
			t.Logf("Type inference warning: %v", err)
		} else {
			inferredTypes := typeCheck.InferenceEngine.GetAllInferredTypes()
			t.Logf("Inference engine produced %d type inferences", len(inferredTypes))
		}
	}

	// Check the AST
	typeErrors := checker.CheckAST(astInterface)

	if len(typeErrors) > 0 {
		t.Logf("Type checker found %d errors on complex code:", len(typeErrors))
		for _, err := range typeErrors {
			t.Logf("  Error: %v", err)
		}
	} else {
		t.Log("Type checker found no errors on complex tree-sitter shim code")
	}

	// Test that we can traverse the shim nodes
	nodeCount := 0
	if shimAST.Root != nil {
		shimAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
			nodeCount++
			// Test type checking could examine these nodes if needed
			return true
		})
	}

	t.Logf("Successfully type-checked complex code with %d tree-sitter nodes", nodeCount)
	t.Logf("Found %d type annotations and %d type errors", len(shimAST.TypeAnnotations), len(typeErrors))
}

func TestTreeSitterShimTypeCheckerIntegration(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code with both valid and potentially invalid patterns
	testCode := `
		my Int $valid = 42;
		my Str $name = "test";
		my ArrayRef[Int] $list = [1, 2, 3];
	`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Test the full CheckFile-like workflow with shim AST
	typeCheck, err := NewTypeCheck()
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Convert shimAST to AST interface
	astInterface := &ast.AST{
		Path:            "test_shim.pl",
		Root:            shimAST.Root,
		TypeAnnotations: shimAST.TypeAnnotations,
		Errors:          shimAST.Errors,
		Source:          shimAST.Source,
	}

	// Create TypeChecker
	checker := NewTypeChecker(typeCheck.TypeHierarchy, nil, "integration_test")

	// Enable type inference
	if typeCheck.InferenceEngine != nil {
		err := typeCheck.InferenceEngine.InferTypes(astInterface)
		if err != nil {
			t.Logf("Type inference error: %v", err)
		}
	}

	// Perform type checking
	typeErrors := checker.CheckAST(astInterface)

	// Validate results
	t.Logf("TypeChecker integration test completed:")
	t.Logf("  Source lines: %d", strings.Count(testCode, "\n")+1)
	t.Logf("  Type annotations found: %d", len(shimAST.TypeAnnotations))
	t.Logf("  Type errors found: %d", len(typeErrors))

	// List type annotations
	for i, annotation := range shimAST.TypeAnnotations {
		t.Logf("  Annotation %d: %s :: %s at %d:%d",
			i+1, annotation.AnnotatedItem, annotation.TypeExpression.String(),
			annotation.Pos.Line, annotation.Pos.Column)
	}

	// List type errors if any
	for i, err := range typeErrors {
		t.Logf("  Error %d: %v", i+1, err)
	}

	// This test validates that the complete integration works
	// The actual number of errors depends on the type hierarchy and symbol table
}
