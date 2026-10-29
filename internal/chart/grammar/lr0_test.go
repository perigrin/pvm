// ABOUTME: Golden tests of the LR(0) tables on small grammars (nullable, left and right recursion), and their errors.
// ABOUTME: Also checks that the cut Guacamole grammar's tables regenerate byte-identically and match the committed ones.

package grammar

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("grammar.update-golden", false, "rewrite testdata/golden/*.tables from the golden grammars")

// TestLR0Golden builds each golden grammar's tables and compares their dump
// to the golden file beside it.
func TestLR0Golden(t *testing.T) {
	for _, name := range []string{"nullable", "leftrec", "rightrec"} {
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join("testdata", "golden", name+".slif"))
			if err != nil {
				t.Fatal(err)
			}
			tab, err := BuildTables(cut(t, string(src), CutSpec{}))
			if err != nil {
				t.Fatalf("BuildTables: %v", err)
			}
			got := tab.Dump()
			path := filepath.Join("testdata", "golden", name+".tables")
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
				t.Errorf("%s: tables differ from the golden file:\n%s", path, firstDifference(path, lines([]byte(got)), lines(want)))
			}
		})
	}
}

// TestLR0NullableClosure checks the property the nullable golden file shows:
// the start state holds the items advanced past A, which is nullable through
// two levels, and its complete items reduce A and C.
func TestLR0NullableClosure(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("testdata", "golden", "nullable.slif"))
	if err != nil {
		t.Fatal(err)
	}
	tab, err := BuildTables(cut(t, string(src), CutSpec{}))
	if err != nil {
		t.Fatal(err)
	}
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

// TestTablesGuacamole regenerates the cut Guacamole grammar's tables twice:
// both runs must be byte-identical, and equal to the committed tables.
func TestTablesGuacamole(t *testing.T) {
	first, err := GenerateTables(guacamolePath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateTables(guacamolePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(second) {
		t.Fatalf("regeneration made %d files, then %d", len(first), len(second))
	}
	for i := range first {
		if first[i].Name != second[i].Name || !bytes.Equal(first[i].Data, second[i].Data) {
			t.Fatalf("regeneration differs at %s", first[i].Name)
		}
	}
	tmp := t.TempDir()
	if err := WriteTables(tmp, first); err != nil {
		t.Fatal(err)
	}
	drift, err := TablesDrift(tablesDir, tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(drift) > 0 {
		t.Errorf("committed tables in %s differ from the generator's: %s (regenerate with make chart-tables)", tablesDir, strings.Join(drift, "; "))
	}
}

const tablesDir = "testdata/tables"

// TestLR0EpsilonEdges: on the Guacamole tables, a kernel state has a
// predicted state exactly when one of its items has a nonterminal after the
// dot, and that state is a predicted one.
func TestLR0EpsilonEdges(t *testing.T) {
	tab, err := BuildTables(guacamoleCut(t))
	if err != nil {
		t.Fatal(err)
	}
	for s, st := range tab.States {
		predicts := false
		for _, i := range st.Items {
			if sym, ok := tab.after(i); ok && !tab.Symbols[sym].Terminal {
				predicts = true
			}
		}
		switch {
		case st.Predicted && st.Epsilon != -1:
			t.Errorf("predicted state %d has an ε-edge to %d", s, st.Epsilon)
		case !st.Predicted && predicts != (st.Epsilon >= 0):
			t.Errorf("kernel state %d predicts = %v, but its ε-state is %d", s, predicts, st.Epsilon)
		case st.Epsilon >= 0 && !tab.States[st.Epsilon].Predicted:
			t.Errorf("kernel state %d's ε-state %d is not a predicted state", s, st.Epsilon)
		}
	}
}
