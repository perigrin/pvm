// ABOUTME: Tests for LSP server inference integration — annotations, diagnostics, symbols.
// ABOUTME: Verifies that OpenDocument triggers inference and query methods return correct data.

package psc_test

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

	// Re-open with source that passes a wrong type to a builtin:
	// push(@arr, $scalar) is fine, but push($x, 1) with a scalar $x is a
	// type-mismatch.
	badSource := []byte("my $x;\npush($x, 1);\n")
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

// TestLSPHoverThroughPositionMap: the parens of `($x)` are erased by the
// lowering, a bit on the Var they wrap, so hover on either paren resolves
// the type of $x through the position map (§6.1.6 item 3).
func TestLSPHoverThroughPositionMap(t *testing.T) {
	s := psc.NewLSPServer()
	source := "my $x = 42;\nmy $y = ($x);\n"
	require.NoError(t, s.OpenDocument("file:///test.pl", []byte(source)))

	open := uint32(strings.Index(source, "("))
	shut := uint32(strings.Index(source, ")"))
	for _, off := range []uint32{open, open + 1, shut} {
		typ, ok := s.TypeAtByte("file:///test.pl", off)
		require.True(t, ok, "offset %d: no type", off)
		assert.Equal(t, types.Int, typ, "offset %d", off)
	}
}

// TestLSPHoverInsideHoleReportsNothing: `(;` is a Hole perl rejects, typed
// Any and never inspected (§4.14.2, Hole row), so hover on it answers
// nothing rather than the Hole's Any or the enclosing declaration's type.
func TestLSPHoverInsideHoleReportsNothing(t *testing.T) {
	s := psc.NewLSPServer()
	source := "my $x = (;\n"
	require.NoError(t, s.OpenDocument("file:///test.pl", []byte(source)))

	typ, ok := s.TypeAtByte("file:///test.pl", uint32(strings.Index(source, "(")))
	assert.False(t, ok, "hover inside a Hole reported %s", typ)
}

// TestLSPEditToBrokenSourceKeepsServing: replacing a clean document with
// source perl rejects still opens, and hover, definition and diagnostics
// answer at every offset without a panic.
func TestLSPEditToBrokenSourceKeepsServing(t *testing.T) {
	s := psc.NewLSPServer()
	require.NoError(t, s.OpenDocument("file:///test.pl", []byte("my $x = 42;\n")))

	broken := "if ($x) {"
	require.NoError(t, s.OpenDocument("file:///test.pl", []byte(broken)))
	for off := uint32(0); off <= uint32(len(broken)); off++ {
		s.TypeAtByte("file:///test.pl", off)
		s.DefinitionAtByte("file:///test.pl", off)
	}
	s.Diagnostics("file:///test.pl")
	assert.Equal(t, []byte(broken), s.Source("file:///test.pl"))
}

// TestLSPDiagnosticsMatchGolden opens every golden fixture internal/infer
// holds, at its own path, and asserts the LSP reports exactly the golden's
// diagnostics: the ones psc check prints for it (TestCheckCommandOnLoweredTree).
// It runs isolated from the host's perl, as the goldens are made.
func TestLSPDiagnosticsMatchGolden(t *testing.T) {
	isolatePerl(t)
	fixtures, err := filepath.Glob(filepath.Join("..", "infer", "testdata", "lowered", "*", "*.pl"))
	require.NoError(t, err)
	require.NotEmpty(t, fixtures)

	s := psc.NewLSPServer()
	for _, fx := range fixtures {
		source, err := os.ReadFile(fx)
		require.NoError(t, err)
		golden, err := os.ReadFile(strings.TrimSuffix(fx, ".pl") + ".golden")
		require.NoError(t, err)
		abs, err := filepath.Abs(fx)
		require.NoError(t, err)
		uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}).String()

		require.NoError(t, s.OpenDocument(uri, source), fx)
		got := []string{}
		for _, d := range s.Diagnostics(uri) {
			got = append(got, fmt.Sprintf("diag %d %d %s %s", d.StartByte, d.EndByte, d.Code, d.Message))
		}
		sort.Strings(got)
		want := []string{}
		for _, line := range strings.Split(string(golden), "\n") {
			if strings.HasPrefix(line, "diag ") {
				want = append(want, line)
			}
		}
		assert.Equal(t, want, got, fx)
	}
}

// TestLSPDefinitionOfCalledSub: a call names its sub, and definition at the
// call finds the sub's declaration.
func TestLSPDefinitionOfCalledSub(t *testing.T) {
	s := psc.NewLSPServer()
	source := "sub greet { return 1; }\ngreet();\n"
	require.NoError(t, s.OpenDocument("file:///test.pl", []byte(source)))

	sym, ok := s.DefinitionAtByte("file:///test.pl", uint32(strings.LastIndex(source, "greet")))
	require.True(t, ok, "DefinitionAtByte should find greet at its call")
	assert.Equal(t, "greet", sym.Name)
}

// TestLSPDefinitionOfElement: $x[0] reads @x and $h{a} reads %h, as do the
// slices @x[0,1] and @h{a}, so definition on their base finds the aggregate
// even when a scalar of the same name is declared. $r->[0] reads $r, the
// control: an arrow subscript's base is the scalar holding the reference.
func TestLSPDefinitionOfElement(t *testing.T) {
	s := psc.NewLSPServer()
	source := "my $x = 'a'; my @x = (1); my $h = 'b'; my %h = (a => 1); my $r = [];\n" +
		"print $x[0]; print $h{a}; print @x[0,1]; print @h{a}; print $r->[0];\n"
	require.NoError(t, s.OpenDocument("file:///test.pl", []byte(source)))

	body := strings.Index(source, "\n")
	for _, c := range []struct{ at, want string }{
		{"$x[0]", "@x"}, {"$h{a}", "%h"}, {"@x[0,1]", "@x"}, {"@h{a}", "%h"}, {"$r->[0]", "$r"},
	} {
		off := uint32(body + strings.Index(source[body:], c.at))
		sym, ok := s.DefinitionAtByte("file:///test.pl", off)
		if assert.True(t, ok, "%s: no definition", c.at) {
			assert.Equal(t, c.want, sym.Name, c.at)
		}
	}
}
