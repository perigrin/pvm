// ABOUTME: Tests the ratchet: that it fails on a regression, fails on an unrecorded improvement, and rewrites only under -update.
// ABOUTME: The point is the failures — perl-lsp shipped a ratchet that could not fail, which is decoration.

package parseoracle_test

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// updateBaseline is the ONLY path that rewrites a committed baseline. It is a
// flag rather than an environment variable so that a `go test` invocation
// records in its own command line that it intended to re-baseline — and so
// that a plain run can never do it by accident, which is what turns a ratchet
// into decoration.
var updateBaseline = flag.Bool("parseoracle.update", false,
	"rewrite the committed ratchet baseline from this run's verdicts")

// corpusReport measures the whole corpus through a shim, the same way
// TestCorpusSweep does. $PARSEORACLE_SHIM names a shim to reuse, which is
// what makes a re-baseline affordable at all.
//
// A named shim that is not built yet is BUILT there rather than assumed
// ready. That is the cache-miss path: CI restores $PARSEORACLE_SHIM from
// actions/cache, and on a miss the directory exists but is empty. Treating an
// empty directory as a corpus would walk zero files, and a run with zero
// files does not fail loudly — it fails as 620 "baseline row has no file in
// this run" lines, which reads like the corpus vanished rather than like the
// cache missed.
//
// The shim's t/ is returned alongside the report because the runner records
// paths relative to it, and the taxonomy has to re-read each file.
func corpusReport(t *testing.T) (parseoracle.Report, string) {
	t.Helper()

	shim := os.Getenv("PARSEORACLE_SHIM")
	if shim == "" {
		shim = t.TempDir()
	}
	if _, err := os.Stat(filepath.Join(shim, "t", "test.pl")); err != nil {
		root := corpusRoot(t)
		if err := parseoracle.BuildShim(root, shim); err != nil {
			t.Fatalf("BuildShim into %s: %v", shim, err)
		}
	}
	shimT := filepath.Join(shim, "t")

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

	report, err := parseoracle.Run(context.Background(), files,
		parseoracle.RunOptions{Dir: shimT, Timeout: sweepTimeout(t)})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return report, shimT
}

// runFixture measures the six-file fixture corpus that TestCorpusRun uses.
func runFixture(t *testing.T) parseoracle.Report {
	t.Helper()
	report, err := parseoracle.Run(context.Background(), fixtureCorpus(t), parseoracle.RunOptions{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return report
}

// downgradeOneFile synthesises a regression: the first file that reached a
// verdict is moved to WRONG, which is the worst bucket and therefore
// unambiguously a regression from anything else.
func downgradeOneFile(t *testing.T, report parseoracle.Report) parseoracle.Report {
	t.Helper()
	out := cloneReport(report)
	for i := range out.Files {
		f := &out.Files[i]
		if f.Excluded || f.Err != "" || f.Verdict.Bucket == parseoracle.BucketWrong {
			continue
		}
		out.Totals[f.Verdict.Bucket]--
		f.Verdict.Bucket = parseoracle.BucketWrong
		f.Verdict.Detail = "synthetic regression"
		out.Totals[parseoracle.BucketWrong]++
		return out
	}
	t.Fatal("no fixture file could be downgraded: the fixture is not exercising the ratchet")
	return out
}

// upgradeOneFile synthesises an improvement: a file that declined is moved to
// exact, which is the best bucket.
func upgradeOneFile(t *testing.T, report parseoracle.Report) parseoracle.Report {
	t.Helper()
	out := cloneReport(report)
	for i := range out.Files {
		f := &out.Files[i]
		if f.Excluded || f.Err != "" || f.Verdict.Bucket == parseoracle.BucketExact {
			continue
		}
		out.Totals[f.Verdict.Bucket]--
		f.Verdict.Bucket = parseoracle.BucketExact
		f.Verdict.Detail = "synthetic improvement"
		out.Totals[parseoracle.BucketExact]++
		return out
	}
	t.Fatal("no fixture file could be upgraded: the fixture is not exercising the ratchet")
	return out
}

func cloneReport(r parseoracle.Report) parseoracle.Report {
	out := r
	out.Files = append([]parseoracle.Result(nil), r.Files...)
	return out
}

func asRatchetError(err error, target **parseoracle.RatchetError) bool {
	return errors.As(err, target)
}

// fixtureBaselinePath is the small committed baseline the ratchet tests gate
// against. It is a six-file fixture rather than the 620-file corpus because
// a full sweep costs ~6 minutes: a ratchet nobody can afford to run is a
// ratchet nobody runs.
func fixtureBaselinePath() string {
	return filepath.Join("testdata", "ratchet", "fixture.txt")
}

// loadFixtureBaseline reads the committed fixture baseline.
func loadFixtureBaseline(t *testing.T) parseoracle.Baseline {
	t.Helper()
	base, err := parseoracle.LoadBaseline(fixtureBaselinePath())
	if err != nil {
		t.Fatalf("LoadBaseline(%s): %v", fixtureBaselinePath(), err)
	}
	return base
}

// TestRatchet is the gate: today's fixture verdicts must match the committed
// baseline exactly. This is the invocation a developer runs, and the one that
// tells them the baseline still describes reality.
func TestRatchet(t *testing.T) {
	report := fixtureReport(t)

	if *updateBaseline {
		pin, err := parseoracle.ReadPin(filepath.Join("testdata", "corpus.pin"))
		if err != nil {
			t.Fatalf("ReadPin: %v", err)
		}
		if err := parseoracle.WriteBaseline(fixtureBaselinePath(),
			parseoracle.NewBaseline(pin, report, "")); err != nil {
			t.Fatalf("WriteBaseline: %v", err)
		}
		t.Logf("rewrote %s from this run; commit it in the same commit as the change that moved it",
			fixtureBaselinePath())
		return
	}

	base := loadFixtureBaseline(t)
	if err := base.Check(report); err != nil {
		t.Fatalf("the fixture no longer matches its committed baseline:\n%v\n\n"+
			"if this is a deliberate change, re-run with -parseoracle.update", err)
	}
}

// fixtureReport measures the six-file fixture corpus.
func fixtureReport(t *testing.T) parseoracle.Report {
	t.Helper()
	return runFixture(t)
}

// corpusBaselinePath is the full 620-file baseline: the frozen form of
// findings §0.11. It is gated behind -parseoracle.corpus because the sweep
// costs ~6 minutes, which is why TestRatchet above gates on the fixture
// instead — but the corpus numbers are the ones that matter, so they are
// frozen too rather than left as a paragraph in a document that decays.
func corpusBaselinePath() string {
	return filepath.Join("testdata", "ratchet", "baseline.txt")
}

// TestCorpusBaselineIsIntact guards the artefact this whole milestone exists
// to produce.
//
// Every other assertion about a committed baseline in this file reads the
// six-row fixture. The real 620-row file was read by exactly one test,
// TestRatchetCorpus, which skips unless -parseoracle.corpus is passed — so
// truncating baseline.txt to its four-line header, deleting every verdict,
// left `go test ./internal/parseoracle/` reporting ok. The file the ratchet
// protects could be destroyed with nothing objecting: perl-lsp's
// `ci/parse_errors_baseline.txt` containing the single line `0`, arrived at
// from the other direction.
//
// This runs unconditionally, because the gating was the defect. It reads the
// committed file only — no perl, no sweep.
func TestCorpusBaselineIsIntact(t *testing.T) {
	base, err := parseoracle.LoadBaseline(corpusBaselinePath())
	if err != nil {
		t.Fatalf("LoadBaseline(%s): %v", corpusBaselinePath(), err)
	}

	// The corpus is 620 .t files (findings §0.11, spec §7.3). An exact
	// count would fail on every legitimate corpus bump, and a floor of 1
	// would pass a file with one row left in it, so the band is wide
	// enough to survive a pin move and narrow enough that a truncation
	// cannot hide inside it.
	const wantRows = 620
	if len(base.Rows) < wantRows/2 {
		t.Fatalf("the committed corpus baseline has %d rows, want ~%d: a baseline "+
			"this short has been truncated or emptied, and it is the artefact "+
			"the ratchet exists to protect", len(base.Rows), wantRows)
	}
	if len(base.Rows) > wantRows*2 {
		t.Errorf("the committed corpus baseline has %d rows, want ~%d: the corpus "+
			"has grown beyond recognition or rows have been duplicated",
			len(base.Rows), wantRows)
	}

	// Every row must be a verdict the ratchet can reason about. A row
	// whose status rank() does not know is a row that can never be
	// compared, which is a silent hole in the denominator.
	counts := make(map[string]int, len(base.Rows))
	paths := make(map[string]string, len(base.Rows))
	for _, row := range base.Rows {
		counts[row.Status]++
		if first, dup := paths[row.Path]; dup {
			t.Errorf("%s appears twice (%s and %s): a duplicated path means one "+
				"verdict silently shadows the other", row.Path, first, row.Status)
		}
		paths[row.Path] = row.Status
		if !parseoracle.ValidCategory(row.Category) {
			t.Errorf("%s has category %q, which is not in the fixed taxonomy",
				row.Path, row.Category)
		}
	}

	// The rows must support the totals recorded in findings §0.11. The doc
	// is the independent record — a reader's claim about what was measured
	// — and a baseline whose header or prose claims totals its rows do not
	// support is a baseline that lies about its own contents. Checking the
	// rows against themselves would assert nothing.
	for _, want := range findingsTotals(t) {
		if got := counts[want.status]; got != want.count {
			t.Errorf("the baseline holds %d %s rows, but findings §0.11 records %d: "+
				"the committed verdicts no longer support the totals the "+
				"documentation claims for them", got, want.status, want.count)
		}
	}

	// WRONG is the one bucket that is a gate rather than a ratchet: a file
	// our parser gets positively wrong is a defect, not a coverage gap.
	if n := counts[parseoracle.BucketWrong.String()]; n != 0 {
		t.Errorf("%d file(s) are baselined as WRONG: that bucket is a gate, "+
			"not a ratchet", n)
	}
}

// findingsBucketTotal is one bucket count as the findings document records it.
type findingsBucketTotal struct {
	status string
	count  int
}

// findingsTotals reads the §0.11 corpus table out of the findings document.
//
// Parsed rather than duplicated as constants here: a copy in the test would
// drift from the document silently, and the point is to hold the two
// together. The rows are `| bucket | count | share |`, with the WRONG row
// bolded.
func findingsTotals(t *testing.T) []findingsBucketTotal {
	t.Helper()

	path := repoFile(t, "docs", "specs", "perl-parser", "00-findings.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the findings document: %v", err)
	}

	wanted := map[string]string{
		"exact":                   parseoracle.BucketExact.String(),
		"wider":                   parseoracle.BucketWider.String(),
		"WRONG":                   parseoracle.BucketWrong.String(),
		"no-answer":               parseoracle.BucketNoAnswer.String(),
		"excluded, environmental": "excluded",
		"runner error":            "error",
	}

	var totals []findingsBucketTotal
	for _, line := range strings.Split(string(data), "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		label := strings.Trim(strings.TrimSpace(cells[1]), "*")
		status, ok := wanted[label]
		if !ok {
			continue
		}
		count, err := strconv.Atoi(strings.Trim(strings.TrimSpace(cells[2]), "*"))
		if err != nil {
			continue
		}
		totals = append(totals, findingsBucketTotal{status: status, count: count})
		delete(wanted, label)
	}

	// Every bucket must have been found. A parser that silently matched
	// nothing would make the assertion above vacuous, which is the same
	// class of defect this test was written to close.
	if len(wanted) != 0 {
		t.Fatalf("findings §0.11 has no row for %v: either the table moved or "+
			"this parser stopped matching it, and an unmatched table makes the "+
			"totals assertion vacuous", wanted)
	}
	return totals
}

// TestRatchetCorpus is the same ratchet at corpus scale.
//
//	PARSEORACLE_SHIM=/tmp/oracletree PERL5_CORPUS=~/dev/perl5 \
//	  go test ./internal/parseoracle/ -run TestRatchetCorpus -parseoracle.corpus
//
// Add -parseoracle.update to re-baseline. Both flags are required together
// for a rewrite, so neither a plain run nor a plain sweep can move it.
//
// $PARSEORACLE_BASELINE substitutes another baseline, and $PARSEORACLE_RECEIPT
// names a file to write a receipt to once the ratchet has passed. Both exist
// for the CI gate: the first lets TestCIGateRunsTheSweep run the workflow's
// own step over a fixture, the second is what the workflow's next step
// refuses to proceed without. Neither changes what a plain run does.
func TestRatchetCorpus(t *testing.T) {
	if !*corpusSweep {
		t.Skip("full corpus ratchet: pass -parseoracle.corpus")
	}

	baselinePath := corpusBaselinePath()
	if override := os.Getenv(baselineEnv); override != "" {
		baselinePath = override
	}

	report, shimT := corpusReport(t)
	pin, err := parseoracle.ReadPin(filepath.Join("testdata", "corpus.pin"))
	if err != nil {
		t.Fatalf("ReadPin: %v", err)
	}

	// What perl ACTUALLY is, not what the pin says it should be. Checking the
	// baseline against a pin read from the same file it was written with
	// compares a value to a copy of itself and cannot fail; the first real CI
	// run passed that check on a perl whose build differed from the
	// baseline's. A re-baseline records the observed world for the same
	// reason: it must describe the perl that produced the verdicts.
	observed, err := parseoracle.ObservePin(context.Background(), pin.Revision)
	if err != nil {
		t.Fatalf("ObservePin: %v", err)
	}

	if *updateBaseline {
		if err := parseoracle.WriteBaseline(baselinePath,
			parseoracle.NewBaseline(observed, report, shimT)); err != nil {
			t.Fatalf("WriteBaseline: %v", err)
		}
		t.Logf("rewrote %s: %s", baselinePath, summarise(report))
		return
	}

	base, err := parseoracle.LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}
	// Check is a symmetric diff, so an empty report against an empty
	// baseline agrees perfectly. That is a shim with no .t files gating a
	// baseline truncated to its header -- perl-lsp's ratchet, whose baseline
	// says `0` -- and it must fail here rather than pass on nothing.
	if len(report.Files) == 0 || len(base.Rows) == 0 {
		t.Fatalf("the sweep found %d files and %s has %d rows: a ratchet over an "+
			"empty denominator agrees on nothing", len(report.Files), baselinePath, len(base.Rows))
	}
	// Skew first: comparing verdicts across a moved pin measures the version
	// bump, not the parser, and reporting that as hundreds of regressions is
	// how a ratchet earns its reputation for crying wolf.
	if err := base.CheckPin(observed); err != nil {
		t.Fatalf("%v", err)
	}
	if err := base.Check(report); err != nil {
		t.Fatalf("the corpus no longer matches its committed baseline:\n%v", err)
	}

	// The receipt is written last, after everything above has passed, so
	// its presence means one thing: this test ran to completion.
	if path := os.Getenv(receiptEnv); path != "" {
		receipt, err := parseoracle.NewReceipt(t.Name(), baselinePath, report)
		if err != nil {
			t.Fatalf("NewReceipt: %v", err)
		}
		if err := receipt.Write(path); err != nil {
			t.Fatalf("writing the receipt: %v", err)
		}
		t.Logf("receipt: %d files against %d rows of %s -> %s",
			receipt.FilesSwept, receipt.BaselineRows, baselinePath, path)
	}
}

// summarise renders the bucket totals for a log line.
func summarise(r parseoracle.Report) string {
	return strings.TrimSpace(r.String())
}

// TestRatchetFailsOnRegression is the property perl-lsp's version lost. Their
// parser_ratchet.rs has its measurements disabled and their baseline file
// contains the single line `0`, so it passes unconditionally. A ratchet that
// cannot fail is decoration, so this proves ours bites: downgrade one file's
// recorded verdict's counterpart in the report and the check must fail.
func TestRatchetFailsOnRegression(t *testing.T) {
	base := loadFixtureBaseline(t)
	report := fixtureReport(t)

	// Synthesise the regression in the REPORT, not the baseline: that is the
	// direction reality moves when the parser breaks.
	regressed := downgradeOneFile(t, report)

	err := base.Check(regressed)
	if err == nil {
		t.Fatal("the ratchet passed a synthetic regression — it cannot fail, so it is decoration")
	}
	var diff *parseoracle.RatchetError
	if !asRatchetError(err, &diff) {
		t.Fatalf("Check returned %T, want *parseoracle.RatchetError", err)
	}
	if len(diff.Regressions) == 0 {
		t.Errorf("a downgraded file produced no regressions: %v", err)
	}
	if len(diff.Improvements) != 0 {
		t.Errorf("a downgrade must not read as an improvement: %v", diff.Improvements)
	}
	// The message must say what broke, not merely that something did.
	if !strings.Contains(err.Error(), "regressed") {
		t.Errorf("a regression message must say a file regressed, got:\n%v", err)
	}
}

// TestRatchetFailsOnImprovement is deliberate, and it is the half people get
// wrong. Conformance plan §4.1: "Any file that regresses fails the build. Any
// file that improves requires a baseline update in the same commit." The
// property protected is that the baseline never silently diverges from
// reality in EITHER direction — a baseline that quietly lags improvements is
// a baseline nobody trusts.
func TestRatchetFailsOnImprovement(t *testing.T) {
	base := loadFixtureBaseline(t)
	report := fixtureReport(t)

	improved := upgradeOneFile(t, report)

	err := base.Check(improved)
	if err == nil {
		t.Fatal("the ratchet passed an unrecorded improvement — the baseline is now allowed to lag reality")
	}
	var diff *parseoracle.RatchetError
	if !asRatchetError(err, &diff) {
		t.Fatalf("Check returned %T, want *parseoracle.RatchetError", err)
	}
	if len(diff.Improvements) == 0 {
		t.Errorf("an upgraded file produced no improvements: %v", err)
	}
	if len(diff.Regressions) != 0 {
		t.Errorf("an improvement must not read as a regression: %v", diff.Regressions)
	}
	// The two directions must not read alike: an improvement is an
	// instruction to re-baseline, not a bug report. The message names the
	// flag exactly as it is typed, so a reader can paste it.
	if !strings.Contains(err.Error(), "-parseoracle.update") {
		t.Errorf("an improvement must tell the author to re-run with -parseoracle.update, got:\n%v", err)
	}
	if !strings.Contains(err.Error(), "same commit") {
		t.Errorf("an improvement must say the new baseline goes in the SAME commit, got:\n%v", err)
	}
	if strings.Contains(err.Error(), "regressed") {
		t.Errorf("an improvement must not be reported as a regression, got:\n%v", err)
	}
}

// TestRatchetLeavingTheDenominatorIsARegression covers the branch whose
// comment says it exists so that a broken shim cannot read as progress.
//
// A file that stops being measured — the shim lost its library, perl became
// unspawnable, the runner timed out — has not got better. But `excluded` and
// `error` sit outside rank's ordering, so without the !wasMeasured guard the
// comparison falls through to the default arm and scores the move as an
// *improvement*: the ratchet then tells the author to re-baseline and record
// the gain. That is precisely how a shim that measures nothing launders
// itself into the committed baseline.
//
// The guard had no test. Replacing its arm with a condition that never fires
// left the whole suite green.
func TestRatchetLeavingTheDenominatorIsARegression(t *testing.T) {
	// Synthetic rather than measured: the property is about rank's
	// ordering, not about any real file, and a unit test that needs no
	// perl sweep is one that actually gets run.
	const path = "t/op/leaves.t"

	for _, tc := range []struct {
		name       string
		from, to   string
		leavingNow bool
	}{
		// Leaving: was scored, now is not. The direction a shim breaks.
		{"exact to error", "exact", "error", true},
		{"exact to excluded", "exact", "excluded", true},
		{"no-answer to error", "no-answer", "error", true},
		{"wrong to excluded", "WRONG", "excluded", true},

		// Entering: was not scored, now is. Also not an improvement —
		// the measurement changed shape and a human should look, even
		// though the new verdict happens to be the best bucket there is.
		{"error to exact", "error", "exact", false},
		{"excluded to exact", "excluded", "exact", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := parseoracle.Baseline{
				Pin: parseoracle.Pin{Interpreter: "5.042000", Revision: "deadbeef"},
				Rows: []parseoracle.Row{{
					Status:   tc.from,
					Category: parseoracle.CategoryNone,
					Path:     path,
				}},
			}

			err := base.Check(reportWithStatus(t, path, tc.to))
			if err == nil {
				t.Fatalf("%s -> %s passed the ratchet: a file that left the "+
					"measured population must never go unreported", tc.from, tc.to)
			}
			var diff *parseoracle.RatchetError
			if !asRatchetError(err, &diff) {
				t.Fatalf("Check returned %T, want *parseoracle.RatchetError", err)
			}

			// The assertion that bites: it must land in Regressions.
			// Merely failing is not enough, because an unrecorded
			// improvement fails too — and tells the author to
			// re-baseline, which would freeze the broken measurement.
			if len(diff.Improvements) != 0 {
				t.Errorf("%s -> %s scored as an IMPROVEMENT (%v): a file leaving "+
					"the denominator would instruct the author to re-baseline, "+
					"laundering a broken shim into the committed verdicts",
					tc.from, tc.to, diff.Improvements)
			}
			if len(diff.Regressions) != 1 {
				t.Fatalf("%s -> %s produced %d regressions, want exactly 1: %v",
					tc.from, tc.to, len(diff.Regressions), err)
			}
			if got := diff.Regressions[0]; got.From != tc.from || got.To != tc.to {
				t.Errorf("regression records %s -> %s, want %s -> %s",
					got.From, got.To, tc.from, tc.to)
			}
			if !strings.Contains(err.Error(), "regressed") {
				t.Errorf("the message must read as a regression, got:\n%v", err)
			}
			if strings.Contains(err.Error(), "-parseoracle.update") {
				t.Errorf("a file leaving the denominator must NOT be reported as "+
					"something to re-baseline away, got:\n%v", err)
			}
		})
	}
}

// reportWithStatus builds a one-file report whose single result carries the
// given frozen status, mirroring ratchet.go's own status() mapping: "error"
// is a runner failure, "excluded" is environmental, anything else is a
// bucket verdict.
func reportWithStatus(t *testing.T, path, status string) parseoracle.Report {
	t.Helper()

	r := parseoracle.Result{Path: path}
	switch status {
	case "error":
		r.Err = "perl: no such file or directory"
	case "excluded":
		r.Excluded = true
	default:
		bucket, ok := bucketNamed(status)
		if !ok {
			t.Fatalf("no bucket named %q", status)
		}
		r.Verdict = parseoracle.Verdict{Bucket: bucket}
	}
	return parseoracle.Report{Files: []parseoracle.Result{r}}
}

// bucketNamed resolves a bucket by its rendered name.
func bucketNamed(name string) (parseoracle.Bucket, bool) {
	for _, b := range []parseoracle.Bucket{
		parseoracle.BucketExact,
		parseoracle.BucketWider,
		parseoracle.BucketWrong,
		parseoracle.BucketNoAnswer,
	} {
		if b.String() == name {
			return b, true
		}
	}
	return 0, false
}

// TestRatchetPinMismatch: skew is not a defect. When the interpreter or the
// corpus revision has moved, every verdict in the baseline was measured
// against a different world, so the honest answer is "re-baseline needed" and
// not "203 files regressed".
func TestRatchetPinMismatch(t *testing.T) {
	base := loadFixtureBaseline(t)

	// Move the interpreter half of the pin. Nothing about the verdicts
	// changed; only the world they were measured in.
	skewed := base
	skewed.Pin.Interpreter = "5.036000"

	err := skewed.CheckPin(parseoracle.Pin{
		Interpreter: base.Pin.Interpreter,
		Revision:    base.Pin.Revision,
	})
	if err == nil {
		t.Fatal("a pin mismatch passed: the baseline is being compared across a skew")
	}
	if !strings.Contains(err.Error(), "re-baseline") {
		t.Errorf("a pin mismatch must report re-baseline needed, got:\n%v", err)
	}
	// It must be distinguishable from a regression, or a skewed run reads as
	// a parser that broke overnight.
	var diff *parseoracle.RatchetError
	if asRatchetError(err, &diff) {
		t.Errorf("a pin mismatch must not be a RatchetError: skew is not a regression, got %v", err)
	}

	// The matching pin must pass, or the check is just always-fail.
	if err := base.CheckPin(base.Pin); err != nil {
		t.Errorf("a matching pin must pass, got %v", err)
	}
}

// TestBaselineCategoriesPopulated: every row carries a taxonomy category from
// the fixed set. A raw count says "203 files are no-answer"; a taxonomy says
// "48 of 203 are QuoteLike", and the second is a work plan.
func TestBaselineCategoriesPopulated(t *testing.T) {
	base := loadFixtureBaseline(t)

	if len(base.Rows) == 0 {
		t.Fatal("the committed fixture baseline has no rows")
	}
	for _, row := range base.Rows {
		if row.Category == "" {
			t.Errorf("%s has an empty category: the taxonomy column is the deliverable", row.Path)
			continue
		}
		if !parseoracle.ValidCategory(row.Category) {
			t.Errorf("%s has category %q, which is not in the fixed taxonomy %v",
				row.Path, row.Category, parseoracle.Categories)
		}
	}
}

// TestBaselineCategoryFollowsConstruct proves the taxonomy is keyed on the
// first error's construct rather than defaulted. A column filled entirely
// with General is a column that was never populated.
func TestBaselineCategoryFollowsConstruct(t *testing.T) {
	// Every case is a construct our grammar genuinely fails on, verified
	// individually — a taxonomy test built on constructs the grammar parses
	// cleanly would assert nothing. All seven categories are covered, so a
	// rule that collapsed into General would fail here rather than silently
	// producing a one-value column.
	for _, tc := range []struct {
		name string
		src  string
		want parseoracle.Category
	}{
		{"class", "use v5.38;\nclass Point { field $x;\n", parseoracle.CategoryModernFeature},
		{"class-field", "use v5.38;\nclass P { field $x = ;\n", parseoracle.CategoryModernFeature},
		{"qw", "my @a = qw( a b c\n", parseoracle.CategoryQuoteLike},
		{"tr", "$x =~ tr/abc/def\n", parseoracle.CategoryQuoteLike},
		{"regex", "$x =~ s{a}{b\n", parseoracle.CategoryRegex},
		{"deref", "my $v = ${ $r->{k}\n", parseoracle.CategoryDereference},
		{"subroutine", "sub f( { }\n", parseoracle.CategorySubroutine},
		{"attributes", "sub f :lvalue :method { }\nsub g( {\n", parseoracle.CategorySubroutine},
		{"controlflow", "if ($x {\n", parseoracle.CategoryControlFlow},

		// Three fixtures here were replaced when the grammar fork landed:
		// `try { f()`, `if ($x) { foo()` and `is(<<"${a}{", "A{")` are now
		// parsed cleanly, so the taxonomy correctly declines to categorise
		// them and the cases asserted nothing. Each was swapped for a
		// truncation in the same category that the current grammar still
		// fails on, verified individually. That `if ($x) { foo()` parses
		// without a closing brace is itself a silent-acceptance gap, filed
		// separately -- it is the degenerate detector's target, not this
		// test's.
		{"heredoc-unterminated-tag", "my $x = <<\"EOT;\n", parseoracle.CategoryQuoteLike},

		// Identifier: the lexer's identifier character class. Non-ASCII
		// names under `use utf8` are 28 of the corpus's General files; the
		// `'` package separator is two more. `sub ᕘ { 1 }` alone parses,
		// so the method-call form is pinned as well as the declaration.
		{"utf8-package", "use utf8;\npackage Føø::Bær;\n", parseoracle.CategoryIdentifier},
		{"utf8-array", "use utf8;\n@ᕘ::ISA = 'x';\n", parseoracle.CategoryIdentifier},
		{"utf8-method", "use utf8;\nsub ᕘ { 'x' . (shift)->SUPER::ᕘ }\n", parseoracle.CategoryIdentifier},
		{"apostrophe-var", "$main'a = 1;\n", parseoracle.CategoryIdentifier},
		{"apostrophe-sub", "sub CORE'print'foo { 43 }\n", parseoracle.CategoryIdentifier},

		// Operator: an operator the lexer cannot separate from its operand.
		// `x` juxtaposed to a closing paren or quote reads as an identifier
		// (`x3`); `&&` after a bareword reads as a sigil; `&.` is the string
		// bitwise family. `(1) x 3` with spaces parses cleanly.
		{"repeat-paren", "my @a = ((1)x3, 2);\n", parseoracle.CategoryOperator},
		{"repeat-quote", "my $s = 'x'x8;\n", parseoracle.CategoryOperator},
		{"string-bitwise", "my $x = 22 &. 66;\n", parseoracle.CategoryOperator},
		{"bareword-and", "my $x = foo && 1;\n", parseoracle.CategoryOperator},

		// Regex: a brace-delimited body whose modifiers sit on their own
		// closing line, which the slash-keyed rule could not see.
		{"regex-brace-modifiers", "$x =~ s{\n a\n}{\n b\n}ge;\n", parseoracle.CategoryRegex},
		{"qr-brace-modifiers", "my $re = qr{\n a\n}x;\n", parseoracle.CategoryRegex},

		// Subroutine: a forward declaration after a statement (alone at
		// the top of a file it parses), and a lexical `our sub`. The
		// qualified form pins that `method` inside a sub NAME is a name,
		// not the class-feature keyword.
		{"forward-declaration", "f(1);\nsub bar;\n", parseoracle.CategorySubroutine},
		{"forward-declaration-qualified", "f(1);\nsub Detached::method;\n", parseoracle.CategorySubroutine},
		{"our-sub", "{\n our sub foo { 42 }\n}\n", parseoracle.CategorySubroutine},

		// ControlFlow: the switch feature.
		{"given", "given ($x) { when (1) { } }\n", parseoracle.CategoryControlFlow},
		{"core-given", "CORE::given(1) { }\n", parseoracle.CategoryControlFlow},

		// ModernFeature: a post-5.36 keyword in call position, which is
		// how legacy code that named a sub `try` or `defer` breaks, and
		// the builtin `true`/`false` surface.
		{"keyword-as-sub", "sub try { 1 }\ntry(1, 2);\n", parseoracle.CategoryModernFeature},
		{"builtin-true", "f(sub { true() });\n", parseoracle.CategoryModernFeature},
		// `new Pack ("a")` parses, so it is the reserved word and not the
		// indirect-object syntax that breaks this one.
		{"keyword-indirect-object", "is(method Pack (\"a\"), \"x\");\n", parseoracle.CategoryModernFeature},

		// QuoteLike, by the span rather than the line: a format or heredoc
		// body is not code, so when the error span begins inside one (or
		// on its header) the construct is the body, whatever the site line
		// says. The picture-line site, the empty-format site and the
		// heredoc-body site each fail with the site on a line no rule
		// claims.
		{"format-body-site", "print 1;\nformat STDOUT =\n@ @<<\n\"#\", $a\n.\nprint 2;\n", parseoracle.CategoryQuoteLike},
		{"format-empty", "format STDERR =\n.\nmy $ref;\n", parseoracle.CategoryQuoteLike},
		{"heredoc-body-site", "my $p = <<\"        --\";\n          /f\n           \\$\n          /x\n        --\n", parseoracle.CategoryQuoteLike},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := parseoracle.CategoriseSource([]byte(tc.src))
			if got != tc.want {
				t.Errorf("CategoriseSource(%q) = %s, want %s", tc.src, got, tc.want)
			}
		})
	}
}

// TestCategoriseSourceSiteWithoutErrorNode: a tree can carry HasError with no
// ERROR node anywhere in it, and an empty category must never be read as "this
// file parsed". Measured on the corpus, eight no-answer files were baselined
// with no category for exactly that reason. Three shapes produce it, each
// pinned here by a minimal source verified to parse that way:
//
//   - a MISSING token, which recovery inserts instead of an ERROR node;
//   - a truncated root, where recovery halted and the source_file node ends
//     before the source does;
//   - a degenerate tree, where a hidden rule leaked and no error was
//     recorded at all.
//
// The first two are the General cases they honestly are; the third lands on
// a line the Regex rule claims, which proves the leaked node's line is the
// site rather than a default.
func TestCategoriseSourceSiteWithoutErrorNode(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
		want parseoracle.Category
	}{
		{"missing-token", "sub f { foo: }\n", parseoracle.CategoryGeneral},
		{"truncated-root", "$x = $#[0];\n", parseoracle.CategoryGeneral},
		{"degenerate", "$x =~ s!a!b!x;\n", parseoracle.CategoryRegex},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := parseoracle.CategoriseSource([]byte(tc.src))
			if got != tc.want {
				t.Errorf("CategoriseSource(%q) = %s, want %s", tc.src, got, tc.want)
			}
		})
	}
}

// TestBaselineHeader: the header records the interpreter version and the
// corpus revision, so a reader can tell whether the numbers below were
// measured in the world they are standing in. It also carries a DO NOT EDIT
// marker, because a hand-edited baseline is a baseline that lies.
func TestBaselineHeader(t *testing.T) {
	data, err := os.ReadFile(fixtureBaselinePath())
	if err != nil {
		t.Fatalf("reading the committed baseline: %v", err)
	}
	text := string(data)

	if !strings.Contains(text, "DO NOT EDIT") {
		t.Error("the baseline must carry a DO NOT EDIT BY HAND marker")
	}

	base := loadFixtureBaseline(t)
	if base.Pin.Interpreter == "" {
		t.Error("the baseline header must record the interpreter version")
	}
	if base.Pin.Revision == "" {
		t.Error("the baseline header must record the corpus revision")
	}
	// Round-tripping proves the header is data rather than a comment that
	// happens to be readable: what Render writes, LoadBaseline reads back.
	reloaded, err := parseoracle.ParseBaseline([]byte(base.Render()))
	if err != nil {
		t.Fatalf("re-parsing a rendered baseline: %v", err)
	}
	if reloaded.Pin != base.Pin {
		t.Errorf("pin did not round-trip: %+v vs %+v", reloaded.Pin, base.Pin)
	}
	if len(reloaded.Rows) != len(base.Rows) {
		t.Errorf("rows did not round-trip: %d vs %d", len(reloaded.Rows), len(base.Rows))
	}
}

// TestBaselineUpdateOnly: -update is the only path that rewrites the
// baseline. Without the flag the ratchet reads and compares; a check that
// silently rewrote what it was checking against could never fail.
func TestBaselineUpdateOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.txt")

	base := loadFixtureBaseline(t)
	original := base.Render()
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatalf("seeding the baseline copy: %v", err)
	}

	report := fixtureReport(t)
	changed := downgradeOneFile(t, report)

	// A check must never write, even when it fails.
	loaded, err := parseoracle.LoadBaseline(path)
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}
	if err := loaded.Check(changed); err == nil {
		t.Fatal("premise: the downgraded report must fail the check")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("re-reading the baseline: %v", err)
	}
	if string(after) != original {
		t.Error("a failing check rewrote the baseline: the ratchet is grading its own homework")
	}

	// -update is the path that rewrites it.
	if err := parseoracle.WriteBaseline(path, parseoracle.NewBaseline(loaded.Pin, changed, "")); err != nil {
		t.Fatalf("WriteBaseline: %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("re-reading after update: %v", err)
	}
	if string(updated) == original {
		t.Error("WriteBaseline did not rewrite the baseline")
	}
	// And after an update, the same report passes.
	reloaded, err := parseoracle.LoadBaseline(path)
	if err != nil {
		t.Fatalf("LoadBaseline after update: %v", err)
	}
	if err := reloaded.Check(changed); err != nil {
		t.Errorf("an updated baseline must accept the report it was written from: %v", err)
	}
}
