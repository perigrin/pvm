// ABOUTME: The Pratt loop: binding powers drive grouping, and nonassoc is checked separately.
// ABOUTME: A table alone cannot reject `1 .. 2 .. 3` — it stops recursion but still accepts.

package parse

import "tamarou.com/pvm/internal/lexer"

// parseExpr consumes infix operators whose binding power exceeds minBP.
//
// The shape is §4.2.1's. What the spec's sketch does not carry, and what the
// measurements forced, is the nonassoc check: a binding power stops the
// recursion at an equal-power operator, but stopping is not rejecting.
//
//	$ perl -e 'my $x = 1 .. 2 .. 3;'
//	syntax error at -e line 1, near "2 .."
//
// Eleven of the 32 levels are %nonassoc, so this is a rule, not an edge.
func (p *parser) parseExpr(minBP int) *Node {
	left := p.parseTerm()
	if left == nil {
		// Input ran out where a term was expected. Nothing to build and
		// nothing consumed; the statement loop decides what that means.
		return nil
	}

	for {
		tok, ok := p.peekSignificant()
		if !ok {
			return left
		}
		text := p.text(tok)

		op, ok := infix[text]
		if !ok || op.BP <= minBP {
			return left
		}

		// Comparisons take their own path: one may START a chain, EXTEND the
		// chain it matches, or REJECT where a chain already stands.
		if cls, _ := compareClass(text); cls != notComparison {
			left = p.parseComparison(left, op, cls, minBP)
			continue
		}

		// Nonassoc: the operator binds, but a second one at the same level
		// is an error rather than a grouping. Detected by parsing the right
		// operand and then looking for a repeat.
		//
		// `++` and `--` are level 27 and nonassoc, but they are POSTFIX --
		// they take no right operand at all. Reaching parseNonassoc for them
		// made `$i++` parse as a binary node with an Unknown for the operand
		// it does not have, which is how `for (...; $i++)` fell to Unknown.
		if op.Assoc == AssocNone && text != "++" && text != "--" {
			left = p.parseNonassoc(left, op, tok)
			continue
		}

		switch text {
		case "?":
			left = p.parseTernary(left, op)
		case "++", "--":
			p.advanceTo(tok)
			left = &Node{
				Kind: Postfix, Text: text,
				Start: left.Start, End: p.prevEnd(),
				Children: []*Node{left},
			}
		case "(", "[", "{":
			left = p.parseSubscript(left, text)
		case "->":
			// `->` before a subscript opener is the SAME operation as the
			// subscript without it -- §4.14 gives both one `Index` node with
			// an `Arrow` flag, not two unrelated shapes.
			//
			// Falling through to the default made the `{k}` of `$h->{k}` a
			// right operand, which parseTerm reads as an anonymous hash
			// because that is what a `{` in term position means. The tree
			// then said a hash was being CONSTRUCTED where one was being
			// indexed. `$a->[0]` said the same about an array.
			//
			// The arrow's own bytes stay in the Index span, so round-trip is
			// unaffected; what changes is that one operation has one shape.
			if next, ok := p.peekAfter(tok); ok {
				if open := p.text(next); open == "[" || open == "{" {
					p.advanceTo(tok)
					left = p.parseSubscript(left, open)
					// The flag this branch's comment has promised since
					// d2ebe02a. `$h{k}` and `$h->{k}` are one shape and
					// one field apart, and they read different variables.
					left.Arrow = true
					continue
				}
			}
			// `->(` is a code dereference and `->name` a method call.
			// Neither is a subscript; both keep the Binary shape.
			p.advanceTo(tok)
			right := p.operand(op.rightBP(), tok)
			left = &Node{
				Kind: Binary, Text: text,
				Start: left.Start, End: right.End,
				Children: []*Node{left, right},
			}
		default:
			p.advanceTo(tok)
			right := p.operand(op.rightBP(), tok)
			left = &Node{
				Kind: Binary, Text: text,
				Start: left.Start, End: right.End,
				Children: []*Node{left, right},
			}
		}
	}
}

// operand parses a right-hand operand, substituting an explicit Unknown when
// the input ran out.
//
// Every infix and prefix form needs this, and putting it in one place is the
// point: a nil check repeated at eleven call sites is a nil check forgotten
// at one of them. The Unknown spans the operator that has no operand, so the
// bytes stay in the tree and the failure is visible rather than silent.
func (p *parser) operand(bp int, after lexer.Token) *Node {
	if n := p.parseExpr(bp); n != nil {
		return n
	}
	return &Node{Kind: Unknown, Start: after.Start, End: after.End}
}

// parseNonassoc builds the node and then rejects a repetition.
//
// `$a .. $b` is fine; `$a .. $b .. $c` is not. The second operator is
// consumed into an Unknown so the bytes stay in the tree -- declining is not
// the same as dropping, and round-trip holds either way.
func (p *parser) parseNonassoc(left *Node, op OpInfo, tok lexer.Token) *Node {
	text := p.text(tok)
	p.advanceTo(tok)
	right := p.operand(op.BP, tok)
	n := &Node{
		Kind: Binary, Text: text,
		Start: left.Start, End: right.End,
		Children: []*Node{left, right},
	}

	next, ok := p.peekSignificant()
	if !ok {
		return n
	}
	if nextOp, isOp := infix[p.text(next)]; isOp && nextOp.Level == op.Level {
		start := n.Start
		p.skipToStatementEnd()
		return &Node{Kind: Unknown, Start: start, End: p.prevEnd()}
	}
	return n
}

// parseTernary: `?:` is right associative, so the else-branch is parsed at
// BP-1 and a second ternary to the right nests inside it. Measured:
//
//	$ perl -MO=Deparse -e 'my $x = $a ? $b : $c ? $d : $e;'
//	my $x = $a ? $b : ($c ? $d : $e);
func (p *parser) parseTernary(cond *Node, op OpInfo) *Node {
	tok, _ := p.peekSignificant()
	p.advanceTo(tok)

	// The then-branch is parsed at the lowest power: everything up to the
	// `:` belongs to it, commas included.
	then := p.operand(0, tok)

	colon, ok := p.peekSignificant()
	if !ok || p.text(colon) != ":" {
		// A `?` with no `:` is not a ternary. Declining rather than
		// inventing the missing branch.
		return &Node{Kind: Unknown, Start: cond.Start, End: then.End}
	}
	p.advanceTo(colon)

	els := p.operand(op.rightBP(), colon)
	return &Node{
		Kind: Ternary, Text: "?:",
		Start: cond.Start, End: els.End,
		Children: []*Node{cond, then, els},
	}
}

// closerFor is the bracket that closes this opener.
func closerFor(open string) string {
	switch open {
	case "(":
		return ")"
	case "[":
		return "]"
	case "{":
		return "}"
	}
	return ""
}

// peekSubscript reports whether a subscript opens here, returning the bracket
// and whether an arrow preceded it. It consumes the arrow and nothing else.
//
// `$h{k}` and `$h->{k}` are one operation and one flag apart (§4.14), so the
// two callers that walk a postfix chain ask the same question in one place.
func (p *parser) peekSubscript() (open string, arrow bool, ok bool) {
	t, ok := p.peekSignificant()
	if !ok {
		return "", false, false
	}
	switch text := p.text(t); text {
	case "[", "{":
		return text, false, true
	case "->":
		next, ok := p.peekAfter(t)
		if !ok {
			return "", false, false
		}
		if o := p.text(next); o == "[" || o == "{" {
			p.advanceTo(t)
			return o, true, true
		}
	}
	return "", false, false
}

// parseSubscript handles the postfix chain: a call, an index, a key.
func (p *parser) parseSubscript(left *Node, open string) *Node {
	tok, _ := p.peekSignificant()
	p.advanceTo(tok)

	// An empty subscript or call -- `f()`, `$r->()` -- has no inner
	// expression, which is not a failure.
	children := []*Node{left}
	if c, ok := p.peekSignificant(); !ok || p.text(c) != closerFor(open) {
		if inner := p.parseExpr(0); inner != nil {
			children = append(children, inner)
		}
	}
	if c, ok := p.peekSignificant(); ok && p.text(c) == closerFor(open) {
		p.advanceTo(c)
	}
	if open == "{" && len(children) == 2 {
		autoquote(children[1])
	}
	return &Node{
		Kind: Index, Text: open,
		Start: left.Start, End: p.prevEnd(),
		Children: children,
	}
}

// autoquote turns a lone bareword hash key into the string it denotes.
//
// `$h{k}` is the key "k" and `$h{k()}` is the sub's return value -- different
// programs. Perl autoquotes even when a sub of that name is in scope:
//
//	$ perl -e 'sub k { "z" } my %h = (k => 1); print $h{k}, "\n";'
//	1
//
// The rule is narrower than "a bareword in braces". A SLICE is not
// autoquoted, so only a lone word filling the whole subscript qualifies:
//
//	$ perl -e 'sub a { "z" } my %h=(a=>1,b=>2); my @s=@h{a,b};
//	           print join(",", map { $_ // "undef" } @s), "\n";'
//	undef,2
//
// `a` called a(); `b` had no sub and autoquoted. That is why this tests the
// node rather than scanning the subscript for words -- a comma leaves a
// Binary here, which is not a Call and is left alone.
//
// Beyond the wrong tree: Call{Resolved:false} means "a call to something I
// have not seen" (§4.8.3), and a hash key is not a call at all. Every
// bareword subscript was a false entry in the set sites.go scores.
func autoquote(key *Node) {
	// The span is what separates `k` from `k()`: an explicit call's node
	// covers its parentheses too, while a bareword covers only the word.
	// Children cannot answer this -- an empty argument list leaves none, so
	// `$h{k()}` and `$h{k}` have the same child count and different extents.
	if key.Kind != Call || key.End-key.Start != len(key.Text) {
		return
	}
	key.Kind = Term
	key.Resolved = false
}
