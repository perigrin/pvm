// ABOUTME: Tests for the psc check command.
// ABOUTME: Covers clean files, files with diagnostics, directory walking, and argument validation.

package psc_test

import (
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
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
	// An undef operand of + triggers a coercion-mismatch diagnostic; perl -w
	// warns "Use of uninitialized value". push() with no arguments is not
	// used: perl refuses it, so the parser leaves a Hole there and inference
	// says nothing about it.
	content := "my $u = undef;\nmy $x = $u + 1;\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", file})

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Error(t, err, "check of file with diagnostics should return an error")
	assert.Contains(t, stderr.String(), "coercion-mismatch", "stderr should name the undef operand's mismatch")
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

func TestCheckCommandWithTypeMismatchNoHint(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "mismatch.pl")
	// push($x, 1) where $x is a scalar — triggers type-mismatch.
	// No guard can narrow Scalar to Array, so no hint line appears.
	content := "my $x;\npush($x, 1);\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", file})

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Error(t, err, "check should report diagnostics")
	assert.Contains(t, stderr.String(), "type-mismatch", "should contain type-mismatch diagnostic")
	assert.NotContains(t, stderr.String(), "hint:", "no hint when no guard helps")
}

// TestCheckStrictReportsUnknown verifies that --strict reports values whose
// type could not be inferred, and that the default does not.
//
// The default is permissive because an un-inferred value is not evidence of a
// bug — PSC simply has nothing to say about it. That silence is also the
// failure mode strictness exists to break: without it the checker is quietest
// exactly where it knows least.
func TestCheckStrictReportsUnknown(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "strict.pl")
	// frobnicate is declared nowhere, neither in CORE.pmt nor in the file,
	// so its result is Unknown. Passing it straight into print's Str
	// argument keeps it un-inferred at the point of the check — assigning it
	// to a declared variable first would give that variable a type from
	// context.
	content := "print frobnicate();\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	runCheck := func(args ...string) string {
		cmd := psc.NewCommand()
		cmd.SetArgs(append([]string{"check", file}, args...))
		var stdout strings.Builder
		var stderr strings.Builder
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		// Diagnostics make Execute return an error; the output is the subject.
		_ = cmd.Execute()
		return stderr.String()
	}

	assert.NotContains(t, runCheck(), "Unknown",
		"default mode should not report un-inferred values")
	assert.Contains(t, runCheck("--strict"), "Unknown",
		"--strict should report the un-inferred operand")
}

// TestCheckWarnLossyCoercionFlag verifies that psc check takes
// --warn-lossy-coercion, which sets infer.Options.WarnLossyCoercion: the
// opt-in for a coercion perl makes silently but that loses the value, such
// as `print $hashref`. A clean file stays clean under it.
func TestCheckWarnLossyCoercionFlag(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "clean.pl")
	require.NoError(t, os.WriteFile(file, []byte("my $x = 1;\nprint $x;\n"), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", file, "--warn-lossy-coercion"})
	var stderr strings.Builder
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&stderr)
	require.NoError(t, cmd.Execute(), "stderr: %s", stderr.String())
}

// TestCheckCommandOnLoweredTree runs psc check over every golden fixture
// internal/infer holds and asserts it prints exactly the fixture's golden
// diagnostics, each formatted by infer.FormatDiagnostic. A golden line is
// "diag <start> <end> <code> <message>" and records no severity, so a
// printed line matches a golden diagnostic at either severity.
func TestCheckCommandOnLoweredTree(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "infer", "testdata", "lowered", "*", "*.pl"))
	require.NoError(t, err)
	require.NotEmpty(t, fixtures)

	for _, fx := range fixtures {
		source, err := os.ReadFile(fx)
		require.NoError(t, err)
		golden, err := os.ReadFile(strings.TrimSuffix(fx, ".pl") + ".golden")
		require.NoError(t, err)

		cmd := psc.NewCommand()
		cmd.SetArgs([]string{"check", fx})
		var stdout, stderr strings.Builder
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		// Diagnostics make Execute return an error, which cobra prints
		// after them; the diagnostics are the subject.
		_ = cmd.Execute()
		printed := strings.TrimSuffix(stderr.String(), "Error: diagnostics found\n")

		var got []string
		if printed != "" {
			got = strings.Split(strings.TrimSuffix(printed, "\n"), "\n")
		}
		var missing []string
		for _, line := range strings.Split(string(golden), "\n") {
			fields := strings.SplitN(line, " ", 5)
			if fields[0] != "diag" {
				continue
			}
			require.Len(t, fields, 5, "%s: golden line %q", fx, line)
			start, err := strconv.ParseUint(fields[1], 10, 32)
			require.NoError(t, err)
			end, err := strconv.ParseUint(fields[2], 10, 32)
			require.NoError(t, err)
			d := infer.Diagnostic{StartByte: uint32(start), EndByte: uint32(end), Code: fields[3], Message: fields[4]}
			found := false
			for _, sev := range []infer.Severity{infer.Error, infer.Warning} {
				d.Severity = sev
				if i := indexOf(got, infer.FormatDiagnostic(fx, source, d)); i >= 0 {
					got = append(got[:i], got[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, line)
			}
		}
		assert.Empty(t, missing, "%s: golden diagnostics psc check did not print", fx)
		assert.Empty(t, got, "%s: lines psc check printed that are no golden diagnostic", fx)
	}
}

// indexOf is the index of the first element of lines equal to s, or -1.
func indexOf(lines []string, s string) int {
	for i, l := range lines {
		if l == s {
			return i
		}
	}
	return -1
}

// importSource holds one mismatch at file level and one inside a parenless
// call to My::Block's block, which parses only when My::Block is found. Both
// are ones perl -w warns about: an undef operand and length(@ARGV).
const importSource = "use My::Block;\nmy $u = undef;\nmy $x = $u + 1;\nblock 'inner' => sub { my $y = length(@ARGV); };\n"

// blockModule exports block, whose prototype lets a call to it go without
// parentheses.
const blockModule = "package My::Block;\nour @EXPORT = qw(block);\nsub block($&) { 1 }\n1;\n"

// isolatePerl leaves psc check no perl to resolve and no PERL5LIB, unless
// the test sets one: its HOME and XDG directories are empty, and pvm's
// version variables are unset.
func isolatePerl(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	for _, v := range []string{"HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME", "XDG_STATE_HOME"} {
		t.Setenv(v, home)
	}
	for _, v := range []string{"PERL5LIB", "PVM_PERL_VERSION", "PLENV_VERSION", "PERLBREW_PERL"} {
		t.Setenv(v, "")
	}
}

// runCheckOn writes source to a file in its own directory and returns what
// psc check prints.
func runCheckOn(t *testing.T, source string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "t.pl")
	require.NoError(t, os.WriteFile(file, []byte(source), 0644))
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", file})
	var stdout, stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	// Diagnostics make Execute return an error; the output is the subject.
	_ = cmd.Execute()
	return stderr.String()
}

// TestCheckCommandResolvesImportsThroughINC verifies psc check parses a file
// as perl compiles it, resolving its use statements through @INC, here a
// PERL5LIB entry: the parenless block call parses, and the mismatch in its
// body is reported.
func TestCheckCommandResolvesImportsThroughINC(t *testing.T) {
	isolatePerl(t)
	lib := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(lib, "My"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(lib, "My", "Block.pm"), []byte(blockModule), 0644))
	t.Setenv("PERL5LIB", lib)

	out := runCheckOn(t, importSource)
	assert.Contains(t, out, "t.pl:3:", "the file-level mismatch is reported")
	assert.Contains(t, out, "t.pl:4:", "the mismatch inside block's body is reported")
}

// TestCheckCommandWithoutPerl verifies that with no resolved perl and no
// PERL5LIB, psc check parses against the file's own directory instead of
// failing: the file-level mismatch is still reported, and the unresolved
// block call is a Hole whose refusal is reported and whose body is not
// checked. perl, finding no My::Block, refuses it too: "syntax error at
// t.pl line 4, near "block 'inner'"".
func TestCheckCommandWithoutPerl(t *testing.T) {
	isolatePerl(t)
	out := runCheckOn(t, importSource)
	assert.Contains(t, out, "t.pl:3:", "the file-level mismatch is reported")
	assert.Contains(t, out, "t.pl:4:1: error: perl refuses this", "the unresolved block call is refused")
	assert.NotContains(t, out, "[type-mismatch]", "nothing is checked inside the Hole")
	assert.NotContains(t, out, "panic")
}

// TestCheckCommandOnBrokenFile verifies psc check on a file perl rejects
// (`my $x = ;` is a syntax error) neither panics nor reports a type
// diagnostic inside the Hole the parser leaves for the missing term: it
// reports the refusal, and checks the rest of the file.
func TestCheckCommandOnBrokenFile(t *testing.T) {
	isolatePerl(t)
	var out string
	require.NotPanics(t, func() { out = runCheckOn(t, "my $x = ;\nmy $u = undef;\nmy $y = $u + 1;\n") })
	line1 := []string{}
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "t.pl:1:") {
			line1 = append(line1, l)
		}
	}
	if assert.Len(t, line1, 1, "only the refusal is reported on line 1") {
		assert.Contains(t, line1[0], "[not_a_term]")
	}
	assert.Contains(t, out, "t.pl:3:", "the rest of the file is still checked")
}

// TestRefusalIsReported: code perl refuses to compile is reported by psc
// check and the LSP, each refusal as its code over its span, and psc check
// exits non-zero. perl 5.42 says "Not enough arguments for push" for the
// first source and "Missing right curly or square bracket" for the second.
// The control compiles, and nothing is reported for it.
func TestRefusalIsReported(t *testing.T) {
	isolatePerl(t)
	for _, c := range []struct{ source, code, at string }{
		{"push(); my $x = length();\n", "call_arity", "push()"},
		{"my $x;\nif ($x) { print \"a\";\n", "unclosed_opener", ";"},
	} {
		file := filepath.Join(t.TempDir(), "t.pl")
		require.NoError(t, os.WriteFile(file, []byte(c.source), 0644))
		cmd := psc.NewCommand()
		cmd.SetArgs([]string{"check", file})
		var stderr strings.Builder
		cmd.SetOut(&strings.Builder{})
		cmd.SetErr(&stderr)
		assert.Error(t, cmd.Execute(), "%q: psc check exits non-zero", c.source)
		assert.Contains(t, stderr.String(), "["+c.code+"]", c.source)

		s := psc.NewLSPServer()
		uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(file)}).String()
		require.NoError(t, s.OpenDocument(uri, []byte(c.source)))
		var refusals []infer.Diagnostic
		for _, d := range s.Diagnostics(uri) {
			if d.Code == c.code {
				refusals = append(refusals, d)
			}
		}
		if assert.Len(t, refusals, 1, "%q: the LSP reports the refusal once", c.source) {
			at := uint32(strings.LastIndex(c.source, c.at))
			assert.LessOrEqual(t, refusals[0].StartByte, at, c.source)
			assert.Greater(t, refusals[0].EndByte, at, "%q: the refusal spans %q", c.source, c.at)
		}
	}

	out := runCheckOn(t, "my @a;\npush(@a, 1);\n")
	assert.Empty(t, out, "control: perl compiles push(@a, 1), and nothing is reported")
}
