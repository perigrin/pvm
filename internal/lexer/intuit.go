// ABOUTME: intuit_curly: whether a `{` after a list operator or at statement start opens a block or a hash.
// ABOUTME: The one brace whose meaning depends on what is inside it rather than on what precedes it.

package lexer

import (
	"bytes"
	"strings"
)

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
// braceIsHash decides the rest: a VARIABLE before a fat comma is not a key,
// which is why the third line is a block and not a hash. The lexer applies
// the same rule to a brace at statement start, because perl runs this one
// function for both.
//
// xref is the caller saying the word is one toke.c leaves at XREF -- `map`,
// `grep`, `sort`, or the handle or program slot of `print`, `printf`,
// `system` or `exec` (emptyBraceIsHash) -- whose brace reaches yyl_leftcurly's
// default case and so takes braceIsHash whole. An EMPTY brace there is a
// hash: toke.c answers HASHBRACK for it before the first-tokens heuristic
// runs (toke.c:6706), measured on 5.42.0:
//
//	map {}, @a                map({}, @a)              a HASHREF
//	map {} @a                 syntax error near "} @a"   a hash, then no comma
//
// That is braceIsHash's empty rule too. After any other word the
// empty brace stays a block: `zzz {} print "b"`, `any {} @a` and `eval {}`
// each compile one, because perl reaches their brace by another path.
//
// Byte scanning rather than tokenizing, for peekIsOpenBrace's reason: the
// lexer stays single-pass, and a lookahead that lexed would have to undo its
// own state changes.
func (l *lexer) intuitCurly(xref bool) bool {
	i := skipSpaceFrom(l.src, l.pos)
	if i >= len(l.src) || l.src[i] != '{' {
		return false
	}
	if xref {
		return !braceIsHash(l.src, i+1)
	}
	// After any other word toke.c's brace never reaches the default case,
	// so the narrower guess pu shipped stands: an empty brace is a block,
	// and a word or plain string before `,` or `=>` a hash.
	i = skipSpaceAndCommentsFrom(l.src, i+1)
	if i >= len(l.src) || l.src[i] == '}' {
		return true
	}
	end := i
	switch c := l.src[i]; {
	case c == '\'' || c == '"':
		if end = skipQuotedFrom(l.src, i); end < 0 {
			return true
		}
	case isAsciiLetter(c) || c == '_':
		for end < len(l.src) && isWordByte(l.src[end]) {
			end++
		}
	default:
		return true
	}
	end = skipSpaceFrom(l.src, end)
	return end >= len(l.src) || l.src[end] != ',' &&
		(l.src[end] != '=' || end+1 >= len(l.src) || l.src[end+1] != '>')
}

// braceIsHash is toke.c's yyl_leftcurly default case (6698-6814): whether
// the `{` just before src[i] reads HASHBRACK. It is empty, or its first term
// -- a quoted string, a q-like construct or a run of word characters -- is
// followed by `=>`, or by `,` when the term does not start with a lower-case
// letter other than q. Measured with -DT on the pinned perl:
//
//	{ 1, 2 }     { Foo, 1 }    { q(a), 1 }    { qux, 1 }    HASHBRACK
//	{ foo, 1 }   { qw(a), 1 }  { -foo => 1 }  { 1 }         a block
//
// shortcut: word characters are ASCII only; a `use utf8` key such as
// `{ é => 1 }` reads as a block. Upgrade with the identifier class.
func braceIsHash(src []byte, i int) bool {
	s := skipSpaceAndCommentsFrom(src, i)
	if s >= len(src) {
		return false
	}
	if src[s] == '}' {
		return true
	}
	// toke.c scans no further than PL_bufend, the end of the first term's line.
	if nl := bytes.IndexByte(src[s:], '\n'); nl >= 0 {
		src = src[:s+nl+1]
	}
	t := s
	switch c := src[s]; {
	case c == '\'' || c == '"' || c == '`':
		if t = skipQuotedFrom(src, s); t < 0 {
			return false
		}
	case c == 'q' && s+1 < len(src) && (!isWordByte(src[s+1]) ||
		(src[s+1] == 'q' || src[s+1] == 'x') && s+2 < len(src) && !isWordByte(src[s+2])):
		d := s + 1
		if isWordByte(src[d]) { // the second letter of qq or qx
			d++
		}
		t = skipSpaceFrom(src, d)
		if t+1 < len(src) && src[t] == '=' && src[t+1] == '>' {
			return true
		}
		if t = skipDelimitedFrom(src, t); t < 0 {
			return false
		}
	default:
		for t < len(src) && isWordByte(src[t]) {
			t++
		}
	}
	t = skipSpaceFrom(src, t)
	if t >= len(src) {
		return false
	}
	if src[t] == ',' {
		return src[s] == 'q' || src[s] < 'a' || src[s] > 'z'
	}
	return src[t] == '=' && t+1 < len(src) && src[t+1] == '>'
}

// blockExpectsTerm is toke.c's block_expectation (6816-6836): whether the
// block an XREF word's `{` just opened wants a TERM first rather than a
// statement. It does when a `{` or `sub` then `:` comes next, so
// `map { { $_->get_columns } } @x` builds an anonymous hash and `map { sub :
// lvalue {1} } @x` an attributed anonymous sub, not a label. toke.c reaches it
// only in XREF: see xref in emit. Measured with -DT on the pinned perl; see
// TestBlockExpectationBraceIsAnonHash.
//
// shortcut: a newline or comment between `sub` and `:` is skipped as spaces
// are, where toke.c loses the `sub` with its line and perl refuses. Upgrade
// if a corpus file has one (TestBlockExpectationBraceKnownLimits).
func (l *lexer) blockExpectsTerm() bool {
	s := l.skipSpaceAt(l.pos)
	if s < len(l.src) && l.src[s] == '{' {
		return true
	}
	d := l.skipSpaceAt(s + 3)
	return bytes.HasPrefix(l.src[s:], []byte("sub")) && d < len(l.src) && l.src[d] == ':'
}

// skipDelimitedFrom returns the index just past the q-like body whose opening
// delimiter is src[i], or -1 if it is unterminated. A bracketing delimiter
// nests, as in toke.c's q-construct skip.
func skipDelimitedFrom(src []byte, i int) int {
	if i >= len(src) {
		return -1
	}
	open := src[i]
	closer := open
	if k := strings.IndexByte("([{<", open); k >= 0 {
		closer = ")]}>"[k]
	}
	for depth, j := 1, i+1; j < len(src); j++ {
		switch {
		case src[j] == '\\' && open != '\\':
			j++
		case src[j] == closer:
			if depth--; depth == 0 {
				return j + 1
			}
		case src[j] == open:
			depth++
		}
	}
	return -1
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
// A plain quote or a backtick; skipDelimitedFrom takes the q-like ones.
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
