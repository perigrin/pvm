// ABOUTME: Go bindings for tree-sitter-perl
// ABOUTME: Provides Go interface to tree-sitter-perl parser

package treesitter

import (
	"fmt"
	"os"
	"strings"

	tree_sitter_typed_perl "github.com/perigrin/pvm/tree-sitter-typed-perl/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Language returns the tree-sitter language for Typed Perl
func Language() *sitter.Language {
	return sitter.NewLanguage(tree_sitter_typed_perl.Language())
}

// PerlParser represents a parser instance for Perl code
type PerlParser struct {
	parser       *sitter.Parser
	typeQueries  *sitter.Query
	debug        bool
	typePatterns []string
}

// NewPerlParser creates a new PerlParser instance
func NewPerlParser(debug bool) (*PerlParser, error) {
	// Create a new tree-sitter parser
	parser := sitter.NewParser()
	language := Language()

	if err := parser.SetLanguage(language); err != nil {
		return nil, fmt.Errorf("failed to set language: %v", err)
	}

	p := &PerlParser{
		parser:       parser,
		typeQueries:  nil, // Will be set up later for type annotation queries
		debug:        debug,
		typePatterns: []string{},
	}

	return p, nil
}

// ParseFile parses a Perl file using tree-sitter
func (p *PerlParser) ParseFile(path string) (*PerlTree, error) {
	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return p.ParseBytes(content)
}

// ParseString parses a string of Perl code using tree-sitter
func (p *PerlParser) ParseString(content string) (*PerlTree, error) {
	return p.ParseBytes([]byte(content))
}

// ParseBytes parses byte content as Perl code using tree-sitter
func (p *PerlParser) ParseBytes(content []byte) (*PerlTree, error) {
	// Parse the content using tree-sitter
	tree := p.parser.Parse(content, nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse content")
	}

	perlTree := &PerlTree{
		Tree:    tree,
		Content: content,
		parser:  p,
	}

	return perlTree, nil
}

// Close frees resources used by the parser
func (p *PerlParser) Close() {
	if p.parser != nil {
		p.parser.Close()
	}
}

// PerlError represents a parsing error
type PerlError struct {
	Message string
	Row     uint32
	Column  uint32
}

// PerlTree represents a parsed Perl syntax tree
type PerlTree struct {
	Tree    *sitter.Tree
	Content []byte
	parser  *PerlParser
	errors  []PerlError
}

// Root returns the root node of the tree
func (t *PerlTree) Root() *sitter.Node {
	if t.Tree == nil {
		return nil
	}
	return t.Tree.RootNode()
}

// Errors returns any parsing errors
func (t *PerlTree) Errors() []PerlError {
	return t.errors
}

// Close frees resources used by the tree
func (t *PerlTree) Close() {
	if t.Tree != nil {
		t.Tree.Close()
	}
}

// FindTypeAnnotations finds all type annotations in the tree
func (t *PerlTree) FindTypeAnnotations() ([]*PerlTypeAnnotation, error) {
	var annotations []*PerlTypeAnnotation

	if t.Tree == nil || t.Tree.RootNode() == nil {
		return annotations, nil
	}

	// Traverse the tree looking for type annotations
	root := t.Tree.RootNode()
	t.traverseForTypeAnnotations(root, &annotations)

	return annotations, nil
}

// traverseForTypeAnnotations recursively traverses the tree looking for type annotations
func (t *PerlTree) traverseForTypeAnnotations(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	if node == nil {
		return
	}

	// DEBUG: Print node kinds to understand what we're seeing (only if DEBUG_PARSER=1)
	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Traversing node: %s (text: %.50s)\n", node.Kind(), strings.ReplaceAll(t.getNodeText(node), "\n", "\\n"))
	}

	// Check if this node represents a type annotation pattern
	switch node.Kind() {
	case "variable_declaration":
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Found variable_declaration node\n")
		}
		t.processVariableDeclaration(node, annotations)
	case "typed_variable_declaration":
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Found typed_variable_declaration node\n")
		}
		t.processTypedVariableDeclaration(node, annotations)
	case "subroutine_declaration_statement":
		t.processSubroutineDeclaration(node, annotations)
	case "method_declaration_statement":
		t.processMethodDeclaration(node, annotations)
	case "type_declaration":
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Found type_declaration node\n")
		}
		t.processTypeDeclaration(node, annotations)
	case "type_assertion":
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Found type_assertion node\n")
		}
		t.processTypeAssertion(node, annotations)
	}

	// Recursively process all child nodes
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child != nil {
			t.traverseForTypeAnnotations(child, annotations)
		}
	}
}

// processTypedVariableDeclaration processes typed variable declarations like "my Type $var"
func (t *PerlTree) processTypedVariableDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	var varName, typeName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing typed variable declaration with %d children\n", node.ChildCount())
	}

	// Walk through child nodes to find the components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "scalar", "array", "hash":
			// Handle all variable types: $var, @array, %hash
			varName = t.getNodeText(child)
		case "type_expression":
			// Extract the full type expression text instead of just identifier
			typeName = t.getNodeText(child)
			
			// Check if the next child is an ERROR node containing package qualification
			if i+1 < int(node.ChildCount()) {
				nextChild := node.Child(uint(i + 1))
				if nextChild != nil && nextChild.Kind() == "ERROR" {
					errorText := t.getNodeText(nextChild)
					if strings.HasPrefix(errorText, "::") {
						// This is a package-qualified type like Package::Type
						typeName = typeName + errorText
					}
				}
			}
		}
	}

	// Create annotation if we found both variable name and type name
	if varName != "" && typeName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating typed annotation for %s: %s\n", varName, typeName)
		}
		annotation := &PerlTypeAnnotation{
			ItemName: varName,
			TypeName: typeName,
			Kind:     "variable",
			StartPos: int(node.StartByte()),
			EndPos:   int(node.EndByte()),
			Content:  t.getNodeText(node),
		}
		*annotations = append(*annotations, annotation)
	}
}

// processVariableDeclaration looks for type annotations in variable declarations
func (t *PerlTree) processVariableDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	// Only handle Typed Perl syntax: my Type $var
	// This appears as: my + ERROR + scalar (where ERROR contains the type)

	var varName, typeName string
	var typeNode *sitter.Node

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing variable declaration with %d children\n", node.ChildCount())
	}

	// Walk through child nodes to find the components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "scalar":
			varName = t.getNodeText(child)
		case "ERROR":
			// This might be a type name in "my Type $var" syntax
			errorText := t.getNodeText(child)
			// Check if it looks like a type name (starts with uppercase)
			if len(errorText) > 0 && errorText[0] >= 'A' && errorText[0] <= 'Z' {
				typeName = errorText
				typeNode = child // Store the ERROR node for position
			}
		}
	}

	// Create annotation if we found both variable name and type name
	if varName != "" && typeName != "" && typeNode != nil {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating annotation for %s: %s\n", varName, typeName)
		}
		annotation := &PerlTypeAnnotation{
			ItemName: varName,
			TypeName: typeName,
			Kind:     "variable",
			StartPos: int(typeNode.StartByte()), // Use the specific type token position
			EndPos:   int(typeNode.EndByte()),   // Use the specific type token position
			Content:  t.getNodeText(typeNode),   // Content of just the type token
		}
		*annotations = append(*annotations, annotation)
	}
}

// processSubroutineDeclaration looks for type annotations in subroutine declarations
func (t *PerlTree) processSubroutineDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	// Extract subroutine name
	var subName string

	// Walk through child nodes to find signature and return type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		switch child.Kind() {
		case "bareword":
			// This should be the subroutine name
			subName = t.getNodeText(child)
		}
	}

	// Create individual annotations for parameter types and return type
	// This allows the stripper to remove them independently
	if subName != "" {
		// Create annotations for parameter types and return type
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(uint(i))
			if child == nil {
				continue
			}

			if child.Kind() == "signature" {
				// Process each ERROR node in the signature
				for j := 0; j < int(child.ChildCount()); j++ {
					sigChild := child.Child(uint(j))
					if sigChild != nil && sigChild.Kind() == "ERROR" {
						errorText := t.getNodeText(sigChild)
						if len(errorText) > 0 && errorText[0] >= 'A' && errorText[0] <= 'Z' {
							annotation := &PerlTypeAnnotation{
								ItemName: fmt.Sprintf("%s_param", subName),
								TypeName: errorText,
								Kind:     "subroutine_param",
								StartPos: int(sigChild.StartByte()),
								EndPos:   int(sigChild.EndByte()),
								Content:  errorText,
							}
							*annotations = append(*annotations, annotation)
						}
					}
				}
			} else if child.Kind() == "ERROR" && strings.HasPrefix(t.getNodeText(child), "->") {
				// Return type annotation
				errorText := t.getNodeText(child)
				annotation := &PerlTypeAnnotation{
					ItemName: fmt.Sprintf("%s_return", subName),
					TypeName: strings.TrimSpace(strings.TrimPrefix(errorText, "->")),
					Kind:     "subroutine_return",
					StartPos: int(child.StartByte()),
					EndPos:   int(child.EndByte()),
					Content:  errorText,
				}
				*annotations = append(*annotations, annotation)
			}
		}
	}
}


// processMethodDeclaration looks for type annotations in method declarations
func (t *PerlTree) processMethodDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	var methodName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing method declaration with %d children\n", node.ChildCount())
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Method child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "bareword":
			if methodName == "" {
				methodName = t.getNodeText(child)
			}
		case "signature":
			t.processMethodSignature(child, methodName, annotations)
		case "method_return_type":
			t.processMethodReturnType(child, methodName, annotations)
		}
	}
}

// processMethodSignature handles signature nodes for typed method parameters
func (t *PerlTree) processMethodSignature(signatureNode *sitter.Node, methodName string, annotations *[]*PerlTypeAnnotation) {
	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing method signature for %s with %d children\n", methodName, signatureNode.ChildCount())
	}

	for i := 0; i < int(signatureNode.ChildCount()); i++ {
		child := signatureNode.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Signature child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		if child.Kind() == "typed_method_parameter" {
			t.processTypedMethodParameter(child, methodName, annotations)
		}
	}
}

// processTypedMethodParameter processes individual typed method parameters
func (t *PerlTree) processTypedMethodParameter(paramNode *sitter.Node, methodName string, annotations *[]*PerlTypeAnnotation) {
	var paramName, typeName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing typed method parameter for %s with %d children\n", methodName, paramNode.ChildCount())
	}

	for i := 0; i < int(paramNode.ChildCount()); i++ {
		child := paramNode.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Parameter child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "type_expression":
			typeName = t.extractTypeExpression(child)
		case "scalar", "array", "hash":
			paramName = t.getNodeText(child)
		}
	}

	if paramName != "" && typeName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating method parameter annotation: %s for %s: %s\n", paramName, methodName, typeName)
		}
		annotation := &PerlTypeAnnotation{
			ItemName: paramName,
			TypeName: typeName,
			Kind:     "method_parameter",
			StartPos: int(paramNode.StartByte()),
			EndPos:   int(paramNode.EndByte()),
			Content:  t.getNodeText(paramNode),
		}
		*annotations = append(*annotations, annotation)
	}
}

// processMethodReturnType processes method return type annotations
func (t *PerlTree) processMethodReturnType(returnTypeNode *sitter.Node, methodName string, annotations *[]*PerlTypeAnnotation) {
	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing method return type for %s with %d children\n", methodName, returnTypeNode.ChildCount())
	}

	for i := 0; i < int(returnTypeNode.ChildCount()); i++ {
		child := returnTypeNode.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Return type child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		if child.Kind() == "type_expression" {
			typeName := t.extractTypeExpression(child)

			if typeName != "" {
				if os.Getenv("DEBUG_PARSER") == "1" {
					fmt.Printf("DEBUG: Creating method return type annotation for %s: %s\n", methodName, typeName)
				}
				annotation := &PerlTypeAnnotation{
					ItemName: methodName + "_return", // Unique identifier for return type
					TypeName: typeName,
					Kind:     "method_return",
					StartPos: int(returnTypeNode.StartByte()),
					EndPos:   int(returnTypeNode.EndByte()),
					Content:  t.getNodeText(returnTypeNode),
				}
				*annotations = append(*annotations, annotation)
			}
			break
		}
	}
}

// processTypeDeclaration processes type declarations like "type MyType = OtherType"
func (t *PerlTree) processTypeDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	var typeName, typeDefinition string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing type declaration with %d children\n", node.ChildCount())
	}

	// Walk through child nodes to find the components
	// Based on grammar: type_declaration: $ => seq('type', field('name', $.identifier), '=', field('definition', $.type_expression), $._semicolon)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "identifier":
			// This should be the type name after 'type' keyword
			if typeName == "" {
				typeName = t.getNodeText(child)
			}
		case "type_expression":
			// This is the type definition after '='
			typeDefinition = t.extractTypeExpression(child)
		}
	}

	if typeName != "" && typeDefinition != "" {
		annotation := &PerlTypeAnnotation{
			ItemName: typeName,
			TypeName: typeDefinition,
			Kind:     "type_declaration",
			StartPos: int(node.StartByte()),
			EndPos:   int(node.EndByte()),
			Content:  t.getNodeText(node),
		}
		*annotations = append(*annotations, annotation)

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Created type declaration annotation: %s = %s\n", typeName, typeDefinition)
		}
	}
}

// extractTypeExpression extracts the type definition from a type_expression node
func (t *PerlTree) extractTypeExpression(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	// For simple cases, just return the text content
	// In more complex cases, we might need to parse union types, intersection types, etc.
	content := t.getNodeText(node)

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Extracting type expression: %s (kind: %s)\n", content, node.Kind())
	}

	return content
}

// PerlTypeAnnotation represents a type annotation found in Perl code
type PerlTypeAnnotation struct {
	ItemName string
	TypeName string
	Kind     string // variable, subroutine, method, etc.
	StartPos int    // byte position
	EndPos   int    // byte position
	Content  string // original source text
	Children []*PerlTypeAnnotation
}

// String returns a string representation of the annotation
func (a *PerlTypeAnnotation) String() string {
	return fmt.Sprintf("%s: %s", a.ItemName, a.TypeName)
}

// These extraction functions are placeholders

// GetNodeText extracts the text content of a tree-sitter node
func (t *PerlTree) GetNodeText(node *sitter.Node) string {
	if node == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= uint(len(t.Content)) || end > uint(len(t.Content)) {
		return ""
	}
	return string(t.Content[start:end])
}

// PerlPosition represents a position in the source code
type PerlPosition struct {
	Row    uint32
	Column uint32
}

// GetPosition returns the position of a node
func (t *PerlTree) GetPosition(node interface{}) PerlPosition {
	return PerlPosition{}
}

// Helper functions for extracting type information from text

// getNodeText is an alias for GetNodeText to match the usage in the methods above
func (t *PerlTree) getNodeText(node *sitter.Node) string {
	return t.GetNodeText(node)
}










// processTypeAssertion processes type assertion expressions like "$value as Type"
func (t *PerlTree) processTypeAssertion(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	var expressionText, typeName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing type assertion with %d children\n", node.ChildCount())
		fmt.Printf("DEBUG: Type assertion text: %s\n", t.getNodeText(node))
	}

	// Walk through child nodes to find the expression and type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "type_expression":
			// Extract the type being asserted to
			typeName = t.getNodeText(child)
		default:
			// The first non-"as" child should be the expression being asserted
			childText := t.getNodeText(child)
			if childText != "as" && expressionText == "" {
				expressionText = childText
			}
		}
	}

	// Create a type assertion annotation if we found both parts
	if expressionText != "" && typeName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating type assertion annotation: %s :: %s\n", expressionText, typeName)
		}

		annotation := &PerlTypeAnnotation{
			ItemName: expressionText,
			TypeName: typeName,
			Kind:     "type_assertion",
			StartPos: int(node.StartByte()),
			EndPos:   int(node.EndByte()),
		}

		*annotations = append(*annotations, annotation)
	}
}
