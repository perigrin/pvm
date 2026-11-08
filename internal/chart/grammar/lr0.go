// ABOUTME: Builds the LR(0) automaton with nullable closure over a grammar, with integer core items.
// ABOUTME: Each state carries its transition list (per symbol) and its reduce list.

package grammar

import (
	"fmt"
	"slices"
)

// TableSymbol is a grammar symbol in the tables: a terminal or a nonterminal,
// and whether it derives the empty string.
type TableSymbol struct {
	Name     string
	Terminal bool
	Nullable bool
}

// BNFRule is a rule with sequences expanded: symbol ids, an empty RHS for an
// empty alternative.
type BNFRule struct {
	LHS int
	RHS []int
}

// CoreItem is a rule and a dot position; its integer id is its index in
// Tables.Items.
type CoreItem struct {
	Rule, Dot int
}

// Transition is a state's move on one symbol.
type Transition struct {
	Symbol, Target int
}

// State is one state of the split LR(0) automaton (Aycock and Horspool's
// ε-DFA). A kernel state holds the items a transition reached, with their
// dots advanced past nullable symbols: their origin is wherever the
// transition's source items began. A predicted state holds the items the
// kernel predicts, closed the same way: their origin is the current position.
// Epsilon is a kernel state's predicted state, -1 if it predicts nothing.
// Expected is the terminals after a dot in Items, in symbol order; Leo is
// the items eligible for Leo's memo, at most one per awaited symbol (leo.go).
type State struct {
	Items       []int
	Predicted   bool
	Epsilon     int
	Transitions []Transition
	Reduces     []int
	Expected    []int
	Leo         []int
}

// Tables are the generator's output.
type Tables struct {
	Start   int
	Symbols []TableSymbol
	Rules   []BNFRule
	Items   []CoreItem
	States  []State
	// First and Follow are per symbol, indexed by symbol id: the sorted ids of
	// the terminals that can begin it, and that can follow it. EOF, one past
	// the last symbol, is the end of input and appears only in Follow.
	First, Follow [][]int
	EOF           int
	// TerminalSets is the terminal-set table a token's id indexes, each entry
	// the sorted ids of the terminals that admit the token; id 0 is the empty
	// set. BuildTables leaves it empty; the generator fills it from the
	// grammar's token classes, complete before recognize.New.
	TerminalSets [][]int
	// Tokens is the compiled token table whose ids index TerminalSets; nil
	// until the generator compiles the grammar's classes.
	Tokens *TokenTable
}

// BuildLR0 builds the tables for rules, whose terminals are named, starting
// from start. Symbol ids are the nonterminals in order of their first rule,
// then the terminals in name order; core item ids run rule by rule, dot by
// dot; states are numbered breadth first with transitions in symbol order.
// So the tables depend only on the rules' order, never on map iteration.
func BuildLR0(start string, rules []Rule, terminals []string) (*Tables, error) {
	t := &Tables{}
	id := map[string]int{}
	for _, r := range rules {
		if _, ok := id[r.LHS]; !ok {
			id[r.LHS] = len(t.Symbols)
			t.Symbols = append(t.Symbols, TableSymbol{Name: r.LHS})
		}
	}
	terms := slices.Clone(terminals)
	slices.Sort(terms)
	for _, name := range slices.Compact(terms) {
		if _, ok := id[name]; ok {
			return nil, fmt.Errorf("lr0: %s is both a terminal and a rule's left-hand side", name)
		}
		id[name] = len(t.Symbols)
		t.Symbols = append(t.Symbols, TableSymbol{Name: name, Terminal: true})
	}
	s, ok := id[start]
	if !ok || t.Symbols[s].Terminal {
		return nil, fmt.Errorf("lr0: no rule for the start symbol %s", start)
	}
	t.Start = s

	for _, r := range rules {
		var rhs []int
		for _, sym := range r.RHS {
			n, ok := id[sym.Text]
			if sym.Kind != Name || !ok {
				return nil, fmt.Errorf("lr0: %s (line %d) names %s, which is neither a terminal nor a rule's left-hand side", ruleText(r, "::="), r.Line, symbolText(sym))
			}
			rhs = append(rhs, n)
		}
		lhs := id[r.LHS]
		if r.Seq != NotSeq && len(rhs) != 1 {
			return nil, fmt.Errorf("lr0: %s (line %d) is a sequence rule, whose right-hand side must be one symbol", ruleText(r, "::="), r.Line)
		}
		switch r.Seq {
		case NotSeq:
			t.Rules = append(t.Rules, BNFRule{lhs, rhs})
		case Star: // X ::= Y* is X ::= (empty) | X Y
			t.Rules = append(t.Rules, BNFRule{lhs, nil}, BNFRule{lhs, []int{lhs, rhs[0]}})
		case Plus: // X ::= Y+ is X ::= Y | X Y
			t.Rules = append(t.Rules, BNFRule{lhs, rhs}, BNFRule{lhs, []int{lhs, rhs[0]}})
		}
	}

	t.nullable()
	itemOf := make([]int, len(t.Rules)) // the id of rule r at dot 0
	for r, rule := range t.Rules {
		itemOf[r] = len(t.Items)
		for dot := 0; dot <= len(rule.RHS); dot++ {
			t.Items = append(t.Items, CoreItem{r, dot})
		}
	}
	t.automaton(itemOf)
	t.expected()
	t.leo()
	t.firstFollow()
	return t, nil
}

// nullable marks every nonterminal that derives the empty string, by
// iterating to a fixed point, which terminates on cycles such as A ::= A.
func (t *Tables) nullable() {
	for changed := true; changed; {
		changed = false
		for _, r := range t.Rules {
			if t.Symbols[r.LHS].Nullable {
				continue
			}
			if !slices.ContainsFunc(r.RHS, func(s int) bool { return !t.Symbols[s].Nullable }) {
				t.Symbols[r.LHS].Nullable = true
				changed = true
			}
		}
	}
}

// automaton builds the split automaton: the start symbol's predicted state,
// then every state's goto on each symbol after a dot, which is always a
// kernel state, and each kernel state's predicted state.
func (t *Tables) automaton(itemOf []int) {
	byLHS := map[int][]int{}
	for r, rule := range t.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
	}
	in := make([]bool, len(t.Items))
	// advance closes seed under Aycock and Horspool's dot advancement: an item
	// with its dot before a nullable symbol brings the item with its dot past
	// it, with the same origin. It returns the set, sorted, and the
	// nonterminals it predicts, in the order first met.
	advance := func(seed []int) (items, predicts []int) {
		work := slices.Clone(seed)
		seen := map[int]bool{}
		for _, i := range seed {
			in[i] = true
		}
		for len(work) > 0 {
			i := work[len(work)-1]
			work = work[:len(work)-1]
			items = append(items, i)
			sym, ok := t.after(i)
			if !ok || t.Symbols[sym].Terminal {
				continue
			}
			if !seen[sym] {
				seen[sym] = true
				predicts = append(predicts, sym)
			}
			if t.Symbols[sym].Nullable && !in[i+1] {
				in[i+1] = true
				work = append(work, i+1)
			}
		}
		for _, i := range items {
			in[i] = false
		}
		slices.Sort(items)
		return items, predicts
	}
	// predict is the closed set of items the nonterminals predict: their rules
	// at dot 0, advanced, and whatever those predict in turn.
	predict := func(nts []int) []int {
		var all []int
		done := map[int]bool{}
		for len(nts) > 0 {
			nt := nts[0]
			nts = nts[1:]
			if done[nt] {
				continue
			}
			done[nt] = true
			var seed []int
			for _, r := range byLHS[nt] {
				seed = append(seed, itemOf[r])
			}
			items, more := advance(seed)
			all = append(all, items...)
			nts = append(nts, more...)
		}
		slices.Sort(all)
		return slices.Compact(all)
	}

	index := map[string]int{}
	state := func(items []int, predicted bool) (int, bool) {
		key := fmt.Sprint(predicted, items)
		if s, ok := index[key]; ok {
			return s, false
		}
		index[key] = len(t.States)
		t.States = append(t.States, State{Items: items, Predicted: predicted, Epsilon: -1})
		return len(t.States) - 1, true
	}
	kernel := func(seed []int) int {
		items, predicts := advance(seed)
		s, isNew := state(items, false)
		if isNew && len(predicts) > 0 {
			// Not t.States[s].Epsilon, _ = state(...): Go leaves unspecified
			// whether the slice is read before state appends to it, and an
			// append may move it.
			eps, _ := state(predict(predicts), true)
			t.States[s].Epsilon = eps
		}
		return s
	}
	// The start state is the start symbol's predictions: every item in it
	// begins at position 0, which is the current position.
	state(predict([]int{t.Start}), true)

	for s := 0; s < len(t.States); s++ {
		bySym := map[int][]int{}
		var reduces []int
		for _, i := range t.States[s].Items {
			if sym, ok := t.after(i); ok {
				bySym[sym] = append(bySym[sym], i)
			} else {
				reduces = append(reduces, t.Items[i].Rule)
			}
		}
		slices.Sort(reduces)
		t.States[s].Reduces = slices.Compact(reduces)
		syms := make([]int, 0, len(bySym))
		for sym := range bySym {
			syms = append(syms, sym)
		}
		slices.Sort(syms)
		for _, sym := range syms {
			items := bySym[sym]
			seed := make([]int, len(items))
			for k, i := range items {
				seed[k] = i + 1
			}
			target := kernel(seed)
			t.States[s].Transitions = append(t.States[s].Transitions, Transition{sym, target})
		}
	}
}

// after is the symbol after item i's dot, if the item is not complete.
func (t *Tables) after(i int) (int, bool) {
	it := t.Items[i]
	rhs := t.Rules[it.Rule].RHS
	if it.Dot == len(rhs) {
		return 0, false
	}
	return rhs[it.Dot], true
}
