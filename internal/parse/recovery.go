// ABOUTME: Error recovery for a live buffer, §6.8.2: the Error and Missing nodes, HasError, and missing-token synthesis.
// ABOUTME: CatchAll is the quiet/reported split perigrin's R-a′ and R-g draw over the refusal codes.

package parse

import (
	"bytes"

	"tamarou.com/pvm/internal/lexer"
)

// CatchAll reports whether code is one of the catch-all refusals a syntax
// error falls into rather than a construct it names: not_a_term,
// trailing_tokens and not_an_expression (perigrin, 2026-10-26, R-a′). Their
// sites build Error nodes, which are structure only: they publish no syntax
// diagnostic and gate no type diagnostic, because those codes fire on Perl
// perl compiles -- trailing_tokens on 318 of 620 perl.git t/ files.
//
// missing_operand is not one (R-g): its Missing node is §6.8.2's missing
// expression after an operator, reported like every synthesized token.
func CatchAll(code RefusalCode) bool {
	switch code {
	case NotATerm, TrailingTokens, NotAnExpression:
		return true
	}
	return false
}

// recover counts a node recovery built, an Error or a Missing one, so the
// statement holding it is marked (markErrors) and a statement without one is
// not walked.
func (p *parser) recover(n *Node) *Node {
	p.recovered++
	return n
}

// markErrors sets HasError on n and everything under it that is, or holds,
// an Error or Missing node, and reports whether n does. parseWith runs it on
// each statement as the statement is finished, and only on one that built a
// recovery node, so Perl with nothing to recover is never walked.
func markErrors(n *Node) bool {
	has := n.Kind == Error || n.Kind == Missing
	for _, c := range n.Children {
		if markErrors(c) {
			has = true
		}
	}
	n.HasError = has
	return has
}

// missingOperand is the zero-width Missing node for an operator's operand
// that is not there, at the operator's end: input ran out, or what follows
// ends the statement or block, `my $x = ;` and `sub f { 2 * }`. The `;` or
// `}` is left to end what it ends (R-g).
func (p *parser) missingOperand(after lexer.Token) *Node {
	return p.recover(&Node{Kind: Missing, Refusal: MissingOperand, Start: after.End, End: after.End})
}

// operandAbsent reports whether no operand follows: end of input, or a `;`
// or `}`, which no term begins.
func (p *parser) operandAbsent() bool {
	tok, ok := p.peekSignificant()
	return !ok || endsStatement(tok, p.src)
}

// missingIdentifier gives a not_a_term Error standing on a `$` or a `sub`
// with nothing after it -- `print $` at end of input, `f(sub, 1)` -- the
// zero-width Missing identifier §6.8.2 synthesizes there, so the variable
// or declaration is still recognizable inside the Error.
func (p *parser) missingIdentifier(n *Node, tok lexer.Token) *Node {
	if text := p.text(tok); n.End == tok.End && (text == "$" || text == "sub") {
		n.Children = []*Node{p.recover(&Node{Kind: Missing, Start: tok.End, End: tok.End})}
	}
	return n
}

// missingSemicolon is the zero-width `;` §6.8.2 synthesizes after a statement
// whose next token, on a later line, begins a statement of its own: `my $x =
// 1` and `my $y = 2;` on the next line, which perl rejects as one. Nil where
// the statement is refused as before.
//
// shortcut: only `my` and `our` count as beginning a statement, and not after
// a word, which may be a sub taking `my $y` as its argument (`sub foo {} foo
// my $y` compiles); widen when panic-mode recovery (§6.8.3) lands.
func (p *parser) missingSemicolon() *Node {
	next, ok := p.peekSignificant()
	if !ok || p.pos == 0 || next.Kind != lexer.Word || p.text(next) != "my" && p.text(next) != "our" {
		return nil
	}
	last := p.toks[p.pos-1]
	if last.Kind == lexer.Word || !bytes.Contains(p.src[last.End:next.Start], []byte("\n")) {
		return nil
	}
	return p.recover(&Node{Kind: Missing, Text: ";", Start: last.End, End: last.End})
}

// closers are the zero-width Missing closers for the brackets still open at
// end of input, innermost first, each at at: one `}`, `)` or `]` per opener
// (§6.8.2). starts holds where each opener was read, in order.
func (p *parser) closers(starts []int, at int) []*Node {
	out := make([]*Node, 0, len(starts))
	for i := len(starts) - 1; i >= 0; i-- {
		text := map[byte]string{'(': ")", '[': "]", '{': "}"}[p.src[starts[i]]]
		out = append(out, &Node{Kind: Missing, Text: text, Start: at, End: at, HasError: true})
	}
	return out
}
