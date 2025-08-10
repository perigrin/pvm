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

// NewParser returns a new Parser instance with enhanced error handling and caching
// This is the unified parser creation path used throughout the system
func NewParser() (Parser, error) {
	// Create enhanced parser with type error identification
	enhancedParser, err := NewEnhancedParser()
	if err != nil {
		return nil, err
	}

	// Wrap with caching for better performance
	return &CachedParser{
		parser: enhancedParser,
		cache:  NewParserCache(100),
	}, nil
}

// NewParserWithOptions creates a parser with specific options
// Note: useScanner parameter is ignored - always returns tree-sitter parser
func NewParserWithOptions(useScanner bool) (Parser, error) {
	// Always use tree-sitter parser - scanner infrastructure has been removed
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

// VersionRequirement represents a Perl version requirement found in a script
type VersionRequirement struct {
	// Version is the required Perl version (e.g., "5.38", "5.40")
	Version string

	// Operator is the comparison operator (">=", "==", ">", etc.)
	// If empty, defaults to ">=" for "use v5.xx" statements
	Operator string

	// Line is the line number where the requirement was found
	Line int

	// Column is the column number where the requirement was found
	Column int
}

// ExtractVersionRequirements parses a Perl script and extracts version requirements
// from "use v5.xx;" statements in the code
func ExtractVersionRequirements(scriptPath string) (*VersionRequirement, error) {
	// Create a parser
	parser, err := NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	// Parse the script
	ast, err := parser.ParseFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse script %s: %w", scriptPath, err)
	}

	// Search for version requirements in the AST
	return extractVersionFromAST(ast)
}

// extractVersionFromAST searches an AST for version requirements
func extractVersionFromAST(ast *AST) (*VersionRequirement, error) {
	if ast.Root == nil {
		return nil, nil
	}

	// Recursively search for use statements with version requirements
	return findVersionInNode(ast.Root)
}

// findVersionInNode recursively searches a node and its children for version requirements
func findVersionInNode(node Node) (*VersionRequirement, error) {
	// Check if this node is a use statement with version
	if req := extractVersionFromUseStatement(node); req != nil {
		return req, nil
	}

	// Recursively check child nodes
	children := node.Children()
	for _, child := range children {
		if child != nil {
			if req, err := findVersionInNode(child); err != nil {
				return nil, err
			} else if req != nil {
				return req, nil
			}
		}
	}

	return nil, nil
}

// extractVersionFromUseStatement extracts version requirement from a use statement node
func extractVersionFromUseStatement(node Node) *VersionRequirement {
	// Check if this is a use statement
	if node.Type() != "use_statement" && node.Type() != "statement" {
		return nil
	}

	// Get the text content of the node
	text := node.Text()

	// Look for patterns like "use v5.38;" or "use 5.038;"
	// This is a simple pattern match - in production you'd want more sophisticated parsing
	if len(text) < 8 { // Minimum length for "use v5.x;"
		return nil
	}

	// Simple pattern matching for version requirements
	// Look for "use v" followed by version number
	if pos := findVersionPattern(text); pos != nil {
		return &VersionRequirement{
			Version:  pos.Version,
			Operator: ">=",                  // Default to >= for "use v5.xx" statements
			Line:     node.Start().Line + 1, // Convert to 1-based
			Column:   node.Start().Column + 1,
		}
	}

	return nil
}

// versionMatch represents a found version pattern
type versionMatch struct {
	Version string
}

// findVersionPattern searches for version patterns in text
func findVersionPattern(text string) *versionMatch {
	// Simple regex-like matching for "use v5.xx" patterns
	// Look for "use v" followed by digits and dots

	// Find "use v" or "use "
	useIndex := -1
	for i := 0; i < len(text)-6; i++ {
		if text[i:i+6] == "use v5" || text[i:i+5] == "use 5" {
			useIndex = i
			break
		}
	}

	if useIndex == -1 {
		return nil
	}

	// Extract version after "use v" or "use "
	var versionStart int
	if text[useIndex:useIndex+6] == "use v5" {
		versionStart = useIndex + 4 // Skip "use "
	} else {
		versionStart = useIndex + 4 // Skip "use "
	}

	// Find the end of the version (until semicolon, whitespace, or end)
	versionEnd := versionStart
	for versionEnd < len(text) &&
		text[versionEnd] != ';' &&
		text[versionEnd] != ' ' &&
		text[versionEnd] != '\t' &&
		text[versionEnd] != '\n' &&
		text[versionEnd] != '\r' {
		versionEnd++
	}

	if versionEnd <= versionStart {
		return nil
	}

	versionStr := text[versionStart:versionEnd]

	// Validate that this looks like a version number
	if isValidVersionString(versionStr) {
		return &versionMatch{
			Version: normalizeVersion(versionStr),
		}
	}

	return nil
}

// isValidVersionString checks if a string looks like a valid Perl version
func isValidVersionString(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Should start with 'v' or digit
	if s[0] != 'v' && (s[0] < '0' || s[0] > '9') {
		return false
	}

	// Count dots and digits
	dotCount := 0
	digitCount := 0

	start := 0
	if s[0] == 'v' {
		start = 1
	}

	for i := start; i < len(s); i++ {
		if s[i] == '.' {
			dotCount++
		} else if s[i] >= '0' && s[i] <= '9' {
			digitCount++
		} else {
			return false // Invalid character
		}
	}

	// Should have at least one digit and reasonable number of dots
	return digitCount > 0 && dotCount <= 3
}

// normalizeVersion normalizes a version string for comparison
func normalizeVersion(s string) string {
	// Remove leading 'v' if present
	if len(s) > 0 && s[0] == 'v' {
		s = s[1:]
	}

	// Convert 5.038 style to 5.38 style if needed
	// This is a simplified conversion
	if len(s) >= 5 && s[1] == '.' && s[2] == '0' {
		// Handle cases like "5.038" -> "5.38"
		return s[:2] + s[3:]
	}

	return s
}
