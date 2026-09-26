// ABOUTME: Control flow: if/unless chains, while/until, both for forms, and labels.
// ABOUTME: A label is recognised by position — could a statement start here — not by the colon.

package parse

import "tamarou.com/pvm/internal/lexer"

// parseControlFlow parses one control-flow statement, or returns nil if this
// word does not start one.
func (p *parser) parseControlFlow(word lexer.Token) *Node {
	switch p.text(word) {
	case "if", "unless":
		return p.parseConditional(word)
	case "while", "until":
		return p.parseWhile(word)
	case "for", "foreach":
		return p.parseFor(word)
	case "last", "next", "redo":
		return p.parseLoopControl(word)
	case "goto":
		// §4.14.2 groups goto with the loop controls under `LoopEx{Op,Label}`,
		// and a BAREWORD after it is a label exactly as it is after `last`.
		// Measured on perl 5.42.0:
		//
		//	$ perl -e 'sub FOO { 42 } goto FOO;'
		//	Can't find label FOO
		//
		// It looks for a label, not for the sub -- so `goto FOO` must not
		// become a call. What goto adds is the other two forms, `goto &f`
		// and `goto $where`, which are expressions rather than labels.
		return p.parseGoto(word)
	case "return":
		// perly.y level 7, a statement form taking an optional list. It
		// arrives with the loop controls because a sub body without `return`
		// is as rare as a loop without `last` -- 27 of the class corpus's
		// Unknowns started here.
		return p.parseReturn(word)
	case "given", "when", "default":
		return p.parseSwitch(word)
	}
	return nil
}

// parseSwitch: `given (EXPR) BLOCK`, `when (EXPR) BLOCK` and `default BLOCK`.
//
// DEPRECATED, and implemented anyway. 5.42.0's own `feature.pm` calls it "the
// Raku given/when construct" and records that it is "enabled by feature bundles
// 5.10 through 5.34, and disabled from the 5.36 feature" bundle onward -- so
// nothing written against a modern bundle gets it without asking. It still
// COMPILES when asked for, perl's own suite still tests it, and `t/op/switch.t`
// is 292 occurrences of nothing else. A parser whose goal is reading perl.git's
// `t/` has to read what perl reads: being out of the default bundle is a
// statement about what programs SHOULD be written, not about what the compiler
// accepts. Verified on 5.42.0:
//
//	$ perl -e 'use feature "switch";
//	      given (1) { when (1) { print "one\n" } default { print "o\n" } }'
//	one
//
// It does not in fact warn here, which is worth recording because the opposite
// is widely assumed: `use warnings` and `-w` both compile the form silently on
// this 5.42.0, so no warning-suppression pragma is needed to exercise it.
//
// Returning nil rather than a keyword reading when no block follows, because
// these three are keywords only in THIS position. Perl lets all three be
// ordinary subs, and a call is what it reads when the block is absent:
//
//	$ perl -e 'sub given { 7 } print given(1), "\n";'
//	7
//	$ perl -MO=Deparse -e 'sub default { 42 } default {a=>1};'
//	default {'a', 1};                       a call with a HASHREF argument
//
// So the block has to be there before the keyword reading is taken, and when
// it is not, nil sends the word back to parseWordTerm as the call it is. The
// lexer settled whether the brace opens a block at all (intuitCurly), which is
// what distinguishes `default { print "d" }` from that hashref argument.
//
// Gated on the keyword rather than added to the general `WORD BLOCK ARG`
// machinery (`01a0d087`, `internal/lexer/keyword.go`'s blockTaking) because
// perl REJECTS the parenthesised shape on an arbitrary word, declared or not:
//
//	$ perl -MO=Deparse -e 'zzz (1) { print "a" } print "b";'
//	syntax error at -e line 1, near ") {"
//	$ perl -MO=Deparse -e 'sub zzz {} zzz (1) { print "a" } print "b";'
//	syntax error at -e line 1, near ") {"
//
// `default BLOCK` alone DID reach the parenless `WORD BLOCK` path and parsed
// there, which is why only the parenthesised two needed this -- but it arrives
// here too, so the three clauses of one construct have one reading and canon
// emits them alike.
//
// Conditional rather than Loop, measured: `given` is not a loop block, so the
// `elsif`-family kind is the honest one and the `Conditional` canon arm emits
// `KEYWORD (COND) BLOCK` already.
//
//	$ perl -e 'use feature "switch"; no warnings; given (1) { last }'
//	Can't "last" outside a loop block
func (p *parser) parseSwitch(word lexer.Token) *Node {
	text := p.text(word)

	// `default` takes NO parens at all, so its block is the very next token.
	// The other two take a parenthesised argument first.
	if text == "default" {
		if next, ok := p.peekAfter(word); !ok || !next.OpensBlock || p.text(next) != "{" {
			return nil
		}
		p.advanceTo(word)
		n := &Node{Kind: Conditional, Text: text, Start: word.Start}
		if blk := p.parseBlockOrDecline(); blk != nil {
			n.Children = append(n.Children, blk)
		}
		n.End = p.prevEnd()
		return n
	}

	if next, ok := p.peekAfter(word); !ok || p.text(next) != "(" {
		return nil
	}

	// Speculate over the head and REWIND if no block follows it, because a
	// `(` is not enough to tell the keyword from the sub call. `if` and
	// `while` can commit on their keyword alone -- nothing else spells them --
	// but `given(1)` with no block is a call perl accepts, and committing here
	// turned it into a Conditional with a condition and no body: a tree that
	// is not a parse of its source, which is the guess the plan forbids.
	//
	// A rewind rather than a balanced-paren lookahead because the head is an
	// arbitrary expression -- `when (@list[0..2])` nests three bracket kinds --
	// and parseExpr is the thing that already knows where it ends. The same
	// save/restore parseLabels uses, for the same reason.
	save := p.pos
	p.advanceTo(word)
	n := &Node{Kind: Conditional, Text: text, Start: word.Start}

	// `when`'s argument is an ORDINARY expression, not a shape of its own.
	// The smartmatch that interprets it is runtime semantics; syntactically a
	// list, a regex and a slice are all just expressions. All compile on
	// 5.42.0:
	//
	//	when ([1,2,3]) {...}      when (/24/) {...}     when (@list[0..2]) {...}
	//
	// so parseParenCondition -- the same head `if` and `while` get -- reads
	// every spelling and no smartmatch-specific rule is needed.
	if cond := p.parseParenCondition(); cond != nil {
		n.Children = append(n.Children, cond)
	}

	blk := p.parseBlockOrDecline()
	if blk == nil {
		p.pos = save
		return nil
	}
	n.Children = append(n.Children, blk)
	n.End = p.prevEnd()
	return n
}

// parseLoopControl: `last`, `next`, `redo`, each with an optional label.
//
// perly.y puts these at level 2 (LOOPEX), a statement form rather than an
// expression operator, and they take an optional term. Measured:
//
//	perl -MO=Deparse -e 'L: while(1){ last }'     ->  last;
//	perl -MO=Deparse -e 'L: while(1){ next L }'   ->  next L;
//
// Without these a loop body containing one falls to Unknown, which makes
// every real loop in the corpus unparseable -- the forms arrive together.
func (p *parser) parseLoopControl(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: LoopControl, Text: p.text(word), Start: word.Start}

	// The label, if there is one. A bareword here is a label rather than a
	// function call: `last FOO` never calls FOO.
	if next, ok := p.peekSignificant(); ok && next.Kind == lexer.Word {
		if !statementKeywords[p.text(next)] {
			p.advanceTo(next)
			n.Children = append(n.Children, &Node{
				Kind: Label, Text: p.text(next),
				Start: next.Start, End: next.End,
			})
		}
	}
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// parseGoto: `goto LABEL`, `goto &NAME`, `goto EXPR`.
//
// A LoopControl like its siblings (§4.14.2 puts all four under `LoopEx`),
// differing only in what it may take. A bareword is a label -- perl looks for
// one and says "Can't find label FOO" rather than calling FOO -- and anything
// else is an expression: `goto &f` is the tail call, `goto $where` a computed
// target.
//
// The semicolon is consumed here, as parseLoopControl and parseReturn do,
// because `statement()` returns a control-flow node directly (`parse.go:430`)
// without a terminator step. Leaving it produced an `Unknown` for the `;`
// alone -- measured before this line was added.
//
// A modifier is left alone: `goto HERE if $x` must reach applyModifier, so
// the label test refuses a modifier word. The terminator bug of 89b1bf3e was
// the inverse -- a modifier crossing a `;` that had already been eaten --
// and it is why the two are separated rather than both handled here.
func (p *parser) parseGoto(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: LoopControl, Text: p.text(word), Start: word.Start}

	next, ok := p.peekSignificant()
	if !ok {
		n.End = p.prevEnd()
		return n
	}

	// A bareword that is not a statement keyword is a label. `goto HERE if
	// $x` must leave `if` to the modifier, which is what the keyword test
	// does.
	if next.Kind == lexer.Word && !statementKeywords[p.text(next)] &&
		!modifiers[p.text(next)] {
		p.advanceTo(next)
		n.Children = append(n.Children, &Node{
			Kind: Label, Text: p.text(next),
			Start: next.Start, End: next.End,
		})
		if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
			p.advanceTo(tok)
		}
		n.End = p.prevEnd()
		return n
	}

	// `goto &f`, `goto $where`. Parsed at the named-unary level so the
	// argument binds the way perl's does -- measured, `((goto $x), $y)`.
	if !endsArgumentList(next, p.src) {
		if arg := p.parseExpr(bpNamedUnary); arg != nil {
			n.Children = append(n.Children, arg)
		}
	}
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// parseReturn: `return;` and `return LIST;`.
func (p *parser) parseReturn(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: LoopControl, Text: p.text(word), Start: word.Start}

	if next, ok := p.peekSignificant(); ok && next.Kind != lexer.Semicolon {
		// Below the comma, so the whole list belongs to the return.
		if arg := p.parseExpr(bpListOp); arg != nil {
			n.Children = append(n.Children, arg)
		}
	}
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// parseConditional: `if (EXPR) BLOCK` with optional elsif and else.
//
// The elsif chain is flattened into siblings rather than nested: `elsif` is
// spelled as one word in Perl and nesting it as `else { if ... }` would make
// the tree disagree with the source an LSP has to render.
func (p *parser) parseConditional(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Conditional, Text: p.text(word), Start: word.Start}

	if cond := p.parseParenCondition(); cond != nil {
		n.Children = append(n.Children, cond)
	}
	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}

	for {
		next, ok := p.peekSignificant()
		if !ok || next.Kind != lexer.Word {
			break
		}
		switch p.text(next) {
		case "elsif":
			p.advanceTo(next)
			branch := &Node{Kind: Conditional, Text: "elsif", Start: next.Start}
			if cond := p.parseParenCondition(); cond != nil {
				branch.Children = append(branch.Children, cond)
			}
			if blk := p.parseBlockOrDecline(); blk != nil {
				branch.Children = append(branch.Children, blk)
			}
			branch.End = p.prevEnd()
			n.Children = append(n.Children, branch)
		case "else":
			p.advanceTo(next)
			branch := &Node{Kind: Conditional, Text: "else", Start: next.Start}
			if blk := p.parseBlockOrDecline(); blk != nil {
				branch.Children = append(branch.Children, blk)
			}
			branch.End = p.prevEnd()
			n.Children = append(n.Children, branch)
			n.End = p.prevEnd()
			return n
		default:
			n.End = p.prevEnd()
			return n
		}
	}
	n.End = p.prevEnd()
	return n
}

// parseContinueOrDecline reads the optional `continue BLOCK` that follows a
// loop body, and returns nil when there is none.
//
// A Loop node with `continue` in Text and the block as its only child, which
// is the kind's documented pattern -- "one kind with the keyword in Text" --
// and which canon's `case Conditional, Loop` arm already emits as
// `KEYWORD BLOCK` with no further change.
//
// There is no `continue` op to model. The destination is metadata on the loop's
// own `enterloop`, which is why the block is part of the loop it follows rather
// than a construct of its own. Measured with `perl -MO=Concise,-exec` over the
// corpus case in `conformance/mdtest/loops.md`:
//
//	d  <{> enterloop(next->12 last->1c redo->e) v
//	12     <0> pushmark s        the continue block starts at the `next` target
//	17     <0> unstack v
//	1c <2> leaveloop vKP/2       and `last` lands past it
//
// So `next` reaches the block and `last` jumps over it, which is observable:
//
//	while ($i<5) { $i++; next if $i==2; push @s,"b$i" } continue { push @s,"c$i" }
//	  ->  b1 c1 c2 b3 c3 b4 c4 b5 c5      it RUNS on `next`
//	while ($i<5) { $i++; last if $i==3; push @s,"b$i" } continue { push @s,"c$i" }
//	  ->  b1 c1 b2 c2                     it does NOT run on `last`
//
// Called from the three sites perl accepts one at -- `while`/`until`, the LIST
// form of `for`/`foreach`, and a bare block -- rather than from
// parseBlockOrDecline, because the C-style `for (;;)` head shares that helper
// and perl REJECTS a continue block after it (see parseFor).
func (p *parser) parseContinueOrDecline() *Node {
	word, ok := p.peekSignificant()
	if !ok || word.Kind != lexer.Word || p.text(word) != "continue" {
		return nil
	}
	// The block has to be there before this reading is taken, because BARE
	// `continue;` is a DIFFERENT statement form -- the jump out of a `when`
	// block. Measured on 5.42.0:
	//
	//	$ perl -e 'use feature "switch"; no warnings;
	//	      for (1) { when (1) { print "a\n"; continue } print "b\n" }'
	//	a
	//	b
	//	$ perl -e 'continue: while (1) { last continue }'
	//	Can't "continue" outside a when block at -e line 1.
	//
	// That form is not this one, so with no block following, nil sends the word
	// back to the ordinary paths -- it reads as a Call, the same reading any
	// undeclared bareword gets, which is `wider` rather than WRONG and not a
	// claim to have understood the `when`-jump. Folding it into a loop clause it
	// is not WOULD have been such a claim.
	next, ok := p.peekAfter(word)
	if !ok || !next.OpensBlock || p.text(next) != "{" {
		return nil
	}
	p.advanceTo(word)
	n := &Node{Kind: Loop, Text: p.text(word), Start: word.Start}
	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}
	n.End = p.prevEnd()
	return n
}

// parseWhile: `while (EXPR) BLOCK`, and `until` which differs only in Text.
func (p *parser) parseWhile(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Loop, Text: p.text(word), Start: word.Start}

	if cond := p.parseParenCondition(); cond != nil {
		n.Children = append(n.Children, cond)
	}
	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}
	if cont := p.parseContinueOrDecline(); cont != nil {
		n.Children = append(n.Children, cont)
	}
	n.End = p.prevEnd()
	return n
}

// parseFor handles both heads.
//
// `for (INIT; COND; STEP)` and `for VAR (LIST)` share a keyword and nothing
// else structurally, and which one this is cannot be known until the head is
// read: the C-style form is three expressions separated by semicolons, and
// the list form is one expression. Reading the head and counting its
// semicolons settles it without lookahead over the whole parenthesised group.
func (p *parser) parseFor(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Loop, Text: p.text(word), Start: word.Start}

	// An optional loop variable: `for my $x (...)`, `for $x (...)`.
	//
	// Only when it comes BEFORE the parenthesised head. `for (my $i = 0; ...)`
	// has its `my` INSIDE the head, where it is the C-style init clause, and
	// consuming it here leaves the head unparseable -- measured: the whole
	// statement fell to Unknown.
	if next, ok := p.peekSignificant(); ok && p.text(next) != "(" {
		if next.Kind == lexer.Word && declarators[p.text(next)] {
			p.advanceTo(next)
			decl := &Node{Kind: Declaration, Text: p.text(next), Start: next.Start}
			if v, ok := p.peekSignificant(); ok && v.Kind == lexer.Variable {
				p.advanceTo(v)
				decl.Children = append(decl.Children, &Node{
					Kind: Term, Text: p.text(v), Start: v.Start, End: v.End,
				})
			}
			decl.End = p.prevEnd()
			n.Children = append(n.Children, decl)
		} else if next.Kind == lexer.Variable {
			p.advanceTo(next)
			n.Children = append(n.Children, &Node{
				Kind: Term, Text: p.text(next), Start: next.Start, End: next.End,
			})
		}
	}

	head, cStyle := p.parseForHead()
	if head != nil {
		n.Children = append(n.Children, head...)
	}
	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}
	// Only the LIST form takes a continue block. perl rejects it after the
	// C-style head, measured on 5.42.0:
	//
	//	$ perl -e 'for (my $i=0; $i<3; $i++) { } continue { }'
	//	syntax error at -e line 1, near "} continue "
	//
	// so folding one in there would read a construct perl does not accept, and
	// refusing the PAIRING is the point. The `continue` then becomes its own
	// statement -- measured, a Call with a Block, the general `WORD BLOCK`
	// reading -- which claims nothing about a loop it does not belong to.
	if !cStyle {
		if cont := p.parseContinueOrDecline(); cont != nil {
			n.Children = append(n.Children, cont)
		}
	}
	n.End = p.prevEnd()
	return n
}

// parseForHead reads `( ... )` and returns its parts: one node for a list
// head, three for a C-style head. An empty slot in `for (;;)` contributes no
// node, which is why the count is not load-bearing anywhere.
//
// The second return says the head was C-STYLE, decided by whether a semicolon
// separated its clauses rather than by counting the parts -- `for (;;)`
// contributes zero nodes and is still C-style. It is what tells parseFor which
// of the two forms may take a continue block.
func (p *parser) parseForHead() ([]*Node, bool) {
	open, ok := p.peekSignificant()
	if !ok || p.text(open) != "(" {
		return nil, false
	}
	p.advanceTo(open)

	var parts []*Node
	cStyle := false
	for {
		tok, ok := p.peekSignificant()
		if !ok {
			return parts, cStyle
		}
		switch {
		case p.text(tok) == ")":
			p.advanceTo(tok)
			return parts, cStyle
		case tok.Kind == lexer.Semicolon:
			// An empty slot: `for (;;)`.
			cStyle = true
			p.advanceTo(tok)
			continue
		}
		// The init clause of a C-style head is often a declaration --
		// `for (my $i = 0; ...)` -- and a declaration is not an expression,
		// so the expression parser cannot read it.
		var part *Node
		if tok.Kind == lexer.Word && declarators[p.text(tok)] {
			part = p.parseVarDeclNoSemi(tok)
		} else {
			part = p.parseExpr(0)
		}
		if part == nil {
			return parts, cStyle
		}
		parts = append(parts, part)

		if sep, ok := p.peekSignificant(); ok && sep.Kind == lexer.Semicolon {
			cStyle = true
			p.advanceTo(sep)
		}
	}
}

// parseParenCondition reads `( EXPR )`, which every conditional and while
// loop has and which is not optional in Perl.
func (p *parser) parseParenCondition() *Node {
	open, ok := p.peekSignificant()
	if !ok || p.text(open) != "(" {
		return nil
	}
	p.advanceTo(open)

	cond := p.parseExpr(0)
	if close, ok := p.peekSignificant(); ok && p.text(close) == ")" {
		p.advanceTo(close)
	}
	return cond
}

// parseBlockOrDecline reads the `{ ... }` a control-flow statement requires.
//
// Returns nil when there is no block -- a half-typed `if ($x)` with nothing
// after it, which an LSP sees constantly. The caller's node still spans what
// was read, so round-trip holds.
func (p *parser) parseBlockOrDecline() *Node {
	tok, ok := p.peekSignificant()
	if !ok || p.text(tok) != "{" {
		return nil
	}
	return p.parseBlock(tok)
}

// modifiers are the six statement-modifier keywords.
var modifiers = map[string]bool{
	"if": true, "unless": true,
	"while": true, "until": true,
	"for": true, "foreach": true,
}

// applyModifier wraps an already-parsed expression in the modifier that
// follows it, if one does.
//
// The modifier INVERTS the tree: `print if $x` is a conditional whose body is
// a print. And it binds looser than everything in the expression grammar --
// looser than `or`, which is itself level 4, the loosest operator there is.
// Measured:
//
//	$ perl -MO=Deparse -e 'print("a") or die("b") if $x;'
//	print 'a' or die 'b' if $x;
//
// That is why this runs AFTER parseExpr rather than inside it: the expression
// parser has already taken everything it can, and whatever it took is the
// modifier's body entire. A binding power low enough to express this would
// have to sit below 0, which is the loop's own floor.
// endsInBlock reports whether an UNRESOLVED `WORD BLOCK` call's text ends at
// the block's `}`, so the statement is already closed and a following word
// starts a new one.
//
// Three conditions, each one measured rather than assumed:
//
//   - a Call whose LAST child is the Block. `defer { ... }` ends there;
//     `defer { ... } print "b"` has a list after it and ends where the list
//     does; `$x = sub { 1 }` is an assignment, not a call.
//
//   - UNRESOLVED. `do BLOCK while COND` is a genuine modifier and perl spells
//     it that way -- measured on 5.42.0:
//
//     $ perl -MO=Deparse -e '$x=0; do {$x++;} while ($x < 10);'
//     do { ++$x } while $x < 10;
//
//     `do` is a builtin, so its call is Resolved; the bare `WORD BLOCK` form is
//     not, because its callee is a bareword whose declaration this parser has
//     not seen. `cmd/mod.t` and `comp/require.t` are where treating the two
//     alike refused the `do` form's modifier.
//
// The narrowness is the point -- a wider rule refuses the genuine modifier on
// any statement whose last token happens to be a brace.
func endsInBlock(body *Node) bool {
	if body == nil || body.Kind != Call || body.Resolved || len(body.Children) == 0 {
		return false
	}
	return body.Children[len(body.Children)-1].Kind == Block
}

func (p *parser) applyModifier(body *Node, start int) *Node {
	// A statement that already ended cannot take a modifier. `my $z;` is
	// finished at its semicolon, and the `foreach` on the NEXT line starts a
	// new statement -- it is not a modifier on the declaration.
	//
	// Without this check every declaration followed by a control-flow
	// statement swallowed it: `my $z;\nforeach my $e (@a) { ... }` parsed as
	// one Loop containing the declaration, and the loop's body fell to
	// Unknown. Measured across T1, that shape is one of the largest single
	// causes.
	//
	// The declaration path is the only one that reaches here after
	// consuming a terminator -- parseVarDecl takes the `;` so the statement
	// owns its punctuation -- which is why the check lives here rather than
	// in each caller.
	//
	// A BLOCK closes a statement the same way, without a `;`. `defer { ... }`
	// on one line and `if (...) { ... }` on the next are two statements, and
	// perl says so -- measured on 5.42.0, they come back as siblings:
	//
	//	$ perl -MO=Deparse -e 'use feature "defer"; while(1){
	//	      defer { print "d" } if ($i == 3) { last; } }'
	//	    defer { print 'd'; }
	//	    if ($i == 3) { last; }
	//
	// Read as a modifier instead, the `if` took `($i == 3)` as its condition
	// and orphaned the `{last}` after it, emitting
	// `if (defer {...}) ($i == 3){last}` -- two paren groups and not a parse of
	// anything. `defer.t` is where this shows up, and it only became reachable
	// once `WORD BLOCK` stopped being declined.
	//
	// A block-form statement is NOT always finished, which is why the kind of
	// the closing brace is not enough on its own: `$x = sub { 1 } if $y` is a
	// genuine modifier on an assignment whose last token is also a `}`. The
	// difference is whether the BODY is a block form, and blockForm is the
	// predicate canon already uses to decide the same thing about a `;`.
	if p.pos > 0 && p.toks[p.pos-1].Kind == lexer.Semicolon {
		return nil
	}
	if endsInBlock(body) {
		return nil
	}

	word, ok := p.peekSignificant()
	if !ok || word.Kind != lexer.Word || !modifiers[p.text(word)] {
		return nil
	}
	p.advanceTo(word)

	text := p.text(word)
	kind := Loop
	if text == "if" || text == "unless" {
		kind = Conditional
	}
	n := &Node{Kind: kind, Text: text, Start: start}

	// The condition is a bare expression here, not a parenthesised one:
	// `$y = 1 if $x` has no parens and `$y = 1 if ($x)` merely has a
	// parenthesised expression as its condition.
	//
	// Children are in SOURCE order -- body then condition -- even though the
	// tree's meaning is the other way round. SourceText walks children in
	// order and emits the bytes between them, so a child that starts before
	// its predecessor ends makes it emit the same span twice: measured, the
	// statement came back as "$y = 1 if $x$y = 1 if $x;".
	//
	// Which child is the condition is recorded by Text (the modifier
	// keyword) plus position, not by ordering. Reordering the tree to match
	// evaluation would break the invariant that a node's children tile it.
	if cond := p.parseExpr(0); cond != nil {
		n.Children = append(n.Children, body, cond)
	} else {
		n.Children = append(n.Children, body)
	}
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// parseLabels reads any run of `NAME:` before a statement.
//
// The decision is POSITIONAL, not punctuational. `$h{LOOP}` and
// `LOOP ? 1 : 0` contain the same bareword-then-colon shape, and what
// separates them is that a statement can start here and cannot start there.
// This is only ever called at a statement boundary, so the question is
// already answered by where we are.
//
// What remains is telling `LABEL:` from a package-qualified name: `Foo::bar`
// has a colon too. Requiring exactly one `:` and a following non-colon
// distinguishes them, which is what perl's own lexer does.
func (p *parser) parseLabels() []*Node {
	var labels []*Node
	for {
		save := p.pos
		word, ok := p.peekSignificant()
		if !ok || word.Kind != lexer.Word {
			return labels
		}
		p.advanceTo(word)

		colon, ok := p.peekSignificant()
		if !ok || p.text(colon) != ":" {
			p.pos = save
			return labels
		}
		// `Foo::bar` lexes its separator as one token, so a `::` here is a
		// qualified name rather than a label. Checked on the text rather
		// than by peeking further, since the lexer already made the call.
		p.advanceTo(colon)
		labels = append(labels, &Node{
			Kind: Label, Text: p.text(word),
			Start: word.Start, End: p.prevEnd(),
		})
	}
}
