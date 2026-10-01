// ABOUTME: Declared syntax: keywords and sub prefixes a module's declaration file adds once imported.
// ABOUTME: Modelled on XS::Parse::Keyword and XS::Parse::Sublike, which state a plugin's grammar as data.

package parse

import (
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// declaredSyntax is one grammar extension a declaration file states. A
// keyword plugin built on XS::Parse::Keyword or XS::Parse::Sublike registers
// its grammar as data -- a sequence of pieces, or flags on a sub-like
// keyword -- and a declaration file restates that data:
//
//	sub async  :sublike(PREFIX);     a prefix on `sub`: `async sub f {...}`
//	sub await  :keyword(TERMEXPR);   an expression: `await $f`
//	sub CANCEL :statement(ANONSUB);  a statement ending at its last piece
//
// The pieces are XS::Parse::Keyword's, without the XPK_ prefix: TERMEXPR is
// an expression running down to assignment and stopping at a comma
// (parse_termexpr, toke.c:14265); BLOCK and ANONSUB are a braced block.
type declaredSyntax struct {
	// prefix marks a sub-like keyword that stands before `sub`.
	prefix bool

	// statement marks a keyword that builds a statement rather than an
	// expression: it ends at its last piece, as `defer BLOCK` does.
	statement bool

	pieces []string
}

// syntaxPieces are the pieces this parser reads.
var syntaxPieces = map[string]bool{"TERMEXPR": true, "BLOCK": true, "ANONSUB": true}

// readSyntaxAttribute reads one declaration-file attribute: `:sublike(PREFIX)`,
// `:keyword(PIECES)` or `:statement(PIECES)`. ok is false for any other
// attribute, and for a sub-like keyword or piece this parser does not read --
// declining to declare a keyword leaves its word a word, where guessing at its
// grammar would build a wrong tree.
func readSyntaxAttribute(text string) (s declaredSyntax, ok bool) {
	name, args, found := strings.Cut(strings.TrimPrefix(text, ":"), "(")
	if !found || !strings.HasSuffix(args, ")") {
		return declaredSyntax{}, false
	}
	words := strings.Fields(strings.TrimSuffix(args, ")"))
	switch name {
	case "sublike":
		return declaredSyntax{prefix: true}, len(words) == 1 && words[0] == "PREFIX"
	case "keyword", "statement":
		for _, w := range words {
			if !syntaxPieces[w] {
				return declaredSyntax{}, false
			}
		}
		return declaredSyntax{statement: name == "statement", pieces: words}, len(words) > 0
	}
	return declaredSyntax{}, false
}

// parseDeclaredStatement: a declared statement keyword, `CANCEL { ... }`, or
// nil when word is not one here. Shaped as parseDefer shapes `defer`, a
// Conditional named by its keyword.
func (p *parser) parseDeclaredStatement(word lexer.Token) *Node {
	s, ok := p.syntax[p.text(word)]
	if !ok || !s.statement || !p.piecesStart(word, s.pieces) {
		return nil
	}
	p.advanceTo(word)
	n := &Node{Kind: Conditional, Text: p.text(word), Start: word.Start}
	p.parsePieces(n, s.pieces)
	n.End = p.prevEnd()
	return n
}

// parseDeclaredExpression: a declared expression keyword, `await $f`, or nil
// when word is not one here. A Call named by its keyword, as `defined $x` is.
func (p *parser) parseDeclaredExpression(word lexer.Token) *Node {
	s, ok := p.syntax[p.text(word)]
	if !ok || s.prefix || s.statement || !p.piecesStart(word, s.pieces) {
		return nil
	}
	p.advanceTo(word)
	n := &Node{Kind: Call, Text: p.text(word), Start: word.Start}
	p.parsePieces(n, s.pieces)
	n.End = p.prevEnd()
	return n
}

// prefixedSub reports whether word is a declared sub prefix with `sub`
// after it, returning that `sub`.
func (p *parser) prefixedSub(word lexer.Token) (lexer.Token, bool) {
	if s, ok := p.syntax[p.text(word)]; !ok || !s.prefix {
		return lexer.Token{}, false
	}
	sub, ok := p.peekAfter(word)
	return sub, ok && sub.Kind == lexer.Word && p.text(sub) == "sub"
}

// parsePrefixedSubDecl: `async sub NAME ...`, or nil when word is not a
// declared prefix before a named sub. The prefix wraps the sub as `my` wraps
// a lexical one, a Declaration whose one child is the sub.
func (p *parser) parsePrefixedSubDecl(word lexer.Token) *Node {
	sub, ok := p.prefixedSub(word)
	if !ok {
		return nil
	}
	if name, ok := p.peekAfter(sub); !ok || name.Kind != lexer.Word {
		return nil
	}
	p.advanceTo(word)
	inner := p.parseSubDecl(sub)
	return &Node{Kind: Declaration, Text: p.text(word), Start: word.Start, End: inner.End, Children: []*Node{inner}}
}

// parsePrefixedAnonSub: `async sub { ... }` in an expression, or nil.
func (p *parser) parsePrefixedAnonSub(word lexer.Token) *Node {
	sub, ok := p.prefixedSub(word)
	if !ok {
		return nil
	}
	if next, ok := p.peekAfter(sub); !ok || !(p.text(next) == "{" || p.text(next) == ":" ||
		next.Kind == lexer.Prototype || p.text(next) == "(") {
		return nil
	}
	p.advanceTo(word)
	inner := p.parseAnonSub(sub)
	return &Node{Kind: Declaration, Text: p.text(word), Start: word.Start, End: inner.End, Children: []*Node{inner}}
}

// piecesStart reports whether what follows word can begin the first piece: a
// block piece needs its `{`.
func (p *parser) piecesStart(word lexer.Token, pieces []string) bool {
	if pieces[0] != "BLOCK" && pieces[0] != "ANONSUB" {
		return true
	}
	next, ok := p.peekAfter(word)
	return ok && p.text(next) == "{"
}

// parsePieces reads each piece into n's children, stopping at the first that
// is not there.
func (p *parser) parsePieces(n *Node, pieces []string) {
	for _, piece := range pieces {
		var c *Node
		switch piece {
		case "TERMEXPR":
			// Down to assignment, stopping at the comma, as a list item is.
			c = p.parseExpr(infix[","].BP)
		case "BLOCK", "ANONSUB":
			c = p.parseBlockOrDecline()
		}
		if c == nil {
			return
		}
		n.Children = append(n.Children, c)
	}
}
