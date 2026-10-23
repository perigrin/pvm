// ABOUTME: Tests the shim's interpreter stamp: BuildShim records which perl built it, and a shim another perl built is rebuilt, never measured with.
// ABOUTME: A 5.42 lib/Config.pm under a 5.44 perl mis-measures silently, so the stamp is the only thing that notices.

package parseoracle_test

import (
	"os"
	"path/filepath"
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
	if err := parseoracle.BuildShim(perl, fakeCorpus(t), shim); err != nil {
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
	if err := parseoracle.BuildShim(shimPerl(t, pin42.Interpreter), fakeCorpus(t), shim); err != nil {
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
	buildFixtureShim(t, p, shim)

	if !parseoracle.ShimReady(shim, pin44) {
		t.Error("the fixture shim is not ready for a pin naming the perl it was built for")
	}
	if parseoracle.ShimReady(shim, pin42) {
		t.Error("the fixture shim is ready for a pin naming another perl")
	}
}
