// ABOUTME: Tests the dependency lint: a file's ops must be introduced at or before its tier.
// ABOUTME: Ops come from perl's own optree, so the check is parser-independent.
package conformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

// TestLintRejectsUndeclaredConstruct is the check that makes the tier
// numbering a claim rather than an opinion.
//
// A file placed at tier 01 that performs arithmetic is using something
// tier 04 introduces. The lint must say so, naming both the op and the
// tier that owns it, because "this file is misplaced" is useless without
// "and it belongs no earlier than here".
func TestLintRejectsUndeclaredConstruct(t *testing.T) {
	// `1+$y` rather than `1+2`: the optimiser folds a constant sum away
	// entirely, and a folded file is the subject of the next test.
	src := "my $y = 2;\nmy $x = 1+$y;\nprint \"$x\\n\";\n"

	err := lintOps(t, src, "01_literals", testTiers)
	if err == nil {
		t.Fatal("lintOps accepted a tier-01 file doing arithmetic, want a rejection")
	}
	if !strings.Contains(err.Error(), "add") {
		t.Errorf("error does not name the offending op: %v", err)
	}
	if !strings.Contains(err.Error(), "04_operators") {
		t.Errorf("error does not name the tier that introduces it: %v", err)
	}
}

// TestLintToleratesFoldedOps pins the limit of what ops can prove.
//
// `my $x = 1+2` arrives as `const[IV 3] s/FOLD` with NO add op, so a
// literals file and an arithmetic file can produce identical op lists.
// The lint must not demand ops it cannot see: it catches a file using
// what it did not declare, and cannot derive a tier on its own. That is
// why the `# TIER` header declares and the ops only lint.
func TestLintToleratesFoldedOps(t *testing.T) {
	src := "my $x = 1+2;\nprint \"$x\\n\";\n"

	if err := lintOps(t, src, "01_literals", testTiers); err != nil {
		t.Errorf("lintOps rejected a file whose arithmetic the optimiser erased: %v", err)
	}
}

// TestUnclaimedOpReported keeps the union honest.
//
// An op no tier claims must be reported rather than silently allowed or
// silently denied. Allowed-by-default lets the union grow by accident,
// which is how a dependency check stops checking; denied-by-default makes
// every new construct look like a tier violation.
func TestUnclaimedOpReported(t *testing.T) {
	// `sort` is claimed by no tier in testTiers, and the optimiser
	// cannot fold it away the way it folds sprintf("%d", 1) to a
	// constant -- which is what makes it a usable fixture here.
	src := "my @a = (3,1,2);\nmy @s = sort @a;\nprint \"@s\\n\";\n"

	err := lintOps(t, src, "01_literals", testTiers)
	if err == nil {
		t.Fatal("lintOps accepted an op no tier claims, want it reported")
	}
	if !strings.Contains(err.Error(), "sort") {
		t.Errorf("error does not name the unclaimed op: %v", err)
	}
	if !strings.Contains(err.Error(), "no tier") {
		t.Errorf("error does not distinguish unclaimed from misplaced: %v", err)
	}
}

// TestOpTableDerivedFromReadmes keeps the lint's input in one place.
//
// The ops a tier introduces could live in a Go table beside this file.
// Then the tier's contents and its declared contents would be two lists
// that must agree, and two lists that must agree are how drift starts --
// this milestone has already watched a hand-copied op list lose an entry
// on its first day. In the tier directory, a disagreement shows up in the
// same diff as the file that caused it.
//
// So the READMEs are the source and this is the reader.
//
// The assertion has to be that the ops came FROM the README, not merely
// that some ops arrived. An earlier version of this test checked that the
// map was non-empty and contained `const` -- which a hard-coded Go table
// satisfies exactly, so it passed against a readTierOps that never opened
// a file. That is the defect this AC exists to prevent, reaching the AC
// itself. The op set below is therefore deliberately unlike any real
// tier's: nothing but the README can produce it.
func TestOpTableDerivedFromReadmes(t *testing.T) {
	dir := t.TempDir()
	for tier, ops := range map[string]string{
		"01_literals":  "padsv_store const",
		"04_operators": "add subtract",
	} {
		writeTierReadme(t, dir, tier, ops)
	}

	tiers, err := readTierOps(dir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for tier, want := range map[string][]string{
		"01_literals":  {"padsv_store", "const"},
		"04_operators": {"add", "subtract"},
	} {
		got, ok := tiers[tier]
		if !ok {
			t.Errorf("%s missing; found %v", tier, keysOf(tiers))
			continue
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s = %v, want exactly %v", tier, got, want)
		}
	}

	if len(tiers) != 2 {
		t.Errorf("read %d tiers, want exactly the 2 written: %v", len(tiers), keysOf(tiers))
	}
}

// TestRealCorpusReadmesAreReadable is the companion to the test above.
//
// That one proves derivation with fixtures; this one proves the real
// corpus has READMEs the reader can actually parse. Split because a
// single test doing both cannot fail informatively: "the reader is
// hard-coded" and "a tier README is malformed" want different fixes.
func TestRealCorpusReadmesAreReadable(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	if len(tiers) == 0 {
		t.Fatal("no tier READMEs found; the lint would have nothing to check against")
	}

	ops, ok := tiers["01_literals"]
	if !ok {
		t.Fatalf("01_literals has no INTRODUCES set; found %v", keysOf(tiers))
	}

	// The ops a literals tier must own, whatever else it gains: a
	// constant, and the statement machinery every file needs.
	for _, want := range []string{"const", "nextstate"} {
		if !contains(ops, want) {
			t.Errorf("01_literals does not introduce %q; has %v", want, ops)
		}
	}
}

// writeTierReadme creates a tier directory with a minimal README.
func writeTierReadme(t *testing.T, dir, tier, ops string) {
	t.Helper()

	path := filepath.Join(dir, tier)
	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatalf("creating %s: %v", tier, err)
	}
	body := "# " + tier + "\n\n## INTRODUCES\n\n    " + ops + "\n"
	if err := os.WriteFile(filepath.Join(path, "README.md"), []byte(body), 0o600); err != nil {
		t.Fatalf("writing %s README: %v", tier, err)
	}
}

// TestReadTierOpsRejectsMalformed covers what a bad README does.
//
// The INTRODUCES block is machine-read, so a malformed one must fail
// loudly rather than yield an empty set -- an empty set would make every
// file in that tier report unclaimed ops, which reads as a corpus-wide
// failure rather than as one broken README.
func TestReadTierOpsRejectsMalformed(t *testing.T) {
	dir := t.TempDir()
	tier := filepath.Join(dir, "01_literals")
	if err := os.MkdirAll(tier, 0o750); err != nil {
		t.Fatalf("creating the tier directory: %v", err)
	}
	readme := filepath.Join(tier, "README.md")
	if err := os.WriteFile(readme, []byte("# 01_literals\n\nNo introduces block.\n"), 0o600); err != nil {
		t.Fatalf("writing the README: %v", err)
	}

	_, err := readTierOps(dir)
	if err == nil {
		t.Fatal("readTierOps accepted a README with no INTRODUCES block, want an error")
	}
	if !strings.Contains(err.Error(), "INTRODUCES") {
		t.Errorf("error does not say what is missing: %v", err)
	}
}

// TestCorpusLints runs the lint over the real corpus.
//
// The three tests above prove the lint works on fixtures; this is the one
// that makes it a gate. A lint nothing runs is a lint that rots, and the
// claim "every file's constructs were introduced at or before its tier"
// is only true if something checks it on every commit.
//
// It runs in `go test` rather than as a script someone remembers, for the
// same reason.
func TestCorpusLints(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	// A directory the tier reader cannot see contributes no subtests, and
	// a skip is otherwise indistinguishable from an absence: a
	// renumbering typo like `9_regex` would disable the gate for a whole
	// tier while the suite stayed green. This is the check that a lint
	// nothing runs is a lint that rots, applied to the lint itself.
	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus root: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() || isTierDir(e.Name()) {
			continue
		}
		cases, err := filepath.Glob(filepath.Join(corpusDir, e.Name(), "*.t"))
		if err != nil {
			t.Fatalf("globbing %s: %v", e.Name(), err)
		}
		if len(cases) > 0 {
			t.Errorf("%s holds %d case(s) but is not a numbered tier, so nothing lints them",
				e.Name(), len(cases))
		}
	}

	// Every op any file emits, so a README claiming something no file
	// uses can be reported below.
	emitted := map[string]bool{}

	for _, tier := range keysOf(tiers) {
		paths, err := filepath.Glob(filepath.Join(corpusDir, tier, "*.t"))
		if err != nil {
			t.Fatalf("globbing %s: %v", tier, err)
		}
		for _, path := range paths {
			t.Run(filepath.Join(tier, filepath.Base(path)), func(t *testing.T) {
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("reading the case: %v", err)
				}
				f, err := ParseFile(string(raw))
				if err != nil {
					t.Fatalf("parsing sections: %v", err)
				}
				if err := lintOps(t, f.Source, tier, tiers); err != nil {
					t.Errorf("%s", err)
				}

				ops, err := opsOf(t, f.Source)
				if err != nil {
					t.Fatalf("collecting ops: %v", err)
				}
				for _, op := range ops {
					emitted[op] = true
				}
			})
		}
	}

	// The lint is one-directional: it catches a file using an op its
	// tier does not claim, and nothing catches a tier CLAIMING an op no
	// file uses. An op added to a README "just in case" permanently
	// widens that tier and every tier after it, invisibly -- which is
	// the drift this whole arrangement exists to stop.
	//
	// Checked against the corpus-wide union rather than per tier,
	// because tier 01 claims the statement machinery (`enter`, `leave`,
	// `nextstate`, `pushmark`) that every later tier also emits.
	for _, tier := range keysOf(tiers) {
		for _, op := range tiers[tier] {
			if !emitted[op] {
				t.Errorf("%s/README.md claims %q, which no corpus file emits", tier, op)
			}
		}
	}
}

// TestPerlIsTheMeasuredVersion fails loudly on the wrong interpreter.
//
// Every corpus file's header says MEASURED perl 5.42.0, and both
// `opsOf` here and `askPerl` in run.go invoke bare `perl` from PATH. On a
// different perl the suite does not break -- it quietly measures a
// different language and reports the disagreement as CORPUS BUG, blaming
// the file rather than the environment.
//
// This is the guard, not the fix. Resolving perl through the pinned
// 5.42.0 belongs to the runner issue, which owns both call sites; until
// then a wrong version is at least named as such.
func TestPerlIsTheMeasuredVersion(t *testing.T) {
	out, err := exec.Command("perl", "-e", "print $]").Output()
	if err != nil {
		t.Fatalf("running perl: %v", err)
	}
	// $] is the numeric form: 5.042000 is 5.42.0.
	if got, want := string(out), "5.042000"; got != want {
		t.Fatalf("perl on PATH reports $] = %s, want %s\n"+
			"\tthe corpus records MEASURED perl 5.42.0, and every expectation "+
			"in it was taken from that interpreter", got, want)
	}
}

// TestTierOrderIsLexicographic pins the property `reachable` depends on.
//
// It decides "at or before this tier" with a string comparison, which is
// correct only while every tier name carries a zero-padded two-digit
// prefix. Unpadded, "10_io" would sort before "9_regex" and the whole
// ordering claim would invert at that boundary without any test noticing.
//
// The spec's "two digits, not four, and no gaps" decision is what makes
// this hold; this is where that decision is enforced rather than assumed.
func TestTierOrderIsLexicographic(t *testing.T) {
	sorted := append([]string(nil), specTiers...)
	sort.Strings(sorted)
	for i := range specTiers {
		if specTiers[i] != sorted[i] {
			t.Fatalf("lexicographic order diverges from tier order at %d: %q vs %q",
				i, sorted[i], specTiers[i])
		}
	}

	// The boundary that would break first under an unpadded scheme.
	if !("09_regex" <= "10_io") {
		t.Error(`"09_regex" does not sort before "10_io"`)
	}
}

// TestReadTierOpsRejectsProse guards the lint's own input.
//
// The INTRODUCES block is read with strings.Fields, so prose inside it
// would become ops: "These are the ops: const and nextstate." yields
// seven entries including "These" and "nextstate." with a trailing stop.
//
// That failure is silent and it widens the allowed set, which is the
// worst direction -- a misplaced file would then pass its lint. It is the
// "union grows by accident" case this lint exists to prevent, reaching
// the lint through its own configuration.
func TestReadTierOpsRejectsProse(t *testing.T) {
	dir := t.TempDir()
	tier := filepath.Join(dir, "01_literals")
	if err := os.MkdirAll(tier, 0o750); err != nil {
		t.Fatalf("creating the tier directory: %v", err)
	}
	readme := filepath.Join(tier, "README.md")
	// Indented, so it IS the block rather than prose sitting outside one.
	// The unindented case is caught by the regex and reported as a missing
	// block; this is the case only the name check can reach, and the one
	// that would otherwise widen the tier by seven entries.
	body := "# 01_literals\n\n## INTRODUCES\n\n    These are the ops: const and nextstate.\n"
	if err := os.WriteFile(readme, []byte(body), 0o600); err != nil {
		t.Fatalf("writing the README: %v", err)
	}

	_, err := readTierOps(dir)
	if err == nil {
		t.Fatal("readTierOps accepted prose in the INTRODUCES block, want an error")
	}
	if !strings.Contains(err.Error(), "These") {
		t.Errorf("error does not name what it rejected: %v", err)
	}
}

// TestReadTierOpsRejectsUnindentedBlock is the companion case.
//
// Prose that is NOT indented is not a malformed block, it is the absence
// of one -- and must be reported that way, since the fixes differ: indent
// what you meant as the block, versus delete prose from inside it.
func TestReadTierOpsRejectsUnindentedBlock(t *testing.T) {
	dir := t.TempDir()
	tier := filepath.Join(dir, "01_literals")
	if err := os.MkdirAll(tier, 0o750); err != nil {
		t.Fatalf("creating the tier directory: %v", err)
	}
	body := "# 01_literals\n\n## INTRODUCES\n\nThese are the ops: const and nextstate.\n"
	if err := os.WriteFile(filepath.Join(tier, "README.md"), []byte(body), 0o600); err != nil {
		t.Fatalf("writing the README: %v", err)
	}

	_, err := readTierOps(dir)
	if err == nil {
		t.Fatal("readTierOps accepted an unindented INTRODUCES section, want an error")
	}
	if !strings.Contains(err.Error(), "INTRODUCES") {
		t.Errorf("error does not say what is missing: %v", err)
	}
}

func keysOf(m tierOps) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// testTiers is a deliberately small op-to-tier map, so these tests state
// what the lint DOES rather than depending on the corpus's real tier
// READMEs, which grow. TestOpTableDerivedFromReadmes covers the real one.
//
// `multiconcat` rather than `concat` is not a typo. The lint reads ops
// AFTER the peephole optimiser, and `print "$x\n"` compiles to a single
// multiconcat rather than to the concat the source implies -- measured,
// `my $x = 5; print "$x\n"` gives
//
//	enter nextstate const padsv_store nextstate pushmark padsv
//	multiconcat print leave
//
// so a tier README must claim the ops perl EMITS, not the ones the
// source appears to call for. Writing this table is where that rule
// first bites, and it will bite every tier README the same way.
var testTiers = tierOps{
	"01_literals":  {"enter", "nextstate", "const", "padsv_store", "leave", "padsv", "print", "pushmark", "multiconcat", "gv", "readline", "stringify"},
	"04_operators": {"add", "subtract", "multiply", "divide"},
}

// specTiers is the tier list from the spec, in dependency order.
//
// Written here rather than derived from the corpus directory. A list read
// from the thing it checks cannot report an absence: a missing tier would
// remove both the directory and the expectation, and every check over it
// would pass over the gap it exists to catch.
//
// Fourteen tiers, zero-padded two digits, no gaps. The padding is what
// makes `reachable`'s string comparison equal a numeric one, which
// TestTierOrderIsLexicographic pins against this same list.
var specTiers = []string{
	"01_literals",
	"02_variables",
	"03_context",
	"04_operators",
	"05_scoping",
	"06_control",
	"07_subroutines",
	"08_references",
	"09_regex",
	"10_io",
	"11_oo",
	"12_packages",
	"13_opaque",
	"14_recursive",
}

// TestTierNecessityCheck gives check 2 teeth.
//
// The ordering claim is that a tier sits where it does because the next
// cannot proceed without it. Until something verifies the declared
// prerequisite, that claim is prose. A prerequisite at a LATER tier number
// is the case that makes the ordering incoherent -- the tier is in the
// wrong place, or the prerequisite is -- so it must fail.
func TestTierNecessityCheck(t *testing.T) {
	for _, tc := range []struct {
		name string
		deps map[string]string
		want string
	}{{
		name: "a later prerequisite is rejected",
		deps: map[string]string{"04_operators": "09_regex", "09_regex": "nothing"},
		want: "not earlier",
	}, {
		name: "a tier depending on itself is rejected",
		deps: map[string]string{"04_operators": "04_operators"},
		want: "not earlier",
	}, {
		name: "a prerequisite no tier provides is rejected",
		deps: map[string]string{"04_operators": "03_context"},
		want: "not a tier in this corpus",
	}, {
		name: "a prerequisite that is not a tier name is rejected",
		deps: map[string]string{"04_operators": "literals"},
		want: "not a tier name",
	}, {
		name: "an earlier prerequisite is accepted",
		deps: map[string]string{"01_literals": "nothing", "04_operators": "01_literals"},
		want: "",
	}, {
		name: "a non-adjacent earlier prerequisite is accepted",
		deps: map[string]string{
			"01_literals": "nothing",
			"08_refs":     "01_literals",
			"09_regex":    "01_literals",
		},
		want: "",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkTierNecessity(tc.deps)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("checkTierNecessity rejected a valid ordering: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("checkTierNecessity accepted %v, want an error saying %q", tc.deps, tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not say %q", err, tc.want)
			}
		})
	}
}

// TestRealCorpusNecessityHolds runs the check over the corpus itself.
//
// The table above proves the check works; this is what makes it a gate.
func TestRealCorpusNecessityHolds(t *testing.T) {
	deps, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier prerequisites: %v", err)
	}
	if len(deps) == 0 {
		t.Fatal("no tier prerequisites found; the check would have nothing to verify")
	}
	if err := checkTierNecessity(deps); err != nil {
		t.Errorf("the corpus ordering is incoherent: %v", err)
	}
}

// TestEveryTierHasAnAdjacencyFile checks the generic rule, not one tier's.
//
// TestReadTierOpsIgnoresTrailingProse pins which part of the section is
// machine-read.
//
// Both machine-read blocks are indented code blocks, and the FIRST such
// block is the value -- not everything up to the next `## ` heading.
// Reading to the next heading makes any explanatory paragraph after the
// block part of the value, which for INTRODUCES silently WIDENS the tier:
// a misplaced file then passes its own lint against ops that are really
// English words.
//
// Measured, not hypothetical. The first DEPENDS ON regex had exactly this
// shape and turned tier 01's closing paragraph into thirty-odd declared
// prerequisites.
func TestReadTierOpsIgnoresTrailingProse(t *testing.T) {
	dir := t.TempDir()
	tier := filepath.Join(dir, "01_literals")
	if err := os.MkdirAll(tier, 0o750); err != nil {
		t.Fatalf("creating the tier directory: %v", err)
	}

	body := "# 01_literals\n\n" +
		"## INTRODUCES\n\n    const nextstate\n\n" +
		"Those are the ops. This sentence is prose and must not be read.\n\n" +
		"## DEPENDS ON\n\n    nothing\n\n" +
		"This sentence is prose too.\n"
	if err := os.WriteFile(filepath.Join(tier, "README.md"), []byte(body), 0o600); err != nil {
		t.Fatalf("writing the README: %v", err)
	}

	tiers, err := readTierOps(dir)
	if err != nil {
		t.Fatalf("readTierOps rejected prose after the block: %v", err)
	}
	if want := []string{"const", "nextstate"}; !slices.Equal(tiers["01_literals"], want) {
		t.Errorf("INTRODUCES = %v, want exactly %v", tiers["01_literals"], want)
	}

	deps, err := readTierDeps(dir)
	if err != nil {
		t.Fatalf("readTierDeps rejected prose after the block: %v", err)
	}
	if deps["01_literals"] != "nothing" {
		t.Errorf("DEPENDS ON = %q, want %q", deps["01_literals"], "nothing")
	}
}
