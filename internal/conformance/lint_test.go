// ABOUTME: Tests the dependency lint: a file's ops must be introduced at or before its tier.
// ABOUTME: Ops come from perl's own optree, so the check is parser-independent.
package conformance

import (
	"os"
	"path/filepath"
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
func TestOpTableDerivedFromReadmes(t *testing.T) {
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
	if len(ops) == 0 {
		t.Error("01_literals declares no ops, which cannot be right for a tier with files in it")
	}

	// The ops a literals tier must own, whatever else it gains: a
	// constant, and the statement machinery every file needs.
	for _, want := range []string{"const", "nextstate"} {
		if !contains(ops, want) {
			t.Errorf("01_literals does not introduce %q; has %v", want, ops)
		}
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
			})
		}
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
