// ABOUTME: Parser implementation using tree-sitter-perl
// ABOUTME: Connects our parser API with tree-sitter-perl

package treesitter

import (
	"io"
	"os"
	"strings"

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

	// rawTree is the underlying tree-sitter tree
	rawTree *PerlTree
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
func (p *Parser) ParseFile(path string) (*AST, error) {
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
	ast := &AST{
		Path:            path,
		Root:            newSimpleNode("root"),
		TypeAnnotations: typeAnnotations,
		Errors:          []error{},
		rawTree:         tree,
	}

	return ast, nil
}

// convertPerlTypeAnnotation converts a PerlTypeAnnotation to the standard TypeAnnotation format
func (p *Parser) convertPerlTypeAnnotation(perlAnn *PerlTypeAnnotation, content string) (*TypeAnnotation, error) {
	// Calculate line and column from byte position
	pos := p.calculatePosition(perlAnn.StartPos, content)

	// Create TypeExpression
	typeExpr := &TypeExpression{
		BaseType: perlAnn.TypeName,
	}

	// Determine annotation kind
	var kind AnnotationKind
	switch perlAnn.Kind {
	case "variable":
		kind = VarAnnotation
	case "subroutine":
		kind = SubParamAnnotation // Simplified for now
	case "method":
		kind = MethodParamAnnotation // Simplified for now
	default:
		kind = VarAnnotation // Default fallback
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

// ParseString parses a string containing Perl code and returns its AST
func (p *Parser) ParseString(content string) (*AST, error) {
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

	// Create an AST from the parse tree
	ast := &AST{
		Root:            newSimpleNode("root"),
		TypeAnnotations: typeAnnotations,
		Errors:          []error{},
		rawTree:         tree,
	}

	return ast, nil
}

// ParseReader parses Perl code from a reader and returns its AST
func (p *Parser) ParseReader(reader io.Reader) (*AST, error) {
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
