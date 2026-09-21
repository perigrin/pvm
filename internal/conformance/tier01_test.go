// ABOUTME: Tier 01 re-validated against the finished tooling, which is what proves the tooling works.
// ABOUTME: Six checks the tier's own issue names, each tier-specific rather than corpus-wide.
package conformance

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierLiterals is the tier this file is about.
//
// A constant rather than a literal at six call sites, because the tier
// number is a POSITION and positions move -- see TestTierNumberingRegenerates
// for the spec's statement that they do.
const tierLiterals = "01_literals"

// tierFiles returns the parsed corpus files of one tier, keyed by base
// name, adjacency file included.
func tierFiles(t *testing.T, tier string) map[string]*File {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(corpusDir, tier, "*.t"))
	if err != nil {
		t.Fatalf("globbing %s: %v", tier, err)
	}
	if len(paths) == 0 {
		t.Fatalf("%s holds no corpus files", tier)
	}

	out := map[string]*File{}
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("reading %s: %v", p, err)
		}
		f, err := ParseFile(string(raw))
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		out[filepath.Base(p)] = f
	}
	return out
}

// TestTierLiteralsCoversGlossary checks the tier against the vocabulary
// it asserts in.
//
// GLOSSARY.md is written FROM perl -- each entry cites perldata/perlop or
// a measured invocation -- so its boundary cases are a list of the ways a
// lexer can be wrong about a literal. A tier that asserts in that
// vocabulary while covering only the cases it happened to think of is a
// tier whose coverage is an accident.
//
// The list below is the glossary's own boundary cases for the two
// literal categories this tier owns, not a fresh opinion about what
// matters. Each entry names the SPELLING that must appear in some file's
// source; a spelling nowhere in the tier is a boundary the corpus makes
// no claim about.
//
// Two of them are why this check exists rather than being taken on trust:
//
//   - `-1` IS TWO TOKENS. A unary minus and the literal `1`. Measured,
//     perl folds it to `const[IV -1]` and the optree shows one constant,
//     so the optree cannot adjudicate it and only the token stream can.
//   - A V-STRING IS NOT A NUMERIC LITERAL. `v65.66.67` and the bare
//     three-part `65.66.67` are strings built from ordinals, which
//     `perldata` documents under "Version Strings", away from numbers.
//
// Matching is on the SOURCE TEXT rather than on tokens on purpose. A
// token-based check would ask our lexer whether the tier covers a case
// our lexer gets wrong, and answer no -- which is backwards: the whole
// point of a refusing file is to cover a case we get wrong.
func TestTierLiteralsCoversGlossary(t *testing.T) {
	// Each entry is a glossary boundary and a spelling that exercises it.
	// Spellings are substrings of a file's `--- source`, so `0.5` would
	// be satisfied by `10.51` -- which is why each is written with the
	// surrounding syntax that pins it.
	boundaries := []struct{ boundary, spelling string }{
		{"decimal integer", "= 42;"},
		{"decimal with fractional part", "= 0.5;"},
		{"leading decimal point", "= .5;"},
		{"signed exponent", "= 5e-1;"},
		{"underscore separators", "= 4_294_967_296;"},
		{"hexadecimal", "= 0xff;"},
		{"binary", "= 0b1010;"},
		{"octal by leading zero", "= 0377;"},
		{"octal by explicit prefix", "= 0o377;"},
		{"trailing decimal point", "= 1.;"},
		{"`-1` is two tokens", "= -1;"},
		{"v-string with a leading v", "= v65.66.67;"},
		{"v-string by three parts", "= 65.66.67;"},
		{"single-quoted string literal", "= 'plain';"},
		{"double-quoted string literal", `= "`},
	}

	files := tierFiles(t, tierLiterals)

	var sources []string
	for _, f := range files {
		sources = append(sources, f.Source)
	}
	joined := strings.Join(sources, "\n")

	for _, b := range boundaries {
		if !strings.Contains(joined, b.spelling) {
			t.Errorf("no file in %s spells %s (%s)\n"+
				"\tGLOSSARY.md names this boundary for a category this "+
				"tier owns, so the tier asserts in a vocabulary it does "+
				"not cover.", tierLiterals, b.spelling, b.boundary)
		}
	}
}

// reNumbered matches a corpus file's `NN_name.t` spelling.
var reNumbered = regexp.MustCompile(`^(\d\d)_(.+\.t)$`)

// regenerateNumbering returns the names a tier's construct files would
// have if their numbering were regenerated from scratch.
//
// The spec makes numbering a DERIVED quantity, and says so where it
// rejects sparse numbers:
//
//	"two digits, not four, and no gaps. ... renumbering is not
//	expensive: the number is DERIVED from the classification, so a
//	reorder is a regeneration rather than a hand-edit."
//	"A file's identity is its NAME; the number is its current
//	position, and positions move."
//
// So this is a function that PRODUCES the numbering, not a predicate that
// inspects it. The difference matters: a predicate asserting "the current
// numbers have no gap" is satisfied by any gapless sequence, including
// one a human maintained by hand, and says nothing about whether a
// regeneration is possible. Running the regeneration and comparing is the
// only thing that checks the spec's actual promise -- that the numbers
// can be thrown away and rebuilt from the names.
//
// The order is the file names' own, since a corpus file carries no
// declared position; the number is assigned by sorting on the identity
// part and counting from 01. The adjacency file is excluded and keeps 00,
// which is reserved for it: it is not one of the tier's constructs but
// the check that they compose, so it sorts ahead of them rather than
// among them.
func regenerateNumbering(names []string) map[string]string {
	var constructs []string
	for _, n := range names {
		if n == adjacencyFile {
			continue
		}
		m := reNumbered.FindStringSubmatch(n)
		if m == nil {
			continue
		}
		constructs = append(constructs, m[2])
	}
	sort.Strings(constructs)

	out := map[string]string{}
	for i, identity := range constructs {
		// Two digits with no gaps, counting from 01. Zero-padded for the
		// same reason the tiers are: lexicographic order must equal
		// numeric order, which an unpadded 10 would break.
		out[identity] = string(rune('0'+(i+1)/10)) + string(rune('0'+(i+1)%10)) + "_" + identity
	}
	return out
}

// TestTierNumberingRegenerates checks that the tier's numbering is what a
// regeneration would produce.
//
// Tier 01 shipped numbered 02, 03, 04 with no 01, because the file that
// would have been 01 was never written. That gap is harmless until the
// numbering is claimed to be derived, at which point it is a number
// nothing produces.
//
// What this establishes: the on-disk names equal `regenerateNumbering`'s
// output, so the numbers could be deleted and rebuilt from the names
// alone. What it does NOT establish is that the ORDER is meaningful --
// alphabetical order is the corpus's convention, not a claim that
// `02_decimal.t` teaches something `03_hexadecimal.t` needs.
func TestTierNumberingRegenerates(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(corpusDir, tierLiterals, "*.t"))
	if err != nil {
		t.Fatalf("globbing %s: %v", tierLiterals, err)
	}

	var names []string
	for _, p := range paths {
		base := filepath.Base(p)
		if base != adjacencyFile && !reNumbered.MatchString(base) {
			t.Errorf("%s is not numbered `NN_name.t`, so nothing can place it", base)
		}
		names = append(names, base)
	}

	want := regenerateNumbering(names)
	have := map[string]bool{}
	for _, n := range names {
		have[n] = true
	}

	for identity, regenerated := range want {
		if !have[regenerated] {
			t.Errorf("regenerating %s's numbering puts %s at %s, which is not on disk",
				tierLiterals, identity, regenerated)
		}
	}

	// The adjacency file keeps 00 and is not part of the regeneration.
	if !have[adjacencyFile] {
		t.Errorf("%s has no %s; 00 is reserved for it and the construct "+
			"numbering starts at 01", tierLiterals, adjacencyFile)
	}
}

// TestTierLiteralsPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// Every file's header says MEASURED perl 5.42.0. That is a claim about an
// interpreter, and `perlPath` is what makes it checkable -- it resolves
// the one perl reporting `$] = 5.042000` and refuses to guess. Measured
// on the machine this was written on, PATH carried three perls and the
// right one won by ordering alone.
//
// Two things are asserted per file, and they are different claims:
//
//   - perl AGREES with the file's `--- expect parses` / `--- expect
//     parsent` bit. A file whose expectation perl contradicts measured
//     nothing about our parser.
//   - perl PRINTS what the file pins, byte for byte, when it pins
//     anything.
//
// This is `verdict`'s own first phase, run for its answer rather than for
// its effect on the suite, so a CORPUS BUG here is reported as one file's
// problem and not as a refusal.
func TestTierLiteralsPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierLiterals, perl)

	for name, f := range tierFiles(t, tierLiterals) {
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

// TestTierLiteralsLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, and it exists because the tier issue's
// acceptance is checkable on its own: a tier landing with a file that
// reaches forward must fail HERE, named as this tier's problem, rather
// than as one subtest among a hundred and thirteen.
//
// What the lint can prove is one direction only: no file uses an op that
// no tier at or before 01 introduces. It cannot derive the tier -- `my $x
// = 1+2` arrives as `const[IV 3] s/FOLD` with no `add` at all -- which is
// why a file DECLARES its tier and the ops lint the declaration.
func TestTierLiteralsLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierLiterals) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierLiterals, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// TestTierLiteralsAdjacency checks the tier's adjacency file holds every
// construct the tier introduces, each next to another.
//
// WHAT OPS CANNOT DO HERE, and it is the whole design of this test.
//
// The obvious check is that the adjacency file emits the tier's declared
// INTRODUCES set. Perl refutes it. Measured 5.42.0: several `my`
// declarations in a row -- MORE constructs, more adjacent, which is
// exactly what an adjacency file is -- let the optimiser fuse them into
// one `padrange` and emit FEWER ops than the one-statement files do, the
// `pushmark` among them. The declared set is a UNION across the tier's
// files and is not a property of any single one, so demanding it of the
// adjacency file demands something perl will not produce.
//
// So the adjacency claim is checked on the SOURCE, which is where it
// actually lives: every construct spelling the tier's construct files
// introduce must also appear in the adjacency file's source, and the
// adjacency file must be a single multi-statement body rather than a
// concatenation of unrelated programs.
//
// What that ESTABLISHES: every construct this tier teaches appears in one
// compiled body, so a parser that handles each alone and mis-handles the
// pair is reachable from this corpus. That is the bug a
// one-construct-per-file corpus cannot see -- measured, `class Foo {
// ADJUST { 1 } }` parses with 0 Unknowns while `class Foo { ADJUST { 1 }
// method m { 2 } }` returns 1 Unknown swallowing both.
//
// What it does NOT establish: that the constructs are adjacent in any
// stronger sense than "in the same body". Whether `my $dec = 0.5;` sits
// next to `my $sq = 'plain';` or six statements away is not checkable
// from ops, from output, or from the tree while the parser is
// incomplete -- and a source-position check would pin a layout, which the
// file is free to change. The file's comment makes the adjacency claim in
// prose and review reads it; this test makes the prose falsifiable on the
// part that can be falsified, which is COVERAGE.
func TestTierLiteralsAdjacency(t *testing.T) {
	files := tierFiles(t, tierLiterals)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierLiterals, adjacencyFile)
	}

	// The construct each non-adjacency file introduces, taken from that
	// file's own source rather than from a list here: the tier's files
	// ARE the enumeration of what it introduces, and a second list beside
	// them is how this package has drifted four times already.
	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		spelling := literalSpelling(f.Source)
		if spelling == "" {
			t.Errorf("%s: no `my $x = <literal>;` to take a construct from", name)
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

// reLiteralBinding matches the `= <literal>;` a construct file binds.
//
// Every construct file in this tier has the same shape -- bind one
// literal, print it -- and the literal after the `=` is the construct.
// Anchored on `= ` and `;` so the captured text is the whole spelling,
// which is what makes a substring search for it in the adjacency file
// mean "this construct appears there" rather than "these characters do".
var reLiteralBinding = regexp.MustCompile(`(?m)^my \$\w+ = (.+);$`)

// literalSpelling returns the literal a construct file binds, as it is
// written, or "" when the file is not of that shape.
func literalSpelling(source string) string {
	m := reLiteralBinding.FindStringSubmatch(source)
	if m == nil {
		return ""
	}
	return "= " + m[1] + ";"
}

// TestTierLiteralsRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// A message is prose: reword it and a file that quoted it breaks over a
// change that moved nothing. A code changes only when the reason the
// parser declines changes, which is the event a refusing file wants to be
// told about -- and `run.go` already fails a file whose named code is not
// among the ones it actually produces.
//
// THE PART THAT NEEDED DECIDING. `parse.RefusalCode` names a site in the
// PARSER, and this tier holds a refusal the parser cannot see. Measured:
// `my $x = 5e-1;` lexes to `Number(5e) Operator(-) Number(1)`, which the
// parser reads as perfectly good subtraction and returns ZERO Unknowns
// for. There is no parser site to name, and naming one anyway would make
// the file fail as a stale marker -- correctly, because the claim would
// be false.
//
// So the rule is biconditional rather than "every refusing file names a
// code":
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
func TestTierLiteralsRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierLiterals) {
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
