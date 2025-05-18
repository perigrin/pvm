// ABOUTME: Error handling framework for PVM Ecosystem
// ABOUTME: Provides structured error types with context and categorization

package errors

import (
	"fmt"
	"strings"
)

// Component prefixes for error codes
const (
	PrefixPVM = "PVM" // Perl Version Manager
	PrefixPVX = "PVX" // Perl Version eXecutor
	PrefixPVI = "PVI" // Perl Version Installer
	PrefixPSC = "PSC" // Perl Script Compiler
	PrefixCFG = "CFG" // Configuration
	PrefixSYS = "SYS" // System
)

// Error categories
const (
	// Configuration-related errors
	CategoryConfig = "Configuration Error"
	
	// Version-related errors
	CategoryVersion = "Version Error"
	
	// Module-related errors
	CategoryModule = "Module Error"
	
	// Execution-related errors
	CategoryExecution = "Execution Error"
	
	// Type-checking related errors
	CategoryType = "Type Error"
	
	// System-related errors
	CategorySystem = "System Error"
	
	// User input errors
	CategoryUserInput = "User Input Error"
)

// Error represents a structured error with context
type Error struct {
	// Prefix identifies the component (PVM, PVX, etc.)
	Prefix string
	
	// Category identifies the type of error
	Category string
	
	// Code is a unique identifier for this error
	Code string
	
	// Message is a short description of the error
	Message string
	
	// Detail provides additional information about the error
	Detail string
	
	// Location indicates where the error occurred (file, line, etc.)
	Location string
	
	// Hint provides a suggestion for resolving the error
	Hint string
	
	// InnerErr is the underlying error
	InnerErr error
}

// Error implements the error interface
func (e *Error) Error() string {
	var builder strings.Builder
	
	// Format the error code with the prefix
	errCode := fmt.Sprintf("%s-%s", e.Prefix, e.Code)
	
	// Start with the error code and message
	builder.WriteString(fmt.Sprintf("%s: %s", errCode, e.Message))
	
	// Add the category if set
	if e.Category != "" {
		builder.WriteString(fmt.Sprintf(" (%s)", e.Category))
	}
	
	// Add detail if present
	if e.Detail != "" {
		builder.WriteString(fmt.Sprintf("\n  Detail: %s", e.Detail))
	}
	
	// Add location if present
	if e.Location != "" {
		builder.WriteString(fmt.Sprintf("\n  Location: %s", e.Location))
	}
	
	// Add hint if present
	if e.Hint != "" {
		builder.WriteString(fmt.Sprintf("\n  Hint: %s", e.Hint))
	}
	
	// Add inner error if present and verbose is enabled
	if e.InnerErr != nil {
		builder.WriteString(fmt.Sprintf("\n  Cause: %v", e.InnerErr))
	}
	
	return builder.String()
}

// Is implements the errors.Is interface for error comparison
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	
	// Match by code and prefix
	return e.Code == t.Code && e.Prefix == t.Prefix
}

// Unwrap implements the errors.Unwrap interface for error chains
func (e *Error) Unwrap() error {
	return e.InnerErr
}

// WithDetail adds detail to the error
func (e *Error) WithDetail(detail string) *Error {
	e.Detail = detail
	return e
}

// WithLocation adds location information to the error
func (e *Error) WithLocation(location string) *Error {
	e.Location = location
	return e
}

// WithHint adds a hint for resolving the error
func (e *Error) WithHint(hint string) *Error {
	e.Hint = hint
	return e
}

// New creates a new error with the specified parameters
func New(prefix, category, code, message string, inner error) *Error {
	return &Error{
		Prefix:   prefix,
		Category: category,
		Code:     code,
		Message:  message,
		InnerErr: inner,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, prefix, category, code, message string) *Error {
	if err == nil {
		return nil
	}
	
	return &Error{
		Prefix:   prefix,
		Category: category,
		Code:     code,
		Message:  message,
		InnerErr: err,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(code, message string, inner error) *Error {
	return New(PrefixCFG, CategoryConfig, code, message, inner)
}

// NewVersionError creates a new version error
func NewVersionError(code, message string, inner error) *Error {
	return New(PrefixPVM, CategoryVersion, code, message, inner)
}

// NewModuleError creates a new module error
func NewModuleError(code, message string, inner error) *Error {
	return New(PrefixPVI, CategoryModule, code, message, inner)
}

// NewExecutionError creates a new execution error
func NewExecutionError(code, message string, inner error) *Error {
	return New(PrefixPVX, CategoryExecution, code, message, inner)
}

// NewTypeError creates a new type error
func NewTypeError(code, message string, inner error) *Error {
	return New(PrefixPSC, CategoryType, code, message, inner)
}

// NewSystemError creates a new system error
func NewSystemError(code, message string, inner error) *Error {
	return New(PrefixSYS, CategorySystem, code, message, inner)
}

// NewUserInputError creates a new user input error
func NewUserInputError(prefix, code, message string, inner error) *Error {
	return New(prefix, CategoryUserInput, code, message, inner)
}