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

	// Move all contents of perl/ to the install root.
	// Known subdirs include bin, lib, site, and vendor, but we move
	// everything to handle future Strawberry layout changes.
	entries, err := os.ReadDir(perlDir)
	if err != nil {
		return fmt.Errorf("strawberry layout: read perl subdir: %w", err)
	}

	for _, entry := range entries {
		src := filepath.Join(perlDir, entry.Name())
		dst := filepath.Join(installDir, entry.Name())

		// Skip if destination already exists at root (e.g., c/ is at both levels)
		if _, err := os.Stat(dst); err == nil {
			continue
		}

		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("strawberry layout: move %s -> %s: %w", src, dst, err)
		}
	}

	// Remove the now-empty perl/ directory.
	if err := os.Remove(perlDir); err != nil {
		// Use RemoveAll as a fallback in case any files remain
		if err2 := os.RemoveAll(perlDir); err2 != nil {
			return fmt.Errorf("strawberry layout: remove perl subdir: %w", err2)
		}
	}

	return nil
}
