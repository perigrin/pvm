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

	// Note: ERROR node checking should be handled by the parser, not the compiler.
	// The compiler should only receive valid, successfully parsed ASTs.

	// Walk the AST and generate clean Perl
	result, err := c.walkNode(rootNode)
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed,
			fmt.Sprintf("failed to walk AST: %v", err)).
			WithCause(err)
	}

	// Post-process the result to clean up any remaining type annotations
	result = c.postProcessTypeAnnotations(result)

	// Add v5.36 pragma if signatures are detected and not already present
	if c.hasSignatures(result) && !strings.Contains(result, "use v5.36") && !strings.Contains(result, "use feature 'signatures'") {
		result = "use v5.36;\n" + result
	}

	return result, nil
}

// SetOptions updates the compiler options
func (c *ASTCompiler) SetOptions(options *CompilerOptions) {
	c.options = options
}

// getRootNode extracts the root node from the AST
func (c *ASTCompiler) getRootNode(inputAST AST) (ast.Node, error) {
	// Use the AST method to get the root node directly
	return inputAST.GetRootNode()
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

	// ERROR nodes should have been caught at the top level and failed gracefully
	// If we reach here with an ERROR node, something is wrong with our detection
	if nodeType == "ERROR" {
		text := c.extractNodeText(node)
		return "", NewCompilerError(ErrCompilationFailed,
			fmt.Sprintf("Unexpected ERROR node encountered: %q. "+
				"This should have been caught during initial AST validation.", text))
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
	case "variable_declaration", "var_decl":
		return c.handleVariable(node)
	case "variable":
		return c.handleSimpleVariable(node)
	case "typed_variable_declaration":
		return c.handleTypedVariable(node)
	case "field_declaration":
		return c.handleField(node)
	case "for_statement", "for_loop", "for":
		return c.handleForLoop(node)
	case "block_stmt", "block", "BlockStmt":
		return c.handleBlock(node)
	case "interpolated_string_literal", "string_literal":
		return c.handleStringLiteral(node)
	case "identifier":
		return c.handleIdentifier(node)
	default:
		// For other nodes, process children and combine their text
		return c.walkChildren(node)
	}
}

// isTypeAnnotationNode checks if a node represents a type annotation
func (c *ASTCompiler) isTypeAnnotationNode(nodeType string) bool {
	typeNodes := map[string]bool{
		// Basic type annotations
		"type_expression": true,
		"type_annotation": true,
		"scalar_type":     true,
		"array_type":      true,
		"hash_type":       true,

		// Method/function types
		"method_return_type":     true,
		"typed_method_parameter": true,
		"return_type":            true,
		"parameter_type":         true,

		// Complex type constructs
		"type_assertion":     true,
		"type_declaration":   true,
		"union_type":         true,
		"intersection_type":  true,
		"negation_type":      true,
		"parameterized_type": true,

		// Named types that start with capital letters (heuristic)
		// These might be parsed as identifiers but represent types
		"Int":      true,
		"Str":      true,
		"Bool":     true,
		"Num":      true,
		"ArrayRef": true,
		"HashRef":  true,
		"CodeRef":  true,
		"Any":      true,
		"Undef":    true,
		"Maybe":    true,
		"Union":    true,
	}
	return typeNodes[nodeType]
}

// walkChildren processes all children of a node
func (c *ASTCompiler) walkChildren(node ast.Node) (string, error) {
	var result strings.Builder

	children := node.Children()

	// Check if this is a block node
	nodeType := node.Type()
	isBlock := nodeType == "block" || nodeType == "block_stmt"

	// Process children with whitespace preservation
	// First, collect all non-empty children with their output
	type childInfo struct {
		node ast.Node
		text string
	}
	var processedChildren []childInfo

	for _, child := range children {
		childType := child.Type()

		// Skip structural tokens in blocks to prevent double formatting
		if isBlock && (childType == "{" || childType == "}" || childType == ";") {
			continue
		}

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

	// For block nodes, they should be handled by handleBlock, not here
	if isBlock {
		// This should not normally be reached - blocks should go through handleBlock
		// Just return concatenated children without formatting
		for i, childInfo := range processedChildren {
			if i > 0 {
				result.WriteString(" ")
			}
			result.WriteString(strings.TrimSpace(childInfo.text))
		}
		return result.String(), nil
	}

	// For non-block nodes, build the result with proper whitespace
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
		return text, nil
	}

	return result.String(), nil
}

// extractNodeText extracts the source text for a node using positions
func (c *ASTCompiler) extractNodeText(node ast.Node) string {
	// First try to use the node's Text() method directly
	nodeText := node.Text()
	if nodeText != "" {
		return nodeText
	}

	// Fallback to position-based extraction if we have valid offsets
	start := node.Start()
	end := node.End()

	if c.source != "" && start.Offset >= 0 && end.Offset > start.Offset && end.Offset <= len(c.source) {
		// Extract text using byte offsets
		return c.source[start.Offset:end.Offset]
	}

	// If all else fails, return empty string
	return ""
}

// handleSubroutine processes subroutine definitions by reconstructing from children
func (c *ASTCompiler) handleSubroutine(node ast.Node) (string, error) {
	// Check if this is a SubDecl node with proper structure
	if subDecl, ok := node.(*ast.SubDecl); ok {
		return c.reconstructSubroutine(subDecl)
	}

	// Fallback: Use walkChildren which should preserve structure better
	return c.walkChildren(node)
}

// reconstructSubroutine reconstructs a subroutine from its AST representation
func (c *ASTCompiler) reconstructSubroutine(subDecl *ast.SubDecl) (string, error) {
	// Try to extract the full subroutine text from source and remove only type annotations
	if c.source != "" {
		// Get the subroutine's complete text
		fullText := c.extractNodeText(subDecl)
		if fullText != "" && false { // Disable regex approach
			// Remove type annotations from the signature and return type
			// This is a simplified approach - just remove type names before variables
			result := fullText

			// Remove parameter type annotations (e.g., "Int $a" -> "$a" or "Int $a = 5" -> "$a = 5")
			// Handle any type name (starting with capital letter) followed by variable and optional default
			result = regexp.MustCompile(`\b[A-Z][a-zA-Z0-9_]*(?:\[[^\]]*\])?\s+(\$\w+(?:\s*=\s*[^,)]+)?)`).
				ReplaceAllString(result, "$1")

			// Remove return type annotations (e.g., "-> Int" -> "" or "-> ArrayRef[Int]" -> "")
			result = regexp.MustCompile(`\s*->\s*[A-Z][a-zA-Z0-9_]*(?:\[[^\]]*\])?\s*`).
				ReplaceAllString(result, " ")

			return result, nil
		}
	}

	// Fallback: reconstruct from parts

	var result strings.Builder

	// Write 'sub' keyword
	result.WriteString("sub")

	// Write subroutine name
	if subDecl.Name != "" {
		result.WriteString(" ")
		result.WriteString(subDecl.Name)
	}

	// Write parameters (without type annotations)
	logicalParams := subDecl.LogicalParameters()
	if len(logicalParams) > 0 {
		result.WriteString("(")
		for i, param := range logicalParams {
			if i > 0 {
				result.WriteString(", ")
			}
			// Write parameter without type annotation
			result.WriteString("$")
			result.WriteString(param.Name)

			// Include default value if present
			if param.Default != nil {
				result.WriteString(" = ")
				defaultText, err := c.walkNode(param.Default)
				if err != nil {
					return "", err
				}
				result.WriteString(defaultText)
			}
		}
		result.WriteString(")")
	}

	// Skip return type annotation (-> Type)

	// Write body
	if subDecl.Body != nil {
		result.WriteString(" ")
		bodyText, err := c.walkNode(subDecl.Body)
		if err != nil {
			return "", err
		}
		result.WriteString(bodyText)
	}

	return result.String(), nil
}

// handleVariable processes variable declarations by reconstructing from children
func (c *ASTCompiler) handleVariable(node ast.Node) (string, error) {
	// For most variable declarations, just use walkChildren which will
	// delegate to handleSimpleVariable for the actual variable nodes
	return c.walkChildren(node)
}

// handleSimpleVariable processes simple variable nodes with potential type annotations
func (c *ASTCompiler) handleSimpleVariable(node ast.Node) (string, error) {
	// Get the full text of the variable node
	fullText := c.extractNodeText(node)

	// Parse the variable declaration text to remove type annotations
	// This handles patterns like "my Int $x = 42" -> "my $x = 42"
	cleaned := c.cleanVariableDeclarationText(fullText)

	return cleaned, nil
}

// handleIdentifier processes identifier nodes, filtering out type names
func (c *ASTCompiler) handleIdentifier(node ast.Node) (string, error) {
	text := c.extractNodeText(node)

	// If this identifier looks like a type name, skip it
	if c.looksLikeTypeName(text) {
		return "", nil
	}

	// Otherwise, return the identifier text
	return text, nil
}

// handleTypedVariable processes typed variable declarations by reconstructing from children
func (c *ASTCompiler) handleTypedVariable(node ast.Node) (string, error) {
	// Find and remove the type expression by reconstructing without it
	var result strings.Builder

	children := node.Children()
	lastEnd := node.Start().Offset

	for _, child := range children {
		childType := child.Type()

		// Skip type expression nodes
		if childType == "type_expression" {
			// Add everything before the type expression (including "my ")
			if child.Start().Offset > lastEnd {
				result.WriteString(c.source[lastEnd:child.Start().Offset])
			}
			// Skip to after the type expression
			lastEnd = child.End().Offset
			// Skip one space after the type if present
			if lastEnd < len(c.source) && c.source[lastEnd] == ' ' {
				lastEnd++
			}
			continue
		}

		// For other nodes, include them with surrounding whitespace
		if child.Start().Offset > lastEnd {
			result.WriteString(c.source[lastEnd:child.Start().Offset])
		}

		// Extract the node's text
		childText := c.extractNodeText(child)
		result.WriteString(childText)
		lastEnd = child.End().Offset
	}

	// Add any remaining text after the last child
	if lastEnd < node.End().Offset {
		result.WriteString(c.source[lastEnd:node.End().Offset])
	}

	// If we couldn't extract properly, fall back to a simpler approach
	if result.Len() == 0 {
		// Process children and reconstruct
		var parts []string
		for _, child := range children {
			if child.Type() == "type_expression" {
				continue
			}
			text := c.extractNodeText(child)
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " "), nil
	}

	return result.String(), nil
}

// handleMethod processes method declarations by reconstructing from children
func (c *ASTCompiler) handleMethod(node ast.Node) (string, error) {
	// Process children but skip type expression nodes
	var parts []string

	for _, child := range node.Children() {
		childType := child.Type()

		// Skip type-related nodes entirely
		if c.isTypeAnnotationNode(childType) {
			continue
		}

		// Process the remaining nodes
		childText, err := c.walkNode(child)
		if err != nil {
			return "", err
		}

		if childText != "" {
			parts = append(parts, childText)
		}
	}

	// Reconstruct with proper spacing
	return strings.Join(parts, " "), nil
}

// handleField processes field declarations by reconstructing from children
func (c *ASTCompiler) handleField(node ast.Node) (string, error) {
	// Process children but skip type expression nodes
	var parts []string

	for _, child := range node.Children() {
		childType := child.Type()

		// Skip type-related nodes entirely
		if c.isTypeAnnotationNode(childType) {
			continue
		}

		// Process the remaining nodes
		childText, err := c.walkNode(child)
		if err != nil {
			return "", err
		}

		if childText != "" {
			parts = append(parts, childText)
		}
	}

	// Reconstruct with proper spacing
	return strings.Join(parts, " "), nil
}

// handleBlock processes block statements with braces
func (c *ASTCompiler) handleBlock(node ast.Node) (string, error) {
	// Check if this is an ast.BlockStmt
	if blockStmt, ok := node.(*ast.BlockStmt); ok {
		logicalStmts := blockStmt.LogicalStatements()
		// Process BlockStmt statements
		var statements []string
		for _, stmt := range logicalStmts {
			stmtText, err := c.walkNode(stmt)
			if err != nil {
				return "", err
			}
			stmtText = strings.TrimSpace(stmtText)

			if stmtText != "" {
				// Don't add semicolons to statements that already have them
				if !strings.HasSuffix(stmtText, ";") && !strings.HasSuffix(stmtText, "}") {
					stmtText += ";"
				}
				statements = append(statements, stmtText)
			}
		}

		// Build the block with proper formatting
		if len(statements) == 0 {
			return "{ }", nil
		}

		return "{ " + strings.Join(statements, " ") + " }", nil
	}

	// For tree-sitter block nodes, try to extract source text directly
	nodeType := node.Type()
	if nodeType == "block" || nodeType == "block_stmt" {
		// Extract the source text directly
		text := c.extractNodeText(node)
		text = strings.TrimSpace(text)

		// If we got valid block text, return it directly
		if text != "" && strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
			return text, nil
		}
	}

	// For other nodes, just process children normally
	return c.walkChildren(node)
}

// handleStringLiteral processes string literals by preserving their exact content
func (c *ASTCompiler) handleStringLiteral(node ast.Node) (string, error) {
	// For string literals, we want to preserve the exact source text
	// since there shouldn't be any type annotations to strip
	nodeText := c.extractNodeText(node)
	if nodeText != "" {
		return nodeText, nil
	}

	// Fallback: process children normally
	return c.walkChildren(node)
}

// handleForLoop processes for statements by removing type annotations from the loop variable
func (c *ASTCompiler) handleForLoop(node ast.Node) (string, error) {
	// Use walkChildren which will skip type annotations but preserve structure
	return c.walkChildren(node)
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

// Note: ERROR node handling has been moved to the parser layer where it belongs.
// The compiler should only receive valid, successfully parsed ASTs.

// cleanVariableDeclarationText removes type annotations from variable declaration text
func (c *ASTCompiler) cleanVariableDeclarationText(text string) string {
	// Parse variable declaration text without regex
	// Handle patterns like:
	// - "my Int $x = 42" -> "my $x = 42"
	// - "my Int $i" -> "my $i" (in for loops)

	words := strings.Fields(text)
	if len(words) < 3 {
		// Not enough words to have a type annotation
		return text
	}

	// Look for pattern: "my" + TypeName + "$variable" + optional rest
	if words[0] == "my" && len(words) >= 3 {
		// Check if second word looks like a type name and third word is a variable
		if c.looksLikeTypeName(words[1]) && strings.HasPrefix(words[2], "$") {
			// Remove the type name (second word)
			result := words[0] // "my"
			for i := 2; i < len(words); i++ {
				result += " " + words[i]
			}
			return result
		}
	}

	// Handle cases where the entire text needs type annotation removal
	// Look for any pattern of TypeName followed by $variable
	for i := 0; i < len(words)-1; i++ {
		if c.looksLikeTypeName(words[i]) && strings.HasPrefix(words[i+1], "$") {
			// Remove the type name at position i
			var result []string
			result = append(result, words[:i]...)   // Everything before type
			result = append(result, words[i+1:]...) // Everything after type
			return strings.Join(result, " ")
		}
	}

	// If no type annotation pattern found, return original text
	return text
}

// postProcessTypeAnnotations performs final cleanup of any remaining type annotations
func (c *ASTCompiler) postProcessTypeAnnotations(code string) string {
	// Split into lines and process each line
	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		cleanedLine := c.cleanVariableDeclarationText(line)
		result = append(result, cleanedLine)
	}

	return strings.Join(result, "\n")
}

// hasSignatures checks if the generated code contains signatures
func (c *ASTCompiler) hasSignatures(code string) bool {
	// Check for function/method signatures
	sigPattern := regexp.MustCompile(`\b(sub|method)\s+[a-zA-Z_][a-zA-Z0-9_]*\s*\([^)]*\)`)
	return sigPattern.MatchString(code)
}
