// ABOUTME: The position-dependent scanners: variables, words, numbers, and the ambiguous punctuation.
// ABOUTME: What / < & and x mean depends on the expect state, which is why they live together.

package lexer

// scanVariable lexes a sigil and its name.
//
// The `$#` forms are the trap this exists for, and t/base/lex.t line 10 is
// the canonical case. Measured with B::Concise:
//
//	$ perl -MO=Concise,-exec -e '$x = $#[0];'
//	3  <#> aelemfast[*#] s
//
//	$ perl -e '@# = (42); print $#[0]'
//	42
//
// So `$#[0]` is element 0 of the array `@#`: the sigil is `$`, the NAME is
// `#`, and `[0]` is an ordinary subscript. It is NOT "$# followed by a
// subscript" -- there is no last-index term in it at all. The spec's gloss
// and this issue's own text both said otherwise, and both were wrong.
//
// The contrast: `$#x` IS last-index-of-@x, one lexeme. Same two opening
// bytes; what follows decides.
func scanVariable(l *lexer) bool {
	start := l.pos
	c := l.src[l.pos]
	if c != '$' && c != '@' && c != '%' {
		return false
	}

	// `%` and `@` are only sigils in term position; in operator position
	// they are modulus and... also `@` is never an operator, but `%` is, and
	// dispatching on position here is what keeps `$a % $b` from lexing as a
	// hash.
	if c == '%' && !l.expect.wantsTerm() {
		return false
	}

	l.pos++
	if l.pos >= len(l.src) {
		l.emit(Variable, start)
		return true
	}

	// `$#` is either last-index (`$#name`, `$#{...}`, `$#$ref`) or the
	// variable named `#`. A following identifier character or brace means
	// last-index; anything else -- including `[` -- means the name is `#`.
	if c == '$' && l.src[l.pos] == '#' {
		l.pos++
		if l.pos < len(l.src) && (isWordByte(l.src[l.pos]) || l.src[l.pos] == '{' || l.src[l.pos] == '$') {
			l.scanVarName()
		}
		l.emit(Variable, start)
		return true
	}

	l.scanVarName()
	l.emit(Variable, start)
	return true
}

// scanVarName consumes an identifier, a punctuation variable, or a braced
// name after a sigil.
func (l *lexer) scanVarName() {
	if l.pos >= len(l.src) {
		return
	}
	switch c := l.src[l.pos]; {
	case isWordByte(c):
		for l.pos < len(l.src) && (isWordByte(l.src[l.pos]) ||
			(l.src[l.pos] == ':' && l.pos+1 < len(l.src) && l.src[l.pos+1] == ':')) {
			if l.src[l.pos] == ':' {
				l.pos += 2
				continue
			}
			l.pos++
		}
	case c == '{':
		// A braced name: ${name}. The brace-matching here is deliberately
		// shallow; the full rule needs the block-vs-hash distinction, which
		// belongs to a later issue.
		depth := 0
		for l.pos < len(l.src) {
			if l.src[l.pos] == '{' {
				depth++
			} else if l.src[l.pos] == '}' {
				depth--
				l.pos++
				if depth == 0 {
					return
				}
				continue
			}
			l.pos++
		}
	case c == '$':
		// A dereference: $$ref, @$ref.
		l.pos++
		l.scanVarName()
	default:
		// A punctuation variable: $_, $0, $!, $@, $/ and the rest. One byte.
		l.pos++
	}
}

// scanWord lexes a bareword or keyword.
//
// Whether a word is a keyword, a function name or a bareword string cannot be
// decided here -- it needs the symbol table and the parser -- so the lexer
// reports Word and leaves the classification alone.
func scanWord(l *lexer) bool {
	if !isAsciiLetter(l.src[l.pos]) && l.src[l.pos] != '_' {
		return false
	}
	start := l.pos
	for l.pos < len(l.src) && isWordByte(l.src[l.pos]) {
		l.pos++
	}
	l.emit(Word, start)
	return true
}

// scanNumber lexes an integer or float literal. The exotic forms -- `0x_1234`,
// `0x0p0` -- are §0.13 rank 9 and belong to a later issue; this covers what
// the M0 corpus reaches.
func scanNumber(l *lexer) bool {
	c := l.src[l.pos]
	if c < '0' || c > '9' {
		return false
	}
	start := l.pos
	for l.pos < len(l.src) && (isWordByte(l.src[l.pos]) || l.src[l.pos] == '.') {
		l.pos++
	}
	l.emit(Number, start)
	return true
}

// scanAngle lexes `<FH>` in term position, and reports false in operator
// position so the caller emits a comparison operator instead.
//
// Spec §3.2: the same `<` is a readline where a value is expected and a
// comparison where one just ended.
func scanAngle(l *lexer) bool {
	if l.src[l.pos] != '<' || !l.expect.wantsTerm() {
		return false
	}
	// A readline's contents are a filehandle name, a scalar, or nothing.
	// Anything else -- a space, an operator -- means this was a comparison
	// after all, so scan ahead before committing.
	j := l.pos + 1
	for j < len(l.src) && (isWordByte(l.src[j]) || l.src[j] == '$' || l.src[j] == ':') {
		j++
	}
	if j >= len(l.src) || l.src[j] != '>' {
		return false
	}
	start := l.pos
	l.pos = j + 1
	l.emit(Readline, start)
	return true
}

// scanAmp lexes `&` as a function sigil in term position.
//
// §0.13 rank 7, 3 corpus files. Spec §3.2's `&` row: operator position yields
// `&&` or bitwise-and, term position yields `&name`.
func scanAmp(l *lexer) bool {
	if l.src[l.pos] != '&' || !l.expect.wantsTerm() {
		return false
	}
	// `&&` is an operator even where a term is expected -- it is the
	// low-precedence form. Only a single `&` before a name is a sigil.
	if l.pos+1 < len(l.src) && l.src[l.pos+1] == '&' {
		return false
	}
	start := l.pos
	l.pos++
	l.emit(FuncSigil, start)
	return true
}

// scanComment lexes `#` to end of line. Trivia, and therefore a token.
func scanComment(l *lexer) bool {
	if l.src[l.pos] != '#' {
		return false
	}
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
	}
	l.emit(Comment, start)
	return true
}

// operators is checked longest-first so `<=>` is not read as `<=` then `>`.
var operators = []string{
	"<=>", "**=", "||=", "&&=", "//=", "...", "<<=", ">>=",
	"=~", "!~", "->", "++", "--", "**", "==", "!=", "<=", ">=",
	"&&", "||", "//", "..", "::", "+=", "-=", "*=", "/=", ".=",
	"%=", "^=", "|=", "&=", "=>", "<<", ">>",
	"+", "-", "*", "/", "%", ".", ",", "=", "<", ">", "!", "?", ":",
	"&", "|", "^", "~", "(", ")", "[", "]", "{", "}", "\\",
}

// scanOperator lexes punctuation.
func scanOperator(l *lexer) bool {
	for _, op := range operators {
		if l.pos+len(op) > len(l.src) {
			continue
		}
		if string(l.src[l.pos:l.pos+len(op)]) == op {
			start := l.pos
			l.pos += len(op)
			// A closing bracket ends a term, so an operator follows it;
			// every other operator opens a term. Without the distinction,
			// `$x[0] <FH>` reads its angle brackets as a readline.
			kind := Operator
			if op == ")" || op == "]" || op == "}" {
				kind = CloseBracket
			}
			l.emit(kind, start)
			return true
		}
	}
	return false
}

// scanSemicolon ends a statement, returning the machine to XSTATE.
func scanSemicolon(l *lexer) bool {
	if l.src[l.pos] != ';' {
		return false
	}
	start := l.pos
	l.pos++
	l.emit(Semicolon, start)
	return true
}

// scanBarePattern lexes `/.../` in TERM position.
//
// scanQuoteLike handles the keyword forms (`m//`, `s///`); this one has no
// keyword, so the expect state is the ONLY thing distinguishing a pattern
// from division. Spec §3.2's `/` row. Measured:
//
//	$ perl -MO=Deparse -e 'my @a=(1); my $n = @a / 2;'
//	my $n = @a / 2;
//
// t/base/lex.t:20 is the harder case, `eval '$foo{1} / 1;'` -- division,
// because a hash subscript closes a term.
//
// Bare `?...?` is deliberately absent: it was removed as a match operator in
// perl 5.22 and is a syntax error in 5.42, verified.
func scanBarePattern(l *lexer) bool {
	if l.src[l.pos] != '/' || !l.expect.wantsTerm() {
		return false
	}
	start := l.pos
	l.pos++
	if !l.scanDelimitedBody('/', 0) {
		l.emit(UnknownRest, start)
		return true
	}
	l.scanModifiers()
	l.emit(Quote, start)
	return true
}
