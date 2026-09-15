// ABOUTME: A hand-written Perl lexer: byte offsets, trivia as tokens, forward progress guaranteed.
// ABOUTME: Go stdlib only — the whole point of replacing the tree-sitter grammar.

package lexer

// Kind names what a token is. Trivia kinds (Whitespace, and later Comment and
// Pod) are ordinary kinds rather than a separate channel: every byte of input
// belongs to exactly one token, which is what makes lossless round-trip
// reachable. A lexer that discards whitespace can never satisfy it.
type Kind int

const (
	// Whitespace is any run of spaces, tabs, newlines and carriage returns.
	Whitespace Kind = iota

	// Error is a byte that cannot start any Perl token. It spans exactly the
	// bytes that could not be lexed, so the input is still covered.
	//
	// An error is a KIND rather than a return value because Tokenize has no
	// error result: returning early would drop the rest of the file, and
	// skipping the byte would break the round-trip invariant. Spec §2.15
	// item 18.
	Error

	// UnknownRest is a construct that began and never ended -- an
	// unterminated heredoc, quote-like operator or POD block. It spans from
	// the opening to the end of input.
	//
	// Distinct from Error on purpose: "this byte is wrong" and "this
	// construct never finished" are different failures, and a consumer that
	// cannot tell them apart reports the wrong thing to the user.
	UnknownRest

	// Quote is one whole quote-like operator: the keyword if any, every
	// delimiter, both bodies of a three-part form, and the trailing
	// modifiers.
	//
	// One token rather than several because the parts are not independently
	// meaningful -- `s/a/b/` and `s/a/b/g` differ, and a consumer that has to
	// re-lex the following token to learn which it got has the wrong
	// boundaries. Interpolation splits this later, when there is a parser to
	// consume the pieces.
	Quote
)

func (k Kind) String() string {
	switch k {
	case Whitespace:
		return "Whitespace"
	case Error:
		return "Error"
	case UnknownRest:
		return "UnknownRest"
	case Quote:
		return "Quote"
	}
	return "Kind(?)"
}

// Token is one lexeme, as a half-open byte range into the source.
//
// Byte offsets rather than line/column: line/column is a presentation concern
// the LSP layer derives on demand, and carrying it here doubles the state that
// has to stay consistent under edit. The source is not held either -- the
// caller already has it, and a token that owns a string copy makes an
// incremental re-lex allocate per token.
type Token struct {
	Kind       Kind
	Start, End int
}

// Tokenize splits src into tokens covering every byte exactly once.
//
// It has no error return by design. A byte that cannot be lexed becomes an
// Error token and lexing continues, because an LSP sees half-typed buffers
// constantly and a lexer that gives up on the first bad byte is useless to
// it. Spec §7.6.2 invariant 1: never panic, on any input.
func Tokenize(src []byte) []Token {
	l := &lexer{src: src}
	for l.pos < len(l.src) {
		l.step(scanOne)
	}
	return l.toks
}

// lexer is the cursor and the output. Everything that will later need lexer
// state -- the expect bit, the heredoc queue, the utf8 flag -- hangs here.
type lexer struct {
	src  []byte
	pos  int
	toks []Token
}

// step runs one scan and enforces spec §7.6.2 invariant 4: every step
// consumes at least one byte.
//
// The guard is structural rather than a timeout because `go test -fuzz` does
// not detect infinite loops -- it hangs, and a hung fuzzer reports nothing at
// all. A Perl lexer reaches this state by ordinary means: an unterminated
// quote-like operator whose scan returns without finding its closer is the
// classic case, and it is a bug the fuzzer would otherwise find by stalling
// for an hour.
//
// Panicking rather than returning an error is deliberate. This cannot happen
// through any input -- it is a defect in a scan function, caught the moment it
// is introduced rather than after it ships.
func (l *lexer) step(scan func(*lexer)) {
	before := l.pos
	scan(l)
	if l.pos <= before {
		panic("lexer: forward progress violated: a scan step consumed no bytes at offset " +
			itoa(before))
	}
}

// scanOne consumes one token at the cursor.
//
// This is the dispatch point every later issue extends: identifiers,
// quote-like operators, heredocs, POD. Today it knows whitespace and nothing
// else, so anything that is not whitespace is one Error byte.
func scanOne(l *lexer) {
	start := l.pos
	if isSpace(l.src[l.pos]) {
		for l.pos < len(l.src) && isSpace(l.src[l.pos]) {
			l.pos++
		}
		l.emit(Whitespace, start)
		return
	}
	if scanQuoteLike(l) {
		return
	}
	// Not yet lexable. One byte, so the cursor always advances and the rest
	// of the file is still lexed.
	l.pos++
	l.emit(Error, start)
}

// emit appends a token spanning start..l.pos.
func (l *lexer) emit(k Kind, start int) {
	l.toks = append(l.toks, Token{Kind: k, Start: start, End: l.pos})
}

// isSpace is Perl's whitespace for lexing purposes. Vertical tab and form
// feed are included: perl's isSPACE covers them, and a file using form feed
// as a page separator is real Perl that must still round-trip.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}

// itoa avoids importing strconv into the panic path, which keeps the hot
// lexing loop free of the dependency.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
