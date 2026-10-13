// ABOUTME: The special-form rules of AnalyzeLowered: MapGrepSort, eval and do, Ref, Match, and the forms the legacy path left untyped.
// ABOUTME: Each is the legacy rule rewritten against the lowered kinds: map, grep and sort are told apart by Op, never by source text.

package infer

import (
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/types"
)

// mapGrepSort walks a map, grep or sort. Its own value is not typed, as on
// the legacy path; what it hands back is recorded by elemType when a
// declaration takes it.
func (a *loweredAnalysis) mapGrepSort(n *lower.MapGrepSort) types.Type {
	a.walkChildren(n)
	return types.Unknown
}

// mapGrepSortElemType is the legacy callElementType for map, grep and sort.
// sort reorders and grep selects, so their elements are the source's; map
// transforms, so its elements are what its block yields. perl 5.42: with
// @n = (3, 1), `map { $_ * 2 } @n` is (6, 2) and `map { "x$_" } @n` is
// ("x3", "x1"). map's expression form is not typed.
func (a *loweredAnalysis) mapGrepSortElemType(n *lower.MapGrepSort) types.Type {
	if n.Op == "map" {
		if n.Block == nil {
			return types.Unknown
		}
		return a.blockValue(n.Block)
	}
	switch len(n.List) {
	case 0:
		return types.Unknown
	case 1:
		return a.aggElemType(n.List[0])
	}
	return a.joinTypes(n.List)
}

// blockValue is the legacy blockResultType: a block's value is its last
// expression statement's. A block perl rejected is a Hole, whose value is
// Any, never a guess.
func (a *loweredAnalysis) blockValue(n lower.Node) types.Type {
	b, ok := n.(*lower.Block)
	if !ok {
		return a.typeOf(n)
	}
	t := types.Unknown
	for _, s := range b.Stmts {
		if es, ok := s.(*lower.ExprStmt); ok && es.X != nil {
			t = a.typeOf(es.X)
		}
	}
	return t
}
