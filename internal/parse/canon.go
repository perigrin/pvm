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
		start := b.Len()
		for _, c := range n.Children {
			emit(b, c, src, 0)
		}
		// A statement whose emission already ends in `;` or `}` takes no
		// terminator: `use strict;` carries its own, and `if (...) { }` is a
		// block form that never had one.
		//
		// Decided from the emitted TEXT rather than from a list of block-like
		// kinds, because that list drifts -- every kind added later would have
		// to be remembered here, and forgetting one writes a semicolon the
		// source never had.
		if s := b.String()[start:]; !endsWith(s, ';') && !endsWith(s, '}') {
			b.WriteByte(';')
		}

	case Trivia:
		// Dropped. See Canon.

	case Call:
		// Always parenthesised, and that is the whole case chapter 7 names.
		// `print (1+2)*3` is `print(1+2) * 3`, and a tree that read it as
		// `print((1+2)*3)` has the same leaves in the same order -- so the
		// ARGUMENT BOUNDARY is the only thing that tells the two apart, and
		// writing it explicitly is what makes the difference visible.
		//
		// A call's arguments are one child: the comma is a Binary operator,
		// not a separator this case supplies. Writing ", " here as well
		// would emit `f($x , 1)`.
		b.WriteString(n.Text)

		// `eval { ... }` is the BLOCK form and takes no parens. Wrapping it
		// as `eval({...})` re-parses to a different tree -- an anonymous
		// hash in argument position rather than a block -- so the paren
		// would be the emitter inventing a construct the source never had.
		if len(n.Children) == 1 && n.Children[0].Kind == Block {
			b.WriteByte(' ')
			emit(b, n.Children[0], src, 0)
			return
		}

		b.WriteByte('(')
		for _, c := range n.Children {
			emit(b, c, src, 0)
		}
		b.WriteByte(')')

	case Unary:
		// Text is the OPERATION, not the source token: `-$x` is "neg" and
		// `\$x` is "ref". The table is keyed by the token, so the operator
		// has to be spelled back out.
		b.WriteString(unaryToken(n.Text))
		if len(n.Children) == 1 {
			emit(b, n.Children[0], src, prefix[unaryToken(n.Text)])
		}

	case Postfix:
		if len(n.Children) == 1 {
			emit(b, n.Children[0], src, infix[n.Text].BP)
		}
		b.WriteString(n.Text)

	case Ternary:
		if len(n.Children) != 3 {
			b.WriteString(n.SourceText(src))
			return
		}
		bp := infix["?"].BP
		emit(b, n.Children[0], src, bp+1)
		b.WriteString(" ? ")
		emit(b, n.Children[1], src, bp)
		b.WriteString(" : ")
		emit(b, n.Children[2], src, bp)

	case CmpChain:
		// N operands, one operator, evaluated once each -- so it is emitted
		// as the chain it is rather than as nested binaries, which would say
		// something the tree does not.
		bp := infix[n.Text].BP
		for i, c := range n.Children {
			if i > 0 {
				b.WriteByte(' ')
				b.WriteString(n.Text)
				b.WriteByte(' ')
			}
			emit(b, c, src, bp+1)
		}

	case List:
		// A List is parenthesised when the SOURCE grouped it, because that
		// grouping changes meaning -- `("a") x 3` is three elements and
		// `"a" x 3` is one (§4.10), and `my ($x) = f()` is not `my $x = f()`
		// (§4.12.2). This is the one place the source's paren is the tree's
		// own fact rather than decoration.
		if n.Paren {
			b.WriteByte('(')
		}
		emitCommaSeparated(b, n, src)
		if n.Paren {
			b.WriteByte(')')
		}

	case AnonArray, AnonHash:
		open, shut := byte('['), byte(']')
		if n.Kind == AnonHash {
			open, shut = '{', '}'
		}
		b.WriteByte(open)
		emitCommaSeparated(b, n, src)
		b.WriteByte(shut)

	case Index:
		// Text is the bracket that opened it; Arrow says whether it was
		// reached through `->`. The two read DIFFERENT VARIABLES -- `$h{k}`
		// is the hash %h and `$h->{k}` is the hashref $h -- so the arrow has
		// to survive into the emission.
		if len(n.Children) != 2 {
			b.WriteString(n.SourceText(src))
			return
		}
		emit(b, n.Children[0], src, atomBP)
		if n.Arrow {
			b.WriteString("->")
		}
		b.WriteString(n.Text)
		emit(b, n.Children[1], src, 0)
		b.WriteByte(closer(n.Text))

	case Block:
		b.WriteByte('{')
		for _, c := range n.Children {
			emit(b, c, src, 0)
		}
		b.WriteByte('}')

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

// emitCommaSeparated writes n's children with ", " between them.
//
// The bracketed forms hold their elements directly -- unlike a call, whose
// arguments arrive as one Binary "," child that emits its own separator.
func emitCommaSeparated(b *strings.Builder, n *Node, src []byte) {
	for i, c := range n.Children {
		if i > 0 {
			b.WriteString(", ")
		}
		emit(b, c, src, 0)
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
	case Binary, CmpChain:
		return infix[n.Text].BP
	case Ternary:
		return infix["?"].BP
	case Unary:
		return prefix[unaryToken(n.Text)]
	case Postfix:
		return infix[n.Text].BP
	case List:
		// An unparenthesised list is a comma expression and binds like one;
		// a parenthesised one is already grouped and cannot be split.
		if !n.Paren {
			return infix[","].BP
		}
		return atomBP
	default:
		return atomBP
	}
}

// unaryToken spells a Unary node's operation back as its source token.
//
// Node.Text records the OPERATION -- "neg", "ref" -- because `-$x` and a
// subtraction are not the same thing and the tree must not say they are. The
// precedence table is keyed by the token, so emitting needs the mapping back.
func unaryToken(op string) string {
	switch op {
	case "neg":
		return "-"
	case "ref":
		return "\\"
	default:
		return op
	}
}

// closer is the bracket that closes the one an Index opened.
func closer(open string) byte {
	switch open {
	case "[":
		return ']'
	case "{":
		return '}'
	default:
		return ')'
	}
}

// atomBP is above every level in the table: nothing can split an atom.
const atomBP = 1000

// endsWith reports whether s ends in c, ignoring trailing whitespace.
func endsWith(s string, c byte) bool {
	for i := len(s) - 1; i >= 0; i-- {
		switch s[i] {
		case ' ', '\t', '\n', '\r':
			continue
		}
		return s[i] == c
	}
	return false
}
