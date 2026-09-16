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

	switch {
	case namedUnary[text]:
		// One argument, parsed at level 19 so arithmetic binds into it and
		// comparison does not.
		if arg := p.parseExpr(bpNamedUnary); arg != nil {
			n.Children = append(n.Children, arg)
		}
		n.Resolved = true

	case listOperator[text]:
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

// parseCallArgs reads the inside of a parenthesised argument list.
func (p *parser) parseCallArgs() *Node {
	if tok, ok := p.peekSignificant(); ok && p.text(tok) == ")" {
		return nil
	}
	return p.parseExpr(0)
}
