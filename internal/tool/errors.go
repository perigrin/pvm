// ABOUTME: Tool-specific error handling and error types
// ABOUTME: Provides structured error handling for tool operations

package tool

import (
	"fmt"
	"strings"
)

// Error codes for tool operations
const (
	ErrToolNotFound     = "501" // Tool not found in any source
	ErrAmbiguousMode    = "502" // Cannot determine execution mode
	ErrInvalidToolName  = "503" // Tool name is invalid
	ErrConfigInvalid    = "504" // Configuration is invalid
	ErrMappingFailed    = "505" // Tool to module mapping failed
	ErrValidationFailed = "506" // Tool validation failed
	ErrInvalidMapping   = "507" // Invalid tool mapping
)

// ToolError represents a tool-specific error
type ToolError struct {
	Code        string
	Operation   string
	ToolName    string
	Message     string
	Cause       error
	Suggestions []string
}

// Error implements the error interface
func (e *ToolError) Error() string {
	if e.ToolName != "" {
		return fmt.Sprintf("tool '%s': %s", e.ToolName, e.Message)
	}
	return e.Message
}

// Unwrap returns the underlying cause of the error
func (e *ToolError) Unwrap() error {
	return e.Cause
}

// NewToolNotFoundError creates a new tool not found error
func NewToolNotFoundError(toolName string, suggestions []string) *ToolError {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("tool '%s' not found", toolName))

	if len(suggestions) > 0 {
		message.WriteString(". Did you mean:")
		for _, suggestion := range suggestions {
			message.WriteString(fmt.Sprintf("\n  - %s", suggestion))
		}
	}

	return &ToolError{
		Code:        ErrToolNotFound,
		Operation:   "resolve",
		ToolName:    toolName,
		Message:     message.String(),
		Suggestions: suggestions,
	}
}

// NewAmbiguousModeError creates a new ambiguous mode error
func NewAmbiguousModeError(input string, reason string, alternatives []string) *ToolError {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("cannot determine execution mode for '%s'", input))
	if reason != "" {
		message.WriteString(fmt.Sprintf(": %s", reason))
	}

	if len(alternatives) > 0 {
		message.WriteString("\nPossible interpretations:")
		for _, alt := range alternatives {
			message.WriteString(fmt.Sprintf("\n  - %s", alt))
		}
	}

	return &ToolError{
		Code:        ErrAmbiguousMode,
		Operation:   "detect",
		ToolName:    input,
		Message:     message.String(),
		Suggestions: alternatives,
	}
}

// NewInvalidToolNameError creates a new invalid tool name error
func NewInvalidToolNameError(toolName string, reason string) *ToolError {
	return &ToolError{
		Code:      ErrInvalidToolName,
		Operation: "validate",
		ToolName:  toolName,
		Message:   fmt.Sprintf("invalid tool name '%s': %s", toolName, reason),
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(operation string, cause error) *ToolError {
	return &ToolError{
		Code:      ErrConfigInvalid,
		Operation: operation,
		Message:   fmt.Sprintf("configuration error: %v", cause),
		Cause:     cause,
	}
}

// NewMappingError creates a new mapping error
func NewMappingError(toolName string, cause error) *ToolError {
	return &ToolError{
		Code:      ErrMappingFailed,
		Operation: "map",
		ToolName:  toolName,
		Message:   fmt.Sprintf("failed to map tool '%s' to module: %v", toolName, cause),
		Cause:     cause,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value string, reason string, suggestions []string) *ToolError {
	message := fmt.Sprintf("validation failed for %s '%s': %s", field, value, reason)

	if len(suggestions) > 0 {
		message += "\nSuggestions:"
		for _, suggestion := range suggestions {
			message += fmt.Sprintf("\n  - %s", suggestion)
		}
	}

	return &ToolError{
		Code:        ErrValidationFailed,
		Operation:   "validate",
		Message:     message,
		Suggestions: suggestions,
	}
}

// NewToolError creates a new generic tool error
func NewToolError(code string, message string) *ToolError {
	return &ToolError{
		Code:    code,
		Message: message,
	}
}
