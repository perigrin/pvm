// ABOUTME: POD, __END__/__DATA__ and format bodies: three line-oriented modes.
// ABOUTME: All three end at a line rather than a delimiter, which is why they share a file.

package lexer

import "bytes"

// scanPod lexes a POD block: a `=word` line where a statement could start,
// through `=cut` inclusive.
//
// "Where a statement could start" is the expect state. t/base/lex.t:491-505
// puts `=pod` immediately after a line ending in `map{`, and perl accepts it
// -- measured, the map still returns 42 -- because a `{` leaves a statement
// expected (§3.1.4, XSTATE).
func scanPod(l *lexer) bool {
	if l.src[l.pos] != '=' || !l.atLineStart() {
		return false
	}
	if !l.expect.wantsTerm() {
		return false // `=` in expression position is assignment
	}
	// A directive is `=` followed by an identifier: `=pod`, `=head1`, `=cute`.
	// A bare `=` at line start is still assignment.
	if l.pos+1 >= len(l.src) || !isAsciiLetter(l.src[l.pos+1]) {
		return false
	}

	start := l.pos
	for l.pos < len(l.src) {
		lineStart := l.pos
		lineEnd := lineStart + lineLen(l.src[lineStart:])
		line := l.src[lineStart:lineEnd]
		l.pos = lineEnd
		if l.pos < len(l.src) {
			l.pos++ // the newline
		}
		if isPodCut(line) {
			l.emit(Pod, start)
			return true
		}
	}
	// POD that runs to EOF without `=cut` is still POD -- perl accepts it --
	// so this is Pod rather than UnknownRest.
	l.emit(Pod, start)
	return true
}

// isPodCut implements spec §2.5.1's three-conjunct rule: a line terminates
// POD when it starts `=cut` AND what follows is not an identifier character.
//
// t/base/lex.t:590-630 is a battery built to break a prefix match: `=cute`,
// `=cut2` and `=cut_` all look like `=cut` and none of them terminates.
// Measured -- a file whose POD opens at `=cute` and closes at a later `=cut`
// prints the statements on both sides, so `=cute` was inside the block.
func isPodCut(line []byte) bool {
	if !bytes.HasPrefix(line, []byte("=cut")) {
		return false
	}
	rest := line[len("=cut"):]
	if len(rest) == 0 {
		return true
	}
	return !isWordByte(rest[0])
}

// scanDataSection lexes `__END__` or `__DATA__` and everything after it.
//
// The rest of the file is data, not Perl. One trivia token, so the bytes
// still round-trip while nothing tries to lex `) not ( perl $$$` as code.
func scanDataSection(l *lexer) bool {
	if l.src[l.pos] != '_' || !l.atLineStart() {
		return false
	}
	rest := l.src[l.pos:]
	var marker []byte
	switch {
	case bytes.HasPrefix(rest, []byte("__END__")):
		marker = []byte("__END__")
	case bytes.HasPrefix(rest, []byte("__DATA__")):
		marker = []byte("__DATA__")
	default:
		return false
	}
	// The marker owns its whole line, so `__END__x` is an identifier.
	after := l.pos + len(marker)
	if after < len(l.src) && isWordByte(l.src[after]) {
		return false
	}

	start := l.pos
	l.pos = len(l.src)
	l.emit(DataSection, start)
	return true
}

// scanFormatBody lexes the picture lines of a `format NAME =` declaration,
// ending at a lone `.` on its own line.
//
// §0.13 rank 3: 8 corpus files, 3 in T2. Spec §2.13, §2.15 item 16; perl's
// LEX_FORMLINE. The body is the same SHAPE as a heredoc -- line-oriented,
// terminated by a sentinel line -- which is why §0.13 groups them.
//
// Picture lines are not Perl. `@<<<<< @>>>>>` would otherwise lex as an
// array sigil followed by left shifts, and `$a,     $b` on the next line is
// the argument list rather than an expression.
func scanFormatBody(l *lexer) bool {
	if !l.inFormat {
		return false
	}
	l.inFormat = false

	start := l.pos
	for l.pos < len(l.src) {
		lineStart := l.pos
		lineEnd := lineStart + lineLen(l.src[lineStart:])
		line := l.src[lineStart:lineEnd]
		l.pos = lineEnd
		if l.pos < len(l.src) {
			l.pos++
		}
		if bytes.Equal(bytes.TrimRight(line, " \t"), []byte(".")) {
			l.emit(FormatBody, start)
			return true
		}
	}
	l.emit(UnknownRest, start)
	return true
}

// noteFormat watches for the `=` that ends a `format NAME =` line, so the
// next line begins the picture body.
//
// The declaration is ordinary tokens; only what follows the newline is
// special, which is the same deferral heredocs need.
func (l *lexer) noteFormat(k Kind, start int) {
	switch {
	case k == Word && string(l.src[start:l.pos]) == "format":
		l.sawFormatWord = true
	case k == Operator && l.sawFormatWord && string(l.src[start:l.pos]) == "=":
		l.sawFormatWord = false
		l.pendingFormat = true
	case k == Whitespace || k == Comment:
		// Trivia does not end the declaration.
	case k == Word:
		// The format's name, between `format` and `=`.
	default:
		l.sawFormatWord = false
	}
}

// atLineStart reports whether the cursor is at the beginning of a line.
// POD, __END__ and format bodies are all line-oriented, so all three ask.
func (l *lexer) atLineStart() bool {
	return l.pos == 0 || l.src[l.pos-1] == '\n'
}

// lineLen returns the length of the first line in b, excluding its newline.
func lineLen(b []byte) int {
	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		return i
	}
	return len(b)
}
