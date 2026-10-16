// ABOUTME: The position map (§6.1.6 item 3): a side table from lowered node ids to CST spans and quote tokens, and from offsets back.
// ABOUTME: Bytes the lowering erased -- parens, trivia, an Unknown's children -- resolve to the lowered node enclosing them.

package lower

import (
	"sort"

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
type Positions struct {
	root  *parse.Node
	ids   map[*parse.Node]ID
	nodes map[ID]Node
	spans map[ID]span        // every CST node's span, by its id
	cst   map[ID]*parse.Node // every CST node, by its id
	order []Node             // every lowered node, by start offset
}

// span is a CST node's bytes, [start, end).
type span struct{ start, end int }

// NewPositions builds the map for f, which Lower(root) returned.
func NewPositions(root *parse.Node, f *File) *Positions {
	p := &Positions{root: root, ids: AssignIDs(root), nodes: map[ID]Node{}, spans: map[ID]span{}, cst: map[ID]*parse.Node{}}
	for n, id := range p.ids {
		p.spans[id] = span{n.Start, n.End}
		p.cst[id] = n
	}
	p.spans[p.ids[root]] = statements(root)
	var walk func(n Node)
	walk = func(n Node) {
		p.nodes[n.Meta().ID] = n
		if s := p.span(n.Meta().ID); s.start < s.end {
			p.order = append(p.order, n)
		}
		for _, c := range Children(n) {
			walk(c)
		}
	}
	walk(f)
	// Stable, so a parent stays before a child that starts where it does.
	sort.SliceStable(p.order, func(i, j int) bool {
		return p.span(p.order[i].Meta().ID).start < p.span(p.order[j].Meta().ID).start
	})
	return p
}

// statements is the File's span: from its first statement's start to its
// last one's end. Trivia between statements is the File's; trivia before
// the first or after the last is nobody's, and an empty file has no bytes.
func statements(root *parse.Node) span {
	var s span
	first := true
	for _, c := range root.Children {
		if c.Kind == parse.Trivia {
			continue
		}
		if first {
			s.start, first = c.Start, false
		}
		s.end = c.End
	}
	return s
}

// span is the bytes of the CST node id names. A lowered node with no parse
// node of its own has its parent's.
func (p *Positions) span(id ID) span {
	id.Role = 0
	return p.spans[id]
}

// Span returns the bytes [start, end) of the lowered node id names, and
// false if the lowered tree has no node with that id.
func (p *Positions) Span(id ID) (start, end int, ok bool) {
	if _, ok := p.nodes[id]; !ok {
		return 0, 0, false
	}
	s := p.span(id)
	return s.start, s.end, true
}

// LastBefore returns the lowered node keep accepts that starts last before
// byte off, in source order, or nil if none does: the question $1 asks of
// the matches before it (nearestPatternBefore, infer.go:1726). A node with
// no bytes, an empty file's File, is never before anything.
//
// order is sorted by start, so a binary search finds the last node that
// starts before off, and the scan back from there stops at the first node
// keep accepts. $1 asks once per capture, and its match is near.
//
// ponytail: the scan back is linear in the nodes between off and the
// answer, so many captures far below one match are still quadratic; keep a
// per-predicate index if a caller ever asks that way.
func (p *Positions) LastBefore(off int, keep func(Node) bool) Node {
	i := sort.Search(len(p.order), func(i int) bool {
		return p.span(p.order[i].Meta().ID).start >= off
	})
	for i--; i >= 0; i-- {
		if n := p.order[i]; keep(n) {
			return n
		}
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
	n := p.cst[id]
	if n == nil || id.Role != 0 {
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
	var out Node
	for n := p.root; n != nil; {
		if s := p.spans[p.ids[n]]; off < s.start || off >= s.end {
			break
		}
		if ln, ok := p.nodes[p.ids[n]]; ok {
			out = ln
		}
		var next *parse.Node
		for _, c := range n.Children {
			if off >= c.Start && off < c.End {
				next = c
				break
			}
		}
		n = next
	}
	return out
}
