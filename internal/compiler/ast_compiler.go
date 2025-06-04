// ABOUTME: Proper AST-based clean Perl compiler using source positions
// ABOUTME: Replaces regex approach with correct AST traversal and text extraction

package compiler

import (
	"fmt"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// ASTCompiler compiles AST to clean Perl using proper AST traversal
type ASTCompiler struct {
	options *CompilerOptions
	source  string // Original source text for position-based extraction
}

// NewASTCompiler creates a new AST-based compiler
func NewASTCompiler() *ASTCompiler {
	return &ASTCompiler{
		options: &CompilerOptions{
			PreserveComments:   true,
			PreserveFormatting: true,
			StrictMode:         false,
		},
	}
}

// Target returns the compilation target
func (c *ASTCompiler) Target() Target {
	return TargetCleanPerl
}

// Validate checks if the AST is suitable for compilation
func (c *ASTCompiler) Validate(ast AST) error {
	if ast == nil {
		return NewCompilerError(ErrInvalidAST, "AST cannot be nil")
	}
	return nil
}

// Compile converts an AST to clean Perl code using AST traversal
func (c *ASTCompiler) Compile(ast AST) (string, error) {
	if err := c.Validate(ast); err != nil {
		return "", err
	}

	// Get source content
	source, err := ast.GetContent()
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed,
			fmt.Sprintf("failed to get source content: %v", err)).
			WithCause(err)
	}
	c.source = source

	// Get the root node
	rootNode, err := c.getRootNode(ast)
	if err != nil {
		return "", err
	}

	// Walk the AST and generate clean Perl
	result, err := c.walkNode(rootNode)
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed,
			fmt.Sprintf("failed to walk AST: %v", err)).
			WithCause(err)
	}

	// Add feature pragma if signatures are detected and not already present
	if c.hasSignatures(result) && !strings.Contains(result, "use feature 'signatures'") {
		result = "use feature 'signatures';\n" + result
	}

	return result, nil
}

// SetOptions updates the compiler options
func (c *ASTCompiler) SetOptions(options *CompilerOptions) {
	c.options = options
}

// getRootNode extracts the root node from the AST adapter
func (c *ASTCompiler) getRootNode(ast AST) (ast.Node, error) {
	if adapter, ok := ast.(*ParserASTAdapter); ok {
		if adapter.ast == nil {
			return nil, NewCompilerError(ErrInvalidAST, "adapter contains nil AST")
		}
		if adapter.ast.Root == nil {
			return nil, NewCompilerError(ErrInvalidAST, "AST has no root node")
		}
		return adapter.ast.Root, nil
	}
	return nil, NewCompilerError(ErrInvalidAST, "unsupported AST type")
}

// walkNode recursively walks an AST node and generates clean Perl
func (c *ASTCompiler) walkNode(node ast.Node) (string, error) {
	if node == nil {
		return "", nil
	}

	nodeType := node.Type()

	// Skip type annotation nodes entirely
	if c.isTypeAnnotationNode(nodeType) {
		return "", nil
	}

	// Handle token nodes - preserve their text directly
	if tokenNode, ok := node.(*ast.TokenNode); ok {
		return c.handleTokenNode(tokenNode)
	}

	// Handle special nodes that need custom processing
	switch nodeType {
	case "subroutine_definition", "sub_decl":
		return c.handleSubroutine(node)
	case "method_declaration", "method_decl":
		return c.handleMethod(node)
	case "variable_declaration":
		return c.handleVariable(node)
	case "field_declaration":
		return c.handleField(node)
	default:
		// For other nodes, process children and combine their text
		return c.walkChildren(node)
	}
}

// isTypeAnnotationNode checks if a node represents a type annotation
func (c *ASTCompiler) isTypeAnnotationNode(nodeType string) bool {
	typeNodes := map[string]bool{
		"type_expression":        true,
		"method_return_type":     true,
		"typed_method_parameter": true,
		"type_assertion":         true,
		"type_declaration":       true,
		"union_type":             true,
		"intersection_type":      true,
		"negation_type":          true,
		"parameterized_type":     true,
	}
	return typeNodes[nodeType]
}

// walkChildren processes all children of a node
func (c *ASTCompiler) walkChildren(node ast.Node) (string, error) {
	var result strings.Builder

	children := node.Children()

	// Process children with whitespace preservation
	// First, collect all non-empty children with their output
	type childInfo struct {
		node ast.Node
		text string
	}
	var processedChildren []childInfo

	for _, child := range children {
		childText, err := c.walkNode(child)
		if err != nil {
			return "", err
		}

		// Only include children that produce output
		if childText != "" {
			processedChildren = append(processedChildren, childInfo{
				node: child,
				text: childText,
			})
		}
	}

	// Now build the result with proper whitespace between non-empty children
	for i, childInfo := range processedChildren {
		if i > 0 {
			// Extract whitespace between the previous and current non-empty children
			prevChild := processedChildren[i-1].node
			whitespace := c.extractWhitespaceBefore(prevChild, childInfo.node)
			result.WriteString(whitespace)
		}
		result.WriteString(childInfo.text)
	}

	// If no children or empty result, extract text directly
	if len(children) == 0 || result.Len() == 0 {
		text := c.extractNodeText(node)
		// Check if this text contains method/subroutine declarations that need cleaning
		if strings.Contains(text, "method ") || strings.Contains(text, "sub ") {
			return c.cleanSubroutineText(text), nil
		}
		return text, nil
	}

	return result.String(), nil
}

// extractNodeText extracts the source text for a node using positions
func (c *ASTCompiler) extractNodeText(node ast.Node) string {
	start := node.Start()
	end := node.End()

	if c.source == "" || start.Offset >= len(c.source) || end.Offset > len(c.source) {
		// Fallback to node's Text() method
		return node.Text()
	}

	// Extract text using byte offsets
	return c.source[start.Offset:end.Offset]
}

// handleSubroutine processes subroutine definitions with signature cleaning
func (c *ASTCompiler) handleSubroutine(node ast.Node) (string, error) {
	// Extract the full subroutine text
	text := c.extractNodeText(node)

	// Use regex to clean the signature (reusing the working regex logic)
	return c.cleanSubroutineText(text), nil
}

// handleVariable processes variable declarations
func (c *ASTCompiler) handleVariable(node ast.Node) (string, error) {
	text := c.extractNodeText(node)
	return c.cleanVariableText(text), nil
}

// handleMethod processes method declarations (same as subroutines for now)
func (c *ASTCompiler) handleMethod(node ast.Node) (string, error) {
	// Methods use the same transformation logic as subroutines
	text := c.extractNodeText(node)
	return c.cleanSubroutineText(text), nil
}

// handleField processes field declarations
func (c *ASTCompiler) handleField(node ast.Node) (string, error) {
	text := c.extractNodeText(node)
	return c.cleanFieldText(text), nil
}

// cleanSubroutineText cleans a subroutine definition (reusing regex logic)
func (c *ASTCompiler) cleanSubroutineText(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = c.cleanSubroutineLine(line)
	}
	return strings.Join(lines, "\n")
}

// cleanSubroutineLine cleans a single line of subroutine/method text
func (c *ASTCompiler) cleanSubroutineLine(line string) string {
	// Handle function/method parameters (handle both sub and method keywords)
	funcPattern := regexp.MustCompile(`\b(sub|method)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)`)
	if funcPattern.MatchString(line) {
		line = funcPattern.ReplaceAllStringFunc(line, func(match string) string {
			parts := funcPattern.FindStringSubmatch(match)
			if len(parts) != 4 { // Now we have 4 parts: full match, keyword, name, params
				return match
			}

			keyword := parts[1]  // "sub" or "method"
			funcName := parts[2] // function name
			params := parts[3]   // parameters

			// Extract parameter names (skip type annotations) for Perl 5.36+ signatures
			paramPattern := regexp.MustCompile(`[A-Z][^$]*\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)`)
			cleanParams := paramPattern.ReplaceAllString(params, `$1`)

			// Keep signature syntax for Perl 5.36+ - just remove type annotations
			return fmt.Sprintf("%s %s(%s)", keyword, funcName, cleanParams)
		})
	}

	// Clean up return type annotations
	returnTypePattern := regexp.MustCompile(`\s*->\s*[A-Z][a-zA-Z_:]*(?:\[[^\]]*\])*`)
	line = returnTypePattern.ReplaceAllString(line, "")

	return line
}

// cleanVariableText cleans variable declarations
func (c *ASTCompiler) cleanVariableText(text string) string {
	// Handle variable declarations: my Type $var
	varPattern := regexp.MustCompile(`\b(my|our|state)\s+[A-Z][^$]+\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)`)
	return varPattern.ReplaceAllString(text, `$1 $2`)
}

// cleanFieldText cleans field declarations
func (c *ASTCompiler) cleanFieldText(text string) string {
	// Handle field declarations: field Type $field
	fieldPattern := regexp.MustCompile(`\bfield\s+[A-Z][^$]+\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)`)
	return fieldPattern.ReplaceAllString(text, `field $1`)
}

// handleTokenNode processes token nodes for whitespace preservation
func (c *ASTCompiler) handleTokenNode(tokenNode *ast.TokenNode) (string, error) {
	switch tokenNode.TokenType {
	case ast.Whitespace, ast.Newline:
		// Preserve all whitespace as-is
		return tokenNode.Text(), nil
	case ast.LeftBrace, ast.RightBrace, ast.LeftParen, ast.RightParen,
		ast.Semicolon, ast.Equals, ast.Dollar:
		// Preserve structural punctuation as-is
		return tokenNode.Text(), nil
	case ast.Arrow:
		// Remove arrow tokens (return type annotations)
		return "", nil
	case ast.SubKeyword, ast.MethodKeyword:
		// Keep sub/method keywords as-is
		return tokenNode.Text(), nil
	case ast.MyKeyword, ast.FieldKeyword:
		// Keep declaration keywords as-is
		return tokenNode.Text(), nil
	case ast.Identifier:
		// For identifiers, check if they appear to be type names that should be removed
		text := tokenNode.Text()
		if c.looksLikeTypeName(text) {
			return "", nil
		}
		return text, nil
	case ast.Number, ast.String:
		// Preserve literals as-is
		return tokenNode.Text(), nil
	default:
		// For unknown token types, preserve the text
		return tokenNode.Text(), nil
	}
}

// looksLikeTypeName checks if an identifier looks like a type name that should be stripped
func (c *ASTCompiler) looksLikeTypeName(identifier string) bool {
	// Simple heuristic: starts with uppercase letter
	if len(identifier) == 0 {
		return false
	}
	firstChar := identifier[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}

// extractWhitespaceBefore extracts whitespace between two adjacent nodes
func (c *ASTCompiler) extractWhitespaceBefore(prevNode, currentNode ast.Node) string {
	if c.source == "" {
		return ""
	}

	prevEnd := prevNode.End()
	currentStart := currentNode.Start()

	// Extract text between the nodes
	if prevEnd.Offset < currentStart.Offset && currentStart.Offset <= len(c.source) {
		between := c.source[prevEnd.Offset:currentStart.Offset]

		// Only return actual whitespace characters, filter out any non-whitespace
		var whitespace strings.Builder
		var lastWasSpace bool

		for _, char := range between {
			if char == ' ' || char == '\t' {
				// Normalize multiple spaces/tabs to single space
				if !lastWasSpace {
					whitespace.WriteRune(' ')
					lastWasSpace = true
				}
			} else if char == '\n' || char == '\r' {
				// Preserve newlines as-is
				whitespace.WriteRune(char)
				lastWasSpace = false
			}
			// For non-whitespace characters, don't reset lastWasSpace
			// since we're filtering them out - consecutive spaces should still be normalized
		}

		return whitespace.String()
	}

	return ""
}

// hasSignatures checks if the generated code contains signatures
func (c *ASTCompiler) hasSignatures(code string) bool {
	// Check for function/method signatures
	sigPattern := regexp.MustCompile(`\b(sub|method)\s+[a-zA-Z_][a-zA-Z0-9_]*\s*\([^)]*\)`)
	return sigPattern.MatchString(code)
}
