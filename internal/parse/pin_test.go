// ABOUTME: The T1 corpus pin is well-formed, and CI actually supplies the corpora it names.
// ABOUTME: A ratchet that skips is a guard that cannot fail, so the workflow is tested too.

package parse_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestT1PinRecordsARevision: the pin is present and names a full SHA.
//
// Only a revision, and no interpreter: the oracle's corpus.pin needs one
// because it RUNS perl to get an answer to compare against, while this ratchet
// reads the .t files as text. An interpreter cannot skew it.
func TestT1PinRecordsARevision(t *testing.T) {
	src, err := os.ReadFile("testdata/t1corpus.pin")
	if err != nil {
		t.Fatalf("the T1 baseline must record the corpus it was measured against: %v", err)
	}

	rev := pinField(string(src), "revision")
	if rev == "" {
		t.Fatal("t1corpus.pin records no revision")
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(rev) {
		t.Errorf("revision %q is not a full 40-character SHA; an abbreviated "+
			"one can become ambiguous as the corpus grows", rev)
	}
}

// TestCIRunsTheRatchets: the workflow supplies both corpora and fails on a
// skip.
//
// Asserted against the YAML as text rather than by running it. The defect this
// guards is precisely that nothing noticed a gate had stopped gating, and a
// test that only runs where the corpus already exists would repeat it.
func TestCIRunsTheRatchets(t *testing.T) {
	src, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatalf("read ci.yml: %v", err)
	}
	ci := string(src)

	for _, want := range []struct{ what, needle string }{
		{"the T1 corpus is supplied", "T1_CORPUS:"},
		{"the perl5 corpus is supplied", "PERL5_CORPUS:"},
		{"the T1 pin is read", "t1corpus.pin"},
		{"the perl5 pin is read", "corpus.pin"},
		{"the checkout is verified against the pin", "rev-parse HEAD"},
		{"the corpora are cached on the pin", "actions/cache"},
		{"a skipped ratchet fails the build", "--- SKIP: $t"},
	} {
		if !strings.Contains(ci, want.needle) {
			t.Errorf("ci.yml: %s -- expected to find %q", want.what, want.needle)
		}
	}

	// Every corpus-backed ratchet must be named in the skip check. A ratchet
	// added later and left out of the list would skip in CI unnoticed, which
	// is this issue happening again.
	for _, name := range []string{
		"TestLexerRatchet",
		"TestParseRatchet",
		"TestParsedFilesRoundTrip",
		"TestLexDotTGoldenStream",
		"TestCanonRatchet",
		"TestCanonTokenIdentity",
		"TestCanonFidelityRatchet",
	} {
		if !strings.Contains(ci, name) {
			t.Errorf("ci.yml does not check whether %s skipped", name)
		}
	}
}

// pinField reads one `key = value` line, ignoring comments.
func pinField(src, key string) string {
	for _, line := range strings.Split(src, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		k, v, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(k) == key {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
