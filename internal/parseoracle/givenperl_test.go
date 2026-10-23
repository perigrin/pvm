// ABOUTME: Tests that every perl the oracle starts is the one it is given, never whichever perl is first on PATH.
// ABOUTME: Each test writes its own pin and puts a decoy of another version first on PATH, so passing proves the path was used.

package parseoracle_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/perltest"
)

// decoyVersion is what the decoy perl claims to be. No real perl reports
// it, so a measurement that came from the decoy cannot pass for one that
// came from the perl the test gave.
const decoyVersion = "5.000001"

// givenPerl finds a real perl through a pin of the test's own: the perl
// first on PATH is measured, a temp corpus.pin is written naming it, and
// PinnedPerlAt finds a perl matching that pin. It returns that perl, the
// pin, and the pin's path. The committed pin is never
// read, so the test means the same thing whichever perl it names.
func givenPerl(t *testing.T) (perltest.Perl, parseoracle.Pin, string) {
	t.Helper()
	requirePOSIXOracle(t)
	path, err := exec.LookPath("perl")
	if err != nil {
		t.Skipf("no perl on PATH to measure: %v", err)
	}
	pin, err := parseoracle.ObservePin(context.Background(), path, "deadbeef")
	if err != nil {
		t.Fatalf("ObservePin(%s): %v", path, err)
	}
	pinPath := filepath.Join(t.TempDir(), "corpus.pin")
	if err := parseoracle.WritePin(pinPath, pin); err != nil {
		t.Fatal(err)
	}
	return perltest.PinnedPerlAt(t, pinPath), pin, pinPath
}

// decoyOnPath puts first on PATH a perl that claims to be decoyVersion
// and then fails, so anything that runs `perl` from PATH either fails or
// reports a version no real perl has.
func decoyOnPath(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n%s\\n' '" + decoyVersion + "' define\nexit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "perl"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// TestRunUsesTheGivenPerl: the oracle's INNER compiles, not just the
// process that runs parse_facts.pl, use the perl in RunOptions.Perl.
// Measured before this was so: under 5.44 with 5.42 first on PATH,
// parse_facts.pl's backticked `perl` ran 5.42, and op/for-many.t came
// back as 5.42's syntax error instead of 5.44's ok.
//
// The fixture's own BEGIN records $] once per compile that runs it, so a
// fake perl printing canned JSON cannot pass: the version comes from the
// file, from inside each inner compile.
func TestRunUsesTheGivenPerl(t *testing.T) {
	p, pin, _ := givenPerl(t)
	subject := goSubject(t)
	decoyOnPath(t)

	dir := t.TempDir()
	log := filepath.Join(dir, "versions.log")
	t.Setenv("ORACLE_VERSION_LOG", log)
	src := "BEGIN { open my $fh, '>>', $ENV{ORACLE_VERSION_LOG} or die $!; print $fh \"$]\\n\"; close $fh }\n" +
		"my @a = (1, 2);\nmy $r = \\@a;\n"
	file := filepath.Join(dir, "version.pl")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := parseoracle.Run(context.Background(), []string{file},
		parseoracle.RunOptions{Perl: p.Path, Subject: subject})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	r := report.Files[0]
	if r.Err != "" {
		t.Fatalf("the oracle failed under %s with a decoy first on PATH: %s", p.Path, r.Err)
	}
	if !r.Facts.OK || !r.Facts.Walked {
		t.Errorf("facts are ok=%v walked=%v, want both: an inner compile ran the decoy (stderr %q)",
			r.Facts.OK, r.Facts.Walked, r.Facts.Stderr)
	}

	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("no inner compile ran the fixture's BEGIN: %v", err)
	}
	versions := strings.Fields(string(data))
	if len(versions) != 2 {
		t.Errorf("the fixture's BEGIN ran %d time(s), want 2 (the Concise compile and the probe): %q",
			len(versions), versions)
	}
	for _, v := range versions {
		if v != pin.Interpreter {
			t.Errorf("an inner compile ran perl %s, want the given %s's %s", v, p.Path, pin.Interpreter)
		}
	}
}

// TestObservePinUsesTheGivenPerl: ObservePin asks the perl it is given,
// so with a decoy of another version first on PATH it still reports the
// interpreter the temp pin names. The perl is PATH's own, found before the
// decoy goes in, rather than PinnedPerlAt's: PinnedPerlAt itself trusts
// ObservePin, so a broken ObservePin would only make it skip.
func TestObservePinUsesTheGivenPerl(t *testing.T) {
	requirePOSIXOracle(t)
	path, err := exec.LookPath("perl")
	if err != nil {
		t.Skipf("no perl on PATH to measure: %v", err)
	}
	measured, err := parseoracle.ObservePin(context.Background(), path, "deadbeef")
	if err != nil {
		t.Fatalf("ObservePin(%s): %v", path, err)
	}
	pinPath := filepath.Join(t.TempDir(), "corpus.pin")
	if err := parseoracle.WritePin(pinPath, measured); err != nil {
		t.Fatal(err)
	}
	pin, err := parseoracle.ReadPin(pinPath)
	if err != nil {
		t.Fatal(err)
	}
	decoyOnPath(t)

	observed, err := parseoracle.ObservePin(context.Background(), path, pin.Revision)
	if err != nil {
		t.Fatalf("ObservePin(%s) with a decoy first on PATH: %v", path, err)
	}
	if observed != pin {
		t.Errorf("ObservePin(%s) = %+v, want the temp pin %+v", path, observed, pin)
	}
}

// TestOpenCacheUsesTheGivenPerl: the cache's identity probe asks the perl
// the cache serves, so the key names the interpreter that measured. With
// a decoy that fails first on PATH, opening the cache for the given perl
// still succeeds and records its $].
func TestOpenCacheUsesTheGivenPerl(t *testing.T) {
	p, pin, _ := givenPerl(t)
	decoyOnPath(t)

	dir := t.TempDir()
	cache, err := parseoracle.OpenCache(dir, p.Path)
	if err != nil {
		t.Fatalf("OpenCache(%s) with a decoy first on PATH: %v", p.Path, err)
	}
	src := filepath.Join(t.TempDir(), "probe.pl")
	if err := os.WriteFile(src, []byte("my $x = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AskFile(context.Background(), src, parseoracle.Options{Perl: p.Path}); err != nil {
		t.Fatalf("AskFile: %v", err)
	}
	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("want one cache entry, got %v (%v)", entries, err)
	}
	data, err := os.ReadFile(entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"interpreter": "`+pin.Interpreter+`"`) {
		t.Errorf("the cache entry does not record %s's %s:\n%s", p.Path, pin.Interpreter, data)
	}
}

// TestRatchetPinObservedWithPinnedPerl: the pin check TestRatchetCorpus
// performs, run on the gate's five-file fixture against a temp pin with a
// decoy of another version first on PATH, observes the temp pin's
// interpreter, and the baseline that sweep would write passes CheckPin.
// Before ObservePin took a path, this observed whichever perl PATH named.
func TestRatchetPinObservedWithPinnedPerl(t *testing.T) {
	p, pin, pinPath := givenPerl(t)
	shim := t.TempDir()
	buildFixtureShim(t, p, shim)
	subject := goSubject(t)
	decoyOnPath(t)

	if !parseoracle.ShimReady(shim, pin) {
		t.Fatalf("the fixture shim built for %s is not ready for the pin naming it", p.Path)
	}
	shimT := filepath.Join(shim, "t")
	files, err := filepath.Glob(filepath.Join(shimT, "*.t"))
	if err != nil || len(files) == 0 {
		t.Fatalf("the fixture shim has no .t files: %v", err)
	}
	for i, f := range files {
		files[i] = filepath.Base(f)
	}
	report, err := parseoracle.Run(context.Background(), files,
		parseoracle.RunOptions{Dir: shimT, Subject: subject, Perl: p.Path})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.Errors != 0 {
		t.Errorf("%d runner error(s) over the fixture shim: an oracle ran the decoy\n%s", report.Errors, report)
	}

	observed := observedPin(t, p, pinPath)
	if observed.Interpreter != pin.Interpreter || observed.Threads != pin.Threads {
		t.Errorf("the sweep observed %+v with a decoy (%s) first on PATH, want the temp pin's %+v",
			observed, decoyVersion, pin)
	}
	if err := parseoracle.NewBaseline(pin, report, shimT).CheckPin(observed); err != nil {
		t.Errorf("the baseline this sweep would write fails its own pin check: %v", err)
	}
}
