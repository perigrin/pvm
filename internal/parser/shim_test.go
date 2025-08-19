// ABOUTME: Tests for tree-sitter shim parser functionality
// ABOUTME: Validates ShimParser interface and TreeSitterAST direct CST access

package parser

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestNewShimParser(t *testing.T) {
	shimParser, err := NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	if shimParser == nil {
		t.Fatal("NewShimParser returned nil")
	}

	// Verify it implements ShimParser interface
	_, ok := shimParser.(ShimParser)
	if !ok {
		t.Fatal("Returned parser does not implement ShimParser interface")
	}
}

func TestShimParserBasicFunctionality(t *testing.T) {
	shimParser, err := NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	testCode := `my Int $x = 42;`

	// Test ParseStringShim (direct CST access)
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("ParseStringShim failed: %v", err)
	}

	if shimAST == nil {
		t.Fatal("ParseStringShim returned nil AST")
	}

	// Verify it's a TreeSitterAST
	if shimAST.GetTree() == nil {
		t.Fatal("TreeSitterAST has no underlying tree")
	}

	// Test ParseString (legacy interface)
	legacyAST, err := shimParser.ParseString(testCode)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if legacyAST == nil {
		t.Fatal("ParseString returned nil AST")
	}

	// Both should parse the same content
	if shimAST.Source != legacyAST.Source {
		t.Error("Source content differs between shim and legacy parsing")
	}
}

func TestVarDeclNodeExtraction(t *testing.T) {
	shimParser, err := NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	testCode := `my Int $value = 123;`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("ParseStringShim failed: %v", err)
	}

	// Find the variable declaration in the tree
	found := false
	if shimAST.Root != nil {
		shimAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
			if node.Type() == "variable_declaration" {
				varDecl := node.AsVarDecl()
				if varDecl != nil {
					// Test VarDecl interface methods
					if varDecl.GetKeyword() != "my" {
						t.Errorf("Expected keyword 'my', got '%s'", varDecl.GetKeyword())
					}

					variables := varDecl.GetVariables()
					if len(variables) != 1 {
						t.Errorf("Expected 1 variable, got %d", len(variables))
					} else {
						if variables[0].Name != "value" {
							t.Errorf("Expected variable name 'value', got '%s'", variables[0].Name)
						}
						if variables[0].Sigil != "$" {
							t.Errorf("Expected sigil '$', got '%s'", variables[0].Sigil)
						}
					}

					typeExpr := varDecl.GetTypeExpr()
					if typeExpr == nil {
						t.Error("Expected type expression, got nil")
					} else if typeExpr.Name != "Int" {
						t.Errorf("Expected type 'Int', got '%s'", typeExpr.Name)
					}

					init := varDecl.GetInit()
					if init == nil {
						t.Error("Expected initialization value, got nil")
					} else if !strings.Contains(init.Text(), "123") {
						t.Errorf("Expected init to contain '123', got '%s'", init.Text())
					}

					found = true
				}
				return false // Stop walking
			}
			return true // Continue walking
		})
	}

	if !found {
		t.Error("No variable declaration found in parsed AST")
	}
}

func TestShimParserDirectCSTAccess(t *testing.T) {
	shimParser, err := NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	testCode := `my Str $name = "hello";`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("ParseStringShim failed: %v", err)
	}

	// Access the underlying tree-sitter tree directly
	tree := shimAST.GetTree()
	if tree == nil {
		t.Fatal("No underlying tree-sitter tree")
	}

	rootNode := tree.RootNode()
	if rootNode == nil {
		t.Fatal("No root node in tree")
	}

	// Verify we can access raw tree-sitter functionality
	if rootNode.Kind() == "" {
		t.Error("Root node has no kind")
	}

	if rootNode.ChildCount() == 0 {
		t.Error("Root node has no children")
	}

	// Test that our shim preserves tree-sitter node information
	if shimAST.Root != nil {
		shimNode := shimAST.Root.GetTreeSitterNode()
		if shimNode == nil {
			t.Error("TreeSitterNode wrapper doesn't provide access to underlying node")
		}

		if shimNode.Kind() != rootNode.Kind() {
			t.Error("Shim node kind doesn't match raw tree-sitter node kind")
		}
	}
}

func TestTypeAnnotationExtraction(t *testing.T) {
	shimParser, err := NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	testCode := `my ArrayRef[Int] $numbers = [1, 2, 3];`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("ParseStringShim failed: %v", err)
	}

	// Check that type annotations are extracted
	if len(shimAST.TypeAnnotations) == 0 {
		t.Error("Expected type annotations to be extracted")
	}

	// Find the parameterized type
	found := false
	if shimAST.Root != nil {
		shimAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
			if node.Type() == "variable_declaration" {
				varDecl := node.AsVarDecl()
				if varDecl != nil {
					typeExpr := varDecl.GetTypeExpr()
					if typeExpr != nil && strings.Contains(typeExpr.String(), "ArrayRef") {
						found = true
						if !strings.Contains(typeExpr.String(), "Int") {
							t.Error("Expected parameterized type to contain 'Int'")
						}
					}
				}
				return false
			}
			return true
		})
	}

	if !found {
		t.Error("Did not find expected parameterized type annotation")
	}
}
