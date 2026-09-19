// ABOUTME: Calls: named unaries take one argument, list operators take the whole list.
// ABOUTME: An undeclared callee is Call{Resolved:false}, never Unknown — §4.8.3.

package parse

import "tamarou.com/pvm/internal/lexer"

// Binding powers for the two call shapes, spec §4.2.
//
// A named unary is level 19: tighter than comparison at 17, looser than
// arithmetic at 22. That is why `length $x + 1` is `length($x + 1)` and
// `length $x < 5` is `length($x) < 5` -- both measured on the optree.
//
// A list operator is level 7, below the comma at 8, which is exactly how it
// swallows the whole list.
const (
	bpNamedUnary = 190
	bpListOp     = 70

	// A dereference binds tighter than every infix operator, `->` at level
	// 29 included, so its operand is the braced expression or the single
	// variable and nothing more.
	//
	// `$$x[0]` is `${$x}[0]` -- the subscript applies to the DEREFERENCE,
	// not to `$x` -- so the sigil must take its operand before any postfix
	// gets a chance. Parsing at 300 leaves `[0]` to the caller's led loop,
	// which then wraps the whole Unary in an Index. Measured on perl 5.42.0:
	//
	//	$ perl -MO=Deparse -e 'my $r = [7]; print $$r[0];'
	//	print $r->[0];
	//
	// Deparse prints the arrow form, which is the same operation spelled the
	// other way -- and is why §4.14 gives both one node with an `Arrow` flag.
	bpDeref = 300
)

// parseWordTerm turns a bareword in term position into a call, a bareword
// term, or a declaration-like keyword the caller handles.
func (p *parser) parseWordTerm(word lexer.Token) *Node {
	text := p.text(word)

	// A niladic builtin takes nothing: `time`, `wantarray`.
	if niladicParse[text] {
		p.advanceTo(word)
		return &Node{
			Kind: Call, Text: text, Resolved: true,
			Start: word.Start, End: word.End,
		}
	}

	p.advanceTo(word)
	n := &Node{Kind: Call, Text: text, Start: word.Start}

	// The paren cliff, §4.8.1. A `(` immediately after the name makes this a
	// FUNC1 -- the parens delimit the arguments and nothing beyond them
	// belongs to the call. toke.c's UNI3 returns FUNC1 here and UNIOP
	// otherwise, and the difference is visible:
	//
	//	length $x + 1      length(add($x, 1))    the + is inside
	//	length ($x) + 1    add(length($x), 1)    the + is outside
	if next, ok := p.peekSignificant(); ok && p.text(next) == "(" {
		p.advanceTo(next)
		if arg := p.parseCallArgs(); arg != nil {
			n.Children = append(n.Children, arg)
		}
		if close, ok := p.peekSignificant(); ok && p.text(close) == ")" {
			p.advanceTo(close)
		}
		n.End = p.prevEnd()
		n.Resolved = namedUnary[text] || listOperator[text] || niladicParse[text]
		return n
	}

	// An operator with NO argument: `shift;`, `die;`, `$x or die;`.
	//
	// Most named unaries default their argument -- `shift` takes @_ or @ARGV,
	// `die` re-raises $@ -- so the argument is optional in practice even
	// where the grammar allows one. Calling parseExpr here returns nil, and
	// the caller then saw an unconsumed `;` and declared the statement
	// Unknown.
	//
	// Measured: `shift;` and `die;` between them are the first failure in
	// dozens of T1 files, and `... or die;` appears in almost every file
	// that opens a filehandle.
	if next, ok := p.peekSignificant(); !ok || endsArgumentList(next, p.src) {
		n.End = p.prevEnd()
		n.Resolved = namedUnary[text] || listOperator[text] || niladicParse[text]
		return n
	}

	switch {
	case namedUnary[text]:
		// One argument, parsed at level 19 so arithmetic binds into it and
		// comparison does not.
		if arg := p.parseExpr(bpNamedUnary); arg != nil {
			n.Children = append(n.Children, arg)
		}
		n.Resolved = true

	case listOperator[text]:
		// A BLOCK slot comes before the list, and what makes it a slot
		// rather than a first argument is the ABSENCE of a comma:
		//
		//	map { $_ } @a          a block, then the list
		//	map +{ x => 1 }, @a    a hashref, and the comma says so
		//
		// Whether the brace is a block at all was settled by the lexer,
		// which ran perl's intuit_curly over it, so this reads OpensBlock
		// rather than deciding a second time. The hashref form leaves it
		// unset and falls through to the ordinary list below, which is
		// right: it IS an ordinary first argument.
		//
		// Without this the list was orphaned. `map { $_ => 1 } @a` became a
		// call holding an AnonHash, and `@a` a statement of its own -- every
		// byte in the tree, round-tripping, and not a parse of its source.
		if blk := p.parseListOpBlock(text); blk != nil {
			n.Children = append(n.Children, blk)
		}
		// A filehandle slot works the same way and for the same reason:
		//
		//	print STDERR "a";      bareword handle
		//	print $fh "a";         scalar handle
		//	print {$fh} "a";       block handle
		//
		// Without this the list parser reads `STDERR`, then finds a string
		// with no operator between them and the statement falls to Unknown.
		// Measured: 5,319 of T1's 13,558 Unknown nodes started at `print`,
		// 39% of the whole gap from this one omission.
		if fh := p.parseFilehandleSlot(text); fh != nil {
			// Marked here rather than in parseFilehandleSlot's three
			// branches: one place that cannot be forgotten when a fourth
			// handle form arrives, and the fact is about the SLOT rather
			// than about the node's own shape.
			//
			// Without it the only evidence is child shape -- two children
			// means a handle, one means a comma list -- which
			// `internal/infer/infer.go:1155-1165` records as having made
			// every typed-handle print a false Str mismatch.
			fh.Handle = true
			n.Children = append(n.Children, fh)
		}
		// The whole comma list, parsed below the comma at level 7.
		if arg := p.parseExpr(bpListOp); arg != nil {
			n.Children = append(n.Children, arg)
		}
		n.Resolved = true

	default:
		// A bareword this parser does not know. It might be a user sub taking
		// a list, a class name, or a hash key -- perl decides with the symbol
		// table and this parser cannot.
		//
		// §4.8.3: produce a Call with Resolved false rather than an Unknown.
		// A call has a known SHAPE and an unknown callee, and the harness
		// scores that as an Unresolved site -- `wider`, not WRONG.
		//
		// No arguments are consumed. Guessing that a bareword takes a list
		// would swallow the rest of the statement on every hash key and
		// class name in the corpus.
		n.Resolved = false
	}

	n.End = p.prevEnd()
	return n
}

// endsArgumentList reports whether a token closes the context an operator's
// arguments would live in, so there is nothing left for it to take.
//
// A terminator, a closing bracket, or an infix operator: `$x or die;` reaches
// `die` with a `;` next, and `f(shift)` reaches `shift` with a `)` next.
func endsArgumentList(tok lexer.Token, src []byte) bool {
	switch tok.Kind {
	case lexer.Semicolon, lexer.CloseBracket:
		return true
	case lexer.Operator:
		// An infix operator cannot start an argument, so the call takes
		// none: `shift || 1`. A PREFIX operator can -- `die -1` -- so only
		// the unambiguously-infix ones count.
		switch string(src[tok.Start:tok.End]) {
		case ",", "=>", "||", "&&", "//", "=", "?", ":":
			return true
		}
	case lexer.Word:
		// The word-spelled logical operators, which the lexer emits as Word
		// because they are identifiers by shape.
		switch string(src[tok.Start:tok.End]) {
		case "or", "and", "xor", "if", "unless", "while", "until", "for", "foreach":
			return true
		}
	}
	return false
}

// takesFilehandle is the set of list operators whose first slot may be a
// filehandle with no comma after it.
//
// Not every list operator has one -- `push @a, 1` has no handle slot and
// treating `@a` as one would be wrong -- so the set is explicit.
var takesFilehandle = map[string]bool{
	"print": true, "printf": true, "say": true,
}

// takesBlock is the set of list operators whose first slot may be a BLOCK
// with no comma after it.
//
// The same three the lexer's blockTaking table names, and for the same
// reason: these are the operators where `{` is genuinely ambiguous. The two
// tables are separate because they answer different questions -- the lexer's
// decides how to LEX the brace, this one decides whether to read a slot --
// and a single shared table would couple the packages for three strings.
var takesBlock = map[string]bool{
	"map": true, "grep": true, "sort": true,
}

// parseListOpBlock reads the leading BLOCK of `map`, `grep` or `sort`, or
// returns nil when there is none.
//
// It asks the token, not the source: the lexer already ran intuit_curly and
// recorded the answer as OpensBlock. A parser that re-derived it could
// disagree with the token stream that produced it -- the same argument
// TestBraceDecisionUsesExpectState makes for every other brace.
func (p *parser) parseListOpBlock(op string) *Node {
	if !takesBlock[op] {
		return nil
	}
	tok, ok := p.peekSignificant()
	if !ok || !tok.OpensBlock || p.text(tok) != "{" {
		return nil
	}
	return p.parseBlock(tok)
}

// parseFilehandleSlot reads `STDERR`, `$fh` or `{$fh}` before a list, or
// returns nil when there is none.
//
// The distinguishing feature is the ABSENCE of a comma: `print $fh "a"` has a
// handle, `print $x, "a"` does not. So the slot is taken only when the token
// after the candidate begins a new term rather than continuing the list.
func (p *parser) parseFilehandleSlot(op string) *Node {
	if !takesFilehandle[op] {
		return nil
	}
	tok, ok := p.peekSignificant()
	if !ok {
		return nil
	}

	switch {
	case tok.OpensBlock && p.text(tok) == "{":
		// `print {$fh} "a"` -- the block form, which exists precisely to
		// disambiguate an expression in the slot.
		return p.parseBlock(tok)

	case tok.Kind == lexer.Word && isBarewordHandle(p.text(tok)):
		// A bareword handle. Only ALL-CAPS names qualify, which is perl's
		// own convention and what keeps `print foo 1` from stealing a
		// function call into the slot.
		p.advanceTo(tok)
		return &Node{
			Kind: Term, Text: p.text(tok),
			Start: tok.Start, End: tok.End,
		}

	case tok.Kind == lexer.Variable && p.src[tok.Start] == '$':
		// `print $fh "a"`. Only when what FOLLOWS starts a new term with no
		// comma -- otherwise `print $x, "a"` would lose its first argument.
		next, ok := p.peekAfter(tok)
		if !ok || !startsTerm(next, p.src) {
			return nil
		}
		p.advanceTo(tok)
		return &Node{
			Kind: Term, Text: p.text(tok),
			Start: tok.Start, End: tok.End,
		}
	}
	return nil
}

// isBarewordHandle reports whether a bareword looks like a filehandle.
//
// perl's rule is the symbol table's, which a static parser does not have, so
// this uses the convention perl's own documentation recommends and every
// corpus file follows: an all-caps name. STDERR, STDOUT, FH, OUT.
func isBarewordHandle(word string) bool {
	if word == "" {
		return false
	}
	for i := 0; i < len(word); i++ {
		c := word[i]
		if (c >= 'A' && c <= 'Z') || c == '_' || (c >= '0' && c <= '9') {
			continue
		}
		return false
	}
	return true
}

// startsTerm reports whether a token begins a new term rather than continuing
// an expression -- the test that separates `print $fh "a"` from `print $x, 1`.
func startsTerm(tok lexer.Token, src []byte) bool {
	switch tok.Kind {
	case lexer.Variable, lexer.Number, lexer.Quote, lexer.HeredocOpen:
		return true
	case lexer.Word:
		return true
	}
	return false
}

// parseCallArgs reads the inside of a parenthesised argument list.
func (p *parser) parseCallArgs() *Node {
	if tok, ok := p.peekSignificant(); ok && p.text(tok) == ")" {
		return nil
	}
	return p.parseExpr(0)
}
