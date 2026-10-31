// ABOUTME: Tests the per-state expected-terminal sets and Leo eligibility.
// ABOUTME: Small grammars pin each definition.

package grammar

import (
	"slices"
	"strings"
	"testing"
)

// stateNames is each state's expected terminals, by name, in id order.
func stateNames(tab *Tables) []string {
	var out []string
	for _, s := range tab.States {
		var names []string
		for _, id := range s.Expected {
			names = append(names, tab.Symbols[id].Name)
		}
		out = append(out, strings.Join(names, " "))
	}
	return out
}

// TestExpectedTerminals: a state expects exactly the terminals after a dot
// in its items. In S ::= A S | B, state 0 predicts both alternatives, the
// kernel after A predicts S again through state 0, and complete states
// expect nothing.
func TestExpectedTerminals(t *testing.T) {
	tab := build(t, toy([]string{"A", "B"}, alts("S", "A S", "B")))
	want := []string{"A B", "", "", ""}
	if got := stateNames(tab); !slices.Equal(got, want) {
		t.Errorf("expected terminals per state = %q, want %q\n%s", got, want, tab.Dump())
	}
}

// leoStates is the item text of each Leo-eligible item.
func leoStates(tab *Tables) []string {
	var out []string
	for _, s := range tab.States {
		for _, it := range s.Leo {
			out = append(out, tab.itemText(it))
		}
	}
	return out
}

// TestLeoEligibility: in S ::= A S | B the kernel state after A holds the one
// item S ::= A . S, a right-recursive completion candidate.
func TestLeoEligibility(t *testing.T) {
	tab := build(t, toy([]string{"A", "B"}, alts("S", "A S", "B")))
	if got, want := leoStates(tab), []string{"S ::= A . S"}; !slices.Equal(got, want) {
		t.Errorf("Leo-eligible items = %q, want %q", got, want)
	}
}

// TestLeoEligibilityPerSymbol: eligibility is per (state, symbol), so the
// other items in a state do not matter. In S ::= A S | A the kernel after A
// holds S ::= A . S beside S ::= A ., and S ::= A . S is eligible. A right
// edge reaches past trailing nullables (2e94's gate note F2), but an item
// with anything after B is never eligible: its β could still read a token.
func TestLeoEligibilityPerSymbol(t *testing.T) {
	for _, c := range []struct {
		g    *Grammar
		want []string
	}{
		{toy([]string{"A"}, alts("S", "A S", "A")), []string{"S ::= A . S"}},
		{toy([]string{"A", "C"}, alts("S", "A T", "A"), alts("T", "S O"), []Rule{seq("O", "C", Star)}), []string{"S ::= A . T"}},
		{toy([]string{"A", "C"}, alts("S", "A S O", "A"), []Rule{seq("O", "C", Star)}), nil},
	} {
		tab := build(t, c.g)
		if got := leoStates(tab); !slices.Equal(got, c.want) {
			t.Errorf("Leo-eligible items = %q, want %q\n%s", got, c.want, tab.Dump())
		}
	}
}

// TestLeoEligibilityRefusals: a left-recursive rule's single item L ::= L . X
// has a nonterminal last, but X never derives L at its right edge; and a
// state with two items awaiting S, S ::= A . S and S ::= A . S B, has no
// Leo item for S.
func TestLeoEligibilityRefusals(t *testing.T) {
	for _, g := range []*Grammar{
		toy([]string{"C", "Y"}, alts("L", "L X", "Y"), alts("X", "C")),
		toy([]string{"A", "B"}, alts("S", "A S", "A S B", "A")),
	} {
		tab := build(t, g)
		if got := leoStates(tab); len(got) > 0 {
			t.Errorf("Leo-eligible items = %q, want none\n%s", got, tab.Dump())
		}
	}
}
