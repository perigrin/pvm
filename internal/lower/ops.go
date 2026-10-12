// ABOUTME: Lowering of operators (§4.14.2): Binary, Repeat, Assign, Range and the lone comparison.
// ABOUTME: The operator is the parse node's Text and every flag a field read, so nothing rescans the source.

package lower

import "tamarou.com/pvm/internal/parse"

// binary lowers any `binary` a narrower arm left: Op is the node's Text.
func (l *lowerer) binary(n *parse.Node) Node {
	b := &Binary{Op: n.Text}
	b.L, b.R = l.operands(n)
	return b
}

// repeat lowers `binary "x"`. A left operand in parentheses or a qw list
// repeats a list (§4.10): the Paren flag says which, so no span is read.
func (l *lowerer) repeat(n *parse.Node) Node {
	if n.Text != "x" {
		return nil
	}
	r := &Repeat{}
	r.Left, r.Count = l.operands(n)
	if len(n.Children) > 0 {
		left := n.Children[0]
		r.ListRepeat = left.Paren || isQw(left)
	}
	return r
}

// isQw reports whether n is a qw list's term.
func isQw(n *parse.Node) bool {
	if n.Kind != parse.Term || len(n.Children) != 0 {
		return false
	}
	m := quoteOp.FindStringSubmatch(n.Text)
	return m != nil && m[1] == "qw"
}

// rangeOp lowers `binary ".."` and `binary "..."`.
func (l *lowerer) rangeOp(n *parse.Node) Node {
	if n.Text != ".." && n.Text != "..." {
		return nil
	}
	r := &Range{Exclusive: n.Text == "..."}
	r.Lo, r.Hi = l.operands(n)
	return r
}

// assignOps are CORE's `:infix(ASSIGN)` operators.
var assignOps = map[string]bool{
	"=": true, "+=": true, "-=": true, "*=": true, "/=": true, "**=": true,
	"%=": true, ".=": true, "x=": true, "&=": true, "|=": true, "^=": true,
	"<<=": true, ">>=": true, "&.=": true, "|.=": true, "^.=": true,
	"&&=": true, "||=": true, "//=": true, "^^=": true,
}

// assign lowers a `binary` whose Text is an assignment operator.
func (l *lowerer) assign(n *parse.Node) Node {
	if !assignOps[n.Text] {
		return nil
	}
	a := &Assign{Op: n.Text}
	a.L, a.R = l.operands(n)
	return a
}

// comparison lowers a `cmp_chain` of one link, which is a plain binary
// operator: perl compiles `$a < $b` to `lt` and only a longer chain to
// cmpchain_and.
func (l *lowerer) comparison(n *parse.Node) Node {
	if len(n.Children) != 2 {
		return nil
	}
	return l.binary(n)
}

// operands lowers a node's first two children, either of which may be
// missing.
func (l *lowerer) operands(n *parse.Node) (left, right Node) {
	if len(n.Children) > 0 {
		left = l.expr(n.Children[0])
	}
	if len(n.Children) > 1 {
		right = l.expr(n.Children[1])
	}
	return left, right
}
