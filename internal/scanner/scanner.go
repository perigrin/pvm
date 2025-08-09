// ABOUTME: Scanner package provides lexical analysis interface following TypeScript-Go patterns
// ABOUTME: Separates tokenization concerns from parsing for better modularity and performance

//go:generate stringer -type=TokenType -output=token_type_string.go
//go:generate moq -out scanner_mock.go . Scanner Token TokenIterator

package scanner

import (
	"io"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// TokenType represents the type of a token
type TokenType int

const (
	// Basic token types
	TokenEOF TokenType = iota
	TokenError
	TokenWhitespace
	TokenComment
	TokenNewline

	// Literals
	TokenString
	TokenNumber
	TokenIdentifier
	TokenVariable
	TokenArrayVariable
	TokenHashVariable

	// Keywords
	TokenMy
	TokenOur
	TokenState
	TokenSub
	TokenMethod
	TokenField
	TokenTypeKeyword
	TokenUse
	TokenPackage
	TokenClass
	TokenAs
	TokenWhere
	TokenDoes
	TokenCan
	TokenRole

	// Operators and punctuation
	TokenAssign
	TokenComma
	TokenSemicolon
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenArrow
	TokenPipe
	TokenAmpersand
	TokenExclamation

	// Type annotation specific
	TokenTypeAnnotation
	TokenTypeExpression
	TokenUnionOperator
	TokenIntersectionOperator
	TokenNegationOperator
)

// Position represents a position in the source code
type Position struct {
	Line   int
	Column int
	Offset int
}

// Token represents a lexical token
type Token interface {
	// Type returns the type of the token
	Type() TokenType

	// Value returns the string value of the token
	Value() string

	// Position returns the position of the token in the source
	Position() Position

	// Length returns the length of the token
	Length() int
}

// Scanner provides lexical analysis capabilities
// Deprecated: Use TokenExtractor interface for new code
type Scanner interface {
	// ScanFile scans a file and returns a token channel
	ScanFile(path string) (TokenIterator, error)

	// ScanString scans a string and returns a token channel
	ScanString(content string) (TokenIterator, error)

	// ScanReader scans from a reader and returns a token channel
	ScanReader(reader io.Reader) (TokenIterator, error)
}

// TokenExtractor provides pure token extraction from already-parsed AST structures
// This is the new interface that follows the single-parse, multiple-consumers architecture
type TokenExtractor interface {
	// ExtractTokens extracts tokens from a pre-parsed AST
	ExtractTokens(ast *ast.AST) (TokenIterator, error)
}

// TokenIterator provides sequential access to tokens
type TokenIterator interface {
	// Next returns the next token, or nil if EOF
	Next() Token

	// Peek returns the next token without advancing
	Peek() Token

	// HasNext returns true if there are more tokens
	HasNext() bool

	// Reset resets the iterator to the beginning
	Reset()

	// Position returns the current position in the token stream
	Position() int
}

// treeSitterToken wraps a tree-sitter node as a Token
type treeSitterToken struct {
	tokenType TokenType
	value     string
	position  Position
	length    int
}

func (t *treeSitterToken) Type() TokenType {
	return t.tokenType
}

func (t *treeSitterToken) Value() string {
	return t.value
}

func (t *treeSitterToken) Position() Position {
	return t.position
}

func (t *treeSitterToken) Length() int {
	return t.length
}

// treeSitterScanner implements both Scanner (deprecated) and TokenExtractor interfaces
type treeSitterScanner struct {
	parser      *treesitter.Parser // Only used for deprecated Scanner methods
	debug       bool
	poolManager *TokenPoolManager
}

// NewScanner creates a new Scanner instance using tree-sitter
func NewScanner(debug bool) (Scanner, error) {
	parser, err := treesitter.NewParser(debug)
	if err != nil {
		return nil, errors.NewSystemError("001",
			"Failed to create tree-sitter parser", err)
	}

	return &treeSitterScanner{
		parser:      parser,
		debug:       debug,
		poolManager: GetGlobalTokenPoolManager(),
	}, nil
}

// NewScannerWithPool creates a new Scanner instance with a specific token pool manager
func NewScannerWithPool(debug bool, poolManager *TokenPoolManager) (Scanner, error) {
	parser, err := treesitter.NewParser(debug)
	if err != nil {
		return nil, errors.NewSystemError("001",
			"Failed to create tree-sitter parser", err)
	}

	return &treeSitterScanner{
		parser:      parser,
		debug:       debug,
		poolManager: poolManager,
	}, nil
}

// NewTokenExtractor creates a new TokenExtractor instance
// This follows the new architecture where parsing is done separately
func NewTokenExtractor(poolManager *TokenPoolManager) TokenExtractor {
	if poolManager == nil {
		poolManager = GetGlobalTokenPoolManager()
	}

	return &treeSitterScanner{
		parser:      nil, // No parser needed - we consume pre-parsed AST
		debug:       false,
		poolManager: poolManager,
	}
}

// NewTokenExtractorWithDebug creates a new TokenExtractor instance with debug enabled
func NewTokenExtractorWithDebug(debug bool, poolManager *TokenPoolManager) TokenExtractor {
	if poolManager == nil {
		poolManager = GetGlobalTokenPoolManager()
	}

	return &treeSitterScanner{
		parser:      nil, // No parser needed - we consume pre-parsed AST
		debug:       debug,
		poolManager: poolManager,
	}
}

// ScanFile implements Scanner interface
// Deprecated: Use TokenExtractor.ExtractTokens() with pre-parsed AST instead
func (s *treeSitterScanner) ScanFile(path string) (TokenIterator, error) {
	return nil, errors.NewSystemError("009",
		"ScanFile is deprecated - use TokenExtractor.ExtractTokens() with pre-parsed AST", nil).
		WithLocation(path)
}

// ScanString implements Scanner interface
// Deprecated: Use TokenExtractor.ExtractTokens() with pre-parsed AST instead
func (s *treeSitterScanner) ScanString(content string) (TokenIterator, error) {
	return nil, errors.NewSystemError("008",
		"ScanString is deprecated - use TokenExtractor.ExtractTokens() with pre-parsed AST", nil)
}

// ScanReader implements Scanner interface
// Deprecated: Use TokenExtractor.ExtractTokens() with pre-parsed AST instead
func (s *treeSitterScanner) ScanReader(reader io.Reader) (TokenIterator, error) {
	return nil, errors.NewSystemError("010",
		"ScanReader is deprecated - use TokenExtractor.ExtractTokens() with pre-parsed AST", nil)
}

// ExtractTokens implements TokenExtractor interface
// This is the new method that extracts tokens from a pre-parsed AST
func (s *treeSitterScanner) ExtractTokens(astTree *ast.AST) (TokenIterator, error) {
	if astTree == nil {
		return nil, errors.NewSystemError("006",
			"AST is nil", nil)
	}

	if astTree.Root == nil {
		return nil, errors.NewSystemError("007",
			"AST root is nil", nil)
	}

	// Use pooled token slice for collecting tokens
	tokenSlice := s.poolManager.GetTokenSlice(128) // Start with reasonable capacity
	defer s.poolManager.ReleaseTokenSlice(tokenSlice)

	tokens := *tokenSlice
	tokens = tokens[:0] // Clear slice but keep capacity

	// Extract tokens from AST by traversing the tree
	s.extractTokensFromAST(astTree.Root, &tokens)

	// Create a new slice to return (not pooled since it needs to outlive this function)
	result := make([]Token, len(tokens))
	copy(result, tokens)

	return s.poolManager.CreateTokenIterator(result), nil
}

// extractTokensFromAST recursively traverses the AST and converts nodes to tokens
func (s *treeSitterScanner) extractTokensFromAST(node ast.Node, tokens *[]Token) {
	if node == nil {
		return
	}

	// Convert AST node to token based on node type
	if token := s.convertASTNodeToToken(node); token != nil {
		*tokens = append(*tokens, token)
	}

	// Recursively process children
	for _, child := range node.Children() {
		s.extractTokensFromAST(child, tokens)
	}
}

// convertASTNodeToToken converts a single AST node to a token
func (s *treeSitterScanner) convertASTNodeToToken(node ast.Node) Token {
	nodeType := node.Type()
	text := node.Text()
	start := node.Start()

	// Skip empty or whitespace-only nodes
	if text == "" || strings.TrimSpace(text) == "" {
		return nil
	}

	// Map AST node types to scanner token types
	tokenType := s.mapNodeTypeToTokenType(nodeType, text)

	// Create position from AST position
	position := Position{
		Line:   start.Line,
		Column: start.Column,
		Offset: start.Offset,
	}

	return s.poolManager.CreateToken(
		tokenType,
		text,
		position,
		len(text),
	)
}

// mapNodeTypeToTokenType maps AST node types to scanner token types
func (s *treeSitterScanner) mapNodeTypeToTokenType(nodeType, text string) TokenType {
	switch nodeType {
	// Keywords
	case "my_declaration", "variable_declaration":
		if text == "my" || text == "our" || text == "state" {
			return TokenMy
		}
	case "subroutine_declaration", "sub_declaration":
		return TokenSub
	case "method_declaration":
		return TokenMethod
	case "field_declaration":
		return TokenField
	case "type_declaration":
		return TokenTypeKeyword
	case "use_statement":
		return TokenUse
	case "package_statement":
		return TokenPackage
	case "class_declaration":
		return TokenClass
	case "role_declaration":
		return TokenRole
	case "type_assertion":
		return TokenAs

	// Variables
	case "scalar_variable", "variable":
		if strings.HasPrefix(text, "$") {
			return TokenVariable
		} else if strings.HasPrefix(text, "@") {
			return TokenArrayVariable
		} else if strings.HasPrefix(text, "%") {
			return TokenHashVariable
		}

	// Literals
	case "string_literal", "quoted_string":
		return TokenString
	case "number_literal", "integer_literal", "float_literal":
		return TokenNumber
	case "identifier":
		return TokenIdentifier

	// Operators and punctuation
	case "assignment_operator":
		return TokenAssign
	case "comma":
		return TokenComma
	case "semicolon":
		return TokenSemicolon
	case "left_parenthesis", "(":
		return TokenLParen
	case "right_parenthesis", ")":
		return TokenRParen
	case "left_brace", "{":
		return TokenLBrace
	case "right_brace", "}":
		return TokenRBrace
	case "left_bracket", "[":
		return TokenLBracket
	case "right_bracket", "]":
		return TokenRBracket
	case "arrow", "->":
		return TokenArrow
	case "pipe", "|":
		return TokenPipe
	case "ampersand", "&":
		return TokenAmpersand
	case "exclamation", "!":
		return TokenExclamation

	// Type annotation specific
	case "type_annotation":
		return TokenTypeAnnotation
	case "type_expression":
		return TokenTypeExpression
	case "union_type":
		return TokenUnionOperator
	case "intersection_type":
		return TokenIntersectionOperator
	case "negation_type":
		return TokenNegationOperator

	// Comments and whitespace
	case "comment":
		return TokenComment
	case "whitespace":
		return TokenWhitespace
	case "newline":
		return TokenNewline
	}

	// Fallback: analyze text content for basic token classification
	return s.classifyTokenByText(text)
}

// classifyTokenByText provides fallback token classification based on text content
func (s *treeSitterScanner) classifyTokenByText(text string) TokenType {
	text = strings.TrimSpace(text)
	if text == "" {
		return TokenWhitespace
	}

	switch text {
	case "my", "our", "state":
		return TokenMy
	case "sub":
		return TokenSub
	case "method":
		return TokenMethod
	case "field":
		return TokenField
	case "type":
		return TokenTypeKeyword
	case "use":
		return TokenUse
	case "package":
		return TokenPackage
	case "class":
		return TokenClass
	case "role":
		return TokenRole
	case "as":
		return TokenAs
	case "where":
		return TokenWhere
	case "does":
		return TokenDoes
	case "can":
		return TokenCan
	case "=":
		return TokenAssign
	case ",":
		return TokenComma
	case ";":
		return TokenSemicolon
	case "(":
		return TokenLParen
	case ")":
		return TokenRParen
	case "{":
		return TokenLBrace
	case "}":
		return TokenRBrace
	case "[":
		return TokenLBracket
	case "]":
		return TokenRBracket
	case "->":
		return TokenArrow
	case "|":
		return TokenPipe
	case "&":
		return TokenAmpersand
	case "!":
		return TokenExclamation
	}

	// Variable patterns
	if strings.HasPrefix(text, "$") {
		return TokenVariable
	} else if strings.HasPrefix(text, "@") {
		return TokenArrayVariable
	} else if strings.HasPrefix(text, "%") {
		return TokenHashVariable
	}

	// Number patterns
	if strings.ContainsAny(text, "0123456789") && !strings.ContainsAny(text, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_") {
		return TokenNumber
	}

	// String patterns (quoted)
	if (strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"")) ||
		(strings.HasPrefix(text, "'") && strings.HasSuffix(text, "'")) {
		return TokenString
	}

	// Default to identifier
	return TokenIdentifier
}

// tokenIterator implements TokenIterator
type tokenIterator struct {
	tokens []Token
	pos    int
}

func (ti *tokenIterator) Next() Token {
	if ti.pos >= len(ti.tokens) {
		return nil
	}
	token := ti.tokens[ti.pos]
	ti.pos++
	return token
}

func (ti *tokenIterator) Peek() Token {
	if ti.pos >= len(ti.tokens) {
		return nil
	}
	return ti.tokens[ti.pos]
}

func (ti *tokenIterator) HasNext() bool {
	return ti.pos < len(ti.tokens)
}

func (ti *tokenIterator) Reset() {
	ti.pos = 0
}

func (ti *tokenIterator) Position() int {
	return ti.pos
}
