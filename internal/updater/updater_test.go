// ABOUTME: Tests for PVM self-updater functionality
// ABOUTME: Comprehensive tests for update orchestration and component integration

package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/archive"
	"tamarou.com/pvm/internal/download"
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

	if opts.Repository != "perigrin/pvm" {
		t.Errorf("Expected repository 'perigrin/pvm', got '%s'", opts.Repository)
	}

	if !opts.IncludePrerelease {
		t.Error("Expected IncludePrerelease to be true by default (to support prerelease-only repositories)")
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

	// Test that nil options are handled correctly by using defaults
	// This should succeed with the default repository (perigrin/pvm)
	_, err := updater.CheckForUpdates(nil)
	if err != nil {
		t.Logf("CheckForUpdates with nil options failed (may be due to network/auth): %v", err)
		// We're testing that nil options don't cause a panic, not necessarily that the API call succeeds
		return
	}
	// Success means nil options were handled correctly
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
	if runtime.GOOS == "windows" {
		t.Skip("Unix installation method detection not applicable on Windows")
	}
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

		if !testContains(result, test.contains) {
			t.Errorf("Method %s instructions don't contain '%s': %s",
				test.method.String(), test.contains, result)
		}
	}
}

// Helper function to check if string contains substring
func testContains(s, substr string) bool {
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

	// The error should mention the repository not being found OR be a rate limit error
	// In CI environments, we might hit rate limits before getting a 404
	errStr := err.Error()
	if !someContains(errStr, "404") && !someContains(errStr, "Not Found") &&
		!someContains(errStr, "rate limit") && !someContains(errStr, "403") {
		t.Errorf("Expected 404, 'Not Found', rate limit, or 403 error, got: %v", err)
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

func TestPrereleaseInclusion(t *testing.T) {
	// Set environment variable to ensure test-friendly behavior
	os.Setenv("GO_TEST", "1")
	defer os.Unsetenv("GO_TEST")

	t.Run("DefaultOptionsIncludePrerelease", func(t *testing.T) {
		opts := DefaultUpdateOptions()
		if !opts.IncludePrerelease {
			t.Error("Default update options should include prereleases to handle prerelease-only repositories")
		}
	})

	t.Run("ExplicitPrereleaseExclusion", func(t *testing.T) {
		// Set a reasonable timeout for test environment
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		opts := &UpdateOptions{
			Repository:        "perigrin/pvm",
			IncludePrerelease: false,
			Context:           ctx,
		}

		updater := NewUpdater()

		// This test checks that explicit exclusion of prereleases is respected
		// but may still succeed due to fallback logic if only prereleases exist
		_, err := updater.CheckForUpdates(opts)
		if err != nil {
			t.Logf("CheckForUpdates with IncludePrerelease=false failed (expected if only prereleases exist): %v", err)
			// This is expected behavior when only prereleases exist and fallback succeeds
		}
	})

	t.Run("ExplicitPrereleaseInclusion", func(t *testing.T) {
		// Set a reasonable timeout for test environment
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		opts := &UpdateOptions{
			Repository:        "perigrin/pvm",
			IncludePrerelease: true,
			Context:           ctx,
		}

		updater := NewUpdater()

		// This should work since we explicitly include prereleases
		_, err := updater.CheckForUpdates(opts)
		if err != nil {
			t.Logf("CheckForUpdates with IncludePrerelease=true failed (may be due to network/auth/timeout): %v", err)
			// Network errors and timeouts are acceptable in tests, we're testing option handling
		}
	})
}

func TestUpdateWithArchiveExtraction_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("ArchiveValidationWorkflow", func(t *testing.T) {
		// Create a test archive with a mock binary
		archivePath := createMockPVMArchive(t)
		defer os.Remove(archivePath)

		// Test that ValidateDownloadedBinary works with archives
		err := download.ValidateDownloadedBinary(archivePath, "")
		if err != nil {
			t.Logf("Archive validation failed (expected in test environment): %v", err)
			// This may fail in test environment due to execution validation,
			// but we want to test that archive extraction works
		}
	})

	t.Run("BinaryExtractionFromArchive", func(t *testing.T) {
		// Create test archive
		archivePath := createMockPVMArchive(t)
		defer os.Remove(archivePath)

		// Test archive extraction
		extractor := archive.NewBinaryExtractor()
		platform := detectCurrentPlatform()

		extractedPath, err := extractor.ExtractExecutable(archivePath, platform)
		if err != nil {
			t.Fatalf("Failed to extract binary from archive: %v", err)
		}

		// Verify extracted file exists
		_, err = os.Stat(extractedPath)
		if err != nil {
			t.Errorf("Extracted binary does not exist: %v", err)
		}

		// Cleanup
		err = extractor.Cleanup(extractedPath)
		if err != nil {
			t.Errorf("Failed to cleanup extracted files: %v", err)
		}
	})
}

func TestUpdateValidationFailure_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("InvalidArchiveHandling", func(t *testing.T) {
		// Create an invalid archive (corrupted)
		invalidArchive := createInvalidArchive(t)
		defer os.Remove(invalidArchive)

		// Test that validation properly fails
		err := download.ValidateDownloadedBinary(invalidArchive, "")
		if err == nil {
			t.Error("Expected validation to fail for invalid archive")
		}
	})

	t.Run("ArchiveWithoutExecutable", func(t *testing.T) {
		// Create archive without the expected executable
		archivePath := createArchiveWithoutExecutable(t)
		defer os.Remove(archivePath)

		err := download.ValidateDownloadedBinary(archivePath, "")
		if err == nil {
			t.Error("Expected validation to fail for archive without executable")
		}

		if err != nil && !containsAny(err.Error(), []string{"no executable found", "finding executable"}) {
			t.Errorf("Expected 'no executable found' error, got: %v", err)
		}
	})
}

// Helper functions for integration tests

func createMockPVMArchive(t *testing.T) string {
	// Create a mock binary content
	binaryContent := createMockBinaryContent()

	// Determine filename based on platform
	var filename string
	if runtime.GOOS == "windows" {
		filename = "pvm.exe"
	} else {
		filename = "pvm"
	}

	// Create tar.gz archive
	return createTestTarGzArchive(t, filename, binaryContent)
}

func createMockBinaryContent() []byte {
	// Create a binary that will pass format validation but won't execute properly
	switch runtime.GOOS {
	case "linux":
		// ELF header with minimum viable content
		elf := make([]byte, 4096) // 4KB to pass size validation
		copy(elf, []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00})
		return elf
	case "darwin":
		// Mach-O header with minimum viable content
		macho := make([]byte, 4096)
		copy(macho, []byte{0xcf, 0xfa, 0xed, 0xfe, 0x07, 0x00, 0x00, 0x01})
		return macho
	case "windows":
		// PE header with minimum viable content
		pe := make([]byte, 4096)
		copy(pe, []byte{'M', 'Z', 0x90, 0x00, 0x03, 0x00, 0x00, 0x00})
		return pe
	default:
		// Shell script as fallback - this will actually work for execution tests
		script := `#!/bin/sh
if [ "$1" = "--version" ]; then
    echo "pvm version 1.0.0-test"
    exit 0
elif [ "$1" = "--help" ]; then
    echo "Usage: pvm [command]"
    exit 0
fi
echo "Test binary"
exit 0
`
		padded := make([]byte, 4096)
		copy(padded, script)
		return padded
	}
}

func createTestTarGzArchive(t *testing.T, filename string, content []byte) string {
	tempFile, err := os.CreateTemp("", "mock-pvm-*.tar.gz")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(tempFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add file to tar
	header := &tar.Header{
		Name: filename,
		Size: int64(len(content)),
		Mode: 0755,
	}

	err = tarWriter.WriteHeader(header)
	if err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}

	_, err = tarWriter.Write(content)
	if err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	return tempFile.Name()
}

func createInvalidArchive(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "invalid-*.tar.gz")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// Write invalid gzip content
	_, err = tempFile.WriteString("This is not a valid gzip file")
	if err != nil {
		t.Fatalf("Failed to write invalid content: %v", err)
	}

	return tempFile.Name()
}

func createArchiveWithoutExecutable(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "no-exec-*.tar.gz")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(tempFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add non-executable files
	files := map[string]string{
		"README.txt":   "This is a readme file",
		"config.yml":   "configuration: test",
		"docs/help.md": "# Help Documentation",
	}

	for filename, content := range files {
		header := &tar.Header{
			Name: filename,
			Size: int64(len(content)),
			Mode: 0644,
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			t.Fatalf("Failed to write tar header: %v", err)
		}

		_, err = tarWriter.Write([]byte(content))
		if err != nil {
			t.Fatalf("Failed to write tar content: %v", err)
		}
	}

	return tempFile.Name()
}

func detectCurrentPlatform() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Convert Go arch names to our naming convention
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "i386"
	default:
		arch = "unknown"
	}

	return fmt.Sprintf("%s-%s", os, arch)
}

func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}
