// ABOUTME: The nud side of the Pratt parser: §4.4's term forms, one per shape.
// ABOUTME: The backslash of §4.4.5 is here — it is the srefgen the fidelity harness measures.

package parse

import (
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// parseTerm parses one term: a literal, a variable, a prefix operator and its
// operand, a parenthesised list, or an anonymous constructor.
//
// Anything it cannot read becomes Unknown spanning what it consumed. The
// plan's rule applies inside expressions as much as at statement level: a
// guessed operand produces a tree that is not a parse of its source.
func (p *parser) parseTerm() *Node {
	tok, ok := p.peekSignificant()
	if !ok {
		// Nothing left to read. Returning an Unknown here would span no
		// bytes, and an empty node is indistinguishable from a loop that did
		// not advance -- which is the invariant the fuzzer checks. The
		// caller handles a nil term.
		return nil
	}
	text := p.text(tok)

	// A filetest: `-e $f`, `-d $dir`. perl returns UNIOP for these
	// (toke.c:6255 FTST), so they bind exactly like a named unary.
	//
	// `lexer.IsFileTest` recognises the token, which arrives whole: the whole
	// `-e` is one Operator, because the minus is part of the operator's name
	// and a lexer that left a bare `-` in front of a word had not yet decided
	// which of perl's three readings of `-` it meant.
	//
	// The decision does NOT need the parser's position, which is what this
	// code assumed while it was gluing two tokens back together. Measured
	// 5.42.0, there is no subtraction reading to choose between:
	//
	//	$ perl -e 'my $x = 1 -e "/etc";'   syntax error near "1 -e "
	//
	// perl forms `-e` there too and then has nowhere to put it. So the lexer
	// decides it on the two bytes alone and this reads the result.
	//
	// THE PAREN CLIFF APPLIES, §4.8.1, and it is not decoration. `-e($f)`
	// spells the argument list with parens, and a branch that handed `($f)`
	// to parseExpr got a paren LIST as the child -- so canon wrote its own
	// parens around one the source already had, `-e(($f))`, and added another
	// on every pass. That breaks the canon fixpoint rather than merely
	// looking odd, and `open_mode_whitespace.t` is where it was caught.
	if lexer.IsFileTest(text) {
		p.advanceTo(tok)
		n := &Node{
			Kind: Call, Text: text, Resolved: true, Start: tok.Start,
			End: tok.End,
		}
		if next, ok := p.peekSignificant(); ok && p.text(next) == "(" {
			p.advanceTo(next)
			if arg := p.parseCallArgs(); arg != nil {
				n.Children = append(n.Children, arg)
			}
			if close, ok := p.peekSignificant(); ok && p.text(close) == ")" {
				p.advanceTo(close)
			}
			n.End = p.prevEnd()
			return n
		}
		// No parens: one argument at level 19, so arithmetic binds into the
		// test and comparison does not. An ABSENT argument is legal --
		// `$_="/etc"; print -e;` deparses as `print -e $_`, so the node keeps
		// the operator's own span and has no child.
		//
		// The absence needs a BRANCH, which is what this comment used to deny.
		// `parseExpr` does not return nil at a terminator: it returns an
		// Unknown spanning whatever it found, so the `;` became the argument
		// and canon emitted `print(-e(;));`. In the `if` form the whole BLOCK
		// was absorbed -- `if (-e){print 1}` canon'd as `if (-e(){print(1)}))`
		// with the parens left unbalanced, which is a wrong SHAPE rather than
		// a wrong count.
		//
		// A terminator is anything that cannot begin a term: the statement's
		// `;`, a closer (`)` of the `if`, `}` of a `grep` block, `]`), end of
		// input, or an INFIX OPERATOR. Testing the token is narrower than
		// testing the result, because an Unknown that spans a REAL operand
		// must stay in the tree.
		//
		// The infix case is not an edge: perl supplies `$_` there too, and
		// both of these are ordinary in its own suite --
		//
		//	perl -MO=Deparse -e '$_="/etc"; -e or die;'
		//	  ->  die unless -e $_;
		//	perl -MO=Deparse -e '$_="/etc"; print -e ? 1 : 0;'
		//	  ->  print -e $_ ? 1 : 0;
		//
		// A first version of this guard stopped at `;` and closers alone and
		// left both of those refusing, which is why the set is the `infix`
		// table rather than a hand-listed pair of kinds.
		next, ok := p.peekSignificant()
		if !ok || next.Kind == lexer.Semicolon || next.Kind == lexer.CloseBracket {
			return n
		}
		if _, isInfix := infix[p.text(next)]; isInfix {
			return n
		}
		if arg := p.parseExpr(bpNamedUnary); arg != nil {
			n.Children = append(n.Children, arg)
			n.End = p.prevEnd()
		}
		return n
	}

	// Prefix operators. `\` is one of these, and §4.4.5 is not optional
	// scaffolding: it is the reference perl reports with srefgen, which is
	// the signal the whole fidelity harness was built to measure.
	if bp, isPrefix := prefix[text]; isPrefix && p.prefixAllowed(text) {
		p.advanceTo(tok)
		operand := p.operand(bp, tok)
		return &Node{
			Kind: Unary, Text: prefixName(text),
			Start: tok.Start, End: operand.End,
			Children: []*Node{operand},
		}
	}

	// A glob: `*foo`, `*{$name}`, `*$glob`. A term, not multiplication -- and
	// the parser knows which because parseTerm is only called where a term
	// is expected.
	if text == "*" {
		// A name that is not an identifier: `*1`, `*^R`, `*-`. perl accepts
		// any variable name here -- measured on 5.42.0, `local *a = *1;`,
		// `*^R = *foo;` and `*X = *-;` each deparse as written -- and the
		// lexer emits the star and the name apart, so the glob is assembled
		// here. Only TOUCHING the star: a digit, `^` and a letter, or a
		// single punctuation byte.
		if glob := p.globWithSpecialName(tok); glob != nil {
			return glob
		}
		if name, ok := p.peekAfter(tok); ok &&
			(name.Kind == lexer.Word || name.Kind == lexer.Variable ||
				p.text(name) == "{") {
			p.advanceTo(tok)
			if p.text(name) == "{" {
				// The braces of `*{EXPR}` GROUP; they do not construct.
				// Calling parseTerm here read the `{` as an anonymous hash,
				// so the tree said a hash was being BUILT where a
				// symbol-table slot is being named -- a wrong tree that
				// round-trips.
				//
				// Consumed here rather than delegated, for the same reason
				// the `${EXPR}` deref does it: parseTerm cannot know the
				// brace is a group, because in every other position it is
				// not.
				p.advanceTo(name)
				inner := p.parseExpr(0)
				if close, ok := p.peekSignificant(); ok && p.text(close) == "}" {
					p.advanceTo(close)
				}
				n := &Node{
					Kind: Term, Text: "*", Start: tok.Start, End: p.prevEnd(),
				}
				if inner != nil {
					n.Children = append(n.Children, inner)
				}
				return n
			}
			p.advanceTo(name)
			if name.Kind == lexer.Variable {
				// `*$glob` names the slot the SCALAR points at, so the
				// variable is an operand and has to be in the tree. Folding
				// it into Text as `*NAME` does would put `$glob` back out of
				// it -- the same leaf this issue exists to stop making.
				//
				//	$ perl -MO=Deparse -e 'my $g = \*STDOUT; my $x = *$g{IO};'
				//	my $x = *$g{'IO'};
				return &Node{
					Kind: Term, Text: "*",
					Start: tok.Start, End: name.End,
					Children: []*Node{{
						Kind: Term, Text: p.text(name),
						Start: name.Start, End: name.End,
					}},
				}
			}
			return &Node{
				Kind: Term, Text: "*" + p.text(name),
				Start: tok.Start, End: name.End,
			}
		}
	}

	switch {
	case text == "(":
		return p.parseParenList(tok)
	case text == "[":
		return p.parseBracketed(tok, "]", AnonArray)
	case text == "{":
		// In term position a `{` is an anonymous hash -- one of the five
		// oracle markers. Whether it is instead a block is the statement
		// level's decision (§4.9.2), and the lexer's brace stack already
		// classified this one as a term.
		return p.parseBracketed(tok, "}", AnonHash)
	}

	switch tok.Kind {
	case lexer.DerefSigil:
		// A sigil applied to an expression: `${$h->{k}}`, `@{[ 1, 2 ]}`,
		// `$$x`. The lexer emits the sigil alone precisely so the interior
		// reaches the parser; making a leaf of it here would put the names
		// inside back out of the tree.
		//
		// A Unary rather than a new kind: a dereference IS a prefix operator
		// applied to one operand, and the sigil is in Text. §4.14 names it
		// `Deref{Sigil, Expr}`, which is this shape with the kind spelled
		// out; the lowering (chapter 6 §6.1.6) is where that rename belongs,
		// because a new CST kind is a change every consumer must learn.
		p.advanceTo(tok)

		// The braces of `${EXPR}` GROUP; they do not construct. Letting
		// parseTerm see the `{` made an AnonHash of it, so the tree said a
		// hash was being built around the very expression being
		// dereferenced.
		if open, ok := p.peekSignificant(); ok && p.text(open) == "{" {
			// The braces hold a BLOCK in perly.y (`indirob: ... | block`),
			// and almost always one expression. A `;` -- first, or after
			// that expression -- says there are statements: `${; do { ... }
			// }`, `${no strict; \$_}`. Measured on 5.42.0, Deparse keeps
			// both as blocks. They are read as one then, and canon's Block
			// writes the braces the sigil needs.
			if p.derefHoldsStatements(open) {
				blk := p.parseBlock(open)
				return &Node{
					Kind: Unary, Text: text,
					Start: tok.Start, End: blk.End,
					Children: []*Node{blk},
				}
			}
			p.advanceTo(open)
			inner := p.parseExpr(0)
			if close, ok := p.peekSignificant(); ok && p.text(close) == "}" {
				p.advanceTo(close)
			}
			if inner == nil {
				// `${}` -- nothing to dereference. The bytes stay in the
				// tree as an Unknown rather than a Unary with no operand.
				return &Node{Kind: Unknown, Refusal: EmptyDeref, Start: tok.Start, End: p.prevEnd()}
			}
			return &Node{
				Kind: Unary, Text: text,
				Start: tok.Start, End: p.prevEnd(),
				Children: []*Node{inner},
			}
		}

		operand := p.operand(bpDeref, tok)
		return &Node{
			Kind: Unary, Text: text,
			Start: tok.Start, End: operand.End,
			Children: []*Node{operand},
		}

	case lexer.FuncSigil:
		// `&` in term position introduces a function name, and the lexer
		// emits it as its own token so that `&f` and `&&` stay
		// distinguishable. The name is the next token, so the two are joined
		// here into one term: `\&f` is a code reference, not a reference to
		// an ampersand.
		p.advanceTo(tok)
		if name, ok := p.peekSignificant(); ok && p.ampTakes(name) {
			p.advanceTo(name)
			return &Node{
				Kind: Term, Text: text + p.text(name),
				Start: tok.Start, End: name.End,
			}
		}
		return &Node{Kind: Term, Text: text, Start: tok.Start, End: tok.End}

	case lexer.Word:
		text := p.text(tok)
		switch {
		case declarators[keywordName(text)] && p.declaratorTakesTarget(tok):
			// A declaration in EXPRESSION position: `open my $fh, $p`,
			// `f(my $x)`. perly.y makes `my` a named unary at level 19
			// (§4.6), so it is a term here as much as a statement form.
			//
			// 50 of T1's files first fail on `open my $fh, ...`, which is
			// the idiomatic three-argument open and appears in almost every
			// file that touches a filehandle.
			return p.parseVarDeclNoSemi(tok)

		case text == "sub":
			// An anonymous sub: `sub { ... }` with no name. Distinguished
			// from a declaration by what follows -- a `{` rather than a
			// name -- which is the same test perl makes.
			//
			// An ATTRIBUTE may stand between: `my $f = sub :lvalue { 1 }`
			// has no name either, and a `:` here cannot be anything else --
			// `sub` is a keyword, so the colon is not a label's and not a
			// ternary's.
			//
			// So may a prototype or signature: `sub (&) { ... }`, `sub ($x)
			// { ... }`.
			if next, ok := p.peekAfter(tok); ok && (p.text(next) == "{" || p.text(next) == ":" ||
				next.Kind == lexer.Prototype || p.text(next) == "(") {
				return p.parseAnonSub(tok)
			}

		case text == "eval" || text == "do":
			// `eval BLOCK` and `do BLOCK` take a block, not an expression.
			// The named-unary path would parse the `{` as an anonymous hash.
			if next, ok := p.peekAfter(tok); ok && p.text(next) == "{" {
				return p.parseBlockOperator(tok)
			}
		}
		return p.parseWordTerm(tok)

	case lexer.Variable, lexer.Number, lexer.Quote,
		lexer.Readline, lexer.HeredocOpen:
		p.advanceTo(tok)
		return &Node{
			Kind: Term, Text: text,
			Start: tok.Start, End: tok.End,
		}
	}

	// Not a term. Consumed so the loop advances; the bytes stay in the tree.
	p.advanceTo(tok)
	return &Node{Kind: Unknown, Refusal: NotATerm, Start: tok.Start, End: tok.End}
}

// declaratorTakesTarget reports whether a declarator word is introducing a
// variable rather than being used as an identifier.
//
// `my`, `our`, `state` and `field` are ordinary barewords in most positions:
// `$h{field}` is a hash key, `f(state => 1)` is a fat-comma pair. Treating
// every occurrence as a declaration made `$h{field}` parse its key as one --
// caught by TestClassSyntax, which asserts exactly that.
//
// A declarator takes a target when a VARIABLE or a `(` follows it. Nothing
// else is a declaration: `my $x`, `my @a`, `my ($a, $b)`, and that is the
// whole grammar (§4.4.3).
func (p *parser) declaratorTakesTarget(word lexer.Token) bool {
	next, ok := p.peekAfter(word)
	if !ok {
		return false
	}
	return next.Kind == lexer.Variable || p.text(next) == "("
}

// derefHoldsStatements reports whether the brace at open, after a deref
// sigil, holds statements rather than one expression: anything but one
// expression running to the closing `}`. A statement form such as `use
// strict` is no expression at all, and its refusal skips past the `;`, so
// the test is what the expression reaches rather than whether a `;` follows
// it. The cursor is left where it was.
func (p *parser) derefHoldsStatements(open lexer.Token) bool {
	save := p.pos
	defer func() { p.pos = save }()
	p.advanceTo(open)
	first, ok := p.peekSignificant()
	if !ok || p.text(first) == "}" {
		return false // `${}`: EmptyDeref's, below.
	}
	if first.Kind == lexer.Semicolon {
		return true
	}
	p.parseExpr(0)
	tok, ok := p.peekSignificant()
	return !ok || p.text(tok) != "}"
}

// parseAnonSub: `sub { ... }` and `sub ($x) { ... }` with no name.
//
// A Declaration with no name child, so consumers that walk declarations see
// it without a second kind to learn. Whether a sub is named is a question
// about its children, not about what kind of node it is.
func (p *parser) parseAnonSub(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	// A prototype or signature, if the lexer found one.
	if proto, ok := p.peekSignificant(); ok && proto.Kind == lexer.Prototype {
		p.advanceTo(proto)
		n.Children = append(n.Children, &Node{
			Kind: PrototypeNode, Text: p.text(proto),
			Start: proto.Start, End: proto.End,
		})
	} else if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
		// A signature: the lexer leaves the parens as code when the feature
		// is on, as it does after a named sub.
		p.parseSignature(n)
	}

	// Attributes: `my $f = sub :lvalue { 1 }`. Same list, same position, and
	// the same reason parseSubDecl reads them -- an unread attribute ends the
	// sub at the colon and leaves the whole body trailing.
	p.parseAttributes(n)

	// A signature after the attributes, the order the feature requires:
	// `sub :lvalue ($x) { $x }`. See parseSubDecl.
	if !hasHead(n) {
		if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
			p.parseSignature(n)
		}
	}

	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}
	n.End = p.prevEnd()
	return n
}

// parseBlockOperator: `eval BLOCK` and `do BLOCK`, spec §4.7.
//
// These take a BLOCK where an expression would be, so the named-unary path
// cannot handle them -- it would parse the `{` as an anonymous hash, which is
// what `eval { $x; };` did before this.
//
// `eval EXPR` and `do EXPR` are different operators with the same spelling,
// and they go through parseWordTerm as ordinary named unaries. The `{` is
// what separates them, which is also how perl decides.
func (p *parser) parseBlockOperator(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Call, Text: p.text(word), Resolved: true, Start: word.Start}

	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}
	n.End = p.prevEnd()
	return n
}

// prefixAllowed keeps an infix operator from being read as a prefix one.
//
// `-` and `+` are in both tables, and which they are depends on position:
// after a term they are ADDOP, at the start of one they are UMINUS. The
// parser knows its own position, so it does not need the lexer's expect
// state for this -- parseTerm is only ever called where a term is expected.
func (p *parser) prefixAllowed(text string) bool {
	switch text {
	case "-", "+", "\\", "!", "~", "~.", "++", "--", "not":
		return true
	}
	return false
}

// globWithSpecialName reads `*1`, `*^R` or `*-` at the star tok, or returns
// nil. See the glob branch of parseTerm.
func (p *parser) globWithSpecialName(star lexer.Token) *Node {
	name, ok := p.peekAfter(star)
	if !ok || name.Start != star.End {
		return nil
	}
	// last is the final token of the name: the name itself, or the letter
	// after a caret.
	var last lexer.Token
	switch text := p.text(name); {
	case name.Kind == lexer.Number && isAllDigits(text):
		last = name
	case text == "^":
		letter, ok := p.peekAfter(name)
		if !ok || letter.Start != name.End || letter.Kind != lexer.Word {
			return nil
		}
		last = letter
	case name.Kind == lexer.Operator && len(text) == 1 && strings.IndexByte("-+/!&@]", text[0]) >= 0:
		last = name
	default:
		return nil
	}
	p.advanceTo(last)
	return &Node{
		Kind: Term, Text: string(p.src[star.Start:last.End]),
		Start: star.Start, End: last.End,
	}
}

// isAllDigits reports whether s is one or more ASCII digits.
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// prefixName distinguishes the unary spelling from the infix one in the
// tree, so `- $a` and `$a - $b` do not both read as "-".
func prefixName(text string) string {
	switch text {
	case "-":
		return "neg"
	case "+":
		return "pos"
	case "\\":
		return "ref"
	}
	return text
}

// parseParenList: `(...)`. A single expression in parens is that expression;
// a comma-separated one is a List.
func (p *parser) parseParenList(open lexer.Token) *Node {
	p.advanceTo(open)

	var items []*Node
	for {
		tok, ok := p.peekSignificant()
		if !ok {
			break
		}
		if p.text(tok) == ")" {
			p.advanceTo(tok)
			break
		}
		// An empty slot: a comma where an element would start. perl drops it,
		// measured on 5.42.0 -- `my @a = (1,,,2)` has 2 elements -- and
		// op/for-many.t spells forty in one variable list. Only AFTER an
		// element: a leading `(,1)` is a syntax error in perl, so that refusal
		// stands.
		//
		// A fat comma in that slot is empty the same way -- `(foo, => 1)` is
		// `(foo(), 1)`, measured -- and it quotes nothing: the word it could
		// have quoted is behind the comma.
		if (p.text(tok) == "," || p.text(tok) == "=>") && len(items) > 0 {
			p.advanceTo(tok)
			continue
		}
		// Parsed above the comma so each element is its own node. A nil
		// element means the input ran out mid-list, which the loop's next
		// peek handles.
		if item := p.parseExpr(infix[","].BP); item != nil {
			items = append(items, item)
		}

		next, ok := p.peekSignificant()
		if !ok {
			break
		}
		switch p.text(next) {
		case ",", "=>":
			// A fat comma quotes the word to its LEFT (§4.5.4), so the flag
			// belongs to the element already read rather than the next one.
			// Without it `(a => 1)` and `(a, 1)` are the same tree, and the
			// lowering cannot tell the string "a" from a call to a().
			if p.text(next) == "=>" && len(items) > 0 {
				items[len(items)-1].Fat = true
			}
			p.advanceTo(next)
		case ")":
			p.advanceTo(next)
			return p.finishList(items, open)
		default:
			// Neither a separator nor the closer. The three word operators
			// below the comma get here: `and`, `or` and `xor` are levels 4
			// and 5, the comma is 8, so the element loop above -- which
			// parses at the comma's power -- stops without consuming them
			// and every `($a and $b)` refused with `not_a_term`.
			//
			// perly.y puts them above the comma, so the paren holds a full
			// `expr` and the list assembled so far is that expr's LEFT
			// operand. Measured on perl 5.42.0:
			//
			//	$ perl -MO=Deparse -e 'my @x = (1, 2 and 3);'
			//	my(@x) = ('???', 2) && 3;
			//
			// So the loop is resumed with the list in hand rather than
			// restarted, and the closer is then taken by the code below.
			if p.atOperatorBelowComma() {
				items = []*Node{p.parseInfix(p.finishList(items, open), 0)}
				if c, ok := p.peekSignificant(); ok && p.text(c) == ")" {
					p.advanceTo(c)
				}
				return p.finishList(items, open)
			}
			// Anything else: stop rather than spin.
			return p.finishList(items, open)
		}
	}
	return p.finishList(items, open)
}

func (p *parser) finishList(items []*Node, open lexer.Token) *Node {
	if len(items) == 1 {
		// `($x)` is `$x`, but the span covers the parens so round-trip holds
		// -- and the PAREN itself is recorded, because two of Perl's rules
		// turn on it. §4.10: `("a") x 3` gives three elements where
		// `"a" x 3` gives one. §4.12.2: with `sub f {(1,2,3)}`,
		// `my ($x) = f()` is 1 and `my $y = f()` is 3. Both measured on
		// perl 5.42.0.
		//
		// Copying field by field rather than dereferencing the node: the
		// span must widen to cover the parens, so this cannot alias. Every
		// flag is carried across -- dropping one here would erase an Arrow
		// on `($h->{k})`, which is the bug this function already had for
		// Paren.
		n := items[0]
		return &Node{
			Kind: n.Kind, Text: n.Text,
			Start: open.Start, End: p.prevEnd(),
			Children: n.Children,
			Resolved: n.Resolved,
			Arrow:    n.Arrow,
			Fat:      n.Fat,
			Handle:   n.Handle,
			Paren:    true,
		}
	}
	return &Node{
		Kind: List, Start: open.Start, End: p.prevEnd(),
		Children: items,
		Paren:    true,
	}
}

// commaList is the element list as ONE node, for a caller that must hand it
// to parseInfix as a left operand.
//
// One element is itself: `[$a and $b]` has a single `$a` on the operator's
// left, and Concise shows no list op there at all. Several become a List,
// which is the same node parseParenList's finishList builds -- but without
// its Paren flag, because a bracket is not a paren and the two rules that
// turn on the flag (§4.10's `("a") x 3`, §4.12.2's list-vs-scalar context)
// do not apply to a constructor.
//
// An empty list cannot reach parseInfix as a left operand, so the caller is
// told to leave the operator alone -- `[and $b]` is a syntax error in perl and
// inventing an operand for it would be a worse tree than declining.
func (p *parser) commaList(items []*Node) *Node {
	if len(items) == 0 {
		return nil
	}
	if len(items) == 1 {
		return items[0]
	}
	return &Node{
		Kind: List, Start: items[0].Start, End: p.prevEnd(),
		Children: items,
	}
}

// parseBracketed: `[...]` and `{...}` in term position.
func (p *parser) parseBracketed(open lexer.Token, closer string, kind Kind) *Node {
	p.advanceTo(open)

	var items []*Node
	for {
		tok, ok := p.peekSignificant()
		if !ok {
			break
		}
		if p.text(tok) == closer {
			p.advanceTo(tok)
			break
		}
		if item := p.parseExpr(infix[","].BP); item != nil {
			items = append(items, item)
		}

		next, ok := p.peekSignificant()
		if !ok {
			break
		}
		switch p.text(next) {
		case ",", "=>":
			// Same rule inside `[...]` and `{...}`: the fat comma marks the
			// element to its left. An anon hash written `{a => 1}` and one
			// written `{a, 1}` mean the same thing to perl but not to a
			// consumer reading the key, which §4.5.4 says is a string in the
			// first and a call in the second.
			if p.text(next) == "=>" && len(items) > 0 {
				items[len(items)-1].Fat = true
			}
			p.advanceTo(next)
		case closer:
			p.advanceTo(next)
			return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
		default:
			// The same resume parseParenList's default arm describes, and for
			// the same reason: a bracket holds a full `expr`, not a comma
			// list, so the three word operators below the comma stop the
			// element loop and take what it built as their LEFT operand.
			// Measured on perl 5.42.0:
			//
			//	$ perl -MO=Deparse -e 'my $y = [1, 2 and 3];'
			//	my $y = [('???', 2) && 3];
			//
			// Without it the operand ESCAPED the bracket -- `[$a and $b]`
			// gave `(and (my $y [$a]) $b)`, a tree in which `and` sits above
			// the declaration with a one-element arrayref on its left. The
			// source has both operands inside the brackets, so that was not
			// a parse of it.
			//
			// The left operand is the ELEMENT LIST, not a nested constructor:
			// Concise shows one `anonlist` for the whole `[...]` with the
			// `and` inside it, over a plain `list` op.
			//
			//	9  <@> anonlist sK*/1
			//	7     <|> and(other->8) lK/1
			//	6        <@> list sK
			//	8        <$> const[IV 3] s
			//
			// So the resume hands parseInfix the items assembled so far and
			// the single result becomes the constructor's only element.
			if left := p.commaList(items); left != nil && p.atOperatorBelowComma() {
				items = []*Node{p.parseInfix(left, 0)}
				if c, ok := p.peekSignificant(); ok && p.text(c) == closer {
					p.advanceTo(c)
				}
			}
			return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
		}
	}
	return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
}

// ampTakes reports whether `&` joins this token as the name of the sub it
// calls.
//
// A Word always: `&f`. A SCALAR variable too, because `&$coderef` is perl's
// older calling convention and is how perl's own suite spells a call that
// passes the caller's `@_` implicitly. Measured on 5.42.0, inside a sub called
// with (1,2):
//
//	&$s;     the callee sees @_ as (1,2)
//	&$s();   the callee sees @_ empty
//
// The other sigils and the subscripted forms are REFUSED, because perl refuses
// them and agreement is the point:
//
//	&@a;      Bareword found where operator expected
//	&$s[0];   syntax error at -e line 1, near "$s["
//	&$s{k};   syntax error at -e line 1, near "$s{k"
//
// Joining a Variable unconditionally made all three parse clean, which is a
// wrong tree rather than a wider one -- so the arity of the check is the fix,
// not an ornament on it.
func (p *parser) ampTakes(name lexer.Token) bool {
	if name.Kind == lexer.Word {
		return true
	}
	if name.Kind != lexer.Variable {
		return false
	}
	if text := p.text(name); text == "" || text[0] != '$' {
		return false
	}
	// A subscript after the scalar has no call reading in perl, so the `&`
	// must not swallow the name and leave the bracket stranded.
	if next, ok := p.peekAfter(name); ok {
		if t := p.text(next); t == "[" || t == "{" {
			return false
		}
	}
	return true
}
