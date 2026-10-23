// ABOUTME: Tier 04 checked against the finished tooling: the tier where a parse can be WRONG, not Unknown.
// ABOUTME: Precedence and associativity leave no op behind, so every claim here is behavioural or nothing.
package conformance

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierOperators is the tier this file is about.
//
// A constant rather than a literal at eight call sites, for the reason
// tier01_test.go gives for tierLiterals: the tier number is a POSITION
// and positions move.
const tierOperators = "04_operators"

// TestTierOperatorsPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// The same two claims tier 01 checks -- perl agrees with the parses bit,
// perl prints what the file pins -- and THIS TIER IS WHERE THEY CARRY THE
// MOST WEIGHT IN THE CORPUS, because they are the only claims it has.
//
// Every earlier tier can fall back on ops. This one cannot for its
// central subject: the README measures that `$a + $b * $c` and `($a + $b)
// * $c` emit the SAME four ops in the SAME order, so grouping -- the
// thing the tier is about -- is invisible to every op-based check in this
// package. What separates the two groupings is the VALUE, and a pinned
// output is what a value is.
//
// That makes a wrong pin worse here than anywhere else. Elsewhere a pin
// perl does not produce leaves a tier with its op claims intact. Here it
// leaves the tier measuring nothing at all.
func TestTierOperatorsPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierOperators, perl)

	for name, f := range tierFiles(t, tierOperators) {
		t.Run(name, func(t *testing.T) {
			compiles, output := askPerl(t, f.Source)

			switch {
			case f.ExpectParses && !compiles:
				t.Errorf("file says `expect parses`, perl -c refuses it")
			case f.ExpectParsent && compiles:
				t.Errorf("file says `expect parsent`, perl -c accepts it")
			}
			if f.ExpectOutput != nil && output != *f.ExpectOutput {
				t.Errorf("pinned output %q, perl prints %q", *f.ExpectOutput, output)
			}
		})
	}
}

// TestTierOperatorsLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a file that reaches forward fails HERE,
// named as this tier's problem rather than as one subtest among a
// hundred and thirteen.
//
// The lint has a live and unusually sharp constraint in this tier, and it
// is why several operators perl HAS are absent from the corpus. `..` is
// level 11 and compiles to `flip`/`flop`/`range`; `++` is level 27 and
// compiles to `preinc`/`postinc`; `|` and `<<` compile to `bit_or` and
// `left_shift`. NO TIER IN THE CORPUS CLAIMS ANY OF THOSE OPS. A file
// reaching for them to widen this tier's operator coverage would satisfy
// its own output pin and fail here, which is the lint doing exactly what
// it is for -- and is why TestTierOperatorsAdjacentPairs records those
// levels as out of budget rather than writing fixtures for them.
func TestTierOperatorsLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierOperators) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierOperators, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// nonassocLevel is one of perly.y's eleven `%nonassoc` levels, and what
// perl DOES when that level's operator is repeated.
//
// `Probe` is the source that repeats the level's operator; empty when the
// level carries no lexable operator at all and there is nothing to
// repeat. `Rejects` is what perl was MEASURED to do with that source --
// the field this whole test exists to fill from the interpreter rather
// than from the `%nonassoc` keyword.
type nonassocLevel struct {
	Level  int
	Token  string
	Probe  string
	Reject bool
	Why    string
}

// nonassocLevels is perly.y's eleven `%nonassoc` levels with a probe for
// each, and the MEASURED answer.
//
// THE POINT OF THIS TABLE, which is the tier's second-sharpest finding.
// `%nonassoc` in a yacc grammar is not a promise that repetition is
// rejected. It is a promise about how the PARSER RESOLVES A CONFLICT, and
// a conflict only arises where an operator has a left operand to fight
// over. Nine of these eleven levels are prefix forms, list operators or
// pseudo-tokens; for those the declaration settles nothing observable,
// and a parser that inferred "nonassoc means repetition is an error" from
// the table would reject three programs perl accepts.
//
// Measured, against the pinned 5.42.0:
//
//	 7 LSTOP    print print @a          -e syntax OK
//	19 UNIOP    scalar scalar @a        -e syntax OK
//	20 REQUIRE  require require         -e syntax OK
//	 2 LOOPEX   last LOOP last LOOP     syntax error
//	11 DOTDOT   1 .. 2 .. 3             syntax error
//	27 POSTINC  $a++ ++                 Can't modify postincrement (++)
//
// Level 27 is the one worth staring at. Perl PARSES `$a++ ++` -- the
// rejection is an lvalue check on the built optree, not a yacc syntax
// error -- so its `%nonassoc` is not a parse-time rule either. `expr.go`
// already excludes `++` and `--` from `parseNonassoc` by name, with a
// comment saying they are postfix and take no right operand. This table
// is the measurement that says that exclusion is right for a second
// reason: even if they had a right operand, perl would not have refused
// them there.
//
// The five unlexable levels carry no probe. That is not a gap this test
// papers over -- an empty probe is an assertion in its own right, checked
// below, that NO source can reach the level, which is why nothing
// measures it.
var nonassocLevels = []nonassocLevel{
	{1, "PREC_LOW", "", false,
		"pseudo-token, used only via %prec to force a reduction; never lexed"},
	{2, "LOOPEX", "LOOP: { last LOOP last LOOP; }", true,
		"a loop exit takes an optional term, and a second one is a yacc syntax error"},
	{3, "PLUGIN_LOW_OP", "", false,
		"XS infix-plugin hook; no core operator is lexed at this level"},
	{7, "LSTOP", "my @a = (1); print print @a;", false,
		"a list operator swallows everything rightward, so the second is its ARGUMENT and no conflict arises"},
	{11, "DOTDOT", "my $x = 1 .. 2 .. 3;", true,
		"the one level where %nonassoc means what the keyword suggests: a yacc syntax error"},
	{18, "PLUGIN_REL_OP", "", false,
		"XS infix-plugin hook; no core operator is lexed at this level"},
	{19, "UNIOP", "my @a = (1); my $x = scalar scalar @a;", false,
		"a named unary takes one term rightward, so the second is its ARGUMENT and no conflict arises"},
	{20, "KW_REQUIRE", "require require;", false,
		"require takes a term rightward; the second parses as the bareword it requires"},
	{27, "POSTINC", "my $a = 1; my $x = $a++ ++;", true,
		"rejected, but NOT by the grammar: perl builds the optree and the lvalue check declines it"},
	{28, "PLUGIN_HIGH_OP", "", false,
		"XS infix-plugin hook; no core operator is lexed at this level"},
	{30, "PERLY_PAREN_CLOSE", "", false,
		"conflict resolution only; `)` closes a term and never stands as an operator"},
}

// TestTierOperatorsNonassocMeasured measures each of perly.y's eleven
// `%nonassoc` levels against perl rather than inferring it from the
// binding-power table.
//
// WHY THE TABLE CANNOT ANSWER THIS. `precedence.go` encodes assoc as a
// field on `OpInfo`, and `AssocNone`'s own comment says a binding power
// "stops the recursion but still accepts the input" -- so nonassoc is
// checked separately in the driving loop. That separate check is a
// DECISION, and a decision taken from a keyword in a grammar file is an
// inference. The issue asks for a measurement, and this is the
// difference: perl is run.
//
// What running perl says, and what no reading of `%nonassoc` would have
// said: THREE of the eleven levels ACCEPT the repetition. `print print
// @a`, `scalar scalar @a` and `require require` all compile. A parser
// that read the eleven `%nonassoc` lines and refused repetition at each
// would be wrong about three of them -- and wrong in the direction this
// tier is named for, refusing a program perl runs.
//
// The count is asserted, because eleven is a fact about perly.y that the
// spec cites a `grep -c` for and `precedence.go` repeats. A twelfth
// nonassoc level would be a change to the grammar this parser models,
// and it should fail here rather than pass quietly with one level
// unmeasured.
func TestTierOperatorsNonassocMeasured(t *testing.T) {
	const perlyNonassocLevels = 11
	if len(nonassocLevels) != perlyNonassocLevels {
		t.Fatalf("perly.y declares %d %%nonassoc levels, this table holds %d.\n"+
			"\tThe number is `grep -cE '^%%nonassoc' perly.y`, cited by both "+
			"the spec and precedence.go. A table that has drifted from it "+
			"measures some other grammar.", perlyNonassocLevels, len(nonassocLevels))
	}

	// Every level named here must be one perly.y actually has, and must
	// be named once. A duplicate would let the table reach eleven entries
	// while leaving a level unmeasured -- which is the exact failure the
	// count above is meant to catch, arriving through the back door.
	levels := parse.LevelsAccountedFor()
	seen := map[int]bool{}
	for _, l := range nonassocLevels {
		if _, ok := levels[l.Level]; !ok {
			t.Errorf("level %d (%s) is not one of the 32 levels parse.LevelsAccountedFor names",
				l.Level, l.Token)
		}
		if seen[l.Level] {
			t.Errorf("level %d (%s) appears twice; one of the eleven is then unmeasured",
				l.Level, l.Token)
		}
		seen[l.Level] = true
	}

	var measured, unlexable []string
	for _, l := range nonassocLevels {
		t.Run(fmt.Sprintf("%02d_%s", l.Level, l.Token), func(t *testing.T) {
			if l.Probe == "" {
				// An empty probe is an ASSERTION, not a skip: it says no
				// source reaches this level. A level with a probe-less
				// entry that nonetheless claims a rejection would be
				// claiming a measurement it never took.
				if l.Reject {
					t.Errorf("claims perl rejects a repetition at this level "+
						"while carrying no probe to repeat.\n\t%s", l.Why)
				}
				unlexable = append(unlexable, fmt.Sprintf("%d %s", l.Level, l.Token))
				return
			}

			compiles, _ := askPerl(t, l.Probe)
			if compiles == l.Reject {
				verb := "accepts"
				if l.Reject {
					verb = "rejects"
				}
				t.Errorf("this table says perl %s %q, and the pinned "+
					"interpreter disagrees.\n\t%s\n\tThe whole point of this "+
					"test is that the answer comes from perl rather than "+
					"from the %%nonassoc keyword; a table that has drifted "+
					"from perl is an inference again.", verb, l.Probe, l.Why)
				return
			}
			measured = append(measured, fmt.Sprintf("%d %s", l.Level, l.Token))
		})
	}

	sort.Strings(measured)
	sort.Strings(unlexable)
	t.Logf("%d of perly.y's 11 %%nonassoc levels measured against perl: %s",
		len(measured), strings.Join(measured, ", "))
	t.Logf("%d carry no lexable operator and are asserted unreachable: %s",
		len(unlexable), strings.Join(unlexable, ", "))
}

// cmpchainProbe is `3 > 2 > 1` and the forced left grouping beside it.
//
// Both spellings in ONE program so the two answers are one measurement.
// Split across two runs they would be two facts about two programs, and
// a parser that got one right and one wrong would satisfy both.
//
// Bracketed for the reason the tier's comparison files are: a false
// comparison stringifies to the EMPTY string, so an unbracketed
// `print ((3>2)>1)` emits a line ending in a space and the repo's
// `trailing-whitespace` hook would quietly strip it out of any pinned
// expectation.
const cmpchainProbe = `print "chain [", (3 > 2 > 1), "] left [", ((3 > 2) > 1), "]\n";`

// cmpchainExpected is what perl 5.42.0 prints for cmpchainProbe.
//
//	$ perl -e 'print "chain [", (3 > 2 > 1), "] left [", ((3 > 2) > 1), "]\n"'
//	chain [1] left []
const cmpchainExpected = "chain [1] left []\n"

// TestTierOperatorsCmpchain measures 5.32+ comparison chaining.
//
// THE CLAIM. `3 > 2 > 1` is 1 under a perl with `cmpchain`, because the
// three operands form ONE n-ary comparison rather than two binary ones.
// Without chaining it would be `(3 > 2) > 1`, which is `1 > 1`, which is
// false -- and false prints as the empty string, so the two readings
// differ in the output by a whole bracket's worth of content.
//
// WHY IT BELONGS IN THIS TIER AND NOT IN THE PARSER'S OWN TESTS. It is
// the purest instance of the thing the tier is named for. Both readings
// are COMPLETE, CONFIDENT PARSES with no Unknown anywhere: a parser
// without chaining builds a perfectly well-formed tree of two binary
// comparisons, reports no refusal, and is WRONG. Nothing in the op
// stream separates them either -- both compile to comparison ops over
// the same three constants. The output is the only witness.
//
// WHY THE PROBE IS CONSTANTS AND NOT `$ARGV[0]`. This is the one place in
// the tier where constant folding is not the enemy. The README's rule --
// every operator needs a runtime operand or the optimiser erases it --
// exists so an op-based claim is not made about an op perl never
// compiled. THIS claim is not about ops. It is about the VALUE perl
// computes, and perl computes it the same way folded or not, so the
// cheapest spelling is the honest one. `askPerl` runs the program rather
// than inspecting its optree, so there is nothing for a fold to hide.
//
// WHY THIS SURVIVES ALONGSIDE `32_chained_comparison.t`. This paragraph
// once said a chaining probe "cannot BE a corpus file at this tier",
// which named the wrong cause: the blocker was that NO TIER CLAIMED
// `cmpchain_and` and `cmpchain_dup`, so any file emitting them failed
// the dependency lint. Tier 04 claims them now and the corpus file
// exists.
//
// The direct measurement stays, and the reason is PLACEMENT rather
// than impossibility -- the earlier version of this comment claimed an
// impossibility twice and was wrong twice. A corpus file CAN pin both
// readings: `print "chain [", (9 < 1 < 5), "] left [", ((9 < 1) < 5), "]"`
// prints `chain [] left [1]` and is a legal tier-04 body.
//
// What this probe has that a corpus file does not is that it runs
// against the pinned interpreter directly, with no pinned output of its
// own to drift. `32_chained_comparison.t` asserts which reading perl
// takes; this asserts the same thing about a different pair of operands
// and fails independently of any corpus expectation.
func TestTierOperatorsCmpchain(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}

	compiles, output := askPerl(t, cmpchainProbe)
	if !compiles {
		t.Fatalf("%s refuses %q, which every perl since 5.32 compiles", perl, cmpchainProbe)
	}
	if output != cmpchainExpected {
		t.Fatalf("pinned %q, %s prints %q.\n"+
			"\t`3 > 2 > 1` is 1 only where comparisons CHAIN: the three "+
			"operands form one n-ary comparison rather than two binary "+
			"ones. Printing the empty string here would mean the pinned "+
			"interpreter read it as `(3 > 2) > 1`, which is `1 > 1`.",
			cmpchainExpected, perl, output)
	}

	// The two readings must actually DIFFER, or the probe measures
	// nothing. A fixture whose two groupings print the same thing is the
	// failure six tiers in this corpus have already shipped, and it is
	// invisible to a pin -- the pin passes, and would go on passing
	// against a parser that grouped the other way.
	chain, left, ok := strings.Cut(strings.TrimSuffix(output, "\n"), "] left ")
	if !ok {
		t.Fatalf("probe output %q does not carry both readings", output)
	}
	if chain == "chain ["+left {
		t.Errorf("the chained and left-grouped readings both print %q.\n"+
			"\tThe probe then discriminates nothing: a parser without "+
			"cmpchain would satisfy it.", left)
	}

	// And our parser must reach a chain node rather than two binaries.
	// The output above is what PERL does; this is the claim about US,
	// and it is the half a behavioural pin cannot make. `chain.go` builds
	// `CmpChain` for the chaining classes, and a tree of nested `Binary`
	// comparisons is precisely the confident-and-wrong parse this tier
	// is named for.
	n := parse.Parse([]byte(cmpchainProbe))
	if chains := countCmpChains(n); chains == 0 {
		t.Errorf("perl chains `3 > 2 > 1` and our parser builds no CmpChain node.\n" +
			"\tA tree of two nested Binary comparisons parses cleanly, " +
			"reports no Unknown, and computes a different answer. That is " +
			"a WRONG parse rather than a refused one, which is what this " +
			"tier exists to catch.")
	}
}

// countCmpChains counts the CmpChain nodes in a tree.
//
// Counted rather than merely detected so the message can say how many: a
// tree with one is the chained reading, and a tree with none is the two
// nested binaries that quietly compute something else.
func countCmpChains(n *parse.Node) int {
	if n == nil {
		return 0
	}
	count := 0
	if n.Kind == parse.CmpChain {
		count++
	}
	for _, c := range n.Children {
		count += countCmpChains(c)
	}
	return count
}

// adjacentPair is one of the 31 adjacent precedence-level pairs, and how
// this corpus accounts for it.
//
// `Loose` and `Tight` name the two levels' operators as they are
// spelled. `Fixture` is a perl program printing the default grouping and
// BOTH forced groupings, or empty when no fixture can be written; `Why`
// then says what stops it.
type adjacentPair struct {
	Loose   int
	Tight   int
	Spell   string
	Fixture string
	Why     string
}

// pairSetup is the operand preamble every pair fixture shares.
//
// `$x` true, `$y` false, `$z` false is the one truth assignment that
// discriminates ALL THREE logical pairs -- found by trying all eight,
// because the first three chosen by eye each printed the same thing for
// both groupings and would have measured nothing. `$n` is 3 and `$m` is
// 2 for the arithmetic and relational pairs.
//
// Every operand comes through `$ARGV[...] //` for the reason the tier
// README states as its operating constraint: with both sides constant the
// optimiser folds the operator out of existence, and a fixture whose
// operator perl never compiled is a fixture about nothing. With no
// arguments the defaults apply, so the output is deterministic, and the
// `//` is also what keeps the run warning-free.
const pairSetup = `my $x = $ARGV[0] // 1; my $y = $ARGV[1] // 0; my $z = $ARGV[2] // 0; ` +
	`my $n = $ARGV[3] // 3; my $m = $ARGV[4] // 2; `

// adjacentPairs accounts for all 31 adjacent precedence pairs of
// perly.y's 32 levels.
//
// EVERY PAIR IS ACCOUNTED FOR; NOT EVERY PAIR CAN BE A FIXTURE, and the
// difference is the honest part of this table. Nine of the 31 have a
// pseudo-token or an XS plugin hook on at least one side -- levels 1, 3,
// 18, 28 and 30 carry no lexable operator at all, which `precedence.go`
// records in `levelsPresent` and which is why those levels appear there
// rather than in `infix`. There is no source that puts two such levels in
// tension, so there is nothing to group two ways and nothing to measure.
//
// Of the 22 that CAN be written, sixteen need an operator this corpus's
// op budget does not reach. `..` compiles to `flip`/`flop`/`range`, `++`
// to `preinc`/`postinc`, `|` to `bit_or`, `<<` to `left_shift`, `?:` to
// `cond_expr`, `=~` to `match` -- and NO TIER IN THE CORPUS CLAIMS ANY OF
// THOSE OPS. A fixture using one would fail `TestTierOperatorsLint`,
// correctly: the lint's whole job is to stop a tier reaching for
// constructs the corpus has not introduced. The bound is the corpus's,
// not perl's, and it moves when a later tier claims those ops.
//
// That leaves SIX pairs measurable here, and all six are. They are the
// six where both levels' operators are ops `04_operators` itself claims.
//
// EACH FIXTURE PRINTS THREE READINGS: the default grouping, then both
// forced groupings. The test below requires the two FORCED readings to
// differ -- a fixture whose groupings print the same thing exercises the
// operators and measures nothing, which is the failure six tiers in this
// corpus have already shipped -- and requires the default to equal the
// tighter-binding one, which is the actual precedence claim.
//
// Every printed value is bracketed. A false comparison and a false `&&`
// stringify to the EMPTY string, so an unbracketed reading ends its line
// in a space and the repo's `trailing-whitespace` hook silently removes
// it. The tier README records this as a trap already paid for once.
var adjacentPairs = []adjacentPair{
	{1, 2, "PREC_LOW / LOOPEX", "",
		"level 1 is a pseudo-token used only via %prec and is never lexed"},
	{2, 3, "LOOPEX / PLUGIN_LOW_OP", "",
		"level 3 is an XS infix-plugin hook with no core operator"},
	{3, 4, "PLUGIN_LOW_OP / or", "",
		"level 3 is an XS infix-plugin hook with no core operator"},

	{4, 5, "or / and",
		pairSetup + `print "[", ($x or $y and $z), "][", (($x or $y) and $z), "][", ($x or ($y and $z)), "]\n";`,
		""},
	// Two things this fixture had to be corrected for, both caught by the
	// checks below rather than by reading.
	//
	// Both operands FALSE, because that is the only shape that
	// discriminates: with a true operand on the right, `not $y and $x`
	// prints 1 under every grouping and the fixture measures nothing.
	//
	// And the two forced readings are in the order the other five use --
	// LOOSER-operator grouping second, TIGHTER third -- which for this
	// pair means `not ($y and $z)` is the loose one. `not` is level 6 and
	// `and` is level 5, so `not` is the TIGHTER of the two despite
	// reading like the outer word; writing the pair the way it reads puts
	// the columns the wrong way round.
	{5, 6, "and / not",
		pairSetup + `print "[", (not $y and $z), "][", (not ($y and $z)), "][", ((not $y) and $z), "]\n";`,
		""},

	{6, 7, "not / LSTOP", "",
		"level 7's list operators are tier 07's and tier 10's subject; `print` appears here only as a fixture"},
	{7, 8, "LSTOP / comma", "",
		"the comma in scalar context is tier 03's construct and its `list` op is tier 03's"},
	{8, 9, "comma / assignment", "",
		"`aassign` and `sassign` are tier 02's ops; a fixture forcing the two groupings is tier 02's to write"},
	{9, 10, "assignment / ternary", "",
		"`?:` compiles to `cond_expr`, which is tier 06's op"},
	{10, 11, "ternary / range", "",
		"`?:` is tier 06's `cond_expr` and `..` compiles to flip/flop/range, which no tier claims"},
	{11, 12, "range / ||", "",
		"`..` compiles to flip/flop/range, which no tier in this corpus claims"},

	{12, 13, "|| / &&",
		pairSetup + `print "[", ($x || $y && $z), "][", (($x || $y) && $z), "][", ($x || ($y && $z)), "]\n";`,
		""},

	{13, 14, "&& / bitwise or", "",
		"`|` compiles to `bit_or`, which no tier in this corpus claims"},
	{14, 15, "bitwise or / bitwise and", "",
		"`|` and `&` compile to `bit_or` and `bit_and`, which no tier claims"},
	{15, 16, "bitwise and / equality", "",
		"`&` compiles to `bit_and`, which no tier in this corpus claims"},

	{16, 17, "equality / relational",
		pairSetup + `print "[", ($m == $n < $n), "][", (($m == $n) < $n), "][", ($m == ($n < $n)), "]\n";`,
		""},

	{17, 18, "relational / PLUGIN_REL_OP", "",
		"level 18 is an XS infix-plugin hook with no core operator"},
	{18, 19, "PLUGIN_REL_OP / UNIOP", "",
		"level 18 is an XS infix-plugin hook with no core operator"},
	{19, 20, "UNIOP / require", "",
		"`require` is tier 12's construct and its op is tier 12's"},
	{20, 21, "require / shift", "",
		"`require` is tier 12's, and `<<` compiles to `left_shift`, which no tier claims"},
	{21, 22, "shift / additive", "",
		"`<<` compiles to `left_shift`, which no tier in this corpus claims"},

	{22, 23, "additive / multiplicative",
		pairSetup + `print "[", ($m + $n * $n), "][", (($m + $n) * $n), "][", ($m + ($n * $n)), "]\n";`,
		""},

	{23, 24, "multiplicative / binding", "",
		"`=~` compiles to `match`, which is tier 09's op"},
	{24, 25, "binding / unary", "",
		"`=~` compiles to `match`, which is tier 09's op"},

	{25, 26, "unary minus / **",
		pairSetup + `print "[", (-$m ** $m), "][", ((-$m) ** $m), "][", (-($m ** $m)), "]\n";`,
		""},

	{26, 27, "** / postfix increment", "",
		"`++` compiles to preinc/postinc, which no tier in this corpus claims"},
	{27, 28, "postfix increment / PLUGIN_HIGH_OP", "",
		"level 28 is an XS infix-plugin hook with no core operator"},
	{28, 29, "PLUGIN_HIGH_OP / arrow", "",
		"level 28 is an XS infix-plugin hook with no core operator"},
	{29, 30, "arrow / PAREN_CLOSE", "",
		"level 30 exists for conflict resolution only and is never an operator"},
	{30, 31, "PAREN_CLOSE / call", "",
		"level 30 exists for conflict resolution only and is never an operator"},
	{31, 32, "call / subscript", "",
		"a postfix call is tier 07's construct; its `entersub` is tier 07's op"},
}

// TestTierOperatorsAdjacentPairs accounts for all 31 adjacent precedence
// pairs and measures every one that can be measured.
//
// THE CLAIM, and why it is not "31 fixtures". The issue asks that the 31
// adjacent pairs be fixtures whose two groupings produce different
// observable output. Measured, that is not writable: nine of the 31 have
// a pseudo-token or a plugin hook on one side and NO SOURCE REACHES THEM,
// and sixteen more need an operator whose op no tier in this corpus
// claims, so a fixture for one would fail `TestTierOperatorsLint`.
// Writing them anyway would mean either widening a tier's INTRODUCES to
// admit ops nothing measures, or pinning a fixture that cannot run. Both
// are how this corpus drifts.
//
// So the table accounts for all 31 and measures the six that are
// reachable, and the unreachable ones carry a REASON rather than a
// silence. That is the same shape `precedence.go`'s `levelsPresent`
// takes for the 32 levels, and for the same stated reason: a count of
// entries passes while a level is missing, and an explicit "this one
// cannot be written, here is what stops it" does not.
//
// WHAT EACH FIXTURE ASSERTS, in order of strength:
//
//   - THE TWO FORCED GROUPINGS DIFFER. Without this the fixture measures
//     nothing: it exercises both operators and would go on passing
//     against a parser that grouped the other way. Three of the six
//     operand choices tried first FAILED this -- `$b or $a and $b` with
//     `$a` true and `$b` false prints `0` under every grouping -- which
//     is exactly the failure six tiers in this corpus already shipped,
//     and it was invisible until something asked.
//   - THE DEFAULT MATCHES THE TIGHTER-BINDING GROUPING. This is the
//     precedence claim itself, and it is only meaningful because of the
//     line above.
//
// WHY THIS IS THE TIER'S CENTRAL TEST. Every earlier tier fails by
// producing an `Unknown`. A mis-grouped expression produces no Unknown at
// all: the tree is complete, the parser is confident, and the answer is
// wrong. `canon.go` reads the `infix[]` table at ten sites, so one wrong
// binding power mis-groups on the way in and un-mis-groups on the way
// out, and a round-trip check cannot see it. A DIFFERENCE IN PRINTED
// OUTPUT BETWEEN TWO GROUPINGS is the only witness that survives that.
func TestTierOperatorsAdjacentPairs(t *testing.T) {
	const perlyLevels = 32
	if want := perlyLevels - 1; len(adjacentPairs) != want {
		t.Fatalf("perly.y has %d levels and so %d adjacent pairs; this table holds %d",
			perlyLevels, want, len(adjacentPairs))
	}

	// The pairs must be the 31 CONSECUTIVE ones, each once. A table that
	// listed level 4-5 twice and skipped 5-6 would reach 31 entries with
	// a pair unaccounted for, which is the failure the count above is
	// meant to catch arriving through the back door.
	for i, p := range adjacentPairs {
		if p.Loose != i+1 || p.Tight != i+2 {
			t.Fatalf("entry %d is %d-%d; the table must run 1-2 through 31-32 in order",
				i, p.Loose, p.Tight)
		}
	}

	var measured, unreachable []string
	for _, p := range adjacentPairs {
		t.Run(fmt.Sprintf("%02d_%02d", p.Loose, p.Tight), func(t *testing.T) {
			if p.Fixture == "" {
				// A missing fixture is an ASSERTION, not a skip: the
				// reason is the claim, and an entry with neither is a
				// pair nothing accounts for.
				if strings.TrimSpace(p.Why) == "" {
					t.Errorf("%d-%d (%s) has no fixture and no reason.\n"+
						"\tA pair that cannot be measured has to say what "+
						"stops it, or the table records a silence as an "+
						"answer.", p.Loose, p.Tight, p.Spell)
				}
				unreachable = append(unreachable, fmt.Sprintf("%d-%d", p.Loose, p.Tight))
				return
			}
			if p.Why != "" {
				t.Errorf("%d-%d (%s) carries both a fixture and a reason it "+
					"cannot be written; one of the two is stale",
					p.Loose, p.Tight, p.Spell)
				return
			}

			compiles, output := askPerl(t, p.Fixture)
			if !compiles {
				t.Errorf("%d-%d (%s): perl -c refuses the fixture", p.Loose, p.Tight, p.Spell)
				return
			}

			def, loose, tight, err := threeReadings(output)
			if err != nil {
				t.Errorf("%d-%d (%s): %v\n\tsource: %s", p.Loose, p.Tight, p.Spell, err, p.Fixture)
				return
			}

			// The discrimination check, and the reason this test is worth
			// more than a pinned output would be.
			if loose == tight {
				t.Errorf("%d-%d (%s): both forced groupings print %q, so the "+
					"fixture measures nothing.\n"+
					"\tIt exercises both operators and would go on passing "+
					"against a parser that grouped the other way. Choose "+
					"operands that make the two readings differ.",
					p.Loose, p.Tight, p.Spell, loose)
				return
			}

			// The precedence claim: the tighter level binds first, so the
			// unparenthesised reading is the one that groups the tighter
			// operator's operands together.
			if def != tight {
				t.Errorf("%d-%d (%s): level %d binds tighter, so the default "+
					"grouping should print %q; perl prints %q, which is the "+
					"looser grouping.\n\tsource: %s",
					p.Loose, p.Tight, p.Spell, p.Tight, tight, def, p.Fixture)
				return
			}
			measured = append(measured, fmt.Sprintf("%d-%d %s", p.Loose, p.Tight, p.Spell))
		})
	}

	t.Logf("%d of 31 adjacent pairs measured with discriminating fixtures:\n\t%s",
		len(measured), strings.Join(measured, "\n\t"))
	t.Logf("%d accounted for as unreachable: %s",
		len(unreachable), strings.Join(unreachable, ", "))
}

// threeReadings splits a pair fixture's output into its default, forced
// looser and forced tighter readings.
//
// Each reading is bracketed, so a false one -- which stringifies to the
// empty string -- is still a distinguishable `[]` rather than a gap. That
// is also what keeps the fixture's own output free of a line-ending
// space, which the repo's `trailing-whitespace` hook would strip.
func threeReadings(output string) (def, loose, tight string, err error) {
	parts := strings.Split(strings.TrimSuffix(output, "\n"), "][")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf(
			"a pair fixture prints exactly three bracketed readings -- default, "+
				"forced-loose, forced-tight -- and this printed %q", output)
	}
	return strings.TrimPrefix(parts[0], "["),
		parts[1],
		strings.TrimSuffix(parts[2], "]"), nil
}

// TestTierOperatorsRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// The reasoning is tier 01's and transfers exactly, so it is stated
// briefly here and in full there. `parse.RefusalCode` names a site in the
// PARSER. A file refusing purely on a token fact has no parser Unknown to
// name, and naming one anyway makes `run.go` fail it as a stale marker --
// correctly, because the claim would be false. So the rule is
// biconditional rather than "every refusing file names a code":
//
//   - A file whose parser produces Unknowns MUST name one of their codes.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// THIS TIER HAS FIVE REFUSING FILES, more than any other in the corpus,
// and the biconditional is what keeps them from being lumped together --
// because they do NOT refuse for the same reason. Four refuse purely
// LEXICALLY: perl spells some operators with letters (`x`, `eq`, `cmp`,
// `and`, `xor`) and our lexer gives all of them the kind `Word` where the
// glossary calls them operators. Those files have no parser Unknown at
// all, so they must NOT cite a code, and their `--- expect tokens`
// section IS their refusal. The fifth, `00_adjacency.t`, refuses in the
// PARSER -- three Unknown nodes over a body whose every construct parses
// in a sibling file, which is the adjacency argument working.
//
// A file that cited a code for a lexical refusal would be making a claim
// about the parser that the parser contradicts, and this tier is where
// that mistake is easiest to make: `06_and_cliff.t`'s whole subject is
// that `&&` and `and` are the same operator, and it would be natural to
// read the lexer's disagreement as a parse failure. It is not one.
func TestTierOperatorsRefusalsCited(t *testing.T) {
	var lexical, parser []string

	for name, f := range tierFiles(t, tierOperators) {
		if f.Refuses == "" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			codes := refusalCodes(parse.Parse([]byte(f.Source)))

			if len(codes) == 0 {
				if f.RefusalCode != "" {
					t.Errorf("names Refusal %s, but the parser returns no "+
						"Unknown at all.\n\tThere is no site to name; the "+
						"refusal is lexical.", f.RefusalCode)
				}
				if len(f.TokenFacts) == 0 {
					t.Errorf("refuses (%s) with no parser Unknown and no "+
						"`--- expect tokens` section.\n\tNothing in this "+
						"file measures the refusal it documents.", f.Refuses)
				}
				lexical = append(lexical, name)
				return
			}

			if f.RefusalCode == "" {
				t.Errorf("refuses (%s) with codes %s but names none.\n"+
					"\tAdd a `Refusal <code>.` clause to the STATUS line -- "+
					"a code is a stable identifier where a message is prose.",
					f.Refuses, joinCodes(codes))
				return
			}
			if _, ok := parse.RefusalSites[f.RefusalCode]; !ok {
				t.Errorf("names Refusal %s, which is not in parse.RefusalSites.\n"+
					"\tThe inventory is the vocabulary; a code outside it is a "+
					"message wearing a code's spelling.", f.RefusalCode)
			}
			if !hasCode(codes, f.RefusalCode) {
				t.Errorf("names Refusal %s, but the parser refuses with %s",
					f.RefusalCode, joinCodes(codes))
			}
			parser = append(parser, name)
		})
	}

	sort.Strings(lexical)
	sort.Strings(parser)
	t.Logf("%d refuse lexically (no parser Unknown, token facts are the claim): %s",
		len(lexical), strings.Join(lexical, ", "))
	t.Logf("%d refuse in the parser (an Unknown, so a code is required): %s",
		len(parser), strings.Join(parser, ", "))
}
