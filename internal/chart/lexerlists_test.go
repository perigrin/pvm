// ABOUTME: Holds the lexer lists, lex-unmapped-known.txt, lexer-defects.txt and tdt/lexer-rooted.txt, to their entry form.
// ABOUTME: Each entry is `<path> <byte> <issue ref>`, the path a corpus file (T1, in testdata/perlonjava/unit or T1_CORPUS), the byte inside it.

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

	"tamarou.com/pvm/internal/chart/lex"
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
// inside it.
func lexerListErrors(name, list, root, t1 string) (out []string) {
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
			if fi, err := os.Stat(t1); err != nil || !fi.IsDir() {
				bad(t1Desc(t1) + " names no directory")
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
	return out
}

// t1Root is T1_CORPUS when set, else the in-tree copy.
func t1Root() string {
	if t1 := os.Getenv("T1_CORPUS"); t1 != "" {
		return t1
	}
	return t1Default()
}

func t1Default() string { return filepath.Join(repoRoot, "testdata", "perlonjava", "unit") }

// t1Desc says where a T1 root came from, for the message about it: the
// in-tree default, or T1_CORPUS.
func t1Desc(root string) string {
	if root == t1Default() {
		return "the in-tree T1 " + root
	}
	return "T1_CORPUS=" + root
}

// TestLexerLists checks every entry, T1's against t1Root.
func TestLexerLists(t *testing.T) {
	t1 := t1Root()
	for _, name := range []string{"lex-unmapped-known.txt", "lexer-defects.txt", filepath.Join("tdt", "lexer-rooted.txt")} {
		list, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		errs := lexerListErrors(name, string(list), repoRoot, t1)
		for _, e := range errs {
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
			errs := lexerListErrors("synthetic", tc.line+"\n", repoRoot, t1Default())
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
	if errs := lexerListErrors("synthetic", list, repoRoot, filepath.Join(dir, "missing")); !strings.Contains(strings.Join(errs, "\n"), "T1_CORPUS") {
		t.Errorf("T1_CORPUS naming no directory: want a failure naming T1_CORPUS, got %v", errs)
	}
	if errs := lexerListErrors("synthetic", list, repoRoot, dir); len(errs) > 0 {
		t.Errorf("T1_CORPUS naming a directory holding the file: %v", errs)
	}
}

// bc7bAccepted is how many lex-unmapped-known.txt entries filed against 01a1ad1d-bc7b perl
// accepts by the committed labels: 90 T1 files and 1 in-tree file.
const bc7bAccepted = 91

// TestLexerListHidesNoChartGap rewrites every brace 01a1ad1d-bc7b leaves unmapped as `+{`
// in each perl-accepted file on lex-unmapped-known.txt under that defect, and holds the chart to
// accepting the result: the list takes out the lexer's defect and nothing of the chart's own.
func TestLexerListHidesNoChartGap(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	accepted := map[string]bool{}
	for name, prefix := range map[string]string{"t1.txt": "T1/", "in-tree.txt": ""} {
		lf, err := ReadLabels(filepath.Join(labelsDir, name))
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range lf.Labels {
			accepted[prefix+l.Path] = !l.Refused()
		}
	}
	list, err := os.ReadFile(filepath.Join("testdata", "lex-unmapped-known.txt"))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(string(list), "\n") {
		f := strings.Fields(line)
		if len(f) != 3 || f[2] != "01a1ad1d-bc7b" || !accepted[f[0]] {
			continue
		}
		n++
		path := filepath.Join(repoRoot, f[0])
		if name, ok := strings.CutPrefix(f[0], "T1/"); ok {
			path = filepath.Join(t1Root(), name)
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// Each pass rewrites the first unmapped brace; one the rewrite leaves unmapped stops it.
		for last := -1; ; {
			_, at := lex.Lex(src, tab.Tokens)
			if at <= last || src[at] != '{' {
				break
			}
			src, last = slices.Insert(src, at, '+'), at+1
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s: refused at byte %d of the rewritten source, expecting %v", f[0], v.at, v.expected)
		}
	}
	if n != bc7bAccepted {
		t.Errorf("%d perl-accepted 01a1ad1d-bc7b entries, want %d", n, bc7bAccepted)
	}
}
