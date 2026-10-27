// ABOUTME: Lowering of the special forms of §4.14.2: MapGrepSort, do, eval, the anonymous constructors, Ref,
// ABOUTME: FileTest, Readline, Glob, LoopEx, Wantarray and Match, each told apart by CST kind, name and flags.

package lower

import (
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

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
// A block lowers through expr like any other child, to a Block; the Unknown
// of an unclosed brace lowers to a Hole carrying its id, so the slot is
// never empty.
func (l *lowerer) mapGrepSort(n *parse.Node) Node {
	if n.Text != "map" && n.Text != "grep" && n.Text != "sort" {
		return nil
	}
	out := &MapGrepSort{Op: n.Text}
	rest := n.Children
	if len(rest) > 0 {
		switch c := rest[0]; {
		case c.Kind == parse.Block || c.Refused() && c.Refusal == parse.UnclosedBrace:
			out.Block, rest = l.expr(c), rest[1:]
		case c.Comparator:
			out.FirstExpr, rest = l.expr(c), rest[1:]
		}
	}
	out.List = l.elements(rest)
	if out.Block == nil && out.FirstExpr == nil && n.Text != "sort" && len(out.List) > 0 {
		out.FirstExpr, out.List = out.List[0], out.List[1:]
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

// doForm lowers `do`: a block child is DoBlock, anything else DoFile.
func (l *lowerer) doForm(n *parse.Node) Node {
	if n.Text != "do" || len(n.Children) > 1 {
		return nil
	}
	if body := l.blockChild(n); body != nil {
		return &DoBlock{Body: body}
	}
	return &DoFile{Path: l.operand(n)}
}

// wantarray lowers the builtin `wantarray`, which takes no operand.
func (l *lowerer) wantarray(n *parse.Node) Node {
	if n.Text != "wantarray" || len(n.Children) != 0 {
		return nil
	}
	return &Wantarray{}
}

// fileTest lowers a file test, `-e $f`: the parser gives it as a call named
// by the operator. A nil Operand means $_.
func (l *lowerer) fileTest(n *parse.Node) Node {
	t := n.Text
	if len(t) != 2 || t[0] != '-' || !strings.ContainsRune(fileTests, rune(t[1])) || len(n.Children) > 1 {
		return nil
	}
	return &FileTest{Op: t[1], Operand: l.operand(n)}
}

// fileTests are the letters toke.c reads as a file test after `-` (§4.15.1).
const fileTests = "rwxoRWXOezsfdlpSbcugktTBAMC"

// globCall lowers `glob EXPR`. A nil Pattern means $_.
func (l *lowerer) globCall(n *parse.Node) Node {
	if n.Text != "glob" || len(n.Children) > 1 {
		return nil
	}
	return &Glob{Pattern: l.operand(n)}
}

// angle lowers a `<...>` term. toke.c's rule (§4.15.2): an empty body, a
// handle name or a simple scalar is a readline, anything else a glob. `<>`
// and `<<>>` read ARGV, which Magic records, and have no Handle.
//
// The handle and the pattern are bytes inside the one term, with no parse
// node of their own, so each takes the term's id under RoleAngleBody. A
// handle is a Bareword or, for `<$fh>`, a Var; a pattern is an
// interpolating StrLit, since perl reads `<$x/*>` as `glob $x . '/*'`.
func (l *lowerer) angle(n *parse.Node) Node {
	t := n.Text
	if len(n.Children) != 0 || len(t) < 2 || t[0] != '<' || t[len(t)-1] != '>' {
		return nil
	}
	if t == "<<>>" {
		return &Readline{Magic: true}
	}
	body := t[1 : len(t)-1]
	switch {
	case body == "":
		return &Readline{Magic: true}
	case handleName.MatchString(body):
		var h Node = &Bareword{Name: body}
		if body[0] == '$' {
			h = &Var{Sigil: "$", Name: body[1:]}
		}
		return &Readline{Handle: l.derive(n, RoleAngleBody, h)}
	}
	return &Glob{Pattern: l.derive(n, RoleAngleBody, &StrLit{Value: body, Interpolated: true})}
}

// handleName matches a readline's body: a bare handle name or a simple
// scalar, package-qualified or not.
var handleName = regexp.MustCompile(`^\$?[A-Za-z_]\w*(::\w+)*$`)

// ref lowers `\EXPR`, which the parser gives as `unary "ref"`. Its `&foo`
// operand names the sub: `\&foo` is a code reference, not a call.
func (l *lowerer) ref(n *parse.Node) Node {
	if n.Text != "ref" || len(n.Children) != 1 {
		return nil
	}
	return &Ref{Operand: l.codeName(n.Children[0])}
}

// arrayLiteral lowers `[...]`.
func (l *lowerer) arrayLiteral(n *parse.Node) Node {
	return &ArrayLiteral{Elems: l.elems(n)}
}

// hashLiteral lowers `{...}`, at statement start too: the parser has
// already read that brace as perl does.
func (l *lowerer) hashLiteral(n *parse.Node) Node {
	return &HashLiteral{Elems: l.elems(n)}
}

// elems lowers each child of n as a list element.
func (l *lowerer) elems(n *parse.Node) []Node {
	var out []Node
	for _, c := range n.Children {
		out = append(out, l.expr(c))
	}
	return out
}

// loopEx lowers a loop_control: last, next, redo, goto, and return, whose
// value stands in Label. A label is a Bareword with the label node's id,
// and goto's `&foo` names the sub it transfers to.
func (l *lowerer) loopEx(n *parse.Node) Node {
	out := &LoopEx{Op: n.Text}
	if len(n.Children) == 1 {
		c := n.Children[0]
		switch {
		case c.Kind == parse.Label:
			out.Label = l.stamp(c, &Bareword{Name: c.Text})
		case n.Text == "goto":
			out.Label = l.codeName(c)
		default:
			out.Label = l.expr(c)
		}
	}
	return out
}

// match lowers a match, substitution, transliteration or qr term to a Match
// with no Target, which means $_. Pattern, Replace and Flags stay inside the
// one quote token (§4.14.2, PARSER).
func (l *lowerer) match(n *parse.Node) Node {
	if len(n.Children) != 0 {
		return nil
	}
	if op := matchOp(n.Text); op != "" {
		return &Match{Op: op}
	}
	return nil
}

// matchOp returns a match term's quote operator, `m` for a bare /.../, or ""
// when the text is not a match.
func matchOp(t string) string {
	if strings.HasPrefix(t, "/") {
		return "m"
	}
	if m := quoteMatch.FindStringSubmatch(t); m != nil {
		return m[1]
	}
	return ""
}

// quoteMatch matches the head of an m, s, tr, y or qr term up to its opening
// delimiter.
var quoteMatch = regexp.MustCompile(`^(m|s|tr|y|qr)\s*[^\w\s]`)

// bind lowers `=~` and `!~`. Over a match term it is that Match with its
// Target set. A bound qr term is a match, Op `m`: perl 5.42 compiles
// `$x =~ qr/b/` to match(), not to a qr constructor. Over anything else -- a pattern in a variable, or an Unknown
// -- it stays a Binary, because there is no quote token to carry.
func (l *lowerer) bind(n *parse.Node) Node {
	if n.Text != "=~" && n.Text != "!~" || len(n.Children) != 2 {
		return nil
	}
	lhs, rhs := n.Children[0], n.Children[1]
	if rhs.Kind == parse.Term && len(rhs.Children) == 0 {
		if op := matchOp(rhs.Text); op != "" {
			if op == "qr" {
				op = "m"
			}
			return &Match{Op: op, Target: l.expr(lhs), Negated: n.Text == "!~"}
		}
	}
	return &Binary{Op: n.Text, L: l.expr(lhs), R: l.expr(rhs)}
}
