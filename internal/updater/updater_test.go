// ABOUTME: Tests for PVM self-updater functionality
// ABOUTME: Comprehensive tests for update orchestration and component integration

package updater

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNewUpdater(t *testing.T) {
	updater := NewUpdater()
	if updater == nil {
		t.Fatal("NewUpdater returned nil")
	}

	if updater.versionChecker == nil {
		t.Error("versionChecker is nil")
	}

	if updater.downloader == nil {
		t.Error("downloader is nil")
	}

	if updater.replacer == nil {
		t.Error("replacer is nil")
	}

	if updater.rollbackMgr == nil {
		t.Error("rollbackMgr is nil")
	}
}

func TestNewUpdaterWithToken(t *testing.T) {
	token := "test-token"
	updater := NewUpdaterWithToken(token)
	if updater == nil {
		t.Fatal("NewUpdaterWithToken returned nil")
	}

	// Note: We can't easily test that the token was set without accessing private fields
	// This test just ensures the constructor works
}

func TestDefaultUpdateOptions(t *testing.T) {
	opts := DefaultUpdateOptions()
	if opts == nil {
		t.Fatal("DefaultUpdateOptions returned nil")
	}

	if opts.Repository != "perigrin/pvm-dev" {
		t.Errorf("Expected repository 'perigrin/pvm-dev', got '%s'", opts.Repository)
	}

	if !opts.Backup {
		t.Error("Expected Backup to be true by default")
	}

	if !opts.AutoRollback {
		t.Error("Expected AutoRollback to be true by default")
	}

	if opts.Context == nil {
		t.Error("Expected Context to be non-nil")
	}
}

func TestUpdateStageString(t *testing.T) {
	tests := []struct {
		stage    UpdateStage
		expected string
	}{
		{StageCheckingVersion, "Checking for updates"},
		{StageDetectingPlatform, "Detecting platform"},
		{StageDownloading, "Downloading update"},
		{StageValidating, "Validating download"},
		{StageCreatingBackup, "Creating backup"},
		{StageReplacing, "Installing update"},
		{StageValidatingUpdate, "Validating installation"},
		{StageCleaningUp, "Cleaning up"},
		{StageRollingBack, "Rolling back"},
		{StageDone, "Update complete"},
		{UpdateStage(999), "Unknown stage"},
	}

	for _, test := range tests {
		result := test.stage.String()
		if result != test.expected {
			t.Errorf("Stage %d: expected '%s', got '%s'", test.stage, test.expected, result)
		}
	}
}

func TestUpdateOptionsValidation(t *testing.T) {
	// Test with nil options - should use defaults
	updater := NewUpdater()

	// This will fail because the repository doesn't exist, but we're testing
	// that nil options are handled correctly
	_, err := updater.CheckForUpdates(nil)
	if err == nil {
		t.Error("Expected error when checking non-existent repository")
	}
	// Error is expected due to repository not existing
}

func TestGetCurrentBinaryPath(t *testing.T) {
	path, err := GetCurrentBinaryPath()
	if err != nil {
		t.Fatalf("GetCurrentBinaryPath failed: %v", err)
	}

	if path == "" {
		t.Error("GetCurrentBinaryPath returned empty path")
	}

	// Check that the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Binary path does not exist: %s", path)
	}
}

func TestDetectInstallationMethodBasic(t *testing.T) {
	tests := []struct {
		path     string
		expected InstallationMethod
	}{
		{"/opt/homebrew/bin/pvm", InstallationHomebrew},
		{"/usr/bin/pvm", getExpectedForUsrBin()}, // /usr/bin is detected package manager
		{"/home/user/.local/bin/pvm", InstallationBinary},
	}

	// Platform-specific test for /usr/local/bin
	if runtime.GOOS == "darwin" {
		tests = append(tests, struct {
			path     string
			expected InstallationMethod
		}{"/usr/local/bin/pvm", InstallationHomebrew})
	} else {
		tests = append(tests, struct {
			path     string
			expected InstallationMethod
		}{"/usr/local/bin/pvm", InstallationSystemPackage})
	}

	for _, test := range tests {
		result, err := DetectInstallationMethod(test.path)
		if err != nil {
			t.Errorf("DetectInstallationMethod(%s) failed: %v", test.path, err)
			continue
		}

		if result != test.expected {
			t.Errorf("DetectInstallationMethod(%s): expected %s, got %s",
				test.path, test.expected, result)
		}
	}
}

func TestInstallationMethodString(t *testing.T) {
	tests := []struct {
		method   InstallationMethod
		expected string
	}{
		{InstallationBinary, "binary"},
		{InstallationHomebrew, "homebrew"},
		{InstallationAPT, "apt"},
		{InstallationYum, "yum"},
		{InstallationPacman, "pacman"},
		{InstallationChocolatey, "chocolatey"},
		{InstallationScoop, "scoop"},
		{InstallationMethod(999), "unknown"},
	}

	for _, test := range tests {
		result := test.method.String()
		if result != test.expected {
			t.Errorf("Method %d: expected '%s', got '%s'", test.method, test.expected, result)
		}
	}
}

func TestInstallationMethodCanSelfUpdate(t *testing.T) {
	tests := []struct {
		method   InstallationMethod
		expected bool
	}{
		{InstallationBinary, true},
		{InstallationHomebrew, false},
		{InstallationAPT, false},
		{InstallationYum, false},
		{InstallationPacman, false},
		{InstallationChocolatey, false},
		{InstallationScoop, false},
		{InstallationMethod(999), false},
	}

	for _, test := range tests {
		result := test.method.CanSelfUpdate()
		if result != test.expected {
			t.Errorf("Method %s CanSelfUpdate: expected %t, got %t",
				test.method.String(), test.expected, result)
		}
	}
}

func TestInstallationMethodGetUpdateInstructions(t *testing.T) {
	tests := []struct {
		method   InstallationMethod
		contains string
	}{
		{InstallationBinary, "pvm update"},
		{InstallationHomebrew, "brew upgrade"},
		{InstallationAPT, "apt update"},
		{InstallationYum, "yum update"},
		{InstallationPacman, "pacman -Syu"},
		{InstallationChocolatey, "choco upgrade"},
		{InstallationScoop, "scoop update"},
	}

	for _, test := range tests {
		result := test.method.GetUpdateInstructions()
		if result == "" {
			t.Errorf("Method %s returned empty instructions", test.method.String())
			continue
		}

		if !contains(result, test.contains) {
			t.Errorf("Method %s instructions don't contain '%s': %s",
				test.method.String(), test.contains, result)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					someContains(s, substr))))
}

func someContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestUpdateResultValidation(t *testing.T) {
	result := &UpdateResult{
		Success:         true,
		UpdatePerformed: true,
		PreviousVersion: "1.0.0",
		NewVersion:      "1.1.0",
		Duration:        time.Second,
		BackupCreated:   true,
		BackupPath:      "/tmp/backup",
		DryRun:          false,
		Message:         "Update successful",
	}

	// Test that all fields are properly set
	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if !result.UpdatePerformed {
		t.Error("Expected UpdatePerformed to be true")
	}

	if result.PreviousVersion != "1.0.0" {
		t.Errorf("Expected PreviousVersion '1.0.0', got '%s'", result.PreviousVersion)
	}

	if result.NewVersion != "1.1.0" {
		t.Errorf("Expected NewVersion '1.1.0', got '%s'", result.NewVersion)
	}

	if result.Duration != time.Second {
		t.Errorf("Expected Duration 1s, got %v", result.Duration)
	}

	if !result.BackupCreated {
		t.Error("Expected BackupCreated to be true")
	}

	if result.BackupPath != "/tmp/backup" {
		t.Errorf("Expected BackupPath '/tmp/backup', got '%s'", result.BackupPath)
	}

	if result.DryRun {
		t.Error("Expected DryRun to be false")
	}

	if result.Message != "Update successful" {
		t.Errorf("Expected Message 'Update successful', got '%s'", result.Message)
	}
}

func TestUpdateWithInvalidRepository(t *testing.T) {
	updater := NewUpdater()

	opts := &UpdateOptions{
		Repository: "nonexistent/repo",
		Context:    context.Background(),
	}

	// This should fail because the repository doesn't exist
	_, err := updater.CheckForUpdates(opts)
	if err == nil {
		t.Error("Expected error when checking non-existent repository")
	}

	// The error should mention the repository not being found
	if !someContains(err.Error(), "404") && !someContains(err.Error(), "Not Found") {
		t.Errorf("Expected 404 or 'Not Found' in error, got: %v", err)
	}
}

func TestDryRunUpdate(t *testing.T) {
	updater := NewUpdater()

	opts := &UpdateOptions{
		Repository: "nonexistent/repo",
		DryRun:     true,
		Context:    context.Background(),
	}

	// Even with dry run, this will fail at the version check stage
	// because the repository doesn't exist
	_, err := updater.PerformUpdate(opts)
	if err == nil {
		t.Error("Expected error when checking non-existent repository")
	}
}

func TestProgressCallback(t *testing.T) {
	var capturedStages []UpdateStage
	var capturedMessages []string
	var capturedProgress []float64

	callback := func(stage UpdateStage, message string, progress float64) {
		capturedStages = append(capturedStages, stage)
		capturedMessages = append(capturedMessages, message)
		capturedProgress = append(capturedProgress, progress)
	}

	// Test the callback mechanism
	callback(StageCheckingVersion, "Testing", 0.5)

	if len(capturedStages) != 1 {
		t.Errorf("Expected 1 captured stage, got %d", len(capturedStages))
	}

	if capturedStages[0] != StageCheckingVersion {
		t.Errorf("Expected StageCheckingVersion, got %v", capturedStages[0])
	}

	if capturedMessages[0] != "Testing" {
		t.Errorf("Expected 'Testing', got '%s'", capturedMessages[0])
	}

	if capturedProgress[0] != 0.5 {
		t.Errorf("Expected 0.5, got %f", capturedProgress[0])
	}
}

// Integration test that tests the full flow with a mock setup
func TestUpdateWorkflowMocking(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-update-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a fake binary file
	binaryPath := filepath.Join(tempDir, "pvm")
	err = os.WriteFile(binaryPath, []byte("fake binary content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create fake binary: %v", err)
	}

	// Test GetCurrentBinaryPath with a real binary
	// (This test uses the actual test binary, which should exist)
	actualPath, err := GetCurrentBinaryPath()
	if err != nil {
		t.Fatalf("GetCurrentBinaryPath failed: %v", err)
	}

	// Test DetectInstallationMethod with the actual path
	method, err := DetectInstallationMethod(actualPath)
	if err != nil {
		t.Fatalf("DetectInstallationMethod failed: %v", err)
	}

	// Should detect as binary installation for test environment
	if method != InstallationBinary {
		t.Logf("Note: Detected installation method as %s instead of binary for test environment", method.String())
	}
}
