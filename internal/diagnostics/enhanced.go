// ABOUTME: Enhanced diagnostics system leveraging symbol information for better error reporting
// ABOUTME: Provides symbol-aware error messages, undefined variable detection, and usage tracking

//go:generate go run ../../scripts/generate_diagnostics.go ../../scripts/diagnostic_definitions.json

package diagnostics

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
)

// DiagnosticKind represents the type of diagnostic
type DiagnosticKind int

const (
	DiagnosticError DiagnosticKind = iota
	DiagnosticWarning
	DiagnosticInfo
	DiagnosticHint
)

// String returns the string representation of diagnostic kind
func (dk DiagnosticKind) String() string {
	switch dk {
	case DiagnosticError:
		return "error"
	case DiagnosticWarning:
		return "warning"
	case DiagnosticInfo:
		return "info"
	case DiagnosticHint:
		return "hint"
	default:
		return "unknown"
	}
}

// Diagnostic represents an enhanced diagnostic with symbol context
type Diagnostic struct {
	// Basic diagnostic information
	Kind    DiagnosticKind
	Message string
	Pos     ast.Position

	// Symbol context
	Symbol     *binder.Symbol
	SymbolName string
	SymbolKind binder.SymbolKind

	// Location context
	FilePath string
	LineText string

	// Enhancement context
	Suggestion     string
	RelatedSymbols []*binder.Symbol
	DidYouMean     []string

	// Code information
	Code        string
	HelpMessage string
}

// UNUSED: EnhancedDiagnosticEngine was part of advanced diagnostics but never integrated
// TODO: Remove completely once LSP dependencies are confirmed not to need these types

// UNUSED: Advanced diagnostics methods removed - these were never integrated

// FormatDiagnostic formats a diagnostic for display
func (d *Diagnostic) FormatDiagnostic(colorEnabled bool) string {
	var builder strings.Builder

	// Error location
	builder.WriteString(fmt.Sprintf("%s:%d:%d: ", d.FilePath, d.Pos.Line, d.Pos.Column))

	// Kind with color
	kindStr := d.Kind.String() + ":"
	if colorEnabled {
		switch d.Kind {
		case DiagnosticError:
			kindStr = "\033[31m" + kindStr + "\033[0m"
		case DiagnosticWarning:
			kindStr = "\033[33m" + kindStr + "\033[0m"
		case DiagnosticInfo:
			kindStr = "\033[34m" + kindStr + "\033[0m"
		case DiagnosticHint:
			kindStr = "\033[36m" + kindStr + "\033[0m"
		}
	}
	builder.WriteString(kindStr + " ")

	// Message
	builder.WriteString(d.Message)
	if d.Code != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", d.Code))
	}
	builder.WriteString("\n")

	// Line text with pointer
	if d.LineText != "" {
		builder.WriteString(fmt.Sprintf("   %d | %s\n", d.Pos.Line, d.LineText))

		// Error pointer
		pointer := strings.Repeat(" ", 6+d.Pos.Column) + "^"
		if colorEnabled {
			pointer = "\033[31m" + pointer + "\033[0m"
		}
		builder.WriteString(pointer + "\n")
	}

	// Suggestion
	if d.Suggestion != "" {
		suggestionLine := "   help: " + d.Suggestion
		if colorEnabled {
			suggestionLine = "\033[36m" + suggestionLine + "\033[0m"
		}
		builder.WriteString(suggestionLine + "\n")
	}

	// Help message
	if d.HelpMessage != "" {
		builder.WriteString("   note: " + d.HelpMessage + "\n")
	}

	// Did you mean suggestions
	if len(d.DidYouMean) > 0 {
		builder.WriteString("   note: Did you mean: " + strings.Join(d.DidYouMean, ", ") + "\n")
	}

	return builder.String()
}
