// ABOUTME: Tests the corpus machinery — shim construction, version pinning, stderr classification.
// ABOUTME: The corpus is referenced, not vendored, so every test skips cleanly without $PERL5_CORPUS.

package parseoracle_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// corpusRoot resolves the corpus and skips the test when it is absent. perl's
// tests are Artistic/GPL, so they are referenced rather than vendored; a
// machine with no perl5 checkout must skip rather than fail.
func corpusRoot(t *testing.T) string {
	t.Helper()
	root, err := parseoracle.CorpusRoot()
	if err != nil {
		t.Skipf("corpus unavailable: %v", err)
	}
	return root
}

// TestCorpusSkipsWithoutEnv proves the suite degrades to a skip rather than a
// failure when PERL5_CORPUS names nothing usable. Someone cloning this repo
// without a perl5 checkout must still be able to run `go test ./...`.
func TestCorpusSkipsWithoutEnv(t *testing.T) {
	t.Setenv("PERL5_CORPUS", filepath.Join(t.TempDir(), "definitely-not-a-corpus"))

	if _, err := parseoracle.CorpusRoot(); err == nil {
		t.Fatal("CorpusRoot must report an error when the corpus is missing, so callers can skip")
	}
}

// TestBuildShimCompilesSubT is the shim's reason to exist. t/test.pl clears
// @INC and unshifts ../lib, so perl's tests only compile inside a tree where
// ../lib is populated. 498 of 620 corpus files use that convention.
func TestBuildShimCompilesSubT(t *testing.T) {
	root := corpusRoot(t)
	shim := t.TempDir()

	if err := parseoracle.BuildShim(root, shim); err != nil {
		t.Fatalf("BuildShim: %v", err)
	}

	cmd := exec.Command("perl", "-c", "--", "op/sub.t")
	cmd.Dir = filepath.Join(shim, "t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("op/sub.t must compile inside the shim, got %v:\n%s", err, out)
	}
}

// TestShimHasConfigPm gets its own assertion because its absence is the exact
// omission that made an earlier measurement report 66.3% instead of 94.2%.
// Config.pm lives in the architecture-specific @INC root, so a shim populated
// from only the pure-perl root silently loses ~170 files.
func TestShimHasConfigPm(t *testing.T) {
	root := corpusRoot(t)
	shim := t.TempDir()

	if err := parseoracle.BuildShim(root, shim); err != nil {
		t.Fatalf("BuildShim: %v", err)
	}

	config := filepath.Join(shim, "lib", "Config.pm")
	if _, err := os.Stat(config); err != nil {
		t.Fatalf("Config.pm missing from the shim's lib/ (%v) — the shim was "+
			"populated from only one library root; this is the omission that "+
			"cost ~170 files and 28 percentage points", err)
	}
}

// TestReadPinRejectsVersionSkew and TestReadPinRejectsRevisionSkew are
// separate on purpose. A pin that checks only one of its two halves is a pin
// that silently drifts, and both halves have already drifted here at least
// once: the corpus is blead 5.45 while the interpreter is 5.42.0, which is why
// t/op/for-many.t is a genuine syntax error rather than a parser bug.
func TestReadPinRejectsVersionSkew(t *testing.T) {
	pin := parseoracle.Pin{Interpreter: "5.999000", Revision: "94e5086608"}

	err := pin.Check("5.042000", "94e5086608")
	if err == nil {
		t.Fatal("a pin must reject an interpreter mismatch")
	}

	var mismatch *parseoracle.PinMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("want a typed *PinMismatchError, got %T: %v", err, err)
	}
	// The error must name BOTH sides: "it does not match" cannot tell you
	// whether the corpus moved or the interpreter did.
	for _, want := range []string{"5.999000", "5.042000"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error must name both sides of the mismatch, %q missing from: %v", want, err)
		}
	}
}

func TestReadPinRejectsRevisionSkew(t *testing.T) {
	pin := parseoracle.Pin{Interpreter: "5.042000", Revision: "94e5086608"}

	err := pin.Check("5.042000", "deadbeef00")
	if err == nil {
		t.Fatal("a pin must reject a corpus-revision mismatch")
	}

	var mismatch *parseoracle.PinMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("want a typed *PinMismatchError, got %T: %v", err, err)
	}
	for _, want := range []string{"94e5086608", "deadbeef00"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error must name both sides of the mismatch, %q missing from: %v", want, err)
		}
	}

	// A pin that agrees on both halves must pass, or Check is a constant
	// failure and the two tests above prove nothing.
	if err := pin.Check("5.042000", "94e5086608"); err != nil {
		t.Errorf("a matching pin must pass, got: %v", err)
	}
}

// TestReadPinReadsBothHalves reads the checked-in pin. Recording both the
// interpreter and the corpus revision is what makes a measurement
// reproducible instead of an anecdote.
func TestReadPinReadsBothHalves(t *testing.T) {
	pin, err := parseoracle.ReadPin(filepath.Join("testdata", "corpus.pin"))
	if err != nil {
		t.Fatalf("ReadPin: %v", err)
	}
	if pin.Interpreter == "" {
		t.Error("pin records no interpreter version")
	}
	if pin.Revision == "" {
		t.Error("pin records no corpus revision")
	}
}

// TestClassifyFixtures runs Classify over stderr captured from real `perl -c`
// runs against this corpus, not over hand-written strings. Exit status cannot
// separate "your parser is wrong" from "this machine lacks Config.pm", and a
// ratchet built on exit status encodes the second as though it were the first.
//
// Each fixture is a recorded failure:
//   - environmental: op/magic.t compiled WITHOUT the shim
//   - versionskew:   op/signatures.t, a warnings category 5.42 lacks
//   - syntaxerror:   op/for-many.t, blead-only refaliasing
//   - other:         op/taint.t, which needs -T on the command line
func TestClassifyFixtures(t *testing.T) {
	for _, tc := range []struct {
		fixture string
		want    parseoracle.FailureKind
	}{
		{"environmental.txt", parseoracle.Environmental},
		{"versionskew.txt", parseoracle.VersionSkew},
		{"syntaxerror.txt", parseoracle.SyntaxError},
		{"other.txt", parseoracle.Other},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			path := filepath.Join("testdata", "stderr", tc.fixture)
			stderr, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading fixture: %v", err)
			}
			if got := parseoracle.Classify(string(stderr)); got != tc.want {
				t.Errorf("Classify(%s) = %v, want %v\nstderr was:\n%s",
					tc.fixture, got, tc.want, stderr)
			}
		})
	}
}

// TestClassifyPrefersEnvironmentalOverSyntax records the ordering that keeps a
// broken setup from being reported as a language failure. `perl -c` stops at
// the FIRST error, so a file whose missing module aborts compilation never
// reaches whatever else is wrong with it.
func TestClassifyPrefersEnvironmentalOverSyntax(t *testing.T) {
	const mixed = "Can't locate Config.pm in @INC (you may need to install the Config module) at t/op/x.t line 3.\n" +
		"syntax error at t/op/x.t line 9, near \"( \\\"\n"

	if got := parseoracle.Classify(mixed); got != parseoracle.Environmental {
		t.Errorf("Classify = %v, want Environmental: a missing module aborts "+
			"compilation before the parser reaches later lines, so this is not "+
			"evidence of a syntax problem", got)
	}
}

// TestClassifyCleanStderrIsNoFailure guards the empty case: a file that
// compiled has nothing to classify.
func TestClassifyCleanStderrIsNoFailure(t *testing.T) {
	if got := parseoracle.Classify("op/sub.t syntax OK\n"); got != parseoracle.NoFailure {
		t.Errorf("Classify(clean) = %v, want NoFailure", got)
	}
}
