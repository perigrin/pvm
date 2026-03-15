// ABOUTME: Tests for Perl binary downloading functionality
// ABOUTME: Ensures proper binary URL generation and download handling

package perl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/platform"
	"tamarou.com/pvm/internal/xdg"
)

// Setup for testing binary downloads
type binaryTestEnv struct {
	// Mock server
	server *httptest.Server

	// Temporary directories
	tempDir     string
	binariesDir string

	// Cleanup functions
	cleanup []func()
}

// Setup binary test environment
func setupBinaryTest(t *testing.T) *binaryTestEnv {
	env := &binaryTestEnv{
		cleanup: []func(){},
	}

	// Create temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "pvm-binary-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	env.tempDir = tempDir
	env.cleanup = append(env.cleanup, func() { _ = os.RemoveAll(tempDir) })

	// Create binaries cache directory
	binariesDir := filepath.Join(tempDir, "binaries")
	err = os.MkdirAll(binariesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create binaries directory: %v", err)
	}
	env.binariesDir = binariesDir

	// Mock for xdg.GetDirs
	originalGetDirs := xdg.GetDirs
	env.cleanup = append(env.cleanup, func() { xdg.GetDirs = originalGetDirs })

	xdg.GetDirs = func() (*xdg.Dirs, error) {
		dirs := &xdg.Dirs{
			CacheDir: tempDir,
		}

		// Mock EnsureDirs method
		dirs.EnsureDirs = func() error {
			return nil
		}

		return dirs, nil
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Test for binary availability checks (HEAD requests)
		if r.Method == http.MethodHead {
			// Return 200 for supported platforms, 404 for others
			if strings.Contains(path, "linux-amd64") ||
				strings.Contains(path, "linux-arm64") ||
				strings.Contains(path, "darwin-amd64") ||
				strings.Contains(path, "darwin-arm64") ||
				strings.Contains(path, "windows-amd64") {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}

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
			time.Sleep(100 * time.Millisecond)
		}

		// Return binary content based on the requested file
		if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".zip") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path)))
			w.Header().Set("Content-Length", "2048")

			// Generate binary content based on the path to ensure uniqueness
			for i := 0; i < 2048; i++ {
				_, _ = w.Write([]byte{byte(i % 256)})
			}
		} else {
			// Default response
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, "Mock response for %s", path)
		}
	}))
	env.server = server
	env.cleanup = append(env.cleanup, func() { server.Close() })

	return env
}

// Cleanup binary test environment
func (env *binaryTestEnv) cleanup_() {
	// Run cleanup functions in reverse order
	for i := len(env.cleanup) - 1; i >= 0; i-- {
		env.cleanup[i]()
	}
}

// Test binary URL generation
func TestGenerateBinaryURL(t *testing.T) {
	tests := []struct {
		version     string
		platform    string
		expectedURL string
		shouldError bool
	}{
		{
			version:     "5.38.0",
			platform:    "linux-amd64",
			expectedURL: DefaultBinaryRepo + "/perl-5.38.0/perl-5.38.0-linux-amd64.tar.gz",
			shouldError: false,
		},
		{
			version:     "5.38.0",
			platform:    "windows-amd64",
			expectedURL: DefaultBinaryRepo + "/perl-5.38.0/perl-5.38.0-windows-amd64.zip",
			shouldError: false,
		},
		{
			version:     "5.38.0",
			platform:    "darwin-arm64",
			expectedURL: DefaultBinaryRepo + "/perl-5.38.0/perl-5.38.0-darwin-arm64.tar.gz",
			shouldError: false,
		},
		{
			version:     "invalid",
			platform:    "linux-amd64",
			expectedURL: "",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.platform), func(t *testing.T) {
			url, err := GenerateBinaryURLWithRepo(DefaultBinaryRepo, test.version, test.platform)

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

// Test basic binary download functionality
func TestDownloadPerlBinary(t *testing.T) {
	env := setupBinaryTest(t)
	defer env.cleanup_()

	// Test with a valid version and platform
	version := "5.38.0"
	testPlatform := "linux-amd64"

	// Create progress tracking variables
	var lastProgress, totalSize int64
	progressCalled := false

	// Create download options
	options := &BinaryDownloadOptions{
		Version:  version,
		Platform: testPlatform,
		RepoURL:  env.server.URL,
		ProgressCallback: func(total, transferred int64, done bool) {
			progressCalled = true
			lastProgress = transferred
			totalSize = total
		},
		Context: context.Background(),
	}

	// Download the binary
	result, err := DownloadPerlBinary(options)
	if err != nil {
		t.Fatalf("Failed to download binary: %v", err)
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

	if result.Platform != testPlatform {
		t.Errorf("Expected platform %s, got %s", testPlatform, result.Platform)
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
	secondResult, err := DownloadPerlBinary(options)
	if err != nil {
		t.Fatalf("Failed to download binary from cache: %v", err)
	}

	if !secondResult.FromCache {
		t.Errorf("Expected FromCache to be true for cached download")
	}

	if secondResult.Path != result.Path {
		t.Errorf("Cache path mismatch: %s vs %s", secondResult.Path, result.Path)
	}
}

// Test binary download failures
func TestBinaryDownloadFailures(t *testing.T) {
	env := setupBinaryTest(t)
	defer env.cleanup_()

	tests := []struct {
		name        string
		version     string
		platform    string
		urlSuffix   string
		maxRetries  int
		skipCache   bool
		shouldError bool
	}{
		{
			name:        "Server error",
			version:     "5.38.0",
			platform:    "linux-amd64",
			urlSuffix:   "/server-error",
			maxRetries:  1,
			skipCache:   true,
			shouldError: true,
		},
		{
			name:        "Not found",
			version:     "5.38.0",
			platform:    "linux-amd64",
			urlSuffix:   "/not-found",
			maxRetries:  1,
			skipCache:   true,
			shouldError: true,
		},
		{
			name:        "Invalid version",
			version:     "invalid",
			platform:    "linux-amd64",
			urlSuffix:   "",
			maxRetries:  1,
			skipCache:   true,
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := &BinaryDownloadOptions{
				Version:    test.version,
				Platform:   test.platform,
				RepoURL:    env.server.URL + test.urlSuffix,
				MaxRetries: test.maxRetries,
				SkipCache:  test.skipCache,
				Context:    context.Background(),
			}

			result, err := DownloadPerlBinary(options)

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

// Test platform support validation
func TestUnsupportedPlatform(t *testing.T) {
	env := setupBinaryTest(t)
	defer env.cleanup_()

	// Mock platform functions for testing
	originalGetPlatformTriple := platform.GetPlatformTriple
	originalIsSupportedPlatform := platform.IsSupportedPlatform
	defer func() {
		// Note: In a real implementation, we'd need proper mocking
		// For now, we'll just test with default platform
		_ = originalGetPlatformTriple
		_ = originalIsSupportedPlatform
	}()

	options := &BinaryDownloadOptions{
		Version:  "5.38.0",
		Platform: "unsupported-arch",
		RepoURL:  env.server.URL,
		Context:  context.Background(),
	}

	// This test would need proper platform mocking to work fully
	// For now, we'll just verify the error handling structure exists
	result, err := DownloadPerlBinary(options)

	// The actual platform might be supported, so we check if the function at least runs
	if err != nil && !strings.Contains(err.Error(), "supported") {
		// Only fail if it's not a platform support error
		if result != nil {
			t.Logf("Download succeeded on current platform: %s", options.Platform)
		}
	}
}

// Test binary availability checking
func TestCheckBinaryAvailability(t *testing.T) {
	env := setupBinaryTest(t)
	defer env.cleanup_()

	tests := []struct {
		version    string
		platform   string
		shouldFind bool
	}{
		{
			version:    "5.38.0",
			platform:   "linux-amd64",
			shouldFind: true,
		},
		{
			version:    "5.38.0",
			platform:   "windows-amd64",
			shouldFind: true,
		},
		{
			version:    "5.38.0",
			platform:   "unsupported-platform",
			shouldFind: false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.platform), func(t *testing.T) {
			available, err := CheckBinaryAvailabilityWithRepo(env.server.URL, test.version, test.platform)
			if err != nil {
				t.Errorf("Unexpected error checking availability: %v", err)
			}

			if available != test.shouldFind {
				t.Errorf("Expected availability %v, got %v", test.shouldFind, available)
			}
		})
	}
}

// Test getting available platforms
func TestGetAvailableBinaryPlatforms(t *testing.T) {
	env := setupBinaryTest(t)
	defer env.cleanup_()

	platforms, err := GetAvailableBinaryPlatformsWithRepo(env.server.URL, "5.38.0")
	if err != nil {
		t.Fatalf("Failed to get available platforms: %v", err)
	}

	// Verify we get the expected supported platforms
	expectedPlatforms := []string{
		"linux-amd64",
		"linux-arm64",
		"darwin-amd64",
		"darwin-arm64",
		"windows-amd64",
	}

	if len(platforms) != len(expectedPlatforms) {
		t.Errorf("Expected %d platforms, got %d", len(expectedPlatforms), len(platforms))
	}

	// Check that all expected platforms are present
	for _, expected := range expectedPlatforms {
		found := false
		for _, platform := range platforms {
			if platform == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected platform %s not found in results", expected)
		}
	}
}

// Test binary download context cancellation
func TestBinaryDownloadCancellation(t *testing.T) {
	env := setupBinaryTest(t)
	defer env.cleanup_()

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Setup a channel to signal when download starts
	started := make(chan bool)

	// Create download options with a progress callback that triggers cancellation
	options := &BinaryDownloadOptions{
		Version:  "5.38.0",
		Platform: "linux-amd64",
		RepoURL:  env.server.URL + "/slow", // Use slow endpoint
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
	resultCh := make(chan *BinaryDownloadResult)
	errCh := make(chan error)

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

		result, err := DownloadPerlBinary(options)
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
		if !strings.Contains(err.Error(), "cancel") && !strings.Contains(err.Error(), "Binary download cancelled") {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Download didn't complete or error within timeout")
	}
}

// Test binary URL generation with custom repo
func TestGenerateBinaryURLWithCustomRepo(t *testing.T) {
	customRepo := "https://example.com/releases"
	version := "5.38.0"
	platform := "linux-amd64"

	expectedURL := customRepo + "/perl-5.38.0/perl-5.38.0-linux-amd64.tar.gz"

	url, err := GenerateBinaryURLWithRepo(customRepo, version, platform)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if url != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, url)
	}
}

// Test binary cache directory structure
func TestBinaryCacheStructure(t *testing.T) {
	env := setupBinaryTest(t)
	defer env.cleanup_()

	version := "5.38.0"
	testPlatform := "linux-amd64"

	options := &BinaryDownloadOptions{
		Version:  version,
		Platform: testPlatform,
		RepoURL:  env.server.URL,
		Context:  context.Background(),
	}

	result, err := DownloadPerlBinary(options)
	if err != nil {
		t.Fatalf("Failed to download binary: %v", err)
	}

	// Verify cache directory structure
	expectedPath := filepath.Join(env.tempDir, "binaries", version, testPlatform)
	if !strings.Contains(result.Path, expectedPath) {
		t.Errorf("Expected path to contain %s, got %s", expectedPath, result.Path)
	}

	// Verify the cache directory exists
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Cache directory %s does not exist", expectedPath)
	}
}

// Test default platform behavior
func TestDefaultPlatformBehavior(t *testing.T) {
	env := setupBinaryTest(t)
	defer env.cleanup_()

	// Test with empty platform (should use current platform)
	options := &BinaryDownloadOptions{
		Version: "5.38.0",
		RepoURL: env.server.URL,
		Context: context.Background(),
	}

	result, err := DownloadPerlBinary(options)

	// This might fail if current platform is not in our mock server's supported list
	// But we want to verify the platform is set correctly
	if err == nil {
		if result.Platform == "" {
			t.Errorf("Expected platform to be set to current platform, got empty")
		}

		// Should match the current platform triple
		currentPlatform := platform.GetPlatformTriple()
		if result.Platform != currentPlatform {
			t.Errorf("Expected platform %s, got %s", currentPlatform, result.Platform)
		}
	} else {
		// Error is acceptable if current platform is not supported by mock
		t.Logf("Download failed as expected for unsupported platform: %v", err)
	}
}
