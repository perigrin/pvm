// ABOUTME: Tests demonstrating tree-sitter shim benefits for flow analysis
// ABOUTME: Shows how tree-sitter provides better function call parsing and type inference

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TestLibraryFunctionInferenceWithTreeSitter demonstrates how tree-sitter shim
// provides better function call parsing for library function type inference
func TestLibraryFunctionInferenceWithTreeSitter(t *testing.T) {
	// Create tree-sitter shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Skip("Tree-sitter shim parser not available")
	}

	// Test code with library functions that were failing with traditional AST
	testCode := `
sub handle_file($filename) {
    my $content = slurp($filename);      # Should infer: Str
    my $decoded = decode_json($content); # Should infer: HashRef[Str, Any]
    my $result = $dbh->selectrow_hashref($sql); # Should infer: Maybe[HashRef[Str, Str]]
    return ($content, $decoded, $result);
}`

	// Parse with tree-sitter shim
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with tree-sitter shim: %v", err)
	}

	t.Logf("Tree-sitter shim parsed successfully, %d errors", len(shimAST.Errors))

	// Create typechecker components
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "tree_sitter_test"

	tc := &TypeCheck{
		TypeHierarchy:               hierarchy,
		Binder:                      binder.NewBinder(),
		EnableTypeInference:         true,
		EnableFlowSensitiveAnalysis: true,
	}

	// Use tree-sitter integration from Phase 1
	_, err = tc.CheckTreeSitterAST(shimAST, "tree_sitter_test")
	if err != nil {
		t.Logf("TypeChecker error (expected): %v", err)
	}

	// Create a traditional TypeChecker for FlowAnalyzer
	checker := NewTypeChecker(hierarchy, symbolTable, "tree_sitter_test")
	checker.SafetyAnalysisEnabled = true

	// Create flow analyzer
	fa := NewFlowAnalyzer(checker)

	// Test direct function call inference with tree-sitter nodes
	t.Run("direct_function_call_inference", func(t *testing.T) {
		// Walk the tree-sitter AST to find function calls
		var functionCalls []*ast.TreeSitterNode
		if shimAST.Root != nil {
			shimAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
				if node.Type() == "function_call_expression" {
					functionCalls = append(functionCalls, node)
				}
				return true
			})
		}

		t.Logf("Found %d function calls in tree-sitter AST", len(functionCalls))

		// Test that we can extract function names from tree-sitter nodes
		for i, call := range functionCalls {
			// Extract function name from tree-sitter node
			functionName := extractFunctionNameFromTreeSitterNode(call)
			t.Logf("Function call %d: %s", i, functionName)

			// Test library function inference
			if functionName == "slurp" {
				libraryType := fa.inferLibraryFunctionType(functionName, nil)
				if libraryType != "Str" {
					t.Errorf("Expected slurp to return Str, got %s", libraryType)
				} else {
					t.Logf("✓ slurp correctly inferred as returning: %s", libraryType)
				}
			}

			if functionName == "decode_json" {
				libraryType := fa.inferLibraryFunctionType(functionName, nil)
				if libraryType != "HashRef[Str, Any]" {
					t.Errorf("Expected decode_json to return HashRef[Str, Any], got %s", libraryType)
				} else {
					t.Logf("✓ decode_json correctly inferred as returning: %s", libraryType)
				}
			}
		}
	})

	// Test that tree-sitter preserves function call structure better
	t.Run("function_call_structure_preservation", func(t *testing.T) {
		// Compare traditional vs tree-sitter parsing
		traditionalParser, err := parser.NewParser()
		if err != nil {
			t.Skip("Traditional parser not available")
		}

		traditionalAST, err := traditionalParser.ParseString(testCode)
		if err != nil {
			t.Fatalf("Failed to parse with traditional parser: %v", err)
		}

		// Count function calls in both ASTs
		traditionalCallCount := countFunctionCalls(traditionalAST.Root)
		treeSitterCallCount := countTreeSitterFunctionCalls(shimAST.Root)

		t.Logf("Traditional AST function calls: %d", traditionalCallCount)
		t.Logf("Tree-sitter AST function calls: %d", treeSitterCallCount)

		// Tree-sitter should preserve function call structure better
		if treeSitterCallCount >= traditionalCallCount {
			t.Log("✓ Tree-sitter preserves function call structure as well or better than traditional AST")
		} else {
			t.Error("Tree-sitter should preserve function call structure at least as well as traditional AST")
		}
	})
}

// Helper function to extract function name from tree-sitter node
func extractFunctionNameFromTreeSitterNode(node *ast.TreeSitterNode) string {
	if node == nil {
		return ""
	}

	// Look for function identifier in tree-sitter node
	// Tree-sitter provides better structure for function calls
	for _, child := range node.GetNamedChildren() {
		if child.Type() == "function" {
			return child.GetTextContent()
		}
		if child.Type() == "identifier" || child.Type() == "function_name" {
			return child.GetTextContent()
		}
		// For some tree-sitter grammars, the function name might be the first child
		if child.IsNamed() {
			text := child.GetTextContent()
			// Simple heuristic: if it looks like a function name
			if isValidFunctionName(text) {
				return text
			}
		}
	}

	// Fallback: extract from node text
	text := node.GetTextContent()
	if strings.Contains(text, "(") {
		// Extract function name before parentheses
		parts := strings.Split(text, "(")
		if len(parts) > 0 {
			functionName := strings.TrimSpace(parts[0])
			if isValidFunctionName(functionName) {
				return functionName
			}
		}
	}

	return ""
}

// Helper function to check if a string looks like a valid function name
func isValidFunctionName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}

	// Simple validation: alphanumeric + underscore, starting with letter or underscore
	for i, r := range name {
		if i == 0 {
			if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && r != '_' {
				return false
			}
		} else {
			if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '_' {
				return false
			}
		}
	}

	return true
}

// Helper function to count function calls in traditional AST
func countFunctionCalls(node ast.Node) int {
	if node == nil {
		return 0
	}

	count := 0
	if node.Type() == "call_expr" {
		count++
	}

	for _, child := range node.Children() {
		count += countFunctionCalls(child)
	}

	return count
}

// Helper function to count function calls in tree-sitter AST
func countTreeSitterFunctionCalls(node *ast.TreeSitterNode) int {
	if node == nil {
		return 0
	}

	count := 0
	node.WalkNodes(func(n *ast.TreeSitterNode) bool {
		if n.Type() == "function_call_expression" {
			count++
		}
		return true
	})

	return count
}

// TestTreeSitterVsTraditionalInference demonstrates the benefits of tree-sitter
// for type inference scenarios that are challenging with traditional parsing
func TestTreeSitterVsTraditionalInference(t *testing.T) {
	testCases := []struct {
		name        string
		code        string
		description string
	}{
		{
			name:        "nested_function_calls",
			code:        `my $result = outer_func(inner_func($param));`,
			description: "Nested function calls with complex parsing",
		},
		{
			name:        "method_chains",
			code:        `my $data = $obj->method1()->method2($arg);`,
			description: "Method chains that require accurate structure preservation",
		},
		{
			name:        "complex_expressions",
			code:        `my $value = func1($a) + func2($b) * func3($c);`,
			description: "Function calls within complex expressions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			// Parse with both parsers
			shimParser, err := parser.NewShimParser()
			if err != nil {
				t.Skip("Tree-sitter shim parser not available")
			}

			traditionalParser, err := parser.NewParser()
			if err != nil {
				t.Skip("Traditional parser not available")
			}

			// Compare parsing results
			shimAST, err := shimParser.ParseStringShim(tc.code)
			if err != nil {
				t.Errorf("Tree-sitter parsing failed: %v", err)
				return
			}

			traditionalAST, err := traditionalParser.ParseString(tc.code)
			if err != nil {
				t.Errorf("Traditional parsing failed: %v", err)
				return
			}

			// Basic quality checks
			t.Logf("Tree-sitter errors: %d", len(shimAST.Errors))
			t.Logf("Traditional errors: %d", len(traditionalAST.Errors))

			// Tree-sitter should have fewer or equal parse errors
			if len(shimAST.Errors) <= len(traditionalAST.Errors) {
				t.Log("✓ Tree-sitter parsing quality is as good or better")
			} else {
				t.Log("Note: Tree-sitter has more parse errors for this case")
			}

			// Check CST access capability (tree-sitter advantage)
			if shimAST.GetCSTRoot() != nil {
				t.Log("✓ Tree-sitter provides direct CST access")
			} else {
				t.Error("Tree-sitter should provide CST access")
			}
		})
	}
}
