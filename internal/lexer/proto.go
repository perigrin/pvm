// ABOUTME: The prototype scan: a balanced (...) after `sub NAME` taken as an opaque string.
// ABOUTME: perl decides prototype-vs-signature on the feature flag, not the contents.

package lexer

import "unicode/utf8"

// scanPrototype takes a balanced `(...)` directly after `sub NAME` as one
// token, without lexing its contents.
//
// The contents are sigils that are not variables. Measured before this
// existed:
//
//	"sub f ($$) { 1 }"      Operator(()  Variable($$))
//	"sub f (\@\@) { 1 }"    Operator(\)  Variable(@\)  Variable(@))
//
// The first swallowed the closing paren into the variable, which takes the
// rest of the declaration with it.
//
// perl scans this with scan_str (toke.c:5923) -- the same routine that reads
// a quote-like body -- and spec §5.5.2 records it as "a balanced (...) as an
// opaque string".
func scanPrototype(l *lexer) bool {
	if !l.expectPrototype || l.pos >= len(l.src) || l.src[l.pos] != '(' {
		return false
	}
	// A signature, not a prototype. perl's own test is the feature and not
	// the contents (toke.c:5923, `if (*s == '(' && !is_sigsub)`), because
	// `($a, $b)` is a valid prototype spelling too -- it just means
	// something else. Measured:
	//
	//	perl -e 'sub f ($$) {1} print prototype(\&f)'              $$
	//	perl -e 'use v5.36; sub g ($a,$b) {1} print prototype(\&g)' (undef)
	if l.signatures {
		l.expectPrototype = false
		l.sigPending = true
		return false
	}

	start := l.pos
	depth := 0
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		l.pos++
		switch c {
		case '\\':
			// A backslash in a prototype quotes the next character, so it
			// cannot close the group. `(\@\@)` is two array references.
			if l.pos < len(l.src) {
				l.pos++
			}
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				l.expectPrototype = false
				l.emit(Prototype, start)
				return true
			}
		case '\n':
			// Unterminated. perl croaks "Prototype not terminated"; an LSP
			// cannot, so the cursor is put back and the parens lex as
			// ordinary tokens. Round-trip holds either way.
			l.pos = start
			l.expectPrototype = false
			return false
		}
	}
	l.pos = start
	l.expectPrototype = false
	return false
}

// noteSubName tracks whether a prototype may start at the next `(`, and
// reports whether the token just emitted was a declaration's NAME.
//
// The return value is what the expect transition needs: after `sub NAME` a
// BLOCK comes next, not a term, which is perl's PREBLOCK (toke.c:6636 for a
// sub, 8862 for a package).
//
// Only directly after the keyword -- otherwise `f ($x)` and `my ($a, $b)`
// would scan their parens as prototypes.
func (l *lexer) noteSubName(k Kind, start int) bool {
	switch k {
	case Whitespace, Comment:
		// Trivia between `sub`, the name and the `(` does not reset the
		// state: `sub  f  ($$)` is one declaration.
		return false
	case Operator:
		// An attribute's colon, which opens or continues a sub's attribute
		// list. The list sits between the name (or the `sub` itself, for an
		// anonymous one) and the body, and a BLOCK still follows it, so the
		// colon must not clear the block expectation. Measured before this,
		// `sub f :lvalue { 1 }` lexed its `{` as an anonymous hash's and the
		// whole body was lost.
		//
		// Exactly one byte, so `Foo::bar`'s `::` -- emitted as one token --
		// is a qualified name and not an attribute.
		if l.inSubAttrs && l.pos-start == 1 && l.src[start] == ':' {
			l.sawAttrColon = true
			return true
		}
		l.closeDeclHead()

	case Word:
		word := string(l.src[start:l.pos])
		// An attribute's name. `:prototype(...)`'s argument IS a prototype,
		// so it must be scanned as one opaque token: `$)` is a real perl
		// punctuation variable, and without this the closing paren is
		// swallowed into a Variable and the sub body goes with it. Measured
		// on `sub :prototype($) { 1 }`:
		//
		//	Operator(:) Word(prototype) Operator(() Variable($))
		//
		// Any other attribute's argument is not perl at all, and the parser
		// scans it by balancing brackets; only `prototype`'s contents can
		// break that, because only they can hide a `)`.
		if l.sawAttrColon {
			l.sawAttrColon = false
			l.expectPrototype = word == "prototype"
			return true
		}
		switch {
		case word == "sub" || word == "method" && l.classSyntax:
			l.sawSubWord = true
			l.sawPackageWord = false
			// An ANONYMOUS sub's prototype or signature follows the keyword
			// directly: `sub (&) { ... }`, `sub ($x) { ... }`. A named one's
			// follows its name, and the name resets this below. After `->`
			// the word is a method name and its parens are arguments.
			//
			// `sub` only: `method` is a keyword under the class feature
			// alone, which the lexer does not track, and elsewhere it is an
			// ordinary name -- op/args.t calls `method('foo', 'bar')`.
			l.expectPrototype = word == "sub" && !l.afterArrow(start)
			// An ANONYMOUS sub's attribute list starts right here, with no
			// name between: `my $f = sub :lvalue { 1 }`.
			l.inSubAttrs = true
		case word == "package" || word == "class":
			// A package or class name is also followed by a block or a
			// semicolon, but never by a prototype.
			l.sawPackageWord = true
			l.sawSubWord = false
			l.expectPrototype = false
			l.inSubAttrs = false
		case l.sawSubWord:
			// The name after `sub`. A prototype may follow it, and so may an
			// attribute list, and so may a block.
			l.sawSubWord = false
			l.noteLexSub(word)
			l.expectPrototype = true
			l.inSubAttrs = true
			return true
		case l.sawPackageWord:
			l.sawPackageWord = false
			l.sawPackageName = true
			if i := l.significantBefore(len(l.toks) - 1); i >= 0 &&
				string(l.src[l.toks[i].Start:l.toks[i].End]) == "class" {
				l.classSyntax = true
			}
			l.expectPrototype = false
			// `class Point :isa(Shape) { }`. A class's attributes take the
			// same shape, and the block after them is still a block.
			l.inSubAttrs = true
			return true
		default:
			l.closeDeclHead()
		}
	case Number:
		// `package Foo 1.0 { }`, `class Point 1.0 :isa(Shape) { }`: a
		// version after a package or class name is part of the head, and
		// the block after it is still the body.
		if l.sawPackageName {
			l.sawPackageName = false
			return true
		}
		l.closeDeclHead()
	case Prototype:
		// `sub f ($$) { ... }` -- the block still follows the prototype, and
		// so may an attribute list: `sub f () :lvalue { }`.
		l.expectPrototype = false
		l.inSubAttrs = true
		return true
	default:
		l.closeDeclHead()
	}
	return false
}

// closeDeclHead clears every declaration-shape carry at once.
//
// They are one state -- "a declaration's head is still open" -- spread over
// four booleans, and clearing three of four is the bug this exists to make
// impossible.
func (l *lexer) closeDeclHead() {
	l.sawSubWord = false
	l.sawPackageWord = false
	l.sawPackageName = false
	l.expectPrototype = false
	l.sawAttrColon = false
	l.inSubAttrs = false
}

// noteSignatureParens keeps sigDepth: the `(` a pending signature was waiting
// for opens it, parens inside a default expression nest, and the `)` that
// returns to depth 0 closes it. Any other significant token first means the
// parens never came.
func (l *lexer) noteSignatureParens(k Kind, start int) {
	if k == Whitespace || k == Comment {
		return
	}
	paren := byte(0)
	if l.pos-start == 1 && (l.src[start] == '(' || l.src[start] == ')') {
		paren = l.src[start]
	}
	if l.sigPending {
		l.sigPending = false
		if paren == '(' {
			l.sigDepth = 1
		}
		return
	}
	if l.sigDepth == 0 {
		// A signature after an attribute list: `sub f :lvalue ($x) {...}`.
		// An attribute's own argument touches its name (toke.c tests
		// `*d == '('` right after it), so a `(` inside the list that does
		// not is the signature.
		if paren == '(' && l.signatures && l.inSubAttrs && !l.touchesAttributeName(start) {
			l.sigDepth = 1
		}
		return
	}
	switch paren {
	case '(':
		l.sigDepth++
	case ')':
		l.sigDepth--
	}
}

// bareSignatureSigil reports whether the sigil at start is a signature
// element on its own -- a placeholder, `($a, $)` -- rather than the start of
// a variable name. toke.c reads signature elements with yyl_sigvar for
// exactly this, and its comment names the trap: "the general yylex code
// would otherwise try to interpret whatever follows as a var; e.g. ($, ...)
// would be seen as the var '$,'".
//
// Only at depth 1, where elements are, and only at an element's start,
// after the `(` or a `,`: a default expression is ordinary code, and `$)`
// inside one is still the effective gid.
func (l *lexer) bareSignatureSigil(start int) bool {
	if l.sigDepth != 1 || start+1 >= len(l.src) {
		return false
	}
	i := l.significantBefore(len(l.toks))
	if i < 0 || l.toks[i].End-l.toks[i].Start != 1 || !byteIn(l.src[l.toks[i].Start], "(,") {
		return false
	}
	if l.nameAfterSpace(start+1) > start+1 {
		return false // `$\n a`: the name is past the space.
	}
	r, _ := utf8.DecodeRune(l.src[start+1:])
	return !identStart(r, l.utf8Pragma)
}

// touchesAttributeName reports whether the token just before start is an
// attribute's name -- a Word after a `:` -- ending exactly there. The `sub`
// of an anonymous sub is a Word too, and `sub($x)` is its signature.
func (l *lexer) touchesAttributeName(start int) bool {
	n := len(l.toks)
	if n < 3 {
		return false
	}
	prev, colon := l.toks[n-2], l.toks[n-3]
	return prev.Kind == Word && prev.End == start &&
		colon.End-colon.Start == 1 && l.src[colon.Start] == ':'
}

// afterArrow reports whether the significant token before start is `->`.
func (l *lexer) afterArrow(start int) bool {
	i := l.significantBefore(len(l.toks) - 1)
	return i >= 0 && l.toks[i].End-l.toks[i].Start == 2 &&
		string(l.src[l.toks[i].Start:l.toks[i].End]) == "->"
}

// lexSub is a lexical sub's name and the bracket depth of its scope.
type lexSub struct {
	name  string
	depth int
}

// noteLexSub records a sub NAME just emitted if a `my`, `state` or `our`
// stands before its `sub`: a lexical sub, in scope until its block closes.
// perl lets one shadow a quote operator -- measured on 5.42.0,
// `{ my sub s { 42 } print s(1) }` prints 42 and `s/a/b/` after the block
// substitutes again.
func (l *lexer) noteLexSub(word string) {
	sub := l.significantBefore(len(l.toks) - 1)
	if sub < 0 {
		return
	}
	decl := l.significantBefore(sub)
	if decl < 0 {
		return
	}
	switch string(l.src[l.toks[decl].Start:l.toks[decl].End]) {
	case "my", "state", "our":
		l.lexSubs = append(l.lexSubs, lexSub{name: word, depth: len(l.brackets)})
	}
}

// closeLexSubScope drops the lexical subs declared inside a block that has
// just closed.
func (l *lexer) closeLexSubScope() {
	for len(l.lexSubs) > 0 && l.lexSubs[len(l.lexSubs)-1].depth > len(l.brackets) {
		l.lexSubs = l.lexSubs[:len(l.lexSubs)-1]
	}
}

// lexSubInScope reports whether a lexical sub of this name is in scope.
func (l *lexer) lexSubInScope(name string) bool {
	for _, s := range l.lexSubs {
		if s.name == name {
			return true
		}
	}
	return false
}
