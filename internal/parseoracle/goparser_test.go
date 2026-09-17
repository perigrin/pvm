// ABOUTME: The hand-written parser measured through the subject contract, as a subprocess.
// ABOUTME: WRONG is a gate; the exact floor is what keeps WRONG=0 from being satisfiable by silence.

package parseoracle_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parseoracle"
)

// t2Dirs is spec §7.3's tier-2 core, the same five directories
// internal/lexer and internal/parse use.
var t2Dirs = []string{"base", "cmd", "comp", "opbasic", "class"}

// goSubject builds the subject binary once and returns a Subject that runs
// it.
//
// Built rather than called: the contract is a subprocess precisely so the
// harness can measure implementations that are not Go (§ SubjectFacts).
// Measuring our own parser through a Go call would test a different path
// from the one every other subject uses, and the JSON round-trip is where
// the "nil means did not look" distinction actually bites.
func goSubject(t *testing.T) *parseoracle.Subject {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "subject")
	build := exec.Command("go", "build", "-o", bin,
		"tamarou.com/pvm/internal/parse/cmd/subject")
	build.Dir = repoRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the subject: %v\n%s", err, out)
	}
	return &parseoracle.Subject{Command: []string{bin}}
}

// t2Corpus returns the T2 core files, relative to the corpus t/ directory.
func t2Corpus(t *testing.T) (root string, files []string) {
	t.Helper()

	corpus, err := parseoracle.CorpusRoot()
	if err != nil {
		t.Skipf("%v", err)
	}
	tDir := filepath.Join(corpus, "t")
	for _, d := range t2Dirs {
		matches, err := filepath.Glob(filepath.Join(tDir, d, "*.t"))
		if err != nil {
			t.Skipf("globbing %s: %v", d, err)
		}
		for _, m := range matches {
			rel, err := filepath.Rel(tDir, m)
			if err != nil {
				continue
			}
			files = append(files, rel)
		}
	}
	sort.Strings(files)
	return tDir, files
}

// runGoSubject measures the Go parser over T2 and returns the report.
func runGoSubject(t *testing.T) parseoracle.Report {
	t.Helper()

	tDir, files := t2Corpus(t)
	shim := os.Getenv("PARSEORACLE_SHIM")
	if shim == "" {
		t.Skip("set PARSEORACLE_SHIM to a built shim to measure the subject;\n" +
			"t/test.pl clears @INC, so without one most files report false failures")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	report, err := parseoracle.Run(ctx, files, parseoracle.RunOptions{
		Subject: goSubject(t),
		// Files are relative to the shim's t/, which is where perl's own
		// tests refer to each other from.
		Dir:     filepath.Join(shim, "t"),
		Timeout: 60 * time.Second,
	})
	if err != nil {
		t.Fatalf("measuring the subject over %d files from %s: %v",
			len(files), tDir, err)
	}
	return report
}

// TestGoParserSubjectMeasures is the M1 gate's subject-contract criterion:
// this parser answers the harness as a subprocess, like any other
// implementation.
//
// It asserts that the measurement HAPPENED, not that it came out well. The
// two criteria below are what judge the answer.
func TestGoParserSubjectMeasures(t *testing.T) {
	report := runGoSubject(t)

	if report.Measured() == 0 {
		t.Fatalf("the subject produced no verdicts over %d files; "+
			"%d environmental, %d runner errors",
			len(report.Files), report.Environmental, report.Errors)
	}
	t.Logf("T2 through the subject contract: %d measured, "+
		"%d exact, %d wider, %d WRONG, %d no-answer (%d environmental, %d errors)",
		report.Measured(),
		report.Totals[parseoracle.BucketExact],
		report.Totals[parseoracle.BucketWider],
		report.Totals[parseoracle.BucketWrong],
		report.Totals[parseoracle.BucketNoAnswer],
		report.Environmental, report.Errors)
}

// TestGoParserWrongIsZero is the gate. Any nonzero value fails.
//
// WRONG means this parser committed to a parse perl did not make. That is
// the one bucket that is never acceptable: a declined construct is honest
// and scores wider, but a confident wrong answer lies to every downstream
// consumer. Chapter 7 §7.2 and the Unknown doc at parse.go:14-26 are the
// same rule stated twice.
func TestGoParserWrongIsZero(t *testing.T) {
	report := runGoSubject(t)

	wrong := report.Totals[parseoracle.BucketWrong]
	if wrong == 0 {
		return
	}

	var names []string
	for _, f := range report.Files {
		if f.Verdict.Bucket == parseoracle.BucketWrong {
			names = append(names, f.Path)
		}
	}
	sort.Strings(names)
	t.Errorf("%d file(s) score WRONG, and WRONG is a gate:\n  %s",
		wrong, strings.Join(names, "\n  "))
}

// TestGoParserExactFloor is what stops WRONG=0 from being vacuous.
//
// A subject that answered nothing at all would score WRONG=0 and pass the
// gate above. The M1 issue says so directly: "WRONG = 0 is trivially
// satisfiable by a parser that answers nothing — every M0 lesson about
// guards that cannot fail applies. The exact floor is what makes WRONG = 0
// mean something."
//
// The issue's target is 40% of markers attempted. This asserts a floor and
// logs the distance, in the ratchet discipline the other M1 metrics use:
// the number moves deliberately, in the commit that moved it.
func TestGoParserExactFloor(t *testing.T) {
	report := runGoSubject(t)

	measured := report.Measured()
	if measured == 0 {
		t.Fatal("no files measured; the floor is meaningless")
	}
	exact := report.Totals[parseoracle.BucketExact]
	rate := 100 * float64(exact) / float64(measured)

	const target = 40.0
	if rate < target {
		t.Errorf("exact is %.1f%% of %d measured files (%d exact), "+
			"below the gate's %.1f%% floor.\n"+
			"WRONG=0 with a low exact rate is a parser that declines rather "+
			"than one that agrees.", rate, measured, exact, target)
	}
	t.Logf("exact %d of %d measured (%.1f%%), floor %.1f%%",
		exact, measured, rate, target)
}
