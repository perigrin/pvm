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

	// Comment is `#` to end of line. Trivia, like Whitespace: it carries no
	// meaning but it carries bytes, and every byte belongs to a token.
	Comment

	// Variable is a sigil and its name: `$x`, `@a`, `%h`, `$#x`, `$#`.
	Variable

	// Word is a bareword or keyword. Which of those it is cannot be decided
	// by the lexer alone, so the distinction is left to the parser.
	Word

	// Number is an integer or float literal.
	Number

	// Operator is punctuation between terms.
	Operator

	// Semicolon ends a statement, returning the machine to XSTATE.
	Semicolon

	// Readline is `<FH>` or `<$fh>` in term position. Distinct from a pair of
	// comparison operators, which is the whole point.
	Readline

	// FuncSigil is `&` introducing a function name in term position, as
	// opposed to `&&` or bitwise-and in operator position.
	FuncSigil

	// DerefSigil is a sigil applied to an EXPRESSION rather than to a name:
	// the `$` of `${$h->{k}}` and of `$$x`, the `@` of `@{[ ... ]}`.
	//
	// `$x` is one Variable token because its name is part of the lexeme.
	// `${EXPR}` is not: the sigil and the expression are separate, and the
	// expression needs the parser. Emitting one Variable token for the whole
	// form -- which is what brace-matching here used to do -- put every name
	// inside it outside the tree, where rename, go-to-definition and
	// inference cannot see them. Round-trip did not notice, because the span
	// still covered every byte (chapter 7 §7.2(c)).
	//
	// `${name}` stays a Variable: the braces are punctuation around a NAME
	// there, which is how `"${foo}bar"` is written, and nothing inside needs
	// parsing. What follows the brace decides which form this is.
	DerefSigil

	// Prototype is a balanced `(...)` directly after `sub NAME`, scanned as
	// an opaque string rather than lexed.
	//
	// Its contents are sigils that are not variables: `($$)` is two scalars
	// in a prototype and `$$` is the process id everywhere else. perl scans
	// it with scan_str (toke.c:5923), and spec §5.5.2 says the same.
	//
	// RECOGNISING one is this milestone's; what a prototype DOES to a call
	// is M4's, and until then a prototyped call stays an honest Unresolved.
	Prototype

	// CloseBracket is `)`, `]` or `}`.
	//
	// Distinct from Operator because it moves the expect state the OTHER
	// way: every other operator opens a term, while a closing bracket ends
	// one. Without the distinction `$x[0] <FH>` lexes its angle brackets as
	// a readline, since the `]` would leave a term expected.
	CloseBracket

	// HeredocOpen is the `<<TERM` marker, without the body.
	HeredocOpen

	// HeredocBody is the body and its terminator line, which begin after the
	// line the marker sits on. Separate from the open because everything
	// between them is ordinary code: `print <<A, <<B;` has two markers, a
	// comma and a semicolon before either body starts.
	HeredocBody

	// Pod is a documentation block, `=word` through `=cut` inclusive. Trivia
	// with a line-oriented terminator rather than a delimiter.
	Pod

	// DataSection is `__END__` or `__DATA__` and everything after it. The
	// contents are data, not Perl, so nothing inside is lexed as code.
	DataSection

	// FormatBody is the picture lines of a `format NAME =` declaration,
	// through the lone `.` that ends them. Not Perl: `@<<<<<` is a picture
	// field, not an array sigil and two left shifts.
	FormatBody
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
	case Comment:
		return "Comment"
	case Variable:
		return "Variable"
	case Word:
		return "Word"
	case Number:
		return "Number"
	case Operator:
		return "Operator"
	case Semicolon:
		return "Semicolon"
	case Readline:
		return "Readline"
	case FuncSigil:
		return "FuncSigil"
	case DerefSigil:
		return "DerefSigil"
	case Prototype:
		return "Prototype"
	case CloseBracket:
		return "CloseBracket"
	case HeredocOpen:
		return "HeredocOpen"
	case HeredocBody:
		return "HeredocBody"
	case Pod:
		return "Pod"
	case DataSection:
		return "DataSection"
	case FormatBody:
		return "FormatBody"
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

	// OpensBlock is set on a `{` that opened a block rather than a subscript
	// or an anonymous hash, and on the `}` that closed one.
	//
	// The lexer already decides this -- the brace stack has to, since a `}`
	// leaves a different expect state depending on what it closed -- and
	// without it on the token the decision is computed and discarded, so the
	// parser re-derives what the lexer knew. That is the duplication the
	// stack exists to prevent, one layer up.
	//
	// It is deliberately NOT the full expect state. A parser that could read
	// PL_expect at every token would start depending on states the lexer may
	// refine later; one bit answering one question is a narrower contract.
	OpensBlock bool
}

// Tokenize splits src into tokens covering every byte exactly once.
//
// It has no error return by design. A byte that cannot be lexed becomes an
// Error token and lexing continues, because an LSP sees half-typed buffers
// constantly and a lexer that gives up on the first bad byte is useless to
// it. Spec §7.6.2 invariant 1: never panic, on any input.
func Tokenize(src []byte) []Token {
	return TokenizeBarewords(src, BarewordOperators)
}

// BarewordOperators are the operators whose left operand Tokenize and
// TokenizeTyped read as a word: CORE.pmt's `:bareword` ones, which a
// caller that has read CORE.pmt passes to TokenizeBarewords. These are
// what CORE.pmt's own read is lexed with, and parse's
// TestLexerBarewordOperatorsAreCores holds them to CORE.pmt's.
var BarewordOperators = []string{"=>"}

// TokenizeBarewords is Tokenize with the operators whose left operand is
// read as a word, `s => 1` quoting `s` rather than opening a substitution.
func TokenizeBarewords(src []byte, barewordOps []string) []Token {
	return tokenize(&lexer{src: src, expect: XState, barewordOps: barewordOps})
}

// TokenizeTyped is Tokenize for typed Perl, the language of a `.pmt`
// declaration file (RFC 0001, "Typed Perl, in `.pmt` only"). The signatures
// feature is on from the first byte, so the `(` after `sub NAME` opens a
// signature and never a prototype: a declaration file spells its prototypes
// `:prototype(...)`, as perl requires once signatures are on.
//
// An operator is declared under its symbol, `sub + :infix(ADD) ...`, so the
// symbol after `sub` is read as the sub's name; see scanOperatorName.
func TokenizeTyped(src []byte) []Token {
	return tokenize(&lexer{src: src, expect: XState, signatures: true, typed: true, barewordOps: BarewordOperators})
}

func tokenize(l *lexer) []Token {
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
	// barewordOps are the operators whose left operand is read as a word.
	barewordOps []string
	// expect is the term-vs-operator state. A file starts at a statement
	// boundary, which is where POD and labels are recognised.
	expect Expect
	// utf8Pragma widens the identifier class to XID. File-level rather than
	// lexically scoped: spec §2.3.1 allows it as a v1. A brace stack now
	// exists (see brackets), so scoping it is reachable, but that is a
	// behaviour change with its own corpus effect and belongs in its own
	// commit rather than riding along with this one.
	utf8Pragma bool
	// brackets records what each open bracket was, so a closer knows what it
	// closed. perl's PL_lex_brackstack. Only `}` is genuinely ambiguous --
	// block or subscript -- but all three are tracked so the stack stays
	// aligned.
	brackets []bracket
	// closedBlock is set by trackBrackets when the token just emitted was a
	// `}` that closed a block, and read by the expect transition. A field
	// rather than a parameter threaded through every scanner, because emit
	// is the single point both go through.
	closedBlock bool
	// openedBlock is the same for a `{` that opened one. Both are copied
	// onto the token as OpensBlock, so the parser does not re-derive them.
	openedBlock bool
	// listOpParen is set when the token just emitted was `map`, `grep` or
	// `sort` with a `(` next, so that `(` can run intuitCurly for the brace
	// it now precedes. A one-token carry, the same shape as sawSubWord and
	// for the same reason: the lookahead that settles the brace is the WORD's
	// to make, and a parenthesised call moves the brace out of its reach.
	listOpParen bool
	// sawSubWord and expectPrototype track `sub NAME`, after which a `(`
	// opens a prototype rather than a list.
	sawSubWord      bool
	expectPrototype bool
	// sigPending and sigDepth track a SIGNATURE, the `(...)` after `sub
	// NAME` when the feature is on. sigPending is set where scanPrototype
	// declines the parens, and the `(` that follows makes sigDepth 1; it
	// counts nested parens in default expressions and is 0 outside. At
	// depth 1 a sigil at an element's start may stand alone -- see
	// bareSignatureSigil.
	sigPending bool
	sigDepth   int
	// globStar is set when the token just emitted was a `*` where no
	// operator was expected -- a glob sigil, not multiplication. toke.c's
	// yyl_star then reads the name with scan_ident, so a punctuation byte
	// after it is the name: `*@` is the glob named `@`.
	globStar bool
	// sawPackageWord is the same for `package NAME` and `class NAME`, which
	// are followed by a block or a semicolon but never by a prototype.
	sawPackageWord bool
	// classSyntax is set once `use feature 'class'` or a `class NAME`
	// declaration is seen: from then `method` declares as `sub` does.
	// Elsewhere it is an ordinary name -- op/method.t calls a sub named
	// method, `method Pack ("a")`.
	classSyntax bool
	// lexSubs are the lexical subs in scope, each with the bracket depth it
	// was declared at: `my sub s { 42 }` makes `s(1)` a call until that
	// block closes. See lexSubInScope.
	lexSubs []lexSub
	// sawPackageName is set on that NAME, so a version after it --
	// `package Foo 1.0 { }` -- keeps the head open for the block.
	sawPackageName bool
	// inSubAttrs is whether an ATTRIBUTE LIST may start or continue here --
	// after `sub`, after a sub's name, after its prototype, and after each
	// attribute already read. A block still follows the list, so this keeps
	// the block expectation alive across the colons: without it
	// `sub f :lvalue { 1 }` lexed its `{` as an anonymous hash's and the
	// whole body was lost.
	inSubAttrs bool
	// sawAttrColon is set when the token just emitted was an attribute's `:`,
	// so the Word after it is the attribute's NAME rather than a bareword. A
	// one-token carry, the same shape as sawSubWord, and it exists for
	// `:prototype(...)`: that argument is a prototype and must be scanned
	// opaquely, because `$)` is a real perl variable that would otherwise eat
	// the closing paren.
	sawAttrColon bool
	// attrArgDepth is the paren depth inside an attribute's argument,
	// `:Foo(bar)`, whose contents the parser reads opaquely. Nothing in it
	// closes the declaration's head, so an attribute after it is still an
	// attribute: `sub f :Foo(bar) :prototype($) { 1 }`.
	attrArgDepth int
	// sawLabelWord is set when the token just emitted was a Word at a
	// STATEMENT boundary, so a `:` following it is a label's colon rather
	// than a ternary's or an attribute's. A one-token carry, the same shape
	// as listOpParen and sawSubWord, and for the same reason: the `:` cannot
	// see where its word stood, because the word left XTerm behind it.
	//
	// The position is the whole discrimination. Measured on the four shapes
	// that put a bareword next to a colon:
	//
	//	SKIP: { ... }        Word at XSTATE   a label
	//	$c ? 1 : 0           `:` at XOPERATOR after a Number
	//	$h{LOOP}             Word at XTERM    a hash key
	//	sub f :lvalue { }    Word at XTERM    an attribute
	//
	// Only the first has a word at a statement boundary, which is the same
	// question parseLabels answers one layer up.
	sawLabelWord bool
	// signatures is whether the signatures feature is on, which decides
	// whether that `(` is a prototype or a signature. File-level, like
	// utf8Pragma and for the same reason.
	signatures bool
	// typed is set for typed Perl, a `.pmt` declaration file, where a
	// container type may stand before a hash parameter: `List[Str] %h`
	// (see containerTypeSigil), and where `sub +` names an operator (see
	// scanOperatorName).
	typed bool
	// pendingVersionMajor held the `5` of a `use v5.36` whose version
	// arrived split across three tokens -- Word("v5"), Operator("."),
	// Number(36) -- because `v5` lexed as an ordinary identifier.
	// `scanVString` now emits one Quote for that spelling, so nothing sets
	// this any more. Tracked by 01a0dc74.
	pendingVersionMajor int
	// pendingPragma remembers a `use` or `no` seen on this statement, so
	// that the `utf8` after it can be recognised. 0 none, 1 use, 2 no.
	pendingPragma int
	// pending holds heredocs whose terminator has been read but whose body
	// has not started. Drained at the next newline, in order.
	pending []pendingHeredoc
	// sawFormatWord and pendingFormat track a `format NAME =` declaration,
	// whose picture body starts on the line after the `=`.
	sawFormatWord bool
	pendingFormat bool
	// inFormat is set when the picture body should be taken next.
	inFormat bool
	// constSubs are the names declared with the empty prototype -- `use
	// constant NAME`, the keys of `use constant {...}`, `sub NAME () {...}`.
	// A value has been produced after one, so an operator follows, as perl's
	// lexer decides from the symbol table. constState and the rest track the
	// declaration being read; see noteConstSub.
	//
	// ponytail: one table for the file; perl scopes these to a package.
	constSubs  map[string]bool
	constState int
	constDepth int
	constKey   string
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
		// With a heredoc queued, stop at the first newline: its body starts
		// immediately after that newline, and a whitespace token that ran
		// past it would put the body in the wrong place.
		//
		// With nothing queued, consume the whole run. Splitting
		// unconditionally would break the "whitespace-only input is ONE
		// trivia token" rule from the skeleton issue -- caught by that
		// issue's own test, which is why it is there.
		stopAtNewline := len(l.pending) > 0
		sawNewline := false
		for l.pos < len(l.src) && isSpace(l.src[l.pos]) {
			if l.src[l.pos] == '\n' && stopAtNewline {
				l.pos++
				sawNewline = true
				break
			}
			l.pos++
		}
		l.emit(Whitespace, start)
		if sawNewline {
			l.takePendingHeredocs()
		}
		// A format body, like a heredoc body, starts after the line ends.
		// NOT wrapped in step(): the scanner declines at EOF -- a `format =`
		// with nothing after it has no body -- and step() requires forward
		// progress, so wrapping it turns a legitimate decline into a panic.
		// The fuzzer found that on its first run, from `format =\n`.
		if l.inFormat {
			scanFormatBody(l)
		}
		return
	}
	// ORDER MATTERS. Each scanner below either claims the cursor or declines,
	// and several decline on the expect state alone:
	//
	//   scanComment    before the quote scanner, so `#` is not a delimiter
	//   scanVariable   before words, so `$x` is not the word `x`
	//   scanAngle      before operators, so `<FH>` is not `<` `FH` `>`
	//   scanAmp        before operators, so `&foo` is not bitwise-and
	//   scanQuoteLike  before words, so `q` is not the bareword `q`
	//                  -- but it declines in operator position for `x`,
	//                  `y` and `s`, which are repetition and names there
	//   scanVString    before words, so `v65.66.67` is not `v65` `.` `66.67`
	//                  -- but it declines with NO dot, so `v5` stays a Word,
	//                  and on capital V, which perl never reads as a version
	//   scanNumber     before operators, so `1.5` is not `1` `.` `5`
	//   scanFileTest   before operators, so `-e` is not `-` `e`
	for _, scan := range []func(*lexer) bool{
		scanPod,
		scanDataSection,
		scanComment,
		// Before every scanner that reads punctuation as something else:
		// `sub <<`, `sub /` and `sub %` name operators, not a heredoc, a
		// pattern and a hash.
		scanOperatorName,
		scanHeredocOpen,
		// Before scanVariable: the sigils inside a prototype are not
		// variables, and scanVariable is what was reading `($$)` as
		// Variable("$$)") -- closing paren included.
		scanPrototype,
		scanPostDerefStar,
		scanVariable,
		scanAngle,
		scanAmp,
		scanQuoteLike,
		scanBarePattern,
		scanVString,
		scanWord,
		scanNumber,
		scanSemicolon,
		scanFileTest,
		scanOperator,
	} {
		if scan(l) {
			return
		}
	}
	// Not yet lexable. One byte, so the cursor always advances and the rest
	// of the file is still lexed.
	l.pos++
	l.emit(Error, start)
}

// emit appends a token spanning start..l.pos and advances the expect state.
//
// The state transition lives here rather than at each call site so that a new
// scanner cannot forget it. Forgetting would not fail loudly -- it would make
// the NEXT token wrong, somewhere else in the file, which is the hardest kind
// of lexer bug to find.
func (l *lexer) emit(k Kind, start int) {
	l.toks = append(l.toks, Token{Kind: k, Start: start, End: l.pos})
	l.trackBrackets(k, start)
	// trackBrackets has just classified this token, so the flags it set
	// belong to the token appended above.
	l.toks[len(l.toks)-1].OpensBlock = l.openedBlock || l.closedBlock

	// Before the expect transition: noteSubName decides whether this token
	// is the NAME of a `sub NAME`, and the transition needs that answer to
	// know a block comes next rather than a term.
	//
	// noteSignatures runs first because the version or feature name arrives
	// one token after the `use` that set pendingPragma, and noteSubName does
	// not touch that state.
	l.noteSignatures(k, start)
	l.noteSignatureParens(k, start)
	afterDeclName := l.noteSubName(k, start)

	// A block-taking word's `(` inherits the lookahead. `map({...} @a)` puts
	// the brace one token further along than `map {...} @a`, and the answer
	// belongs to the same call either way; see listOpParen.
	inListOpParen := l.listOpParen && k == Operator && l.pos-start == 1 &&
		l.src[start] == '('
	// Trivia does not clear the carry: `map ({...} @a)` has a space between
	// the word and its paren, and the call is the same one.
	if k != Whitespace && k != Comment {
		l.listOpParen = k == Word && takesBlock(string(l.src[start:l.pos])) &&
			l.peekIsOpenParen()
	}

	// A label's `:`, carried from the Word before it. Read BEFORE the carry
	// is reset, for inListOpParen's reason.
	//
	// Exactly one byte, so `Foo::bar` -- whose separator the operator scanner
	// emits as one `::` token -- is a qualified name and not a label. That is
	// the same test parseLabels applies, and perl's own lexer applies it too.
	labelColon := l.sawLabelWord && k == Operator && l.pos-start == 1 &&
		l.src[start] == ':'
	// Set on a Word at a statement boundary. XState is where a statement may
	// start -- including the first statement INSIDE a block, since a block's
	// `{` leaves XState -- and XBlock is the state `if (...)`'s `)` leaves,
	// where the next thing is a block.
	//
	// A label STACKS, and `A: B: { ... }` needs the second word to qualify
	// as well. The `:` branch returns XState for exactly that, so the word
	// after it arrives at a statement boundary and this test holds again
	// without a second rule.
	//
	// Trivia does not clear it, for the reason the expect machine returns `e`
	// unchanged on Whitespace and Comment: `SKIP : { }` and a label with a
	// comment after it are the same label. Clearing on trivia would make the
	// rule hold only when the colon touches its word.
	if k != Whitespace && k != Comment {
		l.sawLabelWord = k == Word &&
			(l.expect == XState || l.expect == XBlock)
	}

	l.globStar = k == Operator && l.pos-start == 1 && l.src[start] == '*' &&
		l.expect != XOperator && l.expect != XPostDeref
	l.expect = l.expect.after(k, transition{
		text:            l.src[start:l.pos],
		closedBlock:     l.closedBlock,
		nextIsOpenBrace: l.peekIsOpenBrace(),
		afterDeclName:   afterDeclName,
		declaredNiladic: k == Word && l.constSubs[string(l.src[start:l.pos])],
		// Only a WORD can need the lookahead, or the `(` of a block-taking
		// word's parenthesised call. Every other token would pay a byte scan
		// for an answer nothing reads. Which word it is does not narrow this:
		// any of them may be followed by a block, so intuitCurly decides for
		// all of them -- see keyword.go.
		// A label's `:` does NOT need it, even though perl runs intuit_curly
		// after a label too. It leaves XState, where the brace is classified
		// the way a statement-start brace already is -- and the parser's
		// braceOpensAnonHash is what applies the lookahead there, for every
		// statement-start brace at once. See the labelColon branch in
		// expect.go.
		nextBraceIsBlock: (k == Word || inListOpParen) && l.intuitCurly(),
		listOpParen:      inListOpParen,
		labelColon:       labelColon,
		openedBlock:      l.openedBlock,
	})
	if k == Variable && l.handleSlotTakesTerm(start) {
		l.expect = XTerm
	}
	l.noteFormat(k, start)
	l.noteConstSub(k, start)
	// The picture body begins after the newline that ends the declaration.
	if l.pendingFormat && k == Whitespace && l.pos > start &&
		bytesContainNewline(l.src[start:l.pos]) {
		l.pendingFormat = false
		l.inFormat = true
	}
}

// handleTakers are the builtins whose first slot may hold a scalar
// filehandle or program with no comma after it: `print $fh LIST`,
// `system $prog LIST`.
//
// perl applies the heuristic below after ANY list operator (toke.c tests
// PL_last_lop), including a user's. The lexer cannot know a user's list
// operators, so it is limited to the builtins whose slot this is. A user sub
// taking a handle the same way is the ceiling, and the upgrade is the
// parser's declared-shape table.
var handleTakers = map[string]bool{
	"print": true, "printf": true, "say": true, "exec": true, "system": true,
}

// handleSlotTakesTerm reports whether the scalar just emitted at start fills a
// handleTakers slot, so a TERM follows it rather than an operator.
//
// It is toke.c's yyl_dollar heuristic for a scalar in a list operator's first
// slot followed by whitespace: perl peeks at the next character and sets
// XTERM when it can only start a term. Measured on 5.42.0 with -MO=Deparse,
// `print $f <<"EOT"` is a heredoc, `print $f /2/` a pattern and `print $f -1`
// a negative number, while `print $x - 1` and `print $x == 1` stay operators.
func (l *lexer) handleSlotTakesTerm(start int) bool {
	if l.src[start] != '$' || l.pos >= len(l.src) || !isSpace(l.src[l.pos]) {
		return false
	}
	prev := l.significantBefore(len(l.toks) - 1)
	// The parenthesised call counts too: toke.c compares PL_last_lop with the
	// token TWO back, and for `print($fh` that is still `print`. Measured,
	// `print($f <<"EOT")` is a heredoc on 5.42.0 -- and it is canon's own
	// spelling, so without this its emission did not re-parse.
	if prev >= 0 && l.src[l.toks[prev].Start] == '(' && l.toks[prev].End-l.toks[prev].Start == 1 {
		prev = l.significantBefore(prev)
	}
	if prev < 0 || l.toks[prev].Kind != Word ||
		!handleTakers[string(l.src[l.toks[prev].Start:l.toks[prev].End])] {
		return false
	}
	s := l.pos
	for s < len(l.src) && isSpace(l.src[s]) {
		s++
	}
	if s >= len(l.src) {
		return false
	}
	c := l.src[s]
	next := byte(0)
	if s+1 < len(l.src) {
		next = l.src[s+1]
	}
	isIDFirst := func(b byte) bool {
		return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= 0x80
	}
	switch {
	case byteIn(c, "$@\"'`q"):
		return true // print $fh "foo"
	case byteIn(c, "&*<%") && isIDFirst(next):
		return true // print $fh &sub
	case isIDFirst(c):
		e := s
		for e < len(l.src) && (isIDFirst(l.src[e]) || l.src[e] >= '0' && l.src[e] <= '9') {
			e++
		}
		// The binary operators spelled as words exclude the handle reading.
		switch string(l.src[s:e]) {
		case "x", "eq", "ne", "gt", "lt", "ge", "le", "cmp":
			return false
		}
		return true // print $fh length(), print $fh subr()
	case c >= '0' && c <= '9':
		return true // print $fh 3
	case c == '.' && next >= '0' && next <= '9':
		return true // print $fh .3
	case (c == '?' || c == '-' || c == '+') && next != 0 && !isSpace(next) && next != '=':
		return true // print $fh -1
	case c == '/' && next != 0 && !isSpace(next) && next != '=' && next != '/':
		return true // print $fh /.../
	case c == '<' && next == '<' && s+2 < len(l.src) && !isSpace(l.src[s+2]) && l.src[s+2] != '=':
		return true // print $fh <<"EOF"
	}
	return false
}

// significantBefore is the index of the last non-trivia token before i, or -1.
func (l *lexer) significantBefore(i int) int {
	for i--; i >= 0; i-- {
		if k := l.toks[i].Kind; k != Whitespace && k != Comment {
			return i
		}
	}
	return -1
}

// byteIn reports whether c is one of the bytes in set.
func byteIn(c byte, set string) bool {
	for i := 0; i < len(set); i++ {
		if set[i] == c {
			return true
		}
	}
	return false
}

// bytesContainNewline reports whether b holds a newline.
func bytesContainNewline(b []byte) bool {
	for _, c := range b {
		if c == '\n' {
			return true
		}
	}
	return false
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
