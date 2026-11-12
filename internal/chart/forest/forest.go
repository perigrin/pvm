// ABOUTME: The shared packed forest, built after recognition from a chart's completed items, Leo memos and inserted ;s.
// ABOUTME: Nodes are (symbol or rule position, span); nothing is added to an Earley item, and the recognizer stays Boolean.

// Package forest builds the chart's forest after recognition (spec §9.4,
// §3 rule 12). The chart's items stay (state, distance): what a node derives
// is read from the completed items, the Leo memos' skipped completions and
// the items a slipper ; advanced, and the forest is built top-down from the
// root over those facts. Each fact is a derivation the chart already proved,
// so every node has one; every node on a derivation of the file is a fact
// the chart recorded, so none is missing.
package forest

import (
	"slices"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// Forest is a file's shared packed forest, built only for a file the chart
// accepts. No node has an Alt whose child is the node itself.
type Forest struct {
	Root  *Node
	size  int // the nodes built, which bounds Derivations' work
	steps int // facts recorded and splits tried, for the scaling test
}

// Node is a symbol over a span of lattice nodes, From to To, or a rule
// position there: the first Dot symbols of Rule, Sym -1. A terminal is a
// leaf, one token wide, or zero-width where it is a ; the slipper inserted.
// A nonterminal or a rule position has one Alt per way its span divides.
type Node struct {
	Sym, Rule, Dot int
	From, To       int
	Alts           []Alt
}

// Alt is one packed derivation: Rule's symbols up to the node's last, as
// Left (nil where there are none; that symbol's node where there is one; a
// rule position where there are more), then the last symbol's node as Right
// (nil for an empty rule).
type Alt struct {
	Rule        int
	Left, Right *Node
}

// Tree is one derivation of a symbol over From to To: the rule, and its
// children in order. A leaf is a terminal, Rule -1 and no children.
type Tree struct {
	Sym, Rule, From, To int
	Kids                []*Tree
}

// Refusal is why a file has no forest: the byte it is refused at, and what
// refused it. Its zero value is no refusal.
type Refusal struct {
	Byte   int
	Reason string
}

// Parse recognizes src over tab's grammar and, where the chart accepts it,
// builds its forest. A refused file has a refusal and no forest: at the
// first token no terminal admits (lex-unmapped), else at the token whose
// set died, or the end of the file (grammar).
func Parse(tab *grammar.Tables, src []byte) (*Forest, Refusal) {
	s, unmapped := lex.Lex(src, tab.Tokens)
	if unmapped >= 0 {
		return nil, Refusal{Byte: unmapped, Reason: "lex-unmapped"}
	}
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		return nil, Refusal{Reason: err.Error()}
	}
	if !c.Accepted() {
		at := len(src) // every set survived: the file ends too soon
		if d := slices.Index(c.Sets, nil); d > 0 {
			at = s.Raw[s.Tokens[d-1].Index].Start
		}
		return nil, Refusal{Byte: at, Reason: "grammar"}
	}
	return build(tab, c, l), Refusal{}
}

// builder holds what the chart proved, and the nodes built from it.
type builder struct {
	tab      *grammar.Tables
	c        *recognize.Chart
	l        *lex.Lattice[lex.None]
	semi     int
	inserted []bool
	// facts is each (symbol, origin, rule) the chart completed, those ending
	// at node j being facts[at[j]:at[j+1]], sorted; node is the node of a
	// (symbol, origin), on the first of its facts.
	facts []fact
	at    []int
	node  []*Node
	// items is every item by origin, those from node i being
	// items[from[i]:from[i+1]], in node order: where a rule position from i
	// stands, which is how a split is checked against the chart.
	items []placed
	from  []int
	// core is each rule's first core item id: rule r with its dot at m is
	// core item core[r]+m.
	core []int
	// leaf is each token's leaves, one per terminal of its set, edge k's from
	// leafAt[k]; semiLeaf is the zero-width ; at each inserted node.
	leaf     []*Node
	leafAt   []int
	semiLeaf map[int]*Node
	pre      []*Node       // the prefix node kept beside each item
	prefixes map[int]*Node // the others
	maxDot   int
	// leo is each symbol some Leo item awaits, numbered from 1; memo is the
	// Leo memo at node at for the symbol numbered x, at memo[at*nleo+x-1]:
	// 0 not yet looked up, -1 none, else 1 + its index in links.
	leo   []int
	nleo  int
	memo  []int32
	links []link
	slab  []Node
	alts  []Alt
	work  []int // facts whose node waits for its Alts
	size  int
	steps int
}

// fact is a rule the chart completed for sym from from, small to keep the
// table compact.
type fact struct{ sym, from, rule int32 }

// mk is the fact of rule completing sym from from.
func mk(sym, from, rule int) fact { return fact{int32(sym), int32(from), int32(rule)} }

// placed is an item at node at, by its state.
type placed struct{ at, state int32 }

// link is a Leo memo's first link (recognize.Chart.LeoLink), the next
// link's and the chain's last's (the top's) indexes in links, next -1 where
// there is none, and the last node whose facts walked it.
type link struct {
	rule, origin int
	next, top    int
	walked       int
}

// cmpFact orders facts by symbol, origin and rule.
func cmpFact(x, y fact) int {
	if x.sym != y.sym {
		return int(x.sym - y.sym)
	}
	if x.from != y.from {
		return int(x.from - y.from)
	}
	return int(x.rule - y.rule)
}

// build is the forest of c, a chart of l that accepts; nil where l is not linear.
// shortcut: l must be linear, as up-front lexing makes it, so the token
// ending at node j is edge j-1 and an inserted ; is followed only by its };
// index the edges by node, and keep a ;'s facts apart, when interleaved
// lexing branches (TestForestLinearLatticesOnly).
func build(tab *grammar.Tables, c *recognize.Chart, l *lex.Lattice[lex.None]) *Forest {
	n := len(c.Sets)
	for k, e := range l.Edges {
		if e.From != k || e.To != k+1 {
			return nil
		}
	}
	b := &builder{tab: tab, c: c, l: l, inserted: make([]bool, n), at: make([]int, n+1), from: make([]int, n+1),
		leafAt: make([]int, len(l.Edges)+1), semiLeaf: map[int]*Node{}, prefixes: map[int]*Node{},
		leo: make([]int, len(tab.Symbols)), core: make([]int, len(tab.Rules))}
	b.semi = slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Terminal && s.Name == "PERLY_SEMICOLON" })
	for r, rule := range tab.Rules {
		b.maxDot = max(b.maxDot, len(rule.RHS)+1)
		if r+1 < len(tab.Rules) {
			b.core[r+1] = b.core[r] + len(rule.RHS) + 1
		}
	}
	for _, st := range tab.States {
		for _, i := range st.Leo {
			it := tab.Items[i]
			if sym := tab.Rules[it.Rule].RHS[it.Dot]; b.leo[sym] == 0 {
				b.nleo++
				b.leo[sym] = b.nleo
			}
		}
	}
	b.memo = make([]int32, n*b.nleo)
	for _, p := range c.Inserted() {
		b.inserted[p] = true
	}
	slipped := make([][]recognize.Item, n)
	for _, p := range c.Inserted() {
		next := -1
		if p < len(l.Edges) {
			next = l.Edges[p].Token.TerminalSet
		}
		slipped[p] = c.Slipped(p, next)
	}
	for j, s := range c.Sets { // count each origin's items, then sum the counts into run starts
		for _, items := range [2][]recognize.Item{s.Items(), slipped[j]} {
			for _, it := range items {
				b.from[j-it.Dist+1]++
			}
		}
	}
	for i := range n {
		b.from[i+1] += b.from[i]
	}
	size := b.from[n]
	b.items, b.pre = make([]placed, size), make([]*Node, size)
	fill := slices.Clone(b.from)
	b.facts = make([]fact, 0, size)
	for j, s := range c.Sets {
		b.at[j] = len(b.facts)
		for _, items := range [2][]recognize.Item{s.Items(), slipped[j]} {
			for _, it := range items {
				o := j - it.Dist
				b.items[fill[o]] = placed{int32(j), int32(it.State)}
				fill[o]++
				for _, r := range tab.States[it.State].Reduces {
					b.facts = append(b.facts, mk(tab.Rules[r].LHS, o, r))
				}
			}
		}
		b.leoFacts(j)
	}
	b.at[n] = len(b.facts)
	b.steps += len(b.facts)
	b.node = make([]*Node, len(b.facts))
	for k, e := range l.Edges {
		b.leafAt[k+1] = b.leafAt[k] + len(tab.TerminalSets[e.Token.TerminalSet])
	}
	b.leaf = make([]*Node, b.leafAt[len(l.Edges)])
	root := b.child(tab.Start, 0, n-1)
	for len(b.work) > 0 {
		i := b.work[len(b.work)-1]
		b.work = b.work[:len(b.work)-1]
		nd, f := b.node[i], b.facts[i]
		for ; i < b.at[nd.To+1] && b.facts[i].sym == f.sym && b.facts[i].from == f.from; i++ {
			r := int(b.facts[i].rule)
			b.split(nd, r, len(tab.Rules[r].RHS))
		}
	}
	return &Forest{Root: root, size: b.size, steps: b.steps}
}

// leoFacts sorts node j's facts and adds the completions a Leo memo skipped
// there: the links of each chain whose top is among j's facts, which is
// where recognition took the chain. A chain whose top is absent was not
// taken at j, and its links, where they hold, are items of their own.
func (b *builder) leoFacts(j int) {
	seg := b.facts[b.at[j]:]
	slices.SortFunc(seg, cmpFact)
	seg = slices.Compact(seg)
	b.facts = b.facts[:b.at[j]+len(seg)]
	for _, f := range seg {
		if b.leo[f.sym] == 0 {
			continue
		}
		l := b.link(int(f.from), int(f.sym))
		if l < 0 {
			continue
		}
		top := b.links[b.links[l].top]
		if _, ok := slices.BinarySearchFunc(seg, mk(b.tab.Rules[top.rule].LHS, top.origin, top.rule), cmpFact); !ok {
			continue
		}
		for ; l >= 0 && b.links[l].walked != j; l = b.links[l].next {
			b.links[l].walked = j
			b.facts = append(b.facts, mk(b.tab.Rules[b.links[l].rule].LHS, b.links[l].origin, b.links[l].rule))
		}
	}
	if len(b.facts) > b.at[j]+len(seg) {
		seg = b.facts[b.at[j]:]
		slices.SortFunc(seg, cmpFact)
		b.facts = b.facts[:b.at[j]+len(slices.Compact(seg))]
	}
}

// link is the index in links of the Leo memo at node at for sym, made
// once; -1 where none applies.
func (b *builder) link(at, sym int) int {
	if b.leo[sym] == 0 {
		return -1 // no Leo item awaits sym, so no memo is made for it
	}
	k := at*b.nleo + b.leo[sym] - 1
	if b.memo[k] != 0 {
		return int(b.memo[k]) - 1
	}
	b.memo[k] = -1 // none, unless LeoLink finds one
	rule, origin, more, ok := b.c.LeoLink(at, sym)
	if !ok {
		return -1
	}
	l := len(b.links)
	b.links = append(b.links, link{rule: rule, origin: origin, next: -1, top: l, walked: -1})
	b.memo[k] = int32(l + 1)
	if more {
		if next := b.link(origin, b.tab.Rules[rule].LHS); next >= 0 {
			b.links[l].next, b.links[l].top = next, b.links[next].top
		}
	}
	return l
}

// find is the index of the first fact of sym from from to to, and whether there is one.
func (b *builder) find(sym, from, to int) (int, bool) {
	seg := b.facts[b.at[to]:b.at[to+1]]
	i, _ := slices.BinarySearchFunc(seg, mk(sym, from, -1), cmpFact) // no rule is -1: i is where its facts begin
	i += b.at[to]
	return i, i < b.at[to+1] && int(b.facts[i].sym) == sym && int(b.facts[i].from) == from
}

// item is the index in items of an item at node k from node i whose state
// holds rule r with its dot at m, or -1 where the chart has none.
func (b *builder) item(r, m, i, k int) int {
	run := b.items[b.from[i]:b.from[i+1]]
	core := b.core[r] + m
	for x, _ := slices.BinarySearchFunc(run, k, func(p placed, k int) int { return int(p.at) - k }); x < len(run) && int(run[x].at) == k; x++ {
		if _, ok := slices.BinarySearch(b.tab.States[int(run[x].state)].Items, core); ok {
			return b.from[i] + x
		}
	}
	return -1
}

// alloc is a fresh node, from a slab.
func (b *builder) alloc(n Node) *Node {
	if len(b.slab) == 0 {
		b.slab = make([]Node, 1024)
	}
	p := &b.slab[0]
	b.slab = b.slab[1:]
	*p = n
	b.size++
	return p
}

// child is sym's node over from to to, made once, or nil where the chart
// does not prove sym derives there: a completed fact, a token, or a ;
// inserted at the node.
func (b *builder) child(sym, from, to int) *Node {
	if !b.tab.Symbols[sym].Terminal {
		if i, ok := b.find(sym, from, to); ok {
			return b.factNode(i, to)
		}
		return nil
	}
	if from == to {
		if sym != b.semi || !b.inserted[to] {
			return nil
		}
		if b.semiLeaf[to] == nil {
			b.semiLeaf[to] = b.alloc(Node{Sym: sym, Rule: -1, From: from, To: to})
		}
		return b.semiLeaf[to]
	}
	if to != from+1 {
		return nil
	}
	i, ok := slices.BinarySearch(b.tab.TerminalSets[b.l.Edges[from].Token.TerminalSet], sym)
	if !ok {
		return nil
	}
	slot := &b.leaf[b.leafAt[from]+i]
	if *slot == nil {
		*slot = b.alloc(Node{Sym: sym, Rule: -1, From: from, To: to})
	}
	return *slot
}

// factNode is the node of fact i, which ends at node to, made once; it waits for its Alts.
func (b *builder) factNode(i, to int) *Node {
	if b.node[i] == nil {
		f := b.facts[i]
		b.node[i] = b.alloc(Node{Sym: int(f.sym), Rule: -1, From: int(f.from), To: to})
		b.work = append(b.work, i)
	}
	return b.node[i]
}

// prefix is the node of rule r's first m symbols, m at least 2, over from
// to to, made once, or nil where the chart holds no item for them there. It
// is kept beside that item, in a map where the item already keeps another.
func (b *builder) prefix(r, m, from, to int) *Node {
	x := b.item(r, m, from, to)
	if x < 0 {
		return nil
	}
	if n := b.pre[x]; n != nil && n.Rule == r && n.Dot == m {
		return n
	}
	k := x*len(b.tab.Items) + b.core[r] + m
	if n := b.prefixes[k]; n != nil {
		return n
	}
	n := b.alloc(Node{Sym: -1, Rule: r, Dot: m, From: from, To: to})
	if b.pre[x] == nil {
		b.pre[x] = n
	} else {
		b.prefixes[k] = n
	}
	b.split(n, r, m)
	return n
}

// split adds to n an Alt for each way rule r's first m symbols cover its
// span, by where the m-th begins: a token's start, the node itself for a ;
// inserted there, or an origin the chart completed it from. The starts are
// read from whichever is shorter, the m-th symbol's completions at n.To or
// the nodes where an item from n.From stands; alt checks the first m-1.
func (b *builder) split(n *Node, r, m int) {
	if m == 0 {
		if n.From == n.To {
			b.add(n, Alt{Rule: r})
		}
		return
	}
	x := b.tab.Rules[r].RHS[m-1]
	switch {
	case b.tab.Symbols[x].Terminal:
		for _, k := range [2]int{n.To - 1, n.To} {
			if k >= n.From {
				b.alt(n, r, m, k, b.child(x, k, n.To))
			}
		}
		return
	case m == 1:
		b.alt(n, r, m, n.From, b.child(x, n.From, n.To))
		return
	}
	lo, _ := b.find(x, n.From, n.To)
	hi, _ := b.find(x+1, -1, n.To)
	run := b.items[b.from[n.From]:b.from[n.From+1]]
	end, _ := slices.BinarySearchFunc(run, n.To+1, func(p placed, k int) int { return int(p.at) - k })
	if hi-lo <= end {
		for i := lo; i < hi; i++ {
			if i == lo || b.facts[i-1].from != b.facts[i].from {
				b.alt(n, r, m, int(b.facts[i].from), b.factNode(i, n.To))
			}
		}
		return
	}
	for i, p := range run[:end] {
		if i == 0 || run[i-1].at != p.at {
			b.alt(n, r, m, int(p.at), b.child(x, int(p.at), n.To))
		}
	}
}

// alt adds to n the Alt of rule r whose m-th symbol is right, from k, where
// the chart proves right (non-nil) and the first m-1 symbols from n.From to
// k, as a symbol it completed or as an item standing at k. An Alt whose
// child is n itself is dropped: no derivation can use it.
func (b *builder) alt(n *Node, r, m, k int, right *Node) {
	b.steps++
	if right == nil || right == n {
		return
	}
	var left *Node
	switch {
	case m == 1:
		if k != n.From {
			return
		}
	case m == 2:
		left = b.child(b.tab.Rules[r].RHS[0], n.From, k)
	default:
		left = b.prefix(r, m-1, n.From, k)
	}
	if m > 1 && left == nil || left == n {
		return
	}
	b.add(n, Alt{Rule: r, Left: left, Right: right})
}

// add appends a to n's Alts; most nodes have one, so the first comes from a slab.
func (b *builder) add(n *Node, a Alt) {
	if n.Alts != nil {
		n.Alts = append(n.Alts, a)
		return
	}
	if len(b.alts) == 0 {
		b.alts = make([]Alt, 1024)
	}
	b.alts[0] = a
	n.Alts, b.alts = b.alts[:1:1], b.alts[1:]
}

// Derivations is at most limit of n's derivations, and whether a bound cut
// them short: the limit, or a budget of steps proportional to limit and the
// forest's size, so a forest whose cycles hide exponentially many dead ends
// still returns. A derivation never holds a node inside itself, so a cycle
// adds none, and a slipper ; appears only where a derivation needs it, never
// as an extra null statement beside one that does not.
func (f *Forest) Derivations(n *Node, limit int) (trees []*Tree, cut bool) {
	e := &enum{path: map[*Node]bool{}, budget: 64 * max(limit, 1) * (f.size + 1)}
	trees = e.trees(n, limit)
	return trees, e.budget < 0 || len(trees) >= limit
}

// enum is one enumeration: the symbol nodes on the current path, and the steps left.
type enum struct {
	path   map[*Node]bool
	budget int
}

// trees is at most limit derivations of symbol node n that avoid every node on the path.
func (e *enum) trees(n *Node, limit int) []*Tree {
	if e.budget--; limit <= 0 || e.budget < 0 {
		return nil
	}
	if n.Alts == nil {
		return []*Tree{{Sym: n.Sym, Rule: -1, From: n.From, To: n.To}}
	}
	if e.path[n] {
		return nil
	}
	e.path[n] = true
	defer delete(e.path, n)
	var out []*Tree
	for _, a := range n.Alts {
		for _, kids := range e.seqs(a, limit-len(out)) {
			out = append(out, &Tree{Sym: n.Sym, Rule: a.Rule, From: n.From, To: n.To, Kids: kids})
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

// seqs is at most limit of the children sequences a derives.
func (e *enum) seqs(a Alt, limit int) [][]*Tree {
	if e.budget--; limit <= 0 || e.budget < 0 {
		return nil
	}
	var lefts [][]*Tree
	switch {
	case a.Left == nil:
		lefts = [][]*Tree{nil}
	case a.Left.Sym >= 0:
		for _, t := range e.trees(a.Left, limit) {
			lefts = append(lefts, []*Tree{t})
		}
	default:
		for _, la := range a.Left.Alts {
			lefts = append(lefts, e.seqs(la, limit-len(lefts))...)
			if len(lefts) >= limit {
				break
			}
		}
	}
	if a.Right == nil {
		return lefts
	}
	var out [][]*Tree
	for _, l := range lefts {
		for _, r := range e.trees(a.Right, limit-len(out)) {
			out = append(out, append(slices.Clip(l), r))
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}
