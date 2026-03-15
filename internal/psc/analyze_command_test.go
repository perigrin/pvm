// ABOUTME: Tests for the psc analyze command.
// ABOUTME: Covers single file analysis, directory walking, and use/require extraction.

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

func TestAnalyzeCommandExists(t *testing.T) {
	cmd := psc.NewCommand()
	require.NotNil(t, cmd)

	analyzeCmd, _, err := cmd.Find([]string{"analyze"})
	require.NoError(t, err)
	require.NotNil(t, analyzeCmd)
	assert.Equal(t, "analyze", analyzeCmd.Name())
}

func TestAnalyzeCommandSingleFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.pl")
	content := "use strict;\nuse warnings;\nmy $x = 1;\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"analyze", file})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	require.NoError(t, err)

	// The output should mention the dependencies found
	output := out.String()
	t.Logf("analyze output: %q", output)
	// Grammar may not fully parse 'use strict' due to known gotreesitter limitations,
	// but the analyze command should still report what it finds.
	assert.Contains(t, output, "strict")
	assert.Contains(t, output, "warnings")
}

func TestAnalyzeCommandDirectory(t *testing.T) {
	dir := t.TempDir()

	// Write a few Perl files
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "main.pl"),
		[]byte("use strict;\nmy $x = 1;\n"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "Module.pm"),
		[]byte("package Module;\n1;\n"),
		0644,
	))
	// Write a non-Perl file that should be ignored
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "README.txt"),
		[]byte("This is a readme.\n"),
		0644,
	))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"analyze", dir})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	require.NoError(t, err, "analyze directory should not return an error")
}

func TestAnalyzeCommandMissingPath(t *testing.T) {
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"analyze", "/nonexistent/path"})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	assert.Error(t, err, "analyze of missing path should return an error")
}

func TestAnalyzeCommandNoArgs(t *testing.T) {
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"analyze"})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	assert.Error(t, err, "analyze with no argument should return an error")
}
