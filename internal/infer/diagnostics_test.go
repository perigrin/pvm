// ABOUTME: Tests for the diagnostics types used by the PSC type inference engine.
// ABOUTME: Covers Severity, Diagnostic struct, code constants, and FormatDiagnostic.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
)

func TestSuggestGuardDefined(t *testing.T) {
	// Scalar actual vs Int expected. Scalar includes Undef; Int does not.
	// Removing Undef from Scalar leaves Bool|Str|DualVar|Regex|Ref, which
	// still contains Int (IsSubtype(Int, narrowed) = true), so defined()
	// is suggested.
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Int)
	assert.Equal(t, "Add guard: if (defined($x)) { ... }", suggestion)
}

func TestSuggestGuardRef(t *testing.T) {
	// Scalar actual, Ref expected — defined() fires first because removing
	// Undef from Scalar still contains Ref (IsSubtype(Ref, Scalar&^Undef)).
	// The defined() guard is less invasive than ref().
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Ref)
	assert.Equal(t, "Add guard: if (defined($x)) { ... }", suggestion)
}

func TestSuggestGuardRefDirect(t *testing.T) {
	// Ref|Int actual (no Undef), Ref expected — P1 skips (no Undef).
	// P2: narrowed = Ref. guardNarrowingSatisfies(Ref, Ref) = true.
	suggestion := infer.SuggestGuard("$x", types.Ref|types.Int, types.Ref)
	assert.Equal(t, "Add guard: if (ref($x)) { ... }", suggestion)
}

func TestSuggestGuardNoMatchStructural(t *testing.T) {
	// Scalar actual, Array expected — no guard can make Scalar satisfy Array.
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Array)
	assert.Equal(t, "", suggestion)
}

func TestSuggestGuardNoMatchSubtype(t *testing.T) {
	// Array vs Hash — structural mismatch, no suggestion.
	suggestion := infer.SuggestGuard("@arr", types.Array, types.Hash)
	assert.Equal(t, "", suggestion)
}

func TestSuggestGuardEmptyVarName(t *testing.T) {
	suggestion := infer.SuggestGuard("", types.Scalar, types.Ref)
	assert.Equal(t, "", suggestion)
}

func TestSuggestGuardObject(t *testing.T) {
	// Scalar actual, Object expected — defined() fires because removing
	// Undef from Scalar still contains Object (IsSubtype(Object, narrowed)).
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Object)
	assert.Equal(t, "Add guard: if (defined($x)) { ... }", suggestion)
}

func TestSuggestGuardBool(t *testing.T) {
	// Scalar actual, Bool expected — defined() fires first because
	// Scalar &^ Undef still contains Bool.
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Bool)
	assert.Equal(t, "Add guard: if (defined($x)) { ... }", suggestion)
}

func TestSuggestGuardBoolDirect(t *testing.T) {
	// Bool|Int actual, Bool expected — no Undef bit, so P1 skips.
	// P2 (ref): no Ref bits, skips. P3 (is_bool): Bool & (Bool|Int) = Bool,
	// guardNarrowingSatisfies(Bool, Bool) = true.
	suggestion := infer.SuggestGuard("$x", types.Bool|types.Int, types.Bool)
	assert.Equal(t, "Add guard: if (builtin::is_bool($x)) { ... }", suggestion)
}

func TestDiagnosticCreation(t *testing.T) {
	d := infer.Diagnostic{
		StartByte: 10,
		EndByte:   20,
		Severity:  infer.Error,
		Message:   "too many arguments",
		Code:      infer.CodeArityMismatch,
	}

	assert.Equal(t, uint32(10), d.StartByte)
	assert.Equal(t, uint32(20), d.EndByte)
	assert.Equal(t, infer.Error, d.Severity)
	assert.Equal(t, "too many arguments", d.Message)
	assert.Equal(t, infer.CodeArityMismatch, d.Code)
}

func TestSeverityString(t *testing.T) {
	assert.Equal(t, "error", infer.Error.String())
	assert.Equal(t, "warning", infer.Warning.String())
	assert.Equal(t, "info", infer.Info.String())
}

func TestDiagnosticCodes(t *testing.T) {
	assert.Equal(t, "arity-mismatch", infer.CodeArityMismatch)
	assert.Equal(t, "type-mismatch", infer.CodeTypeMismatch)
	assert.Equal(t, "unknown-builtin", infer.CodeUnknownBuiltin)
}

func TestFormatDiagnostic(t *testing.T) {
	// "abc\n" — 'a' is at byte 0, column 1 of line 1
	source := []byte("abc\ndef\n")

	d := infer.Diagnostic{
		StartByte: 0,
		EndByte:   3,
		Severity:  infer.Error,
		Message:   "something went wrong",
		Code:      infer.CodeTypeMismatch,
	}

	result := infer.FormatDiagnostic("foo.pl", source, d)
	require.Equal(t, "foo.pl:1:1: error: something went wrong [type-mismatch]", result)
}

func TestFormatDiagnosticMultiLine(t *testing.T) {
	// source: "abc\ndef\nghi\n"
	// line 1: bytes 0-3   (a=0, b=1, c=2, \n=3)
	// line 2: bytes 4-7   (d=4, e=5, f=6, \n=7)
	// line 3: bytes 8-11  (g=8, h=9, i=10, \n=11)
	source := []byte("abc\ndef\nghi\n")

	t.Run("line 2 column 1", func(t *testing.T) {
		d := infer.Diagnostic{
			StartByte: 4,
			EndByte:   7,
			Severity:  infer.Warning,
			Message:   "unused variable",
			Code:      infer.CodeTypeMismatch,
		}
		result := infer.FormatDiagnostic("script.pl", source, d)
		assert.Equal(t, "script.pl:2:1: warning: unused variable [type-mismatch]", result)
	})

	t.Run("line 2 column 2", func(t *testing.T) {
		d := infer.Diagnostic{
			StartByte: 5,
			EndByte:   7,
			Severity:  infer.Info,
			Message:   "consider renaming",
			Code:      infer.CodeUnknownBuiltin,
		}
		result := infer.FormatDiagnostic("script.pl", source, d)
		assert.Equal(t, "script.pl:2:2: info: consider renaming [unknown-builtin]", result)
	})

	t.Run("line 3 column 1", func(t *testing.T) {
		d := infer.Diagnostic{
			StartByte: 8,
			EndByte:   11,
			Severity:  infer.Error,
			Message:   "arity error",
			Code:      infer.CodeArityMismatch,
		}
		result := infer.FormatDiagnostic("script.pl", source, d)
		assert.Equal(t, "script.pl:3:1: error: arity error [arity-mismatch]", result)
	})
}

func TestFormatDiagnosticEdgeCases(t *testing.T) {
	t.Run("offset beyond source length", func(t *testing.T) {
		source := []byte("abc\n")
		d := infer.Diagnostic{StartByte: 100, EndByte: 110, Severity: infer.Error, Message: "bad offset", Code: "test"}
		// Should not panic; clamps to end of source
		result := infer.FormatDiagnostic("test.pl", source, d)
		assert.Contains(t, result, "test.pl:")
	})

	t.Run("empty source", func(t *testing.T) {
		d := infer.Diagnostic{StartByte: 0, EndByte: 0, Severity: infer.Warning, Message: "empty", Code: "test"}
		result := infer.FormatDiagnostic("test.pl", []byte{}, d)
		assert.Equal(t, "test.pl:1:1: warning: empty [test]", result)
	})

	t.Run("offset at newline boundary", func(t *testing.T) {
		source := []byte("abc\ndef\n")
		d := infer.Diagnostic{StartByte: 3, EndByte: 4, Severity: infer.Info, Message: "at newline", Code: "test"}
		result := infer.FormatDiagnostic("test.pl", source, d)
		// Byte 3 is the \n character, which is column 4 of line 1
		assert.Equal(t, "test.pl:1:4: info: at newline [test]", result)
	})

	t.Run("offset at end of source", func(t *testing.T) {
		source := []byte("abc\ndef\nghi\n")
		d := infer.Diagnostic{StartByte: 12, EndByte: 12, Severity: infer.Info, Message: "at end", Code: "test"}
		result := infer.FormatDiagnostic("test.pl", source, d)
		// Byte 12 is past the last \n — line 4, col 1
		assert.Equal(t, "test.pl:4:1: info: at end [test]", result)
	})
}

func TestSeverityStringUnknown(t *testing.T) {
	assert.Equal(t, "unknown", infer.Severity(99).String())
}
