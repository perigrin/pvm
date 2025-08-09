// ABOUTME: Error handling framework for PVM Ecosystem
// ABOUTME: Provides structured error types with context and categorization

//go:generate go run ../../scripts/generate_errors.go ../../scripts/error_definitions.json

package errors

import (
	"fmt"
	"strings"
)

// Component prefixes for error codes
const (
	PrefixPVM  = "PVM"  // Perl Version Manager
	PrefixPVX  = "PVX"  // Perl Version eXecutor
	PrefixPVI  = "PVI"  // Perl Version Installer
	PrefixPSC  = "PSC"  // Perl Script Compiler
	PrefixCFG  = "CFG"  // Configuration
	PrefixSYS  = "SYS"  // System
	PrefixDOCS = "DOCS" // Documentation
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

	// Documentation-related errors
	CategoryDocumentation = "Documentation Error"
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

// NewDocumentationError creates a new documentation error
func NewDocumentationError(code, message string, inner error) *Error {
	return New(PrefixDOCS, CategoryDocumentation, code, message, inner)
}

// ErrorSeverity defines the severity level of an error
type ErrorSeverity int

const (
	// SeverityInfo for informational messages
	SeverityInfo ErrorSeverity = iota

	// SeverityWarning for warnings that don't prevent operation
	SeverityWarning

	// SeverityError for errors that prevent operation
	SeverityError

	// SeverityCritical for critical errors that may cause system instability
	SeverityCritical
)

// String returns string representation of severity
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// EnhancedError extends Error with additional context and severity
type EnhancedError struct {
	baseError *Error

	// Severity indicates the error severity level
	severity ErrorSeverity

	// Context provides additional context about the operation
	context map[string]interface{}

	// RecoveryActions suggests possible recovery actions
	recoveryActions []string

	// RelatedErrors contains related or caused errors
	relatedErrors []*Error

	// Timestamp when the error occurred
	timestamp int64
}

// NewEnhancedError creates a new enhanced error
func NewEnhancedError(prefix, category, code, message string, inner error, severity ErrorSeverity) *EnhancedError {
	return &EnhancedError{
		baseError:       New(prefix, category, code, message, inner),
		severity:        severity,
		context:         make(map[string]interface{}),
		recoveryActions: make([]string, 0),
		relatedErrors:   make([]*Error, 0),
		timestamp:       0, // Could use time.Now().Unix() if needed
	}
}

// WithSeverity sets the error severity
func (e *EnhancedError) WithSeverity(severity ErrorSeverity) *EnhancedError {
	e.severity = severity
	return e
}

// WithContext adds context information
func (e *EnhancedError) WithContext(key string, value interface{}) *EnhancedError {
	e.context[key] = value
	return e
}

// WithRecoveryAction adds a recovery action suggestion
func (e *EnhancedError) WithRecoveryAction(action string) *EnhancedError {
	e.recoveryActions = append(e.recoveryActions, action)
	return e
}

// WithRelatedError adds a related error
func (e *EnhancedError) WithRelatedError(err *Error) *EnhancedError {
	e.relatedErrors = append(e.relatedErrors, err)
	return e
}

// Severity returns the error severity
func (e *EnhancedError) Severity() ErrorSeverity {
	return e.severity
}

// Context returns the error context
func (e *EnhancedError) Context() map[string]interface{} {
	return e.context
}

// RecoveryActions returns suggested recovery actions
func (e *EnhancedError) RecoveryActions() []string {
	return e.recoveryActions
}

// RelatedErrors returns related errors
func (e *EnhancedError) RelatedErrors() []*Error {
	return e.relatedErrors
}

// Delegate TypedError interface methods to baseError

// Category returns the error category
func (e *EnhancedError) Category() string {
	return e.baseError.Category()
}

// Code returns the error code
func (e *EnhancedError) Code() string {
	return e.baseError.Code()
}

// Prefix returns the error prefix
func (e *EnhancedError) Prefix() string {
	return e.baseError.Prefix()
}

// Message returns the error message
func (e *EnhancedError) Message() string {
	return e.baseError.Message()
}

// Description returns the error message (used for backward compatibility)
func (e *EnhancedError) Description() string {
	return e.baseError.Description()
}

// Location returns the error location
func (e *EnhancedError) Location() string {
	return e.baseError.Location()
}

// Hint returns the error hint
func (e *EnhancedError) Hint() string {
	return e.baseError.Hint()
}

// Error implements the error interface with enhanced formatting
func (e *EnhancedError) Error() string {
	var builder strings.Builder

	// Start with base error
	builder.WriteString(e.baseError.Error())

	// Add severity
	builder.WriteString(fmt.Sprintf("\n  Severity: %s", e.severity.String()))

	// Add context if present
	if len(e.context) > 0 {
		builder.WriteString("\n  Context:")
		for key, value := range e.context {
			builder.WriteString(fmt.Sprintf("\n    %s: %v", key, value))
		}
	}

	// Add recovery actions if present
	if len(e.recoveryActions) > 0 {
		builder.WriteString("\n  Suggested Actions:")
		for i, action := range e.recoveryActions {
			builder.WriteString(fmt.Sprintf("\n    %d. %s", i+1, action))
		}
	}

	// Add related errors if present
	if len(e.relatedErrors) > 0 {
		builder.WriteString("\n  Related Errors:")
		for i, relErr := range e.relatedErrors {
			builder.WriteString(fmt.Sprintf("\n    %d. %s", i+1, relErr.Error()))
		}
	}

	return builder.String()
}

// ErrorFormatter provides different formatting options for errors
type ErrorFormatter struct {
	colorEnabled bool
	verbose      bool
}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter(colorEnabled, verbose bool) *ErrorFormatter {
	return &ErrorFormatter{
		colorEnabled: colorEnabled,
		verbose:      verbose,
	}
}

// Format formats an error with the specified options
func (f *ErrorFormatter) Format(err error) string {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		return f.formatEnhanced(enhancedErr)
	}

	if typedErr, ok := err.(*Error); ok {
		return f.formatTyped(typedErr)
	}

	return err.Error()
}

func (f *ErrorFormatter) formatEnhanced(err *EnhancedError) string {
	var builder strings.Builder

	// Format with color if enabled
	if f.colorEnabled {
		switch err.severity {
		case SeverityInfo:
			builder.WriteString("\033[36m") // Cyan
		case SeverityWarning:
			builder.WriteString("\033[33m") // Yellow
		case SeverityError:
			builder.WriteString("\033[31m") // Red
		case SeverityCritical:
			builder.WriteString("\033[35m") // Magenta
		}
	}

	// Add error content
	if f.verbose {
		builder.WriteString(err.Error())
	} else {
		builder.WriteString(fmt.Sprintf("%s-%s: %s", err.Prefix(), err.Code(), err.Message()))
	}

	// Reset color if enabled
	if f.colorEnabled {
		builder.WriteString("\033[0m")
	}

	return builder.String()
}

func (f *ErrorFormatter) formatTyped(err *Error) string {
	if f.colorEnabled {
		return fmt.Sprintf("\033[31m%s\033[0m", err.Error()) // Red
	}
	return err.Error()
}

// ErrorCollector collects and categorizes multiple errors
type ErrorCollector struct {
	errors   []error
	warnings []error
	infos    []error
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors:   make([]error, 0),
		warnings: make([]error, 0),
		infos:    make([]error, 0),
	}
}

// Add adds an error to the appropriate category
func (c *ErrorCollector) Add(err error) {
	if err == nil {
		return
	}

	if enhancedErr, ok := err.(*EnhancedError); ok {
		switch enhancedErr.severity {
		case SeverityInfo:
			c.infos = append(c.infos, err)
		case SeverityWarning:
			c.warnings = append(c.warnings, err)
		default:
			c.errors = append(c.errors, err)
		}
		return
	}

	// Default to error category
	c.errors = append(c.errors, err)
}

// HasErrors returns true if any errors were collected
func (c *ErrorCollector) HasErrors() bool {
	return len(c.errors) > 0
}

// HasWarnings returns true if any warnings were collected
func (c *ErrorCollector) HasWarnings() bool {
	return len(c.warnings) > 0
}

// Errors returns all collected errors
func (c *ErrorCollector) Errors() []error {
	return c.errors
}

// Warnings returns all collected warnings
func (c *ErrorCollector) Warnings() []error {
	return c.warnings
}

// Infos returns all collected info messages
func (c *ErrorCollector) Infos() []error {
	return c.infos
}

// All returns all collected errors, warnings, and infos
func (c *ErrorCollector) All() []error {
	all := make([]error, 0, len(c.errors)+len(c.warnings)+len(c.infos))
	all = append(all, c.errors...)
	all = append(all, c.warnings...)
	all = append(all, c.infos...)
	return all
}
