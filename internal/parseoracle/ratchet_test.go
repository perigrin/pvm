// ABOUTME: Tests the ratchet: that it fails on a regression, fails on an unrecorded improvement, and rewrites only under -update.
// ABOUTME: The point is the failures — perl-lsp shipped a ratchet that could not fail, which is decoration.

package parseoracle_test

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
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

// runFixture measures the five-file fixture corpus that TestCorpusRun uses.
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
// against. It is a five-file fixture rather than the 620-file corpus because
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
			parseoracle.NewBaseline(pin, report)); err != nil {
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

// fixtureReport measures the five-file fixture corpus.
func fixtureReport(t *testing.T) parseoracle.Report {
	t.Helper()
	return runFixture(t)
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
		{"try", "use feature 'try';\ntry { f()\n", parseoracle.CategoryModernFeature},
		{"qw", "my @a = qw( a b c\n", parseoracle.CategoryQuoteLike},
		{"tr", "$x =~ tr/abc/def\n", parseoracle.CategoryQuoteLike},
		{"regex", "$x =~ s{a}{b\n", parseoracle.CategoryRegex},
		{"deref", "my $v = ${ $r->{k}\n", parseoracle.CategoryDereference},
		{"subroutine", "sub f( { }\n", parseoracle.CategorySubroutine},
		{"attributes", "sub f :lvalue :method { }\nsub g( {\n", parseoracle.CategorySubroutine},
		{"controlflow", "if ($x) { foo()\n", parseoracle.CategoryControlFlow},
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
	if err := parseoracle.WriteBaseline(path, parseoracle.NewBaseline(loaded.Pin, changed)); err != nil {
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
