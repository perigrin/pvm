// ABOUTME: Proves the croak extraction splits by perl -c rather than by filename.
// ABOUTME: The directory is named croak; 40% of its cases compile fine.
package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRuntimeCroakIsNotANegative pins the whole point of the issue: a case
// perl COMPILES is not a must-not-parse case, whatever directory it lives in.
//
// The check is anchored on `pp_ctl`, `pp_hot`, `pp` and `pp_sys` -- files
// measured at ZERO compile failures. Every case in them croaks at run time,
// from `goto` into a foreach or `pipe()` on a bad left side: diagnostics
// from RUNNING, not from parsing. A runner that took the must-fail bit from
// the directory name would emit `parsent` for all of them and mark a
// correct parser wrong.
func TestRuntimeCroakIsNotANegative(t *testing.T) {
	cases := requireCroakCorpus(t)
	split := measureCroakSplit(t, cases)

	// Every runtime-only case must classify as a positive, never a negative.
	for _, c := range split.RuntimeOnly {
		if c.Verdict != CroakRuntimeOnly {
			t.Fatalf("%s: verdict %v, want %v", c.Ref(), c.Verdict, CroakRuntimeOnly)
		}
		if c.IsNegative() {
			t.Errorf("%s compiles under perl -c but was emitted as a "+
				"must-not-parse negative", c.Ref())
		}
	}

	// The runtime-only set must not be empty, or the loop above proves
	// nothing. These four files are wholly runtime croaks.
	runtimeWhole := map[string]bool{"pp": true, "pp_ctl": true, "pp_hot": true, "pp_sys": true}
	seen := map[string]int{}
	for _, c := range split.Cases {
		if !runtimeWhole[c.File] {
			continue
		}
		if c.Verdict == CroakCompileFail {
			t.Errorf("%s classified compile-fail; %s was measured at zero "+
				"compile failures, so either perl changed or the split is "+
				"reading the wrong thing", c.Ref(), c.File)
		}
		seen[c.File]++
	}
	for f := range runtimeWhole {
		if seen[f] == 0 {
			t.Errorf("no cases extracted from %s, so its zero-compile-failure "+
				"claim is untested", f)
		}
	}
	if len(split.RuntimeOnly) == 0 {
		t.Fatal("no runtime-only cases at all; the split found nothing to protect")
	}
}

// TestCroakSplitIsMeasured pins that the three-way split comes from RUNNING
// perl, not from a table someone typed in.
//
// A checked-in list goes stale silently: perl gains a diagnostic, a case
// moves from runtime to compile time, and the list keeps reporting the old
// answer forever. So this asserts the split is a function of the
// interpreter -- proven by feeding the classifier known source whose
// verdict is not in any corpus file.
func TestCroakSplitIsMeasured(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Skipf("no pinned perl: %v", err)
	}

	// Three synthetic cases, each with an unambiguous verdict, none of
	// which appears in t/lib/croak/. A classifier reading a checked-in
	// list could not answer these at all.
	for _, tc := range []struct {
		name string
		src  string
		want CroakVerdict
	}{
		{"syntax error", "my $x = ;\n", CroakCompileFail},
		{"dies at run time", "die \"boom\";\n", CroakRuntimeOnly},
		{"exits clean", "my $x = 1;\n", CroakExitsZero},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := classifyCroakCase(perl, tc.src)
			if err != nil {
				t.Fatalf("classify: %v", err)
			}
			if got != tc.want {
				t.Errorf("%q classified %v, want %v", tc.src, got, tc.want)
			}
		})
	}

	// And the corpus split itself is computed through that same
	// classifier, so every case carries a verdict.
	cases := requireCroakCorpus(t)
	split := measureCroakSplit(t, cases)
	if n := len(split.CompileFail) + len(split.RuntimeOnly) + len(split.ExitsZero); n != len(split.Cases) {
		t.Errorf("the three buckets hold %d cases, but %d were classified",
			n, len(split.Cases))
	}
	for _, c := range split.Cases {
		if c.Verdict == CroakUnclassified {
			t.Errorf("%s carries no verdict, so it was never measured", c.Ref())
		}
	}
}

// TestCroakCountsReported pins that drift from the recorded 201/64/75 is
// REPORTED, not absorbed.
//
// The numbers in the spec were measured once. If perl changes, or the
// extraction changes, the counts move -- and a runner that silently
// rebalances its buckets would hide exactly the fact that matters. So the
// recorded split is compared against the measured one and any difference
// is logged with its per-file detail.
func TestCroakCountsReported(t *testing.T) {
	cases := requireCroakCorpus(t)
	split := measureCroakSplit(t, cases)

	report := split.DriftReport()
	t.Log("\n" + report)

	if report == "" {
		t.Fatal("DriftReport produced nothing; drift cannot be reported by an empty string")
	}

	// The report must name the recorded figures, or a reader cannot tell
	// what the measurement drifted FROM.
	for _, want := range []string{"201", "64", "75"} {
		if !strings.Contains(report, want) {
			t.Errorf("report does not name the recorded count %s:\n%s", want, report)
		}
	}

	// And it must name the measured ones.
	for _, want := range []string{
		itoa(len(split.CompileFail)),
		itoa(len(split.RuntimeOnly)),
		itoa(len(split.ExitsZero)),
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report does not name the measured count %s:\n%s", want, report)
		}
	}

	// A drifted measurement must SAY so rather than read as agreement.
	drifted := len(split.CompileFail) != recordedCompileFail ||
		len(split.RuntimeOnly) != recordedRuntimeOnly ||
		len(split.ExitsZero) != recordedExitsZero
	if drifted && !strings.Contains(report, "DRIFT") {
		t.Errorf("measured %d/%d/%d against recorded %d/%d/%d, but the "+
			"report does not say DRIFT:\n%s",
			len(split.CompileFail), len(split.RuntimeOnly), len(split.ExitsZero),
			recordedCompileFail, recordedRuntimeOnly, recordedExitsZero, report)
	}
	if !drifted && strings.Contains(report, "DRIFT") {
		t.Errorf("counts match the record but the report claims drift:\n%s", report)
	}
}

// TestCroakExtractionCitesRatherThanCopies pins that the corpus is
// REFERENCED, never vendored.
//
// perl is licensed separately from this repo, and a copied corpus is a
// copy that goes stale the moment blead moves. So every extracted case
// carries a citation -- source filename plus the pinned revision of the
// checkout it came from -- and the repo itself holds none of the content.
func TestCroakExtractionCitesRatherThanCopies(t *testing.T) {
	t.Run("cases cite file and revision", func(t *testing.T) {
		cases := requireCroakCorpus(t)
		if len(cases) == 0 {
			t.Fatal("no cases extracted")
		}

		rev, err := croakRevision(croakCorpusDir())
		if err != nil {
			t.Fatalf("resolving the pinned revision: %v", err)
		}
		// A revision is a full 40-hex sha, not a branch name or a guess.
		if len(rev) != 40 {
			t.Errorf("revision %q is %d characters, want a 40-character sha",
				rev, len(rev))
		}
		for _, r := range rev {
			if !strings.ContainsRune("0123456789abcdef", r) {
				t.Fatalf("revision %q is not hex", rev)
			}
		}

		for _, c := range cases {
			cite := c.Citation(rev)
			if !strings.Contains(cite, c.File) {
				t.Errorf("citation %q does not name the source file %q", cite, c.File)
			}
			if !strings.Contains(cite, rev) {
				t.Errorf("citation %q does not name the pinned revision", cite)
			}
			// Case identity is (file, ordinal), since 87 cases have no
			// `# NAME`. A citation without the ordinal cannot resolve.
			if !strings.Contains(cite, itoa(c.Ordinal)) {
				t.Errorf("citation %q does not name the ordinal %d", cite, c.Ordinal)
			}
		}
	})

	// The honest form of "copies no file": grep the repo.
	t.Run("no croak file is vendored", func(t *testing.T) {
		// Distinctive strings from the corpus. Each appears in
		// t/lib/croak/ and nowhere a hand-written file would put it, so
		// finding one in the repo means content was copied in.
		needles := []string{
			"Cannot invoke method \"m\" on a non-instance",
			"Can't \"goto\" into a binary or list expression",
			"Type of arg 1 to main::f must be array",
		}

		root, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			t.Fatal(err)
		}
		err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // unreadable is not vendored
			}
			if d.IsDir() {
				switch d.Name() {
				case ".git", "vendor", "node_modules":
					return filepath.SkipDir
				}
				return nil
			}
			// This test file names the needles itself, and the spec
			// quotes the corpus by design. Neither is a vendored copy.
			if filepath.Base(path) == "croak_test.go" {
				return nil
			}
			if strings.HasSuffix(path, ".md") {
				return nil
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			for _, n := range needles {
				if strings.Contains(string(src), n) {
					t.Errorf("%s contains croak corpus content %q; the "+
						"corpus must be cited, not copied", path, n)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		// And no directory named after the corpus was checked in.
		if _, err := os.Stat(filepath.Join(root, "internal", "conformance", "croak")); err == nil {
			t.Error("internal/conformance/croak/ exists; the corpus was vendored")
		}
	})
}

// requireCroakCorpus extracts the corpus, or skips with a named reason.
//
// A test that silently passes when it cannot run is worse than no test:
// it reports green about a measurement that never happened.
func requireCroakCorpus(t *testing.T) []CroakCase {
	t.Helper()
	dir := croakCorpusDir()
	cases, err := ExtractCroakCases(dir)
	if err != nil {
		t.Skipf("croak corpus unavailable at %s: %v\n"+
			"\tset %s to a perl checkout to measure the split", dir, err, croakCorpusEnv)
	}
	return cases
}

// measureCroakSplit runs the classification, or skips if perl is absent.
func measureCroakSplit(t *testing.T, cases []CroakCase) *CroakSplit {
	t.Helper()
	split, err := MeasureCroakSplit(cases)
	if err != nil {
		t.Skipf("cannot measure the croak split: %v", err)
	}
	return split
}
