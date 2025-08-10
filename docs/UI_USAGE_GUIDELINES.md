# PVM UI Framework Usage Guidelines

This document provides comprehensive guidelines for using the PVM Fang UI Framework effectively, including architectural patterns, best practices, and coding standards.

## Architecture Guidelines

### Separation of Concerns

The UI framework follows a strict separation between business logic and presentation:

#### Internal Packages (Business Logic)
- **Purpose**: Handle core functionality, data processing, and business rules
- **Output**: Return structured data, errors, and results
- **Dependencies**: No UI framework dependencies allowed
- **Testing**: Pure unit tests without UI concerns

```go
// CORRECT - Internal package returns structured data
package pvm

func GetInstalledVersions() ([]Version, error) {
    versions := make([]Version, 0)
    // Business logic here
    return versions, nil
}

type Version struct {
    Name      string
    Path      string
    IsCurrent bool
    IsSystem  bool
}
```

#### CLI Layer (Presentation)
- **Purpose**: Handle user interaction and output formatting
- **Input**: Get data from internal packages
- **Output**: Use UI framework for all display
- **Dependencies**: UI framework and cobra commands

```go
// CORRECT - CLI layer handles presentation
package cmd

func listVersionsCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    versions, err := pvm.GetInstalledVersions()
    if err != nil {
        ui.Error("Failed to get versions: %v", err)
        return err
    }

    // Format and display using UI framework
    ui.Header("Installed Perl Versions")
    // ... rest of presentation logic

    return nil
}
```

### Component Integration Pattern

Every PVM component follows the same integration pattern:

#### 1. UI Access Pattern
```go
func commandFunction(cmd *cobra.Command, args []string) error {
    // ALWAYS start with getting UI reference
    ui := cli.GetUI(cmd)

    // Use UI for all output
    ui.Info("Starting operation...")

    // Continue with business logic
    return nil
}
```

#### 2. Error Handling Pattern
```go
func commandWithErrorHandling(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    // Internal operation
    result, err := internal.DoSomething(args)
    if err != nil {
        // Format error using UI framework
        ui.Error("Operation failed: %v", err)
        return err
    }

    // Success feedback
    ui.Success("Operation completed successfully")
    return nil
}
```

#### 3. Context Awareness Pattern
```go
func contextAwareCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    // Basic operation info
    ui.Info("Processing %d items", len(args))

    // Verbose details
    if ui.Context().Verbose {
        ui.Debug("Detailed configuration: %+v", config)
    }

    // Process with appropriate feedback
    for i, arg := range args {
        if !ui.Context().Quiet {
            ui.Progress(i+1, len(args), "Processing")
        }

        // Process item
        if err := processItem(arg); err != nil {
            ui.Error("Failed to process %s: %v", arg, err)
            continue
        }

        if ui.Context().Verbose {
            ui.Success("Processed %s", arg)
        }
    }

    ui.Success("Completed processing")
    return nil
}
```

## Output Method Guidelines

### Choosing the Right Method

#### Success Messages
Use `ui.Success()` for:
- Successful completion of operations
- Positive outcomes
- Confirmation of actions taken

```go
ui.Success("File processed successfully")
ui.Success("Installed %d packages", count)
ui.Success("Configuration updated")
```

#### Error Messages
Use `ui.Error()` for:
- Operation failures
- Critical issues
- User input errors

```go
ui.Error("File not found: %s", filename)
ui.Error("Permission denied")
ui.Error("Invalid configuration: %v", err)
```

#### Warning Messages
Use `ui.Warning()` for:
- Non-critical issues
- Deprecation notices
- Potential problems

```go
ui.Warning("Using deprecated flag --old-flag")
ui.Warning("Large file detected, processing may be slow")
ui.Warning("Configuration file not found, using defaults")
```

#### Information Messages
Use `ui.Info()` for:
- General status updates
- Process information
- Configuration details

```go
ui.Info("Scanning directory for files...")
ui.Info("Using Perl version %s", version)
ui.Info("Found %d items to process", count)
```

#### Debug Messages
Use `ui.Debug()` for:
- Detailed diagnostic information
- Internal state details
- Development information

```go
ui.Debug("Internal state: %+v", state)
ui.Debug("Configuration loaded from %s", configPath)
ui.Debug("Processing with options: %v", options)
```

#### Status Messages
Use `ui.Status()` for:
- Ongoing operations
- Current activity
- Process updates

```go
ui.Status("Downloading package...")
ui.Status("Compiling source code...")
ui.Status("Installing dependencies...")
```

### Structured Output Guidelines

#### Tables
Use for tabular data with clear columns:

```go
// Good table usage
headers := []string{"Name", "Version", "Status", "Location"}
rows := make([][]string, len(items))
for i, item := range items {
    rows[i] = []string{item.Name, item.Version, item.Status, item.Location}
}
ui.Table(headers, rows)

// Avoid for single-column data (use lists instead)
// WRONG
headers := []string{"Files"}
rows := [][]string{{"file1.pl"}, {"file2.pl"}}
ui.Table(headers, rows)

// CORRECT
files := []string{"file1.pl", "file2.pl"}
ui.List(files)
```

#### Lists
Use for simple item enumerations:

```go
// Simple list
items := []string{"Initialize", "Process", "Finalize"}
ui.List(items)

// List with title
ui.Header("Setup Steps")
ui.List(items)

// Numbered list
ui.ListWithOptions(ui.ListOptions{
    Items:    items,
    Numbered: true,
    Title:    "Execution Order",
})
```

#### Progress Indicators
Use for long-running operations:

```go
// File processing
for i, file := range files {
    ui.Progress(i+1, len(files), "Processing files")
    processFile(file)
}

// Multi-phase operations
phases := []string{"download", "extract", "install"}
for i, phase := range phases {
    ui.Progress(i+1, len(phases), "Installation")
    executePhase(phase)
}
```

#### Key-Value Pairs
Use for configuration or metadata display:

```go
// System information
info := map[string]string{
    "Version":    "1.0.0",
    "Build Date": "2024-01-15",
    "Go Version": "1.21.0",
}
ui.KeyValue(info)

// With header
ui.SubHeader("Build Information")
ui.KeyValue(info)
```

## Context Handling Guidelines

### Quiet Mode Guidelines

#### What to Show in Quiet Mode
- Critical errors (always visible)
- Final results (success/failure)
- Essential warnings

#### What to Hide in Quiet Mode
- Progress indicators
- Status updates
- Informational messages
- Debug output

```go
// CORRECT - Respect quiet mode
func quietModeExample(ui *ui.Output) {
    // Always show critical errors
    if criticalError {
        ui.Error("Critical system error")
    }

    // Respect quiet mode for progress
    if !ui.Context().Quiet {
        ui.Progress(current, total, "Processing")
    }

    // Always show final results
    ui.Success("Operation completed")
}
```

### Verbose Mode Guidelines

#### Additional Information in Verbose Mode
- Detailed configuration
- Internal state information
- Processing details
- Performance metrics

```go
// CORRECT - Add detail in verbose mode
func verboseModeExample(ui *ui.Output) {
    ui.Info("Starting operation")

    if ui.Context().Verbose {
        ui.Debug("Configuration: %+v", config)
        ui.Debug("Environment: %s", env)
        ui.Debug("Memory usage: %d MB", memoryMB)
    }

    // ... process ...

    if ui.Context().Verbose {
        ui.Debug("Processing completed in %v", duration)
        ui.Debug("Files processed: %d", fileCount)
    }

    ui.Success("Operation completed")
}
```

### Color Mode Guidelines

#### Design for All Color Modes
- Use symbols and formatting beyond just color
- Ensure information is clear without colors
- Test with `--color=never` flag

```go
// GOOD - Clear without colors
ui.Success("✓ Operation completed")  // Symbol + message
ui.Error("✗ Operation failed")      // Symbol + message
ui.Warning("⚠ Potential issue")     // Symbol + message

// AVOID - Color-only differentiation
ui.Printf(colorGreen + "Success" + colorReset)
ui.Printf(colorRed + "Error" + colorReset)
```

## Performance Guidelines

### Efficient Output Patterns

#### Batch Operations
```go
// WRONG - Individual UI calls in loop
for _, item := range manyItems {
    ui.Info("Processing %s", item)
    processItem(item)
}

// CORRECT - Batched progress updates
ui.Info("Processing %d items", len(manyItems))
for i, item := range manyItems {
    if i%100 == 0 || i == len(manyItems)-1 {
        ui.Progress(i+1, len(manyItems), "Processing")
    }
    processItem(item)
}
ui.Success("All items processed")
```

#### Lazy Evaluation
```go
// WRONG - Always compute expensive debug info
debugInfo := generateExpensiveDebugInfo()
ui.Debug("Debug info: %s", debugInfo)

// CORRECT - Only compute when needed
if ui.Context().Verbose {
    debugInfo := generateExpensiveDebugInfo()
    ui.Debug("Debug info: %s", debugInfo)
}
```

#### String Building
```go
// WRONG - String concatenation
var output string
for _, item := range items {
    output += fmt.Sprintf("  %s\n", item)
}
ui.Printf(output)

// CORRECT - String builder
var builder strings.Builder
for _, item := range items {
    builder.WriteString(fmt.Sprintf("  %s\n", item))
}
ui.Printf(builder.String())
```

## Testing Guidelines

### Unit Testing UI Output

#### Basic Output Testing
```go
func TestCommandOutput(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever,
        Quiet:     false,
        Verbose:   false,
    }
    output := ui.NewOutput(ctx)

    // Test the operation
    err := performOperation(output, "test-input")
    assert.NoError(t, err)

    // Verify output
    result := buf.String()
    assert.Contains(t, result, "✓")
    assert.Contains(t, result, "test-input")
}
```

#### Context Testing
```go
func TestVerboseMode(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever,
        Quiet:     false,
        Verbose:   true,
    }
    output := ui.NewOutput(ctx)

    performOperation(output, "test")

    result := buf.String()
    assert.Contains(t, result, "Debug:")
    assert.Greater(t, strings.Count(result, "\n"), 5)
}

func TestQuietMode(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever,
        Quiet:     true,
        Verbose:   false,
    }
    output := ui.NewOutput(ctx)

    performOperation(output, "test")

    result := buf.String()
    assert.NotContains(t, result, "Processing")
    assert.Contains(t, result, "✓")  // Final result should show
}
```

### Integration Testing

#### Command Integration
```go
func TestCommandIntegration(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    // Test normal execution
    stdout := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "command", "arg"},
        "Command should succeed")

    assert.Contains(t, stdout, "expected output")

    // Test verbose mode
    verboseOut := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "--verbose", "command", "arg"},
        "Verbose command should succeed")

    assert.Greater(t, len(verboseOut), len(stdout))
    assert.Contains(t, verboseOut, "Debug:")
}
```

## Error Handling Guidelines

### Structured Error Handling

#### Internal Package Errors
```go
// Internal package - return structured errors
func InternalOperation(path string) error {
    if _, err := os.Stat(path); err != nil {
        return errors.NewUserInputError("PVM", "001",
            "File not found", err).WithLocation(path)
    }
    return nil
}
```

#### CLI Layer Error Display
```go
// CLI layer - format errors for display
func cliCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    for _, arg := range args {
        if err := InternalOperation(arg); err != nil {
            // Use UI framework for error display
            ui.Error("Failed to process %s: %v", arg, err)

            // Provide helpful suggestions
            if errors.IsUserInputError(err) {
                ui.Info("Check that the file exists and is readable")
            }

            continue
        }

        ui.Success("Processed %s", arg)
    }

    return nil
}
```

### Error Recovery Patterns

#### Graceful Degradation
```go
func robustOperation(ui *ui.Output) error {
    results, failures := processAllItems(items)

    if len(failures) > 0 {
        ui.Warning("Operation completed with %d failures:", len(failures))
        for _, failure := range failures {
            ui.Error("Failed: %s - %v", failure.Item, failure.Error)
        }
        ui.Info("Successfully processed: %d/%d items",
            len(results), len(items))
    } else {
        ui.Success("Successfully processed all %d items", len(results))
    }

    return nil
}
```

## Component-Specific Guidelines

### PVM Component Patterns

#### Version Management
```go
func pvmListCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    ui.Header("Perl Version Manager")

    versions, err := pvm.GetVersions()
    if err != nil {
        ui.Error("Failed to list versions: %v", err)
        return err
    }

    if len(versions) == 0 {
        ui.Warning("No Perl versions installed")
        ui.Info("Install with: pvm install <version>")
        return nil
    }

    // Display as table
    headers := []string{"Version", "Status", "Path"}
    rows := make([][]string, len(versions))
    for i, v := range versions {
        status := "Available"
        if v.IsCurrent {
            status = "Current"
        }
        rows[i] = []string{v.Version, status, v.Path}
    }

    ui.Table(headers, rows)
    ui.Info("Use 'pvm use <version>' to switch versions")

    return nil
}
```

### PVX Component Patterns

#### Script Execution
```go
func pvxRunCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    scriptPath := args[0]

    ui.Info("Executing script: %s", scriptPath)

    // Dependency analysis with progress
    ui.Status("Analyzing dependencies...")
    deps, err := pvx.AnalyzeDependencies(scriptPath)
    if err != nil {
        ui.Error("Dependency analysis failed: %v", err)
        return err
    }

    if len(deps.Missing) > 0 {
        ui.Warning("Missing dependencies:")
        ui.List(deps.Missing)
        ui.Info("Install with: pm install %s",
            strings.Join(deps.Missing, " "))
    }

    // Execution with timing
    ui.Status("Executing script...")
    start := time.Now()

    result, err := pvx.Execute(scriptPath, args[1:])
    duration := time.Since(start)

    if err != nil {
        ui.Error("Execution failed: %v", err)
        return err
    }

    ui.Success("Script completed in %v", duration)

    if ui.Context().Verbose {
        ui.SubHeader("Execution Details")
        ui.KeyValue(map[string]string{
            "Exit Code": fmt.Sprintf("%d", result.ExitCode),
            "Duration":  duration.String(),
            "Memory":    fmt.Sprintf("%.1f MB", result.MemoryMB),
        })
    }

    return nil
}
```

### PM Component Patterns

#### Package Installation
```go
func pmInstallCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    packageName := args[0]

    ui.Header(fmt.Sprintf("Installing %s", packageName))

    // Check existing installation
    if pm.IsInstalled(packageName) {
        ui.Warning("Package %s is already installed", packageName)
        ui.Info("Use 'pm upgrade %s' to upgrade", packageName)
        return nil
    }

    // Multi-phase installation with progress
    phases := []struct {
        name string
        fn   func() error
    }{
        {"Downloading", func() error { return pm.Download(packageName) }},
        {"Resolving dependencies", func() error { return pm.ResolveDeps(packageName) }},
        {"Installing", func() error { return pm.Install(packageName) }},
    }

    for i, phase := range phases {
        ui.Progress(i+1, len(phases), "Installation")
        ui.Status(phase.name + "...")

        if err := phase.fn(); err != nil {
            ui.Error("%s failed: %v", phase.name, err)
            return err
        }
    }

    ui.Success("Successfully installed %s", packageName)
    return nil
}
```

### PSC Component Patterns

#### Type Checking
```go
func pscCheckCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    filePath := args[0]

    ui.Info("Type checking %s", filePath)

    result, err := psc.CheckTypes(filePath)
    if err != nil {
        ui.Error("Type checking failed: %v", err)
        return err
    }

    // Display results with appropriate formatting
    if len(result.Errors) > 0 {
        ui.SubHeader("Type Errors")
        for _, e := range result.Errors {
            ui.Error("Line %d: %s", e.Line, e.Message)
        }
        ui.Warning("Found %d type errors", len(result.Errors))
    } else {
        ui.Success("No type errors found")
    }

    // Verbose mode details
    if ui.Context().Verbose && len(result.Annotations) > 0 {
        ui.SubHeader("Type Annotations")
        for _, ann := range result.Annotations {
            ui.Debug("%s: %s (line %d)",
                ann.Variable, ann.Type, ann.Line)
        }
    }

    return nil
}
```

## Migration Guidelines

### Converting Existing Commands

#### Step-by-Step Conversion
1. **Add UI Access**: Start command with `ui := cli.GetUI(cmd)`
2. **Replace Direct Output**: Convert all `fmt.Print*` to UI methods
3. **Update Error Handling**: Use `ui.Error()` for error display
4. **Add Context Awareness**: Respect quiet/verbose modes
5. **Enhance with Structure**: Use tables/lists where appropriate
6. **Update Tests**: Modify tests to work with UI framework

#### Before and After Example
```go
// BEFORE - Direct output
func oldCommand(cmd *cobra.Command, args []string) error {
    fmt.Printf("Processing %d files\n", len(args))

    for _, file := range args {
        fmt.Printf("Processing: %s\n", file)
        if err := process(file); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            return err
        }
    }

    fmt.Println("All files processed")
    return nil
}

// AFTER - UI framework
func newCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    ui.Info("Processing %d files", len(args))

    for i, file := range args {
        ui.Progress(i+1, len(args), "Processing files")
        if err := process(file); err != nil {
            ui.Error("Failed to process %s: %v", file, err)
            return err
        }

        if ui.Context().Verbose {
            ui.Success("Processed %s", file)
        }
    }

    ui.Success("All files processed successfully")
    return nil
}
```

These guidelines ensure consistent, maintainable, and user-friendly CLI experiences across all PVM components while maintaining clean architecture and excellent performance.
