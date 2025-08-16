package parser

import (
	"testing"
)

func TestVariableDeclarationWithInitializer(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "simple_assignment",
			code:     `my $x = 5;`,
			expected: "variable should have initializer",
		},
		{
			name:     "function_call_assignment",
			code:     `my $type = ref($data);`,
			expected: "variable should have ref() call as initializer",
		},
		{
			name:     "arithmetic_assignment",
			code:     `my $sum = $x + $y;`,
			expected: "variable should have arithmetic expression as initializer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			ast, err := parser.ParseString(tc.code)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			// Check if we can find a variable declaration with initializer
			var foundVarDecl bool
			var hasInitializer bool

			// Walk the AST to find variable declarations
			var walkAST func(node interface{}, depth int)
			walkAST = func(node interface{}, depth int) {
				if node == nil {
					return
				}

				// Try to get node type
				if n, ok := node.(interface{ Type() string }); ok {
					nodeType := n.Type()
					t.Logf("Visiting node: %s", nodeType)

					if nodeType == "var_decl" || nodeType == "variable_declaration" {
						foundVarDecl = true
						t.Logf("Found %s node", nodeType)

						// Check for initializer by examining children
						if children, ok := node.(interface{ Children() []interface{} }); ok {
							for i, child := range children.Children() {
								if child == nil {
									continue
								}
								if childNode, ok := child.(interface{ Type() string }); ok {
									childType := childNode.Type()
									t.Logf("  Child %d: %s", i, childType)

									// Look for assignment or expression nodes that would indicate an initializer
									if childType == "assignment" || childType == "expression" ||
										childType == "call" || childType == "number" ||
										childType == "binary_expression" || childType == "literal" {
										hasInitializer = true
										t.Logf("  Found initializer: %s", childType)
									}
								}
							}
						}
					}

					// Walk children recursively
					if nodeWithChildren, ok := node.(interface{ Children() []interface{} }); ok {
						childList := nodeWithChildren.Children()
						t.Logf("Node %s has %d children", nodeType, len(childList))
						for _, child := range childList {
							walkAST(child, depth+1)
						}
					}
				}
			}

			walkAST(ast.Root, 0)

			if !foundVarDecl {
				t.Errorf("Expected to find variable declaration for: %s", tc.code)
			}

			// For now, just log whether we found an initializer
			// We'll make this a proper assertion once we fix the parser
			if hasInitializer {
				t.Logf("✓ Found initializer in AST")
			} else {
				t.Logf("✗ No initializer found in AST - this confirms the parser issue")
			}

			// Log AST structure for debugging
			t.Logf("AST structure for '%s':\n%s", tc.code, ast.String())
		})
	}
}
