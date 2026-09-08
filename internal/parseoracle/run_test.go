// ABOUTME: Tests the corpus runner: bucket totals, environmental exclusion, determinism, and the WRONG list.
// ABOUTME: The default run touches a five-file fixture; the full corpus sweep is opt-in behind -parseoracle.corpus.

package parseoracle_test

import (
	"context"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parseoracle"
)

// corpusSweep opts into the full 620-file measurement. It is a flag rather
// than the default because an AC bounds the default run at ten seconds, and
// the whole corpus needs rather more than that.
var corpusSweep = flag.Bool("parseoracle.corpus", false,
	"run the oracle across the whole perl5 corpus and print the fidelity table")

// TestCorpusSweep is the milestone's deliverable: how often our parser agrees
// with perl across perl's own test suite. Opt in with
//
//	go test ./internal/parseoracle/ -run TestCorpusSweep -parseoracle.corpus -v
//
// A shim is built into a temp dir, because t/test.pl clears @INC and unshifts
// ../lib: without a populated shim, 498 of 620 files report false failures.
// $PARSEORACLE_SHIM reuses an already-built one.
func TestCorpusSweep(t *testing.T) {
	if !*corpusSweep {
		t.Skip("full corpus sweep: pass -parseoracle.corpus")
	}

	shim := os.Getenv("PARSEORACLE_SHIM")
	if shim == "" {
		root := corpusRoot(t)
		shim = t.TempDir()
		if err := parseoracle.BuildShim(root, shim); err != nil {
			t.Fatalf("BuildShim: %v", err)
		}
	}
	shimT := filepath.Join(shim, "t")

	// Paths are relative to the shim's t/, matching how perl's tests refer to
	// each other and how the report's file names read.
	var files []string
	err := filepath.WalkDir(shimT, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".t") {
			rel, err := filepath.Rel(shimT, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the corpus: %v", err)
	}
	sort.Strings(files)

	start := time.Now()
	report, err := parseoracle.Run(context.Background(), files,
		parseoracle.RunOptions{Dir: shimT})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	elapsed := time.Since(start)

	t.Logf("full corpus run: %d files in %s across %d workers\n%s",
		len(files), elapsed.Round(time.Millisecond), runtime.NumCPU(), report)

	// A ceiling on WRONG is not enough on its own -- see ExactRate's comment.
	// Both are asserted so a failure names which one moved.
	if wrong := report.Wrong(); len(wrong) > 0 {
		t.Errorf("%d file(s) committed to a parse perl did not make: %v", len(wrong), wrong)
	}
	if rate := report.ExactRate(); rate <= 0 {
		t.Errorf("exact rate is %v: a parser that never agrees exactly is not "+
			"vindicated by a WRONG count of zero", rate)
	}
}

// fixtureCorpus is the five-file corpus the default test run measures. It is
// deliberately tiny: the full 620-file sweep is opt-in, because an AC requires
// this test to finish inside ten seconds.
func fixtureCorpus(t *testing.T) []string {
	t.Helper()

	dir := filepath.Join("testdata", "fixture")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading fixture corpus: %v", err)
	}

	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".pl") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	if len(files) != 5 {
		t.Fatalf("fixture corpus has %d files, want 5", len(files))
	}
	return files
}

// TestCorpusRun is the runner's fixture test, and the AC that bounds the
// default run's cost. Five files, every bucket represented, under ten seconds.
func TestCorpusRun(t *testing.T) {
	files := fixtureCorpus(t)

	report, err := parseoracle.Run(context.Background(), files, parseoracle.RunOptions{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(report.Files) != len(files) {
		t.Fatalf("report covers %d files, want %d", len(report.Files), len(files))
	}

	// Bucket totals must account for every file that stayed in the
	// denominator, or the percentages below it are arithmetic on nothing.
	if got := report.Measured(); got != report.Totals.Sum() {
		t.Errorf("bucket totals sum to %d but %d files were measured", report.Totals.Sum(), got)
	}

	// The fixture was built so each bucket is exercised. A runner that
	// reports one bucket for everything would still pass a count check.
	for _, want := range []struct {
		file   string
		bucket parseoracle.Bucket
	}{
		{"exact.pl", parseoracle.BucketExact},
		{"explicit_ref.pl", parseoracle.BucketExact},
		{"wider.pl", parseoracle.BucketWider},
		{"noanswer.pl", parseoracle.BucketNoAnswer},
	} {
		r := findResult(t, report, want.file)
		if r.Verdict.Bucket != want.bucket {
			t.Errorf("%s bucketed %s, want %s (%s)",
				want.file, r.Verdict.Bucket, want.bucket, r.Verdict.Detail)
		}
	}
}

// findResult locates one file's result by basename.
func findResult(t *testing.T, report parseoracle.Report, base string) parseoracle.Result {
	t.Helper()
	for _, r := range report.Files {
		if filepath.Base(r.Path) == base {
			return r
		}
	}
	t.Fatalf("no result for %s in report", base)
	return parseoracle.Result{}
}

// TestRunExcludesEnvironmental is the guard against the specific failure this
// harness exists to avoid: blaming the parser for the machine. A file that
// fails because a module is missing says nothing about how Perl parses, so it
// leaves the denominator and is counted on its own.
func TestRunExcludesEnvironmental(t *testing.T) {
	files := fixtureCorpus(t)

	report, err := parseoracle.Run(context.Background(), files, parseoracle.RunOptions{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	env := findResult(t, report, "environmental.pl")
	if env.Failure != parseoracle.Environmental {
		t.Fatalf("environmental.pl classified %s, want Environmental (stderr: %s)",
			env.Failure, env.Facts.Stderr)
	}
	if !env.Excluded {
		t.Error("an environmental failure must be excluded from the denominator")
	}
	if report.Environmental != 1 {
		t.Errorf("report counts %d environmental failures, want 1", report.Environmental)
	}
	if report.Measured() != len(files)-1 {
		t.Errorf("denominator is %d, want %d — the environmental failure is still being counted",
			report.Measured(), len(files)-1)
	}
	// The excluded file must not appear in any bucket total either.
	if report.Totals.Sum() != len(files)-1 {
		t.Errorf("bucket totals sum to %d, want %d", report.Totals.Sum(), len(files)-1)
	}
}

// TestRunIsDeterministic runs twice and diffs the rendered reports. The runner
// is parallel, so without an ordering step the file list arrives in whatever
// order the workers finished — and a report that reorders between runs cannot
// be committed as a baseline, which is what the next issue needs from it.
func TestRunIsDeterministic(t *testing.T) {
	files := fixtureCorpus(t)

	first, err := parseoracle.Run(context.Background(), files, parseoracle.RunOptions{})
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}
	second, err := parseoracle.Run(context.Background(), files, parseoracle.RunOptions{})
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}

	if a, b := first.String(), second.String(); a != b {
		t.Errorf("two runs produced different reports:\n--- first ---\n%s\n--- second ---\n%s", a, b)
	}
}

// TestReportNamesWrongFiles: the list of WRONG files is more useful than the
// percentage. A report that says "3 WRONG" sends a reader back to the corpus
// to find which three.
func TestReportNamesWrongFiles(t *testing.T) {
	report := parseoracle.Report{
		Files: []parseoracle.Result{
			{Path: "t/op/fine.t", Verdict: parseoracle.Verdict{Bucket: parseoracle.BucketExact}},
			{Path: "t/op/lying.t", Verdict: parseoracle.Verdict{
				Bucket: parseoracle.BucketWrong,
				Marker: parseoracle.MarkerSrefgen,
				Detail: "perl took 1 reference(s) the source did not write",
			}},
		},
	}
	report.Totals[parseoracle.BucketExact] = 1
	report.Totals[parseoracle.BucketWrong] = 1

	if names := report.Wrong(); len(names) != 1 || names[0] != "t/op/lying.t" {
		t.Errorf("Wrong() = %v, want [t/op/lying.t]", names)
	}

	rendered := report.String()
	if !strings.Contains(rendered, "t/op/lying.t") {
		t.Errorf("the report must name each WRONG file individually, got:\n%s", rendered)
	}
	if strings.Contains(rendered, "t/op/fine.t") {
		t.Error("only the WRONG files are named individually; listing all of them buries the signal")
	}
	// The reason matters as much as the name: a bare filename sends the
	// reader back to run the comparison by hand.
	if !strings.Contains(rendered, "did not write") {
		t.Errorf("the WRONG list must carry each verdict's detail, got:\n%s", rendered)
	}
}

// TestRunBoundsAWedgedFile is what keeps one pathological file from stalling a
// whole corpus run. It is not hypothetical: t/re/pat_psycho.t calls test.pl's
// `watchdog(5 * 60)` from a BEGIN block, so `perl -c` FORKS a monitor process
// that inherits the pipe the oracle reads. Measured, that file is the sweep's
// one runner error, and it costs exactly the timeout and nothing more.
//
// A run must therefore charge a wedged file its timeout and move on, rather
// than waiting out a grandchild's own lifetime.
func TestRunBoundsAWedgedFile(t *testing.T) {
	dir := t.TempDir()
	// A compile-time fork that outlives its parent and inherits its pipes,
	// which is the shape watchdog() creates.
	const src = "BEGIN { my $pid = fork(); if (!$pid) { sleep 60; exit } }\nmy $x = 1;\n"
	path := filepath.Join(dir, "wedged.pl")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	// Two seconds is plenty for a two-line file, and far less than the 60s
	// the forked child lives.
	const budget = 2 * time.Second
	start := time.Now()
	report, err := parseoracle.Run(context.Background(), []string{path},
		parseoracle.RunOptions{Dir: dir, Timeout: budget})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	t.Logf("Run returned after %s", elapsed.Round(time.Millisecond))
	if elapsed > 10*time.Second {
		t.Errorf("Run took %s with a %s per-file timeout: a compile-time fork "+
			"is holding the output pipe open, so one wedged corpus file stalls "+
			"the whole run", elapsed.Round(time.Millisecond), budget)
	}
	// A file the runner could not measure is an error, never a verdict.
	// Bucketing it would put a timeout in the fidelity numbers.
	if report.Errors != 1 {
		t.Errorf("report counts %d runner errors, want 1", report.Errors)
	}
	if report.Measured() != 0 {
		t.Errorf("a file the oracle could not answer for must not reach a bucket, got %d measured",
			report.Measured())
	}
}

// TestReportAssertsExactFloor is the trap the previous issue documented on
// BucketWider. Our grammar emits `ambiguous_function_call_expression`, which
// honestly declines to commit, so every prototype-driven call buckets wider
// rather than WRONG. A parser that hedged on EVERYTHING would therefore score
// 100% non-WRONG and mean nothing.
//
// So the report exposes an exact RATE, and the corpus run asserts a floor on
// it — not merely a ceiling on WRONG.
func TestReportAssertsExactFloor(t *testing.T) {
	// The degenerate parser: hedges on everything, commits to nothing.
	hedging := parseoracle.Report{}
	hedging.Totals[parseoracle.BucketWider] = 100

	if hedging.Wrong() != nil {
		t.Fatal("premise: the degenerate parser has no WRONG files")
	}
	if rate := hedging.ExactRate(); rate != 0 {
		t.Errorf("ExactRate of an all-hedging parser = %v, want 0", rate)
	}

	honest := parseoracle.Report{}
	honest.Totals[parseoracle.BucketExact] = 75
	honest.Totals[parseoracle.BucketWider] = 25
	if rate := honest.ExactRate(); rate != 0.75 {
		t.Errorf("ExactRate = %v, want 0.75", rate)
	}

	// no-answer is in the denominator: a parser that emits an error node for
	// everything must not score 100% exact on the one file it managed.
	partial := parseoracle.Report{}
	partial.Totals[parseoracle.BucketExact] = 1
	partial.Totals[parseoracle.BucketNoAnswer] = 99
	if rate := partial.ExactRate(); rate != 0.01 {
		t.Errorf("ExactRate = %v, want 0.01 — no-answer must stay in the denominator", rate)
	}
}
