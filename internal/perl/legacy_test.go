// ABOUTME: Tests for legacy tool integration
// ABOUTME: Ensures proper detection and import from plenv and perlbrew

package perl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPerlVersionFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "perl-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a .perl-version file
	versionFile := filepath.Join(tempDir, ".perl-version")
	err = os.WriteFile(versionFile, []byte("5.32.1\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .perl-version file: %v", err)
	}

	// Test reading the file
	version, err := ReadPerlVersionFile(tempDir)
	if err != nil {
		t.Fatalf("Failed to read .perl-version file: %v", err)
	}

	if version != "5.32.1" {
		t.Errorf("Expected version '5.32.1', got '%s'", version)
	}

	// Test non-existent file
	nonExistentDir := filepath.Join(tempDir, "non-existent")
	if err := os.Mkdir(nonExistentDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	_, err = ReadPerlVersionFile(nonExistentDir)
	if err == nil {
		t.Error("Expected error when reading non-existent .perl-version file, got nil")
	}

	// Test empty file
	emptyDir := filepath.Join(tempDir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	emptyFile := filepath.Join(emptyDir, ".perl-version")
	err = os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty .perl-version file: %v", err)
	}

	_, err = ReadPerlVersionFile(emptyDir)
	if err == nil {
		t.Error("Expected error when reading empty .perl-version file, got nil")
	}
}

func TestFindDotPerlVersionFiles(t *testing.T) {
	// Create a nested directory structure
	baseDir, err := os.MkdirTemp("", "perl-version-search")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(baseDir) }()

	// Create subdirectories
	subDir1 := filepath.Join(baseDir, "sub1")
	subDir2 := filepath.Join(baseDir, "sub1", "sub2")
	subDir3 := filepath.Join(baseDir, "sub1", "sub2", "sub3")

	// Create directories
	for _, dir := range []string{subDir1, subDir2, subDir3} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create .perl-version files
	baseVersionFile := filepath.Join(baseDir, ".perl-version")
	err = os.WriteFile(baseVersionFile, []byte("5.32.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create base .perl-version file: %v", err)
	}

	subDir2VersionFile := filepath.Join(subDir2, ".perl-version")
	err = os.WriteFile(subDir2VersionFile, []byte("5.34.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sub2 .perl-version file: %v", err)
	}

	// Test finding files from the deepest directory
	files, err := FindDotPerlVersionFiles(subDir3)
	if err != nil {
		t.Fatalf("Failed to find .perl-version files: %v", err)
	}

	// We should find 2 files (from subDir2 and baseDir)
	if len(files) != 2 {
		t.Errorf("Expected to find 2 .perl-version files, got %d", len(files))
	}
}

func TestMockPlenv(t *testing.T) {
	// Create a mock plenv directory structure
	homeDir, err := os.MkdirTemp("", "mock-home")
	if err != nil {
		t.Fatalf("Failed to create mock home dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(homeDir) }()

	// Create plenv structure
	plenvDir := filepath.Join(homeDir, ".plenv")
	versionsDir := filepath.Join(plenvDir, "versions")
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		t.Fatalf("Failed to create plenv versions dir: %v", err)
	}

	// Create a global version file
	globalVersionFile := filepath.Join(plenvDir, "version")
	err = os.WriteFile(globalVersionFile, []byte("5.32.1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create global version file: %v", err)
	}

	// Create mock perl versions
	versions := []string{"5.30.3", "5.32.1", "5.34.0"}
	for _, version := range versions {
		versionDir := filepath.Join(versionsDir, version)
		binDir := filepath.Join(versionDir, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			t.Fatalf("Failed to create bin dir for %s: %v", version, err)
		}

		// Create mock perl binary
		perlBin := filepath.Join(binDir, "perl")
		err = os.WriteFile(perlBin, []byte("#!/bin/sh\necho test"), 0755)
		if err != nil {
			t.Fatalf("Failed to create mock perl binary for %s: %v", version, err)
		}
	}

	// Patch the userHomeDir function for testing
	unpatch := patchUserHomeDir(homeDir)
	defer unpatch()

	// Now test detecting plenv
	installations, err := DetectPlenv()
	if err != nil {
		t.Fatalf("Failed to detect mock plenv: %v", err)
	}

	if len(installations) != 3 {
		t.Errorf("Expected 3 plenv installations, got %d", len(installations))
	}

	// Check that 5.32.1 is marked as default
	found := false
	for _, inst := range installations {
		if inst.Version == "5.32.1" && inst.IsDefault {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 5.32.1 as the default version")
	}
}

func TestMockPerlbrew(t *testing.T) {
	// Create a mock perlbrew directory structure
	homeDir, err := os.MkdirTemp("", "mock-home")
	if err != nil {
		t.Fatalf("Failed to create mock home dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(homeDir) }()

	// Create perlbrew structure
	perlbrewDir := filepath.Join(homeDir, "perl5", "perlbrew")
	perlsDir := filepath.Join(perlbrewDir, "perls")
	if err := os.MkdirAll(perlsDir, 0755); err != nil {
		t.Fatalf("Failed to create perlbrew perls dir: %v", err)
	}

	// Create a current version file
	currentVersionFile := filepath.Join(perlbrewDir, "CURRENT")
	err = os.WriteFile(currentVersionFile, []byte("perl-5.32.1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create CURRENT file: %v", err)
	}

	// Create aliases file
	aliasesFile := filepath.Join(perlbrewDir, "aliases")
	aliasContent := "stable=perl-5.32.1\ndev=perl-5.39.0\n"
	err = os.WriteFile(aliasesFile, []byte(aliasContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create aliases file: %v", err)
	}

	// Create mock perl versions
	versions := []struct {
		dir string
		ver string
	}{
		{"perl-5.30.3", "5.30.3"},
		{"perl-5.32.1", "5.32.1"},
		{"perl-5.39.0", "5.39.0"},
	}

	for _, v := range versions {
		versionDir := filepath.Join(perlsDir, v.dir)
		binDir := filepath.Join(versionDir, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			t.Fatalf("Failed to create bin dir for %s: %v", v.dir, err)
		}

		// Create mock perl binary
		perlBin := filepath.Join(binDir, "perl")
		err = os.WriteFile(perlBin, []byte("#!/bin/sh\necho test"), 0755)
		if err != nil {
			t.Fatalf("Failed to create mock perl binary for %s: %v", v.dir, err)
		}
	}

	// Patch the userHomeDir function for testing
	unpatch := patchUserHomeDir(homeDir)
	defer unpatch()

	// Now test detecting perlbrew
	installations, err := DetectPerlbrew()
	if err != nil {
		t.Fatalf("Failed to detect mock perlbrew: %v", err)
	}

	if len(installations) != 3 {
		t.Errorf("Expected 3 perlbrew installations, got %d", len(installations))
	}

	// Check that perl-5.32.1 is marked as default
	found := false
	for _, inst := range installations {
		if inst.Version == "5.32.1" && inst.IsDefault {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 5.32.1 as the default version")
	}

	// Test getting aliases
	aliases, err := GetPerlbrewAliases()
	if err != nil {
		t.Fatalf("Failed to get perlbrew aliases: %v", err)
	}

	if len(aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(aliases))
	}

	if aliases["stable"] != "perl-5.32.1" {
		t.Errorf("Expected 'stable' alias to be 'perl-5.32.1', got '%s'", aliases["stable"])
	}
}

// Patch the userHomeDir function to return the homeDir for testing
func patchUserHomeDir(homeDir string) func() {
	original := userHomeDir
	userHomeDir = func() (string, error) {
		return homeDir, nil
	}
	return func() {
		userHomeDir = original
	}
}
