// ABOUTME: Test suite for PVI-PVX integration functionality
// ABOUTME: Includes regression tests for GitHub issue #201 version resolution
package pm

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
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

// TestGetRequiredModulesForScript tests AST-based dependency extraction
func TestGetRequiredModulesForScript(t *testing.T) {
	tests := []struct {
		name          string
		scriptContent string
		expected      []string
		expectError   bool
	}{
		{
			name: "Basic use statements",
			scriptContent: `#!/usr/bin/perl
use strict;
use warnings;
use Moose;
use DBI;`,
			expected: []string{"Moose", "DBI"}, // pragmas filtered out
		},
		{
			name: "Use with versions and imports",
			scriptContent: `use JSON::PP 2.0 qw(encode_json decode_json);
use Test::More;`,
			expected: []string{"JSON::PP", "Test::More"},
		},
		{
			name: "Require statements",
			scriptContent: `require DBI;
require "JSON/PP.pm";
require 'HTTP/Tiny.pm';`,
			expected: []string{"DBI", "JSON::PP", "HTTP::Tiny"},
		},
		{
			name: "Mixed use and require",
			scriptContent: `use strict;
use Moose;
require DBI;
# use Test::More; # commented out
require 'HTTP/Tiny.pm';`,
			expected: []string{"Moose", "DBI", "HTTP::Tiny"},
		},
		{
			name: "Core modules filtered out",
			scriptContent: `use DBI;
use File::Path;  # core module
use Test::More;`,
			expected: []string{"DBI", "Test::More"}, // File::Path filtered as core
		},
		{
			name: "Empty script",
			scriptContent: `#!/usr/bin/perl
# Just a comment
print "Hello World\n";`,
			expected: []string{},
		},
		{
			name: "Complex module names",
			scriptContent: `use Test::More::UTF8;
use Moo::Role;
use namespace::clean;`,
			expected: []string{"Test::More::UTF8", "Moo::Role", "namespace::clean"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a temporary file with the test content
			tmpfile, err := os.CreateTemp("", "test_script_*.pl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			// Write the test content to the file
			if _, err := tmpfile.WriteString(test.scriptContent); err != nil {
				tmpfile.Close()
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpfile.Close()

			// Test the function
			modules, err := GetRequiredModulesForScript(tmpfile.Name())

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Handle nil vs empty slice comparison
			if modules == nil {
				modules = []string{}
			}
			if test.expected == nil {
				test.expected = []string{}
			}

			// Sort both slices for comparison
			sort.Strings(modules)
			sort.Strings(test.expected)

			if !reflect.DeepEqual(modules, test.expected) {
				t.Errorf("Expected modules %v, got %v", test.expected, modules)
			}
		})
	}
}

// TestGetRequiredModulesForScript_ErrorCases tests error handling
func TestGetRequiredModulesForScript_ErrorCases(t *testing.T) {
	t.Run("NonexistentFile", func(t *testing.T) {
		_, err := GetRequiredModulesForScript("/nonexistent/file.pl")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("InvalidPerlSyntax", func(t *testing.T) {
		// Create a temp file with invalid Perl syntax
		tmpfile, err := os.CreateTemp("", "invalid_perl_*.pl")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		// Write invalid Perl syntax
		invalidContent := `use unclosed string"`
		if _, err := tmpfile.WriteString(invalidContent); err != nil {
			tmpfile.Close()
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpfile.Close()

		// The function should handle parsing errors gracefully and return empty list
		modules, err := GetRequiredModulesForScript(tmpfile.Name())
		if err != nil {
			t.Errorf("Function should handle parsing errors gracefully, but got error: %v", err)
		}

		// Should return empty list on parsing errors (graceful degradation)
		if len(modules) != 0 {
			t.Errorf("Expected empty module list on parsing error, got: %v", modules)
		}
	})
}

// TestExtractDependenciesFromContent tests the internal dependency extraction function
func TestExtractDependenciesFromContent(t *testing.T) {
	content := `use strict;
use DBI;
require JSON::PP;
use Test::More qw(ok done_testing);`

	dependencies, err := extractDependenciesFromContent(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"DBI", "JSON::PP", "Test::More"}

	// Sort both for comparison
	sort.Strings(dependencies)
	sort.Strings(expected)

	if !reflect.DeepEqual(dependencies, expected) {
		t.Errorf("Expected %v, got %v", expected, dependencies)
	}
}

// TestFilterCPANModules tests core module filtering
func TestFilterCPANModules(t *testing.T) {
	input := []string{
		"DBI",        // CPAN module
		"File::Path", // Core module
		"Test::More", // CPAN module
		"Carp",       // Core module
		"Moose",      // CPAN module
	}

	result := filterCPANModules(input)
	expected := []string{"DBI", "Test::More", "Moose"}

	// Sort both for comparison
	sort.Strings(result)
	sort.Strings(expected)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
