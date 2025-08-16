// ABOUTME: Parser implementation using tree-sitter-perl
// ABOUTME: Connects our parser API with tree-sitter-perl

package treesitter

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// AnnotationKind represents the kind of type annotation
type AnnotationKind int

const (
	// VarAnnotation is for variable annotations like "my Type $var"
	VarAnnotation AnnotationKind = iota
	// SubParamAnnotation is for subroutine parameter annotations like "sub name(Type $param)"
	SubParamAnnotation
	// SubReturnAnnotation is for subroutine return annotations like "sub ReturnType name()"
	SubReturnAnnotation
	// MethodParamAnnotation is for method parameter annotations like "method name(Type $param)"
	MethodParamAnnotation
	// MethodReturnAnnotation is for method return annotations like "method ReturnType name()"
	MethodReturnAnnotation
	// AttrAnnotation is for attribute annotations like "field Type $attr"
	AttrAnnotation
	// TypeDeclAnnotation is for type declarations like "type MyType = Type"
	TypeDeclAnnotation
	// TypeAssertionAnnotation is for type assertions like "$expr as Type"
	TypeAssertionAnnotation
)

// AST represents the Abstract Syntax Tree of a parsed Perl file
type AST struct {
	// Path is the path to the parsed file
	Path string

	// Root is the root node of the AST
	Root Node

	// TypeAnnotations is a list of type annotations found in the code
	TypeAnnotations []*TypeAnnotation

	// Errors is a list of syntax errors found during parsing
	Errors []error
}

// Node represents a node in the AST
type Node interface {
	// Type returns the type of the node
	Type() string

	// Start returns the start position of the node
	Start() Position

	// End returns the end position of the node
	End() Position

	// Children returns the child nodes
	Children() []Node

	// Text returns the text content of the node
	Text() string
}

// Position represents a position in the source code
type Position struct {
	Line   int
	Column int
	Offset int
}

// TypeExpression represents a parsed type expression
type TypeExpression struct {
	// BaseType is the base type (e.g., "Int", "ArrayRef", etc.)
	BaseType string

	// Parameters are the type parameters for parameterized types
	Parameters []*TypeExpression

	// IsUnion indicates if this is a union type (Type1|Type2)
	IsUnion bool

	// IsIntersection indicates if this is an intersection type (Type1&Type2)
	IsIntersection bool

	// IsNegation indicates if this is a negation type (!Type)
	IsNegation bool

	// UnionTypes are the types in a union type
	UnionTypes []*TypeExpression

	// IntersectionTypes are the types in an intersection type
	IntersectionTypes []*TypeExpression

	// NegatedType is the type being negated
	NegatedType *TypeExpression

	// IsStructural indicates if this is a structural type
	IsStructural bool

	// StructuralMembers are the members for structural types
	StructuralMembers []*StructuralTypeMember

	// IsConditional indicates if this is a conditional type
	IsConditional bool

	// ConditionalCondition is the check type for conditional types
	ConditionalCondition *TypeExpression

	// ConditionalTarget is the type being checked against for conditional types
	ConditionalTarget *TypeExpression

	// ConditionalRelationship is the relationship operator (extends, implements, isa, does)
	ConditionalRelationship string

	// ConditionalTrueType is the type returned when condition is true
	ConditionalTrueType *TypeExpression

	// ConditionalFalseType is the type returned when condition is false
	ConditionalFalseType *TypeExpression

	// IsTypeGuard indicates if this is a type guard type
	IsTypeGuard bool

	// TypeGuardTarget is the type being guarded for TypeGuard<T> types
	TypeGuardTarget *TypeExpression

	// OriginalString is the original string representation
	OriginalString string

	// Position is the position of the type expression in the source code
	Pos Position
}

// StructuralTypeMember represents a member in a structural type for the parser
type StructuralTypeMember struct {
	Key      string              // field key/name
	Type     *ast.TypeExpression // field type (AST format)
	Position ast.Position        // source position
}

// String returns a string representation of the type expression
func (t *TypeExpression) String() string {
	if t == nil {
		return ""
	}

	if t.OriginalString != "" {
		return t.OriginalString
	}

	if t.IsUnion && len(t.UnionTypes) > 0 {
		unionStrings := make([]string, len(t.UnionTypes))
		for i, unionType := range t.UnionTypes {
			unionStrings[i] = unionType.String()
		}
		return strings.Join(unionStrings, "|")
	}

	if t.IsIntersection && len(t.IntersectionTypes) > 0 {
		intersectionStrings := make([]string, len(t.IntersectionTypes))
		for i, intersectionType := range t.IntersectionTypes {
			intersectionStrings[i] = intersectionType.String()
		}
		return strings.Join(intersectionStrings, "&")
	}

	if t.IsNegation && t.NegatedType != nil {
		return "!" + t.NegatedType.String()
	}

	if t.IsConditional && t.ConditionalCondition != nil && t.ConditionalTarget != nil {
		return "(" + t.ConditionalCondition.String() + " " + t.ConditionalRelationship + " " +
			t.ConditionalTarget.String() + " ? " + t.ConditionalTrueType.String() + " : " +
			t.ConditionalFalseType.String() + ")"
	}

	if t.IsTypeGuard && t.TypeGuardTarget != nil {
		return "TypeGuard<" + t.TypeGuardTarget.String() + ">"
	}

	if len(t.Parameters) > 0 {
		paramStrings := make([]string, len(t.Parameters))
		for i, param := range t.Parameters {
			paramStrings[i] = param.String()
		}
		return t.BaseType + "[" + strings.Join(paramStrings, ", ") + "]"
	}

	return t.BaseType
}

// TypeAnnotation represents a type annotation found in the code
type TypeAnnotation struct {
	// AnnotatedItem is the item that has the type annotation
	AnnotatedItem string

	// TypeExpression is the type expression
	TypeExpression *TypeExpression

	// Position is the position of the type annotation
	Pos Position

	// Kind is the kind of annotation (variable, subroutine, method, attribute, etc.)
	Kind AnnotationKind
}

// Parser is the interface for a parser that uses tree-sitter-perl
type Parser struct {
	perlParser *PerlParser
	debug      bool
}

// NewParser creates a new Parser instance
func NewParser(debug bool) (*Parser, error) {
	// Initialize tree-sitter-perl parser
	perlParser, err := NewPerlParser(debug)
	if err != nil {
		return nil, err
	}

	parser := &Parser{
		perlParser: perlParser,
		debug:      debug,
	}

	if debug {
		log.Debugf("Created new tree-sitter-based Parser instance")
	}

	return parser, nil
}

// ParseFile parses a Perl file and returns its AST
func (p *Parser) ParseFile(path string) (*ast.AST, error) {
	ctx := context.Background()

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.NewSystemError("001",
			"File does not exist", err).
			WithLocation(path)
	}

	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewSystemError("002",
			"Failed to read file", err).
			WithLocation(path)
	}

	// Use ParseString to ensure consistent behavior and validation
	astResult, err := p.ParseStringWithContext(ctx, string(content))
	if err != nil {
		return nil, err
	}

	// Set the correct path in the result
	astResult.Path = path

	return astResult, nil
}

// convertPerlTypeAnnotation converts a PerlTypeAnnotation to the standard TypeAnnotation format
func (p *Parser) convertPerlTypeAnnotation(perlAnn *PerlTypeAnnotation, content string) (*TypeAnnotation, error) {
	// Calculate line and column from byte position
	pos := p.calculatePosition(perlAnn.StartPos, content)

	// Create TypeExpression - parse the type string into structured components
	typeExpr := p.parseTypeExpression(perlAnn.TypeName, pos)

	// Determine annotation kind
	var kind AnnotationKind
	switch perlAnn.Kind {
	case "variable":
		kind = VarAnnotation
	case "subroutine":
		kind = SubParamAnnotation // Legacy support
	case "subroutine_param":
		kind = SubParamAnnotation
	case "subroutine_return":
		kind = SubReturnAnnotation
	case "method":
		kind = MethodParamAnnotation // Keep existing for backwards compatibility
	case "method_parameter":
		kind = MethodParamAnnotation
	case "method_return":
		kind = MethodReturnAnnotation
	case "type_declaration":
		kind = TypeDeclAnnotation
	case "type_assertion_expression":
		kind = TypeAssertionAnnotation
	default:
		kind = VarAnnotation // Default fallback
	}

	// Add debug output for method annotation conversion
	if os.Getenv("DEBUG_PARSER") == "1" && (perlAnn.Kind == "method_parameter" || perlAnn.Kind == "method_return") {
		log.Debugf("DEBUG: Converting method annotation: %s = %s (kind: %s)",
			perlAnn.ItemName, perlAnn.TypeName, perlAnn.Kind)
	}

	return &TypeAnnotation{
		AnnotatedItem:  perlAnn.ItemName,
		TypeExpression: typeExpr,
		Pos:            pos,
		Kind:           kind,
	}, nil
}

// calculatePosition converts a byte offset to line/column position
func (p *Parser) calculatePosition(byteOffset int, content string) Position {
	if byteOffset < 0 || byteOffset > len(content) {
		return Position{Line: 1, Column: 1, Offset: byteOffset}
	}

	line := 1
	column := 1

	for i := 0; i < byteOffset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	return Position{
		Line:   line,
		Column: column,
		Offset: byteOffset,
	}
}

// parseTypeExpression parses a type expression string into a structured TypeExpression
func (p *Parser) parseTypeExpression(typeStr string, pos Position) *TypeExpression {
	typeStr = strings.TrimSpace(typeStr)

	// Handle structural types (struct { ... })
	if strings.HasPrefix(typeStr, "struct ") && strings.Contains(typeStr, "{") {
		return p.parseStructuralType(typeStr, pos)
	}

	// Handle negation types (!Type)
	if strings.HasPrefix(typeStr, "!") {
		innerType := p.parseTypeExpression(strings.TrimSpace(typeStr[1:]), pos)
		return &TypeExpression{
			BaseType:       innerType.BaseType, // Use the inner type's base type as the name
			Parameters:     innerType.Parameters,
			IsUnion:        innerType.IsUnion,
			IsIntersection: innerType.IsIntersection,
			IsNegation:     true,
			NegatedType:    innerType,
			OriginalString: typeStr,
			Pos:            pos,
		}
	}

	// Handle union types (Type1|Type2) - enhanced to support complex expressions
	if strings.Contains(typeStr, "|") {
		// Parse union types with proper bracket awareness
		parts := p.splitUnionTypes(typeStr)
		if len(parts) >= 2 {
			var unionTypes []*TypeExpression
			for _, part := range parts {
				unionTypes = append(unionTypes, p.parseTypeExpression(strings.TrimSpace(part), pos))
			}
			return &TypeExpression{
				BaseType:       typeStr, // Keep full expression as name for now
				IsUnion:        true,
				UnionTypes:     unionTypes,
				OriginalString: typeStr,
				Pos:            pos,
			}
		}
	}

	// Handle intersection types (Type1&Type2) - enhanced to support complex expressions
	if strings.Contains(typeStr, "&") {
		// Parse intersection types with proper bracket awareness
		parts := p.splitIntersectionTypes(typeStr)
		if len(parts) >= 2 {
			var intersectionTypes []*TypeExpression
			for _, part := range parts {
				intersectionTypes = append(intersectionTypes, p.parseTypeExpression(strings.TrimSpace(part), pos))
			}
			return &TypeExpression{
				BaseType:          typeStr, // Keep full expression as name for now
				IsIntersection:    true,
				IntersectionTypes: intersectionTypes,
				OriginalString:    typeStr,
				Pos:               pos,
			}
		}
	}

	// Handle conditional types ((T extends U ? X : Y))
	if strings.HasPrefix(typeStr, "(") && strings.HasSuffix(typeStr, ")") {
		innerStr := strings.TrimSpace(typeStr[1 : len(typeStr)-1])

		// Look for conditional type pattern with multiple operators
		if strings.Contains(innerStr, "?") && strings.Contains(innerStr, ":") {
			return p.parseConditionalType(innerStr, pos, typeStr)
		}
	}

	// Handle type guards (TypeGuard<T>)
	if strings.HasPrefix(typeStr, "TypeGuard<") && strings.HasSuffix(typeStr, ">") {
		targetStr := strings.TrimSpace(typeStr[10 : len(typeStr)-1]) // Remove "TypeGuard<" and ">"
		targetType := p.parseTypeExpression(targetStr, pos)

		return &TypeExpression{
			BaseType:        typeStr,
			IsTypeGuard:     true,
			TypeGuardTarget: targetType,
			OriginalString:  typeStr,
			Pos:             pos,
		}
	}

	// Handle parameterized types (Type[Param1, Param2])
	if idx := strings.Index(typeStr, "["); idx > 0 && strings.HasSuffix(typeStr, "]") {
		baseName := strings.TrimSpace(typeStr[:idx])
		paramStr := strings.TrimSpace(typeStr[idx+1 : len(typeStr)-1])

		var params []*TypeExpression
		if paramStr != "" {
			// Parse parameters, handling nested brackets
			paramParts := p.splitParameters(paramStr)
			for _, part := range paramParts {
				params = append(params, p.parseTypeExpression(strings.TrimSpace(part), pos))
			}
		}

		return &TypeExpression{
			BaseType:       baseName,
			Parameters:     params,
			OriginalString: typeStr,
			Pos:            pos,
		}
	}

	// Simple type (no special syntax)
	return &TypeExpression{
		BaseType:       typeStr,
		OriginalString: typeStr,
		Pos:            pos,
	}
}

// splitUnionTypes splits a union type string into components while respecting bracket nesting
func (p *Parser) splitUnionTypes(typeStr string) []string {
	var parts []string
	var currentPart strings.Builder
	bracketDepth := 0
	parenDepth := 0

	for _, char := range typeStr {
		switch char {
		case '[':
			bracketDepth++
			currentPart.WriteRune(char)
		case ']':
			bracketDepth--
			currentPart.WriteRune(char)
		case '(':
			parenDepth++
			currentPart.WriteRune(char)
		case ')':
			parenDepth--
			currentPart.WriteRune(char)
		case '|':
			if bracketDepth == 0 && parenDepth == 0 {
				// This is a top-level union separator
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			} else {
				// This | is inside brackets or parentheses, keep it
				currentPart.WriteRune(char)
			}
		default:
			currentPart.WriteRune(char)
		}
	}

	// Add the final part
	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}

// splitIntersectionTypes splits an intersection type string into components while respecting bracket nesting
func (p *Parser) splitIntersectionTypes(typeStr string) []string {
	var parts []string
	var currentPart strings.Builder
	bracketDepth := 0
	parenDepth := 0

	for _, char := range typeStr {
		switch char {
		case '[':
			bracketDepth++
			currentPart.WriteRune(char)
		case ']':
			bracketDepth--
			currentPart.WriteRune(char)
		case '(':
			parenDepth++
			currentPart.WriteRune(char)
		case ')':
			parenDepth--
			currentPart.WriteRune(char)
		case '&':
			if bracketDepth == 0 && parenDepth == 0 {
				// This is a top-level intersection separator
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			} else {
				// This & is inside brackets or parentheses, keep it
				currentPart.WriteRune(char)
			}
		default:
			currentPart.WriteRune(char)
		}
	}

	// Add the final part
	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}

// splitParameters splits parameter strings, handling nested brackets
func (p *Parser) splitParameters(paramStr string) []string {
	var params []string
	var current strings.Builder
	bracketCount := 0

	for _, char := range paramStr {
		switch char {
		case '[':
			bracketCount++
			current.WriteRune(char)
		case ']':
			bracketCount--
			current.WriteRune(char)
		case ',':
			if bracketCount == 0 {
				params = append(params, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		params = append(params, current.String())
	}

	return params
}

// ParseString parses a string containing Perl code and returns its AST
// ParseStringWithContext parses a Perl code string with context
func (p *Parser) ParseStringWithContext(ctx context.Context, content string) (*ast.AST, error) {
	return p.parseStringInternal(ctx, content)
}

func (p *Parser) ParseString(content string) (*ast.AST, error) {
	ctx := context.Background()
	return p.parseStringInternal(ctx, content)
}

// parseStringInternal contains the actual parsing logic
func (p *Parser) parseStringInternal(ctx context.Context, content string) (*ast.AST, error) {
	// Parse the string using tree-sitter
	tree, err := p.perlParser.ParseString(content)
	if err != nil {
		return nil, errors.NewSystemError("004",
			"Failed to parse string", err)
	}

	// Extract type annotations from the tree
	perlAnnotations, err := tree.FindTypeAnnotations()
	if err != nil {
		return nil, errors.NewSystemError("005",
			"Failed to extract type annotations", err)
	}

	// Convert PerlTypeAnnotations to TypeAnnotations
	var typeAnnotations []*TypeAnnotation
	for _, perlAnn := range perlAnnotations {
		annotation, convErr := p.convertPerlTypeAnnotation(perlAnn, content)
		if convErr != nil {
			// Log the error but don't fail the parse
			log.Debugf("Failed to convert annotation %s: %v", perlAnn.ItemName, convErr)
			continue
		}
		typeAnnotations = append(typeAnnotations, annotation)
	}

	// Convert tree-sitter root to our Node interface
	var rootNode Node
	if tree.Root() != nil {
		rootNode = p.convertTreeSitterNode(tree.Root(), tree)
	} else {
		rootNode = newSimpleNode("root")
	}

	// Convert to proper ast.AST with ast.Node types
	astRoot := p.convertToASTNode(rootNode)

	// Convert type annotations to ast.TypeAnnotation
	var astTypeAnnotations []*ast.TypeAnnotation
	for _, ta := range typeAnnotations {
		astTypeAnnotations = append(astTypeAnnotations, p.convertToASTTypeAnnotation(ta))
	}
	// Ensure slice is never nil, even if empty
	if astTypeAnnotations == nil {
		astTypeAnnotations = make([]*ast.TypeAnnotation, 0)
	}

	// Add constraint-based type annotations by scanning source text
	constraintAnnotations := p.extractConstraintAnnotations(content)
	astTypeAnnotations = append(astTypeAnnotations, constraintAnnotations...)

	// Check for ERROR nodes in the tree-sitter parse result
	if os.Getenv("PARSER_DEBUG") != "" {
		fmt.Printf("[PARSER_DEBUG] Checking for ERROR nodes in content: %q\n", content)
		fmt.Printf("[PARSER_DEBUG] AST root type: %s, text: %q\n", astRoot.Type(), astRoot.Text())
	}
	if errorNodes := p.findErrorNodes(astRoot); len(errorNodes) > 0 {
		if os.Getenv("PARSER_DEBUG") != "" {
			fmt.Printf("[PARSER_DEBUG] Found %d ERROR nodes\n", len(errorNodes))
			for i, node := range errorNodes {
				fmt.Printf("[PARSER_DEBUG] ERROR node %d: %q at %d:%d\n", i, node.Content, node.StartPoint.Line, node.StartPoint.Column)
			}
		}

		// Try to identify specific type errors first
		// Note: Type error identification should be handled at the parser layer, not tree-sitter layer
		// For now, continue with generic tree-sitter errors

		// Fall back to generic tree-sitter error formatting
		errorMessage := p.formatParseErrors(errorNodes, "", content)
		return nil, errors.NewSystemError("007", errorMessage, nil)
	}
	if os.Getenv("PARSER_DEBUG") != "" {
		fmt.Printf("[PARSER_DEBUG] No ERROR nodes found\n")
	}

	// Create an AST from the parse tree
	astResult := &ast.AST{
		Source:          content,
		Root:            astRoot,
		TypeAnnotations: astTypeAnnotations,
		Errors:          nil, // ERROR nodes are now handled above, no additional validation needed
	}

	return astResult, nil
}

// ParseReader parses Perl code from a reader and returns its AST
func (p *Parser) ParseReader(reader io.Reader) (*ast.AST, error) {
	// Read all content from the reader
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewSystemError("006",
			"Failed to read from reader", err)
	}

	// Parse the content as a string
	return p.ParseString(string(content))
}

// Close releases resources used by the parser
func (p *Parser) Close() {
	if p.perlParser != nil {
		p.perlParser.Close()
		p.perlParser = nil
	}
}

// SimpleNode is a simple implementation of the Node interface
type SimpleNode struct {
	NodeType     string
	Text_        string
	StartPos     Position
	EndPos       Position
	NodeChildren []Node
}

// Type implements the Node interface
func (n *SimpleNode) Type() string {
	return n.NodeType
}

// Start implements the Node interface
func (n *SimpleNode) Start() Position {
	return n.StartPos
}

// End implements the Node interface
func (n *SimpleNode) End() Position {
	return n.EndPos
}

// Children implements the Node interface
func (n *SimpleNode) Children() []Node {
	return n.NodeChildren
}

// Text implements the Node interface
func (n *SimpleNode) Text() string {
	return n.Text_
}

// parseStructuralType parses structural type expressions (struct { key: Type, ... })
func (p *Parser) parseStructuralType(typeStr string, pos Position) *TypeExpression {
	// Extract content between braces
	startBrace := strings.Index(typeStr, "{")
	endBrace := strings.LastIndex(typeStr, "}")

	if startBrace == -1 || endBrace == -1 || endBrace <= startBrace {
		// Malformed structural type, return simple type
		return &TypeExpression{
			BaseType:       typeStr,
			OriginalString: typeStr,
			Pos:            pos,
		}
	}

	membersStr := strings.TrimSpace(typeStr[startBrace+1 : endBrace])

	var members []*StructuralTypeMember
	if membersStr != "" {
		// Split by commas (simple parsing for now)
		memberStrs := strings.Split(membersStr, ",")
		for _, memberStr := range memberStrs {
			memberStr = strings.TrimSpace(memberStr)
			if memberStr == "" {
				continue
			}

			// Parse "key: Type" format
			colonPos := strings.Index(memberStr, ":")
			if colonPos == -1 {
				continue
			}

			key := strings.TrimSpace(memberStr[:colonPos])
			typeStr := strings.TrimSpace(memberStr[colonPos+1:])

			if key != "" && typeStr != "" {
				memberType := p.parseTypeExpression(typeStr, pos)
				memberAST := p.convertToASTTypeExpression(memberType)

				member := &StructuralTypeMember{
					Key:      key,
					Type:     memberAST,
					Position: ast.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
				}
				members = append(members, member)
			}
		}
	}

	return &TypeExpression{
		BaseType:          "struct",
		OriginalString:    typeStr,
		Pos:               pos,
		IsStructural:      true,
		StructuralMembers: members,
	}
}

// newSimpleNode creates a new SimpleNode
func newSimpleNode(nodeType string) Node {
	return &SimpleNode{
		NodeType:     nodeType,
		NodeChildren: []Node{},
		StartPos:     Position{Line: 1, Column: 1, Offset: 0},
		EndPos:       Position{Line: 1, Column: 1, Offset: 0},
	}
}

// convertTreeSitterNode converts a tree-sitter node to our Node interface
func (p *Parser) convertTreeSitterNode(tsNode *sitter.Node, tree *PerlTree) Node {
	if tsNode == nil {
		return newSimpleNode("root")
	}

	// Get node information
	nodeType := tsNode.Kind()
	nodeText := tree.GetNodeText(tsNode)

	// Convert positions
	startPos := Position{
		Line:   int(tsNode.StartPosition().Row) + 1, // tree-sitter is 0-based, we want 1-based
		Column: int(tsNode.StartPosition().Column) + 1,
		Offset: int(tsNode.StartByte()),
	}
	endPos := Position{
		Line:   int(tsNode.EndPosition().Row) + 1,
		Column: int(tsNode.EndPosition().Column) + 1,
		Offset: int(tsNode.EndByte()),
	}

	// Convert children
	var children []Node
	childCount := tsNode.ChildCount()
	for i := uint(0); i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil {
			children = append(children, p.convertTreeSitterNode(child, tree))
		}
	}

	return &SimpleNode{
		NodeType:     nodeType,
		Text_:        nodeText,
		StartPos:     startPos,
		EndPos:       endPos,
		NodeChildren: children,
	}
}

// convertToASTNode converts a tree-sitter Node to an ast.Node
func (p *Parser) convertToASTNode(node Node) ast.Node {
	if node == nil {
		return nil
	}

	// Convert position from tree-sitter to ast format
	start := ast.Position{
		Line:   node.Start().Line,
		Column: node.Start().Column,
		Offset: node.Start().Offset,
	}
	end := ast.Position{
		Line:   node.End().Line,
		Column: node.End().Column,
		Offset: node.End().Offset,
	}

	nodeType := node.Type()
	nodeText := node.Text()

	// Don't skip type-related nodes - let AST compiler handle them
	// The AST compiler needs to see type nodes to properly remove them
	// if p.isTypeRelatedNode(nodeType) {
	//     return nil
	// }

	// Handle special cases that need custom processing
	if nodeType == "typed_method_parameter" {
		return p.convertTypedParameterToToken(nodeText, start, end)
	}

	// First, check if this is a structural token that should be preserved as a token
	if tokenNode := p.convertToTokenNodeWithParams(nodeType, nodeText, start, end); tokenNode != nil {
		return tokenNode
	}

	// Convert children recursively - include ALL children to preserve CST structure
	var children []ast.Node
	for _, child := range node.Children() {
		if childAST := p.convertToASTNode(child); childAST != nil {
			children = append(children, childAST)
		}
	}

	// Handle subroutine declarations specifically
	if nodeType == "subroutine_declaration_statement" {
		return p.convertToSubDeclAST(node, start, end)
	}

	// Handle method declarations specifically
	if nodeType == "method_declaration_statement" {
		return p.extractMethodDecl(node)
	}

	// Handle class declarations specifically
	if nodeType == "class_statement" {
		return p.convertToClassDeclAST(node, start, end)
	}

	// Handle role declarations specifically
	if nodeType == "role_statement" {
		return p.convertToRoleDeclAST(node, start, end)
	}

	// Handle use statements specifically
	if nodeType == "use_statement" {
		return p.convertToUseStmtAST(node, start, end)
	}

	// Handle block statements specifically
	if nodeType == "block" {
		return p.parseBlock(node, start, end)
	}

	// Handle higher-level constructs that need semantic processing
	if nodeType == "assignment_expression" &&
		(strings.HasPrefix(strings.TrimSpace(nodeText), "my ") ||
			strings.HasPrefix(strings.TrimSpace(nodeText), "our ") ||
			strings.HasPrefix(strings.TrimSpace(nodeText), "state ")) {
		// fmt.Printf("DEBUG: Processing assignment_expression as VarDecl: %q\n", nodeText)
		return p.convertAssignmentToVarDeclAST(node, start, end)
	}

	// Handle variable_declaration nodes
	if nodeType == "variable_declaration" {
		// fmt.Printf("DEBUG: Processing variable_declaration node: %q\n", nodeText)
		// Skip processing this node directly - let the assignment_expression handle it
		// This prevents duplicate VarDecl nodes from being created
		return nil
	}

	// For nodes with children, create a container node that preserves structure
	if len(children) > 0 {
		return p.createContainerNode(nodeType, children, start, end)
	}

	// For leaf nodes without children, create a generic expression
	return ast.NewExpressionStmt(
		ast.NewLiteralExpr(nodeText, ast.StringLiteral, start, end),
		start, end)
}

// convertAssignmentToVarDeclAST converts an assignment_expression containing a variable declaration to ast.VarDecl
func (p *Parser) convertAssignmentToVarDeclAST(node Node, start, end ast.Position) ast.Node {
	fmt.Printf("DEBUG: convertAssignmentToVarDeclAST called\n")
	// Parse the assignment_expression structure:
	// assignment_expression
	//   variable_declaration (left side)
	//   = (operator)
	//   <initializer> (right side)

	var varDeclNode Node
	var initializerNode Node

	children := node.Children()
	for _, child := range children {
		childType := child.Type()

		if childType == "variable_declaration" {
			varDeclNode = child
		} else if childType == "=" {
			// Skip the assignment operator
			continue
		} else {
			// This should be the initializer (number, function call, etc.)
			initializerNode = child
		}
	}

	if varDeclNode == nil {
		// Fallback to old string-based parsing if structure is unexpected
		return p.convertToVarDeclAST(node.Text(), start, end)
	}

	// Parse the variable declaration part
	var declType string
	var variables []*ast.VariableExpr

	for _, child := range varDeclNode.Children() {
		childType := child.Type()
		childText := strings.TrimSpace(child.Text())

		switch childType {
		case "my", "state", "our", "field", "local":
			declType = childText
		case "scalar", "array", "hash":
			varName := p.extractVariableName(child)
			if varName != "" {
				variables = append(variables, ast.NewVariableExpr("", varName, start, end))
			}
		}
	}

	// Parse the initializer if present
	var initializer ast.ExpressionNode
	if initializerNode != nil {
		initStart := ast.Position{
			Line:   initializerNode.Start().Line,
			Column: initializerNode.Start().Column,
			Offset: initializerNode.Start().Offset,
		}
		initEnd := ast.Position{
			Line:   initializerNode.End().Line,
			Column: initializerNode.End().Column,
			Offset: initializerNode.End().Offset,
		}
		initializer = p.parseInitializerExpression(initializerNode, initStart, initEnd)
	}

	// Create VarDecl with the initializer
	if len(variables) > 0 {
		result := ast.NewVarDecl(declType, variables, nil, initializer, start, end)
		fmt.Printf("DEBUG: Created VarDecl with initializer: %v\n", result.Initializer != nil)
		if result.Initializer != nil {
			fmt.Printf("DEBUG: Initializer type: %s, text: %q\n", result.Initializer.Type(), result.Initializer.Text())
		}
		return result
	}

	return nil
}

// convertToVarDeclAST converts a variable declaration text to ast.VarDecl (fallback)
func (p *Parser) convertToVarDeclAST(nodeText string, start, end ast.Position) ast.Node {
	fmt.Printf("DEBUG: convertToVarDeclAST (fallback) called with text: %q\n", nodeText)
	// Parse the variable declaration
	parts := strings.Fields(strings.TrimSpace(nodeText))
	if len(parts) < 2 {
		return nil // Invalid declaration
	}

	declType := parts[0] // "my", "our", "state"
	varName := parts[1]  // "$foo"

	// Remove sigil to get just the variable name and determine sigil
	var name, sigil string
	if len(varName) > 1 && (varName[0] == '$' || varName[0] == '@' || varName[0] == '%') {
		sigil = string(varName[0])
		name = varName[1:]
	} else {
		name = varName
		sigil = "$" // default
	}

	// Create variable expression
	varExpr := ast.NewVariableExpr(name, sigil, start, end)

	// Create variable declaration
	return ast.NewVarDecl(declType, []*ast.VariableExpr{varExpr}, nil, nil, start, end)
}

// convertToASTTypeAnnotation converts tree-sitter TypeAnnotation to ast.TypeAnnotation
func (p *Parser) convertToASTTypeAnnotation(ta *TypeAnnotation) *ast.TypeAnnotation {
	// Convert the type annotation from tree-sitter format to ast format
	astPos := ast.Position{Line: ta.Pos.Line, Column: ta.Pos.Column}

	// Convert treesitter.AnnotationKind to ast.AnnotationKind
	var astKind ast.AnnotationKind
	switch ta.Kind {
	case VarAnnotation:
		astKind = ast.VarAnnotation
	case SubParamAnnotation:
		astKind = ast.SubParamAnnotation
	case SubReturnAnnotation:
		astKind = ast.SubReturnAnnotation
	case MethodParamAnnotation:
		astKind = ast.MethodParamAnnotation
	case MethodReturnAnnotation:
		astKind = ast.MethodReturnAnnotation
	case AttrAnnotation:
		astKind = ast.FieldAnnotation
	case TypeDeclAnnotation:
		astKind = ast.TypeDeclAnnotation
	case TypeAssertionAnnotation:
		astKind = ast.TypeAssertionAnnotation
	default:
		astKind = ast.VarAnnotation // Default fallback
	}

	return &ast.TypeAnnotation{
		Kind:           astKind,
		AnnotatedItem:  ta.AnnotatedItem,
		TypeExpression: p.convertToASTTypeExpression(ta.TypeExpression),
		Pos:            astPos,
	}
}

// convertToASTTypeExpression converts tree-sitter TypeExpression to ast.TypeExpression
func (p *Parser) convertToASTTypeExpression(te *TypeExpression) *ast.TypeExpression {
	if te == nil {
		return nil
	}

	astPos := ast.Position{Line: te.Pos.Line, Column: te.Pos.Column}

	// Convert parameters recursively
	var astParams []*ast.TypeExpression
	for _, param := range te.Parameters {
		astParams = append(astParams, p.convertToASTTypeExpression(param))
	}

	// Convert union types recursively
	var astUnionTypes []*ast.TypeExpression
	for _, unionType := range te.UnionTypes {
		astUnionTypes = append(astUnionTypes, p.convertToASTTypeExpression(unionType))
	}

	// Convert intersection types recursively
	var astIntersectionTypes []*ast.TypeExpression
	for _, intersectionType := range te.IntersectionTypes {
		astIntersectionTypes = append(astIntersectionTypes, p.convertToASTTypeExpression(intersectionType))
	}

	// Convert structural members
	var astStructuralMembers []*ast.StructuralTypeMember
	for _, member := range te.StructuralMembers {
		astStructuralMembers = append(astStructuralMembers, &ast.StructuralTypeMember{
			Key:      member.Key,
			Type:     member.Type, // Already in AST format
			Position: member.Position,
		})
	}

	// Reconstruct full type name including parameters
	fullTypeName := te.BaseType
	if len(astParams) > 0 {
		var paramNames []string
		for _, param := range astParams {
			paramNames = append(paramNames, param.Name)
		}
		fullTypeName = te.BaseType + "[" + strings.Join(paramNames, ",") + "]"
	}

	astTypeExpr := ast.NewTypeExpression(fullTypeName, astParams, astPos, astPos)
	astTypeExpr.IsUnion = te.IsUnion
	astTypeExpr.IsIntersection = te.IsIntersection
	astTypeExpr.IsNegation = te.IsNegation
	astTypeExpr.UnionTypes = astUnionTypes
	astTypeExpr.IntersectionTypes = astIntersectionTypes
	astTypeExpr.NegatedType = p.convertToASTTypeExpression(te.NegatedType)
	astTypeExpr.StructuralMembers = astStructuralMembers
	astTypeExpr.OriginalString = te.OriginalString

	// Handle conditional types
	if te.IsConditional {
		astTypeExpr.Kind = ast.ConditionalTypeKind
		astTypeExpr.ConditionalCondition = p.convertToASTTypeExpression(te.ConditionalCondition)
		astTypeExpr.ConditionalTarget = p.convertToASTTypeExpression(te.ConditionalTarget)
		astTypeExpr.ConditionalRelationship = te.ConditionalRelationship
		astTypeExpr.ConditionalTrueType = p.convertToASTTypeExpression(te.ConditionalTrueType)
		astTypeExpr.ConditionalFalseType = p.convertToASTTypeExpression(te.ConditionalFalseType)
	}

	// Handle type guards
	if te.IsTypeGuard {
		astTypeExpr.Kind = ast.TypeGuardKind
		astTypeExpr.TypeGuardTarget = p.convertToASTTypeExpression(te.TypeGuardTarget)
	}

	// Set the kind for structural types
	if te.IsStructural {
		astTypeExpr.Kind = ast.StructuralTypeKind
	}

	return astTypeExpr
}

// ParseTypeExpression parses a type expression string and returns a TypeExpression
func ParseTypeExpression(text string, pos Position) (*TypeExpression, error) {
	// Parse in order of precedence: operators first (union, intersection, negation), then parameterized types

	// Check for union types (Type1|Type2) - highest precedence for operators
	if containsTopLevelOperator(text, "|") {
		unionParts := splitTopLevelOperator(text, "|")
		unionTypes := make([]*TypeExpression, len(unionParts))

		for i, part := range unionParts {
			partExpr, err := ParseTypeExpression(strings.TrimSpace(part), pos)
			if err != nil {
				return nil, err
			}
			unionTypes[i] = partExpr
		}

		return &TypeExpression{
			IsUnion:        true,
			UnionTypes:     unionTypes,
			OriginalString: text,
			Pos:            pos,
		}, nil
	}

	// Check for intersection types (Type1&Type2) - only at top level
	if containsTopLevelOperator(text, "&") {
		intersectionParts := splitTopLevelOperator(text, "&")
		intersectionTypes := make([]*TypeExpression, len(intersectionParts))

		for i, part := range intersectionParts {
			partExpr, err := ParseTypeExpression(strings.TrimSpace(part), pos)
			if err != nil {
				return nil, err
			}
			intersectionTypes[i] = partExpr
		}

		return &TypeExpression{
			IsIntersection:    true,
			IntersectionTypes: intersectionTypes,
			OriginalString:    text,
			Pos:               pos,
		}, nil
	}

	// Check for negation types (!Type)
	if strings.HasPrefix(text, "!") {
		negatedType, err := ParseTypeExpression(strings.TrimSpace(text[1:]), pos)
		if err != nil {
			return nil, err
		}

		return &TypeExpression{
			IsNegation:     true,
			NegatedType:    negatedType,
			OriginalString: text,
			Pos:            pos,
		}, nil
	}

	// Check for parameterized types (Type[Param1, Param2, ...]) after operators
	if strings.Contains(text, "[") && strings.HasSuffix(text, "]") {
		openBracket := strings.Index(text, "[")
		baseType := strings.TrimSpace(text[:openBracket])
		paramsText := text[openBracket+1 : len(text)-1]

		// Split parameters by comma, handling nested brackets
		paramsList := splitParams(paramsText)
		params := make([]*TypeExpression, len(paramsList))

		for i, paramText := range paramsList {
			paramExpr, err := ParseTypeExpression(strings.TrimSpace(paramText), pos)
			if err != nil {
				return nil, err
			}
			params[i] = paramExpr
		}

		return &TypeExpression{
			BaseType:       baseType,
			Parameters:     params,
			OriginalString: text,
			Pos:            pos,
		}, nil
	}

	// Simple type
	return &TypeExpression{
		BaseType:       text,
		OriginalString: text,
		Pos:            pos,
	}, nil
}

// containsTopLevelOperator checks if the text contains the operator at the top level (not inside brackets)
func containsTopLevelOperator(text, operator string) bool {
	bracketDepth := 0
	for i, char := range text {
		switch char {
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		default:
			if bracketDepth == 0 && strings.HasPrefix(text[i:], operator) {
				return true
			}
		}
	}
	return false
}

// splitTopLevelOperator splits text by operator only at the top level (not inside brackets)
func splitTopLevelOperator(text, operator string) []string {
	var parts []string
	bracketDepth := 0
	lastSplit := 0

	for i, char := range text {
		switch char {
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		default:
			if bracketDepth == 0 && strings.HasPrefix(text[i:], operator) {
				parts = append(parts, text[lastSplit:i])
				lastSplit = i + len(operator)
			}
		}
	}
	parts = append(parts, text[lastSplit:])
	return parts
}

// splitParams splits a parameter list by comma, handling nested brackets
func splitParams(params string) []string {
	var result []string
	var currentParam strings.Builder
	bracketCount := 0

	for _, c := range params {
		switch c {
		case '[':
			bracketCount++
			currentParam.WriteRune(c)
		case ']':
			bracketCount--
			currentParam.WriteRune(c)
		case ',':
			if bracketCount == 0 {
				// End of parameter
				result = append(result, currentParam.String())
				currentParam.Reset()
			} else {
				// Comma inside brackets, part of the parameter
				currentParam.WriteRune(c)
			}
		default:
			currentParam.WriteRune(c)
		}
	}

	// Add the last parameter
	if currentParam.Len() > 0 {
		result = append(result, currentParam.String())
	}

	return result
}

// convertToClassDeclAST converts a class_statement node to ast.ClassDecl
func (p *Parser) convertToClassDeclAST(node Node, start, end ast.Position) ast.Node {
	className := ""

	// Extract constraints from the class signature
	constraints := p.extractConstraintsFromText(node.Text())

	// Extract class name from tree-sitter node
	for _, child := range node.Children() {
		if child.Type() == "package" {
			className = child.Text()
			break
		}
	}

	// Create class declaration
	classDecl := ast.NewClassDecl(className, start, end)
	classDecl.Constraints = constraints

	// Process children to find fields and methods
	for _, child := range node.Children() {
		switch child.Type() {
		case "block":
			// Process block contents for fields and methods
			p.processClassBlock(child, classDecl)
		case "package":
			// Already handled above
			continue
		}
	}

	return classDecl
}

// convertToRoleDeclAST converts a role_statement node to ast.RoleDecl
func (p *Parser) convertToRoleDeclAST(node Node, start, end ast.Position) ast.Node {
	roleName := ""

	// Extract constraints from the role signature
	constraints := p.extractConstraintsFromText(node.Text())

	// Extract role name from tree-sitter node
	for _, child := range node.Children() {
		if child.Type() == "package" {
			roleName = child.Text()
			break
		}
	}

	// Create role declaration
	roleDecl := ast.NewRoleDecl(roleName, start, end)
	roleDecl.Constraints = constraints

	// Process children to find fields and methods
	for _, child := range node.Children() {
		switch child.Type() {
		case "block":
			// Process block contents for fields and methods
			p.processRoleBlock(child, roleDecl)
		case "package":
			// Already handled above
			continue
		}
	}

	return roleDecl
}

// convertToUseStmtAST converts a use_statement node to ast.UseStmt
func (p *Parser) convertToUseStmtAST(node Node, start, end ast.Position) ast.Node {
	module := ""
	version := ""
	var imports []string

	// Extract fields from tree-sitter node children
	for _, child := range node.Children() {
		switch child.Type() {
		case "package":
			module = child.Text()
		case "_version":
			version = child.Text()
		case "_listexpr":
			// Parse import list from expressions like qw(has with)
			imports = p.parseImportList(child)
		}
	}

	return ast.NewUseStmt(module, version, imports, start, end)
}

// parseImportList extracts imports from a _listexpr node
func (p *Parser) parseImportList(listNode Node) []string {
	var imports []string

	// Handle qw() syntax and other list expressions
	text := listNode.Text()

	// Simple parsing for qw(symbol1 symbol2) format
	if strings.Contains(text, "qw(") {
		// Extract content between qw( and )
		start := strings.Index(text, "qw(") + 3
		end := strings.LastIndex(text, ")")
		if start < end && end > 0 {
			content := text[start:end]
			// Split by whitespace to get individual imports
			for _, item := range strings.Fields(content) {
				item = strings.TrimSpace(item)
				if item != "" {
					imports = append(imports, item)
				}
			}
		}
	} else {
		// For non-qw syntax, try to extract identifiers
		for _, child := range listNode.Children() {
			if child.Type() == "identifier" || child.Type() == "string" {
				imports = append(imports, strings.Trim(child.Text(), `"'`))
			}
		}
	}

	return imports
}

// convertToSubDeclAST converts a subroutine_declaration_statement node to ast.SubDecl
func (p *Parser) convertToSubDeclAST(node Node, start, end ast.Position) ast.Node {
	subroutineName := ""
	var params []*ast.Parameter
	var returnType *ast.TypeExpression
	var body *ast.BlockStmt

	// Extract subroutine name from tree-sitter node structure
	// For typed subroutines: subroutine_declaration_statement has children: sub, bareword, signature, return_type, block
	// For simple subroutines: subroutine_declaration_statement has children: sub, bareword, block
	children := node.Children()

	for _, child := range children {
		childType := child.Type()

		switch childType {
		case "bareword":
			// This should be the subroutine name (e.g., "hello", "add")
			subroutineName = child.Text()
		case "signature":
			// Parse typed method parameters
			params = p.parseSignatureParams(child)
		case "type_expression":
			// Parse return type - tree-sitter presents it as a direct type_expression
			returnType = p.parseTreeSitterTypeExpression(child)
		case "block":
			// Convert the block to a BlockStmt
			blockStart := ast.Position{
				Line:   child.Start().Line,
				Column: child.Start().Column,
			}
			blockEnd := ast.Position{
				Line:   child.End().Line,
				Column: child.End().Column,
			}

			// Parse block contents using new pattern
			body = p.parseBlock(child, blockStart, blockEnd)
		}
	}

	// Create subroutine declaration
	return ast.NewSubDecl(subroutineName, params, returnType, body, false, start, end)
}

// parseSignatureParams parses the signature node to extract typed parameters
func (p *Parser) parseSignatureParams(signatureNode Node) []*ast.Parameter {
	var params []*ast.Parameter

	// The signature contains various parameter types
	children := signatureNode.Children()
	for _, child := range children {
		childType := child.Type()
		if p.debug {
			fmt.Printf("DEBUG: Signature child type: %s, text: '%s'\n", childType, child.Text())
		}
		switch childType {
		case "typed_method_parameter", "mandatory_parameter", "optional_parameter":
			if p.debug {
				fmt.Printf("DEBUG: Parsing parameter of type %s\n", childType)
			}
			param := p.parseTypedParameter(child)
			if param != nil {
				if p.debug {
					fmt.Printf("DEBUG: Created parameter %s, default: %v\n", param.Name, param.Default != nil)
				}
				params = append(params, param)
			}
		case "scalar":
			// Handle untyped parameters
			paramName := strings.TrimPrefix(child.Text(), "$")
			if paramName != "" {
				param := &ast.Parameter{
					Name: paramName,
					Pos: ast.Position{
						Line:   child.Start().Line,
						Column: child.Start().Column,
					},
				}
				params = append(params, param)
			}
		}
	}

	return params
}

// parseTypedParameter parses a typed_method_parameter node
func (p *Parser) parseTypedParameter(paramNode Node) *ast.Parameter {
	var paramName string
	var paramType *ast.TypeExpression
	var defaultValue ast.ExpressionNode

	// Handle different parameter node types
	nodeType := paramNode.Type()

	// typed_method_parameter has children: type_expression, scalar
	// optional_parameter has children: type_expression?, scalar, default
	for _, child := range paramNode.Children() {
		childType := child.Type()
		switch childType {
		case "type_expression":
			// Extract the type name from the type_expression
			paramType = p.parseTreeSitterTypeExpression(child)
		case "scalar":
			// Extract variable name from scalar (e.g., "$a" -> "a")
			varText := child.Text()
			if len(varText) > 1 && varText[0] == '$' {
				paramName = varText[1:] // Remove the $ sigil
			}
		case "interpolated_string_literal", "string_literal", "number", "integer":
			// This is likely a default value for optional parameters
			defaultValue = p.parseSimpleExpression(child)
		case "anonymous_array_expression":
			// This is an array literal default value
			defaultValue = p.parseSimpleExpression(child)
		}
	}

	// For optional parameters, extract the default value from the text
	if nodeType == "optional_parameter" && defaultValue == nil {
		paramText := paramNode.Text()
		if p.debug {
			fmt.Printf("DEBUG parseTypedParameter: optional_parameter text='%s'\n", paramText)
		}
		// Look for pattern: Type $var = value or $var = value
		if equalPos := strings.Index(paramText, " = "); equalPos != -1 {
			// Extract everything after " = "
			defaultStr := strings.TrimSpace(paramText[equalPos+3:])
			if p.debug {
				fmt.Printf("DEBUG parseTypedParameter: found default value='%s'\n", defaultStr)
			}
			if defaultStr != "" {
				// Create a simple literal expression for the default value
				pos := ast.Position{
					Line:   paramNode.Start().Line,
					Column: paramNode.Start().Column,
				}
				// Determine the literal type based on the content
				var literalKind ast.LiteralKind
				if strings.HasPrefix(defaultStr, "\"") && strings.HasSuffix(defaultStr, "\"") {
					literalKind = ast.StringLiteral
				} else if strings.HasPrefix(defaultStr, "'") && strings.HasSuffix(defaultStr, "'") {
					literalKind = ast.StringLiteral
				} else if strings.Contains(defaultStr, ".") {
					literalKind = ast.NumberLiteral
				} else {
					// Integers, booleans, etc.
					literalKind = ast.NumberLiteral
				}
				defaultValue = ast.NewLiteralExpr(defaultStr, literalKind, pos, pos)
				if p.debug {
					fmt.Printf("DEBUG parseTypedParameter: created default value node\n")
				}
			}
		}
	}

	if paramName != "" {
		pos := ast.Position{
			Line:   paramNode.Start().Line,
			Column: paramNode.Start().Column,
		}
		// Create a Parameter struct directly since there's no NewParameter function
		return &ast.Parameter{
			Name:       paramName,
			TypeExpr:   paramType,
			Default:    defaultValue,
			IsOptional: nodeType == "optional_parameter",
			Pos:        pos,
		}
	}

	return nil
}

// parseReturnType parses a method_return_type node
func (p *Parser) parseReturnType(returnTypeNode Node) *ast.TypeExpression {
	// method_return_type contains a type_expression child
	for _, child := range returnTypeNode.Children() {
		if child.Type() == "type_expression" {
			return p.parseTreeSitterTypeExpression(child)
		}
	}
	return nil
}

// parseTreeSitterTypeExpression parses a type_expression node from tree-sitter
func (p *Parser) parseTreeSitterTypeExpression(typeExprNode Node) *ast.TypeExpression {
	// type_expression contains an identifier child with the type name
	for _, child := range typeExprNode.Children() {
		if child.Type() == "identifier" {
			typeName := child.Text()
			startPos := ast.Position{
				Line:   typeExprNode.Start().Line,
				Column: typeExprNode.Start().Column,
			}
			endPos := ast.Position{
				Line:   typeExprNode.End().Line,
				Column: typeExprNode.End().Column,
			}
			// NewTypeExpression signature: (name string, params []*TypeExpression, start, end Position)
			return ast.NewTypeExpression(typeName, nil, startPos, endPos)
		}
	}

	// Fallback: use the text of the type_expression node itself
	typeName := typeExprNode.Text()
	startPos := ast.Position{
		Line:   typeExprNode.Start().Line,
		Column: typeExprNode.Start().Column,
	}
	endPos := ast.Position{
		Line:   typeExprNode.End().Line,
		Column: typeExprNode.End().Column,
	}
	return ast.NewTypeExpression(typeName, nil, startPos, endPos)
}

// parseBlock parses a block using the new encapsulated pattern
func (p *Parser) parseBlock(blockNode Node, start, end ast.Position) *ast.BlockStmt {
	// Create empty block statement
	block := ast.NewBlockStmt([]ast.StatementNode{}, start, end)

	// Process each child node in the block
	for _, child := range blockNode.Children() {
		childType := child.Type()

		// Skip structural tokens that shouldn't be statements
		if childType == "{" || childType == "}" || childType == ";" {
			// Create token node for structural elements
			tokenNode := p.convertToTokenNode(child)
			if tokenNode != nil {
				block.AddChild(tokenNode)
			}
			continue
		}

		// Convert actual statements
		if stmt := p.convertToStatement(child); stmt != nil {
			block.AddChild(stmt)
		} else {
			// For non-statement nodes, try to convert to token
			if tokenNode := p.convertToTokenNode(child); tokenNode != nil {
				block.AddChild(tokenNode)
			}
		}
	}

	return block
}

// parseBlockStatements parses the statements inside a block (deprecated - use parseBlock)
func (p *Parser) parseBlockStatements(blockNode Node) []ast.StatementNode {
	var statements []ast.StatementNode

	// Process each child node in the block
	for _, child := range blockNode.Children() {
		childType := child.Type()

		// Skip structural tokens that shouldn't be statements
		if childType == "{" || childType == "}" || childType == ";" {
			continue
		}

		// Convert the child node to an AST statement
		if stmt := p.convertToStatement(child); stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// convertToStatement converts a tree-sitter node to an AST statement
func (p *Parser) convertToStatement(node Node) ast.StatementNode {
	if node == nil {
		return nil
	}

	nodeType := node.Type()
	start := ast.Position{
		Line:   node.Start().Line,
		Column: node.Start().Column,
		Offset: node.Start().Offset,
	}
	end := ast.Position{
		Line:   node.End().Line,
		Column: node.End().Column,
		Offset: node.End().Offset,
	}

	// Handle different statement types
	fmt.Printf("DEBUG: convertToStatement called with nodeType: %s, text: %q\n", nodeType, node.Text())
	switch nodeType {
	case "assignment_expression":
		// Handle variable declarations with assignments (my $var = value)
		nodeText := node.Text()
		if strings.HasPrefix(strings.TrimSpace(nodeText), "my ") ||
			strings.HasPrefix(strings.TrimSpace(nodeText), "our ") ||
			strings.HasPrefix(strings.TrimSpace(nodeText), "state ") {
			fmt.Printf("DEBUG: convertToStatement processing assignment_expression as VarDecl: %q\n", nodeText)
			if result := p.convertAssignmentToVarDeclAST(node, start, end); result != nil {
				if stmt, ok := result.(ast.StatementNode); ok {
					return stmt
				}
			}
		}
		// Fall back to expression statement
		return p.parseExpressionStatement(node, start, end)
	case "expression_statement":
		// Check if this expression statement contains a variable declaration
		nodeText := node.Text()
		if strings.HasPrefix(strings.TrimSpace(nodeText), "my ") ||
			strings.HasPrefix(strings.TrimSpace(nodeText), "our ") ||
			strings.HasPrefix(strings.TrimSpace(nodeText), "state ") {
			fmt.Printf("DEBUG: convertToStatement found variable declaration in expression_statement: %q\n", nodeText)
			// Look for assignment_expression child
			for _, child := range node.Children() {
				if child.Type() == "assignment_expression" {
					fmt.Printf("DEBUG: convertToStatement processing assignment_expression child\n")
					if result := p.convertAssignmentToVarDeclAST(child, start, end); result != nil {
						if stmt, ok := result.(ast.StatementNode); ok {
							return stmt
						}
					}
				}
			}
		}
		// Parse expression statements (e.g., "return $a + $b;")
		return p.parseExpressionStatement(node, start, end)
	case "variable_declaration":
		// Skip standalone variable declarations - they should be part of assignment_expression
		fmt.Printf("DEBUG: convertToStatement skipping variable_declaration: %q\n", node.Text())
		return nil
	case "typed_variable_declaration":
		// Parse typed variable declarations
		return p.parseTypedVariableDeclaration(node, start, end)
	case "return_expression":
		// Parse return statements
		return p.parseReturnStatement(node, start, end)
	case "class_statement":
		// Parse class declarations
		return p.convertToClassDeclAST(node, start, end).(ast.StatementNode)
	case "role_statement":
		// Parse role declarations
		return p.convertToRoleDeclAST(node, start, end).(ast.StatementNode)
	case "type_alias_statement":
		// Parse type alias statements
		return p.parseTypeAliasStatement(node, start, end)
	case "conditional_statement":
		// Parse conditional statements (if/elsif/else)
		return p.parseConditionalStatement(node, start, end)
	case "for_statement", "foreach_statement", "while_statement":
		// Parse loop statements
		return p.parseLoopStatement(node, start, end)
	default:
		// For other statement types, create a generic expression statement
		// using the node's text content
		nodeText := node.Text()
		if nodeText != "" {
			expr := ast.NewLiteralExpr(nodeText, ast.StringLiteral, start, end)
			return ast.NewExpressionStmt(expr, start, end)
		}
	}

	return nil
}

// parseExpressionStatement parses an expression statement
func (p *Parser) parseExpressionStatement(node Node, start, end ast.Position) ast.StatementNode {
	// Parse the expression statement by examining its children
	// This replaces the placeholder literal implementation

	if p.debug {
		fmt.Printf("DEBUG parseExpressionStatement: Processing node with %d children\n", len(node.Children()))
	}

	// Look for the actual expression within the expression statement
	for _, child := range node.Children() {
		childType := child.Type()

		if p.debug {
			fmt.Printf("DEBUG parseExpressionStatement: Child type: %s\n", childType)
		}

		switch childType {
		case "variable_declaration", "typed_variable_declaration":
			// This is a variable declaration like "my $var = value"
			return p.parseVariableDeclaration(child, start, end)

		case "assignment_expression":
			// This is an assignment expression like "my $var = $hash->{key}"
			// Need to handle this as a complete assignment
			return p.parseAssignmentExpression(child, start, end)

		case "hash_ref", "hash_access", "deref_expression", "hash_element_expression":
			// This is a hash access expression
			expr := p.parseSimpleExpression(child)
			if expr != nil {
				return ast.NewExpressionStmt(expr, start, end)
			}

		case "function_call", "call_expression":
			// This is a function call
			expr := p.parseSimpleExpression(child)
			if expr != nil {
				return ast.NewExpressionStmt(expr, start, end)
			}

		case "return_expression", "return_statement":
			// This is a return statement
			return p.parseReturnStatement(child, start, end)

		default:
			// Try to parse as a general expression
			if expr := p.parseSimpleExpression(child); expr != nil {
				return ast.NewExpressionStmt(expr, start, end)
			}
		}
	}

	// If we couldn't parse any specific construct, fall back to parsing the whole thing as an expression
	if expr := p.parseSimpleExpression(node); expr != nil {
		return ast.NewExpressionStmt(expr, start, end)
	}

	// Final fallback: create a literal (but log this as it indicates a parsing gap)
	nodeText := strings.TrimSpace(node.Text())
	if nodeText != "" {
		expr := ast.NewLiteralExpr(nodeText, ast.StringLiteral, start, end)
		return ast.NewExpressionStmt(expr, start, end)
	}

	return nil
}

// parseAssignmentExpression parses an assignment expression node
func (p *Parser) parseAssignmentExpression(node Node, start, end ast.Position) ast.StatementNode {
	if node == nil {
		return nil
	}

	var varDeclNode Node
	var initializerNode Node

	// Parse assignment expression structure: left = right
	children := node.Children()
	for _, child := range children {
		childType := child.Type()

		if childType == "variable_declaration" || childType == "typed_variable_declaration" {
			varDeclNode = child
		} else if childType == "hash_element_expression" || childType == "scalar" ||
			childType == "number" || childType == "string_literal" ||
			childType == "array_element_expression" || childType == "method_call_expression" ||
			childType == "function_call_expression" || childType == "binary_expression" {
			// This is the right side of the assignment (initializer)
			initializerNode = child
		}
	}

	// If we found a variable declaration, parse it with the initializer
	if varDeclNode != nil {
		stmt := p.parseVariableDeclarationWithInitializer(varDeclNode, initializerNode, start, end)
		if stmt != nil {
			return stmt
		}
	}

	// If no variable_declaration found, fall back to expression statement
	if expr := p.parseSimpleExpression(node); expr != nil {
		return ast.NewExpressionStmt(expr, start, end)
	}

	// Final fallback
	nodeText := strings.TrimSpace(node.Text())
	if nodeText != "" {
		expr := ast.NewLiteralExpr(nodeText, ast.StringLiteral, start, end)
		return ast.NewExpressionStmt(expr, start, end)
	}

	return nil
}

// parseVariableDeclarationWithInitializer parses a variable declaration with an optional initializer
func (p *Parser) parseVariableDeclarationWithInitializer(varDeclNode Node, initializerNode Node, start, end ast.Position) ast.StatementNode {
	if varDeclNode == nil {
		return nil
	}

	var declType string
	var variables []*ast.VariableExpr
	var typeExpr *ast.TypeExpression
	var initializer ast.ExpressionNode

	// Parse the variable declaration node first
	for _, child := range varDeclNode.Children() {
		childType := child.Type()
		childText := strings.TrimSpace(child.Text())

		switch childType {
		case "my", "state", "our", "field", "local":
			declType = childText
		case "type_expression":
			pos := Position{Line: start.Line, Column: start.Column, Offset: start.Offset}
			if parsedType := p.parseTypeExpression(childText, pos); parsedType != nil {
				typeExpr = &ast.TypeExpression{
					BaseNode: ast.NewBaseNode("type_expression", start, end),
					Name:     parsedType.BaseType,
					Kind:     ast.SimpleTypeKind, // Default to simple type for now
				}
			}
		case "scalar", "array", "hash":
			varName := p.extractVariableName(child)
			if varName != "" {
				variables = append(variables, ast.NewVariableExpr("", varName, start, end))
			}
		}
	}

	// Parse the initializer if provided
	if initializerNode != nil {
		initializerStart := ast.Position{
			Line:   initializerNode.Start().Line,
			Column: initializerNode.Start().Column,
			Offset: initializerNode.Start().Offset,
		}
		initializerEnd := ast.Position{
			Line:   initializerNode.End().Line,
			Column: initializerNode.End().Column,
			Offset: initializerNode.End().Offset,
		}

		// Parse the initializer expression
		initializer = p.parseInitializerExpression(initializerNode, initializerStart, initializerEnd)
	}

	// Create VarDecl with the initializer
	if len(variables) > 0 {
		varDecl := ast.NewVarDecl(declType, variables, typeExpr, initializer, start, end)
		return varDecl
	}

	return nil
}

// parseInitializerExpression parses an expression used as a variable initializer
func (p *Parser) parseInitializerExpression(node Node, start, end ast.Position) ast.ExpressionNode {
	if node == nil {
		return nil
	}

	nodeType := node.Type()
	nodeText := strings.TrimSpace(node.Text())

	switch nodeType {
	case "hash_element_expression":
		// This is critical for safety analysis: $input->{name}
		return ast.NewLiteralExpr(nodeText, ast.HashAccessLiteral, start, end)
	case "array_element_expression":
		// Array access: $data->[0]
		return ast.NewLiteralExpr(nodeText, ast.ArrayAccessLiteral, start, end)
	case "method_call_expression":
		// Method call: $obj->method()
		return ast.NewLiteralExpr(nodeText, ast.MethodCallLiteral, start, end)
	case "function_call_expression":
		// Function call: length($string)
		return ast.NewLiteralExpr(nodeText, ast.FunctionCallLiteral, start, end)
	case "func1op_call_expression":
		// Single-operand function call: ref($data)
		return ast.NewLiteralExpr(nodeText, ast.FunctionCallLiteral, start, end)
	case "scalar":
		// Variable reference: $variable
		varName := strings.TrimPrefix(nodeText, "$")
		return ast.NewVariableExpr("", varName, start, end)
	case "number", "integer":
		return ast.NewLiteralExpr(nodeText, ast.NumberLiteral, start, end)
	case "string_literal", "string":
		return ast.NewLiteralExpr(nodeText, ast.StringLiteral, start, end)
	case "binary_expression":
		// Complex expressions like $config->{timeout} // 30
		return ast.NewLiteralExpr(nodeText, ast.BinaryExpressionLiteral, start, end)
	default:
		// Fallback: treat as generic expression
		return ast.NewLiteralExpr(nodeText, ast.ExpressionLiteral, start, end)
	}
}

// parseSimpleExpression parses a simple expression node (used for default values)
func (p *Parser) parseSimpleExpression(node Node) ast.ExpressionNode {
	if node == nil {
		return nil
	}

	nodeType := node.Type()
	nodeText := node.Text()
	start := ast.Position{
		Line:   node.Start().Line,
		Column: node.Start().Column,
		Offset: node.Start().Offset,
	}
	end := ast.Position{
		Line:   node.End().Line,
		Column: node.End().Column,
		Offset: node.End().Offset,
	}

	// Handle different expression types
	switch nodeType {
	case "number", "integer":
		return ast.NewLiteralExpr(nodeText, ast.NumberLiteral, start, end)
	case "string", "string_literal":
		return ast.NewLiteralExpr(nodeText, ast.StringLiteral, start, end)
	case "scalar":
		// Variable reference
		varName := strings.TrimPrefix(nodeText, "$")
		return ast.NewVariableExpr("", varName, start, end)
	default:
		// For other types, treat as literal
		return ast.NewLiteralExpr(nodeText, ast.StringLiteral, start, end)
	}
}

// parseVariableDeclaration parses a variable declaration statement
func (p *Parser) parseVariableDeclaration(node Node, start, end ast.Position) ast.StatementNode {
	var declType string
	var variables []*ast.VariableExpr
	var typeExpr *ast.TypeExpression
	var initializer ast.ExpressionNode

	// DEBUG: Log children of variable declaration node
	fmt.Printf("DEBUG: parseVariableDeclaration - node has %d children\n", len(node.Children()))
	for i, child := range node.Children() {
		fmt.Printf("DEBUG: Child %d: type=%s, text=%q\n", i, child.Type(), child.Text())
	}

	// Parse children to extract components
	for _, child := range node.Children() {
		childType := child.Type()

		switch childType {
		case "my", "our", "state", "field", "local":
			declType = child.Text()
		case "scalar", "variable":
			// This is a variable like $name
			varText := child.Text()
			if len(varText) > 1 && varText[0] == '$' {
				name := varText[1:]
				sigil := "$"
				varExpr := ast.NewVariableExpr(name, sigil, start, end)
				variables = append(variables, varExpr)
			}
		case "array":
			// This is an array variable like @name
			varText := child.Text()
			if len(varText) > 1 && varText[0] == '@' {
				name := varText[1:]
				sigil := "@"
				varExpr := ast.NewVariableExpr(name, sigil, start, end)
				variables = append(variables, varExpr)
			}
		case "type_expression":
			// Parse type annotation
			typeStr := child.Text()
			typeExpr = p.convertToASTTypeExpression(p.parseTypeExpression(typeStr, Position{Line: start.Line, Column: start.Column}))
		case "assignment_expression":
			// This is the initializer part
			initializer = p.parseSimpleExpression(child)
		case "=":
			// Assignment operator - check next sibling for the initializer
			continue
		default:
			// Check if this could be an initializer expression
			if initializer == nil {
				// Try to parse as expression for things like hash access
				if expr := p.parseSimpleExpression(child); expr != nil {
					initializer = expr
				}
			}
		}
	}

	// Always check text for assignment even if children parsing succeeded
	// because tree-sitter may not include assignment as child nodes
	if initializer == nil {
		nodeText := strings.TrimSpace(node.Text())
		fmt.Printf("DEBUG: Fallback text parsing with nodeText: %q\n", nodeText)
		if nodeText != "" {
			// Check for assignment in the text
			if strings.Contains(nodeText, "=") {
				// Split on = to separate declaration from initializer
				parts := strings.SplitN(nodeText, "=", 2)
				if len(parts) == 2 {
					declPart := strings.TrimSpace(parts[0])
					initPart := strings.TrimSpace(parts[1])

					// Parse declaration part only if we don't already have it
					if declType == "" || len(variables) == 0 {
						declFields := strings.Fields(declPart)
						if len(declFields) >= 2 {
							declType = declFields[0]
							varName := declFields[1]

							if len(varName) > 1 && varName[0] == '$' {
								name := varName[1:]
								sigil := "$"
								varExpr := ast.NewVariableExpr(name, sigil, start, end)
								variables = append(variables, varExpr)
							}
						}
					}

					// Create a simple literal expression for the initializer
					if initPart != "" {
						initializer = ast.NewLiteralExpr(initPart, ast.StringLiteral, start, end)
					}
				}
			} else {
				// No assignment, just parse as simple declaration
				parts := strings.Fields(nodeText)
				if len(parts) >= 2 {
					declType = parts[0]
					varName := parts[1]

					if len(varName) > 1 && varName[0] == '$' {
						name := varName[1:]
						sigil := "$"
						varExpr := ast.NewVariableExpr(name, sigil, start, end)
						variables = append(variables, varExpr)
					}
				}
			}
		}
	}

	// Create the variable declaration with proper initializer
	if declType != "" && len(variables) > 0 {
		return ast.NewVarDecl(declType, variables, typeExpr, initializer, start, end)
	}

	return nil
}

// parseTypedVariableDeclaration parses a typed variable declaration statement
func (p *Parser) parseTypedVariableDeclaration(node Node, start, end ast.Position) ast.StatementNode {
	var declType string
	var typeExpr *ast.TypeExpression
	var variables []*ast.VariableExpr

	// Process children to extract declaration type, type expression, and variables
	for _, child := range node.Children() {
		switch child.Type() {
		case "my", "our", "state", "local", "field":
			declType = child.Text()
		case "type_expression":
			typeExpr = p.extractTypeExpression(child)
		case "scalar", "array", "hash":
			varName := p.extractVariableName(child)
			if varName != "" {
				sigil := child.Text()[:1] // Extract sigil ($, @, %)
				varExpr := ast.NewVariableExpr(varName, sigil,
					ast.Position{Line: child.Start().Line, Column: child.Start().Column},
					ast.Position{Line: child.End().Line, Column: child.End().Column})
				variables = append(variables, varExpr)
			}
		}
	}

	// If no variables found, return nil
	if len(variables) == 0 {
		return nil
	}

	return ast.NewVarDecl(declType, variables, typeExpr, nil, start, end)
}

// parseReturnStatement parses a return statement
func (p *Parser) parseReturnStatement(node Node, start, end ast.Position) ast.StatementNode {
	// Parse return statement properly by examining its children
	var returnExpr ast.ExpressionNode

	for _, child := range node.Children() {
		childType := child.Type()

		// Skip the "return" keyword itself
		if childType == "return" {
			continue
		}

		// Parse the return value expression
		if expr := p.parseSimpleExpression(child); expr != nil {
			returnExpr = expr
			break
		}
	}

	// Create a proper return statement
	return ast.NewReturnStmt(returnExpr, start, end)
}

// parseConditionalStatement parses conditional statements (if/elsif/else)
func (p *Parser) parseConditionalStatement(node Node, start, end ast.Position) ast.StatementNode {
	// For now, create a simple AST node for conditional statements
	// This will need more sophisticated parsing in the future
	return ast.NewIfStmt(nil, nil, nil, start, end)
}

// parseLoopStatement parses loop statements (for/foreach/while)
func (p *Parser) parseLoopStatement(node Node, start, end ast.Position) ast.StatementNode {
	// For now, create a simple AST node for loop statements
	// This will need more sophisticated parsing in the future
	return ast.NewForStmt(nil, nil, nil, start, end)
}

// parseTypeAliasStatement parses a type alias statement (type UserID = Int)
func (p *Parser) parseTypeAliasStatement(node Node, start, end ast.Position) ast.StatementNode {
	// Simple text-based parsing for now, following the pattern of other parser functions
	nodeText := strings.TrimSpace(node.Text())
	if nodeText == "" {
		return nil
	}

	// Parse "type Name = Definition;" or "type Name<T> = Definition;"
	// Remove semicolon if present
	nodeText = strings.TrimSuffix(nodeText, ";")

	// Find the equal sign to split name and definition
	equalPos := strings.Index(nodeText, "=")
	if equalPos == -1 {
		return nil
	}

	// Extract left side (type name and parameters) and right side (definition)
	leftPart := strings.TrimSpace(nodeText[4:equalPos]) // Skip "type "
	rightPart := strings.TrimSpace(nodeText[equalPos+1:])

	if leftPart == "" || rightPart == "" {
		return nil
	}

	// Parse type name and optional parameters from left part
	var typeName string
	var typeParams []*ast.TypeParameter

	if strings.Contains(leftPart, "<") && strings.Contains(leftPart, ">") {
		// Has type parameters: "Name<T, U>"
		paramStart := strings.Index(leftPart, "<")
		paramEnd := strings.LastIndex(leftPart, ">")
		if paramStart > 0 && paramEnd > paramStart {
			typeName = strings.TrimSpace(leftPart[:paramStart])
			paramText := strings.TrimSpace(leftPart[paramStart+1 : paramEnd])

			// Simple parameter parsing - split by comma
			if paramText != "" {
				paramNames := strings.Split(paramText, ",")
				for _, paramName := range paramNames {
					paramName = strings.TrimSpace(paramName)
					if paramName != "" {
						param := &ast.TypeParameter{
							Name:     paramName,
							Position: ast.Position{Line: start.Line, Column: start.Column},
						}
						typeParams = append(typeParams, param)
					}
				}
			}
		}
	} else {
		// No type parameters: just the name
		typeName = leftPart
	}

	// Parse the type definition
	typeExpr := p.parseTypeExpression(rightPart, Position{Line: start.Line, Column: start.Column})
	definition := p.convertToASTTypeExpression(typeExpr)

	if typeName != "" && definition != nil {
		return ast.NewTypeAliasStatement(typeName, typeParams, definition, start, end)
	}

	return nil
}

// processClassBlock processes the contents of a class block
func (p *Parser) processClassBlock(blockNode Node, classDecl *ast.ClassDecl) {
	for _, stmt := range blockNode.Children() {
		switch stmt.Type() {
		case "expression_statement":
			// Check if the expression statement contains a variable declaration
			for _, child := range stmt.Children() {
				if child.Type() == "variable_declaration" && strings.Contains(child.Text(), "field ") {
					if field := p.extractFieldDecl(child); field != nil {
						classDecl.AddField(field)
					}
				}
			}
		case "variable_declaration":
			// Direct variable declaration
			if strings.Contains(stmt.Text(), "field ") {
				if field := p.extractFieldDecl(stmt); field != nil {
					classDecl.AddField(field)
				}
			}
		case "method_declaration_statement":
			// Method declaration
			if method := p.extractMethodDecl(stmt); method != nil {
				classDecl.AddMethod(method)
			}
		}
	}
}

// processRoleBlock processes the contents of a role block
func (p *Parser) processRoleBlock(blockNode Node, roleDecl *ast.RoleDecl) {
	for _, stmt := range blockNode.Children() {
		switch stmt.Type() {
		case "expression_statement":
			// Check if the expression statement contains a typed variable declaration
			for _, child := range stmt.Children() {
				if child.Type() == "typed_variable_declaration" && strings.Contains(child.Text(), "field ") {
					if field := p.extractFieldDecl(child); field != nil {
						roleDecl.AddField(field)
					}
				}
			}
		case "typed_variable_declaration":
			// Direct typed variable declaration
			if strings.Contains(stmt.Text(), "field ") {
				if field := p.extractFieldDecl(stmt); field != nil {
					roleDecl.AddField(field)
				}
			}
		case "method_declaration_statement":
			// Method declaration
			if method := p.extractMethodDecl(stmt); method != nil {
				roleDecl.AddProvidedMethod(method)
			}
		}
	}
}

// extractFieldDecl extracts a field declaration from a statement node
func (p *Parser) extractFieldDecl(stmt Node) *ast.FieldDecl {
	start := ast.Position{Line: stmt.Start().Line, Column: stmt.Start().Column}
	end := ast.Position{Line: stmt.End().Line, Column: stmt.End().Column}

	var fieldName string
	var typeExpr *ast.TypeExpression
	var varExpr ast.ExpressionNode

	// Handle typed_variable_declaration nodes specifically
	if stmt.Type() == "typed_variable_declaration" {
		for _, child := range stmt.Children() {
			switch child.Type() {
			case "type_expression":
				// Extract type information
				typeExpr = p.extractTypeExpression(child)
			case "scalar":
				// Extract variable name from scalar node
				fieldName = p.extractVariableName(child)
				if fieldName != "" {
					varExpr = ast.NewVariableExpr(fieldName, "$", start, end)
				}
			}
		}
	} else if strings.Contains(stmt.Text(), "field ") {
		// Fallback to text parsing for non-structured nodes
		text := stmt.Text()
		parts := strings.Fields(text)
		if len(parts) >= 3 {
			fieldName = strings.TrimPrefix(parts[2], "$")
			varExpr = ast.NewVariableExpr(fieldName, "$", start, end)
		}
	}

	if fieldName == "" {
		return nil
	}

	// Type assert varExpr to *VariableExpr as required by NewFieldDecl
	var varPtr *ast.VariableExpr
	if ve, ok := varExpr.(*ast.VariableExpr); ok {
		varPtr = ve
	}
	return ast.NewFieldDecl(fieldName, typeExpr, varPtr, nil, start, end)
}

// extractTypeExpression extracts type information from a type_expression node
func (p *Parser) extractTypeExpression(node Node) *ast.TypeExpression {
	if node == nil {
		return nil
	}

	// For now, extract the base type from the text
	// This could be enhanced to parse complex types like ArrayRef[Int]
	typeText := strings.TrimSpace(node.Text())
	if typeText == "" {
		return nil
	}

	start := ast.Position{Line: node.Start().Line, Column: node.Start().Column}
	end := ast.Position{Line: node.End().Line, Column: node.End().Column}

	// Handle parameterized types like ArrayRef[Int] or Optional[Email]
	if strings.Contains(typeText, "[") && strings.Contains(typeText, "]") {
		openBracket := strings.Index(typeText, "[")
		closeBracket := strings.LastIndex(typeText, "]")

		if openBracket < closeBracket {
			paramText := strings.TrimSpace(typeText[openBracket+1 : closeBracket])

			// Create parameterized type expression
			var params []*ast.TypeExpression
			if paramText != "" {
				paramType := ast.NewTypeExpression(paramText, nil, start, end)
				params = []*ast.TypeExpression{paramType}
			}
			// Use full typeText as name (e.g., "ArrayRef[MyClass]")
			return ast.NewTypeExpression(typeText, params, start, end)
		}
	}

	// Simple type
	return ast.NewTypeExpression(typeText, nil, start, end)
}

// extractVariableName extracts variable name from a scalar node
func (p *Parser) extractVariableName(node Node) string {
	if node == nil {
		return ""
	}

	// Extract variable name from scalar/array/hash node children
	for _, child := range node.Children() {
		if child.Type() != "$" && child.Type() != "@" && child.Type() != "%" && child.Type() != "token" {
			text := strings.TrimSpace(child.Text())
			if text != "$" && text != "@" && text != "%" && text != "" {
				return text
			}
		}
	}

	// Fallback - extract from full text, handling all sigils
	text := strings.TrimSpace(node.Text())
	// Remove common Perl sigils
	text = strings.TrimPrefix(text, "$")
	text = strings.TrimPrefix(text, "@")
	text = strings.TrimPrefix(text, "%")
	return text
}

// extractMethodParameters extracts parameters from a method signature node
func (p *Parser) extractMethodParameters(sigNode Node) []*ast.Parameter {
	var parameters []*ast.Parameter

	if sigNode == nil {
		return parameters
	}

	// Parse parameters from signature node
	for _, child := range sigNode.Children() {
		if child.Type() == "mandatory_parameter" {
			param := p.extractParameter(child)
			if param != nil {
				parameters = append(parameters, param)
			}
		}
	}

	return parameters
}

// extractParameter extracts a single parameter from a parameter node
func (p *Parser) extractParameter(paramNode Node) *ast.Parameter {
	if paramNode == nil {
		return nil
	}

	start := ast.Position{Line: paramNode.Start().Line, Column: paramNode.Start().Column}
	end := ast.Position{Line: paramNode.End().Line, Column: paramNode.End().Column}

	var paramName string
	var typeExpr *ast.TypeExpression

	// Extract type and name from parameter children
	for _, child := range paramNode.Children() {
		switch child.Type() {
		case "type_expression":
			typeExpr = p.extractTypeExpression(child)
		case "scalar":
			paramName = p.extractVariableName(child)
		}
	}

	if paramName == "" {
		return nil
	}

	// Create Parameter struct directly since there's no constructor
	param := &ast.Parameter{
		Name:       paramName,
		TypeExpr:   typeExpr,
		Variable:   ast.NewVariableExpr(paramName, "$", start, end),
		Default:    nil,
		IsOptional: false,
	}
	return param
}

// extractConstraintsFromText extracts constraint information from method or class signatures
func (p *Parser) extractConstraintsFromText(text string) []*ast.TypeConstraint {
	var constraints []*ast.TypeConstraint

	// Look for "where" clause
	whereIndex := strings.Index(text, " where ")
	if whereIndex == -1 {
		return constraints
	}

	// Extract constraint text after "where"
	constraintText := strings.TrimSpace(text[whereIndex+7:])

	// Find the end of constraints (before opening brace)
	braceIndex := strings.Index(constraintText, "{")
	if braceIndex != -1 {
		constraintText = strings.TrimSpace(constraintText[:braceIndex])
	}

	if constraintText == "" {
		return constraints
	}

	// Parse individual constraints separated by commas
	constraintParts := strings.Split(constraintText, ",")
	for _, part := range constraintParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		constraint := p.parseConstraintText(part)
		if constraint != nil {
			constraints = append(constraints, constraint)
		}
	}

	return constraints
}

// parseConstraintText parses a single constraint expression
func (p *Parser) parseConstraintText(text string) *ast.TypeConstraint {
	text = strings.TrimSpace(text)

	// Type constraint: T: SomeType
	if colonIndex := strings.Index(text, ":"); colonIndex != -1 {
		parameter := strings.TrimSpace(text[:colonIndex])
		typeExpr := strings.TrimSpace(text[colonIndex+1:])

		// Create a simple expression node for the constraint
		expr := ast.NewLiteralExpr(typeExpr, ast.StringLiteral, ast.Position{}, ast.Position{})

		return &ast.TypeConstraint{
			Parameter:  parameter,
			Kind:       ast.TypeConstraintKind,
			Expression: expr,
			Position:   ast.Position{},
		}
	}

	// Protocol constraint: T does Role
	if doesIndex := strings.Index(text, " does "); doesIndex != -1 {
		parameter := strings.TrimSpace(text[:doesIndex])
		roleExpr := strings.TrimSpace(text[doesIndex+6:])

		expr := ast.NewLiteralExpr(roleExpr, ast.StringLiteral, ast.Position{}, ast.Position{})

		return &ast.TypeConstraint{
			Parameter:  parameter,
			Kind:       ast.ProtocolConstraint,
			Expression: expr,
			Position:   ast.Position{},
		}
	}

	// Capability constraint: T can 'method'
	if canIndex := strings.Index(text, " can "); canIndex != -1 {
		parameter := strings.TrimSpace(text[:canIndex])
		methodExpr := strings.TrimSpace(text[canIndex+5:])

		expr := ast.NewLiteralExpr(methodExpr, ast.StringLiteral, ast.Position{}, ast.Position{})

		return &ast.TypeConstraint{
			Parameter:  parameter,
			Kind:       ast.CapabilityConstraint,
			Expression: expr,
			Position:   ast.Position{},
		}
	}

	// Value constraint: $param > 0 or version constraint: T->VERSION >= 1.0
	if strings.Contains(text, ">") || strings.Contains(text, "<") || strings.Contains(text, "=") || strings.Contains(text, "->VERSION") {
		// For now, treat as value constraint - more sophisticated parsing could be added
		parameter := extractParameterFromValueConstraint(text)

		expr := ast.NewLiteralExpr(text, ast.StringLiteral, ast.Position{}, ast.Position{})

		kind := ast.ValueConstraint
		if strings.Contains(text, "->VERSION") {
			kind = ast.VersionConstraint
		}

		return &ast.TypeConstraint{
			Parameter:  parameter,
			Kind:       kind,
			Expression: expr,
			Position:   ast.Position{},
		}
	}

	return nil
}

// extractConstraintAnnotations extracts constraint-based type annotations from source text
func (p *Parser) extractConstraintAnnotations(content string) []*ast.TypeAnnotation {
	var annotations []*ast.TypeAnnotation

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// ONLY look for method signatures with constraints (where clauses)
		// Basic type annotations are already handled by the tree-sitter parser
		if strings.Contains(line, "method ") && strings.Contains(line, " where ") {
			annotations = append(annotations, p.extractMethodConstraintAnnotations(line, lineNum+1)...)
		}

		// Note: Field declarations are now handled by regular variable parsing logic
		// No special field annotation processing needed - field is just a scope keyword like my

		// Look for class declarations with constraints
		if strings.Contains(line, "class ") && strings.Contains(line, " where ") {
			annotations = append(annotations, p.extractClassConstraintAnnotations(line, lineNum+1)...)
		}

		// Look for role declarations with constraints
		if strings.Contains(line, "role ") && strings.Contains(line, " where ") {
			annotations = append(annotations, p.extractRoleConstraintAnnotations(line, lineNum+1)...)
		}
	}

	return annotations
}

// extractMethodConstraintAnnotations extracts constraint annotations from method signatures
func (p *Parser) extractMethodConstraintAnnotations(line string, lineNum int) []*ast.TypeAnnotation {
	var annotations []*ast.TypeAnnotation

	// Extract method name
	methodNameMatch := strings.Index(line, "method ")
	if methodNameMatch == -1 {
		return annotations
	}

	// Find parameter list
	parenStart := strings.Index(line, "(")
	parenEnd := strings.LastIndex(line, ")")
	if parenStart == -1 || parenEnd == -1 {
		return annotations
	}

	paramList := line[parenStart+1 : parenEnd]

	// Extract parameter type annotations
	params := strings.Split(paramList, ",")
	for _, param := range params {
		param = strings.TrimSpace(param)
		if strings.Contains(param, " $") {
			// Extract parameter type and name
			parts := strings.Fields(param)
			if len(parts) >= 2 {
				typeExpr := parts[0]
				varName := parts[len(parts)-1]

				// Create type expression
				typeExpression := &ast.TypeExpression{
					Name: typeExpr,
					Kind: ast.SimpleTypeKind,
				}

				annotation := &ast.TypeAnnotation{
					AnnotatedItem:  varName,
					TypeExpression: typeExpression,
					Pos:            ast.Position{Line: lineNum, Column: 1},
					Kind:           ast.MethodParamAnnotation,
				}
				annotations = append(annotations, annotation)
			}
		}
	}

	// Extract return type annotation
	arrowIndex := strings.Index(line, "->")
	whereIndex := strings.Index(line, " where ")
	if arrowIndex != -1 {
		endIndex := len(line)
		if whereIndex != -1 && whereIndex > arrowIndex {
			endIndex = whereIndex
		}

		returnType := strings.TrimSpace(line[arrowIndex+2 : endIndex])
		if returnType != "" {
			typeExpression := &ast.TypeExpression{
				Name: returnType,
				Kind: ast.SimpleTypeKind,
			}

			// Extract method name for annotation
			methodParts := strings.Fields(line[methodNameMatch+7:])
			methodName := ""
			if len(methodParts) > 0 {
				methodName = strings.Split(methodParts[0], "(")[0]
			}

			annotation := &ast.TypeAnnotation{
				AnnotatedItem:  methodName,
				TypeExpression: typeExpression,
				Pos:            ast.Position{Line: lineNum, Column: 1},
				Kind:           ast.MethodReturnAnnotation,
			}
			annotations = append(annotations, annotation)
		}
	}

	return annotations
}

// extractFieldAnnotations extracts type annotations from field declarations
func (p *Parser) extractFieldAnnotations(line string, lineNum int) []*ast.TypeAnnotation {
	var annotations []*ast.TypeAnnotation

	// Look for field declarations: field Type $varname
	fieldIndex := strings.Index(line, "field ")
	if fieldIndex == -1 {
		return annotations
	}

	// Extract the part after "field "
	fieldPart := strings.TrimSpace(line[fieldIndex+6:])

	// Split on whitespace to get type and variable
	parts := strings.Fields(fieldPart)
	if len(parts) >= 2 {
		typeExpr := parts[0]
		varName := parts[1]

		// Remove semicolon from variable name if present
		varName = strings.TrimSuffix(varName, ";")

		// Create type expression
		typeExpression := &ast.TypeExpression{
			Name: typeExpr,
			Kind: ast.SimpleTypeKind,
		}

		annotation := &ast.TypeAnnotation{
			AnnotatedItem:  varName,
			TypeExpression: typeExpression,
			Pos:            ast.Position{Line: lineNum, Column: 1},
			Kind:           ast.FieldAnnotation,
		}
		annotations = append(annotations, annotation)
	}

	return annotations
}

// extractClassConstraintAnnotations extracts constraint annotations from class declarations
func (p *Parser) extractClassConstraintAnnotations(line string, lineNum int) []*ast.TypeAnnotation {
	var annotations []*ast.TypeAnnotation
	// For now, just create a basic annotation for the class
	// More sophisticated extraction could be added later
	return annotations
}

// extractRoleConstraintAnnotations extracts constraint annotations from role declarations
func (p *Parser) extractRoleConstraintAnnotations(line string, lineNum int) []*ast.TypeAnnotation {
	var annotations []*ast.TypeAnnotation
	// For now, just create a basic annotation for the role
	// More sophisticated extraction could be added later
	return annotations
}

// extractParameterFromValueConstraint extracts the parameter name from a value constraint
func extractParameterFromValueConstraint(text string) string {
	// Look for variable pattern ($param) or type pattern (T->)
	if strings.HasPrefix(text, "$") {
		spaceIndex := strings.Index(text, " ")
		if spaceIndex != -1 {
			return strings.TrimSpace(text[:spaceIndex])
		}
	}

	if arrowIndex := strings.Index(text, "->"); arrowIndex != -1 {
		return strings.TrimSpace(text[:arrowIndex])
	}

	// Default to first word
	parts := strings.Fields(text)
	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

func (p *Parser) extractMethodDecl(stmt Node) *ast.MethodDecl {
	start := ast.Position{Line: stmt.Start().Line, Column: stmt.Start().Column}
	end := ast.Position{Line: stmt.End().Line, Column: stmt.End().Column}

	var methodName string
	var returnType *ast.TypeExpression
	var parameters []*ast.Parameter
	var block *ast.BlockStmt

	// Handle method_declaration_statement nodes specifically
	if stmt.Type() == "method_declaration_statement" {
		// Extract method name from children
		for _, child := range stmt.Children() {
			if child.Type() == "bareword" {
				methodName = strings.TrimSpace(child.Text())
				break
			}
		}

		for _, child := range stmt.Children() {
			switch child.Type() {
			case "token":
				// Skip - already handled above
			case "signature":
				// Extract parameters from signature
				parameters = p.extractMethodParameters(child)
			case "type_expression":
				// Parse return type - tree-sitter presents it as a direct type_expression
				returnType = p.parseTreeSitterTypeExpression(child)
			case "return_type":
				// Extract return type from return_type node (legacy case)
				returnType = p.parseReturnType(child)
			case "block":
				// Convert block to BlockStmt
				if blockNode := p.convertToASTNode(child); blockNode != nil {
					if bs, ok := blockNode.(*ast.BlockStmt); ok {
						block = bs
					}
				}
			}
		}

		// Return type is now properly extracted from return_type node above
	} else if strings.Contains(stmt.Text(), "method ") {
		// Fallback to text parsing for non-structured nodes
		text := stmt.Text()
		lines := strings.Split(text, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimSpace(lines[0])
			parts := strings.Fields(firstLine)
			if len(parts) >= 2 && parts[0] == "method" {
				methodName = parts[1]
				if strings.Contains(methodName, "(") {
					methodName = strings.Split(methodName, "(")[0]
				}
			}
		}
	}

	if methodName == "" {
		return nil
	}

	// Extract constraints from the method signature
	constraints := p.extractConstraintsFromText(stmt.Text())

	// Create method declaration
	methodDecl := ast.NewMethodDecl(methodName, parameters, returnType, block, start, end)
	methodDecl.Constraints = constraints

	return methodDecl
}

// validateSyntax performs additional syntax validation on parsed content
func (p *Parser) validateSyntax(content string, typeAnnotations []*ast.TypeAnnotation) []error {
	var errors []error

	// Check for common malformed type patterns
	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for malformed type patterns
		if err := p.validateTypeSyntax(line, lineNum+1); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateTypeSyntax checks for specific malformed type syntax patterns
func (p *Parser) validateTypeSyntax(line string, lineNum int) error {
	// Conservative validation - catch malformed syntax patterns that should be errors

	trimmedLine := strings.TrimSpace(line)

	// Check for malformed assignment expressions that are clearly incomplete
	if strings.Contains(trimmedLine, "my ") && strings.HasSuffix(trimmedLine, "= ;") {
		return fmt.Errorf("IncompleteAssignmentError: line %d: incomplete assignment expression", lineNum)
	}

	// Check for missing closing bracket in parameterized types
	if strings.Contains(trimmedLine, "my ") && strings.Contains(trimmedLine, "ArrayRef[") &&
		!strings.Contains(trimmedLine, "]") && strings.Contains(trimmedLine, ";") {
		return fmt.Errorf("MissingClosingBracketError: line %d: missing closing bracket in parameterized type", lineNum)
	}

	// Check for double union operators in type declarations
	if strings.Contains(trimmedLine, "my ") && strings.Contains(trimmedLine, "||") &&
		!strings.Contains(trimmedLine, "=") && !strings.Contains(trimmedLine, "{") {
		return fmt.Errorf("InvalidUnionSyntaxError: line %d: invalid union syntax '||', use '|' for union types", lineNum)
	}

	// Check for double intersection operators in type declarations
	if strings.Contains(trimmedLine, "my ") && strings.Contains(trimmedLine, "&&") &&
		!strings.Contains(trimmedLine, "=") && !strings.Contains(trimmedLine, "{") && !strings.Contains(trimmedLine, ">=") {
		return fmt.Errorf("InvalidIntersectionSyntaxError: line %d: invalid intersection syntax '&&', use '&' for intersection types", lineNum)
	}

	// Check for double negation operators
	if strings.Contains(trimmedLine, "my ") && strings.Contains(trimmedLine, "!!") {
		return fmt.Errorf("InvalidNegationSyntaxError: line %d: invalid negation syntax '!!', use '!' for negation types", lineNum)
	}

	// Check for incomplete type assertions
	if strings.Contains(trimmedLine, " as ;") {
		return fmt.Errorf("IncompleteTypeAssertionError: line %d: incomplete type assertion, missing target type", lineNum)
	}

	// Check for specific malformed spacing pattern: space after [ but not before ]
	if strings.Contains(trimmedLine, "my ") &&
		strings.Contains(trimmedLine, "[ ") &&
		strings.Contains(trimmedLine, "]") &&
		!strings.Contains(trimmedLine, " ]") &&
		strings.Contains(trimmedLine, ";") {
		// Only flag the very specific pattern "[ Int]" (space after [ but not before ])
		return fmt.Errorf("InvalidParameterizedTypeError: line %d: invalid spacing in parameterized type", lineNum)
	}

	return nil
}

// convertToTokenNode converts a tree-sitter node to a token node
func (p *Parser) convertToTokenNode(node Node) ast.Node {
	if node == nil {
		return nil
	}

	start := ast.Position{
		Line:   node.Start().Line,
		Column: node.Start().Column,
		Offset: node.Start().Offset,
	}
	end := ast.Position{
		Line:   node.End().Line,
		Column: node.End().Column,
		Offset: node.End().Offset,
	}

	return p.convertToTokenNodeWithParams(node.Type(), node.Text(), start, end)
}

// convertToTokenNodeWithParams converts structural tokens to AST token nodes
func (p *Parser) convertToTokenNodeWithParams(nodeType, nodeText string, start, end ast.Position) ast.Node {
	// Map tree-sitter node types to our token types
	switch nodeType {
	case "{":
		return ast.NewTokenNode(ast.LeftBrace, nodeText, start, end)
	case "}":
		return ast.NewTokenNode(ast.RightBrace, nodeText, start, end)
	case "(":
		return ast.NewTokenNode(ast.LeftParen, nodeText, start, end)
	case ")":
		return ast.NewTokenNode(ast.RightParen, nodeText, start, end)
	case ";":
		return ast.NewTokenNode(ast.Semicolon, nodeText, start, end)
	case "->":
		return ast.NewTokenNode(ast.Arrow, nodeText, start, end)
	case "=":
		return ast.NewTokenNode(ast.Equals, nodeText, start, end)
	case "$":
		return ast.NewTokenNode(ast.Dollar, nodeText, start, end)
	case "sub":
		return ast.NewTokenNode(ast.SubKeyword, nodeText, start, end)
	case "method":
		return ast.NewTokenNode(ast.MethodKeyword, nodeText, start, end)
	case "my":
		return ast.NewTokenNode(ast.MyKeyword, nodeText, start, end)
	case "field":
		return ast.NewTokenNode(ast.FieldKeyword, nodeText, start, end)
	case "bareword", "varname", "identifier":
		return ast.NewTokenNode(ast.Identifier, nodeText, start, end)
	case "number":
		return ast.NewTokenNode(ast.Number, nodeText, start, end)
	case "string", "string_literal":
		return ast.NewTokenNode(ast.String, nodeText, start, end)
	default:
		// Check for whitespace content
		if strings.TrimSpace(nodeText) == "" && nodeText != "" {
			if strings.Contains(nodeText, "\n") {
				return ast.NewTokenNode(ast.Newline, nodeText, start, end)
			} else {
				return ast.NewTokenNode(ast.Whitespace, nodeText, start, end)
			}
		}
		return nil // Not a token we want to preserve
	}
}

// createContainerNode creates a container node that holds child tokens and preserves structure
func (p *Parser) createContainerNode(nodeType string, children []ast.Node, start, end ast.Position) ast.Node {
	// Create a BaseNode that can hold children and preserve the structure
	container := ast.NewBaseNode(nodeType, start, end)

	for _, child := range children {
		container.AddChild(child)
	}

	// For source_file, just return the container with all children
	if nodeType == "source_file" {
		return container
	}

	// Return the container directly - it implements ast.Node
	return container
}

// isTypeRelatedNode checks if a node represents type information that should be skipped
func (p *Parser) isTypeRelatedNode(nodeType string) bool {
	typeNodeTypes := map[string]bool{
		"type_expression":    true,
		"method_return_type": true,
		// "typed_method_parameter" is handled specially - don't skip it entirely
		"type_assertion_expression": true,
		"type_declaration":          true,
		"union_type":                true,
		"intersection_type":         true,
		"negation_type":             true,
		"parameterized_type":        true,
	}
	return typeNodeTypes[nodeType]
}

// convertTypedParameterToToken extracts just the variable name from a typed parameter
func (p *Parser) convertTypedParameterToToken(paramText string, start, end ast.Position) ast.Node {
	// Extract variable name from text like "Int $a" -> "$a"
	// Use regex to find the variable name (starts with $)
	varPattern := regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`)
	match := varPattern.FindString(paramText)

	if match != "" {
		// Create a token node with just the variable name
		return ast.NewTokenNode(ast.Identifier, match, start, end)
	}

	// Fallback - return the original text if no variable found
	return ast.NewTokenNode(ast.Identifier, paramText, start, end)
}

// ErrorNodeInfo contains detailed information about an ERROR node
type ErrorNodeInfo struct {
	Node       ast.Node
	Content    string
	StartPoint ast.Position
	EndPoint   ast.Position
}

// findErrorNodes recursively finds all ERROR nodes in the AST and returns detailed information
func (p *Parser) findErrorNodes(node ast.Node) []ErrorNodeInfo {
	var errors []ErrorNodeInfo

	if node == nil {
		return errors
	}

	if node.Type() == "ERROR" {
		errors = append(errors, ErrorNodeInfo{
			Node:       node,
			Content:    node.Text(),
			StartPoint: node.Start(),
			EndPoint:   node.End(),
		})
	}

	// Check children
	for _, child := range node.Children() {
		errors = append(errors, p.findErrorNodes(child)...)
	}

	return errors
}

// formatParseErrors creates Rust-style error messages for parse failures
func (p *Parser) formatParseErrors(errorNodes []ErrorNodeInfo, filePath, sourceContent string) string {
	var result strings.Builder

	// Split source into lines for context display
	lines := strings.Split(sourceContent, "\n")

	result.WriteString(fmt.Sprintf("error[TSP001]: parse error (%d ERROR nodes detected)\n", len(errorNodes)))

	for i, errNode := range errorNodes {
		if i > 0 {
			result.WriteString("\n")
		}

		lineNum := errNode.StartPoint.Line + 1  // Convert 0-based to 1-based
		colNum := errNode.StartPoint.Column + 1 // Convert 0-based to 1-based

		result.WriteString(fmt.Sprintf("  --> %s:%d:%d\n", filePath, lineNum, colNum))
		result.WriteString("   |\n")

		// Show the problematic line with context
		if errNode.StartPoint.Line < len(lines) {
			line := lines[errNode.StartPoint.Line]
			result.WriteString(fmt.Sprintf("%2d | %s\n", lineNum, line))
			result.WriteString("   | ")

			// Add pointer to the exact error location
			for i := 0; i < errNode.StartPoint.Column; i++ {
				if i < len(line) && line[i] == '\t' {
					result.WriteString("\t")
				} else {
					result.WriteString(" ")
				}
			}

			// Add underline for the error span
			errorLen := errNode.EndPoint.Column - errNode.StartPoint.Column
			if errorLen <= 0 {
				errorLen = 1
			}
			for i := 0; i < errorLen; i++ {
				result.WriteString("^")
			}

			result.WriteString(fmt.Sprintf(" unexpected token: '%s'\n", errNode.Content))
		}
	}

	result.WriteString("\n")
	result.WriteString("note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.\n")
	result.WriteString("      Please add test cases for this syntax to improve parser coverage.\n")

	return result.String()
}

// parseConditionalType parses conditional type expressions like (T extends U ? X : Y)
func (p *Parser) parseConditionalType(innerStr string, pos Position, originalStr string) *TypeExpression {
	// Find the question mark and colon, being careful about nested parentheses
	questionIndex := -1
	colonIndex := -1
	parenLevel := 0

	for i, ch := range innerStr {
		switch ch {
		case '(':
			parenLevel++
		case ')':
			parenLevel--
		case '?':
			if parenLevel == 0 && questionIndex == -1 {
				questionIndex = i
			}
		case ':':
			if parenLevel == 0 && colonIndex == -1 && questionIndex != -1 {
				colonIndex = i
			}
		}
	}

	if questionIndex == -1 || colonIndex == -1 {
		// Not a valid conditional type, return as simple type
		return &TypeExpression{
			BaseType:       originalStr,
			OriginalString: originalStr,
			Pos:            pos,
		}
	}

	// Extract condition part (T extends U)
	conditionPart := strings.TrimSpace(innerStr[:questionIndex])
	truePart := strings.TrimSpace(innerStr[questionIndex+1 : colonIndex])
	falsePart := strings.TrimSpace(innerStr[colonIndex+1:])

	// Parse the condition for relationship operators
	relationship, conditionType, targetType := p.parseConditionalCondition(conditionPart)

	if relationship == "" || conditionType == nil || targetType == nil {
		// Invalid conditional structure, return as simple type
		return &TypeExpression{
			BaseType:       originalStr,
			OriginalString: originalStr,
			Pos:            pos,
		}
	}

	// Parse true and false types
	trueTypeExpr := p.parseTypeExpression(truePart, pos)
	falseTypeExpr := p.parseTypeExpression(falsePart, pos)

	return &TypeExpression{
		BaseType:                originalStr,
		IsConditional:           true,
		ConditionalCondition:    conditionType,
		ConditionalTarget:       targetType,
		ConditionalRelationship: relationship,
		ConditionalTrueType:     trueTypeExpr,
		ConditionalFalseType:    falseTypeExpr,
		OriginalString:          originalStr,
		Pos:                     pos,
	}
}

// parseConditionalCondition parses the condition part of conditional types
func (p *Parser) parseConditionalCondition(conditionStr string) (relationship string, conditionType, targetType *TypeExpression) {
	// Look for relationship operators
	relationships := []string{"extends", "implements", "isa", "does"}

	for _, rel := range relationships {
		parts := strings.Split(conditionStr, " "+rel+" ")
		if len(parts) == 2 {
			leftType := p.parseTypeExpression(strings.TrimSpace(parts[0]), Position{})
			rightType := p.parseTypeExpression(strings.TrimSpace(parts[1]), Position{})
			return rel, leftType, rightType
		}
	}

	return "", nil, nil
}
