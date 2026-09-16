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
	return &Node{
		Kind: Index, Text: open,
		Start: left.Start, End: p.prevEnd(),
		Children: children,
	}
}
