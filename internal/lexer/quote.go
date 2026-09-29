// ABOUTME: Quote-like operators: q qq qw m s tr y qr and the plain string forms.
// ABOUTME: Delimiter scanning only — no interpolation, no regex parsing, no heredocs.

package lexer

import "unicode/utf8"

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
	// `qx//` runs a command, like backticks. §4.14.2 records that a backtick
	// string already lexes as one Quote whose Text keeps its delimiters, and
	// that B::SoN names the form `BacktickExpr` -- a language fact. This is
	// the same node reached by the other spelling.
	//
	// No modifiers: perl reads `qx/a/b` as the string then the bareword `b`,
	// which is the same rule that gives q, qq and qw a false here.
	{"qx", 2, false},
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
//
// A rune rather than a byte to match the delimiter it compares against, but
// the table stays ASCII: perl rejects a paired multi-byte delimiter outright
// ("Use of '«' is deprecated as a string delimiter", then "Can't find string
// terminator"), so there is no pair to add.
func pairedCloser(open rune) rune {
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
	if c := rune(l.src[l.pos]); c == '\'' || c == '"' || c == '`' {
		l.pos++
		if !l.scanDelimitedBody(c, pairedCloser(c)) {
			l.emit(UnknownRest, start)
			return true
		}
		l.emit(Quote, start)
		return true
	}

	// After `->` a name is a method, never a quote operator. `s`, `y`, `q`,
	// `m` and `tr` are all legal method names and perl reads every one of
	// them as a name here -- measured:
	//
	//	$ perl -MO=Deparse -e 'my $o; my @x = ($o->s, $o->tr);'
	//	my(@x) = ($o->s, $o->tr);
	//
	// The plain string forms above are still quotes: `$o->"x"` is not a
	// method name in any spelling, so only the keyword forms decline.
	//
	// The NAME of a `sub` or a `method` is a name for the same reason --
	// perl deparses `sub y { 1 }` back to itself -- and sawSubWord is
	// already the lexer's record that the next word is one. `method y { }`
	// in t/class/field.t:166 was reading as a transliteration whose
	// delimiter was `{`, which swallowed the rest of the class body.
	if l.expect == XPostDeref || l.sawSubWord {
		return false
	}
	// After the `&` sigil the word is a sub's NAME as well: `&m('x')` calls
	// the sub m. Measured on 5.42.0, Deparse keeps `&m('a')`, `&s('b')` and
	// `&y('c')` as calls.
	if n := len(l.toks); n > 0 && l.toks[n-1].Kind == FuncSigil && l.toks[n-1].End == l.pos {
		return false
	}

	op, ok := quoteOpAt(l.src, l.pos)
	if !ok {
		return false
	}
	// A lexical sub of the operator's name shadows it: `my sub s` makes
	// `s(1)` a call in its scope. See noteLexSub.
	if l.lexSubInScope(op.name) {
		return false
	}
	after := l.pos + len(op.name)

	// A keyword is only a quote operator if what follows can delimit. `q` in
	// `$q` or `sub q_thing` is an identifier, and `s` in `$s = 1` is not a
	// substitution. A word character immediately after the keyword means it
	// is part of a longer name.
	//
	// The word character may be MULTI-BYTE. Under `use utf8` perl widens the
	// identifier class, and t/uni/gv.t:488 has a sub whose name begins with a
	// quote keyword:
	//
	//	my $rv = \*sምḲ;
	//
	// `sምḲ` is one identifier -- verified by running it -- so `s` is not a
	// substitution here. Testing only the next BYTE reads `ም` as a delimiter
	// and the substitution then runs to EOF, swallowing 10KB of the file.
	//
	// identContinue is the same predicate the identifier scanner uses, so the
	// two cannot disagree about where a name ends.
	if after < len(l.src) {
		r, _ := utf8.DecodeRune(l.src[after:])
		if identContinue(r, l.utf8Pragma) {
			return false
		}
	}

	// A fat comma quotes the word to its left (§4.5.4), and that includes
	// these keywords. Measured on perl 5.42.0:
	//
	//	$ perl -MO=Deparse -e 'my $h = { s => 1, y => 2, tr => 3, m => 4 };'
	//	my $h = {'s', 1, 'y', 2, 'tr', 3, 'm', 4};
	//
	// Without this, `s` takes `=` as its delimiter and runs to the next one,
	// so `{ s => 1 }` lexes as a substitution and everything after it
	// becomes one opaque token. Whitespace before the `=>` is skipped
	// because perl skips it.
	if fatCommaFollows(l.src, after) {
		return false
	}

	// A `}` never delimits a quote operator in a program perl will COMPILE,
	// so a quote-op name that reaches one is the bareword. This holds in
	// every context and not only in a subscript -- measured on 5.42.0, where
	// each of these is a fatal error rather than an operator:
	//
	//	$ perl -e 'sub f { m }'
	//	Search pattern not terminated at -e line 1.
	//	$ perl -e 'for (1) { s }'
	//	Substitution pattern not terminated at -e line 1.
	//	$ perl -e 'my @a = map { qw } 1;'
	//	Can't find string terminator "}" anywhere before EOF at -e line 1.
	//
	// The third is the one that shows WHY this is safe to decide on the byte
	// alone. perl does take the `}` there and then cannot find its closer; a
	// bareword and a run-to-EOF quote are both refusals of source perl
	// refuses, so declining costs nothing, while the SUBSCRIPT -- the only
	// spelling that compiles -- becomes right.
	//
	// That is why this asks the BYTE and not the bracket stack. The
	// construct it appears in is a subscript, where perl autoquotes the key
	// exactly as it does before a fat comma:
	//
	//	$ perl -MO=Deparse -e 'my %h; my $a=$h{m}; my $b=$h{s}; my $c=$h{tr};'
	//	  ->  $h{'m'}  $h{'s'}  $h{'tr'}
	//
	// Without this, `$h{m}` lexed as a match delimited by `}` whose body ran
	// to the NEXT `}`: the emission gained a spurious `};` and the enclosing
	// construct lost its closer. `$h{m}{q}` swallowed BOTH subscripts into
	// one Quote and still scored Unknown=0, so only the canon could see it.
	//
	// It declines on the byte AFTER the name, so a brace-delimited operator
	// in the same position is untouched -- perl reads `$h{ m{a} }` as the
	// match `/a/`, and there the byte is `{`.
	if closeBraceFollows(l.src, after) {
		return false
	}

	l.pos = after
	// A '#' GLUED to the keyword is the delimiter; only a '#' reached after
	// skipping whitespace is a comment. Measured: `q#a#` is the string "a",
	// `q #a#` is a comment and the delimiter follows it.
	var open rune
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
	l.pos += utf8.RuneLen(open) // past the opening delimiter, whole
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
			l.pos += utf8.RuneLen(open2)
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
// The delimiter is returned as a RUNE, not a byte. perl compares the whole
// character, and `qq ϟ a ϟ` closes on the second `ϟ` rather than on its first
// byte -- matching a byte at a time ends the token mid-sequence, which is a
// span that cannot be re-lexed or shown to a user. Measured on 5.42.0:
//
//	qq ϟ hello ϟ        works
//	q«paired»           "Use of '«' is deprecated as a string delimiter"
//	                    then "Can't find string terminator"
//
// so only the NON-PAIRED multi-byte forms need to work; perl itself rejects
// the paired ones, and pairedCloser stays byte-oriented because every
// bracketing delimiter perl accepts is ASCII.
func (l *lexer) skipToDelimiter() (rune, bool) {
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
			r, _ := utf8.DecodeRune(l.src[l.pos:])
			return r, true
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
//
// A delimiter is a CHARACTER, so both the opener and the closer are runes and
// every advance moves a whole one. `ϟ` is two bytes, and comparing the first
// of them ends the token inside a UTF-8 sequence.
func (l *lexer) scanDelimitedBody(open, close rune) bool {
	closer := close
	if closer == 0 {
		closer = open
	}
	depth := 1
	for l.pos < len(l.src) {
		c, width := utf8.DecodeRune(l.src[l.pos:])
		switch {
		case c == closer:
			depth--
			l.pos += width
			if depth == 0 {
				return true
			}
		case close != 0 && c == open:
			// Only a bracketing pair nests; a self-closing delimiter reaches
			// the case above first and ends the body.
			depth += 1
			l.pos += width
		case c == '\\' && closer != '\\':
			// An escape consumes the next CHARACTER too, so a closer cannot
			// hide behind a backslash -- unless the backslash IS the closer,
			// which the case above already handled.
			l.pos += width
			if l.pos < len(l.src) {
				_, w := utf8.DecodeRune(l.src[l.pos:])
				l.pos += w
			}
		default:
			l.pos += width
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
// fatCommaFollows reports whether the next significant bytes at pos are `=>`.
//
// Only whitespace is skipped, and only horizontal whitespace plus newlines --
// a comment between a word and its fat comma is legal Perl but vanishingly
// rare, and reaching for it here would mean re-implementing comment skipping
// in a function whose whole job is one two-byte lookahead.
func fatCommaFollows(src []byte, pos int) bool {
	for pos < len(src) {
		switch src[pos] {
		case ' ', '\t', '\n', '\r':
			pos++
		default:
			return pos+1 < len(src) && src[pos] == '=' && src[pos+1] == '>'
		}
	}
	return false
}

// closeBraceFollows reports whether the next significant byte at pos is `}`.
//
// Whitespace is skipped for fatCommaFollows's reason: perl skips it too, and
// `$h{ m }` deparses to `$h{'m'}`. skipSpaceFrom is the lexer's own space
// predicate, so this cannot disagree with the rest of the scanner about
// where whitespace ends.
func closeBraceFollows(src []byte, pos int) bool {
	pos = skipSpaceFrom(src, pos)
	return pos < len(src) && src[pos] == '}'
}

// HasQuoteOperator reports whether a Quote token's text RUNS an operator
// rather than being a plain string literal.
//
// Exported for the conformance corpus, which distinguishes `qw(a b)` (a
// list) from `"hi"` (a string) while both arrive as Kind Quote. Keeping
// the answer here rather than in a second table is what stops the two
// from drifting: a copy of `quoteOps` made elsewhere lost `qx` on its
// first day, which silently reclassified `qx/ls/` as a string.
func HasQuoteOperator(text string) bool {
	// Backticks run a command without naming an operator, so no prefix
	// test reaches them. They are the same operation as `qx//`, which
	// the table above records.
	if len(text) > 0 && text[0] == '`' {
		return true
	}

	op, ok := quoteOpAt([]byte(text), 0)
	if !ok {
		return false
	}
	// What follows the name must be a delimiter rather than more of a
	// longer word: `sort` begins with `s` and is not a substitution.
	rest := text[len(op.name):]
	return rest != "" && !isWordByte(rest[0])
}

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
