// ABOUTME: Tests for the relocatable-RPATH post-install step.
// ABOUTME: Fakes an install tree and checks the RPATH-rewrite logic.

package perl

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestComputeRelativeCoreDir(t *testing.T) {
	// Given an install tree with libperl.so at
	// <prefix>/lib/perl5/5.42.0/x86_64-linux-thread-multi/CORE/libperl.so
	// the relative path from <prefix>/bin/<binary> to the CORE dir
	// should be ../lib/perl5/5.42.0/x86_64-linux-thread-multi/CORE.
	tmp := t.TempDir()
	coreDir := filepath.Join(tmp, "lib", "perl5", "5.42.0", "x86_64-linux-thread-multi", "CORE")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir CORE: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "libperl.so"), []byte{0x7f, 0x45, 0x4c, 0x46}, 0o644); err != nil {
		t.Fatalf("write fake libperl.so: %v", err)
	}

	rel, err := computeRelativeCoreDir(tmp)
	if err != nil {
		t.Fatalf("computeRelativeCoreDir: %v", err)
	}
	want := filepath.Join("..", "lib", "perl5", "5.42.0", "x86_64-linux-thread-multi", "CORE")
	if rel != want {
		t.Errorf("relative CORE dir: got %q, want %q", rel, want)
	}
}

func TestComputeRelativeCoreDir_NoLibperl(t *testing.T) {
	tmp := t.TempDir()
	_, err := computeRelativeCoreDir(tmp)
	if err == nil {
		t.Errorf("expected error when libperl.so is absent, got nil")
	}
}

// TestMakeRelocatable_EndToEnd only runs when patchelf is on PATH and we're
// on linux — the actual rpath rewrite needs a real ELF toolchain.
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
	// Build a fake perl tree with a real ELF binary + libperl.so.
	binDir := filepath.Join(tmp, "bin")
	coreDir := filepath.Join(tmp, "lib", "perl5", "5.99.0", "x86_64-linux-thread-multi", "CORE")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir CORE: %v", err)
	}

	// libperl.so
	libSrc := filepath.Join(tmp, "libperl.c")
	if err := os.WriteFile(libSrc, []byte("void perl_hello(void){}\n"), 0o644); err != nil {
		t.Fatalf("write libperl.c: %v", err)
	}
	libOut := filepath.Join(coreDir, "libperl.so")
	if err := exec.Command("gcc", "-shared", "-fPIC", libSrc, "-o", libOut).Run(); err != nil {
		t.Fatalf("compile libperl.so: %v", err)
	}

	// perl binary with HARDCODED bad RPATH (simulating the current bug).
	binSrc := filepath.Join(tmp, "perl.c")
	if err := os.WriteFile(binSrc, []byte("int main(void){return 0;}\n"), 0o644); err != nil {
		t.Fatalf("write perl.c: %v", err)
	}
	binOut := filepath.Join(binDir, "perl")
	cmd := exec.Command("gcc", binSrc, "-o", binOut,
		"-L", coreDir, "-lperl",
		"-Wl,-rpath,/nonsense/path/that/doesnt/exist")
	if err := cmd.Run(); err != nil {
		t.Fatalf("compile perl: %v", err)
	}

	// Apply the relocatable fix.
	if err := makeRelocatable(tmp); err != nil {
		t.Fatalf("makeRelocatable: %v", err)
	}

	// Verify the RPATH now contains $ORIGIN and resolves correctly.
	out, err := exec.Command("patchelf", "--print-rpath", binOut).Output()
	if err != nil {
		t.Fatalf("patchelf --print-rpath: %v", err)
	}
	got := string(out)
	want := "$ORIGIN/../lib/perl5/5.99.0/x86_64-linux-thread-multi/CORE"
	if !containsString(got, want) {
		t.Errorf("rewritten RPATH: got %q, want contains %q", got, want)
	}

	// And critically: the binary runs.
	if err := exec.Command(binOut).Run(); err != nil {
		t.Errorf("binary fails to run after relocatable fix: %v", err)
	}

	// Move the whole tree to a new location and verify it still runs
	// (proving the RPATH is genuinely relocatable).
	newTmp := t.TempDir()
	if err := os.Rename(tmp, filepath.Join(newTmp, "moved")); err != nil {
		t.Fatalf("rename: %v", err)
	}
	movedBin := filepath.Join(newTmp, "moved", "bin", "perl")
	if err := exec.Command(movedBin).Run(); err != nil {
		t.Errorf("binary fails to run after being moved: %v", err)
	}
}

func containsString(hay, needle string) bool {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
