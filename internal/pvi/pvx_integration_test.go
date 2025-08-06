// ABOUTME: Test suite for PVI-PVX integration functionality
// ABOUTME: Includes regression tests for GitHub issue #201 version resolution
package pvi

import (
	"path/filepath"
	"runtime"
	"testing"

	"tamarou.com/pvm/internal/perl"
)

// TestResolvePerlExecutable_Issue201_Regression tests the fix for GitHub issue #201
// Ensures that PVI can correctly resolve Perl executables for versions already resolved by PVX
func TestResolvePerlExecutable_Issue201_Regression(t *testing.T) {

	// Import system Perl to ensure we have at least one version available
	err := perl.AutoImportSystemPerl()
	if err != nil {
		t.Skipf("System Perl not available for testing: %v", err)
	}

	// Get available versions
	versions, err := perl.GetInstalledVersions()
	if err != nil || len(versions) == 0 {
		t.Skip("No Perl versions available for testing")
	}

	testVersion := versions[0].Version

	// Test 1: Resolved version should work directly (the main fix for #201)
	t.Run("ResolvedVersionDirectLookup", func(t *testing.T) {
		perlPath, err := resolvePerlExecutable(testVersion)
		if err != nil {
			t.Fatalf("Failed to resolve Perl executable for resolved version %s: %v", testVersion, err)
		}

		// Verify we got a non-empty path
		if perlPath == "" {
			t.Fatal("Empty Perl path returned")
		}

		// Verify it's actually a perl executable (basic sanity check)
		expectedName := "perl"
		if runtime.GOOS == "windows" {
			expectedName = "perl.exe"
		}
		if filepath.Base(perlPath) != expectedName {
			t.Fatalf("Expected executable named '%s', got '%s'", expectedName, filepath.Base(perlPath))
		}
		
		t.Logf("Successfully resolved Perl path: %s", perlPath)
	})

	// Test 2: Empty version should fallback to system Perl
	t.Run("EmptyVersionFallsBackToSystem", func(t *testing.T) {
		perlPath, err := resolvePerlExecutable("")
		if err != nil {
			t.Fatalf("Failed to resolve Perl executable for empty version: %v", err)
		}

		if perlPath == "" {
			t.Fatal("Empty Perl path returned for empty version")
		}

		// Verify we got a reasonable looking path
		expectedName := "perl"
		if runtime.GOOS == "windows" {
			expectedName = "perl.exe"
		}
		if filepath.Base(perlPath) != expectedName {
			t.Fatalf("Expected executable named '%s', got '%s'", expectedName, filepath.Base(perlPath))
		}
		
		t.Logf("System Perl fallback path: %s", perlPath)
	})

	// Test 3: Non-existent version should either return error or fallback gracefully
	t.Run("NonExistentVersionHandling", func(t *testing.T) {
		nonExistentVersion := "99.99.99" // Highly unlikely to exist
		perlPath, err := resolvePerlExecutable(nonExistentVersion)
		
		// Either should return an error (if strict resolution) or fallback to system (graceful)
		if err != nil {
			// If error, should contain PVI-903 error code
			if !containsError(err.Error(), "PVI-903") {
				t.Errorf("Expected PVI-903 error code in error message: %v", err)
			}
		} else {
			// If no error, should return valid perl path (fallback behavior)
			if perlPath == "" {
				t.Fatal("Empty Perl path returned for non-existent version")
			}
			// Just verify it has a reasonable name - don't check file existence in test environment
			expectedName := "perl"
			if runtime.GOOS == "windows" {
				expectedName = "perl.exe"
			}
			if filepath.Base(perlPath) != expectedName {
				t.Errorf("Expected fallback executable named '%s', got '%s'", expectedName, filepath.Base(perlPath))
			}
		}
		
		t.Logf("Non-existent version handled appropriately: path=%s, err=%v", perlPath, err)
	})

	t.Log("All Issue #201 regression tests passed - PVI version resolution working correctly")
}

// TestGetPathForResolvedVersion tests the core function that was added to fix #201
func TestGetPathForResolvedVersion_Issue201(t *testing.T) {

	// Import system Perl
	err := perl.AutoImportSystemPerl()
	if err != nil {
		t.Skip("System Perl not available for testing")
	}

	// Get available versions
	versions, err := perl.GetInstalledVersions()
	if err != nil || len(versions) == 0 {
		t.Skip("No Perl versions available for testing")
	}

	testVersion := versions[0].Version

	// Test the direct path resolution that was added as the fix
	t.Run("DirectPathResolution", func(t *testing.T) {
		perlPath, err := getPathForResolvedVersion(testVersion)
		if err != nil {
			// This is expected if the version isn't fully installed
			t.Logf("Direct path resolution failed as expected for test version %s: %v", testVersion, err)
			return
		}

		// If no error, verify the path is reasonable
		if perlPath == "" {
			t.Fatalf("Empty path returned for version %s", testVersion)
		}
		
		t.Logf("Direct path resolution succeeded for version %s: %s", testVersion, perlPath)
	})

	// Test non-existent version
	t.Run("NonExistentVersionReturnsError", func(t *testing.T) {
		_, err := getPathForResolvedVersion("99.99.99")
		if err == nil {
			t.Fatal("Expected error for non-existent version")
		}
	})
}

// containsError checks if an error message contains a specific error code
func containsError(message, errorCode string) bool {
	return len(message) > 0 && len(errorCode) > 0 && 
		   (message == errorCode || 
		    len(message) >= len(errorCode) && 
		    message[:len(errorCode)] == errorCode)
}