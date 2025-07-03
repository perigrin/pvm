// ABOUTME: Enhanced parser implementation that integrates type error identification
// ABOUTME: Wraps tree-sitter parser with structured error analysis and classification

package parser

import (
	"io"
	"regexp"
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
	astResult, err := ep.tsParser.ParseFile(path)
	if err != nil {
		// Check if this is a type-related error we can identify more specifically
		if typeErr := ep.tryIdentifyTypeError(err, "", "file parsing"); typeErr != nil {
			return nil, typeErr
		}
		// Return original error if not identified as type error
		return nil, err
	}

	// TODO: Re-enable post-parsing AST validation once error identification is more precise
	// Currently this is too aggressive and flags valid parses as errors
	// Even if parsing succeeded, check for malformed types in the AST
	// This catches cases where tree-sitter partially parses malformed syntax
	// if astResult != nil && astResult.Root != nil {
	//	// For file parsing, we need the content to analyze malformed types
	//	content := ""
	//	if astResult.Source != "" {
	//		content = astResult.Source
	//	}
	//	if malformedErr := ep.errorIdentifier.IdentifyMalformedTypeInAST(astResult.Root, content); malformedErr != nil {
	//		return nil, malformedErr
	//	}
	// }

	return astResult, nil
}

// ParseString implements the Parser interface with enhanced error handling
func (ep *enhancedParser) ParseString(content string) (*ast.AST, error) {
	astResult, err := ep.tsParser.ParseString(content)
	if err != nil {
		// Check if this is a type-related error we can identify more specifically
		if typeErr := ep.tryIdentifyTypeError(err, content, "string parsing"); typeErr != nil {
			return nil, typeErr
		}
		// Return original error if not identified as type error
		return nil, err
	}

	// TODO: Re-enable post-parsing AST validation once error identification is more precise
	// Currently this is too aggressive and flags valid parses as errors
	// Even if parsing succeeded, check for malformed types in the AST
	// This catches cases where tree-sitter partially parses malformed syntax
	// if astResult != nil && astResult.Root != nil {
	//	if malformedErr := ep.errorIdentifier.IdentifyMalformedTypeInAST(astResult.Root, content); malformedErr != nil {
	//		return nil, malformedErr
	//	}
	// }

	return astResult, nil
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
	isTypeError := ep.isLikelyTypeError(originalErr, source)

	if !isTypeError {
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
		// Check if this looks like a complete Perl file rather than just a type expression
		if ep.looksLikeCompleteFile(source) {
			// Don't classify complete file parse failures as type errors
			// These should be handled as general parsing failures
			return false
		}

		// Check source for type-like patterns, but be more specific
		// Don't consider package-qualified variables as type errors
		if strings.Contains(source, "::") && strings.Contains(source, "$") &&
			(strings.HasPrefix(source, "our ") || strings.HasPrefix(source, "my ") ||
				strings.HasPrefix(source, "state ") || strings.HasPrefix(source, "local ")) {
			// This looks like a package-qualified variable declaration, not a type error
			return false
		}

		// Check for actual type annotation patterns
		return containsActualTypePattern(source)
	}
	return false
}

// looksLikeCompleteFile checks if the source appears to be a complete Perl file
// rather than just a type expression fragment
func (ep *enhancedParser) looksLikeCompleteFile(source string) bool {
	// Trim leading/trailing whitespace for analysis
	trimmed := strings.TrimSpace(source)

	// Check for clear indicators of complete files
	completeFilePatterns := []string{
		"package ",     // Package declarations
		"use v",        // Version pragmas
		"use strict",   // Common pragmas
		"use warnings", // Common pragmas
	}

	for _, pattern := range completeFilePatterns {
		if strings.HasPrefix(trimmed, pattern) {
			return true
		}
	}

	// Check if it contains multiple lines with Perl statements
	lines := strings.Split(trimmed, "\n")
	statementCount := 0
	typeExpressionCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Check if this looks like a type expression (malformed or otherwise)
		if strings.HasPrefix(line, "my ") && containsTypePatterns(line) {
			typeExpressionCount++
		}

		// Count lines that look like Perl statements
		if strings.HasSuffix(line, ";") ||
			strings.HasSuffix(line, "{") ||
			strings.HasSuffix(line, "}") ||
			strings.HasPrefix(line, "package ") ||
			strings.HasPrefix(line, "use ") ||
			strings.HasPrefix(line, "sub ") ||
			strings.HasPrefix(line, "method ") ||
			strings.HasPrefix(line, "class ") ||
			strings.HasPrefix(line, "role ") {
			statementCount++
		}
	}

	// If most lines look like type expressions, it's probably a type fragment, not a complete file
	if typeExpressionCount >= 3 && typeExpressionCount >= statementCount/2 {
		return false
	}

	// If we have multiple statement-like lines, it's probably a complete file
	return statementCount >= 2
}

// containsTypePatterns checks if a line contains type-related patterns
func containsTypePatterns(line string) bool {
	typePatterns := []string{
		"ArrayRef[", "HashRef[", "Container[", "Map[", "Wrapper[",
		"|", "&", "!", // Type operators
		"[", "]", // Brackets for parameterized types
	}

	for _, pattern := range typePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}

// containsActualTypePattern checks for patterns that indicate actual type annotations
func containsActualTypePattern(source string) bool {
	// Look for specific type annotation patterns
	patterns := []string{
		" as ",      // Type assertions
		"ArrayRef[", // Parameterized types
		"HashRef[",  // Parameterized types
		"Maybe[",    // Parameterized types
		"|",         // Union types (but be careful of bitwise or)
		"&",         // Intersection types (but be careful of bitwise and)
		"!Undef",    // Negation types
		"returns ",  // Method return types
		"field ",    // Field declarations with types
	}

	for _, pattern := range patterns {
		if strings.Contains(source, pattern) {
			return true
		}
	}

	// Check if it's a typed variable declaration (my Type $var)
	// This regex looks for: <keyword> <CapitalizedWord> <$var>
	if match, _ := regexp.MatchString(`\b(my|our|state|field)\s+[A-Z]\w*\s+\$`, source); match {
		return true
	}

	return false
}
