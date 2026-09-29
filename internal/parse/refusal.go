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

// Nine codes for ten `&Node{Kind: Unknown, ...}` sites. The tenth is
// parse.go's `expr.Kind == Unknown`, which propagates an existing code
// outward rather than declaring one; see the note in RefusalSites.
//
// Declared here rather than at the sites so that adding a site means
// adding a constant AND an inventory entry, which is one edit in one
// file. RefusalSites below is what makes the pair inseparable.
//
// Two of the nine -- EmptyDeref and NotAnExpression -- are wired but not
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
	// parseTerm, consumed so the loop advances.
	NotATerm RefusalCode = "not_a_term"

	// MissingOperand: an infix or prefix operator whose right operand ran
	// out. The Unknown spans the operator.
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
	// The tenth site, parse.go's `expr.Kind == Unknown`, has no code of
	// its own and is deliberately absent from this table. It widens an
	// Unknown from the expression to the statement -- a change of SPAN,
	// not of cause -- and carries the inner code outward unchanged. A code
	// there would erase the distinction the codes exist to make.
	TrailingTokens: {
		What:  "an expression parsed but bytes remain before the terminator",
		Where: "parse.go, parseStatement: !endsStatement",
	},
	EmptyDeref: {
		What:  "a `${}` dereference with nothing inside the braces",
		Where: "term.go, parseTerm: inner == nil",
	},
	NotATerm: {
		What:  "a token that cannot begin a term",
		Where: "term.go, parseTerm: fallthrough",
	},
	MissingOperand: {
		What:  "an operator whose right operand ran out",
		Where: "expr.go, operand",
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
}
