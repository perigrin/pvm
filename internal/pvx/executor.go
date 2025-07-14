// ABOUTME: PVX script execution functionality
// ABOUTME: Runs Perl scripts with specific versions and environment control

package pvx

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/pvi"
	"tamarou.com/pvm/internal/xdg"
)

// PVX execution error codes
const (
	ErrExecutionFailed  = "401" // Script execution failed
	ErrScriptNotFound   = "402" // Script file not found
	ErrVersionNotFound  = "403" // Specified Perl version not found
	ErrInvalidIsolation = "404" // Invalid isolation level specified
)

// Variables for mocking in tests
var execCommand = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

// resolvePerlExecutable is a variable to allow mocking in tests
var resolvePerlExecutable = resolvePerlExecutableImpl

// IsolationLevel defines the level of isolation for script execution
type IsolationLevel string

// Available isolation levels
const (
	// IsolationGlobal uses the system's Perl environment completely
	// - No isolation directory created
	// - Uses system modules and environment as-is
	// - Should be explicit opt-in, never default
	IsolationGlobal IsolationLevel = "global"

	// IsolationLocal creates local module installation capability
	// - Uses existing Perl environment, but allows installing modules locally
	// - Adds isolation directory to PERL5LIB (preserves system modules)
	// - Inherits all environment variables
	// - Default for project development work
	IsolationLocal IsolationLevel = "local"

	// IsolationClean provides clean module environment
	// - Creates clean PERL5LIB environment (no system modules)
	// - Still inherits most environment variables
	// - Only isolation directory modules available
	// - Default for tool execution and testing
	IsolationClean IsolationLevel = "clean"
)

// All valid isolation levels
var ValidIsolationLevels = map[IsolationLevel]bool{
	IsolationGlobal: true,
	IsolationLocal:  true,
	IsolationClean:  true,
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

	// Environment variables to preserve in isolation
	// This is useful for preserving specific environment variables when using PreserveEnv filtering
	PreserveEnv []string

	// Environment variables to clear/remove for isolation
	// This is useful for removing specific environment variables in all isolation modes
	ClearEnv []string

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

	// Name for persistent isolation environment
	EnvName string

	// ReadOnly paths for filesystem isolation - creates PVM_READONLY_PATHS environment variable
	// Any directories in this list will be accessible for reading but not writing
	ReadOnlyPaths []string

	// ReadWrite paths for filesystem isolation - creates PVM_READWRITE_PATHS environment variable
	// Any directories in this list will be accessible for both reading and writing
	ReadWritePaths []string

	// Whether to create a temporary directory for script output - creates PVM_ISOLATED_OUTPUT environment variable
	// When set to true, a temporary directory will be created and set as the working directory
	IsolatedOutput bool

	// Directory to save isolated output files to after execution
	// If set and IsolatedOutput is true, files will be copied from the isolated output directory
	SaveOutputDir string

	// Additional module paths to add to PERL5LIB
	// These paths will be added to the PERL5LIB environment variable in addition to the default ones
	AdditionalModulePaths []string

	// Custom module installation path
	// If set, this path will be used for module installation (PERL_LOCAL_LIB_ROOT)
	CustomModulePath string

	// Required modules to install before execution
	// These modules will be automatically installed using PVI if they're not available
	RequiredModules []string

	// Whether to automatically install required modules using PVI
	AutoInstallModules bool

	// Whether to automatically detect dependencies from use/require statements
	AutoDetectDependencies bool

	// Timeout for script execution (0 means no timeout)
	Timeout time.Duration
}

// ExecuteResult contains the result of script execution
type ExecuteResult struct {
	// The combined output (stdout and stderr)
	Output string

	// The exit code of the script
	ExitCode int
}

// ExecuteScript runs a Perl script with the specified options
func ExecuteScript(options *ExecutionOptions, uiOutput ...*ui.Output) (string, error) {
	// Get UI for user feedback (optional parameter for backward compatibility)
	var ui *ui.Output
	if len(uiOutput) > 0 && uiOutput[0] != nil {
		ui = uiOutput[0]
	}
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

	// Auto-detect dependencies using PSC parsing if enabled (superior to manual metadata)
	if options.AutoDetectDependencies {
		if ui != nil && options.Verbose {
			ui.Status("Analyzing script dependencies...")
		}
		autoDeps, err := AutoDetectDependenciesWithOptions(options.ScriptPath, false) // Filter out core modules
		if err != nil {
			if options.Verbose {
				if ui != nil {
					ui.Warning("Could not auto-detect dependencies (continuing without): %v", err)
				} else {
					log.Infof("Could not auto-detect dependencies (continuing without): %v", err)
				}
			}
			autoDeps = []string{}
		}

		// Merge auto-detected dependencies with execution options
		if len(autoDeps) > 0 {
			if options.Verbose {
				if ui != nil {
					ui.Info("Auto-detected %d dependencies from script", len(autoDeps))
					ui.List(autoDeps)
				} else {
					log.Infof("Auto-detected %d dependencies from script: %v", len(autoDeps), autoDeps)
				}
			}
			// Add auto-detected dependencies to required modules
			options.RequiredModules = append(options.RequiredModules, autoDeps...)
		}
	}

	// Resolve Perl version to use
	perlExe, err := resolvePerlExecutable(options)
	if err != nil {
		return "", err
	}

	// Install required modules using PVI if needed
	if options.AutoInstallModules && len(options.RequiredModules) > 0 {
		if ui != nil {
			ui.Status(fmt.Sprintf("Installing %d required modules using PVI", len(options.RequiredModules)))
		} else if options.Verbose {
			log.Infof("Installing %d required modules using PVI", len(options.RequiredModules))
		}

		// Create PVI integration options
		pviOptions := &pvi.PVXIntegrationOptions{
			PerlVersion:     options.PerlVersion,
			RequiredModules: options.RequiredModules,
			InstallDir:      options.CustomModulePath,
			Verbose:         options.Verbose,
			MaxRetries:      2,
			SkipTests:       true, // Skip tests for faster installation
		}

		// Install required modules
		installResult, err := pvi.InstallModulesForPVX(pviOptions)
		if err != nil {
			return "", errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Failed to install required modules: %v", err),
				err)
		}

		// Check if any modules failed to install
		if len(installResult.FailedModules) > 0 {
			return "", errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Failed to install modules: %v", installResult.FailedModules),
				nil)
		}

		if options.Verbose {
			if ui != nil {
				ui.Success("Successfully installed %d modules, skipped %d already installed",
					len(installResult.InstalledModules), len(installResult.SkippedModules))
			} else {
				log.Infof("Successfully installed %d modules, skipped %d already installed",
					len(installResult.InstalledModules), len(installResult.SkippedModules))
			}
		}
	}

	// Set up timeout context
	var ctx context.Context
	var cancel context.CancelFunc
	if options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), options.Timeout)
		defer cancel()
	} else {
		// Default timeout of 5 minutes for safety
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
	}

	// Create the command to execute the script with timeout
	cmd := exec.CommandContext(ctx, perlExe, buildArguments(options)...)

	// Set environment variables
	env, err := buildEnvironment(options)
	if err != nil {
		return "", err
	}
	cmd.Env = env

	// Note: Working directory setup for isolated output was removed - output directory is created separately

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
		// Check if this is a timeout error
		if ctx.Err() == context.DeadlineExceeded {
			cleanupIsolationDir(options)
			return outputBuffer.String(), errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Script execution timed out after %v", options.Timeout),
				err)
		}

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
func ExecuteInlineCode(options *ExecutionOptions, uiOutput ...*ui.Output) (string, error) {
	// Note: UI parameter preserved for backward compatibility but currently unused
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

	// Note: Working directory setup for isolated output was removed - output directory is created separately

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
		options.IsolationLevel == IsolationGlobal {
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

	// Check if output directory was created with isolated output flag
	if options.IsolatedOutput {
		outputDir := filepath.Join(options.IsolationDir, "output")
		if _, err := os.Stat(outputDir); err == nil {
			// Process output directory before cleanup if needed
			if options.Verbose {
				log.Infof("Output directory contains generated files at: %s", outputDir)

				// List files in the output directory
				files, err := os.ReadDir(outputDir)
				if err == nil && len(files) > 0 {
					log.Infof("Generated files in output directory:")
					for _, file := range files {
						log.Infof("  - %s", file.Name())
					}
				}
			}

			// If SaveOutputDir is specified, save the output files
			if options.SaveOutputDir != "" {
				if options.Verbose {
					log.Infof("Saving output files to: %s", options.SaveOutputDir)
				}

				savedFiles, err := saveOutputFiles(options, options.SaveOutputDir)
				if err != nil {
					log.Warnf("Failed to save output files: %v", err)
				} else if options.Verbose {
					log.Infof("Successfully saved %d output files", len(savedFiles))
				}
			}
		}
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

// resolvePerlExecutableImpl finds the appropriate Perl executable
// This is the actual implementation used by the resolvePerlExecutable variable
func resolvePerlExecutableImpl(options *ExecutionOptions) (string, error) {
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
	// The resolver now always provides the path to the perl executable
	if resolvedVersion.Path == "" {
		return "", errors.NewExecutionError(
			ErrVersionNotFound,
			"Resolver did not provide perl executable path for version "+resolvedVersion.Version,
			nil)
	}

	perlExe = resolvedVersion.Path

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

	// Process any environment variables that should be cleared
	if len(options.ClearEnv) > 0 {
		// Create a map for faster lookup
		clearEnvMap := make(map[string]bool)
		for _, key := range options.ClearEnv {
			clearEnvMap[key] = true
		}

		// Filter out environment variables that should be cleared
		filteredEnv := []string{}
		for _, envVar := range env {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) < 2 {
				continue
			}

			if !clearEnvMap[parts[0]] {
				filteredEnv = append(filteredEnv, envVar)
			} else if options.Verbose {
				log.Debugf("Clearing environment variable: %s", parts[0])
			}
		}

		env = filteredEnv
	}

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

	// If legacy isolation flag is set, default to local isolation
	if options.Isolated && options.IsolationLevel == "" {
		options.IsolationLevel = IsolationLocal
	}

	// Validate isolation level
	if options.IsolationLevel != "" && !ValidIsolationLevels[options.IsolationLevel] {
		return nil, errors.NewExecutionError(
			ErrInvalidIsolation,
			fmt.Sprintf("Invalid isolation level: %s", options.IsolationLevel),
			nil)
	}

	// If global isolation requested, return current environment
	if options.IsolationLevel == IsolationGlobal {
		if options.Verbose {
			log.Infof("Using global isolation - script will run in the current environment")
		}
		return env, nil
	}

	// For all other isolation levels, set up the isolation environment
	isolationDir := options.IsolationDir

	// If no isolation directory is specified, create one
	if isolationDir == "" {
		if options.EnvName != "" {
			// Create named persistent environment
			dirs, err := xdg.GetDirs()
			if err != nil {
				return nil, errors.NewExecutionError(
					ErrExecutionFailed,
					"Failed to determine XDG directories",
					err)
			}

			// Create environments directory
			envDir := filepath.Join(dirs.DataDir, "environments")
			if err := os.MkdirAll(envDir, 0755); err != nil {
				return nil, errors.NewExecutionError(
					ErrExecutionFailed,
					"Failed to create environments directory",
					err)
			}

			// Use named environment directory
			isolationDir = filepath.Join(envDir, options.EnvName)

			// Create the named environment directory
			if err := os.MkdirAll(isolationDir, 0755); err != nil {
				return nil, errors.NewExecutionError(
					ErrExecutionFailed,
					fmt.Sprintf("Failed to create named environment '%s'", options.EnvName),
					err)
			}

			// Force no cleanup for named environments
			options.NoCleanup = true

			if options.Verbose {
				log.Infof("Created/using named environment '%s' at: %s", options.EnvName, isolationDir)
			}
		} else {
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
	siteDir := filepath.Join(isolationDir, "lib", "site_perl")
	vendorDir := filepath.Join(isolationDir, "lib", "vendor_perl")

	// Create the standard directories
	dirsToCreate := []string{libDir, archLibDir, binDir}

	// If clean isolation, create more complete directory structure
	if options.IsolationLevel == IsolationClean {
		dirsToCreate = append(dirsToCreate,
			siteDir,
			vendorDir,
			filepath.Join(isolationDir, "man"),
			filepath.Join(isolationDir, "etc"),
			filepath.Join(isolationDir, "share"))

		if options.Verbose {
			log.Infof("Creating complete Perl module directory structure for %s isolation", options.IsolationLevel)
		}
	}

	// Create the directories
	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Failed to create directory %s", dir),
				err)
		}
	}

	// Set up the environment based on isolation level
	switch options.IsolationLevel {
	case IsolationLocal:
		// Local isolation: Add local::lib equivalent while preserving existing PERL5LIB
		if options.Verbose {
			log.Infof("Using local isolation - preserving environment but adding local module paths")
		}

		perl5lib := fmt.Sprintf("%s:%s", libDir, archLibDir)

		// Add any user-specified additional module paths
		for _, path := range options.AdditionalModulePaths {
			perl5lib = fmt.Sprintf("%s:%s", perl5lib, path)
			if options.Verbose {
				log.Infof("Adding module path to PERL5LIB: %s", path)
			}
		}

		// Add to existing PERL5LIB if present
		perl5LibFound := false
		for i, existing := range env {
			if strings.HasPrefix(existing, "PERL5LIB=") {
				currentValue := strings.TrimPrefix(existing, "PERL5LIB=")
				if currentValue != "" {
					perl5lib = fmt.Sprintf("%s:%s", perl5lib, currentValue)
				}
				env[i] = fmt.Sprintf("PERL5LIB=%s", perl5lib)
				perl5LibFound = true
				break
			}
		}

		// Add PERL5LIB if not found in environment
		if !perl5LibFound {
			env = append(env, fmt.Sprintf("PERL5LIB=%s", perl5lib))
		}

		// Use the custom module path if provided, otherwise use the isolation directory
		modulePath := isolationDir
		if options.CustomModulePath != "" {
			modulePath = options.CustomModulePath
			if options.Verbose {
				log.Infof("Using custom module path: %s", modulePath)
			}
		}

		// Set up the local::lib equivalent environment variables
		setEnvVar(&env, "PERL_LOCAL_LIB_ROOT", modulePath)
		setEnvVar(&env, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", modulePath))
		setEnvVar(&env, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", modulePath))

		// Add the bin directory to PATH
		pathFound := false
		for i, existing := range env {
			if strings.HasPrefix(existing, "PATH=") {
				currentPath := strings.TrimPrefix(existing, "PATH=")
				env[i] = fmt.Sprintf("PATH=%s:%s", binDir, currentPath)
				pathFound = true
				break
			}
		}

		// Add PATH if not found (unlikely, but just in case)
		if !pathFound {
			env = append(env, fmt.Sprintf("PATH=%s", binDir))
		}

	case IsolationClean:
		// Clean isolation: Clean PERL5LIB but preserve most environment variables
		if options.Verbose {
			log.Infof("Using clean isolation - preserving most environment but cleaning PERL5LIB")
		}

		// Set PERL5LIB to include all the isolation directory paths
		perl5lib := fmt.Sprintf("%s:%s:%s:%s",
			libDir,
			archLibDir,
			siteDir,
			vendorDir)

		// Add any user-specified additional module paths
		for _, path := range options.AdditionalModulePaths {
			perl5lib = fmt.Sprintf("%s:%s", perl5lib, path)
			if options.Verbose {
				log.Infof("Adding module path to PERL5LIB: %s", path)
			}
		}

		setEnvVar(&env, "PERL5LIB", perl5lib)

		// Set up the local::lib equivalent environment variables

		// Use the custom module path if provided, otherwise use the isolation directory
		modulePath := isolationDir
		if options.CustomModulePath != "" {
			modulePath = options.CustomModulePath
			if options.Verbose {
				log.Infof("Using custom module path: %s", modulePath)
			}
		}

		setEnvVar(&env, "PERL_LOCAL_LIB_ROOT", modulePath)
		setEnvVar(&env, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", modulePath))
		setEnvVar(&env, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", modulePath))

		// Add the bin directory to PATH
		pathFound := false
		for i, existing := range env {
			if strings.HasPrefix(existing, "PATH=") {
				currentPath := strings.TrimPrefix(existing, "PATH=")
				env[i] = fmt.Sprintf("PATH=%s:%s", binDir, currentPath)
				pathFound = true
				break
			}
		}

		// Add PATH if not found
		if !pathFound {
			env = append(env, fmt.Sprintf("PATH=%s", binDir))
		}

	}

	// Store the isolation directory in options for potential cleanup
	options.IsolationDir = isolationDir

	// Set PVM_ISOLATION_DIR environment variable for scripts to use
	setEnvVar(&env, "PVM_ISOLATION_DIR", isolationDir)

	// Generate shims for named environments
	if options.EnvName != "" {
		err := generateEnvironmentShims(options, isolationDir)
		if err != nil {
			log.Warnf("Failed to generate shims for environment '%s': %v", options.EnvName, err)
		} else if options.Verbose {
			log.Infof("Generated shims for environment '%s'", options.EnvName)
		}
	}

	return env, nil
}

// saveOutputFiles copies generated files from the isolated output directory to a specified target location
// Returns a map of saved file paths to their content, and any error encountered
func saveOutputFiles(options *ExecutionOptions, targetDir string) (map[string]string, error) {
	// Exit early if isolated output is not enabled or isolation directory is not set
	if !options.IsolatedOutput || options.IsolationDir == "" {
		if options.Verbose {
			log.Infof("Output files were not saved: isolated output is not enabled or isolation directory is not set")
		}
		return nil, nil
	}

	// Check if the output directory exists
	outputDir := filepath.Join(options.IsolationDir, "output")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if options.Verbose {
			log.Infof("Output directory does not exist: %s", outputDir)
		}
		return nil, nil
	}

	// Expand any environment variables in the target directory
	// This is particularly useful for configurations that use $PWD or $HOME
	expandedTargetDir := targetDir
	if strings.Contains(targetDir, "$") {
		for _, envVar := range os.Environ() {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				if strings.Contains(expandedTargetDir, "$"+key) {
					expandedTargetDir = strings.ReplaceAll(expandedTargetDir, "$"+key, value)
				}
			}
		}
		// Special handling for $PWD if not already replaced
		if strings.Contains(expandedTargetDir, "$PWD") {
			pwd, err := os.Getwd()
			if err == nil {
				expandedTargetDir = strings.ReplaceAll(expandedTargetDir, "$PWD", pwd)
			}
		}
		if options.Verbose && expandedTargetDir != targetDir {
			log.Debugf("Expanded target directory from %s to %s", targetDir, expandedTargetDir)
		}
		targetDir = expandedTargetDir
	}

	// Make sure the target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, errors.NewExecutionError(
			ErrExecutionFailed,
			fmt.Sprintf("Failed to create target directory %s", targetDir),
			err)
	}

	// Copy all files from the output directory to the target directory
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, errors.NewExecutionError(
			ErrExecutionFailed,
			fmt.Sprintf("Failed to read output directory %s", outputDir),
			err)
	}

	// Check if there are any files to copy
	if len(files) == 0 {
		if options.Verbose {
			log.Infof("No output files found in %s", outputDir)
			// List contents of parent directory for debugging
			parentFiles, _ := os.ReadDir(options.IsolationDir)
			log.Infof("Contents of isolation directory %s:", options.IsolationDir)
			for _, f := range parentFiles {
				log.Infof("  - %s", f.Name())
			}
		}
		return nil, nil
	}

	if options.Verbose {
		log.Infof("Copying %d files from %s to %s", len(files), outputDir, targetDir)
		log.Infof("Files to be copied:")
		for _, f := range files {
			log.Infof("  - %s", f.Name())
		}
	}

	result := make(map[string]string)
	copiedCount := 0
	skippedDirs := 0
	skippedErrors := 0

	// Copy each file
	for _, file := range files {
		if file.IsDir() {
			if options.Verbose {
				log.Debugf("Skipping directory: %s", file.Name())
			}
			skippedDirs++
			continue // Skip directories
		}

		// Read the source file
		srcPath := filepath.Join(outputDir, file.Name())
		data, err := os.ReadFile(srcPath)
		if err != nil {
			log.Warnf("Failed to read file %s: %v", srcPath, err)
			skippedErrors++
			continue // Skip files that can't be read
		}

		// Get file permissions
		srcInfo, err := os.Stat(srcPath)
		var srcMode os.FileMode
		if err == nil {
			srcMode = srcInfo.Mode().Perm()
		} else {
			srcMode = 0644 // Default permissions if stat fails
		}

		// Write to the target location
		targetPath := filepath.Join(targetDir, file.Name())
		err = os.WriteFile(targetPath, data, srcMode)
		if err != nil {
			log.Warnf("Failed to write file %s: %v", targetPath, err)
			skippedErrors++
			continue // Skip files that can't be written
		}

		// Store in the result map
		result[targetPath] = string(data)
		copiedCount++

		if options.Verbose {
			log.Infof("Saved output file: %s (%d bytes, mode %s)",
				targetPath, len(data), srcMode.String())
		}
	}

	// Log summary of operation
	if options.Verbose {
		log.Infof("Successfully saved %d output files to %s (skipped %d directories, %d errors)",
			copiedCount, targetDir, skippedDirs, skippedErrors)
	}

	return result, nil
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

	// Create context with timeout for perl command
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run perl with the Config module to get the architecture directory
	cmd := exec.CommandContext(ctx, perlCmd, "-MConfig", "-e", "print $Config{archname}")
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

// generateEnvironmentShims creates shims for a named environment that preserve the isolation settings
func generateEnvironmentShims(options *ExecutionOptions, isolationDir string) error {
	// Create bin directory in the environment
	binDir := filepath.Join(isolationDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Get the PVM executable path
	pvmPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine PVM executable path: %w", err)
	}

	// Standard Perl commands to create shims for
	commands := []string{"perl", "cpan", "prove", "perldoc", "h2ph", "h2xs", "enc2xs", "xsubpp", "corelist", "cpanm"}

	// Generate shim for each command
	for _, command := range commands {
		shimPath := filepath.Join(binDir, command)

		// Create shim script that uses this environment
		shimContent := generateEnvironmentShimScript(command, options.EnvName, pvmPath, options.PerlVersion, options.IsolationLevel)

		// Write shim file
		if err := os.WriteFile(shimPath, []byte(shimContent), 0755); err != nil {
			return fmt.Errorf("failed to create shim for %s: %w", command, err)
		}
	}

	return nil
}

// generateEnvironmentShimScript creates the content for an environment-specific shim
func generateEnvironmentShimScript(command, envName, pvmPath, perlVersion string, isolationLevel IsolationLevel) string {
	template := `#!/usr/bin/env sh
# Generated by PVM - Environment-specific shim for: %s
# Environment: %s

# Get the PVM executable
PVM_EXEC="%s"

# Execute the command through PVM with environment settings
if [ -x "$PVM_EXEC" ]; then
  exec "$PVM_EXEC" pvx --name "%s" --isolation=%s%s --no-cleanup -e 'exec { "%s" } "%s", @ARGV;' -- "$@"
else
  echo "Error: PVM executable not found at '$PVM_EXEC'"
  echo "Please ensure PVM is installed correctly"
  exit 1
fi
`

	// Add perl version flag if specified
	perlFlag := ""
	if perlVersion != "" {
		perlFlag = fmt.Sprintf(" --perl=%s", perlVersion)
	}

	return fmt.Sprintf(template, command, envName, pvmPath, envName, isolationLevel, perlFlag, command, command)
}

// ExecuteTool executes a Perl tool directly (similar to uvx)
func ExecuteTool(options *ExecutionOptions, toolName string, toolArgs []string, uiOutput ...*ui.Output) (string, error) {
	// Get UI for user feedback (optional parameter for backward compatibility)
	var ui *ui.Output
	if len(uiOutput) > 0 && uiOutput[0] != nil {
		ui = uiOutput[0]
	}
	// Create a temporary inline code that invokes the tool
	// This allows us to leverage existing isolation infrastructure

	// Build the command to execute the tool
	var toolCode string
	if len(toolArgs) == 0 {
		// Just run the tool with no arguments
		toolCode = fmt.Sprintf("exec { '%s' } '%s';", toolName, toolName)
	} else {
		// Run the tool with the provided arguments
		// Need to properly quote arguments for Perl
		quotedArgs := make([]string, len(toolArgs))
		for i, arg := range toolArgs {
			// Simple quoting - escape single quotes and wrap in single quotes
			escaped := strings.ReplaceAll(arg, "'", "\\'")
			quotedArgs[i] = fmt.Sprintf("'%s'", escaped)
		}
		argsStr := strings.Join(quotedArgs, ", ")
		toolCode = fmt.Sprintf("exec { '%s' } '%s', %s;", toolName, toolName, argsStr)
	}

	// Set up options for inline code execution
	options.InlineCode = toolCode

	// Add the tool to required modules if it's not a standard command
	standardCommands := []string{"perl", "cpan", "prove", "perldoc", "h2ph", "h2xs", "enc2xs", "xsubpp", "corelist"}
	isStandard := false
	for _, cmd := range standardCommands {
		if toolName == cmd {
			isStandard = true
			break
		}
	}

	if !isStandard {
		// For non-standard tools, we need to map tool names to module names
		// This is a basic mapping - could be expanded
		moduleMap := map[string]string{
			"cpanm":      "App::cpanminus",
			"plackup":    "Plack",
			"carton":     "Carton",
			"dzil":       "Dist::Zilla",
			"perlcritic": "Perl::Critic",
			"perltidy":   "Perl::Tidy",
		}

		if module, exists := moduleMap[toolName]; exists {
			options.RequiredModules = append(options.RequiredModules, module)
			options.AutoInstallModules = true
		} else {
			// Auto-discovery: try to install a module with the same name as the tool
			// Convert tool name to likely module name (capitalize first letter)
			moduleName := strings.ToTitle(toolName)
			options.RequiredModules = append(options.RequiredModules, moduleName)
			options.AutoInstallModules = true
		}
	}

	// Execute as inline code
	return ExecuteInlineCode(options, ui)
}
