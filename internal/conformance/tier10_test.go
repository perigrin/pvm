// ABOUTME: Tier 10 validated against the finished tooling: the checks its issue names, plus the angle split.
// ABOUTME: The tier whose subject -- printing SOMEWHERE -- the op table cannot express.
package conformance

import (
	"regexp"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierIo is the tier this file is about.
//
// A constant rather than a literal at five call sites, for the reason
// `tierLiterals` is one: the tier number is a POSITION, and this tier's
// own README says so at length -- it could sit anywhere after 03 and the
// dependency check would not object.
const tierIo = "10_io"

// TestTierIoPerlValidated runs every file in the tier through the pinned
// interpreter before it counts.
//
// The same two claims tier 01 checks -- perl agrees with the parses bit,
// perl prints what the file pins -- and one reason of this tier's own to
// want them checked. Every file here opens an IN-MEMORY handle rather
// than a path, because a path is not reproducible across machines. That
// substitution is only sound if perl actually produces the pinned bytes
// from it, and this is where that is established rather than asserted:
// `open(my $fh, "<", \"one\ntwo\n")` is a claim about an interpreter's
// behaviour on a scalar ref, and the interpreter is the only thing that
// can confirm it.
func TestTierIoPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierIo, perl)

	for name, f := range tierFiles(t, tierIo) {
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

// TestTierIoLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so that a tier landing with a file that reaches
// forward fails HERE, named as this tier's problem.
//
// This tier gives the lint more to say than most, because its files
// deliberately reach BACKWARD in three places that each look like a
// violation and are not: `srefgen` from `\my $buf` is tier 08's, `gv`
// from a bareword handle is tier 02's, and `cond_expr` from the
// adjacency file's ternary is tier 06's. Using a construct is not
// introducing it, and the lint's one direction -- nothing reaches
// FORWARD -- is what keeps that distinction checkable while the tier's
// INTRODUCES set stays small.
func TestTierIoLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierIo) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierIo, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// tierContext is the tier that 10_io declares a dependency on.
//
// Named rather than spelled at three call sites, and checked against the
// README rather than assumed, because the whole point of the DEPENDS ON
// block is that the prerequisite is DECLARED. A constant here that
// disagreed with the README would be the two-lists-that-must-agree
// failure this package has paid for four times; the check above is what
// makes them one list.
const tierContext = "03_context"

// reHandleRead matches a readline on a named lexical handle: the `<$fh>`
// and nothing else.
var reHandleRead = regexp.MustCompile(`<(\$\w+)>`)

// readsOneHandleTwice reports whether a source reads the SAME lexical
// handle at least twice.
//
// Text rather than ops, because the op stream records `readline` twice
// and cannot say whose handle each read. The handle is a `padsv` before
// it, and two `padsv` ops are indistinguishable in a stream that carries
// no pad indices -- `opsOf` keeps only the op name, deliberately, so the
// lint sees what perl built and not which slot it built it in.
func readsOneHandleTwice(source string) bool {
	counts := map[string]int{}
	for _, m := range reHandleRead.FindAllStringSubmatch(source, -1) {
		counts[m[1]]++
		if counts[m[1]] >= 2 {
			return true
		}
	}
	return false
}

// TestTierIoRefusalsCited checks that a refusing file names WHICH site
// declines, from the stable inventory, rather than describing it.
//
// The rule is tier 01's and is BICONDITIONAL for the same reason, which
// this tier arrives at by a different road. `parse.RefusalCode` names a
// site in the PARSER, and this tier's hard cases are LEXICAL: measured,
// `my $l = <$fh>;` produces zero parser Unknowns, because the lexer
// already resolved `<$fh>` into one Readline token and the parser sees an
// ordinary term. A file refusing on a token fact therefore has no site to
// name, and naming one anyway would make `run.go` fail it as a stale
// marker -- correctly, since the claim would be false.
//
//   - A file whose parser produces Unknowns MUST name one of their codes.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// Both halves are needed. Without the first a parse refusal may stay
// uncited; without the second a lexical refusal may cite a parse site it
// does not have, which is the false claim the codes exist to forbid.
//
// The tier is green today and the loop below runs over nothing, which is
// the correct state of a check on refusals in a tier that has none. It is
// here for the file that refuses tomorrow: the cost of a check with no
// subject is nothing, and the cost of finding out later that a refusal
// landed uncited is a commit.
func TestTierIoRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierIo) {
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

// TestTierIoAngleSplitIsExplicit checks that this tier asserts `<$fh>` is
// ONE TOKEN, which is the only place the corpus can say so.
//
// THE SPLIT THIS TIER'S ISSUE ASKS FOR. `glob-angle` is one of the twelve
// hard markers and it places at tier 13, not here: `<*>` is a delimiting
// problem and `<$fh>` is an IO one, and they are different tiers despite
// the shared spelling. The issue asks the files to make that split
// explicit, and the split is not visible from either half alone.
//
// Measured under 5.42.0, the two halves diverge at the optree and are
// identical at the lexer:
//
//	<*.nonexistent-xyz>    pushmark const gv glob
//	<$fh>                  padsv readline
//
// Different ops. But `conformance/categories.go` maps BOTH onto the one
// `readline operator` category and says why: our lexer cannot tell a
// handle name from a glob pattern without knowing what the name means,
// which is a parsing question and not a lexing one. So the token category
// is the thing the two tiers SHARE, and the ops are where they part.
//
// Tier 13 already makes its half: `05_glob_angle.t` asserts one readline
// operator whose text is `<*.nonexistent-xyz>` and pins `glob`'s empty
// result. Without the matching half here, the corpus asserts the token
// category for the glob spelling and never for the handle spelling -- and
// a lexer that resolved `<$fh>` into three tokens (`<`, `$fh`, `>`) would
// pass every file in this tier, because the parser reads that as a
// perfectly good comparison chain and returns ZERO Unknowns for it. That
// is not hypothetical: it is what our parser does with `<STDIN>` today if
// the lexer hands it the pieces, and the behavioural files here could not
// tell.
//
// So the check is on the tier's own token facts rather than on its
// sources: at least one file must assert the readline operator category
// for a HANDLE spelling. Asserting it in a comment would be prose; a
// token fact is run.
func TestTierIoAngleSplitIsExplicit(t *testing.T) {
	var asserted []string
	for name, f := range tierFiles(t, tierIo) {
		for _, fact := range f.TokenFacts {
			if !strings.Contains(fact, "readline operator") {
				continue
			}
			// A handle spelling, not a glob pattern. `<$fh>` and `<>` and
			// `<STDIN>` are handles; a `*` inside the angles is tier 13's
			// subject and asserting it here would satisfy this check with
			// the other tier's fact.
			if strings.Contains(fact, "*") {
				t.Errorf("%s asserts %s, which is a GLOB spelling.\n"+
					"\tThe angle-bracket glob is tier 13's construct; a file "+
					"here asserting it records the split in the wrong tier.",
					name, fact)
				continue
			}
			asserted = append(asserted, name+": "+fact)
		}
	}

	if len(asserted) == 0 {
		t.Errorf("no file in %s asserts a `readline operator` token fact for a "+
			"handle spelling.\n"+
			"\tTier 13's 05_glob_angle.t asserts that category for `<*...>`, so "+
			"without a file here the corpus claims the category only for the "+
			"glob half of a spelling two tiers share.\n"+
			"\tThe assertion matters because our parser reads `<`, `$fh`, `>` "+
			"as a comparison chain and returns zero Unknowns for it: a lexer "+
			"that split the angles would pass every behavioural file in this "+
			"tier unnoticed.", tierIo)
		return
	}
	t.Logf("the handle half of the angle split, asserted by %d fact(s):\n\t%s",
		len(asserted), strings.Join(asserted, "\n\t"))
}
