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
			if err := lintFile(t, f, tierOo, tiers); err != nil {
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
// `tied` precedes `tie` in the alternation. Measured, the `\b` on both
// ends already prevents `tie` from matching inside `tied`, so the order
// changes nothing today -- but the longer-first spelling is what stays
// correct if the boundary is ever relaxed, and it costs nothing.
var reOoConstruct = regexp.MustCompile(`\b(bless|class|field|method|ADJUST|tied|tie)\b`)

// ungatedHalf reports whether a tier-11 file pins what a construct means
// with its feature DISABLED (issue 01a0cc04-339f).
//
// DERIVED FROM THE NAME rather than listed, and the difference matters.
// A hand-maintained list suppresses a check in two places and nothing
// forces a future ungated file into it or flags a stale entry -- so the
// list drifts silently and the checks quietly stop covering what they
// were written for. Tier 07's `parsent` exemption reads the file itself
// (`f.ExpectParsent`) and cannot drift; this now has the same property.
//
// The suffix is not a convention invented for the check: every file of
// this kind is named for the construct plus `_ungated`, because that is
// what distinguishes it from the gated half sitting beside it.
func ungatedHalf(name string) bool {
	return strings.HasSuffix(name, "_ungated.t")
}

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

// reOoInfixIsa matches the INFIX `isa`, and only that.
//
// A variable, the bare word, then a capitalised bareword class -- which
// is the one shape the operator has. `->isa(` is the METHOD and a
// different construct entirely: measured, the infix form emits `<2> isa`
// with no `entersub` and no `method_named`, where the method form emits
// both. A pattern matching the bare word alone would report the two as
// one, and the tier would look like it had named a construct it had not.
var reOoInfixIsa = regexp.MustCompile(`\$\w+\s+isa\s+[A-Z]\w*`)

// reOoCanChain matches the `->can(` that opens the chain.
//
// Anchored on the arrow because a bare `can` is satisfied by the word in
// a comment, and this tier's files carry long ones. What the chain IS --
// two `entersub` under one `method_named` -- is measured in
// `11_can_chain.t`; this pattern only has to identify where it appears.
var reOoCanChain = regexp.MustCompile(`->\s*can\s*\(`)

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
	if reOoInfixIsa.MatchString(source) {
		out = append(out, infixIsaForm)
	}
	if reOoCanChain.MatchString(source) {
		out = append(out, canChainForm)
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

	// The argument-extent slice's two (issue 01a0c730), keyword-less in
	// the same way: `isa` is an OPERATOR rather than a declaration, and
	// the `can` chain is named for its arrows rather than for any word.
	infixIsaForm = "infix isa"
	canChainForm = "can chain"
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
	// The argument-extent slice's two (issue 01a0c730). Each names the
	// SPELLING that identifies it in a body holding four other method
	// calls: `isa` bare is the infix operator, which `ooForms` finds as
	// an operator and not as a call, and `->can(` is the chain's first
	// arrow, which a plain `can` would not tell apart from the word in
	// a comment.
	case "isa_infix":
		return infixIsaForm
	case "can_chain":
		return canChainForm
	// `tie` and `tied` (issue 01a0c730), which are in THIS tier and not
	// an earlier one because a working `tie` requires a BLESSED object.
	// An unblessed constructor does not error -- measured, it yields an
	// empty value silently -- and the minimal blessed one emits `bless`
	// and `emptyavhv`, both this tier's. Placements in tiers 02, 07 and
	// 12 were each tried and each refused by the op-budget lint.
	case "tie_variable":
		return "tie"
	case "tied_boolean":
		return "tied"
	default:
		return ""
	}
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
// WHY THE POSITIVE FACT AND NOT A NEGATIVE ONE, which this test asked
// for and got wrong. It used to require BOTH a file declaring
// `one operator whose text is "->"` and a file declaring
// `no operator whose text is "->"`, reasoning that the negative caught a
// lexer which never emits an arrow at all.
//
// MEASURED, it does not. Delete `"->"` from `operators` in
// `internal/lexer/scan.go` -- a lexer that cannot tell an arrow from a
// minus and a `>`, which is the exact case the old comment named -- and
// run the corpus. Ten files fail, every one of them on a POSITIVE fact.
// Not one negative fails. `06_indirect_new.t`'s source contains no `->`
// bytes, so `checkTokenFact`, which counts tokens whose text equals the
// claimed string, can never count one there whatever the lexer does.
//
// So the positive is the whole guard, and it is a real one: it fails
// when the arrow is dropped, folded into the word beside it, or split.
//
// THE OTHER HALF OF THE CONTRAST IS A CLAIM ABOUT THE SOURCE, not about
// tokens, and is asserted as one below. `new Foo` and `Foo->new` compile
// to the same op stream, so the corpus needs a file spelled each way --
// and "spelled indirectly" means the source has the call without an
// arrow, which is a question to ask of the source text.
func TestTierOoArrowIsLexical(t *testing.T) {
	const arrowPresent = `one operator whose text is "->"`

	var declaredPresent, spelledIndirectly []string
	for name, f := range tierFiles(t, tierOo) {
		for _, fact := range f.TokenFacts {
			if fact == arrowPresent {
				declaredPresent = append(declaredPresent, name)
			}
		}
		// The indirect spelling, asked of the source because that is
		// where it lives: a method call written with no arrow in it.
		if ooCallsIndirectly(f.Source) {
			spelledIndirectly = append(spelledIndirectly, name)
		}
	}

	if len(declaredPresent) == 0 {
		t.Errorf("no file in %s declares %q.\n"+
			"\tThis is the tier's only guard against a lexer that drops "+
			"the arrow, folds it into the word beside it, or splits it "+
			"into a minus and a `>`. Measured: removing `->` from the "+
			"lexer's operator table fails exactly the files carrying "+
			"this fact.", tierOo, arrowPresent)
	}
	if len(spelledIndirectly) == 0 {
		t.Errorf("no file in %s spells a method call INDIRECTLY.\n"+
			"\t`new Foo` and `Foo->new` emit an identical op stream, so "+
			"the spelling is the only place the construct is visible and "+
			"a tier holding only the direct form cannot see it at all.",
			tierOo)
	}
	t.Logf("arrow declared present in %s; indirect spelling in %s",
		strings.Join(declaredPresent, ", "),
		strings.Join(spelledIndirectly, ", "))
}

// ooCallsIndirectly reports whether a source spells a method call in the
// indirect-object form -- `new Foo`, a bareword method followed by a
// bareword class, with no arrow between them.
//
// Deliberately narrow: it looks for the `new Foo` shape this tier
// actually writes rather than trying to recognise indirect object
// notation in general, which is undecidable without a parser and would
// make this test assert something about our parser rather than about
// the corpus.
func ooCallsIndirectly(src string) bool {
	return reIndirectCall.MatchString(src)
}

// reIndirectCall matches `= new Foo;` and `= new Foo(...)`: an
// assignment to a bareword method applied to a bareword class.
//
// Anchored on the `=` so a `sub new {` declaration does not match, and
// requiring an initial capital on the class so an ordinary two-word
// call like `print STDERR` does not.
var reIndirectCall = regexp.MustCompile(`=\s*[a-z]\w*\s+[A-Z]\w*\s*[;(]`)

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
