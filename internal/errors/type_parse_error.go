// ABOUTME: Type parsing error structures for structured error handling
// ABOUTME: Contains error types without formatting logic - presentation handled by formatters

package errors

import (
	"tamarou.com/pvm/internal/ast"
)

// TypeParseError represents a structured type parsing error
type TypeParseError struct {
	// ErrorType is the specific error classification (e.g., "InvalidUnionSyntaxError")
	ErrorType string

	// Message is the primary error message
	Message string

	// Position is the exact location where the error occurred
	Position ast.Position

	// Suggestion provides a helpful suggestion for fixing the error
	Suggestion string

	// Context describes the type expression context where the error occurred
	Context string

	// ErrorCode is a specific error code for programmatic handling
	ErrorCode TypeErrorCode

	// Source is the original source text that caused the error
	Source string

	// SourceLine is the specific line where the error occurred
	SourceLine string
}

// Error implements the error interface with Rust-style formatting
// Provides detailed error information with source context
func (tpe *TypeParseError) Error() string {
	// Use Rust-style formatter for detailed output
	formatter := NewRustStyleFormatter()
	return formatter.FormatTypeParseError(tpe, "<input>")
}

// TypeErrorCode represents specific type error categories
type TypeErrorCode int

const (
	// UnknownTypeError is for unrecognized errors
	UnknownTypeError TypeErrorCode = iota

	// MissingClosingBracketError for parameterized types
	MissingClosingBracketError

	// InvalidUnionSyntaxError for malformed union types
	InvalidUnionSyntaxError

	// IncompleteTypeAssertionError for malformed 'as' expressions
	IncompleteTypeAssertionError

	// InvalidParameterizedTypeError for malformed parameterized types
	InvalidParameterizedTypeError

	// MissingTypeNameError for empty type annotations
	MissingTypeNameError

	// InvalidWhereClauseError for malformed where constraints
	InvalidWhereClauseError

	// InvalidIntersectionSyntaxError for malformed intersection types
	InvalidIntersectionSyntaxError

	// InvalidNegationSyntaxError for malformed negation types
	InvalidNegationSyntaxError

	// DeepNestingError for excessively nested types
	DeepNestingError
)
