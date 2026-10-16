// ABOUTME: Lowering of calls (§4.14.2): the five call sites to one Call, calls through a reference, and MethodCall.
// ABOUTME: The handle slot, the indirect form and resolution are parse.Node flags; parentheses are read off the span.

package lower

import (
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// call lowers a `call` no special form claimed: a builtin or a sub by name,
// wherever the name stood in the old grammar, or an indirect method call. A first child carrying the
// Handle flag is the filehandle slot of print, printf or say, which is not
// an argument. The arguments are the children's comma chains, flattened.
func (l *lowerer) call(n *parse.Node) Node {
	if n.Indirect {
		return l.indirect(n)
	}
	c := &Call{Name: n.Text, Resolved: n.Resolved, Parenthesized: parenthesized(n)}
	args := n.Children
	if len(args) > 0 && args[0].Handle {
		c.Handle, args = l.handle(args[0]), args[1:]
	}
	if namesCode[n.Text] && len(args) == 1 {
		c.Args = []Node{l.codeName(args[0])}
	} else {
		c.Args = l.elements(args)
	}
	return c
}

// namesCode are the builtins whose `&foo` operand names the sub rather than
// calling it. perl 5.42's -MO=Concise compiles `defined &foo`, `exists
// &foo` and `undef &foo` with no entersub.
var namesCode = map[string]bool{"defined": true, "exists": true, "undef": true}

// parenthesized reports whether a call's arguments stand in parentheses.
// A node's span covers its children exactly, so the span running past the
// last child -- or, with no child, past the name -- is the closing paren:
// `f(1,2)` does, `join ",", @a` does not.
func parenthesized(n *parse.Node) bool {
	end := n.Start + len(n.Text)
	if k := len(n.Children); k > 0 {
		end = n.Children[k-1].End
	}
	return n.End > end
}

// ampCall lowers `&foo` standing alone: a call that bypasses prototypes and
// passes the caller's @_ (§4.8.3), so both Ampersand and BareAmpersand.
// `&foo(...)` is an `index "("` over the term, which indexCall lowers
// without reaching this arm.
func (l *lowerer) ampCall(n *parse.Node) Node {
	if !isCodeName(n) {
		return nil
	}
	return &Call{Name: n.Text[1:], Ampersand: true, BareAmpersand: true}
}

// codeName lowers n where an `&foo` names the sub without calling it -- the
// operand of `\`, defined, exists, undef or goto -- to the Var &foo, and
// lowers anything else as an expression.
func (l *lowerer) codeName(n *parse.Node) Node {
	if !isCodeName(n) {
		return l.expr(n)
	}
	return l.stamp(n, &Var{Sigil: "&", Name: n.Text[1:]})
}

// isCodeName reports whether n is `&foo`: one term, an ampersand and a
// sub's name.
func isCodeName(n *parse.Node) bool {
	return n.Kind == parse.Term && len(n.Children) == 0 && len(n.Text) > 1 && n.Text[0] == '&' && word.MatchString(n.Text[1:])
}

// indexCall lowers an `index "("` over an ampersand, a call that bypasses
// prototypes with its arguments in the parentheses: `&foo(1)` by name, and
// `&$code(1)` and `&{$code}(1)` through a reference. `&$code` is one term
// whose inner expression the parser never built, so its Fn is a Hole with
// the term's id.
//
// An `index "("` with no arguments over a childless call is that call with
// empty parentheses, `time()` or `wantarray()`: the parser builds a builtin
// that takes no operand this way. Every other base declines.
func (l *lowerer) indexCall(n *parse.Node) Node {
	if n.Text != "(" || len(n.Children) == 0 {
		return nil
	}
	base := n.Children[0]
	if len(n.Children) == 1 && base.Kind == parse.Call && len(base.Children) == 0 && !base.Paren {
		return parenthesize(l.expr(base))
	}
	c := &Call{Parenthesized: true, Ampersand: true}
	switch {
	case isCodeName(base):
		c.Name = base.Text[1:]
	case base.Kind == parse.Term && len(base.Children) == 0 && strings.HasPrefix(base.Text, "&$"):
		c.Fn = l.stamp(base, &Hole{})
	case base.Kind == parse.Index && base.Text == "{" && len(base.Children) == 2 && base.Children[0].Text == "&":
		c.Fn = l.expr(base.Children[1])
	default:
		return nil
	}
	c.Args = l.elements(n.Children[1:])
	return c
}

// arrowCall lowers a `binary "->"` that names no dereference. Over a
// parenthesised right side it is a call through a code reference,
// `$code->(1)`, whose Fn is the left side. Otherwise it is a method call:
// the method is a `call` when it has parentheses (`$obj->method(1)`), a
// word when it has none (`Foo->new`), and a scalar when it is dynamic
// (`$obj->$name`, under an `index "("` when it has arguments).
func (l *lowerer) arrowCall(n *parse.Node) Node {
	if n.Text != "->" || len(n.Children) != 2 {
		return nil
	}
	inv, r := n.Children[0], n.Children[1]
	if r.Paren {
		return &Call{Fn: l.expr(inv), Args: l.parenArgs(r), Parenthesized: true}
	}
	m := &MethodCall{}
	switch {
	case r.Kind == parse.Call:
		m.Name, m.Args, m.HasParens = r.Text, l.elements(r.Children), parenthesized(r)
	case r.Kind == parse.Index && r.Text == "(" && len(r.Children) > 0:
		m.Dynamic, m.Args, m.HasParens = l.expr(r.Children[0]), l.elements(r.Children[1:]), true
	case r.Kind == parse.Term && len(r.Children) == 0 && sigilOf(r.Text) == "$":
		m.Dynamic = l.expr(r)
	case r.Kind == parse.Term && len(r.Children) == 0 && word.MatchString(r.Text):
		m.Name = r.Text
	default:
		return nil
	}
	m.Invocant = l.expr(inv)
	return m
}

// indirect lowers indirect object syntax, `new Foo(1)`: the parser's
// Indirect flag says the first child is the invocant and the rest are the
// arguments.
func (l *lowerer) indirect(n *parse.Node) Node {
	m := &MethodCall{Name: n.Text, Indirect: true, HasParens: parenthesized(n)}
	if len(n.Children) > 0 {
		m.Invocant = l.expr(n.Children[0])
		m.Args = l.elements(n.Children[1:])
	}
	return m
}

// parenArgs lowers the parenthesised arguments of `$code->(...)`: a `list`
// when there are none or several, and otherwise the one argument itself.
func (l *lowerer) parenArgs(r *parse.Node) []Node {
	if r.Kind == parse.List {
		return l.elements(r.Children)
	}
	return []Node{l.expr(r)}
}

// handle lowers a filehandle slot. A block slot, `print {$fh} "x"`, of one
// statement holding one expression is that expression; any other block is
// the block, lowered.
func (l *lowerer) handle(n *parse.Node) Node {
	if n.Kind == parse.Block {
		var stmts []*parse.Node
		for _, c := range n.Children {
			if c.Kind == parse.Statement {
				stmts = append(stmts, c)
			}
		}
		if len(stmts) == 1 && len(stmts[0].Children) == 1 {
			return l.expr(stmts[0].Children[0])
		}
	}
	return l.expr(n)
}
