// ABOUTME: Captures the legacy tree-sitter Analyze's report for each testdata/lowered fixture into its .golden.
// ABOUTME: Runs only under -infer.capture, and writes a golden only where none exists yet.

package infer_test

import (
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parser"
)

var captureLegacy = flag.Bool("infer.capture", false,
	"write a .golden from the legacy path for each testdata/lowered fixture that has none")

// TestCaptureLegacyGolden is the one place the golden ratchet calls the
// legacy Analyze. It goes when the legacy path does; the goldens it wrote
// stay as the guard.
//
// An existing golden is left alone: it may carry a deliberate edit, made
// under a Diagnostics-changed trailer, that a recapture would undo. To
// recapture one, delete it first.
func TestCaptureLegacyGolden(t *testing.T) {
	if !*captureLegacy {
		t.Skip("run with -infer.capture to write missing goldens from the legacy path")
	}
	for _, fx := range loweredFixtures(t) {
		golden := goldenPath(fx)
		if _, err := os.Stat(golden); !errors.Is(err, fs.ErrNotExist) {
			continue
		}
		src, err := os.ReadFile(filepath.Join(loweredDir, filepath.FromSlash(fx)))
		if err != nil {
			t.Fatalf("%s: %v", fx, err)
		}
		tree, err := parser.New().Parse(src)
		if err != nil {
			t.Fatalf("%s: legacy parse: %v", fx, err)
		}
		_, diags, st := infer.Analyze(tree, src, nil)
		if err := os.WriteFile(golden, []byte(goldenReport(diags, st)), 0o644); err != nil {
			t.Fatalf("%s: %v", golden, err)
		}
		t.Logf("captured %s", golden)
	}
}
