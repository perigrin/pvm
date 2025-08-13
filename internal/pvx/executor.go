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
	"runtime"
	"strings"
	"syscall"
	"time"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/compiler"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/pm"
	"tamarou.com/pvm/internal/tool"
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

	// Whether to enable all features when executing inline code (used with -E flag)
	EnableFeatures bool

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

	// Whether to skip the type checking phase entirely (PSC integration)
	SkipTypeCheck bool

	// Whether to enable flow-sensitive type analysis (PSC integration)
	FlowSensitive bool

	// Whether to skip flow checks but still perform type refinements (PSC integration)
	SkipFlowChecks bool

	// Whether to enable verbose output
	Verbose bool

	// Whether to suppress all non-error output
	Quiet bool

	// Whether to enable debug output showing detailed version resolution process
	Debug bool

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
	var uiOut *ui.Output
	if len(uiOutput) > 0 && uiOutput[0] != nil {
		uiOut = uiOutput[0]
	}
	if options == nil {
		return "", errors.NewExecutionError(
			ErrExecutionFailed,
			"No execution options provided",
			nil)
	}

	// Set quiet mode if requested
	if options.Quiet {
		log.SetGlobalQuiet(true)
		// Defer restoring previous quiet state
		defer log.SetGlobalQuiet(false)
	}

	// Validate script path
	if _, err := os.Stat(options.ScriptPath); os.IsNotExist(err) {
		return "", errors.NewExecutionError(
			ErrScriptNotFound,
			fmt.Sprintf("Script file not found: %s", options.ScriptPath),
			err)
	}

	// Handle typed module dependencies if script contains type annotations
	var strippedModulePaths map[string]string
	var moduleCleanupFuncs []func()

	// Auto-detect dependencies using PSC parsing if enabled (superior to manual metadata)
	if options.AutoDetectDependencies {
		if uiOut != nil && options.Verbose {
			uiOut.Status("Analyzing script dependencies...")
		}

		var autoDeps []string
		var err error

		// Check if this is a typed script that needs enhanced dependency handling
		if ContainsTypeAnnotations(options.ScriptPath) {
			if uiOut != nil && options.Verbose {
				uiOut.Info("Script contains type annotations, using enhanced dependency detection...")
			}

			// Enhanced dependency detection that handles typed modules
			autoDeps, strippedModulePaths, moduleCleanupFuncs, err = AutoDetectAndStripTypedDependencies(options.ScriptPath, options)
			if err != nil {
				if options.Verbose {
					if uiOut != nil {
						uiOut.Warning("Could not auto-detect typed dependencies (continuing without): %v", err)
					} else {
						log.Infof("Could not auto-detect typed dependencies (continuing without): %v", err)
					}
				}
				autoDeps = []string{}
			}
		} else {
			// Standard dependency detection for non-typed scripts
			autoDeps, err = AutoDetectDependenciesWithOptions(options.ScriptPath, false) // Filter out core modules
			if err != nil {
				if options.Verbose {
					if uiOut != nil {
						uiOut.Warning("Could not auto-detect dependencies (continuing without): %v", err)
					} else {
						log.Infof("Could not auto-detect dependencies (continuing without): %v", err)
					}
				}
				autoDeps = []string{}
			}
		}

		// Merge auto-detected dependencies with execution options
		if len(autoDeps) > 0 {
			if options.Verbose {
				if uiOut != nil {
					uiOut.Info("Auto-detected %d dependencies from script", len(autoDeps))
					uiOut.List(autoDeps)
					if len(strippedModulePaths) > 0 {
						uiOut.Info("Stripped type annotations from %d typed modules", len(strippedModulePaths))
					}
				} else {
					log.Infof("Auto-detected %d dependencies from script: %v", len(autoDeps), autoDeps)
					if len(strippedModulePaths) > 0 {
						log.Infof("Stripped type annotations from %d typed modules", len(strippedModulePaths))
					}
				}
			}
			// Add auto-detected dependencies to required modules
			options.RequiredModules = append(options.RequiredModules, autoDeps...)
		}
	}

	// Add stripped module directories to module paths for typed modules
	if len(strippedModulePaths) > 0 {
		// Extract unique directories containing stripped modules
		strippedDirs := make(map[string]bool)
		for _, strippedPath := range strippedModulePaths {
			dir := filepath.Dir(strippedPath)
			strippedDirs[dir] = true
		}

		// Add these directories to AdditionalModulePaths so they're included in PERL5LIB
		for dir := range strippedDirs {
			options.AdditionalModulePaths = append(options.AdditionalModulePaths, dir)
			if options.Verbose {
				if uiOut != nil {
					uiOut.Info("Adding stripped module directory to module path: %s", dir)
				} else {
					log.Infof("Adding stripped module directory to module path: %s", dir)
				}
			}
		}
	}

	// Resolve Perl version to use - we need the resolved version for both execution and module installation
	if options.Debug {
		log.Infof("[DEBUG] Starting Perl version resolution process...")
		log.Infof("[DEBUG] Explicit version: %s", options.PerlVersion)
		log.Infof("[DEBUG] Script path: %s", options.ScriptPath)
		log.Infof("[DEBUG] Environment variables:")
		log.Infof("[DEBUG]   PVM_PERL_VERSION: %s", os.Getenv("PVM_PERL_VERSION"))
		log.Infof("[DEBUG]   PLENV_VERSION: %s", os.Getenv("PLENV_VERSION"))
		log.Infof("[DEBUG]   PERLBREW_PERL: %s", os.Getenv("PERLBREW_PERL"))
	}

	// Create resolution options with debug callback if needed
	resolutionOptions := &perl.ResolutionOptions{
		ExplicitVersion:     options.PerlVersion,
		ScriptPath:          options.ScriptPath,
		SkipVersionResolved: !options.Verbose, // Only log if verbose
	}

	// Add debug callback if debug mode is enabled
	if options.Debug {
		stepNumber := 0
		resolutionOptions.DebugCallback = func(step string, details interface{}) {
			stepNumber++
			log.Infof("[DEBUG]   %d. %s: %v", stepNumber, step, details)
		}
	}

	resolvedVersion, err := perl.ResolveVersion(resolutionOptions)
	if err != nil {
		if options.Debug {
			log.Infof("[DEBUG] Version resolution failed: %v", err)
		}
		return "", errors.NewExecutionError(
			ErrVersionNotFound,
			"Failed to resolve Perl version",
			err)
	}

	// Log resolution result
	if options.Debug {
		log.Infof("[DEBUG] Resolution result: %s (from %s)", resolvedVersion.Version, resolvedVersion.Source)
		log.Infof("[DEBUG] Source path: %s", resolvedVersion.SourcePath)
	}

	// Get the Perl executable path using the existing function
	if options.Debug {
		log.Infof("[DEBUG] Resolving Perl executable path...")
	}
	perlExe, err := resolvePerlExecutable(options)
	if err != nil {
		if options.Debug {
			log.Infof("[DEBUG] Failed to resolve Perl executable: %v", err)
		}
		return "", err
	}
	if options.Debug {
		log.Infof("[DEBUG] Perl executable: %s", perlExe)
	}

	// Install required modules using PVI if needed
	if options.AutoInstallModules && len(options.RequiredModules) > 0 {
		if uiOut != nil {
			uiOut.Status(fmt.Sprintf("Installing %d required modules using PVI", len(options.RequiredModules)))
		} else if options.Verbose {
			log.Infof("Installing %d required modules using PVI", len(options.RequiredModules))
		}

		// Create PVI integration options - use resolved version instead of raw options.PerlVersion
		pviOptions := &pm.PVXIntegrationOptions{
			PerlVersion:     resolvedVersion.Version, // Use resolved version instead of potentially empty options.PerlVersion
			RequiredModules: options.RequiredModules,
			InstallDir:      options.CustomModulePath,
			Verbose:         options.Verbose,
			MaxRetries:      2,
			SkipTests:       true, // Skip tests for faster installation
		}

		// Install required modules
		installResult, err := pm.InstallModulesForPVX(pviOptions)
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
			if uiOut != nil {
				uiOut.Success("Successfully installed %d modules, skipped %d already installed",
					len(installResult.InstalledModules), len(installResult.SkippedModules))
			} else {
				log.Infof("Successfully installed %d modules, skipped %d already installed",
					len(installResult.InstalledModules), len(installResult.SkippedModules))
			}
		}

		// Add the PVI installation directory to module paths so installed modules are available during execution
		// This ensures that modules installed by PVI are found in the execution environment
		pviInstallDir := pviOptions.InstallDir
		if pviInstallDir == "" {
			// If no custom install dir was specified, PVI will determine the install directory based on context
			// From the verbose log, we saw it uses "/home/perigrin/dev/pvm/local", which suggests it's using
			// a project-specific installation directory. For now, we'll need to get this from somewhere else
			// or make PVI return the actual installation directory used.
			//
			// As a temporary workaround, we'll check if there's a project context
			cfg, cfgErr := config.LoadEffectiveConfig()
			if cfgErr == nil && cfg != nil {
				// Check if we're in a project directory (pvm.toml exists)
				if _, err := os.Stat("pvm.toml"); err == nil {
					// We're in a project directory, use the local subdirectory
					pviInstallDir = "./local"
				}
			}
		}

		if pviInstallDir != "" {
			// Add the lib/perl5 subdirectory to the additional module paths
			pviModulePath := filepath.Join(pviInstallDir, "lib", "perl5")
			options.AdditionalModulePaths = append(options.AdditionalModulePaths, pviModulePath)
			if options.Verbose {
				if uiOut != nil {
					uiOut.Info("Adding PVI installation directory to module path: %s", pviModulePath)
				} else {
					log.Infof("Adding PVI installation directory to module path: %s", pviModulePath)
				}
			}
		}
	}

	// Check if script contains type annotations and handle PSC type checking if needed
	var needsCleanup bool

	if ContainsTypeAnnotations(options.ScriptPath) {
		// Perform PSC-style type checking if requested and not explicitly skipped
		if (options.TypeCheck || options.FlowSensitive) && !options.SkipTypeCheck {
			if options.Verbose {
				if uiOut != nil {
					uiOut.Info("Performing PSC-style type checking...")
				} else {
					log.Infof("Performing PSC-style type checking...")
				}
			}

			// Create type checker
			typeChecker, err := parser.NewTypeCheck()
			if err != nil {
				return "", errors.NewExecutionError(
					ErrExecutionFailed,
					fmt.Sprintf("Failed to create type checker: %v", err),
					err)
			}

			// Configure type checking options
			typeChecker.EnableFlowSensitiveAnalysis = options.FlowSensitive
			typeChecker.SkipFlowChecks = options.SkipFlowChecks

			// Perform type checking
			checkResult, err := typeChecker.CheckFile(options.ScriptPath)
			if err != nil {
				return "", errors.NewExecutionError(
					ErrExecutionFailed,
					fmt.Sprintf("Type checking failed: %v", err),
					err)
			}

			// Report type errors if any
			if len(checkResult.Errors) > 0 {
				errorMsg := fmt.Sprintf("Type checking found %d error(s):", len(checkResult.Errors))
				for _, typeErr := range checkResult.Errors {
					errorMsg += fmt.Sprintf("\n  %s:%d:%d: %s", typeErr.Path, typeErr.Line, typeErr.Column, typeErr.Message)
				}
				return "", errors.NewExecutionError(
					ErrExecutionFailed,
					errorMsg,
					nil)
			}

			if options.Verbose {
				if uiOut != nil {
					uiOut.Success("Type checking completed successfully")
				} else {
					log.Infof("Type checking completed successfully")
				}
			}
		}

		if options.Verbose {
			if uiOut != nil {
				uiOut.Info("Detected type annotations in script, stripping types for execution")
			} else {
				log.Infof("Detected type annotations in script, stripping types for execution")
			}
		}

		// Create a temporary file for the stripped code
		strippedPath, cleanupFunc, err := stripTypeAnnotationsFromScript(options.ScriptPath)
		if err != nil {
			return "", errors.NewExecutionError(
				ErrExecutionFailed,
				fmt.Sprintf("Failed to strip type annotations: %v", err),
				err)
		}

		// Update script path to the stripped version
		options.ScriptPath = strippedPath
		needsCleanup = true

		// Set up cleanup of the temporary stripped script file and modules
		defer func() {
			if needsCleanup {
				cleanupFunc()
			}
			// Clean up stripped module files
			for _, moduleCleanup := range moduleCleanupFuncs {
				if moduleCleanup != nil {
					moduleCleanup()
				}
			}
		}()

		if options.Verbose {
			if uiOut != nil {
				uiOut.Success("Successfully stripped type annotations, executing clean Perl code")
			} else {
				log.Infof("Successfully stripped type annotations, executing clean Perl code")
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

	// Set quiet mode if requested
	if options.Quiet {
		log.SetGlobalQuiet(true)
		// Defer restoring previous quiet state
		defer log.SetGlobalQuiet(false)
	}

	if options.InlineCode == "" {
		return "", errors.NewExecutionError(
			ErrExecutionFailed,
			"No Perl code provided to execute",
			nil)
	}

	// Resolve Perl version to use
	if options.Debug {
		log.Infof("[DEBUG] Starting Perl version resolution for inline code execution...")
		log.Infof("[DEBUG] Explicit version: %s", options.PerlVersion)
		log.Infof("[DEBUG] Environment variables:")
		log.Infof("[DEBUG]   PVM_PERL_VERSION: %s", os.Getenv("PVM_PERL_VERSION"))
		log.Infof("[DEBUG]   PLENV_VERSION: %s", os.Getenv("PLENV_VERSION"))
		log.Infof("[DEBUG]   PERLBREW_PERL: %s", os.Getenv("PERLBREW_PERL"))
	}
	perlExe, err := resolvePerlExecutable(options)
	if err != nil {
		if options.Debug {
			log.Infof("[DEBUG] Failed to resolve Perl executable for inline code: %v", err)
		}
		return "", err
	}
	if options.Debug {
		log.Infof("[DEBUG] Using Perl executable for inline code: %s", perlExe)
	}

	// Build command arguments for inline code execution
	var args []string
	if options.EnableFeatures {
		// Use -E flag to enable all features (equivalent to perl's -E)
		args = []string{"-E", options.InlineCode}
	} else {
		// Use standard -e flag
		args = []string{"-e", options.InlineCode}
	}

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

	// Use version resolution to get Perl executable
	resolvedVersion, err := perl.ResolveVersion(&perl.ResolutionOptions{
		ExplicitVersion:     options.PerlVersion,
		ScriptPath:          options.ScriptPath,
		SkipVersionResolved: !options.Verbose, // Only log if verbose
	})

	if err != nil {
		return "", errors.NewExecutionError(
			ErrVersionNotFound,
			"Failed to resolve Perl version",
			err)
	}

	// Get the path to the Perl executable for the resolved version
	var perlExe string
	if resolvedVersion.Path != "" {
		// Use the path provided by the resolver (system perl case)
		perlExe = resolvedVersion.Path
	} else {
		// Resolve the perl executable path for installed versions
		switch resolvedVersion.Source {
		case perl.SystemPerlSource:
			// This should not happen since system perl resolver provides the path
			return "", errors.NewExecutionError(
				ErrVersionNotFound,
				"System perl resolver did not provide executable path",
				nil)
		default:
			// For installed versions, get the installation info from the registry
			versionInfo, err := perl.GetVersionInfo(resolvedVersion.Version)
			if err != nil {
				// If ForceVersion is enabled and we can't find the version, fall back to system Perl
				if options.ForceVersion {
					systemPerl, sysErr := perl.DetectSystemPerl()
					if sysErr == nil {
						perlExe = systemPerl.Path
						break // Skip the rest of the switch case
					}
				}
				return "", errors.NewExecutionError(
					ErrVersionNotFound,
					"Failed to get version info for "+resolvedVersion.Version,
					err)
			}

			if versionInfo == nil {
				// If ForceVersion is enabled and version not found, fall back to system Perl
				if options.ForceVersion {
					systemPerl, sysErr := perl.DetectSystemPerl()
					if sysErr == nil {
						perlExe = systemPerl.Path
						break // Skip the rest of the switch case
					}
				}
				return "", errors.NewExecutionError(
					ErrVersionNotFound,
					"Version not found in registry: "+resolvedVersion.Version,
					nil)
			}

			// Construct the path to the Perl executable based on the source
			if versionInfo.Source == "system" {
				// For system perl, InstallPath is the directory containing the perl executable
				// Check if InstallPath already points to the perl executable
				if filepath.Base(versionInfo.InstallPath) == "perl" || filepath.Base(versionInfo.InstallPath) == "perl.exe" {
					perlExe = versionInfo.InstallPath
				} else {
					// InstallPath is the bin directory, append perl
					perlExe = filepath.Join(versionInfo.InstallPath, "perl")
					if runtime.GOOS == "windows" {
						perlExe = filepath.Join(versionInfo.InstallPath, "perl.exe")
					}
				}
			} else {
				// For PVM-installed versions, InstallPath is the installation root
				perlExe = filepath.Join(versionInfo.InstallPath, "bin", "perl")
				if runtime.GOOS == "windows" {
					perlExe = filepath.Join(versionInfo.InstallPath, "bin", "perl.exe")
				}
			}
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

	// Check that the executable exists and is executable (skip in test mode)
	if os.Getenv("PVM_SKIP_NETWORK_CALLS") == "" {
		if _, err := os.Stat(perlExe); os.IsNotExist(err) {
			return "", errors.NewExecutionError(
				ErrVersionNotFound,
				fmt.Sprintf("Perl executable not found: %s", perlExe),
				err)
		}
	}

	log.Debugf("Using Perl executable: %s (version %s from %s)",
		perlExe, resolvedVersion.Version, resolvedVersion.Source)
	if options.Verbose {
		log.Infof("Using Perl version %s via %s", resolvedVersion.Version, perlExe)
		log.Infof("Source: %s", getSourceDisplayName(resolvedVersion.Source, resolvedVersion.SourcePath))
	}

	return perlExe, nil
}

// getSourceDisplayName converts a resolution source to a user-friendly display name
func getSourceDisplayName(source perl.ResolutionSource, sourcePath string) string {
	switch source {
	case perl.ExplicitVersion:
		return "explicit version (command line)"
	case perl.ProjectVersionFile:
		if sourcePath != "" {
			return fmt.Sprintf(".perl-version file (%s)", sourcePath)
		}
		return ".perl-version file"
	case perl.ProjectConfig:
		if sourcePath != "" {
			return fmt.Sprintf("project configuration (%s)", sourcePath)
		}
		return "project configuration"
	case perl.EnvironmentVariable:
		if sourcePath != "" {
			return fmt.Sprintf("environment variable (%s)", sourcePath)
		}
		return "environment variable"
	case perl.UserConfig:
		if sourcePath != "" {
			return fmt.Sprintf("user configuration (%s)", sourcePath)
		}
		return "user configuration"
	case perl.SystemPerlSource:
		return "system Perl"
	default:
		return string(source)
	}
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

	// Ensure PVM-specific variables are preserved
	// These variables should always be inherited to ensure consistent behavior
	pvmVars := []string{
		"PVM_PERL_VERSION",
		"PVM_SUPPRESS_WARNINGS",
		"PVM_DEBUG",
		"PVM_CONFIG_DIR",
		"PVM_HOME",
		"PVM_SKIP_NETWORK_CALLS",
	}

	// Capture PVM-specific variables before any processing
	pvmVarValues := make(map[string]string)
	for _, key := range pvmVars {
		if value := os.Getenv(key); value != "" {
			pvmVarValues[key] = value
			if options.Verbose {
				log.Debugf("Preserving PVM variable: %s=%s", key, value)
			}
		}
	}

	// Process any environment variables that should be cleared
	if len(options.ClearEnv) > 0 {
		// Create a map for faster lookup
		clearEnvMap := make(map[string]bool)
		for _, key := range options.ClearEnv {
			clearEnvMap[key] = true
		}

		// Filter out environment variables that should be cleared
		// BUT preserve PVM-specific variables even if they're in ClearEnv
		filteredEnv := []string{}
		for _, envVar := range env {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) < 2 {
				continue
			}

			// Always preserve PVM-specific variables
			isPVMVar := false
			for _, pvmVar := range pvmVars {
				if parts[0] == pvmVar {
					isPVMVar = true
					break
				}
			}

			if !clearEnvMap[parts[0]] || isPVMVar {
				filteredEnv = append(filteredEnv, envVar)
			} else if options.Verbose {
				log.Debugf("Clearing environment variable: %s", parts[0])
			}
		}

		env = filteredEnv
	}

	// Ensure PVM-specific variables are present in the environment
	// This handles the case where they might have been filtered out or not inherited
	for key, value := range pvmVarValues {
		setEnvVar(&env, key, value)
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

	// Debug environment inheritance if verbose
	if options.Verbose {
		logEnvironmentInheritance(env, options)
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

// logEnvironmentInheritance logs environment inheritance information for debugging
func logEnvironmentInheritance(env []string, options *ExecutionOptions) {
	log.Infof("Environment inheritance debug information:")
	log.Infof("  Isolation level: %s", options.IsolationLevel)
	log.Infof("  Total environment variables: %d", len(env))

	// Log PVM-specific variables
	pvmVars := []string{
		"PVM_PERL_VERSION",
		"PVM_SUPPRESS_WARNINGS",
		"PVM_DEBUG",
		"PVM_CONFIG_DIR",
		"PVM_HOME",
		"PVM_SKIP_NETWORK_CALLS",
	}

	log.Infof("  PVM-specific variables:")
	for _, pvmVar := range pvmVars {
		found := false
		for _, envVar := range env {
			if strings.HasPrefix(envVar, pvmVar+"=") {
				log.Infof("    %s", envVar)
				found = true
				break
			}
		}
		if !found {
			log.Infof("    %s=<not set>", pvmVar)
		}
	}

	// Log key environment variables
	keyVars := []string{"PATH", "PERL5LIB", "HOME", "USER", "SHELL"}
	log.Infof("  Key environment variables:")
	for _, keyVar := range keyVars {
		found := false
		for _, envVar := range env {
			if strings.HasPrefix(envVar, keyVar+"=") {
				// Truncate long values for readability
				value := strings.TrimPrefix(envVar, keyVar+"=")
				if len(value) > 100 {
					value = value[:97] + "..."
				}
				log.Infof("    %s=%s", keyVar, value)
				found = true
				break
			}
		}
		if !found {
			log.Infof("    %s=<not set>", keyVar)
		}
	}
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

	// Skip cpanm shim for tool installation environments to prevent fork bomb
	// Tool installations use cpanm directly and creating a shim causes recursive pvm calls
	if strings.HasPrefix(options.EnvName, "tool-") {
		// Filter out cpanm from commands for tool installation environments
		filteredCommands := []string{}
		for _, cmd := range commands {
			if cmd != "cpanm" {
				filteredCommands = append(filteredCommands, cmd)
			}
		}
		commands = filteredCommands
		if options.Verbose {
			log.Infof("Excluded cpanm shim for tool installation environment '%s' to prevent fork bomb", options.EnvName)
		}
	}

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

	// Set quiet mode if requested
	if options != nil && options.Quiet {
		log.SetGlobalQuiet(true)
		// Defer restoring previous quiet state
		defer log.SetGlobalQuiet(false)
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
		// Use shared tool mapping from internal/tool package
		moduleMap := tool.GetBuiltinMappings()

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

// tryAutoInstallPerl attempts to auto-install a Perl version if auto-install is enabled
// This uses external pvm command to avoid import cycles
func tryAutoInstallPerl(version string, ui *ui.Output, verbose bool) error {
	// Load configuration to check if auto-install is enabled
	cfg, err := config.LoadEffectiveConfig()
	if err != nil {
		if verbose {
			if ui != nil {
				ui.Warning("Could not load configuration for auto-install check: %v", err)
			} else {
				log.Infof("Could not load configuration for auto-install check: %v", err)
			}
		}
		return fmt.Errorf("configuration not available")
	}

	// Check if auto-install is enabled
	if cfg == nil || cfg.PVX == nil || !cfg.PVX.AutoInstallPerl {
		if verbose {
			if ui != nil {
				ui.Info("Auto-install not enabled for Perl versions (set auto_install_perl = true in config)")
			} else {
				log.Infof("Auto-install not enabled for Perl versions")
			}
		}
		return fmt.Errorf("auto-install not enabled")
	}

	if verbose {
		if ui != nil {
			ui.Status(fmt.Sprintf("Auto-installing Perl %s using pvm install", version))
		} else {
			log.Infof("Auto-installing Perl %s using pvm install...", version)
		}
	}

	// Use external pvm command to install the required version
	// This avoids import cycles by calling pvm as external process
	pvmPath, err := exec.LookPath("pvm")
	if err != nil {
		return fmt.Errorf("pvm command not found in PATH: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, pvmPath, "install", version)
	if !verbose {
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	err = cmd.Run()
	if err != nil {
		if verbose {
			if ui != nil {
				ui.Error("Failed to auto-install Perl %s: %v", version, err)
			} else {
				log.Infof("Failed to auto-install Perl %s: %v", version, err)
			}
		}
		return fmt.Errorf("pvm install failed: %w", err)
	}

	if verbose {
		if ui != nil {
			ui.Success("Successfully auto-installed Perl %s", version)
		} else {
			log.Infof("Successfully auto-installed Perl %s", version)
		}
	}

	return nil
}

// ContainsTypeAnnotations checks if a Perl script contains type annotations
func ContainsTypeAnnotations(scriptPath string) bool {
	// Read the file content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		// If we can't read the file, assume no types (execution will fail later anyway)
		return false
	}

	source := string(content)

	// Quick text-based patterns to detect type annotations
	// These patterns are based on the ones found in enhanced_parser_impl.go
	// Note: We need to be more careful with common characters like !, |, & that can appear in strings
	typePatterns := []string{
		" as ",       // Type assertions: $value as Int
		"ArrayRef[",  // Parameterized types: ArrayRef[Int]
		"HashRef[",   // Parameterized types: HashRef[Str]
		"Maybe[",     // Maybe types: Maybe[Int]
		"Container[", // Container types
		"Map[",       // Map types
		"Wrapper[",   // Wrapper types
	}

	// Check for parameterized type patterns (these are quite specific)
	for _, pattern := range typePatterns {
		if strings.Contains(source, pattern) {
			return true
		}
	}

	// Check for type operators but be more careful about context
	// Look for type operators that are likely not in strings
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip string literals when looking for type operators
		if containsTypeOperators(line) {
			return true
		}
	}

	// Check for typed variable declarations (more complex pattern)
	// Look for patterns like: my Int $var, our String $VERSION, field Str $name
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for typed variable declarations
		if containsTypedVariableDeclaration(line) {
			return true
		}
	}

	return false
}

// containsTypedVariableDeclaration checks if a line contains a typed variable declaration
func containsTypedVariableDeclaration(line string) bool {
	// Remove leading/trailing whitespace and skip comments
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#") {
		return false
	}

	// Patterns for typed variable declarations:
	// my Type $var, our Type $var, field Type $name, etc.
	declarationKeywords := []string{"my ", "our ", "field ", "has "}

	for _, keyword := range declarationKeywords {
		if strings.HasPrefix(line, keyword) {
			// Extract the part after the keyword
			rest := strings.TrimSpace(line[len(keyword):])

			// Look for a pattern like "TypeName $variable"
			// This is a simple heuristic - a word followed by a space and then $variable
			parts := strings.Fields(rest)
			if len(parts) >= 2 {
				// Check if first part looks like a type name (starts with uppercase)
				typeName := parts[0]
				variablePart := parts[1]

				// Type names typically start with uppercase or contain "::"
				if (len(typeName) > 0 && (typeName[0] >= 'A' && typeName[0] <= 'Z')) || strings.Contains(typeName, "::") {
					// Variable part should start with $ or @or %
					if strings.HasPrefix(variablePart, "$") || strings.HasPrefix(variablePart, "@") || strings.HasPrefix(variablePart, "%") {
						return true
					}
				}
			}
		}
	}

	return false
}

// stripTypeAnnotationsFromScript strips type annotations from a script and returns a temporary file
// Returns: (tempFilePath, cleanupFunc, error)
func stripTypeAnnotationsFromScript(scriptPath string) (string, func(), error) {
	// Parse the file using the same approach as PSC strip command
	standardParser, err := parser.NewParser()
	if err != nil {
		return "", nil, fmt.Errorf("failed to create parser: %v", err)
	}

	ast, err := standardParser.ParseFile(scriptPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse file %s: %v", scriptPath, err)
	}

	// Use unified compiler to strip type annotations (same as PSC strip command)
	cleanCompiler := compiler.NewCleanPerlCompilerUnified()
	strippedCode, err := cleanCompiler.Compile(ast)
	if err != nil {
		return "", nil, fmt.Errorf("failed to strip type annotations from %s: %v", scriptPath, err)
	}

	// Create a temporary file for the stripped code
	// Use a predictable name pattern for easier debugging
	dir := filepath.Dir(scriptPath)
	baseName := filepath.Base(scriptPath)
	tempFile, err := os.CreateTemp(dir, ".pvx-stripped-"+baseName+"-*.pl")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary file: %v", err)
	}

	// Write the stripped code to the temporary file
	if _, err := tempFile.WriteString(strippedCode); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return "", nil, fmt.Errorf("failed to write stripped code: %v", err)
	}

	// Close the file so it can be executed
	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return "", nil, fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Return the temp file path and cleanup function
	tempPath := tempFile.Name()
	cleanupFunc := func() {
		if err := os.Remove(tempPath); err != nil {
			// Log the warning but don't fail
			log.Warnf("Failed to cleanup temporary stripped script file %s: %v", tempPath, err)
		}
	}

	return tempPath, cleanupFunc, nil
}

// containsTypeOperators checks if a line contains type operators in a type context
// (not inside strings or other contexts where they might appear naturally)
func containsTypeOperators(line string) bool {
	// Simple heuristic: if the line contains type operators and looks like a type context,
	// then it's likely a type annotation

	// Remove strings to avoid false positives
	withoutStrings := removeStringLiterals(line)

	// Look for type operators in non-string context
	typeOperators := []string{
		"|", // Union types: Int|Str
		"&", // Intersection types: Object&Serializable
		"!", // Negation types: !Undef
	}

	for _, op := range typeOperators {
		if strings.Contains(withoutStrings, op) {
			// Additional check: make sure this looks like a type context
			if looksLikeTypeContext(withoutStrings, op) {
				return true
			}
		}
	}

	return false
}

// removeStringLiterals removes string literals from a line to avoid false positives
func removeStringLiterals(line string) string {
	// Simple approach: remove content between quotes
	// This is not perfect but good enough for type detection
	result := line

	// Remove double-quoted strings
	for {
		start := strings.Index(result, "\"")
		if start == -1 {
			break
		}
		end := start + 1
		for end < len(result) {
			if result[end] == '"' && (end == 0 || result[end-1] != '\\') {
				break
			}
			end++
		}
		if end < len(result) {
			result = result[:start] + result[end+1:]
		} else {
			result = result[:start]
		}
	}

	// Remove single-quoted strings
	for {
		start := strings.Index(result, "'")
		if start == -1 {
			break
		}
		end := start + 1
		for end < len(result) {
			if result[end] == '\'' && (end == 0 || result[end-1] != '\\') {
				break
			}
			end++
		}
		if end < len(result) {
			result = result[:start] + result[end+1:]
		} else {
			result = result[:start]
		}
	}

	return result
}

// looksLikeTypeContext checks if a type operator appears in what looks like a type context
func looksLikeTypeContext(line, operator string) bool {
	// Look for patterns that suggest this is a type annotation:
	// - Variable declarations: my Type|Other $var
	// - Type definitions: type Foo = Bar|Baz
	// - Field declarations: field Type&Trait $field

	declarationKeywords := []string{"my ", "our ", "field ", "has ", "type ", "typedef ", "sub "}

	for _, keyword := range declarationKeywords {
		if strings.Contains(line, keyword) {
			// Check if the operator appears after the keyword
			keywordIndex := strings.Index(line, keyword)
			operatorIndex := strings.Index(line, operator)
			if operatorIndex > keywordIndex {
				return true
			}
		}
	}

	// Additional heuristic: if the operator is surrounded by what looks like type names
	// (words that start with uppercase or contain ::)
	operatorIndex := strings.Index(line, operator)
	if operatorIndex > 0 && operatorIndex < len(line)-1 {
		before := ""
		after := ""

		// Get word before operator
		i := operatorIndex - 1
		for i >= 0 && (line[i] == ' ' || line[i] == '\t') {
			i--
		}
		if i >= 0 {
			end := i + 1
			for i >= 0 && (line[i] != ' ' && line[i] != '\t') {
				i--
			}
			before = line[i+1 : end]
		}

		// Get word after operator
		i = operatorIndex + 1
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		if i < len(line) {
			start := i
			for i < len(line) && (line[i] != ' ' && line[i] != '\t') {
				i++
			}
			after = line[start:i]
		}

		// Check if before and after look like type names
		if looksLikeTypeName(before) && looksLikeTypeName(after) {
			return true
		}
	}

	return false
}

// looksLikeTypeName checks if a string looks like a type name
func looksLikeTypeName(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Type names typically start with uppercase or contain "::"
	return (s[0] >= 'A' && s[0] <= 'Z') || strings.Contains(s, "::")
}
