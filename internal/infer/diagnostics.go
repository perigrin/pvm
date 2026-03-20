// ABOUTME: Diagnostic types for the PSC type inference engine.
// ABOUTME: Provides Severity, Diagnostic struct, code constants, and FormatDiagnostic.

package infer

import (
	"bytes"
	"fmt"

	"tamarou.com/pvm/internal/types"
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
	StartByte  uint32
	EndByte    uint32
	Severity   Severity
	Message    string
	Code       string // machine-readable identifier, e.g. "arity-mismatch"
	Suggestion string // guard suggestion text, e.g. "Add guard: if (defined($x)) { ... }"
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
// When Suggestion is non-empty, a hint line is appended.
func FormatDiagnostic(filename string, source []byte, d Diagnostic) string {
	line, col := byteOffsetToLineCol(source, d.StartByte)
	result := fmt.Sprintf("%s:%d:%d: %s: %s [%s]", filename, line, col, d.Severity, d.Message, d.Code)
	if d.Suggestion != "" {
		result += "\n  hint: " + d.Suggestion
	}
	return result
}

// byteOffsetToLineCol converts a byte offset into 1-based line and column
// numbers by counting newlines in the prefix of source up to the offset.
// Column is byte-based (not character or UTF-16 code unit count),
// matching tree-sitter's byte offset model.
func byteOffsetToLineCol(source []byte, offset uint32) (line, col int) {
	if int(offset) > len(source) {
		offset = uint32(len(source))
	}
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

// SuggestGuard returns a guard suggestion string for a type mismatch where
// actual is the variable's current type and expected is what the callee
// requires. Returns empty string when no guard can help.
//
// Each candidate guard is tried in priority order. A guard is only suggested
// if applying its narrowing to actual would make the result compatible with
// expected. Compatibility is checked two ways:
//   - types.TypeSatisfies(narrowed, expected) — standard subtype/polymorphic check
//   - types.IsSubtype(expected, narrowed) — "expected fits within narrowed",
//     handles cases where narrowed is a broad union that lost polymorphic status
//     (e.g. Scalar &^ Undef is not in polymorphicMasks but still contains Int)
//
// Priority order (first match wins):
//  1. defined() guard — narrows by removing Undef bit
//  2. ref() guard — narrows to Ref mask
//  3. builtin::is_bool() guard — narrows to Bool
//  4. No match → empty string
func SuggestGuard(varName string, actual, expected types.Type) string {
	if varName == "" {
		return ""
	}

	// Priority 1: defined() — would removing Undef make actual compatible?
	if actual&types.Undef != 0 && expected&types.Undef == 0 {
		narrowed := actual &^ types.Undef
		if guardNarrowingSatisfies(narrowed, expected) {
			return "Add guard: if (defined(" + varName + ")) { ... }"
		}
	}

	// Priority 2: ref() — would narrowing to Ref make actual compatible?
	if actual&types.Ref != 0 {
		narrowed := actual & types.Ref
		if guardNarrowingSatisfies(narrowed, expected) {
			return "Add guard: if (ref(" + varName + ")) { ... }"
		}
	}

	// Priority 3: builtin::is_bool() — would narrowing to Bool help?
	if actual&types.Bool != 0 {
		narrowed := actual & types.Bool
		if guardNarrowingSatisfies(narrowed, expected) {
			return "Add guard: if (builtin::is_bool(" + varName + ")) { ... }"
		}
	}

	return ""
}

// guardNarrowingSatisfies checks whether a narrowed type is compatible with
// the expected type. It uses TypeSatisfies first (handles polymorphic types
// and standard subtype checks), then falls back to checking if expected is
// a subtype of narrowed. The fallback handles cases like Scalar &^ Undef:
// a broad union that lost its polymorphic status in polymorphicMasks but
// still contains the expected type's bits (e.g. Int is a subtype of
// Scalar &^ Undef).
func guardNarrowingSatisfies(narrowed, expected types.Type) bool {
	if types.TypeSatisfies(narrowed, expected) {
		return true
	}
	return types.IsSubtype(expected, narrowed)
}
