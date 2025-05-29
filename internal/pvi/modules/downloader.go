// ABOUTME: CPAN module download functionality
// ABOUTME: Functions for downloading CPAN modules from appropriate sources

package modules

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

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/xdg"
)

// Error codes for module download operations
const (
	ErrDownloadFailed     = "PVI-4001" // Failed to download module archive
	ErrChecksumMismatch   = "PVI-4002" // Checksum validation failed
	ErrCacheFailed        = "PVI-4003" // Failed to cache downloaded file
	ErrInvalidMirror      = "PVI-4004" // Invalid or unreachable mirror
	ErrInvalidDestination = "PVI-4005" // Invalid destination path
	ErrModuleNotFound     = "PVI-4006" // Module not found in the registry
)

// Default values and constants
const (
	MaxRetries  = 3       // Maximum number of download retries
	RetryDelay  = 3       // Delay between retries in seconds
	DefaultTTL  = 24 * 7  // Default cache TTL in hours (1 week)
	MaxProgress = 10      // Maximum number of progress updates per second
	MaxTimeout  = 10 * 60 // Maximum download timeout in seconds (10 minutes)
)

// ProgressCallback is a function that reports download progress
type ProgressCallback func(total, transferred int64, done bool)

// DownloadOptions contains options for downloading modules
type DownloadOptions struct {
	// Module name to download
	ModuleName string

	// Version constraint for the module
	VersionConstraint string

	// CPAN mirror URL to use
	Mirror string

	// Directory to store downloaded files
	CacheDir string

	// TTL for cached files in hours
	CacheTTL int

	// Skip using cache
	SkipCache bool

	// Progress callback function
	ProgressCallback ProgressCallback

	// Maximum number of retries for failed downloads
	MaxRetries int

	// Skip checksum validation
	SkipChecksum bool

	// CPAN provider for metadata
	Provider cpan.Provider

	// Context for cancellation
	Context context.Context
}

// DownloadResult contains information about the downloaded module
type DownloadResult struct {
	// Path to the downloaded file
	Path string

	// Module name
	ModuleName string

	// Version of the downloaded module
	Version string

	// Size of the downloaded file in bytes
	Size int64

	// Checksum of the downloaded file
	Checksum string

	// Whether the file was loaded from cache
	FromCache bool

	// Time taken to download
	Duration time.Duration

	// Distribution name (module distribution on CPAN)
	Distribution string

	// Author of the module
	Author string
}

// DownloadModuleFunc is the function type for downloading a CPAN module
type DownloadModuleFunc func(options *DownloadOptions) (*DownloadResult, error)

// DownloadModule is a variable that holds the module downloading function
// It can be replaced in tests
var DownloadModule DownloadModuleFunc = downloadModule

// downloadModule is the actual implementation of the CPAN module download function
func downloadModule(options *DownloadOptions) (*DownloadResult, error) {
	// Use default options if not specified
	if options == nil {
		options = &DownloadOptions{}
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Ensure we have a module name
	if options.ModuleName == "" {
		return nil, errors.NewSystemError(
			ErrModuleNotFound,
			"No module name specified",
			nil)
	}

	// Set default retries if not specified
	if options.MaxRetries <= 0 {
		options.MaxRetries = MaxRetries
	}

	// Ensure we have a provider
	if options.Provider == nil {
		return nil, errors.NewSystemError(
			ErrModuleNotFound,
			"No CPAN provider specified",
			nil)
	}

	// Get module information from provider
	moduleInfo, err := options.Provider.GetModuleInfo(options.Context, options.ModuleName)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrModuleNotFound,
			fmt.Sprintf("Failed to get information for module %s", options.ModuleName),
			err)
	}

	// Get cache directory
	cacheDir := options.CacheDir
	if cacheDir == "" {
		// Use XDG cache directory
		dirs, err := xdg.GetDirs()
		if err != nil {
			return nil, errors.NewSystemError("001",
				"Failed to determine XDG directories", err)
		}

		// Use modules cache directory
		cacheDir = filepath.Join(dirs.CacheDir, "modules")
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, errors.NewSystemError(
			ErrCacheFailed,
			"Failed to create cache directory",
			err).
			WithLocation(cacheDir)
	}

	// Determine destination filename from distribution file path
	log.Infof("Module info: Name=%s, Distribution=%s, DistributionVersion=%s, DistributionFile=%s", 
		moduleInfo.Name, moduleInfo.Distribution, moduleInfo.DistributionVersion, moduleInfo.DistributionFile)
	
	filename := filepath.Base(moduleInfo.DistributionFile)
	log.Debugf("Initial filename from DistributionFile: %s", filename)
	
	if filename == "" || filename == "." || filename == "/" {
		// Use distribution and version if file path not available
		filename = fmt.Sprintf("%s-%s.tar.gz", moduleInfo.Distribution, moduleInfo.DistributionVersion)
		log.Debugf("Using fallback filename: %s", filename)
	}
	
	// Additional safety check - ensure filename is not empty
	if filename == "" {
		filename = fmt.Sprintf("%s.tar.gz", moduleInfo.Name)
		log.Debugf("Using final fallback filename: %s", filename)
	}

	// Create cache path
	cachePath := filepath.Join(cacheDir, filename)
	log.Debugf("Cache path: %s", cachePath)

	// Check if file exists in cache
	if !options.SkipCache {
		if fileInfo, err := os.Stat(cachePath); err == nil && fileInfo.Size() > 0 {
			// Check TTL if specified
			if options.CacheTTL > 0 {
				modTime := fileInfo.ModTime()
				ttl := time.Duration(options.CacheTTL) * time.Hour
				if time.Since(modTime) > ttl {
					// Cache expired, will download again
					log.Debugf("Cache expired for %s", options.ModuleName)
				} else {
					// Cache still valid
					if !options.SkipChecksum {
						// Validate checksum if available and required
						checksum, err := calculateFileChecksum(cachePath)
						if err == nil {
							// In a real implementation, we would compare with known checksums
							// For now, we'll just accept the cached file
							log.Debugf("Using cached file for %s: %s", options.ModuleName, cachePath)
							return &DownloadResult{
								Path:         cachePath,
								ModuleName:   options.ModuleName,
								Version:      moduleInfo.Version,
								Size:         fileInfo.Size(),
								Checksum:     checksum,
								FromCache:    true,
								Duration:     0,
								Distribution: moduleInfo.Distribution,
								Author:       moduleInfo.Author,
							}, nil
						}
					} else {
						// Skip checksum validation and use cached file
						log.Debugf("Using cached file for %s: %s", options.ModuleName, cachePath)
						return &DownloadResult{
							Path:         cachePath,
							ModuleName:   options.ModuleName,
							Version:      moduleInfo.Version,
							Size:         fileInfo.Size(),
							FromCache:    true,
							Duration:     0,
							Distribution: moduleInfo.Distribution,
							Author:       moduleInfo.Author,
						}, nil
					}
				}
			} else {
				// No TTL specified, use cached file
				log.Debugf("Using cached file for %s: %s", options.ModuleName, cachePath)
				return &DownloadResult{
					Path:         cachePath,
					ModuleName:   options.ModuleName,
					Version:      moduleInfo.Version,
					Size:         fileInfo.Size(),
					FromCache:    true,
					Duration:     0,
					Distribution: moduleInfo.Distribution,
					Author:       moduleInfo.Author,
				}, nil
			}
		}
	}

	// Determine download URL from the mirror or module info
	var downloadURL string
	switch {
	case options.Mirror != "":
		// Use specified mirror
		// Format: mirror/authors/id/A/AU/AUTHOR/Distribution-Version.tar.gz
		// Extract author path from distribution file
		authorPath := filepath.Dir(moduleInfo.DistributionFile)
		downloadURL = fmt.Sprintf("%s/%s/%s", options.Mirror, authorPath, filename)
	case moduleInfo.DistributionFile == "":
		// Use default CPAN URL format if distribution file not provided
		firstChar := moduleInfo.Author[0:1]
		firstTwo := moduleInfo.Author[0:2]
		downloadURL = fmt.Sprintf("https://cpan.metacpan.org/authors/id/%s/%s/%s/%s",
			firstChar, firstTwo, moduleInfo.Author, filename)
	default:
		// Use the distribution file URL directly if it's already a full URL
		if strings.HasPrefix(moduleInfo.DistributionFile, "http://") || strings.HasPrefix(moduleInfo.DistributionFile, "https://") {
			downloadURL = moduleInfo.DistributionFile
		} else {
			// Use the distribution file path from module info with default CPAN URL
			downloadURL = fmt.Sprintf("https://cpan.metacpan.org/%s", moduleInfo.DistributionFile)
		}
	}

	// Download the file
	startTime := time.Now()
	var downloadErr error
	for retry := 0; retry < options.MaxRetries; retry++ {
		if retry > 0 {
			select {
			case <-options.Context.Done():
				return nil, errors.NewSystemError(
					ErrDownloadFailed,
					"Download cancelled",
					options.Context.Err())
			case <-time.After(time.Duration(RetryDelay) * time.Second):
				// Wait before retrying
			}
		}

		log.Debugf("Downloading %s from %s", options.ModuleName, downloadURL)
		downloadErr = downloadFile(downloadURL, cachePath, options)
		if downloadErr == nil {
			break
		}
	}

	if downloadErr != nil {
		return nil, errors.NewSystemError(
			ErrDownloadFailed,
			fmt.Sprintf("Failed to download module %s after %d attempts",
				options.ModuleName, options.MaxRetries),
			downloadErr)
	}

	// Get file info
	fileInfo, err := os.Stat(cachePath)
	if err != nil {
		return nil, errors.NewSystemError("003",
			"Failed to get information about downloaded file", err).
			WithLocation(cachePath)
	}

	// Calculate checksum
	checksum := ""
	if !options.SkipChecksum {
		checksum, err = calculateFileChecksum(cachePath)
		if err != nil {
			return nil, errors.NewSystemError(
				ErrChecksumMismatch,
				"Failed to calculate checksum for downloaded file",
				err).
				WithLocation(cachePath)
		}

		// In a real implementation, we would compare with known checksums
		// For now, we'll just accept the file without validation
	}

	// Return download result
	result := &DownloadResult{
		Path:         cachePath,
		ModuleName:   options.ModuleName,
		Version:      moduleInfo.Version,
		Size:         fileInfo.Size(),
		Checksum:     checksum,
		FromCache:    false,
		Duration:     time.Since(startTime),
		Distribution: moduleInfo.Distribution,
		Author:       moduleInfo.Author,
	}
	log.Debugf("Download result: Path=%s, ModuleName=%s, Size=%d", result.Path, result.ModuleName, result.Size)
	return result, nil
}

// downloadFile downloads a file from a URL to a destination path
func downloadFile(url, destPath string, options *DownloadOptions) error {
	// Create a new HTTP client with reasonable timeouts
	client := &http.Client{
		Timeout: time.Duration(MaxTimeout) * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(options.Context, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Add user agent
	req.Header.Set("User-Agent", "PVM/1.0 (Perl Version Manager; +https://github.com/perigrin/pvm)")

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
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
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
			maxUpdatesPerSec: MaxProgress,
		}
	}

	// Copy data from response to file
	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}

	// Close file before renaming
	if err := out.Close(); err != nil {
		return err
	}

	// Move temporary file to destination
	if err := os.Rename(tmpPath, destPath); err != nil {
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
	defer func() { _ = file.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
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
	maxUpdatesPerSec int
}

// Read reads from the underlying reader and reports progress
func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		r.read += int64(n)

		// Limit progress reports based on maxUpdatesPerSec
		now := time.Now()
		updateInterval := time.Second / time.Duration(r.maxUpdatesPerSec)
		if now.Sub(r.lastReport) >= updateInterval || err == io.EOF {
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
