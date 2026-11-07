// ABOUTME: Holds the lexer lists, lex-unmapped-known.txt, lexer-defects.txt and tdt/lexer-rooted.txt, to their entry form.
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

// issueRef is a git-zhi uuid prefix of whole segments, at least two. The
// first two are the creation time to the millisecond and issues from one run
// share them, so a ref is extended segment by segment until unique; that it
// resolves to one issue is checked by git zhi, never here.
var issueRef = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}(-[0-9a-f]{4}(-[0-9a-f]{4}(-[0-9a-f]{12})?)?)?$`)

// lexerListErrors returns one message per malformed entry of the lexer list
// called name: an entry is `<path> <byte> <issue ref>`, the path a corpus
// file under root (or `T1/<name>.t` under t1, the T1_CORPUS value), the byte
// inside it. With t1 empty it checks no T1 entry and returns how many it left.
func lexerListErrors(name, list, root, t1 string) (out []string, unchecked int) {
	for i, line := range strings.Split(list, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") { // as the ratchet's textLines skips them
			continue
		}
		bad := func(why string) { out = append(out, fmt.Sprintf("%s:%d: %s: %q", name, i+1, why, line)) }
		if len(f) != 3 {
			bad("want <path> <byte> <issue ref>")
			continue
		}
		if !issueRef.MatchString(f[2]) {
			bad("the issue ref is not a uuid prefix of two or more whole segments, like 01a1882d-d8a4")
		}
		var file string
		if t1name, ok := strings.CutPrefix(f[0], "T1/"); ok {
			if t1 == "" {
				unchecked++
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
		// The ratchet matches "<path> <byte>" as text, so only the plain decimal form can match.
		if b, err := strconv.ParseUint(f[1], 10, 63); err != nil || strconv.FormatUint(b, 10) != f[1] || int64(b) >= fi.Size() {
			bad(fmt.Sprintf("the byte is not a plain decimal offset below the file's size, %d", fi.Size()))
		}
	}
	return out, unchecked
}

// TestLexerLists checks every entry, and without T1_CORPUS every entry but
// T1's and then skips, as the repository's corpus-backed ratchets do; CI's
// corpus-backed job sets T1_CORPUS and fails on the skip.
func TestLexerLists(t *testing.T) {
	t1 := os.Getenv("T1_CORPUS")
	unchecked := 0
	for _, name := range []string{"lex-unmapped-known.txt", "lexer-defects.txt", filepath.Join("tdt", "lexer-rooted.txt")} {
		list, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		errs, n := lexerListErrors(name, string(list), repoRoot, t1)
		for _, e := range errs {
			t.Error(e)
		}
		unchecked += n
	}
	if unchecked > 0 {
		t.Skipf("T1_CORPUS unset: %d T1 entries not checked", unchecked)
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
		{"comment", "# the entries below are reviewed", true},
		{"last byte", std5k + " " + itoa(size-1) + " 01a1882d-d8a4", true},
		{"guacamole snippet", "internal/chart/testdata/guacamole/parses/Statements-Block-001.pl 0 01a1882d-d8a4", true},
		{"byte at the end", std5k + " " + itoa(size) + " 01a1882d-d8a4", false},
		{"byte past the end", std5k + " " + itoa(size+10) + " 01a1882d-d8a4", false},
		{"negative byte", std5k + " -1 01a1882d-d8a4", false},
		{"leading zero", std5k + " 07 01a1882d-d8a4", false},
		{"plus sign", std5k + " +7 01a1882d-d8a4", false},
		{"byte not a number", std5k + " x 01a1882d-d8a4", false},
		{"not a corpus file", "internal/chart/README.md 0 01a1882d-d8a4", false},
		{"no such file", "internal/chart/testdata/guacamole/parses/no-such.pl 0 01a1882d-d8a4", false},
		{"missing ref", std5k + " 0", false},
		{"missing byte", std5k + " 01a1882d-d8a4", false},
		{"extra field", std5k + " 0 01a1882d-d8a4 x", false},
		{"ref through the third segment", std5k + " 0 01a18fdc-bf40-7928", true},
		{"ref through the fourth segment", std5k + " 0 01a18fdc-bf40-7928-8d9a", true},
		{"ref as the full uuid", std5k + " 0 01a18fdc-bf40-7928-8d9a-618bb50324da", true},
		{"ref cut inside a segment", std5k + " 0 01a18fdc-bf40-79", false},
		{"ref as the second segment", std5k + " 0 d8a4", false},
		{"ref as the short id", std5k + " 0 01a1882d", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs, _ := lexerListErrors("synthetic", tc.line+"\n", repoRoot, "")
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
	if errs, n := lexerListErrors("synthetic", list, repoRoot, ""); len(errs) > 0 || n != 1 {
		t.Errorf("T1_CORPUS unset: got %v and %d unchecked, want no failure and the entry unchecked", errs, n)
	}
	if errs, _ := lexerListErrors("synthetic", list, repoRoot, filepath.Join(dir, "missing")); !strings.Contains(strings.Join(errs, "\n"), "T1_CORPUS") {
		t.Errorf("T1_CORPUS naming no directory: want a failure naming T1_CORPUS, got %v", errs)
	}
	if errs, n := lexerListErrors("synthetic", list, repoRoot, dir); len(errs) > 0 || n != 0 {
		t.Errorf("T1_CORPUS naming a directory holding the file: %v, %d unchecked", errs, n)
	}
}
