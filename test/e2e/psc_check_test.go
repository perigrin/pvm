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
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
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

// runPSCCheck invokes the psc check command via the Go API against target
// (file or directory) and returns (stdout, stderr, error). It checks a copy
// of target's directory in a fresh one, with PERL5LIB empty: the copy names
// no perl, so what psc check prints depends neither on the .perl-version
// above the testdata nor on which perls the host has installed.
func runPSCCheck(t *testing.T, target string, flags ...string) (string, string, error) {
	t.Helper()

	src, name := target, ""
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		src, name = filepath.Dir(target), filepath.Base(target)
	}
	copied := filepath.Join(t.TempDir(), filepath.Base(src))
	require.NoError(t, os.CopyFS(copied, os.DirFS(src)))
	t.Setenv("PERL5LIB", "")

	cmd := psc.NewCommand()
	cmd.SetArgs(append([]string{"check", filepath.Join(copied, name)}, flags...))

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
// type-mismatch diagnostics when run against a file containing type errors.
// Its `push();` is one perl refuses ("Not enough arguments for push"): the
// parser refuses it too, and the refusal is reported on that line.
func TestPSCCheckDiagnostics(t *testing.T) {
	diagFile := filepath.Join(testdataDir(t), "diagnostics.pl")
	_, statErr := os.Stat(diagFile)
	require.NoError(t, statErr, "testdata/check/diagnostics.pl must exist")

	_, stderr, err := runPSCCheck(t, diagFile)
	assert.Error(t, err, "psc check on a file with errors should return a non-zero exit")

	// push() with no arguments is refused: line 1 reports the refusal.
	assert.Contains(t, stderr, "diagnostics.pl:1:1: error: perl refuses this",
		"push() with no arguments is refused")
	assert.Contains(t, stderr, "[call_arity]", "the refusal carries its code")

	// Container-level type mismatches: wrong sigil for the builtin
	expectedDiagnostics := []struct {
		fragment string
		reason   string
	}{
		{"push", "push($scalar, 1) — push expects Array, got Scalar"},
		{"keys", "keys($scalar) — keys expects Hash, got Scalar"},
		{"length", "length(@arr) — length expects Str, got Array"},
		{"defined", "defined(@arr) — defined expects Scalar, got Array"},
		{"splice", "splice($scalar) — splice expects Array, got Scalar"},
	}
	for _, diag := range expectedDiagnostics {
		assert.Contains(t, stderr, diag.fragment,
			"stderr should contain diagnostic for %s", diag.reason)
	}

	// Verify we get type-mismatch codes (not just the function names)
	assert.Contains(t, stderr, "type-mismatch",
		"stderr should contain type-mismatch diagnostic codes")
}

// TestTypeInferenceAccuracy exercises infer.Analyze directly to verify
// that specific Perl constructs are inferred with the expected types.
func TestTypeInferenceAccuracy(t *testing.T) {
	cases := []struct {
		name      string
		source    string
		fragment  string // substring of source whose node type we check
		want      types.Type
		minOffset int // skip annotations before this byte offset (avoids matching declarations)
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
		{
			name:     "narrowed scalar to Int",
			source:   "my $x = 42;\n$x;\n",
			fragment: "$x",
			want:     types.Int,
			// The reference $x on line 2 starts at byte 12; skip the
			// declaration to test the narrowed annotation.
			minOffset: 12,
		},
		{
			name:      "narrowed scalar to Num",
			source:    "my $n = 3.14;\n$n;\n",
			fragment:  "$n",
			want:      types.Num,
			minOffset: 14,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			root := parse.Parse(src)
			f := lower.Lower(root)
			pos := lower.NewPositions(root, f)
			annotations, _, _ := infer.Analyze(f, pos, nil, infer.Options{})

			// Search the lowered tree, outermost first, for the node whose
			// source text is the fragment, skipping nodes that start below
			// minOffset to distinguish references from declarations.
			var found lower.Node
			var walk func(n lower.Node)
			walk = func(n lower.Node) {
				if found != nil {
					return
				}
				start, end, ok := pos.Span(n.Meta().ID)
				if ok && start >= tc.minOffset && string(src[start:end]) == tc.fragment {
					found = n
					return
				}
				for _, c := range lower.Children(n) {
					walk(c)
				}
			}
			walk(f)
			require.NotNil(t, found, "node for %q not found in source %q", tc.fragment, tc.source)
			assert.Equal(t, tc.want, annotations[found.Meta().ID],
				"node %q should have type %s", tc.fragment, tc.want)
		})
	}
}

// TestPSCCheckValueDiagnostics verifies that psc check fires type-mismatch
// diagnostics for value-level anti-patterns where the sigil is correct ($scalar)
// but assignment narrowing has refined the type to something incompatible with
// the builtin's expected argument type.
func TestPSCCheckValueDiagnostics(t *testing.T) {
	diagFile := filepath.Join(testdataDir(t), "value_diagnostics.pl")
	_, statErr := os.Stat(diagFile)
	require.NoError(t, statErr, "testdata/check/value_diagnostics.pl must exist")

	_, stderr, err := runPSCCheck(t, diagFile)
	assert.Error(t, err, "value-level anti-patterns should produce diagnostics")

	// keys(42) — keys expects Hash, got Int
	assert.Contains(t, stderr, "keys")
	assert.Contains(t, stderr, "type-mismatch")

	// push(100, 1) — push expects Array, got Int
	assert.Contains(t, stderr, "push")

	// values(3.14) — values expects Hash, got Num
	assert.Contains(t, stderr, "values")

	// splice(42) — splice expects Array, got Int
	assert.Contains(t, stderr, "splice")
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
	assert.Contains(t, stderr, "diagnostics.pl:4:",
		"stderr should contain push($x, 1)'s type-mismatch from diagnostics.pl when scanning the directory")
}

// TestPSCCheckUndefPropagation verifies that psc check detects uninitialized
// variables used in arithmetic context.
func TestPSCCheckUndefPropagation(t *testing.T) {
	undefFile := filepath.Join(testdataDir(t), "undef_propagation.pl")
	_, statErr := os.Stat(undefFile)
	require.NoError(t, statErr, "testdata/check/undef_propagation.pl must exist")

	_, stderr, err := runPSCCheck(t, undefFile)
	assert.Error(t, err, "psc check on undef_propagation.pl should return a non-zero exit")

	assert.Contains(t, stderr, "coercion-mismatch",
		"stderr should contain coercion-mismatch for undef in arithmetic")
}

// TestPSCCheckCoercionDiagnostics verifies that a reference in a numeric
// slot is a lossy coercion: perl takes it silently, so psc check says
// nothing by default and reports it under --warn-lossy-coercion.
func TestPSCCheckCoercionDiagnostics(t *testing.T) {
	coercionFile := filepath.Join(testdataDir(t), "coercion.pl")
	_, statErr := os.Stat(coercionFile)
	require.NoError(t, statErr, "testdata/check/coercion.pl must exist")

	_, stderr, err := runPSCCheck(t, coercionFile)
	assert.NoError(t, err, "a lossy coercion is silent by default")
	assert.Empty(t, stderr, "a lossy coercion is silent by default")

	_, stderr, err = runPSCCheck(t, coercionFile, "--warn-lossy-coercion")
	assert.Error(t, err, "--warn-lossy-coercion reports the references in numeric slots")
	assert.Contains(t, stderr, "lossy-coercion",
		"stderr should contain lossy-coercion diagnostic codes under the opt-in")
}

// TestPSCCheckUsesDeclarations: Carp.tpm declares `sub carp :export
// (List[Str] @message) Boolean`, so psc check reports the undef carp is
// handed. Measured, perl 5.42 -w on declared_mismatch.pl warns "Use of
// uninitialized value $error[1] in join or string at .../Carp.pm line
// 588."
func TestPSCCheckUsesDeclarations(t *testing.T) {
	_, stderr, err := runPSCCheck(t, filepath.Join(testdataDir(t), "declared_mismatch.pl"))
	assert.Error(t, err)
	assert.Contains(t, stderr, "declared_mismatch.pl:7:")
	assert.Contains(t, stderr, "call to carp")
	assert.Contains(t, stderr, "coercion-mismatch")
}

// TestPSCCheckDeclarationsNoFalsePositive: calls whose arguments fit
// Carp's declarations report nothing; perl 5.42 -w runs declared_fit.pl
// with no warning of its own.
func TestPSCCheckDeclarationsNoFalsePositive(t *testing.T) {
	_, stderr, err := runPSCCheck(t, filepath.Join(testdataDir(t), "declared_fit.pl"))
	assert.NoError(t, err)
	assert.Empty(t, stderr)
}

// TestPSCCheckUndecidedCallReportsNothing: `return localtime($_[0])` is in
// a context its call site cannot decide, so the call joins localtime's
// candidates and nothing is reported. perl 5.42 -w runs undecided_call.pl
// silently, in list and in scalar context. A library multi is joined the
// same way: undecided_declared.pl returns the parse fixture My::Lib's
// when_, Str in scalar context and List[Str] in list context, which perl
// 5.42 -w runs silently with a My::Lib whose when_ answers wantarray.
func TestPSCCheckUndecidedCallReportsNothing(t *testing.T) {
	_, stderr, err := runPSCCheck(t, filepath.Join(testdataDir(t), "undecided_call.pl"))
	assert.NoError(t, err)
	assert.Empty(t, stderr)

	t.Cleanup(parse.OverlayDeclarations(os.DirFS(filepath.Join(testdataDir(t), "..", "..", "..", "..", "internal", "parse", "testdata"))))
	_, stderr, err = runPSCCheck(t, filepath.Join(testdataDir(t), "undecided_declared.pl"))
	assert.NoError(t, err)
	assert.Empty(t, stderr)
}

// TestPSCCheckSingleCandidateMismatchIsCoercion: cluck has one declared
// candidate, whose arity `cluck(undef)` fits, so the candidate is selected
// and the undef is coerced to its Str: a coercion mismatch, not a call no
// candidate accepts. perl 5.42 -w warns "Use of uninitialized value
// $error[0] in join or string at .../Carp.pm line 588."
func TestPSCCheckSingleCandidateMismatchIsCoercion(t *testing.T) {
	_, stderr, err := runPSCCheck(t, filepath.Join(testdataDir(t), "declared_single.pl"))
	assert.Error(t, err)
	assert.Contains(t, stderr, "declared_single.pl:7:")
	assert.Contains(t, stderr, "coercion-mismatch")
	assert.NotContains(t, stderr, "arity-mismatch")
	assert.NotContains(t, stderr, "no declared candidate")
}
