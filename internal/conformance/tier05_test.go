// ABOUTME: Tier 05 checked against the finished tooling: my, our, state, local, blocks.
// ABOUTME: Five checks the tier's own issue names, each tier-specific rather than corpus-wide.
package conformance

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierScoping is the tier this file is about.
//
// A constant rather than a literal at five call sites, for the reason
// tier01_test.go gives for tierLiterals: the tier number is a POSITION
// and positions move.
const tierScoping = "05_scoping"

// TestTierScopingPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// Identical in intent to TestTierLiteralsPerlValidated, and it carries
// more weight here. Four of this tier's five constructs emit no op of
// their own -- `my`'s pad ops are tier 01's, `our` and `local` share
// tier 02's `gvsv`/`sassign` and differ only by a flag the ops list
// cannot see -- so the OUTPUT is the only thing that tells them apart. A
// pinned output that perl does not actually produce would leave this
// tier with no discriminating measurement at all.
func TestTierScopingPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierScoping, perl)

	for name, f := range tierFiles(t, tierScoping) {
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

// TestTierScopingLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a file that reaches forward fails HERE,
// named as this tier's problem rather than as one subtest among a
// hundred and thirteen.
//
// The lint has a specific thing to catch in this tier, and it is not
// hypothetical. 05_bare_block.t records that probing the loop-ness of a
// bare block behaviourally would take a `last`, which compiles to a
// literal `last` op that no tier at or before 05 claims. A file that
// added the probe anyway would satisfy its own output pin and fail here,
// which is the lint doing exactly what it is for.
func TestTierScopingLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierScoping) {
		t.Run(name, func(t *testing.T) {
			if err := lintOps(t, f.Source, tierScoping, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// reScopingKeyword matches a declaration keyword at the head of a
// statement.
//
// The KEYWORD rather than a whole spelling, which is where this differs
// from tier 01's literalSpelling. Tier 01's constructs are VALUES and
// each file's value is unique, so the spelling identifies the construct.
// This tier's constructs are DECLARATION FORMS and every file declares a
// scalar: `my $x = 1` and `my $n = 1` are the same construct, so a check
// keyed on the whole statement would ask the adjacency file to repeat a
// variable name rather than to exercise a form.
//
// Anchored at line start after optional indentation, so `$mystery` and a
// `local` inside a comment are not mistaken for declarations.
var reScopingKeyword = regexp.MustCompile(`(?m)^[ \t]*(my|our|state|local)\b`)

// reBareBlock matches a bare block: a `{` that opens a statement.
//
// Its own pattern because a bare block has no keyword. What makes it
// BARE is that nothing precedes the brace -- an `if (...) {` or a `sub f
// {` would put something there, and those compile to `enter`/`leave`
// rather than to this tier's `enterloop`/`leaveloop`, which is precisely
// the distinction 05_bare_block.t exists to make.
var reBareBlock = regexp.MustCompile(`(?m)^[ \t]*\{[ \t]*$`)

// scopingForms returns the scoping forms a source exercises, named as
// this tier's README names them.
func scopingForms(source string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range reScopingKeyword.FindAllStringSubmatch(source, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	if reBareBlock.MatchString(source) {
		out = append(out, bareBlockForm)
	}
	return out
}

// bareBlockForm is how the one keyword-less construct is named.
const bareBlockForm = "bare block"

// formFromName returns the scoping form a construct file is named for,
// or "" when its name names none.
//
// THE NAME, not the source, and this is the substantive difference from
// tier 01's literalSpelling. Every file in this tier contains a `my`:
// 01_my.t because `my` is its subject, 02_our.t and 03_state.t because a
// shadow needs one, 05_bare_block.t because a scope needs something to
// close over. Reading each file's subject from its source would report
// `my` five times and the tier would look like one construct.
//
// The spec already makes the name a file's identity -- "A file's
// identity is its NAME; the number is its current position" -- so taking
// the subject from it is reading a declaration rather than inventing a
// second one. The source is then used to check the name is not lying.
func formFromName(name string) string {
	m := reNumbered.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	switch identity := strings.TrimSuffix(m[2], ".t"); identity {
	case "my", "our", "state", "local":
		return identity
	case "bare_block":
		return bareBlockForm
	default:
		return ""
	}
}

// TestTierScopingAdjacency checks the tier's adjacency file holds every
// construct the tier introduces, each next to another, and that it pairs
// with the DECLARED prerequisite.
//
// WHAT OPS CANNOT DO HERE, and this tier is where the refutation is
// sharpest. The obvious check is that the adjacency file emits the
// tier's declared INTRODUCES set. Perl refutes it: `padrange` ABSORBS
// `pushmark`, so consecutive `my` declarations fuse into one op where
// several would otherwise stand and a file with MORE constructs emits
// FEWER ops. The declared set is a UNION ACROSS THE TIER'S FILES and
// never a property of any one of them. This tier's adjacency file shows
// it in the other direction too: it holds more declarations than any
// construct file and emits no `padrange` at all, because its
// declarations are separated by other statements.
//
// So COVERAGE is checked on the SOURCE, as at tier 01. What is new here
// is the second half.
//
// THE PREREQUISITE PAIRING, which tier 01 could not test. Tier 01
// depends on `nothing`, so it had no earlier tier for its adjacency file
// to pair with. This tier's README declares `04_operators`, and the
// issue asks the pairing be with the DECLARED prerequisite rather than
// with tier N-1 -- a distinction the spec draws because the corpus holds
// tiers whose N-1 asserts nothing they need.
//
// That half IS checkable from ops, and only that half. The prerequisite
// is a TIER, tiers declare op sets, and an op is a fact about what perl
// compiled rather than about what the source appears to say. So: the
// adjacency file must emit at least one op `04_operators` introduces.
// Measured, this one does -- `add` and `multiply`, from the `$g + 10`,
// `$x * 2` and `$s + 100` in its initialisers.
//
// Why the OP and not the spelling. A source search for `+` is satisfied
// by a `+` inside a string, and worse, by a `+` the optimiser folds
// away: `my $x = 1+2` compiles to `const[IV 3] s/FOLD` with no `add` at
// all. A pairing asserted through a folded operator is a pairing that
// never happened. Requiring the op requires that the two tiers'
// constructs actually met in one compiled body, which is the claim.
//
// What this does NOT establish: that the constructs are adjacent in any
// stronger sense than "in the same body" -- the same limit tier 01
// records, for the same reason.
func TestTierScopingAdjacency(t *testing.T) {
	files := tierFiles(t, tierScoping)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierScoping, adjacencyFile)
	}
	adjacent := scopingForms(adj.Source)

	// The construct each non-adjacency file introduces, taken from that
	// file's own identity rather than from a list here: the tier's files
	// ARE the enumeration of what it introduces, and a second list beside
	// them is how this package has drifted four times already.
	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		form := formFromName(name)
		if form == "" {
			t.Errorf("%s: its name names no scoping form, so nothing places it", name)
			continue
		}
		if !slices.Contains(scopingForms(f.Source), form) {
			t.Errorf("%s is named for %q, which its own source does not use.\n"+
				"\tThe name is the file's identity; a name the source does "+
				"not honour makes every other check here ask about the "+
				"wrong construct.", name, form)
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
	dep, ok := deps[tierScoping]
	if !ok {
		t.Fatalf("%s declares no DEPENDS ON", tierScoping)
	}
	if dep == "nothing" {
		t.Fatalf("%s declares no prerequisite, but the issue asks its "+
			"adjacency file to pair with one", tierScoping)
	}

	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	depOps := tiers[dep]
	if len(depOps) == 0 {
		t.Fatalf("%s declares %s, which introduces no op to pair with", tierScoping, dep)
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
			tierScoping, dep, adjacencyFile, strings.Join(depOps, " "))
		return
	}
	t.Logf("%s pairs with %s through %s", adjacencyFile, dep, strings.Join(paired, ", "))
}

// TestTierScopingRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// The reasoning is tier 01's and transfers exactly, so it is stated
// briefly here and in full there. `parse.RefusalCode` names a site in
// the PARSER. A file refusing purely on a token fact has no parser
// Unknown to name, and naming one anyway makes `run.go` fail it as a
// stale marker -- correctly, because the claim would be false. So the
// rule is biconditional rather than "every refusing file names a code":
//
//   - A file whose parser produces Unknowns MUST name one of their
//     codes.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// AS OF TODAY THIS TIER HAS NO REFUSING FILE, so the check is vacuous.
// That is a measurement rather than an oversight, and a notable one:
// every one of the five scoping forms parses today, which is worth
// knowing given that four of them are distinguished only by output. The
// check is still written, because the event it guards is a file
// ACQUIRING a refusal -- the next parser change that declines `state` or
// `local` lands a `# STATUS refuses` line here, and this is what then
// requires it to cite a code rather than a sentence.
func TestTierScopingRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierScoping) {
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
