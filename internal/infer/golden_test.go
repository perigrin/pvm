// ABOUTME: The golden-diagnostics ratchet: Analyze over testdata/lowered fixtures, against their committed goldens.
// ABOUTME: matching.txt lists the fixtures that match, and the test fails in both directions, as the parse ratchets do.

package infer_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

var updateLoweredRatchet = flag.Bool("infer.ratchet.update", false,
	"rewrite testdata/lowered/matching.txt with the fixtures the lowered path matches")

// The golden ratchet holds what PSC reports for each fixture.
//
// A fixture is testdata/lowered/<slice>/<name>.pl, one directory per rule
// slice. Beside it, <name>.golden holds what PSC must report for it, in
// goldenReport's format. Nothing regenerates a golden: each is committed,
// and is edited only by hand.
//
// A new fixture's golden is written by hand, from what perl does with the
// fixture. Record what now matches with
//
//	go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -infer.ratchet.update -count=1
//
// and run the ratchet without the flag to green.
//
// A golden is never edited to make a fixture match. One that is wrong is a
// PSC defect, and its edit changes what PSC reports: the commit that
// modifies it carries one trailer per changed golden, naming the cause,
//
//	Diagnostics-changed: decl/our.golden: <why the old answer was wrong>
//
// Adding a new golden needs no trailer; changing one always does.

// loweredDir holds the fixtures, their goldens and matching.txt.
const loweredDir = "testdata/lowered"

// loweredFixtures returns every fixture, as a path relative to loweredDir
// with forward slashes ("decl/my.pl"), sorted. matching.txt lists the same
// spelling.
func loweredFixtures(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(loweredDir, "*", "*.pl"))
	if err != nil {
		t.Fatalf("globbing %s: %v", loweredDir, err)
	}
	var out []string
	for _, m := range matches {
		rel, err := filepath.Rel(loweredDir, m)
		if err != nil {
			t.Fatalf("%s: %v", m, err)
		}
		out = append(out, filepath.ToSlash(rel))
	}
	sort.Strings(out)
	return out
}

// goldenPath is the golden file beside the fixture fixture names.
func goldenPath(fixture string) string {
	return filepath.Join(loweredDir, filepath.FromSlash(strings.TrimSuffix(fixture, ".pl")+".golden"))
}

// goldenReport renders what an analysis reports, in the format the goldens
// hold: a line per diagnostic, its span, code and message, and a line per
// symbol, its name and type, sorted. Annotations are not in it: they are
// keyed by node id, which a golden cannot spell, and what PSC reports is the
// diagnostics and the symbol table.
func goldenReport(diags []infer.Diagnostic, st *infer.SymbolTable) string {
	var lines []string
	for _, d := range diags {
		lines = append(lines, fmt.Sprintf("diag %d %d %s %s", d.StartByte, d.EndByte, d.Code, d.Message))
	}
	for _, s := range st.AllSymbols() {
		lines = append(lines, fmt.Sprintf("sym %s %s", s.Name, s.Type))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

// analyzeLoweredReport parses, lowers and analyses src on the lowered path
// and renders the result as goldenReport does.
func analyzeLoweredReport(src []byte) string {
	root := parse.Parse(src)
	f := lower.Lower(root)
	_, diags, st := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
	return goldenReport(diags, st)
}

// readMatching reads matching.txt: one fixture per line, blank lines and
// #-comments ignored. A missing file lists nothing.
func readMatching(t *testing.T) map[string]bool {
	t.Helper()
	listed := map[string]bool{}
	data, err := os.ReadFile(filepath.Join(loweredDir, "matching.txt"))
	if os.IsNotExist(err) {
		return listed
	}
	if err != nil {
		t.Fatalf("reading matching.txt: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			listed[line] = true
		}
	}
	return listed
}

// TestLoweredGoldenRatchet fails in both directions: a fixture listed in
// matching.txt that no longer matches its golden is a regression, and one
// that matches but is not listed must be added. Either way the list is
// what the lowered path is known to get right, and it only grows.
func TestLoweredGoldenRatchet(t *testing.T) {
	fixtures := loweredFixtures(t)
	if len(fixtures) == 0 {
		t.Fatalf("no fixtures under %s/*/*.pl", loweredDir)
	}

	matching := map[string]bool{}
	diffs := map[string]string{}
	for _, fx := range fixtures {
		src, err := os.ReadFile(filepath.Join(loweredDir, filepath.FromSlash(fx)))
		if err != nil {
			t.Fatalf("%s: %v", fx, err)
		}
		want, err := os.ReadFile(goldenPath(fx))
		if err != nil {
			t.Errorf("%s: no golden; write one beside the fixture: %v", fx, err)
			continue
		}
		got := analyzeLoweredReport(src)
		if got == string(want) {
			matching[fx] = true
		} else {
			diffs[fx] = fmt.Sprintf("--- golden\n%s--- lowered\n%s", want, got)
		}
	}

	if *updateLoweredRatchet {
		body := "# Fixtures under testdata/lowered whose lowered analysis matches its golden.\n" +
			"# Rewritten by -infer.ratchet.update; TestLoweredGoldenRatchet holds it in both directions.\n"
		for _, fx := range fixtures {
			if matching[fx] {
				body += fx + "\n"
			}
		}
		if err := os.WriteFile(filepath.Join(loweredDir, "matching.txt"), []byte(body), 0o644); err != nil {
			t.Fatalf("writing matching.txt: %v", err)
		}
		return
	}

	listed := readMatching(t)
	for _, fx := range fixtures {
		switch {
		case listed[fx] && !matching[fx]:
			t.Errorf("%s: listed in matching.txt but no longer matches its golden (a regression):\n%s", fx, diffs[fx])
		case !listed[fx] && matching[fx]:
			t.Errorf("%s: matches its golden but is not listed; record it with -infer.ratchet.update", fx)
		case !matching[fx]:
			t.Logf("%s: does not match yet", fx)
		}
		delete(listed, fx)
	}
	for fx := range listed {
		t.Errorf("%s: listed in matching.txt but no such fixture", fx)
	}
}
