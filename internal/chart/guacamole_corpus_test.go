// ABOUTME: Guards the checked-in Guacamole corpus: no duplicate case, every case starts with the preamble.
// ABOUTME: Which cases perl accepts is the committed labels' to say (labels_test.go), not perl's at test time.

package chart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var guacamoleDir = filepath.Join("testdata", "guacamole")

// TestGuacamoleCasesUnique: no two case files in one directory hold the same text.
func TestGuacamoleCasesUnique(t *testing.T) {
	for _, kind := range []string{"parses", "parsent"} {
		first := map[string]string{}
		for _, f := range caseFiles(t, kind) {
			text := readFile(t, f)
			if prev, ok := first[text]; ok {
				t.Errorf("duplicate case: %s holds the same text as %s", f, prev)
				continue
			}
			first[text] = f
		}
	}
}

// TestGuacamoleCasesStartWithPreamble: every case file starts with the declared preamble.
func TestGuacamoleCasesStartWithPreamble(t *testing.T) {
	preamble := readFile(t, filepath.Join(guacamoleDir, "preamble.pl"))
	for _, kind := range []string{"parses", "parsent"} {
		for _, f := range caseFiles(t, kind) {
			if !strings.HasPrefix(readFile(t, f), preamble) {
				t.Errorf("%s does not start with the preamble %q", f, preamble)
			}
		}
	}
}

// caseFiles lists one kind's case files, failing if there are none, so the
// guards above cannot pass on an empty corpus.
func caseFiles(t *testing.T, kind string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(guacamoleDir, kind, "*.pl"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no %s cases under %s (%v)", kind, guacamoleDir, err)
	}
	return files
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
