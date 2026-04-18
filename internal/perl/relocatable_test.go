// ABOUTME: Tests for the relocatable post-install step (Linux + macOS).
// ABOUTME: Fakes an install tree, runs makeRelocatable, moves the tree, and
// ABOUTME: verifies the main binary + XS module still resolve libperl.

package perl

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// findCoreDir unit tests (cross-platform)
// ---------------------------------------------------------------------------

func TestFindCoreDir(t *testing.T) {
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

func TestFindCoreDir_Dylib(t *testing.T) {
	tmp := t.TempDir()
	coreDir := filepath.Join(tmp, "lib", "perl5", "5.42.0", "darwin-thread-multi-2level", "CORE")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir CORE: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "libperl.dylib"), []byte("fake"), 0o644); err != nil {
		t.Fatalf("write fake libperl.dylib: %v", err)
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

func TestFindCoreDir_StrayLibperlOutsideCORE(t *testing.T) {
	tmp := t.TempDir()
	strayDir := filepath.Join(tmp, "lib", "stray")
	if err := os.MkdirAll(strayDir, 0o755); err != nil {
		t.Fatalf("mkdir stray: %v", err)
	}
	if err := os.WriteFile(filepath.Join(strayDir, "libperl.so"), []byte{0x7f, 'E', 'L', 'F'}, 0o644); err != nil {
		t.Fatalf("write stray libperl.so: %v", err)
	}
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

// ---------------------------------------------------------------------------
// isLibperl unit test
// ---------------------------------------------------------------------------

func TestIsLibperl(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"libperl.so", true},
		{"libperl.so.5.40.2", true},
		{"libperl.dylib", true},
		{"libperl.5.40.dylib", true},
		{"libperl.a", false},
		{"perl", false},
		{"libperl5.so", false},
	}
	for _, tc := range cases {
		if got := isLibperl(tc.name); got != tc.want {
			t.Errorf("isLibperl(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Linux end-to-end test
// ---------------------------------------------------------------------------

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
	xsDir := filepath.Join(tmp, "lib", "perl5", "5.99.0", "x86_64-linux-thread-multi", "auto", "Foo", "Foo")
	for _, d := range []string{binDir, coreDir, xsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	// libperl.so
	libSrc := filepath.Join(tmp, "libperl.c")
	if err := os.WriteFile(libSrc, []byte("int perl_hello(void){return 42;}\n"), 0o644); err != nil {
		t.Fatalf("write libperl.c: %v", err)
	}
	libOut := filepath.Join(coreDir, "libperl.so")
	if err := exec.Command("gcc", "-shared", "-fPIC", libSrc, "-o", libOut).Run(); err != nil {
		t.Fatalf("compile libperl.so: %v", err)
	}

	// Main binary — calls perl_hello() to force DT_NEEDED.
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

	// XS-style .so at a different depth.
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

	if err := makeRelocatable(tmp); err != nil {
		t.Fatalf("makeRelocatable: %v", err)
	}

	out, err := exec.Command("patchelf", "--print-rpath", binOut).Output()
	if err != nil {
		t.Fatalf("patchelf --print-rpath bin/perl: %v", err)
	}
	wantBin := "$ORIGIN/../lib/perl5/5.99.0/x86_64-linux-thread-multi/CORE"
	if !strings.Contains(string(out), wantBin) {
		t.Errorf("bin/perl rpath: got %q, want contains %q", string(out), wantBin)
	}

	out, err = exec.Command("patchelf", "--print-rpath", xsOut).Output()
	if err != nil {
		t.Fatalf("patchelf --print-rpath Foo.so: %v", err)
	}
	wantXS := "$ORIGIN/../../../CORE"
	if !strings.Contains(string(out), wantXS) {
		t.Errorf("Foo.so rpath: got %q, want contains %q (XS depth != bin depth)", string(out), wantXS)
	}

	if err := exec.Command(binOut).Run(); err != nil {
		t.Fatalf("binary fails to run in place after relocatable fix: %v", err)
	}

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

// ---------------------------------------------------------------------------
// macOS end-to-end test
// ---------------------------------------------------------------------------

func TestMakeRelocatable_Darwin_EndToEnd(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin relocatable test is macOS-only")
	}
	if _, err := exec.LookPath("install_name_tool"); err != nil {
		t.Skip("install_name_tool not available")
	}
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not available")
	}

	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	coreDir := filepath.Join(tmp, "lib", "perl5", "5.99.0", "darwin-thread-multi-2level", "CORE")
	xsDir := filepath.Join(tmp, "lib", "perl5", "5.99.0", "darwin-thread-multi-2level", "auto", "Foo", "Foo")
	for _, d := range []string{binDir, coreDir, xsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	// libperl.dylib — exports perl_hello() so the link is kept.
	libSrc := filepath.Join(tmp, "libperl.c")
	if err := os.WriteFile(libSrc, []byte("int perl_hello(void){return 42;}\n"), 0o644); err != nil {
		t.Fatalf("write libperl.c: %v", err)
	}
	libOut := filepath.Join(coreDir, "libperl.dylib")
	// Build with a deliberately bad install_name (simulating what Perl's
	// Configure bakes in on the build host). -headerpad_max_install_names
	// leaves room for install_name_tool to rewrite paths later.
	if out, err := exec.Command("clang", "-shared", "-fPIC", libSrc, "-o", libOut,
		"-Wl,-headerpad_max_install_names",
		"-install_name", "/nonsense/build/host/path/libperl.dylib").CombinedOutput(); err != nil {
		t.Fatalf("compile libperl.dylib: %v (%s)", err, string(out))
	}

	// Main binary — calls perl_hello() so DT_NEEDED (LC_LOAD_DYLIB on
	// macOS) is genuinely needed at runtime.
	binSrc := filepath.Join(tmp, "perl.c")
	if err := os.WriteFile(binSrc, []byte("int perl_hello(void);int main(void){return perl_hello()==42?0:1;}\n"), 0o644); err != nil {
		t.Fatalf("write perl.c: %v", err)
	}
	// -headerpad_max_install_names ensures the Mach-O header has enough
	// room for install_name_tool to add LC_RPATH entries later. Real Perl
	// builds get this from Configure; our stub binaries need it explicitly.
	binOut := filepath.Join(binDir, "perl")
	if out, err := exec.Command("clang", binSrc, "-o", binOut,
		"-L", coreDir, "-lperl",
		"-Wl,-headerpad_max_install_names",
		"-Wl,-rpath,/nonsense/path/that/doesnt/exist").CombinedOutput(); err != nil {
		t.Fatalf("compile perl: %v (%s)", err, string(out))
	}

	// XS-style .bundle at a different depth.
	xsSrc := filepath.Join(tmp, "foo.c")
	if err := os.WriteFile(xsSrc, []byte("int perl_hello(void);int foo_call(void){return perl_hello();}\n"), 0o644); err != nil {
		t.Fatalf("write foo.c: %v", err)
	}
	xsOut := filepath.Join(xsDir, "Foo.bundle")
	if out, err := exec.Command("clang", "-bundle", xsSrc, "-o", xsOut,
		"-L", coreDir, "-lperl",
		"-Wl,-headerpad_max_install_names",
		"-Wl,-rpath,/nonsense/path").CombinedOutput(); err != nil {
		t.Fatalf("compile Foo.bundle: %v (%s)", err, string(out))
	}

	// Apply the relocatable fix.
	if err := makeRelocatable(tmp); err != nil {
		t.Fatalf("makeRelocatable: %v", err)
	}

	// libperl's install name should now be @rpath/libperl.dylib.
	otoolOut, err := exec.Command("otool", "-D", libOut).Output()
	if err != nil {
		t.Fatalf("otool -D libperl.dylib: %v", err)
	}
	if !strings.Contains(string(otoolOut), "@rpath/libperl.dylib") {
		t.Errorf("libperl install name: got %q, want contains @rpath/libperl.dylib", string(otoolOut))
	}

	// The main binary should reference libperl via @rpath, and have an
	// LC_RPATH entry using @loader_path that reaches CORE.
	otoolOut, err = exec.Command("otool", "-L", binOut).Output()
	if err != nil {
		t.Fatalf("otool -L bin/perl: %v", err)
	}
	if !strings.Contains(string(otoolOut), "@rpath/libperl.dylib") {
		t.Errorf("bin/perl LC_LOAD_DYLIB: got %q, want @rpath/libperl.dylib reference", string(otoolOut))
	}

	rpaths := rpathsFromOtool(binOut)
	foundLoaderPath := false
	for _, rp := range rpaths {
		if strings.HasPrefix(rp, "@loader_path/") {
			foundLoaderPath = true
			break
		}
	}
	if !foundLoaderPath {
		t.Errorf("bin/perl LC_RPATH: no @loader_path entry found; rpaths=%v", rpaths)
	}

	// Binary runs from the original location.
	if err := exec.Command(binOut).Run(); err != nil {
		t.Fatalf("binary fails to run in place after relocatable fix: %v", err)
	}

	// Move the whole tree and verify the binary still runs — the @rpath
	// + @loader_path chain must resolve libperl.dylib from the new
	// location.
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
