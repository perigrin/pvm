// ABOUTME: The position map (§6.1.6 item 3): a side table from lowered node ids to CST spans and quote tokens, and from offsets back.
// ABOUTME: Kept per top-level statement, so an edit renumbers only the statements it changed (Relower) and shares the rest.

package lower

import (
	"slices"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// Positions connects a lowered File to the CST it came from, in both
// directions, so the lowered tree itself carries no spans and no syntax
// pointers (§6.1.6 item 4).
//
// The map is many-CST-to-one-lowered: a CST node the lowering erased has no
// lowered node of its own, and its bytes resolve to the nearest lowered node
// whose CST node encloses them. A paren is a bit on the node it wrapped, so
// its bytes resolve to that node; a comment resolves to the Block or File
// around it; an Unknown's children resolve to its Hole.
//
// It is held as one chunk per top-level statement, each chunk in the
// offsets of the tree it was numbered in, so an edit's map (Relower) shares
// every chunk the edit only moved and reads where it moved to off root.
type Positions struct {
	root   *parse.Node
	file   *File
	chunks []*chunk // one per child of root, in order
	// his[i] is one past the last "" Local of chunks[0..i]: a top-level
	// statement's ids outside any named sub are a run of the File's.
	his []int
	// subs is the chunk holding each top-level sub, by anchor.
	subs map[string]int
	// renumbered is, for a map Relower made, the indexes of the chunks it
	// numbered again, in order; numbered and indexed are the CST nodes
	// this map numbered and the lowered nodes it indexed itself, rather
	// than shared from the map it was made from.
	renumbered        []int
	numbered, indexed int
}

// chunk is the map of one top-level statement, node, as the tree it was
// numbered in holds it: its spans are that tree's, moved by however far
// the statement has moved since.
type chunk struct {
	node  *parse.Node
	ids   map[*parse.Node]ID
	cst   map[ID]*parse.Node // every CST node, by its id
	nodes map[ID]Node        // every lowered node, by its id
	order []Node             // every lowered node with bytes, by start
	subs  []string           // the top-level subs in it, by anchor, sorted
}

// span is a CST node's bytes, [start, end).
type span struct{ start, end int }

// NewPositions builds the map for f, which Lower(root) returned.
func NewPositions(root *parse.Node, f *File) *Positions {
	p := &Positions{root: root, file: f, subs: map[string]int{}}
	next, seen := map[string]int{"": 1}, map[string]int{}
	for i, c := range root.Children {
		ch := p.number(c, next, seen)
		for _, a := range ch.subs {
			p.subs[a] = i
		}
		p.chunks = append(p.chunks, ch)
		p.his = append(p.his, next[""])
	}
	for _, s := range f.Stmts {
		if i := p.chunkOf(s.Meta().ID); i >= 0 && i < len(p.chunks) {
			p.index(p.chunks[i], s)
		}
	}
	for _, c := range p.chunks {
		c.sort()
	}
	return p
}

// Renumbered is, for a map Relower made, the indexes among root's
// children of the top-level statements it numbered again, in order: those
// holding a damaged sub or the edit. Every other is a statement of the
// tree before the edit, moved whole. Nil for a map NewPositions made.
func (p *Positions) Renumbered() []int {
	return p.renumbered
}

// Work is the CST nodes the map numbered and the lowered nodes it
// indexed itself, rather than shared from the map Relower made it from.
func (p *Positions) Work() (numbered, indexed int) {
	return p.numbered, p.indexed
}

// number numbers the top-level statement n as AssignIDs does, next and
// seen being AssignIDs' state before it, and returns its chunk, not yet
// indexed.
func (p *Positions) number(n *parse.Node, next, seen map[string]int) *chunk {
	c := &chunk{node: n, ids: map[*parse.Node]ID{}, nodes: map[ID]Node{}}
	assignIDsFrom(c.ids, n, "", next, seen)
	c.cst = make(map[ID]*parse.Node, len(c.ids))
	for n, id := range c.ids {
		c.cst[id] = n
		if id.Local == 0 && id.Anchor != "" && !strings.Contains(id.Anchor, "/") {
			c.subs = append(c.subs, id.Anchor)
		}
	}
	slices.Sort(c.subs)
	p.numbered += len(c.ids)
	return c
}

// index puts n and every lowered node under it in c.
func (p *Positions) index(c *chunk, n Node) {
	p.indexed++
	c.nodes[n.Meta().ID] = n
	if s := c.span(n.Meta().ID); s.start < s.end {
		c.order = append(c.order, n)
	}
	for _, k := range Children(n) {
		p.index(c, k)
	}
}

// sort orders c.order by start; stable, so a parent stays before a child
// that starts where it does.
func (c *chunk) sort() {
	sort.SliceStable(c.order, func(i, j int) bool {
		return c.span(c.order[i].Meta().ID).start < c.span(c.order[j].Meta().ID).start
	})
}

// span is the bytes of the CST node id names, in the chunk's own tree. A
// lowered node with no parse node of its own has its parent's.
func (c *chunk) span(id ID) span {
	id.Role = 0
	if n := c.cst[id]; n != nil {
		return span{n.Start, n.End}
	}
	return span{}
}

// chunkOf is the index of the chunk id is in: -1 for the root's, and
// len(p.chunks) for an id of no chunk.
func (p *Positions) chunkOf(id ID) int {
	if id.Anchor == "" {
		if id.Local == 0 {
			return -1
		}
		return sort.Search(len(p.his), func(i int) bool { return p.his[i] > id.Local })
	}
	top, _, _ := strings.Cut(id.Anchor, "/")
	if i, ok := p.subs[top]; ok {
		return i
	}
	return len(p.chunks)
}

// shift is how far the top-level statement chunks[i] holds has moved
// since its chunk was numbered.
func (p *Positions) shift(i int) int {
	return p.root.Children[i].Start - p.chunks[i].node.Start
}

// lookup is the lowered node id names and the bytes of its CST node, in
// root's offsets, and false if the lowered tree has no node with that id.
func (p *Positions) lookup(id ID) (Node, span, bool) {
	if id == (ID{}) {
		return p.file, statements(p.root), true
	}
	i := p.chunkOf(id)
	if i < 0 || i >= len(p.chunks) {
		return nil, span{}, false
	}
	c := p.chunks[i]
	n, ok := c.nodes[id]
	if !ok {
		return nil, span{}, false
	}
	s, d := c.span(id), p.shift(i)
	return n, span{s.start + d, s.end + d}, true
}

// statements is the File's span: from its first statement's start to its
// last one's end. Trivia between statements is the File's; trivia before
// the first or after the last is nobody's, and an empty file has no bytes.
//
// It scans in from both ends, so it reads only the trivia around the
// statements: At and LastBefore ask it on every call.
func statements(root *parse.Node) span {
	kids := root.Children
	i, j := 0, len(kids)
	for i < j && kids[i].Kind == parse.Trivia {
		i++
	}
	for j > i && kids[j-1].Kind == parse.Trivia {
		j--
	}
	if i == j {
		return span{}
	}
	return span{kids[i].Start, kids[j-1].End}
}

// Span returns the bytes [start, end) of the lowered node id names, and
// false if the lowered tree has no node with that id.
func (p *Positions) Span(id ID) (start, end int, ok bool) {
	_, s, ok := p.lookup(id)
	return s.start, s.end, ok
}

// LastBefore returns the lowered node keep accepts that starts last before
// byte off, in source order, or nil if none does: the question $1 asks of
// the matches before it (lastMatch, internal/infer/special.go). A node with
// no bytes, an empty file's File, is never before anything.
//
// Each chunk's order is sorted by start, so a binary search finds the last
// node that starts before off, and the scan back from there stops at the
// first node keep accepts. $1 asks once per capture, and its match is near.
//
// ponytail: the scan back is linear in the nodes between off and the
// answer, so many captures far below one match are still quadratic; keep a
// per-predicate index if a caller ever asks that way.
func (p *Positions) LastBefore(off int, keep func(Node) bool) Node {
	kids := p.root.Children
	i := sort.Search(len(kids), func(i int) bool { return kids[i].Start >= off })
	for i--; i >= 0; i-- {
		c, d := p.chunks[i], p.shift(i)
		j := sort.Search(len(c.order), func(j int) bool {
			return c.span(c.order[j].Meta().ID).start+d >= off
		})
		for j--; j >= 0; j-- {
			if n := c.order[j]; keep(n) {
				return n
			}
		}
	}
	if s := statements(p.root); s.start < s.end && s.start < off && keep(p.file) {
		return p.file
	}
	return nil
}

// Quote returns the quote token of the Match id names -- `/(\d+)/` for
// `/(\d+)/` and `$s =~ /(\d+)/` alike -- or "" if id names no Match.
//
// A Match's pattern stays inside its one quote token (§4.14.2, PARSER), so
// the lowered tree carries no text; this is where $1 reads the groups of the
// match before it (perigrin, 2026-10-14: the position map hands back the
// token, the lowered tree stays free of text). A bound Match takes the bind's
// id, and its token is the bind's right operand.
func (p *Positions) Quote(id ID) string {
	if id.Role != 0 {
		return ""
	}
	n := p.root
	if id != (ID{}) {
		i := p.chunkOf(id)
		if i < 0 || i >= len(p.chunks) {
			return ""
		}
		n = p.chunks[i].cst[id]
	}
	if n == nil {
		return ""
	}
	if (n.Text == "=~" || n.Text == "!~") && len(n.Children) == 2 {
		n = n.Children[1]
	}
	if len(n.Children) != 0 || matchOp(n.Text) == "" {
		return ""
	}
	return n.Text
}

// At returns the innermost lowered node whose CST node covers byte off, or
// nil if none does.
func (p *Positions) At(off int) Node {
	if s := statements(p.root); off < s.start || off >= s.end {
		return nil
	}
	var out Node = p.file
	kids := p.root.Children
	i := sort.Search(len(kids), func(i int) bool { return kids[i].End > off })
	if i == len(kids) || off < kids[i].Start {
		return out
	}
	c := p.chunks[i]
	off -= p.shift(i)
	for n := c.node; n != nil; {
		id := c.ids[n]
		if s := c.span(id); off < s.start || off >= s.end {
			break
		}
		if ln, ok := c.nodes[id]; ok {
			out = ln
		}
		var next *parse.Node
		for _, k := range n.Children {
			if off >= k.Start && off < k.End {
				next = k
				break
			}
		}
		n = next
	}
	return out
}
