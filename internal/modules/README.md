# Core Module Management Interfaces

This package defines the core interfaces and types for module management operations across all PVM components. These interfaces provide a clean contract that enables consistent module operations while allowing for different implementations.

## Overview

The module management system is built around several key interfaces:

- **ModuleManager**: High-level module management operations (list, install, remove, update)
- **ModuleInstaller**: Focused module installation operations
- **ProgressTracker**: Progress reporting for operations
- **ParallelProgressTracker**: Progress tracking for parallel operations
- **ProgressReporter**: Callback-based progress reporting

## Key Types

### Module

Represents a Perl module with comprehensive metadata:

```go
type Module struct {
    Name             string    `json:"name"`
    Version          string    `json:"version"`
    Description      string    `json:"description,omitempty"`
    Author           string    `json:"author,omitempty"`
    Path             string    `json:"path,omitempty"`
    InstallationTime time.Time `json:"installation_time,omitempty"`
    CoreModule       bool      `json:"core_module,omitempty"`
    Dependencies     []string  `json:"dependencies,omitempty"`
}
```

### ModuleFilter

Defines criteria for filtering modules:

```go
type ModuleFilter struct {
    Pattern     string // Name pattern filter
    IncludeCore bool   // Include Perl core modules
    IncludeDev  bool   // Include development dependencies
    Phase       string // Dependency phase filter
    LatestOnly  bool   // Return only latest versions
}
```

### InstallOptions

Comprehensive installation configuration:

```go
type InstallOptions struct {
    PerlPath           string // Perl interpreter path
    InstallDir         string // Installation directory
    VersionConstraint  string // Version requirements
    Force              bool   // Force installation on test failures
    RunTests           bool   // Execute tests during installation
    SkipDependencies   bool   // Skip dependency installation
    Verbose            bool   // Detailed output
    Cleanup            bool   // Remove build artifacts
    Parallel           bool   // Enable parallel installation
    Workers            int    // Number of parallel workers
}
```

### InstallResult

Contains comprehensive installation results:

```go
type InstallResult struct {
    ModuleName   string        `json:"module_name"`
    Version      string        `json:"version"`
    Success      bool          `json:"success"`
    Duration     time.Duration `json:"duration"`
    Dependencies []string      `json:"dependencies,omitempty"`
    Warnings     []string      `json:"warnings,omitempty"`
    Errors       []string      `json:"errors,omitempty"`
    Path         string        `json:"path,omitempty"`
}
```

## Interface Usage

### ModuleManager

The primary interface for module management operations:

```go
// List modules with filtering
modules, err := manager.List(ctx, ModuleFilter{
    Pattern:     "DBI*",
    IncludeCore: false,
    LatestOnly:  true,
})

// Install multiple modules
err := manager.Install(ctx, []string{"DBI", "Moose"}, InstallOptions{
    RunTests: true,
    Verbose:  true,
})

// Update specific modules
err := manager.Update(ctx, []string{"DBI"})

// Remove modules
err := manager.Remove(ctx, []string{"Test::Module"})
```

### ModuleInstaller

Focused on installation operations with detailed results:

```go
// Install single module
result, err := installer.InstallModule(ctx, "DBI", InstallOptions{
    RunTests:          true,
    VersionConstraint: ">=1.640",
})

// Install multiple modules with detailed results
results, err := installer.InstallBatch(ctx, []string{"DBI", "Moose"}, InstallOptions{
    Parallel: true,
    Workers:  4,
})
```

### Progress Tracking

Track operation progress with different granularities:

```go
// Simple progress tracking
tracker.Start("Installing modules", 5)
tracker.Update(1, "Installing DBI")
tracker.Update(2, "Installing Moose")
tracker.Finish(&OperationResult{Success: true})

// Parallel progress tracking
parallelTracker.StartParallel([]string{"DBI", "Moose", "Dancer2"})
parallelTracker.UpdateOperation("DBI", StatusCompleted, "Installation successful")
parallelTracker.UpdateOperation("Moose", StatusRunning, "Running tests")
parallelTracker.FinishParallel(results)
```

## Design Principles

### 1. Interface Segregation

Interfaces are focused on specific responsibilities:
- `ModuleManager` for high-level operations
- `ModuleInstaller` for detailed installation operations
- `ProgressTracker` for operation monitoring

### 2. Context Support

All operations accept `context.Context` for:
- Cancellation support
- Timeout handling
- Request tracing

### 3. Comprehensive Error Handling

- Detailed error information in results
- Separation of warnings and errors
- Operation status tracking

### 4. JSON Serialization

All types support JSON marshaling for:
- Configuration persistence
- API responses
- Logging and debugging

### 5. Extensibility

- Interfaces can be extended with new methods
- Types include optional fields for future enhancements
- Plugin-style architecture support

## Implementation Guidelines

### For Module Managers

1. Implement all interface methods with proper error handling
2. Support context cancellation in long-running operations
3. Provide meaningful progress updates
4. Handle edge cases (missing modules, network failures, etc.)

### For Progress Trackers

1. Provide real-time updates without blocking operations
2. Support both simple and parallel tracking scenarios
3. Include timing information for performance analysis
4. Handle cancellation gracefully

### For Consumers

1. Always provide context with appropriate timeouts
2. Handle partial failures in batch operations
3. Check both success flags and error fields in results
4. Use appropriate progress tracking for user experience

## Future Enhancements

Planned additions to the interface:

1. **Dependency Resolution**: Enhanced dependency handling
2. **Cache Management**: Module cache control
3. **Bundle Operations**: Module bundle import/export
4. **Health Checking**: Module validation and health checks
5. **Metrics Collection**: Installation and usage metrics

## Testing

The package includes comprehensive tests demonstrating:

- Interface compliance verification
- JSON serialization/deserialization
- Mock implementations for testing
- Performance benchmarks

Run tests with:
```bash
go test ./internal/modules/
```

## Integration

This package is designed to be used by:

- **PVI**: Module installation and management
- **PVX**: Script dependency resolution
- **PSC**: Type definition module management
- **PVM**: Version-specific module operations

Each component can provide its own implementation while sharing the common interface definitions.
