// ABOUTME: Rust-style error formatter for CLI presentation
// ABOUTME: Converts structured errors into user-friendly Rust compiler-style messages

package errors

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// RustStyleFormatter formats errors in Rust compiler style
type RustStyleFormatter struct {
	// ShowHelp controls whether to show helpful notes and suggestions
	ShowHelp bool
	// ShowContext controls whether to show source context around errors
	ShowContext bool
}

// NewRustStyleFormatter creates a new Rust-style error formatter
func NewRustStyleFormatter() *RustStyleFormatter {
	return &RustStyleFormatter{
		ShowHelp:    true,
		ShowContext: true,
	}
}

// FormatTypeParseError formats a TypeParseError in Rust style
func (rsf *RustStyleFormatter) FormatTypeParseError(err *TypeParseError, filePath string) string {
	var result strings.Builder

	// Main error header with specific error code
	result.WriteString(fmt.Sprintf("error[TSP%03d]: %s", int(err.ErrorCode)+1, err.Message))

	if rsf.ShowContext && err.SourceLine != "" {
		result.WriteString("\n")
		result.WriteString(rsf.formatErrorLocation(err.Position, filePath))
		result.WriteString(rsf.formatSourceContext(err.SourceLine, err.Position))
	}

	// Add suggestion if available
	if rsf.ShowHelp && err.Suggestion != "" {
		result.WriteString("\n")
		result.WriteString(fmt.Sprintf("help: %s\n", err.Suggestion))
	}

	return result.String()
}

// FormatTreeSitterError formats generic tree-sitter errors in Rust style
func (rsf *RustStyleFormatter) FormatTreeSitterError(errNodes []ErrorNodeInfo, filePath, sourceContent string) string {
	var result strings.Builder

	// Split source into lines for context display
	lines := strings.Split(sourceContent, "\n")

	result.WriteString(fmt.Sprintf("error[TSP001]: parse error (%d ERROR nodes detected)\n", len(errNodes)))

	for i, errNode := range errNodes {
		if i > 0 {
			result.WriteString("\n")
		}

		lineNum := errNode.StartPoint.Line + 1  // Convert 0-based to 1-based
		colNum := errNode.StartPoint.Column + 1 // Convert 0-based to 1-based

		position := ast.Position{Line: lineNum, Column: colNum}
		result.WriteString(rsf.formatErrorLocation(position, filePath))

		// Show the problematic line with context
		if errNode.StartPoint.Line < len(lines) {
			line := lines[errNode.StartPoint.Line]
			result.WriteString(rsf.formatSourceContext(line, position))
			result.WriteString(rsf.formatErrorPointer(errNode, line))
		}
	}

	if rsf.ShowHelp {
		result.WriteString("\n")
		result.WriteString("note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.\n")
		result.WriteString("      Please add test cases for this syntax to improve parser coverage.\n")
	}

	return result.String()
}

// formatErrorLocation formats the location pointer line
func (rsf *RustStyleFormatter) formatErrorLocation(position ast.Position, filePath string) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("  --> %s:%d:%d\n", filePath, position.Line, position.Column))
	result.WriteString("   |\n")
	return result.String()
}

// formatSourceContext formats the source line with line number
func (rsf *RustStyleFormatter) formatSourceContext(sourceLine string, position ast.Position) string {
	return fmt.Sprintf("%2d | %s\n", position.Line, sourceLine)
}

// formatErrorPointer creates the pointer line pointing to the error
func (rsf *RustStyleFormatter) formatErrorPointer(errNode ErrorNodeInfo, line string) string {
	var result strings.Builder
	result.WriteString("   | ")

	// Add spacing to align with error location
	for i := 0; i < errNode.StartPoint.Column; i++ {
		if i < len(line) && line[i] == '\t' {
			result.WriteString("\t")
		} else {
			result.WriteString(" ")
		}
	}

	// Add underline for the error span
	errorLen := errNode.EndPoint.Column - errNode.StartPoint.Column
	if errorLen <= 0 {
		errorLen = 1
	}
	for i := 0; i < errorLen; i++ {
		result.WriteString("^")
	}

	result.WriteString(fmt.Sprintf(" unexpected token: '%s'\n", errNode.Content))
	return result.String()
}

// ErrorNodeInfo represents information about a tree-sitter error node
// This duplicates the struct from parser package to avoid circular imports
type ErrorNodeInfo struct {
	StartPoint struct {
		Line   int
		Column int
	}
	EndPoint struct {
		Line   int
		Column int
	}
	Content string
}
