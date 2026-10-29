// ABOUTME: Guards the checked-in Guacamole corpus: no duplicate case, every case starts with the preamble.
// ABOUTME: Also checks the perl-rejects rule on hand-built cases; the corpus-wide sweep is the ratchet's.

package chart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/perltest"
)

var guacamoleDir = filepath.Join("testdata", "guacamole")

// TestRejectsRuleHandBuilt: of a case perl accepts, a case perl rejects that
// is listed, and one perl rejects that is not, the rule names only the third.
// The accepted case uses a stubbed module, so it compiles only with the stubs.
func TestRejectsRuleHandBuilt(t *testing.T) {
	p := perltest.Perl542(t)
	preamble := readFile(t, filepath.Join(guacamoleDir, "preamble.pl"))
	dir := t.TempDir()
	cases := map[string]string{
		"accepted.pl": "use My::Module 4.10;\n1;\n",
		"listed.pl":   "sub {\n",
		"unlisted.pl": "$#\n",
	}
	for name, body := range cases {
		writeFile(t, filepath.Join(dir, name), preamble+body)
	}
	rejects := filepath.Join(t.TempDir(), "perl-rejects.txt")
	writeFile(t, rejects, "listed.pl\n")

	got, err := RejectsDisagree(p.Path, dir, rejects, filepath.Join(guacamoleDir, "stubs"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !strings.HasPrefix(got[0], "unlisted.pl:") {
		t.Fatalf("RejectsDisagree = %q, want one disagreement naming unlisted.pl", got)
	}
}

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
