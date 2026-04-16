// ABOUTME: Post-install step that rewrites ELF RPATH entries to $ORIGIN-
// ABOUTME: relative form so the built Perl tarball works at any install path.
// Perl's Configure bakes the build-time absolute prefix into DT_RUNPATH, which
// breaks every user who installs to a different path — including our own
// release binaries (built on GH Actions runners whose $HOME is
// /home/runner/…). Rewriting RPATH to $ORIGIN/<rel-to-CORE> per-file makes
// the dynamic linker resolve libperl.so relative to each ELF's own location.

package perl

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// makeRelocatable rewrites RPATH/RUNPATH on every ELF object under
// installDir so references to libperl.so resolve via $ORIGIN rather than an
// absolute path baked at link time. The relative portion of the rpath is
// computed per-file so XS .so modules deep under lib/perl5/.../auto/ get a
// correct relative path to CORE (not the bin/-relative path).
//
// Requires patchelf on PATH. Returns a descriptive error when patchelf is
// missing so the caller can decide whether to continue with a warning or
// fail the build. No-op on non-linux platforms — Mach-O relocation for
// macOS is tracked as a follow-up (see issue filed alongside PR #438).
func makeRelocatable(installDir string) error {
	if runtime.GOOS != "linux" {
		return nil
	}

	patchelf, err := exec.LookPath("patchelf")
	if err != nil {
		return fmt.Errorf("patchelf not found on PATH: %w "+
			"(required to make the built Perl relocatable; "+
			"install via apt/dnf/brew or build a Go patchelf equivalent)", err)
	}

	coreDir, err := findCoreDir(installDir)
	if err != nil {
		return fmt.Errorf("locate CORE dir: %w", err)
	}

	// Walk the install tree and patch every ELF file. Errors from
	// individual files are accumulated so one stubborn file doesn't leave
	// the whole tree half-patched silently.
	var walkErrs []error
	err = filepath.WalkDir(installDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// A per-entry walk error (permission denied, broken symlink)
			// shouldn't abort the whole post-install step; record and move on.
			walkErrs = append(walkErrs, fmt.Errorf("walk %s: %w", path, walkErr))
			return nil
		}
		if d.IsDir() {
			return nil
		}
		// Skip symlinks explicitly. Following them risks patching system
		// binaries a malicious tarball could symlink into the install tree.
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if !isELF(path) {
			return nil
		}
		// Per-file relative rpath: $ORIGIN/<rel from file's dir to CORE>.
		// This makes bin/perl, lib/perl5/.../CORE/libperl.so, and every XS
		// .so under lib/perl5/.../auto/**/*.so each find libperl.so
		// correctly despite being at different depths.
		rel, relErr := filepath.Rel(filepath.Dir(path), coreDir)
		if relErr != nil {
			walkErrs = append(walkErrs, fmt.Errorf("compute rel path for %s: %w", path, relErr))
			return nil
		}
		rpath := "$ORIGIN/" + rel
		if patchErr := patchRpath(patchelf, path, rpath); patchErr != nil {
			walkErrs = append(walkErrs, patchErr)
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(walkErrs) > 0 {
		return errors.Join(walkErrs...)
	}
	return nil
}

// findCoreDir returns the absolute path to the directory containing
// libperl.so under installDir. Requires that the directory be named "CORE"
// — a stray libperl.so* elsewhere in the tree (leftover build artifacts,
// bundled shim libraries) would otherwise silently misroute every RPATH.
func findCoreDir(installDir string) (string, error) {
	var coreDir string
	err := filepath.WalkDir(installDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base != "libperl.so" && !strings.HasPrefix(base, "libperl.so.") {
			return nil
		}
		parent := filepath.Base(filepath.Dir(path))
		if parent != "CORE" {
			return nil
		}
		coreDir = filepath.Dir(path)
		return fs.SkipAll
	})
	if err != nil {
		return "", err
	}
	if coreDir == "" {
		return "", fmt.Errorf("libperl.so not found in a CORE/ directory under %s", installDir)
	}
	return coreDir, nil
}

// isELF returns true when path names a regular file whose first four bytes
// match the ELF magic. Uses io.ReadFull so short reads don't accidentally
// match zero-filled tail bytes.
func isELF(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return false
	}
	return magic == [4]byte{0x7f, 'E', 'L', 'F'}
}

// patchRpath runs `patchelf --set-rpath <rpath> <path>`, unless the
// current RPATH already starts with $ORIGIN (idempotent). patchelf (as of
// 0.18) does not support a `--` argv terminator, so we guard against
// filenames starting with `-` explicitly — extremely unlikely inside a
// real Perl install tree, but cheap defense against a malicious tarball.
func patchRpath(patchelf, path, rpath string) error {
	if strings.HasPrefix(filepath.Base(path), "-") {
		return fmt.Errorf("refusing to patchelf file with flag-like name: %s", path)
	}
	cur, err := exec.Command(patchelf, "--print-rpath", path).Output()
	if err == nil {
		curStr := strings.TrimSpace(string(cur))
		if strings.HasPrefix(curStr, "$ORIGIN") {
			return nil
		}
	}
	cmd := exec.Command(patchelf, "--set-rpath", rpath, path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("patchelf --set-rpath on %s: %w (%s)", path, err, string(out))
	}
	return nil
}
