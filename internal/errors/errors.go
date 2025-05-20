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

// TypedError is an interface for errors that provide additional type information
type TypedError interface {
	error
	Category() string
	Code() string
	Prefix() string
	Message() string
	Description() string
	Location() string
	Hint() string
}

// Error represents a structured error with context
type Error struct {
	// Prefix identifies the component (PVM, PVX, etc.)
	prefix string

	// Category identifies the type of error
	category string

	// Code is a unique identifier for this error
	code string

	// Message is a short description of the error
	message string

	// Detail provides additional information about the error
	detail string

	// Location indicates where the error occurred (file, line, etc.)
	location string

	// Hint provides a suggestion for resolving the error
	hint string

	// InnerErr is the underlying error
	innerErr error
}

// Error implements the error interface
func (e *Error) Error() string {
	var builder strings.Builder

	// Format the error code with the prefix
	errCode := fmt.Sprintf("%s-%s", e.prefix, e.code)

	// Start with the error code and message
	builder.WriteString(fmt.Sprintf("%s: %s", errCode, e.message))

	// Add the category if set
	if e.category != "" {
		builder.WriteString(fmt.Sprintf(" (%s)", e.category))
	}

	// Add detail if present
	if e.detail != "" {
		builder.WriteString(fmt.Sprintf("\n  Detail: %s", e.detail))
	}

	// Add location if present
	if e.location != "" {
		builder.WriteString(fmt.Sprintf("\n  Location: %s", e.location))
	}

	// Add hint if present
	if e.hint != "" {
		builder.WriteString(fmt.Sprintf("\n  Hint: %s", e.hint))
	}

	// Add inner error if present and verbose is enabled
	if e.innerErr != nil {
		builder.WriteString(fmt.Sprintf("\n  Cause: %v", e.innerErr))
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
	return e.code == t.code && e.prefix == t.prefix
}

// Unwrap implements the errors.Unwrap interface for error chains
func (e *Error) Unwrap() error {
	return e.innerErr
}

// WithDetail adds detail to the error
func (e *Error) WithDetail(detail string) *Error {
	e.detail = detail
	return e
}

// WithLocation adds location information to the error
func (e *Error) WithLocation(location string) *Error {
	e.location = location
	return e
}

// WithHint adds a hint for resolving the error
func (e *Error) WithHint(hint string) *Error {
	e.hint = hint
	return e
}

// Implementations of the TypedError interface

// Category returns the error category
func (e *Error) Category() string {
	return e.category
}

// Code returns the error code
func (e *Error) Code() string {
	return e.code
}

// Prefix returns the error prefix
func (e *Error) Prefix() string {
	return e.prefix
}

// Message returns the error message
func (e *Error) Message() string {
	return e.message
}

// Description returns the error message (used for backward compatibility)
func (e *Error) Description() string {
	return e.message
}

// Location returns the error location
func (e *Error) Location() string {
	return e.location
}

// Hint returns the error hint
func (e *Error) Hint() string {
	return e.hint
}

// New creates a new error with the specified parameters
func New(prefix, category, code, message string, inner error) *Error {
	return &Error{
		prefix:   prefix,
		category: category,
		code:     code,
		message:  message,
		innerErr: inner,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, prefix, category, code, message string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		prefix:   prefix,
		category: category,
		code:     code,
		message:  message,
		innerErr: err,
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