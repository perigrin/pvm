// ABOUTME: Tests for system version identifier resolution
// ABOUTME: Ensures "system" resolves to imported system Perl version

package perl

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"tamarou.com/pvm/internal/config"
)

func TestResolveExplicitVersion_SystemIdentifier(t *testing.T) {
	// Save original functions
	origGetInstalledVersions := GetInstalledVersions
	origLoadRegistry := LoadRegistry
	defer func() {
		GetInstalledVersions = origGetInstalledVersions
		LoadRegistry = origLoadRegistry
	}()

	t.Run("SystemIdentifierResolvesToImportedSystemPerl", func(t *testing.T) {
		// Mock GetInstalledVersions to return a system Perl
		GetInstalledVersions = func() ([]VersionInfo, error) {
			return []VersionInfo{
				{
					Version:     "5.38.0",
					InstallPath: "/usr/bin/perl",
					Source:      "system",
				},
				{
					Version:     "5.40.0",
					InstallPath: "/home/user/.pvm/versions/5.40.0",
					Source:      "pvm",
				},
			}, nil
		}

		// Test resolving "system"
		resolved, err := resolveExplicitVersion("system", []string{"5.38.0", "5.40.0"}, nil)

		assert.NoError(t, err)
		assert.NotNil(t, resolved)
		assert.Equal(t, "5.38.0", resolved.Version)
		assert.Equal(t, SystemPerlSource, resolved.Source)
		assert.Equal(t, "/usr/bin/perl", resolved.Path)
	})

	t.Run("SystemIdentifierWhenNoSystemPerlImported", func(t *testing.T) {
		// Mock GetInstalledVersions to return no system Perl
		GetInstalledVersions = func() ([]VersionInfo, error) {
			return []VersionInfo{
				{
					Version:     "5.40.0",
					InstallPath: "/home/user/.pvm/versions/5.40.0",
					Source:      "pvm",
				},
			}, nil
		}

		// Test resolving "system" when no system Perl is imported
		resolved, err := resolveExplicitVersion("system", []string{"5.40.0"}, nil)

		assert.Error(t, err)
		assert.Nil(t, resolved)
		assert.Contains(t, err.Error(), "System Perl not found. Use 'pvm import-system' to import it first.")
	})

	t.Run("SystemIdentifierWithGetInstalledVersionsError", func(t *testing.T) {
		// Mock GetInstalledVersions to return an error
		GetInstalledVersions = func() ([]VersionInfo, error) {
			return nil, assert.AnError
		}

		// Test resolving "system" when GetInstalledVersions fails
		resolved, err := resolveExplicitVersion("system", []string{}, nil)

		assert.Error(t, err)
		assert.Nil(t, resolved)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("RegularVersionStillWorks", func(t *testing.T) {
		// Create a temp directory with perl binary structure
		tempDir := t.TempDir()
		binDir := filepath.Join(tempDir, "bin")
		err := os.MkdirAll(binDir, 0755)
		assert.NoError(t, err)

		// Create a fake perl binary (use .exe on Windows)
		perlName := "perl"
		if runtime.GOOS == "windows" {
			perlName = "perl.exe"
		}
		perlPath := filepath.Join(binDir, perlName)
		err = os.WriteFile(perlPath, []byte("#!/bin/sh\necho fake perl"), 0755)
		assert.NoError(t, err)

		// Reset mock for normal operation
		GetInstalledVersions = func() ([]VersionInfo, error) {
			return []VersionInfo{
				{
					Version:     "5.38.0",
					InstallPath: tempDir,
					Source:      "manual", // Regular version, not system
				},
			}, nil
		}

		// Mock LoadRegistry to return the version info
		LoadRegistry = func() (*VersionRegistry, error) {
			return &VersionRegistry{
				Versions: map[string]VersionInfo{
					"5.38.0": {
						Version:     "5.38.0",
						InstallPath: tempDir,
						Source:      "manual",
					},
				},
			}, nil
		}

		// Test resolving a regular version (non-system)
		resolved, err := resolveExplicitVersion("5.38.0", []string{"5.38.0"}, nil)

		assert.NoError(t, err)
		assert.NotNil(t, resolved)
		assert.Equal(t, "5.38.0", resolved.Version)
		assert.Equal(t, ExplicitVersion, resolved.Source) // Should be ExplicitVersion, not SystemPerlSource
	})

	t.Run("AliasResolutionStillWorks", func(t *testing.T) {
		// Create a temp directory with perl binary structure
		tempDir := t.TempDir()
		binDir := filepath.Join(tempDir, "bin")
		err := os.MkdirAll(binDir, 0755)
		assert.NoError(t, err)

		// Create a fake perl binary (use .exe on Windows)
		perlName := "perl"
		if runtime.GOOS == "windows" {
			perlName = "perl.exe"
		}
		perlPath := filepath.Join(binDir, perlName)
		err = os.WriteFile(perlPath, []byte("#!/bin/sh\necho fake perl"), 0755)
		assert.NoError(t, err)

		// Reset mock for normal operation
		GetInstalledVersions = func() ([]VersionInfo, error) {
			return []VersionInfo{}, nil
		}

		// Mock LoadRegistry to return the version info
		LoadRegistry = func() (*VersionRegistry, error) {
			return &VersionRegistry{
				Versions: map[string]VersionInfo{
					"5.38.0": {
						Version:     "5.38.0",
						InstallPath: tempDir,
						Source:      "manual",
					},
				},
			}, nil
		}

		cfg := &config.Config{
			PVM: &config.PVMConfig{
				VersionAliases: map[string]string{
					"stable": "5.38.0",
				},
			},
		}

		// Test resolving an alias
		resolved, err := resolveExplicitVersion("@stable", []string{"5.38.0"}, cfg)

		assert.NoError(t, err)
		assert.NotNil(t, resolved)
		assert.Equal(t, "5.38.0", resolved.Version)
		assert.Equal(t, ExplicitVersion, resolved.Source)
	})
}
