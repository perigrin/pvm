// ABOUTME: The M1 parse-rate ratchet: per-file Unknown counts over T1, failing in either direction.
// ABOUTME: A locally correct fix that makes the corpus worse is the failure this exists to catch.

package parse_test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/conformance"
	"tamarou.com/pvm/internal/parse"
)

var updateRatchet = flag.Bool("parse.update-ratchet", false,
	"rewrite the committed per-file Unknown-node baseline")

// A THIRD ratchet, and the decision is deliberate rather than defaulted.
//
// `internal/lexer` has one keyed on error tokens over perl5's `t/`, and
// `internal/parseoracle` has one keyed on four verdict buckets. This one is
// keyed on Unknown nodes over T1. The issue that asked for it also asked
// whether to extract the shared shape instead, and the answer is no for a
// reason that is about Go rather than about taste: lexer's renderRatchet and
// parseRatchet are unexported in `package lexer`, and this file is
// `package parse_test`. Exporting them to share ~40 lines of format code
// would widen a package's API to serve its tests.
//
// What IS shared is the discipline, and that is the part that matters:
// per-file counts, a committed baseline, failure in EITHER direction, and
// regeneration only under an explicit flag. Three copies is where "two is a
// coincidence" stops being true, so if a fourth is ever wanted, extract then.

// t1Files returns T1 graded: PerlOnJava's unit/*.t at TOP LEVEL, 986 files.
//
// Top level only, and that is not an accident of globbing. Walking the tree
// recursively gives 1,508 and adding module/ gives 1,935; both are different
// corpora with different numbers. Spec §7.3.5 and the M1 gate both mean the
// 986, and a measurement quoted against the wrong set is how this project
// once read a change as removing 277 Unknown nodes when on T1 it added 42.
func t1Files(t *testing.T) (string, []string) {
	t.Helper()

	root := os.Getenv("T1_CORPUS")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no T1_CORPUS and no home directory: %v", err)
		}
		root = filepath.Join(home, "dev", "PerlOnJava",
			"src", "test", "resources", "unit")
	}
	if _, err := os.Stat(root); err != nil {
		t.Skipf("T1 corpus not present at %s: %v\n"+
			"set T1_CORPUS to a PerlOnJava checkout's unit/ to run the M1 metrics",
			root, err)
	}

	matches, err := filepath.Glob(filepath.Join(root, "*.t"))
	if err != nil {
		t.Skipf("globbing the corpus: %v", err)
	}
	var files []string
	for _, m := range matches {
		files = append(files, filepath.Base(m))
	}
	sort.Strings(files)
	return root, files
}

// t1CorpusNote says WHICH corpus the three T1 baselines measure, and goes in
// each of their headers.
//
// It is here rather than in renderRatchet because renderRatchet also writes
// perlgitt.ratchet, which is over perl's own 620 files. A header that named the
// wrong corpus would be the same defect as a header whose count disagrees with
// its body: a number that reads as authoritative and is not.
const t1CorpusNote = "T1 graded is PerlOnJava's unit/*.t at TOP LEVEL. Walking\n" +
	"recursively gives 1508 files and adding module/ gives 1935;\n" +
	"both are different corpora. This is the 986."

// countUnknown counts Unknown nodes in a tree.
//
// Every Unknown in T1 is a childless leaf as of 3c997a6b, so the recursion is
// defensive rather than load-bearing. It stays because that is a measured
// property of today's parser, not an invariant anything enforces.
func countUnknown(n *parse.Node) int {
	total := 0
	if n.Kind == parse.Unknown {
		total = 1
	}
	for _, c := range n.Children {
		total += countUnknown(c)
	}
	return total
}

// TestParseRatchet is the M1 gate's parse-rate metric.
//
// It fails when the number moves in EITHER direction, which is the whole
// point. An improvement that fails is not a nuisance: it forces the commit
// that earned it to carry the new baseline, so the number in the repository
// is always the number the tests produce.
//
// Measured at 3c997a6b: 408 of 986 files clean (41.4%), 4,944 Unknown nodes.
// This session attempted three parser fixes, each verifiably correct against
// toke.c or against running perl, and two of them made T1 worse -- caught
// only by a throwaway sweep that no longer exists. That is what this replaces.
//
// A COUNT IS NOT A SIZE, and that is this ratchet's known blind spot. It
// counts Unknown NODES, so one refusal swallowing a whole region scores
// better than several small ones covering less of the file. Measured while
// 01a0ad52 landed:
//
//	jvm_eval_nested_compound_assignment.t   1 -> 3 nodes   1,317 -> 53 bytes
//	overload_compound_assignment.t          1 -> 10 nodes  2,561 -> 1,424
//
// and again while 01a0de8b landed, the labelled bare block:
//
//	socket_options.t                        29 -> 39 nodes  4,139 -> 3,160 bytes
//
// The file holds seven `SKIP: { ... }` blocks. Each one now OPENS, so the
// parenless `is`/`ok` calls inside it refuse one statement at a time instead
// of disappearing into a single swallowed span -- 979 bytes moved out of
// Unknown, from 76.1% of the file to 58.1%. A rise here is what a cascade
// looks like when it unwinds.
//
// Both read as regressions here and both are large improvements. The
// discipline that keeps that honest is not a second metric -- it is that
// every rise must be measured in BYTES before the baseline is regenerated,
// and the finding recorded in the commit. A rise nobody explained is a
// regression; a rise with a byte count beside it is evidence.
//
// Each file is parsed as perl compiles it, with its `use` statements
// resolved: through a loader rooted at the file's own directory and at the
// @INC of the perl the corpus is measured against (perigrin, 2026-09-30).
// Without one, every file that calls Test::More's `is` or `ok` without
// parentheses refused on a name perl knows -- about 200 files of T1.
func TestParseRatchet(t *testing.T) {
	dir, files := t1Files(t)
	roots := t1Roots(t)

	now := make(map[string]int, len(files))
	for _, rel := range files {
		n, err := parse.ParseFileFrom(filepath.Join(dir, rel), roots...)
		if err != nil {
			continue
		}
		now[rel] = countUnknown(n)
	}

	checkRatchet(t, filepath.Join("testdata", "t1.ratchet"),
		"Unknown nodes per T1 file, through parse.ParseFileFrom with perl 5.42's @INC.\n"+t1CorpusNote, now)
}

// t1Roots is the @INC of the perl the corpus is measured against: where it
// finds the modules a T1 file uses. Skips when that perl is not installed,
// as t1Files skips when the corpus is not.
//
// An installed perl ships no .xs, so the perl.git checkout's dist/, cpan/
// and ext/ follow, when it is present, for a core module's XS source:
// Storable's `dclone` is `$` only in dist/Storable/Storable.xs (perigrin,
// 2026-09-30). Being after @INC, they are reached for a module's .pm only
// when perl 5.42 does not have it.
func t1Roots(t *testing.T) []string {
	t.Helper()
	perl, err := conformance.PerlPath()
	if err != nil {
		t.Skipf("no perl 5.42 to resolve T1's modules from: %v", err)
	}
	out, err := exec.Command(perl, "-e", `print "$_\n" for grep { !ref && -d } @INC`).Output()
	if err != nil {
		t.Skipf("asking %s for @INC: %v", perl, err)
	}
	roots := strings.Fields(string(out))
	if root := perl5Root(); root != "" {
		for _, tree := range []string{"dist", "cpan", "ext"} {
			if dir := filepath.Join(root, tree); isDir(dir) {
				roots = append(roots, dir)
			}
		}
	}
	return roots
}

// isDir reports whether path names a directory.
func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// TestParsedFilesRoundTrip is the M1 gate's round-trip metric: everything
// that parses also reproduces its source from the tree.
//
// Unlike the ratchet this is a hard gate, not a moving number. A tree that
// cannot reproduce its input has lost bytes, and no later pass can recover
// them. It holds over the WHOLE corpus rather than over the clean files,
// because an Unknown spans its bytes and must emit them too.
func TestParsedFilesRoundTrip(t *testing.T) {
	dir, files := t1Files(t)

	var failed []string
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		if got := parse.Parse(src).SourceText(src); got != string(src) {
			failed = append(failed, rel)
		}
	}
	if len(failed) > 0 {
		sort.Strings(failed)
		t.Errorf("%d of %d file(s) do not round-trip:\n  %s",
			len(failed), len(files), strings.Join(failed, "\n  "))
	}
}

// renderRatchet writes the baseline: one "count path" line per file, sorted,
// with a header naming what the numbers are and how to move them.
//
// what describes the measurement in one line; the rest of the header is the
// same for every ratchet, because the rule about regenerating it is.
func renderRatchet(what string, counts map[string]int) string {
	files := make([]string, 0, len(counts))
	total := 0
	clean := 0
	for rel, n := range counts {
		files = append(files, rel)
		total += n
		if n == 0 {
			clean++
		}
	}
	sort.Strings(files)

	var b strings.Builder
	// what may be several lines: the measurement, then a note saying WHICH
	// corpus this is. Naming the population is not decoration -- t1, t1canon
	// and t1fidelity are all over PerlOnJava's 986, while perlgitt is over
	// perl's own 620, and a brief that conflated two such populations already
	// cost this project a claim that could not have been true.
	for _, line := range strings.Split(what, "\n") {
		b.WriteString("# " + line + "\n")
	}
	b.WriteString("#\n")
	b.WriteString(fmt.Sprintf("# %d files, %d clean (%.1f%%), %d total.\n",
		len(files), clean, 100*float64(clean)/float64(len(files)), total))
	b.WriteString("#\n")
	b.WriteString("# Regenerate with -parse.update-ratchet and commit the result\n")
	b.WriteString("# WITH the change that moved it. A baseline updated on its own\n")
	b.WriteString("# is a number nobody can attribute.\n")
	for _, rel := range files {
		b.WriteString(strconv.Itoa(counts[rel]))
		b.WriteByte(' ')
		b.WriteString(rel)
		b.WriteByte('\n')
	}
	return b.String()
}

// checkRatchet compares a measurement against its baseline and fails in
// EITHER direction: a rise is a regression, and a fall is good news that
// still has to be committed with the change that earned it.
//
// Shared by every ratchet in this package. The comparison is the part that
// must not drift between them -- a ratchet that failed in only one direction
// would silently let the number it guards walk.
func checkRatchet(t *testing.T, path, what string, now map[string]int) {
	t.Helper()

	if *updateRatchet {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(renderRatchet(what, now)), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
		t.Logf("rewrote %s (%d files); commit it with the change that moved it",
			path, len(now))
		return
	}

	want, err := parseRatchetFile(path)
	if err != nil {
		t.Fatalf("%v\nrun with -parse.update-ratchet to create it", err)
	}

	var regressed, improved, added, removed []string
	for rel, n := range now {
		was, ok := want[rel]
		if !ok {
			added = append(added, rel)
			continue
		}
		switch {
		case n > was:
			regressed = append(regressed, fmt.Sprintf("%s: %d -> %d", rel, was, n))
		case n < was:
			improved = append(improved, fmt.Sprintf("%s: %d -> %d", rel, was, n))
		}
	}
	for rel := range want {
		if _, ok := now[rel]; !ok {
			removed = append(removed, rel)
		}
	}
	sort.Strings(regressed)
	sort.Strings(improved)
	sort.Strings(added)
	sort.Strings(removed)

	if len(regressed) > 0 {
		t.Errorf("%d file(s) are WORSE than the baseline:\n  %s",
			len(regressed), strings.Join(regressed, "\n  "))
	}
	if len(improved) > 0 {
		t.Errorf("%d file(s) are BETTER than the baseline:\n  %s\n\n"+
			"This is good news and still fails: re-run with -parse.update-ratchet "+
			"and commit the baseline with the change that earned it.",
			len(improved), strings.Join(improved, "\n  "))
	}
	if len(added) > 0 {
		t.Errorf("%d file(s) are in the corpus and not the baseline:\n  %s",
			len(added), strings.Join(added, "\n  "))
	}
	if len(removed) > 0 {
		t.Errorf("%d file(s) are in the baseline and not the corpus:\n  %s\n\n"+
			"Leaving the denominator is a regression, not a cleanup.",
			len(removed), strings.Join(removed, "\n  "))
	}
}

// parseRatchetFile reads a baseline written by renderRatchet.
func parseRatchetFile(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading the baseline at %s: %w", path, err)
	}
	out := make(map[string]int)
	for i, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		count, rel, ok := strings.Cut(line, " ")
		if !ok {
			return nil, fmt.Errorf("%s:%d: no space in %q", path, i+1, line)
		}
		n, err := strconv.Atoi(count)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %q is not a count: %w", path, i+1, count, err)
		}
		out[rel] = n
	}
	return out, nil
}
