// ABOUTME: Tests for core download functionality and progress tracking
// ABOUTME: Validates secure downloads, progress callbacks, and error handling

package download

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewDownloader(t *testing.T) {
	downloader := NewDownloader()

	if downloader == nil {
		t.Fatal("NewDownloader returned nil")
	}

	if downloader.userAgent != "PVM-Updater/1.0" {
		t.Errorf("Expected user agent 'PVM-Updater/1.0', got '%s'", downloader.userAgent)
	}

	if downloader.client.Timeout != 30*time.Minute {
		t.Errorf("Expected timeout 30m, got %v", downloader.client.Timeout)
	}
}

func TestNewDownloaderWithTimeout(t *testing.T) {
	timeout := 5 * time.Minute
	downloader := NewDownloaderWithTimeout(timeout)

	if downloader.client.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, downloader.client.Timeout)
	}
}

func TestDownloadBasic(t *testing.T) {
	// Create test server
	testData := "Hello, PVM world!"
	expectedChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(testData)))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testData))
	}))
	defer server.Close()

	// Create temp directory for test
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test-file.txt")

	// Test download
	downloader := NewDownloader()
	opts := &DownloadOptions{
		URL:              server.URL,
		DestinationPath:  destPath,
		ExpectedChecksum: expectedChecksum,
		Context:          context.Background(),
	}

	result, err := downloader.Download(opts)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Download result is nil")
	}

	if result.Path != destPath {
		t.Errorf("Expected path %s, got %s", destPath, result.Path)
	}

	if result.Size != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), result.Size)
	}

	if result.Checksum != expectedChecksum {
		t.Errorf("Expected checksum %s, got %s", expectedChecksum, result.Checksum)
	}

	if result.FromCache {
		t.Error("Expected FromCache to be false for fresh download")
	}

	// Verify file was created correctly
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(data) != testData {
		t.Errorf("Expected file content '%s', got '%s'", testData, string(data))
	}
}

func TestDownloadWithProgress(t *testing.T) {
	testData := strings.Repeat("A", 1024) // 1KB of data

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testData))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test-file.txt")

	// Track progress calls
	var progressCalls []struct {
		total, transferred int64
		done               bool
	}

	progressCallback := func(total, transferred int64, done bool) {
		progressCalls = append(progressCalls, struct {
			total, transferred int64
			done               bool
		}{total, transferred, done})
	}

	downloader := NewDownloader()
	opts := &DownloadOptions{
		URL:              server.URL,
		DestinationPath:  destPath,
		ProgressCallback: progressCallback,
		Context:          context.Background(),
	}

	_, err := downloader.Download(opts)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify progress callbacks were called
	if len(progressCalls) == 0 {
		t.Error("Expected progress callbacks to be called")
	}

	// Verify final callback marked as done
	lastCall := progressCalls[len(progressCalls)-1]
	if !lastCall.done {
		t.Error("Expected final progress callback to be marked as done")
	}

	if lastCall.transferred != int64(len(testData)) {
		t.Errorf("Expected final transferred %d, got %d", len(testData), lastCall.transferred)
	}
}

func TestDownloadFromCache(t *testing.T) {
	testData := "Cached content"
	expectedChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(testData)))

	// Create temp directory and pre-existing file
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "cached-file.txt")

	// Write file that should be found in cache
	err := os.WriteFile(destPath, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create cached file: %v", err)
	}

	// Server should not be called for cached download
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not be called for cached download")
	}))
	defer server.Close()

	downloader := NewDownloader()
	opts := &DownloadOptions{
		URL:              server.URL,
		DestinationPath:  destPath,
		ExpectedChecksum: expectedChecksum,
		Context:          context.Background(),
	}

	result, err := downloader.Download(opts)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if !result.FromCache {
		t.Error("Expected download to be served from cache")
	}

	if result.Size != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), result.Size)
	}
}

func TestDownloadChecksumMismatch(t *testing.T) {
	testData := "Test data"
	wrongChecksum := "deadbeef" // Intentionally wrong

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testData))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test-file.txt")

	downloader := NewDownloader()
	opts := &DownloadOptions{
		URL:              server.URL,
		DestinationPath:  destPath,
		ExpectedChecksum: wrongChecksum,
		Context:          context.Background(),
	}

	_, err := downloader.Download(opts)
	if err == nil {
		t.Fatal("Expected checksum mismatch error")
	}

	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("Expected checksum mismatch error, got: %v", err)
	}

	// Verify file was cleaned up after checksum failure
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		t.Error("Expected file to be cleaned up after checksum failure")
	}
}

func TestDownloadHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test-file.txt")

	downloader := NewDownloader()
	opts := &DownloadOptions{
		URL:             server.URL,
		DestinationPath: destPath,
		Context:         context.Background(),
	}

	_, err := downloader.Download(opts)
	if err == nil {
		t.Fatal("Expected HTTP error")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected 404 error, got: %v", err)
	}
}

func TestDownloadCancellation(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test-file.txt")

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	downloader := NewDownloader()
	opts := &DownloadOptions{
		URL:             server.URL,
		DestinationPath: destPath,
		Context:         ctx,
	}

	// Cancel context immediately
	cancel()

	_, err := downloader.Download(opts)
	if err == nil {
		t.Fatal("Expected cancellation error")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

func TestDownloadWithResume(t *testing.T) {
	fullData := "This is a test file for resume functionality"
	partialData := fullData[:20]   // First 20 bytes
	remainingData := fullData[20:] // Rest of the data

	// Create temp directory and partial file
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "resume-test.txt")

	// Write partial file
	err := os.WriteFile(destPath, []byte(partialData), 0644)
	if err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	// Server that supports range requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			// Handle range request
			if rangeHeader == "bytes=20-" {
				w.Header().Set("Content-Range", fmt.Sprintf("bytes 20-%d/%d", len(fullData)-1, len(fullData)))
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(remainingData)))
				w.WriteHeader(http.StatusPartialContent)
				w.Write([]byte(remainingData))
				return
			}
		}

		// Full request
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fullData)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fullData))
	}))
	defer server.Close()

	downloader := NewDownloader()
	opts := &DownloadOptions{
		URL:             server.URL,
		DestinationPath: destPath,
		Resume:          true,
		Context:         context.Background(),
	}

	result, err := downloader.Download(opts)
	if err != nil {
		t.Fatalf("Resume download failed: %v", err)
	}

	if result.BytesResummed != int64(len(partialData)) {
		t.Errorf("Expected %d bytes resumed, got %d", len(partialData), result.BytesResummed)
	}

	// Verify complete file
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read resumed file: %v", err)
	}

	if string(data) != fullData {
		t.Errorf("Expected complete file content, got partial: %s", string(data))
	}
}

func TestGetFileSize(t *testing.T) {
	testData := "Size test data"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	downloader := NewDownloader()
	size, err := downloader.GetFileSize(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("GetFileSize failed: %v", err)
	}

	if size != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), size)
	}
}

func TestDownloadNilOptions(t *testing.T) {
	downloader := NewDownloader()
	_, err := downloader.Download(nil)

	if err == nil {
		t.Fatal("Expected error for nil options")
	}

	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Expected nil options error, got: %v", err)
	}
}
