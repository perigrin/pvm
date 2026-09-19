// ABOUTME: The term-vs-operator state machine, all eleven states as perl.h defines them.
// ABOUTME: This is what makes / division or a pattern, < comparison or a readline, x an operator or a name.

package lexer

// Expect is where the lexer is in a statement: is the next thing a TERM
// (a value) or an OPERATOR (something between values)?
//
// Perl cannot be lexed without it. The same bytes mean different things on
// either side:
//
//	@a / 2      division      -- a term just closed
//	split /,/   a pattern     -- an operator just opened one
//	(1)x3       repetition    -- operator position
//	x3 => 1     an identifier -- term position
//
// All eleven of perl's states are modelled, in perl.h:5972-5985 order, even
// though this milestone exercises four of them. The spec is emphatic -- §2.15
// item 3 "Everything else depends on it", §3.9.4 "Do not economise here",
// §3.12 "all 11 states" -- and the reason is the same one that makes trivia
// tokens the right call: adding the missing states later means revisiting
// every transition, which is a rewrite rather than an addition.
type Expect int

const (
	// XOperator: a term just ended, so the next token is an operator.
	XOperator Expect = iota
	// XTerm: a value is expected.
	XTerm
	// XRef: a term is expected, and a bare `{` is a hash reference.
	XRef
	// XState: the start of a statement. POD recognition needs this, and so
	// does telling a label from an expression.
	XState
	// XBlock: a `{` here opens a block, never a hash constructor.
	XBlock
	// XAttrBlock: an attribute list, then a block -- `sub f :lvalue {`.
	XAttrBlock
	// XAttrTerm: an attribute list, then a term -- `my $x :shared = 1`.
	XAttrTerm
	// XTermBlock: a term, then a block.
	XTermBlock
	// XBlockTerm: a block, then a term.
	XBlockTerm
	// XPostDeref: after `->` in a postfix dereference, `->@*` and friends.
	XPostDeref
	// XTermOrDorDor: perl's own comment on this one is "evil hack". It
	// distinguishes `//` as defined-or from `//` as an empty pattern.
	XTermOrDorDor
)

// String names each state exactly as perl.h spells it, so the two lists can
// be diffed without translating between them.
func (e Expect) String() string {
	switch e {
	case XOperator:
		return "XOPERATOR"
	case XTerm:
		return "XTERM"
	case XRef:
		return "XREF"
	case XState:
		return "XSTATE"
	case XBlock:
		return "XBLOCK"
	case XAttrBlock:
		return "XATTRBLOCK"
	case XAttrTerm:
		return "XATTRTERM"
	case XTermBlock:
		return "XTERMBLOCK"
	case XBlockTerm:
		return "XBLOCKTERM"
	case XPostDeref:
		return "XPOSTDEREF"
	case XTermOrDorDor:
		return "XTERMORDORDOR"
	}
	return "Expect(?)"
}

// wantsTerm reports whether a value is expected here. Several states are
// term-ish in the way that matters for dispatching `/`, `<` and `&`, and
// collapsing them at the point of use keeps the dispatch readable.
func (e Expect) wantsTerm() bool {
	switch e {
	case XTerm, XRef, XState, XTermBlock, XBlockTerm, XAttrTerm, XTermOrDorDor,
		// XPostDeref is term-ish for the sigil scanners and only for them:
		// `$r->@*`, `$r->%*` and `$r->$m` need a term expected, or `%`
		// reads as modulus. A WORD there is a method name, which is the
		// half scanQuoteLike declines on.
		XPostDeref:
		return true
	}
	return false
}

// transition is what the state machine needs to know about the token just
// emitted, beyond its kind.
//
// A struct rather than three more parameters: every scanner routes through
// emit, so a new input here is added in one place, and a call site that reads
// `after(k, t)` cannot silently pass its booleans in the wrong order.
type transition struct {
	// text is the token's source bytes, for the keyword table.
	text []byte
	// closedBlock is set when this token is a `}` that closed a block rather
	// than a subscript. See trackBrackets.
	closedBlock bool
	// closedSubscript is set when this token closed a subscript -- a `]`, or
	// a `}` that closed one. Not closedBlock's negation: a `)` is neither.
	closedSubscript bool
	// nextIsOpenBrace is whether a `{` follows, ignoring whitespace. perl
	// looks ahead exactly this far at a `)`; see yyl_rightparen.
	nextIsOpenBrace bool
	// afterDeclName is set on the NAME of a `sub NAME` or `package NAME`,
	// after which a block is expected rather than a term.
	afterDeclName bool
	// nextBraceIsBlock is intuitCurly's answer for the `{` that follows,
	// and is meaningful only when nextIsOpenBrace is set.
	nextBraceIsBlock bool
}

// after returns the state following a token of kind k.
//
// A value closes a term, so an operator comes next; an operator opens one, so
// a term comes next. Trivia does not move the machine at all -- that is the
// whole reason trivia are tokens rather than a skipped channel, since a
// lexer that dropped them would have to re-derive the state.
func (e Expect) after(k Kind, t transition) Expect {
	switch k {
	case Whitespace, Comment:
		return e
	case CloseBracket:
		// Which state a closer leaves depends on what it closed, and for `}`
		// the byte alone does not say. perl keeps PL_lex_brackstack for this;
		// see trackBrackets, which sets closedBlock.
		//
		//	$h{a} / 2      subscript closed -- a value, so `/` is division
		//	if (..) { .. }
		//	%h = ()        block closed -- a statement follows, so `%` is a sigil
		//
		// Measured before the fix: the second case lexed as Operator "%" and
		// Word "h", reading a hash as modulus.
		if t.closedBlock {
			return XState
		}
		// A `)` with a `{` next opens a block, not a subscript or an
		// anonymous hash: `if (...) {`, `while (...) {`, `for (...) {`.
		//
		// perl's rule, and it needs no keyword table -- toke.c's
		// yyl_rightparen (7156) skips space and checks the byte:
		//
		//	if (*s == '{')
		//	    PREBLOCK(PERLY_PAREN_CLOSE);
		//
		// This is where XBlock finally gets assigned. M0 declared it and
		// never reached it, which is why `%h = ()` after a conditional lexed
		// as modulus: the `{` was classified from XOperator and its `}` then
		// reported a closed subscript.
		//
		// NOT narrowed to `)`, though perl's own rule is. It used to cost 16
		// T1 files; the phaser table below took that to ONE, and the one is
		// named rather than guessed:
		//
		//	my @s = $r->@[ 2, 1 ];
		//
		// `@[` lexes as a Variable, so no `[` ever reaches the bracket
		// stack, and its `]` pops whatever was underneath. The stack is
		// corrupted from there on, and the broad rule happens to paper over
		// the next brace. A postfix-slice sigil problem, not a brace one --
		// 01a0ad52 owns it, and narrowing this rule is 01a0b6a2's last step.
		//
		// A closed SUBSCRIPT is excluded, and that part is not a hedge. A
		// `{` after one continues the chain:
		//
		//	$a[0]{k}    $h{a}{b}    $x->[0]{k}    ${$y}{Keys}
		//
		// The last of those is reachable only since a dereference became
		// several tokens: while `${$y}` lexed as one Variable, no
		// CloseBracket ever preceded the subscript and nothing could expose
		// it. Before the exclusion every second brace in a chain carried
		// OpensBlock, which is a token saying a block starts where a
		// subscript does. The parser happened to survive it -- it reaches
		// those braces through the postfix path, which never asks -- so
		// nothing failed, and a flag that lies until someone reads it is
		// exactly the kind of defect the brace stack exists to prevent.
		if t.nextIsOpenBrace && !t.closedSubscript {
			return XBlock
		}
		return XOperator
	case Variable, Number, Quote, Readline, FuncSigil:
		return XOperator
	case DerefSigil:
		// A sigil applied to an expression has NOT produced a value yet: what
		// follows is its operand. Leaving XState here made the `{` of
		// `${*$glob}{Keys}` classify as a BLOCK -- trackBrackets reads
		// l.expect at the brace, and before this token existed the sigil and
		// its brace were one Variable token that trackBrackets never saw.
		//
		// Measured on the PerlOnJava corpus: without this,
		// unit/glob_slot_hash_deref.t went 8 -> 14 Unknown nodes and
		// `${*$glob}{Keys} = 5;` parsed as a bare block.
		return XTerm
	case Word:
		// A niladic builtin has produced a value, so an operator comes next.
		// Measured before the table existed:
		//
		//	"my $t = time / 2;"  ->  Word "time"  UnknownRest "/ 2;"
		//
		// The slash opened a pattern that ran to end of input. See keyword.go
		// for how the 21 niladic keywords were measured.
		if isNiladic(string(t.text)) {
			return XOperator
		}
		// A sub or package NAME is followed by a block, not a term. perl
		// says so with PREBLOCK, which sets XBLOCK -- toke.c:6636 for a sub
		// name, 8862 for a package.
		//
		// Without this the `{` of `sub f { 1 }` is classified from XTerm and
		// reads as an anonymous hash, so its `}` reports a closed subscript
		// and the body never becomes a block.
		if t.afterDeclName {
			return XBlock
		}
		// A PHASER's brace is always a block, never a hash, so the word
		// alone settles it -- no lookahead, the way a sub name needs none.
		// perl reaches PREBLOCK for these the same way.
		if isPhaser(string(t.text)) {
			return XBlock
		}
		// A list operator that takes a block: `map`, `grep`, `sort`. This is
		// the one brace the preceding token cannot settle, because both
		// readings are grammatical after these three, so intuitCurly peeks
		// past it at the first thing INSIDE.
		//
		// Without this the `{` of `map { ; $_ }` is classified from XTerm and
		// reads as an anonymous hash, so its `}` reports a closed subscript,
		// the body never becomes a block, and the list that follows is
		// orphaned into a statement of its own.
		if t.nextIsOpenBrace && takesBlock(string(t.text)) {
			if t.nextBraceIsBlock {
				return XBlock
			}
			return XTerm
		}
		// A bareword is the one case the lexer genuinely cannot settle. It
		// might be a value (`Foo::Bar`, a hash key) and leave an operator
		// expected, or a named unary or list operator (`split`, `grep`,
		// `return`) and leave a TERM expected:
		//
		//	$ perl -MO=Deparse -e 'my @p = split /,/, $s;'
		//	my @p = split(/,/, $s, 0);
		//
		// A pattern, not division. perl decides with the symbol table, which
		// a lexer does not have -- toke.c's yyl_keylookup consults the
		// keyword table AND the stash.
		//
		// Term is the safer default of the two. Guessing operator turns every
		// `split /re/` in the corpus into division and silently mis-lexes the
		// rest of the line; guessing term costs a division after a bareword
		// that was really a value, which is rarer and which the parser can
		// still recover from because the token boundaries stay put.
		//
		// The table above settles the builtins. What remains is a USER sub,
		// which no table can settle: perl consults the stash, and three
		// different parses come out of `zzz / 2` depending on whether and how
		// zzz was declared (see isNiladic). That one stays a hedge.
		return XTerm
	case Operator:
		// `->` is the one operator whose next token is not a plain term.
		// perl's XPOSTDEREF, and the state was declared for exactly this:
		//
		//	$r->@*   $r->%*   $r->$m   -- a sigil, so a term is wanted
		//	$o->s    $o->tr            -- a NAME, never a quote operator
		//
		// Leaving XTerm let scanQuoteLike claim the name: `$o->s, "x"`
		// lexed as UnknownRest `s, "x"` -- the comma delimited a
		// substitution that ran to end of input. Leaving XOperator instead
		// would fix the name and break every postfix dereference, which is
		// why this is its own state rather than one of the other two.
		if string(t.text) == "->" {
			return XPostDeref
		}
		return XTerm
	case Semicolon:
		return XState
	}
	return e
}
