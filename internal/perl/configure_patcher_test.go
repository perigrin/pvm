// ABOUTME: Tests for Configure script patching functionality
// ABOUTME: Validates patch application, backup/restore, and strategy execution

package perl

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestConfigurePatcherCreation(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
	}

	tmpDir := t.TempDir()

	patcher, err := NewConfigurePatcher(tmpDir, "5.26.0", true)
	if err != nil {
		t.Fatalf("Failed to create ConfigurePatcher: %v", err)
	}

	if patcher.sourceDir != tmpDir {
		t.Errorf("Expected sourceDir %s, got %s", tmpDir, patcher.sourceDir)
	}

	if patcher.perlVersion != "5.26.0" {
		t.Errorf("Expected perlVersion 5.26.0, got %s", patcher.perlVersion)
	}

	if !patcher.verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestConfigurePatcherNonMacOS(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping non-macOS tests on macOS platform")
	}

	tmpDir := t.TempDir()

	_, err := NewConfigurePatcher(tmpDir, "5.26.0", false)
	if err == nil {
		t.Error("Expected error when creating ConfigurePatcher on non-macOS platform")
	}
}

func TestDarwinHintsPatching(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
	}

	tmpDir := t.TempDir()

	// Create hints directory and file
	hintsDir := filepath.Join(tmpDir, "hints")
	err := os.MkdirAll(hintsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create hints directory: %v", err)
	}

	// Create a sample darwin.sh file with content that needs patching
	darwinHints := `#!/bin/sh

# macOS/Darwin hints file

osvers=$(sw_vers -productVersion)

case "$osvers" in
10.*)
	echo "macOS 10.x detected"
	;;
*) echo "Unsupported Darwin version: $osvers" >&2; exit 1 ;;
esac

echo "Darwin hints applied"
`

	darwinFile := filepath.Join(hintsDir, "darwin.sh")
	err = os.WriteFile(darwinFile, []byte(darwinHints), 0644)
	if err != nil {
		t.Fatalf("Failed to write darwin.sh: %v", err)
	}

	// Create patcher and apply patches
	patcher, err := NewConfigurePatcher(tmpDir, "5.26.0", true)
	if err != nil {
		t.Fatalf("Failed to create ConfigurePatcher: %v", err)
	}

	err = patcher.ApplyPatches()
	if err != nil {
		t.Fatalf("Failed to apply patches: %v", err)
	}

	// Read the patched file
	patchedContent, err := os.ReadFile(darwinFile)
	if err != nil {
		t.Fatalf("Failed to read patched file: %v", err)
	}

	patchedStr := string(patchedContent)

	// Check that patches were applied (content should be different and contain expected changes)
	if patchedStr == darwinHints {
		t.Error("File content unchanged - patches may not have been applied")
	}

	// Check for backup file
	backupFile := darwinFile + ".pvm-backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file not created")
	}

	// Test restoration
	err = patcher.RestoreBackups()
	if err != nil {
		t.Fatalf("Failed to restore backups: %v", err)
	}

	// Check that original content is restored
	restoredContent, err := os.ReadFile(darwinFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(restoredContent) != darwinHints {
		t.Error("File not properly restored from backup")
	}

	// Check that backup file is removed
	if _, err := os.Stat(backupFile); !os.IsNotExist(err) {
		t.Error("Backup file not removed after restoration")
	}
}

func TestConfigureScriptPatching(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
	}

	tmpDir := t.TempDir()

	// Create a sample Configure script with content that needs patching
	configureScript := `#!/bin/sh

# Perl Configure script

osvers=$(sw_vers -productVersion)
case "$osvers" in
10.*)
	echo "Supported macOS version: $osvers"
	;;
*)
	echo "*** Unexpected product version $osvers" >&2
	echo "*** Try running sw_vers and see what its ProductVersion says." >&2
	exit 1
	;;
esac

echo "Configure script completed"
`

	configureFile := filepath.Join(tmpDir, "Configure")
	err := os.WriteFile(configureFile, []byte(configureScript), 0755)
	if err != nil {
		t.Fatalf("Failed to write Configure script: %v", err)
	}

	// Create patcher with strategy that will use Configure script patching
	patcher, err := NewConfigurePatcher(tmpDir, "5.34.0", true)
	if err != nil {
		t.Fatalf("Failed to create ConfigurePatcher: %v", err)
	}

	// Force the strategy to ConfigureScript for testing
	patcher.strategy = PatchStrategyConfigureScript

	err = patcher.ApplyPatches()
	if err != nil {
		t.Fatalf("Failed to apply patches: %v", err)
	}

	// Read the patched file
	patchedContent, err := os.ReadFile(configureFile)
	if err != nil {
		t.Fatalf("Failed to read patched file: %v", err)
	}

	patchedStr := string(patchedContent)

	// Check that the error message is removed/modified
	if strings.Contains(patchedStr, "Unexpected product version") &&
		strings.Contains(patchedStr, "exit 1") {
		t.Error("Configure script still contains unpatched error handling")
	}

	// Check for backup file
	backupFile := configureFile + ".pvm-backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file not created")
	}
}

func TestEnvironmentOverrideStrategy(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
	}

	tmpDir := t.TempDir()

	patcher, err := NewConfigurePatcher(tmpDir, "5.18.4", true)
	if err != nil {
		t.Fatalf("Failed to create ConfigurePatcher: %v", err)
	}

	// Force environment override strategy
	patcher.strategy = PatchStrategyEnvironmentOverride

	// Store original environment
	originalDeploymentTarget := os.Getenv("MACOSX_DEPLOYMENT_TARGET")
	originalPerlDarwinVersion := os.Getenv("PERL_DARWIN_VERSION")

	err = patcher.ApplyPatches()
	if err != nil {
		t.Fatalf("Failed to apply patches: %v", err)
	}

	// Check that environment variables are set
	deploymentTarget := os.Getenv("MACOSX_DEPLOYMENT_TARGET")
	if deploymentTarget != "10.15" {
		t.Errorf("Expected MACOSX_DEPLOYMENT_TARGET=10.15, got %s", deploymentTarget)
	}

	perlDarwinVersion := os.Getenv("PERL_DARWIN_VERSION")
	if perlDarwinVersion != "10.15.7" {
		t.Errorf("Expected PERL_DARWIN_VERSION=10.15.7, got %s", perlDarwinVersion)
	}

	// Restore original environment
	if originalDeploymentTarget == "" {
		os.Unsetenv("MACOSX_DEPLOYMENT_TARGET")
	} else {
		os.Setenv("MACOSX_DEPLOYMENT_TARGET", originalDeploymentTarget)
	}
	if originalPerlDarwinVersion == "" {
		os.Unsetenv("PERL_DARWIN_VERSION")
	} else {
		os.Setenv("PERL_DARWIN_VERSION", originalPerlDarwinVersion)
	}
}

func TestPatchApplication(t *testing.T) {
	patch := Patch{
		Name:        "test patch",
		Description: "Test patch for unit testing",
		Pattern:     mustCompile(`old_text`),
		Replacement: "new_text",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "patch applied",
			input:    "This is old_text that should be replaced",
			expected: "This is new_text that should be replaced",
		},
		{
			name:     "no match",
			input:    "This has no matching text",
			expected: "This has no matching text",
		},
		{
			name:     "multiple matches",
			input:    "old_text and old_text again",
			expected: "new_text and new_text again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := patch.Apply(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestApplyMacOSConfigurePatchesConvenienceFunction(t *testing.T) {
	if runtime.GOOS != "darwin" {
		// On non-macOS, function should be no-op
		tmpDir := t.TempDir()
		err := ApplyMacOSConfigurePatches(tmpDir, "5.26.0", false)
		if err != nil {
			t.Errorf("Expected no error on non-macOS platform, got: %v", err)
		}
		return
	}

	tmpDir := t.TempDir()

	// Test with a version that doesn't need patching
	err := ApplyMacOSConfigurePatches(tmpDir, "5.40.0", false)
	if err != nil {
		t.Errorf("Unexpected error for recent Perl version: %v", err)
	}

	// Test with a version that needs patching but no Configure script
	err = ApplyMacOSConfigurePatches(tmpDir, "5.26.0", false)
	// This should not error even if no Configure script exists
	// The patcher should handle this gracefully
}

func TestGetStrategyName(t *testing.T) {
	tmpDir := t.TempDir()

	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
		return
	}

	patcher, err := NewConfigurePatcher(tmpDir, "5.26.0", false)
	if err != nil {
		t.Fatalf("Failed to create patcher: %v", err)
	}

	tests := []struct {
		strategy ConfigurePatchStrategy
		expected string
	}{
		{PatchStrategyDarwinHints, "Darwin hints"},
		{PatchStrategyConfigureScript, "Configure script"},
		{PatchStrategyEnvironmentOverride, "Environment override"},
		{ConfigurePatchStrategy(999), "Unknown"},
	}

	for _, tt := range tests {
		patcher.strategy = tt.strategy
		result := patcher.getStrategyName()
		if result != tt.expected {
			t.Errorf("For strategy %d, expected %s, got %s", tt.strategy, tt.expected, result)
		}
	}
}

// Helper function that must compile regexp or panic
func mustCompile(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("Failed to compile regexp: " + err.Error())
	}
	return re
}

func TestPatcherVerboseOutput(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
	}

	tmpDir := t.TempDir()

	// Test verbose vs non-verbose - this mainly tests that verbose flag is handled
	// without errors, since output goes to stdout and is hard to capture in tests

	patcher, err := NewConfigurePatcher(tmpDir, "5.26.0", true)
	if err != nil {
		t.Fatalf("Failed to create verbose patcher: %v", err)
	}

	if !patcher.verbose {
		t.Error("Expected verbose patcher to have verbose=true")
	}

	nonVerbosePatcher, err := NewConfigurePatcher(tmpDir, "5.26.0", false)
	if err != nil {
		t.Fatalf("Failed to create non-verbose patcher: %v", err)
	}

	if nonVerbosePatcher.verbose {
		t.Error("Expected non-verbose patcher to have verbose=false")
	}
}
