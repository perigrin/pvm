// ABOUTME: Tests that PathUnder recognizes a file as being inside a directory, and one that is not.
// ABOUTME: On Windows this canonicalizes both first; elsewhere it is a plain prefix check, exercised here either way.

package perltest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathUnderRecognizesAFileInsideADirectory(t *testing.T) {
	p := Perl542(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "sub", "module.pm")
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
	require.NoError(t, os.WriteFile(file, []byte("1;\n"), 0o644))

	assert.True(t, PathUnder(t, p, file, dir))
}

func TestPathUnderRejectsAFileOutsideADirectory(t *testing.T) {
	p := Perl542(t)
	dir := t.TempDir()
	elsewhere := t.TempDir()
	file := filepath.Join(elsewhere, "module.pm")
	require.NoError(t, os.WriteFile(file, []byte("1;\n"), 0o644))

	assert.False(t, PathUnder(t, p, file, dir))
}
