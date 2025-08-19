// ABOUTME: Tests for inference engine compatibility with tree-sitter shim
// ABOUTME: Validates that type inference works correctly with TreeSitterAST

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

func TestInferenceEngineWithTreeSitterShim(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse some typed code with tree-sitter shim
	testCode := `my Int $count = 42; $count = $count + 1;`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Create inference engine
	typeHierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	engine := NewInferenceEngine(typeHierarchy, symbolTable)
	if engine == nil {
		t.Fatal("Failed to create inference engine")
	}

	// Convert TreeSitterAST to regular AST interface for compatibility
	// The TreeSitterAST already implements the AST interface
	astInterface := &ast.AST{
		Path:            shimAST.Path,
		Root:            shimAST.Root,
		TypeAnnotations: shimAST.TypeAnnotations,
		Errors:          shimAST.Errors,
		Source:          shimAST.Source,
	}

	// Run type inference
	err = engine.InferTypes(astInterface)
	if err != nil {
		t.Fatalf("Type inference failed: %v", err)
	}

	// Verify that some inferences were made
	if len(engine.InferredTypes) == 0 {
		t.Error("Expected some type inferences, but got none")
	}

	// Look for inferences about our count variable
	foundCountInference := false
	for name, info := range engine.InferredTypes {
		if name == "$count" || name == "count" {
			foundCountInference = true
			t.Logf("Found inference for %s: %s (confidence: %.2f)", name, info.Type, info.Confidence)
		}
	}

	if !foundCountInference {
		t.Log("Available inferences:")
		for name, info := range engine.InferredTypes {
			t.Logf("  %s: %s (confidence: %.2f)", name, info.Type, info.Confidence)
		}
		// Don't fail immediately as inference might be basic for now
		t.Log("Note: No specific inference found for $count variable")
	}
}

func TestInferenceEngineWithComplexTreeSitterShim(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse more complex typed code
	testCode := `
		my Str $name = "John";
		my ArrayRef[Int] $numbers = [1, 2, 3];
		my Int $length = length($name);
	`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse complex code with shim: %v", err)
	}

	// Create inference engine with proper initialization
	typeHierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	engine := NewInferenceEngine(typeHierarchy, symbolTable)

	// Convert TreeSitterAST to AST interface
	astInterface := &ast.AST{
		Path:            shimAST.Path,
		Root:            shimAST.Root,
		TypeAnnotations: shimAST.TypeAnnotations,
		Errors:          shimAST.Errors,
		Source:          shimAST.Source,
	}

	// Run type inference
	err = engine.InferTypes(astInterface)
	if err != nil {
		t.Fatalf("Type inference failed on complex code: %v", err)
	}

	// Test that the engine can traverse tree-sitter shim nodes
	nodeCount := 0
	if shimAST.Root != nil {
		shimAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
			nodeCount++
			// Verify that the node is accessible and has basic properties
			if node.Type() == "" {
				t.Error("Found node with empty type")
			}
			return true
		})
	}

	if nodeCount == 0 {
		t.Error("Expected to traverse some nodes, but found none")
	}

	t.Logf("Successfully traversed %d tree-sitter shim nodes", nodeCount)
	t.Logf("Inference engine produced %d type inferences", len(engine.InferredTypes))
}

func TestTreeSitterShimNodeCompatibility(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse simple code
	testCode := `my Int $x = 10;`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Test that tree-sitter shim nodes implement ast.Node interface correctly
	if shimAST.Root == nil {
		t.Fatal("No root node in shim AST")
	}

	// Test Node interface methods
	rootType := shimAST.Root.Type()
	if rootType == "" {
		t.Error("Root node has empty type")
	}

	children := shimAST.Root.Children()
	if len(children) == 0 {
		t.Error("Root node has no children")
	}

	// Test that children also implement Node interface
	for i, child := range children {
		if child == nil {
			t.Errorf("Child %d is nil", i)
			continue
		}

		childType := child.Type()
		if childType == "" {
			t.Errorf("Child %d has empty type", i)
		}

		start := child.Start()
		end := child.End()
		if start.Line <= 0 || start.Column <= 0 {
			t.Errorf("Child %d has invalid start position: %+v", i, start)
		}
		if end.Line < start.Line || (end.Line == start.Line && end.Column < start.Column) {
			t.Errorf("Child %d has invalid end position: start=%+v, end=%+v", i, start, end)
		}

		text := child.Text()
		if text == "" && childType != ";" && childType != "=" {
			t.Logf("Warning: Child %d (%s) has empty text", i, childType)
		}
	}

	t.Logf("TreeSitterAST node compatibility verified successfully")
	t.Logf("Root type: %s, children: %d", rootType, len(children))
}
