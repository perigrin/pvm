//go:build !windows

// ABOUTME: Tests that psc check prints a file name's control characters escaped, never raw to the terminal.
// ABOUTME: A cloned repository's file names are untrusted; an escape sequence in one must not reach the terminal.

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

// TestCheckEscapesControlCharacters checks that a diagnostic in a file
// named with an OSC title sequence prints the name with \xNN escapes, and
// that the note naming an unresolvable .perl-version is escaped the same
// way. The control is a plain name, printed as it is. A name holding a
// byte that is not UTF-8 is TestPrintableEscapes' case: APFS refuses to
// create one.
func TestCheckEscapesControlCharacters(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "p\x1bq")
	require.NoError(t, os.Mkdir(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".perl-version"), []byte("dev\n"), 0o644))
	bad := "my $u = undef;\nmy $x = $u + 1;\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "x\x1b]0;PWNED\a.pl"), []byte(bad), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plain.pl"), []byte(bad), 0o644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", dir})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	_ = cmd.Execute()

	got := out.String()
	assert.NotContains(t, got, "\x1b")
	assert.NotContains(t, got, "\a")
	assert.Contains(t, got, `p\x1bq/x\x1b]0;PWNED\x07.pl:2:`)
	assert.Contains(t, got, `p\x1bq/.perl-version names "dev"`)
	assert.Contains(t, got, `p\x1bq/plain.pl:2:`)
}
