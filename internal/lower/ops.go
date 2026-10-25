// ABOUTME: Lowering of operators (§4.14.2): Binary, Repeat, Assign, Range and the lone comparison.
// ABOUTME: The operator is the parse node's Text and every flag a field read, so nothing rescans the source.

package lower

import "tamarou.com/pvm/internal/parse"

// binary lowers any `binary` a narrower arm left: Op is the node's Text.
// An arrow is a call or a dereference, never an operator, so it declines.
func (l *lowerer) binary(n *parse.Node) Node {
	if n.Text == "->" {
		return nil
	}
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
		r.ListRepeat = left.Paren || isQw(left) || declOfParen(left)
	}
	return r
}

// declOfParen reports whether n is a declaration of a parenthesised list,
// `my ($x)`: measured on perl 5.44.0, `(my ($x)) x 2` repeats a list
// (repeat lKP/DOLIST) where `my $x x 2` repeats a string.
func declOfParen(n *parse.Node) bool {
	return n.Kind == parse.Declaration && len(n.Children) > 0 && n.Children[len(n.Children)-1].Paren
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
	if len(n.Children) > 0 {
		left := n.Children[0]
		a.ListAssign = left.Kind == parse.List || left.Paren || aggregate(left) || refOfParen(left)
	}
	return a
}

// refOfParen reports whether n is `\(...)`. Assigned to, it is a list:
// measured on perl 5.44.0, `\($x) = \$y` is aassign where `\$x = \$y` is
// refassign, as `my \($x) = f()` is a list assignment in decl.go.
func refOfParen(n *parse.Node) bool {
	return n.Kind == parse.Unary && n.Text == "ref" && len(n.Children) == 1 && n.Children[0].Paren
}

// aggregate reports whether n names an array or a hash: a term with an
// aggregate sigil, an aggregate deref, a slice, or a postfix deref or slice.
// perl assigns a list to each, and a scalar to `$#a` and `$h{k}`.
func aggregate(n *parse.Node) bool {
	switch n.Kind {
	case parse.Term:
		return n.Text != "" && (n.Text[0] == '@' || n.Text[0] == '%')
	case parse.Unary:
		return n.Text == "@" || n.Text == "%"
	case parse.Index:
		return len(n.Children) > 0 && aggregate(n.Children[0])
	case parse.Binary:
		return n.Text == "->" && len(n.Children) == 2 && aggregate(n.Children[1])
	}
	return false
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
	return l.elementList(n.Children)
}

// elementList lowers ns, in source order, to one flat List: a comma chain
// among them contributes each of its operands. A comma the source
// parenthesised is one element, a list of its own.
func (l *lowerer) elementList(ns []*parse.Node) *List {
	out := &List{}
	for _, c := range ns {
		l.flatten(c, false, out)
	}
	return out
}

// elements is elementList's elements alone: the arguments of a call, and
// the list of a map, grep or sort.
func (l *lowerer) elements(ns []*parse.Node) []Node {
	return l.elementList(ns).Elems
}

// commaChain lowers a `binary ","` or `"=>"` chain, which is how a list is
// spelled outside parentheses, to the same flat List.
func (l *lowerer) commaChain(n *parse.Node) Node {
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

// prefixOps are CORE's `:prefix` operators by their node Text, but for `\`,
// whose `unary "ref"` lowers to Ref.
var prefixOps = map[string]bool{
	"!": true, "not": true, "neg": true, "pos": true, "~": true, "~.": true,
	"++": true, "--": true,
}

// unary lowers a prefix operator's `unary`: Op is the node's Text. A sigil's
// `unary` is a dereference, and declines.
func (l *lowerer) unary(n *parse.Node) Node {
	if !prefixOps[n.Text] || len(n.Children) != 1 {
		return nil
	}
	return &Unary{Op: n.Text, Operand: l.expr(n.Children[0])}
}

// postfix lowers `postfix "++"` and `"--"`: Postfix is the CST's kind.
func (l *lowerer) postfix(n *parse.Node) Node {
	if len(n.Children) != 1 {
		return nil
	}
	return &Unary{Op: n.Text, Operand: l.expr(n.Children[0]), Postfix: true}
}

// ternary lowers `ternary "?:"`, any of whose operands may be missing.
func (l *lowerer) ternary(n *parse.Node) Node {
	t := &Ternary{}
	t.Cond, t.Then = l.operands(n)
	if len(n.Children) > 2 {
		t.Else = l.expr(n.Children[2])
	}
	return t
}

// eqOps are the CHEQOP operators; every other chain link is CHRELOP.
var eqOps = map[string]bool{"==": true, "!=": true, "eq": true, "ne": true}

// cmpChain lowers a `cmp_chain` of two or more links, flat as the CST has
// it. Every link shares one class, so the first operator names it.
func (l *lowerer) cmpChain(n *parse.Node) Node {
	c := &CmpChain{Ops: append([]string(nil), n.Ops...)}
	for _, o := range n.Children {
		c.Operands = append(c.Operands, l.expr(o))
	}
	if eqOps[n.Text] {
		c.Class = ChEqOp
	}
	return c
}
