// ABOUTME: A YAEP-style Earley recognizer over the generated LR(0) tables and chart/lex's token lattice.
// ABOUTME: Prediction, scanning and completion over (state, distance) items, with one-token lookahead on completion.

// Package recognize is the chart's recognizer (spec §5, "chart/recognize").
// An Earley item is an LR(0) state of the split automaton and its origin,
// held as the distance in lattice nodes back to that origin, so a set's
// content is position-free and equal sets intern to one (sets.go). A new set
// is the old one advanced over a token and closed under completion; the
// transition cache (cache.go) serves a repeat of that step. Items hold no
// context facts (rule 2). Leo memos (leo.go) live beside the sets, not in
// them. Error recovery is not built here.
package recognize

import (
	"fmt"
	"slices"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// Options switch parts of the recognizer off. Neither changes acceptance;
// the tests and the fuzz target hold them to that.
type Options struct {
	// NoCache neither reads nor writes the transition cache.
	NoCache bool
	// NoLookahead completes every complete item, whatever follows.
	NoLookahead bool
	// NoLeo makes no Leo memo, so a right recursion completes link by link.
	NoLeo bool
}

// Recognizer recognizes lattices over one grammar's tables. Its intern
// table and transition cache live as long as it does, across inputs.
type Recognizer struct {
	tab  *grammar.Tables
	opts Options
	term map[string]int // terminal name to symbol id
	// lhs is, per state, the distinct LHS symbols of its reduce rules.
	lhs [][]int
	// follow is, per nonterminal, the terminals that can follow it, with
	// bit eof for the end of input.
	follow []bitset
	eof    int
	// classes interns a sorted terminal list, a token's or a lookahead's,
	// to an id; bits is each class as a bitset.
	classes map[string]int
	bits    []bitset
	sets    interner
	cache   map[cacheKey][]cacheEntry
	leo     map[[2]int]int // (state, symbol) to the LHS of its Leo item
}

// New is a recognizer over tab.
func New(tab *grammar.Tables, opts Options) *Recognizer {
	r := &Recognizer{
		tab: tab, opts: opts, term: map[string]int{}, eof: len(tab.Symbols),
		classes: map[string]int{}, sets: newInterner(), cache: map[cacheKey][]cacheEntry{},
	}
	for id, s := range tab.Symbols {
		if s.Terminal {
			r.term[s.Name] = id
		}
	}
	r.lhs = make([][]int, len(tab.States))
	for s, st := range tab.States {
		for _, rule := range st.Reduces {
			r.lhs[s] = append(r.lhs[s], tab.Rules[rule].LHS)
		}
		slices.Sort(r.lhs[s])
		r.lhs[s] = slices.Compact(r.lhs[s])
	}
	r.follow = follow(tab)
	r.leo = leoItems(r)
	return r
}

// Chart is one recognition: the Earley set at each lattice node, nil where
// no item survives. Interned marks a node whose set was already in the
// intern table; Cached, a node served by the transition cache.
type Chart struct {
	r        *Recognizer
	Sets     []*Set
	Interned []bool
	Cached   []bool
	leo      []map[int]*memo // the side table: per node, Leo memos by symbol
}

// Recognize recognizes l, whose nodes are numbered so that every edge runs
// forward. A node some edge passes over is inside a branch: its set is not
// interned and the cache is not used there (spec §5, "sets intern outside
// branches only").
func Recognize[S lex.State[S]](r *Recognizer, l *lex.Lattice[S]) *Chart {
	n := len(l.Nodes)
	c := &Chart{r: r, Sets: make([]*Set, n), Interned: make([]bool, n), Cached: make([]bool, n), leo: make([]map[int]*memo, n)}
	if n == 0 {
		return c
	}
	in := make([][]lex.Edge, n)
	out := make([][]int, n) // the terminal ids leaving each node
	branch := make([]bool, n)
	for _, e := range l.Edges {
		in[e.To] = append(in[e.To], e)
		out[e.From] = append(out[e.From], r.ids(e.Token.Terminals)...)
		for k := e.From + 1; k < e.To; k++ {
			branch[k] = true
		}
	}
	c.Sets[0], c.Interned[0] = r.sets.intern([]Item{{State: 0}})
	for pos := 1; pos < n; pos++ {
		la := r.lookahead(out[pos])
		var key cacheKey
		// The key names the source set by its id, so it must be interned.
		cacheable := !r.opts.NoCache && !branch[pos] && len(in[pos]) == 1 && c.Sets[in[pos][0].From] != nil && c.Sets[in[pos][0].From].id >= 0
		if cacheable {
			e := in[pos][0]
			key = cacheKey{set: c.Sets[e.From].id, span: pos - e.From, tok: r.class(r.ids(e.Token.Terminals)), la: la}
			if s := r.lookup(c, pos, key); s != nil {
				c.Sets[pos], c.Interned[pos], c.Cached[pos] = s, true, true
				continue
			}
		}
		b := builder{r: r, c: c, pos: pos, seen: map[[2]int]bool{}, interned: cacheable}
		for _, e := range in[pos] {
			b.scan(e)
		}
		var follows bitset // nil completes everything
		if !r.opts.NoLookahead {
			follows = r.bits[la]
		}
		b.complete(follows)
		if len(b.items) == 0 {
			continue
		}
		items := b.relative()
		if branch[pos] {
			c.Sets[pos] = &Set{id: -1, items: items}
			continue
		}
		c.Sets[pos], c.Interned[pos] = r.sets.intern(items)
		if cacheable && b.interned {
			r.store(key, b.deps, c.Sets[pos])
		}
	}
	return c
}

// Accepted reports whether the last node's set holds a complete item of a
// start-symbol rule from node 0.
func (c *Chart) Accepted() bool {
	if len(c.Sets) == 0 {
		return false
	}
	last := len(c.Sets) - 1
	for _, it := range c.Sets[last].Items() {
		if it.Dist == last && slices.Contains(c.r.lhs[it.State], c.r.tab.Start) {
			return true
		}
	}
	return false
}

// Reuse is how many of the nodes that have a set got it from the intern
// table, and how many have a set.
func (c *Chart) Reuse() (hits, positions int) {
	for i, s := range c.Sets {
		if s != nil {
			positions++
			if c.Interned[i] {
				hits++
			}
		}
	}
	return hits, positions
}

// builder makes one node's set: items as (state, absolute origin) while it
// is built, and the origin sets completion read, for the cache.
type builder struct {
	r        *Recognizer
	c        *Chart
	pos      int
	items    [][2]int
	seen     map[[2]int]bool
	deps     []dep
	depSeen  map[dep]bool
	interned bool // the node is cacheable and every set read is interned
}

// add adds an item, and a kernel state's predicted state at this node.
func (b *builder) add(state, origin int) {
	k := [2]int{state, origin}
	if b.seen[k] {
		return
	}
	b.seen[k] = true
	b.items = append(b.items, k)
	if eps := b.r.tab.States[state].Epsilon; eps >= 0 {
		b.add(eps, b.pos)
	}
}

// scan advances every item of e's source set over e's token.
func (b *builder) scan(e lex.Edge) {
	src := b.c.Sets[e.From]
	if src == nil {
		return
	}
	b.read(e.From, src)
	terms := b.r.ids(e.Token.Terminals)
	for _, it := range src.items {
		for _, t := range terms {
			if k, ok := b.r.goTo(it.State, t); ok {
				b.add(k, e.From-it.Dist)
			}
		}
	}
}

// complete closes the set under completion. An item from this node is a
// prediction: its completions are empty, and the automaton's nullable
// closure already made them. With la, a rule completes only if something
// la holds can follow its LHS; a nil la completes every rule. Where a Leo
// memo applies, the top of the recursion is added in one step.
func (b *builder) complete(la bitset) {
	for i := 0; i < len(b.items); i++ {
		state, origin := b.items[i][0], b.items[i][1]
		if origin == b.pos {
			continue
		}
		for _, lhs := range b.r.lhs[state] {
			if la != nil && !b.r.follow[lhs].intersects(la) {
				continue
			}
			src := b.c.Sets[origin]
			b.read(origin, src)
			if top, ok := b.leo(origin, lhs, la); ok {
				b.add(top[0], top[1])
				continue
			}
			for _, it := range src.items {
				if k, ok := b.r.goTo(it.State, lhs); ok {
					b.add(k, origin-it.Dist)
				}
			}
		}
	}
}

// read records that the set at origin was read, by scan or completion.
func (b *builder) read(origin int, s *Set) {
	if s.id < 0 {
		b.interned = false
		return
	}
	d := dep{dist: b.pos - origin, set: s.id}
	if b.depSeen == nil {
		b.depSeen = map[dep]bool{}
	}
	if !b.depSeen[d] {
		b.depSeen[d] = true
		b.deps = append(b.deps, d)
	}
}

// relative is the items as (state, distance), sorted, the set's content.
func (b *builder) relative() []Item {
	items := make([]Item, len(b.items))
	for i, it := range b.items {
		items[i] = Item{State: it[0], Dist: b.pos - it[1]}
	}
	slices.SortFunc(items, func(x, y Item) int {
		if x.State != y.State {
			return x.State - y.State
		}
		return x.Dist - y.Dist
	})
	return items
}

// goTo is state's transition on sym.
func (r *Recognizer) goTo(state, sym int) (int, bool) {
	tr := r.tab.States[state].Transitions
	i, ok := slices.BinarySearchFunc(tr, sym, func(t grammar.Transition, s int) int { return t.Symbol - s })
	if !ok {
		return 0, false
	}
	return tr[i].Target, true
}

// ids is the terminal ids of names, in name order.
func (r *Recognizer) ids(names []string) []int {
	out := make([]int, 0, len(names))
	for _, n := range names {
		if id, ok := r.term[n]; ok {
			out = append(out, id)
		}
	}
	return out
}

// class interns a terminal list, sorted and deduplicated, to an id.
func (r *Recognizer) class(ids []int) int {
	ids = slices.Compact(slices.Sorted(slices.Values(ids)))
	k := fmt.Sprint(ids)
	if c, ok := r.classes[k]; ok {
		return c
	}
	bs := newBitset(r.eof + 1)
	for _, id := range ids {
		bs.set(id)
	}
	r.classes[k] = len(r.bits)
	r.bits = append(r.bits, bs)
	return len(r.bits) - 1
}

// lookahead is the class of the terminals leaving a node, or of the end of
// input where none leave.
func (r *Recognizer) lookahead(out []int) int {
	if len(out) == 0 {
		return r.class([]int{r.eof})
	}
	return r.class(out)
}

// follow is FOLLOW of every symbol, over the terminals and the end of input.
func follow(tab *grammar.Tables) []bitset {
	n := len(tab.Symbols)
	nullable := func(s int) bool { return !tab.Symbols[s].Terminal && tab.Symbols[s].Nullable }
	first := make([]bitset, n)
	fol := make([]bitset, n)
	for s := range tab.Symbols {
		first[s], fol[s] = newBitset(n+1), newBitset(n+1)
		if tab.Symbols[s].Terminal {
			first[s].set(s)
		}
	}
	for changed := true; changed; {
		changed = false
		for _, rule := range tab.Rules {
			for _, x := range rule.RHS {
				changed = first[rule.LHS].or(first[x]) || changed
				if !nullable(x) {
					break
				}
			}
		}
	}
	fol[tab.Start].set(n)
	for changed := true; changed; {
		changed = false
		for _, rule := range tab.Rules {
			for i, x := range rule.RHS {
				if tab.Symbols[x].Terminal {
					continue
				}
				rest := true // everything after x is nullable
				for _, y := range rule.RHS[i+1:] {
					changed = fol[x].or(first[y]) || changed
					if !nullable(y) {
						rest = false
						break
					}
				}
				if rest {
					changed = fol[x].or(fol[rule.LHS]) || changed
				}
			}
		}
	}
	return fol
}

// bitset is a fixed-size set of small integers.
type bitset []uint64

func newBitset(n int) bitset { return make(bitset, (n+63)/64) }
func (b bitset) set(i int)   { b[i/64] |= 1 << (i % 64) }
func (b bitset) intersects(o bitset) bool {
	for i := range b {
		if b[i]&o[i] != 0 {
			return true
		}
	}
	return false
}

// or adds o to b and reports whether b grew.
func (b bitset) or(o bitset) bool {
	grew := false
	for i := range b {
		if v := b[i] | o[i]; v != b[i] {
			b[i], grew = v, true
		}
	}
	return grew
}
