// ABOUTME: Integration tests for disk space validation in download operations
// ABOUTME: Tests the TempFileManager disk space checking with real filesystem operations
package download

import (
	"testing"
)

// TestValidateDiskSpaceIntegration tests the actual disk space checking implementation
func TestValidateDiskSpaceIntegration(t *testing.T) {
	tempDir := t.TempDir()
	tfm := &TempFileManager{
		tempDir: tempDir,
		prefix:  "pvm-test",
	}

	tests := []struct {
		name          string
		requiredBytes int64
		expectError   bool
	}{
		{
			name:          "small file should pass",
			requiredBytes: 1024, // 1KB
			expectError:   false,
		},
		{
			name:          "moderate file should pass",
			requiredBytes: 1024 * 1024, // 1MB
			expectError:   false,
		},
		{
			name:          "extremely large file should fail on most systems",
			requiredBytes: 1024 * 1024 * 1024 * 1024 * 100, // 100TB
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tfm.ValidateDiskSpace(tt.requiredBytes)

			if tt.expectError && err == nil {
				t.Errorf("ValidateDiskSpace() expected error for %d bytes but got none", tt.requiredBytes)
			}

			if !tt.expectError && err != nil {
				t.Errorf("ValidateDiskSpace() unexpected error for %d bytes: %v", tt.requiredBytes, err)
			}

			if err != nil {
				t.Logf("ValidateDiskSpace() error (expected: %v): %v", tt.expectError, err)
			}
		})
	}
}

// TestGetAvailableSpaceReal tests the real disk space implementation
func TestGetAvailableSpaceReal(t *testing.T) {
	tempDir := t.TempDir()
	tfm := &TempFileManager{
		tempDir: tempDir,
		prefix:  "pvm-test",
	}

	availableSpace, err := tfm.getAvailableSpace(tempDir)
	if err != nil {
		t.Fatalf("getAvailableSpace() error: %v", err)
	}

	if availableSpace <= 0 {
		t.Errorf("getAvailableSpace() = %d, want > 0", availableSpace)
	}

	// Should not be the old hardcoded value
	if availableSpace == 1024*1024*1024 {
		t.Logf("WARNING: getAvailableSpace() returned 1GB exactly - this might be the old hardcoded value")
	}

	t.Logf("Available space in %s: %d bytes (%.2f MB)", tempDir, availableSpace, float64(availableSpace)/(1024*1024))
}

// TestGetAvailableSpaceNonExistentPath tests handling of invalid paths
func TestGetAvailableSpaceNonExistentPath(t *testing.T) {
	tfm := &TempFileManager{
		tempDir: "/completely/invalid/path",
		prefix:  "pvm-test",
	}

	// This should succeed by finding a parent directory (eventually root)
	availableSpace, err := tfm.getAvailableSpace(tfm.tempDir)
	if err != nil {
		t.Fatalf("getAvailableSpace() error for non-existent path: %v", err)
	}

	if availableSpace <= 0 {
		t.Errorf("getAvailableSpace() = %d, want > 0", availableSpace)
	}
}

// TestDiskSpaceBufferCalculation tests the 10% buffer is correctly applied
func TestDiskSpaceBufferCalculation(t *testing.T) {
	tempDir := t.TempDir()
	tfm := &TempFileManager{
		tempDir: tempDir,
		prefix:  "pvm-test",
	}

	// Get available space
	availableSpace, err := tfm.getAvailableSpace(tempDir)
	if err != nil {
		t.Fatalf("getAvailableSpace() error: %v", err)
	}

	// Test with exactly available space (should fail due to 10% buffer)
	err = tfm.ValidateDiskSpace(availableSpace)
	if err == nil {
		t.Errorf("ValidateDiskSpace() expected error due to 10%% buffer, but got none")
	}

	// Test with 80% of available space (should pass)
	err = tfm.ValidateDiskSpace(availableSpace * 80 / 100)
	if err != nil {
		t.Errorf("ValidateDiskSpace() unexpected error with 80%% of available space: %v", err)
	}
}
