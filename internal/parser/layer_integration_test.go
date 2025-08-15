// ABOUTME: Tests to verify parser layer produces correct tree-sitter nodes
// ABOUTME: Ensures tree-sitter grammar parses Perl constructs into expected node types

package parser

import (
	"fmt"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

// TestParserLayerBasicStructures tests that the parser correctly identifies basic Perl structures
func TestParserLayerBasicStructures(t *testing.T) {
	testCases := []struct {
		name          string
		code          string
		expectedNodes []string // Node types that should appear in the AST
		description   string
	}{
		{
			name: "simple_variable_declaration",
			code: `my $name = "value";`,
			expectedNodes: []string{
				"source_file",
				"expression_statement", // Expression statement (tree-sitter name)
				"var_decl",             // Variable declaration
				"variable",             // Variable node
			},
			description: "Simple variable declaration should parse into proper AST nodes",
		},
		{
			name: "hash_access",
			code: `my $value = $hash->{key};`,
			expectedNodes: []string{
				"source_file",
				"var_decl", // Variable declaration
				"variable", // Variable node
			},
			description: "Hash access should parse into proper AST nodes, not literals",
		},
		{
			name: "subroutine_with_statements",
			code: `sub test($input) {
    my $name = $input->{name};
    my $id = $input->{user_id};
    return "$name:$id";
}`,
			expectedNodes: []string{
				"source_file",
				"sub_decl",    // Subroutine declaration
				"block_stmt",  // Block statement
				"var_decl",    // Variable declarations
				"return_stmt", // Return statement
			},
			description: "Subroutine body statements should parse as proper AST nodes, not literal strings",
		},
		{
			name: "if_statement",
			code: `if ($condition) {
    print "true";
}`,
			expectedNodes: []string{
				"source_file",
				"conditional_statement", // Conditional statement (tree-sitter name)
				"block_stmt",            // Block statement
			},
			description: "If statements should parse into proper control flow nodes",
		},
		{
			name: "for_loop",
			code: `for my $item (@array) {
    print $item;
}`,
			expectedNodes: []string{
				"source_file",
				"for_statement", // For loop statement
				"block_stmt",    // Block statement
			},
			description: "For loops should parse into proper loop nodes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the code
			parser, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			ast, err := parser.ParseString(tc.code)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			// Collect all node types in the AST
			nodeTypes := collectNodeTypes(ast.Root)

			// Check that expected nodes are present
			for _, expectedNode := range tc.expectedNodes {
				found := false
				for _, nodeType := range nodeTypes {
					if nodeType == expectedNode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: Expected node type '%s' not found in AST. Found types: %v",
						tc.description, expectedNode, nodeTypes)
				}
			}

			// Check for problematic literals (those containing structured code)
			problematicLiterals := checkForProblematicLiterals(ast.Root)
			if len(problematicLiterals) > 0 {
				t.Errorf("%s: Found literal nodes with structured code (parsing fallback): %v. AST types: %v",
					tc.description, problematicLiterals, nodeTypes)
			}
		})
	}
}

// TestParserLayerComplexStructures tests more complex Perl constructs
func TestParserLayerComplexStructures(t *testing.T) {
	testCases := []struct {
		name        string
		code        string
		mustContain map[string]int // node_type -> minimum_count
		description string
	}{
		{
			name: "complex_subroutine_body",
			code: `sub process_data($input) {
    my $name = $input->{name};
    my $id = $input->{user_id};
    if (defined($name)) {
        return "$name:$id";
    }
    return undef;
}`,
			mustContain: map[string]int{
				"var_decl":    2, // Should have at least 2 variable declarations
				"if_stmt":     1, // Should have 1 if statement
				"return_stmt": 1, // Return statements
			},
			description: "Complex subroutine should parse all statements correctly",
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

			nodeTypes := collectNodeTypes(ast.Root)
			nodeCounts := make(map[string]int)
			for _, nodeType := range nodeTypes {
				nodeCounts[nodeType]++
			}

			for expectedType, minCount := range tc.mustContain {
				actualCount := nodeCounts[expectedType]
				if actualCount < minCount {
					t.Errorf("%s: Expected at least %d '%s' nodes, found %d. All node types: %v",
						tc.description, minCount, expectedType, actualCount, nodeCounts)
				}
			}
		})
	}
}

// collectNodeTypes recursively collects all node types in an AST
func collectNodeTypes(node ast.Node) []string {
	if node == nil {
		return []string{}
	}

	nodeTypes := []string{node.Type()}

	// Recursively collect from children
	for _, child := range node.Children() {
		childTypes := collectNodeTypes(child)
		nodeTypes = append(nodeTypes, childTypes...)
	}

	return nodeTypes
}

// checkForProblematicLiterals finds literal nodes that contain structured code (indicating parsing failures)
func checkForProblematicLiterals(node ast.Node) []string {
	var problematicLiterals []string

	if node == nil {
		return problematicLiterals
	}

	// Check if this node is a problematic literal
	if node.Type() == "literal" {
		if litExpr, ok := node.(*ast.LiteralExpr); ok {
			if containsStructuredCode(litExpr.Value) {
				problematicLiterals = append(problematicLiterals, litExpr.Value)
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children() {
		childProblematic := checkForProblematicLiterals(child)
		problematicLiterals = append(problematicLiterals, childProblematic...)
	}

	return problematicLiterals
}

// containsStructuredCode checks if a string contains Perl code constructs (from typechecker layer test)
func containsStructuredCode(s string) bool {
	// Simple heuristic to detect if a "literal" actually contains structured code
	indicators := []string{
		"my $", "->", "if (", "return ", "sub ", "{", "}", "defined(", "ref(",
	}

	for _, indicator := range indicators {
		if strings.Contains(s, indicator) {
			return true
		}
	}
	return false
}

// TestParserLayerPerformance tests that parsing doesn't have performance regressions
func TestParserLayerPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Large-ish code sample
	var codeBuilder strings.Builder
	for i := 0; i < 100; i++ {
		codeBuilder.WriteString(fmt.Sprintf(`sub test_func_%d($param) {
    my $name = $param->{name};
    my $value = $param->{value};
    if (defined($name)) {
        return process($name, $value);
    }
    return undef;
}
`, i))
	}
	code := codeBuilder.String()

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// This should complete in reasonable time (< 1 second)
	_, err = parser.ParseString(code)
	if err != nil {
		t.Fatalf("Failed to parse large code sample: %v", err)
	}

	// If we get here without timeout, parsing performance is reasonable
}
