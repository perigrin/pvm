package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractPlatformFromAsset(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "linux amd64",
			filename: "perl-5.38.0-linux-amd64.tar.gz",
			expected: "linux-amd64",
		},
		{
			name:     "darwin arm64",
			filename: "perl-5.38.0-darwin-arm64.tar.gz",
			expected: "darwin-arm64",
		},
		{
			name:     "windows amd64",
			filename: "perl-5.38.0-windows-amd64.zip",
			expected: "windows-amd64",
		},
		{
			name:     "invalid format",
			filename: "perl-5.38.0.tar.gz",
			expected: "",
		},
		{
			name:     "unknown platform",
			filename: "perl-5.38.0-unknown-platform.tar.gz",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPlatformFromAsset(tt.filename)
			if result != tt.expected {
				t.Errorf("extractPlatformFromAsset(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		arch     string
		expected bool
	}{
		{"linux amd64", "linux", "amd64", true},
		{"darwin arm64", "darwin", "arm64", true},
		{"windows 386", "windows", "386", true},
		{"invalid os", "unknown", "amd64", false},
		{"invalid arch", "linux", "unknown", false},
		{"both invalid", "unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPlatform(tt.os, tt.arch)
			if result != tt.expected {
				t.Errorf("isValidPlatform(%q, %q) = %v, want %v", tt.os, tt.arch, result, tt.expected)
			}
		})
	}
}

func TestCalculateChecksum(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checksum, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("calculateChecksum failed: %v", err)
	}

	// Expected SHA-256 of "Hello, World!"
	expected := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	if checksum != expected {
		t.Errorf("calculateChecksum() = %q, want %q", checksum, expected)
	}
}

func TestScanBinaries(t *testing.T) {
	// Create temporary directory with test files
	tmpDir := t.TempDir()

	testFiles := []struct {
		name     string
		content  string
		isBinary bool
	}{
		{"perl-5.38.0-linux-amd64.tar.gz", "binary content", true},
		{"perl-5.38.0-darwin-arm64.tar.xz", "binary content", true},
		{"perl-5.38.0-windows-amd64.zip", "binary content", true},
		{"not-a-binary.txt", "text content", false},
		{"perl-config.json", "config content", false},
	}

	for _, tf := range testFiles {
		path := filepath.Join(tmpDir, tf.name)
		err := os.WriteFile(path, []byte(tf.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}
	}

	binaries, err := scanBinaries(tmpDir)
	if err != nil {
		t.Fatalf("scanBinaries failed: %v", err)
	}

	expectedCount := 0
	for _, tf := range testFiles {
		if tf.isBinary {
			expectedCount++
		}
	}

	if len(binaries) != expectedCount {
		t.Errorf("scanBinaries found %d binaries, want %d", len(binaries), expectedCount)
	}

	// Verify that all found binaries are expected
	for _, binary := range binaries {
		found := false
		for _, tf := range testFiles {
			if tf.name == binary.Name && tf.isBinary {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected binary found: %s", binary.Name)
		}
	}
}

func TestExtractPlatforms(t *testing.T) {
	binaries := []BinaryFile{
		{Name: "perl-5.38.0-linux-amd64.tar.gz", Path: "/path/to/file1", Size: 1000},
		{Name: "perl-5.38.0-darwin-arm64.tar.gz", Path: "/path/to/file2", Size: 2000},
		{Name: "perl-5.38.0-windows-amd64.zip", Path: "/path/to/file3", Size: 3000},
		{Name: "perl-5.38.0-linux-amd64.tar.gz", Path: "/path/to/file4", Size: 4000}, // duplicate platform
	}

	platforms := extractPlatforms(binaries)

	expected := []string{"darwin-arm64", "linux-amd64", "windows-amd64"}
	if len(platforms) != len(expected) {
		t.Errorf("extractPlatforms returned %d platforms, want %d", len(platforms), len(expected))
	}

	for i, platform := range platforms {
		if platform != expected[i] {
			t.Errorf("Platform[%d] = %q, want %q", i, platform, expected[i])
		}
	}
}

func TestSaveAndLoadMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	metaFile := filepath.Join(tmpDir, "test-meta.json")

	// Create test metadata
	originalMeta := ReleaseMeta{
		Version:   "5.38.0",
		CreatedAt: time.Now().Truncate(time.Second), // Truncate to avoid precision issues
		Platforms: []string{"linux-amd64", "darwin-arm64"},
		Checksums: map[string]string{
			"perl-5.38.0-linux-amd64.tar.gz":  "abc123",
			"perl-5.38.0-darwin-arm64.tar.gz": "def456",
		},
		TotalSize: 12345,
		Verified:  false,
	}

	// Save metadata
	err := saveMetadata(originalMeta, metaFile)
	if err != nil {
		t.Fatalf("saveMetadata failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(metaFile); os.IsNotExist(err) {
		t.Fatalf("Metadata file was not created")
	}

	// Load and verify metadata
	data, err := os.ReadFile(metaFile)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	var loadedMeta ReleaseMeta
	err = json.Unmarshal(data, &loadedMeta)
	if err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	// Compare metadata
	if loadedMeta.Version != originalMeta.Version {
		t.Errorf("Version mismatch: got %q, want %q", loadedMeta.Version, originalMeta.Version)
	}

	if !loadedMeta.CreatedAt.Equal(originalMeta.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", loadedMeta.CreatedAt, originalMeta.CreatedAt)
	}

	if len(loadedMeta.Platforms) != len(originalMeta.Platforms) {
		t.Errorf("Platforms length mismatch: got %d, want %d", len(loadedMeta.Platforms), len(originalMeta.Platforms))
	}

	if len(loadedMeta.Checksums) != len(originalMeta.Checksums) {
		t.Errorf("Checksums length mismatch: got %d, want %d", len(loadedMeta.Checksums), len(originalMeta.Checksums))
	}

	for key, value := range originalMeta.Checksums {
		if loadedMeta.Checksums[key] != value {
			t.Errorf("Checksum mismatch for %s: got %q, want %q", key, loadedMeta.Checksums[key], value)
		}
	}
}

func TestGenerateReleaseBody(t *testing.T) {
	binaries := []BinaryFile{
		{Name: "perl-5.38.0-linux-amd64.tar.gz", Path: "/path/to/file1", Size: 1000},
		{Name: "perl-5.38.0-darwin-arm64.tar.gz", Path: "/path/to/file2", Size: 2000},
	}

	checksums := map[string]string{
		"perl-5.38.0-linux-amd64.tar.gz":  "abc123",
		"perl-5.38.0-darwin-arm64.tar.gz": "def456",
	}

	totalSize := int64(3000)
	version := "5.38.0"

	body := generateReleaseBody(version, binaries, checksums, totalSize)

	// Verify that the body contains expected elements
	expectedElements := []string{
		"# Perl 5.38.0 Binary Distribution",
		"Pre-compiled Perl binaries",
		"## Available Platforms",
		"- darwin-arm64",
		"- linux-amd64",
		"## Checksums (SHA-256)",
		"abc123  perl-5.38.0-linux-amd64.tar.gz",
		"def456  perl-5.38.0-darwin-arm64.tar.gz",
		"**Total download size:** 3000 bytes",
	}

	for _, element := range expectedElements {
		if !containsString(body, element) {
			t.Errorf("Release body missing expected element: %q", element)
		}
	}
}

// Helper function for string contains check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkCalculateChecksum(b *testing.B) {
	// Create a test file
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "benchmark.txt")
	content := make([]byte, 1024*1024) // 1MB of data
	for i := range content {
		content[i] = byte(i % 256)
	}

	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := calculateChecksum(testFile)
		if err != nil {
			b.Fatalf("calculateChecksum failed: %v", err)
		}
	}
}

func BenchmarkExtractPlatformFromAsset(b *testing.B) {
	testFilenames := []string{
		"perl-5.38.0-linux-amd64.tar.gz",
		"perl-5.38.0-darwin-arm64.tar.gz",
		"perl-5.38.0-windows-amd64.zip",
		"invalid-filename.txt",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, filename := range testFilenames {
			extractPlatformFromAsset(filename)
		}
	}
}
