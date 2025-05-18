// ABOUTME: Perl source downloading functionality
// ABOUTME: Provides functions to download Perl source code archives from mirrors

package perl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// Download error codes
const (
	ErrInvalidDownloadVersion = "401" // Invalid version format for download
	ErrDownloadFailed         = "402" // Failed to download source archive
	ErrChecksumMismatch       = "403" // Checksum validation failed
	ErrCacheFailed            = "404" // Failed to cache downloaded file
	ErrInvalidMirror          = "405" // Invalid or unreachable mirror
	ErrInvalidDestination     = "406" // Invalid destination path
)

// Default mirrors and other constants
const (
	DefaultMirror = "https://www.cpan.org/src/5.0"
	MaxRetries    = 3
	RetryDelay    = 3 * time.Second
)

// ProgressCallback is a function that reports download progress
type ProgressCallback func(total, transferred int64, done bool)

// DownloadOptions contains options for downloading Perl source
type DownloadOptions struct {
	// Mirror URL to use
	Mirror string

	// Version to download
	Version string

	// Progress callback function
	ProgressCallback ProgressCallback

	// Maximum number of retries for failed downloads
	MaxRetries int

	// Skip checksum validation
	SkipChecksum bool

	// Skip using cache
	SkipCache bool

	// Context for cancellation
	Context context.Context
}

// DownloadResult contains information about the downloaded source
type DownloadResult struct {
	// Path to the downloaded file
	Path string

	// Version of the downloaded Perl source
	Version string

	// Size of the downloaded file in bytes
	Size int64

	// Checksum of the downloaded file
	Checksum string

	// Whether the file was loaded from cache
	FromCache bool

	// Time taken to download
	Duration time.Duration
}

// GenerateSourceURL generates the URL for a Perl source archive
func GenerateSourceURL(mirror, version string) (string, error) {
	if mirror == "" {
		mirror = DefaultMirror
	}

	// Validate version format
	parsedVersion, err := ParseVersion(version)
	if err != nil {
		return "", errors.NewVersionError(
			ErrInvalidDownloadVersion,
			fmt.Sprintf("Invalid version format for source URL generation: %s", version),
			err)
	}

	// Make sure mirror URL doesn't have a trailing slash
	mirror = strings.TrimSuffix(mirror, "/")

	// Determine the correct archive pattern based on version
	// For versions before 5.14.0, the format is perl-5.X.Y.tar.gz
	// For versions 5.14.0 and newer, the format is perl-5.X.Y.tar.xz
	majorVersion := parsedVersion.Major
	minorVersion := parsedVersion.Minor
	patchVersion := parsedVersion.Patch

	var extension string
	if majorVersion > 5 || (majorVersion == 5 && (minorVersion > 14 || (minorVersion == 14 && patchVersion >= 0))) {
		extension = "tar.xz"
	} else {
		extension = "tar.gz"
	}

	// Construct filename
	filename := fmt.Sprintf("perl-%d.%d.%d.%s",
		majorVersion, minorVersion, patchVersion, extension)

	// Construct URL
	url := fmt.Sprintf("%s/%s", mirror, filename)

	return url, nil
}

// DownloadPerlSourceFunc holds the current implementation of DownloadPerlSource
// It can be replaced with a mock for testing
var DownloadPerlSourceFunc = func(options *DownloadOptions) (*DownloadResult, error) {
	return doDownloadPerlSource(options)
}

// DownloadPerlSource downloads a Perl source archive for the specified version
func DownloadPerlSource(options *DownloadOptions) (*DownloadResult, error) {
	return DownloadPerlSourceFunc(options)
}

// doDownloadPerlSource is the actual implementation of downloading a Perl source archive
func doDownloadPerlSource(options *DownloadOptions) (*DownloadResult, error) {
	// Use default options if not specified
	if options == nil {
		options = &DownloadOptions{}
	}

	// Set default mirror if not specified
	if options.Mirror == "" {
		options.Mirror = DefaultMirror
	}

	// Set default retries if not specified
	if options.MaxRetries <= 0 {
		options.MaxRetries = MaxRetries
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Validate version
	parsedVersion, err := ParseVersion(options.Version)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrInvalidDownloadVersion,
			fmt.Sprintf("Invalid version format: %s", options.Version),
			err)
	}
	version := parsedVersion.String()

	// Generate source URL
	url, err := GenerateSourceURL(options.Mirror, version)
	if err != nil {
		return nil, err
	}

	// Get cache directory
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Ensure cache directory exists
	err = dirs.EnsureDirs()
	if err != nil {
		return nil, errors.NewSystemError("002",
			"Failed to create required directories", err)
	}

	// Determine destination filename based on URL
	filename := filepath.Base(url)
	destPath := filepath.Join(dirs.SourcesDir, filename)

	// Check if file exists in cache
	if !options.SkipCache {
		if fileInfo, err := os.Stat(destPath); err == nil && fileInfo.Size() > 0 {
			// File exists in cache, validate checksum if required
			if !options.SkipChecksum {
				checksum, err := calculateFileChecksum(destPath)
				if err == nil {
					// In a real implementation, we would compare with known checksums
					// For now, we'll just accept the cached file without validation
					return &DownloadResult{
						Path:      destPath,
						Version:   version,
						Size:      fileInfo.Size(),
						Checksum:  checksum,
						FromCache: true,
						Duration:  0,
					}, nil
				}
			} else {
				// Skip checksum validation
				return &DownloadResult{
					Path:      destPath,
					Version:   version,
					Size:      fileInfo.Size(),
					FromCache: true,
					Duration:  0,
				}, nil
			}
		}
	}

	// Download the file
	startTime := time.Now()
	var downloadErr error
	for retry := 0; retry < options.MaxRetries; retry++ {
		if retry > 0 {
			select {
			case <-options.Context.Done():
				return nil, errors.NewVersionError(
					ErrDownloadFailed,
					"Download cancelled",
					options.Context.Err())
			case <-time.After(RetryDelay):
				// Wait before retrying
			}
		}

		downloadErr = downloadFile(url, destPath, options)
		if downloadErr == nil {
			break
		}
	}

	if downloadErr != nil {
		return nil, errors.NewVersionError(
			ErrDownloadFailed,
			fmt.Sprintf("Failed to download Perl source after %d attempts: %s", options.MaxRetries, url),
			downloadErr)
	}

	// Get file info
	fileInfo, err := os.Stat(destPath)
	if err != nil {
		return nil, errors.NewSystemError("003",
			"Failed to get information about downloaded file", err).
			WithLocation(destPath)
	}

	// Calculate checksum
	checksum := ""
	if !options.SkipChecksum {
		checksum, err = calculateFileChecksum(destPath)
		if err != nil {
			return nil, errors.NewVersionError(
				ErrChecksumMismatch,
				"Failed to calculate checksum for downloaded file",
				err).
				WithLocation(destPath)
		}

		// In a real implementation, we would compare with known checksums
		// For now, we'll just accept the file without validation
	}

	// Return download result
	return &DownloadResult{
		Path:      destPath,
		Version:   version,
		Size:      fileInfo.Size(),
		Checksum:  checksum,
		FromCache: false,
		Duration:  time.Since(startTime),
	}, nil
}

// downloadFile downloads a file from a URL to a destination path
func downloadFile(url, destPath string, options *DownloadOptions) error {
	// Create a new HTTP client with reasonable timeouts
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// Create request
	req, err := http.NewRequestWithContext(options.Context, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Create destination directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		return err
	}

	// Create a temporary file to download to
	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		// Clean up temporary file on error
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	// Get content length for progress reporting
	contentLength := resp.ContentLength

	// Create a reader that reports progress
	var reader io.Reader = resp.Body
	if options.ProgressCallback != nil {
		reader = &progressReader{
			reader:           resp.Body,
			total:            contentLength,
			progressCallback: options.ProgressCallback,
		}
	}

	// Copy data from response to file
	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}

	// Close file before renaming
	out.Close()

	// Move temporary file to destination
	err = os.Rename(tmpPath, destPath)
	if err != nil {
		return err
	}

	// Report final progress
	if options.ProgressCallback != nil {
		options.ProgressCallback(contentLength, contentLength, true)
	}

	return nil
}

// calculateFileChecksum calculates SHA-256 checksum of a file
func calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return checksum, nil
}

// progressReader is an io.Reader that reports progress
type progressReader struct {
	reader           io.Reader
	total            int64
	read             int64
	progressCallback ProgressCallback
	lastReport       time.Time
}

// Read reads from the underlying reader and reports progress
func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		r.read += int64(n)

		// Limit progress reports to 10 per second
		now := time.Now()
		if now.Sub(r.lastReport) >= 100*time.Millisecond || err == io.EOF {
			r.progressCallback(r.total, r.read, err == io.EOF)
			r.lastReport = now
		}
	}

	// If this is the end of the file, always report 100% progress
	if err == io.EOF {
		r.progressCallback(r.total, r.read, true)
	}

	return
}

// VerifyMirror checks if a mirror is accessible
func VerifyMirror(mirror string) error {
	if mirror == "" {
		mirror = DefaultMirror
	}

	// Create a new HTTP client with short timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make a HEAD request to the mirror
	resp, err := client.Head(mirror)
	if err != nil {
		return errors.NewVersionError(
			ErrInvalidMirror,
			fmt.Sprintf("Failed to connect to mirror: %s", mirror),
			err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode >= 400 {
		return errors.NewVersionError(
			ErrInvalidMirror,
			fmt.Sprintf("Mirror returned error: %s", resp.Status),
			nil)
	}

	return nil
}
