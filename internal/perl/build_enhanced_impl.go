// ABOUTME: Implementation of enhanced build methods
// ABOUTME: Provides extraction, configuration, compilation, and installation with improvements

package perl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// prepareBuildEnvironment prepares the build directory and environment
func (bm *BuildManager) prepareBuildEnvironment(ctx *BuildContext) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	// Create unique build directory
	timestamp := time.Now().Format("20060102-150405")
	buildDir := filepath.Join(dirs.BuildDir, fmt.Sprintf("perl-%s-%s", ctx.Options.Version, timestamp))

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return errors.NewVersionError(
			ErrBuildDirFailed,
			"Failed to create build directory",
			err,
		).WithLocation(buildDir)
	}

	ctx.BuildDir = buildDir

	// Set install directory
	installDir := ctx.Options.InstallDir
	if installDir == "" {
		installDir = filepath.Join(dirs.VersionsDir, ctx.Options.Version)
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return errors.NewVersionError(
			ErrInstallFailed,
			"Failed to create installation directory",
			err,
		).WithLocation(installDir)
	}

	ctx.InstallDir = installDir

	// Set environment variables for build
	ctx.setupBuildEnvironment()

	return nil
}

// setupBuildEnvironment sets up environment variables for the build
func (ctx *BuildContext) setupBuildEnvironment() {
	// Ensure clean environment for build
	os.Setenv("PERL5LIB", "")
	os.Setenv("PERL5OPT", "")

	// Set compiler flags for better builds
	if runtime.GOOS == "darwin" {
		// macOS specific flags
		os.Setenv("MACOSX_DEPLOYMENT_TARGET", "10.15")
	}
}

// extractSourceWithProgress extracts the source archive with progress reporting
func (bm *BuildManager) extractSourceWithProgress(ctx *BuildContext, progress *Progress) error {
	progress.Update(StageExtract, "Extracting source archive", 0.0)

	// Use enhanced extraction with progress
	extractor := NewArchiveExtractor(ctx.Logger)

	extractedDir, err := extractor.ExtractWithProgress(
		ctx.SourceFile,
		ctx.BuildDir,
		ctx.Options.Context,
		func(current, total int64) {
			if total > 0 {
				prog := float64(current) / float64(total)
				progress.Update(StageExtract, fmt.Sprintf("Extracted %d/%d files", current, total), prog)
			}
		},
	)

	if err != nil {
		return errors.NewVersionError(
			ErrExtractionFailed,
			"Failed to extract source archive",
			err,
		)
	}

	ctx.SourceDir = extractedDir
	progress.Update(StageExtract, "Extraction complete", 1.0)
	return nil
}

// configureBuild runs the Configure script with enhanced error handling
func (bm *BuildManager) configureBuild(ctx *BuildContext, progress *Progress) error {
	progress.Update(StageConfigure, "Preparing configuration", 0.0)

	// Apply PatchPerl patches if needed (cross-platform compatibility)
	err := bm.applyPatchPerlPatches(ctx, progress)
	if err != nil {
		return err
	}

	// Check for cached configure results
	if cachedConfig, ok := bm.buildCache.GetConfigCache(ctx.Options.Version); ok {
		ctx.ConfigCache = cachedConfig
		bm.logger.Infof("Using cached configure results for version %s", ctx.Options.Version)
	}

	// Build configure options
	configOpts := []string{
		"-des", // Default, non-interactive
		fmt.Sprintf("-Dprefix=%s", ctx.InstallDir),
		"-Dusethreads", // Enable threads
	}

	// Add platform-specific options
	platformOpts, err := bm.getPlatformConfigureOptions(ctx)
	if err != nil {
		return err
	}
	configOpts = append(configOpts, platformOpts...)

	// Add user options
	configOpts = append(configOpts, ctx.Options.ConfigureOptions...)

	// Add optimization options based on system
	if ctx.SystemDeps != nil && len(ctx.SystemDeps.Optional) == 0 {
		// All optional deps available, enable more features
		configOpts = append(configOpts, "-Duse64bitall", "-Duselongdouble")
	}

	progress.Update(StageConfigure, "Running Configure script", 0.1)

	// Run configure with timeout
	configCtx, cancel := context.WithTimeout(ctx.Options.Context, 10*time.Minute)
	defer cancel()

	runner := NewCommandRunner(ctx.Logger)
	output, err := runner.RunWithProgress(
		ctx.SourceDir,
		"./Configure",
		configOpts,
		configCtx,
		func(line string, isError bool) {
			if !isError {
				// Parse configure progress
				prog := bm.parseConfigureProgress(line)
				progress.Update(StageConfigure, line, prog)
			}
		},
	)

	if err != nil {
		err := errors.NewVersionError(
			ErrConfigureFailed,
			"Configure script failed",
			err,
		)
		err.WithDetail(fmt.Sprintf("Output: %s", output))
		err.WithDetail(fmt.Sprintf("Options: %v", configOpts))
		return err
	}

	progress.Update(StageConfigure, "Configuration complete", 1.0)
	return nil
}

// parseConfigureProgress attempts to parse progress from configure output
func (bm *BuildManager) parseConfigureProgress(line string) float64 {
	// Configure typically goes through these stages
	stages := []string{
		"Checking for",
		"Looking for",
		"Trying to",
		"Computing",
		"Determining",
		"Creating",
	}

	for i, stage := range stages {
		if strings.Contains(line, stage) {
			return float64(i+1) / float64(len(stages))
		}
	}

	return 0.5 // Default progress
}

// compileWithRetry compiles with retry logic for transient failures
func (bm *BuildManager) compileWithRetry(ctx *BuildContext, progress *Progress) error {
	progress.Update(StageCompile, "Starting compilation", 0.0)

	jobs := ctx.Options.BuildJobs
	if jobs <= 0 {
		jobs = runtime.NumCPU()
	}

	// Reduce jobs on constrained systems
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		// Apple Silicon can be memory constrained
		if jobs > 4 {
			jobs = 4
		}
	}

	var lastErr error
	for attempt := 1; attempt <= ctx.Options.MaxRetries; attempt++ {
		if attempt > 1 {
			bm.logger.Infof("Retrying compilation (attempt %d): %v", attempt, lastErr)
			// Clean partial build
			bm.cleanPartialBuild(ctx)
		}

		err := bm.runMake(ctx, progress, jobs)
		if err == nil {
			progress.Update(StageCompile, "Compilation successful", 1.0)
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !bm.isRetryableError(err) {
			break
		}

		// Reduce parallelism on retry
		if jobs > 1 {
			jobs /= 2
			bm.logger.Infof("Reducing parallel jobs to %d", jobs)
		}
	}

	return errors.NewVersionError(
		ErrCompilationFailed,
		fmt.Sprintf("Compilation failed after %d attempts", ctx.Options.MaxRetries),
		lastErr,
	)
}

// runMake runs the make command with progress tracking
func (bm *BuildManager) runMake(ctx *BuildContext, progress *Progress, jobs int) error {
	makeArgs := []string{fmt.Sprintf("-j%d", jobs)}

	// Add verbose flag if debugging
	if ctx.Options.Verbose {
		makeArgs = append(makeArgs, "VERBOSE=1")
	}

	runner := NewCommandRunner(ctx.Logger)

	// Track compilation progress
	filesCompiled := 0
	totalFiles := bm.estimateTotalFiles(ctx)

	output, err := runner.RunWithProgress(
		ctx.SourceDir,
		"make",
		makeArgs,
		ctx.Options.Context,
		func(line string, isError bool) {
			if !isError && strings.Contains(line, ".c") {
				filesCompiled++
				prog := float64(filesCompiled) / float64(totalFiles)
				if prog > 1.0 {
					prog = 0.99
				}
				progress.Update(StageCompile, fmt.Sprintf("Compiled %d files", filesCompiled), prog)
			}
		},
	)

	if err != nil {
		// Parse make error for common issues
		issue := bm.parseMakeError(output)
		return fmt.Errorf("%w: %s", err, issue)
	}

	return nil
}

// estimateTotalFiles estimates total files to compile
func (bm *BuildManager) estimateTotalFiles(ctx *BuildContext) int {
	// Rough estimate based on Perl version
	version, _ := ParseVersion(ctx.Options.Version)
	if version.Major >= 5 && version.Minor >= 36 {
		return 200 // Newer versions have more files
	}
	return 150
}

// isRetryableError checks if an error is worth retrying
func (bm *BuildManager) isRetryableError(err error) bool {
	errStr := err.Error()

	// Retryable errors
	retryable := []string{
		"internal compiler error",
		"segmentation fault",
		"cannot allocate memory",
		"resource temporarily unavailable",
	}

	for _, pattern := range retryable {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// parseMakeError attempts to identify common make errors
func (bm *BuildManager) parseMakeError(output string) string {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Look for common error patterns
		if strings.Contains(line, "undefined reference") {
			return "Linking error: missing library or symbol"
		}
		if strings.Contains(line, "No such file or directory") {
			return "Missing header file or dependency"
		}
		if strings.Contains(line, "Permission denied") {
			return "Permission error: check file permissions"
		}
	}

	return "Unknown compilation error"
}

// cleanPartialBuild cleans up partial build artifacts
func (bm *BuildManager) cleanPartialBuild(ctx *BuildContext) {
	// Remove object files but keep configure results
	runner := NewCommandRunner(ctx.Logger)
	runner.Run(ctx.SourceDir, "make", []string{"clean"}, context.Background())
}

// runTestsWithTimeout runs tests with a timeout
func (bm *BuildManager) runTestsWithTimeout(ctx *BuildContext, progress *Progress) error {
	progress.Update(StageTest, "Running test suite", 0.0)

	testCtx, cancel := context.WithTimeout(ctx.Options.Context, ctx.Options.TestTimeout)
	defer cancel()

	runner := NewCommandRunner(ctx.Logger)

	// Track test progress
	testsRun := 0
	testsFailed := 0

	output, err := runner.RunWithProgress(
		ctx.SourceDir,
		"make",
		[]string{"test"},
		testCtx,
		func(line string, isError bool) {
			if strings.Contains(line, "ok") || strings.Contains(line, "not ok") {
				testsRun++
				if strings.Contains(line, "not ok") {
					testsFailed++
				}

				status := fmt.Sprintf("Tests: %d run, %d failed", testsRun, testsFailed)
				// Rough progress estimate
				prog := float64(testsRun) / 2000.0 // Perl typically has ~2000 tests
				if prog > 0.95 {
					prog = 0.95
				}
				progress.Update(StageTest, status, prog)
			}
		},
	)

	if err != nil {
		if err == context.DeadlineExceeded {
			return errors.NewVersionError(
				ErrTestFailed,
				fmt.Sprintf("Tests timed out after %v", ctx.Options.TestTimeout),
				err,
			)
		}

		testErr := errors.NewVersionError(
			ErrTestFailed,
			fmt.Sprintf("Test suite failed: %d tests failed", testsFailed),
			err,
		)
		testErr.WithDetail(fmt.Sprintf("Tests run: %d, Failed: %d", testsRun, testsFailed))
		testErr.WithDetail(fmt.Sprintf("Output: %s", output))
		return testErr
	}

	progress.Update(StageTest, fmt.Sprintf("All %d tests passed", testsRun), 1.0)
	return nil
}

// installWithVerification installs and verifies the installation
func (bm *BuildManager) installWithVerification(ctx *BuildContext, progress *Progress) error {
	progress.Update(StageInstall, "Installing Perl", 0.0)

	runner := NewCommandRunner(ctx.Logger)

	// Track installation progress
	filesInstalled := 0

	output, err := runner.RunWithProgress(
		ctx.SourceDir,
		"make",
		[]string{"install"},
		ctx.Options.Context,
		func(line string, isError bool) {
			if !isError && (strings.Contains(line, "Installing") || strings.Contains(line, "install")) {
				filesInstalled++
				progress.Update(StageInstall, fmt.Sprintf("Installed %d files", filesInstalled), 0.5)
			}
		},
	)

	if err != nil {
		installErr := errors.NewVersionError(
			ErrInstallFailed,
			"Installation failed",
			err,
		)
		installErr.WithDetail(fmt.Sprintf("Output: %s", output))
		installErr.WithLocation(ctx.InstallDir)
		return installErr
	}

	// Make the install relocatable by rewriting baked-in dynamic-linker
	// paths. Linux: patchelf rewrites DT_RUNPATH to $ORIGIN-relative.
	// macOS: install_name_tool rewrites LC_LOAD_DYLIB to @rpath and adds
	// per-file @loader_path LC_RPATH entries. Without this, the built
	// Perl carries the build host's absolute prefix and breaks when
	// installed elsewhere.
	//
	// For release builds (--upload), failure is fatal — the tarball would
	// silently ship broken. For local builds we warn; the baked-in RPATH
	// still resolves on the build host.
	progress.Update(StageInstall, "Making binaries relocatable", 0.85)
	if err := makeRelocatable(ctx.InstallDir); err != nil {
		if ctx.Options.ReleaseBuild {
			return errors.NewVersionError(
				ErrInstallFailed,
				"Failed to make the built Perl relocatable "+
					"(required for release builds)",
				err,
			).WithLocation(ctx.InstallDir)
		}
		ctx.Logger.Warningf("Could not make binaries relocatable: %v", err)
		ctx.Logger.Warningf("The built Perl will run at %s but may break if moved", ctx.InstallDir)
		ctx.Logger.Warningf("To fix: install patchelf and re-run `pvm build-perl`")
	}

	// Verify installation
	progress.Update(StageInstall, "Verifying installation", 0.9)

	if err := bm.verifyInstallation(ctx); err != nil {
		return err
	}

	progress.Update(StageInstall, "Installation verified", 1.0)
	return nil
}

// verifyInstallation verifies the Perl installation is working
func (bm *BuildManager) verifyInstallation(ctx *BuildContext) error {
	perlPath := filepath.Join(ctx.InstallDir, "bin", "perl")

	// Check perl binary exists
	if _, err := os.Stat(perlPath); err != nil {
		return errors.NewVersionError(
			ErrInstallFailed,
			"Perl binary not found after installation",
			err,
		).WithLocation(perlPath)
	}

	// Test perl execution
	runner := NewCommandRunner(ctx.Logger)
	output, err := runner.Run(
		ctx.InstallDir,
		perlPath,
		[]string{"-v"},
		context.Background(),
	)

	if err != nil {
		execErr := errors.NewVersionError(
			ErrInstallFailed,
			"Installed Perl binary fails to execute",
			err,
		)
		execErr.WithDetail(fmt.Sprintf("Output: %s", output))
		execErr.WithLocation(perlPath)
		return execErr
	}

	// Verify version matches
	if !strings.Contains(output, ctx.Options.Version) {
		versionErr := errors.NewVersionError(
			ErrInstallFailed,
			fmt.Sprintf("Installed version mismatch: expected %s", ctx.Options.Version),
			nil,
		)
		versionErr.WithDetail(fmt.Sprintf("Output: %s", output))
		return versionErr
	}

	return nil
}

// getPlatformConfigureOptions returns enhanced platform-specific options
func (bm *BuildManager) getPlatformConfigureOptions(ctx *BuildContext) ([]string, error) {
	options := []string{}

	switch runtime.GOOS {
	case "darwin":
		options = append(options,
			"-Duseshrplib",    // Shared library
			"-Duselargefiles", // Large file support
			"-Duse64bitall",   // 64-bit support
		)

		// Apple Silicon specific
		if runtime.GOARCH == "arm64" {
			options = append(options,
				"-Dcc=clang",
				"-Darchname=darwin-thread-multi-arm64",
			)

			// Use system libraries when available
			if ctx.SystemDeps != nil && len(ctx.SystemDeps.Missing) == 0 {
				options = append(options, "-Dusedtrace") // DTrace support
			}
		}

	case "linux":
		options = append(options,
			"-Duseshrplib",
			"-Duselargefiles",
			"-Duse64bitall",
			"-Dcccdlflags=-fPIC",
		)

		// Enable additional features if deps available
		if ctx.SystemDeps != nil {
			if bm.hasLibrary("libssl-dev") {
				options = append(options, "-Duseopenssl")
			}
			if bm.hasLibrary("libgdbm-dev") {
				options = append(options, "-Dusegdbm")
			}
		}

	case "windows":
		// Windows requires special handling
		if bm.isMinGW() {
			options = append(options,
				"-Dcc=gcc",
				"-Dmake=mingw32-make",
			)
		} else {
			options = append(options,
				"-Dcc=cl",
				"-Dmake=nmake",
			)
		}
	}

	return options, nil
}

// hasLibrary checks if a library is available
func (bm *BuildManager) hasLibrary(lib string) bool {
	// This would check system package manager
	// For now, return false
	return false
}

// isMinGW checks if MinGW is available on Windows
func (bm *BuildManager) isMinGW() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	_, err := exec.LookPath("mingw32-make")
	return err == nil
}

// checkCache checks if a cached build is available
func (bm *BuildManager) checkCache(ctx *BuildContext) (*BuildResult, bool) {
	if ctx.Options.ForceRebuild {
		return nil, false
	}

	return bm.buildCache.GetCachedBuild(ctx.Options.Version, ctx.Options)
}

// updateCache updates the build cache
func (bm *BuildManager) updateCache(ctx *BuildContext, result *BuildResult) error {
	return bm.buildCache.SaveBuild(
		ctx.Options.Version,
		ctx.Options,
		result,
		ctx.ConfigCache,
	)
}

// registerInstallation registers the Perl installation
func (bm *BuildManager) registerInstallation(result *BuildResult) error {
	return RegisterVersionAfterBuild(result, "pvm-enhanced")
}

// applyPatchPerlPatches applies Devel::PatchPerl patches for cross-platform compatibility
func (bm *BuildManager) applyPatchPerlPatches(ctx *BuildContext, progress *Progress) error {
	// Check if PatchPerl is disabled
	if ctx.Options.NoPatchPerl {
		bm.logger.Debugf("PatchPerl disabled for Perl %s", ctx.Options.Version)
		return nil
	}

	// Check if PatchPerl is needed for this version
	if !ShouldUsePatchPerl(ctx.Options.Version) {
		bm.logger.Debugf("PatchPerl not needed for Perl %s", ctx.Options.Version)
		return nil
	}

	progress.Update(StageConfigure, "Applying compatibility patches", 0.02)

	bm.logger.Infof("Applying PatchPerl compatibility patches for Perl %s", ctx.Options.Version)

	// Apply patches using PatchPerl
	err := ApplyPatchPerlSafely(ctx.SourceDir, ctx.Options.Version, ctx.Options.Verbose)
	if err != nil {
		return errors.NewVersionError(
			ErrConfigureFailed,
			"Failed to apply PatchPerl compatibility patches",
			err,
		)
	}

	progress.Update(StageConfigure, "Compatibility patches applied", 0.05)
	return nil
}
