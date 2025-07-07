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
	case "subroutine_declaration_statement":
		t.processSubroutineDeclaration(node, annotations)
	case "method_declaration_statement":
		t.processMethodDeclaration(node, annotations)
	case "class_statement":
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Found class_statement node\n")
		}
		t.processClassDeclaration(node, annotations)
	case "type_declaration":
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Found type_declaration node\n")
		}
		t.processTypeDeclaration(node, annotations)
	case "type_assertion_expression":
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

// processVariableDeclaration looks for type annotations in variable declarations
func (t *PerlTree) processVariableDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	var varName, typeName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing variable declaration with %d children\n", node.ChildCount())
	}

	// Use field access to get variable component and manual iteration for type
	// Note: grammar defines field('variable', ...) but no field name for type_expression
	var typeNode *sitter.Node
	variableNode := node.ChildByFieldName("variable")

	// Find type_expression node by iterating through children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child != nil && child.Kind() == "type_expression" {
			typeNode = child
			break
		}
	}

	if os.Getenv("DEBUG_PARSER") == "1" {
		if typeNode != nil {
			fmt.Printf("DEBUG: Found type field: %s (text: %s)\n", typeNode.Kind(), t.getNodeText(typeNode))
		} else {
			fmt.Printf("DEBUG: No type field found\n")
		}
		if variableNode != nil {
			fmt.Printf("DEBUG: Found variable field: %s (text: %s)\n", variableNode.Kind(), t.getNodeText(variableNode))
		} else {
			fmt.Printf("DEBUG: No variable field found\n")
		}
	}

	// Extract type name from type field
	if typeNode != nil {
		typeName = t.getNodeText(typeNode)
	}

	// Extract variable name from variable field
	if variableNode != nil {
		varName = t.getNodeText(variableNode)
	}

	// Fall back to walking children if field access doesn't work or for legacy ERROR node handling
	if typeName == "" || varName == "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Field access failed, falling back to child walking\n")
		}

		var errorTypeNode *sitter.Node

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
				if varName == "" {
					varName = t.getNodeText(child)
				}
			case "type_expression":
				if typeName == "" {
					typeName = t.getNodeText(child)
				}
			case "ERROR":
				// Legacy handling for "my Type $var" syntax where Type appears as ERROR
				errorText := t.getNodeText(child)
				if len(errorText) > 0 && errorText[0] >= 'A' && errorText[0] <= 'Z' && typeName == "" {
					typeName = errorText
					errorTypeNode = child
				}
			}
		}

		// For ERROR node case, use its position
		if errorTypeNode != nil && typeName != "" && varName != "" {
			if os.Getenv("DEBUG_PARSER") == "1" {
				fmt.Printf("DEBUG: Creating annotation from ERROR node for %s: %s\n", varName, typeName)
			}
			annotation := &PerlTypeAnnotation{
				ItemName: varName,
				TypeName: typeName,
				Kind:     "variable",
				StartPos: int(errorTypeNode.StartByte()),
				EndPos:   int(errorTypeNode.EndByte()),
				Content:  t.getNodeText(errorTypeNode),
			}
			*annotations = append(*annotations, annotation)
			return
		}
	}

	// Create annotation if we found both variable name and type name
	if varName != "" && typeName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating annotation for %s: %s\n", varName, typeName)
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

// processSubroutineDeclaration looks for type annotations in subroutine declarations
func (t *PerlTree) processSubroutineDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	// Extract subroutine name
	var subName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing subroutine declaration with %d children\n", node.ChildCount())
	}

	// Walk through child nodes to find signature and return type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Subroutine child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "bareword":
			// This should be the subroutine name
			subName = t.getNodeText(child)
		case "signature":
			// Process typed parameters in the signature
			t.processSubroutineSignature(child, subName, annotations)
		case "method_return_type":
			// Process return type annotation
			t.processSubroutineReturnType(child, subName, annotations)
		}
	}
}

// processSubroutineSignature handles signature nodes for typed subroutine parameters
func (t *PerlTree) processSubroutineSignature(signatureNode *sitter.Node, subName string, annotations *[]*PerlTypeAnnotation) {
	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing subroutine signature for %s with %d children\n", subName, signatureNode.ChildCount())
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
			t.processTypedSubroutineParameter(child, subName, annotations)
		} else if child.Kind() == "optional_parameter" {
			t.processOptionalSubroutineParameter(child, subName, annotations)
		}
	}
}

// processTypedSubroutineParameter processes individual typed subroutine parameters
func (t *PerlTree) processTypedSubroutineParameter(paramNode *sitter.Node, subName string, annotations *[]*PerlTypeAnnotation) {
	var paramName, typeName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing typed subroutine parameter for %s with %d children\n", subName, paramNode.ChildCount())
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
			fmt.Printf("DEBUG: Creating subroutine parameter annotation: %s for %s: %s\n", paramName, subName, typeName)
		}
		annotation := &PerlTypeAnnotation{
			ItemName: paramName,
			TypeName: typeName,
			Kind:     "subroutine_param",
			StartPos: int(paramNode.StartByte()),
			EndPos:   int(paramNode.EndByte()),
			Content:  t.getNodeText(paramNode),
		}
		*annotations = append(*annotations, annotation)
	}
}

// processOptionalSubroutineParameter processes optional subroutine parameters with default values
func (t *PerlTree) processOptionalSubroutineParameter(paramNode *sitter.Node, subName string, annotations *[]*PerlTypeAnnotation) {
	var paramName, typeName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing optional subroutine parameter for %s with %d children\n", subName, paramNode.ChildCount())
	}

	for i := 0; i < int(paramNode.ChildCount()); i++ {
		child := paramNode.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Optional subroutine parameter child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "type_expression":
			typeName = t.extractTypeExpression(child)
		case "scalar", "array", "hash":
			paramName = t.getNodeText(child)
		case "=":
			// Skip the assignment operator
			continue
		case "number", "string", "bareword":
			// Skip default values - we don't need them for type annotation
			continue
		}
	}

	if paramName != "" && typeName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating optional subroutine parameter annotation: %s for %s: %s\n", paramName, subName, typeName)
		}
		annotation := &PerlTypeAnnotation{
			ItemName: paramName,
			TypeName: typeName,
			Kind:     "subroutine_param",
			StartPos: int(paramNode.StartByte()),
			EndPos:   int(paramNode.EndByte()),
			Content:  t.getNodeText(paramNode),
		}
		*annotations = append(*annotations, annotation)
	}
}

// processSubroutineReturnType processes subroutine return type annotations
func (t *PerlTree) processSubroutineReturnType(returnTypeNode *sitter.Node, subName string, annotations *[]*PerlTypeAnnotation) {
	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing subroutine return type for %s with %d children\n", subName, returnTypeNode.ChildCount())
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
					fmt.Printf("DEBUG: Creating subroutine return type annotation for %s: %s\n", subName, typeName)
				}
				annotation := &PerlTypeAnnotation{
					ItemName: subName + "_return", // Unique identifier for return type
					TypeName: typeName,
					Kind:     "subroutine_return",
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

// processMethodDeclaration looks for type annotations in method declarations
func (t *PerlTree) processMethodDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	var methodName string
	var returnTypeNode *sitter.Node

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing method declaration with %d children\n", node.ChildCount())
	}

	// First pass: identify the method name and potential return type
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
				if os.Getenv("DEBUG_PARSER") == "1" {
					fmt.Printf("DEBUG: Found method name: %s\n", methodName)
				}
			}
		case "type_expression":
			// Check if this type_expression appears before the method name
			// If so, it's the return type
			if methodName == "" && returnTypeNode == nil {
				returnTypeNode = child
				if os.Getenv("DEBUG_PARSER") == "1" {
					fmt.Printf("DEBUG: Found return type before method name: %s\n", t.getNodeText(child))
				}
			}
		}
	}

	// Process the return type if we found one before the method name
	if returnTypeNode != nil && methodName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Processing pre-name return type for method %s\n", methodName)
		}
		t.processMethodReturnType(returnTypeNode, methodName, annotations)
	}

	// Check for postfix return type using field access
	if methodName != "" {
		postfixReturnType := node.ChildByFieldName("postfix_return_type")
		if postfixReturnType != nil {
			if os.Getenv("DEBUG_PARSER") == "1" {
				fmt.Printf("DEBUG: Found postfix_return_type field for method %s: %s\n", methodName, t.getNodeText(postfixReturnType))
			}
			t.processMethodReturnType(postfixReturnType, methodName, annotations)
		}
	}

	// Second pass: process other elements
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		switch child.Kind() {
		case "signature":
			t.processMethodSignature(child, methodName, annotations)
		case "return_type":
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
		} else if child.Kind() == "optional_parameter" {
			t.processOptionalMethodParameter(child, methodName, annotations)
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

// processOptionalMethodParameter processes optional method parameters with default values
func (t *PerlTree) processOptionalMethodParameter(paramNode *sitter.Node, methodName string, annotations *[]*PerlTypeAnnotation) {
	var paramName, typeName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing optional method parameter for %s with %d children\n", methodName, paramNode.ChildCount())
	}

	for i := 0; i < int(paramNode.ChildCount()); i++ {
		child := paramNode.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Optional parameter child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "type_expression":
			typeName = t.extractTypeExpression(child)
		case "scalar", "array", "hash":
			paramName = t.getNodeText(child)
		case "=":
			// Skip the assignment operator
			continue
		case "number", "string", "bareword":
			// Skip default values - we don't need them for type annotation
			continue
		}
	}

	if paramName != "" && typeName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating optional method parameter annotation: %s for %s: %s\n", paramName, methodName, typeName)
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

	// The returnTypeNode is already the type_expression, so extract the type directly
	typeName := t.extractTypeExpression(returnTypeNode)

	if typeName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating method return type annotation for %s: %s\n", methodName, typeName)
		}
		annotation := &PerlTypeAnnotation{
			ItemName: methodName, // Use just the method name
			TypeName: typeName,
			Kind:     "method_return",
			StartPos: int(returnTypeNode.StartByte()),
			EndPos:   int(returnTypeNode.EndByte()),
			Content:  t.getNodeText(returnTypeNode),
		}
		*annotations = append(*annotations, annotation)
	} else {
		// Fallback: look for child type expressions
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
						ItemName: methodName, // Use just the method name
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
	Context  string // additional context like class name for fields
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

// processClassDeclaration processes class declarations for type annotations
func (t *PerlTree) processClassDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing class declaration with %d children\n", node.ChildCount())
	}

	var className string

	// Extract class name and process class body for fields and methods
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Class child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "package":
			// This should be the class name
			if className == "" {
				className = t.getNodeText(child)
			}
		case "block":
			// Process the class body for fields and methods
			t.processClassBlock(child, className, annotations)
		}
	}

	// Create a class declaration annotation
	if className != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating class annotation: %s\n", className)
		}

		annotation := &PerlTypeAnnotation{
			ItemName: className,
			TypeName: "class",
			Kind:     "class_declaration",
			StartPos: int(node.StartByte()),
			EndPos:   int(node.EndByte()),
			Content:  t.getNodeText(node),
		}
		*annotations = append(*annotations, annotation)
	}
}

// processClassBlock processes the contents of a class block for field and method declarations
func (t *PerlTree) processClassBlock(node *sitter.Node, className string, annotations *[]*PerlTypeAnnotation) {
	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing class block for %s with %d children\n", className, node.ChildCount())
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Class block child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "variable_declaration":
			// This should be a field declaration - process it as a class field
			t.processClassField(child, className, annotations)
		case "method_declaration_statement":
			// This is a method declaration - let the existing method processor handle it
			t.processMethodDeclaration(child, annotations)
		}
	}
}

// processClassField processes field declarations within a class
func (t *PerlTree) processClassField(node *sitter.Node, className string, annotations *[]*PerlTypeAnnotation) {
	var fieldName, typeName string

	if os.Getenv("DEBUG_PARSER") == "1" {
		fmt.Printf("DEBUG: Processing class field for %s with %d children\n", className, node.ChildCount())
	}

	// Extract field name and type from variable declaration
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Field child %d: %s (text: %s)\n", i, child.Kind(), t.getNodeText(child))
		}

		switch child.Kind() {
		case "type_expression":
			typeName = t.extractTypeExpression(child)
		case "scalar", "array", "hash":
			fieldName = t.getNodeText(child)
		case "keyword":
			// Skip 'field' keyword
			continue
		}
	}

	// Create field annotation if we found both name and type
	if fieldName != "" && typeName != "" {
		if os.Getenv("DEBUG_PARSER") == "1" {
			fmt.Printf("DEBUG: Creating field annotation: %s::%s -> %s\n", className, fieldName, typeName)
		}

		annotation := &PerlTypeAnnotation{
			ItemName: fieldName,
			TypeName: typeName,
			Kind:     "class_field",
			StartPos: int(node.StartByte()),
			EndPos:   int(node.EndByte()),
			Content:  t.getNodeText(node),
			Context:  className, // Store class name in context
		}
		*annotations = append(*annotations, annotation)
	}
}
