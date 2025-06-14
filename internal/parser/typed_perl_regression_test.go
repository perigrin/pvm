// ABOUTME: Regression test for typed Perl subroutine parsing
// ABOUTME: Tests that tree-sitter correctly parses typed function signatures as sub_decl nodes

package parser

import (
	"strings"
	"testing"
)

func TestTypedPerlSubroutineParsing(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string // Expected subroutine names
	}{
		{
			name: "simple_typed_subroutines",
			code: `#!/usr/bin/perl
use strict;
use warnings;

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

sub concat(Str $a, Str $b) -> Str {
    return $a . $b;
}`,
			expected: []string{"add", "concat"},
		},
		{
			name: "method_with_return_type",
			code: `#!/usr/bin/perl
use strict;
use warnings;

method add (Int $a, Int $b) returns Int {
    return $a + $b;
}`,
			expected: []string{"add"},
		},
		{
			name: "mixed_typed_and_untyped",
			code: `#!/usr/bin/perl
use strict;
use warnings;

sub hello {
    print "Hello\n";
}

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

sub goodbye {
    print "Goodbye\n";
}`,
			expected: []string{"hello", "add", "goodbye"},
		},
		{
			name: "complex_typed_signatures",
			code: `#!/usr/bin/perl
use strict;
use warnings;

sub process(ArrayRef[Str] $items, HashRef[Int] $counts) -> Bool {
    return 1;
}

sub transform(Maybe[Str] $input) -> Union[Str, Undef] {
    return $input // undef;
}`,
			expected: []string{"process", "transform"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the code
			result, err := PooledParserFunc(func(p Parser) ([]string, error) {
				ast, err := p.ParseString(tt.code)
				if err != nil {
					return nil, err
				}

				// Find all subroutine and method nodes
				var subDeclNodes []Node
				walkASTForSubDecl(ast.Root, func(node Node) bool {
					if node.Type() == "sub_decl" ||
						node.Type() == "subroutine_declaration_statement" ||
						node.Type() == "method_declaration_statement" ||
						node.Type() == "method_decl" {
						subDeclNodes = append(subDeclNodes, node)
					}
					return true
				})

				// Extract subroutine names
				var names []string
				for _, node := range subDeclNodes {
					name := extractSubroutineName(node, tt.code)
					if name != "" {
						names = append(names, name)
					}
				}

				return names, nil
			})

			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			// Check that we found the expected number of subroutines
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d subroutines, found %d", len(tt.expected), len(result))
				t.Logf("Expected: %v", tt.expected)
				t.Logf("Found: %v", result)

				// Debug: Show what node types we actually found
				debugResult, _ := PooledParserFunc(func(p Parser) (string, error) {
					ast, err := p.ParseString(tt.code)
					if err != nil {
						return "", err
					}

					var debug strings.Builder
					debug.WriteString("Node types found:\n")
					nodeTypes := make(map[string]int)
					walkASTForSubDecl(ast.Root, func(node Node) bool {
						nodeTypes[node.Type()]++
						return true
					})

					for nodeType, count := range nodeTypes {
						debug.WriteString("  " + nodeType + ": " + string(rune(count+'0')) + "\n")
					}

					return debug.String(), nil
				})
				t.Logf("Debug info:\n%s", debugResult)
			}

			// Check that we found all expected subroutines
			for _, expectedName := range tt.expected {
				found := false
				for _, actualName := range result {
					if actualName == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected subroutine '%s' not found", expectedName)
				}
			}
		})
	}
}

// walkASTForSubDecl walks the AST calling the visitor function for each node
func walkASTForSubDecl(node Node, visitor func(Node) bool) {
	if node == nil {
		return
	}

	if !visitor(node) {
		return
	}

	for _, child := range node.Children() {
		walkASTForSubDecl(child, visitor)
	}
}

// extractSubroutineName extracts the subroutine name from a sub_decl node
func extractSubroutineName(node Node, source string) string {
	// Try to extract from children first
	for _, child := range node.Children() {
		if child.Type() == "name" || child.Type() == "identifier" {
			return child.Text()
		}
	}

	// Fall back to parsing the source text
	lines := strings.Split(source, "\n")
	start := node.Start()
	end := node.End()

	if start.Line <= 0 || start.Line > len(lines) {
		return ""
	}

	var text string
	if start.Line == end.Line {
		// Single line
		line := lines[start.Line-1]
		startCol := start.Column - 1 // Adjust for 1-indexed columns
		endCol := end.Column - 1
		if startCol >= 0 && startCol < len(line) && endCol >= 0 && endCol <= len(line) {
			text = line[startCol:endCol]
		}
	} else {
		// Multi-line - extract full text range
		var parts []string
		for i := start.Line - 1; i < end.Line && i < len(lines); i++ {
			line := lines[i]
			if i == start.Line-1 {
				// First line
				startCol := start.Column - 1
				if startCol >= 0 && startCol < len(line) {
					parts = append(parts, line[startCol:])
				}
			} else if i == end.Line-1 {
				// Last line
				endCol := end.Column - 1
				if endCol >= 0 && endCol <= len(line) {
					parts = append(parts, line[:endCol])
				}
			} else {
				// Middle lines
				parts = append(parts, line)
			}
		}
		text = strings.Join(parts, "\n")
	}

	// Extract subroutine name from text like "sub add(Int $a, Int $b) -> Int {" or "method add(Int $a, Int $b) returns Int {"
	for _, keyword := range []string{"sub ", "method "} {
		if strings.Contains(text, keyword) {
			keywordIndex := strings.Index(text, keyword)
			if keywordIndex >= 0 {
				remaining := text[keywordIndex+len(keyword):] // Skip keyword
				parts := strings.Fields(remaining)
				if len(parts) >= 1 {
					name := parts[0]
					// Remove signature or body
					if idx := strings.IndexAny(name, "({"); idx > 0 {
						name = name[:idx]
					}
					return name
				}
			}
		}
	}

	return ""
}
