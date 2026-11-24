// ABOUTME: The precedence constraint (spec §8): bison's resolutions applied to a built forest, after recognition.
// ABOUTME: Constrain judges by state-free keys generated from bison's states; the test-only reference by parser state. Items are untouched.

package forest

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
	tab    *grammar.Tables
	term   []int     // by symbol: a terminal's level, 0 for none
	rule   []int     // by rule, the tables' then the refused ones: its level, 0 for none
	assoc  []string  // by level: left, right or nonassoc
	states []lrState // state 0 is the file's start

	// The state-free tables (decision §3.1), generated with the automaton and read by index, never through a map.
	// Rules and items are the automaton's: the tables', then the refused ones'.
	nt           int        // terminals, EOF included
	base         []int32    // by rule: its item at dot 0
	reduceAt     []uint8    // [r*nt+t]: 0 every state reducing r reduces it with t next, 1 none does, 2 the parent item decides
	mixAt        []int32    // [r*nt+t]: a 2's row of parentForbid, -1 otherwise
	parentForbid []bool     // [row*items+parent item]: the reduction is forbidden under that parent item
	needParent   []bool     // by symbol: some rule of it has a 2
	topAt        []int32    // by item: its row of shiftAt as a spine-top item, -1 where it forbids no shift
	shiftAt      []int32    // [row*items+rule-position item]: its set in shiftForbid, -1 for none
	shiftForbid  [][]uint64 // terminal bitsets: the shifts forbidden after that item under that spine top
}

// lrState is an LR(0) state: the terminals it shifts, sorted, the rules it reduces with each one's LALR(1)
// lookahead (la[i] the sorted terminals reduce[i] reduces on), and its goto on each symbol after a dot.
type lrState struct {
	shift, reduce []int32
	la            [][]int32
	next          map[int32]int32
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
	p := &Precedence{tab: tab, term: make([]int, len(tab.Symbols)), assoc: []string{""}}
	sym := map[string]int{}
	for id, s := range tab.Symbols {
		sym[s.Name] = id
	}
	rules, items, rulePrec := slices.Clone(tab.Rules), slices.Clone(tab.Items), slices.Clone(rulePrec)
	base := slices.Clone(grammar.Derived(tab, buildForestGrammar).core) // base[r] is rule r's item at dot 0
	for _, r := range refused {
		var rhs []int
		for _, s := range append([]grammar.Symbol{{Text: r.LHS}}, r.RHS...) {
			id, ok := sym[s.Text]
			if !ok {
				return nil, fmt.Errorf("forest: the refused rule %s: names %s, which is no symbol of the tables", r.LHS, s.Text)
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
	p.rule = make([]int, len(rules))
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
	for r, rule := range rules {
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
		st := lrState{next: map[int32]int32{}}
		kernels := map[int][]int{} // each symbol after a dot, and the items it advances to
		for _, i := range states[s] {
			it := items[i]
			rhs := rules[it.Rule].RHS
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
	p.lookaheads(states, rules, items, base, byLHS)
	if err := p.stateFree(states, rules, items, base, byLHS); err != nil {
		return nil, err
	}
	return p, nil
}

// stateFree generates the state-free tables (decision §3.1) from bison's resolution of each key in every state of
// the automaton that holds it, and is an error naming each key whose decision still differs between those states
// (§3.2). On the reduce side the key is the rule and the terminal after it, and for a pair the states disagree on,
// the parent item too; on the shift side it is the spine-top item, the rule-position item before the shift and the
// terminal shifted.
func (p *Precedence) stateFree(states [][]int, rules []grammar.BNFRule, items []grammar.CoreItem, base []int, byLHS [][]int) error {
	nt, ni := p.tab.EOF+1, len(items)
	p.nt = nt
	p.base = make([]int32, len(rules))
	for r := range rules {
		p.base[r] = int32(base[r])
	}
	text := func(i int) string { // item i as "lhs: x . y"
		it := items[i]
		w := []string{p.tab.Symbols[rules[it.Rule].LHS].Name + ":"}
		for d, s := range rules[it.Rule].RHS {
			if d == it.Dot {
				w = append(w, ".")
			}
			w = append(w, p.tab.Symbols[s].Name)
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
	for _, st := range p.states {
		for _, r := range st.reduce {
			for t := range nt {
				seen[int(r)*nt+t] |= verdict(p.reducesIn(&st, int(r), t))
			}
		}
	}
	p.reduceAt, p.mixAt, p.needParent = make([]uint8, len(seen)), make([]int32, len(seen)), make([]bool, len(p.tab.Symbols))
	var mixed []int // the [r*nt+t] the states disagree on
	for k, v := range seen {
		p.mixAt[k] = -1
		switch v {
		case 2:
			p.reduceAt[k] = 1
		case 3:
			p.reduceAt[k], p.mixAt[k] = 2, int32(len(mixed))
			mixed = append(mixed, k)
			p.needParent[rules[k/nt].LHS] = true
		}
	}
	parent := make([]uint8, len(mixed)*ni)
	for q, set := range states {
		for _, pi := range set {
			it := items[pi]
			if rhs := rules[it.Rule].RHS; it.Dot < len(rhs) {
				for m, k := range mixed {
					if r := k / nt; rules[r].LHS == rhs[it.Dot] {
						parent[m*ni+pi] |= verdict(p.reducesIn(&p.states[p.walk(int32(q), rules[r].RHS)], r, k%nt))
					}
				}
			}
		}
	}
	p.parentForbid = make([]bool, len(parent))
	for x, v := range parent {
		p.parentForbid[x] = v == 2
		if v == 3 {
			k, pi := mixed[x/ni], x%ni
			errs = append(errs, fmt.Sprintf("reducing %s before %s under the parent item %s", text(base[k/nt]+len(rules[k/nt].RHS)), p.tab.Symbols[k%nt].Name, text(pi)))
		}
	}

	shiftSeen := map[[3]int32]uint8{} // (spine-top item, rule-position item, terminal)
	for q, set := range states {
		for _, top := range set {
			it := items[top]
			if it.Dot == 0 && (q != 0 || rules[it.Rule].LHS != p.tab.Start) {
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
			for r := range in {
				rhs, s := rules[r].RHS, int32(q)
				for j := 1; j < len(rhs); j++ {
					s = p.states[s].next[int32(rhs[j-1])]
					for _, f := range p.states[s].shift {
						if int(f) == rhs[j] || slices.Contains(p.tab.First[rhs[j]], int(f)) {
							shiftSeen[[3]int32{int32(top), int32(base[r] + j), f}] |= verdict(p.shifts(s, int(f)))
						}
					}
				}
			}
		}
	}
	p.topAt = slices.Repeat([]int32{-1}, ni)
	keys := slices.SortedFunc(maps.Keys(shiftSeen), func(a, b [3]int32) int { return slices.Compare(a[:], b[:]) })
	var rows int32
	for _, k := range keys {
		switch v := shiftSeen[k]; {
		case v == 3:
			errs = append(errs, fmt.Sprintf("shifting %s after %s under the spine-top item %s", p.tab.Symbols[k[2]].Name, text(int(k[1])), text(int(k[0]))))
		case v == 2 && p.topAt[k[0]] < 0:
			p.topAt[k[0]] = rows
			rows++
		}
	}
	p.shiftAt = slices.Repeat([]int32{-1}, int(rows)*ni)
	for _, k := range keys {
		if shiftSeen[k] != 2 {
			continue
		}
		x := int(p.topAt[k[0]])*ni + int(k[1])
		if p.shiftAt[x] < 0 {
			p.shiftAt[x] = int32(len(p.shiftForbid))
			p.shiftForbid = append(p.shiftForbid, make([]uint64, (nt+63)/64))
		}
		p.shiftForbid[p.shiftAt[x]][k[2]/64] |= 1 << (k[2] % 64)
	}
	if len(errs) > 0 {
		return fmt.Errorf("forest: a precedence key is state-dependent, decided differently in bison's states: %s", strings.Join(errs, "; "))
	}
	return nil
}

// lookaheads sets each state's LALR(1) lookaheads by propagation to a fixed point over every item of every
// state: the start items read the end of input; an item passes its lookahead to its advance in the goto
// state, and to the first item of each rule its dot predicts, with what follows the predicted symbol (and its
// own lookahead where that derives the empty string). Merging by LR(0) state is what makes it LALR(1).
func (p *Precedence) lookaheads(states [][]int, rules []grammar.BNFRule, items []grammar.CoreItem, base []int, byLHS [][]int) {
	tab := p.tab
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
				changed = union(at(p.states[s].next[int32(x)], i+1), la[s][k]) || changed
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
		st := &p.states[s]
		for _, r := range st.reduce {
			var ts []int32
			for t, w := range at(int32(s), base[r]+len(rules[r].RHS)) {
				for ; w != 0; w &= w - 1 {
					ts = append(ts, int32(t*64+bits.TrailingZeros64(w)))
				}
			}
			st.la = append(st.la, ts)
		}
	}
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
	return p.reducesIn(&p.states[p.walk(s, p.tab.Rules[r].RHS)], r, t)
}

// reducesIn is reduces in st, the state after rule r's symbols.
func (p *Precedence) reducesIn(st *lrState, r, t int) bool {
	if _, conflict := slices.BinarySearch(st.shift, int32(t)); !conflict {
		return true
	}
	res := p.resolve(r, t)
	return res == unresolved || res == reduce
}

// shifts reports whether bison's resolution lets t shift in state s: every rule s reduces on t, by its LALR(1)
// lookahead there, must lose to the shift.
func (p *Precedence) shifts(s int32, t int) bool {
	for i, r := range p.states[s].reduce {
		if _, la := slices.BinarySearch(p.states[s].la[i], int32(t)); !la {
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
// keeps no derivation. It judges with no parser state (decision §3.3), by the keys of p's state-free tables: a
// reduction by its rule and the terminal after it, and the parent item for a pair that needs one; a shift by
// the spine-top item, the rule-position item before it and the terminal shifted. A span left with no
// derivation in its context fails every derivation holding it there. ConstrainReference is what it is
// compared against.
// shortcut: an alternative kept in one context is kept in all, so a node reached in two contexts that keep
// different alternatives keeps both; the verdict is exact, and only the derivations a later reader enumerates
// can include one bison builds in neither. Revisit if READINGS counts one.
func (f *Forest) Constrain(p *Precedence) bool {
	return newJudgement(f, p).constrain()
}

// constrain is Constrain on the judgement's forest, its memo as it stands.
func (c *judgement) constrain() bool {
	f := c.f
	if len(c.eval(f.Root, c.p.tab.EOF, rootTop, -1)) == 0 {
		return false
	}
	c.mark(f.Root, c.p.tab.EOF, rootTop, -1, -1)
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

// ConstrainReference is used only by tests. It is the judgement Constrain made before the state-free evaluator,
// kept under its own name as the fixed point the differential checks compare production Constrain against, so
// production can change without changing what it is judged by; TestPrecedenceReferenceAgreesWithBison holds its
// agreement with bison. It applies p to f in place as Constrain does, verdict, unlinking and shortcut alike, but
// judges a derivation in its edge context: the parser state the derivation holding it reaches at its span's start
// (state 0 at the file's start), and the terminal after it. Each reduction is checked against the terminal after
// it, each shift against the reductions the state before it holds. It must not call Constrain, nor Constrain it.
// Its known defect: shifts forbids a terminal the state does not shift wherever a reduction there with it in its
// lookahead resolves to reduce, where bison has no conflict and reduces. No perly.y forest reaches it;
// TestPrecedenceReferenceKnownDivergence pins the toy where Constrain and bison accept what it refuses.
func (f *Forest) ConstrainReference(p *Precedence) bool {
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

// rootTop stands for the root's spine top: the start item of each root alternative's own rule.
const rootTop = -2

// judgement is one state-free application of a Precedence to a forest. Everything it writes lives here, on the
// call, never on the shared Precedence. Its memo is per node and uses no hashing: evalAt[n] and markAt[n] head
// node n's lists in evals and marks (0 none, so entry 0 is unused), linked by next.
type judgement struct {
	f              *Forest
	p              *Precedence
	evalAt, markAt []int32
	evals          []evalEntry
	marks          []markEntry
	kept           []bool // by alternative
}

// evalEntry is eval's result in one context: the terminal after the node, its spine top and its parent item.
type evalEntry struct {
	t, top, pi, next int32
	out              []int32 // shared with every parent that reads it; never appended into
}

// markEntry is one mark visit: eval's context and the terminal its derivations must begin with (-1 any).
type markEntry struct{ t, top, pi, first, next int32 }

func newJudgement(f *Forest, p *Precedence) *judgement {
	return &judgement{f: f, p: p, evalAt: make([]int32, len(f.Nodes)), markAt: make([]int32, len(f.Nodes)),
		evals: make([]evalEntry, 1, len(f.Nodes)), marks: make([]markEntry, 1, len(f.Nodes)), kept: make([]bool, len(f.Alts))}
}

// canon drops the context node n cannot use, so two encodings of one context share a memo entry: a spine top
// that forbids no shift, and a parent item where n's symbol has no pair that needs one.
func (c *judgement) canon(n, top, pi int32) (int32, int32) {
	if top >= 0 && c.p.topAt[top] < 0 {
		top = -1
	}
	if sym := c.f.Nodes[n].Sym; sym < 0 || !c.p.needParent[sym] {
		pi = -1
	}
	return top, pi
}

// reducesFree is the state-free verdict on reducing rule r with t next, whose node is its parent's child at
// item pi. A pair that needs its parent item and has none is a judgement bug, never an allowance.
func (p *Precedence) reducesFree(r, t int, pi int32) bool {
	k := r*p.nt + t
	switch p.reduceAt[k] {
	case 0:
		return true
	case 1:
		return false
	}
	if pi < 0 {
		panic(fmt.Sprintf("forest: reducing rule %d (%s) before %s needs its parent item, and none was given", r, p.tab.Symbols[p.tab.Rules[r].LHS].Name, p.tab.Symbols[t].Name))
	}
	return !p.parentForbid[int(p.mixAt[k])*len(p.topAt)+int(pi)]
}

// shiftForbidden reports whether shifting t after the rule-position item at, under the spine top top, is forbidden.
func (p *Precedence) shiftForbidden(top, at int32, t int32) bool {
	if top < 0 || p.topAt[top] < 0 {
		return false
	}
	x := p.shiftAt[int(p.topAt[top])*len(p.topAt)+int(at)]
	return x >= 0 && p.shiftForbid[x][t/64]&(1<<(t%64)) != 0
}

// alt is the rule of node n's alternative a, how many of its symbols a covers, and the spine top its own
// symbols hang from: top, or at the root the start item of a's rule.
func (c *judgement) alt(n, a, top int32) (r, m int, at int32) {
	r = int(c.f.Alts[a].Rule)
	m = len(c.p.tab.Rules[r].RHS)
	if nd := c.f.Nodes[n]; nd.Sym < 0 {
		m = int(nd.Dot)
	}
	if top == rootTop {
		top = c.p.base[r]
	}
	return r, m, top
}

// eval is the first terminals of n's derivations that bison's resolutions pass with t next, under the spine top
// top and the parent item pi, sorted; none where every one is forbidden. A span with no tokens begins with t.
func (c *judgement) eval(n int32, t int, top, pi int32) []int32 {
	nd := c.f.Nodes[n]
	if nd.Alt == 0 {
		return []int32{nd.Sym} // a leaf
	}
	top, pi = c.canon(n, top, pi)
	for e := c.evalAt[n]; e != 0; e = c.evals[e].next {
		if x := &c.evals[e]; x.t == int32(t) && x.top == top && x.pi == pi {
			return x.out
		}
	}
	var out []int32
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m, at := c.alt(n, a, top)
		if nd.Sym >= 0 && !c.p.reducesFree(r, t, pi) {
			continue
		}
		out = append(out, c.seq(a, r, m, t, at, nil)...)
	}
	slices.Sort(out)
	out = slices.Compact(out)
	c.evals = append(c.evals, evalEntry{int32(t), top, pi, c.evalAt[n], out})
	c.evalAt[n] = int32(len(c.evals) - 1)
	return out
}

// seq is the first terminals of alternative a's derivations, rule r's first m symbols under the spine top top
// with t next, that bison's resolutions pass. The right child hangs from the rule-position item before it,
// unless it is r's first symbol and inherits top; the left inherits top. With keep it also marks, through keep,
// the children of each derivation whose first terminal it accepts.
func (c *judgement) seq(a int32, r, m int, t int, top int32, keep func(first int32) bool) []int32 {
	if m == 0 {
		return []int32{int32(t)}
	}
	alt := c.f.Alts[a]
	at := c.p.base[r] + int32(m-1) // the rule-position item before the right child
	rtop := top
	if m > 1 {
		rtop = at
	}
	right := c.eval(alt.Right, t, rtop, at)
	if m == 1 {
		for _, f := range right {
			if keep != nil && keep(f) {
				c.mark(alt.Right, t, rtop, at, f)
			}
		}
		return right
	}
	lpi := int32(-1) // the left child's parent item, where it is r's first symbol
	if m == 2 {
		lpi = c.p.base[r]
	}
	var out []int32
	empty := c.f.Nodes[alt.Right].From == c.f.Nodes[alt.Right].To && c.f.Nodes[alt.Right].Alt != 0
	for _, f := range right {
		// A shift of f where the right child begins, unless the right child is empty and reduces before f is read.
		if !empty && c.p.shiftForbidden(top, at, f) {
			continue
		}
		left := c.eval(alt.Left, int(f), top, lpi)
		out = append(out, left...)
		if keep == nil {
			continue
		}
		for _, l := range left {
			if keep(l) {
				c.mark(alt.Right, t, rtop, at, f)
				c.mark(alt.Left, int(f), top, lpi, l)
			}
		}
	}
	return out
}

// mark keeps the alternatives of n's derivations that bison's resolutions pass in eval's context and that begin
// with first (-1 any), and those of their children.
func (c *judgement) mark(n int32, t int, top, pi, first int32) {
	if c.f.Nodes[n].Alt == 0 {
		return
	}
	top, pi = c.canon(n, top, pi)
	for e := c.markAt[n]; e != 0; e = c.marks[e].next {
		if x := c.marks[e]; x.t == int32(t) && x.top == top && x.pi == pi && x.first == first {
			return
		}
	}
	c.marks = append(c.marks, markEntry{int32(t), top, pi, first, c.markAt[n]})
	c.markAt[n] = int32(len(c.marks) - 1)
	keep := func(f int32) bool { return first < 0 || f == first }
	nd := c.f.Nodes[n]
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m, at := c.alt(n, a, top)
		if nd.Sym >= 0 && !c.p.reducesFree(r, t, pi) {
			continue
		}
		if slices.ContainsFunc(c.seq(a, r, m, t, at, keep), keep) {
			c.kept[a] = true
		}
	}
}
