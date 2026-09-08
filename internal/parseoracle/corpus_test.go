// ABOUTME: Tests the corpus machinery — shim construction, version pinning, stderr classification.
// ABOUTME: The corpus is referenced, not vendored, so every test skips cleanly without $PERL5_CORPUS.

package parseoracle_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// corpusRoot resolves the corpus and skips the test when it is absent. perl's
// tests are Artistic/GPL, so they are referenced rather than vendored; a
// machine with no perl5 checkout must skip rather than fail.
func corpusRoot(t *testing.T) string {
	t.Helper()
	root, err := parseoracle.CorpusRoot()
	if err != nil {
		t.Skipf("corpus unavailable: %v", err)
	}
	return root
}

// TestCorpusSkipsWithoutEnv proves the suite degrades to a skip rather than a
// failure when PERL5_CORPUS names nothing usable. Someone cloning this repo
// without a perl5 checkout must still be able to run `go test ./...`.
func TestCorpusSkipsWithoutEnv(t *testing.T) {
	t.Setenv("PERL5_CORPUS", filepath.Join(t.TempDir(), "definitely-not-a-corpus"))

	if _, err := parseoracle.CorpusRoot(); err == nil {
		t.Fatal("CorpusRoot must report an error when the corpus is missing, so callers can skip")
	}
}

// TestBuildShimCompilesSubT is the shim's reason to exist. t/test.pl clears
// @INC and unshifts ../lib, so perl's tests only compile inside a tree where
// ../lib is populated. 498 of 620 corpus files use that convention.
func TestBuildShimCompilesSubT(t *testing.T) {
	root := corpusRoot(t)
	shim := t.TempDir()

	if err := parseoracle.BuildShim(root, shim); err != nil {
		t.Fatalf("BuildShim: %v", err)
	}

	cmd := exec.Command("perl", "-c", "--", "op/sub.t")
	cmd.Dir = filepath.Join(shim, "t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("op/sub.t must compile inside the shim, got %v:\n%s", err, out)
	}
}

// TestShimHasConfigPm gets its own assertion because its absence is the exact
// omission that made an earlier measurement report 66.3% instead of 94.2%.
// Config.pm lives in the architecture-specific @INC root, so a shim populated
// from only the pure-perl root silently loses ~170 files.
func TestShimHasConfigPm(t *testing.T) {
	root := corpusRoot(t)
	shim := t.TempDir()

	if err := parseoracle.BuildShim(root, shim); err != nil {
		t.Fatalf("BuildShim: %v", err)
	}

	config := filepath.Join(shim, "lib", "Config.pm")
	if _, err := os.Stat(config); err != nil {
		t.Fatalf("Config.pm missing from the shim's lib/ (%v) — the shim was "+
			"populated from only one library root; this is the omission that "+
			"cost ~170 files and 28 percentage points", err)
	}
}
