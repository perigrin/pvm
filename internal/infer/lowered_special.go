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

// refTypes is the reference `\` takes to a variable of each sigil: `\$x` is
// a ScalarRef, `\@a` an ArrayRef, `\%h` a HashRef, `\&f` a CodeRef and
// `\*FH` a GlobRef.
var refTypes = map[string]types.Type{
	"$": types.ScalarRef,
	"@": types.ArrayRef,
	"%": types.HashRef,
	"&": types.CodeRef,
	"*": types.GlobRef,
}

// ref is the legacy inferRefgenType: `\EXPR` is typed from what the
// reference is taken TO. A reference to anything but a named variable -- an
// element, a literal, a dereference -- is not typed: it is no evidence of
// which reference results.
//
// A parenthesised array is typed as the array, as on the legacy path,
// although perl takes a reference to each element: perl 5.42, with
// @a = (1, 2), `ref(my $r = \(@a))` is SCALAR.
func (a *loweredAnalysis) ref(n *lower.Ref) types.Type {
	a.operand(n.Operand)
	if v, ok := n.Operand.(*lower.Var); ok {
		if t, ok := refTypes[v.Sigil]; ok {
			return t
		}
	}
	return types.Unknown
}

// glob types `glob EXPR` as the builtin call it is. `<*.c>`, whose pattern
// is the angle term's own bytes, is not typed, as on the legacy path.
func (a *loweredAnalysis) glob(n *lower.Glob) types.Type {
	if n.Pattern != nil && n.Pattern.Meta().ID.Role == lower.RoleAngleBody {
		a.walk(n.Pattern)
		return types.Unknown
	}
	c := &lower.Call{Name: "glob", Parenthesized: true}
	*c.Meta() = *n.Meta()
	if n.Pattern != nil {
		c.Args = []lower.Node{n.Pattern}
	}
	return a.call(c)
}

// match is the legacy rule for `=~` and `!~` over a match, a substitution
// or a transliteration: CORE.pmt's candidates, chosen by the context
// around it, check the target and give its type. A match in scalar context
// is a Boolean, and a substitution a Str: perl 5.42, `"abc" =~ /b/` is 1,
// `"aaa" =~ s/a/b/g` is 3 and `"xxx" =~ s/a/b/g` is "". qr// is a Regex.
// A match against $_, with no target, is not typed.
func (a *loweredAnalysis) match(n *lower.Match) types.Type {
	if n.Op == "qr" {
		return types.Regex
	}
	if n.Target == nil {
		return types.Unknown
	}
	op := "=~"
	if n.Negated {
		op = "!~"
	}
	t := a.infix(n, op, operandShape(n.Target)+"$", n.Target, nil)
	switch {
	case op == "=~" && n.Op == "s":
		return types.Str
	case op == "=~" && n.Op == "m" && a.evalCtx[n.ID] == types.ScalarCtx:
		return types.Boolean
	}
	return t
}

// bindsCaptures is the list half of the legacy inferAssignmentNarrowing for
// a match: declaring several names from a match binds each a capture, and
// a capture is a Str. perl 5.42, `my ($a, $b) = ("x42" =~ /([a-z])(\d+)/)`
// is ("x", "42").
func (a *loweredAnalysis) bindsCaptures(d *lower.Decl) bool {
	m, ok := d.Init.(*lower.Match)
	vars := declaredVars(d)
	if !ok || m.Op != "m" || m.Target == nil || len(vars) < 2 {
		return false
	}
	for _, v := range vars {
		a.st.UpdateType(v.Sigil+v.Name, types.Str)
	}
	return true
}
