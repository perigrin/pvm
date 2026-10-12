// ABOUTME: Lowering of calls (§4.14.2): the five call sites to one Call, calls through a reference, and MethodCall.
// ABOUTME: The handle slot, the indirect form and resolution are parse.Node flags; parentheses are read off the span.

package lower

import "tamarou.com/pvm/internal/parse"

// call lowers a `call` no special form claimed: a builtin or a sub by name,
// wherever the name stood in the old grammar. A first child carrying the
// Handle flag is the filehandle slot of print, printf or say, which is not
// an argument. The arguments are the children's comma chains, flattened.
func (l *lowerer) call(n *parse.Node) Node {
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

// indexCall lowers an `index "("` over `&foo`: a call by name that bypasses
// prototypes, its arguments in the parentheses. Every other base declines.
func (l *lowerer) indexCall(n *parse.Node) Node {
	if n.Text != "(" || len(n.Children) == 0 || !isCodeName(n.Children[0]) {
		return nil
	}
	return &Call{Name: n.Children[0].Text[1:], Args: l.elements(n.Children[1:]), Parenthesized: true, Ampersand: true}
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
