// ABOUTME: Parses compiled.txt, the perl.git t/ files perl 5.44.0 compiles, and holds >= 90% at 0 Unknown.
// ABOUTME: Pins the unclean set in testdata/compiledclean.txt, so a swap or a slide to the floor fails.

package parse_test

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// TestCompiledSetFailsWithoutFile: a missing, empty or unparseable
// compiled.txt is an error, never a set of 0 files that passes 0/0.
func TestCompiledSetFailsWithoutFile(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]*string{
		"missing":        nil,
		"empty":          ptr(""),
		"only a newline": ptr("\n"),
		"blank line":     ptr("base/if.t\n\nbase/lex.t\n"),
		"spaced line":    ptr("base/if.t extra\n"),
		"duplicate line": ptr("base/if.t\nbase/if.t\n"),
	} {
		path := filepath.Join(dir, name)
		if content != nil {
			if err := os.WriteFile(path, []byte(*content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if paths, err := compiledSetAt(path); err == nil {
			t.Errorf("%s: read as %d paths %q, want an error", name, len(paths), paths)
		}
	}
	if paths, err := compiledSetAt(filepath.Join(ratchetDir, "compiled.txt")); err != nil || len(paths) == 0 {
		t.Errorf("the committed compiled.txt: %d paths, %v", len(paths), err)
	}
}

func ptr(s string) *string { return &s }

// TestGradedSetExcludesOnlyExcludedAndError: the graded set the difference
// log compares against keeps every row the oracle graded, the no-answer and
// subject-error rows among them, and leaves out only the files it could not
// measure.
func TestGradedSetExcludesOnlyExcludedAndError(t *testing.T) {
	b, err := parseoracle.ParseBaseline([]byte(
		"interpreter = 5.044000\nrevision = abc\nthreads = define\n" +
			"exact 0 - a.t\nwider 0 - b.t\nno-answer 0 - c.t\nwrong 0 - d.t\n" +
			"subject-error 0 - e.t\nexcluded 0 - f.t\nerror 0 - g.t\n"))
	if err != nil {
		t.Fatal(err)
	}
	got := gradedSet(b)
	want := []string{"a.t", "b.t", "c.t", "d.t", "e.t"}
	if len(got) != len(want) {
		t.Errorf("graded set is %v, want %v", got, want)
	}
	for _, p := range want {
		if !got[p] {
			t.Errorf("graded set leaves out %s", p)
		}
	}
}

// TestCompiledCorpusMissingFileCountsAsUnclean: a compiled-set file the
// corpus does not hold (a wrong pin, a partial checkout) is unclean and named,
// not skipped out of the denominator.
func TestCompiledCorpusMissingFileCountsAsUnclean(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "here.t"), []byte("print 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unclean := measureCompiled(dir, []string{"here.t", "gone.t"})
	if len(unclean) != 1 || unclean["gone.t"] == "" {
		t.Errorf("unclean is %v, want gone.t alone, with its reason", unclean)
	}
}

// TestParseWithoutErrorPinsUncleanSet: compiledclean.txt pins the unclean
// set both ways. A file outside it that turns unclean fails, so a swap cannot
// hide under the floor; a listed file that turns clean fails, so the pin is
// re-recorded with the change that earned it; and a header that disagrees
// with its body fails.
func TestParseWithoutErrorPinsUncleanSet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "compiledclean.txt")
	if err := os.WriteFile(path, []byte(renderCompiledClean(3, map[string]string{"b.t": "2 Unknown"})), 0o644); err != nil {
		t.Fatal(err)
	}
	pinned, err := readCompiledClean(path, 3)
	if err != nil {
		t.Fatal(err)
	}
	if moves := uncleanSetMoves(pinned, map[string]string{"b.t": "1 Unknown"}); len(moves) != 0 {
		t.Errorf("the pinned set reads as moved: %v", moves)
	}
	if moves := uncleanSetMoves(pinned, map[string]string{"b.t": "1 Unknown", "a.t": "1 Unknown"}); len(moves) != 1 {
		t.Errorf("a.t turning unclean: moves %v, want one", moves)
	}
	if moves := uncleanSetMoves(pinned, map[string]string{}); len(moves) != 1 {
		t.Errorf("b.t turning clean: moves %v, want one", moves)
	}
	if _, err := readCompiledClean(path, 4); err == nil {
		t.Error("a pin over 3 files read as a pin over 4")
	}
	for name, content := range map[string]string{
		"wrong clean count": "# 3/3 clean\nb.t\n",
		"no header":         "b.t\n",
		"missing":           "",
	} {
		bad := filepath.Join(t.TempDir(), "compiledclean.txt")
		if content != "" {
			if err := os.WriteFile(bad, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := readCompiledClean(bad, 3); err == nil {
			t.Errorf("%s: read without error", name)
		}
	}
}

// TestParseWithoutErrorOnCompiledCorpus is M3's parse-without-error metric:
// of the files perl 5.44.0 compiles (compiled.txt), at least 90% parse
// through Session.ParseFileFrom with 0 Unknown. A refusal is an Unknown, so
// 0 Unknown is parsed without error.
//
// re/pat_psycho.t is in the set: perl compiles it, and parsing it needs no
// oracle, though the oracle sweep errors on it at its 8m budget.
// comp/final_line_num.t is not: it ends in `print 1+`, and its BEGIN
// __DIE__ handler prints "ok 1" and exits 0, so `perl -c`'s rc 0 is the
// file's own trick (see perlRejects).
//
// The unclean files are pinned in testdata/compiledclean.txt, rewritten with
// -parse.update-ratchet; the docs quote its header.
func TestParseWithoutErrorOnCompiledCorpus(t *testing.T) {
	dir, _ := perlGitTFiles(t)
	paths := readCompiledSet(t)
	if want := m3StartCount(t, "compiled"); len(paths) != want {
		t.Fatalf("compiled.txt has %d files, m3start.txt records compiled = %d", len(paths), want)
	}
	logGradedDifference(t, paths)

	unclean := measureCompiled(dir, paths)
	clean := len(paths) - len(unclean)
	t.Logf("%d/%d clean", clean, len(paths))
	for _, p := range slices.Sorted(maps.Keys(unclean)) {
		t.Logf("unclean: %s: %s", p, unclean[p])
	}
	if clean*10 < len(paths)*9 {
		t.Errorf("%d/%d compiled files parse with 0 Unknown, below the 90%% floor", clean, len(paths))
	}

	pin := filepath.Join("testdata", "compiledclean.txt")
	if *updateRatchet {
		if err := os.WriteFile(pin, []byte(renderCompiledClean(len(paths), unclean)), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("rewrote %s; commit it with the change that moved it", pin)
		return
	}
	pinned, err := readCompiledClean(pin, len(paths))
	if err != nil {
		t.Fatalf("%v\nrun with -parse.update-ratchet to create it", err)
	}
	for _, move := range uncleanSetMoves(pinned, unclean) {
		t.Error(move)
	}
}

// gradedSet is the files baseline.txt's sweep graded: every row but the
// excluded and error ones, which it could not measure.
func gradedSet(b parseoracle.Baseline) map[string]bool {
	graded := map[string]bool{}
	for _, r := range b.Rows {
		if r.Status != "excluded" && r.Status != "error" {
			graded[r.Path] = true
		}
	}
	return graded
}

// logGradedDifference names every file in one of the compiled set and
// baseline.txt's graded set and not the other. The two are different sets,
// so this logs and never asserts.
func logGradedDifference(t *testing.T, compiled []string) {
	t.Helper()
	b, err := parseoracle.LoadBaseline(filepath.Join(ratchetDir, "baseline.txt"))
	if err != nil {
		t.Fatal(err)
	}
	graded := gradedSet(b)
	inCompiled := map[string]bool{}
	for _, p := range compiled {
		inCompiled[p] = true
		if !graded[p] {
			t.Logf("compiled, not graded: %s", p)
		}
	}
	for _, r := range b.Rows {
		if graded[r.Path] && !inCompiled[r.Path] {
			t.Logf("graded, not compiled: %s", r.Path)
		}
	}
}

// measureCompiled parses each compiled-set file under dir and returns the
// unclean ones, each with why: its Unknown count, or the error reading it.
// A file that will not read stays in the set as unclean.
func measureCompiled(dir string, paths []string) map[string]string {
	roots := perlGitTRoots(dir)
	s := parse.NewSession()
	unclean := map[string]string{}
	for _, rel := range paths {
		n, err := s.ParseFileFrom(filepath.Join(dir, rel), roots...)
		if err != nil {
			unclean[rel] = err.Error()
		} else if c := countUnknown(n); c != 0 {
			unclean[rel] = fmt.Sprintf("%d Unknown", c)
		}
	}
	return unclean
}

// renderCompiledClean writes compiledclean.txt: a `# N/M clean` header and
// the unclean files, one path a line, sorted.
func renderCompiledClean(total int, unclean map[string]string) string {
	var out strings.Builder
	out.WriteString("# ABOUTME: The compiled set's files that do not parse with 0 Unknown, and the clean count over it.\n")
	out.WriteString("# ABOUTME: Written by TestParseWithoutErrorOnCompiledCorpus under -parse.update-ratchet; DO NOT EDIT BY HAND.\n")
	fmt.Fprintf(&out, "# %d/%d clean\n", total-len(unclean), total)
	for _, p := range slices.Sorted(maps.Keys(unclean)) {
		out.WriteString(p + "\n")
	}
	return out.String()
}

// cleanHeader is compiledclean.txt's count line.
var cleanHeader = regexp.MustCompile(`^# ([0-9]+)/([0-9]+) clean$`)

// readCompiledClean reads compiledclean.txt's unclean set, failing when its
// header disagrees with its body or counts other than total files.
func readCompiledClean(path string, total int) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	clean, counted := -1, -1
	pinned := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		if m := cleanHeader.FindStringSubmatch(line); m != nil {
			clean, _ = strconv.Atoi(m[1])
			counted, _ = strconv.Atoi(m[2])
		} else if line != "" && !strings.HasPrefix(line, "#") {
			pinned[line] = true
		}
	}
	switch {
	case counted < 0:
		return nil, fmt.Errorf("%s has no `# N/M clean` header", path)
	case counted != total:
		return nil, fmt.Errorf("%s counts %d files, the compiled set has %d", path, counted, total)
	case clean != total-len(pinned):
		return nil, fmt.Errorf("%s says %d/%d clean and lists %d unclean", path, clean, counted, len(pinned))
	}
	return pinned, nil
}

// uncleanSetMoves names each file whose cleanliness differs from the pin.
func uncleanSetMoves(pinned map[string]bool, unclean map[string]string) []string {
	var moves []string
	for _, p := range slices.Sorted(maps.Keys(unclean)) {
		if !pinned[p] {
			moves = append(moves, fmt.Sprintf("%s is unclean (%s) and compiledclean.txt does not list it", p, unclean[p]))
		}
	}
	for p := range pinned {
		if _, ok := unclean[p]; !ok {
			moves = append(moves, fmt.Sprintf("%s is clean and compiledclean.txt lists it unclean: "+
				"re-run with -parse.update-ratchet and commit the pin with the change that earned it", p))
		}
	}
	return moves
}
