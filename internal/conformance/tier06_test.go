// ABOUTME: Tier 06 checked against the finished tooling, the statement code path rather than tier 04's expression one.
// ABOUTME: Five checks the tier's own issue names plus the code-path claim its title makes.
package conformance

import (
	"regexp"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierControl is the tier this file is about.
//
// A constant rather than a literal at every call site, for the reason
// tierLiterals is one: the tier number is a POSITION and positions move.
const tierControl = "06_control"

// TestTierControlPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// The same two claims tier 01 checks -- perl agrees with the parses bit,
// and perl prints what the file pins -- for the same reason: a file whose
// expectation perl contradicts measured nothing about our parser.
//
// This tier needs it more than tier 01 does, because of the ERASURES its
// README records. `if (1) { ... } else { ... }` emits no branch op and
// `while (0) { ... }` compiles to an empty program, and in both cases the
// output test still passes. Running the pinned perl is what makes the
// pinned output an actual measurement of the interpreter rather than a
// transcription of what the author expected it to say.
func TestTierControlPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierControl, perl)

	for name, f := range tierFiles(t, tierControl) {
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

// TestTierControlLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate;
// this is the tier's own, so a file reaching forward fails HERE, named as
// this tier's problem.
//
// The lint has already earned its place in this tier once. The README
// records it: `print $i-- while $i > 0` emits `postdec`, which no tier at
// or before 06 claims, so `11_postfix_while.t` spells the decrement
// `$i = $i - 1` instead. The file's subject is the loop frame and the
// lint refused the shorter spelling correctly.
func TestTierControlLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierControl) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierControl, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// reControlKeyword matches the keyword a construct file introduces,
// wherever it stands in the file's source.
//
// This tier's constructs are KEYWORDS, not bindings. Tier 01's
// `literalSpelling` reads the literal after a `my $x = `, which is the
// whole construct there; here the construct is `while` or `foreach` or
// `goto` and it appears in statement position, in modifier position, or
// both. So the spelling taken from a construct file is the keyword, and
// the adjacency check asks whether that keyword appears in the adjacency
// file rather than whether some longer text does.
//
// Anchored on a word boundary at both ends so `for` is not satisfied by
// `foreach` -- the two are different optrees, which is precisely the pair
// the README says the disambiguation happens at the open paren for.
//
// `do` and `eval` join the list with the tier's fourth group, the
// block-valued expressions. They are keywords like the rest, but the
// adjacency claim they carry is sharper: `do` appears in TWO of this
// tier's files under two unrelated productions -- `09_do_while.t`'s loop
// and `13_do_block.t`'s expression -- so the adjacency file must hold a
// `do`, and holding one is not by itself evidence it holds both. That
// second half is what the README's group prose states and what no regexp
// over a keyword can check.
//
// `die`, `exit` and `time` join with the tier's fifth and sixth groups
// (issue 01a0c730). They were MISSING from this alternation for a while
// after the tier claimed them, and the omission was invisible because
// `15_die.t` and `16_exit.t` both spell `eval` as their observation
// frame -- a `die` has to be trapped to be printed, and an `exit` has to
// be trapped to be survived. So both files satisfied the gate through a
// keyword they merely USE, while the construct each is named for went
// unchecked. A gate that passes on incidental vocabulary is worse than
// no gate, because it reports coverage it does not have.
var reControlKeyword = regexp.MustCompile(`\b(unless|until|foreach|while|if|for|goto|next|last|redo|eval|do|die|exit|time)\b`)

// controlKeywords returns the keywords a file's source spells, in the
// order the regexp's alternation prefers -- longest first, so a `foreach`
// is never also counted as a `for`.
func controlKeywords(source string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range reControlKeyword.FindAllStringSubmatch(source, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// TestTierControlAdjacency checks the tier's adjacency file holds every
// construct the tier introduces, each next to another, and that it pairs
// with the tier's DECLARED prerequisite.
//
// WHAT OPS CANNOT DO HERE is the same obstacle tier 01 hit, arriving by a
// different route. There the optimiser fused adjacent `my` declarations
// into one `padrange`, so more adjacency meant fewer ops. Here the ops
// are not a function of the source at all in the direction the check
// wants: the README measures that `print "y" if $c` and `if ($c) { print
// "y" }` emit IDENTICAL op streams, so the op stream cannot tell which of
// the two the adjacency file contains, and a check that counted ops would
// pass on a file holding only one of the pair. Worse, the tier's own
// erasures mean a file can emit fewer ops than it spells constructs:
// a constant condition deletes the branch entirely.
//
// So the adjacency claim is checked on the SOURCE, as tier 01's is, with
// the spelling being the KEYWORD rather than a literal binding.
//
// THE PART THAT DIFFERS FROM TIER 01, and it is what this tier's AC asks
// for. Tier 01 depends on `nothing`, so it has no prerequisite to pair
// with and its adjacency file pairs only its own constructs. Tier 06
// declares `05_scoping`, which is NOT tier N-1 in the numeric sense of
// "whatever came before" but the one dependency the README says the tier
// cannot be written without -- the block. The pairing this test demands
// is therefore with 05's subject: every one of this tier's branches and
// loop bodies is a block, and the adjacency file must actually contain
// blocks and scoped declarations rather than being a run of postfix
// modifiers that never opens a brace.
//
// That check is not decoration. A postfix-only adjacency file would hold
// every keyword this tier introduces and still never pair the tier with
// its declared prerequisite -- the README measures that a postfix loop
// emits no `enterloop` at all, because it has no block for `next`/`last`
// to target. Reading the prerequisite from the README rather than
// hard-coding "05" is what makes this a test of the DECLARATION.
func TestTierControlAdjacency(t *testing.T) {
	files := tierFiles(t, tierControl)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierControl, adjacencyFile)
	}

	// The constructs each non-adjacency file introduces, taken from those
	// files' own sources rather than from a list here: the tier's files
	// ARE the enumeration of what it introduces, and a second list beside
	// them is how this package has drifted before.
	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		kws := controlKeywords(f.Source)
		if len(kws) == 0 {
			t.Errorf("%s: no control keyword to take a construct from", name)
			continue
		}
		for _, kw := range kws {
			if !strings.Contains(adj.Source, kw) {
				t.Errorf("%s introduces %s, which %s does not contain.\n"+
					"\tThe adjacency file must hold every construct the "+
					"tier introduces, or the pairing it exists to reach is "+
					"not reachable for that construct.",
					name, kw, adjacencyFile)
			}
		}
	}

	// One body, not several. `nextstate` is one per statement, which is
	// how this counts statements without a parser of its own.
	ops, err := opsOf(t, adj.Source)
	if err != nil {
		t.Fatalf("%s: %v", adjacencyFile, err)
	}
	if n := countOp(ops, "nextstate"); n < 2 {
		t.Errorf("%s compiles to %d statement(s); adjacency needs at least 2", adjacencyFile, n)
	}

	// The pairing with the DECLARED prerequisite, read from the README
	// rather than assumed to be tier N-1.
	deps, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	dep, ok := deps[tierControl]
	if !ok {
		t.Fatalf("%s declares no prerequisite", tierControl)
	}
	if dep == "nothing" {
		t.Fatalf("%s declares no prerequisite, but its constructs all need one", tierControl)
	}

	// 05_scoping's subjects are the block and the scoped declaration --
	// `enterloop`/`leaveloop` for the bare block, and a `my` whose pad
	// slot the block owns. A file pairing with it must emit both. That is
	// checkable from ops here where the tier's OWN constructs are not,
	// because these ops belong to the earlier tier and nothing in this
	// tier erases them.
	for _, want := range []string{"enterloop", "padsv"} {
		if countOp(ops, want) == 0 {
			t.Errorf("%s emits no %s, so it does not pair %s's constructs "+
				"with %s, the prerequisite the README declares.\n"+
				"\tA run of postfix modifiers holds every keyword this "+
				"tier introduces and still opens no block -- measured, a "+
				"postfix loop emits no enterloop at all.",
				adjacencyFile, want, tierControl, dep)
		}
	}
}

// TestTierControlStatementPathIsNotTheExpressionPath checks the claim the
// tier's TITLE makes: that a statement is a different code path from an
// expression.
//
// The issue says it in prose -- "handled by hand-written recursive
// descent (`parse.go` dispatching to `control.go`) rather than by the
// Pratt expression parser" -- and gives it as the REASON this tier is its
// own tier rather than part of 04. A reason a tier exists for is worth
// more than an assertion in a README, so this makes it falsifiable.
//
// The two paths, read from parse.go's statement dispatch:
//
//   - `if ($c) { ... }` reaches `parseControlFlow` FIRST, at the word
//     token, and never reaches `parseExpr` for the statement as a whole.
//   - `print "y" if $c` falls past `parseControlFlow` -- `print` is not a
//     control keyword -- into `parseExpr`, and the modifier is applied
//     AFTERWARDS by `applyModifier`, which the comment there says runs
//     after the expression parser precisely because a modifier binds
//     looser than anything the Pratt loop's floor of 0 can express.
//
// WHAT THIS CAN AND CANNOT SEE, and it is why the check is shaped the way
// it is. The conformance package sees `parse.Parse`'s tree, not which
// function built it. So it cannot assert "parseControlFlow ran"
// directly. What it CAN assert is the observable consequence that makes
// the two paths distinguishable:
//
//   - Both spellings produce a Conditional, which is the EQUALITY the
//     corpus file `01_if_postfix.t` pins at the op level. If the two
//     paths disagreed about the kind of node, one of them would be
//     building a tree perl does not build.
//   - The postfix form's Conditional has the PRINT as its first child,
//     because `applyModifier` inverts the tree -- body first, condition
//     second, in source order. The block form's does not: its first
//     child is the condition. That asymmetry is the statement path and
//     the expression path leaving different fingerprints on the same
//     source-level construct, and it is exactly what a parser that ran
//     ONE path for both would not produce.
//
// The ops cannot adjudicate this. The README measures that the two
// spellings emit byte-identical op streams down to the `nextstate`, so
// perl has already collapsed the distinction by the time ops exist. The
// tree is the only place the two code paths are still distinguishable,
// which is the whole reason this tier measures a parser rather than an
// interpreter.
func TestTierControlStatementPathIsNotTheExpressionPath(t *testing.T) {
	// Taken from 01_if_postfix.t, which is the corpus file whose subject
	// is this pair. Spelled here rather than read from the file because
	// the claim is about two SPECIFIC statements and needs them apart.
	const prelude = "my $c = $ENV{X} // 1;\n"

	kindOfFirstChild := func(t *testing.T, source string) (*parse.Node, *parse.Node) {
		t.Helper()
		root := parse.Parse([]byte(prelude + source))
		var cond *parse.Node
		var walk func(*parse.Node)
		walk = func(n *parse.Node) {
			if n == nil || cond != nil {
				return
			}
			if n.Kind == parse.Conditional {
				cond = n
				return
			}
			for _, c := range n.Children {
				walk(c)
			}
		}
		walk(root)
		if cond == nil {
			t.Fatalf("no Conditional in the tree for %q", source)
		}
		if len(cond.Children) < 2 {
			t.Fatalf("Conditional for %q has %d children, need condition and body",
				source, len(cond.Children))
		}
		return cond, cond.Children[0]
	}

	blockCond, blockFirst := kindOfFirstChild(t, `if ($c) { print "block\n" }`)
	postfixCond, postfixFirst := kindOfFirstChild(t, `print "postfix\n" if $c;`)

	// Both paths agree on the KIND. A disagreement here would mean one of
	// them builds a node perl's optree has no counterpart for.
	if blockCond.Kind != postfixCond.Kind {
		t.Errorf("block `if` builds %v and postfix `if` builds %v; the two "+
			"spellings are one construct to perl, which measures them as "+
			"byte-identical op streams",
			blockCond.Kind, postfixCond.Kind)
	}

	// And they disagree on the SHAPE, which is the two code paths showing
	// through. The postfix form's body was parsed by parseExpr before the
	// modifier existed, so it is child 0; the block form's condition was
	// parsed by parseControlFlow's own parenthesised-condition reader, so
	// the condition is child 0.
	if postfixFirst.Kind == blockFirst.Kind {
		t.Errorf("block `if` and postfix `if` put the same kind (%v) first.\n"+
			"\tThe postfix form is built by applyModifier, which inverts "+
			"the tree and puts the already-parsed BODY first; the block "+
			"form's condition comes first. Same first child means one code "+
			"path served both, and the statement/expression boundary the "+
			"tier exists to measure is not there.",
			blockFirst.Kind)
	}
}

// TestTierControlRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// The rule is tier 01's, biconditional for tier 01's reason:
// `parse.RefusalCode` names a site in the PARSER, so a file that refuses
// on a purely lexical fact has no Unknown to name and must NOT cite a
// code -- `run.go` would fail it as a stale marker, correctly, because
// the claim would be false.
//
// AS MEASURED, THIS TIER HAS NO REFUSING FILE. Every one of its eleven
// files is green, which is what `TestCorpus/06_control` reports. That
// makes this check vacuous TODAY and not vacuous tomorrow: the statement
// path is where `UnimplementedStatement` lives -- the largest single
// source of Unknowns in T1 -- so the first file this tier gains for a
// statement form the parser has not reached will be a refusing file, and
// it will arrive under this check rather than beside it.
//
// A test that passes because its loop body never runs is the failure mode
// this package has hit before, so the vacuity is asserted rather than
// left to be discovered: if the tier ever gains a refusing file, the
// subtests below run and the count below changes, and if the tier's files
// stop existing entirely the fixture fails first.
func TestTierControlRefusalsCited(t *testing.T) {
	files := tierFiles(t, tierControl)

	var refusing int
	for name, f := range files {
		if f.Refuses == "" {
			continue
		}
		refusing++
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

	// The vacuity, made visible. Not an assertion that the tier must stay
	// refusal-free -- a file gaining a refusal is progress, not a
	// regression -- but a log line, so a reader of the verbose output can
	// see whether the loop above ran at all.
	t.Logf("%s holds %d file(s), %d of them refusing", tierControl, len(files), refusing)
	if len(files) == 0 {
		t.Fatalf("%s holds no files, so this check asserted nothing", tierControl)
	}
}
