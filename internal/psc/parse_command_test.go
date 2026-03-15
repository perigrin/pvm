// ABOUTME: Tests for the psc parse command.
// ABOUTME: Covers file parsing, tree output, sexpr output, and error cases.

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

func TestParseCommandExists(t *testing.T) {
	cmd := psc.NewCommand()
	require.NotNil(t, cmd)

	// Verify parse subcommand exists
	parseCmd, _, err := cmd.Find([]string{"parse"})
	require.NoError(t, err)
	require.NotNil(t, parseCmd)
	assert.Equal(t, "parse", parseCmd.Name())
}

func TestParseCommandTree(t *testing.T) {
	// Write a temp Perl file
	dir := t.TempDir()
	file := filepath.Join(dir, "test.pl")
	require.NoError(t, os.WriteFile(file, []byte("my $x = 42;\n"), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"parse", file})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, out.String(), "parse output should not be empty")
}

func TestParseCommandSExpr(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.pl")
	require.NoError(t, os.WriteFile(file, []byte("my $x = 42;\n"), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"parse", "--format", "sexpr", file})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "source_file", "sexpr output should contain source_file")
}

func TestParseCommandMissingFile(t *testing.T) {
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"parse", "/nonexistent/path/missing.pl"})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	assert.Error(t, err, "parsing a missing file should return an error")
}

func TestParseCommandNoArgs(t *testing.T) {
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"parse"})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	assert.Error(t, err, "parse with no file argument should return an error")
}
