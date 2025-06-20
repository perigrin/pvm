# PVM UI Framework Migration Guide

This guide provides step-by-step instructions for migrating existing PVM code to use the Fang UI Framework. It covers both new development and conversion of existing commands.

## Overview

The migration replaces all direct output calls (`fmt.Print*`, `cmd.Print*`) with styled UI framework methods that provide consistent, beautiful output across all PVM components.

## Migration Principles

1. **No Direct Output**: All user-facing output goes through the UI framework
2. **Consistent Styling**: All components use identical styling patterns
3. **Context Awareness**: Respect quiet/verbose modes and user preferences
4. **Error Separation**: Internal packages return structured errors, CLI formats them
5. **Backward Compatibility**: All existing functionality is preserved

## Pre-Migration Checklist

Before starting migration:

- [ ] Identify all `fmt.Print*` calls in the component
- [ ] Identify all `cmd.Print*` calls in the component
- [ ] Map current output patterns to UI framework methods
- [ ] Plan error handling improvements
- [ ] Prepare test cases for verification

## Step-by-Step Migration Process

### Step 1: Add UI Framework Access

**Before:**
```go
func myCommand(cmd *cobra.Command, args []string) error {
    fmt.Printf("Processing %d files\n", len(args))
    return nil
}
```

**After:**
```go
func myCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    ui.Info("Processing %d files", len(args))
    return nil
}
```

### Step 2: Replace Basic Output Calls

#### Success Messages

**Before:**
```go
fmt.Printf("✓ File processed successfully\n")
fmt.Println("Operation completed")
```

**After:**
```go
ui.Success("File processed successfully")
ui.Success("Operation completed")
```

#### Error Messages

**Before:**
```go
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
return fmt.Errorf("operation failed: %v", err)
```

**After:**
```go
ui.Error("Operation failed: %v", err)
return fmt.Errorf("operation failed: %v", err)
```

#### Warning Messages

**Before:**
```go
fmt.Printf("Warning: %s\n", message)
```

**After:**
```go
ui.Warning("%s", message)
```

#### Informational Messages

**Before:**
```go
fmt.Printf("Processing file: %s\n", filename)
```

**After:**
```go
ui.Info("Processing file: %s", filename)
```

### Step 3: Convert Structured Output

#### Simple Lists

**Before:**
```go
fmt.Println("Available versions:")
for _, version := range versions {
    fmt.Printf("  - %s\n", version)
}
```

**After:**
```go
ui.Header("Available versions")
ui.List(versions)
```

#### Tables

**Before:**
```go
fmt.Printf("%-10s %-8s %s\n", "Component", "Status", "Path")
fmt.Printf("%-10s %-8s %s\n", "--------", "------", "----")
for _, comp := range components {
    fmt.Printf("%-10s %-8s %s\n", comp.Name, comp.Status, comp.Path)
}
```

**After:**
```go
headers := []string{"Component", "Status", "Path"}
rows := make([][]string, len(components))
for i, comp := range components {
    rows[i] = []string{comp.Name, comp.Status, comp.Path}
}
ui.Table(headers, rows)
```

### Step 4: Enhance Error Handling

#### Internal Package Pattern

**Before:**
```go
// internal/mypackage/processor.go
func ProcessFile(path string) error {
    if _, err := os.Stat(path); err != nil {
        fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", path)
        return err
    }
    return nil
}
```

**After:**
```go
// internal/mypackage/processor.go
func ProcessFile(path string) error {
    if _, err := os.Stat(path); err != nil {
        return errors.NewUserInputError("PVM", "001",
            "File not found", err).WithLocation(path)
    }
    return nil
}
```

#### CLI Command Pattern

**Before:**
```go
// cmd/mycommand.go
func runCommand(cmd *cobra.Command, args []string) error {
    if err := ProcessFile(args[0]); err != nil {
        return err
    }
    fmt.Println("File processed successfully")
    return nil
}
```

**After:**
```go
// cmd/mycommand.go
func runCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    if err := ProcessFile(args[0]); err != nil {
        ui.Error("Processing failed: %v", err)
        return err
    }

    ui.Success("File processed successfully")
    return nil
}
```

### Step 5: Add Progress and Status Updates

**Before:**
```go
fmt.Println("Processing files...")
for _, file := range files {
    fmt.Printf("Processing: %s\n", file)
    processFile(file)
}
fmt.Println("Done")
```

**After:**
```go
ui.Info("Processing %d files", len(files))

for i, file := range files {
    ui.Progress(i+1, len(files), "Processing files")
    processFile(file)
}

ui.Success("Processing completed")
```

## Context-Aware Migration

### Respect Quiet Mode

**Before:**
```go
if verbose {
    fmt.Printf("Detailed information: %s\n", details)
}
fmt.Printf("Operation result: %s\n", result)
```

**After:**
```go
ui := cli.GetUI(cmd)

if ui.Context().Verbose {
    ui.Debug("Detailed information: %s", details)
}

// Important results shown regardless of quiet mode
ui.Success("Operation result: %s", result)
```

### Handle Verbose Mode

**Before:**
```go
fmt.Printf("Starting operation\n")
if verbose {
    fmt.Printf("Configuration: %+v\n", config)
    fmt.Printf("Environment: %s\n", env)
}
```

**After:**
```go
ui.Info("Starting operation")

if ui.Context().Verbose {
    ui.Debug("Configuration: %+v", config)
    ui.Debug("Environment: %s", env)
}
```

## Component-Specific Migration

### PVM Component Migration

**Before:**
```go
func listVersions(cmd *cobra.Command, args []string) error {
    versions, err := getVersions()
    if err != nil {
        return err
    }

    fmt.Println("Available Perl versions:")
    for _, v := range versions {
        status := ""
        if v.Current {
            status = " (current)"
        }
        fmt.Printf("  %s%s\n", v.Version, status)
    }
    return nil
}
```

**After:**
```go
func listVersions(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    versions, err := getVersions()
    if err != nil {
        ui.Error("Failed to list versions: %v", err)
        return err
    }

    if len(versions) == 0 {
        ui.Warning("No Perl versions installed")
        return nil
    }

    ui.Header("Available Perl Versions")

    headers := []string{"Version", "Status", "Path"}
    rows := make([][]string, len(versions))

    for i, v := range versions {
        status := "Available"
        if v.Current {
            status = "Current"
        }
        rows[i] = []string{v.Version, status, v.Path}
    }

    ui.Table(headers, rows)
    return nil
}
```

### PVX Component Migration

**Before:**
```go
func runScript(cmd *cobra.Command, args []string) error {
    scriptPath := args[0]

    fmt.Printf("Analyzing dependencies for %s\n", scriptPath)
    deps, err := analyzeDeps(scriptPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Dependency analysis failed: %v\n", err)
        return err
    }

    if len(deps.Missing) > 0 {
        fmt.Printf("Missing dependencies:\n")
        for _, dep := range deps.Missing {
            fmt.Printf("  - %s\n", dep)
        }
    }

    fmt.Printf("Executing script...\n")
    err = executeScript(scriptPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
        return err
    }

    fmt.Printf("Script completed successfully\n")
    return nil
}
```

**After:**
```go
func runScript(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    scriptPath := args[0]

    ui.Info("Analyzing dependencies for %s", scriptPath)

    deps, err := analyzeDeps(scriptPath)
    if err != nil {
        ui.Error("Dependency analysis failed: %v", err)
        return err
    }

    if len(deps.Missing) > 0 {
        ui.Warning("Missing dependencies detected:")
        ui.List(deps.Missing)
        ui.Info("Install with: pvi install %s", strings.Join(deps.Missing, " "))
    }

    ui.Status("Executing script...")

    err = executeScript(scriptPath)
    if err != nil {
        ui.Error("Execution failed: %v", err)
        return err
    }

    ui.Success("Script completed successfully")
    return nil
}
```

### PSC Component Migration

**Before:**
```go
func checkTypes(cmd *cobra.Command, args []string) error {
    filePath := args[0]

    fmt.Printf("Checking types in %s...\n", filePath)

    result, err := typeCheck(filePath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Type checking failed: %v\n", err)
        return err
    }

    if len(result.Errors) > 0 {
        fmt.Printf("Type errors found:\n")
        for _, e := range result.Errors {
            fmt.Printf("  Line %d: %s\n", e.Line, e.Message)
        }
    } else {
        fmt.Printf("✓ No type errors found\n")
    }

    return nil
}
```

**After:**
```go
func checkTypes(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    filePath := args[0]

    ui.Info("Type checking %s...", filePath)

    result, err := typeCheck(filePath)
    if err != nil {
        ui.Error("Type checking failed: %v", err)
        return err
    }

    if len(result.Errors) > 0 {
        ui.SubHeader("Type Errors")
        for _, e := range result.Errors {
            ui.Error("Line %d: %s", e.Line, e.Message)
        }
        ui.Warning("Found %d type errors", len(result.Errors))
    } else {
        ui.Success("No type errors found")
    }

    if ui.Context().Verbose && len(result.Annotations) > 0 {
        ui.SubHeader("Type Annotations")
        for _, ann := range result.Annotations {
            ui.Debug("%s: %s", ann.Variable, ann.Type)
        }
    }

    return nil
}
```

## Testing Migration

### Update Unit Tests

**Before:**
```go
func TestCommand(t *testing.T) {
    // Capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    err := runCommand(nil, []string{"test.pl"})
    assert.NoError(t, err)

    w.Close()
    os.Stdout = old

    output, _ := io.ReadAll(r)
    assert.Contains(t, string(output), "success")
}
```

**After:**
```go
func TestCommand(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever,
        Quiet:     false,
        Verbose:   false,
    }

    // Create mock command with UI context
    cmd := &cobra.Command{}
    cmd.SetContext(cli.WithUI(context.Background(), ui.NewOutput(ctx)))

    err := runCommand(cmd, []string{"test.pl"})
    assert.NoError(t, err)

    output := buf.String()
    assert.Contains(t, output, "✓")
    assert.Contains(t, output, "success")
}
```

### Update Integration Tests

**Before:**
```go
func TestIntegration(t *testing.T) {
    cmd := exec.Command("pvm", "help")
    output, err := cmd.Output()
    assert.NoError(t, err)
    assert.Contains(t, string(output), "Usage:")
}
```

**After:**
```go
func TestIntegration(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    stdout := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "help"}, "Help should work")

    // Verify styled output
    assert.Contains(t, stdout, "Usage:")

    // Verify UI framework is being used (contains styling indicators)
    hasStyleIndicators := strings.Contains(stdout, "ℹ") ||
                         strings.Contains(stdout, "Commands:") ||
                         len(strings.Split(stdout, "\n")) > 5
    assert.True(t, hasStyleIndicators, "Output should be styled")
}
```

## Common Migration Patterns

### Pattern 1: Simple Command Conversion

**Before:**
```go
func simpleCommand(cmd *cobra.Command, args []string) error {
    fmt.Printf("Starting operation\n")

    for _, arg := range args {
        fmt.Printf("Processing: %s\n", arg)
        if err := process(arg); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            return err
        }
    }

    fmt.Printf("Operation completed\n")
    return nil
}
```

**After:**
```go
func simpleCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    ui.Info("Starting operation")

    for _, arg := range args {
        ui.Status("Processing: %s", arg)
        if err := process(arg); err != nil {
            ui.Error("Processing failed: %v", err)
            return err
        }
    }

    ui.Success("Operation completed")
    return nil
}
```

### Pattern 2: List Command Conversion

**Before:**
```go
func listCommand(cmd *cobra.Command, args []string) error {
    items, err := getItems()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to get items: %v\n", err)
        return err
    }

    if len(items) == 0 {
        fmt.Printf("No items found\n")
        return nil
    }

    fmt.Printf("Found %d items:\n", len(items))
    for _, item := range items {
        fmt.Printf("  %s - %s\n", item.Name, item.Description)
    }

    return nil
}
```

**After:**
```go
func listCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    items, err := getItems()
    if err != nil {
        ui.Error("Failed to get items: %v", err)
        return err
    }

    if len(items) == 0 {
        ui.Warning("No items found")
        return nil
    }

    ui.Header(fmt.Sprintf("Found %d items", len(items)))

    headers := []string{"Name", "Description"}
    rows := make([][]string, len(items))
    for i, item := range items {
        rows[i] = []string{item.Name, item.Description}
    }

    ui.Table(headers, rows)
    return nil
}
```

### Pattern 3: Complex Command with Progress

**Before:**
```go
func complexCommand(cmd *cobra.Command, args []string) error {
    verbose, _ := cmd.Flags().GetBool("verbose")

    fmt.Printf("Starting complex operation\n")

    phases := []string{"initialize", "process", "finalize"}

    for i, phase := range phases {
        fmt.Printf("[%d/%d] %s\n", i+1, len(phases), phase)

        if err := executePhase(phase); err != nil {
            fmt.Fprintf(os.Stderr, "Phase %s failed: %v\n", phase, err)
            return err
        }

        if verbose {
            fmt.Printf("Phase %s completed successfully\n", phase)
        }
    }

    fmt.Printf("Operation completed successfully\n")
    return nil
}
```

**After:**
```go
func complexCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    ui.Header("Complex Operation")

    phases := []string{"initialize", "process", "finalize"}

    for i, phase := range phases {
        ui.Progress(i+1, len(phases), "Executing phases")
        ui.Status("Phase: %s", phase)

        if err := executePhase(phase); err != nil {
            ui.Error("Phase %s failed: %v", phase, err)
            return err
        }

        if ui.Context().Verbose {
            ui.Success("Phase %s completed", phase)
        }
    }

    ui.Success("Operation completed successfully")
    return nil
}
```

## Migration Validation

### Checklist for Completed Migration

- [ ] No `fmt.Print*` calls remain in CLI commands
- [ ] No `cmd.Print*` calls remain in CLI commands
- [ ] All error messages use `ui.Error()`
- [ ] All success messages use `ui.Success()`
- [ ] Warnings use `ui.Warning()`
- [ ] Informational output uses `ui.Info()`
- [ ] Progress indicators added for long operations
- [ ] Verbose mode provides additional detail
- [ ] Quiet mode respected appropriately
- [ ] Tables and lists use structured UI methods
- [ ] Tests updated to work with UI framework
- [ ] Integration tests verify styled output

### Quality Assurance

Run these commands to verify migration quality:

```bash
# Check for remaining direct output calls
grep -r "fmt\.Print" internal/
grep -r "cmd\.Print" cmd/

# Test all components with different modes
pvm --help
pvm --quiet --help
pvm --verbose --help
pvx --help
pvi --help
psc --help

# Test error handling
pvm nonexistent-command
pvx /nonexistent/file.pl
psc check /nonexistent/file.pl

# Run test suite
make test
```

### Performance Validation

```bash
# Test performance impact
time pvm --help
time pvx --help
time pvi --help
time psc --help

# Should complete in under 1 second for help commands
```

## Troubleshooting Migration Issues

### Common Problems and Solutions

**Problem**: Output not appearing
- **Cause**: Quiet mode enabled or UI context not set
- **Solution**: Verify `cli.GetUI(cmd)` is called and context is properly set

**Problem**: Colors not working
- **Cause**: Terminal detection or color mode setting
- **Solution**: Test with `--color=always` and verify terminal support

**Problem**: Tests failing after migration
- **Cause**: Tests expect specific output format
- **Solution**: Update tests to work with UI framework output

**Problem**: Performance regression
- **Cause**: Expensive operations in UI calls
- **Solution**: Move expensive operations outside UI framework calls

### Debug Migration Issues

Add debug output to verify UI framework usage:

```go
ui := cli.GetUI(cmd)
if ui.Context().Verbose {
    ui.Debug("UI framework active - ColorMode: %v, Quiet: %v",
        ui.Context().ColorMode, ui.Context().Quiet)
}
```

## Best Practices

1. **Always get UI reference**: Start every command with `ui := cli.GetUI(cmd)`
2. **Use appropriate methods**: Success for positive outcomes, Error for failures
3. **Respect context**: Check verbose/quiet modes before detailed output
4. **Structured data**: Use Table/List for structured information
5. **Progress feedback**: Show progress for operations taking >2 seconds
6. **Consistent messaging**: Use similar patterns across similar operations
7. **Test thoroughly**: Verify all output modes and error conditions

## Migration Timeline

For existing codebases, follow this timeline:

1. **Week 1**: Core framework setup and basic command migration
2. **Week 2**: Component-specific command migration
3. **Week 3**: Error handling enhancement and structured output
4. **Week 4**: Testing, validation, and polish

This migration guide ensures a smooth transition to the beautiful, consistent UI experience provided by the Fang UI Framework while maintaining all existing functionality.
