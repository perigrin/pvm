// ABOUTME: Tests the exported oracle API — the callable form of perl's own parse facts.
// ABOUTME: Covers the working directory, stderr, taint shebangs and cancellation the corpus runner depends on.

package parseoracle_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parseoracle"
)

// shimT locates the built corpus shim, whose t/ is the only directory perl's
// own tests compile from. Without it the directory-dependent tests have
// nothing to measure, so they skip rather than fail.
func shimT(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("ORACLE_SHIM")
	if dir == "" {
		dir = "/tmp/oracletree/t"
	}
	if _, err := os.Stat(filepath.Join(dir, "op", "sub.t")); err != nil {
		t.Skipf("corpus shim absent at %s (see README.md to build it): %v", dir, err)
	}
	return dir
}

// TestAskDecodesPrototype is the API's own self-test. A prototype changes how
// a call parses, and perl proves it with an srefgen op; if Ask cannot report
// that, it is not reporting perl's parse.
func TestAskDecodesPrototype(t *testing.T) {
	facts, err := parseoracle.Ask(context.Background(),
		[]byte("sub f(\\@){}\nmy @a;\nf(@a);\n"), parseoracle.Options{})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !facts.OK {
		t.Fatalf("probe must compile, stderr: %s", facts.Stderr)
	}
	if facts.Srefgen != 1 {
		t.Errorf("Srefgen = %d, want 1 (ops %v)", facts.Srefgen, facts.Ops)
	}
	if got := facts.Prototypes["f"]; got != `\@` {
		t.Errorf("Prototypes[f] = %q, want %q", got, `\@`)
	}
}

// TestAskFileRequiresDir is the 498-file case. perl's tests do `chdir 't'` and
// then `require './test.pl'` with @INC cleared, so they compile only from
// inside the shim's t/. A corpus runner that omits Dir measures nothing.
func TestAskFileRequiresDir(t *testing.T) {
	dir := shimT(t)

	with, err := parseoracle.AskFile(context.Background(), "op/sub.t",
		parseoracle.Options{Dir: dir})
	if err != nil {
		t.Fatalf("AskFile with Dir: %v", err)
	}
	if !with.OK {
		t.Fatalf("op/sub.t must compile from %s, stderr: %s", dir, with.Stderr)
	}
	if with.OpCount <= 1000 {
		t.Errorf("OpCount = %d, want > 1000", with.OpCount)
	}

	without, err := parseoracle.AskFile(context.Background(),
		filepath.Join(dir, "op", "sub.t"), parseoracle.Options{})
	if err != nil {
		t.Fatalf("AskFile without Dir: %v", err)
	}
	if without.OK {
		t.Error("op/sub.t must NOT compile without Dir; the directory is mandatory")
	}
}

// TestFactsCarriesStderr records why a file failed. Exit status alone cannot
// tell "your parser is wrong" from "this machine lacks Config.pm", so the
// classifier downstream has no input without this.
func TestFactsCarriesStderr(t *testing.T) {
	facts, err := parseoracle.Ask(context.Background(),
		[]byte("use No::Such::Module::Here;\n"), parseoracle.Options{})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if facts.OK {
		t.Fatal("a missing module must not compile")
	}
	if facts.Stderr == "" {
		t.Fatal("Stderr is empty; the cause of the failure was discarded")
	}
	if !strings.Contains(facts.Stderr, "No::Such::Module::Here") {
		t.Errorf("Stderr does not name the cause: %q", facts.Stderr)
	}
}

// TestShebangPassthrough covers the three corpus files carrying `#!./perl -T`.
// Perl refuses to compile them unless -T is also on the command line, so
// without the passthrough they look like syntax errors and are not.
func TestShebangPassthrough(t *testing.T) {
	dir := shimT(t)
	const taint = "op/taint.t"
	if _, err := os.Stat(filepath.Join(dir, taint)); err != nil {
		t.Skipf("%s absent from shim: %v", taint, err)
	}

	off, err := parseoracle.AskFile(context.Background(), taint,
		parseoracle.Options{Dir: dir})
	if err != nil {
		t.Fatalf("AskFile without Shebang: %v", err)
	}
	if off.OK {
		t.Fatal("op/taint.t is expected to fail without -T; the fixture no longer discriminates")
	}

	on, err := parseoracle.AskFile(context.Background(), taint,
		parseoracle.Options{Dir: dir, Shebang: true})
	if err != nil {
		t.Fatalf("AskFile with Shebang: %v", err)
	}
	if !on.OK {
		t.Errorf("op/taint.t must compile with Shebang set, stderr: %s", on.Stderr)
	}
}

// TestAskCancellation guards the corpus runner against a wedged child. The
// script shells out to an inner `sh -c perl`, which holds the stdout pipe
// open, so cancelling the context is not by itself enough to return.
func TestAskCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := parseoracle.Ask(ctx, []byte("sleep 60;\n"), parseoracle.Options{})
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected an error from a cancelled context")
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Ask did not return after its context was cancelled")
	}
}

// TestScriptPathIsPackageRelative keeps the oracle callable from anywhere. The
// original resolved parse_facts.pl against the process working directory,
// which works only for a test run from inside this package.
func TestScriptPathIsPackageRelative(t *testing.T) {
	t.Chdir(t.TempDir())

	facts, err := parseoracle.Ask(context.Background(),
		[]byte("my $x = 1;\n"), parseoracle.Options{})
	if err != nil {
		t.Fatalf("Ask from an unrelated directory: %v", err)
	}
	if !facts.OK {
		t.Fatalf("probe must compile, stderr: %s", facts.Stderr)
	}
}
