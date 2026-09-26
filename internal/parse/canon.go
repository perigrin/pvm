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
			if c.HeredocBody {
				continue // after the terminator; see below
			}
			emit(b, c, src, 0)
		}
		// A HEREDOC BODY GOES AFTER THE TERMINATOR, ON ITS OWN LINE, and
		// both halves of that are required rather than cosmetic. Perl reads
		// a body from the line FOLLOWING the opener's line, so
		// `<<"EOT";body\nEOT\n` on one line leaves the lexer looking for the
		// body on the next -- which holds `EOT`, and the terminator is then
		// consumed as the body of an unterminated heredoc.
		//
		// Emitted here rather than in child order for the same reason. In
		// the tree the body is a child of the statement that opened it; in
		// the text it follows the `;` that closed it. Canon supplies that
		// `;`, so the split is made here, where both are in hand.
		if bodies := heredocBodies(n); len(bodies) > 0 {
			if !blockForm(n) {
				b.WriteByte(';')
			}
			b.WriteByte('\n')
			for _, c := range bodies {
				b.WriteString(c.SourceText(src))
			}
			return
		}
		// Only a BLOCK FORM goes without a terminator. Decided from the last
		// child's kind, not from whether the text ends in `}`: that `}` is
		// genuinely ambiguous, and reading it as "block form" drops the
		// semicolon from every expression statement that happens to end in a
		// brace.
		//
		//	$x = sub { 1 }; $y = 2;   ->  $x = sub {1;}$y = 2;
		//	$x ||= {}; print 1;       ->  $x ||= {}print(1);
		//
		// Both are syntax errors, and they were 404 statements across 218 of
		// T1's 986 files. A kind list can drift; a wrong rule cannot be made
		// right.
		if !blockForm(n) {
			b.WriteByte(';')
			return
		}
		// A block form takes no semicolon, so it needs a space instead --
		// `sub f {1;}f();` runs the next statement into the brace and does
		// not parse.
		b.WriteByte(' ')

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
		// A fat comma autoquotes the word to its left, so this is a STRING
		// and writing `a()` would call it. The parser records the autoquote
		// on the element rather than rewriting the node, because hover and
		// go-to-definition still want to know a word was written there.
		if n.Fat && len(n.Children) == 0 {
			b.WriteString(n.Text)
			return
		}

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

		// `WORD BLOCK ARG` for a word this parser could not resolve: the block
		// and then the list, with NO parens around either. perl spells it that
		// way and accepts only that spelling -- measured on 5.42.0:
		//
		//	$ perl -MO=Deparse -e 'zzz { 1 } print "b";'
		//	zzz {
		//	    1
		//	} print('b');
		//	$ perl -e 'zzz({1;}print "b");'
		//	syntax error at -e line 1, near ";}"
		//
		// So parenthesising this one emits bytes perl rejects, and the reader
		// agrees with perl: `zzz({1;}print("b"))` does not re-parse.
		//
		// Resolved is what separates this from `map({...} @a)` and
		// `print({$fh;} "x")`, which perl DOES accept parenthesised and which
		// canon must keep emitting that way. A resolved callee's block fills a
		// slot its builtin declares; an unresolved one's is the bare
		// `WORD BLOCK` form, whose reading perl decides from the symbol table.
		if !n.Resolved && len(n.Children) > 1 && n.Children[0].Kind == Block {
			b.WriteByte(' ')
			for _, c := range n.Children {
				emit(b, c, src, 0)
			}
			return
		}

		b.WriteByte('(')
		for i, c := range n.Children {
			// A filehandle slot takes no comma after it -- `print $fh "x"`
			// is the handle and then the list, and `print $fh, "x"` prints
			// the handle AS an argument. Only a space separates them, and
			// omitting it emits `print($fh"x")`, which is stable under the
			// fixpoint and still wrong.
			if i > 0 && n.Children[i-1].Handle {
				b.WriteByte(' ')
			}
			emit(b, c, src, 0)
		}
		b.WriteByte(')')

	case Unary:
		// Text is the OPERATION, not the source token: `-$x` is "neg" and
		// `\$x` is "ref". The table is keyed by the token, so the operator
		// has to be spelled back out.
		op := unaryToken(n.Text)
		b.WriteString(op)
		// A WORD operator needs a separator: `not foo` is a call to foo and
		// `notfoo` is the bareword "notfoo". The symbol operators must not
		// have one, because `\ $x` and `- $x` read differently in some
		// positions and a space there is never required.
		if isWordOp(op) {
			b.WriteByte(' ')
		}
		if len(n.Children) == 1 {
			emit(b, n.Children[0], src, prefix[op])
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

	case Declaration:
		// `format NAME = <body>` is emitted AS WRITTEN, before the rest,
		// because there is nothing here a canonical form could normalise.
		// The body is opaque bytes whose only structure is where it starts
		// and stops, and both are whitespace: scanFormatBody begins at the
		// newline after the `=` and ends at a line holding a lone `.`. So
		// every separator in this form is load-bearing, and re-spacing it is
		// how it stops being a format -- measured, the shared ` = ` and a
		// trailing `;` re-lexed as `format STDOUT;` with the body adrift.
		//
		// No terminator either: the `.` line already ended the declaration,
		// which is why blockForm counts this among the forms that write their
		// own.
		if n.Text == "format" {
			b.WriteString(n.SourceText(src))
			return
		}

		// `my`, `our`, `local`, `state`, `sub` and `package`. The keyword,
		// then the children -- a target and optionally an initialiser or a
		// body. An `=` initialiser arrives as a Binary already holding the
		// target, so it is not written here.
		b.WriteString(n.Text)
		for i, c := range n.Children {
			// `my $x = 1` arrives as TWO children, target and initialiser,
			// with the `=` consumed by parseVarDecl rather than left in the
			// tree (decl.go:90-95). Joining them with a space alone emits
			// `my $x 1`, so the operator has to be written back.
			//
			// A `sub` body is the exception: `sub f { }` has a name and a
			// Block and no assignment between them.
			switch {
			case i == 0 || c.Kind == Block:
				b.WriteByte(' ')
			default:
				b.WriteString(" = ")
			}
			emit(b, c, src, 0)
		}

	case Conditional, Loop:
		// `if (COND) BLOCK`, `while (COND) BLOCK`, and the rest. Text is the
		// keyword, kept as written -- `unless` is not rewritten into a
		// negated `if`, because the CST records what was typed.
		b.WriteString(n.Text)
		cond := true
		for _, c := range n.Children {
			b.WriteByte(' ')
			switch {
			// A Block is the body. A nested Conditional is the `else` or
			// `elsif` -- it carries its own keyword and parens, so wrapping
			// it emits `(else {...})`. A Declaration is a foreach's loop
			// variable, which sits OUTSIDE the parens: `for my $x (@l)`.
			case c.Kind == Block || c.Kind == Conditional || c.Kind == Declaration:
				emit(b, c, src, 0)
			case cond:
				// The condition, and only it, is parenthesised: here the
				// parens are grammar rather than grouping, so they are
				// written whether or not the tree grouped anything.
				cond = false
				b.WriteByte('(')
				emit(b, c, src, 0)
				b.WriteByte(')')
			default:
				emit(b, c, src, 0)
			}
		}

	case Use:
		// `use`, `no` or `require`: the keyword, the module, and whatever
		// import list followed. Parsed, not executed.
		b.WriteString(n.Text)
		for _, c := range n.Children {
			b.WriteByte(' ')
			emit(b, c, src, 0)
		}
		b.WriteByte(';')

	case Phaser:
		// `BEGIN`, `END` and the rest, each with a block.
		b.WriteString(n.Text)
		for _, c := range n.Children {
			b.WriteByte(' ')
			emit(b, c, src, 0)
		}

	case LoopControl:
		// `last`, `next` or `redo`, with an optional label.
		//
		// The label here is a TARGET, not a statement label: `next L` names
		// the loop to continue. Emitting the Label node's own form would
		// write `next L: `, which is a label on an empty statement.
		b.WriteString(n.Text)
		for _, c := range n.Children {
			b.WriteByte(' ')
			b.WriteString(c.Text)
		}
		b.WriteByte(';')

	case Label:
		b.WriteString(n.Text)
		b.WriteString(": ")

	case Unknown:
		// The one kind whose own text IS the answer. Unknown spans exactly
		// the bytes this parser could not read, and it has no structure to
		// emit -- writing the source back is what a refusal looks like.
		b.WriteString(n.SourceText(src))

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
		switch info.Assoc {
		case AssocRight:
			l, r = info.BP+1, info.BP
		case AssocNone:
			// Neither side may repeat at equal power, so BOTH need a paren.
			// `precedence.go:14-18` says a binding power alone cannot
			// express nonassoc -- it stops the recursion but still accepts
			// the input -- and this is the second place that has to know.
			// Dropping the paren emits a syntax error:
			//
			//	(1 .. 2) .. 3    ->  1 .. 2 .. 3
			//	$ perl -e '1 .. 2 .. 3'
			//	syntax error, near "2 .."
			l, r = info.BP+1, info.BP+1
		}
		// A fat comma autoquotes the word to its left, so `a => 1` is the
		// string "a" and writing `a()` calls it (§4.5.4). The flag form is
		// handled in emitCommaSeparated; inside a call's arguments the `=>`
		// is a Binary and the autoquote is a property of THIS operator.
		if n.Text == "=>" && n.Children[0].Kind == Call && len(n.Children[0].Children) == 0 {
			b.WriteString(n.Children[0].Text)
		} else {
			emit(b, n.Children[0], src, l)
		}
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
			// Fat is set on the element BEFORE the `=>`, and the two commas
			// are not interchangeable: `(a => 1)` is the string "a" where
			// `(a, 1)` is a call to a (§4.5.4). Emitting the plain comma
			// changes what the list contains.
			if n.Children[i-1].Fat {
				b.WriteString(" => ")
			} else {
				b.WriteString(", ")
			}
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

// isWordOp reports whether an operator is spelled as a word, and so needs a
// space before its operand.
func isWordOp(op string) bool {
	c := op[len(op)-1]
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
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

// blockForm reports whether a statement ends in a block and so takes no
// semicolon.
//
// The kinds that write their own terminator are included: `use strict;` and
// `next L;` emit theirs in their own case, and a second one here would be a
// stray token.
// heredocBodies is the statement's heredoc-body children, in opener order.
func heredocBodies(n *Node) []*Node {
	var out []*Node
	for _, c := range n.Children {
		if c.HeredocBody {
			out = append(out, c)
		}
	}
	return out
}

// lastForm is the statement child whose kind decides whether a terminator is
// wanted: the last one that is not a heredoc body.
//
// A body is always last in child order and never decides the form -- it is
// the statement's VALUE, arriving late. Reading it as the last child made
// every heredoc statement a non-block form, which is right by accident for
// `my $h = <<EOT;` and wrong for `if ($c) { print <<EOT; }`.
func lastForm(n *Node) *Node {
	for i := len(n.Children) - 1; i >= 0; i-- {
		if !n.Children[i].HeredocBody {
			return n.Children[i]
		}
	}
	return nil
}

func blockForm(n *Node) bool {
	last := lastForm(n)
	if last == nil {
		return false
	}
	switch last.Kind {
	case Conditional, Loop, Phaser, Block, Use, LoopControl, Unknown:
		return true
	case Declaration:
		d := last
		// `format NAME = ... .` ended at its own `.` line, which perl accepts
		// with no `;` after it. A semicolon here lands INSIDE the next
		// format body's scan or ahead of the next statement, and either way
		// the emission stops being a fixpoint.
		if d.Text == "format" {
			return true
		}
		// `sub f { }` is a block form; `my $x = 1` is not.
		return len(d.Children) > 0 && d.Children[len(d.Children)-1].Kind == Block
	}
	return false
}
