// ABOUTME: Scanner package provides lexical analysis interface following TypeScript-Go patterns
// ABOUTME: Separates tokenization concerns from parsing for better modularity and performance

//go:generate stringer -type=TokenType -output=token_type_string.go
//go:generate moq -out scanner_mock.go . Scanner Token TokenIterator

package scanner

import (
	"io"
	"os"
	"strings"

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
type Scanner interface {
	// ScanFile scans a file and returns a token channel
	ScanFile(path string) (TokenIterator, error)

	// ScanString scans a string and returns a token channel
	ScanString(content string) (TokenIterator, error)

	// ScanReader scans from a reader and returns a token channel
	ScanReader(reader io.Reader) (TokenIterator, error)
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

// treeSitterScanner implements Scanner using tree-sitter
type treeSitterScanner struct {
	parser      *treesitter.Parser
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

// ScanFile implements Scanner interface
func (s *treeSitterScanner) ScanFile(path string) (TokenIterator, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.NewSystemError("002",
			"File does not exist", err).
			WithLocation(path)
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewSystemError("003",
			"Failed to read file", err).
			WithLocation(path)
	}

	// Scan the content
	return s.ScanString(string(content))
}

// ScanString implements Scanner interface
func (s *treeSitterScanner) ScanString(content string) (TokenIterator, error) {
	// TODO: Temporary fallback - create basic tokens from simple parsing
	// The main parser now returns ast.AST which is incompatible with scanner's tree-sitter AST
	// This should be properly integrated once the scanner is updated to work with the new AST format
	tokens := s.extractTokensBasic(content)

	return s.poolManager.CreateTokenIterator(tokens), nil
}

// ScanReader implements Scanner interface
func (s *treeSitterScanner) ScanReader(reader io.Reader) (TokenIterator, error) {
	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewSystemError("005",
			"Failed to read from reader", err)
	}

	return s.ScanString(string(content))
}

// extractTokensBasic performs basic tokenization without AST analysis
func (s *treeSitterScanner) extractTokensBasic(content string) []Token {
	// Use pooled token slice for collecting tokens
	tokenSlice := s.poolManager.GetTokenSlice(64) // Start with 64 token capacity
	defer s.poolManager.ReleaseTokenSlice(tokenSlice)

	tokens := *tokenSlice
	tokens = tokens[:0] // Clear slice but keep capacity

	// Simple tokenization for basic functionality
	// This is a temporary implementation until scanner-AST integration is fixed

	// Split on whitespace but also handle operators
	parts := strings.Fields(content)

	for _, part := range parts {
		// Handle the "=" operator and ";" separately
		if strings.Contains(part, "=") {
			// Split by = and add tokens for each part
			beforeEq := strings.Split(part, "=")[0]
			if beforeEq != "" {
				tokens = append(tokens, s.createBasicToken(beforeEq))
			}
			tokens = append(tokens, s.poolManager.CreateToken(
				TokenAssign,
				"=",
				Position{Line: 1, Column: 1},
				1,
			))
			afterEq := strings.TrimSuffix(strings.Split(part, "=")[1], ";")
			if afterEq != "" {
				tokens = append(tokens, s.createBasicToken(afterEq))
			}
		} else {
			cleaned := strings.TrimSuffix(part, ";")
			if cleaned != "" {
				tokens = append(tokens, s.createBasicToken(cleaned))
			}
		}

		// Add semicolon if present
		if strings.HasSuffix(part, ";") {
			tokens = append(tokens, s.poolManager.CreateToken(
				TokenSemicolon,
				";",
				Position{Line: 1, Column: 1},
				1,
			))
		}
	}

	// Create a new slice to return (not pooled since it needs to outlive this function)
	result := make([]Token, len(tokens))
	copy(result, tokens)
	return result
}

// createBasicToken creates a token from a string value
func (s *treeSitterScanner) createBasicToken(value string) Token {
	var tokenType TokenType

	switch {
	case value == "my" || value == "our" || value == "state":
		tokenType = TokenMy
	case value == "sub":
		tokenType = TokenSub
	case value == "package":
		tokenType = TokenPackage
	case strings.HasPrefix(value, "$"):
		tokenType = TokenVariable
	case strings.ContainsAny(value, "0123456789") && !strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"):
		tokenType = TokenNumber
	default:
		tokenType = TokenIdentifier
	}

	return s.poolManager.CreateToken(
		tokenType,
		value,
		Position{Line: 1, Column: 1},
		len(value),
	)
}

// extractTokensFromAST converts tree-sitter AST nodes to tokens
func (s *treeSitterScanner) extractTokensFromAST(ast *treesitter.AST, content string) []Token {
	// Use pooled token slice for collecting tokens
	tokenSlice := s.poolManager.GetTokenSlice(128) // Start with 128 token capacity for AST
	defer s.poolManager.ReleaseTokenSlice(tokenSlice)

	tokens := *tokenSlice
	tokens = tokens[:0] // Clear slice but keep capacity

	if ast.Root != nil {
		s.extractTokensFromNode(ast.Root, content, &tokens)
	}

	// Add EOF token using pooled allocation
	tokens = append(tokens, s.poolManager.CreateToken(
		TokenEOF,
		"",
		Position{Line: len(content), Column: 1, Offset: len(content)},
		0,
	))

	// Create a new slice to return (not pooled since it needs to outlive this function)
	result := make([]Token, len(tokens))
	copy(result, tokens)
	return result
}

// extractTokensFromNode recursively extracts tokens from AST nodes
func (s *treeSitterScanner) extractTokensFromNode(node treesitter.Node, content string, tokens *[]Token) {
	nodeType := node.Type()
	nodeText := node.Text()
	startPos := node.Start()
	endPos := node.End()

	// Convert tree-sitter position to our position
	pos := Position{
		Line:   startPos.Line,
		Column: startPos.Column,
		Offset: startPos.Offset,
	}

	// Map tree-sitter node types to our token types
	tokenType := s.mapNodeTypeToTokenType(nodeType, nodeText)

	// Only create tokens for leaf nodes or significant structural nodes
	if len(node.Children()) == 0 || s.isSignificantStructuralNode(nodeType) {
		token := s.poolManager.CreateToken(
			tokenType,
			nodeText,
			pos,
			endPos.Offset-startPos.Offset,
		)
		*tokens = append(*tokens, token)
	}

	// Recursively process children
	for _, child := range node.Children() {
		s.extractTokensFromNode(child, content, tokens)
	}
}

// mapNodeTypeToTokenType maps tree-sitter node types to our token types
func (s *treeSitterScanner) mapNodeTypeToTokenType(nodeType, nodeText string) TokenType {
	switch nodeType {
	case "identifier", "varname":
		return TokenIdentifier
	case "string", "string_literal":
		return TokenString
	case "number":
		return TokenNumber
	case "scalar":
		return TokenVariable
	case "array_variable":
		return TokenArrayVariable
	case "hash_variable":
		return TokenHashVariable
	case "comment":
		return TokenComment
	case "whitespace":
		return TokenWhitespace
	case "newline":
		return TokenNewline
	case "my":
		return TokenMy
	case "our":
		return TokenOur
	case "state":
		return TokenState
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
	case "=":
		return TokenAssign
	case ",":
		return TokenComma
	case ";":
		return TokenSemicolon
	case "$":
		return TokenVariable // $ sigil is part of variable
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
	default:
		// Check text content for keywords
		switch nodeText {
		case "my":
			return TokenMy
		case "our":
			return TokenOur
		case "state":
			return TokenState
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
		case "|":
			return TokenPipe
		case "&":
			return TokenAmpersand
		case "!":
			return TokenExclamation
		default:
			return TokenIdentifier
		}
	}
}

// isSignificantStructuralNode determines if a node type should generate a token
func (s *treeSitterScanner) isSignificantStructuralNode(nodeType string) bool {
	switch nodeType {
	case "type_annotation", "type_expression", "union_type", "intersection_type", "negation_type":
		return true
	default:
		return false
	}
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
