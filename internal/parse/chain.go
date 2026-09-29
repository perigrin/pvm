// ABOUTME: Chained comparisons, 5.32+: one n-ary node so each operand is evaluated once.
// ABOUTME: The two yyerror productions of perly.y:1466-1478 are the entire non-chaining rule.

package parse

// cmpClass is §4.3's split of the comparison tokens. Chaining is a property
// of the CLASS, not of the level: CHRELOP and NCRELOP share level 17 and
// behave differently, which is why a binding-power table cannot express this.
type cmpClass int

const (
	notComparison cmpClass = iota
	// chRelop chains: < > <= >= lt gt le ge
	chRelop
	// chEqop chains: == != eq ne
	chEqop
	// ncRelop does not: isa
	ncRelop
	// ncEqop does not: <=> cmp ~~
	ncEqop
)

var cmpClasses = map[string]cmpClass{
	"<": chRelop, ">": chRelop, "<=": chRelop, ">=": chRelop,
	"lt": chRelop, "gt": chRelop, "le": chRelop, "ge": chRelop,

	"==": chEqop, "!=": chEqop, "eq": chEqop, "ne": chEqop,

	"isa": ncRelop,

	"<=>": ncEqop, "cmp": ncEqop, "~~": ncEqop,
}

// compareClass reports the operator's class and whether it may chain.
func compareClass(text string) (cmpClass, bool) {
	c := cmpClasses[text]
	return c, c == chRelop || c == chEqop
}

// chainClass reports which chain a node already is, so a following operator
// can decide whether to extend it.
//
// A chain may not MIX classes: `$a < $b == $c` is two different comparisons,
// not one three-operand chain, because CHRELOP and CHEQOP are separate
// classes in perly.y. The relop chain is the left operand of the eqop.
func chainClass(n *Node) cmpClass {
	if n.Kind != CmpChain {
		return notComparison
	}
	c, _ := compareClass(n.Text)
	return c
}

// parseComparison is the whole of perly.y:1466-1478:
//
//	termrelop:	relopchain %prec PREC_LOW { cmpchain_finish(...) }
//		|	term NCRELOP term         { newBINOP(...) }
//		|	termrelop NCRELOP  { yyerror("syntax error"); YYERROR; }
//		|	termrelop CHRELOP  { yyerror("syntax error"); YYERROR; }
//
// The two yyerror productions are the entire non-chaining rule: once a
// termrelop has been reduced, NO further relational operator may follow --
// not a chaining one, not a non-chaining one. Measured:
//
//	$ perl -e 'my $q = $a <=> $b <=> $c'
//	syntax error at -e line 1, near "$b <=>"
//	$ perl -e 'my $q = $a < $b isa Foo'
//	Bareword found where operator expected
//
// So a comparison does one of three things, and which one depends on what
// stands to its left:
//
//	left is not a comparison    start: a chain if chainable, else a Binary
//	left is a chain, same class extend it by one operand
//	left is any other comparison of this level  reject
//	left is a comparison of the other level     start, with it as an operand
func (p *parser) parseComparison(left *Node, op OpInfo, cls cmpClass, minBP int) *Node {
	tok, _ := p.peekSignificant()
	text := p.text(tok)
	chainable := cls == chRelop || cls == chEqop

	// Extend: the standing chain is of this operator's own class.
	if isComparison(left) && chainClass(left) == cls {
		p.advanceTo(tok)
		right := p.operand(op.BP, tok)
		left.Children = append(left.Children, right)
		left.End = right.End
		return left
	}

	// Reject: a comparison of this operator's level already stands, and
	// this one cannot join it. Covers `$a <=> $b <=> $c` (ncEqop never
	// chains) and `$a < $b isa Foo` (different class, same level). A
	// comparison of the other level is an ordinary operand: `$a < $b == $c`
	// is a termeqop over a termrelop.
	if isComparison(left) && sameLevel(comparisonClass(left), cls) {
		start := left.Start
		p.skipToStatementEnd()
		return &Node{Kind: Unknown, Refusal: ChainClassMismatch, Start: start, End: p.prevEnd()}
	}

	// Start.
	p.advanceTo(tok)
	right := p.operand(op.BP, tok)
	kind := Binary
	if chainable {
		kind = CmpChain
	}
	return &Node{
		Kind: kind, Text: text,
		Start: left.Start, End: right.End,
		Children: []*Node{left, right},
	}
}

// comparisonClass reports the class of the operator a comparison node was
// built from.
func comparisonClass(n *Node) cmpClass {
	c, _ := compareClass(n.Text)
	return c
}

// sameLevel reports whether two classes share a precedence level: the
// relational classes sit at one level, the equality classes at the next.
func sameLevel(a, b cmpClass) bool {
	relational := func(c cmpClass) bool { return c == chRelop || c == ncRelop }
	return relational(a) == relational(b)
}

// isComparison reports whether this node is already a reduced termrelop --
// a chain, or a Binary built from a comparison operator. A parenthesised one
// is not: the parentheses make it a term, so `($a == $b) == 0` compares the
// first result with 0 rather than chaining.
func isComparison(n *Node) bool {
	if n.Paren {
		return false
	}
	if n.Kind == CmpChain {
		return true
	}
	if n.Kind != Binary {
		return false
	}
	c, _ := compareClass(n.Text)
	return c != notComparison
}
