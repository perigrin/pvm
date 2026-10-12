// ABOUTME: Lowering of slices to Slice{Kind} (§4.14.1 rule 5): the five-way switch on the base's sigil and the bracket.
// ABOUTME: The kind is two field reads off the CST, never re-derived from source text.

package lower

import "tamarou.com/pvm/internal/parse"

// slice lowers an `index` that is a slice to a Slice, and declines an
// element access. A list base -- (1,2)[0], ($x)[0] -- with no
// arrow is a SliceList; otherwise the base's sigil and the bracket select
// the kind: @ over [ or {, % over [ or {. Idx is the subscript as lowered,
// its comma tree intact.
func (l *lowerer) slice(n *parse.Node) Node {
	if len(n.Children) != 2 {
		return nil
	}
	kind, ok := sliceKindOf(n)
	if !ok {
		return nil
	}
	return &Slice{Base: l.expr(n.Children[0]), Idx: l.expr(n.Children[1]), Kind: kind}
}

// sliceKindOf returns the kind of slice the index n is, or false if it is
// not one.
func sliceKindOf(n *parse.Node) (SliceKind, bool) {
	base := n.Children[0]
	if !n.Arrow && (base.Paren || base.Kind == parse.List) {
		return SliceList, true
	}
	if (base.Kind != parse.Term && base.Kind != parse.Unary) || base.Text == "" {
		return 0, false
	}
	switch base.Text[0:1] + n.Text {
	case "@[":
		return SliceArray, true
	case "@{":
		return SliceHash, true
	case "%[":
		return SliceKVArray, true
	case "%{":
		return SliceKVHash, true
	}
	return 0, false
}
