// ABOUTME: Core module management interfaces and types
// ABOUTME: Defines contracts for module operations across all PVM components

package modules

import (
	"context"
	"time"
)

// ModuleManager provides high-level module management operations
type ModuleManager interface {
	// List returns modules matching the given filter
	List(ctx context.Context, filter ModuleFilter) ([]*Module, error)

	// Install installs one or more modules with the given options
	Install(ctx context.Context, modules []string, opts InstallOptions) error

	// Remove uninstalls the specified modules
	Remove(ctx context.Context, modules []string) error

	// Update updates the specified modules to latest versions
	Update(ctx context.Context, modules []string) error
}

// ModuleInstaller provides module installation operations
type ModuleInstaller interface {
	// InstallModule installs a single module
	InstallModule(ctx context.Context, module string, opts InstallOptions) (*InstallResult, error)

	// InstallBatch installs multiple modules, potentially in parallel
	InstallBatch(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error)
}

// ProgressTracker provides progress reporting for operations
type ProgressTracker interface {
	// Start begins tracking an operation
	Start(operation string, total int)

	// Update reports progress on the current operation
	Update(current int, message string)

	// Finish completes the operation with final result
	Finish(result *OperationResult)
}

// ParallelProgressTracker provides progress tracking for parallel operations
type ParallelProgressTracker interface {
	// StartParallel begins tracking multiple parallel operations
	StartParallel(operations []string)

	// UpdateOperation updates the status of a specific operation
	UpdateOperation(id string, status OperationStatus, message string)

	// FinishParallel completes all parallel operations
	FinishParallel(results []*OperationResult)
}

// ProgressReporter provides callback-based progress reporting
type ProgressReporter interface {
	// Subscribe adds a progress callback
	Subscribe(callback ProgressCallback)

	// Unsubscribe removes a progress callback
	Unsubscribe(callback ProgressCallback)
}

// Module represents a Perl module with metadata
type Module struct {
	// Name is the module name (e.g., "DBI")
	Name string `json:"name"`

	// Version is the module version
	Version string `json:"version"`

	// Description is a short description of the module
	Description string `json:"description,omitempty"`

	// Author is the module author
	Author string `json:"author,omitempty"`

	// Path is the filesystem path to the module
	Path string `json:"path,omitempty"`

	// InstallationTime is when the module was installed
	InstallationTime time.Time `json:"installation_time,omitempty"`

	// CoreModule indicates if this is a Perl core module
	CoreModule bool `json:"core_module,omitempty"`

	// Dependencies lists module dependencies
	Dependencies []string `json:"dependencies,omitempty"`
}

// ModuleFilter defines criteria for filtering modules
type ModuleFilter struct {
	// Pattern filters modules by name pattern
	Pattern string

	// IncludeCore includes Perl core modules
	IncludeCore bool

	// IncludeDev includes development dependencies
	IncludeDev bool

	// Phase filters by dependency phase (runtime, build, test, develop)
	Phase string

	// LatestOnly returns only the latest version of each module
	LatestOnly bool
}

// ModuleQuery represents a module search query
type ModuleQuery struct {
	// Query is the search term
	Query string

	// Limit limits the number of results
	Limit int

	// Source specifies the metadata source to search
	Source string
}

// InstallOptions contains options for module installation
type InstallOptions struct {
	// PerlPath is the path to the Perl interpreter
	PerlPath string

	// InstallDir is the target installation directory
	InstallDir string

	// VersionConstraint specifies version requirements
	VersionConstraint string

	// Force installation even if tests fail
	Force bool

	// RunTests enables test execution during installation
	RunTests bool

	// SkipDependencies skips dependency installation
	SkipDependencies bool

	// Verbose enables detailed output
	Verbose bool

	// Cleanup removes build artifacts after installation
	Cleanup bool

	// Parallel enables parallel installation when applicable
	Parallel bool

	// Workers specifies the number of parallel workers
	Workers int
}

// InstallResult contains the result of a module installation
type InstallResult struct {
	// ModuleName is the name of the installed module
	ModuleName string `json:"module_name"`

	// Version is the installed version
	Version string `json:"version"`

	// Success indicates if installation was successful
	Success bool `json:"success"`

	// Duration is the time taken for installation
	Duration time.Duration `json:"duration"`

	// Dependencies lists installed dependencies
	Dependencies []string `json:"dependencies,omitempty"`

	// Warnings contains any installation warnings
	Warnings []string `json:"warnings,omitempty"`

	// Errors contains installation errors
	Errors []string `json:"errors,omitempty"`

	// Path is the installation path
	Path string `json:"path,omitempty"`
}

// OperationResult represents the result of any module operation
type OperationResult struct {
	// Operation is the type of operation performed
	Operation string `json:"operation"`

	// Target is the module or target of the operation
	Target string `json:"target"`

	// Success indicates if the operation was successful
	Success bool `json:"success"`

	// Duration is the time taken for the operation
	Duration time.Duration `json:"duration"`

	// Message provides additional information
	Message string `json:"message,omitempty"`

	// Error contains error information if unsuccessful
	Error error `json:"error,omitempty"`
}

// OperationStatus represents the status of an operation
type OperationStatus int

const (
	// StatusPending indicates the operation is pending
	StatusPending OperationStatus = iota

	// StatusRunning indicates the operation is in progress
	StatusRunning

	// StatusCompleted indicates the operation completed successfully
	StatusCompleted

	// StatusFailed indicates the operation failed
	StatusFailed

	// StatusCancelled indicates the operation was cancelled
	StatusCancelled
)

// String returns a string representation of the operation status
func (s OperationStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// ProgressCallback is called to report operation progress
type ProgressCallback func(operation string, current, total int, message string)

// OutdatedModule represents a module with available updates
type OutdatedModule struct {
	// Name is the module name
	Name string `json:"name"`

	// CurrentVersion is the currently installed version
	CurrentVersion string `json:"current_version"`

	// LatestVersion is the latest available version
	LatestVersion string `json:"latest_version"`

	// CoreModule indicates if this is a Perl core module
	CoreModule bool `json:"core_module,omitempty"`
}
