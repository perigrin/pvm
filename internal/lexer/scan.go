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

	// A POSTFIX SLICE: `$r->@[0,1]`, `$r->%{'a'}`. The sigil is the whole
	// token and the bracket after it is an ordinary opener, so it reaches
	// the bracket stack.
	//
	// Without this the sigil scanner took `@[` as a punctuation variable
	// named `[`. No opener was ever pushed, and the matching `]` then popped
	// whatever was underneath -- corrupting the stack for the rest of the
	// file, which is why a brace much later in postderef.t was misread.
	//
	// `@*` and `%*` are NOT this case: there the star IS the whole
	// dereference and one Variable token is right. Measured:
	//
	//	$ perl -MO=Deparse -e 'my $r=[1,2,3]; my @s = $r->@[0,1]; my @t = $r->@*;'
	//	my(@s) = @$r[0, 1];
	//	my(@t) = @$r;
	if l.expect == XPostDeref && (l.src[l.pos] == '[' || l.src[l.pos] == '{') {
		l.emit(Variable, start)
		return true
	}

	// A sigil applied to an EXPRESSION is its own token, and the expression
	// after it is lexed normally so the parser can read it.
	//
	// `${EXPR}` and `$$x` are the two shapes. Consuming them here produced
	// one Variable token spanning the whole form, and the parser then made a
	// childless leaf of it -- no interior, so nothing inside was reachable.
	if l.startsDerefExpression() {
		l.emit(DerefSigil, start)
		return true
	}

	l.scanVarName()
	l.emit(Variable, start)
	return true
}

// startsDerefExpression reports whether the sigil at l.pos-1 is applied to an
// expression rather than to a name. l.pos is just past the sigil.
//
// Two forms qualify, and each has a near neighbour that does NOT:
//
//	${ $h->{k} }   expression      vs  ${name}  a NAME in braces
//	$$x            deref of $x     vs  $$       the process id
//
// The brace case is decided by what follows it: an identifier that runs to
// the closing brace is a name, anything else is an expression. The sigil case
// is decided by what follows the SECOND sigil: a name makes it a dereference,
// nothing makes it the punctuation variable `$$`.
func (l *lexer) startsDerefExpression() bool {
	if l.pos >= len(l.src) {
		return false
	}

	// Whitespace may stand between the sigil and the brace. `$ {*$glob}{Keys}`
	// is a real form and appears in the corpus; checking the byte immediately
	// after the sigil missed it, and the sigil then fell back to the
	// swallowing path it was meant to replace. Measured on perl 5.42.0:
	//
	//	$ perl -e 'use Symbol qw(gensym); my $g = gensym();
	//	           $ {*$g}{K} = 5; print $ {*$g}{K};'
	//	5
	brace := l.pos
	for brace < len(l.src) && (l.src[brace] == ' ' || l.src[brace] == '\t') {
		brace++
	}
	if brace < len(l.src) && l.src[brace] == '{' {
		return !l.bracedNameFollowsAt(brace)
	}

	// `$$x` and `@$x` are dereferences; a bare `$$` is the pid, and
	// `$$ref[0]` is still a dereference of `$ref`. A second sigil with a
	// NAME after it is the whole rule -- with nothing after it there is
	// nothing to dereference.
	if l.src[l.pos] != '$' {
		return false
	}
	next := l.pos + 1
	return next < len(l.src) &&
		(isWordByte(l.src[next]) || l.src[next] == '{' || l.src[next] == '$')
}

// bracedNameFollowsAt reports whether `{` at brace opens a plain NAME, as in
// `${foo}` -- an identifier, optionally spaced, then the closing brace.
//
// Anything else is an expression: `${$x}`, `${ $h->{k} }`, `@{[ 1, 2 ]}`.
func (l *lexer) bracedNameFollowsAt(brace int) bool {
	i := brace + 1
	for i < len(l.src) && (l.src[i] == ' ' || l.src[i] == '\t') {
		i++
	}

	// `${^TEST}`, `${^TAINT}`, `${^UNICODE}` -- the caret control variables
	// are NAMES, and the caret is part of the name rather than an operator.
	// Splitting one into a sigil and an expression made `^TEST` a term the
	// parser had to guess at, which is what TestLexDotTGoldenStream caught.
	if i < len(l.src) && l.src[i] == '^' {
		i++
	}

	nameStart := i
	for i < len(l.src) && isWordByte(l.src[i]) {
		i++
	}
	if i == nameStart {
		return false
	}
	for i < len(l.src) && (l.src[i] == ' ' || l.src[i] == '\t') {
		i++
	}
	return i < len(l.src) && l.src[i] == '}'
}

// scanVarName consumes an identifier, a punctuation variable, or a braced
// name after a sigil.
func (l *lexer) scanVarName() {
	if l.pos >= len(l.src) {
		return
	}
	switch c := l.src[l.pos]; {
	case c == '\'':
		// `$'` is the postmatch variable: the apostrophe is the whole name.
		// A separator needs an identifier character BEFORE it, and there is
		// none here.
		l.pos++
	case c == ':' && l.leadingPackageSeparator():
		// `$::x` is `$main::x`. scanIdentRunes requires a leading identifier
		// byte and `:` is not one, so this lexed as `Variable("$:")` -- a
		// punctuation variable named `:` -- followed by stray tokens. A
		// plausible wrong answer rather than an error, which is why nothing
		// caught it.
		//
		// Measured: `our $x = 5; print $::x` prints 5.
		//
		// `$:` ALONE is a real punctuation variable, the format line-break
		// set, so the two colons and a name after them are all required --
		// see leadingPackageSeparator.
		l.pos += 2
		l.scanIdentRunes()
	case l.scanIdentRunes():
		// Consumed by the identifier scanner, which handles both package
		// separators and the utf8-widened class.
	case c == '{':
		// A braced NAME: `${name}`, `${^TAINT}`. Reached only when
		// startsDerefExpression has already declined, so the braces here are
		// punctuation around an identifier and the match can be shallow.
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

// leadingPackageSeparator reports whether `::` at the cursor introduces a
// package-qualified name, as the `$::x` shorthand for `$main::x`.
//
// Both colons AND a name are required. `$:` is the format line-break
// variable and `$::` with nothing after it is not a name perl accepts, so
// taking either would trade one wrong answer for another.
func (l *lexer) leadingPackageSeparator() bool {
	return l.pos+2 < len(l.src) &&
		l.src[l.pos+1] == ':' &&
		isWordByte(l.src[l.pos+2])
}

// scanWord lexes a bareword or keyword.
//
// Whether a word is a keyword, a function name or a bareword string cannot be
// decided here -- it needs the symbol table and the parser -- so the lexer
// reports Word and leaves the classification alone.
func scanWord(l *lexer) bool {
	start := l.pos
	if !l.scanIdentRunes() {
		return false
	}
	// `use utf8` widens the class for everything after it, so the pragma has
	// to be noticed as it is lexed rather than in a prepass.
	l.notePragma(start)
	l.emit(Word, start)
	return true
}

// scanVString lexes the `v`-prefixed spelling of a v-string: `v65.66.67`.
//
// It runs BEFORE scanWord, because `v65` is a legal identifier and the word
// scanner claimed it -- leaving `Word(v65) Operator(.) Number(66.67)`, a
// concatenation of a bareword with a number where perl has one string. The
// parser cannot see that: it receives a valid expression and returns no
// Unknown node.
//
// TWO DOTS are still required, which is what keeps `use v5.36` lexing as
// the three tokens `internal/parse/use.go` reassembles a version from.
// Perl calls the one-dot `v5.36` a v-string too -- measured, its length is
// 2 -- and this lexer does not yet, which is a separate gap from the one
// the corpus asserts here.
func scanVString(l *lexer) bool {
	if c := l.src[l.pos]; c != 'v' && c != 'V' {
		return false
	}
	if !l.expect.wantsTerm() {
		return false
	}
	start := l.pos
	l.pos++
	if l.pos >= len(l.src) || !isDigit(l.src[l.pos]) {
		l.pos = start
		return false
	}
	if l.scanNumberRun() < 2 {
		l.pos = start
		return false
	}
	l.emit(Quote, start)
	return true
}

// isDigit reports whether c is an ASCII decimal digit.
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// startsLeadingDecimal reports whether the cursor sits on a `.` that
// begins a numeric literal written with no digit before the point.
//
// Two conditions, and POSITION is the load-bearing one. A digit must
// follow, so `.` alone and `..` are untouched -- but a digit following is
// not sufficient, because `$a .5` is concatenation. Measured 5.42.0, the
// same four bytes parse two ways with nothing but context between them:
//
//	my $r = $a .5;     ->  my $r = $a . '5';    concatenation
//	print $a .5        ->  print $a 0.5;       filehandle and a number
//
// The second prints nothing, because `$a` lands in print's filehandle
// slot. So the rule is the expect state, the same mechanism §3.2 uses for
// a leading `%`, `<`, `&` and `/`: where a TERM is expected nothing else
// can start with a dot, and where an OPERATOR is expected a dot is
// concatenation regardless of what follows.
// TWO version-string exceptions, both measured rather than defensive,
// and both found by a test rather than anticipated.
//
// `use v5.36` reaches here with a term expected and a digit following,
// because `v5` lexes as an ordinary Word -- so without the first guard
// the `.36` becomes one Number and `pendingVersionMajor` never sees the
// Operator it reassembles the version from. The feature bundle then does
// not turn on, which silently changes what the rest of the file means:
// `use v5.36` enables signatures, and `sub g ($a, $b)` is a PROTOTYPE
// without it.
//
// `require(v5.5.630)` needs the second. `pendingVersionMajor` clears on
// the Number it pairs with, so by the third part it is already zero and
// the first guard no longer applies. What rules it out is that the
// previous token is a Number with no space before the dot: a v-string's
// parts are adjacent, and a leading decimal never follows a number
// directly. `(1,.5)` is a real leading decimal and its previous token is
// the comma, not the 1.
func startsLeadingDecimal(l *lexer) bool {
	return l.src[l.pos] == '.' &&
		l.expect.wantsTerm() &&
		!continuesVersionString(l) &&
		l.pos+1 < len(l.src) &&
		isDigit(l.src[l.pos+1])
}

// continuesVersionString reports whether the dot at the cursor separates
// two parts of a v-string such as `v5.36` or `v5.5.630`.
//
// ADJACENCY is the whole rule: a v-string's parts touch, so the token
// before the dot ends exactly where the dot begins. `(1,.5)` is a real
// leading decimal and the token before its dot is the comma; `5 . .5`
// has a space. Neither is adjacent to a digit-bearing token.
//
// Both the `v5` Word and a `Number` count, because the two dots of
// `v5.5.630` have different predecessors. Checking `pendingVersionMajor`
// instead would miss every v-string outside a `use`: `noteSignatures`
// returns early unless a `use` is pending, so `require(v5.5.630)` never
// sets it.
func continuesVersionString(l *lexer) bool {
	if len(l.toks) == 0 {
		return false
	}
	prev := l.toks[len(l.toks)-1]
	if prev.End != l.pos {
		return false
	}
	if prev.Kind == Number {
		return true
	}
	if prev.Kind != Word {
		return false
	}
	_, ok := versionPrefix(string(l.src[prev.Start:prev.End]))
	return ok
}

// scanNumber lexes an integer or float literal, or the v-string that a
// second decimal point turns the same digits into. The exotic forms --
// `0x_1234`, `0x0p0` -- are §0.13 rank 9 and belong to a later issue; this
// covers what the M0 corpus reaches.
//
// DOT COUNTING is what separates the two kinds, and it cannot be decided
// before the run has been scanned: `65.66` is a float and `65.66.67` is the
// three-character string `ABC`, and they are the same bytes up to the second
// point. Measured 5.42.0:
//
//	$ perl -e 'my $n = 65.66; print $n'          65.66
//	$ perl -e 'my $v = 65.66.67; print $v'       ABC
//	$ perl -e 'my $x = 5.42.0; print $x+0'       0
//
// The last is the category claim: perl holds a string of ordinals, so `+0`
// is 0 rather than 5.42. GLOSSARY.md records the decision under "numeric
// literal" -- a v-string is not in that category.
func scanNumber(l *lexer) bool {
	c := l.src[l.pos]
	if !isDigit(c) && !startsLeadingDecimal(l) {
		return false
	}
	start := l.pos
	dots := l.scanNumberRun()
	// One dot is a float; two make a v-string, which this lexer spells as a
	// Quote because that is its kind for a string. `internal/conformance/
	// categories.go` reads a Quote with no operator name as the glossary's
	// "string literal", and `v` is not a quote-operator name.
	if dots >= 2 {
		l.emit(Quote, start)
		return true
	}
	l.emit(Number, start)
	return true
}

// scanNumberRun advances the cursor over one numeric run and reports how
// many decimal points it contained.
//
// Shared with scanVString, which differs only in the `v` it consumes first.
func (l *lexer) scanNumberRun() int {
	dots := 0
	// A RADIX PREFIX rules the exponent sign out, because `e` is a hex
	// digit there rather than an exponent marker. Measured 5.42.0:
	// `print 0x1e-1` is 29, so that `-` is subtraction.
	hex := l.pos+1 < len(l.src) && l.src[l.pos] == '0' &&
		(l.src[l.pos+1] == 'x' || l.src[l.pos+1] == 'X')
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case c == '.':
			// A DOUBLE dot is the range operator, not part of the literal.
			// `1..5` is three tokens; consuming the dots gave one
			// Number("1..5") and the range was gone before the parser saw
			// it.
			//
			// `$a..$b` was never affected, so only literal endpoints broke
			// -- and the cost is measured: `my @r = (1..5)` infers @r as
			// Str, because `internal/infer/infer.go:645-659` matches the
			// anonymous `.` before ever seeing `..` (§4.14.2).
			//
			// One dot still belongs to the number: `1.5` is a float, and
			// perl takes `1_000.5` and `1.5e10` whole.
			if l.pos+1 < len(l.src) && l.src[l.pos+1] == '.' {
				return dots
			}
			dots++
		case isWordByte(c):
			// The SIGN OF AN EXPONENT is part of the literal, and this is
			// the only place a sign ever is. Measured 5.42.0:
			//
			//	$ perl -e 'print 5e-1'   0.5
			//	$ perl -e 'print 5e'     Bareword found where operator
			//	                         expected (Missing operator
			//	                         before "e"?)
			//
			// So `Number("5e")` is a token perl would reject, which is
			// what makes the old split wrong rather than merely different.
			// The sign must be ADJACENT to the `e`: `5-1` is arithmetic,
			// and so is the second `-` of `5e-1-1`.
			if !hex && (c == 'e' || c == 'E') &&
				l.pos+2 < len(l.src) &&
				(l.src[l.pos+1] == '-' || l.src[l.pos+1] == '+') &&
				isDigit(l.src[l.pos+2]) {
				l.pos += 2
				continue
			}
		default:
			return dots
		}
		l.pos++
	}
	return dots
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
	// `<<>>` is the 5.22 double diamond, which reads every argument as a
	// filename and never as a command. Checked first because its inner `<>`
	// would otherwise close the scan at the second `<`.
	if l.pos+3 < len(l.src) && l.src[l.pos+1] == '<' &&
		l.src[l.pos+2] == '>' && l.src[l.pos+3] == '>' {
		start := l.pos
		l.pos += 4
		l.emit(Readline, start)
		return true
	}

	// A readline's contents are a filehandle name, a scalar, or nothing --
	// and a GLOB's are a shell pattern, which is neither. Measured:
	//
	//	$ perl -MO=Deparse -e 'my @g = <*.c>; my @h = <~/x>;'
	//	my(@g) = glob('*.c');
	//	my(@h) = glob('~/x');
	//
	// Both are the same token to the lexer; which one it is depends on the
	// contents, and the parser can read those. What matters here is only
	// where the form ENDS.
	//
	// A newline ends the candidate: `$a < $b` on one line and `$c > $d` on
	// the next must not join into one token across them. Everything else is
	// permitted, because perl accepts `<a b>` -- a glob pattern with a space
	// in it -- as readily as `<*.c>`.
	j := l.pos + 1
	for j < len(l.src) && l.src[j] != '>' && l.src[j] != '\n' {
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
