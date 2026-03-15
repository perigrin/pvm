// ABOUTME: Perl binary downloading functionality
// ABOUTME: Provides functions to download pre-compiled Perl binary distributions

package perl

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/config"
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
	DefaultBinaryRepo         = "https://github.com/perigrin/pvm/releases/download"
	BinaryMaxRetries          = 3
	BinaryRetryDelay          = 2 * time.Second
	DefaultChunkSize          = 8 * 1024 * 1024  // 8MB chunks for parallel downloads
	ParallelDownloadThreshold = 50 * 1024 * 1024 // 50MB minimum for parallel downloads
	DefaultMaxBandwidth       = 0                // 0 means unlimited
	ProgressReportInterval    = 100 * time.Millisecond
	ChecksumBufferSize        = 64 * 1024 // 64KB buffer for streaming checksum
)

// setGitHubAuthIfAvailable sets GitHub authentication header if GH_TOKEN is available
func setGitHubAuthIfAvailable(req *http.Request) {
	if token := os.Getenv("GH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// BinaryDownloadOptions contains options for downloading Perl binaries
type BinaryDownloadOptions struct {
	// Version to download
	Version string

	// Platform triple (e.g., "linux-amd64", "darwin-arm64")
	Platform string

	// Progress callback function (legacy)
	ProgressCallback ProgressCallback

	// Enhanced progress callback function
	EnhancedProgressCallback EnhancedProgressCallback

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

	// Enable resumable downloads
	EnableResume bool

	// Enable parallel chunk downloads for large files
	EnableParallel bool

	// Chunk size for parallel downloads (bytes)
	ChunkSize int64

	// Maximum download bandwidth (bytes per second, 0 = unlimited)
	MaxBandwidth int64

	// Verify checksum during transfer (streaming validation)
	StreamingChecksum bool
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

	// Whether download was resumed
	Resumed bool

	// Number of bytes already downloaded when resuming
	ResumedBytes int64

	// Whether parallel downloading was used
	ParallelDownload bool

	// Number of chunks used for parallel download
	ChunkCount int

	// Average download speed in bytes per second
	AverageSpeed int64

	// Whether streaming checksum validation was used
	StreamingChecksumUsed bool
}

// DownloadProgress contains detailed progress information
type DownloadProgress struct {
	// Total size in bytes
	Total int64
	// Bytes transferred so far
	Transferred int64
	// Current download speed in bytes per second
	Speed int64
	// Estimated time to completion
	ETA time.Duration
	// Whether download is complete
	Done bool
	// Number of active chunks (for parallel downloads)
	ActiveChunks int
	// Current chunk being downloaded (for parallel downloads)
	CurrentChunk int
}

// EnhancedProgressCallback is a function that reports enhanced download progress
type EnhancedProgressCallback func(progress DownloadProgress)

// Convert old progress callback to enhanced version for backward compatibility
func wrapProgressCallback(callback ProgressCallback) EnhancedProgressCallback {
	if callback == nil {
		return nil
	}
	return func(progress DownloadProgress) {
		callback(progress.Total, progress.Transferred, progress.Done)
	}
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

	// Set default chunk size if not specified
	if options.ChunkSize <= 0 {
		options.ChunkSize = DefaultChunkSize
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

	// Download the file with enhanced features
	startTime := time.Now()
	var downloadResult *enhancedDownloadResult
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

		downloadResult, downloadErr = downloadBinaryFileEnhanced(url, destPath, options)
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

	// Calculate final checksum if not already done during streaming
	checksum := downloadResult.Checksum
	if !options.SkipChecksum && checksum == "" {
		checksum, err = calculateFileChecksum(destPath)
		if err != nil {
			return nil, errors.NewVersionError(
				ErrBinaryChecksumFailed,
				"Failed to calculate checksum for downloaded binary file",
				err).
				WithLocation(destPath)
		}
	}

	downloadDuration := time.Since(startTime)
	var averageSpeed int64
	if downloadDuration > time.Millisecond {
		durationSeconds := downloadDuration.Seconds()
		if durationSeconds > 0 {
			averageSpeed = int64(float64(downloadResult.BytesTransferred) / durationSeconds)
		}
	}

	// Return enhanced download result
	return &BinaryDownloadResult{
		Path:                  destPath,
		Version:               version,
		Platform:              options.Platform,
		Size:                  fileInfo.Size(),
		Checksum:              checksum,
		FromCache:             false,
		Duration:              downloadDuration,
		Resumed:               downloadResult.Resumed,
		ResumedBytes:          downloadResult.ResumedBytes,
		ParallelDownload:      downloadResult.ParallelDownload,
		ChunkCount:            downloadResult.ChunkCount,
		AverageSpeed:          averageSpeed,
		StreamingChecksumUsed: downloadResult.StreamingChecksum,
	}, nil
}

// enhancedDownloadResult contains detailed information about an enhanced download
type enhancedDownloadResult struct {
	BytesTransferred  int64
	Resumed           bool
	ResumedBytes      int64
	ParallelDownload  bool
	ChunkCount        int
	StreamingChecksum bool
	Checksum          string
}

// downloadBinaryFileEnhanced downloads a binary file with enhanced features
func downloadBinaryFileEnhanced(url, destPath string, options *BinaryDownloadOptions) (*enhancedDownloadResult, error) {
	// Get remote file size to determine download strategy
	contentLength, supportsRange, err := getRemoteFileInfo(url, options.Context)
	if err != nil {
		return nil, err
	}

	// Check if we should use parallel downloads
	useParallel := options.EnableParallel &&
		contentLength > ParallelDownloadThreshold &&
		supportsRange

	// Check if we should try to resume
	existingSize := int64(0)
	canResume := false
	if options.EnableResume {
		if stat, err := os.Stat(destPath); err == nil {
			existingSize = stat.Size()
			canResume = existingSize > 0 && existingSize < contentLength && supportsRange
		}
	}

	_ = &enhancedDownloadResult{
		StreamingChecksum: options.StreamingChecksum,
	}

	switch {
	case useParallel && !canResume:
		// Use parallel download for new large files
		return downloadParallel(url, destPath, contentLength, options)
	case canResume:
		// Resume existing download
		return resumeDownload(url, destPath, existingSize, contentLength, options)
	default:
		// Use single-threaded download
		return downloadSingle(url, destPath, options)
	}
}

// getRemoteFileInfo gets file size and range support from remote server
func getRemoteFileInfo(url string, ctx context.Context) (contentLength int64, supportsRange bool, err error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, false, err
	}
	setGitHubAuthIfAvailable(req)

	resp, err := client.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	contentLength = resp.ContentLength
	acceptRanges := resp.Header.Get("Accept-Ranges")
	supportsRange = strings.Contains(acceptRanges, "bytes")

	return contentLength, supportsRange, nil
}

// downloadSingle performs a single-threaded download
func downloadSingle(url, destPath string, options *BinaryDownloadOptions) (*enhancedDownloadResult, error) {
	client := &http.Client{Timeout: 10 * time.Minute}

	req, err := http.NewRequestWithContext(options.Context, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	setGitHubAuthIfAvailable(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	err = os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		return nil, err
	}

	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	// Setup progress tracking and optional bandwidth limiting
	progressTracker := &downloadProgressTracker{
		total:            resp.ContentLength,
		enhancedCallback: options.EnhancedProgressCallback,
		legacyCallback:   options.ProgressCallback,
		bandwidthLimiter: newBandwidthLimiter(options.MaxBandwidth),
	}

	// Setup streaming checksum if enabled
	var hasher hash.Hash
	var writer io.Writer = out
	if options.StreamingChecksum {
		hasher = sha256.New()
		writer = io.MultiWriter(out, hasher)
	}

	// Create progress reader
	reader := &enhancedProgressReader{
		reader:  resp.Body,
		tracker: progressTracker,
	}

	// Copy data with progress and bandwidth limiting
	bytesTransferred, err := io.Copy(writer, reader)
	if err != nil {
		return nil, err
	}

	out.Close()

	// Move temporary file to destination
	err = os.Rename(tmpPath, destPath)
	if err != nil {
		return nil, err
	}

	result := &enhancedDownloadResult{
		BytesTransferred:  bytesTransferred,
		StreamingChecksum: options.StreamingChecksum,
	}

	if hasher != nil {
		result.Checksum = fmt.Sprintf("%x", hasher.Sum(nil))
	}

	// Report final progress
	if options.EnhancedProgressCallback != nil {
		options.EnhancedProgressCallback(DownloadProgress{
			Total:       resp.ContentLength,
			Transferred: bytesTransferred,
			Done:        true,
		})
	}
	if options.ProgressCallback != nil {
		options.ProgressCallback(resp.ContentLength, bytesTransferred, true)
	}

	return result, nil
}

// resumeDownload resumes a partially downloaded file
func resumeDownload(url, destPath string, existingSize, contentLength int64, options *BinaryDownloadOptions) (*enhancedDownloadResult, error) {
	client := &http.Client{Timeout: 10 * time.Minute}

	req, err := http.NewRequestWithContext(options.Context, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	setGitHubAuthIfAvailable(req)

	// Set Range header for resuming
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		// Server doesn't support resume, start over
		return downloadSingle(url, destPath, options)
	}

	// Open file in append mode
	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	// Setup progress tracking starting from existing size
	progressTracker := &downloadProgressTracker{
		total:            contentLength,
		transferred:      existingSize,
		enhancedCallback: options.EnhancedProgressCallback,
		legacyCallback:   options.ProgressCallback,
		bandwidthLimiter: newBandwidthLimiter(options.MaxBandwidth),
	}

	// Setup streaming checksum if enabled (note: can't validate existing portion)
	var hasher hash.Hash
	var writer io.Writer = out
	if options.StreamingChecksum {
		hasher = sha256.New()
		writer = io.MultiWriter(out, hasher)
	}

	reader := &enhancedProgressReader{
		reader:  resp.Body,
		tracker: progressTracker,
	}

	bytesTransferred, err := io.Copy(writer, reader)
	if err != nil {
		return nil, err
	}

	result := &enhancedDownloadResult{
		BytesTransferred:  existingSize + bytesTransferred,
		Resumed:           true,
		ResumedBytes:      existingSize,
		StreamingChecksum: options.StreamingChecksum,
	}

	if hasher != nil {
		// Note: This checksum only covers the resumed portion
		result.Checksum = fmt.Sprintf("%x", hasher.Sum(nil))
	}

	// Report final progress
	if options.EnhancedProgressCallback != nil {
		options.EnhancedProgressCallback(DownloadProgress{
			Total:        contentLength,
			Transferred:  existingSize + bytesTransferred,
			Done:         true,
			CurrentChunk: 1,
		})
	}
	if options.ProgressCallback != nil {
		options.ProgressCallback(contentLength, existingSize+bytesTransferred, true)
	}

	return result, nil
}

// downloadParallel performs parallel chunk downloads
func downloadParallel(url, destPath string, contentLength int64, options *BinaryDownloadOptions) (*enhancedDownloadResult, error) {
	numChunks := int(contentLength / options.ChunkSize)
	if contentLength%options.ChunkSize != 0 {
		numChunks++
	}

	// Limit maximum number of chunks to prevent too many connections
	maxChunks := 8
	if numChunks > maxChunks {
		numChunks = maxChunks
		options.ChunkSize = contentLength / int64(numChunks)
	}

	// Create temporary files for chunks
	chunkPaths := make([]string, numChunks)
	for i := 0; i < numChunks; i++ {
		chunkPaths[i] = fmt.Sprintf("%s.chunk.%d", destPath, i)
	}

	// Setup progress tracking
	progressTracker := &downloadProgressTracker{
		total:            contentLength,
		enhancedCallback: options.EnhancedProgressCallback,
		legacyCallback:   options.ProgressCallback,
	}

	// Download chunks in parallel
	var wg sync.WaitGroup
	chunkErrors := make([]error, numChunks)

	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(chunkIndex int) {
			defer wg.Done()

			start := int64(chunkIndex) * options.ChunkSize
			end := start + options.ChunkSize - 1
			if chunkIndex == numChunks-1 {
				end = contentLength - 1
			}

			chunkErrors[chunkIndex] = downloadChunk(url, chunkPaths[chunkIndex], start, end, progressTracker, options)
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range chunkErrors {
		if err != nil {
			// Clean up chunk files
			for j := 0; j < numChunks; j++ {
				os.Remove(chunkPaths[j])
			}
			return nil, fmt.Errorf("chunk %d failed: %w", i, err)
		}
	}

	// Combine chunks into final file
	err := combineChunks(chunkPaths, destPath)
	if err != nil {
		// Clean up chunk files
		for _, chunkPath := range chunkPaths {
			os.Remove(chunkPath)
		}
		return nil, err
	}

	// Clean up chunk files
	for _, chunkPath := range chunkPaths {
		os.Remove(chunkPath)
	}

	result := &enhancedDownloadResult{
		BytesTransferred:  contentLength,
		ParallelDownload:  true,
		ChunkCount:        numChunks,
		StreamingChecksum: false, // Parallel downloads don't support streaming checksum
	}

	// Report final progress
	if options.EnhancedProgressCallback != nil {
		options.EnhancedProgressCallback(DownloadProgress{
			Total:        contentLength,
			Transferred:  contentLength,
			Done:         true,
			ActiveChunks: numChunks,
		})
	}
	if options.ProgressCallback != nil {
		options.ProgressCallback(contentLength, contentLength, true)
	}

	return result, nil
}

// downloadChunk downloads a specific byte range
func downloadChunk(url, chunkPath string, start, end int64, tracker *downloadProgressTracker, options *BinaryDownloadOptions) error {
	client := &http.Client{Timeout: 10 * time.Minute}

	req, err := http.NewRequestWithContext(options.Context, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	setGitHubAuthIfAvailable(req)

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server returned non-partial content status: %s", resp.Status)
	}

	err = os.MkdirAll(filepath.Dir(chunkPath), 0755)
	if err != nil {
		return err
	}

	out, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	defer out.Close()

	reader := &enhancedProgressReader{
		reader:  resp.Body,
		tracker: tracker,
	}

	_, err = io.Copy(out, reader)
	return err
}

// combineChunks combines downloaded chunks into the final file
func combineChunks(chunkPaths []string, destPath string) error {
	err := os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		return err
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, chunkPath := range chunkPaths {
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(out, chunkFile)
		chunkFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// downloadBinaryFile downloads a binary file from a URL to a destination path (legacy function for backward compatibility)
func downloadBinaryFile(url, destPath string, options *BinaryDownloadOptions) error {
	// For legacy compatibility, use simple single download without enhanced features
	legacyOptions := &BinaryDownloadOptions{
		Version:                  options.Version,
		Platform:                 options.Platform,
		ProgressCallback:         options.ProgressCallback,         // Keep the legacy callback
		EnhancedProgressCallback: options.EnhancedProgressCallback, // Also support enhanced if provided
		MaxRetries:               options.MaxRetries,
		SkipChecksum:             options.SkipChecksum,
		SkipCache:                options.SkipCache,
		Context:                  options.Context,
		RepoURL:                  options.RepoURL,
		EnableResume:             false, // Legacy mode doesn't support resume
		EnableParallel:           false, // Legacy mode doesn't support parallel
		StreamingChecksum:        false, // Legacy mode doesn't support streaming checksum
		MaxBandwidth:             options.MaxBandwidth,
		ChunkSize:                options.ChunkSize,
	}

	// Use enhanced download in single-threaded mode
	_, err := downloadBinaryFileEnhanced(url, destPath, legacyOptions)
	return err
}

// CheckBinaryAvailability checks if a binary is available for the specified version and platform
func CheckBinaryAvailability(version, platform string) (bool, error) {
	// Try to load configuration to use configured mirrors
	cfg, err := config.LoadEffectiveConfig()
	if err != nil {
		// If config loading fails, fall back to default
		return CheckBinaryAvailabilityWithRepo(DefaultBinaryRepo, version, platform)
	}

	// Check if PVM Binary config is available
	if cfg.PVM != nil && cfg.PVM.Binary != nil && len(cfg.PVM.Binary.BinaryMirrors) > 0 {
		return CheckBinaryAvailabilityWithMirrors(cfg.PVM.Binary.BinaryMirrors, version, platform)
	}

	// Fall back to default if no mirrors configured
	return CheckBinaryAvailabilityWithRepo(DefaultBinaryRepo, version, platform)
}

// CheckBinaryAvailabilityWithMirrors checks binary availability using configured mirrors with failover
func CheckBinaryAvailabilityWithMirrors(mirrors []string, version, platform string) (bool, error) {
	if len(mirrors) == 0 {
		// Fall back to default if no mirrors configured
		return CheckBinaryAvailabilityWithRepo(DefaultBinaryRepo, version, platform)
	}

	// Try each mirror in order
	for _, mirror := range mirrors {
		available, err := CheckBinaryAvailabilityWithRepo(mirror, version, platform)
		if err != nil {
			// If there's an error, try the next mirror
			continue
		}
		if available {
			// Binary found in this mirror
			return true, nil
		}
	}

	// Binary not found in any mirror
	return false, nil
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

	// Create HEAD request to check if binary exists
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err
	}

	// Add GitHub authentication if available
	setGitHubAuthIfAvailable(req)

	// Make the request
	resp, err := client.Do(req)
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

// downloadProgressTracker tracks download progress with enhanced features
type downloadProgressTracker struct {
	total            int64
	transferred      int64
	enhancedCallback EnhancedProgressCallback
	legacyCallback   ProgressCallback
	bandwidthLimiter *bandwidthLimiter
	startTime        time.Time
	lastReport       time.Time
	mutex            sync.Mutex
}

// enhancedProgressReader is an io.Reader that reports enhanced progress
type enhancedProgressReader struct {
	reader  io.Reader
	tracker *downloadProgressTracker
}

// Read reads from the underlying reader and reports enhanced progress
func (r *enhancedProgressReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		r.tracker.addBytes(int64(n))
	}
	return n, err
}

// addBytes adds bytes to the progress tracker
func (t *downloadProgressTracker) addBytes(bytes int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.startTime.IsZero() {
		t.startTime = time.Now()
		t.lastReport = t.startTime
	}

	t.transferred += bytes

	// Apply bandwidth limiting if configured
	if t.bandwidthLimiter != nil {
		t.bandwidthLimiter.throttle(bytes)
	}

	// Report progress at regular intervals
	now := time.Now()
	if (t.enhancedCallback != nil || t.legacyCallback != nil) &&
		(now.Sub(t.lastReport) >= ProgressReportInterval || t.transferred >= t.total) {
		elapsed := now.Sub(t.startTime)
		var speed int64
		var eta time.Duration

		if elapsed > time.Nanosecond {
			elapsedSeconds := elapsed.Seconds()
			if elapsedSeconds > 0.001 { // Require at least 1ms to avoid division issues
				speed = int64(float64(t.transferred) / elapsedSeconds)
				if speed > 0 && t.transferred < t.total {
					remaining := t.total - t.transferred
					etaSeconds := float64(remaining) / float64(speed)
					eta = time.Duration(etaSeconds * float64(time.Second))
				}
			}
		}

		// Call enhanced callback if available
		if t.enhancedCallback != nil {
			progress := DownloadProgress{
				Total:       t.total,
				Transferred: t.transferred,
				Speed:       speed,
				ETA:         eta,
				Done:        t.transferred >= t.total,
			}
			t.enhancedCallback(progress)
		}

		// Call legacy callback if available
		if t.legacyCallback != nil {
			t.legacyCallback(t.total, t.transferred, t.transferred >= t.total)
		}

		t.lastReport = now
	}
}

// bandwidthLimiter implements bandwidth throttling
type bandwidthLimiter struct {
	maxBytesPerSecond int64
	lastTime          time.Time
	allowedBytes      int64
	mutex             sync.Mutex
}

// newBandwidthLimiter creates a new bandwidth limiter
func newBandwidthLimiter(maxBytesPerSecond int64) *bandwidthLimiter {
	if maxBytesPerSecond <= 0 {
		return nil // No limiting
	}
	return &bandwidthLimiter{
		maxBytesPerSecond: maxBytesPerSecond,
		lastTime:          time.Now(),
		allowedBytes:      maxBytesPerSecond, // Start with one second worth
	}
}

// throttle applies bandwidth limiting
func (bl *bandwidthLimiter) throttle(bytes int64) {
	if bl == nil {
		return
	}

	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(bl.lastTime)

	// Add allowed bytes based on elapsed time
	if elapsed > 0 {
		newAllowedBytes := int64(elapsed.Seconds() * float64(bl.maxBytesPerSecond))
		bl.allowedBytes += newAllowedBytes

		// Cap at 2 seconds worth to prevent bursts
		maxAllowed := bl.maxBytesPerSecond * 2
		if bl.allowedBytes > maxAllowed {
			bl.allowedBytes = maxAllowed
		}
	}

	bl.lastTime = now

	// Check if we need to throttle
	if bytes > bl.allowedBytes {
		// Calculate how long to sleep
		excessBytes := bytes - bl.allowedBytes
		sleepTime := time.Duration(float64(excessBytes)/float64(bl.maxBytesPerSecond)*1000) * time.Millisecond

		if sleepTime > 0 {
			bl.mutex.Unlock()
			time.Sleep(sleepTime)
			bl.mutex.Lock()
		}

		bl.allowedBytes = 0
	} else {
		bl.allowedBytes -= bytes
	}
}
