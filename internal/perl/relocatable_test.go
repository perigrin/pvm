// ABOUTME: Tests for the relocatable-RPATH post-install step.
// ABOUTME: Fakes an install tree with bin/perl + CORE/libperl.so + an XS .so
// ABOUTME: module, runs makeRelocatable, moves the tree, and verifies both
// ABOUTME: the main binary and the dlopened XS module still resolve libperl.

package perl

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFindCoreDir(t *testing.T) {
	// Given an install tree with libperl.so at
	// <prefix>/lib/perl5/5.42.0/x86_64-linux-thread-multi/CORE/libperl.so
	// findCoreDir returns the absolute path to the CORE directory.
	tmp := t.TempDir()
	coreDir := filepath.Join(tmp, "lib", "perl5", "5.42.0", "x86_64-linux-thread-multi", "CORE")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir CORE: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "libperl.so"), []byte{0x7f, 'E', 'L', 'F'}, 0o644); err != nil {
		t.Fatalf("write fake libperl.so: %v", err)
	}

	got, err := findCoreDir(tmp)
	if err != nil {
		t.Fatalf("findCoreDir: %v", err)
	}
	if got != coreDir {
		t.Errorf("CORE dir: got %q, want %q", got, coreDir)
	}
}

func TestFindCoreDir_NoLibperl(t *testing.T) {
	tmp := t.TempDir()
	_, err := findCoreDir(tmp)
	if err == nil {
		t.Errorf("expected error when libperl.so is absent, got nil")
	}
}

// TestFindCoreDir_StrayLibperlOutsideCORE guards against a libperl.so in
// some non-CORE directory silently misrouting the whole rewrite. The
// function must require the parent dir to be named CORE.
func TestFindCoreDir_StrayLibperlOutsideCORE(t *testing.T) {
	tmp := t.TempDir()
	// Stray libperl.so not in a CORE dir — from a previous failed build,
	// or a bundled shim library.
	strayDir := filepath.Join(tmp, "lib", "stray")
	if err := os.MkdirAll(strayDir, 0o755); err != nil {
		t.Fatalf("mkdir stray: %v", err)
	}
	if err := os.WriteFile(filepath.Join(strayDir, "libperl.so"), []byte{0x7f, 'E', 'L', 'F'}, 0o644); err != nil {
		t.Fatalf("write stray libperl.so: %v", err)
	}
	// Real libperl.so in CORE.
	coreDir := filepath.Join(tmp, "lib", "perl5", "5.42.0", "arch", "CORE")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir CORE: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "libperl.so"), []byte{0x7f, 'E', 'L', 'F'}, 0o644); err != nil {
		t.Fatalf("write real libperl.so: %v", err)
	}

	got, err := findCoreDir(tmp)
	if err != nil {
		t.Fatalf("findCoreDir: %v", err)
	}
	if got != coreDir {
		t.Errorf("findCoreDir picked wrong libperl: got %q, want %q (CORE-parented)", got, coreDir)
	}
}

// TestMakeRelocatable_EndToEnd only runs when patchelf is on PATH and we're
// on linux — the actual rpath rewrite needs a real ELF toolchain. The test
// builds a main binary that genuinely calls a libperl symbol (forcing a
// DT_NEEDED even with --as-needed) AND an XS-style .so at a different
// depth, then moves the whole tree and verifies both still resolve.
func TestMakeRelocatable_EndToEnd(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("relocatable RPATH test is linux-only")
	}
	if _, err := exec.LookPath("patchelf"); err != nil {
		t.Skip("patchelf not available")
	}
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not available")
	}

	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	coreDir := filepath.Join(tmp, "lib", "perl5", "5.99.0", "x86_64-linux-thread-multi", "CORE")
	// XS modules live several directories deeper. Their relative path to
	// CORE is different from bin/'s — the per-file rpath computation must
	// handle this correctly.
	xsDir := filepath.Join(tmp, "lib", "perl5", "5.99.0", "x86_64-linux-thread-multi", "auto", "Foo", "Foo")
	for _, d := range []string{binDir, coreDir, xsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	// libperl.so — exports perl_hello() so DT_NEEDED is forced even under
	// --as-needed.
	libSrc := filepath.Join(tmp, "libperl.c")
	if err := os.WriteFile(libSrc, []byte("int perl_hello(void){return 42;}\n"), 0o644); err != nil {
		t.Fatalf("write libperl.c: %v", err)
	}
	libOut := filepath.Join(coreDir, "libperl.so")
	if err := exec.Command("gcc", "-shared", "-fPIC", libSrc, "-o", libOut).Run(); err != nil {
		t.Fatalf("compile libperl.so: %v", err)
	}

	// Main binary. main() calls perl_hello() so the linker keeps
	// DT_NEEDED libperl.so.0 (or unversioned) regardless of --as-needed.
	// The original -rpath points at a nonsense path so without the fix
	// the binary would not resolve libperl at runtime.
	binSrc := filepath.Join(tmp, "perl.c")
	if err := os.WriteFile(binSrc, []byte("int perl_hello(void);int main(void){return perl_hello()==42?0:1;}\n"), 0o644); err != nil {
		t.Fatalf("write perl.c: %v", err)
	}
	binOut := filepath.Join(binDir, "perl")
	cmd := exec.Command("gcc", binSrc, "-o", binOut,
		"-L", coreDir, "-lperl",
		"-Wl,-rpath,/nonsense/path/that/doesnt/exist")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile perl: %v (%s)", err, string(out))
	}

	// XS-style .so at a different depth. It also references perl_hello.
	// When dlopened, it must find libperl.so via its own $ORIGIN-relative
	// RPATH (which is deeper — ../../../CORE, not ../lib/.../CORE).
	xsSrc := filepath.Join(tmp, "foo.c")
	if err := os.WriteFile(xsSrc, []byte("int perl_hello(void);int foo_call(void){return perl_hello();}\n"), 0o644); err != nil {
		t.Fatalf("write foo.c: %v", err)
	}
	xsOut := filepath.Join(xsDir, "Foo.so")
	if err := exec.Command("gcc", "-shared", "-fPIC", xsSrc, "-o", xsOut,
		"-L", coreDir, "-lperl",
		"-Wl,-rpath,/nonsense/path").Run(); err != nil {
		t.Fatalf("compile Foo.so: %v", err)
	}

	// Apply the relocatable fix.
	if err := makeRelocatable(tmp); err != nil {
		t.Fatalf("makeRelocatable: %v", err)
	}

	// Main binary RPATH should be the bin-relative path to CORE.
	out, err := exec.Command("patchelf", "--print-rpath", binOut).Output()
	if err != nil {
		t.Fatalf("patchelf --print-rpath bin/perl: %v", err)
	}
	wantBin := "$ORIGIN/../lib/perl5/5.99.0/x86_64-linux-thread-multi/CORE"
	if !strings.Contains(string(out), wantBin) {
		t.Errorf("bin/perl rpath: got %q, want contains %q", string(out), wantBin)
	}

	// XS .so is deeper — its RPATH must reflect that (shorter/different
	// relative path), NOT the bin-relative path. This is the I2 guard.
	out, err = exec.Command("patchelf", "--print-rpath", xsOut).Output()
	if err != nil {
		t.Fatalf("patchelf --print-rpath Foo.so: %v", err)
	}
	// From <xsDir> back up to <coreDir>: ../../../CORE
	wantXS := "$ORIGIN/../../../CORE"
	if !strings.Contains(string(out), wantXS) {
		t.Errorf("Foo.so rpath: got %q, want contains %q (XS depth ≠ bin depth)", string(out), wantXS)
	}

	// Binary runs from the original location.
	if err := exec.Command(binOut).Run(); err != nil {
		t.Fatalf("binary fails to run in place after relocatable fix: %v", err)
	}

	// Move the whole tree to a new location; binary still runs — proving
	// the RPATH really is relocatable (not an artifact of the build host's
	// library search path).
	newTmp := t.TempDir()
	movedRoot := filepath.Join(newTmp, "moved")
	if err := os.Rename(tmp, movedRoot); err != nil {
		t.Fatalf("rename: %v", err)
	}
	movedBin := filepath.Join(movedRoot, "bin", "perl")
	if err := exec.Command(movedBin).Run(); err != nil {
		t.Errorf("binary fails to run after being moved: %v", err)
	}
}
