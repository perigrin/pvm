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
	"tamarou.com/pvm/internal/perltest"
)

// shimT locates the built corpus shim, whose t/ is the only directory perl's
// own tests compile from, for the tests that measure with the pinned perl p.
//
// $PARSEORACLE_SHIM, which CI sets, is made ready for corpus.pin the way
// the sweep makes it, rebuilt if it was built for anything else. A shim at
// $ORACLE_SHIM (default /tmp/oracletree/t) is a local convenience and is
// never rebuilt: one not stamped for corpus.pin is skipped, because its
// lib/Config.pm is another perl's and the pinned perl fails on it with
// "Insecure dependency" or "Perl lib version" rather than measuring. Without
// a shim the directory-dependent tests have nothing to measure, so they skip
// rather than fail.
func shimT(t *testing.T, p perltest.Perl) string {
	t.Helper()
	if shim := os.Getenv("PARSEORACLE_SHIM"); shim != "" {
		readyShim(t, p, shim)
		return filepath.Join(shim, "t")
	}
	dir := os.Getenv("ORACLE_SHIM")
	if dir == "" {
		dir = "/tmp/oracletree/t"
	}
	if _, err := os.Stat(filepath.Join(dir, "op", "sub.t")); err != nil {
		t.Skipf("corpus shim absent at %s (see README.md to build it): %v", dir, err)
	}
	pin, err := parseoracle.ReadPin(filepath.Join("testdata", "corpus.pin"))
	if err != nil {
		t.Fatalf("ReadPin: %v", err)
	}
	if !parseoracle.ShimReady(filepath.Dir(dir), pin) {
		t.Skipf("the shim at %s is not stamped for corpus.pin (%s, threads %t, revision %s); "+
			"set PARSEORACLE_SHIM to have it built for the pinned perl",
			dir, pin.Interpreter, pin.Threads, pin.Revision)
	}
	return dir
}

// TestAskDecodesPrototype is the API's own self-test. A prototype changes how
// a call parses, and perl proves it with an srefgen op; if Ask cannot report
// that, it is not reporting perl's parse.
func TestAskDecodesPrototype(t *testing.T) {
	requirePOSIXOracle(t)
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
	p := perltest.PinnedPerl(t)
	dir := shimT(t, p)

	with, err := parseoracle.AskFile(context.Background(), "op/sub.t",
		parseoracle.Options{Dir: dir, Perl: p.Path})
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
	requirePOSIXOracle(t)
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
//
// The two halves cost wildly different amounts, and the asymmetry is the
// point rather than an accident. Measured on an idle machine:
//
//	without -T   150ms    perl rejects it and compiles nothing
//	with -T      32s      the full compile, walk included
//
// So this test's cost is one 32-second measurement against DefaultTimeout's
// 60 seconds, and on a loaded machine that margin is not enough -- the
// failure this test was flaking with. An explicit timeout is set rather than
// leaning on the default, because the default is calibrated for a corpus
// file costing milliseconds and this file is three orders of magnitude off
// that. See the note on DefaultTimeout.
//
// Raising the DEFAULT instead would be worse: it would slow every wedged
// file in the corpus sweep by the same amount, to accommodate one fixture.
func TestShebangPassthrough(t *testing.T) {
	p := perltest.PinnedPerl(t)
	dir := shimT(t, p)
	const taint = "op/taint.t"
	if _, err := os.Stat(filepath.Join(dir, taint)); err != nil {
		t.Skipf("%s absent from shim: %v", taint, err)
	}

	// Generous because the measurement is genuinely expensive, not because
	// a slow answer is acceptable: a wedged child still fails, four minutes
	// later instead of one.
	const budget = 4 * time.Minute

	off, err := parseoracle.AskFile(context.Background(), taint,
		parseoracle.Options{Dir: dir, Timeout: budget, Perl: p.Path})
	if err != nil {
		t.Fatalf("AskFile without Shebang: %v", err)
	}
	if off.OK {
		t.Fatal("op/taint.t is expected to fail without -T; the fixture no longer discriminates")
	}

	on, err := parseoracle.AskFile(context.Background(), taint,
		parseoracle.Options{Dir: dir, Shebang: true, Timeout: budget, Perl: p.Path})
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
//
// The probe blocks in BEGIN because the oracle only ever compiles: a plain
// `sleep` is never reached under perl -c, and a test using one would race the
// oracle's own 60ms runtime rather than measure cancellation.
func TestAskCancellation(t *testing.T) {
	requirePOSIXOracle(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := parseoracle.Ask(ctx, []byte("BEGIN { sleep 60 }\n"), parseoracle.Options{})
		done <- err
	}()

	// Cancel once the child is demonstrably blocked, so the test measures
	// cancellation rather than whether a deadline beat a normal exit.
	select {
	case err := <-done:
		t.Fatalf("probe did not block in BEGIN, so there was nothing to cancel (err: %v)", err)
	case <-time.After(2 * time.Second):
	}
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected an error from a cancelled context")
		}
	case <-time.After(20 * time.Second):
		t.Fatal("Ask did not return after its context was cancelled")
	}
}

// TestScriptPathIsPackageRelative keeps the oracle callable from anywhere. The
// original resolved parse_facts.pl against the process working directory,
// which works only for a test run from inside this package.
func TestScriptPathIsPackageRelative(t *testing.T) {
	requirePOSIXOracle(t)
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

// TestExpensiveFixtureStillFitsItsBudget keeps the numbers in
// TestShebangPassthrough's comment and on DefaultTimeout honest.
//
// Both cite 32 seconds for op/taint.t through the probe, and both draw a
// conclusion from it: that 60s is too tight a margin and 4m is enough. A
// comment carrying a measurement nothing re-runs is a comment that goes
// quietly wrong -- this one was already wrong when found, claiming every
// corpus file costs "tens of milliseconds".
//
// Asserted loosely and in one direction. The absolute number moves with the
// machine, so this fails only if the cost approaches the budget it justified
// -- which is the moment the conclusion stops holding.
func TestExpensiveFixtureStillFitsItsBudget(t *testing.T) {
	p := perltest.PinnedPerl(t)
	dir := shimT(t, p)
	const taint = "op/taint.t"
	if _, err := os.Stat(filepath.Join(dir, taint)); err != nil {
		t.Skipf("%s absent from shim: %v", taint, err)
	}

	start := time.Now()
	f, err := parseoracle.AskFile(context.Background(), taint,
		parseoracle.Options{Dir: dir, Shebang: true, Timeout: 4 * time.Minute, Perl: p.Path})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("AskFile: %v", err)
	}
	if !f.OK {
		t.Fatalf("op/taint.t must compile with Shebang set, stderr: %s", f.Stderr)
	}

	// Half the budget. Measured at ~32s, so this has 2x headroom for a
	// loaded machine and still fails long before a real wedge.
	if limit := 2 * time.Minute; elapsed > limit {
		t.Errorf("op/taint.t took %v through the probe, over half the %v budget "+
			"TestShebangPassthrough sets. The comments citing ~32s are stale, "+
			"and the reasoning that 4m is enough needs re-checking.",
			elapsed.Round(time.Second), 4*time.Minute)
	}
	t.Logf("op/taint.t through the probe: %v", elapsed.Round(time.Millisecond))
}

// TestShimTRefusesAShimBuiltForAnotherPerl: the directory-dependent tests
// above measure with the pinned perl, so a shim at $ORACLE_SHIM stamped
// for another interpreter (or not stamped at all) is skipped, not measured
// over: its lib/Config.pm belongs to another perl.
func TestShimTRefusesAShimBuiltForAnotherPerl(t *testing.T) {
	p := perltest.Perl{} // the stamp check needs no perl
	shim := t.TempDir()
	if err := os.MkdirAll(filepath.Join(shim, "t", "op"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"test.pl", "op/sub.t"} {
		if err := os.WriteFile(filepath.Join(shim, "t", f), []byte("1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := parseoracle.WriteShimStamp(shim, parseoracle.Pin{Interpreter: "5.000001", Revision: "deadbeef", Threads: true}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PARSEORACLE_SHIM", "")
	t.Setenv("ORACLE_SHIM", filepath.Join(shim, "t"))

	var sub *testing.T
	t.Run("shimT", func(st *testing.T) {
		sub = st
		shimT(st, p)
	})
	if !sub.Skipped() {
		t.Error("shimT handed out a shim stamped for perl 5.000001")
	}
}
