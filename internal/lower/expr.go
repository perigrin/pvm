// ABOUTME: The expression-lowering dispatcher: one arm list per parse.Kind, a Hole for whatever no arm claims.
// ABOUTME: Each slice of §4.14.2 adds its arm functions to the lists here and keeps their bodies in its own file.

package lower

import "tamarou.com/pvm/internal/parse"

// lowerer carries what every lowered node is stamped with: its id from
// AssignIDs.
type lowerer struct {
	ids map[*parse.Node]ID
}

// arm lowers a node of the kind it is listed under, or returns nil to
// decline it so the next arm in the list may try.
type arm func(*parse.Node) Node

// expr lowers one expression node. A kind with no arm, an arm list that
// declines, and every Unknown give a Hole (§4.14.2, Hole row): opaque, its
// children never lowered. Whatever is returned carries the node's id and
// Paren bit, so no arm sets either.
//
// Where two slices share a kind, the narrower arm comes first: a variable's
// term (sigil-led) before a literal's, a Deref's unary ("$" "@" "%") before
// an operator's.
func (l *lowerer) expr(n *parse.Node) Node {
	var out Node
	switch n.Kind {
	case parse.Term:
		out = first(n, l.variable, l.literal, l.angle, l.match)
	case parse.Call:
		out = first(n, l.fatWord, l.mapGrepSort, l.evalForm, l.doForm, l.wantarray, l.fileTest, l.globCall)
	case parse.Unary:
		out = first(n, l.prefixDeref, l.ref)
	case parse.Binary:
		out = first(n, l.postfixDeref, l.postfixSlice, l.bind)
	case parse.Index:
		out = first(n, l.slice, l.subscript)
	case parse.AnonArray:
		out = first(n, l.arrayLiteral)
	case parse.AnonHash:
		out = first(n, l.hashLiteral)
	case parse.LoopControl:
		out = first(n, l.loopEx)
	case parse.Declaration:
		out = first(n, l.decl)
	}
	if out == nil {
		out = &Hole{}
	}
	return l.stamp(n, out)
}

// stamp gives out the id and Paren bit of n, the parse node it lowers.
func (l *lowerer) stamp(n *parse.Node, out Node) Node {
	b := out.Meta()
	b.ID, b.Paren = l.ids[n], n.Paren
	return out
}

// first returns what the first arm to claim n gives, or nil if none does.
func first(n *parse.Node, arms ...arm) Node {
	for _, a := range arms {
		if out := a(n); out != nil {
			return out
		}
	}
	return nil
}
