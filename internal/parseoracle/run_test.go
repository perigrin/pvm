// ABOUTME: Tests the corpus runner: bucket totals, environmental exclusion, determinism, and the WRONG list.
// ABOUTME: The default run touches a five-file fixture; the full corpus sweep is opt-in behind -parseoracle.corpus.

package parseoracle_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

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
