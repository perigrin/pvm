// ABOUTME: Tests the precedence tables package: the mapping from the recognizer's core items to the automaton's items.
// ABOUTME: The tables are held mostly by the forest's precedence and feasibility tests; semicolon_test.go pins the ; verdicts.

package prectab

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// committed is the committed grammar, its generated tables and their Precedence.
func committed(t *testing.T) (*grammar.Grammar, *grammar.Tables, *Precedence) {
	t.Helper()
	g, ref, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref)
	if err != nil {
		t.Fatal(err)
	}
	p, err := ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	return g, tab, p
}

// TestPrectabItemMapping requires Item to map each of the recognizer's core items onto the automaton item with the
// same rule and dot, judged by the rule's symbols rather than its id, onto no automaton item twice; and the automaton
// items no core item maps to to be exactly the refused rules' items, every one of them.
func TestPrectabItemMapping(t *testing.T) {
	g, tab, p := committed(t)
	mapped := make([]bool, len(p.Items))
	for i, it := range tab.Items {
		a := int(p.Item(i))
		if a < 0 || a >= len(p.Items) {
			t.Fatalf("core item %d maps to %d, outside the automaton's %d items", i, a, len(p.Items))
		}
		at, want, got := p.Items[a], tab.Rules[it.Rule], p.Rules[p.Items[a].Rule]
		if at.Dot != it.Dot || got.LHS != want.LHS || !slices.Equal(got.RHS, want.RHS) {
			t.Errorf("core item %d (rule %d, dot %d) maps to automaton item %d (rule %d, dot %d)", i, it.Rule, it.Dot, a, at.Rule, at.Dot)
		}
		if mapped[a] {
			t.Errorf("automaton item %d is mapped twice", a)
		}
		mapped[a] = true
	}
	unmapped, rules := 0, map[int]bool{}
	for a, ok := range mapped {
		if ok {
			continue
		}
		unmapped++
		rules[p.Items[a].Rule] = true
		if p.Items[a].Rule < len(tab.Rules) {
			t.Errorf("automaton item %d, of the tables' rule %d, is mapped from no core item", a, p.Items[a].Rule)
		}
	}
	refusedItems := 0
	for r := len(tab.Rules); r < len(p.Rules); r++ {
		refusedItems += len(p.Rules[r].RHS) + 1
	}
	t.Logf("core items: %d", len(tab.Items))
	t.Logf("automaton items: %d", len(p.Items))
	t.Logf("unmapped: %d in %d refused rules", unmapped, len(rules))
	if len(tab.Items) != 1212 || len(p.Items) != 1224 || unmapped != 12 || len(rules) != 4 {
		t.Errorf("%d core items, %d automaton items, %d unmapped in %d rules; want 1212, 1224, 12 in 4", len(tab.Items), len(p.Items), unmapped, len(rules))
	}
	if len(rules) != len(g.Refused) || unmapped != refusedItems {
		t.Errorf("%d unmapped items in %d rules; the %d refused rules hold %d items", unmapped, len(rules), len(g.Refused), refusedItems)
	}
}
