// ABOUTME: The parse tree: nodes over byte spans, and the Unknown that declines rather than guesses.
// ABOUTME: Package parse, not ast — internal/parser is the tree-sitter wrapper and will be for a while.

package parse

import "tamarou.com/pvm/internal/lexer"

// Kind names what a node is.
type Kind int

const (
	// SourceFile is the root. It always exists, including for empty input:
	// an empty program is a valid program, and a nil root would make every
	// consumer check for one.
	SourceFile Kind = iota

	// Statement is one statement, whatever its form.
	Statement

	// Unknown is a construct this parser does not understand, spanning
	// exactly the bytes it could not parse.
	//
	// Unknown is the milestone's whole discipline in one node kind. A parser
	// that guesses at an unfamiliar construct produces a tree that is not a
	// parse of its source, and the fidelity harness scores that WRONG -- the
	// only bucket that fails a build. Emitting Unknown instead is a refusal,
	// and a refusal is never a wrong answer.
	//
	// But a refusal only counts as one if it is REPORTED. See Sites: an
	// Unknown that says nothing to the harness is scored as claiming there
	// is nothing there, which is a silent wrong answer rather than a
	// decline. The node and the hedge are one mechanism in two places.
	Unknown

	// Trivia is whitespace, a comment, or POD: bytes that belong to no
	// statement but must still be in the tree, because every byte is.
	Trivia

	// Term is a leaf value: a literal, a variable, a bareword.
	Term

	// Binary is an infix operator and its two operands.
	Binary

	// Unary is a prefix operator and its one operand. The `\` of §4.4.5 is
	// one of these, and it is the srefgen the fidelity harness measures.
	Unary

	// Postfix is `++` or `--` after its operand.
	Postfix

	// Ternary is `?:`, with condition, then-branch and else-branch.
	Ternary

	// CmpChain is a chained comparison, 5.32+. One n-ary node rather than
	// nested binaries, because each operand is evaluated once:
	//
	//	$ perl -MO=Deparse,-p -e 'my $q = $a < $b < $c'
	//	(my($q) = ($a < $b < $c));
	CmpChain

	// List is a parenthesised or comma-separated list.
	List

	// AnonArray is `[...]` and AnonHash is `{...}` in term position. Anonhash
	// is one of the five oracle markers.
	AnonArray
	AnonHash

	// Index is a subscript or dereference: `$r->[0]`, `$h{k}`, `@a[0,1]`.
	Index
)

func (k Kind) String() string {
	switch k {
	case SourceFile:
		return "source_file"
	case Statement:
		return "statement"
	case Unknown:
		return "unknown"
	case Trivia:
		return "trivia"
	case Term:
		return "term"
	case Binary:
		return "binary"
	case Unary:
		return "unary"
	case Postfix:
		return "postfix"
	case Ternary:
		return "ternary"
	case CmpChain:
		return "cmp_chain"
	case List:
		return "list"
	case AnonArray:
		return "anon_array"
	case AnonHash:
		return "anon_hash"
	case Index:
		return "index"
	}
	return "?"
}

// Node is one node of the parse tree.
//
// Byte offsets rather than line/column, for the reason the lexer's Token
// gives: line and column are derived at the LSP layer, and carrying them
// doubles the state that must stay consistent under edit.
//
// A node's span always covers its children's spans exactly, with no gaps.
// That is what makes the round-trip property hold: concatenating every
// leaf's text reproduces the input, Unknown included.
type Node struct {
	Kind       Kind
	Start, End int
	Children   []*Node

	// Text is the operator or literal that names this node -- "+", "?:",
	// "$x". Empty for nodes whose meaning is entirely their kind and
	// children, such as a List.
	//
	// It duplicates bytes the span already covers, which is deliberate: an
	// interior node's span covers its operands too, so `$a + $b` and
	// `$a - $b` are indistinguishable by span alone. A consumer that had to
	// re-lex the span to learn which operator it got would be re-deriving
	// what the parser already knew.
	Text string
}

// SourceText reconstructs the bytes this node covers, walking the tree.
//
// It is the round-trip invariant made checkable by a consumer rather than
// only by a test. An LSP renders from the tree, so "the tree still holds
// every byte" has to be something the tree itself can answer.
//
// Children do not tile their parent's span: `$a + $b` has two children and
// the `+` and the spaces belong to neither. The gaps between children are
// the parent's own bytes and are emitted in place, which is why an operator
// is NOT made a child node -- an operator is not an operand, and a tree that
// said so would force every consumer to filter it back out.
func (n *Node) SourceText(src []byte) string {
	var b []byte
	var walk func(*Node)
	walk = func(n *Node) {
		at := n.Start
		for _, c := range n.Children {
			if c.Start > at {
				b = append(b, src[at:c.Start]...)
			}
			walk(c)
			at = c.End
		}
		if n.End > at {
			b = append(b, src[at:n.End]...)
		}
	}
	walk(n)
	return string(b)
}

// Parse builds a tree covering every byte of src exactly once.
//
// Like Tokenize, it has no error return and never panics. An LSP sees
// half-typed buffers constantly, and a parser that gives up on the first
// unfamiliar construct is useless to it.
//
// At this stage every statement is Unknown: the skeleton establishes the
// invariants before there is any grammar to test them against. Later issues
// replace Unknown with real nodes one construct at a time, and the round-trip
// test is what keeps each replacement honest.
func Parse(src []byte) *Node {
	root := &Node{Kind: SourceFile, Start: 0, End: len(src)}
	toks := lexer.Tokenize(src)

	p := &parser{src: src, toks: toks}
	for p.pos < len(p.toks) {
		before := p.pos
		if n := p.statement(); n != nil {
			root.Children = append(root.Children, n)
		}
		// Forward progress, asserted rather than assumed. The lexer learned
		// this the hard way: scanFormatBody was wrapped in a step() guard and
		// the fuzzer still found a `format =\n` input that spun. A parser
		// loop that can return without consuming is the same bug one layer
		// up, so the guard is here rather than in a comment.
		if p.pos == before {
			p.pos++
		}
	}
	return root
}

// parser is the cursor over the token stream.
type parser struct {
	src  []byte
	toks []lexer.Token
	pos  int
}

// peekSignificant returns the next non-trivia token without consuming it.
//
// Trivia is skipped for LOOKUP but never for CONSUMPTION: advanceTo moves the
// cursor past the trivia as well as the token, so every byte still lands in
// exactly one node and round-trip holds. A parser that dropped trivia here
// would have to put it back somewhere.
func (p *parser) peekSignificant() (lexer.Token, bool) {
	for i := p.pos; i < len(p.toks); i++ {
		if !isTrivia(p.toks[i].Kind) {
			return p.toks[i], true
		}
	}
	return lexer.Token{}, false
}

// advanceTo consumes through the given token, trivia included.
func (p *parser) advanceTo(tok lexer.Token) {
	for p.pos < len(p.toks) && p.toks[p.pos].Start <= tok.Start {
		p.pos++
	}
}

// text is the token's source bytes.
func (p *parser) text(tok lexer.Token) string {
	return string(p.src[tok.Start:tok.End])
}

// prevEnd is where the last consumed token ended, which is where a node that
// closed on the cursor should end.
func (p *parser) prevEnd() int {
	if p.pos == 0 {
		return 0
	}
	return p.toks[p.pos-1].End
}

// skipToStatementEnd consumes to the next `;` or unmatched `}`, which is
// where recovery resumes.
//
// It stops ON the boundary token, not after trailing trivia: the newline
// after a statement belongs to no statement, and a recovery that swallowed it
// would make the Unknown's span disagree with every other node's.
func (p *parser) skipToStatementEnd() {
	depth := 0
	sawOpen := false
	for p.pos < len(p.toks) {
		tok := p.toks[p.pos]
		p.pos++
		switch tok.Kind {
		case lexer.Operator:
			if tok.End-tok.Start == 1 {
				switch p.src[tok.Start] {
				case '{', '(', '[':
					depth++
					sawOpen = true
				}
			}
		case lexer.CloseBracket:
			// A closer with nothing open belongs to an enclosing construct;
			// consuming past it would swallow the rest of the file.
			if depth == 0 {
				return
			}
			depth--
			// Back to depth zero on a `}`: a block-bearing statement such as
			// `if (..) { .. }` ends at its closing brace with no semicolon.
			// The two openers of that form are why "depth is already zero" is
			// not the test -- by the time the `}` is reached, the `(` has
			// already been balanced.
			if depth == 0 && sawOpen && p.src[tok.Start] == '}' {
				return
			}
		case lexer.Semicolon:
			if depth == 0 {
				return
			}
		}
	}
}

// statement consumes one statement, or one run of trivia.
//
// Trivia comes out as its own node rather than being attached to a
// statement, because "which statement does a blank line between two subs
// belong to" has no good answer and every answer complicates round-trip.
func (p *parser) statement() *Node {
	if isTrivia(p.toks[p.pos].Kind) {
		start := p.toks[p.pos].Start
		end := start
		for p.pos < len(p.toks) && isTrivia(p.toks[p.pos].Kind) {
			end = p.toks[p.pos].End
			p.pos++
		}
		return &Node{Kind: Trivia, Start: start, End: end}
	}

	start := p.toks[p.pos].Start

	// A statement FORM -- a declaration, control flow, a phaser -- is not an
	// expression, and the expression parser must not be handed one.
	//
	// It would not fail if it were. `if ($x) { 1 }` reads as a bareword,
	// then a call, then a subscript: index(index(if, ($x)), {1}). Every node
	// is well-formed and the whole thing is a fiction. That is precisely the
	// guess the plan forbids -- a tree that is not a parse of its source,
	// which the harness scores WRONG.
	//
	// So the forms are named and declined until the issues that own them
	// land. Naming them is the cost of not guessing at them.
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Word {
		if statementKeywords[p.text(tok)] {
			p.skipToStatementEnd()
			return &Node{Kind: Unknown, Start: start, End: p.prevEnd()}
		}
	}

	expr := p.parseExpr(0)
	if expr == nil {
		p.skipToStatementEnd()
		return &Node{Kind: Unknown, Start: start, End: p.prevEnd()}
	}

	// Through a terminating `;` if there is one, so the statement owns its
	// punctuation and the next statement starts clean.
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}

	// Anything left before the statement boundary was not consumed by the
	// expression parser -- a statement form it does not know. The whole
	// statement becomes Unknown rather than a half-parsed expression next to
	// a mystery: a partial tree claims to be a parse of bytes it did not
	// read, and that is what the harness scores WRONG.
	if tok, ok := p.peekSignificant(); ok && !endsStatement(tok, p.src) {
		p.skipToStatementEnd()
		return &Node{Kind: Unknown, Start: start, End: p.prevEnd()}
	}

	if expr.Kind == Unknown {
		return &Node{Kind: Unknown, Start: start, End: p.prevEnd()}
	}
	return &Node{
		Kind: Statement, Start: start, End: p.prevEnd(),
		Children: []*Node{expr},
	}
}

// statementKeywords open a statement form this milestone has not implemented
// yet. Each is owned by a later issue in M1:
//
//	my our local state sub package  declarations
//	if elsif else unless            control flow
//	while until for foreach do      control flow
//	use no require BEGIN END ...    the rest of T2
//	class field method              the rest of T2
//
// They are listed rather than detected because the expression parser cannot
// fail on them -- it reads `if ($x) { 1 }` as a call-then-subscript and
// returns a confident wrong answer. A list that must be shortened as each
// issue lands is the honest form of "not yet": the entry disappears when the
// parser can really read the form.
var statementKeywords = map[string]bool{
	"my": true, "our": true, "local": true, "state": true,
	"sub": true, "package": true,

	"if": true, "elsif": true, "else": true, "unless": true,
	"while": true, "until": true, "for": true, "foreach": true,
	"do": true, "continue": true,

	"use": true, "no": true, "require": true,
	"BEGIN": true, "END": true, "CHECK": true, "INIT": true, "UNITCHECK": true,

	"class": true, "field": true, "method": true,
	"try": true, "catch": true, "finally": true, "defer": true,

	// Loop controls and `return` take an optional term and are statement
	// forms in perly.y (levels 2 and 7), not expression operators.
	"return": true, "last": true, "next": true, "redo": true, "goto": true,

	"format": true,
}

// endsStatement reports whether this token closes the statement rather than
// continuing it. A `}` belongs to an enclosing construct, so it stops the
// statement without being consumed by it.
func endsStatement(tok lexer.Token, src []byte) bool {
	return tok.Kind == lexer.Semicolon ||
		(tok.Kind == lexer.CloseBracket && src[tok.Start] == '}')
}

func isTrivia(k lexer.Kind) bool {
	switch k {
	case lexer.Whitespace, lexer.Comment, lexer.Pod:
		return true
	}
	return false
}
