// ABOUTME: Perl build and installation functionality
// ABOUTME: Provides functions to compile and install Perl from source

package perl

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// Build error codes
const (
	ErrExtractionFailed    = "501" // Failed to extract source archive
	ErrBuildDirFailed      = "502" // Failed to create build directory
	ErrConfigureFailed     = "503" // Configure script failed
	ErrCompilationFailed   = "504" // Compilation failed
	ErrTestFailed          = "505" // Tests failed
	ErrInstallFailed       = "506" // Installation failed
	ErrCleanupFailed       = "507" // Cleanup failed
	ErrInvalidBuildOptions = "508" // Invalid build options
	ErrBuildCancelled      = "509" // Build process was cancelled
)

// BuildProgressStage represents a stage in the build process
type BuildProgressStage int

const (
	// Build stages
	StageExtract BuildProgressStage = iota
	StageConfigure
	StageCompile
	StageTest
	StageInstall
	StageCleanup
	StageDone
)

// String returns a string representation of the build stage
func (s BuildProgressStage) String() string {
	switch s {
	case StageExtract:
		return "Extracting"
	case StageConfigure:
		return "Configuring"
	case StageCompile:
		return "Compiling"
	case StageTest:
		return "Testing"
	case StageInstall:
		return "Installing"
	case StageCleanup:
		return "Cleaning up"
	case StageDone:
		return "Done"
	default:
		return "Unknown"
	}
}

// BuildProgressCallback is called to report progress during the build process
type BuildProgressCallback func(stage BuildProgressStage, details string, progress float64)

// BuildOptions contains options for building Perl
type BuildOptions struct {
	// Version to build (used to identify source archive)
	Version string

	// SourceFile is the path to the source archive
	// If empty, the archive will be located or downloaded
	SourceFile string

	// InstallDir is the directory where Perl will be installed
	// If empty, a default location will be used
	InstallDir string

	// BuildJobs is the number of parallel jobs to use during compilation
	// If 0, a default value based on CPU count will be used
	BuildJobs int

	// RunTests indicates whether to run the test suite
	RunTests bool

	// BuildDir is the directory for temporary build files
	// If empty, a default location will be used
	BuildDir string

	// ConfigureOptions contains additional options to pass to the Configure script
	ConfigureOptions []string

	// CleanupBuildDir indicates whether to clean up the build directory after installation
	CleanupBuildDir bool

	// BuildOnly indicates whether to build without installing (creates relocatable build)
	BuildOnly bool

	// ProgressCallback is called to report progress
	ProgressCallback BuildProgressCallback `json:"-"`

	// Context for cancellation
	Context context.Context `json:"-"`

	// TestTimeout is the timeout for running tests
	TestTimeout time.Duration

	// AllowTestFailures allows build to continue even if tests fail
	AllowTestFailures bool

	// ForceRebuild forces a rebuild even if cached
	ForceRebuild bool

	// MaxRetries is the maximum number of retries for build steps
	MaxRetries int

	// Verbose enables verbose output
	Verbose bool
}

// BuildResult contains information about the build
type BuildResult struct {
	// Version that was built
	Version string

	// InstallPath is the path where Perl was installed
	InstallPath string

	// BuildPath is the path where Perl was built
	BuildPath string

	// Duration is the total time taken to build
	Duration time.Duration

	// Stages contains timing information for each stage
	Stages map[BuildProgressStage]time.Duration

	// Timestamp is when the build started
	Timestamp time.Time
}

// BuildPerl builds and installs Perl from source
func BuildPerl(options *BuildOptions) (*BuildResult, error) {
	// Use default options if nil
	if options == nil {
		options = &BuildOptions{}
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Track timing information
	startTime := time.Now()
	stageTimes := make(map[BuildProgressStage]time.Duration)
	stageStartTime := startTime

	// Function to update stage timing and report progress
	updateStage := func(stage BuildProgressStage, details string, progress float64) {
		// Update timing for the previous stage if any
		if stage > StageExtract {
			prevStage := stage - 1
			stageTimes[prevStage] = time.Since(stageStartTime)
		}

		// Reset stage start time
		stageStartTime = time.Now()

		// Report progress if callback is set
		if options.ProgressCallback != nil {
			options.ProgressCallback(stage, details, progress)
		}
	}

	// Initialize the build result
	result := &BuildResult{
		Version: options.Version,
		Stages:  stageTimes,
	}

	// Validate version
	parsedVersion, err := ParseVersion(options.Version)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrInvalidBuildOptions,
			fmt.Sprintf("Invalid version format: %s", options.Version),
			err)
	}
	version := parsedVersion.String()
	result.Version = version

	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		return nil, errors.NewSystemError("002",
			"Failed to create required directories", err)
	}

	// Get or create build directory
	buildDir := options.BuildDir
	if buildDir == "" {
		// Use timestamp to create unique build directory
		timestamp := time.Now().Format("20060102-150405")
		buildDir = filepath.Join(dirs.BuildDir, fmt.Sprintf("perl-%s-%s", version, timestamp))
	}

	// Create build directory
	err = os.MkdirAll(buildDir, 0755)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrBuildDirFailed,
			"Failed to create build directory",
			err).
			WithLocation(buildDir)
	}
	result.BuildPath = buildDir

	// Set installation directory
	installDir := options.InstallDir
	if installDir == "" {
		installDir = filepath.Join(dirs.VersionsDir, version)
	}

	// Create installation directory
	err = os.MkdirAll(installDir, 0755)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrInstallFailed,
			"Failed to create installation directory",
			err).
			WithLocation(installDir)
	}
	result.InstallPath = installDir

	// Find or download source archive
	sourceFile := options.SourceFile
	if sourceFile == "" {
		// Check if the source is already cached
		archiveFilename := ""

		// Determine the correct archive extension based on version
		majorVersion := parsedVersion.Major
		minorVersion := parsedVersion.Minor
		patchVersion := parsedVersion.Patch

		if majorVersion > 5 || (majorVersion == 5 && (minorVersion > 14 || (minorVersion == 14 && patchVersion >= 0))) {
			archiveFilename = fmt.Sprintf("perl-%s.tar.xz", version)
		} else {
			archiveFilename = fmt.Sprintf("perl-%s.tar.gz", version)
		}

		sourceFile = filepath.Join(dirs.SourcesDir, archiveFilename)

		// Check if the source archive exists
		if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
			// Download the source archive
			downloadOptions := &DownloadOptions{
				Version:          version,
				ProgressCallback: nil, // TODO: Adapt download progress to build progress
				Context:          options.Context,
			}

			downloadResult, err := DownloadPerlSource(downloadOptions)
			if err != nil {
				return nil, errors.NewVersionError(
					ErrExtractionFailed,
					"Failed to download source archive",
					err)
			}

			sourceFile = downloadResult.Path
		}
	}

	// Start extraction
	updateStage(StageExtract, "Extracting source archive", 0.0)

	// Extract source archive
	extractedDir, err := extractArchive(sourceFile, buildDir, options.Context)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrExtractionFailed,
			"Failed to extract source archive",
			err)
	}

	// Move to the extracted directory
	srcDir := extractedDir
	if srcDir == "" {
		// Try to find the source directory
		entries, err := os.ReadDir(buildDir)
		if err != nil {
			return nil, errors.NewVersionError(
				ErrExtractionFailed,
				"Failed to read build directory",
				err)
		}

		// Look for the first directory
		for _, entry := range entries {
			if entry.IsDir() {
				srcDir = filepath.Join(buildDir, entry.Name())
				break
			}
		}

		if srcDir == "" {
			return nil, errors.NewVersionError(
				ErrExtractionFailed,
				"Could not find source directory in extracted archive",
				nil)
		}
	}

	// Start configure
	updateStage(StageConfigure, "Running Configure script", 0.0)

	// Construct configure options
	configureOptions := []string{
		"-des",                                 // Default options, no interactive prompts
		"-Dusethreads",                         // Enable threads
		fmt.Sprintf("-Dprefix=%s", installDir), // Installation directory
	}

	// Add relocatable @INC only for build-only mode (binary distribution)
	if options.BuildOnly {
		configureOptions = append(configureOptions, "-Duserelocatableinc")
	}

	// Add user-specified configure options
	configureOptions = append(configureOptions, options.ConfigureOptions...)

	// Add platform-specific configure options
	platformOptions, err := getPlatformConfigureOptions(options.BuildOnly)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrInvalidBuildOptions,
			"Failed to get platform-specific configure options",
			err)
	}
	configureOptions = append(configureOptions, platformOptions...)

	// Run Configure script
	configureErr := runCommandWithProgress(
		srcDir,
		"./Configure",
		configureOptions,
		options.Context,
		func(line string) {
			if options.ProgressCallback != nil {
				options.ProgressCallback(StageConfigure, line, 0.5)
			}
		},
	)

	if configureErr != nil {
		return nil, errors.NewVersionError(
			ErrConfigureFailed,
			"Configure script failed",
			configureErr).
			WithLocation(srcDir)
	}

	// Start compilation
	updateStage(StageCompile, "Compiling Perl", 0.0)

	// Determine number of parallel jobs
	jobs := options.BuildJobs
	if jobs <= 0 {
		jobs = runtime.NumCPU()
	}

	// Run make
	makeErr := runCommandWithProgress(
		srcDir,
		"make",
		[]string{fmt.Sprintf("-j%d", jobs)},
		options.Context,
		func(line string) {
			if options.ProgressCallback != nil {
				// Parse progress from make output (approximate)
				progress := 0.0
				if strings.Contains(line, "[100%]") {
					progress = 1.0
				} else if strings.Contains(line, "[") && strings.Contains(line, "%]") {
					// Extract percentage from line like "[45%]"
					start := strings.Index(line, "[") + 1
					end := strings.Index(line, "%]")
					if start > 0 && end > start {
						percentStr := line[start:end]
						if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
							progress = percent / 100.0
						}
					}
				}
				options.ProgressCallback(StageCompile, line, progress)
			}
		},
	)

	if makeErr != nil {
		return nil, errors.NewVersionError(
			ErrCompilationFailed,
			"Compilation failed",
			makeErr).
			WithLocation(srcDir)
	}

	// Run tests if requested
	if options.RunTests {
		updateStage(StageTest, "Running tests", 0.0)

		testErr := runCommandWithProgress(
			srcDir,
			"make",
			[]string{"test"},
			options.Context,
			func(line string) {
				if options.ProgressCallback != nil {
					// Parse progress from test output (approximate)
					progress := 0.0
					if strings.Contains(line, "All tests successful") {
						progress = 1.0
					} else if strings.Contains(line, "test") && strings.Contains(line, "..") {
						// Very rough estimate based on test count
						progress = 0.5
					}
					options.ProgressCallback(StageTest, line, progress)
				}
			},
		)

		if testErr != nil {
			return nil, errors.NewVersionError(
				ErrTestFailed,
				"Tests failed",
				testErr).
				WithLocation(srcDir)
		}
	}

	// Install only if BuildOnly is false
	if !options.BuildOnly {
		// Start installation
		updateStage(StageInstall, "Installing Perl", 0.0)

		// Run make install
		installErr := runCommandWithProgress(
			srcDir,
			"make",
			[]string{"install"},
			options.Context,
			func(line string) {
				if options.ProgressCallback != nil {
					options.ProgressCallback(StageInstall, line, 0.5)
				}
			},
		)

		if installErr != nil {
			return nil, errors.NewVersionError(
				ErrInstallFailed,
				"Installation failed",
				installErr).
				WithLocation(srcDir)
		}
	}

	// Clean up if requested
	if options.CleanupBuildDir {
		updateStage(StageCleanup, "Cleaning up build directory", 0.0)

		cleanupErr := os.RemoveAll(buildDir)
		if cleanupErr != nil {
			// Non-fatal error, continue but report it
			if options.ProgressCallback != nil {
				options.ProgressCallback(StageCleanup,
					fmt.Sprintf("Warning: Failed to clean up build directory: %v", cleanupErr),
					1.0)
			}
		}
	}

	// Done
	if options.BuildOnly {
		updateStage(StageDone, "Build completed", 1.0)
	} else {
		updateStage(StageDone, "Installation completed", 1.0)
	}

	// Update final timing
	result.Duration = time.Since(startTime)

	// Register the installed Perl version only if not build-only
	if !options.BuildOnly {
		err = RegisterVersionAfterBuild(result, "pvm")
		if err != nil {
			// Log the error but don't fail the build
			if options.ProgressCallback != nil {
				options.ProgressCallback(StageDone,
					fmt.Sprintf("Warning: Failed to register Perl version: %v", err),
					1.0)
			}
		}
	}

	return result, nil
}

// extractArchiveFunc is a variable holding the extract archive function
// This allows tests to replace it with a mock
var extractArchiveFunc = doExtractArchive

// extractArchive is a wrapper function that calls the current implementation
func extractArchive(archivePath, destDir string, ctx context.Context) (string, error) {
	return extractArchiveFunc(archivePath, destDir, ctx)
}

// doExtractArchive is the actual implementation of extracting a source archive
func doExtractArchive(archivePath, destDir string, ctx context.Context) (string, error) {
	// Open the archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	// Determine the archive type based on extension
	baseName := filepath.Base(archivePath)
	var reader io.Reader

	// Extract archive based on its type
	switch {
	case strings.HasSuffix(baseName, ".tar.gz") || strings.HasSuffix(baseName, ".tgz"):
		// Create gzip reader
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return "", err
		}
		defer func() { _ = gzReader.Close() }()
		reader = gzReader
	case strings.HasSuffix(baseName, ".tar.xz"):
		// Create xz reader
		xzReader, err := xz.NewReader(file)
		if err != nil {
			return "", err
		}
		reader = xzReader
	default:
		return "", fmt.Errorf("unsupported archive format: %s", baseName)
	}

	// Create tar reader
	tarReader := tar.NewReader(reader)

	// Extract each file
	extractedDir := ""
	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Continue processing
		}

		// Read next header
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return "", err
		}

		// Get the target path
		target := filepath.Join(destDir, header.Name)

		// Check for directory traversal attacks
		if !strings.HasPrefix(target, destDir) {
			return "", fmt.Errorf("illegal file path: %s", header.Name)
		}

		// Handle different file types
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}

			// Remember the root directory (first directory in the archive)
			if extractedDir == "" && filepath.Dir(header.Name) == "." {
				extractedDir = target
			}

		case tar.TypeReg:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}

			// Create file
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return "", err
			}

			// Copy file content
			if _, err := io.Copy(f, tarReader); err != nil {
				_ = f.Close()
				return "", err
			}
			_ = f.Close()

		case tar.TypeSymlink:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}

			// Create symbolic link
			if err := os.Symlink(header.Linkname, target); err != nil {
				return "", err
			}
		}
	}

	return extractedDir, nil
}

// runCommandWithProgressFunc holds the current implementation of runCommandWithProgress
// It can be replaced with a mock for testing
var runCommandWithProgressFunc = doRunCommandWithProgress

// runCommandWithProgress is a wrapper that calls the current implementation
func runCommandWithProgress(
	dir string,
	command string,
	args []string,
	ctx context.Context,
	progressCb func(line string),
) error {
	return runCommandWithProgressFunc(dir, command, args, ctx, progressCb)
}

// doRunCommandWithProgress is the actual implementation of running a command with progress reporting
func doRunCommandWithProgress(
	dir string,
	command string,
	args []string,
	ctx context.Context,
	progressCb func(line string),
) error {
	// Create command
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Read output in background
	var stdoutBuf, stderrBuf bytes.Buffer

	// Start goroutine to read stdout
	stdoutCh := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutBuf.WriteString(line + "\n")
			if progressCb != nil {
				progressCb(line)
			}
		}
		stdoutCh <- scanner.Err()
	}()

	// Start goroutine to read stderr
	stderrCh := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")
			if progressCb != nil {
				progressCb("ERROR: " + line)
			}
		}
		stderrCh <- scanner.Err()
	}()

	// Wait for command to complete
	err = cmd.Wait()

	// Check for output reading errors
	stdoutErr := <-stdoutCh
	if stdoutErr != nil {
		return fmt.Errorf("error reading stdout: %v", stdoutErr)
	}

	stderrErr := <-stderrCh
	if stderrErr != nil {
		return fmt.Errorf("error reading stderr: %v", stderrErr)
	}

	// Check command exit status
	if err != nil {
		// Add output to error for better diagnostics
		err = fmt.Errorf("%w\nSTDOUT:\n%s\nSTDERR:\n%s",
			err, stdoutBuf.String(), stderrBuf.String())
	}

	return err
}

// getPlatformConfigureOptions returns platform-specific options for the Configure script
func getPlatformConfigureOptions(buildOnly bool) ([]string, error) {
	options := []string{}

	switch runtime.GOOS {
	case "darwin":
		// macOS-specific options
		baseOptions := []string{
			"-Dusedtrace",           // Enable dtrace support
			"-Dusethreads",          // Enable threads
			"-Duselargefiles",       // Enable large file support
			"-Dccflags=-DHAS_TIMES", // Fix for macOS time handling
		}

		// Only add shared library support if not doing relocatable build
		if !buildOnly {
			baseOptions = append(baseOptions, "-Duseshrplib") // Build shared libperl
		}

		options = append(options, baseOptions...)

		// Check if we're on Apple Silicon (arm64)
		if runtime.GOARCH == "arm64" {
			options = append(options,
				"-Dcc=clang", // Use clang compiler
				"-Darchname=darwin-thread-multi-2level-arm64", // Architecture name
			)
		}

	case "linux":
		// Linux-specific options
		baseOptions := []string{
			"-Duselargefiles",    // Enable large file support
			"-Dcccdlflags=-fPIC", // Position-independent code for shared libs
		}

		// Only add shared library support if not doing relocatable build
		if !buildOnly {
			baseOptions = append(baseOptions, "-Duseshrplib") // Build shared libperl
		}

		options = append(options, baseOptions...)

	case "windows":
		// Windows-specific options
		// Note: Building Perl on Windows is complex and may require
		// different approaches depending on the build environment
		// For simplicity, we'll assume using MinGW/MSYS2 here
		baseOptions := []string{
			"-Duseithreads", // Use POSIX threads
			"-Dcc=gcc",      // Use gcc compiler
		}

		// Only add shared library support if not doing relocatable build
		if !buildOnly {
			baseOptions = append(baseOptions, "-Duseshrplib") // Build shared libperl
		}

		options = append(options, baseOptions...)
	}

	return options, nil
}
