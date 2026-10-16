// ABOUTME: Lowering of statements (§5.17's kinds, without spans): blocks, conditionals, loops, use and phasers.
// ABOUTME: Each kind is told apart by CST kind, keyword and the Modifier, LoopVar and CStyle flags, never by child guesses.

package lower

import (
	"regexp"

	"tamarou.com/pvm/internal/parse"
)

// stmt lowers one child of a file or a block. Trivia lowers to nothing.
func (l *lowerer) stmt(n *parse.Node) Node {
	switch n.Kind {
	case parse.Trivia:
		return nil
	case parse.Statement:
		return l.statement(n, nil)
	}
	return l.expr(n)
}

// statement lowers a Statement by its first child that is neither trivia
// nor a label. An expression or a bare block is wrapped in a kind carrying
// the statement's id; a statement form carries its own node's id, the
// Statement mapping forward to it. An empty statement lowers to nothing.
//
// The labels go to a loop or a bare block, where last and next read them.
// ponytail: a label on any other statement is dropped; carry it when a
// consumer resolves goto.
func (l *lowerer) statement(n *parse.Node, labels []Node) Node {
	var body []*parse.Node
	for _, c := range n.Children {
		switch c.Kind {
		case parse.Trivia:
		case parse.Label:
			labels = append(labels, l.stamp(c, &Bareword{Name: c.Text}))
		default:
			body = append(body, c)
		}
	}
	if len(body) == 0 {
		return nil
	}
	switch c := body[0]; c.Kind {
	case parse.Statement:
		// A bare block with a continue is a statement of its own inside
		// this one, which holds the labels.
		return l.statement(c, labels)
	case parse.Block:
		return l.stamp(n, &BlockStmt{Labels: labels, Body: l.expr(c), Continue: l.continueBlock(body[1:])})
	case parse.Conditional:
		return l.conditional(c)
	case parse.Loop:
		return l.loop(c, labels)
	case parse.Use:
		return l.use(c)
	case parse.Phaser:
		return l.stamp(c, &PhaserStmt{Kind: c.Text, Body: l.nth(c.Children, 0)})
	}
	return l.stamp(n, &ExprStmt{X: l.expr(body[0])})
}

// loop lowers a Loop to WhileStmt, CForStmt or ForeachStmt. The modifier
// form's children are body then condition, the block form's the reverse,
// and the Modifier flag says which.
func (l *lowerer) loop(n *parse.Node, labels []Node) Node {
	kids := n.Children
	var out Node
	switch {
	case n.CStyle:
		f := &CForStmt{Labels: labels}
		for i, slot := range []*Node{&f.Init, &f.Cond, &f.Post} {
			if n.ForSlots[i] && len(kids) > 0 {
				*slot, kids = l.expr(kids[0]), kids[1:]
			}
		}
		f.Body = l.nth(kids, 0)
		out = f
	case n.Text == "while" || n.Text == "until":
		w := &WhileStmt{Labels: labels, Negated: n.Text == "until"}
		if n.Modifier {
			w.Body, w.Cond = l.nth(kids, 0), l.nth(kids, 1)
			w.PostCond = len(kids) > 0 && isDoBlock(kids[0])
		} else {
			w.Cond, w.Body = l.nth(kids, 0), l.nth(kids, 1)
			w.Continue = l.continueBlock(kids[min(2, len(kids)):])
		}
		out = w
	case n.Text == "for" || n.Text == "foreach":
		f := &ForeachStmt{Labels: labels}
		if n.Modifier {
			f.Body, f.List = l.nth(kids, 0), l.nth(kids, 1)
		} else {
			if len(kids) > 0 && kids[0].LoopVar {
				f.Var, kids = l.expr(kids[0]), kids[1:]
			}
			f.List, f.Body = l.nth(kids, 0), l.nth(kids, 1)
			f.Continue = l.continueBlock(kids[min(2, len(kids)):])
		}
		out = f
	default:
		out = &Hole{}
	}
	return l.stamp(n, out)
}

// use lowers `use`, `no` and `require`. A bare word first is the module
// and a version-shaped term the version, `v5` included: perl reads `use Foo 1.2` as a
// version check, never an import. Whatever follows is the import list.
func (l *lowerer) use(n *parse.Node) Node {
	out := &UseStmt{Op: n.Text}
	kids := n.Children
	if len(kids) > 0 && bareTerm(kids[0], word) && !bareTerm(kids[0], version.MatchString) {
		out.Module, kids = kids[0].Text, kids[1:]
	}
	if len(kids) > 0 && bareTerm(kids[0], version.MatchString) {
		out.Version, kids = kids[0].Text, kids[1:]
	}
	if out.Module == "" && out.Version == "" && len(kids) > 0 {
		out.Expr, kids = l.expr(kids[0]), kids[1:]
	}
	out.Args = l.nth(kids, 0)
	return l.stamp(n, out)
}

// version matches a version number: 5.010, 1.2, v5.36.
var version = regexp.MustCompile(`^v?[0-9][0-9._]*$`)

// bareTerm reports whether n is a childless term whose text is accepts.
func bareTerm(n *parse.Node, is func(string) bool) bool {
	return n.Kind == parse.Term && len(n.Children) == 0 && is(n.Text)
}

// isDoBlock reports whether n is `do BLOCK`, the body of a loop that tests
// after it runs.
func isDoBlock(n *parse.Node) bool {
	return n.Kind == parse.Call && n.Text == "do" && len(n.Children) == 1 && n.Children[0].Kind == parse.Block
}

// continueBlock lowers the block of a `continue` that leads rest, or returns
// nil when rest has none.
func (l *lowerer) continueBlock(rest []*parse.Node) Node {
	if len(rest) == 0 || rest[0].Kind != parse.Loop || rest[0].Text != "continue" {
		return nil
	}
	return l.nth(rest[0].Children, 0)
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
		out.Then, out.Cond = l.nth(n.Children, 0), l.nth(n.Children, 1)
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
	out := &IfStmt{Cond: l.nth(n.Children, 0), Then: l.nth(n.Children, 1)}
	if len(rest) == 0 {
		return out
	}
	switch e := rest[0]; {
	case e.Kind == parse.Conditional && e.Text == "elsif":
		out.Else = l.stamp(e, l.branch(e, rest[1:]))
	case e.Kind == parse.Conditional && e.Text == "else":
		out.Else = l.nth(e.Children, 0)
	default:
		out.Else = l.expr(e)
	}
	return out
}

// nth lowers the i-th of nodes, or returns nil when there is no such node.
func (l *lowerer) nth(nodes []*parse.Node, i int) Node {
	if i >= len(nodes) {
		return nil
	}
	return l.expr(nodes[i])
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
