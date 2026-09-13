// ABOUTME: Tests the subprocess contract that lets any implementation be measured.
// ABOUTME: A fake subject is a shell script, which is the point: no Go types cross the boundary.

package parseoracle

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// writeFakeSubject creates an executable that prints fixed JSON, standing in
// for a real parser. It is deliberately a shell script rather than a Go test
// helper: if a shell script can be measured, so can PerlOnJava, perl-lsp or
// Chalk, none of which can satisfy a Go interface.
func writeFakeSubject(t *testing.T, body string) string {
	t.Helper()
	return writeSubjectScript(t, "#!/bin/sh\ncat <<'SUBJECT_EOF'\n"+body+"\nSUBJECT_EOF\n")
}

// writeSubjectScript installs an arbitrary shell script as a subject, for the
// tests that need a subject to misbehave rather than to answer.
func writeSubjectScript(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake subject is a POSIX shell script")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-subject")
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

	oracle := withSrefgen(1)
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

	v := CompareFacts(withSrefgen(0), facts)
	if v.Bucket != BucketNoAnswer {
		t.Errorf("Bucket: got %v, want %v", v.Bucket, BucketNoAnswer)
	}
	if v.Detail == "" {
		t.Error("Detail: empty; a declined parse should say why")
	}
}

// TestNullCallSitesIsNotAnAnswer pins that a JSON null is an absent answer,
// not an empty one. A Go subject without omitempty, a JSON::PP undef and a
// Jackson null field all emit `"call_sites": null` to mean "not computed",
// and scoring that as "no calls here" would hand every such subject a verdict
// it never gave.
func TestNullCallSitesIsNotAnAnswer(t *testing.T) {
	subject := writeFakeSubject(t, `{"ok": true, "call_sites": null, "prototypes": null}`)

	facts, err := Subject{Command: []string{subject}}.Parse(context.Background(), "irrelevant.pl")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if facts.KnowsCallSites {
		t.Error("KnowsCallSites: got true for a null call_sites")
	}
	if facts.KnowsPrototypes {
		t.Error("KnowsPrototypes: got true for a null prototypes")
	}
	if v := CompareFacts(withSrefgen(1), facts); v.Bucket != BucketNoAnswer {
		t.Errorf("Bucket: got %v, want %v -- null is silence, and silence is not agreement",
			v.Bucket, BucketNoAnswer)
	}
}

// TestSubjectFactsRoundTrip pins that the exported type can say "no call
// sites" through its own JSON. A Go subject that reuses SubjectFacts must be
// able to report an empty list as an answer and a nil one as no answer; if
// marshalling collapses the two, the type contradicts the contract it defines.
func TestSubjectFactsRoundTrip(t *testing.T) {
	cases := []struct {
		name        string
		in          SubjectFacts
		knowsSites  bool
		knowsProtos bool
	}{
		{"empty answers are answers", SubjectFacts{OK: true,
			CallSites: []SubjectCallSite{}, Prototypes: map[string]string{}}, true, true},
		{"nil answers are silence", SubjectFacts{OK: true}, false, false},
	}
	for _, tc := range cases {
		out, err := json.Marshal(tc.in)
		if err != nil {
			t.Fatalf("%s: Marshal: %v", tc.name, err)
		}
		got, err := decodeSubjectFacts(out)
		if err != nil {
			t.Fatalf("%s: decode %s: %v", tc.name, out, err)
		}
		if got.KnowsCallSites != tc.knowsSites {
			t.Errorf("%s: %s decoded KnowsCallSites=%v, want %v", tc.name, out, got.KnowsCallSites, tc.knowsSites)
		}
		if got.KnowsPrototypes != tc.knowsProtos {
			t.Errorf("%s: %s decoded KnowsPrototypes=%v, want %v", tc.name, out, got.KnowsPrototypes, tc.knowsProtos)
		}
	}
}

// TestSubjectMustAnswerOK pins that a document without an opinion is not a
// verdict. `null` and `{}` are what a subject prints when it crashed before
// deciding anything; recording either as "the subject rejected the file" puts
// a parser opinion in the report that no parser held.
func TestSubjectMustAnswerOK(t *testing.T) {
	for _, doc := range []string{`null`, `{}`, `{"call_sites": []}`} {
		subject := writeFakeSubject(t, doc)
		_, err := Subject{Command: []string{subject}}.Parse(context.Background(), "irrelevant.pl")
		if err == nil {
			t.Errorf("%s: Parse accepted a document with no ok field; a subject that "+
				"never said whether the file is Perl has not answered", doc)
		}
	}
}

// TestSubjectTimeoutIsNamed pins that a timeout reads as a timeout. The oracle
// already maps its context error; a subject reporting "signal: killed" for the
// same event reads as a crash and sends triage in the wrong direction.
func TestSubjectTimeoutIsNamed(t *testing.T) {
	subject := writeSubjectScript(t, "#!/bin/sh\nexec sleep 30\n")

	_, err := Subject{Command: []string{subject}, Timeout: 100 * time.Millisecond}.
		Parse(context.Background(), "irrelevant.pl")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Parse: got %v, want a deadline error", err)
	}
}

// TestSubjectStderrIsReported pins that a subject's own diagnosis survives.
// "exit status 2" tells a reader nothing; "cannot read op/sub.t" tells them the
// working directory is wrong, which is what it was.
func TestSubjectStderrIsReported(t *testing.T) {
	subject := writeSubjectScript(t, "#!/bin/sh\necho \"cannot read $1\" >&2\nexit 2\n")

	_, err := Subject{Command: []string{subject}}.Parse(context.Background(), "op/sub.t")
	if err == nil || !strings.Contains(err.Error(), "cannot read op/sub.t") {
		t.Errorf("Parse: got %v; the subject said why it failed and the error should repeat it", err)
	}
}

// TestSubjectTimeoutKillsGrandchildren pins that a timed-out subject takes its
// children with it. A `#!/bin/sh` subject that does not exec its last command
// is a parent, and killing only the parent leaves the real work running past
// the sweep -- 620 files of that is a machine full of orphans.
func TestSubjectTimeoutKillsGrandchildren(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "pid")
	subject := writeSubjectScript(t,
		"#!/bin/sh\nsleep 300 &\necho $! > "+pidfile+"\nwait\n")

	_, err := Subject{Command: []string{subject}, Timeout: 100 * time.Millisecond}.
		Parse(context.Background(), "irrelevant.pl")
	if err == nil {
		t.Fatal("Parse: expected a timeout")
	}

	raw, err := os.ReadFile(pidfile)
	if err != nil {
		t.Fatalf("the subject never recorded its child's pid: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("pidfile %q: %v", raw, err)
	}
	child, _ := os.FindProcess(pid)
	deadline := time.Now().Add(3 * time.Second)
	for child.Signal(syscall.Signal(0)) == nil {
		if time.Now().After(deadline) {
			_ = child.Kill()
			t.Fatalf("sleep (pid %d) survived the subject's timeout: the subject's grandchildren leaked", pid)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestSubjectOutputIsBounded pins that a subject cannot consume the harness.
// Output is read into memory once per worker; a subject that streams forever
// must be cut off and reported, not buffered until the machine gives up.
func TestSubjectOutputIsBounded(t *testing.T) {
	subject := writeSubjectScript(t,
		"#!/bin/sh\nexec perl -e 'print q({\"ok\":true}) x 2_000_000'\n")

	_, err := Subject{Command: []string{subject}}.Parse(context.Background(), "irrelevant.pl")
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("Parse: got %v; want the output cap named, not a JSON error after buffering it all", err)
	}
}
