// ABOUTME: The position map (§6.1.6 item 3): a side table from lowered node ids to CST spans, and from offsets back.
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
	spans map[ID]span // every CST node's span, by its id
	order []Node      // every lowered node, by start offset
}

// span is a CST node's bytes, [start, end).
type span struct{ start, end int }

// NewPositions builds the map for f, which Lower(root) returned.
func NewPositions(root *parse.Node, f *File) *Positions {
	p := &Positions{root: root, ids: AssignIDs(root), nodes: map[ID]Node{}, spans: map[ID]span{}}
	for n, id := range p.ids {
		p.spans[id] = span{n.Start, n.End}
	}
	var walk func(n Node)
	walk = func(n Node) {
		p.nodes[n.Meta().ID] = n
		p.order = append(p.order, n)
		for _, c := range Children(n) {
			walk(c)
		}
	}
	walk(f)
	// Stable, so a parent stays before a child that starts where it does.
	sort.SliceStable(p.order, func(i, j int) bool { return p.start(p.order[i]) < p.start(p.order[j]) })
	return p
}

// start is where n's bytes begin. A node with no parse node of its own
// starts where its parent's parse node does.
func (p *Positions) start(n Node) int {
	id := n.Meta().ID
	id.Role = 0
	return p.spans[id].start
}

// LastBefore returns the lowered node keep accepts that starts last before
// byte off, in source order, or nil if none does: the question $1 asks of
// the matches before it (nearestPatternBefore, infer.go:1726).
//
// ponytail: a linear scan; binary-search order on off if a caller asks often.
func (p *Positions) LastBefore(off int, keep func(Node) bool) Node {
	for i := len(p.order) - 1; i >= 0; i-- {
		if n := p.order[i]; p.start(n) < off && keep(n) {
			return n
		}
	}
	return nil
}

// At returns the innermost lowered node whose CST node covers byte off, or
// nil if none does.
func (p *Positions) At(off int) Node {
	var out Node
	for n := p.root; n != nil; {
		if off < n.Start || off >= n.End {
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
