// ABOUTME: Skips a test that runs the perl oracle on Windows, where it cannot run.
// ABOUTME: parse_facts.pl re-runs perl through `sh -c` with a single-quoted path.

package parseoracle_test

import (
	"runtime"
	"testing"
)

// requirePOSIXOracle skips the calling test on Windows. The oracle is
// POSIX-only by construction -- parse_facts.pl compiles each file through
// `sh -c perl '...'` -- and its gate, the fidelity sweep, runs on Linux.
func requirePOSIXOracle(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the perl oracle is POSIX-only: parse_facts.pl runs perl through sh -c")
	}
}
