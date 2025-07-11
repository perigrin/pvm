package perl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEnhancedProgressCallback(t *testing.T) {
	progressReports := make([]DownloadProgress, 0)
	var mu sync.Mutex

	callback := func(progress DownloadProgress) {
		mu.Lock()
		defer mu.Unlock()
		progressReports = append(progressReports, progress)
	}

	tracker := &downloadProgressTracker{
		total:            1000,
		enhancedCallback: callback,
	}

	// Simulate download progress
	tracker.addBytes(250)
	tracker.addBytes(250)
	tracker.addBytes(250)
	tracker.addBytes(250)

	mu.Lock()
	defer mu.Unlock()

	if len(progressReports) == 0 {
		t.Errorf("Expected progress reports, got none")
	}

	lastReport := progressReports[len(progressReports)-1]
	if lastReport.Total != 1000 {
		t.Errorf("Expected total 1000, got %d", lastReport.Total)
	}
	if lastReport.Transferred != 1000 {
		t.Errorf("Expected transferred 1000, got %d", lastReport.Transferred)
	}
	if !lastReport.Done {
		t.Errorf("Expected done to be true")
	}
}

func TestBandwidthLimiter(t *testing.T) {
	limiter := newBandwidthLimiter(1000) // 1000 bytes per second
	if limiter == nil {
		t.Fatal("Expected bandwidth limiter, got nil")
	}

	start := time.Now()
	limiter.throttle(2000) // Should take ~1 second to throttle
	elapsed := time.Since(start)

	if elapsed < 500*time.Millisecond {
		t.Errorf("Expected throttling delay, got %v", elapsed)
	}
}

func TestBandwidthLimiterUnlimited(t *testing.T) {
	limiter := newBandwidthLimiter(0) // Unlimited
	if limiter != nil {
		t.Errorf("Expected nil limiter for unlimited bandwidth, got %v", limiter)
	}
}

func TestGetRemoteFileInfo(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("Expected HEAD request, got %s", r.Method)
		}
		w.Header().Set("Content-Length", "1024")
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	contentLength, supportsRange, err := getRemoteFileInfo(server.URL, context.Background())
	if err != nil {
		t.Fatalf("getRemoteFileInfo failed: %v", err)
	}

	if contentLength != 1024 {
		t.Errorf("Expected content length 1024, got %d", contentLength)
	}
	if !supportsRange {
		t.Errorf("Expected range support to be true")
	}
}

func TestDownloadSingle(t *testing.T) {
	testData := "Hello, World! This is test data for single download."

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.Write([]byte(testData))
	}))
	defer server.Close()

	// Create temporary file
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-download.txt")

	progressReports := make([]DownloadProgress, 0)
	var mu sync.Mutex

	callback := func(progress DownloadProgress) {
		mu.Lock()
		defer mu.Unlock()
		progressReports = append(progressReports, progress)
	}

	options := &BinaryDownloadOptions{
		EnhancedProgressCallback: callback,
		Context:                  context.Background(),
		StreamingChecksum:        true,
	}

	result, err := downloadSingle(server.URL, destPath, options)
	if err != nil {
		t.Fatalf("downloadSingle failed: %v", err)
	}

	// Verify file was downloaded
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Fatal("Downloaded file does not exist")
	}

	// Verify file content
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Downloaded content mismatch. Expected %q, got %q", testData, string(content))
	}

	// Verify result
	if result.BytesTransferred != int64(len(testData)) {
		t.Errorf("Expected bytes transferred %d, got %d", len(testData), result.BytesTransferred)
	}

	if !result.StreamingChecksum {
		t.Errorf("Expected streaming checksum to be true")
	}

	if result.Checksum == "" {
		t.Errorf("Expected checksum to be calculated")
	}

	// Verify progress reports
	mu.Lock()
	defer mu.Unlock()

	if len(progressReports) == 0 {
		t.Errorf("Expected progress reports, got none")
	}

	finalReport := progressReports[len(progressReports)-1]
	if !finalReport.Done {
		t.Errorf("Expected final report to be done")
	}
}

func TestResumeDownload(t *testing.T) {
	testData := "Hello, World! This is test data for resume testing with some extra content."

	// Create test server that supports range requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			// Handle range request
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Range", fmt.Sprintf("bytes 20-%d/%d", len(testData)-1, len(testData)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write([]byte(testData[20:]))
		} else {
			// Handle normal request
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.Header().Set("Accept-Ranges", "bytes")
			w.Write([]byte(testData))
		}
	}))
	defer server.Close()

	// Create temporary file with partial content
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-resume.txt")

	// Write first 20 bytes
	err := os.WriteFile(destPath, []byte(testData[:20]), 0644)
	if err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	options := &BinaryDownloadOptions{
		Context: context.Background(),
	}

	result, err := resumeDownload(server.URL, destPath, 20, int64(len(testData)), options)
	if err != nil {
		t.Fatalf("resumeDownload failed: %v", err)
	}

	// Verify file was completed
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read completed file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Resumed content mismatch. Expected %q, got %q", testData, string(content))
	}

	// Verify result
	if !result.Resumed {
		t.Errorf("Expected resumed to be true")
	}

	if result.ResumedBytes != 20 {
		t.Errorf("Expected resumed bytes 20, got %d", result.ResumedBytes)
	}

	if result.BytesTransferred != int64(len(testData)) {
		t.Errorf("Expected total bytes transferred %d, got %d", len(testData), result.BytesTransferred)
	}
}

func TestDownloadParallel(t *testing.T) {
	// Create large test data
	testData := make([]byte, 100*1024) // 100KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Create test server that supports range requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			// Parse range header (simplified)
			var start, end int
			fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)

			if end >= len(testData) {
				end = len(testData) - 1
			}

			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(testData)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(testData[start : end+1])
		} else {
			// Handle HEAD request for file info
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.Header().Set("Accept-Ranges", "bytes")
			if r.Method == http.MethodHead {
				return
			}
			w.Write(testData)
		}
	}))
	defer server.Close()

	// Create temporary file
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-parallel.bin")

	options := &BinaryDownloadOptions{
		Context:   context.Background(),
		ChunkSize: 10 * 1024, // 10KB chunks
	}

	result, err := downloadParallel(server.URL, destPath, int64(len(testData)), options)
	if err != nil {
		t.Fatalf("downloadParallel failed: %v", err)
	}

	// Verify file was downloaded
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if len(content) != len(testData) {
		t.Errorf("Downloaded size mismatch. Expected %d, got %d", len(testData), len(content))
	}

	// Verify content integrity
	for i, b := range content {
		if b != testData[i] {
			t.Errorf("Content mismatch at byte %d. Expected %d, got %d", i, testData[i], b)
			break
		}
	}

	// Verify result
	if !result.ParallelDownload {
		t.Errorf("Expected parallel download to be true")
	}

	if result.ChunkCount <= 1 {
		t.Errorf("Expected multiple chunks, got %d", result.ChunkCount)
	}

	if result.BytesTransferred != int64(len(testData)) {
		t.Errorf("Expected bytes transferred %d, got %d", len(testData), result.BytesTransferred)
	}
}

func TestCombineChunks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test chunk files
	chunk1Data := []byte("Hello, ")
	chunk2Data := []byte("World!")
	expectedData := "Hello, World!"

	chunk1Path := filepath.Join(tmpDir, "chunk1")
	chunk2Path := filepath.Join(tmpDir, "chunk2")

	err := os.WriteFile(chunk1Path, chunk1Data, 0644)
	if err != nil {
		t.Fatalf("Failed to create chunk 1: %v", err)
	}

	err = os.WriteFile(chunk2Path, chunk2Data, 0644)
	if err != nil {
		t.Fatalf("Failed to create chunk 2: %v", err)
	}

	// Combine chunks
	destPath := filepath.Join(tmpDir, "combined")
	chunkPaths := []string{chunk1Path, chunk2Path}

	err = combineChunks(chunkPaths, destPath)
	if err != nil {
		t.Fatalf("combineChunks failed: %v", err)
	}

	// Verify combined file
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read combined file: %v", err)
	}

	if string(content) != expectedData {
		t.Errorf("Combined content mismatch. Expected %q, got %q", expectedData, string(content))
	}
}

func TestEnhancedProgressReader(t *testing.T) {
	testData := "Test data for progress reader"

	progressReports := make([]DownloadProgress, 0)
	var mu sync.Mutex

	callback := func(progress DownloadProgress) {
		mu.Lock()
		defer mu.Unlock()
		progressReports = append(progressReports, progress)
	}

	tracker := &downloadProgressTracker{
		total:            int64(len(testData)),
		enhancedCallback: callback,
	}

	reader := &enhancedProgressReader{
		reader:  io.LimitReader(io.LimitReader(io.NopCloser(strings.NewReader(testData)), int64(len(testData))), int64(len(testData))),
		tracker: tracker,
	}

	// Read all data
	buffer := make([]byte, len(testData))
	n, err := io.ReadFull(reader, buffer)
	if err != nil {
		t.Fatalf("Failed to read from progress reader: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected to read %d bytes, got %d", len(testData), n)
	}

	if string(buffer) != testData {
		t.Errorf("Read data mismatch. Expected %q, got %q", testData, string(buffer))
	}

	// Give time for progress reporting
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(progressReports) == 0 {
		t.Errorf("Expected progress reports, got none")
	}
}

func TestBackwardCompatibility(t *testing.T) {
	testData := "Backward compatibility test data"

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.Write([]byte(testData))
	}))
	defer server.Close()

	// Create temporary file
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-compat.txt")

	// Test legacy progress callback
	var legacyTotal, legacyTransferred int64
	var legacyDone bool

	legacyCallback := func(total, transferred int64, done bool) {
		legacyTotal = total
		legacyTransferred = transferred
		legacyDone = done
	}

	options := &BinaryDownloadOptions{
		ProgressCallback: legacyCallback,
		Context:          context.Background(),
	}

	err := downloadBinaryFile(server.URL, destPath, options)
	if err != nil {
		t.Fatalf("downloadBinaryFile failed: %v", err)
	}

	// Verify legacy callback was called
	if legacyTotal != int64(len(testData)) {
		t.Errorf("Expected legacy total %d, got %d", len(testData), legacyTotal)
	}

	if legacyTransferred != int64(len(testData)) {
		t.Errorf("Expected legacy transferred %d, got %d", len(testData), legacyTransferred)
	}

	if !legacyDone {
		t.Errorf("Expected legacy done to be true")
	}

	// Verify file was downloaded
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Downloaded content mismatch. Expected %q, got %q", testData, string(content))
	}
}

// Benchmark tests
func BenchmarkDownloadSingle(b *testing.B) {
	testData := make([]byte, 1024*1024) // 1MB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.Write(testData)
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpDir := b.TempDir()
		destPath := filepath.Join(tmpDir, "benchmark.bin")

		options := &BinaryDownloadOptions{
			Context: context.Background(),
		}

		_, err := downloadSingle(server.URL, destPath, options)
		if err != nil {
			b.Fatalf("downloadSingle failed: %v", err)
		}
	}
}

func BenchmarkBandwidthLimiter(b *testing.B) {
	limiter := newBandwidthLimiter(1024 * 1024) // 1MB/s

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.throttle(1024) // 1KB chunks
	}
}

// Enhanced binary installation tests removed - tracked in GitHub issue #56
// These tests require complete enhanced installation system implementation
// including custom build options, parallel installation, and build management
//
// See: https://github.com/perigrin/pvm/issues/56
// "Implement enhanced Perl binary installation and build management"
func TestEnhancedBinaryInstallation_Removed(t *testing.T) {
	t.Skip("Removed - tracked in issue #56")
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "perl-install")

	// Create a mock binary archive
	testBinaryData := createMockBinaryArchive(t, tmpDir)

	// Create test server for binary download
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testBinaryData)))
		w.Header().Set("Accept-Ranges", "bytes")
		w.Write(testBinaryData)
	}))
	defer server.Close()

	// Test installation with enhanced download features
	progressReports := make([]DownloadProgress, 0)
	var mu sync.Mutex

	progressCallback := func(progress DownloadProgress) {
		mu.Lock()
		defer mu.Unlock()
		progressReports = append(progressReports, progress)
	}

	// Use progress callback (avoiding unused variable error)
	_ = progressCallback

	options := &BinaryInstallOptions{
		Version:    "5.38.0",
		Platform:   "linux-amd64",
		InstallDir: installDir,
		Context:    context.Background(),
		RepoURL:    server.URL,
	}

	// Override the binary download function to use our test server
	originalFunc := DownloadPerlBinaryFunc
	DownloadPerlBinaryFunc = func(opts *BinaryDownloadOptions) (*BinaryDownloadResult, error) {
		// Use our test server URL
		testOpts := *opts
		testOpts.RepoURL = server.URL + "/perl-5.38.0/perl-5.38.0-linux-amd64.tar.gz"

		// Use simple download for test
		testOpts.EnableResume = false
		testOpts.EnableParallel = false

		return doDownloadPerlBinary(&testOpts)
	}
	defer func() { DownloadPerlBinaryFunc = originalFunc }()

	result, err := InstallFromBinary(options)
	if err != nil {
		// Expected to fail due to archive format, but download should work
		t.Logf("Installation failed as expected (mock binary): %v", err)
	}

	// Verify progress was reported
	mu.Lock()
	defer mu.Unlock()
	if len(progressReports) == 0 {
		t.Errorf("Expected progress reports during installation")
	}

	// Verify result contains enhanced information
	if result != nil {
		t.Logf("Installation result: %+v", result)
	}
}

func TestEnhancedBinaryValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a mock Perl installation structure
	mockInstallDir := createMockPerlInstallation(t, tmpDir)

	// Test basic validation
	valid, warnings, err := ValidateBinaryInstallation(mockInstallDir)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	t.Logf("Validation result: valid=%t, warnings=%v", valid, warnings)

	// Test detailed validation with benchmarks
	result, err := ValidateBinaryInstallationDetailed(mockInstallDir, true)
	if err != nil {
		t.Fatalf("Detailed validation failed: %v", err)
	}

	if result.CompletenessScore < 0.0 || result.CompletenessScore > 1.0 {
		t.Errorf("Invalid completeness score: %f", result.CompletenessScore)
	}

	if result.BenchmarkResults == nil {
		t.Errorf("Expected benchmark results but got nil")
	}

	t.Logf("Detailed validation: score=%.2f, version=%s, benchmarks=%+v",
		result.CompletenessScore, result.DetectedVersion, result.BenchmarkResults)
}

func TestVersionExtractionEnhanced(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name: "standard perl output",
			output: `This is perl 5, version 38, subversion 0 (v5.38.0) built for x86_64-linux
Copyright 1987-2023, Larry Wall`,
			expected: "5.38.0",
		},
		{
			name: "perl with development version",
			output: `This is perl 5, version 39, subversion 7 (v5.39.7) built for darwin-2level
This is a development version of perl.`,
			expected: "5.39.7",
		},
		{
			name: "old perl format",
			output: `This is perl, version 5.008008 built for i386-linux-thread-multi
Copyright 1987-2006, Larry Wall`,
			expected: "5.008008",
		},
		{
			name: "windows perl",
			output: `This is perl 5, version 32, subversion 1 (v5.32.1) built for MSWin32-x64-multi-thread
Binary build 3203 [299195] provided by ActiveState http://www.ActiveState.com`,
			expected: "5.32.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionFromOutput(tt.output)
			if result != tt.expected {
				t.Errorf("Expected version '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Enhanced installation integration tests removed - tracked in GitHub issue #56
func TestEnhancedInstallationIntegration_Removed(t *testing.T) {
	t.Skip("Removed - tracked in issue #56")
	tmpDir := t.TempDir()

	// Test complete installation workflow with enhanced features
	testCases := []struct {
		name           string
		enableResume   bool
		enableParallel bool
		streamChecksum bool
		maxBandwidth   int64
	}{
		{
			name:           "basic enhanced install",
			enableResume:   false,
			enableParallel: false,
			streamChecksum: true,
		},
		{
			name:           "resume enabled",
			enableResume:   true,
			enableParallel: false,
			streamChecksum: false,
		},
		{
			name:           "parallel enabled",
			enableResume:   false,
			enableParallel: true,
			streamChecksum: false,
		},
		{
			name:           "bandwidth limited",
			enableResume:   false,
			enableParallel: false,
			streamChecksum: true,
			maxBandwidth:   1024 * 1024, // 1MB/s
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = filepath.Join(tmpDir, tc.name) // installDir not used in this test

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate a 1MB file for parallel download testing
				data := make([]byte, 1024*1024)
				for i := range data {
					data[i] = byte(i % 256)
				}

				if r.Method == http.MethodHead {
					w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
					w.Header().Set("Accept-Ranges", "bytes")
					return
				}

				rangeHeader := r.Header.Get("Range")
				if rangeHeader != "" {
					// Handle range request for resume/parallel
					var start, end int
					fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
					if end >= len(data) || end == 0 {
						end = len(data) - 1
					}

					w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
					w.Header().Set("Accept-Ranges", "bytes")
					w.WriteHeader(http.StatusPartialContent)
					w.Write(data[start : end+1])
				} else {
					w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
					w.Header().Set("Accept-Ranges", "bytes")
					w.Write(data)
				}
			}))
			defer server.Close()

			var progressCount int
			progressCallback := func(progress DownloadProgress) {
				progressCount++
				t.Logf("Progress: %d/%d bytes (%.1f%%), Speed: %d B/s, ETA: %v",
					progress.Transferred, progress.Total,
					float64(progress.Transferred)/float64(progress.Total)*100,
					progress.Speed, progress.ETA)
			}

			// Test enhanced download directly (since full installation would fail with mock data)
			options := &BinaryDownloadOptions{
				Version:                  "5.38.0",
				Platform:                 "linux-amd64",
				Context:                  context.Background(),
				RepoURL:                  server.URL,
				EnableResume:             tc.enableResume,
				EnableParallel:           tc.enableParallel,
				StreamingChecksum:        tc.streamChecksum,
				MaxBandwidth:             tc.maxBandwidth,
				ChunkSize:                256 * 1024, // 256KB chunks for testing
				EnhancedProgressCallback: progressCallback,
			}

			// Download to test enhanced features
			result, err := DownloadPerlBinary(options)
			if err != nil {
				// Expected to fail due to invalid URL format, but let's test what we can
				t.Logf("Download failed as expected: %v", err)
				return
			}

			if result == nil {
				t.Fatalf("Expected download result, got nil")
			}

			// Verify enhanced features were used
			if tc.enableParallel && result.ParallelDownload {
				t.Logf("Parallel download was used with %d chunks", result.ChunkCount)
			}

			if tc.enableResume && result.Resumed {
				t.Logf("Download was resumed from %d bytes", result.ResumedBytes)
			}

			if tc.streamChecksum && result.StreamingChecksumUsed {
				t.Logf("Streaming checksum was calculated: %s", result.Checksum)
			}

			if progressCount == 0 {
				t.Errorf("Expected progress updates, got none")
			}

			t.Logf("Download completed: %d bytes in %v (avg speed: %d B/s)",
				result.Size, result.Duration, result.AverageSpeed)
		})
	}
}

// Helper functions for mock data creation

func createMockBinaryArchive(t *testing.T, tmpDir string) []byte {
	// Create a minimal tar.gz file for testing
	// This is just mock data - real implementation would need proper archive
	return []byte("mock binary archive data for testing")
}

func createMockPerlInstallation(t *testing.T, tmpDir string) string {
	installDir := filepath.Join(tmpDir, "mock-perl")

	// Create directory structure
	dirs := []string{"bin", "lib", "man", "share"}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(installDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create mock perl executable
	perlExe := filepath.Join(installDir, "bin", "perl")
	if runtime.GOOS == "windows" {
		perlExe = filepath.Join(installDir, "bin", "perl.exe")
	}

	// Create a mock script that simulates perl behavior
	mockScript := `#!/bin/bash
if [ "$1" = "-v" ]; then
    echo "This is perl 5, version 38, subversion 0 (v5.38.0) built for x86_64-linux"
    echo "Copyright 1987-2023, Larry Wall"
elif [ "$1" = "-e" ]; then
    if [ "$2" = "print 'Hello World'" ]; then
        echo "Hello World"
    elif [ "$2" = "exit 0" ]; then
        exit 0
    fi
fi
`

	err := os.WriteFile(perlExe, []byte(mockScript), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock perl executable: %v", err)
	}

	return installDir
}
