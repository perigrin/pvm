// ABOUTME: Canon writes a tree back out as Perl, parenthesised the way the TREE groups.
// ABOUTME: Its parens come from binding power, never from the ones the source happened to have.

package parse

import (
	"slices"
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

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
	var b canonBuf
	emit(&b, n, src, 0)
	return b.String()
}

// canonBuf is canon's output, with the offset of each `(` and `)` it wrote
// because a node's Paren flag is set. See CanonToken.
type canonBuf struct {
	strings.Builder
	parenNode []int
	// follow is the binding power of the operator written right after the
	// node emitted next, which a named unary's operand would absorb. emit
	// reads and clears it, so a parent sets it (emitBefore) only for the
	// child it is about to write; every other child reads 0.
	follow int
}

// writeParenNode writes a paren a node's Paren flag stands for, and marks it.
func (b *canonBuf) writeParenNode(c byte) {
	b.parenNode = append(b.parenNode, b.Len())
	b.WriteByte(c)
}

// writeFrom writes o's text from from to to, and the marks within it.
func (b *canonBuf) writeFrom(o *canonBuf, from, to int) {
	for _, at := range o.parenNode {
		if at >= from && at < to {
			b.parenNode = append(b.parenNode, b.Len()+at-from)
		}
	}
	b.WriteString(o.String()[from:to])
}

// emitBefore emits n with f, the power of the operator written after it.
func emitBefore(b *canonBuf, n *Node, src []byte, outer, f int) {
	b.follow = f
	emit(b, n, src, outer)
}

// emit writes n, wrapping it in parentheses when its own binding power is
// looser than outer -- the power its parent will read it at.
//
// A prefix op's parens depend on what FOLLOWS it as well, which outer does
// not carry: at a right operand, `!defined $x` and `$a . lc $b`, outer says
// nothing about the right edge. f is that power, 0 when a closer or nothing
// follows. A binary's left operand is followed by its operator, its right
// operand by whatever follows the binary, and so on down the right edge.
func emit(b *canonBuf, n *Node, src []byte, outer int) {
	f := b.follow
	b.follow = 0
	// A parenthesised element with a trailing separator keeps its parens
	// whatever its binding power, so the separator stays inside them:
	// `(A,)/0` is written `(A() ,) / 0`, not `A() , / 0`.
	if emitWraps(n, outer) {
		f = 0 // the `)` follows
		b.WriteByte('(')
		defer b.WriteByte(')')
	}
	// A trailing separator the source wrote goes after the node, inside any
	// parens added above: `f 1, || 2` is written `f(1 ,) || 2`. See
	// Node.TrailingComma.
	if n.TrailingComma != "" {
		defer b.WriteString(" " + n.TrailingComma)
	}

	switch n.Kind {
	case SourceFile:
		for _, c := range n.Children {
			emit(b, c, src, 0)
		}

	case Statement:
		var stmt canonBuf
		for _, c := range n.Children {
			if c.HeredocBody {
				continue // after the terminator; see below
			}
			// A bare block with no statement in it is written `{;}`. Written
			// `{}` it re-reads as an empty anonymous hash, which is what perl
			// makes of an empty brace at statement position; measured on
			// 5.42.0, `{;}` is an enterloop and `{}` an emptyavhv.
			if c.Kind == Block && !slices.ContainsFunc(c.Children, func(s *Node) bool {
				return s.Kind == Statement
			}) {
				stmt.WriteString("{;}")
				continue
			}
			emit(&stmt, c, src, 0)
		}
		text := stmt.String()
		// An expression statement whose emission begins with `{` begins with
		// an anonymous hash -- `({0})` or `({0}) . 1` with the parens gone --
		// and a bare `{` at statement position re-reads as a block unless
		// braceOpensAnonHash says otherwise (`{a => 1}`, `{}`). Where it
		// would, perl's Deparse writes `+{0};`, and so does this. The text is
		// re-parsed to ask, so the answer is the parser's own.
		if strings.HasPrefix(text, "{") && n.Children[0].Kind != Block && n.Children[0].Kind != Unknown &&
			rereadsAsBlock(text) {
			b.WriteByte('+')
		}
		// A NEWLINE ALREADY IN THE STATEMENT after its heredoc opener -- a
		// quote that runs on, `... eq q(hello\n)` -- is where perl reads the
		// bodies from: its scan_heredoc takes the lines after the opener's
		// line wherever they fall, and the quote resumes after them. So the
		// bodies go there, not after the terminator, or perl reads the
		// quote's own lines as the body. Measured on 5.42.0,
		// `<<E21 . 'single\nE21 content\nE21\nquoted'` is
		// "E21 content\n" . "single\nquoted".
		if bodies := heredocBodies(n); len(bodies) > 0 {
			if at := newlineAfterOpener(n, text); at >= 0 {
				b.writeFrom(&stmt, 0, at+1)
				for _, c := range bodies {
					b.WriteString(c.SourceText(src))
				}
				b.writeFrom(&stmt, at+1, len(text))
				if !blockForm(n) {
					b.WriteByte(';')
				} else {
					b.WriteByte(' ')
				}
				return
			}
		}
		b.writeFrom(&stmt, 0, len(text))
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
		// `require MODULE` keeps its bareword bare: written `require(mro)`
		// it would re-parse with `mro` as a call rather than a module name.
		if keywordName(n.Text) == "require" && len(n.Children) == 1 &&
			n.Children[0].Kind == Term && isBarewordText(n.Children[0].Text) {
			b.WriteString(n.Text)
			b.WriteByte(' ')
			b.WriteString(n.Children[0].Text)
			break
		}
		// Indirect object notation keeps its spelling, method then class,
		// and gains the argument parens every call gets: `new Foo "a"` is
		// written `new Foo("a")`. That re-parses to the same call -- the
		// class is a package this file declares or loads, which is what
		// made parseIndirect read it -- and it differs from the source only
		// in the parens Faithful already forgives. The arrow form perl's
		// Deparse writes would reorder the tokens, and fidelity would count
		// every correct indirect call as a misparse.
		if n.Indirect && len(n.Children) > 0 {
			b.WriteString(n.Text)
			b.WriteByte(' ')
			b.WriteString(n.Children[0].Text)
			emit(b, &Node{Kind: Call, Resolved: true, Children: n.Children[1:]}, src, 0)
			return
		}
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
		// Parenthesised, it is written as its source, `(A=>)`: the word
		// alone re-reads as a call.
		if n.Fat && len(n.Children) == 0 {
			if n.Paren {
				b.WriteString(n.SourceText(src))
				return
			}
			b.WriteString(n.Text)
			return
		}

		// A named unary op is written as the operator it is, measured on
		// perl 5.44.0: `chdir "t"` and `chdir("t")` build the same `chdir
		// vK/1`, so its call parens are not in perl's tree. They are written
		// only where the source wrote them (ArgParen), or where the operator
		// after it is tighter and would bind into the operand: `defined($x)
		// ** 2`. The operand is read at the named-unary level, so a looser
		// one is wrapped and a tree with `defined` over `||` writes
		// `defined ($x || $y)`, a paren the source lacks. The BLOCK forms of
		// `eval` and `do` are written below.
		if n.NamedUnary && !n.ArgParen && len(n.Children) == 0 {
			b.WriteString(n.Text)
			return
		}
		if n.NamedUnary && !n.ArgParen && len(n.Children) == 1 &&
			n.Children[0].Kind != Block && f <= bpNamedUnary {
			var arg canonBuf
			emitBefore(&arg, n.Children[0], src, bpNamedUnary+1, f)
			text := arg.String()
			b.WriteString(n.Text)
			// perl reads `NAME (` as the call's parens, so an operand written
			// with a leading `(` that closes before the operand ends would
			// re-read as the argument list: `defined ($x) . "a"` is
			// `defined($x) . "a"`. That operand takes the call parens.
			if leadingParenCloses(text) {
				b.WriteByte(' ')
				b.writeFrom(&arg, 0, len(text))
			} else {
				b.WriteByte('(')
				b.writeFrom(&arg, 0, len(text))
				b.WriteByte(')')
			}
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
			//
			// A sort comparator is the same shape: `sort foo @a` orders by foo,
			// and `sort foo, @a` sorts foo's result with the rest.
			if i > 0 && (n.Children[i-1].Handle || n.Children[i-1].Comparator) {
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
		if op == "not" && len(n.Children) == 1 {
			emitNot(b, n, src, f)
			break
		}
		b.WriteString(op)
		// A WORD operator needs a separator: `not foo` is a call to foo and
		// `notfoo` is the bareword "notfoo". The symbol operators must not
		// have one, because `\ $x` and `- $x` read differently in some
		// positions and a space there is never required.
		if isWordOp(op) {
			b.WriteByte(' ')
		}
		if len(n.Children) == 1 {
			// A dereference keeps its braces unless the operand is a plain
			// scalar variable: `$$name` reads only that, so `${$h->{k}}`
			// written bare is `${$h}->{k}`, a different program -- measured
			// on 5.42.0, one yields the value and the other dies with "Not a
			// SCALAR reference". A Block operand writes its own braces.
			if c := n.Children[0]; isDerefSigil(op) && c.Kind != Block && !bareDerefOperand(c) {
				b.WriteByte('{')
				emit(b, c, src, 0)
				b.WriteByte('}')
				break
			}
			// `-` or `+` before an operand that begins with the same byte
			// would write `--` or `++`, the decrement or increment: perl's
			// Deparse writes `- -$x` as `-(-$x)`, and so does this. `!`
			// before `~` would write `!~`, the negated match, which perl
			// refuses there; Deparse writes that one bare and so does not
			// re-read, so this parenthesises it the same way.
			if joins, ok := map[string]string{"-": "-", "+": "+", "!": "~"}[op]; ok {
				var operand canonBuf
				emitBefore(&operand, n.Children[0], src, prefix[op], f)
				// The operand is emitted once and its text reused: emitting it
				// again here doubles the work at every level of `!!!!$x`.
				// `-` before a one-letter name writes a filetest: `- A()` is
				// a negated call and `-A()` is `-A $_`, measured on 5.42.0.
				if text := operand.String(); strings.HasPrefix(text, joins) ||
					op == "-" && joinsFiletest(text) {
					b.WriteByte('(')
					b.writeFrom(&operand, 0, len(text))
					b.WriteByte(')')
				} else {
					b.writeFrom(&operand, 0, len(text))
				}
				break
			}
			emitBefore(b, n.Children[0], src, prefix[op], f)
		}

	case Postfix:
		if len(n.Children) == 1 {
			emitBefore(b, n.Children[0], src, infix[n.Text].BP, infix[n.Text].BP)
		}
		b.WriteString(n.Text)

	case Ternary:
		if len(n.Children) != 3 {
			b.WriteString(n.SourceText(src))
			return
		}
		// The middle part is closed by `:` as by a closer, so it gets 0:
		// `$c ? print "a" : print "b"` stays bare.
		bp := infix["?"].BP
		emitBefore(b, n.Children[0], src, bp+1, bp)
		b.WriteString(" ? ")
		emit(b, n.Children[1], src, bp)
		b.WriteString(" : ")
		emitBefore(b, n.Children[2], src, bp, f)

	case CmpChain:
		// N operands, each link with its own operator, evaluated once each
		// -- so it is emitted as the chain it is rather than as nested
		// binaries, which would say something the tree does not.
		bp := infix[n.Text].BP
		for i, c := range n.Children {
			if i > 0 {
				b.WriteByte(' ')
				b.WriteString(n.Ops[i-1])
				b.WriteByte(' ')
			}
			next := bp
			if i == len(n.Children)-1 {
				next = f
			}
			emitBefore(b, c, src, bp+1, next)
		}

	case List:
		// A List is parenthesised when the SOURCE grouped it, because that
		// grouping changes meaning -- `("a") x 3` is three elements and
		// `"a" x 3` is one (§4.10), and `my ($x) = f()` is not `my $x = f()`
		// (§4.12.2). This is the one place the source's paren is the tree's
		// own fact rather than decoration.
		if n.Paren {
			b.writeParenNode('(')
		}
		emitCommaSeparated(b, n, src)
		if n.Paren {
			b.writeParenNode(')')
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
		emitBefore(b, n.Children[0], src, atomBP, atomBP)
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
		//
		// The `.` line is closed with a newline when the source ended on it.
		// perl ends a format at a `.` followed by blanks and then a newline OR
		// the end of the file, so the source is right without one -- but the
		// statement separator written after this form would land on that
		// line, the re-lex would take it into the body, and every emission
		// would grow by a space.
		if n.Text == "format" {
			text := n.SourceText(src)
			b.WriteString(text)
			if !strings.HasSuffix(text, "\n") {
				b.WriteByte('\n')
			}
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
			//
			// A declaration's HEAD -- its prototype and its attributes -- is
			// separated by a space too, not an `=`. `sub x () { 8 }` emitted
			// `sub x = () {8;}` before this, which re-lexes as an assignment
			// to a bareword and Unknown could not see it: the tree was right
			// and only the emission was wrong, so the count stayed at zero.
			//
			// A SIGNATURE is part of the head as well, and arrives as the
			// parenthesised list parseSubDecl read it as. ` = ` there wrote
			// `sub f = ($a) {...}`, which perl rejects as an illegal
			// declaration. A one-element signature is that element's own
			// node with Paren set rather than a List. A leaf's span covers
			// the parentheses and its source text keeps them; a compound
			// node is rebuilt from its children and loses them, so they are
			// written here when missing: otherwise `sub f ($x = 1)` came
			// back as `sub f $x = 1`.
			switch {
			case i == 0 || c.Kind == Block ||
				c.Kind == PrototypeNode || c.Kind == Attribute:
				b.WriteByte(' ')
			case n.Text == "package" || n.Text == "class":
				// A name, a version, attributes and a block: no initialiser,
				// so nothing takes ` = `. `package Foo = 1.0;` is not Perl.
				b.WriteByte(' ')
			case i == 1 && n.Children[0].Kind == TypeName:
				// `my Foo $f`: the variable follows its class.
				b.WriteByte(' ')
			case (n.Text == "sub" || n.Text == "method") && c.Paren:
				b.WriteByte(' ')
				if c.Kind != List {
					var sig canonBuf
					emit(&sig, c, src, 0)
					if !strings.HasPrefix(sig.String(), "(") {
						b.WriteByte('(')
						b.writeFrom(&sig, 0, sig.Len())
						b.WriteByte(')')
					} else {
						b.writeFrom(&sig, 0, sig.Len())
					}
					continue
				}
			default:
				// The initialiser is the right side of an assignment and is
				// read at its power: `my $x = (%h and 1)` written bare is
				// `(my $x = %h) and 1`, measured on 5.42.0.
				op := "="
				if n.AssignOp != "" {
					op = n.AssignOp
				}
				b.WriteString(" " + op + " ")
				emitBefore(b, c, src, infix[op].BP, f)
				continue
			}
			emit(b, c, src, 0)
		}

	case Conditional, Loop:
		// Two forms share this Kind, with their children in OPPOSITE orders,
		// and the Modifier flag is which one this is.
		//
		// The STATEMENT MODIFIER form -- `$y = 1 if $x`, `$s += $_ foreach
		// 1..3` -- reads body first and condition second, and its keyword sits
		// BETWEEN them. Emitted by the block form's rule below it came back as
		// `if ($y = 1) $x`: the keyword moved to the front, the body
		// parenthesised as though it were the condition, the condition left
		// bare, and the `;` never written. That is not a parse of anything, and
		// it scored Unknown = 0 -- a swapped emission is not a refusal, so
		// nothing but canon could see it. 251 lines of perl.git t/ are this
		// shape.
		if n.Modifier {
			for i, c := range n.Children {
				next := 0 // the keyword follows the body
				if i > 0 {
					// The keyword separates the halves, and the condition
					// after it is a BARE expression -- `$y = 1 if $x` has no
					// parens and `$y = 1 if ($x)` merely has a parenthesised
					// expression as its condition, which the child's own Paren
					// flag already carries.
					b.WriteByte(' ')
					b.WriteString(n.Text)
					b.WriteByte(' ')
					next = f
				}
				emitBefore(b, c, src, 0, next)
			}
			// The terminator is the enclosing Statement's to write, and
			// blockForm is what decides it wants one -- see there for why a
			// modifier is not a block form even though it shares a Kind with
			// two that are.
			return
		}

		// `if (COND) BLOCK`, `while (COND) BLOCK`, and the rest. Text is the
		// keyword, kept as written -- `unless` is not rewritten into a
		// negated `if`, because the CST records what was typed.
		b.WriteString(n.Text)
		children := n.Children
		if n.CStyle {
			// `for (INIT; COND; STEP)`: the clauses lead the children in slot
			// order, and an empty slot is written as nothing between its
			// semicolons. The spelling is Deparse's on 5.42.0 -- `for ($a; ;
			// $c)`, `for ($a; $b;)`, `for (; ; $c)` -- each `;` followed by a
			// space except the last when its slot is empty.
			b.WriteString(" (")
			for i, filled := range n.ForSlots {
				if i > 0 {
					b.WriteByte(';')
					if i < len(n.ForSlots)-1 || filled {
						b.WriteByte(' ')
					}
				}
				if filled && len(children) > 0 {
					emit(b, children[0], src, 0)
					children = children[1:]
				}
			}
			b.WriteByte(')')
		}
		cond := !n.CStyle
		for _, c := range children {
			b.WriteByte(' ')
			switch {
			// A Block is the body. A nested Conditional is the `else` or
			// `elsif` -- it carries its own keyword and parens, so wrapping
			// it emits `(else {...})`. A foreach's loop variable sits
			// OUTSIDE the parens, declared or bare -- `for my $x (@l)`,
			// `for $x (@l)` -- and the parser marks it LoopVar. The Kind
			// cannot say so: a condition that declares, `if (my $x = f())`,
			// is a Declaration too, and so is a list whose one element is
			// `sub {}`.
			case c.Kind == Block || c.Kind == Conditional || c.LoopVar:
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

	case TypeName:
		b.WriteString(n.Text)

	case Phaser:
		// `BEGIN`, `END` and the rest, each with a block.
		b.WriteString(n.Text)
		for _, c := range n.Children {
			b.WriteByte(' ')
			emit(b, c, src, 0)
		}

	case LoopControl:
		// `last`, `next`, `redo` and `goto` with an optional label, and
		// `return` with an optional EXPRESSION.
		//
		// The label here is a TARGET, not a statement label: `next L` names
		// the loop to continue. Emitting the Label node's own form would
		// write `next L: `, which is a label on an empty statement -- so a
		// Label is written as its bare text.
		//
		// Everything else is EMITTED, not spelled. Writing `c.Text` for an
		// expression child deletes the expression and keeps its punctuation:
		// a Binary's Text is the operator alone, a parenthesised List's is
		// empty, a Ternary's is "?:" and a Call's is the callee name, so
		// every `return` with structure under it canon'd as a fragment --
		// `return ,;`, `return ;`, `return ?:;`, `return f;` -- at
		// Unknown = 0, which is issue 01a0e013. The parse was RIGHT all
		// along: the comma is level 8 / BP 80 and parseReturn reads at
		// bpListOp 70, so it binds and both operands are in the tree. Only
		// this loop threw them away.
		//
		// NO TERMINATOR. The `;` is the enclosing Statement's to write,
		// decided by blockForm, and every other non-block kind leaves it
		// there. Writing one here put a `;` in the MIDDLE of a
		// statement-modifier form -- `last if $x` came back as `last; if $x`
		// -- because a LoopControl is the modifier's BODY there, not a
		// statement of its own. That is the emission half of 01a0dee8.
		b.WriteString(n.Text)
		for _, c := range n.Children {
			b.WriteByte(' ')
			if c.Kind == Label {
				b.WriteString(c.Text)
				continue
			}
			emitBefore(b, c, src, 0, f)
		}

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
			// `AssocNone`'s comment in precedence.go says a binding power alone cannot
			// express nonassoc -- it stops the recursion but still accepts
			// the input -- and this is the second place that has to know.
			// Dropping the paren emits a syntax error:
			//
			//	(1 .. 2) .. 3    ->  1 .. 2 .. 3
			//	$ perl -e '1 .. 2 .. 3'
			//	syntax error, near "2 .."
			l, r = info.BP+1, info.BP+1
		}
		// A comparison that reached a Binary is a non-chaining one, and
		// perly.y's yyerror productions give it the same both-sides rule
		// (chain.go): `($a <=> $b) <=> 0` written bare is a syntax error.
		if c, _ := compareClass(n.Text); c != notComparison {
			l, r = info.BP+1, info.BP+1
		}
		// A fat comma autoquotes the word to its left, so `a => 1` is the
		// string "a" and writing `a()` calls it (§4.5.4). The flag form is
		// handled in emitCommaSeparated; inside a call's arguments the `=>`
		// is a Binary and the autoquote is a property of THIS operator.
		if n.Text == "=>" && n.Children[0].Kind == Call && len(n.Children[0].Children) == 0 {
			b.WriteString(n.Children[0].Text)
		} else {
			emitBefore(b, n.Children[0], src, l, info.BP)
		}
		b.WriteByte(' ')
		b.WriteString(n.Text)
		b.WriteByte(' ')
		// The parens of `->(` are the call's argument list, not grouping,
		// so they are written even where the argument alone would not need
		// them. Measured on 5.42.0, `$r->($v->[0])` is `&$r($v->[0])` and
		// `$r -> $v->[0]` is `$r->$v->[0]`, a method call through $v.
		if c := n.Children[1]; n.Text == "->" && c.Paren && src[c.Start] == '(' && !writesParens(c, r) {
			b.WriteByte('(')
			emit(b, c, src, 0)
			b.WriteByte(')')
			return
		}
		emitBefore(b, n.Children[1], src, r, f)

	default:
		// A parenthesised Term's span covers its parens, so its source text
		// writes them, and the outermost pair is the one its Paren flag
		// stands for: `(($x))` keeps both, and only the outer is marked.
		text := n.SourceText(src)
		if n.Kind == Term && n.Paren && strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")") {
			b.writeParenNode('(')
			b.WriteString(text[1 : len(text)-1])
			b.writeParenNode(')')
			break
		}
		b.WriteString(text)
	}
}

// emitNot writes n, a `not`, followed by an operator of power f.
//
// `not` is a term whose operand runs right until a looser `and`, `or` or
// `xor`, so the operand is read just above NOTOP -- a tree with `not` over
// `or` writes `not ($x or 1)` -- and only what follows decides whether the
// whole `not` needs grouping: measured on perl 5.44.0, `not $x == 1` is
// `not(($x == 1))`, so `==` over `not $x` is written `(not $x) == 1`. A
// `not` the source parenthesised keeps that pair wherever it stands.
//
// perl reads `not (` as a function of the parenthesised operand alone
// (term.go), tighter than anything after it: `not ($y) + 1` is
// `(!$y) + 1`. So a parenthesised operand is written with its parens and
// needs no grouping, and an operand whose emission begins with a `(` that
// closes early takes call parens: `not` over `(1) + 1` is `not((1) + 1)`,
// where `not (1) + 1` would be a different program.
func emitNot(b *canonBuf, n *Node, src []byte, f int) {
	c := n.Children[0]
	// With a TrailingComma, emit has written the source's pair already.
	if n.Paren && n.TrailingComma == "" || f > prefix["not"] && !c.Paren {
		b.WriteByte('(')
		defer b.WriteByte(')')
		f = 0 // the `)` follows
	}
	var arg canonBuf
	if c.Paren {
		emit(&arg, c, src, 0)
		text := arg.String()
		b.WriteString("not ")
		if strings.HasPrefix(text, "(") && leadingParenCloses(text) {
			b.writeFrom(&arg, 0, len(text))
			return
		}
		b.WriteByte('(')
		b.writeFrom(&arg, 0, len(text))
		b.WriteByte(')')
		return
	}
	emitBefore(&arg, c, src, prefix["not"]+1, f)
	text := arg.String()
	if strings.HasPrefix(text, "(") && !leadingParenCloses(text) {
		b.WriteString("not(")
		b.writeFrom(&arg, 0, len(text))
		b.WriteByte(')')
		return
	}
	b.WriteString("not ")
	b.writeFrom(&arg, 0, len(text))
}

// emitWraps reports whether emit parenthesises n when read at outer.
func emitWraps(n *Node, outer int) bool {
	return bindingPower(n) < outer || n.Paren && n.TrailingComma != ""
}

// leadingParenCloses reports whether text, an emitted operand, either does
// not begin with `(` or begins with a `(` that closes at its end.
func leadingParenCloses(text string) bool {
	if !strings.HasPrefix(text, "(") {
		return true
	}
	toks := significantTokens([]byte(text))
	end, ok := matchParen(toks, 0)
	return ok && end == len(toks)-1
}

// writesParens reports whether emit writes a parenthesised node's parens
// itself at outer: by wrapping it, as a parenthesised List, as a Term,
// which is written as its source and whose span covers them, or as a
// parenthesised `not`, whose pair emitNot writes.
func writesParens(n *Node, outer int) bool {
	return emitWraps(n, outer) || n.Kind == List || n.Kind == Term ||
		n.Kind == Unary && n.Text == "not" && n.Paren
}

// emitCommaSeparated writes n's children with ", " between them.
//
// The bracketed forms hold their elements directly -- unlike a call, whose
// arguments arrive as one Binary "," child that emits its own separator.
func emitCommaSeparated(b *canonBuf, n *Node, src []byte) {
	for i, c := range n.Children {
		if i > 0 {
			// Fat is set on the element BEFORE the `=>`, and the two commas
			// are not interchangeable: `(a => 1)` is the string "a" where
			// `(a, 1)` is a call to a (§4.5.4). Emitting the plain comma
			// changes what the list contains.
			//
			// An element that kept its separators (TrailingComma) has
			// already written them; only a space follows.
			switch {
			case n.Children[i-1].TrailingComma != "":
				b.WriteByte(' ')
			case n.Children[i-1].Fat:
				b.WriteString(" => ")
			default:
				b.WriteString(", ")
			}
		}
		// Every element but the last is followed by its comma, which a
		// `not` would take into its operand: `[(not $x), 1]`.
		next := 0
		if i < len(n.Children)-1 {
			next = infix[","].BP
		}
		emitBefore(b, c, src, 0, next)
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
		// A dereference is a term: a subscript after it applies to it, so
		// `$$r[0]` is written bare, not as the list slice `($$r)[0]`. So is
		// a named parameter's `:$name`, which is one parameter. So is `not`
		// (perly.y's `term: NOTOP listexpr`): nothing before it splits it,
		// and emitNot groups it by what follows.
		if isDerefSigil(n.Text) || n.Text == ":" || n.Text == "not" {
			return atomBP
		}
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
	case Declaration:
		// A variable declaration WITH an initialiser is an assignment, and
		// binds like one. At an atom's power it was never parenthesised, so
		// `(my $s = "abc") =~ /x/` re-emitted as `my $s = "abc" =~ /x/` --
		// the match bound to "abc" first, a different program, at Unknown=0.
		// The initialiser is either a second child (`my $x = 1`) or an
		// assignment Binary the target list was parsed into (`my ($a, $b) =
		// @_`).
		if declarators[keywordName(n.Text)] && len(n.Children) > 0 {
			kids := n.Children
			if len(kids) > 1 && kids[0].Kind == TypeName {
				kids = kids[1:]
			}
			last := kids[len(kids)-1:]
			if len(kids) > 1 ||
				len(last) == 1 && last[0].Kind == Binary && isAssignment(infix[last[0].Text]) {
				return infix["="].BP
			}
		}
		return atomBP
	default:
		return atomBP
	}
}

// isDerefSigil reports whether a Unary's operator is a sigil applied to an
// expression: `${...}`, `@{...}`, `%{...}`, `*{...}`, `&{...}`.
func isDerefSigil(op string) bool {
	switch op {
	case "$", "@", "%", "*", "&", "$#":
		return true
	}
	return false
}

// bareDerefOperand reports whether a dereference's operand may follow its
// sigil without braces: a scalar variable, or a scalar dereference of one --
// `@$r`, `$$$r`.
//
// The variable's name must be one the lexer reads after a second sigil -- a
// word byte, `{`, `$` or `::`, toke.c's test -- or the bare form re-reads as
// something else: `${$;}` written `$$;` is the pid and a terminator.
func bareDerefOperand(c *Node) bool {
	switch {
	case c.Kind == Term && strings.HasPrefix(c.Text, "$") && !c.Paren:
		rest := c.Text[1:]
		return rest != "" && (isWordByte(rest[0]) || rest[0] == '{' || rest[0] == '$' ||
			strings.HasPrefix(rest, "::"))
	case c.Kind == Unary && c.Text == "$" && len(c.Children) == 1:
		return bareDerefOperand(c.Children[0])
	}
	return false
}

// isBarewordText reports whether s is spelled as a bareword: a name, maybe
// package-qualified, and not a variable, number or string.
func isBarewordText(s string) bool {
	if s == "" {
		return false
	}
	c := s[0]
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// unaryToken spells a Unary node's operation back as its source token.
//
// Node.Text records the OPERATION -- "neg", "pos", "ref" -- because `-$x` and a
// subtraction are not the same thing and the tree must not say they are. The
// precedence table is keyed by the token, so emitting needs the mapping back.
func unaryToken(op string) string {
	switch op {
	case "neg":
		return "-"
	case "pos":
		return "+"
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

// joinsFiletest reports whether a `-` written before text lexes as a
// filetest: a filetest letter with no word byte after it, as the lexer's
// scanFileTest reads it.
func joinsFiletest(text string) bool {
	return text != "" && lexer.IsFileTest("-"+text[:1]) &&
		(len(text) == 1 || !isWordByte(text[1]))
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
// heredocBodies is the statement's heredoc bodies, in opener order: first the
// ones that fell inside its span, then the ones after its terminator. Every
// inner body precedes every trailing one in the source, so the order holds.
// newlineAfterOpener is the offset, in a statement's emitted text, of the
// first newline after its first heredoc opener, or -1 when the opener's line
// runs to the statement's end.
//
// ponytail: the opener is found by its spelling, so a string holding the
// same `<<TERM` earlier in the statement is taken for it; emit offsets would
// be the upgrade.
func newlineAfterOpener(n *Node, text string) int {
	opener := firstHeredocOpener(n)
	if opener == "" {
		return -1
	}
	at := strings.Index(text, opener)
	if at < 0 {
		return -1
	}
	nl := strings.IndexByte(text[at:], '\n')
	if nl < 0 {
		return -1
	}
	return at + nl
}

// firstHeredocOpener is the spelling of the first `<<TERM` under n.
func firstHeredocOpener(n *Node) string {
	if n.Kind == Term && strings.HasPrefix(n.Text, "<<") && len(n.Children) == 0 {
		return n.Text
	}
	for _, c := range n.Children {
		if c.HeredocBody {
			continue
		}
		if got := firstHeredocOpener(c); got != "" {
			return got
		}
	}
	return ""
}

func heredocBodies(n *Node) []*Node {
	out := append([]*Node{}, n.InnerHeredocBodies...)
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

// rereadsAsBlock reports whether a statement's text, written bare, begins
// with a bare block. ponytail: a re-parse per statement that starts with an
// anonymous hash, which is rare; a text-level braceOpensAnonHash if it shows.
func rereadsAsBlock(text string) bool {
	root := Parse([]byte(text))
	return len(root.Children) > 0 && len(root.Children[0].Children) > 0 &&
		root.Children[0].Children[0].Kind == Block
}

func blockForm(n *Node) bool {
	last := lastForm(n)
	if last == nil {
		return false
	}
	switch last.Kind {
	case Conditional, Loop:
		// The block forms end in a `}` and take no `;`. The STATEMENT
		// MODIFIER form shares their Kind and ends in an expression, so it
		// takes one: `$y = 1 if $x` needs its terminator and `if ($x) {...}`
		// must not grow one. Reading the Kind alone dropped it from every
		// modifier statement, which is half of why they canon'd as something
		// that is not Perl.
		return !last.Modifier
	case Phaser:
		// A phaser name with no block is a forward declaration, `sub END;`
		// to perl's Deparse, and takes its `;` like any other.
		return len(last.Children) > 0
	case Block, Use, Unknown:
		return true
	case LoopControl:
		// NOT a block form. `last`, `next`, `redo`, `goto` and `return` all end
		// in a label or an expression and all take a `;` -- they were listed
		// above only because the LoopControl case used to write its own
		// terminator, and writing one there put a `;` inside `last if $x`.
		//
		// The canon of a bare one loses its trailing space as a result:
		// `{return;}` rather than `{return; }`, which is what every other
		// non-block statement already gets -- `{$x = 1;}` is the same rule --
		// and perl accepts it. Measured on 5.42.0,
		// `perl -e 'sub f {return;} sub g {return 1;}'` is syntax OK.
		return false
	case Declaration:
		d := last
		// A lexical sub is a declarator wrapping the sub declaration, so the
		// sub's own shape decides: `my sub f { 1 }` ends in its block.
		for len(d.Children) == 1 && d.Children[0].Kind == Declaration {
			d = d.Children[0]
		}
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
