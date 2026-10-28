// ABOUTME: Error recovery for a live buffer, §6.8.2-§6.8.3: Error and Missing nodes, HasError, missing tokens, panic mode.
// ABOUTME: CatchAll is R-a′'s quiet split; panicSkip and tierOneMark are §5.13.4's sync points and 100-token cap.

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

// recover counts a node recovery built, an Error or a Missing one, so finish
// marks the tree (markErrors) and a parse that built none is not walked.
func (p *parser) recover(n *Node) *Node {
	p.recovered++
	return n
}

// markErrors sets HasError on n and everything under it that is, or holds,
// an Error or Missing node, clears it elsewhere, and reports whether n
// does. finish runs it once over the tree it finishes, and only when the
// parse built a recovery node or the tree holds a flagged child, so Perl
// with nothing to recover is never walked.
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
// my $y` compiles). Panic mode (panicSkip) now stops before every other
// line-start keyword, so what this leaves is one trailing_tokens Error, not
// a swallowed statement; widen to those keywords only once R-g′ is measured
// against them, since trailing_tokens fires on Perl perl compiles.
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
		out = append(out, &Node{Kind: Missing, Text: text, Start: at, End: at})
	}
	return out
}

// maxSkip caps one panic-mode skip at perl-lsp's 100 tokens (§5.13.4), so
// one error cannot eat a file: the outer loop resumes where the cap stops.
const maxSkip = 100

// panicSkip is panic-mode recovery (§6.8.3): from a catch-all refusal in
// the statement that began at start, it consumes to the nearest §5.13.4
// sync point, and the caller's one Error node covers everything consumed.
//
// It stops after a `;`, or after a `{ ... }` (`if (...) { ... }` takes no
// `;`), with no bracket it opened still open; before a `}` closing the
// enclosing block, a column-0 anchor at any depth (tier 1), a line-start
// keyword where a statement can start (tier 2, tierTwoAnchor), or the 101st
// token. Inside a bracket the statement opened, before the skip or in it,
// only tier 1 and the enclosing block's `}` stop it, and the cap waits for
// the bracket to close: a statement perl compiles that a catch-all declined
// keeps the one Error it had. A closer closes the innermost opener of its
// kind, the ones inside it left open. A `)` or `]` closing an opener read
// before the skip -- the condition of an `if` whose parse stopped inside it
// -- ends the skip after it; one with none open, before the statement or in
// it, is stray and skipped. A sync point counts only once the statement
// holds a token, so every recovery advances.
//
// Every stop is decided by the tokens already read, so the parse of a
// prefix never depends on the text after it (Repair resumes on that).
//
// It ends on the last significant token it took, the trivia after belonging
// to no statement, then takes the heredoc bodies that follow: their openers
// are in what it took, and left to the statement loop each would be a
// second refusal.
func (p *parser) panicSkip(start int) {
	p.pos = p.skipEnd(start)
	p.takeHeredocBodies()
}

// skipEnd is the token index panicSkip stops at.
func (p *parser) skipEnd(start int) int {
	pair := map[byte]byte{')': '(', ']': '[', '}': '{'}
	var open []byte    // the brackets the statement opened, innermost last
	var at []int       // and the token index each opened at
	var count [256]int // of each kind
	var last lexer.Token
	// From the statement's first token: the parse may have stopped inside a
	// bracket it opened, `++$h{refaddr $x} > 1`, whose closer is the
	// statement's and not the block's.
	from := p.pos
	for from > 0 && p.toks[from-1].Start >= start {
		from--
	}
	for i := from - 1; i >= 0; i-- {
		if !isTrivia(p.toks[i].Kind) {
			last = p.toks[i]
			break
		}
	}
	end := p.pos
	for i, skipped := from, 0; i < len(p.toks); i++ {
		tok := p.toks[i]
		if isTrivia(tok.Kind) {
			continue
		}
		b := bracketOf(tok, last, p.src)
		skipping := i >= p.pos
		if skipping && (tok.Start > start && (tierOneAnchor(tok, p.src) || len(open) == 0 && p.tierTwoAnchor(tok, last)) ||
			len(open) == 0 && skipped >= maxSkip || b == '}' && count['{'] == 0 && p.inBlock) {
			return end
		}
		last = tok
		if skipping {
			skipped++
			end = i + 1
		}
		switch b {
		case '(', '[', '{':
			open, at = append(open, b), append(at, i)
			count[b]++
		case ')', ']', '}':
			closed, opened := count[pair[b]] > 0, 0
			for closed {
				top := open[len(open)-1]
				open, opened, at = open[:len(open)-1], at[len(at)-1], at[:len(at)-1]
				count[top]--
				if top == pair[b] {
					break
				}
			}
			switch {
			case !skipping:
			case b == '}':
				// A block form's `{ ... }` the skip read, not a subscript's `}`.
				if closed && len(open) == 0 && opened >= p.pos {
					return end
				}
			case closed && opened < p.pos, !closed && p.enclosedBy(pair[b], from):
				// The `)` or `]` of a condition, argument list or subscript
				// the parse stopped inside, opened before the skip in the
				// statement or before the statement: it contains the error,
				// and what follows it -- a control statement's block -- is
				// parsed as code.
				return end
			}
		}
		if skipping && len(open) == 0 && tok.Kind == lexer.Semicolon {
			return end
		}
	}
	return end
}

// openToEOF reports whether the statement read from token from left only
// trivia after it with a bracket it opened still open, counted as
// refuseUnclosedOpener counts them.
func (p *parser) openToEOF(from int) bool {
	for _, tok := range p.toks[p.pos:] {
		if !isTrivia(tok.Kind) {
			return false
		}
	}
	depth := 0
	var last lexer.Token
	for _, tok := range p.toks[from:p.pos] {
		if isTrivia(tok.Kind) {
			continue
		}
		switch bracketOf(tok, last, p.src) {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth = max(depth-1, 0)
		}
		last = tok
	}
	return depth > 0
}

// maxEnclosingScan bounds enclosedBy's walk back, so a run of stray closers
// costs each skip a bounded walk rather than one to the start of the file.
const maxEnclosingScan = 1000

// enclosedBy reports whether a statement beginning at token from stands
// inside an opener of kind open read before it: walking back to the nearest
// `;`, `{` or `}` outside brackets, an opener nothing after it closed.
// Brackets are counted as refuseUnclosedOpener counts them, and a walk past
// maxEnclosingScan tokens finds none.
func (p *parser) enclosedBy(open byte, from int) bool {
	depth := 0
	for i, n := from-1, 0; i >= 0 && n < maxEnclosingScan; i-- {
		tok := p.toks[i]
		if isTrivia(tok.Kind) {
			continue
		}
		n++
		var last lexer.Token
		for j := i - 1; j >= 0; j-- {
			if !isTrivia(p.toks[j].Kind) {
				last = p.toks[j]
				break
			}
		}
		switch b := bracketOf(tok, last, p.src); {
		case b == ')' || b == ']' || b == '}':
			if b == '}' && depth == 0 {
				return false
			}
			depth++
		case b == '(' || b == '[' || b == '{':
			if depth == 0 {
				return b == open
			}
			depth--
		case tok.Kind == lexer.Semicolon && depth == 0:
			return false
		}
	}
	return false
}

// tierOneWords begin §5.13.4's tier-1 anchors at column 0, where Perl
// convention puts every top-level declaration.
var tierOneWords = map[string]bool{
	"sub": true, "package": true, "class": true, "format": true,
	"BEGIN": true, "END": true, "INIT": true, "CHECK": true, "UNITCHECK": true,
}

// tierTwoWords are §5.13.4's tier-2 anchors, perl-lsp's is_sync_point
// keywords, taken only first on a line. The true ones begin a declaration
// and nothing else; the false ones are also statement modifiers.
var tierTwoWords = map[string]bool{
	"my": true, "our": true, "local": true, "state": true,
	"sub": true, "package": true, "use": true,
	"if": false, "unless": false, "elsif": false, "else": false,
	"while": false, "until": false, "for": false, "foreach": false,
}

// tierOneAnchor reports whether tok is a tier-1 word at column 0. A word
// inside a heredoc body, POD or a string is that token's text, never one.
func tierOneAnchor(tok lexer.Token, src []byte) bool {
	return tok.Kind == lexer.Word && (tok.Start == 0 || src[tok.Start-1] == '\n') && tierOneWords[string(src[tok.Start:tok.End])]
}

// tierTwoAnchor reports whether tok is a tier-2 word first on its line where
// a statement can start after last, the token before it. After a `;`, `{`
// or `}` one always can. After a term -- a variable, number, string or
// closing bracket -- only a declaration can, the statement before having
// lost its `;`: a modifier there continues it (`croak "x"\n unless $ok;`),
// and after a word or an operator a declaration may be its operand
// (`foo\n my $x`, `$cb =\n sub { ... }`).
func (p *parser) tierTwoAnchor(tok, last lexer.Token) bool {
	decl, ok := tierTwoWords[p.text(tok)]
	if tok.Kind != lexer.Word || !ok {
		return false
	}
	line := bytes.LastIndexByte(p.src[:tok.Start], '\n') + 1
	if len(bytes.TrimLeft(p.src[line:tok.Start], " \t")) != 0 {
		return false
	}
	switch {
	case last.Kind == lexer.Semicolon, last.End-last.Start == 1 && (p.src[last.Start] == '{' || p.src[last.Start] == '}'):
		return true
	case !decl:
		return false
	}
	switch last.Kind {
	case lexer.Variable, lexer.Number, lexer.Quote, lexer.Readline, lexer.CloseBracket:
		return true
	}
	return false
}

// tierOneMark is where a block's statements met their first column-0
// anchor (§5.13.4 tier 1): the parser as it stood there, and how many of
// the block's children it had read. parseBlock closes a brace still open at
// end of input at that anchor -- a decision made where the block ends, so
// nothing before the brace depends on it.
type tierOneMark struct {
	set      bool
	children int
	p        parser
}

// bracketOf is tok's bracket as refuseUnclosedOpener counts them, or 0: a
// one-byte token, never a Word (a glob's punctuation name, `*(`), and after
// a `&` only a `{` (the rest name the sub, see ampTakesPunct). last is the
// significant token before tok.
func bracketOf(tok, last lexer.Token, src []byte) byte {
	if tok.End-tok.Start != 1 || tok.Kind == lexer.Word || last.Kind == lexer.FuncSigil && src[tok.Start] != '{' {
		return 0
	}
	if b := src[tok.Start]; bytes.IndexByte([]byte("([{)]}"), b) >= 0 {
		return b
	}
	return 0
}
