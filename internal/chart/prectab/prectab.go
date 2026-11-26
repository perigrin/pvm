// ABOUTME: A grammar's precedence tables (spec §8): bison's LR(0) states and LALR(1) lookaheads, and the state-free
// ABOUTME: tables generated from them. It imports only grammar, so the recognizer and the forest can both read them.

package prectab

import (
	"fmt"
	"maps"
	"math/bits"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/chart/grammar"
)

// Precedence is a grammar's precedence declarations, bison's way: each terminal's level, each rule's (its %prec
// token's, else its last terminal's), each level's associativity, and bison's LR(0) parser states, read from
// the grammar alone.
type Precedence struct {
	Tab       *grammar.Tables
	TermLevel []int              // by symbol: a terminal's level, 0 for none
	RuleLevel []int              // by automaton rule: its level, 0 for none
	Assoc     []string           // by level: left, right or nonassoc
	States    []State            // state 0 is the file's start
	Rules     []grammar.BNFRule  // the automaton's rules: the tables', then the refused ones'
	Items     []grammar.CoreItem // the automaton's items, numbered as Rules are: Base[r]+dot

	// The state-free tables (decision §3.1), generated with the automaton and read by index, never through a map.
	// Rules and items are the automaton's: the tables', then the refused ones'.
	NT           int        // terminals, EOF included
	Base         []int32    // by rule: its item at dot 0
	ReduceAt     []uint8    // [r*NT+t]: 0 every state reducing r reduces it with t next, 1 none does, 2 the parent item decides
	MixAt        []int32    // [r*NT+t]: a 2's row of ParentForbid, -1 otherwise
	ParentForbid []bool     // [row*items+parent item]: the reduction is forbidden under that parent item
	NeedParent   []bool     // by symbol: some rule of it has a 2
	TopAt        []int32    // by item: its row of ShiftAt as a spine-top item, -1 where it forbids no shift
	ShiftAt      []int32    // [row*items+rule-position item]: its set in ShiftForbid, -1 for none
	ShiftForbid  [][]uint64 // terminal bitsets: the shifts forbidden after that item under that spine top
}

// State is an LR(0) state: the terminals it shifts, sorted, the rules it reduces with each one's LALR(1)
// lookahead (LA[i] the sorted terminals Reduce[i] reduces on), and its goto on each symbol after a dot.
type State struct {
	Shift, Reduce []int32
	LA            [][]int32
	Next          map[int32]int32
}

// NewPrecedence is levels, loosest first, over tab, whose rule r carries the %prec token rulePrec[r] ("" for
// none). A token that is no terminal of tab, as a %prec-only token is, still gives its rules their level.
// refused are rules over tab's symbols that refuse in their action: no forest derives them, but bison's states
// hold their items, so they are in the automaton (perly.y's termeqop: termeqop NCEQOP shifts the second <=>).
//
// The states are bison's: the canonical LR(0) automaton of the grammar, closure and goto, which bison's LALR(1)
// tables share, so a conflict is found exactly where bison's state has one. Each reduction's lookahead there is
// bison's LALR(1) set, not the follow set. Both are built from the tables alone, never read from the chart,
// whose split states are not bison's. From them it generates the state-free tables Constrain reads; only
// ConstrainReference carries a state, the one a derivation reaches from the file's start along the symbols before
// it. A refused rule naming a symbol tab lacks is an error, and so is a key whose decision bison's states still
// disagree on with its whole key (decision §3.2).
func NewPrecedence(tab *grammar.Tables, levels []grammar.Level, rulePrec []string, refused []grammar.Rule) (*Precedence, error) {
	p := &Precedence{Tab: tab, TermLevel: make([]int, len(tab.Symbols)), Assoc: []string{""}}
	sym := map[string]int{}
	for id, s := range tab.Symbols {
		sym[s.Name] = id
	}
	rules, items, rulePrec := slices.Clone(tab.Rules), slices.Clone(tab.Items), slices.Clone(rulePrec)
	base := make([]int, len(rules)) // base[r] is rule r's item at dot 0
	for r := 1; r < len(rules); r++ {
		base[r] = base[r-1] + len(rules[r-1].RHS) + 1
	}
	for _, r := range refused {
		var rhs []int
		for _, s := range append([]grammar.Symbol{{Text: r.LHS}}, r.RHS...) {
			id, ok := sym[s.Text]
			if !ok {
				return nil, fmt.Errorf("prectab: the refused rule %s: names %s, which is no symbol of the tables", r.LHS, s.Text)
			}
			rhs = append(rhs, id)
		}
		lhs, rhs := rhs[0], rhs[1:]
		rulePrec = append(rulePrec, r.Prec)
		base = append(base, len(items))
		for dot := 0; dot <= len(rhs); dot++ {
			items = append(items, grammar.CoreItem{Rule: len(rules), Dot: dot})
		}
		rules = append(rules, grammar.BNFRule{LHS: lhs, RHS: rhs})
	}
	// shortcut: tab's FIRST sets and nullability leave the refused rules out, exact while each begins with its own
	// left-hand side, which derives no empty string, as perly.y's four do (TestPrecedenceRefusedRulesLeftRecursive);
	// derive them over rules if one does not.
	p.RuleLevel = make([]int, len(rules))
	level := map[string]int{}
	for i, l := range levels {
		p.Assoc = append(p.Assoc, l.Assoc)
		for _, tok := range l.Tokens {
			level[tok] = i + 1
		}
	}
	for id, s := range tab.Symbols {
		if s.Terminal {
			p.TermLevel[id] = level[s.Name]
		}
	}
	for r, rule := range rules {
		if rulePrec[r] != "" {
			p.RuleLevel[r] = level[rulePrec[r]]
			continue
		}
		for _, s := range slices.Backward(rule.RHS) {
			if tab.Symbols[s].Terminal {
				p.RuleLevel[r] = p.TermLevel[s]
				break
			}
		}
	}
	byLHS := make([][]int, len(tab.Symbols))
	var start []int
	for r, rule := range rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
		if rule.LHS == tab.Start {
			start = append(start, base[r])
		}
	}
	closure := func(set []int) []int {
		set = slices.Clone(set)
		in := map[int]bool{}
		for i := 0; i < len(set); i++ {
			if in[set[i]] {
				continue
			}
			in[set[i]] = true
			it := items[set[i]]
			if rhs := rules[it.Rule].RHS; it.Dot < len(rhs) {
				for _, r := range byLHS[rhs[it.Dot]] {
					set = append(set, base[r])
				}
			}
		}
		return slices.Sorted(maps.Keys(in))
	}
	id := map[string]int32{} // each state's items, as text, to its index
	var states [][]int
	intern := func(set []int) int32 {
		k := fmt.Sprint(set)
		if i, ok := id[k]; ok {
			return i
		}
		id[k] = int32(len(states))
		states = append(states, set)
		return id[k]
	}
	intern(closure(start))
	for s := 0; s < len(states); s++ { // breadth first; intern appends the states it finds
		st := State{Next: map[int32]int32{}}
		kernels := map[int][]int{} // each symbol after a dot, and the items it advances to
		for _, i := range states[s] {
			it := items[i]
			rhs := rules[it.Rule].RHS
			switch {
			case it.Dot == len(rhs):
				st.Reduce = append(st.Reduce, int32(it.Rule))
			case tab.Symbols[rhs[it.Dot]].Terminal:
				st.Shift = append(st.Shift, int32(rhs[it.Dot]))
			}
			if it.Dot < len(rhs) {
				kernels[rhs[it.Dot]] = append(kernels[rhs[it.Dot]], i+1)
			}
		}
		slices.Sort(st.Shift)
		st.Shift = slices.Compact(st.Shift)
		for _, x := range slices.Sorted(maps.Keys(kernels)) {
			st.Next[int32(x)] = intern(closure(kernels[x]))
		}
		p.States = append(p.States, st)
	}
	p.Rules, p.Items = rules, items
	p.lookaheads(states, rules, items, base, byLHS)
	if err := p.stateFree(states, rules, items, base, byLHS); err != nil {
		return nil, err
	}
	return p, nil
}

// Item is the automaton item of the tables' core item core: the item of its rule with its dot. The tables' rules
// keep their ids in the automaton, whose items add only the refused rules'; TestPrectabItemMapping holds the rest.
func (p *Precedence) Item(core int) int32 {
	it := p.Tab.Items[core]
	return p.Base[it.Rule] + int32(it.Dot)
}

// stateFree generates the state-free tables (decision §3.1) from bison's resolution of each key in every state of
// the automaton that holds it, and is an error naming each key whose decision still differs between those states
// (§3.2). On the reduce side the key is the rule and the terminal after it, and for a pair the states disagree on,
// the parent item too; on the shift side it is the spine-top item, the rule-position item before the shift and the
// terminal shifted.
func (p *Precedence) stateFree(states [][]int, rules []grammar.BNFRule, items []grammar.CoreItem, base []int, byLHS [][]int) error {
	nt, ni := p.Tab.EOF+1, len(items)
	p.NT = nt
	p.Base = make([]int32, len(rules))
	for r := range rules {
		p.Base[r] = int32(base[r])
	}
	text := func(i int) string { // item i as "lhs: x . y"
		it := items[i]
		w := []string{p.Tab.Symbols[rules[it.Rule].LHS].Name + ":"}
		for d, s := range rules[it.Rule].RHS {
			if d == it.Dot {
				w = append(w, ".")
			}
			w = append(w, p.Tab.Symbols[s].Name)
		}
		if it.Dot == len(rules[it.Rule].RHS) {
			w = append(w, ".")
		}
		return strings.Join(w, " ")
	}
	verdict := func(ok bool) uint8 { // bit 1 allowed in some state, bit 2 forbidden in some state
		if ok {
			return 1
		}
		return 2
	}
	var errs []string

	seen := make([]uint8, len(rules)*nt)
	for _, st := range p.States {
		for _, r := range st.Reduce {
			for t := range nt {
				seen[int(r)*nt+t] |= verdict(p.reducesIn(&st, int(r), t))
			}
		}
	}
	p.ReduceAt, p.MixAt, p.NeedParent = make([]uint8, len(seen)), make([]int32, len(seen)), make([]bool, len(p.Tab.Symbols))
	var mixed []int // the [r*nt+t] the states disagree on
	for k, v := range seen {
		p.MixAt[k] = -1
		switch v {
		case 2:
			p.ReduceAt[k] = 1
		case 3:
			p.ReduceAt[k], p.MixAt[k] = 2, int32(len(mixed))
			mixed = append(mixed, k)
			p.NeedParent[rules[k/nt].LHS] = true
		}
	}
	parent := make([]uint8, len(mixed)*ni)
	for q, set := range states {
		for _, pi := range set {
			it := items[pi]
			if rhs := rules[it.Rule].RHS; it.Dot < len(rhs) {
				for m, k := range mixed {
					if r := k / nt; rules[r].LHS == rhs[it.Dot] {
						parent[m*ni+pi] |= verdict(p.reducesIn(&p.States[p.Walk(int32(q), rules[r].RHS)], r, k%nt))
					}
				}
			}
		}
	}
	p.ParentForbid = make([]bool, len(parent))
	for x, v := range parent {
		p.ParentForbid[x] = v == 2
		if v == 3 {
			k, pi := mixed[x/ni], x%ni
			errs = append(errs, fmt.Sprintf("reducing %s before %s under the parent item %s", text(base[k/nt]+len(rules[k/nt].RHS)), p.Tab.Symbols[k%nt].Name, text(pi)))
		}
	}

	shiftSeen := map[[3]int32]uint8{} // (spine-top item, rule-position item, terminal)
	for q, set := range states {
		for _, top := range set {
			it := items[top]
			if it.Dot == 0 && (q != 0 || rules[it.Rule].LHS != p.Tab.Start) {
				continue // a spine hangs from a kernel item, or from a start item at the root
			}
			// The rules a spine from top can begin with: the start item's own at the root, else those its dot
			// predicts, closed over their first symbols.
			work := []int{it.Rule}
			if it.Dot > 0 {
				work = nil
				if rhs := rules[it.Rule].RHS; it.Dot < len(rhs) {
					work = slices.Clone(byLHS[rhs[it.Dot]])
				}
			}
			in := map[int]bool{}
			for len(work) > 0 {
				r := work[len(work)-1]
				work = work[:len(work)-1]
				if in[r] {
					continue
				}
				in[r] = true
				if rhs := rules[r].RHS; len(rhs) > 0 {
					work = append(work, byLHS[rhs[0]]...)
				}
			}
			// shortcut: f in FIRST(rhs[j]) only after a nullable symbol derives ε is judged in s, where bison reduces
			// the empty rule first and shifts f later, so a key forbidden by another item's shift of f refuses what
			// bison accepts (TestPrecedenceNullableShiftKnownLimit; none on perly.y,
			// TestPrecedenceNullableShiftLimitAbsent); judge such an f in the state after the empty reductions when
			// a grammar has the shape, a follow-up decision (§6).
			for r := range in {
				rhs, s := rules[r].RHS, int32(q)
				for j := 1; j < len(rhs); j++ {
					s = p.States[s].Next[int32(rhs[j-1])]
					// A terminal's FIRST is itself. A state that does not shift f (rhs[j] begins with a nullable
					// prefix there) has no shift/reduce conflict on f, so the shift is allowed in it; precedence
					// does not decide a reduce/reduce conflict, which bison resolves by rule order.
					for _, f := range p.Tab.First[rhs[j]] {
						_, sh := slices.BinarySearch(p.States[s].Shift, int32(f))
						shiftSeen[[3]int32{int32(top), int32(base[r] + j), int32(f)}] |= verdict(!sh || p.Shifts(s, f))
					}
				}
			}
		}
	}
	p.TopAt = slices.Repeat([]int32{-1}, ni)
	keys := slices.SortedFunc(maps.Keys(shiftSeen), func(a, b [3]int32) int { return slices.Compare(a[:], b[:]) })
	var rows int32
	for _, k := range keys {
		switch v := shiftSeen[k]; {
		case v == 3:
			errs = append(errs, fmt.Sprintf("shifting %s after %s under the spine-top item %s", p.Tab.Symbols[k[2]].Name, text(int(k[1])), text(int(k[0]))))
		case v == 2 && p.TopAt[k[0]] < 0:
			p.TopAt[k[0]] = rows
			rows++
		}
	}
	p.ShiftAt = slices.Repeat([]int32{-1}, int(rows)*ni)
	for _, k := range keys {
		if shiftSeen[k] != 2 {
			continue
		}
		x := int(p.TopAt[k[0]])*ni + int(k[1])
		if p.ShiftAt[x] < 0 {
			p.ShiftAt[x] = int32(len(p.ShiftForbid))
			p.ShiftForbid = append(p.ShiftForbid, make([]uint64, (nt+63)/64))
		}
		p.ShiftForbid[p.ShiftAt[x]][k[2]/64] |= 1 << (k[2] % 64)
	}
	if len(errs) > 0 {
		return fmt.Errorf("prectab: a precedence key is state-dependent, decided differently in bison's states: %s", strings.Join(errs, "; "))
	}
	return nil
}

// lookaheads sets each state's LALR(1) lookaheads by propagation to a fixed point over every item of every
// state: the start items read the end of input; an item passes its lookahead to its advance in the goto
// state, and to the first item of each rule its dot predicts, with what follows the predicted symbol (and its
// own lookahead where that derives the empty string). Merging by LR(0) state is what makes it LALR(1).
func (p *Precedence) lookaheads(states [][]int, rules []grammar.BNFRule, items []grammar.CoreItem, base []int, byLHS [][]int) {
	tab := p.Tab
	words := (tab.EOF + 64) / 64
	la := make([][][]uint64, len(states)) // la[s][k]: the lookahead of state s's k-th item
	at := func(s int32, item int) []uint64 {
		k, _ := slices.BinarySearch(states[s], item)
		return la[s][k]
	}
	for s := range states {
		la[s] = make([][]uint64, len(states[s]))
		for k := range la[s] {
			la[s][k] = make([]uint64, words)
		}
	}
	for _, i := range states[0] {
		if rules[items[i].Rule].LHS == tab.Start && items[i].Dot == 0 {
			at(0, i)[tab.EOF/64] |= 1 << (tab.EOF % 64)
		}
	}
	union := func(dst, src []uint64) bool {
		changed := false
		for w := range dst {
			if dst[w]|src[w] != dst[w] {
				dst[w] |= src[w]
				changed = true
			}
		}
		return changed
	}
	first := make([]uint64, words)
	for changed := true; changed; {
		changed = false
		for s := range states {
			for k, i := range states[s] {
				it := items[i]
				rhs := rules[it.Rule].RHS
				if it.Dot == len(rhs) {
					continue
				}
				x := rhs[it.Dot]
				changed = union(at(p.States[s].Next[int32(x)], i+1), la[s][k]) || changed
				if tab.Symbols[x].Terminal {
					continue
				}
				clear(first) // what can follow x here: FIRST of the rest, and this item's lookahead if it can vanish
				nullable := true
				for _, y := range rhs[it.Dot+1:] {
					for _, t := range tab.First[y] {
						first[t/64] |= 1 << (t % 64)
					}
					if nullable = tab.Symbols[y].Nullable; !nullable {
						break
					}
				}
				if nullable {
					union(first, la[s][k])
				}
				for _, r := range byLHS[x] {
					changed = union(at(int32(s), base[r]), first) || changed
				}
			}
		}
	}
	for s := range states {
		st := &p.States[s]
		for _, r := range st.Reduce {
			var ts []int32
			for t, w := range at(int32(s), base[r]+len(rules[r].RHS)) {
				for ; w != 0; w &= w - 1 {
					ts = append(ts, int32(t*64+bits.TrailingZeros64(w)))
				}
			}
			st.LA = append(st.LA, ts)
		}
	}
}

// Walk is the state reached from state s over syms. Every state a derivation begins a rule in holds the
// rule's first item, so the goto exists; a missing one is a constrainer bug.
func (p *Precedence) Walk(s int32, syms []int) int32 {
	for _, x := range syms {
		next, ok := p.States[s].Next[int32(x)]
		if !ok {
			panic("prectab: no goto in the precedence constraint's parser state")
		}
		s = next
	}
	return s
}

// resolution is bison's resolution of a conflict between reducing a rule and shifting a terminal.
type resolution int

const (
	unresolved resolution = iota // either has no level: precedence does not resolve it
	reduce
	shift
	refuse // %nonassoc: bison's resolution is an error action
)

// resolve is bison's resolution of the conflict between reducing rule r and shifting terminal t.
func (p *Precedence) resolve(r, t int) resolution {
	rl, tl := p.RuleLevel[r], 0
	if t < len(p.TermLevel) {
		tl = p.TermLevel[t]
	}
	switch {
	case rl == 0 || tl == 0:
		return unresolved
	case rl > tl:
		return reduce
	case rl < tl:
		return shift
	}
	switch p.Assoc[rl] {
	case "left":
		return reduce
	case "right":
		return shift
	}
	return refuse
}

// Reduces reports whether bison's resolution lets rule r, begun in state s, reduce with t next: where the
// state after its symbols can also shift t, the resolution must be reduce.
func (p *Precedence) Reduces(r int, s int32, t int) bool {
	return p.reducesIn(&p.States[p.Walk(s, p.Tab.Rules[r].RHS)], r, t)
}

// Reduces in st, the state after rule r's symbols; unexported because nothing outside prectab calls it.
func (p *Precedence) reducesIn(st *State, r, t int) bool {
	if _, conflict := slices.BinarySearch(st.Shift, int32(t)); !conflict {
		return true
	}
	res := p.resolve(r, t)
	return res == unresolved || res == reduce
}

// Shifts reports whether bison's resolution lets t shift in state s: every rule s reduces on t, by its LALR(1)
// lookahead there, must lose to the shift.
func (p *Precedence) Shifts(s int32, t int) bool {
	for i, r := range p.States[s].Reduce {
		if _, la := slices.BinarySearch(p.States[s].LA[i], int32(t)); !la {
			continue
		}
		if res := p.resolve(int(r), t); res == reduce || res == refuse {
			return false
		}
	}
	return true
}

// ReducesFree is the state-free verdict on reducing rule r with t next, whose node is its parent's child at
// item pi. A pair that needs its parent item and has none is a judgement bug, never an allowance.
func (p *Precedence) ReducesFree(r, t int, pi int32) bool {
	k := r*p.NT + t
	switch p.ReduceAt[k] {
	case 0:
		return true
	case 1:
		return false
	}
	if pi < 0 {
		panic(fmt.Sprintf("prectab: reducing rule %d (%s) before %s needs its parent item, and none was given", r, p.Tab.Symbols[p.Tab.Rules[r].LHS].Name, p.Tab.Symbols[t].Name))
	}
	return !p.ParentForbid[int(p.MixAt[k])*len(p.TopAt)+int(pi)]
}

// ShiftForbidden reports whether shifting t after the rule-position item at, under the spine top top, is forbidden.
func (p *Precedence) ShiftForbidden(top, at int32, t int32) bool {
	if top < 0 || p.TopAt[top] < 0 {
		return false
	}
	x := p.ShiftAt[int(p.TopAt[top])*len(p.TopAt)+int(at)]
	return x >= 0 && p.ShiftForbid[x][t/64]&(1<<(t%64)) != 0
}

// Parsed is the committed grammar's precedence over a tables value, ParsePrecedence's grammar.Derived value, built
// once per value; Err where the tables are not the committed grammar's, whose rule ids the %prec list is indexed by.
// A test stands an error in for it by deriving a Parsed on its own tables value first.
type Parsed struct {
	P   *Precedence
	Err error
}

func buildParsed(tab *grammar.Tables) Parsed {
	g, _, err := grammar.Committed()
	if err != nil {
		return Parsed{Err: err}
	}
	if len(tab.Rules) != len(g.G1) {
		return Parsed{Err: fmt.Errorf("prectab: %d table rules for perly.y's %d: not the committed grammar's tables", len(tab.Rules), len(g.G1))}
	}
	rulePrec := make([]string, len(g.G1))
	for r, rule := range g.G1 {
		if tab.Symbols[tab.Rules[r].LHS].Name != rule.LHS {
			return Parsed{Err: fmt.Errorf("prectab: table rule %d is %s's, perly.y's is %s's", r, tab.Symbols[tab.Rules[r].LHS].Name, rule.LHS)}
		}
		rulePrec[r] = rule.Prec
	}
	p, err := NewPrecedence(tab, g.Levels, rulePrec, g.Refused)
	return Parsed{p, err}
}

// ParsePrecedence is the committed grammar's Precedence over tab as forest.ParseWith derives it, once per tables
// value; an error where tab is not the committed grammar's tables.
func ParsePrecedence(tab *grammar.Tables) (*Precedence, error) {
	p := grammar.Derived(tab, buildParsed)
	return p.P, p.Err
}
