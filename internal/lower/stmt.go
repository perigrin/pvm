// ABOUTME: Lowering of statements (§5.17's kinds, without spans): blocks, conditionals, loops, use and phasers.
// ABOUTME: Each kind is told apart by CST kind, keyword and the Modifier, LoopVar and CStyle flags, never by child guesses.

package lower

import "tamarou.com/pvm/internal/parse"

// stmt lowers one child of a file or a block. Trivia lowers to nothing.
func (l *lowerer) stmt(n *parse.Node) Node {
	switch n.Kind {
	case parse.Trivia:
		return nil
	case parse.Statement:
		return l.statement(n)
	}
	return l.expr(n)
}

// statement lowers a Statement by its one non-trivia child. An expression
// is wrapped in an ExprStmt carrying the statement's id; a statement form
// carries its own node's id, the Statement mapping forward to it. An empty
// statement lowers to nothing.
func (l *lowerer) statement(n *parse.Node) Node {
	var body []*parse.Node
	for _, c := range n.Children {
		if c.Kind != parse.Trivia {
			body = append(body, c)
		}
	}
	if len(body) == 0 {
		return nil
	}
	c := body[0]
	if c.Kind == parse.Conditional {
		return l.conditional(c)
	}
	return l.stamp(n, &ExprStmt{X: l.expr(c)})
}

// conditional lowers `if`, `unless` and `elsif`, in the block form or the
// modifier form, to an IfStmt. The other conditionals -- `given`, `when`,
// `try` (with its `catch` and `finally`) and `defer` -- are a Hole until a
// consumer needs them.
func (l *lowerer) conditional(n *parse.Node) Node {
	if n.Text != "if" && n.Text != "unless" {
		return l.stamp(n, &Hole{})
	}
	out := &IfStmt{}
	if n.Modifier {
		out.Then, out.Cond = l.child(n, 0), l.child(n, 1)
	} else {
		out = l.branch(n, n.Children[min(2, len(n.Children)):])
	}
	if n.Text == "unless" {
		out.Then, out.Else = out.Else, out.Then
	}
	return l.stamp(n, out)
}

// branch lowers a block-form conditional's condition and block, and rest,
// the `elsif` and `else` siblings that follow them in the CST, as its Else:
// an `elsif` is an IfStmt of its own, over the siblings after it, and an
// `else` is its block.
func (l *lowerer) branch(n *parse.Node, rest []*parse.Node) *IfStmt {
	out := &IfStmt{Cond: l.child(n, 0), Then: l.child(n, 1)}
	if len(rest) == 0 {
		return out
	}
	switch e := rest[0]; {
	case e.Kind == parse.Conditional && e.Text == "elsif":
		out.Else = l.stamp(e, l.branch(e, rest[1:]))
	case e.Kind == parse.Conditional && e.Text == "else":
		out.Else = l.child(e, 0)
	default:
		out.Else = l.expr(e)
	}
	return out
}

// child lowers n's i-th child, or returns nil when n has no such child.
func (l *lowerer) child(n *parse.Node, i int) Node {
	if i >= len(n.Children) {
		return nil
	}
	return l.expr(n.Children[i])
}

// block lowers a Block's statements, trivia dropped.
func (l *lowerer) block(n *parse.Node) Node {
	out := &Block{}
	for _, c := range n.Children {
		if s := l.stmt(c); s != nil {
			out.Stmts = append(out.Stmts, s)
		}
	}
	return out
}
