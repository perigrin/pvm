// ABOUTME: Parser for Perl code with type annotations
// ABOUTME: Implements parsing capabilities for PSC type checking

package parser

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"

	"tamarou.com/pvm/internal/parser/treesitter"
)

// Parser represents a parser for Perl code with type annotations
type Parser interface {
	// ParseFile parses a Perl file and returns its AST
	ParseFile(path string) (*AST, error)

	// ParseString parses a string containing Perl code and returns its AST
	ParseString(content string) (*AST, error)

	// ParseReader parses Perl code from a reader and returns its AST
	ParseReader(reader io.Reader) (*AST, error)
}

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

// NewParser returns a new Parser instance using tree-sitter
// This maintains backward compatibility with existing code
func NewParser() (Parser, error) {
	// Use the tree-sitter parser for compatibility
	tsParser, err := treesitter.NewParser(false)
	if err != nil {
		return nil, err
	}

	// Create the base parser
	baseParser := &treeSitterParserWrapper{
		parser: tsParser,
	}

	// Wrap with caching for better performance
	return &CachedParser{
		parser: baseParser,
		cache:  NewParserCache(100), // Cache up to 100 files
	}, nil
}

// NewParserWithOptions creates a parser with specific options
// This is the new interface for choosing scanner vs tree-sitter parsing
func NewParserWithOptions(useScanner bool) (Parser, error) {
	if useScanner {
		// Use the new scanner-based parser
		return NewTokenBasedParser(true)
	}

	// Use traditional tree-sitter parser
	return NewParser()
}

// hashContent generates a hash of the content string for caching purposes
func hashContent(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

// treeSitterParserWrapper adapts our tree-sitter parser to the Parser interface
type treeSitterParserWrapper struct {
	parser *treesitter.Parser
}

// CachedParser is a Parser implementation that caches ASTs
type CachedParser struct {
	parser Parser       // Underlying parser
	cache  *ParserCache // Cache for parsed ASTs
}

// ParseFile implements the Parser interface with caching
func (cp *CachedParser) ParseFile(path string) (*AST, error) {
	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{
			Message: "Failed to read file: " + err.Error(),
			Line:    0,
			Column:  0,
		}
	}

	// Check if we have a cached version
	if cachedAST := cp.cache.Get(path, string(content)); cachedAST != nil {
		return cachedAST, nil
	}

	// Parse with the underlying parser
	ast, err := cp.parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cp.cache.Put(path, string(content), ast)

	return ast, nil
}

// ParseString implements the Parser interface with caching
func (cp *CachedParser) ParseString(content string) (*AST, error) {
	// For strings, use a synthetic path based on content hash
	path := "memory:" + hashContent(content)

	// Check if we have a cached version
	if cachedAST := cp.cache.Get(path, content); cachedAST != nil {
		return cachedAST, nil
	}

	// Parse with the underlying parser
	ast, err := cp.parser.ParseString(content)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cp.cache.Put(path, content, ast)

	return ast, nil
}

// ParseReader implements the Parser interface with caching
func (cp *CachedParser) ParseReader(reader io.Reader) (*AST, error) {
	// For reader input, we can't cache effectively without consuming the reader
	// So we'll read it into memory first
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, &ParseError{
			Message: "Failed to read from reader: " + err.Error(),
			Line:    0,
			Column:  0,
		}
	}

	// Use the string parser which handles caching
	return cp.ParseString(string(content))
}

// We're already using hashContent from above, so no need to redefine it here.

// ParseFile implements the Parser interface
func (w *treeSitterParserWrapper) ParseFile(path string) (*AST, error) {
	tsAst, err := w.parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Convert the tree-sitter AST to our AST format
	return w.convertAst(tsAst), nil
}

// ParseString implements the Parser interface
func (w *treeSitterParserWrapper) ParseString(content string) (*AST, error) {
	tsAst, err := w.parser.ParseString(content)
	if err != nil {
		return nil, err
	}

	// Convert the tree-sitter AST to our AST format
	return w.convertAst(tsAst), nil
}

// ParseReader implements the Parser interface
func (w *treeSitterParserWrapper) ParseReader(reader io.Reader) (*AST, error) {
	tsAst, err := w.parser.ParseReader(reader)
	if err != nil {
		return nil, err
	}

	// Convert the tree-sitter AST to our AST format
	return w.convertAst(tsAst), nil
}

// convertAst converts a tree-sitter AST to our AST format
func (w *treeSitterParserWrapper) convertAst(tsAst *treesitter.AST) *AST {
	// Convert the root node
	var rootNode Node
	if tsAst.Root != nil {
		rootNode = &nodeWrapper{node: tsAst.Root}
	}

	// Convert type annotations
	var annotations []*TypeAnnotation
	for _, tsAnnot := range tsAst.TypeAnnotations {
		annotations = append(annotations, &TypeAnnotation{
			AnnotatedItem:  tsAnnot.AnnotatedItem,
			TypeExpression: convertTypeExpression(tsAnnot.TypeExpression),
			Pos:            convertPosition(tsAnnot.Pos),
			Kind:           AnnotationKind(tsAnnot.Kind),
		})
	}

	ast := &AST{
		Path:            tsAst.Path,
		Root:            rootNode,
		TypeAnnotations: annotations,
		Errors:          tsAst.Errors,
	}

	return ast
}

// nodeWrapper adapts treesitter.Node to parser.Node
type nodeWrapper struct {
	node treesitter.Node
}

func (n *nodeWrapper) Type() string {
	return n.node.Type()
}

func (n *nodeWrapper) Start() Position {
	tsPos := n.node.Start()
	return convertPosition(tsPos)
}

func (n *nodeWrapper) End() Position {
	tsPos := n.node.End()
	return convertPosition(tsPos)
}

func (n *nodeWrapper) Children() []Node {
	tsChildren := n.node.Children()
	children := make([]Node, len(tsChildren))
	for i, child := range tsChildren {
		children[i] = &nodeWrapper{node: child}
	}
	return children
}

func (n *nodeWrapper) Text() string {
	return n.node.Text()
}

// convertPosition converts treesitter.Position to parser.Position
func convertPosition(tsPos treesitter.Position) Position {
	return Position{
		Line:   tsPos.Line,
		Column: tsPos.Column,
		Offset: tsPos.Offset,
	}
}

// convertTypeExpression converts treesitter.TypeExpression to parser.TypeExpression
func convertTypeExpression(tsExpr *treesitter.TypeExpression) *TypeExpression {
	if tsExpr == nil {
		return nil
	}

	expr := &TypeExpression{
		Name:         tsExpr.BaseType,
		Union:        tsExpr.IsUnion,
		Intersection: tsExpr.IsIntersection,
		Negation:     tsExpr.IsNegation,
		Pos:          convertPosition(tsExpr.Pos),
	}

	// Convert parameters
	if len(tsExpr.Parameters) > 0 {
		expr.Params = make([]*TypeExpression, len(tsExpr.Parameters))
		for i, param := range tsExpr.Parameters {
			expr.Params[i] = convertTypeExpression(param)
		}
	}

	// For union types, put them in Params
	if len(tsExpr.UnionTypes) > 0 {
		expr.Params = make([]*TypeExpression, len(tsExpr.UnionTypes))
		for i, unionType := range tsExpr.UnionTypes {
			expr.Params[i] = convertTypeExpression(unionType)
		}
	}

	// For intersection types, put them in Params
	if len(tsExpr.IntersectionTypes) > 0 {
		expr.Params = make([]*TypeExpression, len(tsExpr.IntersectionTypes))
		for i, intType := range tsExpr.IntersectionTypes {
			expr.Params[i] = convertTypeExpression(intType)
		}
	}

	// For negated types, use the negated type's name
	if tsExpr.NegatedType != nil {
		expr.Name = tsExpr.NegatedType.BaseType
	}

	return expr
}

// We exclusively use tree-sitter for parsing now
