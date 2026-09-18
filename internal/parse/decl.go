// ABOUTME: Declarations: my/our/local/state, sub in both body forms, package in both forms.
// ABOUTME: A declaration's target list is not a call — `my ($a, $b)` declares two variables.

package parse

import "tamarou.com/pvm/internal/lexer"

// declarators are the variable-introducing keywords. `local` is not one
// strictly -- it saves and restores a global rather than creating a lexical --
// but it takes the same target forms and parses identically.
var declarators = map[string]bool{
	"my": true, "our": true, "local": true, "state": true,

	// 5.38's `field $x :param = $default;`. Structurally a declarator with
	// attributes, which `my` also accepts (`my $x :shared`), so it takes the
	// same path rather than one of its own.
	"field": true,
}

// parseDeclaration parses one declaration, or returns nil if this word does
// not start one.
func (p *parser) parseDeclaration(word lexer.Token) *Node {
	text := p.text(word)
	switch {
	case declarators[text]:
		return p.parseVarDecl(word)
	case text == "sub" || text == "method":
		// `sub {` with no name is an anonymous sub, which is an EXPRESSION
		// and not a declaration. Declining here sends it to parseTerm, whose
		// anon-sub branch (`term.go:126-132`) has always read it correctly --
		// `my $f = sub { 1 }` was right all along because the expression
		// path is reached there.
		//
		// In statement position parseDeclaration ran first and produced a
		// Declaration spanning only `sub`, then a SECOND statement holding
		// an AnonHash for the body. A tree that is not a parse of its
		// source, and it round-trips.
		//
		// perl makes the same test: a name after `sub` is a declaration, a
		// `{` is an anonymous sub.
		if next, ok := p.peekAfter(word); ok && p.text(next) == "{" {
			return nil
		}
		return p.parseSubDecl(word)
	case text == "package":
		return p.parsePackageDecl(word)
	}
	return nil
}

// parseVarDecl: `my $x`, `my ($a, $b) = @_`, and the rest.
//
// The target is parsed as an expression above the comma, so `my ($a, $b)`
// yields the parenthesised list and `= @_` is picked up by the caller as an
// ordinary assignment. That keeps the declaration from re-implementing
// assignment, which is already at level 9.
func (p *parser) parseVarDecl(word lexer.Token) *Node {
	n := p.parseVarDeclNoSemi(word)
	// The terminating `;` belongs to the declaration, like any other
	// statement's. Without it the statement ends before the semicolon and
	// the leftover becomes an Unknown sitting beside a perfectly good tree.
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
		n.End = p.prevEnd()
	}
	return n
}

// parseVarDeclNoSemi is parseVarDecl without consuming a terminator.
//
// The init clause of a C-style for head is a declaration whose `;` is the
// head's SEPARATOR, not the declaration's terminator: `for (my $i = 0; ...)`.
// Eating it there loses the head's structure.
func (p *parser) parseVarDeclNoSemi(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	// The variable, then any attributes, then an optional initialiser.
	// Attributes come between: `field $x :param = 1` and `my $x :shared = 1`
	// both put them after the name, where an expression parser would read
	// the `:` as a ternary's colon with nothing to match.
	if v, ok := p.peekSignificant(); ok && v.Kind == lexer.Variable {
		p.advanceTo(v)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(v), Start: v.Start, End: v.End,
		})
		p.parseAttributes(n)

		// The initialiser, if there is one.
		if eq, ok := p.peekSignificant(); ok && p.text(eq) == "=" {
			p.advanceTo(eq)
			if init := p.parseExpr(0); init != nil {
				n.Children = append(n.Children, init)
			}
		}
		n.End = p.prevEnd()
		return n
	}

	// A list target -- `my ($a, $b) = @_` -- or anything else, parsed as one
	// expression so the assignment at level 9 is already inside it.
	if target := p.parseExpr(0); target != nil {
		n.Children = append(n.Children, target)
	}
	n.End = p.prevEnd()
	return n
}

// parseSubDecl: `sub NAME [PROTO] ( BLOCK | ";" )`, spec §5.5.2.
//
// The bodiless form is §0.13 rank 4, 7 corpus files. For a fresh parser it is
// just the `;` alternative of the grammar -- the "statement-level leak" the
// findings describe was a tree-sitter recovery artifact, not a subtlety of
// the language.
func (p *parser) parseSubDecl(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	// The name. An anonymous sub has none, and `sub { ... }` is an
	// expression rather than a declaration -- but it reaches here only as a
	// statement, where perl also treats it as a declaration of nothing.
	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		})
	}

	// The prototype, if the lexer found one. Recognition is its job; this
	// only carries the result so M4 has something to resolve.
	if proto, ok := p.peekSignificant(); ok && proto.Kind == lexer.Prototype {
		p.advanceTo(proto)
		n.Children = append(n.Children, &Node{
			Kind: PrototypeNode, Text: p.text(proto),
			Start: proto.Start, End: proto.End,
		})
	} else if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
		// A SIGNATURE. With `use v5.36` the lexer declines to scan a
		// prototype (`proto.go:32-35`) because the two are different things
		// -- measured, `prototype(\&g)` is undef for a signatured sub -- and
		// nothing downstream picked the parens up. The body was then read as
		// a hash slice OF the signature and the sub body was LOST, so
		// turning the feature ON made the parse worse.
		//
		// Read as an ordinary parenthesised list, which is what a signature
		// is shaped like: `($x, $y)`, `($x = 1)`, `(@rest)`. Whether each
		// element is a parameter, and what its default means, is M2's --
		// this milestone owns the syntax.
		if sig := p.parseTerm(); sig != nil {
			n.Children = append(n.Children, sig)
		}
	}

	p.finishBodyOrSemicolon(n)
	return n
}

// parsePackageDecl: `package NAME;` and `package NAME { ... }`.
//
// The two differ in scope -- to the end of the enclosing block, or to its own
// braces -- which is a later concern. Here they differ only in which
// alternative finishes them.
func (p *parser) parsePackageDecl(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		})
	}

	// An optional version: `package Foo 1.0;`.
	if v, ok := p.peekSignificant(); ok && v.Kind == lexer.Number {
		p.advanceTo(v)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(v),
			Start: v.Start, End: v.End,
		})
	}

	p.finishBodyOrSemicolon(n)
	return n
}

// finishBodyOrSemicolon consumes the `( Block | ";" )` that ends a sub or
// package declaration, appending the block when there is one.
func (p *parser) finishBodyOrSemicolon(n *Node) {
	if tok, ok := p.peekSignificant(); ok {
		switch {
		case tok.OpensBlock && p.text(tok) == "{":
			n.Children = append(n.Children, p.parseBlock(tok))
		case tok.Kind == lexer.Semicolon:
			p.advanceTo(tok)
		}
	}
	n.End = p.prevEnd()
}
