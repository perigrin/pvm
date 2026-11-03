// ABOUTME: Checks the -DT alignment fixtures in TDT_FIXTURES: one per corpus file, spans in order inside the file, families listed.
// ABOUTME: Also shows the check failing on synthetic malformed fixtures; nothing here runs perl or the aligner.

package chart

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var tdtDir = filepath.Join("testdata", "tdt")

// readFittingFamilies returns the family names fitting-families.txt lists,
// the first field of each line that is not blank or a comment.
func readFittingFamilies(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(tdtDir, "fitting-families.txt"))
	if err != nil {
		t.Fatal(err)
	}
	fams := map[string]bool{}
	for _, l := range strings.Split(string(data), "\n") {
		if f := strings.Fields(l); len(f) > 0 && !strings.HasPrefix(f[0], "#") {
			fams[f[0]] = true
		}
	}
	return fams
}

// tdtFixtureErrors returns one message per defect in an alignment fixture
// for a corpus file of size bytes (the form is in offline/README.md):
// a `# capture:` header; for a whole capture, pair lines `<start> <end>
// <TYPE> [<family> <TYPE>...]...`, unaligned lines `<start> <end> -
// [<family>]` and unmatched toke.c runs `unmatched <byte> <TYPE>...`, in
// file order, without overlap, inside the file, every family listed.
func tdtFixtureErrors(text string, size int, fams map[string]bool) []string {
	var out []string
	bad := func(n int, format string, a ...any) {
		out = append(out, fmt.Sprintf("line %d: ", n)+fmt.Sprintf(format, a...))
	}
	capture, prevStart, prevEnd := "", -1, 0
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(nil, 1<<20)
	for n := 1; sc.Scan(); n++ {
		l := sc.Text()
		if c, ok := strings.CutPrefix(l, "# capture: "); ok {
			capture = c
			continue
		}
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		if capture != "whole" {
			bad(n, "a %q capture has no alignment, but this line aligns", capture)
			continue
		}
		f := strings.Fields(l)
		if len(f) >= 3 && f[0] == "unmatched" {
			b, err := strconv.Atoi(f[1])
			switch {
			case err != nil:
				bad(n, "byte %q is not a number", f[1])
			case b > size:
				bad(n, "unmatched run at byte %d is past the end of the file (%d bytes)", b, size)
			case b < prevEnd:
				bad(n, "unmatched run at byte %d is out of order (previous span ends at %d)", b, prevEnd)
			}
			continue
		}
		if len(f) < 3 {
			bad(n, "want `<start> <end> <TYPE>...`, `<start> <end> - [<family>]` or `unmatched <byte> <TYPE>...`: %q", l)
			continue
		}
		s, err1 := strconv.Atoi(f[0])
		e, err2 := strconv.Atoi(f[1])
		switch {
		case err1 != nil || err2 != nil || s < 0 || e <= s:
			bad(n, "span %s %s is not a byte range", f[0], f[1])
			continue
		case e > size:
			bad(n, "span %d %d is past the end of the file (%d bytes)", s, e, size)
		case s < prevEnd && s >= prevStart:
			bad(n, "span %d %d overlaps the span before it, which ends at %d", s, e, prevEnd)
		case s < prevEnd:
			bad(n, "span %d %d is out of order (previous span ends at %d)", s, e, prevEnd)
		}
		prevStart, prevEnd = s, e
		for _, w := range f[3:] {
			if w[0] >= 'a' && w[0] <= 'z' && !fams[w] {
				bad(n, "family %q is not in fitting-families.txt", w)
			}
		}
	}
	if capture == "" {
		out = append(out, "no `# capture:` line")
	}
	return out
}

// checkTDTFixtures checks the fixture in dir of each corpus file, keys
// mapped to the source path whose size bounds the spans.
func checkTDTFixtures(t *testing.T, dir string, keys, paths []string, fams map[string]bool) {
	t.Helper()
	for i, k := range keys {
		fx, err := os.ReadFile(filepath.Join(dir, k+".tdt"))
		if err != nil {
			t.Errorf("%s: no aligned-pairs fixture (%v)", k, err)
			continue
		}
		fi, err := os.Stat(paths[i])
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range tdtFixtureErrors(string(fx), int(fi.Size()), fams) {
			t.Errorf("%s.tdt: %s", k, p)
		}
	}
}

// TestTDTFixtures: every in-tree corpus file has a well-formed aligned-pairs
// fixture in TDT_FIXTURES, and every T1 file when T1_CORPUS is set. The
// fixtures are generated offline by tdt-align.sh, too large to commit
// (perigrin: "Don't commit fixtures"), so an unset TDT_FIXTURES skips.
func TestTDTFixtures(t *testing.T) {
	dir := os.Getenv("TDT_FIXTURES")
	if dir == "" {
		t.Skip("TDT_FIXTURES unset: the -DT fixtures are generated offline by offline/tdt-align.sh")
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("TDT_FIXTURES=%s names no directory (%v)", dir, err)
	}
	fams := readFittingFamilies(t)
	files := inTreeFiles(t)
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = filepath.Join(repoRoot, f)
	}
	checkTDTFixtures(t, dir, files, paths, fams)

	root := os.Getenv("T1_CORPUS")
	if root == "" {
		t.Log("T1_CORPUS unset: T1's fixtures are not checked")
		return
	}
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		t.Fatalf("T1_CORPUS=%s names no directory (%v)", root, err)
	}
	matches, err := filepath.Glob(filepath.Join(root, "*.t"))
	if err != nil || len(matches) == 0 {
		t.Fatalf("T1_CORPUS=%s holds no top-level *.t (%v)", root, err)
	}
	keys := make([]string, len(matches))
	for i, m := range matches {
		keys[i] = "T1/" + filepath.Base(m)
	}
	checkTDTFixtures(t, dir, keys, matches, fams)
}

// TestTDTFixturesRefusesMalformed: each defect in a synthetic fixture is
// reported, and the well-formed fixture passes.
func TestTDTFixturesRefusesMalformed(t *testing.T) {
	fams := map[string]bool{"variable": true, "brace-semicolon": true}
	const head = "# tdt: x.pl\n# capture: whole\n"
	const body = "0 2 KW_MY\n3 5 PERLY_DOLLAR variable PRIVATEREF\n6 7 -\nunmatched 8 KW_USE_or_NO BAREWORD\n8 9 PERLY_BRACE_CLOSE brace-semicolon PERLY_SEMICOLON\n"
	cases := []struct {
		name, text string
		want       string // a reported problem contains this; "" means none
	}{
		{"well formed", head + body, ""},
		{"refused capture without alignment", "# capture: refused: perl exited 255\n", ""},
		{"overlapping spans", head + "0 3 KW_MY\n2 5 PERLY_DOLLAR\n", "overlaps"},
		{"spans out of order", head + "3 5 PERLY_DOLLAR\n0 2 KW_MY\n", "out of order"},
		{"span past the end", head + "8 11 PERLY_DOLLAR\n", "past the end"},
		{"unmatched run past the end", head + "unmatched 11 KW_USE_or_NO\n", "past the end"},
		{"family missing from the list", head + "3 5 PERLY_DOLLAR quote PRIVATEREF\n", `"quote" is not in fitting-families.txt`},
		{"unaligned family missing from the list", head + "6 7 - heredoc\n", `"heredoc" is not in fitting-families.txt`},
		{"refused capture with alignment", "# capture: refused: perl exited 255\n0 2 KW_MY\n", "has no alignment"},
		{"no capture line", "0 2 KW_MY\n", "no `# capture:` line"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := tdtFixtureErrors(c.text, 10, fams)
			if c.want == "" {
				if len(got) != 0 {
					t.Fatalf("tdtFixtureErrors = %q, want no problem", got)
				}
				return
			}
			for _, p := range got {
				if strings.Contains(p, c.want) {
					return
				}
			}
			t.Fatalf("tdtFixtureErrors = %q, want a problem containing %q", got, c.want)
		})
	}
}
