// ABOUTME: The position map (§6.1.6 item 3): a side table from lowered node ids to CST spans, and from offsets back.
// ABOUTME: Bytes the lowering erased -- parens, trivia, an Unknown's children -- resolve to the lowered node enclosing them.

package lower

import "tamarou.com/pvm/internal/parse"

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
}

// NewPositions builds the map for f, which Lower(root) returned.
func NewPositions(root *parse.Node, f *File) *Positions {
	p := &Positions{root: root, ids: AssignIDs(root), nodes: map[ID]Node{}}
	var walk func(n Node)
	walk = func(n Node) {
		p.nodes[n.Meta().ID] = n
		for _, c := range Children(n) {
			walk(c)
		}
	}
	walk(f)
	return p
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
