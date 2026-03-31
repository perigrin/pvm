// ABOUTME: Tests for the LSP server and lsp command skeleton.
// ABOUTME: Verifies server creation, document lifecycle, and command registration.

package psc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/psc"
)

func TestNewLSPServer(t *testing.T) {
	s := psc.NewLSPServer()
	require.NotNil(t, s, "NewLSPServer should return a non-nil server")
}

func TestLSPServerOpenDocument(t *testing.T) {
	s := psc.NewLSPServer()

	source := []byte("my $x = 42;\n")
	err := s.OpenDocument("file:///test.pl", source)
	require.NoError(t, err, "OpenDocument should not return an error for valid Perl")
}

func TestLSPServerOpenDocumentReplaces(t *testing.T) {
	s := psc.NewLSPServer()

	err := s.OpenDocument("file:///test.pl", []byte("my $x = 1;\n"))
	require.NoError(t, err)

	// Open the same URI again with different content
	err = s.OpenDocument("file:///test.pl", []byte("my $x = 42;\n"))
	require.NoError(t, err, "re-opening a document should not return an error")
}

func TestLSPServerCloseDocument(t *testing.T) {
	s := psc.NewLSPServer()

	err := s.OpenDocument("file:///test.pl", []byte("my $x = 42;\n"))
	require.NoError(t, err)

	// Closing should not panic or error
	s.CloseDocument("file:///test.pl")
}

func TestLSPServerCloseNonExistent(t *testing.T) {
	s := psc.NewLSPServer()
	// Closing a document that was never opened should be a no-op
	s.CloseDocument("file:///never-opened.pl")
}

func TestLSPCommandExists(t *testing.T) {
	cmd := psc.NewCommand()
	require.NotNil(t, cmd)

	lspCmd, _, err := cmd.Find([]string{"lsp"})
	require.NoError(t, err)
	require.NotNil(t, lspCmd)
	assert.Equal(t, "lsp", lspCmd.Name())
}
