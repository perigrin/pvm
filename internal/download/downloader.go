// ABOUTME: Core download functionality with progress tracking for PVM updates
// ABOUTME: Provides secure HTTPS downloads with progress callbacks and timeout handling

package download

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// ProgressCallback is called during download progress
type ProgressCallback func(total, transferred int64, done bool)

// Downloader handles secure file downloads
type Downloader struct {
	client    *http.Client
	userAgent string
}

// NewDownloader creates a new downloader with default settings
func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Minute, // Long timeout for large downloads
		},
		userAgent: "PVM-Updater/1.0",
	}
}

// NewDownloaderWithTimeout creates a downloader with custom timeout
func NewDownloaderWithTimeout(timeout time.Duration) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: timeout,
		},
		userAgent: "PVM-Updater/1.0",
	}
}

// DownloadOptions configures download behavior
type DownloadOptions struct {
	URL              string
	DestinationPath  string
	ExpectedChecksum string           // SHA256 checksum for validation
	ProgressCallback ProgressCallback // Called during download
	Context          context.Context  // For cancellation
	Resume           bool             // Whether to resume partial downloads
	MaxRetries       int              // Maximum number of retries (0 = no retries)
	RetryDelay       time.Duration    // Delay between retries
}

// DownloadResult contains information about the completed download
type DownloadResult struct {
	Path          string        // Final file path
	Size          int64         // File size in bytes
	Checksum      string        // SHA256 checksum of downloaded file
	Duration      time.Duration // Time taken to download
	FromCache     bool          // Whether file was already present and valid
	BytesResummed int64         // Bytes that were already downloaded (for resume)
}

// Download downloads a file with progress tracking and validation
func (d *Downloader) Download(opts *DownloadOptions) (*DownloadResult, error) {
	startTime := time.Now()

	if opts == nil {
		return nil, fmt.Errorf("download options cannot be nil")
	}

	if opts.Context == nil {
		opts.Context = context.Background()
	}

	// Check if file already exists and is valid
	if existingSize, exists := d.checkExistingFile(opts); exists {
		if opts.ExpectedChecksum != "" {
			if checksum, err := d.calculateChecksum(opts.DestinationPath); err == nil {
				if checksum == opts.ExpectedChecksum {
					return &DownloadResult{
						Path:      opts.DestinationPath,
						Size:      existingSize,
						Checksum:  checksum,
						Duration:  time.Since(startTime),
						FromCache: true,
					}, nil
				}
			}
		}
	}

	// Create destination directory
	destDir := filepath.Dir(opts.DestinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("creating destination directory: %w", err)
	}

	// Perform download with retries
	var lastErr error
	maxAttempts := opts.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			delay := opts.RetryDelay
			if delay == 0 {
				delay = time.Duration(attempt) * time.Second // Exponential backoff
			}

			select {
			case <-opts.Context.Done():
				return nil, opts.Context.Err()
			case <-time.After(delay):
			}
		}

		result, err := d.attemptDownload(opts, startTime)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if err == context.Canceled || err == context.DeadlineExceeded {
			break
		}
	}

	return nil, fmt.Errorf("download failed after %d attempts: %w", maxAttempts, lastErr)
}

// attemptDownload performs a single download attempt
func (d *Downloader) attemptDownload(opts *DownloadOptions, startTime time.Time) (*DownloadResult, error) {
	// Check for partial download if resume is enabled
	var resumeOffset int64
	if opts.Resume {
		if info, err := os.Stat(opts.DestinationPath); err == nil {
			resumeOffset = info.Size()
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(opts.Context, "GET", opts.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", d.userAgent)
	if resumeOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
	}

	// Perform request
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	switch {
	case resumeOffset > 0 && resp.StatusCode == http.StatusPartialContent:
		// Resuming download - OK
	case resp.StatusCode == http.StatusOK:
		// Full download - OK, reset resume offset
		resumeOffset = 0
	default:
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Get content length
	contentLength := resp.ContentLength
	totalSize := contentLength
	if resumeOffset > 0 {
		totalSize += resumeOffset
	}

	// Open destination file
	var outFile *os.File
	if resumeOffset > 0 {
		outFile, err = os.OpenFile(opts.DestinationPath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		outFile, err = os.Create(opts.DestinationPath)
	}
	if err != nil {
		return nil, fmt.Errorf("creating destination file: %w", err)
	}
	defer outFile.Close()

	// Download with progress tracking
	transferred, err := d.copyWithProgress(outFile, resp.Body, totalSize, resumeOffset, opts.ProgressCallback)
	if err != nil {
		// Clean up partial file on error (unless resuming)
		if resumeOffset == 0 {
			os.Remove(opts.DestinationPath)
		}
		return nil, fmt.Errorf("download failed: %w", err)
	}

	// Final progress callback
	if opts.ProgressCallback != nil {
		opts.ProgressCallback(totalSize, transferred, true)
	}

	// Calculate checksum if expected
	var checksum string
	if opts.ExpectedChecksum != "" {
		checksum, err = d.calculateChecksum(opts.DestinationPath)
		if err != nil {
			return nil, fmt.Errorf("calculating checksum: %w", err)
		}

		if checksum != opts.ExpectedChecksum {
			os.Remove(opts.DestinationPath)
			return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", opts.ExpectedChecksum, checksum)
		}
	}

	return &DownloadResult{
		Path:          opts.DestinationPath,
		Size:          transferred,
		Checksum:      checksum,
		Duration:      time.Since(startTime),
		FromCache:     false,
		BytesResummed: resumeOffset,
	}, nil
}

// copyWithProgress copies data while calling progress callback
func (d *Downloader) copyWithProgress(dst io.Writer, src io.Reader, totalSize, initialOffset int64, callback ProgressCallback) (int64, error) {
	const bufferSize = 32 * 1024 // 32KB buffer
	buffer := make([]byte, bufferSize)

	var transferred int64 = initialOffset
	lastUpdate := time.Now()
	updateInterval := 100 * time.Millisecond // Update progress every 100ms

	for {
		n, err := src.Read(buffer)
		if n > 0 {
			_, writeErr := dst.Write(buffer[:n])
			if writeErr != nil {
				return transferred, writeErr
			}

			transferred += int64(n)

			// Call progress callback periodically
			if callback != nil && time.Since(lastUpdate) >= updateInterval {
				callback(totalSize, transferred, false)
				lastUpdate = time.Now()
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return transferred, err
		}
	}

	return transferred, nil
}

// checkExistingFile checks if a file already exists and returns its size
func (d *Downloader) checkExistingFile(opts *DownloadOptions) (int64, bool) {
	info, err := os.Stat(opts.DestinationPath)
	if err != nil {
		return 0, false
	}

	if info.IsDir() {
		return 0, false
	}

	return info.Size(), true
}

// calculateChecksum calculates SHA256 checksum of a file
func (d *Downloader) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetFileSize gets the size of a remote file without downloading it
func (d *Downloader) GetFileSize(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("creating HEAD request: %w", err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("Content-Length header not found")
	}

	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing Content-Length: %w", err)
	}

	return size, nil
}
