// ABOUTME: Tests the precedence, resolve and readings stage keys: each stage alone, on fresh copies of its prepared input.
// ABOUTME: Also the refused-file rule, that the stage chain reproduces forest.ParseWith, and that the keys are never gated.

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/parse"
)

var stageKeys = []string{
	"speed-stage-precedence-aggregate-longlived", "speed-stage-resolve-aggregate-longlived",
	"speed-stage-readings-aggregate-longlived",
}

// One file each that the chart, LibraryOperators (under an empty Scope), the constraint and resolve refuse, and one
// none does.
var stageSrcs = map[string][]byte{
	"chart.pl": []byte("my $x = ;\n"), "libop.pl": []byte("1 plus 2;\n"), "prec.pl": []byte("1 <=> 2 <=> 3;\n"),
	"resolve.pl": []byte("print FH x 3;\n"), "ok.pl": []byte("my $x = 1 + 2 * 3;\n"),
}

func noScope(string) forest.Scope { return forest.Scope{} }

// forestDigest names f's content, "nil" for none.
func forestDigest(f *forest.Forest) string {
	if f == nil {
		return "nil"
	}
	return fmt.Sprintf("%x", sha256.Sum256(fmt.Append(nil, f.Root, f.Nodes, f.Alts)))
}

// libopForest is src's forest after LibraryOperators under scope, nil where the chart or LibraryOperators refuses it;
// with prec, also after the constraint, nil where it refuses.
func libopForest(t *testing.T, tab *grammar.Tables, src []byte, scope forest.Scope, prec *prectab.Precedence) *forest.Forest {
	t.Helper()
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		t.Fatal(err)
	}
	f := forest.Build(tab, c, l)
	if f == nil || forest.LibraryOperators(tab, f, s, src, scope) >= 0 || prec != nil && !f.Constrain(prec) {
		return nil
	}
	return f
}

// wrapStages has see watch each stage's every run, i its index in pipelineStages, before the stage runs.
func wrapStages(t *testing.T, see func(i int, in stageInput)) {
	t.Helper()
	old := pipelineStages
	t.Cleanup(func() { pipelineStages = old })
	pipelineStages = slices.Clone(old)
	for i := range pipelineStages {
		run := old[i].run
		pipelineStages[i].run = func(tab *grammar.Tables, p *prectab.Precedence, in stageInput) bool {
			see(i, in)
			return run(tab, p, in)
		}
	}
}

// fakeStageProtocol is a Protocol on the fake clock now, recursive descent costing 10 per file.
func fakeStageProtocol(now *time.Duration, stderr io.Writer) Protocol {
	return Protocol{Ref: func() Runner { return func(string) { *now += 10 } },
		Now: func() time.Duration { return *now }, GC: func() {}, Stderr: stderr}
}

// TestPipelineStageKeys: the three keys follow the forest's stage key in the print order; on the fake clock the
// constraint costs 1, resolve 2 and the readings count 4 per file against recursive descent's 10, so each key reads
// its own stage alone (0.1, 0.2, 0.4), printed as a number; and the precedence stage is handed the forest after
// LibraryOperators under the file's own scope, which here differs from an empty Scope's.
func TestPipelineStageKeys(t *testing.T) {
	i := slices.Index(keys, "speed-stage-forest-aggregate-longlived")
	if i < 0 || len(keys) < i+4 || !slices.Equal(keys[i+1:i+4], stageKeys) {
		t.Fatalf("keys %q do not hold %q right after speed-stage-forest-aggregate-longlived", keys, stageKeys)
	}
	tab := committedTables(t)
	t.Cleanup(parse.OverlayDeclarations(os.DirFS(filepath.Join("..", "..", "testdata", "libops"))))
	srcs := map[string][]byte{"ops.pl": []byte("use My::Ops; 1 plus 2 + 3;\n"), "ok.pl": stageSrcs["ok.pl"]}
	root := parse.ParseWithLoader(srcs["ops.pl"], func(string) ([]byte, bool) { return nil, false })
	own := forest.Scope{Imports: parse.Imports(root), LibraryOperators: parse.LibraryOperators(root)}
	scope := func(p string) forest.Scope {
		if p == "ops.pl" {
			return own
		}
		return forest.Scope{}
	}
	if libopForest(t, tab, srcs["ops.pl"], forest.Scope{}, nil) != nil {
		t.Fatal("ops.pl survives LibraryOperators under an empty Scope, so the test cannot tell the scopes apart")
	}
	want := map[string]string{}
	for p, src := range srcs {
		if want[p] = forestDigest(libopForest(t, tab, src, scope(p), nil)); want[p] == "nil" {
			t.Fatalf("%s is refused before the constraint under its own scope", p)
		}
	}
	var now time.Duration
	wrapStages(t, func(i int, in stageInput) {
		if i == 0 && forestDigest(in.f) != want[in.path] {
			t.Errorf("the precedence stage ran %s on a forest other than the one after LibraryOperators under its scope", in.path)
		}
		now += 1 << i
	})
	m := map[string]string{}
	if err := measurePipelineStages(fakeStageProtocol(&now, io.Discard), tab, []string{"ops.pl", "ok.pl"}, srcs, scope, m); err != nil {
		t.Fatal(err)
	}
	for i, k := range stageKeys {
		if wantV := fmt.Sprintf("%.4f", float64(int(1)<<i)/10); m[k] != wantV {
			t.Errorf("%s = %q, want %s: the stage alone over recursive descent", k, m[k], wantV)
		}
	}
	for _, ms := range report(m) {
		if slices.Contains(stageKeys, ms.Key) {
			if _, err := strconv.ParseFloat(ms.Value, 64); err != nil {
				t.Errorf("%s printed %q, want a number", ms.Key, ms.Value)
			}
		}
	}
}

// TestPipelineStageKeysRefusedFiles: every file stays in the timed population, and a stage whose input the chart or
// an earlier stage refused does nothing for it, counting it on stderr; resolve's input is the forest after the
// constraint (M7).
func TestPipelineStageKeysRefusedFiles(t *testing.T) {
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	files := slices.Sorted(maps.Keys(stageSrcs))
	afterConstraint := map[string]string{}
	for _, p := range []string{"resolve.pl", "ok.pl"} {
		afterConstraint[p] = forestDigest(libopForest(t, tab, stageSrcs[p], forest.Scope{}, prec))
	}
	seen := make([]map[string]bool, len(stageKeys))
	for i := range seen {
		seen[i] = map[string]bool{}
	}
	wrapStages(t, func(i int, in stageInput) {
		seen[i][in.path] = true
		if i == 1 && forestDigest(in.f) != afterConstraint[in.path] {
			t.Errorf("resolve ran %s on a forest other than the one after the constraint", in.path)
		}
	})
	var now time.Duration
	var stderr strings.Builder
	visits := 0
	p := fakeStageProtocol(&now, &stderr)
	p.Ref = func() Runner { return func(string) { visits++; now += 10 } }
	m := map[string]string{}
	if err := measurePipelineStages(p, tab, files, stageSrcs, noScope, m); err != nil {
		t.Fatal(err)
	}
	for i, want := range [][]string{{"ok.pl", "prec.pl", "resolve.pl"}, {"ok.pl", "resolve.pl"}, {"ok.pl"}} {
		if got := slices.Sorted(maps.Keys(seen[i])); !slices.Equal(got, want) {
			t.Errorf("%s ran on %q, want %q", stageKeys[i], got, want)
		}
		line := fmt.Sprintf("chart-ratchet: %s: %d of 5 files refused before the stage", stageKeys[i], 5-len(want))
		if !strings.Contains(stderr.String(), line) {
			t.Errorf("stderr %q does not report %q", stderr.String(), line)
		}
	}
	if want := 3 * 3 * (warmupPairs + timedPairs) * len(files); visits != want {
		t.Errorf("recursive descent timed %d files, want %d: every file, refused or not, in every stage's population", visits, want)
	}
}

// TestPipelineStagesSeeFreshInputEachPass: each stage is handed the same input on every one of the protocol's
// passes, so a hook that ran once, or a copy shared with the prepared input, shows as a changed digest or a missing run.
func TestPipelineStagesSeeFreshInputEachPass(t *testing.T) {
	tab := committedTables(t)
	files := []string{"ok.pl", "resolve.pl"}
	digests := make([]map[string][]string, len(stageKeys))
	for i := range digests {
		digests[i] = map[string][]string{}
	}
	wrapStages(t, func(i int, in stageInput) { digests[i][in.path] = append(digests[i][in.path], forestDigest(in.f)) })
	var now time.Duration
	if err := measurePipelineStages(fakeStageProtocol(&now, io.Discard), tab, files, stageSrcs, noScope, map[string]string{}); err != nil {
		t.Fatal(err)
	}
	passes := 3 * (warmupPairs + timedPairs)
	for i, k := range stageKeys {
		for _, p := range files {
			ds := digests[i][p]
			if p == "resolve.pl" && i == 2 {
				continue // resolve refuses it, so the readings count never sees it
			}
			if len(ds) < passes {
				t.Errorf("%s ran on %s %d times, want at least once per pass (%d)", k, p, len(ds), passes)
			}
			for _, d := range ds {
				if d != ds[0] {
					t.Errorf("%s saw %s change between passes", k, p)
					break
				}
			}
		}
	}
}

// TestPipelineStagesReproduceParseWith: over every file of the in-tree T1 corpus, LibraryOperators under the file's
// own scope, then the constraint, resolve and the readings count, run as the stage keys run them, refuse where
// forest.ParseWith refuses, for the same reason, and count the same readings where it accepts.
func TestPipelineStagesReproduceParseWith(t *testing.T) {
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join("..", "..", "..", "..", defaultT1, "*.t"))
	if err != nil || len(files) == 0 {
		t.Fatalf("the in-tree T1 corpus: %d files, %v", len(files), err)
	}
	srcs := map[string][]byte{}
	for _, p := range files {
		if srcs[p], err = os.ReadFile(p); err != nil {
			t.Fatal(err)
		}
	}
	scope := chartInputs(tab, srcs, nil, nil, ownDir).Scope
	readings := map[string]forest.Readings{}
	wrapStages(t, func(i int, in stageInput) {
		if i == 2 {
			readings[in.path] = in.f.CountReadings(tab)
		}
	})
	ins, err := stageInputs(tab, files, srcs, scope)
	if err != nil {
		t.Fatal(err)
	}
	chain := map[string]string{}
	for _, p := range files {
		if ins[p].f == nil {
			chain[p] = "before the constraint"
		}
	}
	for i, st := range pipelineStages {
		prepare, side := stageSide(tab, prec, st, ins)
		prepare()
		run := side()
		for _, p := range files {
			run(p)
		}
		advance(tab, prec, st, ins)
		for _, p := range files {
			if ins[p].f == nil && chain[p] == "" {
				chain[p] = []string{"precedence", "scope", "readings"}[i]
			}
		}
	}
	r := recognize.New(tab, recognize.Options{})
	accepted := 0
	for _, p := range files {
		f, ref := forest.ParseWith(tab, r, srcs[p], scope(p))
		switch got := chain[p]; {
		case got == "before the constraint":
			if !slices.Contains([]string{"grammar", "lex-unmapped", "scope"}, ref.Reason) {
				t.Errorf("%s: the chain refuses it before the constraint, ParseWith %q", p, ref.Reason)
			}
		case got != ref.Reason:
			t.Errorf("%s: the chain refuses it at %q, ParseWith %q", p, got, ref.Reason)
		case got == "" && readings[p] != f.Readings:
			t.Errorf("%s: the chain counts %+v readings, ParseWith %+v", p, readings[p], f.Readings)
		case got == "":
			accepted++
		}
	}
	if accepted == 0 {
		t.Fatal("ParseWith accepts no T1 file")
	}
}

// TestPipelineStageKeysAllRefused: a population refused before the constraint does not panic, and the three keys
// still print as finite numbers.
func TestPipelineStageKeysAllRefused(t *testing.T) {
	var now time.Duration
	m := map[string]string{}
	if err := measurePipelineStages(fakeStageProtocol(&now, io.Discard), committedTables(t), []string{"chart.pl", "libop.pl"},
		stageSrcs, noScope, m); err != nil {
		t.Fatal(err)
	}
	for _, k := range stageKeys {
		if v, err := strconv.ParseFloat(m[k], 64); err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("%s = %q, want a finite number", k, m[k])
		}
	}
}

// TestStageKeysNeverGated: each stage key at ten times its baseline still passes, as every speed- key does.
func TestStageKeysNeverGated(t *testing.T) {
	for _, k := range stageKeys {
		if got := compareOne(k, "0.5000", "5.0000", ""); len(got) > 0 {
			t.Errorf("%s 0.5 -> 5.0: %q, want no regression", k, got)
		}
	}
}

// TestPipelineStagePrecedenceTimesConstrain: the precedence stage key times production Constrain, not the reference.
// On a variant of the reference's known-divergence toy (forest's TestPrecedenceReferenceKnownDivergence, without
// its Z ::= HIGH rule: S: NUM A; A: E PLUS NUM; E: %empty %prec HIGH, PLUS below HIGH) Constrain accepts
// NUM PLUS NUM, as bison 3.8.2 does, and ConstrainReference refuses it, so the stage's verdict says which one ran.
func TestPipelineStagePrecedenceTimesConstrain(t *testing.T) {
	g := &grammar.Grammar{}
	for _, r := range [][]string{{"S", "NUM", "A"}, {"A", "E", "PLUS", "NUM"}, {"E"}} {
		rule := grammar.Rule{LHS: r[0]}
		for _, w := range r[1:] {
			rule.RHS = append(rule.RHS, grammar.Symbol{Kind: grammar.Name, Text: w})
		}
		g.G1 = append(g.G1, rule)
	}
	g.G0 = []grammar.Rule{{LHS: "NUM"}, {LHS: "PLUS"}}
	tab, err := grammar.BuildTables(g)
	if err != nil {
		t.Fatal(err)
	}
	tab.TerminalSets = [][]int{{}}
	l := &lex.Lattice[lex.None]{Nodes: make([]lex.Node[lex.None], 4)}
	for i, w := range []string{"NUM", "PLUS", "NUM"} {
		tab.TerminalSets = append(tab.TerminalSets, []int{slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Name == w })})
		l.Edges = append(l.Edges, lex.Edge{From: i, To: i + 1, Token: lex.Token{TerminalSet: len(tab.TerminalSets) - 1}})
	}
	prec, err := prectab.NewPrecedence(tab, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"PLUS"}}, {N: 2, Assoc: "left", Tokens: []string{"HIGH"}}},
		[]string{"", "", "HIGH"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil || !c.Accepted() {
		t.Fatalf("the chart refuses NUM PLUS NUM: %v", err)
	}
	f := forest.Build(tab, c, l)
	if f.Copy().ConstrainReference(prec) || !f.Copy().Constrain(prec) {
		t.Fatal("Constrain and ConstrainReference no longer diverge on the toy, so it cannot tell them apart")
	}
	st := pipelineStages[0]
	if st.key != "speed-stage-precedence-aggregate-longlived" {
		t.Fatalf("pipelineStages[0] is %s", st.key)
	}
	if !st.run(tab, prec, stageInput{f: f.Copy()}) {
		t.Error("the precedence stage refuses NUM PLUS NUM, as ConstrainReference does; want Constrain's accept")
	}
}
