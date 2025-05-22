// ABOUTME: Enhanced error formatting with context lines and helpful suggestions
// ABOUTME: Provides rich error display similar to modern compiler error messages

package psc

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/typechecker"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	SeverityError ErrorSeverity = iota
	SeverityWarning
	SeverityInfo
)

// EnhancedError extends TypeCheckError with additional context and suggestions
type EnhancedError struct {
	*typechecker.TypeCheckError
	Severity    ErrorSeverity
	Suggestion  string
	ContextCode []string
}

// ErrorFormatter provides enhanced error formatting capabilities
type ErrorFormatter struct {
	showContext     bool
	contextLines    int
	colorEnabled    bool
	sourceCodeCache map[string][]string
}

// NewErrorFormatter creates a new error formatter with default settings
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{
		showContext:     true,
		contextLines:    2,
		colorEnabled:    true,
		sourceCodeCache: make(map[string][]string),
	}
}

// SetContextLines sets the number of context lines to show around errors
func (ef *ErrorFormatter) SetContextLines(lines int) {
	ef.contextLines = lines
}

// SetColorEnabled enables or disables color output
func (ef *ErrorFormatter) SetColorEnabled(enabled bool) {
	ef.colorEnabled = enabled
}

// FormatError formats a single error with enhanced display
func (ef *ErrorFormatter) FormatError(err *typechecker.TypeCheckError) string {
	enhanced := ef.enhanceError(err)
	return ef.formatEnhancedError(enhanced)
}

// FormatErrors formats multiple errors with enhanced display
func (ef *ErrorFormatter) FormatErrors(errors []typechecker.TypeCheckError) string {
	if len(errors) == 0 {
		return ""
	}

	var result strings.Builder

	for i, err := range errors {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(ef.FormatError(&err))
	}

	return result.String()
}

// enhanceError converts a TypeCheckError to an EnhancedError with additional context
func (ef *ErrorFormatter) enhanceError(err *typechecker.TypeCheckError) *EnhancedError {
	enhanced := &EnhancedError{
		TypeCheckError: err,
		Severity:       SeverityError,
		Suggestion:     ef.generateSuggestion(err),
		ContextCode:    ef.getContextLines(err.Path, err.Line),
	}

	return enhanced
}

// generateSuggestion provides helpful suggestions for common type errors
func (ef *ErrorFormatter) generateSuggestion(err *typechecker.TypeCheckError) string {
	message := strings.ToLower(err.Message)

	// Type mismatch suggestions
	if strings.Contains(message, "type mismatch") {
		if strings.Contains(message, "int") && strings.Contains(message, "str") {
			return "Consider using string interpolation: \"$value\" or explicit conversion"
		}
		if strings.Contains(message, "str") && strings.Contains(message, "int") {
			return "Consider using numeric conversion: int($value) or 0 + $value"
		}
		return "Check that the assigned value matches the declared type"
	}

	// Undefined variable suggestions
	if strings.Contains(message, "undefined") || strings.Contains(message, "not found") {
		return "Make sure the variable is declared before use, or check for typos"
	}

	// Type annotation suggestions
	if strings.Contains(message, "annotation") {
		return "Review the type annotation syntax: my TypeName $variable = value;"
	}

	// Assignment suggestions
	if strings.Contains(message, "assignment") {
		return "Ensure the right side of the assignment is compatible with the declared type"
	}

	return "Review the Typed Perl documentation for more information"
}

// getContextLines retrieves source code lines around the error location
func (ef *ErrorFormatter) getContextLines(filePath string, errorLine int) []string {
	if !ef.showContext || errorLine <= 0 {
		return nil
	}

	// Get or cache the source code
	lines := ef.getCachedSourceLines(filePath)
	if lines == nil {
		return nil
	}

	// Calculate the range of lines to show
	start := errorLine - ef.contextLines - 1 // Convert to 0-based indexing
	if start < 0 {
		start = 0
	}

	end := errorLine + ef.contextLines
	if end > len(lines) {
		end = len(lines)
	}

	// Extract the context lines
	var contextLines []string
	for i := start; i < end; i++ {
		lineNum := i + 1 // Convert back to 1-based indexing
		prefix := "   "
		if lineNum == errorLine {
			prefix = ">> "
		}
		contextLines = append(contextLines, fmt.Sprintf("%s%3d: %s", prefix, lineNum, lines[i]))
	}

	return contextLines
}

// getCachedSourceLines gets source code lines from cache or reads from file
func (ef *ErrorFormatter) getCachedSourceLines(filePath string) []string {
	// Check cache first
	if lines, exists := ef.sourceCodeCache[filePath]; exists {
		return lines
	}

	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if scanner.Err() != nil {
		return nil
	}

	// Cache the lines
	ef.sourceCodeCache[filePath] = lines
	return lines
}

// formatEnhancedError formats an EnhancedError with full context and suggestions
func (ef *ErrorFormatter) formatEnhancedError(err *EnhancedError) string {
	var result strings.Builder

	// Main error line
	severity := ef.formatSeverity(err.Severity)
	result.WriteString(fmt.Sprintf("%s:%d:%d: %s %s\n",
		err.TypeCheckError.Path, err.TypeCheckError.Line, err.TypeCheckError.Column, severity, err.TypeCheckError.Message))

	// Context lines
	if len(err.ContextCode) > 0 {
		for _, line := range err.ContextCode {
			result.WriteString(line)
			result.WriteString("\n")
		}

		// Add error marker pointing to the problem
		if err.TypeCheckError.Column > 0 {
			marker := strings.Repeat(" ", 7+err.TypeCheckError.Column) + "^"
			if ef.colorEnabled {
				marker = ef.colorize(marker, "red")
			}
			result.WriteString(marker)
			result.WriteString("\n")
		}
	}

	// Helpful suggestion
	if err.Suggestion != "" {
		suggestionLine := fmt.Sprintf("   help: %s", err.Suggestion)
		if ef.colorEnabled {
			suggestionLine = ef.colorize(suggestionLine, "cyan")
		}
		result.WriteString(suggestionLine)
		result.WriteString("\n")
	}

	return result.String()
}

// formatSeverity formats the severity level with appropriate styling
func (ef *ErrorFormatter) formatSeverity(severity ErrorSeverity) string {
	switch severity {
	case SeverityError:
		text := "error:"
		if ef.colorEnabled {
			return ef.colorize(text, "red")
		}
		return text
	case SeverityWarning:
		text := "warning:"
		if ef.colorEnabled {
			return ef.colorize(text, "yellow")
		}
		return text
	case SeverityInfo:
		text := "info:"
		if ef.colorEnabled {
			return ef.colorize(text, "blue")
		}
		return text
	default:
		return "error:"
	}
}

// colorize applies ANSI color codes to text
func (ef *ErrorFormatter) colorize(text, color string) string {
	if !ef.colorEnabled {
		return text
	}

	colors := map[string]string{
		"red":    "\033[31m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"cyan":   "\033[36m",
		"reset":  "\033[0m",
	}

	if colorCode, exists := colors[color]; exists {
		return colorCode + text + colors["reset"]
	}
	return text
}
