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
	case Word:
		word := string(l.src[start:l.pos])
		switch {
		case word == "sub" || word == "method":
			l.sawSubWord = true
			l.sawPackageWord = false
			l.expectPrototype = false
		case word == "package" || word == "class":
			// A package or class name is also followed by a block or a
			// semicolon, but never by a prototype.
			l.sawPackageWord = true
			l.sawSubWord = false
			l.expectPrototype = false
		case l.sawSubWord:
			// The name after `sub`. A prototype may follow it, and so may a
			// block.
			l.sawSubWord = false
			l.expectPrototype = true
			return true
		case l.sawPackageWord:
			l.sawPackageWord = false
			l.expectPrototype = false
			return true
		default:
			l.expectPrototype = false
		}
	case Prototype:
		// `sub f ($$) { ... }` -- the block still follows the prototype.
		return true
	default:
		l.sawSubWord = false
		l.sawPackageWord = false
		l.expectPrototype = false
	}
	return false
}
