// ABOUTME: Runs offline/criteria-lint.sh on its bad and clean fixtures and on inline chains, pinning each defect line it prints.
// ABOUTME: Also shows it refusing input it could not read; the lint needs sh and jq, and the tests skip on Windows.

package chart

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var criteriaLintDir = filepath.Join("testdata", "criteria-lint")

// runCriteriaLint runs the lint with input on stdin and returns its stdout,
// its stderr and its exit status.
func runCriteriaLint(t *testing.T, input []byte) (stdout, stderr string, code int) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("criteria-lint.sh is a POSIX sh script")
	}
	cmd := exec.Command("sh", filepath.Join("offline", "criteria-lint.sh"))
	cmd.Stdin = bytes.NewReader(input)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		code = ee.ExitCode()
	default:
		t.Fatal(err)
	}
	return out.String(), errb.String(), code
}

func readLintFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(criteriaLintDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestCriteriaLint: bad.json holds one issue per defect (the third a pin on
// speed.go, which an issue it blocks transitively edits), and the lint names
// each on its own line and exits 1; clean.json passes with exit 0 and no
// output. A missing script fails both halves (sh exits 127).
func TestCriteriaLint(t *testing.T) {
	t.Run("bad", func(t *testing.T) {
		out, stderr, code := runCriteriaLint(t, readLintFixture(t, "bad.json"))
		want := strings.Join([]string{
			`01aa0001-0000 "Criterion in prose": no command: the other thing holds, said in prose only`,
			`01aa0002-0000 "Criterion green before its work": passes on the current tree: it already holds (` + "`true`" + `)`,
			`01aa0003-0000 "Pin on speed.go": pins internal/chart/cmd/ratchet/speed.go, which 01aa0005-0000 edits: guard: speed.go is untouched (` + "`git diff --quiet c6de5c76 -- internal/chart/cmd/ratchet/speed.go`" + `)`,
		}, "\n") + "\n"
		if code != 1 || out != want {
			t.Errorf("exit %d, want 1\nstdout:\n%s\nwant:\n%s\nstderr: %s", code, out, want, stderr)
		}
	})
	t.Run("clean", func(t *testing.T) {
		out, stderr, code := runCriteriaLint(t, readLintFixture(t, "clean.json"))
		if code != 0 || out != "" {
			t.Errorf("exit %d, want 0 and no output\nstdout:\n%s\nstderr: %s", code, out, stderr)
		}
	})
}

// TestCriteriaLintRejectsMalformedInput: the lint cannot pass on input it did
// not read. Malformed JSON and a non-array exit non-zero naming the script on
// stderr; an empty array exits 0 and prints nothing.
func TestCriteriaLintRejectsMalformedInput(t *testing.T) {
	for _, in := range []string{"{not json", `{"id": "01aa0001", "body": ""}`, ""} {
		out, stderr, code := runCriteriaLint(t, []byte(in))
		if code == 0 || out != "" || !strings.HasPrefix(stderr, "criteria-lint.sh: ") {
			t.Errorf("input %q: exit %d, stdout %q, stderr %q; want non-zero, no stdout, a criteria-lint.sh: message", in, code, out, stderr)
		}
	}
	out, stderr, code := runCriteriaLint(t, []byte("[]\n"))
	if code != 0 || out != "" || stderr != "" {
		t.Errorf("empty array: exit %d, stdout %q, stderr %q; want 0 and nothing", code, out, stderr)
	}
}

// TestCriteriaLintSectionsAndGuards pins the line rules one case at a time.
func TestCriteriaLintSectionsAndGuards(t *testing.T) {
	const ac = `## Acceptance Criteria\n\n### Positive Scenarios\n- [ ] it is red (` + "`false`" + `)\n`
	for _, c := range []struct{ name, input, want string }{
		{"passing guard not flagged",
			`[{"id":"01cc0001-0000-7","title":"G","state":"pending","body":"` + ac + `- [ ] guard: it still builds (` + "`true`" + `)\n"}]`,
			""},
		{"negative scenario linted",
			`[{"id":"01cc0002-0000-7","title":"N","state":"pending","body":"` + ac + `\n### Negative Scenarios\n- [ ] the bad input is refused (` + "`true`" + `)\n"}]`,
			`01cc0002-0000 "N": passes on the current tree: the bad input is refused (` + "`true`" + `)` + "\n"},
		{"bare backtick span is no command",
			`[{"id":"01cc0003-0000-7","title":"B","state":"pending","body":"` + ac + `- [ ] it holds ` + "`false`" + `\n"}]`,
			`01cc0003-0000 "B": no command: it holds ` + "`false`" + "\n"},
		{"pin on a file only an earlier issue edits not flagged",
			`[{"id":"01cc0004-0000-7","title":"E","state":"pending","blocks":["01cc0005-0000-7"],"body":"- paths: ` + "`internal/chart/README.md`" + `\n` + ac + `"},` +
				`{"id":"01cc0005-0000-7","title":"P","state":"pending","body":"` + ac + `- [ ] guard: the README is pinned (` + "`cmp internal/chart/README.md /dev/null`" + `)\n"}]`,
			""},
	} {
		t.Run(c.name, func(t *testing.T) {
			out, stderr, code := runCriteriaLint(t, []byte(c.input))
			wantCode := 0
			if c.want != "" {
				wantCode = 1
			}
			if code != wantCode || out != c.want {
				t.Errorf("exit %d, want %d\nstdout:\n%s\nwant:\n%s\nstderr: %s", code, wantCode, out, c.want, stderr)
			}
		})
	}
}
