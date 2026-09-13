// ABOUTME: Proves the sweep runs against an injected subject, not just our own parser.
// ABOUTME: The subject here is a shell script, so passing means no Go type crossed the boundary.

package parseoracle

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestRunAcceptsInjectedParser is the end-to-end form of the issue's claim.
// TestSubjectFromShellScript shows one file can be measured through the
// contract; this shows a whole sweep can, which is what the ratchet and the
// fidelity number are built on.
func TestRunAcceptsInjectedParser(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake subject is a POSIX shell script")
	}

	dir := t.TempDir()

	// A subject that always declines. Trivial, and that is deliberate: what
	// is under test is that the sweep ASKS it rather than asking our parser.
	subjectPath := filepath.Join(dir, "always-declines")
	script := "#!/bin/sh\n" +
		"cat <<'EOF'\n" +
		`{"ok": true, "call_sites": [], "declined": true, ` +
		`"declined_reason": "injected subject declines everything"}` + "\n" +
		"EOF\n"
	if err := os.WriteFile(subjectPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write subject: %v", err)
	}

	// One trivially valid Perl file, so perl supplies ground truth and the
	// verdict turns on the subject's answer alone.
	src := filepath.Join(dir, "simple.pl")
	if err := os.WriteFile(src, []byte("my $x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	report, err := Run(context.Background(), []string{src}, RunOptions{
		Subject: &Subject{Command: []string{subjectPath}},
		Workers: 1,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(report.Files) != 1 {
		t.Fatalf("Results: got %d, want 1", len(report.Files))
	}
	got := report.Files[0]
	if got.Err != "" {
		t.Fatalf("unexpected error: %s", got.Err)
	}
	if got.Verdict.Bucket != BucketNoAnswer {
		t.Errorf("Bucket: got %v, want %v -- the sweep did not consult the injected subject",
			got.Verdict.Bucket, BucketNoAnswer)
	}
	if got.Verdict.Detail != "injected subject declines everything" {
		t.Errorf("Detail: got %q; the verdict did not come from the subject",
			got.Verdict.Detail)
	}
}

// TestRunResolvesSubjectPathsAgainstDir pins the working-directory half of the
// contract. The runner names corpus files relative to RunOptions.Dir, and the
// oracle is run from there; an external subject must be too, or it is handed a
// path it cannot open and the report records that as five parser verdicts.
//
// The subject is honest: it refuses a file it cannot read. The quiet
// alternative -- report ok:false, the `perl -c` convention -- is worse, because
// it prints "exact rate 0.0%" as though it were a fidelity measurement.
func TestRunResolvesSubjectPathsAgainstDir(t *testing.T) {
	subject := writeSubjectScript(t, "#!/bin/sh\n"+
		"test -r \"$1\" || { echo \"cannot read $1\" >&2; exit 2; }\n"+
		"echo '{\"ok\": true, \"call_sites\": []}'\n")
	dir := writeCorpusFile(t, "simple.pl", "my $x = 1;\n")

	report, err := Run(context.Background(), []string{"simple.pl"}, RunOptions{
		Dir:     dir,
		Subject: &Subject{Command: []string{subject}},
		Workers: 1,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := report.Files[0]
	if got.Err != "" {
		t.Fatalf("runner error: %s -- the subject was not run from RunOptions.Dir", got.Err)
	}
	if got.Verdict.Bucket != BucketExact {
		t.Errorf("Bucket: got %v (%s), want %v", got.Verdict.Bucket, got.Verdict.Detail, BucketExact)
	}
}

// TestRunTimeoutReachesSubject pins the other half: RunOptions.Timeout is
// documented as bounding ONE file, and a subject is part of measuring a file.
// Without this a subject runs on its own 60s default whatever the sweep was
// told, and a wedged corpus file costs a minute per worker while the report
// claims a shorter budget.
func TestRunTimeoutReachesSubject(t *testing.T) {
	subject := writeSubjectScript(t, "#!/bin/sh\nexec sleep 5\n")
	dir := writeCorpusFile(t, "simple.pl", "my $x = 1;\n")

	start := time.Now()
	report, err := Run(context.Background(), []string{"simple.pl"}, RunOptions{
		Dir:     dir,
		Subject: &Subject{Command: []string{subject}},
		Timeout: time.Second,
		Workers: 1,
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := report.Files[0]
	if !strings.Contains(got.Err, "subject") || !strings.Contains(got.Err, context.DeadlineExceeded.Error()) {
		t.Errorf("Err: %q; want the subject cut off by the run's timeout", got.Err)
	}
	if elapsed > 4*time.Second {
		t.Errorf("Run took %s against a 1s per-file timeout", elapsed.Round(time.Millisecond))
	}
}

// TestPerlSubjectMeasuresFixture is the claim made good: an implementation
// that is not Go, not in this module and not a fake, measured end to end
// through the contract with the paths and working directory the corpus sweep
// uses. The subject is perl's own parser, which makes this a calibration as
// well: measured against itself, the reference must score exact on every file
// it compiles, including the prototype call our grammar can only hedge on.
func TestPerlSubjectMeasuresFixture(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the perl subject is invoked through a POSIX shell")
	}
	script, err := filepath.Abs(filepath.Join("testdata", "perl_subject.pl"))
	if err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.Abs(filepath.Join("testdata", "fixture"))
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]Bucket{
		"exact.pl":        BucketExact,
		"explicit_ref.pl": BucketExact,
		"wider.pl":        BucketExact,
		"noanswer.pl":     BucketNoAnswer,
	}
	files := []string{"environmental.pl"}
	for name := range want {
		files = append(files, name)
	}

	report, err := Run(context.Background(), files, RunOptions{
		Dir:     dir,
		Subject: &Subject{Command: []string{"perl", script}},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, f := range report.Files {
		switch {
		case f.Err != "":
			t.Errorf("%s: runner error: %s", f.Path, f.Err)
		case f.Path == "environmental.pl":
			if !f.Excluded {
				t.Errorf("%s: not excluded; a missing module is about the machine", f.Path)
			}
		case f.Verdict.Bucket != want[f.Path]:
			t.Errorf("%s: got %v (%s), want %v", f.Path, f.Verdict.Bucket, f.Verdict.Detail, want[f.Path])
		}
	}
	if report.Errors != 0 || report.Measured() != 4 || report.Environmental != 1 {
		t.Errorf("report: measured %d, environmental %d, errors %d; want 4, 1, 0\n%s",
			report.Measured(), report.Environmental, report.Errors, report)
	}
}

// writeCorpusFile creates a one-file corpus and returns its directory, so a
// test can hand the runner a RELATIVE path the way the sweep does.
func writeCorpusFile(t *testing.T, name, src string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return dir
}

// TestNoParserTypesInAPI guards the boundary against regression. The portable
// core must not name a parser type: if it does, a subject that is not written
// in Go cannot be measured, which is the defect this issue exists to fix.
func TestNoParserTypesInAPI(t *testing.T) {
	for _, file := range []string{"compare_facts.go", "subject.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if contains(string(src), "internal/parser") {
			t.Errorf("%s imports internal/parser; the portable core must not "+
				"depend on any one parser (adapter.go is where that belongs)", file)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack == needle ||
			len(needle) == 0 ||
			indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
