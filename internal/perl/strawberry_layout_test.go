// ABOUTME: Tests for Strawberry Perl directory layout relocation
// ABOUTME: Verifies that Strawberry Perl's nested perl/ layout is rearranged to match PVM's expected layout

package perl

import (
	"os"
	"path/filepath"
	"testing"
)

// createFile creates a file at the given path, creating parent directories as needed.
func createFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create parent dirs for %s: %v", path, err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
	f.Close()
}

// fileExists returns true if the path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestRelocateStrawberryLayout(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Strawberry Perl's nested layout under perl/
	createFile(t, filepath.Join(tmpDir, "perl", "bin", "perl.exe"))
	createFile(t, filepath.Join(tmpDir, "perl", "bin", "perl5.38.0.exe"))
	createFile(t, filepath.Join(tmpDir, "perl", "lib", "strict.pm"))
	createFile(t, filepath.Join(tmpDir, "perl", "site", "lib", "Moose.pm"))
	// Also create the C toolchain at root (should not be touched)
	createFile(t, filepath.Join(tmpDir, "c", "bin", "gcc.exe"))

	err := relocateStrawberryLayout(tmpDir)
	if err != nil {
		t.Fatalf("relocateStrawberryLayout returned unexpected error: %v", err)
	}

	// perl/bin/* should now be at bin/*
	if !fileExists(filepath.Join(tmpDir, "bin", "perl.exe")) {
		t.Error("expected bin/perl.exe to exist after relocation")
	}
	if !fileExists(filepath.Join(tmpDir, "bin", "perl5.38.0.exe")) {
		t.Error("expected bin/perl5.38.0.exe to exist after relocation")
	}

	// perl/lib should now be at lib/
	if !fileExists(filepath.Join(tmpDir, "lib", "strict.pm")) {
		t.Error("expected lib/strict.pm to exist after relocation")
	}

	// perl/site should now be at site/
	if !fileExists(filepath.Join(tmpDir, "site", "lib", "Moose.pm")) {
		t.Error("expected site/lib/Moose.pm to exist after relocation")
	}

	// The nested perl/ directory should be gone
	if fileExists(filepath.Join(tmpDir, "perl")) {
		t.Error("expected perl/ subdirectory to be removed after relocation")
	}

	// The c/ toolchain directory must remain untouched
	if !fileExists(filepath.Join(tmpDir, "c", "bin", "gcc.exe")) {
		t.Error("expected c/bin/gcc.exe to remain at root after relocation")
	}
}

func TestRelocateStrawberryLayout_AlreadyCorrect(t *testing.T) {
	tmpDir := t.TempDir()

	// Layout already matches PVM expectations — no perl/ subdirectory
	createFile(t, filepath.Join(tmpDir, "bin", "perl.exe"))
	createFile(t, filepath.Join(tmpDir, "lib", "strict.pm"))

	err := relocateStrawberryLayout(tmpDir)
	if err != nil {
		t.Fatalf("relocateStrawberryLayout returned unexpected error for already-correct layout: %v", err)
	}

	// Existing files must still be in place
	if !fileExists(filepath.Join(tmpDir, "bin", "perl.exe")) {
		t.Error("expected bin/perl.exe to still exist after no-op relocation")
	}
	if !fileExists(filepath.Join(tmpDir, "lib", "strict.pm")) {
		t.Error("expected lib/strict.pm to still exist after no-op relocation")
	}
}

func TestRelocateStrawberryLayout_PreservesToolchain(t *testing.T) {
	tmpDir := t.TempDir()

	// Full Strawberry layout including toolchain
	createFile(t, filepath.Join(tmpDir, "perl", "bin", "perl.exe"))
	createFile(t, filepath.Join(tmpDir, "perl", "lib", "strict.pm"))
	createFile(t, filepath.Join(tmpDir, "c", "bin", "gcc.exe"))
	createFile(t, filepath.Join(tmpDir, "c", "bin", "gmake.exe"))
	createFile(t, filepath.Join(tmpDir, "c", "lib", "libgcc.a"))

	err := relocateStrawberryLayout(tmpDir)
	if err != nil {
		t.Fatalf("relocateStrawberryLayout returned unexpected error: %v", err)
	}

	// Perl binaries relocated correctly
	if !fileExists(filepath.Join(tmpDir, "bin", "perl.exe")) {
		t.Error("expected bin/perl.exe to exist after relocation")
	}

	// All C toolchain entries must be untouched at root
	if !fileExists(filepath.Join(tmpDir, "c", "bin", "gcc.exe")) {
		t.Error("expected c/bin/gcc.exe to remain after relocation")
	}
	if !fileExists(filepath.Join(tmpDir, "c", "bin", "gmake.exe")) {
		t.Error("expected c/bin/gmake.exe to remain after relocation")
	}
	if !fileExists(filepath.Join(tmpDir, "c", "lib", "libgcc.a")) {
		t.Error("expected c/lib/libgcc.a to remain after relocation")
	}

	// perl/ subdirectory should be gone
	if fileExists(filepath.Join(tmpDir, "perl")) {
		t.Error("expected perl/ subdirectory to be removed after relocation")
	}
}
