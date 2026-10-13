// ABOUTME: The variable, subscript and slice rules of AnalyzeLowered: Var, Deref, Index, Key and Slice, and the element types a declaration records.
// ABOUTME: Each is the legacy rule rewritten against the lowered kinds: the sigil, the arrow and the slice kind are fields, never re-read from text.

package infer

import (
	"fmt"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/types"
)

// variable is the legacy sigil arms: a scalar, array or hash is the type
// the symbol table narrowed it to, else its sigil's type, and $#a is an
// Int. A glob and a &name are not typed.
//
// A capture variable such as $1 is read as any other scalar here: its type
// comes from the match before it, which this rule does not read.
func (a *loweredAnalysis) variable(v *lower.Var) types.Type {
	if v.Sigil == "$#" {
		return types.Int
	}
	return a.lookupVar(v.Sigil, v.Name)
}

// lookupVar is the legacy lookupNarrowedType: sigil+name's narrowed type,
// or the sigil's type when nothing narrowed it. Unknown for a sigil that
// names no scalar, array or hash.
func (a *loweredAnalysis) lookupVar(sigil, name string) types.Type {
	fallback, ok := sigilTypes[sigil]
	if !ok {
		return types.Unknown
	}
	if sym, ok := a.st.Lookup(sigil + name); ok && sym.Type != types.Unknown {
		return sym.Type
	}
	return fallback
}

// derefTypes is what a dereference yields: `@$r` and `$r->@*` the whole
// array, `%$r` and `$r->%*` the whole hash, and `$#$r` and `$#{$r}` the
// last index, an Int. Any other dereference, `$$r` included, is not typed.
var derefTypes = map[string]types.Type{"@": types.Array, "%": types.Hash, "$#": types.Int}

// index is the legacy array_element_expression arm: an element read, and
// the check on its index.
func (a *loweredAnalysis) index(n *lower.Index) types.Type {
	a.walkBase(n.Base, n.Arrow, "@")
	if n.Idx != nil {
		a.checkIndex(n.Idx, a.walk(n.Idx))
	}
	return a.element(containerName(n.Base, n.Arrow, "@"))
}

// key is the legacy hash_element_expression arm. A key is stringified, not
// numified, so it is not checked.
func (a *loweredAnalysis) key(n *lower.Key) types.Type {
	a.walkBase(n.Base, n.Arrow, "%")
	if n.Key != nil {
		a.walk(n.Key)
	}
	return a.element(containerName(n.Base, n.Arrow, "%"))
}

// slice walks a slice. The slice itself is not typed, as on the legacy
// path; what it holds is recorded by elemType when a declaration takes it.
func (a *loweredAnalysis) slice(n *lower.Slice) {
	if n.Kind == lower.SliceList {
		a.walkChildren(n)
		return
	}
	a.walkBase(n.Base, false, sliceSigil(n.Kind))
	if n.Idx != nil {
		a.walk(n.Idx)
	}
}

// sliceSigil is the sigil of the aggregate a slice reads: @a for @a[..]
// and %a[..], %h for @h{..} and %h{..}.
func sliceSigil(k lower.SliceKind) string {
	if k == lower.SliceHash || k == lower.SliceKVHash {
		return "%"
	}
	return "@"
}

// walkBase walks a subscript's base. A variable read without an arrow
// names the aggregate, not the scalar its sigil spells -- $a[0] reads @a
// -- so it is typed as that aggregate.
func (a *loweredAnalysis) walkBase(base lower.Node, arrow bool, aggSigil string) {
	if v, ok := base.(*lower.Var); ok && !arrow {
		if t := a.lookupVar(aggSigil, v.Name); t != types.Unknown {
			a.annotations[v.ID] = t
		}
		return
	}
	if base != nil {
		a.walk(base)
	}
}

// containerName is the variable whose elements a subscript reads: @a for
// $a[0], %h for $h{k}, and the scalar holding the reference for $r->[0],
// $$r[0] and @{$r}[0,1]. A subscript is a dereference when it has an arrow
// or its base is a Deref. "" when the base is not a variable, as in
// $h{a}[0].
func containerName(base lower.Node, arrow bool, aggSigil string) string {
	switch b := base.(type) {
	case *lower.Var:
		if !arrow {
			return aggSigil + b.Name
		}
		if b.Sigil == "$" {
			return "$" + b.Name
		}
	case *lower.Deref:
		if v, ok := b.Expr.(*lower.Var); ok && !arrow && v.Sigil == "$" {
			return "$" + v.Name
		}
	}
	return ""
}

// element is the legacy inferElementType: one element of the container
// name, which is its recorded element type in scalar context. An element of
// a container nothing was recorded for is a Scalar, not a guess.
func (a *loweredAnalysis) element(name string) types.Type {
	elem := a.elemOf(name)
	if elem == types.Unknown {
		return types.Scalar
	}
	// An element is one value: an element of a List is not a List.
	if narrowed, ok := types.NarrowByContext(elem, types.ScalarCtx); ok {
		return narrowed
	}
	return elem
}

// elemOf is the element type recorded for name, Unknown if none.
func (a *loweredAnalysis) elemOf(name string) types.Type {
	if name == "" {
		return types.Unknown
	}
	if sym, ok := a.st.Lookup(name); ok {
		return sym.ElemType
	}
	return types.Unknown
}

// checkIndex is the legacy checkElementIndex: an array index known to be a
// reference is reported, since a reference numifies to its address and
// perl warns "Use of reference ... as array index". An index of no known
// type is passed over.
func (a *loweredAnalysis) checkIndex(idx lower.Node, actual types.Type) {
	if actual == types.Unknown || !types.IsSubtype(actual, types.Ref) {
		return
	}
	start, end, _ := a.pos.Span(idx.Meta().ID)
	a.diags = append(a.diags, Diagnostic{
		StartByte: uint32(start),
		EndByte:   uint32(end),
		Severity:  argumentSeverity(actual, types.Num),
		Code:      CodeTypeMismatch,
		Message:   fmt.Sprintf("array index expects Num, got %s", actual),
	})
}

// elemType is the legacy listElementType: the element type of what a
// declaration fills an aggregate with. A list or constructor holds the join
// of its elements; a slice holds the elements of what it slices, and a
// key/value slice the keys too; anything else is one element, its own type.
func (a *loweredAnalysis) elemType(n lower.Node) types.Type {
	switch n := n.(type) {
	case *lower.List:
		return a.joinTypes(n.Elems)
	case *lower.ArrayLiteral:
		return a.joinTypes(n.Elems)
	case *lower.HashLiteral:
		return a.joinTypes(n.Elems)
	case *lower.Slice:
		return a.sliceElemType(n)
	case *lower.Repeat:
		// The elements are copies of the left operand, not values of
		// the operator's own Str: `(0) x $n` holds Ints.
		if n.ListRepeat {
			return a.elemType(n.Left)
		}
		return a.typeOf(n.Left)
	case *lower.Range:
		return a.rangeElemType(n)
	case *lower.Call:
		return a.callElemType(n)
	case *lower.MapGrepSort:
		return a.mapGrepSortElemType(n)
	}
	return a.annotations[n.Meta().ID]
}

// rangeElemType is the element type of a range: perl counts between two
// numbers in integers, and steps between strings by magic increment.
//
// perl: `1..3` is (1, 2, 3), `0..2.7` is (0, 1, 2), and `"a".."c"` is
// ("a", "b", "c").
func (a *loweredAnalysis) rangeElemType(r *lower.Range) types.Type {
	if types.IsSubtype(a.typeOf(r.Lo), types.Num) && types.IsSubtype(a.typeOf(r.Hi), types.Num) {
		return types.Int
	}
	return types.Str
}

// sliceElemType is the element type of what a slice returns.
//
// perl: with @a = (1, 2) and %h = (a => 1), @a[0,1] is (1, 2), %a[0] is
// (0, 1) and %h{a} is ("a", 1): a key/value slice returns each index or key
// before its value.
func (a *loweredAnalysis) sliceElemType(s *lower.Slice) types.Type {
	if s.Kind == lower.SliceList {
		if s.Base == nil {
			return types.Unknown
		}
		return a.elemType(s.Base)
	}
	elem := a.elemOf(containerName(s.Base, false, sliceSigil(s.Kind)))
	if elem == types.Unknown {
		return types.Unknown
	}
	switch s.Kind {
	case lower.SliceKVArray:
		return types.Join(types.Int, elem)
	case lower.SliceKVHash:
		return types.Join(types.Str, elem)
	}
	return elem
}

// joinTypes joins the annotations of ns.
func (a *loweredAnalysis) joinTypes(ns []lower.Node) types.Type {
	t := types.Unknown
	for _, n := range ns {
		if n != nil {
			t = types.Join(t, a.annotations[n.Meta().ID])
		}
	}
	return t
}

// recordElemType is the element half of the legacy
// inferAssignmentNarrowing: an array or hash records the element type of
// what fills it, and so does a scalar holding an anonymous array or hash.
func (a *loweredAnalysis) recordElemType(v *lower.Var, init lower.Node) {
	switch init.(type) {
	case *lower.ArrayLiteral, *lower.HashLiteral:
	default:
		if v.Sigil != "@" && v.Sigil != "%" {
			return
		}
	}
	if elem := a.elemType(init); elem != types.Unknown {
		a.st.UpdateElemType(v.Sigil+v.Name, elem)
	}
}
