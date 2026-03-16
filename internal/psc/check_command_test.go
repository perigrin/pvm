// ABOUTME: Tests for the psc check command.
// ABOUTME: Covers clean files, files with diagnostics, directory walking, and argument validation.

package psc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/psc"
)

func TestCheckCommandExists(t *testing.T) {
	cmd := psc.NewCommand()
	require.NotNil(t, cmd)

	checkCmd, _, err := cmd.Find([]string{"check"})
	require.NoError(t, err)
	require.NotNil(t, checkCmd)
	assert.Equal(t, "check", checkCmd.Name())
}

func TestCheckCommandCleanFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "clean.pl")
	content := "my @arr;\npush(@arr, 1);\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", file})

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.NoError(t, err, "check of clean file should return no error")
	assert.Empty(t, stderr.String(), "no diagnostics should be printed for a clean file")
}

func TestCheckCommandWithDiagnostics(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "bad.pl")
	// push() with no arguments triggers an arity-mismatch diagnostic.
	content := "push();\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", file})

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Error(t, err, "check of file with diagnostics should return an error")
	assert.Contains(t, stderr.String(), "arity", "stderr should mention arity for push() with no args")
}

func TestCheckCommandDirectory(t *testing.T) {
	dir := t.TempDir()

	// Write a clean Perl file.
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "clean.pl"),
		[]byte("my @arr;\npush(@arr, 1);\n"),
		0644,
	))
	// Write a non-Perl file that should be ignored.
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "README.txt"),
		[]byte("This is a readme.\n"),
		0644,
	))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", dir})

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.NoError(t, err, "check of clean directory should return no error")
}

func TestCheckCommandMissingPath(t *testing.T) {
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", "/nonexistent/path/to/file.pl"})

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Error(t, err, "check of nonexistent path should return an error")
}

func TestCheckCommandNoArgs(t *testing.T) {
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check"})

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Error(t, err, "check with no arguments should return an error")
}
