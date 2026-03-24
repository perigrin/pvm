// ABOUTME: Module build and installation functionality
// ABOUTME: Functions for building and installing CPAN modules

package modules

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/perl"
)

// Error codes for build operations
const (
	ErrBuildFailed      = "PVI-4201" // Failed to build module
	ErrTestFailed       = "PVI-4202" // Module tests failed
	ErrInstallFailed    = "PVI-4203" // Failed to install module
	ErrDependencyFailed = "PVI-4204" // Failed to resolve dependencies
	ErrInvalidBuildOpts = "PVI-4205" // Invalid build options
	ErrBuildCancelled   = "PVI-4206" // Build process cancelled
	ErrBadBuildSystem   = "PVI-4207" // Unsupported build system
	ErrPerlNotFound     = "PVI-4208" // Perl interpreter not found
	ErrCleanupFailed    = "PVI-4209" // Failed to clean up build files
	ErrPrereqFailed     = "PVI-4210" // Failed to install prerequisites
)

// BuildProgressStage represents a stage in the module build process
type BuildProgressStage int

const (
	// Module build stages
	StagePrepare BuildProgressStage = iota
	StageCreateBuildScript
	StageBuild
	StageTest
	StageInstall
	StageCleanup
	StageDone
)

// String returns a string representation of the build stage
func (s BuildProgressStage) String() string {
	switch s {
	case StagePrepare:
		return "Preparing"
	case StageCreateBuildScript:
		return "Creating build script"
	case StageBuild:
		return "Building"
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

// ModuleBuildOptions contains options for building a module
type ModuleBuildOptions struct {
	// Path to the extracted module directory
	ModuleDir string

	// Module name
	ModuleName string

	// Distribution name
	Distribution string

	// Path to the Perl interpreter to use
	PerlPath string

	// Installation directory (usually site_perl)
	InstallDir string

	// BuildDir is the directory for temporary build files
	BuildDir string

	// Run tests before installation
	RunTests bool

	// Skip tests completely
	NoTest bool

	// Force installation even if tests fail
	Force bool

	// Clean build directory after installation
	Cleanup bool

	// Include verbose output
	Verbose bool

	// Additional arguments to pass to build commands
	BuildArgs []string

	// Additional arguments to pass to test commands
	TestArgs []string

	// Additional arguments to pass to install commands
	InstallArgs []string

	// Skip prerequisite installation
	SkipPrereqs bool

	// Environment variables for the build process (e.g., local::lib setup)
	Environment map[string]string

	// Progress callback
	ProgressCallback BuildProgressCallback

	// Context for cancellation
	Context context.Context
}

// ModuleBuildResult contains information about the build
type ModuleBuildResult struct {
	// Module name
	ModuleName string

	// Distribution name
	Distribution string

	// Whether the module was successfully built
	Success bool

	// Whether the module was successfully installed
	Installed bool

	// Whether tests were run and passed
	TestsPassed bool

	// Detailed test results if tests were run
	TestResults *errors.TestResults

	// Warning messages from the build process
	Warnings []string

	// Error messages from the build process
	Errors []string

	// Output from the build process
	Output string

	// Duration is the total time taken to build
	Duration time.Duration

	// Stages contains timing information for each stage
	Stages map[BuildProgressStage]time.Duration
}

// BuildAndInstallModuleFunc is the function type for building and installing a Perl module
type BuildAndInstallModuleFunc func(options *ModuleBuildOptions) (*ModuleBuildResult, error)

// BuildAndInstallModule is a variable that holds the module build and install function
// It can be replaced in tests
var BuildAndInstallModule BuildAndInstallModuleFunc = buildAndInstallModule

// buildAndInstallModule is the actual implementation of the module build and install function
func buildAndInstallModule(options *ModuleBuildOptions) (*ModuleBuildResult, error) {
	// Use default options if nil
	if options == nil {
		return nil, errors.NewSystemError(
			ErrInvalidBuildOpts,
			"No build options provided",
			nil)
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
		if stage > StagePrepare {
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
	result := &ModuleBuildResult{
		ModuleName:   options.ModuleName,
		Distribution: options.Distribution,
		Success:      false,
		Installed:    false,
		TestsPassed:  false,
		Warnings:     []string{},
		Errors:       []string{},
		Output:       "",
		Stages:       stageTimes,
	}

	// Buffer for collecting output
	var outputBuffer bytes.Buffer

	// Start with preparation stage
	updateStage(StagePrepare, "Checking build environment", 0.0)

	// Validate the module directory
	if options.ModuleDir == "" {
		return result, errors.NewSystemError(
			ErrInvalidBuildOpts,
			"No module directory specified",
			nil)
	}

	// Check if the module directory exists
	if _, err := os.Stat(options.ModuleDir); os.IsNotExist(err) {
		return result, errors.NewSystemError(
			ErrInvalidBuildOpts,
			"Module directory does not exist",
			err).
			WithLocation(options.ModuleDir)
	}

	// Ensure Perl path is specified
	if options.PerlPath == "" {
		// Try to find the current Perl version
		perlPath, err := perl.GetCurrentPerlPath()
		if err != nil {
			return result, errors.NewSystemError(
				ErrPerlNotFound,
				"No Perl interpreter specified and could not find current Perl",
				err)
		}
		options.PerlPath = perlPath
	}

	// Check if Perl exists
	if _, err := os.Stat(options.PerlPath); os.IsNotExist(err) {
		return result, errors.NewSystemError(
			ErrPerlNotFound,
			"Perl interpreter not found",
			err).
			WithLocation(options.PerlPath)
	}

	// Detect the build system
	buildSystem, err := DetectBuildSystem(options.ModuleDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Could not detect build system: %v", err))
		return result, errors.NewSystemError(
			ErrBadBuildSystem,
			"Could not detect build system",
			err).
			WithLocation(options.ModuleDir)
	}

	// Log detected build system
	log.Debugf("Detected build system: %s", buildSystem)
	updateStage(StagePrepare, fmt.Sprintf("Detected build system: %s", buildSystem), 0.5)

	// Detect the correct make command for this Perl installation
	makeCmd, err := detectMakeCommand(options.PerlPath)
	if err != nil {
		if runtime.GOOS == "windows" {
			makeCmd = "gmake"
		}
		log.Warnf("Could not detect make command, falling back to %q: %v", makeCmd, err)
	}

	// Create build script based on the detected build system
	updateStage(StageCreateBuildScript, "Creating build script", 0.0)

	var installCmd []string

	// Install prerequisites if not skipped
	if !options.SkipPrereqs {
		// Create the install command based on build system
		switch buildSystem {
		case "Build.PL":
			// For Module::Build (Build.PL)
			// Run the build script generator
			log.Debugf("Running Build.PL")
			updateStage(StageCreateBuildScript, "Running Build.PL", 0.2)

			cmd := []string{options.PerlPath, "Build.PL"}
			if options.Verbose {
				cmd = append(cmd, "--verbose")
			}

			output, err := runCommandWithEnv(options.ModuleDir, cmd, options.Context, options.Environment)
			outputBuffer.WriteString(output)

			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Build.PL failed: %v", err))
				return result, errors.NewSystemError(
					ErrBuildFailed,
					"Failed to create build script",
					err).
					WithLocation(options.ModuleDir)
			}

			// Now the ./Build script should exist
			installCmd = buildPLCommand(options.PerlPath, "install")
			if options.Force {
				installCmd = append(installCmd, "--force")
			}
			if options.Verbose {
				installCmd = append(installCmd, "--verbose")
			}

		case "Makefile.PL":
			// For ExtUtils::MakeMaker (Makefile.PL)
			log.Debugf("Running Makefile.PL")
			updateStage(StageCreateBuildScript, "Running Makefile.PL", 0.2)

			cmd := []string{options.PerlPath, "Makefile.PL"}

			// Note: Installation directory is now handled via PERL_MM_OPT environment variable
			// set up in the isolation environment (local::lib style)

			if options.Verbose {
				cmd = append(cmd, "VERBOSE=1")
			}

			output, err := runCommandWithEnv(options.ModuleDir, cmd, options.Context, options.Environment)
			outputBuffer.WriteString(output)

			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Makefile.PL failed: %v", err))
				return result, errors.NewSystemError(
					ErrBuildFailed,
					"Failed to create Makefile",
					err).
					WithLocation(options.ModuleDir)
			}

			// Now make should work
			installCmd = []string{makeCmd, "install"}
			if options.Force {
				installCmd = append(installCmd, "FORCE=1")
			}

		case "Makefile":
			// Pre-existing Makefile
			log.Debugf("Using existing Makefile")
			updateStage(StageCreateBuildScript, "Using existing Makefile", 0.5)

			// Use make directly
			installCmd = []string{makeCmd, "install"}
			if options.Force {
				installCmd = append(installCmd, "FORCE=1")
			}

		default:
			result.Errors = append(result.Errors, fmt.Sprintf("Unsupported build system: %s", buildSystem))
			return result, errors.NewSystemError(
				ErrBadBuildSystem,
				fmt.Sprintf("Unsupported build system: %s", buildSystem),
				nil).
				WithLocation(options.ModuleDir)
		}
	}

	// Build the module
	updateStage(StageBuild, "Building module", 0.0)

	switch buildSystem {
	case "Build.PL":
		// Build with Module::Build
		cmd := buildPLCommand(options.PerlPath)
		if options.Verbose {
			cmd = append(cmd, "--verbose")
		}

		output, err := runCommandWithEnv(options.ModuleDir, cmd, options.Context, options.Environment)
		outputBuffer.WriteString(output)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Build failed: %v", err))
			return result, errors.NewSystemError(
				ErrBuildFailed,
				"Failed to build module",
				err).
				WithLocation(options.ModuleDir)
		}

	case "Makefile.PL", "Makefile":
		// Build with make
		cmd := []string{makeCmd}

		// Determine number of parallel jobs (for make -j)
		jobs := runtime.NumCPU()
		cmd = append(cmd, fmt.Sprintf("-j%d", jobs))

		output, err := runCommandWithEnv(options.ModuleDir, cmd, options.Context, options.Environment)
		outputBuffer.WriteString(output)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Make failed: %v", err))
			return result, errors.NewSystemError(
				ErrBuildFailed,
				"Failed to build module",
				err).
				WithLocation(options.ModuleDir)
		}
	}

	// Run tests if requested
	result.TestsPassed = !options.RunTests // If tests aren't run, consider them "passed"

	if options.RunTests && !options.NoTest {
		updateStage(StageTest, "Running tests", 0.0)

		var testCmd []string
		switch buildSystem {
		case "Build.PL":
			testCmd = buildPLCommand(options.PerlPath, "test")
			if options.Verbose {
				testCmd = append(testCmd, "--verbose")
			}
		case "Makefile.PL", "Makefile":
			testCmd = []string{makeCmd, "test"}
		}

		// Add any user-specified test arguments
		testCmd = append(testCmd, options.TestArgs...)

		output, err := runCommandWithEnv(options.ModuleDir, testCmd, options.Context, options.Environment)
		outputBuffer.WriteString(output)

		// Parse test output using the new parser
		parser := NewTAPParser(options.Verbose)
		testResults := parser.ParseTestOutput(output)
		result.TestResults = testResults

		if err != nil {
			// Create user-friendly error using new error type
			testError := errors.NewModuleTestFailureError(options.ModuleName, "", testResults)

			result.Warnings = append(result.Warnings, testError.Error())
			log.Warnf("Tests failed for %s: %v", options.ModuleName, err)

			if !options.Force {
				return result, testError
			}

			// If Force is true, continue despite test failures
			log.Warnf("Installing %s despite test failures (--force)", options.ModuleName)
		} else {
			result.TestsPassed = true
		}
	} else if options.NoTest {
		log.Infof("Skipping tests for %s (--notest flag)", options.ModuleName)
		result.TestsPassed = true // Consider tests "passed" when skipped
	}

	// Install the module
	updateStage(StageInstall, "Installing module", 0.0)

	// Add any user-specified install arguments
	installCmd = append(installCmd, options.InstallArgs...)

	output, err := runCommandWithEnv(options.ModuleDir, installCmd, options.Context, options.Environment)
	outputBuffer.WriteString(output)

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Installation failed: %v", err))
		return result, errors.NewSystemError(
			ErrInstallFailed,
			"Failed to install module",
			err).
			WithLocation(options.ModuleDir)
	}

	// Module is now installed
	result.Installed = true

	// Clean up if requested
	if options.Cleanup && options.BuildDir != "" {
		updateStage(StageCleanup, "Cleaning up build files", 0.0)

		// Only remove the build directory if it's not the module directory
		// (in case the module was built directly in its final location)
		if options.BuildDir != options.ModuleDir {
			err := os.RemoveAll(options.BuildDir)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to clean up build directory: %v", err))
				log.Warnf("Failed to clean up build directory %s: %v", options.BuildDir, err)
			}
		}
	}

	// All done
	updateStage(StageDone, "Module installed successfully", 1.0)

	// Update final timing
	result.Duration = time.Since(startTime)
	result.Success = true
	result.Output = outputBuffer.String()

	return result, nil
}

// detectMakeCommand queries the active Perl for its configured make command.
// Perl's Config.pm knows which make to use for the current installation
// (e.g., "make" on Unix, "gmake" on Strawberry Perl, "nmake" on MSVC Perl).
func detectMakeCommand(perlPath string) (string, error) {
	out, err := exec.Command(perlPath, "-MConfig", "-e", `print $Config{make}`).Output()
	if err != nil {
		return "make", err
	}
	cmd := strings.TrimSpace(string(out))
	if cmd == "" {
		return "make", nil
	}
	return cmd, nil
}

// buildPLCommand returns the command to run a Module::Build script.
// On Unix, ./Build is directly executable via shebang.
// On Windows, it must be invoked through the Perl interpreter.
func buildPLCommand(perlPath string, args ...string) []string {
	if runtime.GOOS == "windows" {
		return append([]string{perlPath, "Build"}, args...)
	}
	return append([]string{"./Build"}, args...)
}

// runCommandWithEnv runs a command in the specified directory with custom environment variables
func runCommandWithEnv(dir string, command []string, ctx context.Context, envVars map[string]string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Create command
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir

	// Create pipes for stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set up environment variables
	env := os.Environ()
	env = append(env, "PERL_MM_USE_DEFAULT=1") // Non-interactive installation

	// Add custom environment variables
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	cmd.Env = env

	// Log the command being run
	log.Debugf("Running command in %s: %s", dir, strings.Join(command, " "))
	if len(envVars) > 0 {
		log.Debugf("With environment variables: %v", envVars)
	}

	// Execute the command
	err := cmd.Run()

	// Combine stdout and stderr
	output := stdout.String()
	if stderr.Len() > 0 {
		output += stderr.String()
	}

	return output, err
}
