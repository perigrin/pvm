// ABOUTME: The bracket stack: whether a `}` closed a block or a subscript, which decides what follows.
// ABOUTME: perl keeps PL_lex_brackstack for the same reason — the two leave different states behind.

package lexer

import "bytes"

// peekIsOpenBrace reports whether the next non-space byte is `{`.
//
// One byte of lookahead past whitespace, which is exactly what perl does at a
// closing paren (toke.c yyl_rightparen: `s = skipspace(s); if (*s == '{')`).
// It does not tokenize ahead: the lexer stays single-pass, and a lookahead
// that lexed would have to undo its own state changes.
func (l *lexer) peekIsOpenBrace() bool {
	pending := l.pending
	for i := l.pos; i < len(l.src); i++ {
		switch l.src[i] {
		case '\n':
			// Queued heredoc bodies start after this newline, and perl has
			// already cut them out of the line it looks ahead in: the `{`
			// after `for (split /\n/, <<'EOF')` and its body is the loop's.
			j := i + 1
			for _, h := range pending {
				j, _ = heredocBodyEnd(l.src, j, h)
			}
			pending = nil
			i = j - 1
			continue
		case ' ', '\t', '\r':
			continue
		case '#':
			// A comment, skipped as perl's skipspace skips it: `if ($x) # c`
			// with its `{` on the next line. Up to the newline, which the
			// case above then takes.
			nl := bytes.IndexByte(l.src[i:], '\n')
			if nl < 0 {
				return false
			}
			i += nl - 1
			continue
		case '{':
			return true
		default:
			return false
		}
	}
	return false
}

// peekIsOpenParen reports whether the next non-space byte is `(`, so a
// block-taking word can tell a parenthesised call from a bare one -- the
// spelling whose brace its own lookahead cannot reach.
func (l *lexer) peekIsOpenParen() bool {
	i := skipSpaceFrom(l.src, l.pos)
	return i < len(l.src) && l.src[i] == '('
}

// bracket is what an opener was, recorded so its closer knows what it closed.
type bracket int

const (
	// braceBlock is a `{` that opened a block: a sub body, a conditional, a
	// bare block. A statement follows its `}`, so a TERM is expected.
	braceBlock bracket = iota

	// braceTerm is a `{` that opened a subscript or an anonymous hash. A
	// value has just been produced, so an OPERATOR is expected after its `}`.
	braceTerm

	// paren and square are `(` and `[`. Both always produce a value, so
	// their closers behave alike and the distinction is only for matching.
	paren
	square
)

// trackBrackets maintains the stack. Called from emit, before the expect
// transition, so that the transition can ask what a closer just closed.
//
// perl does this with PL_lex_brackstack and the same reasoning: a `}` is the
// one closer whose meaning is not determined by the byte. `$h{a}` and
// `if (..) { .. }` end with the same character and leave opposite states.
func (l *lexer) trackBrackets(k Kind, start int) {
	l.closedBlock = false
	l.openedBlock = false

	if l.pos-start != 1 {
		return
	}
	switch l.src[start] {
	case '(':
		l.brackets = append(l.brackets, paren)
	case '[':
		l.brackets = append(l.brackets, square)
	case '{':
		// The decision is made here, from the state the `{` was READ in, not
		// from anything after it.
		//
		// In OPERATOR position a value has just been produced, so the `{`
		// subscripts it: `$h{a}`, `$r->{k}`. In XSTATE a statement is
		// starting, so it opens a block. Everywhere else a term is expected,
		// where perl's own answer is an anonymous hash -- also a value.
		//
		// That is intuit_curly minus its lookahead at the first token inside.
		// The lookahead settles `map { $_ => 1 }` versus `map { ; ... }`,
		// which is the parser's decision (spec §4.9.2). Here the question is
		// only which state the CLOSER leaves behind, and the opener's
		// position answers it.
		switch {
		case l.expect == XState || l.expect == XBlock:
			l.brackets = append(l.brackets, braceBlock)
			l.openedBlock = true
		default:
			l.brackets = append(l.brackets, braceTerm)
		}
	case ')', ']', '}':
		if n := len(l.brackets); n > 0 {
			top := l.brackets[n-1]
			l.brackets = l.brackets[:n-1]
			l.closeLexSubScope()
			l.closedBlock = top == braceBlock
		}
		// A closer with nothing open leaves closedBlock false. An LSP sees
		// half-typed buffers where that happens constantly, and treating the
		// unmatched case as a value is the same hedge as everywhere else:
		// it keeps token boundaries recoverable.
	}
}
