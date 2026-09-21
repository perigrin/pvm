// ABOUTME: Tier 07 checked against the finished tooling, the call-form slice excepted.
// ABOUTME: Six checks the tier's own issue names, the `@_` aliasing one among them.
package conformance

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierSubroutines is the tier this file is about.
//
// A constant rather than a literal at six call sites, for tier01_test.go's
// reason: the tier number is a POSITION and positions move.
const tierSubroutines = "07_subroutines"

// TestTierSubroutinesPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// Identical in shape to the tier 01 check and deliberately so: a per-tier
// acceptance that differed per tier would be six different claims wearing
// one name. `perlPath` resolves the one perl reporting `$] = 5.042000`,
// which is what makes every file's MEASURED perl 5.42.0 header a checkable
// statement rather than a comment.
//
// What is asserted per file: perl agrees with the `--- expect parses` /
// `--- expect parsent` bit, and perl prints what the file pins, byte for
// byte, when it pins anything.
func TestTierSubroutinesPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierSubroutines, perl)

	for name, f := range tierFiles(t, tierSubroutines) {
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

// TestTierSubroutinesLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate; this
// is the tier's own, so a file here that reaches forward fails named as
// this tier's problem.
//
// WHAT THIS PROVES IS NARROWER HERE THAN ANYWHERE EARLIER, and the tier
// README's last section is the record of why. `opsOf` runs `perl
// -MO=Concise,-exec` with no sub named, which prints the MAIN program's
// optree; a sub body is a separate CV and is not in it. So for this tier
// the lint sees the call sites and the closure construction and nothing of
// the bodies -- `leavesub`, `argcheck`, `argelem`, `argdefelem` and
// `return` are all emitted by these files and none is reachable.
//
// It is still worth running in that narrowed form. The direction it checks
// -- no file here uses a LATER tier's op -- runs over the main program and
// keeps its force: it is why `08_references`' `srefgen` stays out of these
// files. Issue 01a0c547-516b-73d3-ba9e-ae5b6c40fa29 tracks teaching the
// lint to descend; until it does, claiming the five invisible ops in
// INTRODUCES would make `TestCorpusLints` fail from the other direction,
// which is the correct answer to a claim nothing can reproduce.
func TestTierSubroutinesLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierSubroutines) {
		t.Run(name, func(t *testing.T) {
			if err := lintOps(t, f.Source, tierSubroutines, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// subroutineConstructs are the spellings that stand for this tier's
// constructs, each keyed by the construct file that introduces it.
//
// WHY A TABLE HERE WHERE TIER 01 DERIVED ONE. Tier 01's construct is a
// literal bound by `my $x = <literal>;`, so one regex over a file's source
// extracts it and the files themselves remain the enumeration. This tier
// has no such uniform shape: a named declaration, an anonymous one, a
// signature with a default, an `@_` read, an `@_` write and a guarded
// return are six different syntactic forms with no common frame to match
// on. A regex general enough to catch all six would match so loosely that
// containment in the adjacency file would stop meaning anything.
//
// What keeps the table from drifting away from the directory is the loop
// below: every construct file must have an entry, so adding a file to the
// tier without adding its spelling here fails rather than passing
// silently. That is the property tier 01 got from derivation, obtained the
// other way round.
//
// Each spelling is written with the surrounding syntax that pins it. A
// bare `return` would be satisfied by any of the adjacency file's three;
// `$_[0]++` is the aliasing write and nothing else spells it.
var subroutineConstructs = map[string]string{
	"01_named_sub.t":  "sub pick",
	"02_anon_sub.t":   "sub {",
	"03_call_forms.t": "&pick(",
	"04_args_array.t": "$_[0]",
	"05_return.t":     "return $label if",
	"06_signature.t":  "($n, $label = ",
	"07_args_alias.t": "$_[0]++",

	// The call-form slice's files (issue 01a0c432-fbd5). Registered here
	// rather than left out, because this table's own loop is what stops
	// a file joining the tier without joining the adjacency check -- the
	// property tier 01 gets by derivation, obtained the other way round.
	// A file absent from it fails loudly, which is how these three
	// arrived.
	//
	// Each is spelled as the CALL SITE rather than the declaration,
	// because the call site is what the form is: `print f 1, 2;` and
	// `print g 1, 2;` are the same shape and print different things, f
	// greedily and g cut to one argument by its prototype. The
	// declarations that decide that difference are not interchangeable
	// between the two, so a spelling taken from one would not identify
	// the other.
	//
	// `g` is declared inside `no feature "signatures"` in the adjacency
	// file, because `use v5.36` turns `($)` from a PROTOTYPE into a
	// SIGNATURE -- measured 5.42.0, `use v5.36; sub g ($) {...} print g
	// 1, 2;` is "Too many arguments for subroutine 'main::g'". That
	// block is what lets both readings live in one body.
	"08_parenless_extent.t": "print f 1, 2;",
	"09_prototype_extent.t": "print g 1, 2;",
	"11_code_ref_call.t":    "->(",
}

// TestTierSubroutinesAdjacency checks the tier's adjacency file holds
// every construct the tier introduces, each next to another, and that it
// pairs with the tier's DECLARED prerequisite.
//
// COVERAGE, NOT OPS, for tier 01's reason and more sharply. There the
// optimiser refuted an op-based check -- more adjacent constructs fused
// into `padrange` and emitted FEWER ops. Here the ops are not merely fewer
// but absent: measured 5.42.0, `00_adjacency.t`'s source carries a
// signature, a default, a `for` loop, a `next`, three `return`s, an
// aliasing write through `$_[0]` and an ampersand call, and its main
// optree contains `anoncode` and `entersub` and not one other op from
// tier 06 or 07, because every one of them compiled into a CV. Seven
// constructs, two visible ops. An op-based adjacency check would be
// measuring the wrong program.
//
// THE PAIRING IS WITH THE DECLARED PREREQUISITE, READ FROM THE README.
// `readTierDeps` is what reads it, and the reason to read rather than
// assume N-1 is that the spec names tiers where N-1 asserts nothing --
// 09_regex needs nothing from 08_references. This tier declares
// 06_control, and it declares it for a reason the adjacency file has to
// honour: without a branch, `return` compiles to no op at all, so the
// construct the tier is named for is observable only once control exists.
// The check below therefore demands the prerequisite's own construct be
// present in the adjacency body -- a `next`, which is tier 06's and is in
// no other file in this tier -- so the pairing is a fact about the source
// rather than a sentence in a comment.
func TestTierSubroutinesAdjacency(t *testing.T) {
	files := tierFiles(t, tierSubroutines)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierSubroutines, adjacencyFile)
	}

	for name := range files {
		if name == adjacencyFile {
			continue
		}
		spelling, ok := subroutineConstructs[name]
		if !ok {
			t.Errorf("%s has no entry in subroutineConstructs.\n"+
				"\tEvery construct file names the spelling that stands for "+
				"it, or the adjacency check silently stops covering it.", name)
			continue
		}
		if !strings.Contains(adj.Source, spelling) {
			t.Errorf("%s introduces %s, which %s does not contain.\n"+
				"\tThe adjacency file must hold every construct the tier "+
				"introduces, or the pairing it exists to reach is not "+
				"reachable for that construct.",
				name, spelling, adjacencyFile)
		}
	}

	// The DECLARED prerequisite, and its construct in the same body.
	deps, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier DEPENDS ON blocks: %v", err)
	}
	dep := deps[tierSubroutines]
	if dep != "06_control" {
		t.Fatalf("%s declares %s as its prerequisite; this check knows how to "+
			"pair with 06_control and nothing else", tierSubroutines, dep)
	}
	// `next` is 06_control's and appears nowhere else in this tier, which
	// is what makes its presence here evidence of the pairing rather than
	// a coincidence of vocabulary.
	if !strings.Contains(adj.Source, "next if") {
		t.Errorf("%s declares %s as its prerequisite, but %s carries no "+
			"loop-control statement from it.\n"+
			"\tThe adjacency file pairs this tier's constructs with the "+
			"DECLARED prerequisite's, not with tier N-1's by position.",
			tierSubroutines, dep, adjacencyFile)
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
}

// TestTierSubroutinesArgAliasing measures what makes `@_` different from
// an array, against perl, rather than assuming it.
//
// The issue says it plainly: "`@_` is the one worth care: it aliases, so
// `$_[0] = 1` modifies the caller's variable. A file that treats it as a
// plain array measures nothing about what makes it different."
// `04_args_array.t` READS `$_[0]`, which any array supports; the aliasing
// file writes through it, which no copy does.
//
// HOW THIS IS MEASURED, and why a pinned output alone would not be enough.
// The corpus file's `--- expect output` already pins what perl prints, and
// `TestTierSubroutinesPerlValidated` already checks that byte for byte. So
// a check here that only re-ran the file would restate an existing one. The
// claim that needs its own measurement is a DIFFERENCE: that the pinned
// output is what it is BECAUSE of aliasing. So the file's source is run
// twice -- once as written, once with the aliasing write replaced by a
// write to a copy -- and the two outputs must differ.
//
// That is the only form of this check that can fail for the right reason.
// An aliasing file whose sub happened to return the modified value would
// print the same thing under both spellings, and its pinned output would
// then be evidence of nothing; this catches that, and a pinned-output
// check cannot.
//
// The substitution is a plain string replacement on the source rather than
// anything parser-shaped, for the reason the tier 01 glossary check uses
// source text: asking our own parser to adjudicate a file is asking the
// thing under test to grade itself.
func TestTierSubroutinesArgAliasing(t *testing.T) {
	const aliasFile = "07_args_alias.t"

	files := tierFiles(t, tierSubroutines)
	f, ok := files[aliasFile]
	if !ok {
		t.Fatalf("%s has no %s.\n"+
			"\tThe tier's `@_` file reads arguments; nothing in the tier "+
			"writes through one, which is the property that distinguishes "+
			"`@_` from an array.", tierSubroutines, aliasFile)
	}

	// The aliasing write, and the same write aimed at a copy instead.
	// `my $copy = $_[0]; $copy++` modifies a pad variable the caller
	// cannot see, which is exactly what `@_` is not.
	const aliasingWrite = "$_[0]++"
	const copyingWrite = "my $copy = $_[0]; $copy++"

	if !strings.Contains(f.Source, aliasingWrite) {
		t.Fatalf("%s does not write through `%s`.\n"+
			"\tReading `$_[0]` measures an array; writing through it is "+
			"what measures the alias.", aliasFile, aliasingWrite)
	}

	_, aliased := askPerl(t, f.Source)
	_, copied := askPerl(t, strings.Replace(f.Source, aliasingWrite, copyingWrite, 1))

	if aliased == copied {
		t.Errorf("replacing `%s` with `%s` leaves perl printing %q either way.\n"+
			"\tThe file's output does not depend on the aliasing, so it "+
			"measures nothing about it.", aliasingWrite, copyingWrite, aliased)
	}
	if f.ExpectOutput != nil && *f.ExpectOutput != aliased {
		t.Errorf("pinned output %q, perl prints %q under the aliasing spelling",
			*f.ExpectOutput, aliased)
	}
}

// TestTierSubroutinesRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// The rule is BICONDITIONAL, and tier 01 is where that was settled.
// `parse.RefusalCode` names a site in the PARSER, so a file whose refusal
// is purely lexical has no Unknown to name and naming one anyway would
// make the file fail as a stale marker -- correctly, since the claim would
// be false. So:
//
//   - A file whose parser produces Unknowns MUST name one of their codes.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// THE TIER AS IT STANDS REFUSES NOWHERE, which is not a reason to omit the
// check. Every file here parses today, so this loop runs zero subtests;
// the moment one of them starts refusing -- and the tier's own README
// expects that, since signatures and ampersand calls are not all
// implemented -- the citation rule is already in force rather than being
// remembered by whoever adds the marker.
func TestTierSubroutinesRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierSubroutines) {
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
