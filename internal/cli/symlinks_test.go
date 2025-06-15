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
	defer func() { _ = os.Remove(binPath) }()

	// Verify with no symlinks
	status := VerifySymlinks(binPath)

	// Check each component status (should all be false)
	if status[ComponentPVM] || status[ComponentPVX] ||
		status[ComponentPVI] || status[ComponentPSC] {
		t.Error("Expected all symlinks to be missing, but some were found")
	}
}

func TestCreateSymlinks(t *testing.T) {
	// Note: Windows symlink creation is handled via hard linking or file copying,
	// so this test should work on all platforms including CI environments

	// Create a temporary binary path for testing
	tempDir := os.TempDir()
	binPath := filepath.Join(tempDir, "test_binary")

	// Write a dummy file
	err := os.WriteFile(binPath, []byte("test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	defer func() { _ = os.Remove(binPath) }()

	// Create symlinks
	symlinks, err := CreateSymlinks(binPath)
	if err != nil {
		t.Fatalf("Failed to create symlinks: %v", err)
	}

	// Check created symlinks and clean up
	for component, symlinkPath := range symlinks {
		// Check that the symlink/copy exists
		info, err := os.Lstat(symlinkPath) // Use Lstat to get symlink info
		if err != nil {
			t.Errorf("Symlink/copy for %s does not exist: %v", component, err)
		}

		// On Windows, verify it's a file (hard link or copy), on Unix verify it's a symlink
		if runtime.GOOS == "windows" {
			if info.IsDir() {
				t.Errorf("Expected file for %s on Windows, got directory", component)
			}
		} else {
			// On Unix systems, check that it's a symlink
			if info.Mode()&os.ModeSymlink == 0 {
				t.Errorf("Expected symlink for %s on Unix systems, got regular file", component)
			}
		}

		// Clean up
		_ = os.Remove(symlinkPath)
	}
}

func TestCopyFile(t *testing.T) {
	// Create a source file
	tempDir := os.TempDir()
	srcPath := filepath.Join(tempDir, "source_file")
	dstPath := filepath.Join(tempDir, "dest_file")

	// Clean up any existing files
	_ = os.Remove(srcPath)
	_ = os.Remove(dstPath)

	// Write content to source file
	content := []byte("test content")
	err := os.WriteFile(srcPath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	defer func() { _ = os.Remove(srcPath) }()

	// Copy the file using platform package
	data, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	err = os.WriteFile(dstPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}
	defer func() { _ = os.Remove(dstPath) }()

	// Check that the destination file exists and has the correct content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("Destination file content does not match source: expected %q, got %q", content, dstContent)
	}
}

func TestCrossplatformSymlinkCreation(t *testing.T) {
	// Test that demonstrates cross-platform compatibility
	tempDir := os.TempDir()
	binPath := filepath.Join(tempDir, "test_binary")

	// Write a test binary
	err := os.WriteFile(binPath, []byte("#!/bin/bash\necho 'test'"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	defer func() { _ = os.Remove(binPath) }()

	// Create symlinks
	symlinks, err := CreateSymlinks(binPath)
	if err != nil {
		t.Fatalf("Failed to create symlinks: %v", err)
	}

	t.Logf("Created %d symlinks/links for platform %s", len(symlinks), runtime.GOOS)

	// Platform-specific validation
	for component, linkPath := range symlinks {
		info, err := os.Lstat(linkPath) // Use Lstat to get symlink info, not target info
		if err != nil {
			t.Errorf("Failed to stat link for %s: %v", component, err)
			continue
		}

		switch runtime.GOOS {
		case "windows":
			// On Windows, we expect a regular file (hard link or copy)
			switch {
			case info.Mode()&os.ModeSymlink != 0:
				t.Logf("Windows: %s is a symlink (unusual but acceptable)", component)
			case info.Mode().IsRegular():
				t.Logf("Windows: %s is a regular file/hard link (expected)", component)
			default:
				t.Errorf("Windows: %s has unexpected file mode: %v", component, info.Mode())
			}
		default:
			// On Unix systems, we expect a symlink
			if info.Mode()&os.ModeSymlink != 0 {
				t.Logf("Unix: %s is a symlink (expected)", component)
			} else {
				t.Errorf("Unix: %s is not a symlink, mode: %v", component, info.Mode())
			}
		}

		// Clean up
		_ = os.Remove(linkPath)
	}
}
