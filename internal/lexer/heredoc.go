// ABOUTME: Heredocs: recognised in term position only, bodies deferred to the end of the line.
// ABOUTME: The pending queue is what makes `print <<A, <<B;` work, and what a per-scan approach cannot do.

package lexer

import "bytes"

// pendingHeredoc is a heredoc whose terminator has been read but whose body
// has not started yet.
//
// The split is forced by the language: the terminator appears at the `<<`,
// the body begins after the current LINE ends. Everything between them is
// ordinary code, which is why the queue lives on the lexer rather than inside
// one scan.
type pendingHeredoc struct {
	term    []byte
	indent  bool // <<~ strips leading whitespace from the terminator line
	openPos int  // where the `<<` was, for diagnostics
}

// scanHeredocOpen lexes `<<TERM` in TERM position and queues the body.
//
// toke.c:7173-7177 gates this on the expect state:
//
//	if (PL_expect != XOPERATOR) {
//	    if (s[1] == '<' && s[2] != '>')
//	        s = scan_heredoc(s);
//
// In operator position the same bytes are a left shift, measured:
// `my $n = 1 << 3` is 8.
func scanHeredocOpen(l *lexer) bool {
	if l.src[l.pos] != '<' || l.pos+1 >= len(l.src) || l.src[l.pos+1] != '<' {
		return false
	}
	if !l.expect.wantsTerm() {
		return false // a left shift
	}
	// `<<>>` is the double-diamond readline, not a heredoc. toke.c checks
	// `s[2] != '>'` for exactly this.
	if l.pos+2 < len(l.src) && l.src[l.pos+2] == '>' {
		return false
	}

	start := l.pos
	j := l.pos + 2

	indent := false
	if j < len(l.src) && l.src[j] == '~' {
		indent = true
		j++
	}

	// Whitespace is permitted before a QUOTED terminator and forbidden before
	// a bare one. Measured in real files, because -e mangles the newlines:
	//
	//	print <<"EOF";   -> body
	//	print << "EOF";  -> body
	//	print << EOF;    -> Use of bare << to mean <<"" is forbidden
	//
	// An earlier draft of this issue had the asymmetry on the quoted form,
	// which is backwards. perl refuses the spaced bareword rather than
	// guessing an empty terminator.
	spaced := false
	for j < len(l.src) && (l.src[j] == ' ' || l.src[j] == '\t') {
		spaced = true
		j++
	}
	if j >= len(l.src) {
		return false
	}

	var term []byte
	switch q := l.src[j]; q {
	case '"', '\'', '`':
		k := j + 1
		for k < len(l.src) && l.src[k] != q {
			k++
		}
		if k >= len(l.src) {
			return false // unterminated: not a heredoc opener
		}
		term = l.src[j+1 : k]
		j = k + 1
	case '\\':
		// <<\EOF is the backslash-quoted form: non-interpolating, and the
		// terminator is the bare word after the backslash.
		j++
		s := j
		for j < len(l.src) && isWordByte(l.src[j]) {
			j++
		}
		if j == s {
			return false
		}
		term = l.src[s:j]
	default:
		if spaced {
			// `<< EOF` -- perl forbids this outright, so it is not a heredoc
			// and not a left shift either. Declining here leaves the `<<` to
			// the operator scanner, which is the honest shape: we do not
			// invent a construct perl rejects.
			return false
		}
		s := j
		for j < len(l.src) && isWordByte(l.src[j]) {
			j++
		}
		if j == s {
			return false
		}
		term = l.src[s:j]
	}

	l.pos = j
	l.pending = append(l.pending, pendingHeredoc{
		term: term, indent: indent, openPos: start,
	})
	l.emit(HeredocOpen, start)
	return true
}

// takePendingHeredocs consumes the queued bodies, in the order the heredocs
// were opened.
//
// Called when a newline is reached. FIFO rather than LIFO: `print <<A, <<B;`
// puts A's body first, which t/base/lex.t:56 exercises and which a stack
// would get backwards.
func (l *lexer) takePendingHeredocs() {
	for len(l.pending) > 0 {
		h := l.pending[0]
		l.pending = l.pending[1:]

		start := l.pos
		for l.pos < len(l.src) {
			lineStart := l.pos
			lineEnd := bytes.IndexByte(l.src[l.pos:], '\n')
			if lineEnd < 0 {
				// No terminator before EOF. The body is everything that is
				// left, reported as unterminated rather than silently
				// swallowed.
				l.pos = len(l.src)
				l.emit(UnknownRest, start)
				return
			}
			line := l.src[lineStart : lineStart+lineEnd]
			l.pos = lineStart + lineEnd + 1

			candidate := line
			if h.indent {
				candidate = bytes.TrimLeft(candidate, " \t")
			}
			if bytes.Equal(candidate, h.term) {
				l.emit(HeredocBody, start)
				break
			}
		}
		if l.pos >= len(l.src) && len(l.pending) == 0 {
			// Ran out without matching: already emitted above.
			return
		}
	}
}
