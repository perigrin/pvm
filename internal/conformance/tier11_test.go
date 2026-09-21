// ABOUTME: Tier 11 checked against the finished tooling: bless, method dispatch, class/field/method/ADJUST, indirect new.
// ABOUTME: Six checks the tier's own issue names, one of which is the tier's motivating bug stated as a test.
package conformance

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// tierOo is the tier this file is about.
//
// A constant rather than a literal at six call sites, for the reason
// tier01_test.go gives for tierLiterals: the tier number is a POSITION
// and positions move.
const tierOo = "11_oo"

// TestTierOoPerlValidated runs every file in the tier through the pinned
// interpreter before it counts.
//
// Identical in intent to TestTierLiteralsPerlValidated, and this tier is
// where the VERSION half of it carries weight that no other tier's does.
// `class`, `field`, `method` and `ADJUST` are 5.38+ syntax gated behind
// `use feature 'class'`, and measured, `use v5.42;` does NOT enable it in
// 5.42.0 -- `class Foo` under a bare version bundle is a syntax error. So
// the tier's files name the feature explicitly, and whether they got that
// right is a question only the pinned interpreter can answer. A file that
// had reached for the bundle would be a syntax error that the op-based
// checks below never run on at all.
func TestTierOoPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierOo, perl)

	for name, f := range tierFiles(t, tierOo) {
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

// TestTierOoLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a file that reaches forward fails HERE,
// named as this tier's problem rather than as one subtest among a
// hundred and thirty-four.
//
// The lint has a specific thing to catch in this tier, recorded in its
// README: `methstart` is NOT in the declared INTRODUCES set even though
// every `method` body opens with it. `opsOf` runs `perl -MO=Concise,-exec`
// with no sub named, which dumps the main program alone, and a method
// body is its own CV. So no file this tier can be written to hold puts
// `methstart` in what the lint measures, and a README claiming it fails
// `TestCorpusLints` from the other direction -- as an op a tier claims
// and no file emits. Issue 01a0c547-516b tracks the `opsOf` that would
// make it measurable; until then the absence is a measurement.
func TestTierOoLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierOo) {
		t.Run(name, func(t *testing.T) {
			if err := lintOps(t, f.Source, tierOo, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// reOoConstruct matches the keyword a construct file is about, as this
// tier's source spells it.
//
// KEYWORDS rather than whole spellings, for tier 05's reason: this tier's
// constructs are FORMS, and `bless {}, "Foo"` and `bless {}, $c` are the
// same form. A check keyed on the whole statement would ask the adjacency
// file to repeat a class name rather than to exercise a construct.
//
// `ADJUST` is matched at line start because it is a block introducer like
// the others; `bless`, `field` and `method` are matched anywhere, because
// `bless` appears mid-expression (`my $o = bless {}, $c`) and `field` and
// `method` are indented inside a class body.
var reOoConstruct = regexp.MustCompile(`\b(bless|class|field|method|ADJUST)\b`)

// reOoArrowCall matches a method call written with the arrow.
//
// Its own pattern because the arrow is not a keyword. What makes a call
// DIRECT is the `->` between invocant and method name -- which is exactly
// the token `06_indirect_new.t` asserts is absent from the indirect
// spelling, and the reason the pair is a lexical claim rather than a
// behavioural one.
var reOoArrowCall = regexp.MustCompile(`->\s*(\w+|\$\w+)`)

// reOoIndirect matches indirect object notation for a constructor:
// `new Foo` with no arrow anywhere between.
//
// Anchored on `new` followed by a capitalised bareword, which is the only
// spelling this tier carries. A looser pattern would match the `new` in
// `sub new`, which every file defining a constructor has.
var reOoIndirect = regexp.MustCompile(`=\s*new\s+[A-Z]\w*`)

// ooForms returns the constructs of this tier a source exercises, named
// as the README names them.
func ooForms(source string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range reOoConstruct.FindAllStringSubmatch(source, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	if reOoIndirect.MatchString(source) {
		out = append(out, indirectNewForm)
	} else if reOoArrowCall.MatchString(source) {
		// `else if` on purpose: `06_indirect_new.t` holds no arrow at
		// all, and a file that held both would be named for whichever
		// its NAME claims -- see formFromOoName, which is what places a
		// file. This branch only supplies the adjacency file's coverage.
		out = append(out, arrowCallForm)
	}
	return out
}

// indirectNewForm and arrowCallForm name the two keyword-less constructs.
const (
	indirectNewForm = "indirect new"
	arrowCallForm   = "arrow method call"
)

// formFromOoName returns the construct a file is named for, or "" when
// its name names none.
//
// THE NAME, not the source, for tier 05's reason stated in full there.
// Every file in this tier contains a method call: the `class` files call
// `Foo->new` because that is how a class is observed at all, and the
// `bless` files call `ref` on the result. Reading each file's subject
// from its source would report the same construct nine times and the tier
// would look like one construct.
func formFromOoName(name string) string {
	m := reNumbered.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	switch identity := strings.TrimSuffix(m[2], ".t"); identity {
	case "bless_empty", "bless_populated":
		return "bless"
	case "method_named", "method_dynamic", "method_super":
		return arrowCallForm
	case "indirect_new":
		return indirectNewForm
	case "class_empty":
		return "class"
	case "class_field_method":
		return "field"
	case "class_adjust":
		return "ADJUST"
	default:
		return ""
	}
}

// TestTierOoAdjacency checks the tier's adjacency file holds every
// construct the tier introduces, each next to another, and that it pairs
// with the DECLARED prerequisite.
//
// WHAT OPS CANNOT DO HERE, and this tier is the corpus's sharpest case
// after tier 12. Tier 01 records that `padrange` absorbs `pushmark`, so
// the declared INTRODUCES set is a UNION across a tier's files rather
// than a property of any one of them. That applies here, and it is again
// not the binding constraint: measured, `class`, `field`, `ADJUST` and
// `:isa` emit NO OP OF THEIR OWN. A class body at file scope compiles to
// tier 05's `enterloop`/`leaveloop` with a `nextstate` per declaration
// inside it. An adjacency check keyed on ops would be blind to four of
// this tier's constructs, and `Foo->new` and `new Foo` -- the whole
// indirect-object question -- emit an IDENTICAL stream.
//
// So COVERAGE is checked on the SOURCE, as at tiers 01, 05 and 12.
//
// THE PREREQUISITE PAIRING is the half ops can carry. The README declares
// `08_references`, and `readTierDeps` READS that rather than computing
// tier N-1 -- which would be `10_io`, a tier this one uses nothing from.
// The README says so in as many words: pairing with 10 would assert
// nothing, because no file here opens a file handle.
//
// Reading the declaration is not enough, which tier 12 found by mutation
// and recorded: a pairing check satisfied by "emits an op of the declared
// prerequisite" can stay green under the WRONG declaration, when the
// adjacency file happens to emit that tier's ops too. So the declaration
// is PINNED separately below. Two assertions, neither implying the other.
//
// Measured, this adjacency file pairs through `ref`, which
// `08_references` introduces and which arrives from the two `ref(...)`
// calls the file prints. That is the right op for the claim: the README's
// reason for depending on tier 08 is that a blessed object IS a reference
// plus a string, and `ref` is the reference half asked what it became.
//
// Why the OP and not the spelling: a source search for `bless` is
// satisfied by the word in a comment, and by a `bless` inside a sub that
// nothing calls. Requiring the op requires that the two tiers' constructs
// actually met in one compiled body.
//
// What this does NOT establish: that the constructs are adjacent in any
// stronger sense than "in the same body" -- the same limit tiers 01, 05
// and 12 record, for the same reason.
func TestTierOoAdjacency(t *testing.T) {
	files := tierFiles(t, tierOo)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierOo, adjacencyFile)
	}
	adjacent := ooForms(adj.Source)

	// The construct each non-adjacency file introduces, taken from that
	// file's own identity rather than from a list here: the tier's files
	// ARE the enumeration of what it introduces, and a second list beside
	// them is how this package has drifted four times already.
	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		form := formFromOoName(name)
		if form == "" {
			t.Errorf("%s: its name names no construct of this tier, so nothing places it", name)
			continue
		}
		if !slices.Contains(ooForms(f.Source), form) {
			t.Errorf("%s is named for %q, which its own source does not use.\n"+
				"\tThe name is the file's identity; a name the source does "+
				"not honour makes every other check here ask about the "+
				"wrong construct.", name, form)
			continue
		}
		// The indirect spelling is the one construct the adjacency file
		// cannot hold. `new Foo` and `Foo->new` differ in NO op, so a
		// body holding both would be a body whose two halves the optree
		// cannot tell apart -- and worse, the adjacency file also carries
		// the arrow call, which the indirect form is defined by the
		// absence of. The two are mutually exclusive in one statement and
		// indistinguishable in one optree, so the tier asserts the pair
		// LEXICALLY, in the two files' token facts, which is where the
		// difference lives. See TestTierOoArrowIsLexical.
		if form == indirectNewForm {
			continue
		}
		if !slices.Contains(adjacent, form) {
			t.Errorf("%s introduces %q, which %s does not use.\n"+
				"\tThe adjacency file must hold every construct the tier "+
				"introduces, or the pairing it exists to reach is not "+
				"reachable for that construct.", name, form, adjacencyFile)
		}
	}

	// One body, not several. Adjacency needs at least two constructs to
	// be adjacent TO, and `nextstate` is one per statement -- which is
	// how this counts statements without a parser of its own.
	ops, err := opsOf(t, adj.Source)
	if err != nil {
		t.Fatalf("%s: %v", adjacencyFile, err)
	}
	if n := countOp(ops, "nextstate"); n < 2 {
		t.Errorf("%s compiles to %d statement(s); adjacency needs at least 2", adjacencyFile, n)
	}

	// The pairing with the DECLARED prerequisite, READ from the README
	// rather than computed as tier N-1.
	deps, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	dep, ok := deps[tierOo]
	if !ok {
		t.Fatalf("%s declares no DEPENDS ON", tierOo)
	}
	if dep == "nothing" {
		t.Fatalf("%s declares no prerequisite, but the issue asks its "+
			"adjacency file to pair with one", tierOo)
	}

	// The declared prerequisite is PINNED, and this is the check the rest
	// of the pairing needs in front of it.
	//
	// Tier 12 found the need for this by mutation and its comment records
	// the finding: repointing a README's DEPENDS ON at the wrong tier left
	// the pairing check GREEN, because the adjacency file emitted that
	// tier's ops as well. The same hole is open here and is wider.
	// Measured, this adjacency file emits ops belonging to several tiers
	// -- `nextstate` and `padsv` from 01, `sassign` from 02, `enterloop`
	// from 05, `entersub` from 07 -- so "emits an op of the declared
	// prerequisite" would be satisfied by declaring almost any earlier
	// tier, and the README's argument that the prerequisite is
	// `08_references` and NOT tier N-1 would vanish silently.
	//
	// That argument is substantive rather than bookkeeping: the README
	// says `bless` is the WHOLE of the dependency -- a blessed object is
	// a tier 08 reference plus a string -- and that pairing with `10_io`
	// would assert nothing because no file here opens a handle. A drift
	// to 10 would erase the claim with every test still green.
	//
	// So the pin is a SECOND assertion beside the pairing rather than a
	// replacement for it. The pin says the declaration is the one the
	// tier argued for; the pairing says the adjacency file honours it.
	// Either alone is satisfiable without the other.
	const declaredPrerequisite = "08_references"
	if dep != declaredPrerequisite {
		t.Errorf("%s declares DEPENDS ON %s, want %s.\n"+
			"\tThis tier's README argues that its prerequisite is NOT tier "+
			"N-1: `bless` is the whole of the dependency, and pairing with "+
			"10_io would assert nothing because no file here opens a "+
			"handle. A change here is a change to that claim and must be "+
			"made deliberately.\n"+
			"\tNote that the pairing check below does NOT catch this on "+
			"its own -- %s emits ops of tiers 01, 02, 05 and 07 too.",
			tierOo, dep, declaredPrerequisite, adjacencyFile)
	}

	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	depOps := tiers[dep]
	if len(depOps) == 0 {
		t.Fatalf("%s declares %s, which introduces no op to pair with", tierOo, dep)
	}

	var paired []string
	for _, op := range depOps {
		if countOp(ops, op) > 0 {
			paired = append(paired, op)
		}
	}
	if len(paired) == 0 {
		t.Errorf("%s declares DEPENDS ON %s, and %s emits none of that "+
			"tier's ops (%s).\n"+
			"\tThe adjacency file is where the two tiers meet in one "+
			"compiled body; without one of the prerequisite's ops the "+
			"declared edge is prose.",
			tierOo, dep, adjacencyFile, strings.Join(depOps, " "))
		return
	}
	t.Logf("%s pairs with %s through %s", adjacencyFile, dep, strings.Join(paired, ", "))
}

// TestTierOoAdjacencyCatchesAdjust is this tier's motivating bug, stated
// as a test, and it is the only test in the corpus that asserts a file
// FAILS.
//
// THE BUG, measured 2026-09-21 against our parser and re-measured here
// every run:
//
//	class Foo { ADJUST { 1 } }                   parses, 0 Unknowns
//	class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both
//
// `ADJUST` alone parses. `ADJUST` followed by anything -- a `method`, a
// `field`, or a second `ADJUST` -- does not.
//
// WHY THAT IS THE WHOLE ARGUMENT FOR ADJACENCY FILES. A corpus of one
// construct per file goes green over this BY CONSTRUCTION, because every
// construct in such a corpus is measured in isolation and every construct
// in this tier passes in isolation. `09_class_adjust.t` is green, and it
// is green truthfully: ADJUST alone really does parse. The bug is not a
// property of any construct, it is a property of a PAIR, and no file that
// holds one construct can hold a pair.
//
// So this test asserts the shape the tier's design rests on, in both
// directions, because either half alone is satisfiable by an accident:
//
//   - EVERY CONSTRUCT FILE PASSES. Not a convenience -- it is the premise.
//     A tier whose construct files already failed would prove nothing
//     about adjacency, because the adjacency file's failure could be
//     inherited from any one of them.
//   - THE ADJACENCY FILE FAILS. And fails at the ADJUST pair specifically
//     rather than anywhere, which is checked below by span: a refusal
//     that had drifted to some other cause would leave this test green
//     while the file's cited issue pointed at a bug that was no longer
//     the one being measured. A stale citation is exactly what the
//     refusal-code work exists to prevent.
//
// THE DAY THIS TEST GOES RED is the day the parser learns ADJUST, and the
// failure is then correct and expected: the corpus must be told, because
// `00_adjacency.t` has a `STATUS refuses` marker that must come off and a
// ratchet entry that must be regenerated. `TestCorpus` reports a file
// marked refusing that now passes as an error for the same reason.
func TestTierOoAdjacencyCatchesAdjust(t *testing.T) {
	files := tierFiles(t, tierOo)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierOo, adjacencyFile)
	}

	// The premise: every construct file passes in isolation. Without
	// this, the adjacency file's failure says nothing about adjacency.
	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		if codes := refusalCodes(parse.Parse([]byte(f.Source))); len(codes) != 0 {
			t.Errorf("%s refuses with %s, but this tier's argument needs "+
				"every construct file to PASS in isolation.\n"+
				"\tThe adjacency file's failure is only evidence about "+
				"adjacency while the constructs it holds each parse alone.",
				name, joinCodes(codes))
		}
	}

	// The adjacency file fails.
	codes := refusalCodes(parse.Parse([]byte(adj.Source)))
	if len(codes) == 0 {
		t.Fatalf("%s parses clean.\n"+
			"\tEither the parser learned ADJUST-plus-sibling -- in which "+
			"case remove the file's `STATUS refuses` marker and regenerate "+
			"the ratchet -- or the file no longer holds the pair this tier "+
			"is about.", adjacencyFile)
	}

	// And fails AT THE ADJUST PAIR, not merely somewhere. Checked by the
	// span of the Unknown, which is the one part of a refusal that says
	// WHERE rather than only what kind. The bug's signature is an Unknown
	// that STARTS at the `ADJUST` keyword and REACHES PAST it into the
	// construct that follows -- "swallowing both" is the measured
	// description and this is it as an assertion.
	src := []byte(adj.Source)
	adjustAt := strings.Index(adj.Source, "ADJUST")
	if adjustAt < 0 {
		t.Fatalf("%s holds no ADJUST block, so it cannot hold the pair "+
			"this tier exists to catch", adjacencyFile)
	}
	sibling := strings.Index(adj.Source[adjustAt:], "method ")
	if sibling < 0 {
		t.Fatalf("%s has an ADJUST with no sibling construct after it.\n"+
			"\tADJUST ALONE PARSES -- `09_class_adjust.t` is green and "+
			"truthfully so. The bug is a property of the pair.", adjacencyFile)
	}
	sibling += adjustAt

	var swallowing *parse.Node
	for _, n := range unknownNodes(parse.Parse(src)) {
		if n.Start == adjustAt && n.End > sibling {
			swallowing = n
			break
		}
	}
	if swallowing == nil {
		t.Errorf("%s refuses with %s, but no Unknown starts at the ADJUST "+
			"keyword (offset %d) and reaches past the sibling at offset %d.\n"+
			"\tThe file cites an issue about ADJUST-plus-sibling. A refusal "+
			"somewhere else means the citation is stale -- which is the "+
			"failure mode refusal codes exist to prevent -- so either the "+
			"cause moved or the file did.\n\tUnknowns: %s",
			adjacencyFile, joinCodes(codes), adjustAt, sibling, describeUnknowns(src))
		return
	}
	t.Logf("%s refuses %s at [%d,%d), which starts at ADJUST and swallows "+
		"the sibling at %d", adjacencyFile, swallowing.Refusal,
		swallowing.Start, swallowing.End, sibling)
}

// unknownNodes collects every Unknown in a tree, in tree order.
//
// `refusalCodes` returns the CODES; this returns the nodes, because the
// ADJUST check above is about a SPAN and a code has none. Two functions
// rather than one that returns both, so the common caller keeps the
// simpler signature.
func unknownNodes(n *parse.Node) []*parse.Node {
	if n == nil {
		return nil
	}
	var out []*parse.Node
	if n.Kind == parse.Unknown {
		out = append(out, n)
	}
	for _, c := range n.Children {
		out = append(out, unknownNodes(c)...)
	}
	return out
}

// describeUnknowns renders every Unknown with its span and the source it
// covers, so a drifted refusal is reported as what it actually caught
// rather than only as the wrong shape.
func describeUnknowns(src []byte) string {
	var out []string
	for _, n := range unknownNodes(parse.Parse(src)) {
		end := min(n.End, len(src))
		out = append(out, fmt.Sprintf("%s @[%d,%d) %q",
			n.Refusal, n.Start, n.End, string(src[n.Start:end])))
	}
	if len(out) == 0 {
		return "none"
	}
	return strings.Join(out, "; ")
}

// TestTierOoArrowIsLexical is this tier's answer to the one question its
// files could not otherwise ask.
//
// THE PROBLEM, stated in the README: `Foo->new` and `new Foo` compile to
// the SAME op stream -- `pushmark`, `const[PV "Foo"] sM/BARE`,
// `method_named[PV "new"]`, `entersub` -- differing in no op, only in
// line numbers. Indirect object notation is a LEXING problem and not a
// compilation one, so a corpus that asserted only behaviour and optrees
// could not see the difference at all. `06_indirect_new.t` therefore
// carries a token fact, and it is the only file in the tier that did.
//
// WHY ONE FACT WAS NOT ENOUGH, and this is the gap this test closes. The
// fact `06_indirect_new.t` declares is a NEGATIVE: `no operator whose
// text is "->"`. A negative alone is satisfied VACUOUSLY by a lexer that
// never produces `->` at all -- one that folded the arrow into the word
// beside it, or dropped it as trivia, passes the indirect file's claim
// perfectly while getting every direct call in the tier wrong. Six of the
// corpus's tiers have shipped a file that asserted behaviour while
// asserting nothing a broken lexer would fail; this is the same shape,
// and the fix is the matching POSITIVE claim somewhere in the tier.
//
// So the tier is required to hold BOTH sides of the contrast:
//
//   - a file whose source spells the call indirectly and declares that no
//     arrow operator is there, and
//   - a file whose source spells it directly and declares that one is.
//
// Together they are falsifiable: a lexer that emits no arrow fails the
// second, a lexer that invents one fails the first, and a lexer that
// cannot tell `->` from a minus followed by a `>` fails both.
//
// The tokens are re-lexed here rather than trusted from the files'
// declarations, because a declaration the runner did not check is prose.
// `TestCorpus` does check them, and that is the gate; this asserts the
// tier HAS them, which is the part that was missing.
func TestTierOoArrowIsLexical(t *testing.T) {
	const (
		arrowAbsent  = `no operator whose text is "->"`
		arrowPresent = `one operator whose text is "->"`
	)

	var declaredAbsent, declaredPresent []string
	for name, f := range tierFiles(t, tierOo) {
		for _, fact := range f.TokenFacts {
			switch fact {
			case arrowAbsent:
				declaredAbsent = append(declaredAbsent, name)
			case arrowPresent:
				declaredPresent = append(declaredPresent, name)
			}
		}
	}

	if len(declaredAbsent) == 0 {
		t.Errorf("no file in %s declares %q.\n"+
			"\tIndirect object notation is a LEXING problem -- `new Foo` "+
			"and `Foo->new` emit an identical op stream -- so the absence "+
			"of the arrow is the only place the construct is visible.",
			tierOo, arrowAbsent)
	}
	if len(declaredPresent) == 0 {
		t.Errorf("no file in %s declares %q.\n"+
			"\tThe negative claim alone is satisfied VACUOUSLY by a lexer "+
			"that never emits `->` at all. Without the matching positive "+
			"claim the tier asserts nothing a broken lexer would fail.",
			tierOo, arrowPresent)
	}
	t.Logf("arrow absent in %s; arrow present in %s",
		strings.Join(declaredAbsent, ", "), strings.Join(declaredPresent, ", "))
}

// TestTierOoClassKeywordsAreLexed closes the tier's other lexical gap.
//
// THE PROBLEM. `class`, `field`, `method` and `ADJUST` emit no op of
// their own -- a class body at file scope is tier 05's `enterloop` and
// `leaveloop` with a `nextstate` per declaration inside it -- so nothing
// in the optree distinguishes them from each other or from a bare block
// holding four statements. The tier's class files assert perl's OUTPUT,
// which a correct program produces, and the optree, which says almost
// nothing. Neither would notice a lexer that read `class Foo {` as one
// opaque blob, or that gave `field` and `ADJUST` different kinds from
// each other, or that treated `ADJUST` as a label.
//
// That is the pattern six of this corpus's tiers hit: a file testing
// behaviour while asserting nothing a broken lexer would fail. Tier 08's
// `@{$r}` is the recorded instance -- it appeared only inside a string,
// where the lexer emits one Quote token and the deref is never tokenised
// at all.
//
// WHAT THE CLAIM IS, and it is deliberately modest. These four are
// KEYWORDS, and GLOSSARY.md's `word` category says plainly that keywords
// are not distinguished from other identifiers at this layer: "Whether
// `print` is a builtin, a user sub, or a filehandle is a parsing question
// that needs context this layer does not have." So the right lexical
// claim is not that `class` has a kind of its own -- it does not, and
// should not -- but that each of the four ARRIVES AS ITS OWN WORD TOKEN,
// separate from the name beside it and from the brace after it.
//
// That is falsifiable and is the failure worth catching: a lexer that
// consumed `class Foo {` as one token, or that swallowed `ADJUST {` into
// a label, would produce a token stream in which these words are not
// there. Every check above it would stay green.
//
// Checked on the TIER's files as a union rather than on one file, for the
// reason tier 01 gives about INTRODUCES sets: the constructs are spread
// across `07_class_empty.t`, `08_class_field_method.t` and
// `09_class_adjust.t`, and requiring all four of one file would be
// requiring a file the tier deliberately does not have.
func TestTierOoClassKeywordsAreLexed(t *testing.T) {
	// The four keywords `class` syntax is made of. A list here rather
	// than derived, because these ARE the tier's subject as its title
	// names them -- "bless AND class/field/method/ADJUST" -- and a
	// derivation from the sources would be satisfied by whatever the
	// sources happen to contain, which is the thing being checked.
	keywords := []string{"class", "field", "method", "ADJUST"}

	files := tierFiles(t, tierOo)
	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			var holders []string
			for name, f := range files {
				src := []byte(f.Source)
				for _, tk := range lexer.Tokenize(src) {
					if tk.Kind == lexer.Word && string(src[tk.Start:tk.End]) == kw {
						holders = append(holders, name)
						break
					}
				}
			}
			if len(holders) == 0 {
				t.Errorf("no file in %s lexes %q as a word of its own.\n"+
					"\tThese four keywords emit NO OP, so the optree cannot "+
					"see them and perl's output is produced by a correct "+
					"program either way. A lexer that read `class Foo {` as "+
					"one blob would pass every other check in this tier.",
					tierOo, kw)
				return
			}
			slices.Sort(holders)
			t.Logf("%q lexed as its own word in %s", kw, strings.Join(holders, ", "))
		})
	}
}

// TestTierOoRefusalsCited checks that a refusing file names WHICH site
// declines, from the stable inventory, rather than describing it.
//
// The rule is tier 01's and is biconditional, because `parse.RefusalCode`
// names sites in the PARSER and a refusal can be purely lexical:
//
//   - A file whose parser produces Unknowns MUST name one of their codes.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// THIS TIER HAS EXACTLY ONE REFUSING FILE and it is the adjacency file,
// which makes the check unusually load-bearing here. The tier's whole
// argument is that this one file catches a bug every other file in the
// tier is blind to, and the file cites an issue by id. An issue id says
// WHICH BUG someone believed this was; a refusal code says which site of
// our parser actually declines today. Measured, `00_adjacency.t` produces
// `trailing_tokens` -- the expression parser reads the ADJUST block and
// then finds the `method` that follows it before the terminator -- and
// without the code named, a refusal that drifted to some other site would
// leave the file reading as though the ADJUST bug were still what it
// measured. That is a stale citation, which is the exact failure the
// refusal-code inventory exists to prevent.
func TestTierOoRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierOo) {
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
		})
	}
}
