// ABOUTME: The Pratt binding powers, derived from CORE.pmt's precedence relations.
// ABOUTME: Also the census of perly.y's 32 levels, which the conformance tiers measure against.

package parse

import (
	"slices"
	"sort"
	"sync"
)

// Assoc is how an operator groups when it meets one of equal power.
type Assoc uint8

const (
	AssocLeft Assoc = iota
	AssocRight
	// AssocNone is perly.y's %nonassoc: a second occurrence at the same
	// level is a syntax ERROR, not a grouping. A binding power alone cannot
	// express that -- it stops the recursion but still accepts the input --
	// so nonassoc is checked separately in the driving loop.
	AssocNone
)

// OpInfo is one infix or postfix entry.
type OpInfo struct {
	BP    int
	Assoc Assoc
}

// rightBP is the power to pass down when parsing the right operand. Left
// associative passes BP, so an equal-power operator to the right stops and
// the left grouping wins; right associative passes BP-1, so it continues.
func (o OpInfo) rightBP() int {
	if o.Assoc == AssocRight {
		return o.BP - 1
	}
	return o.BP
}

// The parser's binding powers, spec §4.2, derived from the precedence
// relations CORE.pmt states (RFC 0001, "Precedence is a relation between
// operators") by derivePowers: infix holds the infix and postfix operators,
// prefix the power a prefix operator passes down for its operand,
// bpNamedUnary and bpListOp the powers a named unary's and a list
// operator's operands are parsed at.
//
// A named unary is tighter than comparison and looser than the shifts, so
// `length $x + 1` is `length($x + 1)` and `length $x < 5` is `length($x) <
// 5` -- both measured on the optree. A list operator is below the comma,
// which is exactly how it swallows the whole list.
var (
	infix                  map[string]OpInfo
	prefix                 map[string]int
	bpNamedUnary, bpListOp int

	// bpDeref is the subscripts' own power: a dereference binds tighter
	// than every infix and postfix operator below them -- `->` included
	// -- so its operand is the braced expression or the single variable
	// and nothing more, and the parse stops before a subscript: below
	// that, `@$r[1,2]` read as a deref of `$r[1,2]`.
	//
	// `$$x[0]` is `${$x}[0]` -- the subscript applies to the DEREFERENCE,
	// not to `$x` -- so the sigil must take its operand before any postfix
	// gets a chance. Parsing at a power below the subscripts' leaves `[0]`
	// to the caller's led loop, which then wraps the whole Unary in an
	// Index. Measured on perl 5.42.0:
	//
	//	$ perl -MO=Deparse -e 'my $r = [7]; print $$r[0];'
	//	print $r->[0];
	//
	// Deparse prints the arrow form, which is the same operation spelled
	// the other way -- and is why §4.14 gives both one node with an
	// `Arrow` flag.
	bpDeref int

	// bpBelowComma is the floor that admits the comma and excludes
	// everything under it -- which, in the infix table, is exactly `and`,
	// `or` and `xor`.
	//
	// parseExpr stops at `op.BP <= minBP`, so this is `and`'s own power:
	// `and` and `or`/`xor` stop, the comma does not. `not` and the list
	// operators sit in the gap and carry no infix entry -- `not` is prefix
	// and the list operators are prefix -- so the three word operators are
	// the whole set. Below the comma is the `and or not` versus `&& ||`
	// cliff of §4.1.1.
	//
	// Named because three sites need the same boundary and a literal at
	// each would be three chances to write the wrong one. Derived from the
	// table so it cannot drift from it.
	bpBelowComma int

	// coreOrderDecls are the relations the powers are derived from, and
	// levelPowers the power of each operator's level, by its label in
	// the order: what a library's operators are placed among
	// (libraryInfix).
	coreOrderDecls []operatorDecl
	levelPowers    map[string]int

	powersOnce sync.Once
)

// derivePowers derives the binding powers, once, from CORE.pmt's relations
// and the operators it has no line for (undeclaredOperators). Every
// parse but the read of CORE.pmt below derives them first, through
// parseWith, so a test that reads CORE.pmt alone gets them too.
//
// CORE.pmt is read for them before any power exists, so a default
// expression a power is needed to read, chdir's `$dir = $_`, is misread and
// its declaration is in error. What the powers are derived from is the
// operator lines and the levels the builtins' shapes are in, which that
// read does not miss: TestParserPrecedenceFromCore holds the powers to the
// ones a read with them derives, and readCore, reading with them, refuses
// a CORE.pmt in error.
func derivePowers() {
	powersOnce.Do(func() {
		src, ok := declaration("CORE")
		if !ok {
			panic("parse: declarations/CORE.pmt is not embedded")
		}
		p := newParser(src, nil, true)
		p.buildingCore, p.derivingPowers = true, true
		facts := readDeclarationWith(p)
		decls := precedenceDeclsOf(facts, deriveShapes(protoTable(facts), facts.signatures))
		coreOrderDecls = slices.Concat(decls, undeclaredOperators)
		bp, err := deriveBindingPowers(coreOrderDecls)
		if err != nil {
			panic("parse: declarations/CORE.pmt: " + err.Error())
		}
		levelPowers = bp.levels
		for _, op := range coreQualifiedOps {
			bp.infix["CORE::"+op] = bp.infix[op]
		}
		infix, prefix, bpNamedUnary, bpListOp = bp.infix, bp.prefix, bp.namedUnary, bp.listOp
		bpDeref = infix["["].BP
		bpBelowComma = infix["and"].BP
	})
}

// isAssignment reports whether op is `=` or one of its compound forms.
// They are ONE token class in toke.c:250, so code that asks "is this an
// assignment" tests the level rather than listing the spellings.
func isAssignment(op OpInfo) bool {
	return op.BP != 0 && op.BP == infix["="].BP
}

// IsWordShapedOperator reports whether text spells one of perl's operators
// with letters rather than punctuation: `x`, `cmp`, `eq`, `and`, `not` and
// the rest, plus the compound `x=`.
//
// EXPORTED FOR THE CONFORMANCE GLOSSARY, which needs the question answered
// and must not answer it with a second list. `conformance/GLOSSARY.md`
// defines `word-shaped operator` as a category and
// `internal/conformance/categories.go` maps it here, the same way it maps
// `quote-like operator` to `lexer.HasQuoteOperator` rather than copying
// the lexer's quote-op table -- a copy of that one lost `qx` on its first
// day.
//
// The tables above ARE the answer, so this reads them rather than naming
// spellings. Both are consulted because `not` is prefix and the rest are
// infix, and the category is about SPELLING rather than arity: `perlop`
// calls `not` an operator ("Unary C<"not"> returns the logical negation")
// and it is spelled with letters, which is the whole of the test.
//
// A leading letter is what separates a word from punctuation. Every key in
// these tables is one or the other, never mixed -- `x=` begins with a
// letter and `+=` does not.
func IsWordShapedOperator(text string) bool {
	if text == "" {
		return false
	}
	c := text[0]
	isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
	if !isLetter {
		return false
	}
	derivePowers()
	if _, ok := infix[text]; ok {
		return true
	}
	_, ok := prefix[text]
	return ok
}

// atOperatorBelowComma reports whether the next significant token is one of
// the three word operators below the comma.
//
// Three sites stand where an element loop or an initialiser has just stopped
// and must decide whether what stopped it belongs ABOVE what was built:
// parseParenList, parseBracketed and parseVarDecl. Two copies of the test
// were defensible; the third earned a name, and the bracketed one was missing
// the test entirely -- `[$a and $b]` let the operand escape the bracket, so
// the tree said `and` sat above the declaration with a one-element arrayref
// as its left operand.
//
// Nothing is consumed: the caller resumes the Pratt loop with the left
// operand it assembled itself.
func (p *parser) atOperatorBelowComma() bool {
	next, ok := p.peekSignificant()
	if !ok {
		return false
	}
	op, isOp := infix[p.text(next)]
	return isOp && op.BP <= bpBelowComma
}

// levelsPresent records which of the 32 perly.y levels this parser
// accounts for, and how. Thirty-two levels, counted rather than remembered:
//
//	$ grep -cE '^%(left|right|nonassoc)' perly.y
//	32
//	$ grep -cE '^%nonassoc' perly.y
//	11
//
// An earlier pass in this project wrote "33" into a README from memory and it
// propagated. The number is cheap to check.
//
// perly.y has no row of its own for `isa`: toke.c:8697 lexes it as an
// NCRELOP, at level 17 with the relational operators. perlop's table, and
// CORE.pmt's relations, give it its own row, tighter than theirs; perl
// refuses the two unparenthesised together either way (chain.go).
//
// Five levels carry no lexable operator: 1 PREC_LOW is a pseudo-token used
// only via %prec, 3/18/28 are XS infix-plugin hooks, and 30 is
// PERLY_PAREN_CLOSE, which exists for conflict resolution. Listing them
// explicitly is the difference between "32 levels are handled" and "32 levels
// are in a map" -- a count of map entries would silently pass while a level
// was missing, because several levels share a binding power.
// LevelsAccountedFor returns each perly.y precedence level and how this
// parser handles it. Exported so the completeness test asserts against the
// table rather than against a count of map entries.
func LevelsAccountedFor() map[int]string {
	out := make(map[int]string, len(levelsPresent))
	for k, v := range levelsPresent {
		out[k] = v
	}
	return out
}

// NonassocOperators returns the infix operators declared %nonassoc, whose
// repetition is a syntax error rather than a grouping.
func NonassocOperators() []string {
	derivePowers()
	var out []string
	for op, info := range infix {
		if info.Assoc == AssocNone {
			out = append(out, op)
		}
	}
	sort.Strings(out)
	return out
}

var levelsPresent = map[int]string{
	1:  "PREC_LOW, pseudo-token, never lexed",
	2:  "LOOPEX: goto last next redo dump -- statement forms, ch5",
	3:  "PLUGIN_LOW_OP, XS plugin hook",
	4:  "infix: or xor",
	5:  "infix: and",
	6:  "prefix: not",
	7:  "LSTOP: list operators -- their own issue",
	8:  "infix: , =>",
	9:  "infix: = and the compound assignments",
	10: "infix: ?:",
	11: "infix: .. ... (nonassoc)",
	12: "infix: || // ^^",
	13: "infix: &&",
	14: "infix: | ^ |. ^.",
	15: "infix: & &.",
	16: "infix: == != eq ne <=> cmp ~~",
	17: "infix: < > <= >= lt gt le ge isa",
	18: "PLUGIN_REL_OP, XS plugin hook",
	19: "UNIOP: named unaries and file tests -- their own issue",
	20: "KW_REQUIRE -- statement form, ch5",
	21: "infix: << >>",
	22: "infix: + - .",
	23: "infix: * / % x",
	24: "infix: =~ !~",
	25: "prefix: ! ~ ~. unary- unary+ backslash",
	26: "infix: **",
	27: "infix: ++ -- (postfix); prefix: ++ --",
	28: "PLUGIN_HIGH_OP, XS plugin hook",
	29: "infix: ->",
	30: "PERLY_PAREN_CLOSE, conflict resolution only",
	31: "infix: ( -- postfix call",
	32: "infix: [ { -- postfix subscript",
}

// coreQualifiedOps are the word operators perl also accepts spelled with
// CORE::, meaning the bare word -- measured on 5.42.0 with -MO=Deparse,
// `$r = $a CORE::eq $b` is `$r = $a eq $b`. Entered in the tables under the
// spelled name, so every lookup sees them and canon keeps the spelling:
// derivePowers enters them in infix, and init in cmpClasses.
var coreQualifiedOps = []string{
	"eq", "ne", "lt", "gt", "le", "ge", "cmp", "x", "and", "or", "xor", "isa",
}

func init() {
	for _, op := range coreQualifiedOps {
		if c, ok := cmpClasses[op]; ok {
			cmpClasses["CORE::"+op] = c
		}
	}
}
