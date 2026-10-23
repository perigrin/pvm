// ABOUTME: Holds the op counts mined from perl's t/perf/opcount.t (testdata/opcount/) to the pinned perl.
// ABOUTME: Each fixture is a snippet sub and the count of each named op perl compiles it to, traced to its opcount.t line.

package parseoracle_test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/perltest"
)

// How opcount.t counts, read from its test_opcount (line 49) and
// B::OP::test_opcount_callback (line 32): it walks
// B::svref_2object($coderef)->ROOT with B::walkoptree, so only the snippet
// sub's own optree is counted -- its leavesub and its nextstates, not the
// main program's leave, enter and nextstates. walkoptree follows op kids
// only, and a nested named sub (sub f{} at line 946) is compiled into its
// own CV, so its body is not counted. Every op is counted, none skipped; a
// null op whose targ records the op it replaced is named "ex-" plus that
// op's ppname without its "pp_", and a null op with targ 0 stays "null". An
// op absent from the tree counts 0, so every => 0 pair passes for any
// snippet that lacks the op.
//
// perl 5.42.0 segfaults compiling opcount.t whole (exit 139). perl 5.44.0,
// the pinned perl, compiles it: `perl -c perf/opcount.t` in the oracle's
// shim tree prints "perf/opcount.t syntax OK", and running it prints
// 1..2658 with all 2658 ok. The fixtures are still checked one snippet at a
// time, never by compiling the whole file.

const (
	opcountDir = "testdata/opcount"

	// opcountFloor is the fewest distinct (opcount.t line, op) pairs the
	// fixtures must hold.
	opcountFloor = 200
	// opcountNonzeroFloor is the fewest of those pairs whose count is not
	// zero, so the zero counts, which any snippet lacking the op passes,
	// cannot carry the floor.
	opcountNonzeroFloor = 100
)

// opcountFixture is one test_opcount call: the snippet sub, the
// declarations and pragmas from opcount.t it needs to compile, and the
// count of each op opcount.t expects.
type opcountFixture struct {
	Line    int
	Name    string
	Prelude string
	Snippet string
	Counts  map[string]int
}

var (
	opcountHeading = regexp.MustCompile(`^## L(\d+): (.+)$`)
	opcountRow     = regexp.MustCompile(`^\|\s*([\w-]+)\s*\|\s*(\d+)\s*\|$`)
)

// loadOpcountFixtures reads every fixture in dir's Markdown files other
// than README.md. A missing directory or one with no fixtures is an error,
// not zero assertions.
func loadOpcountFixtures(dir string) ([]opcountFixture, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	var fixtures []opcountFixture
	for _, path := range paths {
		if filepath.Base(path) == "README.md" {
			continue
		}
		got, err := parseOpcountFile(path)
		if err != nil {
			return nil, err
		}
		fixtures = append(fixtures, got...)
	}
	if len(fixtures) == 0 {
		return nil, fmt.Errorf("no opcount fixtures in %s", dir)
	}
	return fixtures, nil
}

// parseOpcountFile reads one fixture file: each fixture is a "## L<line>:
// <name>" heading, an optional ```perl prelude fence, a ```perl fence
// holding the snippet, and a | op | count | table.
func parseOpcountFile(path string) ([]opcountFixture, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var fixtures []opcountFixture
	var fence *string
	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		line := scanner.Text()
		where := fmt.Sprintf("%s:%d", path, n)
		if fence != nil {
			if line == "```" {
				fence = nil
			} else {
				*fence += line + "\n"
			}
			continue
		}
		if m := opcountHeading.FindStringSubmatch(line); m != nil {
			src, _ := strconv.Atoi(m[1])
			fixtures = append(fixtures, opcountFixture{Line: src, Name: m[2], Counts: map[string]int{}})
			continue
		}
		if len(fixtures) == 0 {
			continue
		}
		cur := &fixtures[len(fixtures)-1]
		switch {
		case line == "```perl prelude":
			fence = &cur.Prelude
		case line == "```perl":
			fence = &cur.Snippet
		case opcountRow.MatchString(line):
			m := opcountRow.FindStringSubmatch(line)
			if _, dup := cur.Counts[m[1]]; dup {
				return nil, fmt.Errorf("%s: %s is counted twice in L%d", where, m[1], cur.Line)
			}
			cur.Counts[m[1]], _ = strconv.Atoi(m[2])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	for _, fx := range fixtures {
		if fx.Line <= 0 || strings.TrimSpace(fx.Snippet) == "" || len(fx.Counts) == 0 {
			return nil, fmt.Errorf("%s: fixture L%d %q needs an opcount.t line, a snippet and counts", path, fx.Line, fx.Name)
		}
	}
	return fixtures, nil
}

// opcountPairs counts the distinct (opcount.t line, op) pairs, and how many
// of them expect a nonzero count. A fixture repeated under the same line
// adds nothing.
func opcountPairs(fixtures []opcountFixture) (total, nonzero int) {
	seen := map[string]bool{}
	for _, fx := range fixtures {
		for op, n := range fx.Counts {
			key := fmt.Sprintf("%d %s", fx.Line, op)
			if seen[key] {
				continue
			}
			seen[key] = true
			total++
			if n != 0 {
				nonzero++
			}
		}
	}
	return total, nonzero
}

func mustLoadOpcountFixtures(t *testing.T) []opcountFixture {
	t.Helper()
	fixtures, err := loadOpcountFixtures(opcountDir)
	if err != nil {
		t.Fatal(err)
	}
	return fixtures
}

// TestOpcountFixturesParse: the fixtures hold at least opcountFloor
// distinct op=>count assertions, each from a named opcount.t line.
func TestOpcountFixturesParse(t *testing.T) {
	fixtures := mustLoadOpcountFixtures(t)
	total, _ := opcountPairs(fixtures)
	if total < opcountFloor {
		t.Errorf("the fixtures hold %d distinct op=>count assertions; at least %d are required", total, opcountFloor)
	}
	t.Logf("%d fixtures, %d op=>count assertions", len(fixtures), total)
}

// TestOpcountFixturesNotCarriedByZeroCounts: a zero count passes for any
// snippet lacking the op, so every fixture holds a nonzero count and at
// least opcountNonzeroFloor of the assertions are nonzero.
func TestOpcountFixturesNotCarriedByZeroCounts(t *testing.T) {
	fixtures := mustLoadOpcountFixtures(t)
	for _, fx := range fixtures {
		if !hasNonzeroCount(fx) {
			t.Errorf("L%d %q holds only zero counts", fx.Line, fx.Name)
		}
	}
	_, nonzero := opcountPairs(fixtures)
	if nonzero < opcountNonzeroFloor {
		t.Errorf("%d nonzero op=>count assertions; at least %d are required", nonzero, opcountNonzeroFloor)
	}

	// The floor is measured over nonzero counts, so a set of zero counts
	// large enough to pass opcountFloor cannot pass it.
	zeros := opcountFixture{Line: 1, Counts: map[string]int{}}
	for i := 0; i < opcountFloor; i++ {
		zeros.Counts[fmt.Sprintf("op%d", i)] = 0
	}
	if total, nonzero := opcountPairs([]opcountFixture{zeros}); total < opcountFloor || nonzero != 0 {
		t.Errorf("all-zero fixture measured as %d pairs, %d nonzero; want %d, 0", total, nonzero, opcountFloor)
	}
}

func hasNonzeroCount(fx opcountFixture) bool {
	for _, n := range fx.Counts {
		if n != 0 {
			return true
		}
	}
	return false
}

// TestOpcountFixturesHaveDistinctSources: no two fixtures claim one
// opcount.t line, and duplicating a fixture does not raise the count.
func TestOpcountFixturesHaveDistinctSources(t *testing.T) {
	fixtures := mustLoadOpcountFixtures(t)
	claimed := map[int]string{}
	for _, fx := range fixtures {
		if other, dup := claimed[fx.Line]; dup {
			t.Errorf("L%d is claimed by both %q and %q", fx.Line, other, fx.Name)
		}
		claimed[fx.Line] = fx.Name
	}

	total, nonzero := opcountPairs(fixtures)
	doubled, doubledNonzero := opcountPairs(append(append([]opcountFixture{}, fixtures...), fixtures...))
	if doubled != total || doubledNonzero != nonzero {
		t.Errorf("duplicating every fixture moved the count from %d (%d nonzero) to %d (%d nonzero)",
			total, nonzero, doubled, doubledNonzero)
	}
}

// TestOpcountFixturesRejectEmptyDir: a missing directory, or one with no
// fixtures, is an error rather than zero assertions.
func TestOpcountFixturesRejectEmptyDir(t *testing.T) {
	empty := t.TempDir()
	if err := os.WriteFile(filepath.Join(empty, "README.md"), []byte("# not a fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, dir := range map[string]string{
		"empty":   empty,
		"missing": filepath.Join(empty, "absent"),
	} {
		if fixtures, err := loadOpcountFixtures(dir); err == nil {
			t.Errorf("%s directory loaded %d fixtures without an error", name, len(fixtures))
		}
	}
}

// sortedOps lists a fixture's ops in order, for stable messages.
func sortedOps(counts map[string]int) []string {
	ops := make([]string, 0, len(counts))
	for op := range counts {
		ops = append(ops, op)
	}
	sort.Strings(ops)
	return ops
}

// opcountCounter counts the ops in $opcount_sub's optree the way
// opcount.t's test_opcount does, and prints each "name count".
const opcountCounter = `
my %opcount;
sub B::OP::pvm_opcount_callback {
    my ($op) = @_;
    my $name = $op->name;
    if ($name eq 'null' && $op->targ) {
        $name = "ex-" . substr(B::ppname($op->targ), 3);
    }
    $opcount{$name}++;
}
B::walkoptree(B::svref_2object($opcount_sub)->ROOT, 'pvm_opcount_callback');
print "$_ $opcount{$_}\n" for sort keys %opcount;
`

// opcountOf compiles snippet after prelude, under opcount.t's own
// "use warnings; use strict;", and returns perl's count of each op in the
// snippet sub. A snippet perl cannot compile is an error carrying perl's
// stderr, never an empty count. So is one perl warns about: opcount.t runs
// warning-free under the pinned perl, and a snippet missing its enclosing
// "use feature 'signatures'" (line 269) still compiles, as an anon sub with
// the prototype "$s = $a[0]", but warns "Illegal character in prototype".
func opcountOf(perl perltest.Perl, prelude, snippet string) (map[string]int, error) {
	script := "use warnings;\nuse strict;\nuse B ();\n" + prelude +
		"\nmy $opcount_sub = " + strings.TrimSpace(snippet) + ";\n" + opcountCounter
	cmd := perl.Command("-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil || stderr.Len() > 0 {
		return nil, fmt.Errorf("%s did not compile the snippet cleanly (%v):\n%s", perl.Path, err, stderr.String())
	}
	counts := map[string]int{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name, n, ok := strings.Cut(line, " ")
		count, convErr := strconv.Atoi(n)
		if !ok || convErr != nil {
			return nil, fmt.Errorf("unreadable count line %q from %s", line, perl.Path)
		}
		counts[name] = count
	}
	return counts, nil
}

// checkOpcountFixture compares each of fx's counts with perl's and names
// every op whose count differs.
func checkOpcountFixture(perl perltest.Perl, fx opcountFixture) error {
	got, err := opcountOf(perl, fx.Prelude, fx.Snippet)
	if err != nil {
		return fmt.Errorf("L%d %q: %w", fx.Line, fx.Name, err)
	}
	var drift []string
	for _, op := range sortedOps(fx.Counts) {
		if got[op] != fx.Counts[op] {
			drift = append(drift, fmt.Sprintf("%s: expected %d, perl counted %d", op, fx.Counts[op], got[op]))
		}
	}
	if len(drift) > 0 {
		return fmt.Errorf("L%d %q: %s", fx.Line, fx.Name, strings.Join(drift, "; "))
	}
	return nil
}

func opcountFixtureAt(t *testing.T, fixtures []opcountFixture, line int) opcountFixture {
	t.Helper()
	for _, fx := range fixtures {
		if fx.Line == line {
			return fx
		}
	}
	t.Fatalf("no fixture for opcount.t line %d", line)
	return opcountFixture{}
}

// TestOpcountFixturesHeldToPerl: the pinned perl counts every fixture's
// ops as the fixture says, one snippet at a time.
func TestOpcountFixturesHeldToPerl(t *testing.T) {
	perl := perltest.PinnedPerl(t)
	for _, fx := range mustLoadOpcountFixtures(t) {
		if err := checkOpcountFixture(perl, fx); err != nil {
			t.Error(err)
		}
	}
}

// TestOpcountFixtureCompileFailureIsAFailure: a snippet stripped of the
// feature its enclosing block in opcount.t enables fails the check with
// perl's stderr rather than being scored. Without try (line 677) perl
// 5.44.0 reports 'Global symbol "$e" requires explicit package name' and a
// syntax error; without signatures (line 269) it warns "Illegal character
// in prototype".
func TestOpcountFixtureCompileFailureIsAFailure(t *testing.T) {
	perl := perltest.PinnedPerl(t)
	fixtures := mustLoadOpcountFixtures(t)
	for _, line := range []int{679, 272} {
		fx := opcountFixtureAt(t, fixtures, line)
		fx.Prelude = ""
		err := checkOpcountFixture(perl, fx)
		if err == nil || !strings.Contains(err.Error(), "-e line") {
			t.Errorf("L%d without its prelude: want a failure quoting perl's stderr, got %v", line, err)
		}
	}
}

// TestOpcountCountsSubRootLikeOpcountT: counts come from the snippet
// sub's ROOT, as test_opcount walks it. Measured with perl 5.44.0, the
// main program holding the same two statements has enter 1, leave 1, no
// leavesub, nextstate 3 and ex-nextstate 2; the sub has one leavesub, one
// nextstate and one ex-nextstate, and the null op that was an aelem
// counts as ex-aelem.
func TestOpcountCountsSubRootLikeOpcountT(t *testing.T) {
	perl := perltest.PinnedPerl(t)
	got, err := opcountOf(perl, "", "sub { our @a; $a[0] = 1 }")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"leavesub": 1, "nextstate": 1, "ex-nextstate": 1, "leave": 0, "enter": 0, "ex-aelem": 1, "aelemfast": 1}
	for _, op := range sortedOps(want) {
		if got[op] != want[op] {
			t.Errorf("%s: counted %d, want %d (all counts: %v)", op, got[op], want[op], got)
		}
	}
}

// TestOpcountDetectsDriftedCount: a fixture with one count altered from
// perl's fails, naming the op, the fixture's count and perl's.
func TestOpcountDetectsDriftedCount(t *testing.T) {
	perl := perltest.PinnedPerl(t)
	fx := opcountFixtureAt(t, mustLoadOpcountFixtures(t), 76)
	fx.Counts = map[string]int{"aelem": 0, "aelemfast": 2, "ex-aelem": 1}
	err := checkOpcountFixture(perl, fx)
	if err == nil || !strings.Contains(err.Error(), "aelemfast: expected 2, perl counted 1") {
		t.Errorf("a drifted aelemfast count: want it named with both counts, got %v", err)
	}
}

// TestOpcountFixturesMatchTheirSourceLines: each fixture's snippet is in
// the test_opcount call at the opcount.t line it names, and the op=>count
// pairs after the snippet in that call are exactly the fixture's.
func TestOpcountFixturesMatchTheirSourceLines(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(corpusRoot(t), "t", "perf", "opcount.t"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitAfter(string(src), "\n")
	pair := regexp.MustCompile(`'?([\w-]+)'?\s*=>\s*(\d+)`)
	comment := regexp.MustCompile(`(?m)#.*$`)
	for _, fx := range mustLoadOpcountFixtures(t) {
		call, err := opcountCallAt(lines, fx.Line)
		if err != nil {
			t.Errorf("L%d %q: %v", fx.Line, fx.Name, err)
			continue
		}
		snippet := strings.TrimSuffix(fx.Snippet, "\n")
		at := strings.Index(call, snippet)
		if at < 0 {
			t.Errorf("L%d %q: the snippet is not in opcount.t's call:\n%s", fx.Line, fx.Name, call)
			continue
		}
		source := map[string]int{}
		for _, m := range pair.FindAllStringSubmatch(comment.ReplaceAllString(call[at+len(snippet):], ""), -1) {
			source[m[1]], _ = strconv.Atoi(m[2])
		}
		if fmt.Sprint(source) != fmt.Sprint(fx.Counts) {
			t.Errorf("L%d %q: opcount.t expects %v, the fixture %v", fx.Line, fx.Name, source, fx.Counts)
		}
	}
}

// opcountCallAt returns the test_opcount call that starts on 1-based line
// n, through its balanced closing parenthesis.
func opcountCallAt(lines []string, n int) (string, error) {
	if n < 1 || n > len(lines) || !strings.Contains(lines[n-1], "test_opcount(") {
		return "", fmt.Errorf("opcount.t line %d does not start a test_opcount call", n)
	}
	text := strings.Join(lines[n-1:], "")
	start := strings.Index(text, "test_opcount(") + len("test_opcount")
	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return text[start : i+1], nil
			}
		}
	}
	return "", fmt.Errorf("the test_opcount call at line %d never closes", n)
}
