// ABOUTME: Tests for module extraction functionality
// ABOUTME: Covers root directory detection and archive handling edge cases

package modules

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestArchive creates a tar.gz archive with the specified structure
func createTestArchive(t *testing.T, structure map[string]string) string {
	t.Helper()

	// Create temporary file for the archive
	tmpFile, err := os.CreateTemp("", "test-archive-*.tar.gz")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	// Create gzip writer
	gzw := gzip.NewWriter(tmpFile)
	defer gzw.Close()

	// Create tar writer
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add files to archive
	for path, content := range structure {
		// Determine if this is a directory or file
		isDir := strings.HasSuffix(path, "/")
		if isDir {
			path = strings.TrimSuffix(path, "/")
		}

		if isDir {
			// Add directory
			header := &tar.Header{
				Name:     path + "/",
				Mode:     0755,
				Typeflag: tar.TypeDir,
			}
			if err := tw.WriteHeader(header); err != nil {
				t.Fatalf("Failed to write directory header: %v", err)
			}
		} else {
			// Add file
			header := &tar.Header{
				Name:     path,
				Mode:     0644,
				Size:     int64(len(content)),
				Typeflag: tar.TypeReg,
			}
			if err := tw.WriteHeader(header); err != nil {
				t.Fatalf("Failed to write file header: %v", err)
			}
			if _, err := tw.Write([]byte(content)); err != nil {
				t.Fatalf("Failed to write file content: %v", err)
			}
		}
	}

	return tmpFile.Name()
}

func TestSelectBestRootDirectory_SingleDirectory(t *testing.T) {
	// Test basic case with single directory containing Build.PL
	seenDirs := map[string]bool{
		"Module-Name-1.0": true,
	}

	allEntries := map[string]tar.Header{
		"Module-Name-1.0/Build.PL":           {Name: "Module-Name-1.0/Build.PL"},
		"Module-Name-1.0/lib/Module/Name.pm": {Name: "Module-Name-1.0/lib/Module/Name.pm"},
	}

	rootDir, err := selectBestRootDirectory("/tmp/test", seenDirs, allEntries)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rootDir != "Module-Name-1.0" {
		t.Errorf("Expected root directory 'Module-Name-1.0', got: %s", rootDir)
	}
}

func TestSelectBestRootDirectory_MultipleDirectoriesWithBuildFiles(t *testing.T) {
	// Test case with multiple directories containing different build systems
	// Build.PL should be preferred over Makefile.PL
	seenDirs := map[string]bool{
		"Module-A-1.0": true,
		"Module-B-1.0": true,
		"docs":         true,
	}

	allEntries := map[string]tar.Header{
		"Module-A-1.0/Makefile.PL": {Name: "Module-A-1.0/Makefile.PL"},
		"Module-B-1.0/Build.PL":    {Name: "Module-B-1.0/Build.PL"},
		"docs/README.txt":          {Name: "docs/README.txt"},
	}

	rootDir, err := selectBestRootDirectory("/tmp/test", seenDirs, allEntries)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Build.PL should be preferred over Makefile.PL
	if rootDir != "Module-B-1.0" {
		t.Errorf("Expected root directory 'Module-B-1.0' (Build.PL), got: %s", rootDir)
	}
}

func TestSelectBestRootDirectory_DeterministicSelection(t *testing.T) {
	// Test that selection is deterministic when multiple directories have same build system
	seenDirs := map[string]bool{
		"Module-Z-1.0": true,
		"Module-A-1.0": true,
		"Module-M-1.0": true,
	}

	allEntries := map[string]tar.Header{
		"Module-Z-1.0/Build.PL": {Name: "Module-Z-1.0/Build.PL"},
		"Module-A-1.0/Build.PL": {Name: "Module-A-1.0/Build.PL"},
		"Module-M-1.0/Build.PL": {Name: "Module-M-1.0/Build.PL"},
	}

	// Run multiple times to ensure deterministic behavior
	var results []string
	for i := 0; i < 5; i++ {
		rootDir, err := selectBestRootDirectory("/tmp/test", seenDirs, allEntries)
		if err != nil {
			t.Fatalf("Expected no error on iteration %d, got: %v", i, err)
		}
		results = append(results, rootDir)
	}

	// All results should be the same and alphabetically first
	expectedRoot := "Module-A-1.0"
	for i, result := range results {
		if result != expectedRoot {
			t.Errorf("Iteration %d: expected %s, got %s", i, expectedRoot, result)
		}
	}
}

func TestSelectBestRootDirectory_NoBuildFiles(t *testing.T) {
	// Test fallback behavior when no directories contain build files
	seenDirs := map[string]bool{
		"docs":    true,
		"scripts": true,
		"data":    true,
	}

	allEntries := map[string]tar.Header{
		"docs/README.txt":  {Name: "docs/README.txt"},
		"scripts/setup.sh": {Name: "scripts/setup.sh"},
		"data/config.json": {Name: "data/config.json"},
	}

	rootDir, err := selectBestRootDirectory("/tmp/test", seenDirs, allEntries)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should select alphabetically first directory
	if rootDir != "data" {
		t.Errorf("Expected root directory 'data' (alphabetically first), got: %s", rootDir)
	}
}

func TestSelectBestRootDirectory_BuildSystemPriority(t *testing.T) {
	// Test that Build.PL > Makefile.PL > Makefile in priority
	testCases := []struct {
		name     string
		entries  map[string]tar.Header
		expected string
	}{
		{
			name: "Build.PL vs Makefile.PL",
			entries: map[string]tar.Header{
				"module-a/Makefile.PL": {Name: "module-a/Makefile.PL"},
				"module-b/Build.PL":    {Name: "module-b/Build.PL"},
			},
			expected: "module-b",
		},
		{
			name: "Build.PL vs Makefile",
			entries: map[string]tar.Header{
				"module-a/Makefile": {Name: "module-a/Makefile"},
				"module-b/Build.PL": {Name: "module-b/Build.PL"},
			},
			expected: "module-b",
		},
		{
			name: "Makefile.PL vs Makefile",
			entries: map[string]tar.Header{
				"module-a/Makefile":    {Name: "module-a/Makefile"},
				"module-b/Makefile.PL": {Name: "module-b/Makefile.PL"},
			},
			expected: "module-b",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			seenDirs := map[string]bool{
				"module-a": true,
				"module-b": true,
			}

			rootDir, err := selectBestRootDirectory("/tmp/test", seenDirs, tc.entries)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if rootDir != tc.expected {
				t.Errorf("Expected root directory '%s', got: %s", tc.expected, rootDir)
			}
		})
	}
}

func TestSelectBestRootDirectory_EmptyDirs(t *testing.T) {
	// Test error handling when no directories are found
	seenDirs := map[string]bool{}
	allEntries := map[string]tar.Header{}

	_, err := selectBestRootDirectory("/tmp/test", seenDirs, allEntries)
	if err == nil {
		t.Error("Expected error for empty directories, got none")
	}

	expectedError := "no directories found in archive"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

func TestExtractModuleArchive_Integration(t *testing.T) {
	// Integration test that creates actual archive and tests full extraction
	ctx := context.Background()

	// Create test archive structure simulating Perl::Critic-like layout
	archiveStructure := map[string]string{
		"Perl-Critic-1.156/":                                  "",
		"Perl-Critic-1.156/Build.PL":                          "use lib 'inc';\nuse Perl::Critic::BuildUtilities;",
		"Perl-Critic-1.156/inc/":                              "",
		"Perl-Critic-1.156/inc/Perl/":                         "",
		"Perl-Critic-1.156/inc/Perl/Critic/":                  "",
		"Perl-Critic-1.156/inc/Perl/Critic/BuildUtilities.pm": "package Perl::Critic::BuildUtilities;\n1;",
		"Perl-Critic-1.156/lib/":                              "",
		"Perl-Critic-1.156/lib/Perl/":                         "",
		"Perl-Critic-1.156/lib/Perl/Critic.pm":                "package Perl::Critic;\n1;",
		"Perl-Critic-1.156-docs/":                             "",
		"Perl-Critic-1.156-docs/README.txt":                   "Documentation files",
	}

	archivePath := createTestArchive(t, archiveStructure)
	defer os.Remove(archivePath)

	// Create temporary target directory
	targetDir, err := os.MkdirTemp("", "test-extract-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	// Extract the archive
	result, err := extractModuleArchive(archivePath, targetDir, ctx)
	if err != nil {
		t.Fatalf("Failed to extract archive: %v", err)
	}

	// Verify correct root directory was selected
	if result.RootDir != "Perl-Critic-1.156" {
		t.Errorf("Expected root directory 'Perl-Critic-1.156', got: %s", result.RootDir)
	}

	// Verify Build.PL exists in the extracted directory
	buildPLPath := filepath.Join(result.ExtractedDir, "Build.PL")
	if _, err := os.Stat(buildPLPath); os.IsNotExist(err) {
		t.Errorf("Build.PL not found at expected path: %s", buildPLPath)
	}

	// Verify inc directory structure exists
	incUtilsPath := filepath.Join(result.ExtractedDir, "inc", "Perl", "Critic", "BuildUtilities.pm")
	if _, err := os.Stat(incUtilsPath); os.IsNotExist(err) {
		t.Errorf("BuildUtilities.pm not found at expected path: %s", incUtilsPath)
	}

	// Verify the docs directory was NOT selected as root
	docsPath := filepath.Join(targetDir, "Perl-Critic-1.156-docs")
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		t.Errorf("Docs directory should exist but not be selected as root: %s", docsPath)
	}
}

func TestExtractModuleArchive_DeterministicBehavior(t *testing.T) {
	// Test that the same archive always produces the same result
	ctx := context.Background()

	archiveStructure := map[string]string{
		"Module-Z-1.0/":         "",
		"Module-Z-1.0/Build.PL": "use Module::Build;",
		"Module-A-1.0/":         "",
		"Module-A-1.0/Build.PL": "use Module::Build;",
		"Module-M-1.0/":         "",
		"Module-M-1.0/Build.PL": "use Module::Build;",
	}

	archivePath := createTestArchive(t, archiveStructure)
	defer os.Remove(archivePath)

	var results []*ExtractionResult
	for i := 0; i < 3; i++ {
		targetDir, err := os.MkdirTemp("", "test-deterministic-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(targetDir)

		result, err := extractModuleArchive(archivePath, targetDir, ctx)
		if err != nil {
			t.Fatalf("Failed to extract archive on iteration %d: %v", i, err)
		}
		results = append(results, result)
	}

	// All results should have the same root directory
	expectedRoot := "Module-A-1.0" // Should be alphabetically first
	for i, result := range results {
		if result.RootDir != expectedRoot {
			t.Errorf("Iteration %d: expected root directory '%s', got: %s", i, expectedRoot, result.RootDir)
		}
	}
}

func TestExtractModuleArchive_ContextCancellation(t *testing.T) {
	// Test that extraction properly handles context cancellation
	archiveStructure := map[string]string{
		"Module-1.0/":         "",
		"Module-1.0/Build.PL": "use Module::Build;",
	}

	archivePath := createTestArchive(t, archiveStructure)
	defer os.Remove(archivePath)

	targetDir, err := os.MkdirTemp("", "test-cancel-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Extraction should fail with context cancellation
	_, err = extractModuleArchive(archivePath, targetDir, ctx)
	if err == nil {
		t.Error("Expected error due to context cancellation, got none")
	}

	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("Expected cancellation error, got: %v", err)
	}
}

func TestSymlinkSecurity(t *testing.T) {
	// Test symlink security validation (cannot test directly as createTestArchive
	// doesn't support symlinks, but we test the sanitization logic)

	testCases := []struct {
		name       string
		linkTarget string
		shouldSkip bool
		reason     string
	}{
		{
			name:       "relative safe symlink",
			linkTarget: "lib/Module.pm",
			shouldSkip: false,
			reason:     "relative path should be allowed",
		},
		{
			name:       "absolute path attack",
			linkTarget: "/etc/passwd",
			shouldSkip: true,
			reason:     "absolute paths should be blocked",
		},
		{
			name:       "directory traversal attack",
			linkTarget: "../../../etc/passwd",
			shouldSkip: true,
			reason:     "directory traversal should be blocked",
		},
		{
			name:       "dot directory traversal",
			linkTarget: "./../../etc/passwd",
			shouldSkip: true,
			reason:     "dot-prefixed directory traversal should be blocked",
		},
		{
			name:       "safe subdirectory",
			linkTarget: "subdir/file.txt",
			shouldSkip: false,
			reason:     "safe subdirectory should be allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the sanitization logic
			sanitized := sanitizePath(tc.linkTarget)
			isEmpty := sanitized == ""
			isAbsolute := filepath.IsAbs(sanitized)

			shouldSkip := isEmpty || isAbsolute

			if shouldSkip != tc.shouldSkip {
				t.Errorf("Expected shouldSkip=%v for %s (%s), got shouldSkip=%v (sanitized='%s', isEmpty=%v, isAbsolute=%v)",
					tc.shouldSkip, tc.linkTarget, tc.reason, shouldSkip, sanitized, isEmpty, isAbsolute)
			}
		})
	}
}
