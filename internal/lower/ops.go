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

// list lowers a `list` node: each child is an element, and Fat marks the
// ones carrying the Fat flag.
func (l *lowerer) list(n *parse.Node) Node {
	out := &List{}
	for _, c := range n.Children {
		l.flatten(c, false, out)
	}
	return out
}

// commaList lowers a `binary ","` or `"=>"` chain, which is how a list is
// spelled outside parentheses, to the same flat List.
func (l *lowerer) commaList(n *parse.Node) Node {
	if !isComma(n) {
		return nil
	}
	out := &List{}
	l.flatten(n, false, out)
	return out
}

// isComma reports whether n is one link of a comma chain.
func isComma(n *parse.Node) bool {
	return n.Kind == parse.Binary && !n.Paren && len(n.Children) == 2 && (n.Text == "," || n.Text == "=>")
}

// flatten appends n's elements to out. fat says n stands left of a `=>`,
// which falls to the last element of a chain: the chain is left-nested,
// so `a => 1, b => 2` holds `b` as the right operand of a `,`.
func (l *lowerer) flatten(n *parse.Node, fat bool, out *List) {
	if isComma(n) {
		l.flatten(n.Children[0], n.Text == "=>", out)
		l.flatten(n.Children[1], fat, out)
		return
	}
	fat = fat || n.Fat
	out.Elems = append(out.Elems, l.elem(n, fat))
	out.Fat = append(out.Fat, fat)
}

// elem lowers one list element. §4.5.4: the word left of `=>` is a string,
// whether the Fat flag or a `binary "=>"` says it stands there.
func (l *lowerer) elem(n *parse.Node, fat bool) Node {
	if fat && n.Kind == parse.Call && len(n.Children) == 0 {
		return l.stamp(n, &Bareword{Name: n.Text, Quoted: true})
	}
	return l.expr(n)
}
