// ABOUTME: Tests for the psc parse command.
// ABOUTME: Covers file parsing, the CST and lowered trees, tree and sexpr output, and error cases.

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

	// internal/parse's root and statement, not tree-sitter's
	// (source_file (variable_declaration_expression ...)).
	output := out.String()
	assert.True(t, strings.HasPrefix(output, `(source_file (statement (declaration "my"`), "sexpr output: %s", output)
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

// runParse runs `psc parse` with args over a file holding src.
func runParse(t *testing.T, src string, args ...string) (string, error) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "test.pl")
	require.NoError(t, os.WriteFile(file, []byte(src), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs(append(append([]string{"parse"}, args...), file))
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	return out.String(), err
}

// TestParseCommandLayers: --tree picks internal/parse's CST or the lowered
// tree, in either format, and neither speaks tree-sitter's vocabulary.
func TestParseCommandLayers(t *testing.T) {
	src := "my ($x) = f(); $h->{k}; print $fh \"x\";\n"
	layers := map[string][]string{
		"cst":     {"declaration", "index", "call", "paren", "arrow", "handle"},
		"lowered": {"Decl", "Key", "Call", "ListAssign", "Arrow", "Handle"},
	}
	for tree, want := range layers {
		for _, format := range []string{"tree", "sexpr"} {
			out, err := runParse(t, src, "--tree", tree, "--format", format)
			require.NoError(t, err)
			for _, w := range want {
				assert.Contains(t, out, w, "--tree %s --format %s", tree, format)
			}
			assert.NotContains(t, out, "variable_declaration", "--tree %s --format %s", tree, format)
		}
	}
}

// TestParseCommandRejectsUnknownTree: a --tree outside the two is an error
// naming both.
func TestParseCommandRejectsUnknownTree(t *testing.T) {
	_, err := runParse(t, "1;\n", "--tree", "hir")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hir")
	assert.Contains(t, err.Error(), "cst")
	assert.Contains(t, err.Error(), "lowered")
}

// TestParseCommandRejectsUnknownFormat: a --format outside the two is an
// error naming both, as --tree's is; each of the two is accepted, the
// control.
func TestParseCommandRejectsUnknownFormat(t *testing.T) {
	_, err := runParse(t, "1;\n", "--format", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "json")
	assert.Contains(t, err.Error(), "tree")
	assert.Contains(t, err.Error(), "sexpr")

	for _, format := range []string{"tree", "sexpr"} {
		_, err := runParse(t, "1;\n", "--format", format)
		assert.NoError(t, err, format)
	}
}

// TestParseCommandOnBrokenFile: on a file perl rejects, the CST shows the
// refusal with its code -- here a catch-all, so an Error node -- the lowered
// tree a Hole, and both exit 0.
func TestParseCommandOnBrokenFile(t *testing.T) {
	src := "my $x = );\n"
	cst, err := runParse(t, src, "--tree", "cst")
	require.NoError(t, err)
	assert.Contains(t, cst, "error refusal=not_a_term")

	lowered, err := runParse(t, src, "--tree", "lowered")
	require.NoError(t, err)
	assert.Contains(t, lowered, "Hole")
}

// runPSC runs psc with args and returns what it prints to stdout.
func runPSC(t *testing.T, args ...string) string {
	t.Helper()
	cmd := psc.NewCommand()
	cmd.SetArgs(args)
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&strings.Builder{})
	require.NoError(t, cmd.Execute())
	return out.String()
}

// TestParseCommandResolvesImports: psc parse and psc analyze parse a file
// as psc check does, resolving its use statements, so they show the tree
// check analysed. My::Block beside t.pl exports block with a ($&)
// prototype, so `block 'inner' => sub {...}` is a call perl compiles, and
// the `require` in its body is one t.pl loads. Without My::Block perl
// refuses the call, and the control shows the Hole the parser leaves.
func TestParseCommandResolvesImports(t *testing.T) {
	isolatePerl(t)
	source := "use My::Block;\nblock 'inner' => sub { require Data::Dumper; };\n"
	for _, c := range []struct {
		module string
		hole   bool
	}{{exportingBlock, false}, {"", true}} {
		dir := t.TempDir()
		file := filepath.Join(dir, "t.pl")
		require.NoError(t, os.WriteFile(file, []byte(source), 0644))
		if c.module != "" {
			writeBlockModule(t, dir, c.module)
		}

		lowered := runPSC(t, "parse", "--tree", "lowered", file)
		assert.Equal(t, c.hole, strings.Contains(lowered, "Hole"), "module %q:\n%s", c.module, lowered)
		deps := runPSC(t, "analyze", file)
		assert.Equal(t, !c.hole, strings.Contains(deps, "Data::Dumper"), "module %q:\n%s", c.module, deps)
	}
}
