// ABOUTME: Proves the sweep runs against an injected subject, not just our own parser.
// ABOUTME: The subject here is a shell script, so passing means no Go type crossed the boundary.

package parseoracle

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
