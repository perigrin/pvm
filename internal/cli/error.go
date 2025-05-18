// ABOUTME: Error handling for CLI commands
// ABOUTME: Provides consistent error formatting and handling

package cli

import (
	"fmt"
	"os"
)

// Error component prefixes
const (
	PrefixPVM = "PVM"
	PrefixPVX = "PVX"
	PrefixPVI = "PVI"
	PrefixPSC = "PSC"
	PrefixCFG = "CFG"
	PrefixSYS = "SYS"
)

// Error categories
const (
	CategoryConfig    = "Configuration Error"
	CategoryVersion   = "Version Error"
	CategoryModule    = "Module Error"
	CategoryExecution = "Execution Error"
	CategoryType      = "Type Error"
	CategorySystem    = "System Error"
	CategoryUserInput = "User Input Error"
)

// Error represents a cli error with additional context
type Error struct {
	Prefix    string
	Category  string
	Code      string
	Message   string
	Detail    string
	Location  string
	Hint      string
	InnerErr  error
}

// NewError creates a new CLI error
func NewError(prefix, category, code, message string, err error) *Error {
	return &Error{
		Prefix:   prefix,
		Category: category,
		Code:     code,
		Message:  message,
		InnerErr: err,
	}
}

// WithDetail adds detail information to the error
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

// Error implements the error interface
func (e *Error) Error() string {
	prefix := fmt.Sprintf("%s-%s", e.Prefix, e.Code)
	result := fmt.Sprintf("%s: %s", prefix, e.Message)
	
	if e.Detail != "" {
		result += fmt.Sprintf("\n  Detail: %s", e.Detail)
	}
	
	if e.Location != "" {
		result += fmt.Sprintf("\n  Location: %s", e.Location)
	}
	
	if e.Hint != "" {
		result += fmt.Sprintf("\n  Hint: %s", e.Hint)
	}
	
	if e.InnerErr != nil && Verbose {
		result += fmt.Sprintf("\n  Cause: %v", e.InnerErr)
	}
	
	return result
}

// Unwrap implements the unwrap interface for error chains
func (e *Error) Unwrap() error {
	return e.InnerErr
}

// HandleError handles a CLI error appropriately based on flags
func HandleError(err error) {
	if err == nil {
		return
	}
	
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}