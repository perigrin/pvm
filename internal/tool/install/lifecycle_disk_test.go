// ABOUTME: Tests for disk usage calculation in tool lifecycle management
// ABOUTME: Validates tool disk usage calculation and error handling
package install

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tamarou.com/pvm/internal/diskspace"
)

// TestDiskUsageCalculationUnit tests just the disk usage calculation logic
func TestDiskUsageCalculationUnit(t *testing.T) {
	// Create test directory with known files
	testDir := t.TempDir()

	// Create test files with known sizes
	testFiles := map[string]string{
		"file1.txt":        "Hello, World!",                                 // 13 bytes
		"file2.txt":        "This is a longer test file with more content.", // 44 bytes
		"subdir/file3.txt": "Nested file content",                           // 19 bytes
	}

	var expectedSize int64
	for filename, content := range testFiles {
		filePath := filepath.Join(testDir, filename)

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create parent directory: %v", err)
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		expectedSize += int64(len(content))
	}

	// Test the disk space calculation
	actualSize, err := diskspace.CalculateDirectorySize(testDir)
	if err != nil {
		t.Fatalf("CalculateDirectorySize() error: %v", err)
	}

	if actualSize != expectedSize {
		t.Errorf("CalculateDirectorySize() = %d, want %d", actualSize, expectedSize)
	}

	t.Logf("Directory size calculation: %d bytes (expected: %d)", actualSize, expectedSize)
}

// TestDiskUsageCalculationWithRealMetadata tests disk usage with tool metadata
func TestDiskUsageCalculationWithRealMetadata(t *testing.T) {
	// Create temporary directory for tool storage
	tempDir := t.TempDir()
	storage := &ToolStorage{baseDir: tempDir}

	// Create a test tool directory with files
	toolName := "test-tool"
	toolDir := filepath.Join(tempDir, toolName)

	if err := os.MkdirAll(toolDir, 0755); err != nil {
		t.Fatalf("Failed to create tool directory: %v", err)
	}

	// Create some files
	testFiles := map[string]string{
		"bin/tool":    "#!/bin/bash\necho 'test tool'\n",                           // ~25 bytes
		"lib/Tool.pm": "package Tool;\nsub version { '1.0' }\n1;\n",                // ~35 bytes
		"README.md":   "# Test Tool\nA simple test tool for disk usage testing.\n", // ~50 bytes
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(toolDir, filename)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create parent directory: %v", err)
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create metadata
	metadata := &ToolMetadata{
		ToolName:    toolName,
		ModuleName:  "Test::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
		InstallPath: toolDir,
		Status:      "installed",
	}

	if err := storage.SaveMetadata(metadata); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Test disk usage calculation
	diskUsage, err := diskspace.CalculateDirectorySize(toolDir)
	if err != nil {
		t.Fatalf("CalculateDirectorySize() error: %v", err)
	}

	if diskUsage <= 0 {
		t.Errorf("Disk usage = %d, want > 0", diskUsage)
	}

	// Should be approximately 110 bytes (25+35+50) plus metadata.json file
	// The metadata file can be a few hundred bytes, so allow a wider range
	expectedMin := int64(100)
	expectedMax := int64(1000) // Increased to account for metadata.json
	if diskUsage < expectedMin || diskUsage > expectedMax {
		t.Errorf("Disk usage = %d, expected between %d and %d bytes (includes metadata file)", diskUsage, expectedMin, expectedMax)
	}

	t.Logf("Tool disk usage: %d bytes", diskUsage)
}

// TestDiskUsageCalculationErrorHandling tests error cases
func TestDiskUsageCalculationErrorHandling(t *testing.T) {
	// Test with non-existent directory
	nonExistentPath := "/completely/non/existent/path"

	_, err := diskspace.CalculateDirectorySize(nonExistentPath)
	if err == nil {
		t.Error("CalculateDirectorySize() expected error for non-existent path, got none")
	}

	t.Logf("Expected error for non-existent path: %v", err)
}
