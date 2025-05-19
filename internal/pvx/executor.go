// ABOUTME: PVX script execution functionality
// ABOUTME: Runs Perl scripts with specific versions and environment control

package pvx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/perl"
)

// PVX execution error codes
const (
	ErrExecutionFailed  = "401" // Script execution failed
	ErrScriptNotFound   = "402" // Script file not found
	ErrVersionNotFound  = "403" // Specified Perl version not found
	ErrInvalidIsolation = "404" // Invalid isolation level specified
)

// Variable for execCommand to allow mocking in tests
var execCommand = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

// IsolationLevel defines the level of isolation for script execution
type IsolationLevel string

// Available isolation levels
const (
	// IsolationNone runs the script with no isolation, using the system's Perl environment
	IsolationNone IsolationLevel = "none"

	// IsolationLow creates a minimal isolation layer with local::lib equivalent
	// - Uses existing Perl environment, but allows installing modules locally
	// - Inherits all environment variables
	// - Has full access to the filesystem
	IsolationLow IsolationLevel = "low"

	// IsolationMedium provides stronger isolation by restricting module access
	// - Creates a clean PERL5LIB environment
	// - Still inherits most environment variables
	// - Has full access to the filesystem but restricted module installation
	IsolationMedium IsolationLevel = "medium"

	// IsolationHigh creates the strongest isolation possible without containers
	// - Creates a clean environment with minimal environment variables
	// - Restricts module access to only the isolation directory
	// - Still has filesystem access, but proper isolation is recommended
	IsolationHigh IsolationLevel = "high"
)

// All valid isolation levels
var ValidIsolationLevels = map[IsolationLevel]bool{
	IsolationNone:   true,
	IsolationLow:    true,
	IsolationMedium: true,
	IsolationHigh:   true,
}

// ExecutionOptions contains options for script execution
type ExecutionOptions struct {
	// Path to the script to execute
	ScriptPath string

	// Inline Perl code to execute (used with -e flag)
	InlineCode string

	// Command line arguments to pass to the script
	Args []string

	// Environment variables to set for the script execution
	Env map[string]string

	// Specific Perl version to use (if empty, will be resolved)
	PerlVersion string

	// Force the use of the specified version, even if it doesn't exist
	ForceVersion bool

	// Root directory for execution environment
	RootDir string

	// Whether to enable type checking before execution
	TypeCheck bool

	// Whether to enable verbose output
	Verbose bool

	// Whether to create an isolated environment for the script (deprecated, use IsolationLevel instead)
	Isolated bool

	// The level of isolation to apply when executing the script
	IsolationLevel IsolationLevel

	// Path to the isolation directory (if empty, will be created in a temp directory)
	IsolationDir string

	// Whether to skip cleanup of the isolation directory after execution
	NoCleanup bool
}

// ExecuteResult contains the result of script execution
type ExecuteResult struct {
	// The combined output (stdout and stderr)
	Output string

	// The exit code of the script
	ExitCode int
}

// ExecuteScript runs a Perl script with the specified options
func ExecuteScript(options *ExecutionOptions) (string, error) {
	if options == nil {
		return "", errors.NewExecutionError(
			ErrExecutionFailed,
			"No execution options provided",
			nil)
	}

	// Validate script path
	if _, err := os.Stat(options.ScriptPath); os.IsNotExist(err) {
		return "", errors.NewExecutionError(
			ErrScriptNotFound,
			fmt.Sprintf("Script file not found: %s", options.ScriptPath),
			err)
	}

	// Resolve Perl version to use
	perlExe, err := resolvePerlExecutable(options)
	if err != nil {
		return "", err
	}

	// Create the command to execute the script
	cmd := exec.Command(perlExe, buildArguments(options)...)

	// Set environment variables
	env, err := buildEnvironment(options)
	if err != nil {
		return "", err
	}
	cmd.Env = env

	// Configure output capture
	var outputBuffer bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer

	// Execute the command
	log.Debugf("Executing Perl script with command: %s %s", perlExe, strings.Join(buildArguments(options), " "))
	// Using a function variable to allow for mocking in tests
	err = execCommand(cmd)

	// Check for execution errors
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Get the exit code if possible
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
			// Cleanup isolation directory if needed
			cleanupIsolationDir(options)
			return outputBuffer.String(), errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Script execution failed with exit code %d", exitCode),
				err)
		}
		// Cleanup isolation directory if needed
		cleanupIsolationDir(options)
		return outputBuffer.String(), errors.NewExecutionError(
			ErrExecutionFailed,
			"Script execution failed",
			err)
	}

	// Cleanup isolation directory if needed
	cleanupIsolationDir(options)

	return outputBuffer.String(), nil
}

// ExecuteInlineCode runs Perl code directly with the specified options
func ExecuteInlineCode(options *ExecutionOptions) (string, error) {
	if options == nil {
		return "", errors.NewExecutionError(
			ErrExecutionFailed,
			"No execution options provided",
			nil)
	}

	if options.InlineCode == "" {
		return "", errors.NewExecutionError(
			ErrExecutionFailed,
			"No Perl code provided to execute",
			nil)
	}

	// Resolve Perl version to use
	perlExe, err := resolvePerlExecutable(options)
	if err != nil {
		return "", err
	}

	// Build command arguments for inline code execution
	args := []string{"-e", options.InlineCode}

	// Add any script arguments if provided
	if len(options.Args) > 0 {
		args = append(args, options.Args...)
	}

	// Create the command to execute the Perl code
	cmd := exec.Command(perlExe, args...)

	// Set environment variables
	env, err := buildEnvironment(options)
	if err != nil {
		return "", err
	}
	cmd.Env = env

	// Configure output capture
	var outputBuffer bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer

	// Execute the command
	log.Debugf("Executing Perl code with command: %s -e '%s'", perlExe, options.InlineCode)
	err = execCommand(cmd)

	// Check for execution errors
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Get the exit code if possible
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
			// Cleanup isolation directory if needed
			cleanupIsolationDir(options)
			return outputBuffer.String(), errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Code execution failed with exit code %d", exitCode),
				err)
		}
		// Cleanup isolation directory if needed
		cleanupIsolationDir(options)
		return outputBuffer.String(), errors.NewExecutionError(
			ErrExecutionFailed,
			"Code execution failed",
			err)
	}

	// Cleanup isolation directory if needed
	cleanupIsolationDir(options)

	return outputBuffer.String(), nil
}

// cleanupIsolationDir removes the temporary isolation directory if one was created
// and if cleanup is not disabled
func cleanupIsolationDir(options *ExecutionOptions) {
	// Skip cleanup if explicitly disabled
	if options.NoCleanup {
		if options.Verbose && options.IsolationDir != "" {
			log.Infof("Isolation directory retained (--no-cleanup): %s", options.IsolationDir)
		}
		return
	}

	// Skip if no isolation directory was created
	if options.IsolationDir == "" ||
		options.IsolationLevel == "" ||
		options.IsolationLevel == IsolationNone {
		return
	}

	// Only cleanup directories created by us (auto-generated temp directories)
	// Skip cleanup for user-specified isolation directories
	if !strings.Contains(options.IsolationDir, "pvm-isolated-") {
		if options.Verbose {
			log.Infof("Skipping cleanup of user-specified isolation directory: %s", options.IsolationDir)
		}
		return
	}

	// Perform cleanup
	if options.Verbose {
		log.Infof("Cleaning up isolation directory: %s", options.IsolationDir)
	}

	err := os.RemoveAll(options.IsolationDir)
	if err != nil {
		log.Warnf("Failed to clean up isolation directory %s: %v", options.IsolationDir, err)
	} else if options.Verbose {
		log.Infof("Successfully removed isolation directory")
	}
}

// resolvePerlExecutable finds the appropriate Perl executable
func resolvePerlExecutable(options *ExecutionOptions) (string, error) {
	// Check if PerlVersion is an absolute path to a Perl executable
	if options.PerlVersion != "" && filepath.IsAbs(options.PerlVersion) {
		// Check if the file exists and is executable
		if _, err := os.Stat(options.PerlVersion); err == nil {
			log.Debugf("Using explicit Perl executable path: %s", options.PerlVersion)
			if options.Verbose {
				log.Infof("Using explicit Perl executable: %s", options.PerlVersion)
			}
			return options.PerlVersion, nil
		} else if !os.IsNotExist(err) {
			// Some other error occurred
			return "", errors.NewExecutionError(
				ErrVersionNotFound,
				fmt.Sprintf("Error accessing Perl executable: %s", options.PerlVersion),
				err)
		}
		// If file doesn't exist, continue with normal resolution
	}

	var resolvedVersion *perl.ResolvedVersion
	var err error

	// If a specific version is requested, use that
	if options.PerlVersion != "" {
		// Use version resolution to find the executable
		resolvedVersion, err = perl.ResolveVersion(&perl.ResolutionOptions{
			ExplicitVersion:     options.PerlVersion,
			SkipLocal:           true,
			SkipEnvVars:         true,
			SkipUserConfig:      true,
			SkipSystemPerl:      !options.ForceVersion, // Use system Perl as fallback if forcing version
			SkipVersionResolved: !options.Verbose,      // Only log if verbose
		})

		if err != nil {
			return "", errors.NewExecutionError(
				ErrVersionNotFound,
				fmt.Sprintf("Specified Perl version not found: %s", options.PerlVersion),
				err)
		}
	} else {
		// Use normal version resolution
		resolvedVersion, err = perl.ResolveVersion(&perl.ResolutionOptions{
			SkipVersionResolved: !options.Verbose, // Only log if verbose
		})

		if err != nil {
			return "", errors.NewExecutionError(
				ErrVersionNotFound,
				"Failed to resolve Perl version",
				err)
		}
	}

	// Get the path to the Perl executable for the resolved version
	var perlExe string
	if resolvedVersion.Source == perl.SystemPerlSource {
		// Use the path from the system Perl detection
		perlExe = resolvedVersion.Path
	} else {
		// For installed versions, get the installation info from the registry
		versionInfo, err := perl.GetVersionInfo(resolvedVersion.Version)
		if err == nil && versionInfo != nil {
			// Use the installation path from the version info
			perlExe = filepath.Join(versionInfo.InstallPath, "bin", "perl")
		} else {
			// If version info isn't available or there's an error, use the path from the resolver
			// This handles cases where the resolver has direct path information
			if resolvedVersion.Path != "" {
				perlExe = resolvedVersion.Path
			} else {
				// Fallback to a default path structure as a last resort
				perlExe = filepath.Join("/usr/local/pvm/perls", resolvedVersion.Version, "bin", "perl")
			}
		}

		// If we're forcing a version but it doesn't exist, fall back to system Perl
		fileErr := func() error {
			_, err := os.Stat(perlExe)
			return err
		}()
		if options.ForceVersion && os.IsNotExist(fileErr) {
			systemPerl, err := perl.DetectSystemPerl()
			if err != nil {
				return "", errors.NewExecutionError(
					ErrVersionNotFound,
					"Failed to find system Perl as fallback",
					err)
			}
			perlExe = systemPerl.Path
		}
	}

	// Check that the executable exists and is executable
	if _, err := os.Stat(perlExe); os.IsNotExist(err) {
		return "", errors.NewExecutionError(
			ErrVersionNotFound,
			fmt.Sprintf("Perl executable not found: %s", perlExe),
			err)
	}

	log.Debugf("Using Perl executable: %s (version %s from %s)",
		perlExe, resolvedVersion.Version, resolvedVersion.Source)
	if options.Verbose {
		log.Infof("Using Perl version %s via %s",
			resolvedVersion.Version, perlExe)
	}

	return perlExe, nil
}

// buildArguments creates the command line arguments for Perl
func buildArguments(options *ExecutionOptions) []string {
	args := []string{}

	// Add any Perl specific flags here
	// For example, we might add type checking flags if enabled
	if options.TypeCheck {
		args = append(args, "-T") // Simplified - we'd use more sophisticated type checking in reality
	}

	// Add the script path
	args = append(args, options.ScriptPath)

	// Add any script arguments
	if len(options.Args) > 0 {
		args = append(args, options.Args...)
	}

	return args
}

// buildEnvironment creates the environment variables for the script execution
func buildEnvironment(options *ExecutionOptions) ([]string, error) {
	// Start with the current environment
	env := os.Environ()

	// Add or override with custom environment variables
	for key, value := range options.Env {
		envVar := fmt.Sprintf("%s=%s", key, value)
		found := false

		// Replace existing variable if present
		for i, existing := range env {
			if strings.HasPrefix(existing, key+"=") {
				env[i] = envVar
				found = true
				break
			}
		}

		// Add new variable if not found
		if !found {
			env = append(env, envVar)
		}
	}

	// If legacy isolation flag is set, default to low isolation
	if options.Isolated && options.IsolationLevel == "" {
		options.IsolationLevel = IsolationLow
	}

	// Validate isolation level
	if options.IsolationLevel != "" && !ValidIsolationLevels[options.IsolationLevel] {
		return nil, errors.NewExecutionError(
			ErrInvalidIsolation,
			fmt.Sprintf("Invalid isolation level: %s", options.IsolationLevel),
			nil)
	}

	// If no isolation requested, return current environment
	if options.IsolationLevel == "" || options.IsolationLevel == IsolationNone {
		return env, nil
	}

	// For all other isolation levels, set up the isolation environment
	isolationDir := options.IsolationDir

	// If no isolation directory is specified, create one
	if isolationDir == "" {
		// Determine a unique name for the isolation directory
		scriptName := "inline"
		if options.ScriptPath != "" {
			scriptName = filepath.Base(options.ScriptPath)
			// Remove extension if present
			if ext := filepath.Ext(scriptName); ext != "" {
				scriptName = scriptName[:len(scriptName)-len(ext)]
			}
		}

		// Create a temporary directory for isolation
		var err error
		isolationDir, err = os.MkdirTemp("", fmt.Sprintf("pvm-isolated-%s-", scriptName))
		if err != nil {
			return nil, errors.NewExecutionError(
				ErrExecutionFailed,
				"Failed to create isolation directory",
				err)
		}

		// If verbose, log the isolation directory
		if options.Verbose {
			log.Infof("Created isolation directory: %s", isolationDir)
		}
	} else {
		// Ensure the specified isolation directory exists
		if err := os.MkdirAll(isolationDir, 0755); err != nil {
			return nil, errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Failed to create isolation directory at %s", isolationDir),
				err)
		}
	}

	// Create subdirectories for the Perl installation
	libDir := filepath.Join(isolationDir, "lib", "perl5")

	// Get the architecture using the perl executable
	archDir, err := getPerlArchDir(options.PerlVersion)
	if err != nil {
		// Fall back to a reasonable default
		log.Debugf("Failed to get Perl architecture, using default: %v", err)
		archDir = "darwin-2level" // Default for macOS
	}

	archLibDir := filepath.Join(libDir, archDir)
	binDir := filepath.Join(isolationDir, "bin")

	// Create the directories
	for _, dir := range []string{libDir, archLibDir, binDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Failed to create directory %s", dir),
				err)
		}
	}

	// Set up the environment based on isolation level
	switch options.IsolationLevel {
	case IsolationLow:
		// Low isolation: Add local::lib equivalent while preserving existing PERL5LIB
		perl5lib := fmt.Sprintf("%s:%s", libDir, archLibDir)

		// Add to existing PERL5LIB if present
		for i, existing := range env {
			if strings.HasPrefix(existing, "PERL5LIB=") {
				currentValue := strings.TrimPrefix(existing, "PERL5LIB=")
				if currentValue != "" {
					perl5lib = fmt.Sprintf("%s:%s", perl5lib, currentValue)
				}
				env[i] = fmt.Sprintf("PERL5LIB=%s", perl5lib)
				break
			}
		}

		// Set up the local::lib equivalent environment variables
		setEnvVar(&env, "PERL5LIB", perl5lib)
		setEnvVar(&env, "PERL_LOCAL_LIB_ROOT", isolationDir)
		setEnvVar(&env, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", isolationDir))
		setEnvVar(&env, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", isolationDir))

		// Add the bin directory to PATH
		for i, existing := range env {
			if strings.HasPrefix(existing, "PATH=") {
				currentPath := strings.TrimPrefix(existing, "PATH=")
				env[i] = fmt.Sprintf("PATH=%s:%s", binDir, currentPath)
				break
			}
		}

	case IsolationMedium:
		// Medium isolation: Clean PERL5LIB but preserve most environment variables

		// Set PERL5LIB to only include the isolation directory paths
		perl5lib := fmt.Sprintf("%s:%s", libDir, archLibDir)
		setEnvVar(&env, "PERL5LIB", perl5lib)

		// Set up the local::lib equivalent environment variables
		setEnvVar(&env, "PERL_LOCAL_LIB_ROOT", isolationDir)
		setEnvVar(&env, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", isolationDir))
		setEnvVar(&env, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", isolationDir))

		// Add the bin directory to PATH
		for i, existing := range env {
			if strings.HasPrefix(existing, "PATH=") {
				currentPath := strings.TrimPrefix(existing, "PATH=")
				env[i] = fmt.Sprintf("PATH=%s:%s", binDir, currentPath)
				break
			}
		}

	case IsolationHigh:
		// High isolation: Start with minimal environment and add only what's needed

		// Create a clean environment with only essential variables
		cleanEnv := []string{}

		// Copy only essential environment variables (non-exhaustive list)
		essentialVars := []string{
			"PATH",
			"HOME",
			"USER",
			"SHELL",
			"TMPDIR",
			"TERM",
		}

		for _, key := range essentialVars {
			for _, envVar := range env {
				if strings.HasPrefix(envVar, key+"=") {
					cleanEnv = append(cleanEnv, envVar)
					break
				}
			}
		}

		// Add custom environment variables
		for key, value := range options.Env {
			setEnvVar(&cleanEnv, key, value)
		}

		// Set PERL5LIB to only include the isolation directory paths
		perl5lib := fmt.Sprintf("%s:%s", libDir, archLibDir)
		setEnvVar(&cleanEnv, "PERL5LIB", perl5lib)

		// Set up the local::lib equivalent environment variables
		setEnvVar(&cleanEnv, "PERL_LOCAL_LIB_ROOT", isolationDir)
		setEnvVar(&cleanEnv, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", isolationDir))
		setEnvVar(&cleanEnv, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", isolationDir))

		// Set PATH to include the bin directory first
		for i, existing := range cleanEnv {
			if strings.HasPrefix(existing, "PATH=") {
				currentPath := strings.TrimPrefix(existing, "PATH=")
				cleanEnv[i] = fmt.Sprintf("PATH=%s:%s", binDir, currentPath)
				break
			}
		}

		// Use the clean environment instead of the original one
		env = cleanEnv
	}

	// Store the isolation directory in options for potential cleanup
	options.IsolationDir = isolationDir

	return env, nil
}

// getPerlArchDir determines the architecture directory for Perl modules
func getPerlArchDir(perlPath string) (string, error) {
	// If no perl path is provided, try to use system perl
	perlCmd := "perl"
	if perlPath != "" {
		// Check if it's an absolute path
		if filepath.IsAbs(perlPath) {
			// Use the provided perl executable
			perlCmd = perlPath
		}
	}

	// Run perl with the Config module to get the architecture directory
	cmd := exec.Command(perlCmd, "-MConfig", "-e", "print $Config{archname}")
	var out bytes.Buffer
	cmd.Stdout = &out

	// Execute the command
	if err := cmd.Run(); err != nil {
		return "", errors.NewExecutionError(
			ErrExecutionFailed,
			"Failed to determine Perl architecture",
			err)
	}

	// Get the output
	archName := strings.TrimSpace(out.String())
	if archName == "" {
		return "", errors.NewExecutionError(
			ErrExecutionFailed,
			"Empty architecture name returned",
			nil)
	}

	return archName, nil
}

// setEnvVar sets an environment variable, overriding if it already exists
func setEnvVar(env *[]string, key, value string) {
	envVar := fmt.Sprintf("%s=%s", key, value)
	for i, existing := range *env {
		if strings.HasPrefix(existing, key+"=") {
			(*env)[i] = envVar
			return
		}
	}
	*env = append(*env, envVar)
}
