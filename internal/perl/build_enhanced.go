// ABOUTME: Enhanced Perl build system with improved error handling and features
// ABOUTME: Provides robust compilation, dependency management, and cross-platform support

package perl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// BuildManager manages the enhanced build process
type BuildManager struct {
	logger       *log.Logger
	checksumsDB  *ChecksumDatabase
	depChecker   *DependencyChecker
	buildCache   *BuildCache
	progressPool *ProgressPool
}

// NewBuildManager creates a new enhanced build manager
func NewBuildManager(logger *log.Logger) (*BuildManager, error) {
	checksums, err := NewChecksumDatabase()
	if err != nil {
		return nil, errors.NewSystemError(
			ErrBuildDirFailed,
			"Failed to initialize checksum database",
			err,
		)
	}

	depChecker, err := NewDependencyChecker()
	if err != nil {
		return nil, errors.NewSystemError(
			ErrBuildDirFailed,
			"Failed to initialize dependency checker",
			err,
		)
	}

	cache, err := NewBuildCache()
	if err != nil {
		return nil, errors.NewSystemError(
			ErrBuildDirFailed,
			"Failed to initialize build cache",
			err,
		)
	}

	return &BuildManager{
		logger:       logger,
		checksumsDB:  checksums,
		depChecker:   depChecker,
		buildCache:   cache,
		progressPool: NewProgressPool(),
	}, nil
}

// BuildEnhanced performs an enhanced build with all improvements
func (bm *BuildManager) BuildEnhanced(options *BuildOptions) (*BuildResult, error) {
	// Validate inputs
	if err := bm.validateBuildOptions(options); err != nil {
		return nil, err
	}

	// Create build context with proper cleanup
	buildCtx := &BuildContext{
		Options:   options,
		StartTime: time.Now(),
		Logger:    bm.logger,
	}
	defer buildCtx.Cleanup()

	// Check system dependencies
	if err := bm.checkSystemDependencies(buildCtx); err != nil {
		return nil, err
	}

	// Initialize progress tracking
	progress := bm.progressPool.Get(options.ProgressCallback)
	defer bm.progressPool.Put(progress)

	// Check build cache
	if cachedResult, ok := bm.checkCache(buildCtx); ok && !options.ForceRebuild {
		progress.Update(StageDone, "Using cached build", 1.0)
		return cachedResult, nil
	}

	// Download and validate source
	sourceFile, err := bm.downloadAndValidateSource(buildCtx, progress)
	if err != nil {
		return nil, err
	}
	buildCtx.SourceFile = sourceFile

	// Prepare build environment
	if err := bm.prepareBuildEnvironment(buildCtx); err != nil {
		return nil, err
	}

	// Extract source with progress
	if err := bm.extractSourceWithProgress(buildCtx, progress); err != nil {
		return nil, err
	}

	// Configure build
	if err := bm.configureBuild(buildCtx, progress); err != nil {
		return nil, err
	}

	// Compile with retry logic
	if err := bm.compileWithRetry(buildCtx, progress); err != nil {
		return nil, err
	}

	// Run tests if requested
	if options.RunTests {
		if err := bm.runTestsWithTimeout(buildCtx, progress); err != nil {
			if !options.AllowTestFailures {
				return nil, err
			}
			bm.logger.Warningf("Tests failed but continuing due to AllowTestFailures flag: %v", err)
		}
	}

	// Install with verification
	if err := bm.installWithVerification(buildCtx, progress); err != nil {
		return nil, err
	}

	// Update build cache
	result := buildCtx.ToBuildResult()
	if err := bm.updateCache(buildCtx, result); err != nil {
		bm.logger.Warningf("Failed to update build cache: %v", err)
	}

	// Register installation
	if err := bm.registerInstallation(result); err != nil {
		bm.logger.Warningf("Failed to register installation: %v", err)
	}

	progress.Update(StageDone, "Build completed successfully", 1.0)
	return result, nil
}

// validateBuildOptions validates the build options
func (bm *BuildManager) validateBuildOptions(options *BuildOptions) error {
	if options == nil {
		return errors.NewSystemError(
			ErrInvalidBuildOptions,
			"Build options cannot be nil",
			nil,
		)
	}

	if options.Version == "" {
		return errors.NewVersionError(
			ErrInvalidBuildOptions,
			"Version must be specified",
			nil,
		)
	}

	// Validate version format
	if _, err := ParseVersion(options.Version); err != nil {
		return errors.NewVersionError(
			ErrInvalidBuildOptions,
			"Invalid version format",
			err,
		)
	}

	// Set defaults
	if options.Context == nil {
		options.Context = context.Background()
	}

	if options.BuildJobs <= 0 {
		options.BuildJobs = runtime.NumCPU()
	}

	if options.TestTimeout <= 0 {
		options.TestTimeout = 30 * time.Minute
	}

	if options.MaxRetries <= 0 {
		options.MaxRetries = 3
	}

	return nil
}

// checkSystemDependencies checks for required system dependencies
func (bm *BuildManager) checkSystemDependencies(ctx *BuildContext) error {
	deps, err := bm.depChecker.CheckBuildDependencies()
	if err != nil {
		return errors.NewSystemError(
			ErrBuildDirFailed,
			"Failed to check system dependencies",
			err,
		)
	}

	if len(deps.Missing) > 0 {
		err := errors.NewSystemError(
			ErrBuildDirFailed,
			fmt.Sprintf("Missing required dependencies: %s", strings.Join(deps.Missing, ", ")),
			nil,
		)
		err.WithDetail(fmt.Sprintf("Missing: %s", strings.Join(deps.Missing, ", ")))
		if len(deps.Optional) > 0 {
			err.WithDetail(fmt.Sprintf("Optional: %s", strings.Join(deps.Optional, ", ")))
		}
		// Add platform-specific install hints
		if hint, ok := deps.InstallHint[runtime.GOOS]; ok {
			err.WithHint(hint)
		}
		return err
	}

	// Warn about optional dependencies
	if len(deps.Optional) > 0 {
		bm.logger.Warningf("Optional dependencies missing: %s. Some features may not be available",
			strings.Join(deps.Optional, ", "))
	}

	ctx.SystemDeps = deps
	return nil
}

// downloadAndValidateSource downloads and validates the source archive
func (bm *BuildManager) downloadAndValidateSource(ctx *BuildContext, progress *Progress) (string, error) {
	progress.Update(StageDownload, "Checking source archive", 0.0)

	// Determine source file path
	sourceFile := ctx.Options.SourceFile
	if sourceFile == "" {
		// Download with checksum validation
		downloadOpts := &DownloadOptions{
			Version:          ctx.Options.Version,
			Context:          ctx.Options.Context,
			ProgressCallback: progress.DownloadCallback(),
			SkipChecksum:     false,
			MaxRetries:       ctx.Options.MaxRetries,
		}

		result, err := bm.downloadWithValidation(downloadOpts)
		if err != nil {
			return "", err
		}

		sourceFile = result.Path
	}

	// Validate checksum
	if err := bm.validateChecksum(sourceFile, ctx.Options.Version); err != nil {
		return "", err
	}

	progress.Update(StageDownload, "Source validated", 1.0)
	return sourceFile, nil
}

// downloadWithValidation downloads and validates against known checksums
func (bm *BuildManager) downloadWithValidation(opts *DownloadOptions) (*DownloadResult, error) {
	// First attempt normal download
	result, err := DownloadPerlSource(opts)
	if err != nil {
		return nil, err
	}

	// Validate checksum
	expectedChecksum, err := bm.checksumsDB.GetChecksum(opts.Version)
	if err != nil {
		bm.logger.Warningf("No known checksum for version %s", opts.Version)
		return result, nil
	}

	actualChecksum, err := calculateFileChecksum(result.Path)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrChecksumMismatch,
			"Failed to calculate checksum",
			err,
		)
	}

	if actualChecksum != expectedChecksum {
		// Delete corrupted file
		os.Remove(result.Path)
		return nil, errors.NewVersionError(
			ErrChecksumMismatch,
			fmt.Sprintf("Checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum),
			nil,
		)
	}

	result.Checksum = actualChecksum
	return result, nil
}

// validateChecksum validates file checksum against known good values
func (bm *BuildManager) validateChecksum(filePath, version string) error {
	expectedChecksum, err := bm.checksumsDB.GetChecksum(version)
	if err != nil {
		// No known checksum, skip validation
		bm.logger.Debugf("No known checksum for version %s", version)
		return nil
	}

	actualChecksum, err := calculateFileChecksum(filePath)
	if err != nil {
		return errors.NewVersionError(
			ErrChecksumMismatch,
			"Failed to calculate file checksum",
			err,
		)
	}

	if actualChecksum != expectedChecksum {
		err := errors.NewVersionError(
			ErrChecksumMismatch,
			fmt.Sprintf("Checksum validation failed for %s", filepath.Base(filePath)),
			nil,
		)
		err.WithDetail(fmt.Sprintf("Expected: %s, Actual: %s", expectedChecksum, actualChecksum))
		err.WithLocation(filePath)
		return err
	}

	return nil
}

// BuildContext holds the context for a build operation
type BuildContext struct {
	Options        *BuildOptions
	StartTime      time.Time
	SourceFile     string
	BuildDir       string
	SourceDir      string
	InstallDir     string
	SystemDeps     *DependencyInfo
	ConfigCache    map[string]string
	BuildArtifacts []string
	Logger         *log.Logger
}

// Cleanup cleans up the build context
func (ctx *BuildContext) Cleanup() {
	if ctx.Options.CleanupBuildDir && ctx.BuildDir != "" {
		if err := os.RemoveAll(ctx.BuildDir); err != nil {
			ctx.Logger.Warningf("Failed to cleanup build directory %s: %v", ctx.BuildDir, err)
		}
	}
}

// ToBuildResult converts context to build result
func (ctx *BuildContext) ToBuildResult() *BuildResult {
	return &BuildResult{
		Version:     ctx.Options.Version,
		InstallPath: ctx.InstallDir,
		BuildPath:   ctx.BuildDir,
		Duration:    time.Since(ctx.StartTime),
		Timestamp:   ctx.StartTime,
	}
}

// Progress represents build progress tracking
type Progress struct {
	callback BuildProgressCallback
	stage    BuildProgressStage
	mu       sync.Mutex
}

// Update updates the progress
func (p *Progress) Update(stage BuildProgressStage, details string, progress float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stage = stage
	if p.callback != nil {
		p.callback(stage, details, progress)
	}
}

// DownloadCallback returns a callback for download progress
func (p *Progress) DownloadCallback() ProgressCallback {
	return func(total, transferred int64, done bool) {
		if total > 0 {
			progress := float64(transferred) / float64(total)
			status := fmt.Sprintf("Downloaded %d/%d bytes", transferred, total)
			if done {
				status = "Download complete"
				progress = 1.0
			}
			p.Update(StageDownload, status, progress)
		}
	}
}

// ProgressPool manages progress trackers
type ProgressPool struct {
	pool sync.Pool
}

// NewProgressPool creates a new progress pool
func NewProgressPool() *ProgressPool {
	return &ProgressPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Progress{}
			},
		},
	}
}

// Get gets a progress tracker from the pool
func (pp *ProgressPool) Get(callback BuildProgressCallback) *Progress {
	p := pp.pool.Get().(*Progress)
	p.callback = callback
	p.stage = StageExtract
	return p
}

// Put returns a progress tracker to the pool
func (pp *ProgressPool) Put(p *Progress) {
	p.callback = nil
	pp.pool.Put(p)
}

// Additional build stages
const (
	StageDownload BuildProgressStage = iota - 1 // Before extract
)
