// ABOUTME: Lowering of slices to Slice{Kind} (§4.14.1 rule 5): the five-way switch on the base's sigil and the bracket.
// ABOUTME: The kind is two field reads off the CST, never re-derived from source text.

package lower

import "tamarou.com/pvm/internal/parse"

// slice lowers an `index` that is a slice to a Slice, and declines an
// element access. A list base -- (1,2)[0], ($x)[0], qw(a b)[1] -- with no
// arrow is a SliceList; otherwise the base's sigil and the bracket select
// the kind: @ over [ or {, % over [ or {. Idx is the subscript as lowered,
// its comma tree intact, and a lone word in braces is quoted as in $h{k}.
func (l *lowerer) slice(n *parse.Node) Node {
	if len(n.Children) != 2 {
		return nil
	}
	kind, ok := sliceKindOf(n)
	if !ok {
		return nil
	}
	return &Slice{Base: l.expr(n.Children[0]), Idx: l.sliceIdx(n), Kind: kind}
}

// postfixSlice lowers `binary "->"` whose right side is an index over a
// bare "@" or "%" term -- $x->@[0,1], $x->%{a} -- to a Slice whose base is
// the postfix Deref of the arrow's left side, as perl -MO=Deparse prints
// $x->@[0,1] as @$x[0, 1]. The Deref carries the bare term's id. It
// declines every other arrow.
func (l *lowerer) postfixSlice(n *parse.Node) Node {
	if n.Text != "->" || len(n.Children) != 2 {
		return nil
	}
	r := n.Children[1]
	if r.Kind != parse.Index || r.Arrow || len(r.Children) != 2 {
		return nil
	}
	sigil := r.Children[0]
	if sigil.Kind != parse.Term || len(sigil.Children) != 0 || (sigil.Text != "@" && sigil.Text != "%") {
		return nil
	}
	kind, ok := sliceKindOf(r)
	if !ok {
		return nil
	}
	base := l.stamp(sigil, &Deref{Sigil: sigil.Text, Expr: l.expr(n.Children[0]), Postfix: true})
	return &Slice{Base: base, Idx: l.sliceIdx(r), Kind: kind}
}

// sliceIdx lowers the subscript of the slice index n.
func (l *lowerer) sliceIdx(n *parse.Node) Node {
	if n.Text == "{" {
		return l.hashKey(n.Children[1])
	}
	return l.expr(n.Children[1])
}

// sliceKindOf returns the kind of slice the index n is, or false if it is
// not one.
func sliceKindOf(n *parse.Node) (SliceKind, bool) {
	base := n.Children[0]
	if !n.Arrow && (base.Paren || base.Kind == parse.List || isQw(base)) {
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

// isQw reports whether n is a qw list, which perl subscripts as a list:
// perl -MO=Deparse prints qw(a b)[1] as ('a', 'b')[1].
func isQw(n *parse.Node) bool {
	if n.Kind != parse.Term || len(n.Children) != 0 {
		return false
	}
	m := quoteOp.FindStringSubmatch(n.Text)
	return m != nil && m[1] == "qw"
}
