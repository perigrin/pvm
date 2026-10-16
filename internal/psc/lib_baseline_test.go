// ABOUTME: Holds psc check over perl 5.42's lib to a committed baseline of its diagnostics.
// ABOUTME: A diagnostic not in the baseline fails unless it is a true positive perl -w confirms.

package psc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/perltest"
)

// libDiagShape is one diagnostic psc check prints, split into its path and
// the rest: line, column, severity, message and code. The path is matched
// lazily, so a Windows drive letter's colon stays in it.
var libDiagShape = regexp.MustCompile(`^(.+?):(\d+:\d+: (?:error|warning|info): .+ \[[a-z_-]+\])$`)

// perlWShape is the tail perl -w puts on a warning or a death.
var perlWShape = regexp.MustCompile(` at \S+ line \d+\.`)

// libEntry is a non-blank, non-comment line and its 1-based number, so an
// error can name the line it is about.
type libEntry struct {
	n    int
	line string
}

// libEntries returns the entries of text in order.
func libEntries(text string) []libEntry {
	var entries []libEntry
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, libEntry{i + 1, line})
	}
	return entries
}

// parseLibBaseline reads lib_baseline.txt: one privlib-relative diagnostic
// per line, with forward slashes, and # comments. A line that is not a
// diagnostic is an error naming it.
func parseLibBaseline(text string) (map[string]bool, error) {
	baseline := map[string]bool{}
	for _, e := range libEntries(text) {
		if !libDiagShape.MatchString(e.line) {
			return nil, fmt.Errorf("baseline line %d is not a diagnostic: %q", e.n, e.line)
		}
		baseline[e.line] = true
	}
	return baseline, nil
}

// parseLibTruePositives reads lib_true_positives.txt: a privlib-relative
// diagnostic, a tab, and the line perl -w prints for that file or for a
// minimal file with the same construct, copied verbatim. An entry without
// that evidence is an error naming its line.
func parseLibTruePositives(text string) (map[string]bool, error) {
	admitted := map[string]bool{}
	for _, e := range libEntries(text) {
		diag, evidence, _ := strings.Cut(e.line, "\t")
		if !libDiagShape.MatchString(diag) {
			return nil, fmt.Errorf("true positive line %d is not a diagnostic: %q", e.n, diag)
		}
		if !perlWShape.MatchString(evidence) {
			return nil, fmt.Errorf("true positive line %d has no perl -w output as evidence: %q", e.n, evidence)
		}
		admitted[diag] = true
	}
	return admitted, nil
}

// newLibDiagnostics returns the diagnostics in psc check's output that are
// neither in the baseline nor admitted true positives, each made relative
// to privlib with forward slashes. A line under privlib that is not a
// diagnostic is returned too, so a format psc check has not printed before
// fails rather than passing unread. Indented hints and psc's own notes are
// not under privlib and are not diagnostics.
func newLibDiagnostics(output, privlib string, baseline, admitted map[string]bool) []string {
	// ponytail: backslashes become slashes on every OS, so a Windows path
	// normalizes in tests run anywhere; no file in perl's lib has one.
	root := strings.TrimSuffix(strings.ReplaceAll(privlib, `\`, "/"), "/") + "/"
	var fresh []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		m := libDiagShape.FindStringSubmatch(line)
		if m == nil {
			if rel := strings.ReplaceAll(line, `\`, "/"); strings.HasPrefix(rel, root) {
				fresh = append(fresh, strings.TrimPrefix(rel, root))
			}
			continue
		}
		diag := strings.TrimPrefix(strings.ReplaceAll(m[1], `\`, "/"), root) + ":" + m[2]
		if !baseline[diag] && !admitted[diag] {
			fresh = append(fresh, diag)
		}
	}
	return fresh
}

// libCheckFinished is nil when psc check's run over lib ended as a finished
// run does: with no error, or with the one it returns for having printed
// diagnostics. Any other error means the walk stopped early, so the
// output holds only part of lib and cannot be compared.
func libCheckFinished(err error) error {
	if err == nil || err.Error() == "diagnostics found" {
		return nil
	}
	return fmt.Errorf("psc check did not finish: %w", err)
}

const (
	findDiag  = "File/Find.pm:96:9: error: left operand of \"=~\": expected Str, got Undef|Int|Num|Str|NaN|Inf [coercion-mismatch]"
	zlibDiag  = "IO/Zlib.pm:623:17: warning: left operand of \"==\": expected Num, got Undef|Int [coercion-mismatch]"
	ansiDiag  = "Term/ANSIColor.pm:533:46: error: left operand of \".\": expected Str, got Undef|Int|Num|Str|NaN|Inf [coercion-mismatch]"
	perlWLine = "Use of uninitialized value $x in concatenation (.) or string at -e line 1."
)

// twoLineBaseline is a baseline of two diagnostics with its header.
func twoLineBaseline(t *testing.T) map[string]bool {
	t.Helper()
	baseline, err := parseLibBaseline("# captured for a test\n" + findDiag + "\n" + zlibDiag + "\n")
	require.NoError(t, err)
	return baseline
}

func TestLibBaselineRejectsANewDiagnostic(t *testing.T) {
	out := "/lib/" + findDiag + "\n/lib/" + ansiDiag + "\n"
	assert.Equal(t, []string{ansiDiag}, newLibDiagnostics(out, "/lib", twoLineBaseline(t), nil))
}

func TestLibBaselineAdmitsAListedTruePositive(t *testing.T) {
	out := "/lib/" + ansiDiag + "\n"
	admitted, err := parseLibTruePositives(ansiDiag + "\t" + perlWLine + "\n")
	require.NoError(t, err)
	assert.Empty(t, newLibDiagnostics(out, "/lib", twoLineBaseline(t), admitted))

	_, err = parseLibTruePositives(ansiDiag + "\t\n")
	assert.ErrorContains(t, err, "line 1", "an empty evidence field is rejected")
}

func TestLibBaselineNormalizesWindowsSeparators(t *testing.T) {
	out := `C:\perl\lib\File\Find.pm:96:9: error: left operand of "=~": expected Str, got Undef|Int|Num|Str|NaN|Inf [coercion-mismatch]` + "\r\n"
	assert.Empty(t, newLibDiagnostics(out, `C:\perl\lib`, twoLineBaseline(t), nil))
}

func TestLibBaselineAllowsADisappearedDiagnostic(t *testing.T) {
	assert.Empty(t, newLibDiagnostics("/lib/"+findDiag+"\n", "/lib", twoLineBaseline(t), nil))
}

func TestLibBaselineRejectsAMalformedLine(t *testing.T) {
	baseline, err := parseLibBaseline("# a comment is not a diagnostic\n")
	require.NoError(t, err)
	assert.Empty(t, baseline)

	_, err = parseLibBaseline("# header\n" + findDiag + "\nFile/Find.pm: something went wrong\n")
	assert.ErrorContains(t, err, "line 3")
}

func TestLibTruePositiveRequiresPerlWOutput(t *testing.T) {
	_, err := parseLibTruePositives(ansiDiag + "\tperl warns here because $x may be undef\n")
	assert.ErrorContains(t, err, "line 1")
}

func TestLibTruePositiveMatchesExactDiagnostic(t *testing.T) {
	admitted, err := parseLibTruePositives(ansiDiag + "\t" + perlWLine + "\n")
	require.NoError(t, err)
	otherFile := strings.Replace(ansiDiag, "Term/ANSIColor.pm", "Term/Other.pm", 1)
	otherLine := strings.Replace(ansiDiag, ":533:", ":534:", 1)
	out := "/lib/" + otherFile + "\n/lib/" + otherLine + "\n"
	assert.Equal(t, []string{otherFile, otherLine}, newLibDiagnostics(out, "/lib", map[string]bool{}, admitted))
}

func TestLibBaselineRejectsANewRefusal(t *testing.T) {
	refusal := "Foo.pm:3:1: error: perl refuses this: a call with too many arguments [call_arity]"
	note := "Foo.pm:4:1: info: a note [unknown-builtin]"
	out := "/lib/" + refusal + "\n/lib/" + note + "\n"
	assert.Equal(t, []string{refusal, note}, newLibDiagnostics(out, "/lib", twoLineBaseline(t), nil))

	baseline, err := parseLibBaseline(refusal + "\n" + note + "\n")
	require.NoError(t, err)
	assert.Empty(t, newLibDiagnostics(out, "/lib", baseline, nil), "a listed refusal and info line match the baseline")

	admitted, err := parseLibTruePositives(refusal + "\t" + perlWLine + "\n" + note + "\t" + perlWLine + "\n")
	require.NoError(t, err)
	assert.Empty(t, newLibDiagnostics(out, "/lib", map[string]bool{}, admitted), "an admitted refusal and info line pass")
}

func TestLibBaselineRejectsAnUnparsedLibLine(t *testing.T) {
	odd := "Foo.pm:3:1: fatal: something psc check never printed before"
	out := "/lib/" + odd + "\n  hint: an indented hint is not a diagnostic\npsc: note: not under lib\n"
	assert.Equal(t, []string{odd}, newLibDiagnostics(out, "/lib", twoLineBaseline(t), nil))
}

func TestPSCCheckLibBaselineRequiresAFinishedRun(t *testing.T) {
	assert.NoError(t, libCheckFinished(nil))
	assert.NoError(t, libCheckFinished(errors.New("diagnostics found")))

	_, err := runCheckErr(t, filepath.Join(t.TempDir(), "missing"))
	assert.ErrorContains(t, libCheckFinished(err), "did not finish")
}

// TestPSCCheckLibBaseline runs psc check over perl 5.42.0's privlib, as
// `PVM_PERL_VERSION=5.42.0 psc check <privlib>` does, and fails on any
// diagnostic not in testdata/lib_baseline.txt or admitted by
// testdata/lib_true_positives.txt. It takes about 20s (18-20s measured
// on 2026-10-16), well inside go test's default 10m.
func TestPSCCheckLibBaseline(t *testing.T) {
	p := perltest.Perl542(t)
	t.Setenv("PVM_PERL_VERSION", "5.42.0")
	t.Setenv("PERL5LIB", "")

	read := func(name string) string {
		data, err := os.ReadFile(filepath.Join("testdata", name))
		require.NoError(t, err)
		return string(data)
	}
	baseline, err := parseLibBaseline(read("lib_baseline.txt"))
	require.NoError(t, err)
	admitted, err := parseLibTruePositives(read("lib_true_positives.txt"))
	require.NoError(t, err)

	out, err := runCheckErr(t, p.Privlib)
	require.NoError(t, libCheckFinished(err))
	if fresh := newLibDiagnostics(out, p.Privlib, baseline, admitted); len(fresh) > 0 {
		t.Errorf("psc check over %s reports %d diagnostic(s) not in the baseline:\n  %s\n"+
			"Fix each, or list it in testdata/lib_true_positives.txt with the line perl -w prints for it.",
			p.Privlib, len(fresh), strings.Join(fresh, "\n  "))
	}
}
