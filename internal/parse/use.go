// ABOUTME: use/no/require, the phasers, and the 5.38 class syntax — the rest of what T2 contains.
// ABOUTME: Syntax only: what `use feature` enables is M2's, and BEGIN's compile-time effect is M4's.

package parse

import "tamarou.com/pvm/internal/lexer"

// phasers are the compile-and-run-phase blocks.
var phasers = map[string]bool{
	"BEGIN": true, "END": true, "CHECK": true,
	"INIT": true, "UNITCHECK": true,
}

// parseTheRest handles the statement forms this issue owns, or returns nil.
func (p *parser) parseTheRest(word lexer.Token) *Node {
	text := p.text(word)
	switch {
	case text == "use" || text == "no" || text == "require":
		return p.parseUse(word)
	case phasers[text]:
		return p.parsePhaser(word)
	case text == "class":
		return p.parseClass(word)
	}
	return nil
}

// parseUse: `use MODULE LIST;`, `use VERSION;`, and the `no` and `require`
// spellings.
//
// The import list is an ordinary expression, which covers every shape the T2
// corpus actually contains -- surveyed rather than guessed:
//
//	12  use v5.36
//	12  use feature 'class'
//	12  no warnings 'experimental::class'
//	 2  no warnings qw(syntax deprecated)
//	 1  use test_use { () }
//
// The last one is why the list is parsed as an expression rather than as a
// comma-separated list of literals: `{ () }` is a term, and a narrower parser
// would decline it.
func (p *parser) parseUse(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Use, Text: p.text(word), Start: word.Start}

	// The module name is a BAREWORD, not an expression.
	//
	// Parsing it as one makes `use feature 'class';` leave the string
	// orphaned: `feature` becomes an unresolved Call, which by design
	// consumes no arguments, so `'class'` is never read and the statement
	// falls to Unknown. Measured on five t/class files before this split.
	//
	// So the name is taken directly and the import list is parsed after it.
	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		term := &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		}

		// A version bundle is THREE tokens, not one: `use v5.36;` lexes as
		// Word("v5") Operator(".") Number(36), because `v5` is a valid
		// identifier and the lexer has no reason to know better.
		//
		// Taking only the Word leaves `.36` behind, which is not an
		// expression, so the statement falls to Unknown -- measured on seven
		// t/class files. The whole run is absorbed into one term instead.
		if isVersionPrefix(term.Text) {
			for {
				dot, ok := p.peekSignificant()
				if !ok || p.text(dot) != "." {
					break
				}
				num, ok := p.peekAfter(dot)
				if !ok || num.Kind != lexer.Number {
					break
				}
				p.advanceTo(num)
				term.End = num.End
			}
		}
		// Resolve the module, if this parse has a loader and the name is a
		// module rather than a version bundle.
		//
		// Enrichment only: whether the source is found changes what a later
		// call KNOWS, never whether this statement parses. `use` is keyword,
		// bareword, optional list, semicolon, and that is true whether or not
		// the file exists.
		if !isVersionPrefix(term.Text) {
			p.res.resolve(term.Text)
		}
		n.Children = append(n.Children, term)
	}

	// The import list, or a version. A version lexes as Word("v5") then
	// Operator(".") then Number, and the Word half is taken as the name
	// above -- the rest reads as an expression here, which round-trips even
	// though the shape is not what perl's grammar calls a version.
	if next, ok := p.peekSignificant(); ok && next.Kind != lexer.Semicolon {
		if arg := p.parseExpr(0); arg != nil {
			n.Children = append(n.Children, arg)
		}
	}

	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// isVersionPrefix reports whether a word is the `v5` of a `v5.36`.
func isVersionPrefix(text string) bool {
	if len(text) < 2 || (text[0] != 'v' && text[0] != 'V') {
		return false
	}
	for i := 1; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

// parsePhaser: `BEGIN { ... }` and the other four.
func (p *parser) parsePhaser(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Phaser, Text: p.text(word), Start: word.Start}

	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// parseClass: 5.38's `class NAME { ... }` and `class NAME;`.
//
// Structurally a package with a different keyword, so it shares
// finishBodyOrSemicolon. `field` and `method` inside the body are handled by
// parseDeclaration, which already knows `method` as a sub spelling.
func (p *parser) parseClass(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		})
	}

	// An optional version: `class Point 1.0 { }`.
	if v, ok := p.peekSignificant(); ok && v.Kind == lexer.Number {
		p.advanceTo(v)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(v), Start: v.Start, End: v.End,
		})
	}

	// Attributes: `class Point :isa(Shape) { }`.
	p.parseAttributes(n)

	p.finishBodyOrSemicolon(n)
	return n
}

// parseAttributes reads a run of `:name` or `:name(args)` and appends them.
//
// Attributes hang off classes, fields, methods and subs alike. They are kept
// as Term children rather than given their own kind: nothing in this
// milestone reads them, and a kind nobody reads is a kind that drifts.
func (p *parser) parseAttributes(n *Node) {
	for {
		colon, ok := p.peekSignificant()
		if !ok || p.text(colon) != ":" {
			return
		}
		name, ok := p.peekAfter(colon)
		if !ok || name.Kind != lexer.Word {
			return
		}
		p.advanceTo(name)
		attr := &Node{
			Kind: Term, Text: ":" + p.text(name),
			Start: colon.Start, End: name.End,
		}

		// An optional parenthesised argument, taken as an opaque span: an
		// attribute's argument is not Perl in general (`:isa(Shape)` is, but
		// `:lvalue` takes none and XS attributes take arbitrary text).
		if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
			depth := 0
			for p.pos < len(p.toks) {
				tok := p.toks[p.pos]
				p.pos++
				if p.text(tok) == "(" {
					depth++
				}
				if tok.Kind == lexer.CloseBracket && p.src[tok.Start] == ')' {
					depth--
					if depth == 0 {
						break
					}
				}
			}
			attr.End = p.prevEnd()
		}
		n.Children = append(n.Children, attr)
	}
}

// peekAfter returns the next significant token after the given one.
func (p *parser) peekAfter(tok lexer.Token) (lexer.Token, bool) {
	save := p.pos
	p.advanceTo(tok)
	next, ok := p.peekSignificant()
	p.pos = save
	return next, ok
}
