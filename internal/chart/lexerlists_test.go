// ABOUTME: Holds the two lexer lists, lex-unmapped-known.txt and lexer-defects.txt, to their entry form.
// ABOUTME: Each entry is `<path> <byte> <issue ref>`, the path a corpus file (T1 via T1_CORPUS), the byte inside it.

package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// repoRoot is the repository root seen from this package's directory.
var repoRoot = filepath.Join("..", "..")

// inTreeCorpus is the in-tree corpus files as repository-relative globs.
var inTreeCorpus = []string{
	"internal/chart/testdata/std5k.pl",
	"internal/psc/testdata/latency/edit5k.pl",
	"internal/chart/testdata/guacamole/parses/*.pl",
	"internal/chart/testdata/guacamole/parsent/*.pl",
}

// issueRef is a git-zhi id through its second segment, the shortest form
// that resolves reliably (crochet's how-to-use-git-zhi).
var issueRef = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}$`)

// lexerListErrors returns one message per malformed entry of the lexer list
// called name: an entry is `<path> <byte> <issue ref>`, the path a corpus
// file under root (or `T1/<name>.t` under t1, the T1_CORPUS value), the byte
// inside it.
func lexerListErrors(name, list, root, t1 string) []string {
	var out []string
	for i, line := range strings.Split(list, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		bad := func(why string) { out = append(out, fmt.Sprintf("%s:%d: %s: %q", name, i+1, why, line)) }
		if len(f) != 3 {
			bad("want <path> <byte> <issue ref>")
			continue
		}
		if !issueRef.MatchString(f[2]) {
			bad("the issue ref is not an id through its second segment, like 01a1882d-d8a4")
		}
		var file string
		if t1name, ok := strings.CutPrefix(f[0], "T1/"); ok {
			if t1 == "" {
				bad("a T1 entry needs T1_CORPUS set")
				continue
			}
			if fi, err := os.Stat(t1); err != nil || !fi.IsDir() {
				bad("T1_CORPUS=" + t1 + " names no directory")
				continue
			}
			if strings.Contains(t1name, "/") || !strings.HasSuffix(t1name, ".t") {
				bad("a T1 path is T1/<name>.t, a top-level .t file")
				continue
			}
			file = filepath.Join(t1, t1name)
		} else {
			if !slices.ContainsFunc(inTreeCorpus, func(g string) bool { m, _ := filepath.Match(g, f[0]); return m }) {
				bad("the path is not a corpus file")
				continue
			}
			file = filepath.Join(root, f[0])
		}
		fi, err := os.Stat(file)
		if err != nil {
			bad(err.Error())
			continue
		}
		if b, err := strconv.ParseUint(f[1], 10, 63); err != nil || int64(b) >= fi.Size() {
			bad(fmt.Sprintf("the byte is not a decimal offset below the file's size, %d", fi.Size()))
		}
	}
	return out
}

func TestLexerLists(t *testing.T) {
	t1 := os.Getenv("T1_CORPUS")
	for _, name := range []string{"lex-unmapped-known.txt", "lexer-defects.txt"} {
		list, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range lexerListErrors(name, string(list), repoRoot, t1) {
			t.Error(e)
		}
	}
}

func TestLexerListsRefusesMalformedEntries(t *testing.T) {
	const std5k = "internal/chart/testdata/std5k.pl"
	info, err := os.Stat(filepath.Join(repoRoot, std5k))
	if err != nil {
		t.Fatal(err)
	}
	size := info.Size()
	itoa := func(n int64) string { return strconv.FormatInt(n, 10) }
	for _, tc := range []struct {
		name, line string
		ok         bool
	}{
		{"well-formed", std5k + " 0 01a1882d-d8a4", true},
		{"last byte", std5k + " " + itoa(size-1) + " 01a1882d-d8a4", true},
		{"guacamole snippet", "internal/chart/testdata/guacamole/parses/Statements-Block-001.pl 0 01a1882d-d8a4", true},
		{"byte at the end", std5k + " " + itoa(size) + " 01a1882d-d8a4", false},
		{"byte past the end", std5k + " " + itoa(size+10) + " 01a1882d-d8a4", false},
		{"negative byte", std5k + " -1 01a1882d-d8a4", false},
		{"byte not a number", std5k + " x 01a1882d-d8a4", false},
		{"not a corpus file", "internal/chart/README.md 0 01a1882d-d8a4", false},
		{"no such file", "internal/chart/testdata/guacamole/parses/no-such.pl 0 01a1882d-d8a4", false},
		{"missing ref", std5k + " 0", false},
		{"missing byte", std5k + " 01a1882d-d8a4", false},
		{"extra field", std5k + " 0 01a1882d-d8a4 x", false},
		{"ref as the second segment", std5k + " 0 d8a4", false},
		{"ref as the short id", std5k + " 0 01a1882d", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs := lexerListErrors("synthetic", tc.line+"\n", repoRoot, "")
			if tc.ok && len(errs) > 0 {
				t.Errorf("refused a well-formed entry: %v", errs)
			}
			if !tc.ok && len(errs) == 0 {
				t.Errorf("accepted %q", tc.line)
			}
		})
	}
}

func TestLexerListsT1Unresolvable(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.t"), []byte("print 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	const list = "T1/a.t 0 01a1882d-d8a4\n"
	for _, t1 := range []string{"", filepath.Join(dir, "missing")} {
		errs := lexerListErrors("synthetic", list, repoRoot, t1)
		if len(errs) == 0 || !strings.Contains(strings.Join(errs, "\n"), "T1_CORPUS") {
			t.Errorf("T1_CORPUS=%q: want a failure naming T1_CORPUS, got %v", t1, errs)
		}
	}
	if errs := lexerListErrors("synthetic", list, repoRoot, dir); len(errs) > 0 {
		t.Errorf("T1_CORPUS naming a directory holding the file: %v", errs)
	}
}
