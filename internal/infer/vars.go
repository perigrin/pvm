// ABOUTME: The variable, subscript and slice rules of Analyze: Var, Deref, Index, Key and Slice, and the element types a declaration records.
// ABOUTME: Each reads the lowered kinds: the sigil, the arrow and the slice kind are fields, never re-read from text.

package infer

import (
	"fmt"
	"strconv"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/types"
)

// variable types a variable: a scalar, array or hash is the type
// the symbol table narrowed it to, else its sigil's type, and $#a is an
// Int. A glob and a &name are not typed.
//
// A capture variable such as $1 is typed from the match before it, as
// capture says.
func (a *analysis) variable(v *lower.Var) types.Type {
	if v.Sigil == "$#" {
		return types.Int
	}
	if n, err := strconv.Atoi(v.Name); v.Sigil == "$" && err == nil && n >= 1 {
		return a.capture(v, n)
	}
	return a.lookupVar(v.Sigil, v.Name)
}

// lookupVar is sigil+name's narrowed type,
// or the sigil's type when nothing narrowed it. Unknown for a sigil that
// names no scalar, array or hash.
func (a *analysis) lookupVar(sigil, name string) types.Type {
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

// index types an element read, and
// the check on its index.
func (a *analysis) index(n *lower.Index) types.Type {
	a.walkBase(n.Base, n.Arrow, "@")
	if n.Idx != nil {
		a.checkIndex(n.Idx, a.walk(n.Idx))
	}
	return a.element(containerName(n.Base, n.Arrow, "@"))
}

// key types a hash element read. A key is stringified, not
// numified, so it is not checked.
func (a *analysis) key(n *lower.Key) types.Type {
	a.walkBase(n.Base, n.Arrow, "%")
	if n.Key != nil {
		a.walk(n.Key)
	}
	name := containerName(n.Base, n.Arrow, "%")
	if k, ok := constKey(n.Key); ok && name != "" {
		if sym, ok := a.st.Lookup(name); ok {
			if t, ok := sym.KeyTypes[k]; ok {
				return oneElement(t)
			}
		}
	}
	return a.element(name)
}

// constKey is the string a constant hash key spells: count for `{count}`,
// `count =>` and `{"count"}`. ok is false for a key computed at run time.
func constKey(n lower.Node) (string, bool) {
	switch n := n.(type) {
	case *lower.Bareword:
		return n.Name, true
	case *lower.StrLit:
		return n.Value, !n.Backtick
	}
	return "", false
}

// slice walks a slice. The slice itself is not typed; what it holds is
// recorded by elemType when a declaration takes it.
func (a *analysis) slice(n *lower.Slice) {
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
func (a *analysis) walkBase(base lower.Node, arrow bool, aggSigil string) {
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

// element is one element of the container
// name, which is its recorded element type in scalar context. An element of
// a container nothing was recorded for is a Scalar, not a guess.
func (a *analysis) element(name string) types.Type {
	return oneElement(a.elemOf(name))
}

// oneElement is one element of a container whose elements are elem.
func oneElement(elem types.Type) types.Type {
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
func (a *analysis) elemOf(name string) types.Type {
	if name == "" {
		return types.Unknown
	}
	if sym, ok := a.st.Lookup(name); ok {
		return sym.ElemType
	}
	return types.Unknown
}

// checkIndex is the index check: an array index known to be a
// reference is reported, since a reference numifies to its address and
// perl warns "Use of reference ... as array index". An index of no known
// type is passed over.
func (a *analysis) checkIndex(idx lower.Node, actual types.Type) {
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

// elemType is the element type of what a
// declaration fills an aggregate with. A list or constructor holds the join
// of its elements; a slice holds the elements of what it slices, and a
// key/value slice the keys too; anything else is one element, its own type.
func (a *analysis) elemType(n lower.Node) types.Type {
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
func (a *analysis) rangeElemType(r *lower.Range) types.Type {
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
func (a *analysis) sliceElemType(s *lower.Slice) types.Type {
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
func (a *analysis) joinTypes(ns []lower.Node) types.Type {
	t := types.Unknown
	for _, n := range ns {
		if n != nil {
			t = types.Join(t, a.annotations[n.Meta().ID])
		}
	}
	return t
}

// recordElemType is the element half of assignment narrowing: an array or
// hash records the element type of what fills it, and so does a scalar
// holding an anonymous array or hash.
func (a *analysis) recordElemType(v *lower.Var, init lower.Node) {
	switch init.(type) {
	case *lower.ArrayLiteral, *lower.HashLiteral:
	default:
		if v.Sigil != "@" && v.Sigil != "%" {
			return
		}
	}
	if elems, ok := hashFill(v, init); ok {
		values, keys := a.pairs(elems)
		if values != types.Unknown {
			a.st.UpdateElemType(v.Sigil+v.Name, values)
			a.st.UpdateKeyTypes(v.Sigil+v.Name, keys)
		}
		return
	}
	if elem := a.elemType(init); elem != types.Unknown {
		a.st.UpdateElemType(v.Sigil+v.Name, elem)
	}
}

// hashFill is the key/value list init fills a hash with: a hash literal's,
// or the list a %hash is assigned.
func hashFill(v *lower.Var, init lower.Node) ([]lower.Node, bool) {
	switch n := init.(type) {
	case *lower.HashLiteral:
		return n.Elems, true
	case *lower.List:
		if v.Sigil == "%" {
			return n.Elems, true
		}
	}
	return nil, false
}

// pairs reads a key/value list two at a time: the join of its values, and
// the value type under each constant key, the last one given winning as in
// perl. A hash element is a value, never a key: perl 5.42 -w is silent on
// `$h{$k} == $h{$l}` over `%h = ('multiply' => 19, 'i_multiply' => 19)`.
// When an element may flatten to other than one value, as `%a` in
// `(%a, b => 1)` does, which element is a key is not known, and nothing is.
func (a *analysis) pairs(elems []lower.Node) (types.Type, map[string]types.Type) {
	values, keys := types.Unknown, map[string]types.Type{}
	for i, e := range elems {
		if !isOneValue(e) {
			return types.Unknown, nil
		}
		if i%2 == 0 {
			continue
		}
		t := a.annotations[e.Meta().ID]
		values = types.Join(values, t)
		if k, ok := constKey(elems[i-1]); ok {
			keys[k] = t
		}
	}
	return values, keys
}

// isOneValue reports whether n is sure to be one value in list context: a
// literal, a scalar, an element, a constructor, a reference to a variable,
// or an operator that yields a value of its own.
func isOneValue(n lower.Node) bool {
	switch n := n.(type) {
	case *lower.NumLit, *lower.StrLit, *lower.Interp, *lower.Bareword,
		*lower.Index, *lower.Key, *lower.ArrayLiteral, *lower.HashLiteral,
		*lower.AnonSub, *lower.Unary:
		return true
	case *lower.Var:
		return n.Sigil == "$" || n.Sigil == "$#"
	case *lower.Ref:
		// `\(@a)` is a reference to each element.
		return n.Operand != nil && !n.Operand.Meta().Paren
	case *lower.Binary:
		return !logicalOps[n.Op]
	}
	return false
}
