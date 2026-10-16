// ABOUTME: The special-form rules of Analyze: MapGrepSort, eval and do, Ref, Match, and the forms left untyped.
// ABOUTME: Each reads the lowered kinds: map, grep and sort are told apart by Op, never by source text.

package infer

import (
	"strings"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/types"
)

// mapGrepSort walks a map, grep or sort. Its own value is not typed;
// what it hands back is recorded by elemType when a
// declaration takes it.
func (a *analysis) mapGrepSort(n *lower.MapGrepSort) types.Type {
	a.walkChildren(n)
	return types.Unknown
}

// mapGrepSortElemType is the element type of a map, grep or sort.
// sort reorders and grep selects, so their elements are the source's; map
// transforms, so its elements are what its block yields. perl 5.42: with
// @n = (3, 1), `map { $_ * 2 } @n` is (6, 2) and `map { "x$_" } @n` is
// ("x3", "x1"). map's expression form is not typed, and neither is a
// block that yields a list, whose elements each become one: perl 5.42,
// `map { glob "$_" } @a` is the names each glob found. A Hole's Any is
// kept.
func (a *analysis) mapGrepSortElemType(n *lower.MapGrepSort) types.Type {
	if n.Op == "map" {
		if n.Block == nil {
			return types.Unknown
		}
		if t := a.blockValue(n.Block); t == types.Any || t&(types.Array|types.Hash) == 0 {
			return t
		}
		return types.Unknown
	}
	switch len(n.List) {
	case 0:
		return types.Unknown
	case 1:
		return a.aggElemType(n.List[0])
	}
	return a.joinTypes(n.List)
}

// blockValue is a block's value, its last
// expression statement's. A block perl rejected is a Hole, whose value is
// Any, never a guess.
func (a *analysis) blockValue(n lower.Node) types.Type {
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

// ref types a reference: `\EXPR` is typed from what the
// reference is taken TO. A reference to anything but a named variable -- an
// element, a literal, a dereference -- is not typed: it is no evidence of
// which reference results.
//
// A scalar variable typed Glob holds a glob, so a reference to it is a
// GlobRef: perl 5.42, `my $v = *STDOUT; ref \$v` is GLOB.
//
// A parenthesised array is typed as the array, although perl takes a
// reference to each element: perl 5.42, with @a = (1, 2), `ref(my $r =
// \(@a))` is SCALAR.
func (a *analysis) ref(n *lower.Ref) types.Type {
	a.operand(n.Operand)
	if v, ok := n.Operand.(*lower.Var); ok {
		if v.Sigil == "$" && a.typeOf(v) == types.Glob {
			return types.GlobRef
		}
		if t, ok := refTypes[v.Sigil]; ok {
			return t
		}
	}
	return types.Unknown
}

// glob types `glob EXPR` as the builtin call it is. `<*.c>`, whose pattern
// is the angle term's own bytes, is not typed.
func (a *analysis) glob(n *lower.Glob) types.Type {
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

// match is the rule for `=~` and `!~` over a match, a substitution
// or a transliteration: CORE.tpm's candidates, chosen by the context
// around it, check the target and give its type. A match in scalar context
// is a Boolean, and a substitution a Str: perl 5.42, `"abc" =~ /b/` is 1,
// `"aaa" =~ s/a/b/g` is 3 and `"xxx" =~ s/a/b/g` is "". qr// is a Regex.
// A match against $_, with no target, is not typed.
func (a *analysis) match(n *lower.Match) types.Type {
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

// bindsCaptures is the list half of assignment narrowing for a match:
// declaring several names from a match binds each a capture, and a capture
// is a Str. perl 5.42, `my ($a, $b) = ("x42" =~ /([a-z])(\d+)/)` is ("x",
// "42").
func (a *analysis) bindsCaptures(d *lower.Decl) bool {
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

// capture types $n, the n-th group of the last
// match, substitution or qr before it in source order. A capture is a
// substring, so a Str, and an Int when its group can match nothing but
// digits: perl 5.42 -w, after "abc123" =~ /(\d+)/, `$1 + 1` is 124 and
// silent, and after "abcxyz" =~ /([a-z]+)/ it warns "isn't numeric". No
// match before it, or no such group, leaves Str.
//
// In the body of an `if` modifier whose condition is a match, that match
// runs first, though it follows the body in the source, so it is the one
// read, as ownMatch says. perl 5.42 -w is silent on
// `print $1 + 1 if "a7" =~ /(\d+)/`.
func (a *analysis) capture(v *lower.Var, n int) types.Type {
	start, _, _ := a.pos.Span(v.ID)
	var m lower.Node = lastMatch(a.pos, start)
	if g := a.ownMatch(start, a.modifier); g != nil {
		m = g
	}
	if m == nil {
		return types.Str
	}
	groups := captureGroups(quotePattern(a.pos.Quote(m.Meta().ID)))
	if n <= len(groups) && isNumericOnlyGroup(groups[n-1]) {
		return types.Int
	}
	return types.Str
}

// lastMatch is the last match, substitution or qr that starts before off
// in source order, or nil.
func lastMatch(pos *lower.Positions, off int) lower.Node {
	return pos.LastBefore(off, func(x lower.Node) bool {
		mt, ok := x.(*lower.Match)
		return ok && (mt.Op == "m" || mt.Op == "s" || mt.Op == "qr")
	})
}

// successMatch is the match n's condition is, and the branch that runs
// when it succeeds: Then, or Else under `!~`. nil when the condition is
// no match or substitution.
func successMatch(n *lower.IfStmt) (*lower.Match, lower.Node) {
	m, ok := n.Cond.(*lower.Match)
	if !ok || (m.Op != "m" && m.Op != "s") {
		return nil, nil
	}
	if m.Negated {
		return m, n.Else
	}
	return m, n.Then
}

// ownMatch is the match guarding stmt, when it is the last to run before
// a capture at off in stmt's successful branch: the block form puts the
// match before the capture in the source, and the `if` modifier puts it
// after, with no other match in the body before the capture. nil when
// stmt is nil, guards with no match, or another match intervenes.
func (a *analysis) ownMatch(off int, stmt *lower.IfStmt) *lower.Match {
	if stmt == nil {
		return nil
	}
	g, _ := successMatch(stmt)
	if g == nil {
		return nil
	}
	if last := lastMatch(a.pos, off); last != nil && last != lower.Node(g) && !a.before(last, stmt) {
		return nil
	}
	return g
}

// quotePattern is the pattern inside a match, substitution or qr quote
// token: `(\d+)` from `/(\d+)/`, `m{(\d+)}x` and `s/(\d+)/x/g` alike. A
// bracketing delimiter nests, and a backslash escapes the character after
// it. "" when tok is not such a token.
func quotePattern(tok string) string {
	body := []rune(strings.TrimLeft(strings.TrimLeft(tok, "mqrs"), " \t\n"))
	if len(body) == 0 {
		return ""
	}
	open := body[0]
	closer := map[rune]rune{'(': ')', '[': ']', '{': '}', '<': '>'}[open]
	if closer == 0 {
		closer = open
	}
	depth := 0
	for i := 1; i < len(body); i++ {
		switch r := body[i]; {
		case r == '\\':
			i++
		case r == open && closer != open:
			depth++
		case r == closer && depth > 0:
			depth--
		case r == closer:
			return string(body[1:i])
		}
	}
	return ""
}
