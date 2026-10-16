// ABOUTME: Tests for the LSP server and lsp command skeleton.
// ABOUTME: Verifies server creation, document lifecycle, and command registration.

package psc_test

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
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

// exportingBlock exports block with its ($&) prototype, and hiddenBlock does
// not export it: perl compiles importSource against the first, warning
// "length() used on @ARGV" at line 4, and refuses it against the second
// ("syntax error at t.pl line 4, near "block 'inner'"").
const (
	exportingBlock = "package My::Block;\nuse Exporter 'import';\nour @EXPORT = qw(block);\nsub block($&) { 1 }\n1;\n"
	hiddenBlock    = "package My::Block;\nuse Exporter 'import';\nsub block($&) { 1 }\n1;\n"
)

// blockDiagnosed opens importSource at t.pl in dir and reports whether the
// length(@ARGV) mismatch inside block's body is diagnosed: it is only when
// My::Block, beside t.pl, exports block.
func blockDiagnosed(t *testing.T, s *psc.LSPServer, dir string) bool {
	t.Helper()
	uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(filepath.Join(dir, "t.pl"))}).String()
	require.NoError(t, s.OpenDocument(uri, []byte(importSource)))
	body := uint32(strings.Index(importSource, "block 'inner'"))
	for _, d := range s.Diagnostics(uri) {
		if d.StartByte >= body && d.Code == infer.CodeTypeMismatch {
			return true
		}
	}
	return false
}

// writeBlockModule writes source as My/Block.pm under dir.
func writeBlockModule(t *testing.T, dir, source string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "My"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "My", "Block.pm"), []byte(source), 0644))
}

// TestLSPSeesEditedModule: a module edited on disk is re-read by the next
// open of a buffer that uses it, so its new exports reach the analysis.
func TestLSPSeesEditedModule(t *testing.T) {
	isolatePerl(t)
	dir := t.TempDir()
	writeBlockModule(t, dir, hiddenBlock)
	s := psc.NewLSPServer()
	require.False(t, blockDiagnosed(t, s, dir), "control: block is not exported yet")

	writeBlockModule(t, dir, exportingBlock)
	assert.True(t, blockDiagnosed(t, s, dir), "the edited module's export reaches the buffer")
}

// TestLSPSessionRereadsModules: a module missing at the first open is found
// once it exists, and a module edited on disk is re-read; neither a stale
// module nor a cached "not found" outlives the change on disk.
func TestLSPSessionRereadsModules(t *testing.T) {
	isolatePerl(t)
	dir := t.TempDir()
	s := psc.NewLSPServer()
	require.False(t, blockDiagnosed(t, s, dir), "control: My::Block does not exist yet")

	writeBlockModule(t, dir, exportingBlock)
	assert.True(t, blockDiagnosed(t, s, dir), "a module created after the first open is found")

	writeBlockModule(t, dir, hiddenBlock)
	assert.False(t, blockDiagnosed(t, s, dir), "a module edited to drop its export is re-read")
}
