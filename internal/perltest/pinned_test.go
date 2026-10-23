// ABOUTME: Tests that PinnedPerl finds the perl corpus.pin names: pvm's versions directory first, then the registry, then PATH.
// ABOUTME: A candidate must match the pin's $] and useithreads both; with none, the test skips, or fails under PVM_REQUIRE_PINNED_PERL.

package perltest

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pinnedStandIn writes a shell script at dir/perl that answers the two
// questions PinnedPerl asks a candidate: what it is ($] and useithreads,
// as parseoracle.ObservePin asks) and, for the privlibexp probe, where
// its library is. threads is "define" or "" as $Config{useithreads}
// reports it.
func pinnedStandIn(t *testing.T, dir, version, threads string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	lib := filepath.Join(t.TempDir(), "privlib")
	require.NoError(t, os.MkdirAll(lib, 0o755))
	path := filepath.Join(dir, "perl")
	script := "#!/bin/sh\ncase \"$*\" in\n" +
		"*privlibexp*) printf '%s\\n%s\\n' '" + lib + "' '" + lib + "/arch' ;;\n" +
		"*) printf '%s\\n%s\\n' '" + version + "' '" + threads + "' ;;\n" +
		"esac\n"
	require.NoError(t, os.WriteFile(path, []byte(script), 0o755))
	return path
}

// pinFile writes a corpus.pin naming interpreter and threads.
func pinFile(t *testing.T, interpreter, threads string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "corpus.pin")
	body := "interpreter = " + interpreter + "\nrevision    = deadbeef\nthreads     = " + threads + "\n"
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

// pinnedWorld isolates PinnedPerl's three places to look: an empty
// DataHome, a registry that knows no perl, and a PATH with nothing on
// it. It returns the DataHome, under which a test plants pvm/versions.
func pinnedWorld(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in perl is a shell script")
	}
	withNoPvmPerl(t)
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	t.Setenv("PATH", t.TempDir())
	t.Setenv(requirePinnedEnv, "")
	return data
}

// TestPinnedPerlPrefersVersionsDir: with a matching perl in pvm's versions
// directory and a non-matching one first on PATH, PinnedPerl returns the
// versions-dir perl. PATH is what every oracle command used to trust,
// and on this machine it names 5.42 while the pin may name 5.44.
func TestPinnedPerlPrefersVersionsDir(t *testing.T) {
	data := pinnedWorld(t)
	want := pinnedStandIn(t, filepath.Join(data, "pvm", "versions", "5.44.0", "bin"), "5.044000", "define")
	decoy := t.TempDir()
	pinnedStandIn(t, decoy, "5.042000", "define")
	t.Setenv("PATH", decoy)

	p := PinnedPerlAt(t, pinFile(t, "5.044000", "define"))
	assert.Equal(t, want, p.Path)
}

// TestPinnedPerlFallsBackToPath: with nothing in pvm, a PATH perl that
// matches the pin is the one CI's actions-setup-perl installed, and it is
// used.
func TestPinnedPerlFallsBackToPath(t *testing.T) {
	pinnedWorld(t)
	dir := t.TempDir()
	want := pinnedStandIn(t, dir, "5.044000", "define")
	t.Setenv("PATH", dir)

	assert.Equal(t, want, PinnedPerlAt(t, pinFile(t, "5.044000", "define")).Path)
}

// TestPinnedPerlSkipsOnMismatch: no candidate matches and
// PVM_REQUIRE_PINNED_PERL is unset, so the test skips rather than running
// the wrong perl.
func TestPinnedPerlSkipsOnMismatch(t *testing.T) {
	data := pinnedWorld(t)
	pinnedStandIn(t, filepath.Join(data, "pvm", "versions", "5.44.0", "bin"), "5.042000", "define")
	dir := t.TempDir()
	pinnedStandIn(t, dir, "5.038002", "define")
	t.Setenv("PATH", dir)
	pin := pinFile(t, "5.044000", "define")

	assert.True(t, skipped(t, func(st testing.TB) { PinnedPerlAt(st, pin) }),
		"no perl matching the pin skips the test")
}

// TestPinnedPerlChecksThreads: a candidate with the pinned $] but
// useithreads undef is a different build, and three corpus files parse
// differently across that skew. It is passed over for the threaded perl
// later in the order.
func TestPinnedPerlChecksThreads(t *testing.T) {
	data := pinnedWorld(t)
	pinnedStandIn(t, filepath.Join(data, "pvm", "versions", "5.44.0", "bin"), "5.044000", "")
	dir := t.TempDir()
	want := pinnedStandIn(t, dir, "5.044000", "define")
	t.Setenv("PATH", dir)
	pin := pinFile(t, "5.044000", "define")

	assert.Equal(t, want, PinnedPerlAt(t, pin).Path,
		"the unthreaded versions-dir perl must be rejected for the threaded PATH perl")

	t.Setenv("PATH", t.TempDir())
	assert.True(t, skipped(t, func(st testing.TB) { PinnedPerlAt(st, pin) }),
		"with only the unthreaded perl left, the test skips")
}

// pinnedRequiredFailureChildEnv tells TestPinnedPerlRequiredFailureChild
// to run for real rather than skip itself; only
// TestPinnedPerlFailsUnderRequire's subprocess sets it.
const pinnedRequiredFailureChildEnv = "PVM_PERLTEST_PINNED_FAILURE_CHILD"

// TestPinnedPerlRequiredFailureChild calls PinnedPerl under
// PVM_REQUIRE_PINNED_PERL with no matching perl anywhere, so it must
// fail. Run directly it skips itself.
func TestPinnedPerlRequiredFailureChild(t *testing.T) {
	if os.Getenv(pinnedRequiredFailureChildEnv) == "" {
		t.Skip("only runs as TestPinnedPerlFailsUnderRequire's subprocess")
	}
	pinnedWorld(t)
	dir := t.TempDir()
	pinnedStandIn(t, dir, "5.042000", "define")
	t.Setenv("PATH", dir)
	t.Setenv(requirePinnedEnv, "1")
	PinnedPerlAt(t, pinFile(t, "5.044000", "define"))
}

// TestPinnedPerlFailsUnderRequire: PVM_REQUIRE_PINNED_PERL=1 turns the
// skip into a failure, for the CI steps that installed the pinned perl
// and must not pass by having every held test skip.
func TestPinnedPerlFailsUnderRequire(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestPinnedPerlRequiredFailureChild$", "-test.v")
	cmd.Env = append(os.Environ(), pinnedRequiredFailureChildEnv+"=1")
	out, err := cmd.CombinedOutput()
	assert.Error(t, err, "PinnedPerl under %s should fail the test, not pass:\n%s", requirePinnedEnv, out)
	assert.Contains(t, string(out), requirePinnedEnv)
	assert.Contains(t, string(out), "--- FAIL: TestPinnedPerlRequiredFailureChild")
}

// TestPinnedPerlReadsTheCommittedPin: PinnedPerl with no path reads
// internal/parseoracle/testdata/corpus.pin, whatever perl it names.
func TestPinnedPerlReadsTheCommittedPin(t *testing.T) {
	_, err := os.Stat(committedPin())
	assert.NoError(t, err, "PinnedPerl reads the pin at %s", committedPin())
}

// TestDottedVersion covers the conversion from $]'s packed form to the
// directory name pvm installs a version under.
func TestDottedVersion(t *testing.T) {
	for packed, want := range map[string]string{
		"5.044000": "5.44.0",
		"5.042000": "5.42.0",
		"5.038002": "5.38.2",
		"5.008009": "5.8.9",
		"5.010":    "5.10.0",
	} {
		assert.Equal(t, want, dottedVersion(packed), packed)
	}
}
