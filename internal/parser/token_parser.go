// ABOUTME: Token-based parser following TypeScript-Go separation patterns
// ABOUTME: Consumes tokens from scanner to build AST while maintaining compatibility

package parser

import (
	"io"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/scanner"
)

// TokenBasedParser implements Parser interface using TokenExtractor
// This follows the new architecture: Parse once → Extract tokens → Process
type TokenBasedParser struct {
	parser         Parser                 // Main parser for AST creation
	tokenExtractor scanner.TokenExtractor // Token extractor for AST-based tokenization
	useTokens      bool                   // Whether to use token-based processing
}

// NewTokenBasedParser creates a parser that uses the new token extraction architecture
func NewTokenBasedParser(useTokens bool) (Parser, error) {
	// Create main parser for AST creation
	mainParser, err := NewTreeSitterParser()
	if err != nil {
		return nil, err
	}

	// Create token extractor (no parser needed - it consumes AST)
	tokenExtractor := scanner.NewTokenExtractor(nil)

	return &TokenBasedParser{
		parser:         mainParser,
		tokenExtractor: tokenExtractor,
		useTokens:      useTokens,
	}, nil
}

// ParseFile implements Parser interface using the new architecture
func (p *TokenBasedParser) ParseFile(path string) (*AST, error) {
	// Parse with main parser first to get AST
	ast, err := p.parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// If token processing is enabled, enhance AST with token information
	if p.useTokens {
		return p.enhanceASTWithTokens(ast)
	}

	return ast, nil
}

// ParseString implements Parser interface using the new architecture
func (p *TokenBasedParser) ParseString(content string) (*AST, error) {
	// Parse with main parser first to get AST
	ast, err := p.parser.ParseString(content)
	if err != nil {
		return nil, err
	}

	// If token processing is enabled, enhance AST with token information
	if p.useTokens {
		return p.enhanceASTWithTokens(ast)
	}

	return ast, nil
}

// ParseReader implements Parser interface using the new architecture
func (p *TokenBasedParser) ParseReader(reader io.Reader) (*AST, error) {
	// Parse with main parser first to get AST
	ast, err := p.parser.ParseReader(reader)
	if err != nil {
		return nil, err
	}

	// If token processing is enabled, enhance AST with token information
	if p.useTokens {
		return p.enhanceASTWithTokens(ast)
	}

	return ast, nil
}

// enhanceASTWithTokens enhances an existing AST with token information
func (p *TokenBasedParser) enhanceASTWithTokens(parsedAST *AST) (*AST, error) {
	// Convert AST to ast.AST format for token extraction
	astTree := (*ast.AST)(parsedAST)

	// Extract tokens from the AST using the new token extractor
	tokenIterator, err := p.tokenExtractor.ExtractTokens(astTree)
	if err != nil {
		// If token extraction fails, return the original AST without enhancement
		// This maintains compatibility even if token processing has issues
		return parsedAST, nil
	}

	// Collect tokens for processing
	tokens := p.collectTokens(tokenIterator)

	// Extract type annotations from tokens to enhance the AST
	if astTree.Source != "" {
		additionalAnnotations := p.extractTypeAnnotationsFromTokens(tokens, astTree.Source)

		// Merge with existing annotations (avoid duplicates)
		existingAnnotations := astTree.TypeAnnotations
		astTree.TypeAnnotations = p.mergeTypeAnnotations(existingAnnotations, additionalAnnotations)
	}

	return parsedAST, nil
}

// mergeTypeAnnotations merges additional annotations with existing ones, avoiding duplicates
func (p *TokenBasedParser) mergeTypeAnnotations(existing, additional []*ast.TypeAnnotation) []*ast.TypeAnnotation {
	if len(additional) == 0 {
		return existing
	}

	// Create a map to track existing annotations by their annotated item and position
	existingMap := make(map[string]bool)
	for _, annotation := range existing {
		key := annotation.AnnotatedItem + ":" + annotation.TypeExpression.String() +
			":" + string(rune(annotation.Pos.Line)) + ":" + string(rune(annotation.Pos.Column))
		existingMap[key] = true
	}

	// Add additional annotations that don't already exist
	result := make([]*ast.TypeAnnotation, len(existing))
	copy(result, existing)

	for _, annotation := range additional {
		key := annotation.AnnotatedItem + ":" + annotation.TypeExpression.String() +
			":" + string(rune(annotation.Pos.Line)) + ":" + string(rune(annotation.Pos.Column))
		if !existingMap[key] {
			result = append(result, annotation)
		}
	}

	return result
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

	// Simple pattern matching for "my Type $var" or "field Type $var"
	if (keyword.Type() == scanner.TokenMy || keyword.Type() == scanner.TokenField) && start+2 < len(tokens) {
		typeToken := tokens[start+1]
		varToken := tokens[start+2]

		if typeToken.Type() == scanner.TokenIdentifier && varToken.Type() == scanner.TokenVariable {
			pos := ast.Position{
				Line:   typeToken.Position().Line,
				Column: typeToken.Position().Column,
				Offset: typeToken.Position().Offset,
			}

			// Extract the full type string from the source content
			typeString := p.extractFullTypeString(tokens, start+1, content)
			typeExpr, err := ParseTypeExpression(typeString, Position{
				Line:   pos.Line,
				Column: pos.Column,
				Offset: pos.Offset,
			})
			if err != nil {
				// Fallback to simple parsing if complex parsing fails
				typeExpr = p.parseTypeExpressionFromToken(typeToken)
			}

			return &ast.TypeAnnotation{
				AnnotatedItem:  varToken.Value(),
				TypeExpression: typeExpr,
				Pos:            pos,
				Kind:           ast.VarAnnotation,
			}
		}
	}

	// TODO: Add more sophisticated pattern matching for other annotation types
	return nil
}

// extractFullTypeString extracts the complete type string including parameters from source content
func (p *TokenBasedParser) extractFullTypeString(tokens []scanner.Token, typeStart int, content string) string {
	if typeStart >= len(tokens) {
		return ""
	}

	typeToken := tokens[typeStart]

	// Find the end of the type expression by looking for the variable token
	// We need to find the type string between the type token and variable token
	if typeStart+1 >= len(tokens) {
		return typeToken.Value()
	}

	varToken := tokens[typeStart+1]

	// Extract substring from type token start to just before variable token
	typeStartOffset := typeToken.Position().Offset
	varStartOffset := varToken.Position().Offset

	if typeStartOffset >= 0 && varStartOffset > typeStartOffset && varStartOffset <= len(content) {
		typeSubstring := content[typeStartOffset:varStartOffset]
		// Trim whitespace and clean up the type string
		typeSubstring = strings.TrimSpace(typeSubstring)
		return typeSubstring
	}

	// Fallback to just the token value
	return typeToken.Value()
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
	return ast.NewTypeExpression(token.Value(), nil, pos, pos)
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

// NewParserWithTokenExtraction creates a new parser with token extraction enabled
// This uses the new architecture: Parse AST first, then extract tokens
func NewParserWithTokenExtraction() (Parser, error) {
	return NewTokenBasedParser(true)
}

// NewStandardParser creates a parser without token extraction
// This is the standard parser for cases that don't need token processing
func NewStandardParser() (Parser, error) {
	return NewTokenBasedParser(false)
}
