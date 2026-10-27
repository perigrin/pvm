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
// It runs inside a parenthesised call too, on the same two readings, which is
// why the `(` carries the lookahead forward rather than forcing a block --
// also measured on 5.42.0:
//
//	map({; a => 1} @a)        map({'a', 1;} @a)        a BLOCK
//	map({a => 1}, @a)         map({'a', 1}, @a)        a HASHREF
//
// A word or string followed by `,` or `=>` is a hashref; everything else is a
// block -- a leading `;`, an operator, and a VARIABLE before a fat comma,
// which is why the third line is a block and not a hash. That is the same
// rule the parser's braceOpensAnonHash applies at statement start, and the
// agreement is not a coincidence: perl runs this one function for both.
//
// An EMPTY brace -- its `}` the first thing after the `{`, past whitespace
// and comments -- is a hash when emptyIsHash is set, which is the caller
// saying the word is `map`, `grep`, `sort`, or the handle or program slot of
// `print`, `printf`, `system` or `exec` (emptyBraceIsHash). toke.c answers HASHBRACK for
// it before the first-tokens heuristic runs (toke.c:6706), measured on 5.42.0:
//
//	map {}, @a                map({}, @a)              a HASHREF
//	map {} @a                 syntax error near "} @a"   a hash, then no comma
//
// That is braceOpensAnonHash's empty rule too. After any other word the
// empty brace stays a block: `zzz {} print "b"`, `any {} @a` and `eval {}`
// each compile one, because perl reaches their brace by another path.
//
// Byte scanning rather than tokenizing, for peekIsOpenBrace's reason: the
// lexer stays single-pass, and a lookahead that lexed would have to undo its
// own state changes.
func (l *lexer) intuitCurly(emptyIsHash bool) bool {
	i := skipSpaceFrom(l.src, l.pos)
	if i >= len(l.src) || l.src[i] != '{' {
		return false
	}
	// perl's skipspace, which skips comments: `map { # c\n a => 1 }, @a` is
	// a hash, and `map { # c\n }, @a` an empty one.
	i = skipSpaceAndCommentsFrom(l.src, i+1)
	if i >= len(l.src) {
		return true
	}
	if l.src[i] == '}' {
		return !emptyIsHash
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
		case ' ', '\t', '\n', '\r', '\f', '\v':
		default:
			return i
		}
	}
	return i
}

// skipSpaceAndCommentsFrom is skipSpaceFrom that also skips `#` comments,
// each to the end of its line.
func skipSpaceAndCommentsFrom(src []byte, i int) int {
	for i = skipSpaceFrom(src, i); i < len(src) && src[i] == '#'; i = skipSpaceFrom(src, i) {
		for i < len(src) && src[i] != '\n' {
			i++
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
