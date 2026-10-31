// ABOUTME: Golden tests of the LR(0) tables on small grammars (nullable, left and right recursion), and their errors.
// ABOUTME: Also checks that a kernel state's ε-edges, predicted states and Leo eligibility are consistent.

package grammar

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("grammar.update-golden", false, "rewrite testdata/golden/*.tables from the golden grammars")

func name(s string) Symbol { return Symbol{Name, s} }

// alts is the rules of lhs, one per alternative, each a list of symbol names;
// an empty alternative is the empty string.
func alts(lhs string, alternatives ...string) []Rule {
	var out []Rule
	for _, a := range alternatives {
		r := Rule{LHS: lhs}
		for _, s := range strings.Fields(a) {
			r.RHS = append(r.RHS, name(s))
		}
		out = append(out, r)
	}
	return out
}

// seq is lhs ::= sym* or sym+.
func seq(lhs, sym string, kind Seq) Rule {
	return Rule{LHS: lhs, RHS: []Symbol{name(sym)}, Seq: kind}
}

// toy states a grammar as Go data: the terminals, then G1 rules in order.
func toy(terminals []string, g1 ...[]Rule) *Grammar {
	g := &Grammar{}
	for _, rs := range g1 {
		g.G1 = append(g.G1, rs...)
	}
	for _, term := range terminals {
		g.G0 = append(g.G0, Rule{LHS: term})
	}
	return g
}

func build(t *testing.T, g *Grammar) *Tables {
	t.Helper()
	tab, err := BuildTables(g)
	if err != nil {
		t.Fatalf("BuildTables: %v", err)
	}
	return tab
}

// leftrecToy is E ::= E P T | <second>.
func leftrecToy(second string) *Grammar {
	return toy([]string{"P", "T"}, alts("E", "E P T", second))
}

// nullableToy: A is nullable through two levels, through C, and C through D*.
var nullableToy = toy([]string{"D", "E", "X"},
	alts("S", "A X B"), alts("A", "C"), []Rule{seq("C", "D", Star), seq("B", "E", Plus)})

// goldenToys are the grammars behind testdata/golden/*.tables.
var goldenToys = []struct {
	name string
	g    *Grammar
}{
	{"leftrec", leftrecToy("T")},
	{"rightrec", toy([]string{"A"}, alts("S", "A S", "A"))},
	{"nullable", nullableToy},
}

func goldenPath(name string) string { return filepath.Join("testdata", "golden", name+".tables") }

// TestGoldenToyTables builds each toy's tables and compares their dump to the
// golden file beside it.
func TestGoldenToyTables(t *testing.T) {
	for _, toy := range goldenToys {
		t.Run(toy.name, func(t *testing.T) {
			got := build(t, toy.g).Dump()
			path := goldenPath(toy.name)
			if *updateGolden {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (regenerate with -grammar.update-golden)", err)
			}
			if got != string(want) {
				t.Errorf("%s: tables differ from the golden file:\n%s", path, firstDifference(path, "the generator", lines([]byte(got)), lines(want)))
			}
		})
	}
}

// TestGoldenToyTablesCatchesMutation: one rule changed in the leftrec toy's
// Go data gives tables that differ from its golden, and the unchanged toy
// still equals it.
func TestGoldenToyTablesCatchesMutation(t *testing.T) {
	want, err := os.ReadFile(goldenPath("leftrec"))
	if err != nil {
		t.Fatal(err)
	}
	if got := build(t, leftrecToy("T")).Dump(); got != string(want) {
		t.Fatalf("the unchanged toy differs from its golden:\n%s", firstDifference("leftrec.tables", "the generator", lines([]byte(got)), lines(want)))
	}
	if got := build(t, leftrecToy("P")).Dump(); got == string(want) {
		t.Error("the mutated toy generates its golden's tables")
	}
}

// TestBuildTablesRefusesEmptyGrammar: no rules and no start rule are errors
// naming what is missing, and a real toy still builds.
func TestBuildTablesRefusesEmptyGrammar(t *testing.T) {
	for _, g := range []*Grammar{nil, {}, {G0: []Rule{{LHS: "A"}}}} {
		tab, err := BuildTables(g)
		if err == nil || !strings.Contains(err.Error(), "start rule") || tab != nil {
			t.Errorf("BuildTables(%+v) = %v, %v; want an error naming the missing start rule", g, tab, err)
		}
	}
	build(t, leftrecToy("T"))
}

// TestLR0NullableClosure checks the property the nullable golden file shows:
// the start state holds the items advanced past A, which is nullable through
// two levels, and its complete items reduce A and C.
func TestLR0NullableClosure(t *testing.T) {
	tab := build(t, nullableToy)
	start := tab.States[0]
	var items []string
	for _, id := range start.Items {
		items = append(items, tab.itemText(id))
	}
	for _, want := range []string{"S ::= . A X B", "S ::= A . X B", "A ::= . C", "A ::= C .", "C ::= .", "C ::= . C D"} {
		if !strings.Contains(strings.Join(items, "\n")+"\n", want+"\n") {
			t.Errorf("start state lacks %q; has:\n%s", want, strings.Join(items, "\n"))
		}
	}
	var reduces []string
	for _, r := range start.Reduces {
		reduces = append(reduces, tab.Symbols[tab.Rules[r].LHS].Name)
	}
	if got := strings.Join(reduces, " "); got != "A C" {
		t.Errorf("start state reduces %q, want \"A C\"", got)
	}
	for _, s := range tab.Symbols {
		if want := s.Name == "A" || s.Name == "C"; s.Nullable != want {
			t.Errorf("%s nullable = %v, want %v", s.Name, s.Nullable, want)
		}
	}
}

func TestLR0UndefinedSymbol(t *testing.T) {
	_, err := BuildLR0("S", []Rule{{LHS: "S", RHS: []Symbol{name("Nope"), name("a")}}}, []string{"a"})
	if err == nil || !strings.Contains(err.Error(), "Nope") {
		t.Fatalf("BuildLR0 error = %v, want one naming Nope", err)
	}
	if _, err := BuildLR0("Missing", []Rule{{LHS: "S", RHS: []Symbol{name("a")}}}, []string{"a"}); err == nil || !strings.Contains(err.Error(), "Missing") {
		t.Fatalf("BuildLR0 with no rule for the start symbol: error = %v, want one naming Missing", err)
	}
}

// TestLR0CyclicNullable: A ::= A with an empty alternative must terminate,
// with A and B nullable and S not. The reader has no empty rules, so the
// grammar is built directly.
func TestLR0CyclicNullable(t *testing.T) {
	rules := []Rule{
		{LHS: "S", RHS: []Symbol{name("B"), name("c")}},
		{LHS: "A", RHS: []Symbol{name("A")}},
		{LHS: "A"},
		{LHS: "B", RHS: []Symbol{name("A"), name("A")}},
	}
	tab, err := BuildLR0("S", rules, []string{"c"})
	if err != nil {
		t.Fatal(err)
	}
	nullable := map[string]bool{}
	for _, s := range tab.Symbols {
		nullable[s.Name] = s.Nullable
	}
	if want := map[string]bool{"S": false, "A": true, "B": true, "c": false}; !mapsEqual(nullable, want) {
		t.Errorf("nullable = %v, want %v", nullable, want)
	}
}

func mapsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// TestLR0EpsilonEdges: in each golden toy's tables, a
// kernel state has a predicted state exactly when one of its items has a
// nonterminal after the dot, and that state is a predicted one.
func TestLR0EpsilonEdges(t *testing.T) {
	tabs := map[string]*Tables{}
	for _, toy := range goldenToys {
		tabs[toy.name] = build(t, toy.g)
	}
	for name, tab := range tabs {
		for s, st := range tab.States {
			predicts := false
			for _, i := range st.Items {
				if sym, ok := tab.after(i); ok && !tab.Symbols[sym].Terminal {
					predicts = true
				}
			}
			switch {
			case st.Predicted && st.Epsilon != -1:
				t.Errorf("%s: predicted state %d has an ε-edge to %d", name, s, st.Epsilon)
			case !st.Predicted && predicts != (st.Epsilon >= 0):
				t.Errorf("%s: kernel state %d predicts = %v, but its ε-state is %d", name, s, predicts, st.Epsilon)
			case st.Epsilon >= 0 && !tab.States[st.Epsilon].Predicted:
				t.Errorf("%s: kernel state %d's ε-state %d is not a predicted state", name, s, st.Epsilon)
			}
		}
	}
}

// TestBuildTablesTakesGeneratorInput: tables build from a grammar stated as
// input.go types, and no probe or conflict file is written.
func TestBuildTablesTakesGeneratorInput(t *testing.T) {
	g := &Grammar{
		G1: []Rule{{LHS: "S", RHS: []Symbol{name("A"), name("S")}}, {LHS: "S", RHS: []Symbol{name("A")}}},
		G0: []Rule{{LHS: "A", RHS: []Symbol{{Literal, "a"}}}},
	}
	tab, err := BuildTables(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(tab.Files()) == 0 {
		t.Fatal("no files")
	}
	for _, f := range tab.Files() {
		if f.Name == "probes.txt" || f.Name == "conflicts.txt" {
			t.Errorf("Files names %s", f.Name)
		}
	}
}

// firstDifference describes where the listed lines first differ from the
// lines maker makes. A line is quoted as the list writes it, unescaped, so
// grep -F finds it.
func firstDifference(path, maker string, made, listed []string) string {
	for i := 0; i < len(made) || i < len(listed); i++ {
		switch {
		case i >= len(made):
			return fmt.Sprintf(`%s: line %d lists "%s", which %s no longer makes`, path, i+1, listed[i], maker)
		case i >= len(listed):
			return fmt.Sprintf(`%s: %s makes "%s", which is not listed (after line %d)`, path, maker, made[i], i)
		case made[i] != listed[i]:
			return fmt.Sprintf(`%s: line %d lists "%s", but %s makes "%s"`, path, i+1, listed[i], maker, made[i])
		}
	}
	return path + ": the lines match, but the bytes differ: the list must end in exactly one newline"
}

func lines(b []byte) []string { return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") }

func TestFirstDifference(t *testing.T) {
	for _, tc := range []struct {
		made, listed []string
		want         string
	}{
		{[]string{"a", `b '\'`}, []string{"a"}, `p: the generator makes "b '\'", which is not listed (after line 1)`},
		{[]string{"a"}, []string{"a", "Fake"}, `p: line 2 lists "Fake", which the generator no longer makes`},
		{[]string{`x '\'`}, []string{"y"}, `p: line 1 lists "y", but the generator makes "x '\'"`},
		{[]string{"a"}, []string{"a"}, "p: the lines match, but the bytes differ: the list must end in exactly one newline"},
	} {
		if got := firstDifference("p", "the generator", tc.made, tc.listed); got != tc.want {
			t.Errorf("firstDifference(%q, %q) = %s, want %s", tc.made, tc.listed, got, tc.want)
		}
	}
}

// TestWriteTablesAndDrift: WriteTables leaves exactly the generator's files,
// removing a stray .txt, and TablesDrift reports nothing for equal
// directories and names the one file that changed.
func TestWriteTablesAndDrift(t *testing.T) {
	files := build(t, goldenToys[0].g).Files()
	a, b := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(b, "stray.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{a, b} {
		if err := WriteTables(dir, files); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(b, "stray.txt")); !os.IsNotExist(err) {
		t.Errorf("WriteTables left the stray file (stat error %v)", err)
	}
	if drift, err := TablesDrift(a, b); err != nil || len(drift) != 0 {
		t.Fatalf("TablesDrift of equal directories = %v, %v; want none", drift, err)
	}
	if err := os.WriteFile(filepath.Join(b, "rules.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	drift, err := TablesDrift(a, b)
	if want := "rules.txt differs from the generator's"; err != nil || len(drift) != 1 || drift[0] != want {
		t.Errorf("TablesDrift = %q, %v; want [%q]", drift, err, want)
	}
}
