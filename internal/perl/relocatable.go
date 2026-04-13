// ABOUTME: Post-install step that rewrites ELF RPATH entries to $ORIGIN-
// ABOUTME: relative form so the built Perl tarball works at any install path.
// Perl's Configure bakes the build-time absolute prefix into DT_RUNPATH, which
// breaks every user who installs to a different path — including our own
// release binaries (built on GH Actions runners whose $HOME is
// /home/runner/…). Rewriting RPATH to $ORIGIN/../lib/perl5/VER/ARCH/CORE
// makes the dynamic linker resolve it relative to the binary's own location.

package perl

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// makeRelocatable rewrites RPATH/RUNPATH on every ELF object under
// installDir so references to libperl.so resolve via $ORIGIN rather than an
// absolute path baked at link time.
//
// Requires patchelf on PATH. Returns a descriptive error when patchelf is
// missing so the caller can decide whether to continue with a warning or
// fail the build. No-op on non-linux platforms — only ELF needs this fix;
// macOS uses @rpath/@loader_path with install_name_tool, handled elsewhere.
func makeRelocatable(installDir string) error {
	if runtime.GOOS != "linux" {
		return nil
	}

	if _, err := exec.LookPath("patchelf"); err != nil {
		return fmt.Errorf("patchelf not found on PATH: %w "+
			"(required to make the built Perl relocatable; "+
			"install via apt/dnf/brew or build a Go patchelf equivalent)", err)
	}

	relCore, err := computeRelativeCoreDir(installDir)
	if err != nil {
		return fmt.Errorf("locate CORE dir: %w", err)
	}

	// Use $ORIGIN (literal, not shell-expanded) so the dynamic linker
	// resolves it at runtime relative to the file being loaded.
	rpath := "$ORIGIN/" + relCore

	// Walk the install tree and patch every ELF file that has a RPATH
	// mentioning the install prefix, or a needed libperl.so.
	return filepath.WalkDir(installDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Walk errors on individual entries shouldn't abort the whole
			// post-install step — log and continue.
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !isELF(path) {
			return nil
		}
		return patchRpath(path, rpath)
	})
}

// computeRelativeCoreDir finds libperl.so under installDir and returns the
// path from <installDir>/bin to the directory containing libperl.so, using
// filepath.Rel. When a binary in bin/ has RPATH=$ORIGIN/<this-result>, the
// dynamic linker finds libperl.so regardless of where the install tree sits.
func computeRelativeCoreDir(installDir string) (string, error) {
	var libperl string
	err := filepath.WalkDir(installDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		// libperl.so or libperl.so.5.40.2 etc.
		if base == "libperl.so" || strings.HasPrefix(base, "libperl.so.") {
			libperl = path
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if libperl == "" {
		return "", fmt.Errorf("libperl.so not found under %s", installDir)
	}
	coreDir := filepath.Dir(libperl)
	binDir := filepath.Join(installDir, "bin")
	rel, err := filepath.Rel(binDir, coreDir)
	if err != nil {
		return "", fmt.Errorf("compute relative path from %s to %s: %w", binDir, coreDir, err)
	}
	return rel, nil
}

// isELF returns true when path names a file whose first four bytes are the
// ELF magic. Cheap and avoids shelling out to `file`.
func isELF(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var magic [4]byte
	if _, err := f.Read(magic[:]); err != nil {
		return false
	}
	return magic == [4]byte{0x7f, 'E', 'L', 'F'}
}

// patchRpath runs `patchelf --set-rpath <rpath> <path>`, unless the current
// RPATH already starts with $ORIGIN (idempotent). Also strips any build-host
// paths that might linger from the original link.
func patchRpath(path, rpath string) error {
	// Check current rpath; skip if it already uses $ORIGIN.
	cur, err := exec.Command("patchelf", "--print-rpath", path).Output()
	if err == nil {
		curStr := strings.TrimSpace(string(cur))
		if strings.HasPrefix(curStr, "$ORIGIN") {
			return nil
		}
	}
	cmd := exec.Command("patchelf", "--set-rpath", rpath, path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("patchelf --set-rpath on %s: %w (%s)", path, err, string(out))
	}
	return nil
}
