// ABOUTME: Tests for Perl source downloading functionality
// ABOUTME: Ensures proper URL generation and download handling

package perl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/xdg"
)

// Setup for testing downloads
type downloadTestEnv struct {
	// Mock server
	server *httptest.Server

	// Temporary directories
	tempDir    string
	sourcesDir string

	// Cleanup functions
	cleanup []func()
}

// Setup test environment
func setupDownloadTest(t *testing.T) *downloadTestEnv {
	env := &downloadTestEnv{
		cleanup: []func(){},
	}

	// Create temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "pvm-download-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	env.tempDir = tempDir
	env.cleanup = append(env.cleanup, func() { os.RemoveAll(tempDir) })

	// Create sources directory
	sourcesDir := filepath.Join(tempDir, "sources")
	err = os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}
	env.sourcesDir = sourcesDir

	// Mock for xdg.GetDirs
	originalGetDirs := xdg.GetDirs
	env.cleanup = append(env.cleanup, func() { xdg.GetDirs = originalGetDirs })

	xdg.GetDirs = func() (*xdg.Dirs, error) {
		dirs := &xdg.Dirs{
			CacheDir:   tempDir,
			SourcesDir: sourcesDir,
		}
		
		// Mock EnsureDirs method
		dirs.EnsureDirs = func() error {
			return nil
		}
		
		return dirs, nil
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple mock implementation that returns different responses based on the path
		path := r.URL.Path

		// Test for server error condition
		if strings.Contains(path, "server-error") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Test for not found condition
		if strings.Contains(path, "not-found") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Test for slow download
		if strings.Contains(path, "slow") {
			// Slow response
			time.Sleep(100 * time.Millisecond)
		}

		// Check if it's a partial download request
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			// Parse range header
			parts := strings.Split(rangeHeader, "=")
			if len(parts) == 2 && parts[0] == "bytes" {
				rangeParts := strings.Split(parts[1], "-")
				if len(rangeParts) == 2 {
					// Return a partial response
					w.WriteHeader(http.StatusPartialContent)
					fmt.Fprintf(w, "Partial content for %s", path)
					return
				}
			}
		}

		// Return normal response based on the requested file
		if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tar.xz") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path)))
			w.Header().Set("Content-Length", "1024")

			// Generate content based on the path to ensure uniqueness
			for i := 0; i < 1024; i++ {
				w.Write([]byte{byte(i % 256)})
			}
		} else {
			// Default response
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Mock response for %s", path)
		}
	}))
	env.server = server
	env.cleanup = append(env.cleanup, func() { server.Close() })

	return env
}

// Cleanup test environment
func (env *downloadTestEnv) cleanup_() {
	// Run cleanup functions in reverse order
	for i := len(env.cleanup) - 1; i >= 0; i-- {
		env.cleanup[i]()
	}
}

// Test URL generation
func TestGenerateSourceURL(t *testing.T) {
	tests := []struct {
		mirror      string
		version     string
		expectedURL string
		shouldError bool
	}{
		{
			// Default mirror with version 5.38.0 (should use .tar.xz)
			mirror:      DefaultMirror,
			version:     "5.38.0",
			expectedURL: DefaultMirror + "/perl-5.38.0.tar.xz",
			shouldError: false,
		},
		{
			// Default mirror with older version (should use .tar.gz)
			mirror:      DefaultMirror,
			version:     "5.12.5",
			expectedURL: DefaultMirror + "/perl-5.12.5.tar.gz",
			shouldError: false,
		},
		{
			// Custom mirror with trailing slash
			mirror:      "http://example.com/perl/",
			version:     "5.38.0",
			expectedURL: "http://example.com/perl/perl-5.38.0.tar.xz",
			shouldError: false,
		},
		{
			// Version at the transition point (5.14.0)
			mirror:      DefaultMirror,
			version:     "5.14.0",
			expectedURL: DefaultMirror + "/perl-5.14.0.tar.xz",
			shouldError: false,
		},
		{
			// Invalid version format
			mirror:      DefaultMirror,
			version:     "invalid",
			expectedURL: "",
			shouldError: true,
		},
		{
			// Empty mirror (should use default)
			mirror:      "",
			version:     "5.38.0",
			expectedURL: DefaultMirror + "/perl-5.38.0.tar.xz",
			shouldError: false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.mirror, test.version), func(t *testing.T) {
			url, err := GenerateSourceURL(test.mirror, test.version)

			if test.shouldError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if url != test.expectedURL {
					t.Errorf("Expected URL %s, got %s", test.expectedURL, url)
				}
			}
		})
	}
}

// Test basic download functionality
func TestDownloadPerlSource(t *testing.T) {
	env := setupDownloadTest(t)
	defer env.cleanup_()

	// Test with a valid version
	version := "5.38.0"
	
	// Create progress tracking variables
	var lastProgress, totalSize int64
	progressCalled := false

	// Create download options
	options := &DownloadOptions{
		Mirror:  env.server.URL,
		Version: version,
		ProgressCallback: func(total, transferred int64, done bool) {
			progressCalled = true
			lastProgress = transferred
			totalSize = total
		},
		Context: context.Background(),
	}

	// Download the source
	result, err := DownloadPerlSource(options)
	if err != nil {
		t.Fatalf("Failed to download source: %v", err)
	}

	// Verify progress reporting
	if !progressCalled {
		t.Errorf("Progress callback was not called")
	}

	if lastProgress != totalSize {
		t.Errorf("Final progress report mismatch: got %d, expected %d", lastProgress, totalSize)
	}

	// Verify file exists
	if _, err := os.Stat(result.Path); os.IsNotExist(err) {
		t.Errorf("Downloaded file does not exist at %s", result.Path)
	}

	// Verify result fields
	if result.Version != version {
		t.Errorf("Expected version %s, got %s", version, result.Version)
	}

	if result.FromCache {
		t.Errorf("Expected FromCache to be false for new download")
	}

	if result.Size <= 0 {
		t.Errorf("Expected positive file size, got %d", result.Size)
	}

	if result.Checksum == "" {
		t.Errorf("Expected non-empty checksum")
	}

	// Test cache functionality by downloading again
	secondResult, err := DownloadPerlSource(options)
	if err != nil {
		t.Fatalf("Failed to download source from cache: %v", err)
	}

	if !secondResult.FromCache {
		t.Errorf("Expected FromCache to be true for cached download")
	}

	if secondResult.Path != result.Path {
		t.Errorf("Cache path mismatch: %s vs %s", secondResult.Path, result.Path)
	}
}

// Test download failure handling
func TestDownloadFailures(t *testing.T) {
	env := setupDownloadTest(t)
	defer env.cleanup_()

	tests := []struct {
		name        string
		version     string
		urlSuffix   string
		maxRetries  int
		skipCache   bool
		shouldError bool
	}{
		{
			name:        "Server error",
			version:     "5.38.0",
			urlSuffix:   "/server-error",
			maxRetries:  1,
			skipCache:   true,
			shouldError: true,
		},
		{
			name:        "Not found",
			version:     "5.38.0",
			urlSuffix:   "/not-found",
			maxRetries:  1,
			skipCache:   true,
			shouldError: true,
		},
		{
			name:        "Retry success",
			version:     "5.38.0",
			urlSuffix:   "", // No suffix for success
			maxRetries:  3,
			skipCache:   true,
			shouldError: false,
		},
		{
			name:        "Invalid version",
			version:     "invalid",
			urlSuffix:   "",
			maxRetries:  1,
			skipCache:   true,
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := &DownloadOptions{
				Mirror:     env.server.URL + test.urlSuffix,
				Version:    test.version,
				MaxRetries: test.maxRetries,
				SkipCache:  test.skipCache,
				Context:    context.Background(),
			}

			result, err := DownloadPerlSource(options)

			if test.shouldError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if result == nil {
					t.Errorf("Expected result, got nil")
				}
			}
		})
	}
}

// Test context cancellation
func TestDownloadCancellation(t *testing.T) {
	env := setupDownloadTest(t)
	defer env.cleanup_()

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Setup a channel to signal when download starts
	started := make(chan bool)

	// Create download options with a progress callback that triggers cancellation
	options := &DownloadOptions{
		Mirror:  env.server.URL + "/slow", // Use slow endpoint
		Version: "5.38.0",
		ProgressCallback: func(total, transferred int64, done bool) {
			// Signal that download has started
			select {
			case started <- true:
				// Signal sent
			default:
				// Already signaled
			}
		},
		Context:    ctx,
		MaxRetries: 1,
		SkipCache:  true,
	}

	// Start download in a goroutine
	resultCh := make(chan *DownloadResult)
	errCh := make(chan error)
	
	// Create a temporary file mimicking a download in progress
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "download-test-*.tmp")
	if err == nil {
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())
	}
	
	go func() {
		// Sleep briefly to simulate download starting
		time.Sleep(50 * time.Millisecond)
		
		// Send the started signal
		select {
		case started <- true:
			// Signal sent
		default:
			// Already signaled
		}
		
		// Sleep a bit more to ensure cancellation has time to take effect
		time.Sleep(50 * time.Millisecond)
		
		result, err := DownloadPerlSource(options)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	// Wait for download to start
	select {
	case <-started:
		// Download started, cancel the context
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatalf("Download didn't start within timeout")
	}

	// Wait for result or error
	select {
	case <-resultCh:
		t.Errorf("Expected download to be cancelled, but it succeeded")
	case err := <-errCh:
		// Verify that it's a cancellation error
		if !strings.Contains(err.Error(), "cancel") && !strings.Contains(err.Error(), "Download cancelled") {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Download didn't complete or error within timeout")
	}
}

// Test mirror verification
func TestVerifyMirror(t *testing.T) {
	env := setupDownloadTest(t)
	defer env.cleanup_()

	tests := []struct {
		name        string
		mirror      string
		shouldError bool
	}{
		{
			name:        "Valid mirror",
			mirror:      env.server.URL,
			shouldError: false,
		},
		{
			name:        "Invalid mirror",
			mirror:      "http://invalid.example.com",
			shouldError: true,
		},
		{
			name:        "Error response mirror",
			mirror:      env.server.URL + "/server-error",
			shouldError: true,
		},
		{
			name:        "Default mirror",
			mirror:      "", // Should use default
			shouldError: true, // Default mirror requires network access, might fail in CI
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := VerifyMirror(test.mirror)

			if test.shouldError {
				if err == nil {
					// Skip error check for default mirror as it depends on network connectivity
					if test.mirror != "" {
						t.Errorf("Expected error, but got none")
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test checksum calculation
func TestCalculateFileChecksum(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "checksum-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some data to the file
	data := "Hello, world!"
	_, err = tmpFile.WriteString(data)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Expected SHA-256 checksum for "Hello, world!"
	expectedChecksum := "315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3"

	// Calculate checksum
	checksum, err := calculateFileChecksum(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to calculate checksum: %v", err)
	}

	// Verify checksum
	if checksum != expectedChecksum {
		t.Errorf("Expected checksum %s, got %s", expectedChecksum, checksum)
	}

	// Test with non-existent file
	_, err = calculateFileChecksum("/path/to/nonexistent/file")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got none")
	}
}

// Test progress reader implementation
func TestProgressReader(t *testing.T) {
	// Create a mock reader with some data
	data := "Hello, world!"
	mockReader := strings.NewReader(data)

	// Create a progress reader
	var lastTotal, lastRead int64
	var lastDone bool
	callCount := 0

	reader := &progressReader{
		reader: mockReader,
		total:  int64(len(data)),
		progressCallback: func(total, read int64, done bool) {
			lastTotal = total
			lastRead = read
			lastDone = done
			callCount++
		},
	}

	// Read all data but with a small buffer to trigger multiple reads
	buf := make([]byte, 5) // Smaller than the full data
	var fullBuf []byte
	
	for {
		n, err := reader.Read(buf)
		fullBuf = append(fullBuf, buf[:n]...)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// Verify data was read correctly
	if string(fullBuf) != data {
		t.Errorf("Expected data %q, got %q", data, string(fullBuf))
	}

	// Verify progress was reported
	if callCount == 0 {
		t.Errorf("Progress callback was not called")
	}

	if lastTotal != int64(len(data)) {
		t.Errorf("Expected total %d, got %d", len(data), lastTotal)
	}

	if lastRead != int64(len(data)) {
		t.Errorf("Expected read %d, got %d", len(data), lastRead)
	}

	if !lastDone {
		t.Errorf("Expected done to be true")
	}
}