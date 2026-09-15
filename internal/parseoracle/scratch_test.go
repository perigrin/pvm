// ABOUTME: Pins that the oracle's scratch file never lands inside the corpus tree, killed or not.
// ABOUTME: A timeout kills the oracle with SIGKILL, so nothing it promised to clean up ever runs.

package parseoracle_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestScratchFileNeverLandsInTheCorpus is the leak as it happened: twelve
// `<file>.t.oracle.<pid>.pl` files inside /tmp/oracletree/t, every one a
// timeout victim (pat_psycho.t's watchdog, taint.t against the 60s default).
// The runner cancels with SIGKILL, so an unlink after the probe, an END block
// or a signal handler all run never. The only fix that survives being killed
// is not writing the file there in the first place.
//
// The fixture hangs ONLY in the probe run: the Concise run loads O.pm and the
// probe run does not, so the sleep is skipped where the harness needs it to
// be and taken exactly where capture_with_end is mid-flight when the timeout
// lands.
func TestScratchFileNeverLandsInTheCorpus(t *testing.T) {
	dir := t.TempDir()
	src := "BEGIN { sleep 10 unless $INC{'O.pm'} }\nmy $x = 1;\n"
	if err := os.WriteFile(filepath.Join(dir, "hang.pl"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := parseoracle.AskFile(context.Background(), "hang.pl",
		parseoracle.Options{Dir: dir, Timeout: 2 * time.Second})
	if err == nil {
		t.Fatal("the fixture must time out inside the probe run, or this test proves nothing")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "hang.pl" {
			t.Errorf("the oracle left %q beside the corpus file: a stray .pl inside the "+
				"corpus tree is a measurement artifact created by a timeout", e.Name())
		}
	}
}
