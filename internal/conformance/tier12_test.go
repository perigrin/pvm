// ABOUTME: Tier 12 checked against the finished tooling: package, use, require, import.
// ABOUTME: Five checks the tier's own issue names, plus the one assertion a tier of no-op constructs can make.
package conformance

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierPackages is the tier this file is about.
//
// A constant rather than a literal at five call sites, for the reason
// tier01_test.go gives for tierLiterals: the tier number is a POSITION
// and positions move.
const tierPackages = "12_packages"

// TestTierPackagesPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// Identical in intent to TestTierLiteralsPerlValidated, and this tier is
// where it carries the most weight in the corpus. Three of the tier's
// four constructs emit NO RUNTIME OP AT ALL -- `package` moves the
// compiler's notion of the current package, `use` is a BEGIN block that
// finished before the optree existed, and `import` is an ordinary method
// call indistinguishable from any other. The ops therefore cannot tell
// this tier's files apart, and for `03_use_pragma.t` and
// `04_use_empty_list.t` the pinned OUTPUT is the only measurement that
// separates loading a module from importing from it.
//
// A pinned output perl does not actually produce would leave this tier
// with nothing measured but one `require`.
func TestTierPackagesPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierPackages, perl)

	for name, f := range tierFiles(t, tierPackages) {
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

// TestTierPackagesLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a file that reaches forward fails HERE,
// named as this tier's problem rather than as one subtest among a
// hundred and thirteen.
//
// The lint has a live constraint in this tier, and it is the reason
// several files are written the way they are. `05_require_bareword.t`
// and `06_require_expression.t` observe the load through `%INC` rather
// than through anything the loaded module does, because a module that
// printed or returned would drag its own ops into a file whose tier
// claims exactly one. The lint is what holds that discipline: a file
// that reached for `open` to prove the module's file exists would
// satisfy its own output pin and fail here.
func TestTierPackagesLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierPackages) {
		t.Run(name, func(t *testing.T) {
			if err := lintOps(t, f.Source, tierPackages, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// rePackageStatement matches `package NAME;` -- the statement form.
//
// The trailing `;` is what distinguishes it from the block form, and
// both spellings must be told apart because the tier has a file for
// each. Anchored at line start after optional indentation so a `package`
// inside a string or a comment is not mistaken for a declaration.
var rePackageStatement = regexp.MustCompile(`(?m)^[ \t]*package\s+[\w:]+\s*;`)

// rePackageBlock matches `package NAME {` -- the block form.
var rePackageBlock = regexp.MustCompile(`(?m)^[ \t]*package\s+[\w:]+\s*\{`)

// reUseStatement matches `use MODULE;` -- the form with no import list,
// which is the one that DOES import.
//
// It matches both `use strict;` and `use POSIX;`, which is correct: they
// are one construct, and the tier has a file for each only because a
// pragma and an exporter make the construct observable in different
// ways. `formFromPackageName` is what tells the two files apart, from
// their names, which is where that distinction belongs.
var reUseStatement = regexp.MustCompile(`(?m)^[ \t]*use\s+[\w:]+\s*;`)

// reUseEmptyList matches `use MODULE ();` -- the load-but-import-nothing
// form.
//
// Its own pattern because the empty parentheses are NOT an empty
// argument list, they are the absence of one, and that difference is the
// whole subject of 04_use_empty_list.t. A check that read both `use
// POSIX;` and `use POSIX ();` as "the use statement" would let the tier
// drop the file that separates the two halves of `use` and still pass.
var reUseEmptyList = regexp.MustCompile(`(?m)^[ \t]*use\s+[\w:]+\s*\(\s*\)\s*;`)

// reRequireBareword matches `require NAME;` where the operand is a
// bareword module name rather than an expression.
var reRequireBareword = regexp.MustCompile(`(?m)^[ \t]*require\s+[A-Za-z_][\w:]*\s*;`)

// reRequireExpression matches `require EXPR;` where the operand is not a
// bareword -- a variable, in this tier.
//
// Separate from the bareword form because the tier's claim is that the
// two compile to the SAME op and differ only in how the filename got
// there. A pattern matching both would make that claim untestable from
// the source, and the op stream cannot make it either: `require` is one
// op whichever way it was spelled.
var reRequireExpression = regexp.MustCompile(`(?m)^[ \t]*require\s+[$@%]`)

// reImportCall matches a call to `import` on a package name.
var reImportCall = regexp.MustCompile(`->\s*import\b`)

// packageForms are this tier's constructs, named as the README names
// them, each with the pattern that recognises it in a source.
//
// A slice of pairs rather than a map so failure messages come out in a
// stable order, and so the table reads as the tier's enumeration.
func packageForms() []struct {
	form string
	re   *regexp.Regexp
} {
	return []struct {
		form string
		re   *regexp.Regexp
	}{
		{"package_statement", rePackageStatement},
		{"package_block", rePackageBlock},
		// `use_pragma` and `use_import` share a pattern because they
		// share a CONSTRUCT: `use strict;` and `use POSIX;` are the same
		// statement, and the tier has a file for each only because a
		// pragma's effect is compile-time and unobservable at run time
		// while an exporter's is a symbol in `main::`. Splitting the
		// pattern would assert a syntactic difference that is not there.
		{"use_pragma", reUseStatement},
		{"use_import", reUseStatement},
		{"use_empty_list", reUseEmptyList},
		{"require_bareword", reRequireBareword},
		{"require_expression", reRequireExpression},
		{"import_call", reImportCall},
	}
}

// packageFormsIn returns the constructs a source exercises.
func packageFormsIn(source string) []string {
	var out []string
	for _, pf := range packageForms() {
		if pf.re.MatchString(source) {
			out = append(out, pf.form)
		}
	}
	return out
}

// formFromPackageName returns the construct a file's NAME declares, or
// "" when its name names none.
//
// THE NAME, not the source, for tier 05's reason. Several files here
// carry more than one of the tier's constructs: `07_import_call.t` needs
// two `package` statements to have a boundary to call across, and
// `00_adjacency.t` holds all of them. Reading each file's subject from
// its source would report `package_statement` four times and the tier
// would look like one construct with six spellings.
//
// The spec already makes the name a file's identity -- "A file's
// identity is its NAME; the number is its current position" -- so taking
// the subject from it is reading a declaration rather than inventing a
// second one. The source is then used to check the name is not lying.
func formFromPackageName(name string) string {
	m := reNumbered.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	identity := strings.TrimSuffix(m[2], ".t")
	for _, pf := range packageForms() {
		if pf.form == identity {
			return identity
		}
	}
	return ""
}

// TestTierPackagesAdjacency checks the tier's adjacency file holds every
// construct the tier introduces, each next to another, and that it pairs
// with the DECLARED prerequisite.
//
// WHAT OPS CANNOT DO HERE, and this tier is the corpus's extreme case.
// Tier 01 records that `padrange` absorbs `pushmark`, so a file with
// MORE constructs emits FEWER ops and the declared INTRODUCES set is a
// UNION across the tier's files rather than a property of any one of
// them. That reasoning applies here too, but it is not even the binding
// constraint: this tier's whole INTRODUCES set is ONE op, `require`, and
// three of its four constructs emit nothing whatsoever. An adjacency
// check keyed on ops would be satisfied by a file holding a single
// `require` and no `package` and no `use` at all.
//
// So COVERAGE is checked on the SOURCE, as at tiers 01 and 05.
//
// THE PREREQUISITE PAIRING is the half ops CAN carry, and only that
// half. The README declares `07_subroutines` -- NOT tier N-1, which
// would be `11_oo`, and the README says plainly that this tier needs
// nothing from the object system. `readTierDeps` reads the declaration
// rather than computing N-1.
//
// Reading it is not enough, and mutation is what showed that: with the
// README edited to declare `11_oo`, the pairing check stayed green,
// because this adjacency file emits `11_oo`'s ops as well. So the
// declaration is PINNED separately, and the note at that pin records
// why. Two assertions, neither implying the other: the declaration is
// the one the tier argued for, and the adjacency file honours it.
//
// Measured, the adjacency file pairs through `entersub`, which
// `07_subroutines` introduces and which arrives from `Greet->import("tag")`
// and the two fully-qualified calls. That is the right op for the claim:
// the README's reason for depending on tier 07 is that `import` is a
// plain sub call, and `entersub` is that call as perl compiled it.
//
// Why the OP and not the spelling, for this half: a source search for
// `import` is satisfied by the word inside a comment, and by a sub
// DEFINITION that nothing calls. Requiring the op requires that the two
// tiers' constructs actually met in one compiled body.
//
// What this does NOT establish: that the constructs are adjacent in any
// stronger sense than "in the same body" -- the same limit tiers 01 and
// 05 record, for the same reason.
func TestTierPackagesAdjacency(t *testing.T) {
	files := tierFiles(t, tierPackages)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierPackages, adjacencyFile)
	}
	adjacent := packageFormsIn(adj.Source)

	// The construct each non-adjacency file introduces, taken from that
	// file's own identity rather than from a list here: the tier's files
	// ARE the enumeration of what it introduces, and a second list beside
	// them is how this package has drifted four times already.
	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		form := formFromPackageName(name)
		if form == "" {
			t.Errorf("%s: its name names no construct of this tier, so nothing places it", name)
			continue
		}
		if !slices.Contains(packageFormsIn(f.Source), form) {
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
	dep, ok := deps[tierPackages]
	if !ok {
		t.Fatalf("%s declares no DEPENDS ON", tierPackages)
	}
	if dep == "nothing" {
		t.Fatalf("%s declares no prerequisite, but the issue asks its "+
			"adjacency file to pair with one", tierPackages)
	}

	// The declared prerequisite is PINNED, and this is the check the
	// rest of the pairing needs in front of it.
	//
	// Found by mutation, not by reasoning. Changing the README's DEPENDS
	// ON from `07_subroutines` to `11_oo` and re-running left the
	// pairing test GREEN, because the adjacency file emits `11_oo`'s
	// ops as well -- `method_named` from `Greet->import("tag")` and
	// `stub` from the package block. So "emits an op of the declared
	// prerequisite" is satisfied by the WRONG declaration here, and the
	// test read the declaration faithfully while asserting nothing
	// about what it said.
	//
	// That matters more in this tier than in any other, because the
	// prerequisite being `07_subroutines` and NOT tier N-1 is the
	// tier's substantive claim. The README spends a section on it: the
	// `import` that makes `use` heavier than `require` is a plain sub
	// call, and this tier needs nothing from the object system. A silent
	// drift to `11_oo` would erase the claim and every test would stay
	// green.
	//
	// So the pin is a SECOND assertion beside the pairing rather than a
	// replacement for it. The pin says the declaration is the one the
	// tier argued for; the pairing says the adjacency file honours it.
	// Either alone is satisfiable without the other.
	const declaredPrerequisite = "07_subroutines"
	if dep != declaredPrerequisite {
		t.Errorf("%s declares DEPENDS ON %s, want %s.\n"+
			"\tThis tier's README argues at length that its prerequisite "+
			"is NOT tier N-1: `import` is a plain sub call, and the tier "+
			"needs nothing from the object system. A change here is a "+
			"change to that claim and must be made deliberately.\n"+
			"\tNote that the pairing check below does NOT catch this on "+
			"its own -- 00_adjacency.t emits 11_oo's ops too.",
			tierPackages, dep, declaredPrerequisite)
	}

	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	depOps := tiers[dep]
	if len(depOps) == 0 {
		t.Fatalf("%s declares %s, which introduces no op to pair with", tierPackages, dep)
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
			tierPackages, dep, adjacencyFile, strings.Join(depOps, " "))
		return
	}
	t.Logf("%s pairs with %s through %s", adjacencyFile, dep, strings.Join(paired, ", "))
}

// TestTierPackagesRefusalsCited checks that a refusing file names WHICH
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
// That is a measurement rather than an oversight. It is also less
// surprising here than it looks: `internal/parse/use.go` already handles
// `use`, `no` and `require` as one statement form, and `decl.go` handles
// `package` in both spellings, because T2 -- the corpus these tiers were
// surveyed against -- is full of them. The tier the ops say almost
// nothing about is the tier the parser was built for first.
//
// The check is still written, because the event it guards is a file
// ACQUIRING a refusal.
func TestTierPackagesRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierPackages) {
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
