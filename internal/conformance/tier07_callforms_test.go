// ABOUTME: The tier 07 call-form slice, and the smallest-useful-corpus gate it completes.
// ABOUTME: Parenless argument extent is MEASURED against perl here, not assumed.
package conformance

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// callFormFiles are the corpus files this slice owns, inside
// 07_subroutines.
//
// NAMED RATHER THAN GLOBBED, and that is the whole reason this list
// exists. The tier directory holds two issues' files: declaration, `@_`,
// `return` and signatures belong to 01a0c362-073a and are checked by
// tier07_test.go, and these five are the call-form slice. A glob over
// the directory would make each issue's checks silently answer for the
// other's files -- so when one of them adds a file, the other's
// acceptance would change meaning without anybody editing it.
//
// `03_call_forms.t` predates this issue and is listed anyway: it is the
// four-name-forms file, it IS a call form, and the slice's gate would be
// lying if it excluded the one call-form file that already existed.
// THE UNDECLARED CALLEE IS NOT A FILE HERE, and the reason is a
// corpus-wide gap rather than a judgement about the construct. perl
// REFUSES that source -- that refusal is the whole measurement -- so
// such a file would carry `--- expect parsent`. The format documents
// that section and `ParseFile` reads it, but measured at dc1bea2c no
// corpus file had ever used one, and the lint infrastructure assumes
// every file compiles: `TestCorpusLints` and `TestTierSubroutinesLint`
// both call `opsOf` unconditionally, and `perl -MO=Concise` exits 255
// on a program with a syntax error. A `parsent` file anywhere in the
// corpus fails them, in a package this issue may not edit.
//
// So the fact is measured in TestTierCallFormsParenlessExtent instead,
// where perl is run directly and no optree is asked for. That is where
// it belongs regardless -- it is the fact that the extent question has
// a PRECONDITION, not a construct with an extent of its own -- and the
// corpus file would have been a second copy of it.
var callFormFiles = []string{
	"03_call_forms.t",
	"08_parenless_extent.t",
	"09_prototype_extent.t",
	"11_code_ref_call.t",
}

// callFormSlice returns the slice's parsed files, keyed by base name.
//
// Fails rather than skipping when one is missing. A slice whose gate
// quietly ran over four files instead of five would still say PASS, and
// the one thing this issue delivers is a gate that means something.
func callFormSlice(t *testing.T) map[string]*File {
	t.Helper()

	all := tierFiles(t, tierSubroutines)
	out := map[string]*File{}
	for _, name := range callFormFiles {
		f, ok := all[name]
		if !ok {
			t.Fatalf("%s/%s is named in callFormFiles but is not in the tier.\n"+
				"\tThe list is what keeps this slice's checks from "+
				"answering for the rest of tier 07's files; a name in it "+
				"with no file behind it measures nothing.",
				tierSubroutines, name)
		}
		out[name] = f
	}
	return out
}

// parenlessProbe is one source spelled once and compiled twice, under a
// callee with a prototype and a callee without.
//
// ONE PROGRAM TEXT, TWO DECLARATIONS. The call site below -- `print f 1,
// 2;` -- is byte-identical in both, and only the `sub f` line above it
// differs. That is what makes this a measurement of EXTENT rather than
// of two unrelated programs: if the two runs print the same thing, the
// probe has discriminated nothing and says so.
//
// `%s` is where the prototype goes, empty or `($)`.
const parenlessProbe = `sub f %s{ return "f[" . join(",", @_) . "]" }
print f 1, 2;
print "\n";
`

// TestTierCallFormsParenlessExtent measures how far a parenless call's
// arguments run, against perl, rather than assuming it.
//
// THIS IS THE ISSUE'S POINT and it is worth stating what "assumed" would
// have looked like. The obvious assumption is that `f 1, 2` is ambiguous
// at parse time and that a parser must guess. Perl does not guess, and
// measuring says so in four facts, none of which a reading of the source
// would have given:
//
//  1. AN UNDECLARED CALLEE IS NOT AMBIGUOUS, IT IS A SYNTAX ERROR.
//     Measured 5.42.0, `f 1, 2` where perl has not yet seen `f` is
//     "Number found where operator expected (Do you need to predeclare
//     \"f\"?)" followed by "syntax error" -- and this is true with
//     `use strict` AND without it, and true whether or not `sub f` is
//     defined LATER in the same file. There is no extent question to
//     answer, because there is no call. Measured below rather than in a
//     corpus file, for the reason callFormFiles records: perl refuses
//     the source, and the corpus lint asks every file for an optree.
//
//  2. A DECLARED CALLEE WITH NO PROTOTYPE IS GREEDY. `print f 1, 2`
//     prints `f[1,2]`: the `2` is f's second argument, NOT print's
//     second. The parenless call consumes the whole remaining list even
//     though it is itself nested inside another list operator. Both
//     `sub f;` and a full definition are enough to make this true; the
//     body need not have been seen.
//
//  3. A PROTOTYPE CHANGES THE EXTENT OF THE SAME BYTES. With `sub f
//     ($)`, that identical `print f 1, 2` prints `f[1]2` -- one argument
//     to f, and the `2` falls through to print. The source text is the
//     same; the parse is not. This is the fact that makes the extent a
//     MEASUREMENT: nothing local to the call site decides it.
//
//  4. `+ 3` JOINS THE LAST ARGUMENT RATHER THAN THE LIST. `f 1, 2 + 3`
//     calls f with `(1, 5)`, because `+` binds tighter than `,`. That
//     one is what a precedence table would have predicted, which is why
//     it is the least interesting of the four and is still measured.
//
// HOW THE DISCRIMINATION IS CHECKED, and why a pin alone would not do.
// Both extent files already pin their output, and
// TestTierCallFormsPerlValidated already checks those byte for byte. A
// check here that only re-ran them would restate that. The claim needing
// its own measurement is a DIFFERENCE: that the outputs differ BECAUSE
// of the prototype. So the same call site is run under both declarations
// and the two results must not agree -- the failure mode six tiers in
// this corpus have already shipped is a fixture whose two spellings
// print the same thing, which a pin passes forever.
func TestTierCallFormsParenlessExtent(t *testing.T) {
	greedy := fmt.Sprintf(parenlessProbe, "")
	limited := fmt.Sprintf(parenlessProbe, "($) ")

	compiles, greedyOut := askPerl(t, greedy)
	if !compiles {
		t.Fatalf("perl refuses the unprototyped probe:\n%s", greedy)
	}
	compiles, limitedOut := askPerl(t, limited)
	if !compiles {
		t.Fatalf("perl refuses the prototyped probe:\n%s", limited)
	}

	// Fact 2 and fact 3, pinned. Measured perl 5.42.0.
	const wantGreedy = "f[1,2]\n"
	const wantLimited = "f[1]2\n"

	if greedyOut != wantGreedy {
		t.Errorf("unprototyped `print f 1, 2` prints %q, pinned %q.\n"+
			"\tThe parenless call to a declared callee with no prototype "+
			"is GREEDY: it takes the whole remaining list, including the "+
			"argument that reads as the enclosing print's.",
			greedyOut, wantGreedy)
	}
	if limitedOut != wantLimited {
		t.Errorf("prototyped `print f 1, 2` prints %q, pinned %q.\n"+
			"\tA `($)` prototype cuts the extent to one argument and the "+
			"`2` falls through to print.", limitedOut, wantLimited)
	}

	// And they must DIFFER, or the probe measures nothing. This is the
	// check a pin cannot make: two pins that happened to be equal would
	// both pass while discriminating no parse at all.
	if greedyOut == limitedOut {
		t.Errorf("the prototyped and unprototyped spellings of the same "+
			"call site both print %q.\n"+
			"\tThe probe then measures no extent: a parser that consumed "+
			"the same arguments in both cases would satisfy it.", greedyOut)
	}

	// Fact 1: the undeclared callee, which is where "measure rather than
	// assume" pays. A parser told that `f 1, 2` is an ambiguity to be
	// resolved would be modelling a program perl does not compile.
	const undeclared = "f 1, 2;\nsub f { 1 }\n"
	if compiles, _ := askPerl(t, undeclared); compiles {
		t.Errorf("perl compiles %q.\n"+
			"\tMeasured 5.42.0 it does not: a parenless call to a callee "+
			"perl has not yet SEEN is a syntax error, not an ambiguous "+
			"extent -- and defining the sub later in the file does not "+
			"help, because the parse happens first.", undeclared)
	}

	// Fact 4: `+ 3` joins the final argument rather than ending the list.
	const plusProbe = `sub f { print "f[" . join(",", @_) . "]\n" }
f 1, 2 + 3;
`
	compiles, plusOut := askPerl(t, plusProbe)
	if !compiles {
		t.Fatalf("perl refuses %q", plusProbe)
	}
	if plusOut != "f[1,5]\n" {
		t.Errorf("`f 1, 2 + 3` prints %q, pinned \"f[1,5]\\n\".\n"+
			"\t`+` binds tighter than `,`, so the addition joins the "+
			"second argument rather than closing the list.", plusOut)
	}

	t.Logf("parenless extent measured against perl: undeclared callee is a "+
		"syntax error; unprototyped is greedy (%q); `($)` cuts to one (%q); "+
		"`+ 3` joins the last argument (%q)",
		strings.TrimSuffix(greedyOut, "\n"),
		strings.TrimSuffix(limitedOut, "\n"),
		strings.TrimSuffix(plusOut, "\n"))
}

// TestTierCallFormsPerlValidated runs every file this slice owns through
// the pinned interpreter before it counts.
//
// The same shape as the tier's own check and deliberately so -- see
// TestTierSubroutinesPerlValidated for why a per-tier acceptance that
// differed per tier would be several claims wearing one name. What is
// narrower here is only WHICH files: this one answers for the slice, so
// a failure names the call-form issue rather than tier 07 at large.
func TestTierCallFormsPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating the call-form slice against %s", perl)

	files := callFormSlice(t)

	// The adjacency file too, because this slice ASSERTS AGAINST IT.
	// TestTierCallFormsAdjacency checks that the call forms sit in that
	// body, and the claim those checks rest on is that the body means
	// what it says it means. Found by mutation: deleting the
	// `no feature "signatures"` toggle turns `($)` from a prototype into
	// a signature, `print g 1, 2` stops printing `g[1]2`, and the
	// adjacency check went on passing -- it only looks for the spelling,
	// which is still there. The pinned output is what notices, and it
	// was not being read here.
	//
	// The tier's own validator covers this file as well. Duplication is
	// the right answer rather than a reason to leave it out: a file two
	// issues assert against should fail for BOTH of them, or the one
	// whose check is missing learns nothing when it breaks.
	files[adjacencyFile] = tierFiles(t, tierSubroutines)[adjacencyFile]

	for name, f := range files {
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

// TestTierCallFormsLint runs the dependency lint over the slice's files.
//
// Narrow here for the reason tier07_test.go's own lint check records at
// length: `opsOf` runs `perl -MO=Concise,-exec` with no sub named, so it
// prints the MAIN program's optree and a sub body is a separate CV that
// is not in it. The call SITES are in the main program, which is exactly
// what this slice is about, so the direction that matters -- no file here
// reaches a later tier's op -- keeps its force.
//
// It has a bite the rest of the tier's does not. `$obj->m(1)` compiles to
// `method_named`, which 11_oo introduces, so the method-call form the
// issue lists among the call forms CANNOT be a file in this tier: the
// lint would correctly refuse it. That is why the slice stops at
// `$code->()`, whose `entersub` this tier already claims, and the lint
// running here is what keeps that boundary honest rather than remembered.
func TestTierCallFormsLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range callFormSlice(t) {
		t.Run(name, func(t *testing.T) {
			// A `parsent` file HAS no optree. perl refuses to compile it,
			// which is the file's whole claim, so `perl -MO=Concise`
			// exits non-zero and `opsOf` reports that as an error. That
			// is not a lint finding about the file, it is the lint being
			// asked a question with no answer -- the set of ops a
			// program that does not compile uses is empty, and the
			// direction this lint checks (no op from a later tier) is
			// satisfied by it vacuously and correctly.
			if f.ExpectParsent {
				t.Skipf("`expect parsent`: perl compiles no optree for it, so there are no ops to lint")
			}
			if err := lintOps(t, f.Source, tierSubroutines, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// callFormSpellings are the spellings that stand for each call form, keyed
// by the file that introduces it.
//
// A table rather than a derivation, for the reason subroutineConstructs
// gives: the forms have no common syntactic frame, and a regex loose
// enough to catch `&f`, `f 1, 2` and `$c->()` alike would match so much
// that containment in the adjacency file would stop meaning anything.
//
// A file with no entry here must have one in `adjacencyExempt`, or the
// loop below fails it. An exemption is a measured reason the construct
// CANNOT sit in the shared body; silence is how a corpus stops covering
// a construct without anybody deciding to.
var callFormSpellings = map[string]string{
	"03_call_forms.t":       "&answer,",
	"08_parenless_extent.t": "print f 1, 2;",
	"09_prototype_extent.t": "print g 1, 2;",
	"11_code_ref_call.t":    "->(",
}

// adjacencyExempt names the slice files whose construct cannot appear in
// the adjacency file, and why.
//
// An EXPLICIT exemption rather than silence. A form left out of the
// adjacency check because nobody noticed is indistinguishable from one
// left out because it cannot be there, and the first is how a corpus
// stops covering a construct.
// EMPTY TODAY, AND THAT IS THE FINDING. The prototype form looked like
// it had to be exempt: `sub g ($)` is a PROTOTYPE only where the
// signatures feature is off, the adjacency file says `use v5.36`, and
// under it perl reads the same three characters as a SIGNATURE and
// enforces arity --
//
//	$ perl -e 'use v5.36; sub g ($) { "g" } print g 1, 2;'
//	Too many arguments for subroutine 'main::g' (got 2; expected 1)
//
// -- so the extent-cutting behaviour is simply absent there. Measured
// 5.42.0, `no feature "signatures"` in a block restores it inside a file
// that otherwise enables them, so the form joins the shared body after
// all and the pair that matters -- `print f 1, 2` printing `f[1-2]`
// beside `print g 1, 2` printing `g[1]2` -- is reachable.
//
// The map stays because the check needs somewhere to put a form that
// genuinely cannot share the body, and an exemption must be a written
// measurement rather than an omission nobody noticed.
var adjacencyExempt = map[string]string{}

// TestTierCallFormsAdjacency checks the tier's adjacency file places the
// call forms side by side in one body.
//
// WHAT THIS ADDS OVER TestTierSubroutinesAdjacency, which already checks
// that file. That one asks whether each of the tier's seven constructs is
// present; this one asks the question the spec's adjacency rule is
// actually for, about THESE forms: a parser that handles `f(1)`, `f 1`,
// `&f` and `$c->()` each alone can still get a pair wrong, and the pair
// is only reachable if they share a body.
//
// THE PREREQUISITE IS READ AND THEN PINNED, both. Three tiers found that
// "the adjacency file emits an op the declared prerequisite introduces"
// passes under the WRONG declaration, because an adjacency file touches
// several tiers at once. So `readTierDeps` supplies the declaration and
// the pin below says which declaration this check knows how to pair with;
// a declaration that moved while this test went on passing is the failure
// that pin exists to catch.
func TestTierCallFormsAdjacency(t *testing.T) {
	adj, ok := tierFiles(t, tierSubroutines)[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierSubroutines, adjacencyFile)
	}
	// Through callFormSlice rather than the raw tier map, so a name in
	// callFormFiles with no file behind it fails where it is named
	// instead of nil-dereferencing at the containment check below.
	files := callFormSlice(t)

	for _, name := range callFormFiles {
		if why, exempt := adjacencyExempt[name]; exempt {
			t.Logf("%s is exempt from adjacency: %s", name, why)
			continue
		}
		spelling, known := callFormSpellings[name]
		if !known {
			t.Errorf("%s has no entry in callFormSpellings and no exemption.\n"+
				"\tEvery call form names the spelling that stands for it, "+
				"or the adjacency check silently stops covering it.", name)
			continue
		}
		if !strings.Contains(files[name].Source, spelling) {
			t.Errorf("%s is said to be spelled %q, which its own source does "+
				"not contain.\n\tThe spelling is the file's identity here; "+
				"one the source does not honour makes the containment check "+
				"below ask about the wrong construct.", name, spelling)
			continue
		}
		if !strings.Contains(adj.Source, spelling) {
			t.Errorf("%s introduces the call form spelled %q, which %s does "+
				"not contain.\n\tThe spec's rule is that the tier's "+
				"constructs sit side by side in one body; a form absent "+
				"from it is a form no pair covers.", name, spelling, adjacencyFile)
		}
	}

	// At least two of the forms, in one body, or "side by side" is a
	// description of nothing. `nextstate` is one op per statement, which
	// is how this counts statements without a parser of its own.
	ops, err := opsOf(t, adj.Source)
	if err != nil {
		t.Fatalf("%s: %v", adjacencyFile, err)
	}
	if n := countOp(ops, "nextstate"); n < 2 {
		t.Errorf("%s compiles to %d statement(s); adjacency needs at least 2", adjacencyFile, n)
	}

	// The DECLARED prerequisite, read, then pinned.
	deps, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier DEPENDS ON blocks: %v", err)
	}
	dep, ok := deps[tierSubroutines]
	if !ok {
		t.Fatalf("%s declares no DEPENDS ON", tierSubroutines)
	}
	const pinnedPrerequisite = "06_control"
	if dep != pinnedPrerequisite {
		t.Fatalf("%s declares DEPENDS ON %q, but this check pairs the call "+
			"forms with %q.\n\tA declaration that moved while this test went "+
			"on passing would pair the forms with the WRONG prerequisite and "+
			"nothing would say so.", tierSubroutines, dep, pinnedPrerequisite)
	}
	// `next` is 06_control's and appears nowhere else in this tier, which
	// is what makes its presence evidence of the pairing rather than a
	// coincidence of vocabulary.
	if !strings.Contains(adj.Source, "next if") {
		t.Errorf("%s declares %s as its prerequisite, but %s carries no "+
			"loop-control statement from it.", tierSubroutines, dep, adjacencyFile)
	}
}

// TestTierCallFormsRefusalsCited checks that each refusing file in the
// slice names WHICH site declines, from the stable inventory.
//
// THE RULE IS BICONDITIONAL and tier 01 is where that was settled.
// `parse.RefusalCode` names a site in the PARSER, so:
//
//   - A file whose parser produces Unknowns MUST name one of their codes.
//   - A file whose parser produces NONE must NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// Naming a code for a purely lexical refusal would be a false claim that
// still skipped, which is how a corpus stops measuring what it documents.
func TestTierCallFormsRefusalsCited(t *testing.T) {
	for name, f := range callFormSlice(t) {
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
					"\tThe inventory is the vocabulary; a code outside it is "+
					"a message wearing a code's spelling.", f.RefusalCode)
			}
			if !hasCode(codes, f.RefusalCode) {
				t.Errorf("names Refusal %s, but the parser refuses with %s",
					f.RefusalCode, joinCodes(codes))
			}
		})
	}
}

// smallestUsefulCorpus is the milestone's stated delivery target, spelled
// as the spec spells it: "tiers 01-04 plus a call-form slice from tier 07".
var smallestUsefulCorpus = []string{
	"01_literals",
	"02_variables",
	"03_context",
	"04_operators",
}

// TestSmallestUsefulCorpus is the gate: tiers 01-04 plus the call-form
// slice are all green.
//
// WHAT "GREEN" HAS TO MEAN HERE, because the obvious two readings are
// both worthless and the gate is the issue's one delivery target.
//
// A gate demanding ZERO REFUSALS fails today and would fail forever. Tier
// 01 documents five refusals and tier 04 five; they are the corpus doing
// its job, naming constructs the parser does not handle, and a milestone
// gate that cannot be met until the parser is finished is a gate nobody
// can act on. It would also invert the corpus's design, which SKIPS a
// documented refusal precisely so the suite stays pristine while still
// recording the gap.
//
// A gate IGNORING refusals asserts nothing. Every file could refuse and
// it would still say PASS, which is the failure eight of the fourteen
// tiers in this milestone shipped in one form or another.
//
// So the gate demands what `verdict` in run.go actually distinguishes.
// Its five outcomes divide cleanly into three that are the corpus working
// and two that are it broken:
//
//   - `passed` -- our parser handled the file. Green.
//   - `knownRefusal` -- our parser did not, and the FILE SAYS SO, citing
//     an issue and, where the refusal is in the parser, the site. Green:
//     this is a measurement that succeeded, recording a gap accurately.
//   - `corpusBug` -- perl disagrees with the file's own expectation. NOT
//     green: nothing was measured about our parser at all.
//   - `staleMarker` -- the file claims a refusal it no longer has, or
//     names a site that no longer declines. NOT green: the file is still
//     being skipped on a claim its header no longer describes, which is
//     exactly how a corpus silently stops measuring.
//   - `refused` -- our parser failed and the file does not admit it. NOT
//     green: an undocumented regression.
//
// The gate is therefore: every file in tiers 01-04 and in the slice
// reaches `passed` or `knownRefusal`, and the count of each is REPORTED
// so a refusal budget that grows is visible in the log rather than
// silently absorbed. That is the strongest statement that is true today
// and that stays meaningful as the parser improves: a gap may exist, but
// it must be documented, cited, and still accurate.
func TestSmallestUsefulCorpus(t *testing.T) {
	type outcome struct {
		file string
		v    verdictResult
	}

	var bad []outcome
	green, refusing := 0, 0
	var refusals []string

	check := func(label string, f *File) {
		v := verdict(t, f)
		switch v.kind {
		case passed:
			green++
		case knownRefusal:
			refusing++
			refusals = append(refusals, fmt.Sprintf("%s (%s)", label, f.Refuses))
		default:
			bad = append(bad, outcome{label, v})
		}
	}

	for _, tier := range smallestUsefulCorpus {
		for name, f := range tierFiles(t, tier) {
			check(tier+"/"+name, f)
		}
	}
	for name, f := range callFormSlice(t) {
		check(tierSubroutines+"/"+name+" (slice)", f)
	}

	sort.Strings(refusals)
	t.Logf("smallest useful corpus: %d files pass, %d refuse with a cited gap",
		green, refusing)
	for _, r := range refusals {
		t.Logf("  documented refusal: %s", r)
	}

	// A gate over no files would pass vacuously, which is the one failure
	// this test cannot afford. Four tiers and a five-file slice is far
	// more than this, and the floor is only here so that a glob returning
	// nothing fails instead of reporting success.
	if total := green + refusing + len(bad); total < len(callFormFiles) {
		t.Fatalf("the gate ran over %d files, which cannot be tiers 01-04 "+
			"plus a %d-file slice.\n\tA gate over nothing passes vacuously.",
			total, len(callFormFiles))
	}

	sort.Slice(bad, func(i, j int) bool { return bad[i].file < bad[j].file })
	for _, b := range bad {
		t.Errorf("%s is %s, not green:\n\t%s\n"+
			"\tThe gate admits a file that PASSES and a file that refuses "+
			"with a documented, still-accurate citation. It admits neither "+
			"a corpus bug, nor a marker that has gone stale, nor an "+
			"undocumented refusal.",
			b.file, b.v.kind, strings.Join(b.v.msgs, "\n\t"))
	}
}
