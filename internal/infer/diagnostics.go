// ABOUTME: Diagnostic types for the PSC type inference engine.
// ABOUTME: Provides Severity, Diagnostic struct, code constants, and FormatDiagnostic.

package infer

import (
	"bytes"
	"fmt"
)

// Severity represents the importance level of a diagnostic message.
type Severity int

const (
	// Error indicates a definite problem that prevents correct analysis.
	Error Severity = iota
	// Warning indicates a potential problem that may not be an error.
	Warning
	// Info indicates a purely informational diagnostic.
	Info
)

// String returns the lowercase name of the severity level.
func (s Severity) String() string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warning"
	case Info:
		return "info"
	default:
		return "unknown"
	}
}

// Diagnostic holds a single analysis finding with its source location.
type Diagnostic struct {
	StartByte uint32
	EndByte   uint32
	Severity  Severity
	Message   string
	Code      string // machine-readable identifier, e.g. "arity-mismatch"
}

// Machine-readable diagnostic code constants.
const (
	CodeArityMismatch  = "arity-mismatch"
	CodeTypeMismatch   = "type-mismatch"
	CodeUnknownBuiltin = "unknown-builtin"
)

// FormatDiagnostic converts byte offsets to 1-based line:col positions using
// the source text and returns a human-readable diagnostic string.
//
// Output format: filename:line:col: severity: message [code]
func FormatDiagnostic(filename string, source []byte, d Diagnostic) string {
	line, col := byteOffsetToLineCol(source, d.StartByte)
	return fmt.Sprintf("%s:%d:%d: %s: %s [%s]", filename, line, col, d.Severity, d.Message, d.Code)
}

// byteOffsetToLineCol converts a byte offset into 1-based line and column
// numbers by counting newlines in the prefix of source up to the offset.
func byteOffsetToLineCol(source []byte, offset uint32) (line, col int) {
	prefix := source[:offset]
	line = bytes.Count(prefix, []byte{'\n'}) + 1
	lastNewline := bytes.LastIndexByte(prefix, '\n')
	if lastNewline < 0 {
		col = int(offset) + 1
	} else {
		col = int(offset) - lastNewline
	}
	return line, col
}
