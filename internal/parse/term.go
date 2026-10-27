// ABOUTME: The nud side of the Pratt parser: §4.4's term forms, one per shape.
// ABOUTME: The backslash of §4.4.5 is here — it is the srefgen the fidelity harness measures.

package parse

import (
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// parseTerm parses one term: a literal, a variable, a prefix operator and its
// operand, a parenthesised list, or an anonymous constructor.
//
// Anything it cannot read becomes Unknown spanning what it consumed. The
// plan's rule applies inside expressions as much as at statement level: a
// guessed operand produces a tree that is not a parse of its source.
func (p *parser) parseTerm() *Node {
	// The power parseExpr is reading this term at belongs to this term
	// alone: a block or a parenthesised list inside it starts over.
	termBP := p.termBP
	p.termBP = 0
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > maxDepth {
		return p.refuseTooDeep(true)
	}
	tok, ok := p.peekSignificant()
	if !ok {
		// Nothing left to read. Returning an Unknown here would span no
		// bytes, and an empty node is indistinguishable from a loop that did
		// not advance -- which is the invariant the fuzzer checks. The
		// caller handles a nil term.
		return nil
	}
	text := p.text(tok)

	// A named parameter, blead's `:$name` (perly.y's `optcolon PERLY_DOLLAR
	// sigvar`, perlsub since 5.43.5): in a signature, the one `:` a term can
	// begin with. A ternary's `:` in a default is in operator position and
	// never reaches here.
	if p.inSignature && text == ":" {
		if v, ok := p.peekAfter(tok); ok && v.Kind == lexer.Variable && strings.HasPrefix(p.text(v), "$") {
			p.advanceTo(v)
			return &Node{
				Kind: Unary, Text: ":", Start: tok.Start, End: v.End,
				Children: []*Node{{Kind: Term, Text: p.text(v), Start: v.Start, End: v.End}},
			}
		}
	}

	// A filetest: `-e $f`, `-d $dir`. perl returns UNIOP for these
	// (toke.c:6255 FTST), so they bind exactly like a named unary.
	//
	// `lexer.IsFileTest` recognises the token, which arrives whole: the whole
	// `-e` is one Operator, because the minus is part of the operator's name
	// and a lexer that left a bare `-` in front of a word had not yet decided
	// which of perl's three readings of `-` it meant.
	//
	// The decision does NOT need the parser's position, which is what this
	// code assumed while it was gluing two tokens back together. Measured
	// 5.42.0, there is no subtraction reading to choose between:
	//
	//	$ perl -e 'my $x = 1 -e "/etc";'   syntax error near "1 -e "
	//
	// perl forms `-e` there too and then has nowhere to put it. So the lexer
	// decides it on the two bytes alone and this reads the result.
	//
	// THE PAREN CLIFF APPLIES, §4.8.1, and it is not decoration. `-e($f)`
	// spells the argument list with parens, and a branch that handed `($f)`
	// to parseExpr got a paren LIST as the child -- so canon wrote its own
	// parens around one the source already had, `-e(($f))`, and added another
	// on every pass. That breaks the canon fixpoint rather than merely
	// looking odd, and `open_mode_whitespace.t` is where it was caught.
	if lexer.IsFileTest(text) {
		p.advanceTo(tok)
		n := &Node{
			Kind: Call, Text: text, Resolved: true, NamedUnary: true,
			Start: tok.Start, End: tok.End,
		}
		if next, ok := p.peekSignificant(); ok && p.text(next) == "(" {
			p.advanceTo(next)
			n.ArgParen = true
			if arg := p.parseCallArgs(); arg != nil {
				n.Children = append(n.Children, arg)
			}
			if close, ok := p.peekSignificant(); ok && p.text(close) == ")" {
				p.advanceTo(close)
			}
			n.End = p.prevEnd()
			p.readHandleSlots(n, text)
			return n
		}
		// No parens: one argument at level 19, so arithmetic binds into the
		// test and comparison does not. An ABSENT argument is legal --
		// `$_="/etc"; print -e;` deparses as `print -e $_`, so the node keeps
		// the operator's own span and has no child.
		//
		// The absence needs a BRANCH, which is what this comment used to deny.
		// `parseExpr` does not return nil at a terminator: it returns an
		// Unknown spanning whatever it found, so the `;` became the argument
		// and canon emitted `print(-e(;));`. In the `if` form the whole BLOCK
		// was absorbed -- `if (-e){print 1}` canon'd as `if (-e(){print(1)}))`
		// with the parens left unbalanced, which is a wrong SHAPE rather than
		// a wrong count.
		//
		// A terminator is anything that cannot begin a term: the statement's
		// `;`, a closer (`)` of the `if`, `}` of a `grep` block, `]`), end of
		// input, or an INFIX OPERATOR. Testing the token is narrower than
		// testing the result, because an Unknown that spans a REAL operand
		// must stay in the tree.
		//
		// The infix case is not an edge: perl supplies `$_` there too, and
		// both of these are ordinary in its own suite --
		//
		//	perl -MO=Deparse -e '$_="/etc"; -e or die;'
		//	  ->  die unless -e $_;
		//	perl -MO=Deparse -e '$_="/etc"; print -e ? 1 : 0;'
		//	  ->  print -e $_ ? 1 : 0;
		//
		// A first version of this guard stopped at `;` and closers alone and
		// left both of those refusing, which is why the set is the `infix`
		// table rather than a hand-listed pair of kinds.
		next, ok := p.peekSignificant()
		if !ok || next.Kind == lexer.Semicolon || next.Kind == lexer.CloseBracket {
			return n
		}
		// `++` and `--` are in the table as postfix operators, and begin a
		// term too: measured on 5.42.0, `-f ++ $d` is `-f ++$d`.
		if _, isInfix := p.infixOp(p.text(next)); isInfix && p.text(next) != "++" && p.text(next) != "--" {
			return n
		}
		if arg := p.parseExpr(bpNamedUnary); arg != nil {
			n.Children = append(n.Children, arg)
			n.End = p.prevEnd()
		}
		p.readHandleSlots(n, text)
		return n
	}

	// Prefix operators. `\` is one of these, and §4.4.5 is not optional
	// scaffolding: it is the reference perl reports with srefgen, which is
	// the signal the whole fidelity harness was built to measure.
	if bp, isPrefix := prefix[text]; isPrefix && p.prefixAllowed(text) {
		p.advanceTo(tok)
		// `not (` is a function of the parenthesised list alone: toke.c's
		// KEY_not returns FUN1 when a `(` follows, and NOTOP, which takes
		// the whole listexpr, otherwise. Measured on 5.42.0 with
		// -MO=Deparse,-p, `(not(0), 1)` is `((!0), 1)` and `not (1) + 1`
		// is `(!1) + 1`. At bpDeref nothing after the `)` binds into it.
		if next, ok := p.peekSignificant(); ok && text == "not" && p.text(next) == "(" {
			bp = bpDeref
		}
		operand := p.operand(bp, tok)
		// The `not (` pair is not's, as a call's parens are the call's.
		if text == "not" && bp == bpDeref && operand.Paren && p.src[operand.Start] == '(' {
			takeSyntaxPair(operand)
		}
		return p.refuseAliasedOperand(&Node{
			Kind: Unary, Text: prefixName(text),
			Start: tok.Start, End: operand.End,
			Children: []*Node{operand},
		}, text, "prefix")
	}

	// A glob: `*foo`, `*{$name}`, `*$glob`. A term, not multiplication -- and
	// the parser knows which because parseTerm is only called where a term
	// is expected.
	if text == "*" {
		// A name that is not an identifier: `*1`, `*^R`, `*-`. perl accepts
		// any variable name here -- measured on 5.42.0, `local *a = *1;`,
		// `*^R = *foo;` and `*X = *-;` each deparse as written -- and the
		// lexer emits the star and the name apart, so the glob is assembled
		// here. Only TOUCHING the star: a digit, or `^` and a letter. A
		// punctuation name, `*-` or `*|`, arrives as the Word the lexer's
		// scanGlobName makes it, and is read below as `*foo` is.
		if glob := p.globWithSpecialName(tok); glob != nil {
			return glob
		}
		// A deref'd scalar names the slot too: `*$$foo`, measured on
		// 5.42.0 to deparse as written.
		if next, ok := p.peekAfter(tok); ok && next.Kind == lexer.DerefSigil {
			p.advanceTo(tok)
			operand := p.operand(bpDeref, tok)
			return &Node{
				Kind: Term, Text: "*",
				Start: tok.Start, End: operand.End,
				Children: []*Node{operand},
			}
		}
		if name, ok := p.peekAfter(tok); ok &&
			(name.Kind == lexer.Word || name.Kind == lexer.Variable ||
				p.text(name) == "{") {
			if p.text(name) == "{" {
				if n := p.bracedSubscriptedName(tok, name); n != nil {
					return n
				}
			}
			p.advanceTo(tok)
			// Braces holding statements, as `${; ...}`'s do: `*{;undef}`.
			if p.text(name) == "{" && p.derefHoldsStatements(name) {
				blk := p.parseBlock(name)
				return &Node{
					Kind: Term, Text: "*",
					Start: tok.Start, End: blk.End,
					Children: []*Node{blk},
				}
			}
			if p.text(name) == "{" {
				// The braces of `*{EXPR}` GROUP; they do not construct.
				// Calling parseTerm here read the `{` as an anonymous hash,
				// so the tree said a hash was being BUILT where a
				// symbol-table slot is being named -- a wrong tree that
				// round-trips.
				//
				// Consumed here rather than delegated, for the same reason
				// the `${EXPR}` deref does it: parseTerm cannot know the
				// brace is a group, because in every other position it is
				// not.
				p.advanceTo(name)
				inner := p.parseExpr(0)
				if close, ok := p.peekSignificant(); ok && p.text(close) == "}" {
					p.advanceTo(close)
				}
				n := &Node{
					Kind: Term, Text: "*", Start: tok.Start, End: p.prevEnd(),
				}
				if inner != nil {
					n.Children = append(n.Children, inner)
				}
				return n
			}
			// `*%x` after a declared word: the lexer took `%x` as the star's
			// operand, as an undeclared word's multiplication has it, but a
			// star read as a term is perl's glob `*%`, then `x` -- measured
			// on 5.44.0, `sub A {} A*%x3` is `A(*% x 3)`. So is `*$;`, the
			// glob `*$` before a `;`; `*$x` is the glob `$x` names.
			if name.Kind == lexer.Variable && name.Start == tok.End && (p.src[name.Start] != '$' || name.End-name.Start == 2 &&
				!isWordByte(p.src[name.Start+1]) && !strings.ContainsRune("${:", rune(p.src[name.Start+1]))) {
				return p.ambiguousSigil(tok.Start)
			}
			p.advanceTo(name)
			if name.Kind == lexer.Variable {
				// `*$glob` names the slot the SCALAR points at, so the
				// variable is an operand and has to be in the tree. Folding
				// it into Text as `*NAME` does would put `$glob` back out of
				// it -- the same leaf this issue exists to stop making.
				//
				//	$ perl -MO=Deparse -e 'my $g = \*STDOUT; my $x = *$g{IO};'
				//	my $x = *$g{'IO'};
				return &Node{
					Kind: Term, Text: "*",
					Start: tok.Start, End: name.End,
					Children: []*Node{{
						Kind: Term, Text: p.text(name),
						Start: name.Start, End: name.End,
					}},
				}
			}
			return &Node{
				Kind: Term, Text: "*" + p.text(name),
				Start: tok.Start, End: name.End,
			}
		}
	}

	switch {
	case text == "(":
		n := p.parseParenList(tok)
		// `->(`: the pair is the call's argument list.
		if n.Paren && p.afterArrow(n) {
			takeSyntaxPair(n)
		}
		return n
	case text == "[":
		return p.parseBracketed(tok, "]", AnonArray)
	case text == "{":
		// In term position a `{` is an anonymous hash -- one of the five
		// oracle markers. Whether it is instead a block is the statement
		// level's decision (§4.9.2), and the lexer's brace stack already
		// classified this one as a term.
		return p.parseBracketed(tok, "}", AnonHash)
	}

	switch tok.Kind {
	case lexer.DerefSigil:
		// A sigil applied to an expression: `${$h->{k}}`, `@{[ 1, 2 ]}`,
		// `$$x`. The lexer emits the sigil alone precisely so the interior
		// reaches the parser; making a leaf of it here would put the names
		// inside back out of the tree.
		//
		// A Unary rather than a new kind: a dereference IS a prefix operator
		// applied to one operand, and the sigil is in Text. §4.14 names it
		// `Deref{Sigil, Expr}`, which is this shape with the kind spelled
		// out; the lowering (chapter 6 §6.1.6) is where that rename belongs,
		// because a new CST kind is a change every consumer must learn.
		p.advanceTo(tok)

		// The braces of `${EXPR}` GROUP; they do not construct. Letting
		// parseTerm see the `{` made an AnonHash of it, so the tree said a
		// hash was being built around the very expression being
		// dereferenced.
		if open, ok := p.peekSignificant(); ok && p.text(open) == "{" {
			if n := p.bracedSubscriptedName(tok, open); n != nil {
				return n
			}
			// The braces hold a BLOCK in perly.y (`indirob: ... | block`),
			// and almost always one expression. A `;` -- first, or after
			// that expression -- says there are statements: `${; do { ... }
			// }`, `${no strict; \$_}`. Measured on 5.42.0, Deparse keeps
			// both as blocks. They are read as one then, and canon's Block
			// writes the braces the sigil needs.
			if p.derefHoldsStatements(open) {
				blk := p.parseBlock(open)
				return &Node{
					Kind: Unary, Text: text,
					Start: tok.Start, End: blk.End,
					Children: []*Node{blk},
				}
			}
			p.advanceTo(open)
			inner := p.parseExpr(0)
			if close, ok := p.peekSignificant(); ok && p.text(close) == "}" {
				p.advanceTo(close)
			}
			if inner == nil {
				// `${}` -- nothing to dereference. The bytes stay in the
				// tree as an Unknown rather than a Unary with no operand.
				return &Node{Kind: Unknown, Refusal: EmptyDeref, Start: tok.Start, End: p.prevEnd()}
			}
			return &Node{
				Kind: Unary, Text: text, DerefBraces: true,
				Start: tok.Start, End: p.prevEnd(),
				Children: []*Node{inner},
			}
		}

		operand := p.operand(bpDeref, tok)
		return &Node{
			Kind: Unary, Text: text,
			Start: tok.Start, End: operand.End,
			Children: []*Node{operand},
		}

	case lexer.FuncSigil:
		// `&` in term position introduces a function name, and the lexer
		// emits it as its own token so that `&f` and `&&` stay
		// distinguishable. The name is the next token, so the two are joined
		// here into one term: `\&f` is a code reference, not a reference to
		// an ampersand.
		p.advanceTo(tok)
		// Empty braces are no code to call: perl 5.44.0 has `(&{},)` a
		// syntax error near "&{", and read as a subscript on a bare `&` its
		// canon re-read differently.
		if open, ok := p.peekSignificant(); ok && p.text(open) == "{" {
			if close, ok := p.peekAfter(open); ok && p.text(close) == "}" {
				return p.refuseNotATerm(tok)
			}
		}
		// Braces holding statements, as `${; ...}`'s do: `&{; $c }(1)`,
		// measured on 5.42.0 to deparse as `&{$c;}(1)`.
		if open, ok := p.peekSignificant(); ok && p.text(open) == "{" && p.derefHoldsStatements(open) {
			blk := p.parseBlock(open)
			return &Node{
				Kind: Unary, Text: text,
				Start: tok.Start, End: blk.End,
				Children: []*Node{blk},
			}
		}
		if name, ok := p.peekSignificant(); ok && p.ampTakes(name) {
			p.advanceTo(name)
			return &Node{
				Kind: Term, Text: text + p.text(name),
				Start: tok.Start, End: name.End,
			}
		}
		// perl has no bare `&`: toke.c's scan_ident takes one punctuation
		// character after it as the name, the way `$;` and `$=` are names.
		// Measured on 5.42.0, `&=;`, `&,;` and `&);` each compile as a call,
		// while `&= 2;`, `f(1, &, 2);` and `f(&);` are each a syntax error
		// -- the name leaves a term after a term, or a `(` unclosed.
		name, ok := p.peekSignificant()
		if !ok {
			// `&` at end of file: "syntax error at EOF".
			return p.refuseNotATerm(tok)
		}
		if p.ampTakesPunct(name) {
			p.advanceTo(name)
			return &Node{
				Kind: Term, Text: text + p.text(name),
				Start: tok.Start, End: name.End,
			}
		}
		// Digits after it that are a Number are a name perl refuses: `&00` is "Numeric
		// variables with more than one digit may not start with '0'" on
		// 5.44.0. Read as a bare `&`, `no n!&00`'s canon `&;00` re-read
		// differently.
		if name.Kind == lexer.Number {
			return p.refuseNotATerm(tok)
		}
		return &Node{Kind: Term, Text: text, Start: tok.Start, End: tok.End}

	case lexer.Word:
		text := p.text(tok)
		if d := p.parsePrefixedAnonSub(tok); d != nil {
			return d
		}
		if d := p.parseDeclaredExpression(tok); d != nil {
			return d
		}
		switch {
		case declarators[keywordName(text)] && p.declaratorTakesTarget(tok):
			// A declaration in EXPRESSION position: `open my $fh, $p`,
			// `f(my $x)`. perly.y makes `my` a named unary at level 19
			// (§4.6), so it is a term here as much as a statement form.
			//
			// 50 of T1's files first fail on `open my $fh, ...`, which is
			// the idiomatic three-argument open and appears in almost every
			// file that touches a filehandle.
			if termBP > infix["="].BP {
				return p.parseVarDeclOperand(tok, termBP)
			}
			return p.parseVarDeclNoSemi(tok)

		case (keywordName(text) == "my" || keywordName(text) == "our") && !p.declaratorIsBareword(tok):
			// `my` and `our` are always keywords, and with no target they
			// are no term: measured on 5.42.0, `!my=0`, `f(my)` and `$x =
			// our;` are each a syntax error. Read as a call, canon wrote
			// `my()`, which re-read as a declaration.
			return p.refuseNotATerm(tok)

		case wordInfixOps[keywordName(text)] && !p.wordOpIsString(tok):
			// Nor is a word operator: measured on 5.44.0, `gt*2;`, `print
			// -eq;` and `{ge}` are each a syntax error, while `$h{eq}` and
			// `eq => 1` compile. Read as a call, `gt*&;`'s canon re-read
			// differently.
			return p.refuseNotATerm(tok)

		case text == "sub":
			// An anonymous sub: `sub { ... }` with no name. Distinguished
			// from a declaration by what follows -- a `{` rather than a
			// name -- which is the same test perl makes.
			//
			// An ATTRIBUTE may stand between: `my $f = sub :lvalue { 1 }`
			// has no name either, and a `:` here cannot be anything else --
			// `sub` is a keyword, so the colon is not a label's and not a
			// ternary's.
			//
			// So may a prototype or signature: `sub (&) { ... }`, `sub ($x)
			// { ... }`.
			next, ok := p.peekAfter(tok)
			if ok && (p.text(next) == "{" || p.text(next) == ":" ||
				next.Kind == lexer.Prototype || p.text(next) == "(") {
				return p.parseAnonSub(tok)
			}
			// Nor a sub with no head and no block: measured on 5.42.0, `my
			// $f = sub;`, `f(sub, 1)` and `print sub;` are each "Illegal
			// declaration of anonymous subroutine". Before `=>` it is the
			// string 'sub', before `}` a hash key, `$h{sub}`, and a name
			// after it is parseWordTerm's. A `}` that closes no subscript
			// is no key's: `{sub}` and `do {sub}` are refused as perl
			// refuses them, and FuzzParser's `{sub}` canon'd as `{sub;}`.
			if !ok || next.Kind != lexer.Word && p.text(next) != "=>" && !(p.inHashKey && p.text(next) == "}") {
				return p.refuseNotATerm(tok)
			}

		case text == "eval" || text == "do":
			// `eval BLOCK` and `do BLOCK` take a block, not an expression.
			// The named-unary path would parse the `{` as an anonymous hash.
			if next, ok := p.peekAfter(tok); ok && p.text(next) == "{" {
				return p.parseBlockOperator(tok)
			}
		}
		return p.markReferenceSlots(p.refuseArity(p.refuseRefScalarSlot(p.markResolution(p.parseWordTerm(tok)))))

	case lexer.Variable, lexer.Number, lexer.Quote,
		lexer.Readline, lexer.HeredocOpen:
		p.advanceTo(tok)
		return &Node{
			Kind: Term, Text: text,
			Start: tok.Start, End: tok.End,
		}
	}

	return p.refuseNotATerm(tok)
}

// refuseNotATerm declines a term that begins at tok and cannot be one. Consumed
// so the loop advances; the bytes stay in the tree, through whatever the
// caller already read past tok -- an anonymous sub's head, for one.
//
// The refusal is a catch-all, so it is an Error node (recovery.go).
func (p *parser) refuseNotATerm(tok lexer.Token) *Node {
	p.advanceTo(tok)
	return p.missingIdentifier(p.recover(&Node{Kind: Error, Refusal: NotATerm, Start: tok.Start, End: p.prevEnd()}), tok)
}

// wordInfixOps are the word operators that are never a term: `x` is a
// bareword where a term is read (`x*2` is `'x' * 2` on 5.44.0), and `isa`
// one without its feature.
var wordInfixOps = map[string]bool{
	"eq": true, "ne": true, "lt": true, "gt": true, "le": true, "ge": true,
	"cmp": true, "and": true, "or": true, "xor": true,
}

// wordOpIsString reports whether a word operator read as a term is the
// string perl makes of it: before `=>`, or alone in a hash subscript.
func (p *parser) wordOpIsString(word lexer.Token) bool {
	next, ok := p.peekAfter(word)
	return ok && (p.text(next) == "=>" || p.inHashKey && p.text(next) == "}")
}

// declaratorIsBareword reports whether a `my` or `our` with no variable
// after it is still a word perl reads: a class name before the variable
// (`my Foo $x`), a hash key (`$h{my}`), or the left of a fat comma (`my =>
// 1`). Measured on 5.42.0, the last two compile.
func (p *parser) declaratorIsBareword(word lexer.Token) bool {
	next, ok := p.peekAfter(word)
	return ok && (next.Kind == lexer.Word || p.text(next) == "=>" || p.text(next) == "}")
}

// declaratorTakesTarget reports whether a declarator word is introducing a
// variable rather than being used as an identifier.
//
// `my`, `our`, `state` and `field` are ordinary barewords in most positions:
// `$h{field}` is a hash key, `f(state => 1)` is a fat-comma pair. Treating
// every occurrence as a declaration made `$h{field}` parse its key as one --
// caught by TestClassSyntax, which asserts exactly that.
//
// A declarator takes a target when a VARIABLE or a `(` follows it. Nothing
// else is a declaration: `my $x`, `my @a`, `my ($a, $b)`, and that is the
// whole grammar (§4.4.3).
func (p *parser) declaratorTakesTarget(word lexer.Token) bool {
	next, ok := p.peekAfter(word)
	if !ok {
		return false
	}
	return next.Kind == lexer.Variable || p.text(next) == "("
}

// bracedSubscriptedName reads `${ foo[0] }`, `@{ foo {"a"} }`, `%{ foo{a} }`
// and `*{ foo{a} }` as toke.c's scan_ident does: an identifier other than
// `sub`, then `[` or `{`, inside the braces after a sigil is the variable's
// NAME and the subscripts are its own, `$foo[0]`, `@foo{"a"}`. The closing
// brace is a fake one perl drops. A sub of that name and its prototype play
// no part: measured on 5.44.0, `@{ foo { "a" } }` deparses as `@foo{'a'}`
// with `sub foo (&)`, `sub foo ($)`, `use strict` or no sub at all, and `@{
// sub {1} }` as a deref of an anonymous sub. Nor is a word with a leading
// `::` a name: `@{ ::foo {"a"} }` is rv2av over entersub, with no rv2hv, and
// `@{ ::foo [1] }` a syntax error. The outermost subscript has
// NameBraces set, so canon writes the braces back. nil, with the cursor
// unmoved, when the braces hold anything else.
//
// ponytail: only subscripts running to the closing brace. perl drops that
// brace wherever it falls, so `@{ foo {"a"}; 1 }` is `@foo{'a'}; 1` and `@{
// foo {"a"} + 1 }` is `@foo{'a'} + 1`; psc's tree for those is wrong. So is
// its tree for a quote word, `@{ q {a} }`, which perl reads as the slice
// `@q{'a'}` (likewise qq, qw, m, s, y and tr) and the lexer takes as a
// quote. Issue 01a16e4f-b570-7182-8094-71c56af8c419 tracks both.
func (p *parser) bracedSubscriptedName(sigil, open lexer.Token) *Node {
	text := p.text(sigil)
	word, ok := p.peekAfter(open)
	if !ok || text != "$" && text != "@" && text != "%" && text != "*" || word.Kind != lexer.Word {
		return nil
	}
	name := p.text(word)
	if name == "sub" || strings.HasPrefix(name, "::") || !isIdentifier(strings.ReplaceAll(name, "::", "_")) {
		return nil
	}
	if sub, ok := p.peekAfter(word); !ok || p.text(sub) != "[" && p.text(sub) != "{" {
		return nil
	}
	save := p.pos
	p.advanceTo(word)
	n := &Node{Kind: Term, Text: text + name, Start: sigil.Start, End: word.End}
	for {
		sub, ok := p.peekSignificant()
		if !ok || p.text(sub) != "[" && p.text(sub) != "{" {
			break
		}
		n = p.parseSubscript(n, p.text(sub))
	}
	shut, ok := p.peekSignificant()
	if !ok || p.text(shut) != "}" {
		p.pos = save
		return nil
	}
	p.advanceTo(shut)
	n.End = shut.End
	n.NameBraces = true
	return n
}

// derefHoldsStatements reports whether the brace at open, after a deref
// sigil, holds statements rather than one expression: anything but one
// expression running to the closing `}`. A statement form such as `use
// strict` is no expression at all, and its refusal skips past the `;`, so
// the test is what the expression reaches rather than whether a `;` follows
// it. The cursor is left where it was.
func (p *parser) derefHoldsStatements(open lexer.Token) bool {
	// The trial parse below holds a nested deref's own trial, so unclosed
	// `${${${...` doubled per level: FuzzParser hung on 19 of them. The
	// answers are kept for the outermost call's duration.
	if held, ok := p.derefMemo[open.Start]; ok {
		return held
	}
	if p.derefMemo == nil {
		p.derefMemo = map[int]bool{}
		defer func() { p.derefMemo = nil }()
	}
	held := p.derefHoldsStatementsOnce(open)
	p.derefMemo[open.Start] = held
	return held
}

func (p *parser) derefHoldsStatementsOnce(open lexer.Token) bool {
	save := p.pos
	defer func() { p.pos = save }()
	p.advanceTo(open)
	first, ok := p.peekSignificant()
	if !ok || p.text(first) == "}" {
		return false // `${}`: EmptyDeref's, below.
	}
	// So does the yada, which is a statement and never an expression
	// (perly.y:759): `${...}++` dies Unimplemented, measured on 5.42.0.
	if first.Kind == lexer.Semicolon || p.text(first) == "..." {
		return true
	}
	p.parseExpr(0)
	tok, ok := p.peekSignificant()
	return !ok || p.text(tok) != "}"
}

// parseAnonSub: `sub { ... }` and `sub ($x) { ... }` with no name.
//
// A Declaration with no name child, so consumers that walk declarations see
// it without a second kind to learn. Whether a sub is named is a question
// about its children, not about what kind of node it is.
func (p *parser) parseAnonSub(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	// A prototype or signature, if the lexer found one.
	if proto, ok := p.peekSignificant(); ok && proto.Kind == lexer.Prototype {
		p.advanceTo(proto)
		n.Children = append(n.Children, &Node{
			Kind: PrototypeNode, Text: p.text(proto),
			Start: proto.Start, End: proto.End,
		})
	} else if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
		// A signature: the lexer leaves the parens as code when the feature
		// is on, as it does after a named sub.
		p.parseSignature(n)
	}

	// Attributes: `my $f = sub :lvalue { 1 }`. Same list, same position, and
	// the same reason parseSubDecl reads them -- an unread attribute ends the
	// sub at the colon and leaves the whole body trailing.
	p.parseAttributes(n)

	// A signature after the attributes, the order the feature requires:
	// `sub :lvalue ($x) { $x }`. See parseSubDecl.
	if !hasHead(n) {
		if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
			p.parseSignature(n)
		}
	}

	// The body. A head with no block after it is no sub: measured on 5.42.0,
	// `sub ($$);` and `my $f = sub ();` are each "Illegal declaration of
	// anonymous subroutine", and `my $f = sub :lvalue;` and a signature with
	// no block are syntax errors, so the refusal takes the head with it.
	blk := p.parseBlockOrDecline()
	if blk == nil {
		return p.refuseNotATerm(word)
	}
	n.Children = append(n.Children, blk)
	n.End = p.prevEnd()
	return n
}

// parseBlockOperator: `eval BLOCK` and `do BLOCK`, spec §4.7.
//
// These take a BLOCK where an expression would be, so the named-unary path
// cannot handle them -- it would parse the `{` as an anonymous hash, which is
// what `eval { $x; };` did before this.
//
// `eval EXPR` and `do EXPR` are different operators with the same spelling,
// and they go through parseWordTerm as ordinary named unaries. The `{` is
// what separates them, which is also how perl decides.
func (p *parser) parseBlockOperator(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Call, Text: p.text(word), Resolved: true, Start: word.Start}

	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}
	n.End = p.prevEnd()
	return n
}

// prefixAllowed keeps an infix operator from being read as a prefix one.
//
// `-` and `+` are in both tables, and which they are depends on position:
// after a term they are ADDOP, at the start of one they are UMINUS. The
// parser knows its own position, so it does not need the lexer's expect
// state for this -- parseTerm is only ever called where a term is expected.
func (p *parser) prefixAllowed(text string) bool {
	switch text {
	case "-", "+", "\\", "!", "~", "~.", "++", "--", "not":
		return true
	}
	return false
}

// globWithSpecialName reads `*1` or `*^R` at the star tok, or returns
// nil. See the glob branch of parseTerm.
func (p *parser) globWithSpecialName(star lexer.Token) *Node {
	name, ok := p.peekAfter(star)
	if !ok || name.Start != star.End {
		return nil
	}
	// last is the final token of the name: the name itself, or the letter
	// after a caret.
	var last lexer.Token
	switch text := p.text(name); {
	case name.Kind == lexer.Number && isAllDigits(text):
		last = name
	case text == "^":
		letter, ok := p.peekAfter(name)
		if !ok || letter.Start != name.End || letter.Kind != lexer.Word {
			return nil
		}
		last = letter
	default:
		return nil
	}
	p.advanceTo(last)
	return &Node{
		Kind: Term, Text: string(p.src[star.Start:last.End]),
		Start: star.Start, End: last.End,
	}
}

// isAllDigits reports whether s is one or more ASCII digits.
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// prefixName distinguishes the unary spelling from the infix one in the
// tree, so `- $a` and `$a - $b` do not both read as "-".
func prefixName(text string) string {
	switch text {
	case "-":
		return "neg"
	case "+":
		return "pos"
	case "\\":
		return "ref"
	}
	return text
}

// parseParenList: `(...)`. A single expression in parens is that expression;
// a comma-separated one is a List.
func (p *parser) parseParenList(open lexer.Token) *Node {
	p.advanceTo(open)

	// seps are the separators since the last item. More than one, or any
	// before the closer, are kept on that item for canon; see keepSeps.
	var items []*Node
	var seps []string
	for {
		tok, ok := p.peekSignificant()
		if !ok {
			break
		}
		if p.text(tok) == ")" {
			keepSeps(items, seps, 1)
			p.advanceTo(tok)
			break
		}
		// An empty slot: a comma where an element would start. perl drops it,
		// measured on 5.42.0 -- `my @a = (1,,,2)` has 2 elements -- and
		// op/for-many.t spells forty in one variable list. Only AFTER an
		// element: a leading `(,1)` is a syntax error in perl, so that refusal
		// stands.
		//
		// A fat comma in that slot is empty the same way -- `(foo, => 1)` is
		// `(foo(), 1)`, measured -- and it quotes nothing: the word it could
		// have quoted is behind the comma.
		if (p.text(tok) == "," || p.text(tok) == "=>") && len(items) > 0 {
			p.advanceTo(tok)
			seps = append(seps, p.text(tok))
			continue
		}
		// Parsed above the comma so each element is its own node. A nil
		// element means the input ran out mid-list, which the loop's next
		// peek handles.
		if item := p.parseExpr(infix[","].BP); item != nil {
			keepSeps(items, seps, 2)
			seps = nil
			items = append(items, item)
		}

		next, ok := p.peekSignificant()
		if !ok {
			break
		}
		switch p.text(next) {
		case ",", "=>":
			// A fat comma quotes the word to its LEFT (§4.5.4), so the flag
			// belongs to the element already read rather than the next one.
			// Without it `(a => 1)` and `(a, 1)` are the same tree, and the
			// lowering cannot tell the string "a" from a call to a().
			if p.text(next) == "=>" && len(items) > 0 {
				items[len(items)-1].Fat = true
				markFatOperand(items[len(items)-1])
			}
			p.advanceTo(next)
			seps = []string{p.text(next)}
		case ")":
			p.advanceTo(next)
			return p.finishList(items, open)
		default:
			// Neither a separator nor the closer. The three word operators
			// below the comma get here: `and`, `or` and `xor` are levels 4
			// and 5, the comma is 8, so the element loop above -- which
			// parses at the comma's power -- stops without consuming them
			// and every `($a and $b)` refused with `not_a_term`.
			//
			// perly.y puts them above the comma, so the paren holds a full
			// `expr` and the list assembled so far is that expr's LEFT
			// operand. Measured on perl 5.42.0:
			//
			//	$ perl -MO=Deparse -e 'my @x = (1, 2 and 3);'
			//	my(@x) = ('???', 2) && 3;
			//
			// So the loop is resumed with the list in hand rather than
			// restarted, and the closer is then taken by the code below.
			//
			// The left operand is the elements WITHOUT the paren: the paren
			// wraps the whole expr, and is recorded once, on it. Wrapped by
			// finishList here it was a second Paren node spanning `(%h` with
			// no closer, and canon wrote `(%h and 1;`.
			if left := p.commaList(items); left != nil && p.atOperatorBelowComma() {
				items = []*Node{p.parseInfix(left, 0)}
				if c, ok := p.peekSignificant(); ok && p.text(c) == ")" {
					p.advanceTo(c)
				}
				return p.finishList(items, open)
			}
			// Anything else: stop rather than spin.
			return p.finishList(items, open)
		}
	}
	return p.finishList(items, open)
}

func (p *parser) finishList(items []*Node, open lexer.Token) *Node {
	if len(items) == 1 {
		// `($x)` is `$x`, but the span covers the parens so round-trip holds
		// -- and the PAREN itself is recorded, because two of Perl's rules
		// turn on it. §4.10: `("a") x 3` gives three elements where
		// `"a" x 3` gives one. §4.12.2: with `sub f {(1,2,3)}`,
		// `my ($x) = f()` is 1 and `my $y = f()` is 3. Both measured on
		// perl 5.42.0.
		//
		// A copy rather than the node itself: the span must widen to cover
		// the parens, so this cannot alias. Every flag is carried across,
		// by copying the whole value -- dropping one here would erase an
		// Arrow on `($h->{k})`, which is the bug this function once had for
		// Paren, and an Indirect on `(method $obj ())` in t/op/method.t:79.
		c := *items[0]
		if !c.Paren {
			c.InnerStart, c.InnerEnd = c.Start, c.End
		}
		// The inner pair is a grouping of the same node even with a
		// separator after it: measured on perl 5.44.0, `((0),)[0]` is
		// `const[IV 0] sP` and `(0,)[0]` is `const[IV 0] s`.
		c.nestedParen = c.Paren
		c.Start, c.End = open.Start, p.prevEnd()
		c.Paren = true
		// A leaf is written as its source text, which now holds any
		// trailing comma inside the parens: `(0,)`. Kept as well, canon
		// wrote it twice. A fat comma's word is written the same way when
		// parenthesised: see canon's Call case.
		if len(c.Children) == 0 && (c.Kind == Term || c.Kind == Call && c.Fat) {
			c.TrailingComma = ""
		}
		return &c
	}
	return &Node{
		Kind: List, Start: open.Start, End: p.prevEnd(),
		Children: items,
		Paren:    true,
	}
}

// takeSyntaxPair gives n's outermost pair, which finishList recorded as
// Paren, to the syntax that requires it: an arrow call's arguments, `not
// (`'s operand, or a list slice's list. n keeps Paren only for a grouping
// pair nested directly inside, `$f->(($x))`. See Node.SyntaxParen.
func takeSyntaxPair(n *Node) {
	n.SyntaxParen = true
	n.Paren = n.nestedParen
}

// commaList is the element list as ONE node, for a caller that must hand it
// to parseInfix as a left operand.
//
// One element is itself: `[$a and $b]` has a single `$a` on the operator's
// left, and Concise shows no list op there at all. Several become a List,
// which is the same node parseParenList's finishList builds -- but without
// its Paren flag, because a bracket is not a paren and the two rules that
// turn on the flag (§4.10's `("a") x 3`, §4.12.2's list-vs-scalar context)
// do not apply to a constructor.
//
// An empty list cannot reach parseInfix as a left operand, so the caller is
// told to leave the operator alone -- `[and $b]` is a syntax error in perl and
// inventing an operand for it would be a worse tree than declining.
func (p *parser) commaList(items []*Node) *Node {
	if len(items) == 0 {
		return nil
	}
	if len(items) == 1 {
		return items[0]
	}
	return &Node{
		Kind: List, Start: items[0].Start, End: p.prevEnd(),
		Children: items,
	}
}

// parseBracketed: `[...]` and `{...}` in term position.
func (p *parser) parseBracketed(open lexer.Token, closer string, kind Kind) *Node {
	p.advanceTo(open)

	// seps: see parseParenList.
	var items []*Node
	var seps []string
	for {
		tok, ok := p.peekSignificant()
		if !ok {
			break
		}
		if p.text(tok) == closer {
			keepSeps(items, seps, 1)
			p.advanceTo(tok)
			break
		}
		// An empty slot, as in parseParenList: `["x",, "y"]` has two
		// elements, measured on 5.42.0. Only after an element.
		if (p.text(tok) == "," || p.text(tok) == "=>") && len(items) > 0 {
			p.advanceTo(tok)
			seps = append(seps, p.text(tok))
			continue
		}
		if item := p.parseExpr(infix[","].BP); item != nil {
			keepSeps(items, seps, 2)
			seps = nil
			items = append(items, item)
		}

		next, ok := p.peekSignificant()
		if !ok {
			break
		}
		switch p.text(next) {
		case ",", "=>":
			// Same rule inside `[...]` and `{...}`: the fat comma marks the
			// element to its left. An anon hash written `{a => 1}` and one
			// written `{a, 1}` mean the same thing to perl but not to a
			// consumer reading the key, which §4.5.4 says is a string in the
			// first and a call in the second.
			if p.text(next) == "=>" && len(items) > 0 {
				items[len(items)-1].Fat = true
				markFatOperand(items[len(items)-1])
			}
			p.advanceTo(next)
			seps = []string{p.text(next)}
		case closer:
			p.advanceTo(next)
			return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
		default:
			// The same resume parseParenList's default arm describes, and for
			// the same reason: a bracket holds a full `expr`, not a comma
			// list, so the three word operators below the comma stop the
			// element loop and take what it built as their LEFT operand.
			// Measured on perl 5.42.0:
			//
			//	$ perl -MO=Deparse -e 'my $y = [1, 2 and 3];'
			//	my $y = [('???', 2) && 3];
			//
			// Without it the operand ESCAPED the bracket -- `[$a and $b]`
			// gave `(and (my $y [$a]) $b)`, a tree in which `and` sits above
			// the declaration with a one-element arrayref on its left. The
			// source has both operands inside the brackets, so that was not
			// a parse of it.
			//
			// The left operand is the ELEMENT LIST, not a nested constructor:
			// Concise shows one `anonlist` for the whole `[...]` with the
			// `and` inside it, over a plain `list` op.
			//
			//	9  <@> anonlist sK*/1
			//	7     <|> and(other->8) lK/1
			//	6        <@> list sK
			//	8        <$> const[IV 3] s
			//
			// So the resume hands parseInfix the items assembled so far and
			// the single result becomes the constructor's only element.
			if left := p.commaList(items); left != nil && p.atOperatorBelowComma() {
				items = []*Node{p.parseInfix(left, 0)}
				if c, ok := p.peekSignificant(); ok && p.text(c) == closer {
					p.advanceTo(c)
				}
			}
			return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
		}
	}
	return &Node{Kind: kind, Start: open.Start, End: p.prevEnd(), Children: items}
}

// ampTakes reports whether `&` joins this token as the name of the sub it
// calls.
//
// A Word always: `&f`. A SCALAR variable too, because `&$coderef` is perl's
// older calling convention and is how perl's own suite spells a call that
// passes the caller's `@_` implicitly. Measured on 5.42.0, inside a sub called
// with (1,2):
//
//	&$s;     the callee sees @_ as (1,2)
//	&$s();   the callee sees @_ empty
//
// The other sigils and the subscripted forms are REFUSED, because perl refuses
// them and agreement is the point:
//
//	&@a;      Bareword found where operator expected
//	&$s[0];   syntax error at -e line 1, near "$s["
//	&$s{k};   syntax error at -e line 1, near "$s{k"
//
// Joining a Variable unconditionally made all three parse clean, which is a
// wrong tree rather than a wider one -- so the arity of the check is the fix,
// not an ornament on it.
func (p *parser) ampTakes(name lexer.Token) bool {
	if name.Kind == lexer.Word {
		return true
	}
	if name.Kind != lexer.Variable {
		return false
	}
	if text := p.text(name); text == "" || text[0] != '$' {
		return false
	}
	// A subscript after the scalar has no call reading in perl, so the `&`
	// must not swallow the name and leave the bracket stranded.
	if next, ok := p.peekAfter(name); ok {
		if t := p.text(next); t == "[" || t == "{" {
			return false
		}
	}
	return true
}

// ampTakesPunct reports whether `&` takes this token as a punctuation name:
// one character of an operator, a `;` or a closing bracket. A `{` is not one
// -- `&{$c}` is a call through a block -- and a token of two or more
// characters is not either, since perl would take only its first.
func (p *parser) ampTakesPunct(name lexer.Token) bool {
	switch name.Kind {
	case lexer.Operator, lexer.Semicolon, lexer.CloseBracket:
	default:
		return false
	}
	text := p.text(name)
	return len(text) == 1 && text != "{"
}

// keepSeps keeps, on the last item, the separators written after it when
// there are at least min of them: 1 before a closer, where any separator is
// trailing, and 2 before another item, where one is the item separator.
// Kept, they are all of what canon writes after that item. See
// Node.TrailingComma and emitCommaSeparated.
func keepSeps(items []*Node, seps []string, min int) {
	if len(items) == 0 || len(seps) < min {
		return
	}
	items[len(items)-1].TrailingComma = strings.Join(seps, " ")
}

// markTrailing records a separator with no element after it on the element
// before it, for canon to write back. See Node.TrailingComma.
func markTrailing(n *Node, sep string) {
	if n.TrailingComma == "" {
		n.TrailingComma = sep
		return
	}
	n.TrailingComma += " " + sep
}
