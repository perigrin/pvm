// ABOUTME: Tests for LSP server inference integration — annotations, diagnostics, symbols.
// ABOUTME: Verifies that OpenDocument triggers inference and query methods return correct data.

package psc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/psc"
	"tamarou.com/pvm/internal/types"
)

func TestLSPOpenDocumentTriggersInference(t *testing.T) {
	s := psc.NewLSPServer()

	err := s.OpenDocument("file:///test.pl", []byte("my $x = 42;\n"))
	require.NoError(t, err)

	annotations := s.Annotations("file:///test.pl")
	require.NotNil(t, annotations, "Annotations should be non-nil after opening a document")
}

func TestLSPDocumentChangeDiagnostics(t *testing.T) {
	s := psc.NewLSPServer()

	// Open clean source — no type errors expected.
	cleanSource := []byte("my $x = 42;\n")
	err := s.OpenDocument("file:///test.pl", cleanSource)
	require.NoError(t, err)

	cleanDiags := s.Diagnostics("file:///test.pl")
	assert.Empty(t, cleanDiags, "clean source should have no diagnostics")

	// Re-open with source that passes a wrong type to a builtin.
	// push(@arr, $scalar) is fine; but push() with zero args triggers arity error.
	badSource := []byte("push();\n")
	err = s.OpenDocument("file:///test.pl", badSource)
	require.NoError(t, err)

	updatedDiags := s.Diagnostics("file:///test.pl")
	assert.NotEmpty(t, updatedDiags, "bad source should produce diagnostics")
}

func TestLSPTypeAtByte(t *testing.T) {
	s := psc.NewLSPServer()

	// "42;\n" — the number literal starts at offset 0.
	err := s.OpenDocument("file:///test.pl", []byte("42;\n"))
	require.NoError(t, err)

	typ, ok := s.TypeAtByte("file:///test.pl", 0)
	require.True(t, ok, "TypeAtByte should find a type at offset 0")
	assert.Equal(t, types.Int, typ, "42 should be inferred as Int")
}

func TestLSPTypeAtByteVariable(t *testing.T) {
	s := psc.NewLSPServer()

	// "my @arr;\n" — @arr node starts at byte 3 (after "my ").
	err := s.OpenDocument("file:///test.pl", []byte("my @arr;\n"))
	require.NoError(t, err)

	// Offset 3 is the '@' sigil of @arr.
	typ, ok := s.TypeAtByte("file:///test.pl", 3)
	require.True(t, ok, "TypeAtByte should find a type for @arr")
	assert.Equal(t, types.Array, typ, "@arr should be inferred as Array")
}

func TestLSPSymbolTable(t *testing.T) {
	s := psc.NewLSPServer()

	source := []byte("sub greet { return 1; }\n")
	err := s.OpenDocument("file:///test.pl", source)
	require.NoError(t, err)

	st := s.SymbolTable("file:///test.pl")
	require.NotNil(t, st, "SymbolTable should be non-nil after opening a document")

	sym, found := st.Lookup("greet")
	require.True(t, found, "symbol table should contain 'greet'")
	assert.Equal(t, types.Code, sym.Type, "sub 'greet' should have type Code")
}

func TestLSPCloseDocumentClearsInference(t *testing.T) {
	s := psc.NewLSPServer()

	err := s.OpenDocument("file:///test.pl", []byte("my $x = 42;\n"))
	require.NoError(t, err)

	// Verify annotations are present before closing.
	require.NotNil(t, s.Annotations("file:///test.pl"), "should have annotations before close")

	s.CloseDocument("file:///test.pl")

	assert.Nil(t, s.Annotations("file:///test.pl"), "Annotations should be nil after close")
	assert.Nil(t, s.Diagnostics("file:///test.pl"), "Diagnostics should be nil after close")
	assert.Nil(t, s.SymbolTable("file:///test.pl"), "SymbolTable should be nil after close")
}

func TestLSPAnnotationsNilForUnknownURI(t *testing.T) {
	s := psc.NewLSPServer()

	assert.Nil(t, s.Annotations("file:///not-opened.pl"),
		"Annotations for a never-opened URI should be nil")
	assert.Nil(t, s.Diagnostics("file:///not-opened.pl"),
		"Diagnostics for a never-opened URI should be nil")
	assert.Nil(t, s.SymbolTable("file:///not-opened.pl"),
		"SymbolTable for a never-opened URI should be nil")
}

func TestLSPTypeAtByteUnknownURI(t *testing.T) {
	s := psc.NewLSPServer()

	_, ok := s.TypeAtByte("file:///not-opened.pl", 0)
	assert.False(t, ok, "TypeAtByte should return false for a never-opened URI")
}

func TestLSPDefinitionAtByte(t *testing.T) {
	s := psc.NewLSPServer()

	// "my $x = 42;\n" — $x is declared starting at byte 3.
	source := []byte("my $x = 42;\n")
	err := s.OpenDocument("file:///test.pl", source)
	require.NoError(t, err)

	sym, ok := s.DefinitionAtByte("file:///test.pl", 3)
	require.True(t, ok, "DefinitionAtByte should find $x at offset 3")
	assert.Equal(t, "$x", sym.Name)
}

func TestLSPDiagnosticSuggestionField(t *testing.T) {
	server := psc.NewLSPServer()
	// push($x, 1) where $x is Scalar triggers type-mismatch.
	// Verifies the Suggestion field exists on the Diagnostic struct
	// and is accessible via the LSP API. No guard helps here (Scalar
	// cannot be narrowed to Array), so Suggestion is empty.
	source := []byte("my $x;\npush($x, 1);\n")
	err := server.OpenDocument("file:///test.pl", source)
	require.NoError(t, err)

	diags := server.Diagnostics("file:///test.pl")
	require.True(t, len(diags) > 0, "should have diagnostics")

	var found bool
	for _, d := range diags {
		if d.Code == infer.CodeTypeMismatch {
			// Scalar vs Array: empty suggestion is correct.
			assert.Empty(t, d.Suggestion)
			found = true
		}
	}
	assert.True(t, found, "should find type-mismatch diagnostic via LSP")
}
