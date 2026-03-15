// ABOUTME: Tests for automatic import functionality during pvm init
// ABOUTME: Ensures plenv and perlbrew installations are correctly detected and imported

package perl

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/xdg"
)

func TestAutoImportLegacyVersions(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	origHomeDir := userHomeDir
	origLoadRegistry := LoadRegistry
	origSaveRegistry := SaveRegistry
	origGetDirs := xdg.GetDirs

	defer func() {
		userHomeDir = origHomeDir
		LoadRegistry = origLoadRegistry
		SaveRegistry = origSaveRegistry
		xdg.GetDirs = origGetDirs
	}()

	// Mock userHomeDir to return temp directory
	userHomeDir = func() (string, error) {
		return tempDir, nil
	}

	// Mock XDG directories
	xdg.GetDirs = func() (*xdg.Dirs, error) {
		pvmDataDir := filepath.Join(tempDir, ".local", "share", "pvm")
		return &xdg.Dirs{
			DataDir:   pvmDataDir,
			ConfigDir: filepath.Join(tempDir, ".config", "pvm"),
			CacheDir:  filepath.Join(tempDir, ".cache", "pvm"),
		}, nil
	}

	// Mock registry functions to use temp directory
	registry := &VersionRegistry{
		Versions: make(map[string]VersionInfo),
	}
	LoadRegistry = func() (*VersionRegistry, error) {
		return registry, nil
	}
	SaveRegistry = func(r *VersionRegistry) error {
		registry = r
		return nil
	}

	// Create fake plenv installation
	plenvDir := filepath.Join(tempDir, ".plenv")
	plenvVersionsDir := filepath.Join(plenvDir, "versions")
	plenv538Dir := filepath.Join(plenvVersionsDir, "5.38.0")
	plenv536Dir := filepath.Join(plenvVersionsDir, "5.36.0")

	err := os.MkdirAll(filepath.Join(plenv538Dir, "bin"), 0755)
	if err != nil {
		t.Fatalf("Failed to create plenv test directory: %v", err)
	}
	err = os.MkdirAll(filepath.Join(plenv536Dir, "bin"), 0755)
	if err != nil {
		t.Fatalf("Failed to create plenv test directory: %v", err)
	}

	// Create fake perl binaries
	perlBin538 := filepath.Join(plenv538Dir, "bin", "perl")
	perlBin536 := filepath.Join(plenv536Dir, "bin", "perl")
	err = os.WriteFile(perlBin538, []byte("#!/bin/bash\necho 'perl 5.38.0'\n"), 0755)
	if err != nil {
		t.Fatalf("Failed to create fake perl binary: %v", err)
	}
	err = os.WriteFile(perlBin536, []byte("#!/bin/bash\necho 'perl 5.36.0'\n"), 0755)
	if err != nil {
		t.Fatalf("Failed to create fake perl binary: %v", err)
	}

	// Set 5.38.0 as default in plenv
	err = os.WriteFile(filepath.Join(plenvDir, "version"), []byte("5.38.0\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create plenv version file: %v", err)
	}

	// Create fake perlbrew installation
	perlbrewDir := filepath.Join(tempDir, "perl5", "perlbrew")
	perlbrewPerlsDir := filepath.Join(perlbrewDir, "perls")
	perlbrew532Dir := filepath.Join(perlbrewPerlsDir, "perl-5.32.1")

	err = os.MkdirAll(filepath.Join(perlbrew532Dir, "bin"), 0755)
	if err != nil {
		t.Fatalf("Failed to create perlbrew test directory: %v", err)
	}

	// Create fake perl binary for perlbrew
	perlBin532 := filepath.Join(perlbrew532Dir, "bin", "perl")
	err = os.WriteFile(perlBin532, []byte("#!/bin/bash\necho 'perl 5.32.1'\n"), 0755)
	if err != nil {
		t.Fatalf("Failed to create fake perl binary: %v", err)
	}

	// Run auto import
	results, err := AutoImportLegacyVersions()
	if err != nil {
		t.Fatalf("AutoImportLegacyVersions failed: %v", err)
	}

	// Verify results
	if results.TotalImported != 3 {
		t.Errorf("Expected 3 versions imported, got %d", results.TotalImported)
	}

	if len(results.PlenvImports) != 2 {
		t.Errorf("Expected 2 plenv imports, got %d", len(results.PlenvImports))
	}

	if len(results.PerlbrewImports) != 1 {
		t.Errorf("Expected 1 perlbrew import, got %d", len(results.PerlbrewImports))
	}

	// Verify default version is set correctly
	if results.DefaultVersion != "5.38.0" {
		t.Errorf("Expected default version 5.38.0, got %s", results.DefaultVersion)
	}

	// Verify plenv imports
	foundVersions := make(map[string]bool)
	for _, result := range results.PlenvImports {
		foundVersions[result.Version] = true
		if result.Source != Plenv {
			t.Errorf("Expected source to be plenv, got %s", result.Source)
		}
		if result.Version == "5.38.0" && !result.WasDefault {
			t.Errorf("Expected 5.38.0 to be marked as default")
		}
		if result.Version == "5.36.0" && result.WasDefault {
			t.Errorf("Expected 5.36.0 to not be marked as default")
		}
	}

	if !foundVersions["5.38.0"] || !foundVersions["5.36.0"] {
		t.Errorf("Missing expected plenv versions")
	}

	// Verify perlbrew import
	if results.PerlbrewImports[0].Version != "5.32.1" {
		t.Errorf("Expected perlbrew version 5.32.1, got %s", results.PerlbrewImports[0].Version)
	}
	if results.PerlbrewImports[0].Source != Perlbrew {
		t.Errorf("Expected source to be perlbrew, got %s", results.PerlbrewImports[0].Source)
	}

	// Verify registry was updated
	if len(registry.Versions) != 3 {
		t.Errorf("Expected 3 versions in registry, got %d", len(registry.Versions))
	}

	// Verify symlinks were created
	pvmVersionsDir := filepath.Join(tempDir, ".local", "share", "pvm", "versions")
	expectedSymlinks := []string{"5.38.0", "5.36.0", "5.32.1"}
	for _, version := range expectedSymlinks {
		symlinkPath := filepath.Join(pvmVersionsDir, version)
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Errorf("Expected symlink for %s, but file does not exist: %v", version, err)
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("Expected %s to be a symlink", symlinkPath)
		}
	}
}

func TestShouldAutoImport(t *testing.T) {
	// Test case 1: No versions installed (should import)
	origGetInstalledVersions := GetInstalledVersions
	defer func() {
		GetInstalledVersions = origGetInstalledVersions
	}()

	GetInstalledVersions = func() ([]VersionInfo, error) {
		return []VersionInfo{}, nil
	}

	if !ShouldAutoImport() {
		t.Error("Expected ShouldAutoImport to return true when no versions are installed")
	}

	// Test case 2: Versions already installed (should not import)
	GetInstalledVersions = func() ([]VersionInfo, error) {
		return []VersionInfo{
			{Version: "5.38.0", Source: "pvm"},
		}, nil
	}

	if ShouldAutoImport() {
		t.Error("Expected ShouldAutoImport to return false when versions are already installed")
	}

	// Test case 3: Error getting versions (should import as fallback)
	GetInstalledVersions = func() ([]VersionInfo, error) {
		return nil, os.ErrNotExist
	}

	if !ShouldAutoImport() {
		t.Error("Expected ShouldAutoImport to return true when error occurs getting versions")
	}
}

func TestImportFromPlenvWithExistingVersions(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	origHomeDir := userHomeDir
	origLoadRegistry := LoadRegistry
	origSaveRegistry := SaveRegistry
	origGetDirs := xdg.GetDirs
	origIsVersionInstalled := IsVersionInstalled

	defer func() {
		userHomeDir = origHomeDir
		LoadRegistry = origLoadRegistry
		SaveRegistry = origSaveRegistry
		xdg.GetDirs = origGetDirs
		IsVersionInstalled = origIsVersionInstalled
	}()

	// Mock userHomeDir to return temp directory
	userHomeDir = func() (string, error) {
		return tempDir, nil
	}

	// Mock XDG directories
	xdg.GetDirs = func() (*xdg.Dirs, error) {
		pvmDataDir := filepath.Join(tempDir, ".local", "share", "pvm")
		return &xdg.Dirs{
			DataDir:   pvmDataDir,
			ConfigDir: filepath.Join(tempDir, ".config", "pvm"),
			CacheDir:  filepath.Join(tempDir, ".cache", "pvm"),
		}, nil
	}

	// Mock registry functions
	registry := &VersionRegistry{
		Versions: make(map[string]VersionInfo),
	}
	LoadRegistry = func() (*VersionRegistry, error) {
		return registry, nil
	}
	SaveRegistry = func(r *VersionRegistry) error {
		registry = r
		return nil
	}

	// Mock IsVersionInstalled to return true for 5.38.0 (already installed)
	IsVersionInstalled = func(version string) (bool, error) {
		return version == "5.38.0", nil
	}

	// Create fake plenv installation with two versions
	plenvDir := filepath.Join(tempDir, ".plenv")
	plenvVersionsDir := filepath.Join(plenvDir, "versions")
	plenv538Dir := filepath.Join(plenvVersionsDir, "5.38.0")
	plenv536Dir := filepath.Join(plenvVersionsDir, "5.36.0")

	err := os.MkdirAll(filepath.Join(plenv538Dir, "bin"), 0755)
	if err != nil {
		t.Fatalf("Failed to create plenv test directory: %v", err)
	}
	err = os.MkdirAll(filepath.Join(plenv536Dir, "bin"), 0755)
	if err != nil {
		t.Fatalf("Failed to create plenv test directory: %v", err)
	}

	// Create fake perl binaries
	perlBin538 := filepath.Join(plenv538Dir, "bin", "perl")
	perlBin536 := filepath.Join(plenv536Dir, "bin", "perl")
	err = os.WriteFile(perlBin538, []byte("#!/bin/bash\necho 'perl 5.38.0'\n"), 0755)
	if err != nil {
		t.Fatalf("Failed to create fake perl binary: %v", err)
	}
	err = os.WriteFile(perlBin536, []byte("#!/bin/bash\necho 'perl 5.36.0'\n"), 0755)
	if err != nil {
		t.Fatalf("Failed to create fake perl binary: %v", err)
	}

	// Run import
	results, err := importFromPlenv()
	if err != nil {
		t.Fatalf("importFromPlenv failed: %v", err)
	}

	// Should only import 5.36.0 since 5.38.0 is already installed
	if len(results) != 1 {
		t.Errorf("Expected 1 import result, got %d", len(results))
	}

	if len(results) > 0 && results[0].Version != "5.36.0" {
		t.Errorf("Expected to import 5.36.0, got %s", results[0].Version)
	}
}
