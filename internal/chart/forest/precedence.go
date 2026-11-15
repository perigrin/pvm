// ABOUTME: The precedence constraint (spec §8): bison's resolutions applied to a built forest, after recognition.
// ABOUTME: Each derivation is judged by its edge context, the parser state before it and the terminal after it; items are untouched.

package forest

import (
	"fmt"
	"maps"
	"slices"

	"tamarou.com/pvm/internal/chart/grammar"
)

// Precedence is a grammar's precedence declarations, bison's way: each terminal's level, each rule's (its %prec
// token's, else its last terminal's), each level's associativity, and bison's LR(0) parser states, read from
// the grammar alone.
type Precedence struct {
	tab    *grammar.Tables
	term   []int     // by symbol: a terminal's level, 0 for none
	rule   []int     // by rule: its level, 0 for none
	assoc  []string  // by level: left, right or nonassoc
	states []lrState // state 0 is the file's start
}

// lrState is an LR(0) state: the terminals it shifts and the rules it reduces, both sorted, and its goto on
// each symbol after a dot.
type lrState struct {
	shift, reduce []int32
	next          map[int32]int32
}

// NewPrecedence is levels, loosest first, over tab, whose rule r carries the %prec token rulePrec[r] ("" for
// none). A token that is no terminal of tab, as a %prec-only token is, still gives its rules their level.
//
// The states are bison's: the canonical LR(0) automaton of the grammar, closure and goto, which bison's LALR(1)
// tables share. The constraint carries the one a derivation reaches, from the file's start along the symbols
// before it, so a conflict is found exactly where bison's state has one. They are built from the tables alone,
// never read from the chart, whose split states are not bison's.
func NewPrecedence(tab *grammar.Tables, levels []grammar.Level, rulePrec []string) *Precedence {
	p := &Precedence{tab: tab, term: make([]int, len(tab.Symbols)), rule: make([]int, len(tab.Rules)),
		assoc: []string{""}}
	level := map[string]int{}
	for i, l := range levels {
		p.assoc = append(p.assoc, l.Assoc)
		for _, tok := range l.Tokens {
			level[tok] = i + 1
		}
	}
	for id, s := range tab.Symbols {
		if s.Terminal {
			p.term[id] = level[s.Name]
		}
	}
	for r, rule := range tab.Rules {
		if rulePrec[r] != "" {
			p.rule[r] = level[rulePrec[r]]
			continue
		}
		for _, s := range slices.Backward(rule.RHS) {
			if tab.Symbols[s].Terminal {
				p.rule[r] = p.term[s]
				break
			}
		}
	}
	base := grammar.Derived(tab, buildForestGrammar).core // base[r] is rule r's item at dot 0
	byLHS := make([][]int, len(tab.Symbols))
	var start []int
	for r, rule := range tab.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
		if rule.LHS == tab.Start {
			start = append(start, base[r])
		}
	}
	closure := func(items []int) []int {
		items = slices.Clone(items)
		in := map[int]bool{}
		for i := 0; i < len(items); i++ {
			if in[items[i]] {
				continue
			}
			in[items[i]] = true
			it := tab.Items[items[i]]
			if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) {
				for _, r := range byLHS[rhs[it.Dot]] {
					items = append(items, base[r])
				}
			}
		}
		return slices.Sorted(maps.Keys(in))
	}
	id := map[string]int32{} // each state's items, as text, to its index
	var items [][]int
	intern := func(set []int) int32 {
		k := fmt.Sprint(set)
		if i, ok := id[k]; ok {
			return i
		}
		id[k] = int32(len(items))
		items = append(items, set)
		return id[k]
	}
	intern(closure(start))
	for s := 0; s < len(items); s++ { // breadth first; intern appends the states it finds
		st := lrState{next: map[int32]int32{}}
		kernels := map[int][]int{} // each symbol after a dot, and the items it advances to
		for _, i := range items[s] {
			it := tab.Items[i]
			rhs := tab.Rules[it.Rule].RHS
			switch {
			case it.Dot == len(rhs):
				st.reduce = append(st.reduce, int32(it.Rule))
			case tab.Symbols[rhs[it.Dot]].Terminal:
				st.shift = append(st.shift, int32(rhs[it.Dot]))
			}
			if it.Dot < len(rhs) {
				kernels[rhs[it.Dot]] = append(kernels[rhs[it.Dot]], i+1)
			}
		}
		slices.Sort(st.shift)
		st.shift = slices.Compact(st.shift)
		for _, x := range slices.Sorted(maps.Keys(kernels)) {
			st.next[int32(x)] = intern(closure(kernels[x]))
		}
		p.states = append(p.states, st)
	}
	return p
}

// walk is the state reached from state s over syms. Every state a derivation begins a rule in holds the
// rule's first item, so the goto exists; a missing one is a constrainer bug.
func (p *Precedence) walk(s int32, syms []int) int32 {
	for _, x := range syms {
		next, ok := p.states[s].next[int32(x)]
		if !ok {
			panic("forest: no goto in the precedence constraint's parser state")
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
	rl, tl := p.rule[r], 0
	if t < len(p.term) {
		tl = p.term[t]
	}
	switch {
	case rl == 0 || tl == 0:
		return unresolved
	case rl > tl:
		return reduce
	case rl < tl:
		return shift
	}
	switch p.assoc[rl] {
	case "left":
		return reduce
	case "right":
		return shift
	}
	return refuse
}

// reduces reports whether bison's resolution lets rule r, begun in state s, reduce with t next: where the
// state after its symbols can also shift t, the resolution must be reduce.
func (p *Precedence) reduces(r int, s int32, t int) bool {
	s = p.walk(s, p.tab.Rules[r].RHS)
	if _, conflict := slices.BinarySearch(p.states[s].shift, int32(t)); !conflict {
		return true
	}
	res := p.resolve(r, t)
	return res == unresolved || res == reduce
}

// shifts reports whether bison's resolution lets t shift in state s: every rule s can reduce with t next must
// lose to the shift.
func (p *Precedence) shifts(s int32, t int) bool {
	for _, r := range p.states[s].reduce {
		if _, la := slices.BinarySearch(p.tab.Follow[p.tab.Rules[r].LHS], t); !la {
			continue
		}
		if res := p.resolve(int(r), t); res == reduce || res == refuse {
			return false
		}
	}
	return true
}

// Constrain applies p to f in place: it keeps each alternative some derivation of the file bison's
// resolutions would build passes through and drops the rest, true; or refuses the file, false, where the root
// keeps no derivation. A derivation is judged in its edge context: the parser state the derivation holding it
// reaches at its span's start (state 0 at the file's start), and the terminal after it. Each reduction is
// checked against the terminal after it, each shift against the reductions the state before it holds; a span
// left with no derivation in its context fails every derivation holding it there.
// shortcut: an alternative kept in one context is kept in all, so a node reached in two parser states that keep
// different alternatives keeps both; the verdict is exact, and only the derivations a later reader enumerates
// can include one bison builds in neither state. Revisit if READINGS counts one.
func (f *Forest) Constrain(p *Precedence) bool {
	c := &constrainer{f: f, p: p, firsts: map[ctx][]int32{}, marked: map[ctx]bool{}, kept: make([]bool, len(f.Alts))}
	if len(c.eval(f.Root, 0, p.tab.EOF)) == 0 {
		return false
	}
	c.mark(f.Root, 0, p.tab.EOF, -1)
	for n := range f.Nodes {
		a := &f.Nodes[n].Alt
		for *a != 0 {
			if c.kept[*a] {
				a = &f.Alts[*a].Next
			} else {
				*a = f.Alts[*a].Next
			}
		}
	}
	return true
}

// constrainer is one application of a Precedence to a forest.
type constrainer struct {
	f      *Forest
	p      *Precedence
	firsts map[ctx][]int32 // eval's results
	marked map[ctx]bool    // mark's visits, first the terminal its derivations must begin with
	kept   []bool          // by alternative
}

// ctx is a node in an edge context: the parser state at its start and the terminal after it, and for mark the
// terminal its derivations must begin with (-1 any).
type ctx struct{ n, state, t, first int32 }

// rule is the rule of a symbol node's or rule position's alternative a, and how many of its symbols a covers.
func (c *constrainer) rule(n, a int32) (r, m int) {
	r = int(c.f.Alts[a].Rule)
	if nd := c.f.Nodes[n]; nd.Sym < 0 {
		return r, int(nd.Dot)
	}
	return r, len(c.p.tab.Rules[r].RHS)
}

// eval is the first terminals of n's derivations that bison's resolutions pass in context (s, t), sorted;
// none where every one is forbidden. A span with no tokens begins with t.
func (c *constrainer) eval(n, s int32, t int) []int32 {
	nd := c.f.Nodes[n]
	if nd.Alt == 0 {
		return []int32{nd.Sym} // a leaf
	}
	k := ctx{n, s, int32(t), -1}
	if out, ok := c.firsts[k]; ok {
		return out
	}
	var out []int32
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m := c.rule(n, a)
		if nd.Sym >= 0 && !c.p.reduces(r, s, t) {
			continue
		}
		out = append(out, c.seq(a, r, m, s, t, nil)...)
	}
	slices.Sort(out)
	out = slices.Compact(out)
	c.firsts[k] = out
	return out
}

// seq is the first terminals of alternative a's derivations, rule r's first m symbols begun in state s with t
// next, that bison's resolutions pass. With keep it also marks, through keep, the children of each derivation
// whose first terminal it accepts.
func (c *constrainer) seq(a int32, r, m int, s int32, t int, keep func(first int32) bool) []int32 {
	if m == 0 {
		return []int32{int32(t)}
	}
	alt := c.f.Alts[a]
	before := c.p.walk(s, c.p.tab.Rules[r].RHS[:m-1]) // the state the right child begins in
	right := c.eval(alt.Right, before, t)
	if m == 1 {
		for _, f := range right {
			if keep != nil && keep(f) {
				c.mark(alt.Right, before, t, f)
			}
		}
		return right
	}
	var out []int32
	empty := c.f.Nodes[alt.Right].From == c.f.Nodes[alt.Right].To && c.f.Nodes[alt.Right].Alt != 0
	for _, f := range right {
		// A shift of f where the right child begins, unless the right child is empty and reduces before f is read.
		if !empty && !c.p.shifts(before, int(f)) {
			continue
		}
		left := c.eval(alt.Left, s, int(f))
		out = append(out, left...)
		if keep == nil {
			continue
		}
		for _, l := range left {
			if keep(l) {
				c.mark(alt.Right, before, t, f)
				c.mark(alt.Left, s, int(f), l)
			}
		}
	}
	return out
}

// mark keeps the alternatives of n's derivations that bison's resolutions pass in context (s, t) and that
// begin with first (-1 any), and those of their children.
func (c *constrainer) mark(n, s int32, t int, first int32) {
	nd := c.f.Nodes[n]
	k := ctx{n, s, int32(t), first}
	if nd.Alt == 0 || c.marked[k] {
		return
	}
	c.marked[k] = true
	keep := func(f int32) bool { return first < 0 || f == first }
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m := c.rule(n, a)
		if nd.Sym >= 0 && !c.p.reduces(r, s, t) {
			continue
		}
		if slices.ContainsFunc(c.seq(a, r, m, s, t, keep), keep) {
			c.kept[a] = true
		}
	}
}
