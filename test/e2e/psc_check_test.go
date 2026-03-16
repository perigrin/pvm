// ABOUTME: End-to-end integration tests for the PSC type inference pipeline.
// ABOUTME: Tests the psc check command and infer.Analyze API against real Perl source files.

package e2e

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/psc"
	"tamarou.com/pvm/internal/types"
)

// testdataDir returns the absolute path to the check testdata directory.
// It resolves relative to the current file's location so the path is correct
// regardless of which directory `go test` is invoked from.
func testdataDir(t *testing.T) string {
	t.Helper()

	// Find the project root via go.mod, then navigate to the testdata directory.
	dir, err := os.Getwd()
	require.NoError(t, err, "getwd must succeed")

	// Walk up to find go.mod
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			break
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "could not find project root (no go.mod)")
		dir = parent
	}

	return filepath.Join(dir, "test", "e2e", "testdata", "check")
}

// runPSCCheck invokes the psc check command against target (file or directory)
// via the Go API and returns (stdout, stderr, error).
func runPSCCheck(t *testing.T, target string) (string, string, error) {
	t.Helper()

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", target})

	var stdout, stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

// TestPSCCheckCleanFile verifies that psc check exits successfully (no diagnostics)
// when run against a valid Perl file with no type errors.
func TestPSCCheckCleanFile(t *testing.T) {
	cleanFile := filepath.Join(testdataDir(t), "clean.pl")
	_, statErr := os.Stat(cleanFile)
	require.NoError(t, statErr, "testdata/check/clean.pl must exist")

	_, stderr, err := runPSCCheck(t, cleanFile)
	assert.NoError(t, err, "psc check on a clean file should return no error")
	assert.Empty(t, stderr, "psc check on a clean file should produce no diagnostic output")
}

// TestPSCCheckDiagnostics verifies that psc check returns an error and emits
// arity-mismatch diagnostics when run against a file containing type errors.
func TestPSCCheckDiagnostics(t *testing.T) {
	diagFile := filepath.Join(testdataDir(t), "diagnostics.pl")
	_, statErr := os.Stat(diagFile)
	require.NoError(t, statErr, "testdata/check/diagnostics.pl must exist")

	_, stderr, err := runPSCCheck(t, diagFile)
	assert.Error(t, err, "psc check on a file with errors should return a non-zero exit")
	assert.Contains(t, stderr, "arity-mismatch",
		"stderr should contain an arity-mismatch diagnostic for push() with no arguments")
}

// TestTypeInferenceAccuracy exercises infer.Analyze directly to verify that
// specific Perl constructs are inferred with the expected types.
func TestTypeInferenceAccuracy(t *testing.T) {
	p := parser.New()

	cases := []struct {
		name     string
		source   string
		fragment string // substring of source whose node type we check
		want     types.Type
	}{
		{
			name:     "integer literal",
			source:   "42;",
			fragment: "42",
			want:     types.Int,
		},
		{
			name:     "float literal",
			source:   "3.14;",
			fragment: "3.14",
			want:     types.Num,
		},
		{
			name:     "array variable",
			source:   "my @arr; @arr;",
			fragment: "@arr",
			want:     types.Array,
		},
		{
			name:     "push returns Int",
			source:   "my @arr; push(@arr, 1);",
			fragment: "push(@arr, 1)",
			want:     types.Int,
		},
		{
			name:     "keys returns List",
			source:   "my %hash; keys(%hash);",
			fragment: "keys(%hash)",
			want:     types.List,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			tree, err := p.Parse(src)
			require.NoError(t, err, "parse must succeed for %q", tc.source)

			annotations, _, _ := infer.Analyze(tree, src)

			// Search the annotation map for a node whose source text starts with
			// the expected fragment.
			fragBytes := []byte(tc.fragment)
			fragLen := uint32(len(fragBytes))
			var found bool
			for offset, typ := range annotations {
				if offset+fragLen > uint32(len(src)) {
					continue
				}
				if string(src[offset:offset+fragLen]) == tc.fragment {
					assert.Equal(t, tc.want, typ,
						"node %q should have type %s", tc.fragment, tc.want)
					found = true
					break
				}
			}
			assert.True(t, found, "annotation for %q not found in source %q", tc.fragment, tc.source)
		})
	}
}

// TestPSCCheckDirectoryScan verifies that psc check reports diagnostics when
// scanning a directory that contains a Perl file with type errors.
func TestPSCCheckDirectoryScan(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory walk uses forward slashes in diagnostic output; skipped on Windows")
	}

	dir := testdataDir(t)
	_, statErr := os.Stat(dir)
	require.NoError(t, statErr, "testdata/check directory must exist")

	_, stderr, err := runPSCCheck(t, dir)
	assert.Error(t, err, "psc check on the testdata directory should return an error (diagnostics.pl contains errors)")
	assert.Contains(t, stderr, "arity-mismatch",
		"stderr should contain arity-mismatch from diagnostics.pl when scanning the directory")
}
