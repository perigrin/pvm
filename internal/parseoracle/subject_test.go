// ABOUTME: Tests the subprocess contract that lets any implementation be measured.
// ABOUTME: A fake subject is a shell script, which is the point: no Go types cross the boundary.

package parseoracle

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeFakeSubject creates an executable that prints fixed JSON, standing in
// for a real parser. It is deliberately a shell script rather than a Go test
// helper: if a shell script can be measured, so can PerlOnJava, perl-lsp or
// Chalk, none of which can satisfy a Go interface.
func writeFakeSubject(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake subject is a POSIX shell script")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-subject")
	script := "#!/bin/sh\ncat <<'SUBJECT_EOF'\n" + body + "\nSUBJECT_EOF\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake subject: %v", err)
	}
	return path
}

// TestSubjectFromShellScript is the acceptance test for the whole issue: an
// implementation that is not Go, does not import our packages, and has no
// syntax tree we can see, is measurable all the same.
func TestSubjectFromShellScript(t *testing.T) {
	subject := writeFakeSubject(t, `{
  "ok": true,
  "prototypes": {"f": "\\@"},
  "call_sites": [{"line": 3, "name": "f", "took_reference": true}],
  "declined": false
}`)

	s := Subject{Command: []string{subject}}
	facts, err := s.Parse(context.Background(), "irrelevant.pl")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if !facts.OK {
		t.Error("OK: got false, want true")
	}
	if got := facts.Prototypes["f"]; got != `\@` {
		t.Errorf("Prototypes[f]: got %q, want %q", got, `\@`)
	}
	if len(facts.CallSites) != 1 {
		t.Fatalf("CallSites: got %d, want 1", len(facts.CallSites))
	}
	if !facts.CallSites[0].TookReference {
		t.Error("CallSites[0].TookReference: got false, want true")
	}
	if facts.Declined {
		t.Error("Declined: got true, want false")
	}
}

// TestPartialSubjectScoresNoAnswer pins the rule that a subject which cannot
// answer a question must not be guessed at. Omitting call_sites is not the
// same as reporting no call sites, and scoring the difference as agreement
// would let a subject win by staying silent.
func TestPartialSubjectScoresNoAnswer(t *testing.T) {
	subject := writeFakeSubject(t, `{"ok": true}`)

	s := Subject{Command: []string{subject}}
	facts, err := s.Parse(context.Background(), "irrelevant.pl")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if facts.KnowsCallSites {
		t.Error("KnowsCallSites: got true for a subject that omitted the field")
	}

	oracle := Facts{OK: true, Srefgen: 1}
	v := CompareFacts(oracle, facts)
	if v.Bucket != BucketNoAnswer {
		t.Errorf("Bucket: got %v, want %v -- an unanswered question is not agreement",
			v.Bucket, BucketNoAnswer)
	}
}

// TestSubjectDeclinedScoresNoAnswer checks the portable replacement for
// IsDegenerate. "Did you silently drop source you could not handle" is a
// question every parser can answer about itself; only tree-sitter has hidden
// rules to leak.
func TestSubjectDeclinedScoresNoAnswer(t *testing.T) {
	subject := writeFakeSubject(t, `{
  "ok": true,
  "call_sites": [],
  "declined": true,
  "declined_reason": "dropped source at line 4"
}`)

	s := Subject{Command: []string{subject}}
	facts, err := s.Parse(context.Background(), "irrelevant.pl")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	v := CompareFacts(Facts{OK: true, Srefgen: 0}, facts)
	if v.Bucket != BucketNoAnswer {
		t.Errorf("Bucket: got %v, want %v", v.Bucket, BucketNoAnswer)
	}
	if v.Detail == "" {
		t.Error("Detail: empty; a declined parse should say why")
	}
}
