// ABOUTME: Comprehensive tests for disk space utilities
// ABOUTME: Tests cross-platform functionality and edge cases
package diskspace

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetSpaceInfo(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "current directory",
			path:    ".",
			wantErr: false,
		},
		{
			name:    "temp directory",
			path:    os.TempDir(),
			wantErr: false,
		},
		{
			name:    "root directory",
			path:    "/",
			wantErr: false,
		},
		{
			name:    "non-existent path with existing parent",
			path:    filepath.Join(os.TempDir(), "non-existent-dir"),
			wantErr: false,
		},
		{
			name:    "relative path with non-existent parent",
			path:    "./non-existent-parent/child",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetSpaceInfo(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetSpaceInfo() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetSpaceInfo() unexpected error: %v", err)
				return
			}

			if info == nil {
				t.Error("GetSpaceInfo() returned nil info")
				return
			}

			// Validate space information makes sense
			if info.Total <= 0 {
				t.Errorf("GetSpaceInfo() Total = %d, want > 0", info.Total)
			}

			if info.Free < 0 {
				t.Errorf("GetSpaceInfo() Free = %d, want >= 0", info.Free)
			}

			if info.Available < 0 {
				t.Errorf("GetSpaceInfo() Available = %d, want >= 0", info.Available)
			}

			// Available space should not exceed free space
			if info.Available > info.Free {
				t.Errorf("GetSpaceInfo() Available (%d) > Free (%d), which is invalid",
					info.Available, info.Free)
			}

			// Free space should not exceed total space
			if info.Free > info.Total {
				t.Errorf("GetSpaceInfo() Free (%d) > Total (%d), which is invalid",
					info.Free, info.Total)
			}

			t.Logf("Space info for %s: Total=%d, Free=%d, Available=%d",
				tt.path, info.Total, info.Free, info.Available)
		})
	}
}

func TestCalculateDirectorySize(t *testing.T) {
	// Create a temporary directory with known content
	tempDir := t.TempDir()

	// Create some test files with known sizes
	testFiles := []struct {
		name    string
		content string
	}{
		{"file1.txt", "Hello, World!"},
		{"file2.txt", "This is a test file with more content."},
		{"subdir/file3.txt", "Nested file content"},
	}

	var expectedSize int64
	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)

		// Create subdirectory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		if err := os.WriteFile(filePath, []byte(tf.content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		expectedSize += int64(len(tf.content))
	}

	// Test directory size calculation
	actualSize, err := CalculateDirectorySize(tempDir)
	if err != nil {
		t.Fatalf("CalculateDirectorySize() error: %v", err)
	}

	if actualSize != expectedSize {
		t.Errorf("CalculateDirectorySize() = %d, want %d", actualSize, expectedSize)
	}

	t.Logf("Directory size: %d bytes", actualSize)
}

func TestCalculateDirectorySizeEmptyDir(t *testing.T) {
	tempDir := t.TempDir()

	size, err := CalculateDirectorySize(tempDir)
	if err != nil {
		t.Fatalf("CalculateDirectorySize() error for empty dir: %v", err)
	}

	if size != 0 {
		t.Errorf("CalculateDirectorySize() for empty dir = %d, want 0", size)
	}
}

func TestCalculateDirectorySizeNonExistent(t *testing.T) {
	nonExistentDir := "/path/that/definitely/does/not/exist"

	_, err := CalculateDirectorySize(nonExistentDir)
	if err == nil {
		t.Error("CalculateDirectorySize() expected error for non-existent directory")
	}
}

// Benchmark tests for performance validation
func BenchmarkGetSpaceInfo(b *testing.B) {
	tempDir := os.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetSpaceInfo(tempDir)
		if err != nil {
			b.Fatalf("GetSpaceInfo() error: %v", err)
		}
	}
}

func BenchmarkCalculateDirectorySize(b *testing.B) {
	// Create a directory with moderate content for benchmarking
	tempDir := b.TempDir()

	// Create some test files
	for i := 0; i < 10; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
		content := make([]byte, 1024) // 1KB files
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			b.Fatalf("Failed to create benchmark file: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CalculateDirectorySize(tempDir)
		if err != nil {
			b.Fatalf("CalculateDirectorySize() error: %v", err)
		}
	}
}
