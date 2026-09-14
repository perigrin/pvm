// ABOUTME: Pins perl's threads build flag, which changes the parse and which $] does not record.
// ABOUTME: Proven by the first real CI run: same $], different useithreads, three files disagreed.

package parseoracle_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestPinCatchesThreadsSkew is the defect the first GitHub Actions run of the
// fidelity gate exposed, and it is a hole in the pin rather than in the
// parser.
//
// The runner installed perl 5.042000, matching the pin exactly, and the
// interpreter check passed. But that perl was built `useithreads=undef` while
// the baseline was measured on `useithreads=define`, and three corpus files
// call `skip_all` inside a BEGIN block when threads are absent:
//
//	class/threads.t: exact -> no-answer
//	op/threads-dirh.t: exact -> no-answer
//	op/threads.t: exact -> no-answer
//
// Perl exits DURING compilation there, so it reports ok:1 with no optree and
// the B5 rule correctly declines. The verdicts are right; the pin let two
// perls that parse the corpus differently both satisfy it.
//
// $] does not carry the build configuration, so pinning it is not enough.
// This is the same class as every other defect this harness has found -- an
// environmental fact recorded as a parser fact -- one level up, in the thing
// that is supposed to make measurements comparable.
func TestPinCatchesThreadsSkew(t *testing.T) {
	pinned := parseoracle.Pin{
		Interpreter: "5.042000",
		Revision:    "94e5086608bcf1d01cd2d836f11ca8378a3320b6",
		Threads:     true,
	}

	// The runner's perl: same version, same corpus, different build.
	err := pinned.Check("5.042000", "94e5086608bcf1d01cd2d836f11ca8378a3320b6", false)
	if err == nil {
		t.Fatal("Check accepted a non-threaded perl against a threaded baseline: " +
			"three corpus files parse differently across that skew, and the pin " +
			"exists precisely so a measurement is not taken across one")
	}
	if !strings.Contains(err.Error(), "threads") {
		t.Errorf("the mismatch must name threads, got: %v", err)
	}

	// The matching build still passes, or the pin would reject every run.
	if err := pinned.Check("5.042000", "94e5086608bcf1d01cd2d836f11ca8378a3320b6", true); err != nil {
		t.Errorf("a perl matching every pinned field must verify: %v", err)
	}
}

// TestPinRoundTripsThreads: the flag has to survive the file, or CI reads a
// pin that silently drops the half this test exists to add.
func TestPinRoundTripsThreads(t *testing.T) {
	for _, want := range []bool{true, false} {
		dir := t.TempDir()
		path := dir + "/corpus.pin"
		if err := parseoracle.WritePin(path, parseoracle.Pin{
			Interpreter: "5.042000",
			Revision:    "94e5086608bcf1d01cd2d836f11ca8378a3320b6",
			Threads:     want,
		}); err != nil {
			t.Fatalf("WritePin: %v", err)
		}
		got, err := parseoracle.ReadPin(path)
		if err != nil {
			t.Fatalf("ReadPin: %v", err)
		}
		if got.Threads != want {
			t.Errorf("Threads round-tripped as %v, want %v", got.Threads, want)
		}
	}
}
