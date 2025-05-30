// ABOUTME: Parser for Perl code with type annotations
// ABOUTME: Implements parsing capabilities for PSC type checking

//go:generate moq -out parser_mock.go . Parser

package parser

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// Parser represents a parser for Perl code with type annotations
type Parser interface {
	// ParseFile parses a Perl file and returns its AST
	ParseFile(path string) (*ast.AST, error)

	// ParseString parses a string containing Perl code and returns its AST
	ParseString(content string) (*ast.AST, error)

	// ParseReader parses Perl code from a reader and returns its AST
	ParseReader(reader io.Reader) (*ast.AST, error)
}

// Backward compatibility type aliases
// These allow existing code to continue working while we migrate to consolidated AST types
type AST = ast.AST
type Node = ast.Node
type Position = ast.Position
type TypeAnnotation = ast.TypeAnnotation
type TypeExpression = ast.TypeExpression
type AnnotationKind = ast.AnnotationKind

// Backward compatibility constants
const (
	VarAnnotation          = ast.VarAnnotation
	SubParamAnnotation     = ast.SubParamAnnotation
	SubReturnAnnotation    = ast.SubReturnAnnotation
	MethodParamAnnotation  = ast.MethodParamAnnotation
	MethodReturnAnnotation = ast.MethodReturnAnnotation
	AttrAnnotation         = ast.FieldAnnotation // Map old name to new name
	TypeDeclAnnotation     = ast.TypeDeclAnnotation
)

// ParseError represents a parsing error
type ParseError struct {
	// Message is the error message
	Message string

	// Line is the line number where the error occurred
	Line int

	// Column is the column number where the error occurred
	Column int

	// Path is the path to the file where the error occurred
	Path string
}

// Error implements the error interface
func (pe *ParseError) Error() string {
	if pe.Path != "" {
		return fmt.Sprintf("%s:%d:%d: %s", pe.Path, pe.Line, pe.Column, pe.Message)
	}
	return fmt.Sprintf("%d:%d: %s", pe.Line, pe.Column, pe.Message)
}

// ParseTypeExpression parses a type expression and returns it in the consolidated format
// for backward compatibility
func ParseTypeExpression(text string, pos Position) (*ast.TypeExpression, error) {
	tsPos := treesitter.Position{
		Line:   pos.Line,
		Column: pos.Column,
		Offset: pos.Offset,
	}

	tsExpr, err := treesitter.ParseTypeExpression(text, tsPos)
	if err != nil {
		return nil, err
	}

	// Convert treesitter TypeExpression to consolidated TypeExpression
	return convertTypeExpression(tsExpr), nil
}

// NewParser returns a new Parser instance using pooling for efficiency
// This now uses pooled parsers by default for better performance and thread safety
func NewParser() (Parser, error) {
	// Use pooled parser for better performance and thread safety
	return NewPooledParser()
}

// NewParserDirect creates a new parser instance without pooling (for internal use)
func NewParserDirect() (Parser, error) {
	// For now, use tree-sitter parser directly since it has working type annotation extraction
	// TODO: Integrate scanner-based parser with type annotation extraction
	tsParser, err := treesitter.NewParser(false)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree-sitter parser: %v", err)
	}

	baseParser := &treeSitterParserWrapper{
		parser: tsParser,
	}

	return &CachedParser{
		parser: baseParser,
		cache:  NewParserCache(100),
	}, nil
}

// NewParserWithOptions creates a parser with specific options
// This is the new interface for choosing scanner vs tree-sitter parsing
func NewParserWithOptions(useScanner bool) (Parser, error) {
	if useScanner {
		// Use the new scanner-based parser
		return NewTokenBasedParser(true)
	}

	// Use traditional tree-sitter parser directly
	return NewTreeSitterParser()
}

// NewTreeSitterParser creates a parser that explicitly uses tree-sitter
// This is provided for cases where tree-sitter parsing is specifically required
func NewTreeSitterParser() (Parser, error) {
	tsParser, err := treesitter.NewParser(false)
	if err != nil {
		return nil, err
	}

	baseParser := &treeSitterParserWrapper{
		parser: tsParser,
	}

	return &CachedParser{
		parser: baseParser,
		cache:  NewParserCache(100),
	}, nil
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
func (w *treeSitterParserWrapper) ParseFile(path string) (*ast.AST, error) {
	tsAst, err := w.parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// The tree-sitter parser now returns *ast.AST directly, no conversion needed
	return tsAst, nil
}

// ParseString implements the Parser interface
func (w *treeSitterParserWrapper) ParseString(content string) (*ast.AST, error) {
	tsAst, err := w.parser.ParseString(content)
	if err != nil {
		return nil, err
	}

	// The tree-sitter parser now returns *ast.AST directly, no conversion needed
	return tsAst, nil
}

// ParseReader implements the Parser interface
func (w *treeSitterParserWrapper) ParseReader(reader io.Reader) (*ast.AST, error) {
	tsAst, err := w.parser.ParseReader(reader)
	if err != nil {
		return nil, err
	}

	// The tree-sitter parser now returns *ast.AST directly, no conversion needed
	return tsAst, nil
}

// convertPosition converts treesitter.Position to consolidated ast.Position
func convertPosition(tsPos treesitter.Position) ast.Position {
	return ast.Position{
		Line:   tsPos.Line,
		Column: tsPos.Column,
		Offset: tsPos.Offset,
	}
}

// convertTypeExpression converts treesitter.TypeExpression to consolidated ast.TypeExpression
func convertTypeExpression(tsExpr *treesitter.TypeExpression) *ast.TypeExpression {
	if tsExpr == nil {
		return nil
	}

	start := convertPosition(tsExpr.Pos)
	expr := ast.NewTypeExpression(tsExpr.BaseType, nil, start, start)
	expr.IsUnion = tsExpr.IsUnion
	expr.IsIntersection = tsExpr.IsIntersection
	expr.IsNegation = tsExpr.IsNegation
	expr.OriginalString = tsExpr.OriginalString

	// Convert parameters
	if len(tsExpr.Parameters) > 0 {
		expr.Parameters = make([]*ast.TypeExpression, len(tsExpr.Parameters))
		for i, param := range tsExpr.Parameters {
			expr.Parameters[i] = convertTypeExpression(param)
		}
	}

	// Convert union types
	if len(tsExpr.UnionTypes) > 0 {
		expr.UnionTypes = make([]*ast.TypeExpression, len(tsExpr.UnionTypes))
		for i, unionType := range tsExpr.UnionTypes {
			expr.UnionTypes[i] = convertTypeExpression(unionType)
		}
	}

	// Convert intersection types
	if len(tsExpr.IntersectionTypes) > 0 {
		expr.IntersectionTypes = make([]*ast.TypeExpression, len(tsExpr.IntersectionTypes))
		for i, intType := range tsExpr.IntersectionTypes {
			expr.IntersectionTypes[i] = convertTypeExpression(intType)
		}
	}

	// Convert negated type
	if tsExpr.NegatedType != nil {
		expr.NegatedType = convertTypeExpression(tsExpr.NegatedType)
	}

	return expr
}

// We exclusively use tree-sitter for parsing now
