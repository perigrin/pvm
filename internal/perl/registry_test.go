// ABOUTME: Tests for Perl installation registry
// ABOUTME: Tests registration, listing, and uninstallation of Perl versions

package perl

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/xdg"
)

// These variables are already declared in registry.go
// We reference them here for mocking in tests

// setupTestRegistry creates a temporary registry for testing
func setupTestRegistry(t *testing.T) (*VersionRegistry, string, func()) {
	// Create a temporary directory for the test registry
	tempDir, err := os.MkdirTemp("", "pvm-registry-test")
	require.NoError(t, err)

	// Create an empty registry
	registry := &VersionRegistry{
		Versions: make(map[string]VersionInfo),
	}

	// Create a mock registry file path
	registryPath := filepath.Join(tempDir, registryFileName)

	// Replace the LoadRegistry and SaveRegistry functions with test versions
	originalLoadRegistry := LoadRegistry
	originalSaveRegistry := SaveRegistry

	// Set up mock load function that returns our test registry
	LoadRegistry = func() (*VersionRegistry, error) {
		return registry, nil
	}

	// Set up mock save function that updates our test registry
	SaveRegistry = func(r *VersionRegistry) error {
		// Save the changes back to our test registry
		registry = r

		// Also write to file for checking file operations
		data, err := json.MarshalIndent(registry, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(registryPath, data, 0644)
	}

	// Return cleanup function
	cleanup := func() {
		LoadRegistry = originalLoadRegistry
		SaveRegistry = originalSaveRegistry
		_ = os.RemoveAll(tempDir)
	}

	return registry, registryPath, cleanup
}

// TestRegisterVersion tests registering a new version
func TestRegisterVersion(t *testing.T) {
	// Set up test registry
	registry, registryPath, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Create version info
	versionInfo := VersionInfo{
		Version:     "5.38.0",
		InstallPath: "/opt/perl/5.38.0",
		InstallTime: time.Now(),
		Source:      "pvm",
	}

	// Register the version
	err := RegisterVersion(versionInfo)
	require.NoError(t, err)

	// Verify registry was updated
	assert.Contains(t, registry.Versions, "5.38.0")
	assert.Equal(t, versionInfo.InstallPath, registry.Versions["5.38.0"].InstallPath)
	assert.Equal(t, versionInfo.Source, registry.Versions["5.38.0"].Source)

	// Verify file was written
	data, err := os.ReadFile(registryPath)
	require.NoError(t, err)

	var fileRegistry VersionRegistry
	err = json.Unmarshal(data, &fileRegistry)
	require.NoError(t, err)

	assert.Contains(t, fileRegistry.Versions, "5.38.0")
}

// TestRegisterVersionNormalization tests version normalization during registration
func TestRegisterVersionNormalization(t *testing.T) {
	// Set up test registry
	registry, _, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Create version info with unnormalized version string
	versionInfo := VersionInfo{
		Version:     "v5.38", // Missing patch version, has 'v' prefix
		InstallPath: "/opt/perl/5.38.0",
		InstallTime: time.Now(),
		Source:      "pvm",
	}

	// Register the version
	err := RegisterVersion(versionInfo)
	require.NoError(t, err)

	// Verify registry was updated with normalized version (5.38.0)
	assert.Contains(t, registry.Versions, "5.38.0")
	assert.NotContains(t, registry.Versions, "v5.38")
}

// TestRegisterVersionAlreadyExists tests handling of duplicate version registration
func TestRegisterVersionAlreadyExists(t *testing.T) {
	// Set up test registry
	_, _, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Create version info
	versionInfo := VersionInfo{
		Version:     "5.38.0",
		InstallPath: "/opt/perl/5.38.0",
		InstallTime: time.Now(),
		Source:      "pvm",
	}

	// Register the version
	err := RegisterVersion(versionInfo)
	require.NoError(t, err)

	// Try to register the same version again
	err = RegisterVersion(versionInfo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

// TestRegisterVersionInvalidVersion tests handling of invalid version strings
func TestRegisterVersionInvalidVersion(t *testing.T) {
	// Set up test registry
	_, _, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Create version info with invalid version
	versionInfo := VersionInfo{
		Version:     "not-a-version",
		InstallPath: "/opt/perl/5.38.0",
		InstallTime: time.Now(),
		Source:      "pvm",
	}

	// Register the version (should fail)
	err := RegisterVersion(versionInfo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid version")
}

// TestRegisterVersionErrorPaths tests all error paths in RegisterVersion
func TestRegisterVersionErrorPaths(t *testing.T) {
	// Save original functions to restore after the test
	originalLoadRegistry := LoadRegistry
	originalSaveRegistry := SaveRegistry
	originalParseVersion := ParseVersion
	defer func() {
		LoadRegistry = originalLoadRegistry
		SaveRegistry = originalSaveRegistry
		ParseVersion = originalParseVersion
	}()

	// Test 1: ParseVersion fails
	t.Run("ParseVersionFails", func(t *testing.T) {
		// Already covered in TestRegisterVersionInvalidVersion
	})

	// Test 2: LoadRegistry fails
	t.Run("LoadRegistryFails", func(t *testing.T) {
		// Mock ParseVersion to succeed
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{Major: 5, Minor: 38, Patch: 0}, nil
		}

		// Mock LoadRegistry to fail
		LoadRegistry = func() (*VersionRegistry, error) {
			return nil, errors.New("mock LoadRegistry error")
		}

		// Test registration
		versionInfo := VersionInfo{
			Version:     "5.38.0",
			InstallPath: "/opt/perl/5.38.0",
			InstallTime: time.Now(),
			Source:      "pvm",
		}
		err := RegisterVersion(versionInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock LoadRegistry error")
	})

	// Test 3: SaveRegistry fails
	t.Run("SaveRegistryFails", func(t *testing.T) {
		// Mock ParseVersion to succeed
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{Major: 5, Minor: 38, Patch: 0}, nil
		}

		// Mock LoadRegistry to succeed
		LoadRegistry = func() (*VersionRegistry, error) {
			return &VersionRegistry{Versions: make(map[string]VersionInfo)}, nil
		}

		// Mock SaveRegistry to fail
		SaveRegistry = func(r *VersionRegistry) error {
			return errors.New("mock SaveRegistry error")
		}

		// Test registration
		versionInfo := VersionInfo{
			Version:     "5.38.0",
			InstallPath: "/opt/perl/5.38.0",
			InstallTime: time.Now(),
			Source:      "pvm",
		}
		err := RegisterVersion(versionInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock SaveRegistry error")
	})

	// Test 4: Version already exists
	t.Run("VersionAlreadyExists", func(t *testing.T) {
		// Mock ParseVersion to succeed
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{Major: 5, Minor: 38, Patch: 0}, nil
		}

		// Mock LoadRegistry to return registry with the version already in it
		LoadRegistry = func() (*VersionRegistry, error) {
			registry := &VersionRegistry{
				Versions: map[string]VersionInfo{
					"5.38.0": {
						Version:     "5.38.0",
						InstallPath: "/opt/perl/5.38.0",
						InstallTime: time.Now(),
						Source:      "pvm",
					},
				},
			}
			return registry, nil
		}

		// Test registration
		versionInfo := VersionInfo{
			Version:     "5.38.0",
			InstallPath: "/opt/perl/5.38.0",
			InstallTime: time.Now(),
			Source:      "pvm",
		}
		err := RegisterVersion(versionInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

// TestRegisterVersionAfterBuild tests registering a version after a successful build
func TestRegisterVersionAfterBuild(t *testing.T) {
	// Set up test registry
	registry, _, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Create a mock build result
	buildResult := &BuildResult{
		Version:     "5.38.0",
		InstallPath: "/opt/perl/5.38.0",
		BuildPath:   "/tmp/build/perl-5.38.0",
		Duration:    time.Minute * 5,
	}

	// Register the version
	err := RegisterVersionAfterBuild(buildResult, "pvm")
	require.NoError(t, err)

	// Verify registry was updated
	assert.Contains(t, registry.Versions, "5.38.0")
	assert.Equal(t, buildResult.InstallPath, registry.Versions["5.38.0"].InstallPath)
	assert.Equal(t, "pvm", registry.Versions["5.38.0"].Source)
}

// TestGetInstalledVersions tests listing installed versions
func TestGetInstalledVersions(t *testing.T) {
	// Set up test registry
	registry, _, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Add a few versions
	versions := []VersionInfo{
		{
			Version:     "5.36.0",
			InstallPath: "/opt/perl/5.36.0",
			InstallTime: time.Now().Add(-48 * time.Hour),
			Source:      "pvm",
		},
		{
			Version:     "5.38.0",
			InstallPath: "/opt/perl/5.38.0",
			InstallTime: time.Now(),
			Source:      "pvm",
		},
		{
			Version:     "5.34.1",
			InstallPath: "/opt/perl/5.34.1",
			InstallTime: time.Now().Add(-24 * time.Hour),
			Source:      "pvm",
		},
	}

	for _, v := range versions {
		registry.Versions[v.Version] = v
	}

	// Get installed versions
	installedVersions, err := GetInstalledVersions()
	require.NoError(t, err)

	// Verify we get all versions
	assert.Len(t, installedVersions, 3)

	// Verify versions are sorted by version number (descending)
	assert.Equal(t, "5.38.0", installedVersions[0].Version)
	assert.Equal(t, "5.36.0", installedVersions[1].Version)
	assert.Equal(t, "5.34.1", installedVersions[2].Version)
}

// TestGetInstalledVersionsErrorPaths tests error paths in GetInstalledVersions
func TestGetInstalledVersionsErrorPaths(t *testing.T) {
	// Save original function to restore later
	originalLoadRegistry := LoadRegistry
	defer func() {
		LoadRegistry = originalLoadRegistry
	}()

	// Test case: LoadRegistry fails
	t.Run("LoadRegistryFails", func(t *testing.T) {
		// Mock LoadRegistry to fail
		LoadRegistry = func() (*VersionRegistry, error) {
			return nil, errors.New("mock LoadRegistry error")
		}

		// Call function - should return error
		versions, err := GetInstalledVersions()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock LoadRegistry error")
		assert.Nil(t, versions)
	})
}

// TestLoadRegistryFullCoverage tests all code paths in loadRegistryFunc
func TestLoadRegistryFullCoverage(t *testing.T) {
	// Save original function to restore later
	originalLoadRegistry := LoadRegistry
	originalGetDirs := xdg.GetDirs
	originalStat := osStat
	originalReadFile := ioutilReadFile
	defer func() {
		LoadRegistry = originalLoadRegistry
		xdg.GetDirs = originalGetDirs
		osStat = originalStat
		ioutilReadFile = originalReadFile
	}()

	// Test 1: GetDirs fails
	t.Run("GetDirsFails", func(t *testing.T) {
		// Mock GetDirs to return an error
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			return nil, errors.New("mock GetDirs error")
		}

		// Temporarily restore loadRegistryFunc to invoke it directly
		LoadRegistry = loadRegistryFunc

		// Call function - should return an empty registry with error
		reg, err := LoadRegistry()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to determine XDG directories")
		assert.NotNil(t, reg) // Should still return an initialized registry
		assert.Empty(t, reg.Versions)
	})

	// Test 2: EnsureDirs fails
	t.Run("EnsureDirsFails", func(t *testing.T) {
		// Mock GetDirs to return dirs with failing EnsureDirs
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			dirs := &xdg.Dirs{
				DataDir: "/mock/data/dir",
			}
			// Replace the EnsureDirs method
			dirs.EnsureDirs = func() error {
				return errors.New("mock EnsureDirs error")
			}
			return dirs, nil
		}

		// Temporarily restore loadRegistryFunc to invoke it directly
		LoadRegistry = loadRegistryFunc

		// Call function - should return an empty registry with error
		reg, err := LoadRegistry()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to create required directories")
		assert.NotNil(t, reg)
		assert.Empty(t, reg.Versions)
	})

	// Test 3: Registry file doesn't exist (coverage already provided in other tests)

	// Test 4: Registry file exists but ReadFile fails
	t.Run("ReadFileFails", func(t *testing.T) {
		// Set up mocks
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			dirs := &xdg.Dirs{
				DataDir: "/mock/data/dir",
			}
			dirs.EnsureDirs = func() error { return nil }
			return dirs, nil
		}

		// Mock stat to say file exists
		osStat = func(name string) (os.FileInfo, error) {
			return nil, nil // File exists
		}

		// Mock read file to fail
		ioutilReadFile = func(filename string) ([]byte, error) {
			return nil, errors.New("mock read file error")
		}

		// Temporarily restore loadRegistryFunc to invoke it directly
		LoadRegistry = loadRegistryFunc

		// Call function - should return an empty registry with error
		reg, err := LoadRegistry()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to read registry file")
		assert.NotNil(t, reg)
		assert.Empty(t, reg.Versions)
	})

	// Test 5: Registry file exists but contains invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		// Set up mocks
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			dirs := &xdg.Dirs{
				DataDir: "/mock/data/dir",
			}
			dirs.EnsureDirs = func() error { return nil }
			return dirs, nil
		}

		// Mock stat to say file exists
		osStat = func(name string) (os.FileInfo, error) {
			return nil, nil // File exists
		}

		// Mock read file to return invalid JSON
		ioutilReadFile = func(filename string) ([]byte, error) {
			return []byte("this is not valid JSON"), nil
		}

		// Temporarily restore loadRegistryFunc to invoke it directly
		LoadRegistry = loadRegistryFunc

		// Call function - should return an empty registry with error
		reg, err := LoadRegistry()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to parse registry file")
		assert.NotNil(t, reg)
		assert.Empty(t, reg.Versions)
	})

	// Test 6: Registry file exists and contains valid JSON
	t.Run("ValidJSON", func(t *testing.T) {
		// Set up mocks
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			dirs := &xdg.Dirs{
				DataDir: "/mock/data/dir",
			}
			dirs.EnsureDirs = func() error { return nil }
			return dirs, nil
		}

		// Mock stat to say file exists
		osStat = func(name string) (os.FileInfo, error) {
			return nil, nil // File exists
		}

		// Mock read file to return valid JSON
		expectedRegistry := &VersionRegistry{
			Versions: map[string]VersionInfo{
				"5.38.0": {
					Version:     "5.38.0",
					InstallPath: "/opt/perl/5.38.0",
					InstallTime: time.Now(),
					Source:      "pvm",
				},
			},
		}
	// Create a version without build options for serialization
	simpleRegistry := &VersionRegistry{
		Versions: make(map[string]VersionInfo),
	}
	
	// Copy only the basic fields, explicitly setting BuildOptions to nil
	for k, v := range expectedRegistry.Versions {
		versionCopy := VersionInfo{
			Version:     v.Version,
			InstallPath: v.InstallPath,
			InstallTime: v.InstallTime,
			Source:      v.Source,
			BuildOptions: nil, // Explicitly set to nil to avoid marshaling issues
		}
		simpleRegistry.Versions[k] = versionCopy
	}
	
	// nolint:staticcheck // SA1026: We're deliberately excluding BuildOptions which contains the problematic BuildProgressCallback
	regJSON, _ := json.Marshal(simpleRegistry)
	ioutilReadFile = func(filename string) ([]byte, error) {
			return regJSON, nil
		}

		// Temporarily restore loadRegistryFunc to invoke it directly
		LoadRegistry = loadRegistryFunc

		// Call function - should return the registry with the expected version
		reg, err := LoadRegistry()
		assert.NoError(t, err)
		assert.NotNil(t, reg)
		assert.Contains(t, reg.Versions, "5.38.0")
		assert.Equal(t, "/opt/perl/5.38.0", reg.Versions["5.38.0"].InstallPath)
	})
}

// TestSaveRegistryFullCoverage tests all code paths in saveRegistryFunc
func TestSaveRegistryFullCoverage(t *testing.T) {
	// Save original function to restore later
	originalSaveRegistry := SaveRegistry
	originalGetDirs := xdg.GetDirs
	originalWriteFile := ioutilWriteFile
	originalRename := osRename
	defer func() {
		SaveRegistry = originalSaveRegistry
		xdg.GetDirs = originalGetDirs
		ioutilWriteFile = originalWriteFile
		osRename = originalRename
	}()

	// Test 1: GetDirs fails
	t.Run("GetDirsFails", func(t *testing.T) {
		// Mock GetDirs to return an error
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			return nil, errors.New("mock GetDirs error")
		}

		// Temporarily restore saveRegistryFunc to invoke it directly
		SaveRegistry = saveRegistryFunc

		// Call function - should return an error
		registry := &VersionRegistry{
			Versions: make(map[string]VersionInfo),
		}
		err := SaveRegistry(registry)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to determine XDG directories")
	})

	// Test 2: EnsureDirs fails
	t.Run("EnsureDirsFails", func(t *testing.T) {
		// Mock GetDirs to return dirs with failing EnsureDirs
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			dirs := &xdg.Dirs{
				DataDir: "/mock/data/dir",
			}
			// Replace the EnsureDirs method
			dirs.EnsureDirs = func() error {
				return errors.New("mock EnsureDirs error")
			}
			return dirs, nil
		}

		// Temporarily restore saveRegistryFunc to invoke it directly
		SaveRegistry = saveRegistryFunc

		// Call function - should return an error
		registry := &VersionRegistry{
			Versions: make(map[string]VersionInfo),
		}
		err := SaveRegistry(registry)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to create required directories")
	})

	// Test 3: MarshalIndent fails (difficult to test, skipping)

	// Test 4: WriteFile fails
	t.Run("WriteFileFails", func(t *testing.T) {
		// Set up mocks
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			dirs := &xdg.Dirs{
				DataDir: "/mock/data/dir",
			}
			dirs.EnsureDirs = func() error { return nil }
			return dirs, nil
		}

		// Mock write file to fail
		ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
			return errors.New("mock write file error")
		}

		// Temporarily restore saveRegistryFunc to invoke it directly
		SaveRegistry = saveRegistryFunc

		// Call function - should return an error
		registry := &VersionRegistry{
			Versions: make(map[string]VersionInfo),
		}
		err := SaveRegistry(registry)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to write registry file")
	})

	// Test 5: Rename fails
	t.Run("RenameFails", func(t *testing.T) {
		// Set up mocks
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			dirs := &xdg.Dirs{
				DataDir: "/mock/data/dir",
			}
			dirs.EnsureDirs = func() error { return nil }
			return dirs, nil
		}

		// Mock write file to succeed
		ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
			return nil
		}

		// Mock rename to fail
		osRename = func(oldpath, newpath string) error {
			return errors.New("mock rename error")
		}

		// Temporarily restore saveRegistryFunc to invoke it directly
		SaveRegistry = saveRegistryFunc

		// Call function - should return an error
		registry := &VersionRegistry{
			Versions: make(map[string]VersionInfo),
		}
		err := SaveRegistry(registry)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to update registry file")
	})

	// Test 6: Everything succeeds
	t.Run("SuccessfulSave", func(t *testing.T) {
		// Set up mocks
		xdg.GetDirs = func() (*xdg.Dirs, error) {
			dirs := &xdg.Dirs{
				DataDir: "/mock/data/dir",
			}
			dirs.EnsureDirs = func() error { return nil }
			return dirs, nil
		}

		// Mock write file to succeed
		ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
			return nil
		}

		// Mock rename to succeed
		osRename = func(oldpath, newpath string) error {
			return nil
		}

		// Temporarily restore saveRegistryFunc to invoke it directly
		SaveRegistry = saveRegistryFunc

		// Call function - should not return an error
		registry := &VersionRegistry{
			Versions: map[string]VersionInfo{
				"5.38.0": {
					Version:     "5.38.0",
					InstallPath: "/opt/perl/5.38.0",
					InstallTime: time.Now(),
					Source:      "pvm",
				},
			},
		}
		err := SaveRegistry(registry)
		assert.NoError(t, err)
	})
}

// TestGetVersionInfo tests getting information about a specific version
func TestGetVersionInfo(t *testing.T) {
	// Set up test registry
	registry, _, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Add a test version
	expectedInfo := VersionInfo{
		Version:     "5.38.0",
		InstallPath: "/opt/perl/5.38.0",
		InstallTime: time.Now(),
		Source:      "pvm",
	}
	registry.Versions["5.38.0"] = expectedInfo

	// Get the version info
	versionInfo, err := GetVersionInfo("5.38.0")
	require.NoError(t, err)

	// Verify we got the right version
	assert.Equal(t, expectedInfo.Version, versionInfo.Version)
	assert.Equal(t, expectedInfo.InstallPath, versionInfo.InstallPath)
	assert.Equal(t, expectedInfo.Source, versionInfo.Source)

	// Test with version normalization (v5.38)
	versionInfo, err = GetVersionInfo("v5.38")
	require.NoError(t, err)
	assert.Equal(t, "5.38.0", versionInfo.Version)

	// Test with non-existent version
	_, err = GetVersionInfo("5.40.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

// TestIsVersionInstalled tests checking if a version is installed
func TestIsVersionInstalled(t *testing.T) {
	// Set up test registry
	registry, _, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Add a test version
	registry.Versions["5.38.0"] = VersionInfo{
		Version:     "5.38.0",
		InstallPath: "/opt/perl/5.38.0",
		InstallTime: time.Now(),
		Source:      "pvm",
	}

	// Test installed version
	installed, err := IsVersionInstalled("5.38.0")
	require.NoError(t, err)
	assert.True(t, installed)

	// Test with version normalization
	installed, err = IsVersionInstalled("v5.38")
	require.NoError(t, err)
	assert.True(t, installed)

	// Test non-installed version
	installed, err = IsVersionInstalled("5.40.0")
	require.NoError(t, err)
	assert.False(t, installed)

	// Test invalid version
	_, err = IsVersionInstalled("not-a-version")
	assert.Error(t, err)
}
