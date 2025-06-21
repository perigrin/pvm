// ABOUTME: Perl binary downloading functionality
// ABOUTME: Provides functions to download pre-compiled Perl binary distributions

package perl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/platform"
	"tamarou.com/pvm/internal/xdg"
)

// Binary download error codes
const (
	ErrInvalidBinaryVersion      = "501" // Invalid version format for binary download
	ErrBinaryDownloadFailed      = "502" // Failed to download binary archive
	ErrBinaryChecksumFailed      = "503" // Binary checksum validation failed
	ErrBinaryCacheFailed         = "504" // Failed to cache downloaded binary
	ErrBinaryNotAvailable        = "505" // Binary not available for platform
	ErrBinaryPlatformUnsupported = "506" // Platform not supported for binaries
)

// Binary download constants
const (
	DefaultBinaryRepo = "https://github.com/example/pvm/releases/download"
	BinaryMaxRetries  = 3
	BinaryRetryDelay  = 2 * time.Second
)

// BinaryDownloadOptions contains options for downloading Perl binaries
type BinaryDownloadOptions struct {
	// Version to download
	Version string

	// Platform triple (e.g., "linux-amd64", "darwin-arm64")
	Platform string

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

	// Custom repository URL (defaults to GitHub Releases)
	RepoURL string
}

// BinaryDownloadResult contains information about the downloaded binary
type BinaryDownloadResult struct {
	// Path to the downloaded binary archive
	Path string

	// Version of the downloaded Perl binary
	Version string

	// Platform triple
	Platform string

	// Size of the downloaded file in bytes
	Size int64

	// Checksum of the downloaded file
	Checksum string

	// Whether the file was loaded from cache
	FromCache bool

	// Time taken to download
	Duration time.Duration
}

// GenerateBinaryURL generates the URL for a Perl binary archive
func GenerateBinaryURL(version, platform string) (string, error) {
	return GenerateBinaryURLWithRepo(DefaultBinaryRepo, version, platform)
}

// GenerateBinaryURLWithRepo generates the URL for a Perl binary archive with custom repo
func GenerateBinaryURLWithRepo(repoURL, version, platformTriple string) (string, error) {
	if repoURL == "" {
		repoURL = DefaultBinaryRepo
	}

	// Validate version format
	parsedVersion, err := ParseVersion(version)
	if err != nil {
		return "", errors.NewVersionError(
			ErrInvalidBinaryVersion,
			fmt.Sprintf("Invalid version format for binary URL generation: %s", version),
			err)
	}

	// Validate platform
	if platformTriple == "" {
		platformTriple = platform.GetPlatformTriple()
	}

	// Make sure repo URL doesn't have a trailing slash
	repoURL = strings.TrimSuffix(repoURL, "/")

	// Determine archive extension based on platform
	archiveExt := ".tar.gz"
	if strings.HasPrefix(platformTriple, "windows") {
		archiveExt = ".zip"
	}

	// Construct filename: perl-{version}-{platform}.{ext}
	filename := fmt.Sprintf("perl-%s-%s%s",
		parsedVersion.String(), platformTriple, archiveExt)

	// Construct URL: {repo}/perl-{version}/{filename}
	url := fmt.Sprintf("%s/perl-%s/%s", repoURL, parsedVersion.String(), filename)

	return url, nil
}

// DownloadPerlBinaryFunc holds the current implementation of DownloadPerlBinary
// It can be replaced with a mock for testing
var DownloadPerlBinaryFunc = doDownloadPerlBinary

// DownloadPerlBinary downloads a Perl binary archive for the specified version and platform
func DownloadPerlBinary(options *BinaryDownloadOptions) (*BinaryDownloadResult, error) {
	return DownloadPerlBinaryFunc(options)
}

// doDownloadPerlBinary is the actual implementation of downloading a Perl binary archive
func doDownloadPerlBinary(options *BinaryDownloadOptions) (*BinaryDownloadResult, error) {
	// Use default options if not specified
	if options == nil {
		options = &BinaryDownloadOptions{}
	}

	// Set default repo if not specified
	if options.RepoURL == "" {
		options.RepoURL = DefaultBinaryRepo
	}

	// Set default platform if not specified
	if options.Platform == "" {
		options.Platform = platform.GetPlatformTriple()
	}

	// Check if platform is supported
	if !platform.IsSupportedPlatform() {
		return nil, errors.NewVersionError(
			ErrBinaryPlatformUnsupported,
			fmt.Sprintf("Platform %s is not supported for binary downloads", options.Platform),
			nil)
	}

	// Set default retries if not specified
	if options.MaxRetries <= 0 {
		options.MaxRetries = BinaryMaxRetries
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Validate version
	parsedVersion, err := ParseVersion(options.Version)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrInvalidBinaryVersion,
			fmt.Sprintf("Invalid version format: %s", options.Version),
			err)
	}
	version := parsedVersion.String()

	// Generate binary URL
	url, err := GenerateBinaryURLWithRepo(options.RepoURL, version, options.Platform)
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

	// Create binaries cache subdirectory
	binariesDir := filepath.Join(dirs.CacheDir, "binaries", version, options.Platform)
	err = os.MkdirAll(binariesDir, 0755)
	if err != nil {
		return nil, errors.NewSystemError("003",
			"Failed to create binaries cache directory", err)
	}

	// Determine destination filename based on URL
	filename := filepath.Base(url)
	destPath := filepath.Join(binariesDir, filename)

	// Check if file exists in cache
	if !options.SkipCache {
		if fileInfo, err := os.Stat(destPath); err == nil && fileInfo.Size() > 0 {
			// File exists in cache, validate checksum if required
			if !options.SkipChecksum {
				checksum, err := calculateFileChecksum(destPath)
				if err == nil {
					// For now, we'll accept the cached file
					// In a real implementation, we would compare with known checksums
					return &BinaryDownloadResult{
						Path:      destPath,
						Version:   version,
						Platform:  options.Platform,
						Size:      fileInfo.Size(),
						Checksum:  checksum,
						FromCache: true,
						Duration:  0,
					}, nil
				}
			} else {
				// Skip checksum validation
				return &BinaryDownloadResult{
					Path:      destPath,
					Version:   version,
					Platform:  options.Platform,
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
					ErrBinaryDownloadFailed,
					"Binary download cancelled",
					options.Context.Err())
			case <-time.After(BinaryRetryDelay):
				// Wait before retrying
			}
		}

		downloadErr = downloadBinaryFile(url, destPath, options)
		if downloadErr == nil {
			break
		}
	}

	if downloadErr != nil {
		return nil, errors.NewVersionError(
			ErrBinaryDownloadFailed,
			fmt.Sprintf("Failed to download Perl binary after %d attempts: %s", options.MaxRetries, url),
			downloadErr)
	}

	// Get file info
	fileInfo, err := os.Stat(destPath)
	if err != nil {
		return nil, errors.NewSystemError("004",
			"Failed to get information about downloaded binary file", err).
			WithLocation(destPath)
	}

	// Calculate checksum
	checksum := ""
	if !options.SkipChecksum {
		checksum, err = calculateFileChecksum(destPath)
		if err != nil {
			return nil, errors.NewVersionError(
				ErrBinaryChecksumFailed,
				"Failed to calculate checksum for downloaded binary file",
				err).
				WithLocation(destPath)
		}

		// In a real implementation, we would compare with known checksums
		// For now, we'll just accept the file without validation
	}

	// Return download result
	return &BinaryDownloadResult{
		Path:      destPath,
		Version:   version,
		Platform:  options.Platform,
		Size:      fileInfo.Size(),
		Checksum:  checksum,
		FromCache: false,
		Duration:  time.Since(startTime),
	}, nil
}

// downloadBinaryFile downloads a binary file from a URL to a destination path
func downloadBinaryFile(url, destPath string, options *BinaryDownloadOptions) error {
	// Create a new HTTP client with reasonable timeouts
	client := &http.Client{
		Timeout: 10 * time.Minute, // Binaries might be larger than source
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
	defer func() { _ = resp.Body.Close() }()

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
		_ = out.Close()
		// Clean up temporary file on error
		if err != nil {
			_ = os.Remove(tmpPath)
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
	_ = out.Close()

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

// CheckBinaryAvailability checks if a binary is available for the specified version and platform
func CheckBinaryAvailability(version, platform string) (bool, error) {
	return CheckBinaryAvailabilityWithRepo(DefaultBinaryRepo, version, platform)
}

// CheckBinaryAvailabilityWithRepo checks binary availability with custom repo
func CheckBinaryAvailabilityWithRepo(repoURL, version, platformTriple string) (bool, error) {
	if repoURL == "" {
		repoURL = DefaultBinaryRepo
	}

	if platformTriple == "" {
		platformTriple = platform.GetPlatformTriple()
	}

	// Generate URL
	url, err := GenerateBinaryURLWithRepo(repoURL, version, platformTriple)
	if err != nil {
		return false, err
	}

	// Create a new HTTP client with short timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make a HEAD request to check if binary exists
	resp, err := client.Head(url)
	if err != nil {
		// Network error - return false but don't fail (allow graceful fallback)
		return false, nil
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	return resp.StatusCode == http.StatusOK, nil
}

// GetAvailableBinaryPlatforms returns a list of available platforms for a given version
func GetAvailableBinaryPlatforms(version string) ([]string, error) {
	return GetAvailableBinaryPlatformsWithRepo(DefaultBinaryRepo, version)
}

// GetAvailableBinaryPlatformsWithRepo returns available platforms with custom repo
func GetAvailableBinaryPlatformsWithRepo(repoURL, version string) ([]string, error) {
	if repoURL == "" {
		repoURL = DefaultBinaryRepo
	}

	// Define the platforms we support
	supportedPlatforms := []string{
		"linux-amd64",
		"linux-arm64",
		"darwin-amd64",
		"darwin-arm64",
		"windows-amd64",
	}

	var availablePlatforms []string
	for _, platform := range supportedPlatforms {
		available, err := CheckBinaryAvailabilityWithRepo(repoURL, version, platform)
		if err != nil {
			// Continue checking other platforms even if one fails
			continue
		}
		if available {
			availablePlatforms = append(availablePlatforms, platform)
		}
	}

	return availablePlatforms, nil
}
