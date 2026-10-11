// ABOUTME: Lowering of variables and subscripts (§4.14.2): Var, Deref, Index and Key, with Arrow.
// ABOUTME: The sigil, the name and the arrow are read off the CST's Text and flags, never re-lexed from the span.

package lower

import (
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// variable lowers a childless sigil-led term to a Var, and declines anything
// else. The sigil is the leading bytes and the name is the rest, with the
// braces of ${x} and $#{x} removed: perl -MO=Deparse prints both without
// them.
func (l *lowerer) variable(n *parse.Node) Node {
	sigil := sigilOf(n.Text)
	if sigil == "" || len(n.Children) != 0 {
		return nil
	}
	name := n.Text[len(sigil):]
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		name = strings.TrimSpace(name[1 : len(name)-1])
		if !bracedName.MatchString(name) {
			return nil
		}
	}
	if name == "" {
		// A bare sigil: the base of &{$x} or of a postfix slice.
		return nil
	}
	return &Var{Sigil: sigil, Name: name}
}

// sigilOf returns the sigil t starts with, or "" if it starts with none.
func sigilOf(t string) string {
	if strings.HasPrefix(t, "$#") {
		return "$#"
	}
	if t != "" && strings.IndexByte("$@%*&", t[0]) >= 0 {
		return t[:1]
	}
	return ""
}

// bracedName matches what may stand inside the braces of a variable's
// name: a word, package-qualified or not, or a caret name like ^WARNING_BITS.
var bracedName = regexp.MustCompile(`^\^?\w+(::\w+)*$`)

// subscript lowers an `index` to an Index ("[") or a Key ("{"), the Arrow
// bit copied from the node. It declines an index that is not one element of
// one container: a slice, whose base is @-, %- or list-shaped, and the
// &{$x} and *foo{CODE} forms, whose bases are a code or glob sigil.
func (l *lowerer) subscript(n *parse.Node) Node {
	if len(n.Children) != 2 || !elementBase(n) {
		return nil
	}
	base, sub := n.Children[0], n.Children[1]
	if n.Text == "{" {
		return &Key{Base: l.expr(base), Key: l.hashKey(sub), Arrow: n.Arrow}
	}
	return &Index{Base: l.expr(base), Idx: l.expr(sub), Arrow: n.Arrow}
}

// elementBase reports whether an index's base makes it an element access.
func elementBase(n *parse.Node) bool {
	base := n.Children[0]
	if !n.Arrow && (base.Paren || base.Kind == parse.List) {
		return false
	}
	if (base.Kind == parse.Term || base.Kind == parse.Unary) && base.Text != "" {
		return strings.IndexByte("@%&*", base.Text[0]) < 0
	}
	return true
}
