// ABOUTME: The 32 precedence levels of perly.y as a Pratt binding-power table.
// ABOUTME: Level order is perly.y declaration order; BP is level*10, leaving room to insert.

package parse

import "sort"

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
	// Level is the perly.y declaration index, 1 (loosest) to 32 (tightest).
	// Kept alongside BP so the table can be checked against perly.y by
	// counting rather than by reading binding powers.
	Level int
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

// infix is the complete infix and postfix table, spec §4.2.
//
// Thirty-two levels, counted rather than remembered:
//
//	$ grep -cE '^%(left|right|nonassoc)' perly.y
//	32
//	$ grep -cE '^%nonassoc' perly.y
//	11
//
// An earlier pass in this project wrote "33" into a README from memory and it
// propagated. The number is cheap to check.
//
// Levels 1, 3, 18, 28 and 30 carry no lexable operator -- they are
// pseudo-tokens and plugin hooks -- so they appear in levelsPresent rather
// than here. Levels 2, 7, 19, 20 and 25 are prefix forms and live in prefix.
var infix = map[string]OpInfo{
	// 4-5: the lowest logical operators. Below the comma at 8, which is the
	// `and or not` versus `&& ||` cliff of §4.1.1.
	"or":  {4, 40, AssocLeft},
	"xor": {4, 40, AssocLeft}, // NOT its own level -- toke.c:9224
	"and": {5, 50, AssocLeft},

	// 8: comma. `=>` is a comma that autoquotes its left bareword.
	",":  {8, 80, AssocLeft},
	"=>": {8, 80, AssocLeft},

	// 9: assignment, right associative. One token class in toke.c:250.
	"=": {9, 90, AssocRight}, "+=": {9, 90, AssocRight},
	"-=": {9, 90, AssocRight}, "*=": {9, 90, AssocRight},
	"/=": {9, 90, AssocRight}, ".=": {9, 90, AssocRight},
	"%=": {9, 90, AssocRight}, "**=": {9, 90, AssocRight},
	"x=": {9, 90, AssocRight}, "||=": {9, 90, AssocRight},
	"&&=": {9, 90, AssocRight}, "//=": {9, 90, AssocRight},
	"|=": {9, 90, AssocRight}, "&=": {9, 90, AssocRight},
	"^=": {9, 90, AssocRight}, "<<=": {9, 90, AssocRight},
	">>=": {9, 90, AssocRight}, "|.=": {9, 90, AssocRight},
	"&.=": {9, 90, AssocRight}, "^.=": {9, 90, AssocRight},
	"^^=": {9, 90, AssocRight},

	// 10: ternary, right associative. Measured:
	//   perl -MO=Deparse -e 'my $x = $a ? $b : $c ? $d : $e;'
	//   my $x = $a ? $b : ($c ? $d : $e);
	"?": {10, 100, AssocRight},

	// 11: range, NONASSOC. Measured: `my $x = 1 .. 2 .. 3;` is
	// "syntax error near 2 ..".
	"..": {11, 110, AssocNone}, "...": {11, 110, AssocNone},

	// 12-13.
	"||": {12, 120, AssocLeft}, "//": {12, 120, AssocLeft},
	"^^": {12, 120, AssocLeft}, // toke.c:6441 -- shares || level
	"&&": {13, 130, AssocLeft},

	// 14-15: bitwise.
	"|": {14, 140, AssocLeft}, "^": {14, 140, AssocLeft},
	"|.": {14, 140, AssocLeft}, "^.": {14, 140, AssocLeft},
	"&": {15, 150, AssocLeft}, "&.": {15, 150, AssocLeft},

	// 16: equality. The chaining half is handled in §4.3, not here.
	"==": {16, 160, AssocLeft}, "!=": {16, 160, AssocLeft},
	"eq": {16, 160, AssocLeft}, "ne": {16, 160, AssocLeft},
	"<=>": {16, 160, AssocLeft}, "cmp": {16, 160, AssocLeft},
	"~~": {16, 160, AssocLeft},

	// 17: relational.
	"<": {17, 170, AssocLeft}, ">": {17, 170, AssocLeft},
	"<=": {17, 170, AssocLeft}, ">=": {17, 170, AssocLeft},
	"lt": {17, 170, AssocLeft}, "gt": {17, 170, AssocLeft},
	"le": {17, 170, AssocLeft}, "ge": {17, 170, AssocLeft},
	"isa": {17, 170, AssocLeft}, // toke.c:8697 -- NON-chaining

	// 21-23.
	"<<": {21, 210, AssocLeft}, ">>": {21, 210, AssocLeft},
	"+": {22, 220, AssocLeft}, "-": {22, 220, AssocLeft},
	".": {22, 220, AssocLeft}, // concat is an ADDOP: `"a" . 1 + 2` is 'a1' + 2
	"*": {23, 230, AssocLeft}, "/": {23, 230, AssocLeft},
	"%": {23, 230, AssocLeft},
	"x": {23, 230, AssocLeft}, // repetition is a MULOP

	// 24: binding.
	"=~": {24, 240, AssocLeft}, "!~": {24, 240, AssocLeft},

	// 26: exponentiation, right associative AND tighter than unary minus.
	// Measured: 2**3**2 is 512, and -2**2 is -4.
	"**": {26, 260, AssocRight},

	// 27: postfix inc/dec.
	"++": {27, 270, AssocNone}, "--": {27, 270, AssocNone},

	// 29-32: the postfix chain, tightest in the grammar.
	"->": {29, 290, AssocLeft},
	"(":  {31, 310, AssocLeft},
	"[":  {32, 320, AssocLeft},
	"{":  {32, 320, AssocLeft},
}

// assignLevel is perly.y's assignment level. `=` and the eighteen compound
// forms are ONE token class in toke.c:250, so code that asks "is this an
// assignment" tests the level rather than listing the spellings.
const assignLevel = 9

// bpBelowComma is the floor that admits the comma and excludes everything
// under it -- which, in the infix table, is exactly `and`, `or` and `xor`.
//
// parseExpr stops at `op.BP <= minBP`, so this is `and`'s own power: `and`
// (50) and `or`/`xor` (40) stop, the comma (80) does not. Levels 6 and 7 sit
// in the gap and carry no infix entry -- `not` is prefix and the list
// operators are prefix -- so the three word operators are the whole set.
//
// Named because three sites need the same boundary and a literal 50 at each
// would be three chances to write the wrong one. Derived from the table so
// it cannot drift from it.
var bpBelowComma = infix["and"].BP

// prefix is the power a prefix operator passes down for its operand.
var prefix = map[string]int{
	"not": 60,  // level 6, takes a listexpr -- swallows commas
	"!":   250, // level 25
	"~":   250,
	"~.":  250,
	"-":   250, // UMINUS
	"+":   250, // UMINUS -- a no-op that exists to force term parsing
	"\\":  250, // REFGEN: the srefgen the fidelity harness measures
	"++":  270, // PREINC
	"--":  270, // PREDEC
}

// levelsPresent records which of the 32 perly.y levels this table accounts
// for, and how.
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
