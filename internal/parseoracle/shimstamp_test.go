// ABOUTME: Tests the shim's interpreter stamp: BuildShim records which perl built it, and a shim another perl built is rebuilt, never measured with.
// ABOUTME: A 5.42 lib/Config.pm under a 5.44 perl mis-measures silently, so the stamp is the only thing that notices.

package parseoracle_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/perltest"
)

var (
	pin44 = parseoracle.Pin{Interpreter: "5.044000", Revision: "deadbeef", Threads: true}
	pin42 = parseoracle.Pin{Interpreter: "5.042000", Revision: "deadbeef", Threads: true}
)

// shimPerl writes a stand-in perl that claims to be version (threaded)
// when asked what it is, and names two small library roots when asked
// for @INC, which is everything BuildShim asks a perl.
func shimPerl(t *testing.T, version string) string {
	t.Helper()
	requirePOSIXOracle(t)
	arch, pure := t.TempDir(), t.TempDir()
	for _, f := range []string{filepath.Join(arch, "Config.pm"), filepath.Join(pure, "strict.pm")} {
		if err := os.WriteFile(f, []byte("1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "perl")
	script := "#!/bin/sh\ncase \"$*\" in\n" +
		"*useithreads*) printf '%s\\n%s\\n' '" + version + "' define ;;\n" +
		"*) printf '%s\\n%s\\n' '" + arch + "' '" + pure + "' ;;\n" +
		"esac\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// fakeCorpus is a perl5 checkout as small as BuildShim accepts: a t/
// with test.pl in it.
func fakeCorpus(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "t"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "t", "test.pl"), []byte("1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestShimMatchesThePin: BuildShim writes interpreter.stamp from the perl
// that built it, and ShimReady accepts the shim for a pin naming that
// perl and refuses it for a pin naming any other interpreter or build.
func TestShimMatchesThePin(t *testing.T) {
	perl := shimPerl(t, pin44.Interpreter)
	shim := t.TempDir()
	if err := parseoracle.BuildShim(perl, fakeCorpus(t), shim, pin44.Revision); err != nil {
		t.Fatalf("BuildShim: %v", err)
	}

	if _, err := os.Stat(filepath.Join(shim, "lib", "Config.pm")); err != nil {
		t.Errorf("BuildShim did not populate lib/ from the perl it was given: %v", err)
	}
	if !parseoracle.ShimReady(shim, pin44) {
		t.Error("a shim built by perl 5.044000 (threads) is not ready for a pin naming exactly that")
	}
	if parseoracle.ShimReady(shim, pin42) {
		t.Error("a shim built by 5.044000 was accepted for a 5.042000 pin")
	}
	unthreaded := pin44
	unthreaded.Threads = false
	if parseoracle.ShimReady(shim, unthreaded) {
		t.Error("a shim built by a threaded perl was accepted for an unthreaded pin")
	}
}

// TestShimStampMismatchRebuilds: a shim stamped 5.042000 under a
// 5.044000 pin is not ready. RebuildShim replaces it outright -- BuildShim
// alone copies like `cp -n` and would keep the old lib/ -- and fails if
// the rebuilt shim still disagrees with the pin or the corpus is absent.
func TestShimStampMismatchRebuilds(t *testing.T) {
	shim := t.TempDir()
	if err := parseoracle.BuildShim(shimPerl(t, pin42.Interpreter), fakeCorpus(t), shim, pin42.Revision); err != nil {
		t.Fatalf("BuildShim: %v", err)
	}
	stale := filepath.Join(shim, "lib", "stale-5.42-only.pm")
	if err := os.WriteFile(stale, []byte("1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if parseoracle.ShimReady(shim, pin44) {
		t.Fatal("a shim stamped 5.042000 is ready under a 5.044000 pin: it would be measured with as is")
	}

	if err := parseoracle.RebuildShim(shimPerl(t, pin44.Interpreter), fakeCorpus(t), shim, pin44); err != nil {
		t.Fatalf("RebuildShim: %v", err)
	}
	if !parseoracle.ShimReady(shim, pin44) {
		t.Error("the rebuilt shim is not ready for the pin it was rebuilt for")
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("the rebuilt shim kept a file from the 5.42 build: it was copied over, not rebuilt")
	}

	// A perl that cannot produce the pin rebuilds a shim that still
	// disagrees, and that is an error, not a shim to measure with.
	if err := parseoracle.RebuildShim(shimPerl(t, pin42.Interpreter), fakeCorpus(t), shim, pin44); err == nil {
		t.Error("RebuildShim accepted a shim rebuilt by 5.042000 for a 5.044000 pin")
	}
	// No corpus to rebuild from is an error too.
	if err := parseoracle.RebuildShim(shimPerl(t, pin44.Interpreter), filepath.Join(t.TempDir(), "absent"), shim, pin44); err == nil {
		t.Error("RebuildShim succeeded with no corpus to rebuild from")
	}
}

// TestFixtureShimCarriesTheStamp: the gate's hand-built five-file shim
// carries the stamp of the perl it was built for, so the gate's sweep,
// which withholds the corpus on purpose, measures it rather than trying
// to rebuild it.
func TestFixtureShimCarriesTheStamp(t *testing.T) {
	p := perltest.Perl{Path: shimPerl(t, pin44.Interpreter)}
	shim := t.TempDir()
	buildFixtureShim(t, p, shim, pin44.Revision)

	if !parseoracle.ShimReady(shim, pin44) {
		t.Error("the fixture shim is not ready for a pin naming the perl it was built for")
	}
	if parseoracle.ShimReady(shim, pin42) {
		t.Error("the fixture shim is ready for a pin naming another perl")
	}
}

// TestRebuildShimRefusesAForeignDirectory: RebuildShim removes the old
// shim, so a $PARSEORACLE_SHIM pointed by mistake at a perl5 checkout
// would delete it. A directory holding anything besides t/, lib/ and the
// stamp is refused and left exactly as it was.
func TestRebuildShimRefusesAForeignDirectory(t *testing.T) {
	dir := fakeCorpus(t)
	precious := filepath.Join(dir, "precious.c")
	if err := os.WriteFile(precious, []byte("int main;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := parseoracle.RebuildShim(shimPerl(t, pin44.Interpreter), fakeCorpus(t), dir, pin44); err == nil {
		t.Error("RebuildShim accepted a directory that is not a shim")
	} else if !strings.Contains(err.Error(), "precious.c") && !strings.Contains(err.Error(), ".git") {
		t.Errorf("the refusal does not name what it would have deleted: %v", err)
	}
	for _, keep := range []string{precious, filepath.Join(dir, ".git"), filepath.Join(dir, "t", "test.pl")} {
		if _, err := os.Stat(keep); err != nil {
			t.Errorf("RebuildShim removed %s from a directory it refused: %v", keep, err)
		}
	}
}

// TestRebuildShimIsSafeConcurrently: workers sharing one shim may rebuild
// it at once. Each must succeed and leave a ready shim, never a half-
// removed tree or a stamp check that read a sibling's deletion.
func TestRebuildShimIsSafeConcurrently(t *testing.T) {
	perl := shimPerl(t, pin44.Interpreter)
	corpus := fakeCorpus(t)
	for round := 0; round < 3; round++ {
		shim := filepath.Join(t.TempDir(), "shim")
		var wg sync.WaitGroup
		errs := make([]error, 4)
		for i := range errs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				errs[i] = parseoracle.RebuildShim(perl, corpus, shim, pin44)
			}(i)
		}
		wg.Wait()
		for i, err := range errs {
			if err != nil {
				t.Errorf("round %d, rebuild %d: %v", round, i, err)
			}
		}
		if !parseoracle.ShimReady(shim, pin44) {
			t.Errorf("round %d left a shim that is not ready", round)
		}
	}
}

// TestShimStampRecordsTheRevision: the shim's t/ is the corpus at one
// revision, so a pin that moves only the revision must not accept it.
func TestShimStampRecordsTheRevision(t *testing.T) {
	shim := t.TempDir()
	if err := parseoracle.BuildShim(shimPerl(t, pin44.Interpreter), fakeCorpus(t), shim, pin44.Revision); err != nil {
		t.Fatalf("BuildShim: %v", err)
	}
	moved := pin44
	moved.Revision = "cafef00d"
	if parseoracle.ShimReady(shim, moved) {
		t.Error("a shim built from revision deadbeef was accepted for a pin naming cafef00d")
	}
	if !parseoracle.ShimReady(shim, pin44) {
		t.Error("a shim is not ready for the very pin it was built for")
	}
}

// TestRebuildShimNamesTheMissingFile: a rebuilt shim that is not ready
// says why. A corpus with no t/test.pl builds a shim missing it, and the
// error names that file rather than reporting a stamp mismatch.
func TestRebuildShimNamesTheMissingFile(t *testing.T) {
	corpus := t.TempDir()
	if err := os.MkdirAll(filepath.Join(corpus, "t"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := parseoracle.RebuildShim(shimPerl(t, pin44.Interpreter), corpus, filepath.Join(t.TempDir(), "shim"), pin44)
	if err == nil || !strings.Contains(err.Error(), "test.pl") {
		t.Errorf("want an error naming the missing t/test.pl, got %v", err)
	}
}
