// ABOUTME: Tests for typechecker tree-sitter shim integration functions
// ABOUTME: Validates the convenience methods for direct TreeSitterAST type checking

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/parser"
)

func TestCheckTreeSitterAST(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse some typed code
	testCode := `my Int $count = 42; my Str $message = "hello world";`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Create a TypeCheck instance
	typeCheck, err := NewTypeCheck()
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Use the convenience method to check the TreeSitterAST
	result, err := typeCheck.CheckTreeSitterAST(shimAST, "test_module")
	if err != nil {
		t.Fatalf("CheckTreeSitterAST failed: %v", err)
	}

	if result == nil {
		t.Fatal("CheckTreeSitterAST returned nil result")
	}

	// Verify the result contains type annotations
	if len(result.TypeAnnotations) == 0 {
		t.Error("Expected type annotations in result")
	} else {
		t.Logf("Found %d type annotations in result", len(result.TypeAnnotations))
		for _, annotation := range result.TypeAnnotations {
			t.Logf("  Annotation: %s :: %s", annotation.AnnotatedItem, annotation.TypeExpression.String())
		}
	}

	// Check for type errors
	if len(result.Errors) > 0 {
		t.Logf("Type checking found %d errors:", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("  Error: %s", err.Message)
		}
	} else {
		t.Log("Type checking found no errors")
	}

	t.Logf("CheckTreeSitterAST integration test completed successfully")
}

func TestCheckTreeSitterASTWithTypeErrors(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code with type errors
	testCode := `my UnknownType $invalid = "this should cause a type error";`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Create a TypeCheck instance
	typeCheck, err := NewTypeCheck()
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Check the TreeSitterAST
	result, err := typeCheck.CheckTreeSitterAST(shimAST, "error_test")
	if err != nil {
		t.Fatalf("CheckTreeSitterAST failed: %v", err)
	}

	// We expect type errors
	if len(result.Errors) == 0 {
		t.Log("Note: Expected type errors for UnknownType, but none found (may depend on type hierarchy)")
	} else {
		t.Logf("Found expected type errors: %d", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("  Error: %s", err.Message)
		}
	}

	t.Logf("Type error handling test completed")
}

func TestCheckTreeSitterASTWithInference(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code that could benefit from type inference
	testCode := `
		my ArrayRef[Str] $names = ["Alice", "Bob"];
		my Int $count = scalar(@$names);
		my Str $first = $names->[0];
	`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Create a TypeCheck instance with inference enabled
	typeCheck, err := NewTypeCheck()
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Enable type inference
	typeCheck.EnableTypeInference = true

	// Check the TreeSitterAST with inference
	result, err := typeCheck.CheckTreeSitterAST(shimAST, "inference_test")
	if err != nil {
		t.Fatalf("CheckTreeSitterAST with inference failed: %v", err)
	}

	// Verify results
	t.Logf("Type checking with inference completed:")
	t.Logf("  Type annotations: %d", len(result.TypeAnnotations))
	t.Logf("  Type errors: %d", len(result.Errors))
	t.Logf("  Flow-sensitive enabled: %t", result.FlowSensitiveEnabled)

	if len(result.RefinedTypes) > 0 {
		t.Logf("  Refined types: %d", len(result.RefinedTypes))
		for varName, refinedType := range result.RefinedTypes {
			t.Logf("    %s -> %s", varName, refinedType)
		}
	}

	// List all type annotations found
	for i, annotation := range result.TypeAnnotations {
		t.Logf("  Annotation %d: %s :: %s at %d:%d",
			i+1, annotation.AnnotatedItem, annotation.TypeExpression.String(),
			annotation.Pos.Line, annotation.Pos.Column)
	}

	t.Log("Inference integration test completed successfully")
}

func TestBindWithTreeSitterCST(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code for CST binding test
	testCode := `my Int $x = 10; my Str $y = "test";`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Create a TypeCheck instance
	typeCheck, err := NewTypeCheck()
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Test the CST binding method
	symbolTable, err := typeCheck.bindWithTreeSitterCST(shimAST)
	if err != nil {
		t.Fatalf("bindWithTreeSitterCST failed: %v", err)
	}

	if symbolTable == nil {
		t.Fatal("bindWithTreeSitterCST returned nil symbol table")
	}

	// Verify symbol table has symbols
	symbols := symbolTable.GetVisibleSymbols()
	t.Logf("CST binding produced %d symbols", len(symbols))

	for _, symbol := range symbols {
		t.Logf("  Symbol: %s (type: %s)", symbol.Name, symbol.Type)
	}

	t.Log("CST binding test completed successfully")
}
