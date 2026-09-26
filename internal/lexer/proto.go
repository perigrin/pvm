// ABOUTME: The prototype scan: a balanced (...) after `sub NAME` taken as an opaque string.
// ABOUTME: perl decides prototype-vs-signature on the feature flag, not the contents.

package lexer

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
		case word == "sub" || word == "method":
			l.sawSubWord = true
			l.sawPackageWord = false
			l.expectPrototype = false
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
			l.expectPrototype = true
			l.inSubAttrs = true
			return true
		case l.sawPackageWord:
			l.sawPackageWord = false
			l.expectPrototype = false
			// `class Point :isa(Shape) { }`. A class's attributes take the
			// same shape, and the block after them is still a block.
			l.inSubAttrs = true
			return true
		default:
			l.closeDeclHead()
		}
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
	l.expectPrototype = false
	l.sawAttrColon = false
	l.inSubAttrs = false
}
