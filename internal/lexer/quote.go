// ABOUTME: Quote-like operators: q qq qw m s tr y qr and the plain string forms.
// ABOUTME: Delimiter scanning only — no interpolation, no regex parsing, no heredocs.

package lexer

// quoteOp describes one keyword-introduced quote-like operator.
//
// parts is how many delimited sections it takes: 2 for q/qq/qw/m/qr (the
// operator plus one body), 3 for s/tr/y (operator plus pattern plus
// replacement). The count drives whether a second delimiter pair is read.
type quoteOp struct {
	name  string
	parts int
	// mods is whether trailing letters belong to the operator. Only the
	// matching forms take them: perl reads `q/a/b/` as `q/a/` followed by
	// the bareword `b`, not as a q with modifier `b`.
	mods bool
}

// quoteOps is checked longest-first so `tr` is not read as `t` followed by
// `r`, and `qq`/`qw`/`qr` are not read as `q`.
var quoteOps = []quoteOp{
	{"tr", 3, true},
	{"qq", 2, false},
	{"qw", 2, false},
	{"qr", 2, true},
	{"s", 3, true},
	{"y", 3, true},
	{"m", 2, true},
	{"q", 2, false},
}

// pairedCloser returns the closing delimiter for a bracketing opener, or 0.
//
// These four are the only pairs. Everything else closes with itself, which is
// what toke.c:12412 means by `if (PL_multi_open == PL_multi_close)`: nesting
// is the exception, not the rule.
func pairedCloser(open byte) byte {
	switch open {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	case '<':
		return '>'
	}
	return 0
}

// scanQuoteLike lexes a quote-like operator at the cursor and reports whether
// it found one. The cursor is left after the operator's last byte, modifiers
// included.
func scanQuoteLike(l *lexer) bool {
	start := l.pos

	// A plain string form is its own delimiter: '...', "...", `...`.
	if c := l.src[l.pos]; c == '\'' || c == '"' || c == '`' {
		l.pos++
		if !l.scanDelimitedBody(c, pairedCloser(c)) {
			l.emit(UnknownRest, start)
			return true
		}
		l.emit(Quote, start)
		return true
	}

	op, ok := quoteOpAt(l.src, l.pos)
	if !ok {
		return false
	}
	after := l.pos + len(op.name)

	// A keyword is only a quote operator if what follows can delimit. `q` in
	// `$q` or `sub q_thing` is an identifier, and `s` in `$s = 1` is not a
	// substitution. A word character immediately after the keyword means it
	// is part of a longer name.
	if after < len(l.src) && isWordByte(l.src[after]) {
		return false
	}

	l.pos = after
	// A '#' GLUED to the keyword is the delimiter; only a '#' reached after
	// skipping whitespace is a comment. Measured: `q#a#` is the string "a",
	// `q #a#` is a comment and the delimiter follows it.
	var open byte
	var found bool
	if l.pos < len(l.src) && l.src[l.pos] == '#' {
		open, found = '#', true
	} else {
		open, found = l.skipToDelimiter()
	}
	if !found {
		// A quote operator with nothing to delimit: the keyword is all there
		// is. Treat it as unlexable rather than silently consuming the rest.
		l.pos = after
		l.emit(Error, start)
		return true
	}
	close := pairedCloser(open)
	l.pos++ // past the opening delimiter
	replStart := 0

	if !l.scanDelimitedBody(open, close) {
		l.emit(UnknownRest, start)
		return true
	}

	if op.parts == 3 {
		// A three-part operator reuses a non-bracketing delimiter for the
		// replacement -- `s/a/b/` has closed only the pattern so far -- but
		// takes a whole new pair when the first was bracketing, because
		// `s{a}{b}` has already consumed its closing brace.
		if close != 0 {
			open2, found2 := l.skipToDelimiter()
			if !found2 {
				l.emit(UnknownRest, start)
				return true
			}
			l.pos++
			replStart = l.pos
			if !l.scanDelimitedBody(open2, pairedCloser(open2)) {
				l.emit(UnknownRest, start)
				return true
			}
		} else {
			replStart = l.pos
			if !l.scanDelimitedBody(open, 0) {
				l.emit(UnknownRest, start)
				return true
			}
		}
	}

	replEnd := l.pos
	if op.mods {
		l.scanModifiers()
	}
	l.emit(Quote, start)

	// With /e the replacement is CODE, not a string. t/base/lex.t:114:
	//
	//	$foo =~ s/^not /substr(<<EOF, 0, 0)/e;
	//	  Ignored
	//	EOF
	//
	// A heredoc opened inside the replacement takes its body from the lines
	// AFTER the substitution. Treat the replacement as opaque and the <<EOF
	// is missed, the body lexes as barewords, and the file still round-trips
	// -- which is why round-trip alone cannot gate this milestone.
	//
	// Only the heredoc openers are harvested here, not a full re-lex. The
	// replacement's own tokens are already inside the Quote span, and
	// emitting them again would break the round trip by covering bytes
	// twice. What must escape the span is the PENDING state: the queue entry
	// that makes the body arrive in the right place.
	if op.parts == 3 && replStart > 0 && hasModifier(l.src[replEnd:l.pos], 'e') {
		l.queueHeredocsIn(replStart, replEnd)
	}
	return true
}

// hasModifier reports whether the modifier run contains c.
func hasModifier(mods []byte, c byte) bool {
	for _, m := range mods {
		if m == c {
			return true
		}
	}
	return false
}

// queueHeredocsIn scans an /e replacement for heredoc openers and queues
// them, so their bodies are taken after the current line.
//
// A sub-lexer over the replacement's bytes: it shares nothing with the outer
// cursor, and only the pending queue crosses back.
func (l *lexer) queueHeredocsIn(start, end int) {
	sub := &lexer{src: l.src[:end], pos: start, expect: XTerm}
	for sub.pos < end {
		before := sub.pos
		if scanHeredocOpen(sub) {
			continue
		}
		sub.pos++
		if sub.pos <= before {
			break
		}
	}
	l.pending = append(l.pending, sub.pending...)
}

// skipToDelimiter advances past whitespace and comments to the delimiter,
// returning it without consuming it.
//
// perl's skipspace runs BEFORE delimiter selection, and it skips comments as
// well as whitespace. That is the whole of the `q #a#` asymmetry: glued, the
// `#` delimits; spaced, the `#` opens a comment and the delimiter is whatever
// comes next -- possibly on a later line. Measured:
//
//	$ perl -e 'print q #comment
//	xfoox'
//	foo
//
// The delimiter there is `x`, two lines down. §2.10.5's comment between the
// two halves of `s{a} # c\n {b}` is the same rule reached from the other
// side.
func (l *lexer) skipToDelimiter() (byte, bool) {
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case isSpace(c):
			l.pos++
		case c == '#':
			for l.pos < len(l.src) && l.src[l.pos] != '\n' {
				l.pos++
			}
		default:
			return c, true
		}
	}
	return 0, false
}

// scanDelimitedBody consumes a body and its closing delimiter, reporting
// whether the closer was found. close is 0 for a delimiter that closes with
// itself.
//
// THE ORDER OF THE TWO CHECKS BELOW IS THE WHOLE POINT. The closer is tested
// before the escape, because when the delimiter IS a backslash there are no
// escapes -- every `\` is a delimiter. perl guards it explicitly at
// toke.c:12445 (`close_delim_code != '\\'`); PerlOnJava checks the closer
// first and is right; gotreesitter and perl-lsp check the escape first, eat
// their own closer, and scan to EOF. Measured:
//
//	$ perl -e '$_="a"; s\a\b\; print'
//	b
func (l *lexer) scanDelimitedBody(open, close byte) bool {
	closer := close
	if closer == 0 {
		closer = open
	}
	depth := 1
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case c == closer:
			depth--
			l.pos++
			if depth == 0 {
				return true
			}
		case close != 0 && c == open:
			// Only a bracketing pair nests; a self-closing delimiter reaches
			// the case above first and ends the body.
			depth++
			l.pos++
		case c == '\\' && closer != '\\':
			// An escape consumes the next byte too, so a closer cannot hide
			// behind a backslash -- unless the backslash IS the closer, which
			// the case above already handled.
			l.pos++
			if l.pos < len(l.src) {
				l.pos++
			}
		default:
			l.pos++
		}
	}
	return false
}

// scanModifiers consumes the trailing letters of m//, s///, tr/// and qr//.
//
// They are part of the operator's span rather than a separate token: `s/a/b/`
// and `s/a/b/g` differ in meaning, and a consumer that has to re-lex the next
// token to learn which it got has the wrong token boundaries.
func (l *lexer) scanModifiers() {
	for l.pos < len(l.src) && isAsciiLetter(l.src[l.pos]) {
		l.pos++
	}
}

// quoteOpAt matches the longest quote-operator keyword at pos.
func quoteOpAt(src []byte, pos int) (quoteOp, bool) {
	for _, op := range quoteOps {
		if pos+len(op.name) > len(src) {
			continue
		}
		if string(src[pos:pos+len(op.name)]) == op.name {
			return op, true
		}
	}
	return quoteOp{}, false
}

func isAsciiLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// isWordByte is the ASCII word-character test. The full identifier class,
// including the Unicode rules under `use utf8`, belongs to the identifier
// issue; this is only enough to tell `q` from `q_thing`.
func isWordByte(c byte) bool {
	return isAsciiLetter(c) || c >= '0' && c <= '9' || c == '_'
}
