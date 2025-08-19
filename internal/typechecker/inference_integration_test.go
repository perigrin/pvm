// ABOUTME: Tests for inference engine integration with tree-sitter shim
// ABOUTME: Validates the convenience methods for TreeSitterAST integration

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

func TestInferTypesFromTreeSitterAST(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse some typed code
	testCode := `my Str $message = "hello"; my Int $length = length($message);`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Create inference engine
	typeHierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	engine := NewInferenceEngine(typeHierarchy, symbolTable)

	// Test the convenience method
	err = engine.InferTypesFromTreeSitterAST(shimAST)
	if err != nil {
		t.Fatalf("InferTypesFromTreeSitterAST failed: %v", err)
	}

	// Verify inferences were made
	if len(engine.InferredTypes) == 0 {
		t.Error("Expected some type inferences, but got none")
	}

	t.Logf("Successfully inferred %d types using TreeSitterAST integration", len(engine.InferredTypes))
}

func TestInferTypesFromShimParser(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse some code
	testCode := `my ArrayRef[Str] $names = ["Alice", "Bob"]; my Int $count = scalar(@$names);`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Use the one-shot convenience function
	typeHierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	engine, err := InferTypesFromShimParser(shimAST, typeHierarchy, symbolTable)
	if err != nil {
		t.Fatalf("InferTypesFromShimParser failed: %v", err)
	}

	if engine == nil {
		t.Fatal("InferTypesFromShimParser returned nil engine")
	}

	// Verify the engine has inferences
	if len(engine.InferredTypes) == 0 {
		t.Error("Expected some type inferences from one-shot function")
	}

	t.Logf("One-shot inference produced %d type inferences", len(engine.InferredTypes))

	// Check for specific inferences
	foundNames := false
	foundCount := false
	for name, info := range engine.InferredTypes {
		t.Logf("Inference: %s -> %s (confidence: %.2f)", name, info.Type, info.Confidence)
		if name == "$names" || name == "names" {
			foundNames = true
		}
		if name == "$count" || name == "count" {
			foundCount = true
		}
	}

	if !foundNames {
		t.Log("Note: No inference found for $names variable")
	}
	if !foundCount {
		t.Log("Note: No inference found for $count variable")
	}
}

func TestTreeSitterShimInferenceCompatibility(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse complex typed code
	testCode := `
		my HashRef[Str] $config = {
			host => "localhost",
			port => "8080"
		};
		my Str $host = $config->{host};
		my Int $port = int($config->{port});
	`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse complex code: %v", err)
	}

	// Verify the AST has type annotations
	if len(shimAST.TypeAnnotations) == 0 {
		t.Log("Note: No type annotations found in shimAST")
	} else {
		t.Logf("Found %d type annotations in shimAST", len(shimAST.TypeAnnotations))
		for _, annotation := range shimAST.TypeAnnotations {
			t.Logf("  Annotation: %s :: %s", annotation.AnnotatedItem, annotation.TypeExpression.String())
		}
	}

	// Test inference on this AST
	typeHierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	engine := NewInferenceEngine(typeHierarchy, symbolTable)

	err = engine.InferTypesFromTreeSitterAST(shimAST)
	if err != nil {
		t.Fatalf("Inference failed on complex code: %v", err)
	}

	// Verify we can traverse the shim nodes
	nodeCount := 0
	if shimAST.Root != nil {
		shimAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
			nodeCount++

			// Test that inference engine can work with these nodes
			if node.Type() == "variable_declaration" {
				if varDecl := node.AsVarDecl(); varDecl != nil {
					t.Logf("Found variable declaration: %s", varDecl.String())
				}
			}

			return true
		})
	}

	t.Logf("Successfully processed %d tree-sitter shim nodes", nodeCount)
	t.Logf("Inference engine produced %d inferences on complex code", len(engine.InferredTypes))
}
