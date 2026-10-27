// ABOUTME: Error recovery for a live buffer, §6.8.2-§6.8.3: Error and Missing nodes, HasError, missing tokens, panic mode.
// ABOUTME: CatchAll is R-a′'s quiet split; panicSkip and tierOneCloses are §5.13.4's sync points and 100-token cap.

package parse

import (
	"bytes"
	"maps"

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
// `;`); before a `}` closing the enclosing block, a column-0 anchor (tier
// 1), a line-start statement keyword (tier 2), or a 101st token. A bracket
// group that closes is one token, skipped whole, so nothing inside it is a
// sync point and the cap never cuts it: a statement perl compiles that a
// catch-all declined keeps the one Error it had. An opener that never
// closes opens nothing, and a closer with no block to close is stray; both
// are skipped as tokens. A sync point counts only once the statement holds
// a token, so every recovery advances.
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
	if p.closes == nil {
		p.closes = bracketGroups(p.toks, p.src)
	}
	end := p.pos
	for i, skipped := p.pos, 0; i < len(p.toks); i++ {
		tok := p.toks[i]
		if isTrivia(tok.Kind) {
			continue
		}
		if skipped == maxSkip || tok.Start > start && (tierOneAnchor(tok, p.src) || p.tierTwoAnchor(tok)) ||
			tok.Kind == lexer.CloseBracket && p.src[tok.Start] == '}' && p.inBlock {
			return end
		}
		skipped++
		if j := p.closes[i]; j > i {
			i = j
			if p.src[p.toks[j].Start] == '}' {
				return j + 1
			}
		}
		end = i + 1
		if tok.Kind == lexer.Semicolon {
			return end
		}
	}
	return end
}

// bracketGroups maps each opener in toks to the index of the closer that
// closes it, 0 for one that never closes. A closer closes the innermost
// opener of its own kind, and the openers inside that are left open; a
// closer with none open is stray. Brackets are bracketOf's.
func bracketGroups(toks []lexer.Token, src []byte) []int {
	closes := make([]int, len(toks))
	pair := map[byte]byte{')': '(', ']': '[', '}': '{'}
	var open []int
	count := map[byte]int{}
	var last lexer.Token
	for i, tok := range toks {
		if isTrivia(tok.Kind) {
			continue
		}
		b := bracketOf(tok, last, src)
		last = tok
		switch b {
		case 0:
		case '(', '[', '{':
			open = append(open, i)
			count[b]++
		default:
			if count[pair[b]] == 0 {
				continue
			}
			for {
				top := open[len(open)-1]
				open = open[:len(open)-1]
				count[src[toks[top].Start]]--
				if src[toks[top].Start] == pair[b] {
					closes[top] = i
					break
				}
			}
		}
	}
	return closes
}

// tierOneWords begin §5.13.4's tier-1 anchors at column 0, where Perl
// convention puts every top-level declaration.
var tierOneWords = map[string]bool{
	"sub": true, "package": true, "class": true, "format": true,
	"BEGIN": true, "END": true, "INIT": true, "CHECK": true, "UNITCHECK": true,
}

// tierTwoWords are §5.13.4's tier-2 anchors, perl-lsp's is_sync_point
// keywords, taken only first on a line: mid-line they are modifiers.
var tierTwoWords = map[string]bool{
	"my": true, "our": true, "local": true, "state": true,
	"sub": true, "package": true, "use": true,
	"if": true, "unless": true, "elsif": true, "else": true,
	"while": true, "until": true, "for": true, "foreach": true,
}

// tierOneAnchor reports whether tok is a tier-1 word at column 0. A word
// inside a heredoc body, POD or a string is that token's text, never one.
func tierOneAnchor(tok lexer.Token, src []byte) bool {
	return tok.Kind == lexer.Word && (tok.Start == 0 || src[tok.Start-1] == '\n') && tierOneWords[string(src[tok.Start:tok.End])]
}

// tierTwoAnchor reports whether tok is a tier-2 word first on its line.
func (p *parser) tierTwoAnchor(tok lexer.Token) bool {
	if tok.Kind != lexer.Word || !tierTwoWords[p.text(tok)] {
		return false
	}
	line := bytes.LastIndexByte(p.src[:tok.Start], '\n') + 1
	return len(bytes.TrimLeft(p.src[line:tok.Start], " \t")) == 0
}

// tierOneCloses is §5.13.4's tier 1 for a lost brace: each `{` no `}` closes
// by end of input closes at the first column-0 anchor directly inside it,
// so a sub missing its `}` leaves the column-0 subs after it at top level.
// It returns the token indexes of those anchors and of the braces they
// close, which refuseUnclosedOpener then leaves out of its count.
//
// Braces are matched only with braces, so a lost `)` closes no sub; a
// balanced file takes one pass and closes nothing.
func tierOneCloses(toks []lexer.Token, src []byte) (anchors, closed map[int]bool) {
	scan := func(at func(i int, open []int) []int) []int {
		var open []int
		var last lexer.Token
		for i, tok := range toks {
			if isTrivia(tok.Kind) {
				continue
			}
			if at != nil {
				open = at(i, open)
			}
			switch bracketOf(tok, last, src) {
			case '{':
				open = append(open, i)
			case '}':
				if len(open) > 0 {
					open = open[:len(open)-1]
				}
			}
			last = tok
		}
		return open
	}
	unclosed := scan(nil)
	if len(unclosed) == 0 {
		return nil, nil
	}
	anchors, closed = map[int]bool{}, map[int]bool{}
	for _, i := range unclosed {
		closed[i] = false
	}
	scan(func(i int, open []int) []int {
		if !tierOneAnchor(toks[i], src) {
			return open
		}
		// Popping one changes no other brace's match: nothing closed it.
		for n := len(open); n > 0; n-- {
			if _, ok := closed[open[n-1]]; !ok {
				return open[:n]
			}
			closed[open[n-1]], anchors[i] = true, true
		}
		return nil
	})
	maps.DeleteFunc(closed, func(_ int, v bool) bool { return !v })
	return anchors, closed
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
