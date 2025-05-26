// ABOUTME: Error types and handling for the compiler package
// ABOUTME: Provides structured error reporting for compilation failures

package compiler

import "fmt"

// CompilerError represents an error that occurred during compilation
type CompilerError struct {
	// Code is a unique error code for this type of error
	Code string

	// Message is the human-readable error message
	Message string

	// Path is the file path where the error occurred (if applicable)
	Path string

	// Line is the line number where the error occurred (if applicable)
	Line int

	// Column is the column number where the error occurred (if applicable)
	Column int

	// Cause is the underlying error that caused this error (if any)
	Cause error
}

// Error implements the error interface
func (e *CompilerError) Error() string {
	if e.Path != "" && e.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: [%s] %s", e.Path, e.Line, e.Column, e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *CompilerError) Unwrap() error {
	return e.Cause
}

// WithLocation adds location information to the error
func (e *CompilerError) WithLocation(path string, line, column int) *CompilerError {
	e.Path = path
	e.Line = line
	e.Column = column
	return e
}

// WithCause adds a cause to the error
func (e *CompilerError) WithCause(cause error) *CompilerError {
	e.Cause = cause
	return e
}

// NewCompilerError creates a new compiler error
func NewCompilerError(code, message string) *CompilerError {
	return &CompilerError{
		Code:    code,
		Message: message,
	}
}

// Error code constants
const (
	ErrInvalidAST        = "INVALID_AST"
	ErrCompilationFailed = "COMPILATION_FAILED"
	ErrUnsupportedNode   = "UNSUPPORTED_NODE"
	ErrTypeError         = "TYPE_ERROR"
	ErrSyntaxError       = "SYNTAX_ERROR"
)
