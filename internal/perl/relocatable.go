// ABOUTME: Post-install step that rewrites baked-in dynamic-linker paths so
// ABOUTME: built Perl tarballs work at any install path. Linux uses patchelf
// ABOUTME: ($ORIGIN RPATH); macOS uses install_name_tool (@loader_path).

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

// makeRelocatable rewrites dynamic-linker paths on every binary/shared-lib
// under installDir so references to libperl resolve relative to each file's
// own location rather than via an absolute path baked at build time.
//
// Dispatches to a platform-specific implementation:
//   - Linux: patchelf --set-rpath with $ORIGIN/<rel-to-CORE>
//   - macOS: install_name_tool -change/-id/-add_rpath with @loader_path
//
// Returns a descriptive error when the required tool is missing so the caller
// can decide whether to continue with a warning or fail the build.
func makeRelocatable(installDir string) error {
	switch runtime.GOOS {
	case "linux":
		return makeRelocatableLinux(installDir)
	case "darwin":
		return makeRelocatableDarwin(installDir)
	default:
		return nil
	}
}

// ---------------------------------------------------------------------------
// Linux (ELF / patchelf)
// ---------------------------------------------------------------------------

func makeRelocatableLinux(installDir string) error {
	patchelf, err := exec.LookPath("patchelf")
	if err != nil {
		return fmt.Errorf("patchelf not found on PATH: %w "+
			"(required to make the built Perl relocatable; "+
			"install via apt/dnf/brew)", err)
	}

	coreDir, err := findCoreDir(installDir)
	if err != nil {
		return fmt.Errorf("locate CORE dir: %w", err)
	}

	var walkErrs []error
	err = filepath.WalkDir(installDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			walkErrs = append(walkErrs, fmt.Errorf("walk %s: %w", path, walkErr))
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if !isELF(path) {
			return nil
		}
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

// ---------------------------------------------------------------------------
// macOS (Mach-O / install_name_tool)
// ---------------------------------------------------------------------------

// makeRelocatableDarwin rewrites LC_ID_DYLIB, LC_LOAD_DYLIB, and LC_RPATH
// entries on every Mach-O file under installDir so libperl.dylib resolves
// via @loader_path/<rel-to-CORE> regardless of where the tree is placed.
//
// After mutation, every Mach-O is ad-hoc re-signed with `codesign -f -s -`
// because macOS 11+ rejects binaries whose signatures are invalidated by
// install_name_tool.
func makeRelocatableDarwin(installDir string) error {
	inameTool, err := exec.LookPath("install_name_tool")
	if err != nil {
		return fmt.Errorf("install_name_tool not found on PATH: %w "+
			"(ships with Xcode Command Line Tools)", err)
	}
	if _, err := exec.LookPath("otool"); err != nil {
		return fmt.Errorf("otool not found on PATH: %w "+
			"(ships with Xcode Command Line Tools)", err)
	}

	coreDir, err := findCoreDir(installDir)
	if err != nil {
		return fmt.Errorf("locate CORE dir: %w", err)
	}

	// Discover the absolute build-host path to libperl that will be baked
	// into every LC_LOAD_DYLIB entry. We need this as the "old" argument
	// in install_name_tool -change <old> <new>.
	libperlName, err := findLibperlName(coreDir)
	if err != nil {
		return fmt.Errorf("find libperl dylib name: %w", err)
	}

	codesign, _ := exec.LookPath("codesign")

	var walkErrs []error
	err = filepath.WalkDir(installDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			walkErrs = append(walkErrs, fmt.Errorf("walk %s: %w", path, walkErr))
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if !isMachO(path) {
			return nil
		}

		rel, relErr := filepath.Rel(filepath.Dir(path), coreDir)
		if relErr != nil {
			walkErrs = append(walkErrs, fmt.Errorf("compute rel path for %s: %w", path, relErr))
			return nil
		}
		loaderRpath := "@loader_path/" + rel

		if patchErr := patchMachO(inameTool, codesign, path, loaderRpath, libperlName); patchErr != nil {
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

// findLibperlName returns the filename of libperl inside coreDir — either
// "libperl.dylib" or a versioned variant like "libperl.5.40.dylib". The
// caller needs this to construct the install_name_tool -change argument.
func findLibperlName(coreDir string) (string, error) {
	entries, err := os.ReadDir(coreDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		name := e.Name()
		if name == "libperl.dylib" || (strings.HasPrefix(name, "libperl.") && strings.HasSuffix(name, ".dylib")) {
			return name, nil
		}
	}
	// Perl on macOS may also build with the .so extension when configured
	// with -Duseshrplib but without explicit darwin shared-lib flags.
	for _, e := range entries {
		name := e.Name()
		if name == "libperl.so" || strings.HasPrefix(name, "libperl.so.") {
			return name, nil
		}
	}
	return "", fmt.Errorf("no libperl dylib/so found in %s", coreDir)
}

// patchMachO applies install_name_tool mutations to a single Mach-O file:
//
//  1. If the file IS libperl itself: -id @loader_path/<basename> so other
//     Mach-O files' LC_LOAD_DYLIB entries can reference it via @rpath.
//  2. For every file: query current LC_LOAD_DYLIB entries via otool -L. If
//     any reference libperl via an absolute path, -change <old> @rpath/<name>.
//  3. For every file: -add_rpath <loaderRpath> so @rpath resolves to CORE.
//  4. Ad-hoc re-sign with codesign -f -s - (macOS 11+ requirement).
func patchMachO(inameTool, codesign, path, loaderRpath, libperlName string) error {
	base := filepath.Base(path)

	// If this is libperl itself, set its install name to @rpath/<name> so
	// binaries referencing it via @rpath find it at runtime.
	if base == libperlName {
		if out, err := exec.Command(inameTool, "-id", "@rpath/"+libperlName, path).CombinedOutput(); err != nil {
			return fmt.Errorf("install_name_tool -id on %s: %w (%s)", path, err, string(out))
		}
	}

	// Discover current LC_LOAD_DYLIB entries that reference libperl via
	// an absolute build-host path and rewrite them to @rpath/<name>.
	otoolOut, err := exec.Command("otool", "-L", path).Output()
	if err == nil {
		for _, line := range strings.Split(string(otoolOut), "\n") {
			line = strings.TrimSpace(line)
			// otool -L lines look like:
			//   /absolute/path/libperl.dylib (compatibility version ...)
			if !strings.Contains(line, libperlName) {
				continue
			}
			// Skip the @rpath or @loader_path entries — already relocated.
			if strings.HasPrefix(line, "@") {
				continue
			}
			// Extract the path (everything before the first " (").
			oldPath := line
			if idx := strings.Index(line, " ("); idx > 0 {
				oldPath = line[:idx]
			}
			// Skip if it's already the ID line for the dylib itself (first
			// line of otool -L output for a dylib).
			if oldPath == "@rpath/"+libperlName {
				continue
			}
			if out, err := exec.Command(inameTool, "-change", oldPath, "@rpath/"+libperlName, path).CombinedOutput(); err != nil {
				return fmt.Errorf("install_name_tool -change on %s: %w (%s)", path, err, string(out))
			}
		}
	}

	// Add an LC_RPATH entry so @rpath resolves from this file's location
	// to the CORE dir containing libperl. install_name_tool -add_rpath
	// fails if the rpath already exists; check first.
	existingRpaths := rpathsFromOtool(path)
	if !containsStr(existingRpaths, loaderRpath) {
		if out, err := exec.Command(inameTool, "-add_rpath", loaderRpath, path).CombinedOutput(); err != nil {
			return fmt.Errorf("install_name_tool -add_rpath on %s: %w (%s)", path, err, string(out))
		}
	}

	// Ad-hoc re-sign. macOS 11+ invalidates code signatures after
	// install_name_tool mutations. An unsigned binary triggers SIP
	// enforcement on some configurations.
	if codesign != "" {
		if out, err := exec.Command(codesign, "-f", "-s", "-", path).CombinedOutput(); err != nil {
			return fmt.Errorf("codesign on %s: %w (%s)", path, err, string(out))
		}
	}

	return nil
}

// rpathsFromOtool returns existing LC_RPATH entries for a Mach-O file.
func rpathsFromOtool(path string) []string {
	out, err := exec.Command("otool", "-l", path).Output()
	if err != nil {
		return nil
	}
	var rpaths []string
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if strings.Contains(line, "LC_RPATH") {
			// The path line follows two lines after the LC_RPATH header:
			//   cmd LC_RPATH
			//   cmdsize ...
			//   path /some/path (offset ...)
			for j := i + 1; j < len(lines) && j <= i+3; j++ {
				trimmed := strings.TrimSpace(lines[j])
				if strings.HasPrefix(trimmed, "path ") {
					p := strings.TrimPrefix(trimmed, "path ")
					if idx := strings.Index(p, " (offset"); idx > 0 {
						p = p[:idx]
					}
					rpaths = append(rpaths, strings.TrimSpace(p))
					break
				}
			}
		}
	}
	return rpaths
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// findCoreDir returns the absolute path to the directory containing the
// shared libperl under installDir. Requires that the directory be named
// "CORE". On Linux looks for libperl.so*; on macOS for libperl*.dylib or
// libperl.so* (Perl can build either on darwin depending on Configure flags).
func findCoreDir(installDir string) (string, error) {
	var coreDir string
	err := filepath.WalkDir(installDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if !isLibperl(base) {
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
		return "", fmt.Errorf("libperl.{so,dylib} not found in a CORE/ directory under %s", installDir)
	}
	return coreDir, nil
}

// isLibperl returns true for filenames matching libperl shared libraries on
// either platform: libperl.so, libperl.so.5.40.2, libperl.dylib,
// libperl.5.40.dylib, etc.
func isLibperl(name string) bool {
	if name == "libperl.so" || strings.HasPrefix(name, "libperl.so.") {
		return true
	}
	if name == "libperl.dylib" || (strings.HasPrefix(name, "libperl.") && strings.HasSuffix(name, ".dylib")) {
		return true
	}
	return false
}

// isELF returns true when path names a regular file whose first four bytes
// match the ELF magic. Uses io.ReadFull so short reads don't accidentally
// match zero-filled tail bytes.
func isELF(path string) bool {
	return hasMagic(path, [4]byte{0x7f, 'E', 'L', 'F'})
}

// isMachO returns true when path names a Mach-O file (thin or fat/universal).
// Checks the first 4 bytes against the known Mach-O magic values.
func isMachO(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return false
	}
	switch magic {
	case [4]byte{0xCE, 0xFA, 0xED, 0xFE}: // MH_MAGIC (32-bit LE)
		return true
	case [4]byte{0xCF, 0xFA, 0xED, 0xFE}: // MH_MAGIC_64 (64-bit LE)
		return true
	case [4]byte{0xFE, 0xED, 0xFA, 0xCE}: // MH_CIGAM (32-bit BE)
		return true
	case [4]byte{0xFE, 0xED, 0xFA, 0xCF}: // MH_CIGAM_64 (64-bit BE)
		return true
	case [4]byte{0xCA, 0xFE, 0xBA, 0xBE}: // FAT_MAGIC (universal)
		return true
	case [4]byte{0xBE, 0xBA, 0xFE, 0xCA}: // FAT_CIGAM
		return true
	default:
		return false
	}
}

func hasMagic(path string, want [4]byte) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return false
	}
	return magic == want
}

// patchRpath runs `patchelf --set-rpath <rpath> <path>`, unless the
// current RPATH already starts with $ORIGIN (idempotent).
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

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
