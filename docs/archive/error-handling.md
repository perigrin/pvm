# Error Handling and Logging Guide

This document provides guidelines for using the error handling and logging framework in the PVM Ecosystem.

## Error Categories

The system defines the following error categories:

1. **Configuration Errors** (`CategoryConfig`): Issues with configuration files or settings
   - Example: Missing configuration file, invalid configuration value

2. **Version Errors** (`CategoryVersion`): Problems with Perl version detection, resolution, or installation
   - Example: Version not found, version format invalid

3. **Module Errors** (`CategoryModule`): Issues with CPAN modules installation or dependencies
   - Example: Module not found, dependency conflict

4. **Execution Errors** (`CategoryExecution`): Problems during script or command execution
   - Example: Script execution failure, command not found

5. **Type Errors** (`CategoryType`): Type checking failures or inconsistencies
   - Example: Type mismatch, missing type definition

6. **System Errors** (`CategorySystem`): Issues with the underlying operating system or environment
   - Example: File system error, permission denied

7. **User Input Errors** (`CategoryUserInput`): Problems with command-line arguments or inputs
   - Example: Invalid flag, missing required argument

## Error Components

Error codes are prefixed with the component that generated them:

- `PVM-`: For Perl Version Manager errors
- `PVX-`: For Perl Version eXecutor errors
- `PM-`: For Perl Version Installer errors
- `PSC-`: For Perl Script Compiler errors
- `CFG-`: For configuration errors
- `SYS-`: For system-level errors

## Error Format

All errors follow a consistent structured format:

```
PVM-001: Unable to detect Perl version (Version Error)
  Detail: Version string '5.x.y' does not match expected format
  Location: version.go:123
  Hint: Use a valid version format like '5.32.1'
```

The format includes:
- Error code with component prefix
- Short error message
- Error category
- Detailed explanation (optional)
- Location of the error (file, line, etc.) (optional)
- Hint for resolving the error (optional)

## Creating Errors

To create a new error, use the error helper functions:

```go
// Generic error creation
err := errors.New(PrefixPVM, CategoryVersion, "001", "Failed to detect version", innerErr)

// Category-specific helpers
err := errors.NewVersionError("001", "Failed to detect version", innerErr)
err := errors.NewConfigError("001", "Invalid configuration", innerErr)
err := errors.NewModuleError("001", "Module not found", innerErr)
// etc.
```

You can add additional context to errors:

```go
err := errors.NewVersionError("001", "Failed to detect version", innerErr)
err.WithDetail("Version string does not match expected format")
   .WithLocation("version.go:123")
   .WithHint("Use a valid version format like '5.32.1'")
```

## Error Handling

When handling errors, you should:

1. Log the error at the appropriate level
2. Consider whether to return the error or handle it locally
3. Add context if needed before returning the error

Example:

```go
func resolveVersion(version string) (string, error) {
    result, err := doSomething()
    if err != nil {
        // Add context to the error
        return "", errors.Wrap(err, PrefixPVM, CategoryVersion, "002",
            "Failed to resolve version alias")
    }
    return result, nil
}
```

## Logging

The logging system provides several levels of logging:

1. **Debug** (`LevelDebug`): Detailed information, useful for debugging
2. **Info** (`LevelInfo`): General information about system operation
3. **Warning** (`LevelWarning`): Potential issues that don't prevent operation
4. **Error** (`LevelError`): Errors that prevent a specific operation
5. **Fatal** (`LevelFatal`): Critical errors that prevent system operation

When to use each level:

- **Debug**: Use for detailed information that is useful during development or debugging
- **Info**: Use for general information about system operation (starting up, shutting down, etc.)
- **Warning**: Use for potential issues that don't prevent operation but might need attention
- **Error**: Use for errors that prevent a specific operation but not the entire system
- **Fatal**: Use for critical errors that prevent the system from functioning (will exit the program)

Example:

```go
// In CLI package
cli.LogDebug("Processing argument: %s", arg)
cli.LogInfo("Starting version resolution")
cli.LogWarning("Configuration file not found, using defaults")
cli.LogError("Failed to execute command: %v", err)

// Or using the log package directly
log.Debugf("Processing argument: %s", arg)
log.Infof("Starting version resolution")
log.Warningf("Configuration file not found, using defaults")
log.Errorf("Failed to execute command: %v", err)
```

## Error and Logging Integration

The two systems are integrated through the `errors.LogError` function, which logs an error at the appropriate level based on its category:

```go
// Log an error at the appropriate level
errors.LogError(err)

// Log an error with location information
errors.LogErrorWithLocation(err, "version.go:123")

// Log an error at debug level
errors.LogDebug(err)

// Log a fatal error and exit
errors.LogFatal(err)
```

## Guidelines for Error Messages

1. **Be specific**: Describe what went wrong clearly
2. **Be concise**: Keep messages short but informative
3. **Be actionable**: Include hints for how to resolve the error
4. **Be consistent**: Use similar wording for similar errors
5. **Include context**: Add details that help locate and understand the error

Good error message: "Failed to install module 'Test::More': Module not found in CPAN"
Bad error message: "Installation error occurred"
