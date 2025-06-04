// ABOUTME: Parser implementation using tree-sitter-perl
// ABOUTME: Connects our parser API with tree-sitter-perl

package treesitter

import (
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
	// SubReturnAnnotation is for subroutine return annotations like "sub name() -> ReturnType"
	SubReturnAnnotation
	// MethodParamAnnotation is for method parameter annotations like "method name(Type $param)"
	MethodParamAnnotation
	// MethodReturnAnnotation is for method return annotations like "method name() -> ReturnType"
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

	// OriginalString is the original string representation
	OriginalString string

	// Position is the position of the type expression in the source code
	Pos Position
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
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.NewSystemError("001",
			"File does not exist", err).
			WithLocation(path)
	}

	// Parse the file using a simplified approach until tree-sitter is integrated
	tree, err := p.perlParser.ParseFile(path)
	if err != nil {
		return nil, errors.NewSystemError("002",
			"Failed to parse file", err).
			WithLocation(path)
	}

	// Extract type annotations from the tree
	perlAnnotations, err := tree.FindTypeAnnotations()
	if err != nil {
		return nil, errors.NewSystemError("003",
			"Failed to extract type annotations", err).
			WithLocation(path)
	}

	// Convert PerlTypeAnnotations to TypeAnnotations
	var typeAnnotations []*TypeAnnotation
	for _, perlAnn := range perlAnnotations {
		annotation, convErr := p.convertPerlTypeAnnotation(perlAnn, string(tree.Content))
		if convErr != nil {
			// Log the error but don't fail the parse
			log.Debugf("Failed to convert annotation %s: %v", perlAnn.ItemName, convErr)
			continue
		}
		typeAnnotations = append(typeAnnotations, annotation)
	}

	// Create an AST from the parse tree
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
	constraintAnnotations := p.extractConstraintAnnotations(string(tree.Content))
	astTypeAnnotations = append(astTypeAnnotations, constraintAnnotations...)

	astResult := &ast.AST{
		Source:          string(tree.Content),
		Path:            path,
		Root:            astRoot,
		TypeAnnotations: astTypeAnnotations,
		Errors:          []error{},
	}

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
	case "type_assertion":
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
func (p *Parser) ParseString(content string) (*ast.AST, error) {
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

	// Validate the parsed content for syntax errors
	syntaxErrors := p.validateSyntax(content, astTypeAnnotations)

	// Create an AST from the parse tree
	astResult := &ast.AST{
		Source:          content,
		Root:            astRoot,
		TypeAnnotations: astTypeAnnotations,
		Errors:          syntaxErrors,
	}

	// If there are syntax errors, return them
	if len(syntaxErrors) > 0 {
		return astResult, syntaxErrors[0] // Return the first error
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

	// Skip type-related nodes entirely
	if p.isTypeRelatedNode(nodeType) {
		return nil
	}

	// Handle special cases that need custom processing
	if nodeType == "typed_method_parameter" {
		return p.convertTypedParameterToToken(nodeText, start, end)
	}

	// First, check if this is a structural token that should be preserved as a token
	if tokenNode := p.convertToTokenNode(nodeType, nodeText, start, end); tokenNode != nil {
		return tokenNode
	}

	// Convert children recursively - include ALL children to preserve CST structure
	var children []ast.Node
	for _, child := range node.Children() {
		if childAST := p.convertToASTNode(child); childAST != nil {
			children = append(children, childAST)
		}
	}

	// Handle higher-level constructs that need semantic processing
	if nodeType == "assignment_expression" &&
		(strings.HasPrefix(strings.TrimSpace(nodeText), "my ") ||
			strings.HasPrefix(strings.TrimSpace(nodeText), "our ") ||
			strings.HasPrefix(strings.TrimSpace(nodeText), "state ")) {
		return p.convertToVarDeclAST(nodeText, start, end)
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

// convertToVarDeclAST converts a variable declaration text to ast.VarDecl
func (p *Parser) convertToVarDeclAST(nodeText string, start, end ast.Position) ast.Node {
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

	return &ast.TypeAnnotation{
		Kind:           ast.AnnotationKind(ta.Kind),
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

	astTypeExpr := ast.NewTypeExpression(te.BaseType, astParams, astPos, astPos)
	astTypeExpr.IsUnion = te.IsUnion
	astTypeExpr.IsIntersection = te.IsIntersection
	astTypeExpr.IsNegation = te.IsNegation
	astTypeExpr.UnionTypes = astUnionTypes
	astTypeExpr.IntersectionTypes = astIntersectionTypes
	astTypeExpr.NegatedType = p.convertToASTTypeExpression(te.NegatedType)
	astTypeExpr.OriginalString = te.OriginalString
	return astTypeExpr
}

// ParseTypeExpression parses a type expression string and returns a TypeExpression
func ParseTypeExpression(text string, pos Position) (*TypeExpression, error) {
	// This is a simplified implementation that would need to be enhanced
	// to handle more complex type expressions including unions and intersections

	// Check for union types (Type1|Type2)
	if strings.Contains(text, "|") {
		unionParts := strings.Split(text, "|")
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

	// Check for intersection types (Type1&Type2)
	if strings.Contains(text, "&") {
		intersectionParts := strings.Split(text, "&")
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

	// Check for parameterized types (Type[Param1, Param2, ...])
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

// convertToSubDeclAST converts a subroutine_declaration_statement node to ast.SubDecl
func (p *Parser) convertToSubDeclAST(node Node, start, end ast.Position) ast.Node {
	subroutineName := ""
	var params []*ast.Parameter
	var returnType *ast.TypeExpression
	var body *ast.BlockStmt

	// Extract subroutine name from tree-sitter node structure
	// Based on the tree dump: subroutine_declaration_statement has children: sub, bareword, block
	children := node.Children()
	for _, child := range children {
		switch child.Type() {
		case "bareword":
			// This should be the subroutine name (e.g., "hello")
			subroutineName = child.Text()
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

			// For now, create an empty block - we can enhance this later to parse block contents
			var blockStatements []ast.StatementNode
			body = ast.NewBlockStmt(blockStatements, blockStart, blockEnd)
		}
	}

	// Create subroutine declaration
	// For now, create with empty parameters and no return type - we can enhance this later
	return ast.NewSubDecl(subroutineName, params, returnType, body, false, start, end)
}

// processClassBlock processes the contents of a class block
func (p *Parser) processClassBlock(blockNode Node, classDecl *ast.ClassDecl) {
	for _, stmt := range blockNode.Children() {
		// Process field declarations
		if strings.Contains(stmt.Text(), "field ") {
			if field := p.extractFieldDecl(stmt); field != nil {
				classDecl.AddField(field)
			}
		}

		// Process method declarations
		if strings.Contains(stmt.Text(), "method ") {
			if method := p.extractMethodDecl(stmt); method != nil {
				classDecl.AddMethod(method)
			}
		}
	}
}

// processRoleBlock processes the contents of a role block
func (p *Parser) processRoleBlock(blockNode Node, roleDecl *ast.RoleDecl) {
	for _, stmt := range blockNode.Children() {
		// Process field declarations
		if strings.Contains(stmt.Text(), "field ") {
			if field := p.extractFieldDecl(stmt); field != nil {
				roleDecl.AddField(field)
			}
		}

		// Process method declarations
		if strings.Contains(stmt.Text(), "method ") {
			if method := p.extractMethodDecl(stmt); method != nil {
				roleDecl.AddProvidedMethod(method)
			}
		}
	}
}

// extractFieldDecl extracts a field declaration from a statement node
func (p *Parser) extractFieldDecl(stmt Node) *ast.FieldDecl {
	// Simple parsing - this would need to be enhanced for full parsing
	text := stmt.Text()

	// Basic pattern: field Type $name = value;
	if !strings.HasPrefix(strings.TrimSpace(text), "field ") {
		return nil
	}

	start := ast.Position{Line: stmt.Start().Line, Column: stmt.Start().Column}
	end := ast.Position{Line: stmt.End().Line, Column: stmt.End().Column}

	// Extract field name (simplified)
	parts := strings.Fields(text)
	if len(parts) < 3 {
		return nil
	}

	fieldName := parts[2]                          // Should be $name
	fieldName = strings.TrimPrefix(fieldName, "$") // Remove $

	// Create variable expression for the field
	varExpr := ast.NewVariableExpr(fieldName, "$", start, end)

	// Create field declaration
	return ast.NewFieldDecl(fieldName, nil, varExpr, nil, start, end)
}

// extractMethodDecl extracts a method declaration from a statement node
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

		// Look for method signatures with constraints
		if strings.Contains(line, "method ") && strings.Contains(line, " where ") {
			annotations = append(annotations, p.extractMethodConstraintAnnotations(line, lineNum+1)...)
		}

		// Look for method signatures without constraints (to extract basic type annotations)
		if strings.Contains(line, "method ") && !strings.Contains(line, " where ") {
			annotations = append(annotations, p.extractMethodConstraintAnnotations(line, lineNum+1)...)
		}

		// Look for field declarations
		if strings.Contains(line, "field ") && strings.Contains(line, " $") {
			annotations = append(annotations, p.extractFieldAnnotations(line, lineNum+1)...)
		}

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
	// Simple parsing - this would need to be enhanced for full parsing
	text := stmt.Text()

	if !strings.Contains(text, "method ") {
		return nil
	}

	start := ast.Position{Line: stmt.Start().Line, Column: stmt.Start().Column}
	end := ast.Position{Line: stmt.End().Line, Column: stmt.End().Column}

	// Extract constraints from the method signature
	constraints := p.extractConstraintsFromText(text)

	// Extract method name (simplified)
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return nil
	}

	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "method ") {
		return nil
	}

	// Parse method name from first line
	parts := strings.Fields(firstLine)
	if len(parts) < 2 {
		return nil
	}

	methodName := parts[1]
	if strings.Contains(methodName, "(") {
		methodName = strings.Split(methodName, "(")[0]
	}

	// Create basic method declaration with constraints
	methodDecl := ast.NewMethodDecl(methodName, nil, nil, nil, start, end)
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

// convertToTokenNode converts structural tokens to AST token nodes
func (p *Parser) convertToTokenNode(nodeType, nodeText string, start, end ast.Position) ast.Node {
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
		"type_assertion":     true,
		"type_declaration":   true,
		"union_type":         true,
		"intersection_type":  true,
		"negation_type":      true,
		"parameterized_type": true,
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
