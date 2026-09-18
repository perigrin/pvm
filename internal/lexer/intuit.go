// ABOUTME: intuit_curly: whether a `{` after a list operator opens a block or an anonymous hash.
// ABOUTME: The one brace whose meaning depends on what is inside it rather than on what precedes it.

package lexer

// intuitCurly reports whether the `{` at the next non-space byte opens a
// BLOCK. It is perl's own heuristic, toke.c's intuit_curly (6698-6842).
//
// Everywhere else a brace is decided by what came before it. After `map`,
// `grep` and `sort` both readings are grammatical, so perl peeks PAST the
// brace at the first thing inside -- measured on perl 5.42.0:
//
//	map { ; a => 1 } (1,2)    map({'a', 1;} 1, 2)      a BLOCK
//	map { a => 1 }, (1,2)     map({'a', 1}, (1, 2))    a HASHREF
//	map { $_ => 1 } @a        map({$_, 1;} @a)         a BLOCK
//
// A word or string followed by `,` or `=>` is a hashref; everything else is a
// block -- a leading `;`, an operator, and a VARIABLE before a fat comma,
// which is why the third line is a block and not a hash. That is the same
// rule the parser's braceOpensAnonHash applies at statement start, and the
// agreement is not a coincidence: perl runs this one function for both.
//
// Byte scanning rather than tokenizing, for peekIsOpenBrace's reason: the
// lexer stays single-pass, and a lookahead that lexed would have to undo its
// own state changes.
func (l *lexer) intuitCurly() bool {
	i := skipSpaceFrom(l.src, l.pos)
	if i >= len(l.src) || l.src[i] != '{' {
		return false
	}
	i = skipSpaceFrom(l.src, i+1)
	if i >= len(l.src) {
		return true
	}

	// The first thing in the brace. A quoted string is a key candidate, and
	// so is a bareword. Anything else -- a sigil, a `;`, an operator -- ends
	// it here: that is a block.
	var end int
	switch c := l.src[i]; {
	case c == '\'' || c == '"':
		end = skipQuotedFrom(l.src, i)
		if end < 0 {
			return true
		}
	case isWordStart(c):
		for end = i; end < len(l.src) && isWordByte(l.src[end]); end++ {
		}
	default:
		return true
	}

	// The separator after it. Only `,` and `=>` make a hash.
	end = skipSpaceFrom(l.src, end)
	if end >= len(l.src) {
		return true
	}
	if l.src[end] == ',' {
		return false
	}
	if l.src[end] == '=' && end+1 < len(l.src) && l.src[end+1] == '>' {
		return false
	}
	return true
}

// skipSpaceFrom returns the index of the first non-space byte at or after i.
func skipSpaceFrom(src []byte, i int) int {
	for ; i < len(src); i++ {
		switch src[i] {
		case ' ', '\t', '\n', '\r':
		default:
			return i
		}
	}
	return i
}

// skipQuotedFrom returns the index just past the string starting at i, or -1
// if it is unterminated.
//
// The two plain quotes only, which is all a key can be here. A `q{...}` as
// the first thing inside a `map {` would mean reading a delimiter to decide a
// delimiter, and perl's own heuristic does not go there either.
func skipQuotedFrom(src []byte, i int) int {
	q := src[i]
	for j := i + 1; j < len(src); j++ {
		switch src[j] {
		case '\\':
			j++
		case q:
			return j + 1
		}
	}
	return -1
}

// isWordStart is isWordByte without the digits: a key may not begin with one.
func isWordStart(c byte) bool {
	return isAsciiLetter(c) || c == '_'
}
