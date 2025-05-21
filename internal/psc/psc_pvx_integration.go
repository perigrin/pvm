// ABOUTME: Integration between PSC and PVX components
// ABOUTME: Enables type-checked script execution

package psc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/pvi"
	"tamarou.com/pvm/internal/pvx"
)

// PSC-PVX integration error codes
const (
	ErrIntegrationFailed    = "PSC-801" // Integration between PSC and PVX failed
	ErrTypeCheckFailed      = "PSC-802" // Type checking failed
	ErrExecutionFailed      = "PSC-803" // Script execution failed
	ErrDependencyMissing    = "PSC-804" // Required dependency is missing
	ErrInvalidConfiguration = "PSC-805" // Invalid configuration
)

// TypeCheckedExecutionOptions contains options for type-checked execution
type TypeCheckedExecutionOptions struct {
	// Path to the script to execute
	ScriptPath string

	// Command line arguments to pass to the script
	Args []string

	// Custom Perl version to use
	PerlVersion string

	// Skip type checking
	SkipTypeCheck bool

	// Skip environment isolation
	DisableIsolation bool

	// Enable flow-sensitive analysis
	EnableFlowSensitiveAnalysis bool

	// Skip flow-sensitive checks but still perform refinements
	SkipFlowChecks bool

	// Custom environment variables
	EnvironmentVars map[string]string

	// Required modules
	RequiredModules []string

	// Enable verbose output
	Verbose bool

	// Module installation directory
	ModuleDir string

	// Skip cleanup after execution
	NoCleanup bool
}

// TypeCheckedExecutionResult contains the result of a type-checked execution
type TypeCheckedExecutionResult struct {
	// Output from the script execution
	Output string

	// TypeCheckPassed indicates if type checking was successful
	TypeCheckPassed bool

	// TypeCheckErrors contains any type checking errors
	TypeCheckErrors []parser.TypeCheckError

	// ExitCode is the exit code from the script execution
	ExitCode int
}

// ExecuteWithTypeChecking performs type checking and then executes a Perl script
func ExecuteWithTypeChecking(options *TypeCheckedExecutionOptions) (*TypeCheckedExecutionResult, error) {
	if options == nil {
		return nil, errors.NewTypeError(
			ErrIntegrationFailed,
			"No execution options provided",
			nil)
	}

	// Check if the script exists
	if _, err := os.Stat(options.ScriptPath); os.IsNotExist(err) {
		return nil, errors.NewTypeError(
			ErrIntegrationFailed,
			fmt.Sprintf("Script file not found: %s", options.ScriptPath),
			err).WithLocation(options.ScriptPath)
	}

	// Initialize the result
	result := &TypeCheckedExecutionResult{
		TypeCheckPassed: true,
	}

	// Perform type checking if not skipped
	if !options.SkipTypeCheck {
		tc, err := parser.NewTypeCheck()
		if err != nil {
			return nil, errors.NewTypeError(
				ErrTypeCheckFailed,
				"Failed to create type checker",
				err)
		}

		// Configure flow-sensitive analysis
		tc.EnableFlowSensitiveAnalysis = options.EnableFlowSensitiveAnalysis
		tc.SkipFlowChecks = options.SkipFlowChecks

		// Perform type checking
		checkResult, err := tc.CheckFile(options.ScriptPath)
		if err != nil {
			return nil, errors.NewTypeError(
				ErrTypeCheckFailed,
				fmt.Sprintf("Failed to check file: %s", options.ScriptPath),
				err).WithLocation(options.ScriptPath)
		}

		// Check if there were type errors
		if len(checkResult.Errors) > 0 {
			result.TypeCheckPassed = false
			result.TypeCheckErrors = checkResult.Errors

			// Log the errors
			if options.Verbose {
				fmt.Printf("Type checking failed for %s\n", options.ScriptPath)
				for _, err := range checkResult.Errors {
					fmt.Printf("  %s:%d:%d: %s\n", err.Path, err.Line, err.Column, err.Message)
				}
			}

			// Return without executing if type checking failed
			return result, errors.NewTypeError(
				ErrTypeCheckFailed,
				"Type checking failed, aborting execution",
				nil).WithLocation(options.ScriptPath)
		}

		// Extract required modules from type annotations
		requiredModules := extractRequiredModulesFromTypeAnnotations(checkResult)
		if len(requiredModules) > 0 {
			if options.Verbose {
				fmt.Printf("Detected %d required modules from type annotations\n", len(requiredModules))
				for _, mod := range requiredModules {
					fmt.Printf("  - %s\n", mod)
				}
			}

			// Add to required modules list if not already included
			for _, mod := range requiredModules {
				if !containsModule(options.RequiredModules, mod) {
					options.RequiredModules = append(options.RequiredModules, mod)
				}
			}
		}

		// Log success
		if options.Verbose {
			fmt.Printf("Type checking passed for %s\n", options.ScriptPath)
		}
	}

	// Prepare execution options for PVX
	isolationLevel := pvx.IsolationMedium
	if options.DisableIsolation {
		isolationLevel = pvx.IsolationNone
	}

	// Set up environment variables
	env := make(map[string]string)
	for k, v := range options.EnvironmentVars {
		env[k] = v
	}

	// Add PSC-specific environment variables
	env["PVM_TYPE_CHECKED"] = "1"

	// Create the execution options
	execOptions := &pvx.ExecutionOptions{
		ScriptPath:     options.ScriptPath,
		Args:           options.Args,
		PerlVersion:    options.PerlVersion,
		Env:            env,
		Verbose:        options.Verbose,
		IsolationLevel: isolationLevel,
		NoCleanup:      options.NoCleanup,
	}

	// If modules are required, ensure they are installed using PVI
	if len(options.RequiredModules) > 0 {
		if options.Verbose {
			fmt.Printf("Installing %d required modules using PVI\n", len(options.RequiredModules))
		}

		// Create PVI integration options
		pviOptions := &pvi.PVXIntegrationOptions{
			PerlVersion:     options.PerlVersion,
			RequiredModules: options.RequiredModules,
			InstallDir:      options.ModuleDir,
			Verbose:         options.Verbose,
			MaxRetries:      2,
			SkipTests:       true, // Skip tests for faster installation in execution context
		}

		// Install required modules using PVI
		installResult, err := pvi.InstallModulesForPVX(pviOptions)
		if err != nil {
			return result, errors.NewTypeError(
				ErrDependencyMissing,
				fmt.Sprintf("Failed to install required modules: %v", err),
				err).WithLocation(options.ScriptPath)
		}

		// Check if any modules failed to install
		if len(installResult.FailedModules) > 0 {
			return result, errors.NewTypeError(
				ErrDependencyMissing,
				fmt.Sprintf("Failed to install modules: %v", installResult.FailedModules),
				nil).WithLocation(options.ScriptPath)
		}

		// Set custom module path if installation directory is specified
		if options.ModuleDir != "" {
			execOptions.CustomModulePath = options.ModuleDir
		}

		if options.Verbose {
			fmt.Printf("Successfully installed %d modules, skipped %d already installed\n",
				len(installResult.InstalledModules), len(installResult.SkippedModules))
		}
	}

	// Execute the script
	output, err := pvx.ExecuteScript(execOptions)
	result.Output = output

	// Handle execution errors
	if err != nil {
		// Try to extract exit code
		if exitErr, ok := err.(interface{ ExitCode() int }); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}

		return result, errors.NewTypeError(
			ErrExecutionFailed,
			fmt.Sprintf("Script execution failed: %s", err.Error()),
			err).WithLocation(options.ScriptPath)
	}

	return result, nil
}

// extractRequiredModulesFromTypeAnnotations extracts required modules from type annotations
func extractRequiredModulesFromTypeAnnotations(result *parser.TypeCheckResult) []string {
	if result == nil || len(result.TypeAnnotations) == 0 {
		return nil
	}

	// Map to track unique module names
	moduleMap := make(map[string]bool)

	// Process type annotations
	for _, annotation := range result.TypeAnnotations {
		// Get the type expression as string
		typeStr := annotation.TypeExpression.String()

		// Extract module names from type expressions
		// In simple cases, this would be the part before "::"
		extractModulesFromType(typeStr, moduleMap)
	}

	// Convert map keys to slice
	modules := make([]string, 0, len(moduleMap))
	for module := range moduleMap {
		// Skip core modules and simple built-in types
		if isBuiltinType(module) {
			continue
		}
		modules = append(modules, module)
	}

	return modules
}

// extractModulesFromType extracts module names from a type expression
func extractModulesFromType(typeStr string, moduleMap map[string]bool) {
	// Handle parameterized types
	if strings.Contains(typeStr, "[") {
		baseType, _ := parser.ExtractTypeAndParams(typeStr)

		// Process the base type
		if strings.Contains(baseType, "::") {
			parts := strings.Split(baseType, "::")
			if len(parts) > 1 {
				moduleMap[strings.Join(parts[:len(parts)-1], "::")] = true
			}
		}

		// Process parameters
		return
	}

	// Handle regular types with namespace
	if strings.Contains(typeStr, "::") {
		parts := strings.Split(typeStr, "::")
		if len(parts) > 1 {
			moduleMap[strings.Join(parts[:len(parts)-1], "::")] = true
		}
	}
}

// isBuiltinType checks if a type is a built-in type
func isBuiltinType(typeName string) bool {
	// List of built-in types
	builtinTypes := map[string]bool{
		"Any":         true,
		"Scalar":      true,
		"Str":         true,
		"Num":         true,
		"Int":         true,
		"Float":       true,
		"Bool":        true,
		"Undef":       true,
		"Ref":         true,
		"ArrayRef":    true,
		"HashRef":     true,
		"CodeRef":     true,
		"RegexpRef":   true,
		"GlobRef":     true,
		"FileHandle":  true,
		"List":        true,
		"Array":       true,
		"Hash":        true,
		"Code":        true,
		"Glob":        true,
		"Maybe":       true,
		"Optional":    true,
		"Callable":    true,
		"Iterable":    true,
		"Positional":  true,
		"Associative": true,
		"IO":          true,
		"Path":        true,
		"File":        true,
		"Dir":         true,
	}

	// Check if the type is a built-in type
	// For composite names, check the first part
	if strings.Contains(typeName, "::") {
		parts := strings.Split(typeName, "::")
		return builtinTypes[parts[0]]
	}

	return builtinTypes[typeName]
}

// containsModule checks if a module is in a list of modules
func containsModule(modules []string, module string) bool {
	for _, m := range modules {
		if m == module {
			return true
		}
	}
	return false
}

// StripAndExecute strips type annotations and then executes a Perl script
func StripAndExecute(scriptPath string, args []string, perlVersion string, verbose bool) (string, error) {
	// Create a temporary file for the stripped code
	dir := filepath.Dir(scriptPath)
	baseName := filepath.Base(scriptPath)
	strippedPath := filepath.Join(dir, "."+baseName+".stripped.pl")

	// Strip type annotations
	strippedCode, err := parser.StripAnnotations(scriptPath)
	if err != nil {
		return "", errors.NewTypeError(
			ErrIntegrationFailed,
			"Failed to strip type annotations",
			err).WithLocation(scriptPath)
	}

	// Write the stripped code to a temporary file
	if err := os.WriteFile(strippedPath, []byte(strippedCode), 0644); err != nil {
		return "", errors.NewTypeError(
			ErrIntegrationFailed,
			"Failed to write stripped code",
			err).WithLocation(strippedPath)
	}

	// Set up deferred cleanup of the temporary file
	defer func() {
		if verbose {
			fmt.Printf("Cleaning up temporary file: %s\n", strippedPath)
		}
		_ = os.Remove(strippedPath)
	}()

	// Execute the stripped script using PVX
	execOptions := &pvx.ExecutionOptions{
		ScriptPath:  strippedPath,
		Args:        args,
		PerlVersion: perlVersion,
		Verbose:     verbose,
	}

	output, err := pvx.ExecuteScript(execOptions)
	if err != nil {
		return output, errors.NewTypeError(
			ErrExecutionFailed,
			"Failed to execute stripped script",
			err).WithLocation(strippedPath)
	}

	return output, nil
}
