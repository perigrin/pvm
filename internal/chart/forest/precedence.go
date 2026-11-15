// ABOUTME: The precedence constraint (spec §8): bison's resolutions applied to a built forest, after recognition.
// ABOUTME: Each derivation is judged by its edge context, the symbol before it and the terminal after it; items are untouched.

package forest

import (
	"maps"
	"slices"

	"tamarou.com/pvm/internal/chart/grammar"
)

// Precedence is a grammar's precedence declarations, bison's way: each terminal's level, each rule's (its %prec
// token's, else its last terminal's), each level's associativity, and what a parser state can shift and reduce
// after each pair of symbols, read from the grammar alone.
type Precedence struct {
	tab   *grammar.Tables
	term  []int    // by symbol: a terminal's level, 0 for none
	rule  []int    // by rule: its level, 0 for none
	assoc []string // by level: left, right or nonassoc
	// at is the state after each symbol, at[λ+1] (λ -1 the start of the file): the terminals it shifts.
	// after is the state after λ then X: the terminals it shifts and the rules it reduces.
	at    [][]int32
	after map[[2]int32]step
}

// step is a state's shifts, terminal ids, and reductions, rule ids, both sorted.
type step struct{ shift, reduce []int32 }

// NewPrecedence is levels, loosest first, over tab, whose rule r carries the %prec token rulePrec[r] ("" for
// none). A token that is no terminal of tab, as a %prec-only token is, still gives its rules their level.
//
// The state after a symbol λ is approximated from λ alone: every item whose dot follows λ, closed. That is the
// derivation's left edge context, never the chart's state; bison's state after λ holds a subset of those items,
// so a conflict found here that bison's state lacks is a disagreement the toy referees would show.
func NewPrecedence(tab *grammar.Tables, levels []grammar.Level, rulePrec []string) *Precedence {
	p := &Precedence{tab: tab, term: make([]int, len(tab.Symbols)), rule: make([]int, len(tab.Rules)),
		assoc: []string{""}, after: map[[2]int32]step{}}
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
	seeds := make([][]int, len(tab.Symbols)+1) // seeds[λ+1]: the items whose dot follows λ, the start's at 0
	for r, rule := range tab.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
		if rule.LHS == tab.Start {
			seeds[0] = append(seeds[0], base[r])
		}
		for d, x := range rule.RHS {
			seeds[x+1] = append(seeds[x+1], base[r]+d+1)
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
	steps := func(items []int) step {
		var s step
		for _, i := range items {
			it := tab.Items[i]
			if rhs := tab.Rules[it.Rule].RHS; it.Dot == len(rhs) {
				s.reduce = append(s.reduce, int32(it.Rule))
			} else if tab.Symbols[rhs[it.Dot]].Terminal {
				s.shift = append(s.shift, int32(rhs[it.Dot]))
			}
		}
		slices.Sort(s.shift)
		s.shift = slices.Compact(s.shift)
		return s
	}
	for lambda := -1; lambda < len(tab.Symbols); lambda++ {
		state := closure(seeds[lambda+1])
		p.at = append(p.at, steps(state).shift)
		next := map[int][]int{} // each symbol after a dot, and the items it advances to
		for _, i := range state {
			it := tab.Items[i]
			if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) {
				next[rhs[it.Dot]] = append(next[rhs[it.Dot]], i+1)
			}
		}
		for x, items := range next {
			p.after[[2]int32{int32(lambda), int32(x)}] = steps(closure(items))
		}
	}
	return p
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

// reduces reports whether bison's resolution lets rule r reduce with t next, its last symbol following lambda
// (its own, or the derivation's where the rule has one symbol): where the state can also shift t, the
// resolution must be reduce.
func (p *Precedence) reduces(r int, lambda int32, t int) bool {
	rhs := p.tab.Rules[r].RHS
	shifts := p.at[lambda+1]
	if n := len(rhs); n > 0 {
		if n > 1 {
			lambda = int32(rhs[n-2])
		}
		shifts = p.after[[2]int32{lambda, int32(rhs[n-1])}].shift
	}
	if _, conflict := slices.BinarySearch(shifts, int32(t)); !conflict {
		return true
	}
	res := p.resolve(r, t)
	return res == unresolved || res == reduce
}

// shifts reports whether bison's resolution lets t shift after x, which follows lambda: every rule the state
// there can reduce with t next must lose to the shift.
func (p *Precedence) shifts(lambda, x int32, t int) bool {
	for _, r := range p.after[[2]int32{lambda, x}].reduce {
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
// keeps no derivation. A derivation is judged in its edge context: lambda, the symbol before its span in the
// derivation holding it (-1 at the file's start), and the terminal after it. Each reduction is checked against
// the terminal after it, each shift against the reductions the state before it holds; a span left with no
// derivation in its context fails every derivation holding it there.
// shortcut: an alternative kept in one context is kept in all, so a node shared by contexts that keep different
// alternatives can leave a combination no context passes; the toy forests share none, revisit in PRECPERLY.
func (f *Forest) Constrain(p *Precedence) bool {
	c := &constrainer{f: f, p: p, firsts: map[ctx][]int32{}, marked: map[ctx]bool{}, kept: make([]bool, len(f.Alts))}
	if len(c.eval(f.Root, -1, p.tab.EOF)) == 0 {
		return false
	}
	c.mark(f.Root, -1, p.tab.EOF, -1)
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

// ctx is a node in an edge context: the symbol before it and the terminal after it, and for mark the
// terminal its derivations must begin with (-1 any).
type ctx struct{ n, lambda, t, first int32 }

// rule is the rule of a symbol node's or rule position's alternative a, and how many of its symbols a covers.
func (c *constrainer) rule(n, a int32) (r, m int) {
	r = int(c.f.Alts[a].Rule)
	if nd := c.f.Nodes[n]; nd.Sym < 0 {
		return r, int(nd.Dot)
	}
	return r, len(c.p.tab.Rules[r].RHS)
}

// eval is the first terminals of n's derivations that bison's resolutions pass in context (lambda, t), sorted;
// none where every one is forbidden. A span with no tokens begins with t.
func (c *constrainer) eval(n, lambda int32, t int) []int32 {
	nd := c.f.Nodes[n]
	if nd.Alt == 0 {
		return []int32{nd.Sym} // a leaf
	}
	k := ctx{n, lambda, int32(t), -1}
	if out, ok := c.firsts[k]; ok {
		return out
	}
	var out []int32
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m := c.rule(n, a)
		if nd.Sym >= 0 && !c.p.reduces(r, lambda, t) {
			continue
		}
		out = append(out, c.seq(a, r, m, lambda, t, nil)...)
	}
	slices.Sort(out)
	out = slices.Compact(out)
	c.firsts[k] = out
	return out
}

// seq is the first terminals of alternative a's derivations, rule r's first m symbols after lambda with t next,
// that bison's resolutions pass. With keep it also marks, through keep, the children of each derivation whose
// first terminal it accepts.
func (c *constrainer) seq(a int32, r, m int, lambda int32, t int, keep func(first int32) bool) []int32 {
	if m == 0 {
		return []int32{int32(t)}
	}
	rhs := c.p.tab.Rules[r].RHS
	alt := c.f.Alts[a]
	before := func(i int) int32 { // the symbol before rule r's i-th, from 0
		if i == 0 {
			return lambda
		}
		return int32(rhs[i-1])
	}
	right := c.eval(alt.Right, before(m-1), t)
	if m == 1 {
		for _, f := range right {
			if keep != nil && keep(f) {
				c.mark(alt.Right, before(0), t, f)
			}
		}
		return right
	}
	var out []int32
	empty := c.f.Nodes[alt.Right].From == c.f.Nodes[alt.Right].To && c.f.Nodes[alt.Right].Alt != 0
	for _, f := range right {
		// A shift of f after symbol m-1, unless the right child is empty and reduces before f is read.
		if !empty && !c.p.shifts(before(m-2), int32(rhs[m-2]), int(f)) {
			continue
		}
		left := c.eval(alt.Left, lambda, int(f))
		out = append(out, left...)
		if keep == nil {
			continue
		}
		for _, l := range left {
			if keep(l) {
				c.mark(alt.Right, before(m-1), t, f)
				c.mark(alt.Left, lambda, int(f), l)
			}
		}
	}
	return out
}

// mark keeps the alternatives of n's derivations that bison's resolutions pass in context (lambda, t) and that
// begin with first (-1 any), and those of their children.
func (c *constrainer) mark(n, lambda int32, t int, first int32) {
	nd := c.f.Nodes[n]
	k := ctx{n, lambda, int32(t), first}
	if nd.Alt == 0 || c.marked[k] {
		return
	}
	c.marked[k] = true
	keep := func(f int32) bool { return first < 0 || f == first }
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m := c.rule(n, a)
		if nd.Sym >= 0 && !c.p.reduces(r, lambda, t) {
			continue
		}
		if slices.ContainsFunc(c.seq(a, r, m, lambda, t, keep), keep) {
			c.kept[a] = true
		}
	}
}
