// ABOUTME: Tier 02 checked against the finished tooling, the first tier with a real prerequisite.
// ABOUTME: Six checks the tier's own issue names, the adjacency one pairing with the DECLARED tier.
package conformance

import (
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierVariables is the tier this file is about.
//
// A constant for the same reason tier 01 has one: the tier number is a
// POSITION, and TestTierNumberingRegenerates records the spec's statement
// that positions move.
const tierVariables = "02_variables"

// TestTierVariablesCoversGlossary checks the tier against the vocabulary
// it asserts in.
//
// GLOSSARY.md's `variable` entry is written FROM perl and names the
// boundaries at which a lexer can be wrong about what one variable IS.
// A tier that asserts in that vocabulary while covering only the
// spellings it happened to think of has coverage by accident.
//
// The entries below are the glossary's own boundaries for the category
// this tier owns, plus the two constructs the tier's README names in its
// title line and then does not exercise. Two are worth the words:
//
//   - `${name}` IS ONE VARIABLE. The glossary says so explicitly --
//     "the braces are punctuation around a name" -- and measured,
//     `my $x = 42; print ${x}` prints 42 with the same `padsv` the
//     unbraced spelling gives. A lexer that reads `${` as the start of a
//     block, or as the dereference `${ $ref }` it is NOT, is wrong here
//     and nowhere else in the tier.
//   - `exists` HAS NO OP IN ANY SPELLING. The README devotes its longest
//     passage to this and the tier then shipped without a file that
//     writes it: measured, `exists $h{a}` is `multideref($h{"a"})
//     sK/EXISTS`, the construct surviving only as a FLAG. An op-derived
//     tier would contain no notion of `exists` at all, which is exactly
//     why the declaration has to be checked against files rather than
//     trusted.
//
// Matching is on SOURCE TEXT rather than on tokens, for tier 01's
// reason: a token-based check would ask our lexer whether the tier covers
// a case our lexer gets wrong, and answer no.
func TestTierVariablesCoversGlossary(t *testing.T) {
	// Each entry is a boundary and a spelling that exercises it. The
	// spellings carry their surrounding syntax so a match means the
	// construct is present rather than merely the characters.
	boundaries := []struct{ boundary, spelling string }{
		{"scalar sigil on a package name", "$::"},
		{"array sigil", "my @a ="},
		{"hash sigil", "my %h ="},
		{"last-index sigil", "$#a"},
		{"array element, constant subscript", "$a[0]"},
		{"array element, computed subscript", "$a[$#a]"},
		{"hash element, bareword key", "$h{a}"},
		{"hash element, computed subscript", "$h{$k[0]}"},
		{"array slice", "@a[0"},
		{"hash slice", "@h{"},
		{"`${name}` is one variable", "${x}"},
		{"delete on an element", "delete $h{a}"},
		{"delete on a slice", "delete @h{"},
		{"exists", "exists $h{a}"},
	}

	files := tierFiles(t, tierVariables)

	var sources []string
	for _, f := range files {
		sources = append(sources, f.Source)
	}
	joined := strings.Join(sources, "\n")

	for _, b := range boundaries {
		if !strings.Contains(joined, b.spelling) {
			t.Errorf("no file in %s spells %s (%s)\n"+
				"\tThe tier asserts in a vocabulary it does not cover.",
				tierVariables, b.spelling, b.boundary)
		}
	}
}

// TestTierVariablesPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// Every file's header says MEASURED perl 5.42.0, which is a claim about
// an interpreter; `perlPath` is what makes it checkable, resolving the
// one perl reporting `$] = 5.042000` rather than guessing.
//
// Two claims per file, and they are different:
//
//   - perl AGREES with the file's `--- expect parses` / `--- expect
//     parsent` bit. A file perl contradicts measured nothing about our
//     parser.
//   - perl PRINTS what the file pins, byte for byte, when it pins
//     anything.
//
// This is `verdict`'s first phase run for its answer rather than for its
// effect on the suite, so a CORPUS BUG is reported as one file's problem
// and not as a refusal.
func TestTierVariablesPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierVariables, perl)

	for name, f := range tierFiles(t, tierVariables) {
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

// TestTierVariablesLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a tier landing with a file that reaches
// forward fails HERE, named as this tier's problem, rather than as one
// subtest among a hundred and thirteen.
//
// This tier is where the lint earns its keep rather than merely running.
// Measured, the obvious spellings of the tier's own constructs drag later
// tiers in: `print exists $h{a} ? 1 : 0` adds `cond_expr`, `print
// 0+(exists $h{a})` adds `add`, and `length($$)` adds three ops at once.
// Each is a file that looks like it is about a tier-02 construct and is
// not, and the lint is what says so.
//
// What the lint proves is one direction only: no file uses an op that no
// tier at or before 02 introduces. It cannot derive the tier -- `$h{a}`
// arrives as `multideref` with no `helem`, and `exists` arrives as a FLAG
// with no op at all -- which is why a file DECLARES its tier and the ops
// lint the declaration.
func TestTierVariablesLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierVariables) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierVariables, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// TestTierVariablesAdjacency checks the tier's adjacency file holds every
// construct the tier introduces, each next to another, AND next to a
// construct from the tier it declares itself to depend on.
//
// TWO HALVES, AND THE SECOND IS WHAT MAKES THIS TIER DIFFERENT FROM 01.
//
// Tier 01 depends on `nothing`, so its adjacency file had only the
// coverage half to check. Tier 02's README declares a prerequisite, and
// the spec is explicit that the pairing follows the DECLARATION and not
// the arithmetic:
//
//	"Where a tier genuinely depends on an earlier one, the adjacency
//	file pairs with the DECLARED PREREQUISITE rather than with N-1 --
//	which matters for 09_regex, 12_packages and the other tiers this
//	document already says have no N-1 dependency. Pairing those with
//	N-1 would assert nothing."
//
// So the prerequisite is READ FROM THE README via readTierDeps, not
// computed as 02-1. For this tier the two happen to coincide, which is
// precisely why it must not be written as the subtraction: a check that
// is right here for the wrong reason is a check that is silently wrong
// at 09 and 12, and the tiers that would catch it are other agents'
// files.
//
// WHY BOTH HALVES ARE ON THE SOURCE, WHICH IS NOT THE OBVIOUS CHOICE.
//
// The obvious prerequisite check is on ops: demand that the adjacency
// file emit an op tier 01 introduces. Perl refutes it. Measured 5.42.0,
// a program with NO tier-01 construct in it at all --
//
//	my @a = (1);
//	print scalar(@a), "\n";
//
// -- emits `const enter leave nextstate print pushmark`, six of tier
// 01's nine declared ops, because the fixtures every corpus file uses
// (`print`, a statement, a program) are themselves tier-01 ops. An
// op-based prerequisite check would go green over a file containing
// nothing from the prerequisite tier, which is the failure mode the
// spec's own revision history describes for N-1 pairing: satisfied by a
// file that catches nothing.
//
// The coverage half cannot use ops either, for tier 01's recorded
// reason: `padrange` absorbs `pushmark`, so a file with MORE adjacent
// constructs emits FEWER ops, and the declared set is a UNION across the
// tier's files rather than a property of any single one.
//
// What this ESTABLISHES: every construct this tier teaches, and at least
// one the prerequisite teaches, appear in one compiled body, so a parser
// that handles each alone and mis-handles the pair is reachable. What it
// does NOT establish is adjacency in any sense stronger than "in the
// same body" -- source position would pin a layout the file is free to
// change, and the file's prose makes the stronger claim for review.
func TestTierVariablesAdjacency(t *testing.T) {
	files := tierFiles(t, tierVariables)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierVariables, adjacencyFile)
	}

	// COVERAGE. The construct each non-adjacency file introduces, taken
	// from that file's own `# INTRODUCES` line and matched by the
	// spellings the tier's glossary check already enumerates -- rather
	// than from a third list here, which is how this package has drifted
	// four times.
	//
	// A tier-02 construct is not a single `= <literal>;` binding the way
	// a tier-01 one is: `$a[0] = 7` and `delete @h{...}` are statements,
	// not bindings. So the spelling comes from the file's source through
	// variableSpellings, which returns every variable-ish spelling the
	// file uses.
	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		missing := missingFrom(adj.Source, variableSpellings(f.Source))
		if len(missing) > 0 {
			t.Errorf("%s introduces %s, which %s does not contain.\n"+
				"\tThe adjacency file must hold every construct the tier "+
				"introduces, or the pairing it exists to reach is not "+
				"reachable for that construct.",
				name, strings.Join(missing, ", "), adjacencyFile)
		}
	}

	// THE PREREQUISITE HALF. Declared, not computed.
	deps, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	prereq, ok := deps[tierVariables]
	if !ok {
		t.Fatalf("%s declares no prerequisite", tierVariables)
	}
	if prereq == "nothing" {
		// Not a failure in general -- tier 01 says exactly this -- but
		// for THIS tier it is, because the README's prose depends on a
		// prerequisite and a `nothing` here means the two disagree.
		t.Fatalf("%s declares no prerequisite, so there is nothing for its "+
			"adjacency file to pair with; the README's prose says otherwise",
			tierVariables)
	}

	// The prerequisite's constructs come from ITS files, the same way
	// this tier's come from this tier's: the corpus is the enumeration.
	// A construct of the prerequisite that appears in the adjacency file
	// is the pairing; none appearing means the file asserts nothing
	// about the tier it claims to need.
	prereqFiles := tierFiles(t, prereq)
	var paired []string
	for name, f := range prereqFiles {
		if name == adjacencyFile {
			continue
		}
		spelling := literalSpelling(f.Source)
		if spelling == "" {
			continue
		}
		// The binding's right-hand side alone: the adjacency file holds
		// tier 01's literals inside tier 02's containers -- `(42, 0.5,
		// 'plain')` -- rather than re-binding them to scalars, so the
		// whole `= <literal>;` spelling would never match.
		value := strings.TrimSuffix(strings.TrimPrefix(spelling, "= "), ";")

		// AND IT MUST BE A CONSTRUCT THIS TIER'S OWN FILES DO NOT USE.
		// Without this the pairing is satisfied by a coincidence, and
		// mutation testing caught it doing exactly that: pointed at
		// 03_context, literalSpelling matched that tier's `my $n = @a;`
		// and offered `@a` as the shared construct -- which appears in
		// this tier's adjacency file because `@a` IS a tier-02
		// construct, not because anything of 03_context is present. The
		// check went green over a prerequisite the file exercises
		// nothing of.
		//
		// literalSpelling is tier 01's helper and reads tier 01's
		// `my $x = <literal>;` shape; applied to any other tier it
		// returns whatever sits after the `=`, which is that tier's
		// subject and frequently also ours. The construct files are the
		// filter, and they are the right one: a spelling that appears
		// in a tier-02 file teaching a tier-02 construct is ours no
		// matter which other tier also writes it, so its presence in
		// the adjacency file is explained without the prerequisite.
		//
		// This is what keeps the check meaningful for a prerequisite
		// that is NOT 01_literals -- the case the spec names for
		// 09_regex and 12_packages, and the reason this reads the
		// declaration rather than computing N-1.
		if ours(files, value) {
			continue
		}
		if strings.Contains(adj.Source, value) {
			paired = append(paired, value)
		}
	}
	if len(paired) == 0 {
		t.Errorf("%s declares DEPENDS ON %s, but %s contains no construct "+
			"from it.\n\tThe adjacency file pairs with the DECLARED "+
			"prerequisite; with nothing of %s in the body, the dependency "+
			"the tier claims is unexercised and the ordering claim is prose.",
			tierVariables, prereq, adjacencyFile, prereq)
	} else {
		t.Logf("paired with %s via %s", prereq, strings.Join(paired, ", "))
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
}

// tierVariableSpellings is the vocabulary a tier-02 construct file is
// searched for.
//
// This is deliberately a list of SPELLINGS rather than a pattern. A
// pattern general enough to match "any variable-ish thing" matches
// `$::s = $a[0]` as four constructs and `"$w[0]-$w[1]"` as two more,
// and the adjacency check then demands the adjacency file contain
// fragments of itself. The spellings here are the ones the tier's files
// actually introduce, which is a closed set the glossary check pins.
//
// EACH ENTRY IS KEY-AGNOSTIC WHERE THE CONSTRUCT IS. `delete $h{d}` and
// `delete $h{a}` are one construct -- deleting an element -- and which
// key a file happens to use is a fact about the file, not about what it
// teaches. Writing `delete $h{a}` here would make the adjacency file
// wrong for choosing a different key, which is a check on a coincidence.
// So an entry ends at the `{` or `[` wherever what follows is a key, and
// carries the subscript only where the SUBSCRIPT IS THE CONSTRUCT:
// `$a[$#a]` and `$h{$k[0]}` are the computed-subscript cases, which is
// the whole reason those two files exist.
//
// Ordering does not matter here -- every entry is tested independently
// and nesting between them (`delete @h{` contains `@h{`) only means a
// file using the longer one reports both, which is true.
var tierVariableSpellings = []string{
	"my @a =",
	"my %h =",
	"$a[0]",
	"$a[$#a]",
	"$#a",
	"@a[0",
	"$h{a}",
	"$h{$k[0]}",
	"@h{",
	"delete @h{",
	"delete $h{",
	"exists $h{",
	"${",
	"@{",
	"$::",
	"@::p",
	"%::g",

	// THE AGGREGATE-ARGUMENT OPERATORS. Each entry carries the SIGIL
	// that follows the keyword and stops there, because the sigil is
	// the construct: these four take an AGGREGATE where every other
	// list operator takes an expression, so `push @` is the claim and
	// `push` alone would be satisfied by a body writing it over a
	// scalar.
	//
	// NAME-AGNOSTIC, on this list's own stated principle. Which array
	// is pushed to and which hash is iterated are facts about a file,
	// not about what it teaches, and the adjacency file's `each` runs
	// over a separate one-key hash precisely because hash order makes
	// the four-key one unpinnable. An entry spelling `each %h` would
	// make the check fail for that correct choice.
	//
	// Without these four the adjacency check goes green over the
	// constructs, because it demands only what this list names -- which
	// is how a new file's construct reaches the tier unpaired.
	"push @",
	"unshift @",
	"values %",
	"each %",
}

// variableSpellings returns the tier-02 constructs a file's source uses.
func variableSpellings(source string) []string {
	var out []string
	for _, s := range tierVariableSpellings {
		if strings.Contains(source, s) {
			out = append(out, s)
		}
	}
	return out
}

// ours reports whether a spelling appears in one of this tier's own
// construct files, and is therefore explained without any prerequisite.
//
// The adjacency file is excluded because it is the body under test: a
// spelling is there by definition, and counting it would make every
// spelling this tier's own.
func ours(files map[string]*File, spelling string) bool {
	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		if strings.Contains(f.Source, spelling) {
			return true
		}
	}
	return false
}

// missingFrom returns the spellings a body does not contain.
func missingFrom(body string, spellings []string) []string {
	var out []string
	for _, s := range spellings {
		if !strings.Contains(body, s) {
			out = append(out, s)
		}
	}
	return out
}

// TestTierVariablesRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// A message is prose: reword it and a file that quoted it breaks over a
// change that moved nothing. A code changes only when the reason the
// parser declines changes, which is the event a refusing file wants to
// be told about -- and `run.go` already fails a file whose named code is
// not among the ones it actually produces.
//
// THE RULE IS BICONDITIONAL, and tier 01's reasoning transfers exactly.
// `parse.RefusalCode` names a site in the PARSER, so a file refusing
// purely on a TOKEN fact has no Unknown to name; naming one anyway makes
// the file fail as a stale marker -- correctly, because the claim would
// be false.
//
//   - A file whose parser produces Unknowns MUST name one of their
//     codes. It has a site to name and prose would be the alternative.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead. A lexical
//     refusal's citation is the assertion itself.
//
// Both halves are needed. Without the first, a parse refusal may stay
// uncited; without the second, a lexical refusal may cite a parse site it
// does not have, which is the false claim the codes exist to forbid.
//
// THIS TIER REFUSES NOWHERE TODAY, and the test is written anyway rather
// than deferred. Measured against the parser at this revision, all
// thirteen files return zero Unknowns, so every subtest here is vacuous
// -- which is the normal state for a tier the parser has caught up with,
// not evidence the check is unnecessary. The moment a file here starts
// refusing, this is what decides whether it cites a code or token facts,
// and a tier whose refusal discipline is only written once it has a
// refusal gets the discipline wrong on the file that needed it.
func TestTierVariablesRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierVariables) {
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

// TestTierVariablesNumberingRegenerates checks the tier's numbering is
// what a regeneration would produce.
//
// The spec makes numbering DERIVED -- "the number is DERIVED from the
// classification, so a reorder is a regeneration rather than a
// hand-edit" -- and this tier adds files, which is exactly the event
// that puts a hand-maintained sequence out of step. regenerateNumbering
// is tier 01's; the check is run here because the tier that ADDS files
// is the one that can break it.
func TestTierVariablesNumberingRegenerates(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(corpusDir, tierVariables, "*.t"))
	if err != nil {
		t.Fatalf("globbing %s: %v", tierVariables, err)
	}

	var names []string
	for _, p := range paths {
		base := filepath.Base(p)
		if base != adjacencyFile && !reNumbered.MatchString(base) {
			t.Errorf("%s is not numbered `NN_name.t`, so nothing can place it", base)
		}
		names = append(names, base)
	}

	have := map[string]bool{}
	for _, n := range names {
		have[n] = true
	}
	for identity, regenerated := range regenerateNumbering(names) {
		if !have[regenerated] {
			t.Errorf("regenerating %s's numbering puts %s at %s, which is not on disk",
				tierVariables, identity, regenerated)
		}
	}
	if !have[adjacencyFile] {
		t.Errorf("%s has no %s; 00 is reserved for it and the construct "+
			"numbering starts at 01", tierVariables, adjacencyFile)
	}
}
