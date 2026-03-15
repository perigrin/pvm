// ABOUTME: End-to-end workflows using multiple PVM ecosystem components
// ABOUTME: Stub implementation - type-system dependent workflows return "not yet available"

package integration

import (
	"fmt"
	"time"

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

// TypeCheckError represents a type-checking error from a Perl file
type TypeCheckError struct {
	// Path is the file path where the error occurred
	Path string

	// Line is the line number of the error
	Line int

	// Column is the column number of the error
	Column int

	// Message is the error message
	Message string
}

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

	// PVX-specific options for workflow orchestration
	// EnvironmentName for named persistent environments
	EnvironmentName string

	// PreserveEnv lists environment variables to preserve
	PreserveEnv []string

	// ClearEnv lists environment variables to clear
	ClearEnv []string

	// ModulePaths contains additional PERL5LIB paths
	ModulePaths []string

	// ExecuteCode for direct code execution (-e flag)
	ExecuteCode string

	// ExecuteFeatures for direct code execution with features (-E flag)
	ExecuteFeatures string

	// AutoDetectDeps enables automatic dependency detection
	AutoDetectDeps bool

	// ForceVersion forces using the specified version
	ForceVersion bool

	// NoInstall disables automatic installation
	NoInstall bool

	// ReadOnlyPaths for filesystem isolation
	ReadOnlyPaths []string

	// ReadWritePaths for filesystem isolation
	ReadWritePaths []string

	// IsolatedOutput enables output isolation
	IsolatedOutput bool

	// SaveOutputDir for saving isolated output
	SaveOutputDir string

	// NoCleanup disables cleanup after execution
	NoCleanup bool

	// AdditionalArgs for script execution
	AdditionalArgs []string
}

// WorkflowResult contains the result of an end-to-end workflow
type WorkflowResult struct {
	// VersionUsed is the resolved Perl version
	VersionUsed string

	// TypeCheckPassed indicates if type checking succeeded
	TypeCheckPassed bool

	// TypeErrors contains any type checking errors
	TypeErrors []TypeCheckError

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

// CompleteWorkflow runs a complete end-to-end workflow.
// Type-checking steps are stubbed out until type-system components are available.
func CompleteWorkflow(options *WorkflowOptions) (*WorkflowResult, error) {
	return nil, fmt.Errorf("CompleteWorkflow is not yet available in this build")
}

// TypeCheckWorkflow runs type checking on a Perl script.
// Stubbed out until type-system components are available.
func TypeCheckWorkflow(scriptPath string, perlVersion string, verbose bool) (*WorkflowResult, error) {
	return nil, fmt.Errorf("TypeCheckWorkflow is not yet available in this build")
}

// ExecutionWorkflow runs a Perl script without type checking.
// Stubbed out until type-system components are available.
func ExecutionWorkflow(scriptPath string, perlVersion string, verbose bool) (*WorkflowResult, error) {
	return nil, fmt.Errorf("ExecutionWorkflow is not yet available in this build")
}

// ValidationWorkflow validates a workspace.
// Stubbed out until type-system components are available.
func ValidationWorkflow(testScript string) (*WorkflowResult, error) {
	return nil, fmt.Errorf("ValidationWorkflow is not yet available in this build")
}
