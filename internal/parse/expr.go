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
	return p.parseInfix(left, minBP)
}

// parseInfix is parseExpr's loop with the left operand supplied.
//
// Split out because a parenthesised expression is not a comma list: perly.y
// puts `and`, `or` and `xor` ABOVE the comma, so the paren's contents are a
// full `expr` whose left operand may be the whole list. Measured:
//
//	$ perl -MO=Deparse -e 'my @x = (1, 2 and 3);'
//	my(@x) = ('???', 2) && 3;
//
// The `and` took `(1, 2)` as its left operand. parseParenList assembles that
// list itself, so it needs the loop without the leading parseTerm.
// elementlessAfterComma reports whether a comma followed by tok separates
// nothing: tok cannot begin an element, so the comma is dropped. A closer, a
// terminator, another comma, a statement modifier, and the three word
// operators below the comma.
func elementlessAfterComma(tok lexer.Token, src []byte) bool {
	switch tok.Kind {
	case lexer.CloseBracket, lexer.Semicolon:
		return true
	}
	text := string(src[tok.Start:tok.End])
	// Another separator, plain or fat: `tie $@, => 'main', 1` drops the
	// comma as `(1,,2)` does.
	if text == "," || text == "=>" {
		return true
	}
	if tok.Kind != lexer.Word {
		return false
	}
	op, isInfix := infix[text]
	return modifiers[text] || isInfix && op.BP <= bpBelowComma
}

func (p *parser) parseInfix(left *Node, minBP int) *Node {
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

		// A TRAILING COMMA is a separator with no element after it, not an
		// operator missing its operand.
		//
		// perl allows one before every closer a list can end at, and perl's
		// own test suite leans on it: the multi-line `runperl(`,
		// `foreach my $x (` and `test_opcount(` house style writes the comma
		// after the last argument so that adding one more is a one-line diff.
		//
		// Read as an infix operator, the comma's operand hunt reached the
		// closer and parseTerm consumed it as `not_a_term`. The count that
		// cost was the cheap half; the expensive half is that the closer was
		// then GONE from the construct that owned it, which read the next
		// token as its own:
		//
		//	foreach my $x ($a,) { 1 }
		//
		// canon'd as `foreach my $x ($a , ){1})`, the loop BODY having become
		// a hash subscript on the comma expression. A wrong shape, not a
		// wrong number, and an Unknown count cannot see it.
		//
		// Bucketing every dirty perl.git t/ file's first Unknown span by the
		// construct it stumbled over made this the largest single cause of
		// the 620-file corpus: 27 files and 417 nodes, against 10 files for
		// the runner-up. class/construct.t's note in t2_test.go has named it
		// since the day it was measured -- "nothing in the chain owns
		// trailing commas in argument lists" -- and this is that owner.
		//
		// The comma is CONSUMED before returning, so the closer is what the
		// caller peeks at next -- parseParenList's own loop, parseSubscript's
		// closer check, parseCallArgs' caller -- and each takes it as its
		// own. Returning without consuming would spin: the comma still binds
		// above the caller's minBP.
		//
		// `left` is non-nil by construction here, so there is always an
		// element for the separator to follow. `f(,)` -- a comma with nothing
		// in front of it -- never reaches this loop at all: parseTerm refuses
		// the comma itself, and that refusal is correct and stays.
		//
		// The same holds before anything else that cannot start an element.
		// perl drops the comma, measured on 5.42.0 with -MO=Deparse:
		//
		//	f(1, , 2);               f 1, 2;         another comma
		//	skip "x", 2, if $m;      skip 'x', 2 if $m;   a modifier
		//	close $fh, or die;       die unless close $fh; a word operator
		//
		// So the comma is dropped and the loop goes on. What follows then
		// decides by the loop's own rule: a closer, a terminator or a
		// modifier is not infix and ends the expression; another comma is
		// looked at in turn; `or`, `and` and `xor` are parsed as the
		// operators they are, with this list as their left operand.
		if text == "," || text == "=>" {
			if next, ok := p.peekAfter(tok); ok && elementlessAfterComma(next, p.src) {
				p.advanceTo(tok)
				continue
			}
			// An assignment operator cannot begin an element either, but it
			// binds tighter than the comma, so the list ENDS at the comma
			// rather than becoming its left side: `substr $x, 0, 1, = "a"`
			// is `substr($x, 0, 1) = "a"`, measured on 5.42.0.
			if next, ok := p.peekAfter(tok); ok && next.Kind != lexer.Word &&
				infix[p.text(next)].Level == infix["="].Level {
				p.advanceTo(tok)
				return left
			}
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

			// A WORD after `->` is a METHOD NAME, and a method name is never
			// a parenless list operator -- not even when a sub of that name
			// is declared in this very file. perl dispatches on the invocant
			// at runtime and the name is only ever a name here. Measured on
			// 5.42.0:
			//
			//	$ perl -MO=Deparse -e \
			//	    'package C; sub new { bless {} } package main;
			//	     my $x = C->new->foo;'
			//	my $x = 'C'->new->foo;
			//
			// `new` is declared right there and the chain still reads as two
			// method calls, not as `new` swallowing `->foo` as an argument
			// list.
			//
			// Read here rather than through parseTerm because parseWordTerm's
			// job is to decide what a bareword in TERM position means, and
			// after an arrow the question is already answered. Routing the
			// name through it made every `Foo->new->bar` in a file declaring
			// `sub new` refuse: measured, 19 T1 files regressed that way --
			// `autoload.t`, `tied_hash_autoviv_refloop_cleanup.t` and 17
			// more -- each stranding the second `->` as `not_a_term`.
			//
			// The name takes its OWN argument parens when the source wrote
			// them, and stays a bare Term when it did not. Both halves are
			// forced, in opposite directions, by what canon re-emits:
			//
			//   - With parens, a Term would leave the `(` to the led loop,
			//     which reads it as a subscript ON the arrow: `$o->meth(2)`
			//     came back as `($o -> meth)(2)`, a different grouping. 78 T1
			//     files, measured.
			//   - Without parens, a Call gets canon's "always parenthesised"
			//     treatment and `print $t->join` comes back as
			//     `print($t -> join())`. Faithful forgives a call acquiring
			//     ONE set of parens, not a nested pair, so the outer `print(`
			//     and the inner `join()` together fail it. One T1 file,
			//     measured -- threads_filehandle_inheritance.t.
			//
			// So the node kind follows the source's own spelling, which is
			// what the tree is for. A parenless method call has no argument
			// list to record, and a Term records exactly that.
			if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
				p.advanceTo(name)
				right := &Node{
					Kind: Term, Text: p.text(name),
					Start: name.Start, End: name.End,
				}
				// Only a paren IMMEDIATELY after the name is this method's
				// argument list. Anything else -- another `->`, an operator,
				// a terminator -- belongs to the enclosing expression, which
				// is why there is no parenless argument branch here at all.
				//
				// Resolved is true on the Call: `->name` IS a method call,
				// whatever the invocant turns out to be. The unresolved
				// `Call{Resolved:false}` of §4.8.3 is for a bareword whose
				// MEANING is unsettled, and this one's is settled by the arrow.
				if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
					right.Kind = Call
					right.Resolved = true
					p.advanceTo(open)
					if arg := p.parseCallArgs(); arg != nil {
						right.Children = append(right.Children, arg)
					}
					if close, ok := p.peekSignificant(); ok && p.text(close) == ")" {
						p.advanceTo(close)
					}
					right.End = p.prevEnd()
				}
				left = &Node{
					Kind: Binary, Text: text,
					Start: left.Start, End: right.End,
					Children: []*Node{left, right},
				}
				continue
			}

			right := p.operand(op.rightBP(), tok)
			left = &Node{
				Kind: Binary, Text: text,
				Start: left.Start, End: right.End,
				Children: []*Node{left, right},
			}
		default:
			p.advanceTo(tok)
			if p.emptyDefault(left, op) {
				left = &Node{
					Kind: Binary, Text: text,
					Start: left.Start, End: tok.End,
					Children: []*Node{left},
				}
				continue
			}
			right := p.operand(op.rightBP(), tok)
			left = &Node{
				Kind: Binary, Text: text,
				Start: left.Start, End: right.End,
				Children: []*Node{left, right},
			}
		}
	}
}

// emptyDefault reports whether an assignment has no right operand because it
// is a signature placeholder's empty default: `($=)`, `($y, $ //=)`. perly.y's
// optsigscalardefault may be empty after a placeholder, and only there --
// measured on 5.42.0, `sub c ($x=) {}` is "Optional parameter lacks default
// expression". The Binary keeps its one child and canon writes its source.
func (p *parser) emptyDefault(left *Node, op OpInfo) bool {
	if !p.inSignature || op.Level != assignLevel || left.Kind != Term ||
		len(p.src[left.Start:left.End]) != 1 {
		return false
	}
	next, ok := p.peekSignificant()
	return ok && (p.text(next) == "," || p.text(next) == ")")
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
	return &Node{Kind: Unknown, Refusal: MissingOperand, Start: after.Start, End: after.End}
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
		return &Node{Kind: Unknown, Refusal: NonassocRepeated, Start: start, End: p.prevEnd()}
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
		return &Node{Kind: Unknown, Refusal: TernaryNoColon, Start: cond.Start, End: then.End}
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
