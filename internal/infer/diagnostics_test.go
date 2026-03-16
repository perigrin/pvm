// ABOUTME: Tests for the diagnostics types used by the PSC type inference engine.
// ABOUTME: Covers Severity, Diagnostic struct, code constants, and FormatDiagnostic.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
)

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
