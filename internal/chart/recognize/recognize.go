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
	"sync"
	"sync/atomic"

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
// table and transition cache live as long as it does, across inputs, and
// change with every recognition: a Recognizer is not safe for concurrent use.
type Recognizer struct {
	tab  *grammar.Tables
	opts Options
	// g is the grammar-only tables, shared with every recognizer over tab;
	// cls is the classes as this recognizer last read them from g.
	g     *grammarTables
	cls   *classes
	sets  interner
	cache map[cacheKey]*cacheNode // each key's trie root
	trie  map[cacheEdge]*cacheNode
	// stored counts the cache's stores, to order its entries; longestScan
	// is the most trie distances one lookup examined, for the test of its bound.
	stored, longestScan int
	// expectedComputed counts Expected's computations, for the test of its
	// memo, which lives in the cache's trie (expected.go).
	expectedComputed int
	// probes counts goTo's binary searches, for the test of the postdot
	// index's bound (probes_test.go).
	probes int
	// memoScans counts memoAt's whole-set scans, which only a branch set
	// takes, for the test that NoLeo makes none (leo_test.go).
	memoScans int
	// itemsBuf, depsBuf and relBuf are Recognize's builder's buffers,
	// reused from node to node: what keeps their contents copies them.
	itemsBuf [][2]int
	depsBuf  []dep
	relBuf   []Item
}

// grammarTables is what New derives from the tables alone. It is built once
// per *grammar.Tables, by the first New over it, kept beside it
// (grammar.Derived), and read by every
// recognizer over that value from any goroutine, so the tables' states,
// rules, items, symbols and FOLLOW sets must not change after the first New.
// The rule: nothing in it is written after it is built except cls, the one
// part derived from TerminalSets, which may grow after New. A set appended to
// the tables is covered by extend, under mu, which publishes new classes and
// never writes ones already published; a set id once covered is never
// rebuilt, so TerminalSets may be appended to but its existing sets never
// changed. A tables value with other contents is a copy with its own part.
type grammarTables struct {
	// lhs is, per state, the distinct LHS symbols of its reduce rules.
	lhs [][]int
	// follow is tab.Follow per symbol as bitsets, with bit eof for the end of input.
	follow []bitset
	eof    int
	leo    map[[2]int]int // (state, symbol) to the LHS of its Leo item
	leoSym []bool         // per symbol, whether any state has a Leo item awaiting it
	// semi is PERLY_SEMICOLON's symbol id and brace PERLY_BRACE_CLOSE's,
	// the ruby slipper's (slipper.go); -1 where the grammar lacks them.
	semi, brace int
	// eofBits is the end of input's lookahead class, key eofClass.
	eofBits bitset
	mu      sync.Mutex
	cls     atomic.Pointer[classes]
}

// classes is the lookahead classes of the terminal sets with ids below
// len(terms). Published classes are never written.
type classes struct {
	// terms is the sets as read from tab.TerminalSets: validation, scan
	// and bits all read this one copy.
	terms [][]int
	// bits is each set as a lookahead bitset, by id; closes marks the sets
	// holding PERLY_BRACE_CLOSE, before which the slipper inserts a ;.
	bits   []bitset
	closes []bool
}

// eofClass is the end of input's lookahead class key: no set id, so a set
// added later never takes it.
const eofClass = -2

// buildGrammar derives tab's grammar-only tables.
func buildGrammar(tab *grammar.Tables) *grammarTables {
	terminal := func(name string) int {
		return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Terminal && s.Name == name })
	}
	g := &grammarTables{eof: tab.EOF, semi: terminal("PERLY_SEMICOLON"), brace: terminal("PERLY_BRACE_CLOSE"), leo: leoItems(tab)}
	g.leoSym = make([]bool, len(tab.Symbols))
	for k := range g.leo {
		g.leoSym[k[1]] = true
	}
	g.eofBits = g.bitsOf([]int{g.eof})
	if g.semi >= 0 {
		g.eofBits.set(g.semi) // the end of input closes, so a ; may be inserted there
	}
	g.lhs = make([][]int, len(tab.States))
	for s, st := range tab.States {
		for _, rule := range st.Reduces {
			g.lhs[s] = append(g.lhs[s], tab.Rules[rule].LHS)
		}
		slices.Sort(g.lhs[s])
		g.lhs[s] = slices.Compact(g.lhs[s])
	}
	for _, ids := range tab.Follow {
		g.follow = append(g.follow, g.bitsOf(ids))
	}
	g.cls.Store(&classes{})
	g.extend(tab)
	return g
}

// extend is the classes covering every set tab now has, publishing new
// classes for the sets added since the last; tab is the value g was built from.
func (g *grammarTables) extend(tab *grammar.Tables) *classes {
	g.mu.Lock()
	defer g.mu.Unlock()
	old := g.cls.Load()
	if len(tab.TerminalSets) <= len(old.terms) {
		return old
	}
	// Clipped, so append copies rather than write past what old's readers see.
	c := &classes{terms: slices.Clip(old.terms), bits: slices.Clip(old.bits), closes: slices.Clip(old.closes)}
	for _, ids := range tab.TerminalSets[len(old.terms):] {
		closes := g.brace >= 0 && slices.Contains(ids, g.brace)
		b := g.bitsOf(ids)
		// Where a ; may be inserted, completion must also see it may follow.
		if closes && g.semi >= 0 {
			b.set(g.semi)
		}
		c.terms = append(c.terms, ids)
		c.bits = append(c.bits, b)
		c.closes = append(c.closes, closes)
	}
	g.cls.Store(c)
	return c
}

// New is a recognizer over tab. The grammar-only tables are tab's shared
// ones; the intern table, cache and counters are the recognizer's own.
// The first `New` over a tables value writes beside it, so do not copy
// `*tab` while that `New` runs.
func New(tab *grammar.Tables, opts Options) *Recognizer {
	g := grammar.Derived(tab, buildGrammar)
	return &Recognizer{tab: tab, opts: opts, g: g, cls: g.cls.Load(), sets: newInterner(), cache: map[cacheKey]*cacheNode{}, trie: map[cacheEdge]*cacheNode{}}
}

// Chart is one recognition: the Earley set at each lattice node, nil where
// no item survives. Interned marks a node whose set was already in the
// intern table; Cached, a node served by the transition cache.
type Chart struct {
	r        *Recognizer
	Sets     []*Set
	Interned []bool
	Cached   []bool
	leo      map[[2]int]*memo // the side table: real Leo memos by (node, symbol); nil until one is made
	inserted []int            // nodes where the slipper scanned a ;
	// slips is the slipper's closure at each node recognition built it for,
	// by (node, the next token's terminal set, -1 at the end of input), Leo
	// on as recognition closed it; nil until one is kept (Slipped).
	slips map[[2]int][]Item
	// slipAccepted is whether the ; inserted at end of input completes the start.
	slipAccepted bool
}

// Recognize recognizes l, whose nodes are numbered so that every edge runs
// forward and whose tokens' set ids index r's tables; it is an error where
// they do not. A node some edge passes over is inside a branch: its set is
// not interned and the cache is not used there (spec §5, "sets intern
// outside branches only").
func Recognize[S lex.State[S]](r *Recognizer, l *lex.Lattice[S]) (*Chart, error) {
	n := len(l.Nodes)
	for _, e := range l.Edges {
		if e.Token.TerminalSet >= len(r.cls.terms) {
			r.cls = r.g.extend(r.tab) // a set added to the tables after New
			break
		}
	}
	for i, e := range l.Edges {
		switch {
		case e.From < 0 || e.To >= n:
			return nil, fmt.Errorf("recognize: edge %d, %d to %d, has a node outside the lattice's %d", i, e.From, e.To, n)
		case e.From >= e.To:
			return nil, fmt.Errorf("recognize: edge %d, %d to %d, does not run forward", i, e.From, e.To)
		case e.Token.TerminalSet < 0 || e.Token.TerminalSet >= len(r.cls.terms):
			return nil, fmt.Errorf("recognize: edge %d's terminal set %d is outside the table's %d", i, e.Token.TerminalSet, len(r.cls.terms))
		}
	}
	c := &Chart{r: r, Sets: make([]*Set, n), Interned: make([]bool, n), Cached: make([]bool, n)}
	if n == 0 {
		return c, nil
	}
	in := make([][]lex.Edge, n)
	out := make([][]int, n) // the set ids leaving each node
	branch := make([]bool, n)
	// One backing array each, carved by count, instead of a slice per node.
	nin, nout := make([]int32, n), make([]int32, n)
	for _, e := range l.Edges {
		nin[e.To]++
		nout[e.From]++
	}
	inAll, outAll := make([]lex.Edge, len(l.Edges)), make([]int, len(l.Edges))
	for i, a, b := 0, 0, 0; i < n; i++ {
		in[i], out[i] = inAll[a:a:a+int(nin[i])], outAll[b:b:b+int(nout[i])]
		a, b = a+int(nin[i]), b+int(nout[i])
	}
	for _, e := range l.Edges {
		in[e.To] = append(in[e.To], e)
		out[e.From] = append(out[e.From], e.Token.TerminalSet)
		for k := e.From + 1; k < e.To; k++ {
			branch[k] = true
		}
	}
	c.Sets[0], c.Interned[0] = r.intern([]Item{{State: 0}})
	for pos := 1; pos < n; pos++ {
		la, laBits := r.lookahead(out[pos])
		var key cacheKey
		// The key names the source set by its id, so it must be interned.
		cacheable := la != -1 && !r.opts.NoCache && !branch[pos] && len(in[pos]) == 1 && c.Sets[in[pos][0].From] != nil && c.Sets[in[pos][0].From].id >= 0
		if cacheable {
			e := in[pos][0]
			key = cacheKey{set: c.Sets[e.From].id, span: pos - e.From, tok: e.Token.TerminalSet, la: la}
			if s := r.lookup(c, pos, key); s != nil {
				c.Sets[pos], c.Interned[pos], c.Cached[pos] = s, true, true
				continue
			}
		}
		b := builder{r: r, c: c, pos: pos, interned: cacheable, items: r.itemsBuf[:0], deps: r.depsBuf[:0]}
		for _, e := range in[pos] {
			b.scan(e)
			if r.cls.closes[e.Token.TerminalSet] {
				b.slip(e)
			}
		}
		var follows bitset // nil completes everything
		if !r.opts.NoLookahead {
			follows = laBits
		}
		b.complete(follows)
		r.itemsBuf, r.depsBuf = b.items, b.deps
		if len(b.items) == 0 {
			continue
		}
		items := b.relative()
		if branch[pos] {
			c.Sets[pos] = &Set{id: -1, items: slices.Clone(items)}
			continue
		}
		c.Sets[pos], c.Interned[pos] = r.intern(items)
		if cacheable && b.interned {
			r.store(key, b.deps, c.Sets[pos])
		}
	}
	c.slipped(out)
	return c, nil
}

// Accepted reports whether the last node's set holds a complete item of a
// start-symbol rule from node 0.
func (c *Chart) Accepted() bool {
	if len(c.Sets) == 0 {
		return false
	}
	if c.slipAccepted {
		return true
	}
	last := len(c.Sets) - 1
	for _, it := range c.Sets[last].Items() {
		if it.Dist == last && slices.Contains(c.r.g.lhs[it.State], c.r.tab.Start) {
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
// is built, and the origin sets completion read, for the cache. Each finds
// its duplicates by scanning while small and keeps a map (seen, depSeen)
// only once it holds more than smallSet (YAEP's sit_dist_insert, adapted).
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

// smallSet is the size up to which a builder finds duplicates by scanning.
const smallSet = 24

// isNew reports whether x is not yet in list, and records it in seen once
// the list is past smallSet; the caller appends it.
func isNew[T comparable](list []T, seen *map[T]bool, x T) bool {
	if *seen != nil {
		if (*seen)[x] {
			return false
		}
		(*seen)[x] = true
		return true
	}
	if slices.Contains(list, x) {
		return false
	}
	if len(list) == smallSet {
		*seen = make(map[T]bool, 4*smallSet)
		for _, y := range list {
			(*seen)[y] = true
		}
		(*seen)[x] = true
	}
	return true
}

// add adds an item, and a kernel state's predicted state at this node.
func (b *builder) add(state, origin int) {
	k := [2]int{state, origin}
	if !isNew(b.items, &b.seen, k) {
		return
	}
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
	if src.core != nil {
		for _, t := range b.r.cls.terms[e.Token.TerminalSet] {
			for _, en := range b.r.awaiting(src, t) {
				b.add(int(en.target), e.From-src.items[en.idx].Dist)
			}
		}
		return
	}
	// shortcut: a branch set has no core, so it searches every item; give
	// branch sets a postdot index when interleaved mode makes them common.
	for _, it := range src.items {
		for _, t := range b.r.cls.terms[e.Token.TerminalSet] {
			if k, ok := b.r.goTo(it.State, t); ok {
				b.add(k, e.From-it.Dist)
			}
		}
	}
}

// complete closes the set under completion. A predicted item from this
// node has empty completions, and the automaton's nullable closure already
// made them; a kernel item from this node is one the slipper's zero-width ;
// advanced, and completes into the set before the ;. With la, a rule
// completes only if something la holds can follow its LHS; a nil la completes every rule. Where a Leo
// memo applies, the top of the recursion is added in one step.
func (b *builder) complete(la bitset) {
	for i := 0; i < len(b.items); i++ {
		state, origin := b.items[i][0], b.items[i][1]
		if origin == b.pos && b.r.tab.States[state].Predicted {
			continue
		}
		for _, lhs := range b.r.g.lhs[state] {
			if la != nil && !b.r.g.follow[lhs].intersects(la) {
				continue
			}
			src := b.c.Sets[origin]
			b.read(origin, src)
			if top, ok := b.leo(origin, lhs, la); ok {
				b.add(top[0], top[1])
				continue
			}
			if src.core != nil {
				for _, en := range b.r.awaiting(src, lhs) {
					b.add(int(en.target), origin-src.items[en.idx].Dist)
				}
				continue
			}
			// shortcut: as in scan, a branch set searches every item.
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
	if isNew(b.deps, &b.depSeen, d) {
		b.deps = append(b.deps, d)
	}
}

// relative is the items as (state, distance), sorted, the set's content, in
// the recognizer's buffer: a caller that keeps it copies it.
func (b *builder) relative() []Item {
	items := b.r.relBuf[:0]
	for _, it := range b.items {
		items = append(items, Item{State: it[0], Dist: b.pos - it[1]})
	}
	b.r.relBuf = items
	slices.SortFunc(items, func(x, y Item) int {
		if x.State != y.State {
			return x.State - y.State
		}
		return x.Dist - y.Dist
	})
	return items
}

// goTo is state's transition on sym, counted in probes: recognition's
// lookup. A read of a finished chart uses transition, which writes nothing.
func (r *Recognizer) goTo(state, sym int) (int, bool) {
	r.probes++
	return transition(r.tab, state, sym)
}

// transition is state's transition on sym.
func transition(tab *grammar.Tables, state, sym int) (int, bool) {
	tr := tab.States[state].Transitions
	i, ok := slices.BinarySearchFunc(tr, sym, func(t grammar.Transition, s int) int { return t.Symbol - s })
	if !ok {
		return 0, false
	}
	return tr[i].Target, true
}

// bitsOf is ids as a bitset over the terminals and the end of input.
func (g *grammarTables) bitsOf(ids []int) bitset {
	b := newBitset(g.eof + 1)
	for _, id := range ids {
		b.set(id)
	}
	return b
}

// lookahead is the class key and bitset of the set ids leaving a node: the
// one set id, or eofClass where none leave. Where different
// sets leave, the key is -1, uncacheable, and the bitset their union.
// shortcut: the union is built per node and never cached; intern unions
// when interleaved mode makes forks common.
func (r *Recognizer) lookahead(out []int) (int, bitset) {
	if len(out) == 0 {
		return eofClass, r.g.eofBits
	}
	for _, s := range out[1:] {
		if s != out[0] {
			u := newBitset(r.g.eof + 1)
			for _, s := range out {
				for i, w := range r.cls.bits[s] {
					u[i] |= w
				}
			}
			return -1, u
		}
	}
	return out[0], r.cls.bits[out[0]]
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
