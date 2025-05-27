// ABOUTME: Token-based parser following TypeScript-Go separation patterns
// ABOUTME: Consumes tokens from scanner to build AST while maintaining compatibility

package parser

import (
	"io"
	"os"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/scanner"
)

// TokenBasedParser implements Parser interface using Scanner tokens
// This provides the TypeScript-Go style separation while maintaining compatibility
type TokenBasedParser struct {
	scanner    scanner.Scanner
	fallback   Parser // Fallback to original tree-sitter parser for compatibility
	useScanner bool   // Whether to use scanner or fallback
}

// NewTokenBasedParser creates a parser that uses the scanner for tokenization
func NewTokenBasedParser(useScanner bool) (Parser, error) {
	// Create scanner
	scannerInstance, err := scanner.NewScanner(false)
	if err != nil {
		return nil, err
	}

	// Create fallback parser for compatibility
	fallbackParser, err := NewTreeSitterParser()
	if err != nil {
		return nil, err
	}

	return &TokenBasedParser{
		scanner:    scannerInstance,
		fallback:   fallbackParser,
		useScanner: useScanner,
	}, nil
}

// ParseFile implements Parser interface using tokens
func (p *TokenBasedParser) ParseFile(path string) (*AST, error) {
	if !p.useScanner {
		// Use fallback for compatibility
		return p.fallback.ParseFile(path)
	}

	// Use scanner-based approach
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{
			Message: "Failed to read file: " + err.Error(),
			Line:    0,
			Column:  0,
		}
	}

	return p.parseFromTokens(path, string(content))
}

// ParseString implements Parser interface using tokens
func (p *TokenBasedParser) ParseString(content string) (*AST, error) {
	if !p.useScanner {
		// Use fallback for compatibility
		return p.fallback.ParseString(content)
	}

	return p.parseFromTokens("", content)
}

// ParseReader implements Parser interface using tokens
func (p *TokenBasedParser) ParseReader(reader io.Reader) (*AST, error) {
	if !p.useScanner {
		// Use fallback for compatibility
		return p.fallback.ParseReader(reader)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, &ParseError{
			Message: "Failed to read from reader: " + err.Error(),
			Line:    0,
			Column:  0,
		}
	}

	return p.parseFromTokens("", string(content))
}

// parseFromTokens builds an AST from scanned tokens
func (p *TokenBasedParser) parseFromTokens(path, content string) (*AST, error) {
	// Get token iterator
	iter, err := p.scanner.ScanString(content)
	if err != nil {
		return nil, &ParseError{
			Message: "Failed to scan content: " + err.Error(),
			Line:    0,
			Column:  0,
		}
	}

	// For now, build a simplified AST from tokens
	// In future iterations, this will become a full recursive descent parser
	root := &tokenNode{
		nodeType: "root",
		tokens:   p.collectTokens(iter),
		position: ast.Position{Line: 1, Column: 1, Offset: 0},
	}

	// Extract type annotations from tokens
	annotations := p.extractTypeAnnotationsFromTokens(root.tokens, content)

	astResult := &ast.AST{
		Path:            path,
		Root:            root,
		TypeAnnotations: annotations,
		Errors:          []error{},
		Source:          content,
	}
	return (*AST)(astResult), nil
}

// collectTokens collects all tokens from the iterator
func (p *TokenBasedParser) collectTokens(iter scanner.TokenIterator) []scanner.Token {
	var tokens []scanner.Token
	for iter.HasNext() {
		token := iter.Next()
		if token.Type() == scanner.TokenEOF {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens
}

// extractTypeAnnotationsFromTokens finds type annotations in the token stream
func (p *TokenBasedParser) extractTypeAnnotationsFromTokens(tokens []scanner.Token, content string) []*ast.TypeAnnotation {
	var annotations []*ast.TypeAnnotation

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		// Look for patterns like "my Type $var" or "method name(Type $param)"
		if token.Type() == scanner.TokenMy || token.Type() == scanner.TokenMethod ||
			token.Type() == scanner.TokenField {

			annotation := p.parseTypeAnnotationFromTokens(tokens, i, content)
			if annotation != nil {
				annotations = append(annotations, annotation)
			}
		}
	}

	return annotations
}

// parseTypeAnnotationFromTokens parses a type annotation starting at the given position
func (p *TokenBasedParser) parseTypeAnnotationFromTokens(tokens []scanner.Token, start int, content string) *ast.TypeAnnotation {
	if start >= len(tokens) {
		return nil
	}

	keyword := tokens[start]

	// Simple pattern matching for "my Type $var"
	if keyword.Type() == scanner.TokenMy && start+2 < len(tokens) {
		typeToken := tokens[start+1]
		varToken := tokens[start+2]

		if typeToken.Type() == scanner.TokenIdentifier && varToken.Type() == scanner.TokenVariable {
			pos := ast.Position{
				Line:   typeToken.Position().Line,
				Column: typeToken.Position().Column,
				Offset: typeToken.Position().Offset,
			}

			return &ast.TypeAnnotation{
				AnnotatedItem:  varToken.Value(),
				TypeExpression: p.parseTypeExpressionFromToken(typeToken),
				Pos:            pos,
				Kind:           ast.VarAnnotation,
			}
		}
	}

	// TODO: Add more sophisticated pattern matching for other annotation types
	return nil
}

// parseTypeExpressionFromToken creates a TypeExpression from a token
func (p *TokenBasedParser) parseTypeExpressionFromToken(token scanner.Token) *ast.TypeExpression {
	pos := ast.Position{
		Line:   token.Position().Line,
		Column: token.Position().Column,
		Offset: token.Position().Offset,
	}

	// For now, create a simple type expression
	// This will be enhanced in future iterations to handle complex types
	return &ast.TypeExpression{
		BaseType: token.Value(),
		Pos:      pos,
	}
}

// tokenNode implements Node interface for token-based parsing
type tokenNode struct {
	nodeType string
	tokens   []scanner.Token
	position ast.Position
	children []ast.Node
	parent   ast.Node
}

func (n *tokenNode) Type() string {
	return n.nodeType
}

func (n *tokenNode) Start() ast.Position {
	return n.position
}

func (n *tokenNode) End() ast.Position {
	if len(n.tokens) > 0 {
		lastToken := n.tokens[len(n.tokens)-1]
		pos := lastToken.Position()
		return ast.Position{
			Line:   pos.Line,
			Column: pos.Column + lastToken.Length(),
			Offset: pos.Offset + lastToken.Length(),
		}
	}
	return n.position
}

func (n *tokenNode) Children() []ast.Node {
	if n.children == nil {
		n.children = []ast.Node{}
	}
	return n.children
}

func (n *tokenNode) Text() string {
	var text string
	for _, token := range n.tokens {
		text += token.Value()
	}
	return text
}

func (n *tokenNode) Parent() ast.Node {
	return n.parent
}

func (n *tokenNode) SetParent(parent ast.Node) {
	n.parent = parent
}

// NewParserWithScanner creates a new parser with scanner enabled
// This is the new API for scanner-based parsing
func NewParserWithScanner() (Parser, error) {
	return NewTokenBasedParser(true)
}

// NewCompatParser creates a parser that falls back to tree-sitter
// This maintains compatibility with existing code
func NewCompatParser() (Parser, error) {
	return NewTokenBasedParser(false)
}
