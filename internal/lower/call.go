// ABOUTME: Lowering of calls (§4.14.2): the five call sites to one Call, calls through a reference, and MethodCall.
// ABOUTME: The handle slot, the indirect form and resolution are parse.Node flags; parentheses are read off the span.

package lower

import "tamarou.com/pvm/internal/parse"

// call lowers a `call` no special form claimed: a builtin or a sub by name,
// wherever the name stood in the old grammar. A first child carrying the
// Handle flag is the filehandle slot of print, printf or say, which is not
// an argument. The arguments are the children's comma chains, flattened.
func (l *lowerer) call(n *parse.Node) Node {
	c := &Call{Name: n.Text, Resolved: n.Resolved}
	args := n.Children
	if len(args) > 0 && args[0].Handle {
		c.Handle, args = l.handle(args[0]), args[1:]
	}
	c.Args = l.elements(args)
	return c
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
