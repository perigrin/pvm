// ABOUTME: Canon writes a tree back out as Perl, parenthesised the way the TREE groups.
// ABOUTME: Its parens come from binding power, never from the ones the source happened to have.

package parse

import "strings"

// Canon emits n as Perl, with the parenthesisation the TREE implies.
//
// This is the whole reason the check is worth running. An emitter that
// printed each leaf's original text in order would be SourceText with extra
// steps: it round-trips by construction and proves nothing. Canon writes its
// own grouping, so a tree that grouped differently from its source produces
// a paren the source does not contain -- and that is what the comparison
// sees.
//
// Trivia is dropped on purpose: comments and POD belong to no statement and
// carry no grouping, so they cannot disagree with the tree. Callers compare
// significant tokens, not bytes.
func Canon(n *Node, src []byte) string {
	var b strings.Builder
	emit(&b, n, src, 0)
	return b.String()
}

// emit writes n, wrapping it in parentheses when its own binding power is
// looser than outer -- the power its parent will read it at.
func emit(b *strings.Builder, n *Node, src []byte, outer int) {
	if bindingPower(n) < outer {
		b.WriteByte('(')
		defer b.WriteByte(')')
	}

	switch n.Kind {
	case SourceFile:
		for _, c := range n.Children {
			emit(b, c, src, 0)
		}

	case Statement:
		for _, c := range n.Children {
			emit(b, c, src, 0)
		}
		b.WriteByte(';')

	case Trivia:
		// Dropped. See Canon.

	case Call:
		// Always parenthesised, and that is the whole case chapter 7 names.
		// `print (1+2)*3` is `print(1+2) * 3`, and a tree that read it as
		// `print((1+2)*3)` has the same leaves in the same order -- so the
		// ARGUMENT BOUNDARY is the only thing that tells the two apart, and
		// writing it explicitly is what makes the difference visible.
		b.WriteString(n.Text)
		b.WriteByte('(')
		for i, c := range n.Children {
			if i > 0 {
				b.WriteString(", ")
			}
			emit(b, c, src, 0)
		}
		b.WriteByte(')')

	case Binary:
		info := infix[n.Text]
		if len(n.Children) != 2 {
			b.WriteString(n.SourceText(src))
			return
		}
		// The operand on the side associativity does NOT favour must be
		// parenthesised at equal power; the favoured side must not be. Left
		// associative favours the left, so the right operand is the one
		// that needs BP+1 to clear.
		//
		// This is rightBP() read backwards, which is the point: the emitter
		// asks the same question the parser did, and gets the same answer.
		l, r := info.BP, info.BP+1
		if info.Assoc == AssocRight {
			l, r = info.BP+1, info.BP
		}
		emit(b, n.Children[0], src, l)
		b.WriteByte(' ')
		b.WriteString(n.Text)
		b.WriteByte(' ')
		emit(b, n.Children[1], src, r)

	default:
		b.WriteString(n.SourceText(src))
	}
}

// bindingPower is the power a node is read at by whatever contains it.
//
// A fact about NODES, which is why it lives here and not in precedence.go:
// that table records perly.y's operator levels, and most node kinds are not
// operators at all. A Term or a parenthesised list binds tighter than any
// operator can pull apart, so they answer with the maximum.
func bindingPower(n *Node) int {
	switch n.Kind {
	case Binary:
		return infix[n.Text].BP
	default:
		return atomBP
	}
}

// atomBP is above every level in the table: nothing can split an atom.
const atomBP = 1000
