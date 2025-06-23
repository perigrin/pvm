// ABOUTME: End-to-end workflows using multiple PVM ecosystem components
// ABOUTME: Demonstrates complete integration between PVM, PVX, PVI, and PSC

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/psc"
	"tamarou.com/pvm/internal/pvi"
	"tamarou.com/pvm/internal/pvx"
)

// Workflow error codes
const (
	ErrWorkflowFailed         = "WF-001" // General workflow failure
	ErrVersionResolution      = "WF-002" // Version resolution failed
	ErrTypeCheck              = "WF-003" // Type checking failed
	ErrModuleInstallation     = "WF-004" // Module installation failed
	ErrScriptExecution        = "WF-005" // Script execution failed
	ErrTypeDefinitionGenerate = "WF-006" // Type definition generation failed
)

// WorkflowOptions contains configuration for end-to-end workflows
type WorkflowOptions struct {
	// ScriptPath is the path to the main script
	ScriptPath string

	// PerlVersion specifies which Perl version to use
	PerlVersion string

	// Verbose enables detailed output
	Verbose bool

	// SkipTypeCheck disables type checking
	SkipTypeCheck bool

	// SkipModuleInstall disables automatic module installation
	SkipModuleInstall bool

	// SkipExecution disables script execution
	SkipExecution bool

	// GenerateTypeDefs enables type definition generation
	GenerateTypeDefs bool

	// RequiredModules lists additional modules to install
	RequiredModules []string

	// IsolationLevel for script execution
	IsolationLevel pvx.IsolationLevel

	// SaveTypeDefs saves generated type definitions
	SaveTypeDefs bool
}

// WorkflowResult contains the result of an end-to-end workflow
type WorkflowResult struct {
	// VersionUsed is the resolved Perl version
	VersionUsed string

	// TypeCheckPassed indicates if type checking succeeded
	TypeCheckPassed bool

	// TypeErrors contains any type checking errors
	TypeErrors []parser.TypeCheckError

	// ModulesInstalled lists installed modules
	ModulesInstalled []string

	// ModulesSkipped lists already-installed modules
	ModulesSkipped []string

	// ModulesFailed lists modules that failed to install
	ModulesFailed []string

	// ExecutionOutput contains script output
	ExecutionOutput string

	// ExecutionExitCode is the script's exit code
	ExecutionExitCode int

	// TypeDefGenerated indicates if type definitions were generated
	TypeDefGenerated bool

	// TypeDefPath is the path to generated type definitions
	TypeDefPath string

	// Duration is the total workflow time
	Duration time.Duration

	// Errors contains any workflow errors
	Errors []error
}

// CompleteWorkflow runs a complete end-to-end workflow
func CompleteWorkflow(options *WorkflowOptions) (*WorkflowResult, error) {
	if options == nil {
		return nil, errors.NewSystemError(
			ErrWorkflowFailed,
			"No workflow options provided",
			nil,
		)
	}

	startTime := time.Now()
	result := &WorkflowResult{
		Errors: []error{},
	}

	if options.Verbose {
		fmt.Printf("Starting complete workflow for: %s\n", options.ScriptPath)
	}

	// Step 1: Version Resolution
	if options.Verbose {
		fmt.Printf("Step 1: Resolving Perl version...\n")
	}

	resolvedVersion, err := resolveVersion(options)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result, err
	}
	result.VersionUsed = resolvedVersion.Version

	if options.Verbose {
		fmt.Printf("  Using Perl version: %s at %s\n", resolvedVersion.Version, resolvedVersion.Path)
	}

	// Step 2: Type Checking (if enabled)
	if !options.SkipTypeCheck {
		if options.Verbose {
			fmt.Printf("Step 2: Type checking...\n")
		}

		typeCheckResult, err := performTypeCheck(options)
		if err != nil {
			result.Errors = append(result.Errors, err)
			// Continue with execution even if type checking fails
		} else {
			result.TypeCheckPassed = true
			result.TypeErrors = typeCheckResult.Errors

			if len(typeCheckResult.Errors) > 0 {
				result.TypeCheckPassed = false
				if options.Verbose {
					fmt.Printf("  Type checking found %d errors\n", len(typeCheckResult.Errors))
				}
			} else if options.Verbose {
				fmt.Printf("  Type checking passed\n")
			}

			// Extract required modules from type annotations
			requiredFromTypes := extractModulesFromTypeCheck(typeCheckResult)
			if len(requiredFromTypes) > 0 {
				options.RequiredModules = append(options.RequiredModules, requiredFromTypes...)
				if options.Verbose {
					fmt.Printf("  Detected %d additional modules from type annotations\n", len(requiredFromTypes))
				}
			}
		}
	} else if options.Verbose {
		fmt.Printf("Step 2: Skipping type checking\n")
	}

	// Step 3: Module Installation (if needed)
	if !options.SkipModuleInstall && len(options.RequiredModules) > 0 {
		if options.Verbose {
			fmt.Printf("Step 3: Installing required modules...\n")
		}

		installResult, err := installModules(options, resolvedVersion)
		if err != nil {
			result.Errors = append(result.Errors, err)
			// Continue with execution even if module installation fails
		} else {
			result.ModulesInstalled = installResult.InstalledModules
			result.ModulesSkipped = installResult.SkippedModules
			result.ModulesFailed = installResult.FailedModules

			if options.Verbose {
				fmt.Printf("  Installed: %d, Skipped: %d, Failed: %d\n",
					len(result.ModulesInstalled),
					len(result.ModulesSkipped),
					len(result.ModulesFailed))
			}
		}
	} else if options.Verbose {
		if options.SkipModuleInstall {
			fmt.Printf("Step 3: Skipping module installation\n")
		} else {
			fmt.Printf("Step 3: No modules required\n")
		}
	}

	// Step 4: Script Execution (if not skipped)
	if !options.SkipExecution {
		if options.Verbose {
			fmt.Printf("Step 4: Executing script...\n")
		}

		execResult, err := executeScript(options, resolvedVersion)
		if err != nil {
			result.Errors = append(result.Errors, err)
			// Script execution failure is a terminal error
			result.Duration = time.Since(startTime)
			return result, err
		}

		result.ExecutionOutput = execResult.Output
		result.ExecutionExitCode = execResult.ExitCode

		if options.Verbose {
			fmt.Printf("  Script executed successfully (exit code: %d)\n", result.ExecutionExitCode)
		}
	} else if options.Verbose {
		fmt.Printf("Step 4: Skipping script execution\n")
	}

	// Step 5: Type Definition Generation (if enabled)
	if options.GenerateTypeDefs {
		if options.Verbose {
			fmt.Printf("Step 5: Generating type definitions...\n")
		}

		typeDefResult, err := generateTypeDefinitions(options)
		if err != nil {
			result.Errors = append(result.Errors, err)
			// Type definition generation failure is not terminal
		} else {
			result.TypeDefGenerated = true
			result.TypeDefPath = typeDefResult.SavedPath

			if options.Verbose {
				fmt.Printf("  Type definitions generated: %s\n", result.TypeDefPath)
			}
		}
	} else if options.Verbose {
		fmt.Printf("Step 5: Skipping type definition generation\n")
	}

	result.Duration = time.Since(startTime)

	if options.Verbose {
		fmt.Printf("Workflow completed in %v\n", result.Duration)
	}

	return result, nil
}

// TypeCheckWorkflow performs type checking with type definition support
func TypeCheckWorkflow(scriptPath string, perlVersion string, verbose bool) (*WorkflowResult, error) {
	options := &WorkflowOptions{
		ScriptPath:        scriptPath,
		PerlVersion:       perlVersion,
		Verbose:           verbose,
		SkipModuleInstall: true, // Only type check, don't install modules
		SkipExecution:     true, // Only type check, don't execute
		GenerateTypeDefs:  true,
		SaveTypeDefs:      true,
	}

	return CompleteWorkflow(options)
}

// ExecutionWorkflow performs module installation and script execution
func ExecutionWorkflow(scriptPath string, perlVersion string, verbose bool) (*WorkflowResult, error) {
	options := &WorkflowOptions{
		ScriptPath:     scriptPath,
		PerlVersion:    perlVersion,
		Verbose:        verbose,
		SkipTypeCheck:  true, // Only execute, don't type check
		IsolationLevel: pvx.IsolationClean,
	}

	return CompleteWorkflow(options)
}

// DevelopmentWorkflow performs the complete development cycle
func DevelopmentWorkflow(scriptPath string, perlVersion string) (*WorkflowResult, error) {
	options := &WorkflowOptions{
		ScriptPath:       scriptPath,
		PerlVersion:      perlVersion,
		Verbose:          true,
		GenerateTypeDefs: true,
		SaveTypeDefs:     true,
		IsolationLevel:   pvx.IsolationLocal,
	}

	return CompleteWorkflow(options)
}

// resolveVersion resolves the Perl version to use
func resolveVersion(options *WorkflowOptions) (*perl.ResolvedVersion, error) {
	// For integration tests, when no explicit version is provided, use system Perl directly
	if options.PerlVersion == "" {
		systemPerl, err := perl.DetectSystemPerl()
		if err != nil {
			return nil, errors.NewVersionError(
				ErrVersionResolution,
				"Failed to detect system Perl for integration test",
				err,
			)
		}

		resolved := &perl.ResolvedVersion{
			Version: systemPerl.Version,
			Source:  perl.SystemPerlSource,
			Path:    systemPerl.Path,
		}
		return resolved, nil
	}

	// For explicit versions, use normal resolution
	resolutionOptions := &perl.ResolutionOptions{
		ExplicitVersion: options.PerlVersion,
	}

	resolved, err := perl.ResolveVersion(resolutionOptions)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrVersionResolution,
			fmt.Sprintf("Failed to resolve Perl version: %s", options.PerlVersion),
			err,
		)
	}

	return resolved, nil
}

// performTypeCheck performs type checking on the script
func performTypeCheck(options *WorkflowOptions) (*parser.TypeCheckResult, error) {
	tc, err := parser.NewTypeCheck()
	if err != nil {
		return nil, errors.NewTypeError(
			ErrTypeCheck,
			"Failed to create type checker",
			err,
		)
	}

	result, err := tc.CheckFile(options.ScriptPath)
	if err != nil {
		return nil, errors.NewTypeError(
			ErrTypeCheck,
			fmt.Sprintf("Type checking failed for %s", options.ScriptPath),
			err,
		)
	}

	return result, nil
}

// extractModulesFromTypeCheck extracts required modules from type check results
func extractModulesFromTypeCheck(result *parser.TypeCheckResult) []string {
	moduleMap := make(map[string]bool)

	for _, annotation := range result.TypeAnnotations {
		typeStr := annotation.TypeExpression.String()

		// Extract module names from namespaced types
		if strings.Contains(typeStr, "::") {
			parts := strings.Split(typeStr, "::")
			if len(parts) > 1 {
				// Assume the module name is everything except the last part
				moduleName := strings.Join(parts[:len(parts)-1], "::")
				if !isBuiltinModule(moduleName) {
					moduleMap[moduleName] = true
				}
			}
		}
	}

	modules := make([]string, 0, len(moduleMap))
	for module := range moduleMap {
		modules = append(modules, module)
	}

	return modules
}

// installModules installs required modules
func installModules(options *WorkflowOptions, resolvedVersion *perl.ResolvedVersion) (*pvi.PVXIntegrationResult, error) {
	pviOptions := &pvi.PVXIntegrationOptions{
		PerlVersion:     resolvedVersion.Version,
		RequiredModules: options.RequiredModules,
		Verbose:         options.Verbose,
		SkipTests:       true, // Skip tests for faster workflow execution
		MaxRetries:      2,
	}

	result, err := pvi.InstallModulesForPVX(pviOptions)
	if err != nil {
		return nil, errors.NewModuleError(
			ErrModuleInstallation,
			"Failed to install required modules",
			err,
		)
	}

	return result, nil
}

// executeScript executes the script with the resolved version
func executeScript(options *WorkflowOptions, resolvedVersion *perl.ResolvedVersion) (*ExecutionResult, error) {
	execOptions := &pvx.ExecutionOptions{
		ScriptPath:     options.ScriptPath,
		PerlVersion:    resolvedVersion.Version,
		ForceVersion:   true, // Force fallback to system Perl when version not found
		Verbose:        options.Verbose,
		IsolationLevel: options.IsolationLevel,
		Timeout:        30 * time.Second, // Set reasonable timeout for integration tests
	}

	output, err := pvx.ExecuteScript(execOptions)
	if err != nil {
		// Check if this is an execution error with exit code
		if execErr, ok := err.(interface{ ExitCode() int }); ok {
			return &ExecutionResult{
				Output:   output,
				ExitCode: execErr.ExitCode(),
			}, nil
		}

		return nil, errors.NewExecutionError(
			ErrScriptExecution,
			fmt.Sprintf("Script execution failed: %s", options.ScriptPath),
			err,
		)
	}

	return &ExecutionResult{
		Output:   output,
		ExitCode: 0,
	}, nil
}

// generateTypeDefinitions generates type definitions for the script
func generateTypeDefinitions(options *WorkflowOptions) (*psc.TypeDefinitionResult, error) {
	// Extract module name from script path
	moduleName := extractModuleName(options.ScriptPath)

	typeDefOptions := &psc.TypeDefinitionOptions{
		ModuleName: moduleName,
		SourceFile: options.ScriptPath,
		Save:       options.SaveTypeDefs,
		Verbose:    options.Verbose,
	}

	result, err := psc.GenerateTypeDefinition(typeDefOptions)
	if err != nil {
		return nil, errors.NewTypeError(
			ErrTypeDefinitionGenerate,
			fmt.Sprintf("Failed to generate type definitions for %s", options.ScriptPath),
			err,
		)
	}

	return result, nil
}

// extractModuleName extracts a module name from a script path
func extractModuleName(scriptPath string) string {
	base := filepath.Base(scriptPath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// isBuiltinModule checks if a module name represents a builtin module
func isBuiltinModule(moduleName string) bool {
	builtins := map[string]bool{
		"strict":         true,
		"warnings":       true,
		"utf8":           true,
		"feature":        true,
		"Carp":           true,
		"List::Util":     true,
		"Scalar::Util":   true,
		"File::Basename": true,
		"File::Path":     true,
		"IO::Handle":     true,
		"Time::Local":    true,
		"Getopt::Long":   true,
		"Pod::Usage":     true,
		"Test::More":     true,
		"POSIX":          true,
		"Fcntl":          true,
		"Socket":         true,
		"Storable":       true,
		"Digest::MD5":    true,
		"MIME":           true,
		"Encode":         true,
		"threads":        true,
		"Thread":         true,
		"Config":         true,
		"Exporter":       true,
		"AutoLoader":     true,
		"SelfLoader":     true,
		"Benchmark":      true,
		"Class":          true,
		"base":           true,
		"parent":         true,
		"constant":       true,
		"vars":           true,
		"lib":            true,
		"overload":       true,
		"attributes":     true,
		"fields":         true,
	}

	// Check the first component of the module name
	parts := strings.Split(moduleName, "::")
	if len(parts) > 0 {
		return builtins[parts[0]]
	}

	return builtins[moduleName]
}

// ValidationWorkflow validates that all components work together
func ValidationWorkflow(testScript string) (*WorkflowResult, error) {
	// Create a temporary test script if none provided
	if testScript == "" {
		tempDir := os.TempDir()
		testScript = filepath.Join(tempDir, "pvm_integration_test.pl")

		testCode := `#!/usr/bin/env perl
use strict;
use warnings;

# Validation script compatible with system Perl
my $message = "Hello, PVM Integration!";
my $numbers = [1, 2, 3, 4, 5];

sub add {
    my ($a, $b) = @_;
    return $a + $b;
}

# Test the integration
print $message . "\n";
print "Sum of 2 + 3 = " . add(2, 3) . "\n";
print "Numbers: " . join(", ", @$numbers) . "\n";
`

		err := os.WriteFile(testScript, []byte(testCode), 0644)
		if err != nil {
			return nil, errors.NewSystemError(
				ErrWorkflowFailed,
				"Failed to create test script",
				err,
			)
		}

		defer func() { _ = os.Remove(testScript) }()
	}

	options := &WorkflowOptions{
		ScriptPath:       testScript,
		Verbose:          false, // Keep quiet for validation
		GenerateTypeDefs: true,
		SaveTypeDefs:     false, // Don't save for validation
		IsolationLevel:   pvx.IsolationClean,
	}

	return CompleteWorkflow(options)
}

// ExecutionResult represents the result of script execution
type ExecutionResult struct {
	Output   string
	ExitCode int
}
