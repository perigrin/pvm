// ABOUTME: The refusal codes: stable identifiers naming WHICH site declined, so two refusals differ.
// ABOUTME: RefusalSites is the inventory the constants are declared against, not a second list beside them.

package parse

// RefusalCode names which of the parser's refusal sites produced an
// Unknown.
//
// A CODE, not a message. A message is prose and changes when someone
// rewords it; a corpus file that named one would break on an edit that
// changed nothing about the parser. A code is a promise: it changes only
// when the reason the parser declines changes, which is exactly the event
// a corpus file wants to be told about.
//
// The spelling is lower_snake_case for the same reason -- it is meant to
// be typed into a `# STATUS refuses` header by hand and grepped for
// afterwards.
type RefusalCode string

// Sixteen codes for seventeen construction sites -- `&Node{Kind: Unknown, ...}`,
// or the Error or Missing node recovery builds for the four codes it
// replaces (recovery.go). The seventeenth is parse.go's `expr.Refused()`,
// which propagates an existing code outward rather than declaring one; see
// the note in RefusalSites.
//
// Declared here rather than at the sites so that adding a site means
// adding a constant AND an inventory entry, which is one edit in one
// file. RefusalSites below is what makes the pair inseparable.
//
// Two of the sixteen -- EmptyDeref and NotAnExpression -- are wired but not
// reachable from any source today. They are not dead code: they are the
// codes those sites will carry the day parseExpr can return nil there,
// and an unreachable site that becomes reachable must arrive with a code
// rather than without one. TestEveryUnknownSiteHasACode documents which
// ones and why.
const (
	// UnimplementedStatement: a statement keyword this milestone has not
	// reached. The largest single source of Unknowns in T1.
	UnimplementedStatement RefusalCode = "unimplemented_statement"

	// NotAnExpression: the statement is not a known form and the
	// expression parser produced nothing at all.
	NotAnExpression RefusalCode = "not_an_expression"

	// TrailingTokens: an expression parsed, but bytes remain before the
	// terminator. A half-parsed expression beside a mystery is a tree
	// that claims to be a parse of bytes it did not read.
	TrailingTokens RefusalCode = "trailing_tokens"

	// EmptyDeref: `${}` and friends -- a dereference with nothing to
	// dereference.
	EmptyDeref RefusalCode = "empty_deref"

	// NotATerm: a token that cannot begin a term. The catch-all of
	// parseTerm, consumed so the loop advances; also a `my` or `our` in
	// term position with no variable after it, and an anonymous sub's
	// head with no block after it.
	NotATerm RefusalCode = "not_a_term"

	// MissingOperand: an infix or prefix operator whose right operand ran
	// out, or stands before a `;` or `}`. A zero-width Missing node at the
	// operator's end, and reported (R-g).
	MissingOperand RefusalCode = "missing_operand"

	// NonassocRepeated: `$a .. $b .. $c` -- a non-associative operator
	// used twice at one level, which perl rejects and we decline.
	NonassocRepeated RefusalCode = "nonassoc_repeated"

	// TernaryNoColon: a `?` with no `:`. Declining rather than inventing
	// the branch that is not there.
	TernaryNoColon RefusalCode = "ternary_no_colon"

	// ChainClassMismatch: `$a <=> $b == $c` -- two comparison operators
	// at one precedence level that do not chain with each other.
	ChainClassMismatch RefusalCode = "chain_class_mismatch"

	// UnclosedBrace: `map {` or `grep {` whose brace never closes, which
	// perl reports as a missing right curly.
	UnclosedBrace RefusalCode = "unclosed_brace"

	// RefScalarSlot: a call fills a `\$` prototype slot with something
	// that is not a scalar lvalue -- `sref(1)` -- which perl rejects as
	// "Type of arg 1 to main::sref must be scalar", or an operator's `\$`
	// operand is one -- `++1` -- which perl rejects as "Can't modify
	// constant item in preincrement (++)".
	RefScalarSlot RefusalCode = "ref_scalar_slot"

	// CallArity: a call passes a number of arguments no declaration of its
	// sub accepts -- `each()`, `select(1, 2)` -- which perl rejects as "Not
	// enough arguments" or "Too many arguments".
	CallArity RefusalCode = "call_arity"

	// UnclosedOpener: a `(`, `[` or `{` still open at end of file, which
	// perl reports as "Missing right curly or square bracket" or as a
	// syntax error at EOF. The Unknown spans the last token perl read,
	// whose line is the one perl reports.
	UnclosedOpener RefusalCode = "unclosed_opener"

	// FormatNotTerminated: a `format NAME =` the file ends inside, before
	// a lone `.` line, which perl reports as "Format not terminated".
	FormatNotTerminated RefusalCode = "format_not_terminated"

	// AmbiguousSigil: a `&` or `*` the lexer read as a sigil and the parser
	// as an operator, after an undeclared bareword, before a quote-op name
	// -- `A &y` -- or a `*` before a `#` or quote glob name, `A*#`, `A*"x"`. perl
	// reads the operator and the quote or comment it begins.
	AmbiguousSigil RefusalCode = "ambiguous_sigil"

	// NestingTooDeep: brackets, blocks or operands nested past maxDepth.
	// perl takes any depth; a recursive parser would overflow its stack,
	// which is fatal to the LSP, so the deeper input is declined instead.
	NestingTooDeep RefusalCode = "nesting_too_deep"
)

// RefusalSite says what one code means and where the site lives.
type RefusalSite struct {
	// What the parser could not do, in a sentence a corpus author can
	// read without opening the parser.
	What string

	// Where the site is: file and function. A reader looking for the
	// code's site finds it here rather than by grepping for the constant.
	Where string
}

// RefusalSites is the inventory: every refusal code, what it means, and
// which construction site emits it.
//
// This is the enumeration the issue asked for, and it is deliberately the
// table the TEST reads rather than a comment beside one. Two lists that
// must agree are how drift starts -- this package has been bitten by that
// four times -- so there is one list, the sites are declared against it,
// and TestRefusalCodeInventory holds its size to the number of
// `&Node{Kind: Unknown, ...}` constructions the parser actually has.
var RefusalSites = map[RefusalCode]RefusalSite{
	UnimplementedStatement: {
		What:  "a statement keyword this milestone has not implemented",
		Where: "parse.go, parseStatement: statementKeywords",
	},
	NotAnExpression: {
		What:  "the statement is not a known form and parses as no expression",
		Where: "parse.go, parseStatement: expr == nil",
	},
	// The seventeenth site, parse.go's `expr.Refused()`, has no code of
	// its own and is deliberately absent from this table. It widens a
	// refusal from the expression to the statement -- a change of SPAN,
	// not of cause -- and carries the inner code outward unchanged. A code
	// there would erase the distinction the codes exist to make.
	TrailingTokens: {
		What:  "an expression parsed but bytes remain before the terminator",
		Where: "parse.go, refuseTrailingTokens: !endsStatement, !declarationEnded",
	},
	EmptyDeref: {
		What:  "a `${}` dereference with nothing inside the braces",
		Where: "term.go, parseTerm: inner == nil",
	},
	NotATerm: {
		What:  "a token that cannot begin a term",
		Where: "term.go, refuseNotATerm: parseTerm's fallthrough, a `my` or `our` with no target, parseAnonSub with no block, and a `package` or `class` with no name",
	},
	MissingOperand: {
		What:  "an operator whose right operand ran out",
		Where: "expr.go, operand; recovery.go, missingOperand",
	},
	NonassocRepeated: {
		What:  "a non-associative operator used twice at one level",
		Where: "expr.go, parseNonassoc",
	},
	TernaryNoColon: {
		What:  "a `?` with no `:`",
		Where: "expr.go, parseTernary",
	},
	ChainClassMismatch: {
		What:  "comparison operators from classes that do not chain",
		Where: "chain.go, parseChain",
	},
	UnclosedBrace: {
		What:  "a map or grep brace with no closing `}`",
		Where: "call.go, parseListOpBlock",
	},
	RefScalarSlot: {
		What:  "a `\\$` prototype slot or operand given what perl refuses there",
		Where: "refscalar.go, refuseRefScalarSlot and refuseAliasedOperand",
	},
	CallArity: {
		What:  "a call with a number of arguments no declaration accepts",
		Where: "arity.go, refuseArity",
	},
	UnclosedOpener: {
		What:  "a bracket, or an attribute's `(` argument, still open at end of file",
		Where: "parse.go, refuseUnclosedOpener; use.go, parseAttributes sets attrOpenAtEOF",
	},
	FormatNotTerminated: {
		What:  "a format declaration with no `.` line before end of file",
		Where: "decl.go, parseFormatDecl",
	},
	AmbiguousSigil: {
		What:  "a `&` or `*` after an undeclared bareword, before a quote-op name or (for `*`) a `#` or quote glob name",
		Where: "expr.go, parseInfix: refuseAmbiguousSigil",
	},
	NestingTooDeep: {
		What:  "brackets, blocks or operands nested deeper than the parser's cap",
		Where: "parse.go, refuseTooDeep: parseTerm, parseInfix and parseBlock past maxDepth",
	},
}
