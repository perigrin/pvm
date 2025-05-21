// ABOUTME: Direct tree-sitter to Perl compiler
// ABOUTME: Generates clean Perl code directly from tree-sitter parse tree

package parser

import (
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// CompileTreeSitterToPerl generates clean Perl code directly from tree-sitter tree
func CompileTreeSitterToPerl(path string) (string, error) {
	// Parse directly with tree-sitter
	parser, err := treesitter.NewPerlParser(false)
	if err != nil {
		return "", err
	}
	defer parser.Close()

	tree, err := parser.ParseFile(path)
	if err != nil {
		return "", err
	}
	defer tree.Close()

	// Compile directly from tree-sitter nodes
	compiler := &TreeSitterCompiler{
		content: tree.Content,
	}

	return compiler.CompileNode(tree.Root()), nil
}

// TreeSitterCompiler generates Perl code directly from tree-sitter nodes
type TreeSitterCompiler struct {
	content []byte
}

// CompileNode recursively compiles a tree-sitter node to clean Perl code
func (c *TreeSitterCompiler) CompileNode(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind() {
	case "source_file":
		return c.compileChildren(node)
	case "variable_declaration":
		return c.compileVariableDeclaration(node)
	case "subroutine_declaration_statement":
		return c.compileSubroutineDeclaration(node)
	case "signature":
		return c.compileSignature(node)
	case "ERROR":
		// Skip ERROR nodes (these contain type annotations)
		return ""
	default:
		// For leaf nodes (like keywords, operators, punctuation), return original text
		if node.ChildCount() == 0 || c.isSimpleNode(node) {
			return c.getNodeText(node)
		}
		// For complex nodes, compile children
		return c.compileChildren(node)
	}
}

// isSimpleNode checks if a node should be treated as a simple text node
func (c *TreeSitterCompiler) isSimpleNode(node *sitter.Node) bool {
	kind := node.Kind()
	// These are simple nodes that should preserve their text
	simpleNodes := []string{
		"=", ";", "{", "}", "(", ")", ",", ".", "return", "print",
		"interpolated_string_literal", "string_content", "number",
		"comment", "function", "bareword", "$", "varname",
	}

	for _, simple := range simpleNodes {
		if kind == simple {
			return true
		}
	}
	return false
}

// compileVariableDeclaration handles "my Type $var" -> "my $var"
func (c *TreeSitterCompiler) compileVariableDeclaration(node *sitter.Node) string {
	var parts []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		switch child.Kind() {
		case "my", "our", "state":
			parts = append(parts, c.getNodeText(child))
		case "scalar":
			parts = append(parts, c.getNodeText(child))
		case "ERROR":
			// Skip type annotations (ERROR nodes)
			continue
		case ":":
			// Skip colon in type annotations
			continue
		case "attrlist":
			// Skip attribute lists (type annotations)
			continue
		default:
			compiled := c.CompileNode(child)
			if compiled != "" {
				parts = append(parts, compiled)
			}
		}
	}

	return strings.Join(parts, " ")
}

// compileSubroutineDeclaration handles "sub name(Type $param) -> RetType" -> "sub name($param)"
func (c *TreeSitterCompiler) compileSubroutineDeclaration(node *sitter.Node) string {
	var result strings.Builder

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		switch child.Kind() {
		case "sub":
			result.WriteString(c.getNodeText(child))
		case "bareword":
			result.WriteString(c.getNodeText(child))
		case "signature":
			result.WriteString(c.CompileNode(child))
		case "block":
			result.WriteString(" ")
			result.WriteString(c.CompileNode(child))
		case "ERROR":
			// Skip ERROR nodes (return type annotations like "-> Type")
			continue
		default:
			compiled := c.CompileNode(child)
			if compiled != "" {
				result.WriteString(compiled)
			}
		}
	}

	return result.String()
}

// compileSignature handles "(Type $param)" -> "($param)"
func (c *TreeSitterCompiler) compileSignature(node *sitter.Node) string {
	var result strings.Builder

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		switch child.Kind() {
		case "(", ")":
			result.WriteString(c.getNodeText(child))
		case "mandatory_parameter", "optional_parameter":
			// Extract just the parameter name (scalar variable)
			paramName := c.extractParameterName(child)
			if paramName != "" {
				result.WriteString(paramName)
			}
		case "ERROR":
			// Skip ERROR nodes (parameter type annotations)
			continue
		case ",":
			result.WriteString(c.getNodeText(child))
		default:
			compiled := c.CompileNode(child)
			if compiled != "" {
				result.WriteString(compiled)
			}
		}
	}

	return result.String()
}

// extractParameterName gets just the scalar variable from a parameter node
func (c *TreeSitterCompiler) extractParameterName(node *sitter.Node) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child != nil && child.Kind() == "scalar" {
			return c.getNodeText(child)
		}
	}
	return ""
}

// compileChildren compiles all child nodes
func (c *TreeSitterCompiler) compileChildren(node *sitter.Node) string {
	var result strings.Builder

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child != nil {
			compiled := c.CompileNode(child)
			result.WriteString(compiled)
		}
	}

	return result.String()
}

// getNodeText extracts the original text of a node
func (c *TreeSitterCompiler) getNodeText(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	start := node.StartByte()
	end := node.EndByte()
	if start >= uint(len(c.content)) || end > uint(len(c.content)) {
		return ""
	}

	return string(c.content[start:end])
}
