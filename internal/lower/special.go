// ABOUTME: Lowering of the special forms of §4.14.2: MapGrepSort, do, eval, the anonymous constructors, Ref,
// ABOUTME: FileTest, Readline, Glob, LoopEx, Wantarray and Match, each told apart by CST kind, name and flags.

package lower

import "tamarou.com/pvm/internal/parse"

// mapGrepSort lowers `map`, `grep` and `sort`, block form and expression
// form alike. Op is the call's name, so map and grep are separated by a
// field rather than by the source they were spelled with.
//
// The block form's first child is the block, or the Unknown the parser
// leaves in the block slot when the brace never closes (UnclosedBrace is
// made there alone). sort's comparator slot carries Comparator. Whatever
// follows is the list, one comma tree, flattened. In map's and grep's
// expression form the first comma operand is the expression.
//
// A block lowers through expr like any other child: a Hole until blocks have
// an arm of their own, carrying the block's id so the slot is never empty.
func (l *lowerer) mapGrepSort(n *parse.Node) Node {
	if n.Text != "map" && n.Text != "grep" && n.Text != "sort" {
		return nil
	}
	out := &MapGrepSort{Op: n.Text}
	rest := n.Children
	if len(rest) > 0 {
		switch c := rest[0]; {
		case c.Kind == parse.Block || c.Kind == parse.Unknown && c.Refusal == parse.UnclosedBrace:
			out.Block, rest = l.expr(c), rest[1:]
		case c.Comparator:
			out.FirstExpr, rest = l.expr(c), rest[1:]
		}
	}
	for _, c := range rest {
		out.List = append(out.List, l.commaList(c)...)
	}
	if out.Block == nil && out.FirstExpr == nil && n.Text != "sort" && len(out.List) > 0 {
		out.FirstExpr, out.List = out.List[0], out.List[1:]
	}
	return out
}

// commaList lowers a comma tree to its operands in source order. A comma
// the source parenthesised is one operand, a list of its own.
func (l *lowerer) commaList(n *parse.Node) []Node {
	if n.Kind != parse.Binary || n.Text != "," && n.Text != "=>" || n.Paren {
		return []Node{l.expr(n)}
	}
	var out []Node
	for _, c := range n.Children {
		out = append(out, l.commaList(c)...)
	}
	return out
}

// evalForm lowers `eval`: a block child is EvalBlock, code the analysis can
// read; anything else is EvalStr, the soundness boundary whose result is
// Any. No operand means $_ (perl's Deparse gives `eval;` as `eval $_`).
func (l *lowerer) evalForm(n *parse.Node) Node {
	if n.Text != "eval" || len(n.Children) > 1 {
		return nil
	}
	if body := l.blockChild(n); body != nil {
		return &EvalBlock{Body: body}
	}
	return &EvalStr{Code: l.operand(n)}
}

// blockChild lowers n's one child if it is a block, and returns nil otherwise.
func (l *lowerer) blockChild(n *parse.Node) Node {
	if len(n.Children) == 1 && n.Children[0].Kind == parse.Block {
		return l.expr(n.Children[0])
	}
	return nil
}

// operand lowers n's first child, or returns nil when it has none.
func (l *lowerer) operand(n *parse.Node) Node {
	if len(n.Children) == 0 {
		return nil
	}
	return l.expr(n.Children[0])
}
