// ABOUTME: Test helper that finds the perl the parse oracle's corpus.pin names, without trusting PATH.
// ABOUTME: A test held to that perl calls PinnedPerl, which skips where none matches, or fails under PVM_REQUIRE_PINNED_PERL.

package perltest

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/platform"
	"tamarou.com/pvm/internal/xdg"
)

// requirePinnedEnv turns PinnedPerl's skip into a failure, for a CI step
// that installed the pinned perl and must not pass by every held test
// skipping.
const requirePinnedEnv = "PVM_REQUIRE_PINNED_PERL"

// PinnedPerl returns the perl internal/parseoracle/testdata/corpus.pin
// names: the first candidate whose $] and useithreads both match the pin,
// looking in pvm's versions directory, then pvm's registry, then PATH. It
// skips t where none matches, or fails it when PVM_REQUIRE_PINNED_PERL is
// set.
//
// PATH is last because PATH is what the oracle used to trust, and here it
// names whatever perl the shell defaults to, which need not be the one
// the pin records. The versions directory is first because pvm's registry
// can name an install that has since moved.
//
// Only a test function calls PinnedPerl; a helper takes the Perl it
// returns as a parameter, so TestEveryPinnedPerlTestIsRequired can see
// every held test by name.
func PinnedPerl(t testing.TB) Perl {
	t.Helper()
	return PinnedPerlAt(t, committedPin())
}

// PinnedPerlAt is PinnedPerl for the pin file at pinPath, for a test that
// must not depend on which perl the committed pin names.
func PinnedPerlAt(t testing.TB, pinPath string) Perl {
	t.Helper()
	pin, err := parseoracle.ReadPin(pinPath)
	if err != nil {
		t.Fatalf("PinnedPerl: %v", err)
	}
	for _, path := range pinnedCandidates(pin.Interpreter) {
		if observed, err := parseoracle.ObservePin(context.Background(), path, ""); err == nil &&
			observed.Interpreter == pin.Interpreter && observed.Threads == pin.Threads {
			return locateUnder(t, path, requirePinnedEnv)
		}
	}
	skipOrRequireUnder(t, requirePinnedEnv,
		"no perl in pvm's versions directory, pvm's registry or PATH reports $] %s with useithreads %t, as %s pins",
		pin.Interpreter, pin.Threads, pinPath)
	return Perl{}
}

// pinnedCandidates lists, in lookup order, the perls that might be the
// pinned one. A candidate that does not exist fails ObservePin and is
// passed over.
func pinnedCandidates(packed string) []string {
	version := dottedVersion(packed)
	var paths []string
	if dirs, err := xdg.GetDirs(); err == nil {
		paths = append(paths, filepath.Join(dirs.VersionsDir, version, "bin", platform.ExecutableName("perl")))
	}
	if info, err := perl.GetVersionInfo(version); err == nil && info != nil {
		paths = append(paths, perlExecutable(info.InstallPath))
	}
	if path, err := exec.LookPath("perl"); err == nil {
		paths = append(paths, path)
	}
	return paths
}

// committedPin is the oracle's pin, found beside this source file so a
// test in any package reads the same one.
func committedPin() string {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		panic(errors.New("cannot locate the perltest package source"))
	}
	return filepath.Join(filepath.Dir(self), "..", "parseoracle", "testdata", "corpus.pin")
}

// dottedVersion turns $]'s packed form into the dotted one pvm names its
// installs by: 5.044000 is 5.44.0.
func dottedVersion(packed string) string {
	major, frac, _ := strings.Cut(packed, ".")
	frac = (frac + "000000")[:6]
	minor, _ := strconv.Atoi(frac[:3])
	patch, _ := strconv.Atoi(frac[3:])
	return fmt.Sprintf("%s.%d.%d", major, minor, patch)
}
