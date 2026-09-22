// ABOUTME: Regenerates the corpus refusal baseline, guarded behind an env var so it never runs by accident.
// ABOUTME: The baseline is the record of what our parser refuses, so rewriting it silently would erase the record.
package conformance

import (
	"os"
	"testing"
)

// TestRegenerateCorpusRatchet rewrites `testdata/corpus.ratchet` from the
// corpus as it stands, and does nothing at all unless
// `PVM_REGENERATE_RATCHET` is set.
//
// A regeneration step is the one thing a ratchet must never do on its
// own. The baseline exists to make a change in what our parser refuses
// FAIL, and a test that rewrote it whenever it disagreed would turn
// every regression into a silent update -- the failure mode the
// `t/` sweep's own baseline was built to prevent.
//
// So this is a test rather than a `cmd/` program for proximity -- it
// uses `corpusState` and `renderRatchet`, which are package-private and
// are the same two functions `TestCorpusRatchet` compares with, so the
// file it writes is by construction the file the check expects to read.
// A separate generator would be a second implementation of the format
// and would drift from the reader.
func TestRegenerateCorpusRatchet(t *testing.T) {
	if os.Getenv("PVM_REGENERATE_RATCHET") == "" {
		t.Skip("set PVM_REGENERATE_RATCHET=1 to rewrite the baseline")
	}

	state, err := corpusState(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus: %v", err)
	}

	keyed, err := ratchetKeys(state)
	if err != nil {
		t.Fatalf("keying the corpus: %v", err)
	}

	if err := os.WriteFile(ratchetPath, []byte(renderRatchet(keyed)), 0o644); err != nil {
		t.Fatalf("writing the baseline: %v", err)
	}
	t.Logf("wrote %s from %d files", ratchetPath, len(keyed))
}
