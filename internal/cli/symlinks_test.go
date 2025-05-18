package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVerifySymlinks(t *testing.T) {
	// Create a temporary binary path for testing
	tempDir := os.TempDir()
	binPath := filepath.Join(tempDir, "test_binary")

	// Write a dummy file
	err := os.WriteFile(binPath, []byte("test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	defer os.Remove(binPath)

	// Verify with no symlinks
	status := VerifySymlinks(binPath)

	// Check each component status (should all be false)
	if status[ComponentPVM] || status[ComponentPVX] ||
		status[ComponentPVI] || status[ComponentPSC] {
		t.Error("Expected all symlinks to be missing, but some were found")
	}
}

func TestCreateSymlinks(t *testing.T) {
	// Skip on Windows in CI environments since it might require elevated permissions
	if runtime.GOOS == "windows" && os.Getenv("CI") != "" {
		t.Skip("Skipping symlink test on Windows in CI environment")
	}

	// Create a temporary binary path for testing
	tempDir := os.TempDir()
	binPath := filepath.Join(tempDir, "test_binary")

	// Write a dummy file
	err := os.WriteFile(binPath, []byte("test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	defer os.Remove(binPath)

	// Create symlinks
	symlinks, err := CreateSymlinks(binPath)
	if err != nil {
		t.Fatalf("Failed to create symlinks: %v", err)
	}

	// Check created symlinks and clean up
	for component, symlinkPath := range symlinks {
		// Check that the symlink exists
		_, err := os.Stat(symlinkPath)
		if err != nil {
			t.Errorf("Symlink for %s does not exist: %v", component, err)
		}

		// Clean up
		os.Remove(symlinkPath)
	}
}

func TestCopyFile(t *testing.T) {
	// Create a source file
	tempDir := os.TempDir()
	srcPath := filepath.Join(tempDir, "source_file")
	dstPath := filepath.Join(tempDir, "dest_file")

	// Clean up any existing files
	os.Remove(srcPath)
	os.Remove(dstPath)

	// Write content to source file
	content := []byte("test content")
	err := os.WriteFile(srcPath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	defer os.Remove(srcPath)

	// Copy the file
	err = copyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}
	defer os.Remove(dstPath)

	// Check that the destination file exists and has the correct content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("Destination file content does not match source: expected %q, got %q", content, dstContent)
	}
}
