// ABOUTME: Enhanced parser implementation that integrates type error identification
// ABOUTME: Wraps tree-sitter parser with structured error analysis and classification

package parser

import (
	"io"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// enhancedParser wraps the tree-sitter parser with type error identification
type enhancedParser struct {
	tsParser        *treesitter.Parser
	errorIdentifier *TypeErrorIdentifier
}

// NewEnhancedParser creates a new enhanced parser with type error identification
func NewEnhancedParser() (Parser, error) {
	tsParser, err := treesitter.NewParser(false) // false for non-debug mode
	if err != nil {
		return nil, err
	}

	return &enhancedParser{
		tsParser:        tsParser,
		errorIdentifier: NewTypeErrorIdentifier(),
	}, nil
}

// ParseFile implements the Parser interface with enhanced error handling
func (ep *enhancedParser) ParseFile(path string) (*ast.AST, error) {
	ast, err := ep.tsParser.ParseFile(path)
	if err != nil {
		// Check if this is a type-related error we can identify more specifically
		if typeErr := ep.tryIdentifyTypeError(err, "", "file parsing"); typeErr != nil {
			return nil, typeErr
		}
		// Return original error if not identified as type error
		return nil, err
	}
	return ast, nil
}

// ParseString implements the Parser interface with enhanced error handling
func (ep *enhancedParser) ParseString(content string) (*ast.AST, error) {
	ast, err := ep.tsParser.ParseString(content)
	if err != nil {
		// Check if this is a type-related error we can identify more specifically
		if typeErr := ep.tryIdentifyTypeError(err, content, "string parsing"); typeErr != nil {
			return nil, typeErr
		}
		// Return original error if not identified as type error
		return nil, err
	}
	return ast, nil
}

// ParseReader implements the Parser interface with enhanced error handling
func (ep *enhancedParser) ParseReader(reader io.Reader) (*ast.AST, error) {
	ast, err := ep.tsParser.ParseReader(reader)
	if err != nil {
		// For reader parsing, we don't have the content to analyze
		// Return original error
		return nil, err
	}
	return ast, nil
}

// tryIdentifyTypeError attempts to identify type-specific errors from tree-sitter failures
func (ep *enhancedParser) tryIdentifyTypeError(originalErr error, source, context string) *errors.TypeParseError {
	// Only try to identify if we have source content and it looks like a type-related error
	if source == "" {
		return nil
	}

	// Check if the error appears to be type-related
	if !ep.isLikelyTypeError(originalErr, source) {
		return nil
	}

	// Use error identifier to classify the error
	typeErr := ep.errorIdentifier.IdentifyTreeSitterError(source, originalErr, context)
	if typeParseErr, ok := typeErr.(*errors.TypeParseError); ok {
		return typeParseErr
	}

	return nil
}

// isLikelyTypeError checks if an error and source suggest this is a type-related parsing error
func (ep *enhancedParser) isLikelyTypeError(err error, source string) bool {
	// Check error message for indicators
	errStr := err.Error()
	if strings.Contains(errStr, "parse error") || strings.Contains(errStr, "ERROR nodes") {
		// Check source for type-like patterns
		return containsTypeLikePattern(source)
	}
	return false
}
