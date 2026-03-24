// ABOUTME: Strawberry Perl directory layout relocation for Windows installations
// ABOUTME: Rearranges the nested perl/ subdirectory to match PVM's expected flat layout

package perl

import (
	"fmt"
	"os"
	"path/filepath"
)

// relocateStrawberryLayout rearranges a freshly extracted Strawberry Perl portable
// zip so that its directory layout matches what PVM expects.
//
// Strawberry Perl portable zips extract to a nested structure:
//
//	installDir/
//	  perl/
//	    bin/   <- perl executables
//	    lib/   <- core library tree
//	    site/  <- site library tree (may be absent)
//	  c/       <- C toolchain, stays at root
//
// PVM expects:
//
//	installDir/
//	  bin/
//	  lib/
//	  site/    <- if present
//	  c/       <- toolchain unchanged
//
// If no perl/ subdirectory is found the layout is assumed to be already correct
// and the function returns nil immediately.
func relocateStrawberryLayout(installDir string) error {
	perlDir := filepath.Join(installDir, "perl")

	// Nothing to do if the nested perl/ directory does not exist.
	if _, err := os.Stat(perlDir); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("strawberry layout: stat perl subdir: %w", err)
	}

	// Subdirectories inside perl/ that need to be moved to the root.
	// site/ is optional — skip it if absent.
	subDirs := []string{"bin", "lib", "site"}

	for _, sub := range subDirs {
		src := filepath.Join(perlDir, sub)
		dst := filepath.Join(installDir, sub)

		if _, err := os.Stat(src); os.IsNotExist(err) {
			// Optional subdirectory not present; skip.
			continue
		} else if err != nil {
			return fmt.Errorf("strawberry layout: stat %s: %w", src, err)
		}

		// Ensure the parent of dst exists (dst itself must not exist for Rename).
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("strawberry layout: mkdirall for %s: %w", dst, err)
		}

		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("strawberry layout: move %s -> %s: %w", src, dst, err)
		}
	}

	// Remove the now-empty perl/ directory.
	if err := os.Remove(perlDir); err != nil {
		return fmt.Errorf("strawberry layout: remove perl subdir: %w", err)
	}

	return nil
}
