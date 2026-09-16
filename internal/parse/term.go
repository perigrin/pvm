// ABOUTME: The nud side of the Pratt parser: §4.4's term forms, one per shape.
// ABOUTME: The backslash of §4.4.5 is here — it is the srefgen the fidelity harness measures.

package parse

import "tamarou.com/pvm/internal/lexer"

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
	case lexer.FuncSigil:
		// `&` in term position introduces a function name, and the lexer
		// emits it as its own token so that `&f` and `&&` stay
		// distinguishable. The name is the next token, so the two are joined
		// here into one term: `\&f` is a code reference, not a reference to
		// an ampersand.
		p.advanceTo(tok)
		if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
			p.advanceTo(name)
			return &Node{
				Kind: Term, Text: text + p.text(name),
				Start: tok.Start, End: name.End,
			}
		}
		return &Node{Kind: Term, Text: text, Start: tok.Start, End: tok.End}

	case lexer.Variable, lexer.Number, lexer.Quote, lexer.Word,
		lexer.Readline, lexer.HeredocOpen:
		p.advanceTo(tok)
		return &Node{
			Kind: Term, Text: text,
			Start: tok.Start, End: tok.End,
		}
	}

	// Not a term. Consumed so the loop advances; the bytes stay in the tree.
	p.advanceTo(tok)
	return &Node{Kind: Unknown, Start: tok.Start, End: tok.End}
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
			p.advanceTo(next)
		case ")":
			p.advanceTo(next)
			return p.finishList(items, open)
		default:
			// Neither a separator nor the closer: stop rather than spin.
			return p.finishList(items, open)
		}
	}
	return p.finishList(items, open)
}

func (p *parser) finishList(items []*Node, open lexer.Token) *Node {
	if len(items) == 1 {
		// `($x)` is `$x`, but the span covers the parens so round-trip holds.
		n := items[0]
		return &Node{
			Kind: n.Kind, Text: n.Text,
			Start: open.Start, End: p.prevEnd(),
			Children: n.Children,
		}
	}
	return &Node{
		Kind: List, Start: open.Start, End: p.prevEnd(),
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
			p.advanceTo(next)
		case closer:
			p.advanceTo(next)
			return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
		default:
			return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
		}
	}
	return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
}
