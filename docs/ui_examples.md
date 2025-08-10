# PVM UI Framework Examples

This document provides practical examples of using the PVM Fang UI Framework across different scenarios and components.

## Basic Output Examples

### Success Messages

```go
ui := cli.GetUI(cmd)

// Simple success
ui.Success("File processed successfully")

// Success with details
ui.Success("Installed %d dependencies in %v", count, duration)

// Success with file path
ui.Success("Created project at %s", projectPath)
```

**Output:**
```
✓ File processed successfully
✓ Installed 15 dependencies in 2.3s
✓ Created project at /home/user/myproject
```

### Error Messages

```go
ui := cli.GetUI(cmd)

// Simple error
ui.Error("Command failed")

// Error with context
ui.Error("Failed to parse file %s: %v", filename, err)

// Error with suggestion
ui.Error("Permission denied. Try running with sudo")
```

**Output:**
```
✗ Command failed
✗ Failed to parse file script.pl: syntax error at line 15
✗ Permission denied. Try running with sudo
```

### Warning Messages

```go
ui := cli.GetUI(cmd)

// Configuration warning
ui.Warning("Using default Perl version (system)")

// Deprecation warning
ui.Warning("Flag --old-flag is deprecated, use --new-flag instead")

// Performance warning
ui.Warning("Large file detected, processing may take longer")
```

**Output:**
```
⚠ Using default Perl version (system)
⚠ Flag --old-flag is deprecated, use --new-flag instead
⚠ Large file detected, processing may take longer
```

### Informational Messages

```go
ui := cli.GetUI(cmd)

// Status information
ui.Info("Scanning directory for Perl files...")

// Configuration information
ui.Info("Using Perl %s at %s", version, path)

// Process information
ui.Info("Found %d files to process", len(files))
```

**Output:**
```
ℹ Scanning directory for Perl files...
ℹ Using Perl v5.36.0 at /usr/local/bin/perl
ℹ Found 23 files to process
```

## Structured Output Examples

### Table Display

```go
ui := cli.GetUI(cmd)

// Component status table
headers := []string{"Component", "Version", "Status", "Location"}
rows := [][]string{
    {"PVM", "1.0.0", "Active", "/usr/local/bin/pvm"},
    {"PVX", "1.0.0", "Active", "/usr/local/bin/pvx"},
    {"PM", "1.0.0", "Active", "/usr/local/bin/pm"},
    {"PSC", "1.0.0", "Active", "/usr/local/bin/psc"},
}

ui.Header("PVM Components")
ui.Table(headers, rows)
```

**Output:**
```
PVM Components
══════════════

Component | Version | Status | Location
----------|---------|--------|------------------
PVM       | 1.0.0   | Active | /usr/local/bin/pvm
PVX       | 1.0.0   | Active | /usr/local/bin/pvx
PM       | 1.0.0   | Active | /usr/local/bin/pm
PSC       | 1.0.0   | Active | /usr/local/bin/psc
```

### List Display

```go
ui := cli.GetUI(cmd)

// Installation steps
steps := []string{
    "Download Perl distribution",
    "Extract archive",
    "Configure build options",
    "Compile source code",
    "Install binaries",
    "Update PATH",
}

ui.Header("Installation Steps")
ui.List(steps)
```

**Output:**
```
Installation Steps
══════════════════

• Download Perl distribution
• Extract archive
• Configure build options
• Compile source code
• Install binaries
• Update PATH
```

### Progress Indicators

```go
ui := cli.GetUI(cmd)

// File processing with progress
files := []string{"file1.pl", "file2.pl", "file3.pl"}

ui.Info("Processing %d files", len(files))
for i, file := range files {
    ui.Progress(i+1, len(files), "Processing files")
    processFile(file)
}
ui.Success("All files processed")
```

**Output:**
```
ℹ Processing 3 files
Processing files... [1/3] ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
Processing files... [2/3] ████████████████░░░░░░░░░░░░░░░░
Processing files... [3/3] ████████████████████████████████
✓ All files processed
```

## Component-Specific Examples

### PVM Component

```go
// pvm list command
func listPerlVersions(ui *ui.Output) error {
    ui.Header("Available Perl Versions")

    versions, err := getAvailableVersions()
    if err != nil {
        ui.Error("Failed to list versions: %v", err)
        return err
    }

    if len(versions) == 0 {
        ui.Warning("No Perl versions installed")
        ui.Info("Run 'pvm install <version>' to install a version")
        return nil
    }

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
    ui.Info("Use 'pvm use <version>' to switch versions")

    return nil
}
```

### PVX Component

```go
// pvx run command with dependency analysis
func runScript(ui *ui.Output, scriptPath string, args []string) error {
    ui.Info("Analyzing script dependencies...")

    deps, err := analyzeDependencies(scriptPath)
    if err != nil {
        ui.Error("Dependency analysis failed: %v", err)
        return err
    }

    if len(deps.Missing) > 0 {
        ui.Warning("Missing dependencies detected:")
        ui.List(deps.Missing)
        ui.Info("Install with: pm install %s", strings.Join(deps.Missing, " "))
    }

    ui.Status("Executing script...")

    start := time.Now()
    err = executeScript(scriptPath, args)
    duration := time.Since(start)

    if err != nil {
        ui.Error("Script execution failed: %v", err)
        return err
    }

    ui.Success("Script completed in %v", duration)
    return nil
}
```

### PM Component

```go
// pm install command with progress
func installModule(ui *ui.Output, moduleName string) error {
    ui.Header(fmt.Sprintf("Installing %s", moduleName))

    // Check if already installed
    if isInstalled(moduleName) {
        ui.Warning("Module %s is already installed", moduleName)
        ui.Info("Use 'pm upgrade %s' to upgrade", moduleName)
        return nil
    }

    // Download phase
    ui.Status("Downloading module...")
    if err := downloadModule(moduleName); err != nil {
        ui.Error("Download failed: %v", err)
        return err
    }

    // Dependency resolution
    ui.Status("Resolving dependencies...")
    deps, err := resolveDependencies(moduleName)
    if err != nil {
        ui.Error("Dependency resolution failed: %v", err)
        return err
    }

    if len(deps) > 0 {
        ui.Info("Installing %d dependencies:", len(deps))
        ui.List(deps)

        for i, dep := range deps {
            ui.Progress(i+1, len(deps), "Installing dependencies")
            if err := installDependency(dep); err != nil {
                ui.Error("Failed to install dependency %s: %v", dep, err)
                return err
            }
        }
    }

    // Installation phase
    ui.Status("Installing module...")
    if err := performInstall(moduleName); err != nil {
        ui.Error("Installation failed: %v", err)
        return err
    }

    ui.Success("Successfully installed %s", moduleName)
    return nil
}
```

### PSC Component

```go
// psc check command with detailed reporting
func checkTypes(ui *ui.Output, filePath string, strict bool) error {
    ui.Info("Type checking %s...", filePath)

    result, err := performTypeCheck(filePath)
    if err != nil {
        ui.Error("Type checking failed: %v", err)
        return err
    }

    // Show type annotations found
    if len(result.Annotations) > 0 {
        ui.SubHeader("Type Annotations")
        for _, ann := range result.Annotations {
            ui.Printf("  %s: %s (line %d)",
                ann.Variable, ann.Type, ann.Line)
        }
    }

    // Show type errors
    if len(result.Errors) > 0 {
        ui.SubHeader("Type Errors")
        for _, err := range result.Errors {
            ui.Error("Line %d: %s", err.Line, err.Message)
        }

        if strict {
            ui.Error("Type checking failed (%d errors)", len(result.Errors))
            return fmt.Errorf("type errors found")
        } else {
            ui.Warning("Found %d type errors (use --strict to fail)", len(result.Errors))
        }
    } else {
        ui.Success("No type errors found")
    }

    // Show inferred types if verbose
    if ui.Context().Verbose && len(result.Inferred) > 0 {
        ui.SubHeader("Inferred Types")
        for variable, inferredType := range result.Inferred {
            ui.Debug("%s: %s (inferred)", variable, inferredType)
        }
    }

    return nil
}
```

## Context-Aware Examples

### Quiet Mode Handling

```go
func processFiles(ui *ui.Output, files []string) error {
    // Always show important information
    ui.Info("Processing %d files", len(files))

    for i, file := range files {
        // Only show progress in non-quiet mode
        if !ui.Context().Quiet {
            ui.Progress(i+1, len(files), "Processing")
        }

        if err := processFile(file); err != nil {
            // Always show errors
            ui.Error("Failed to process %s: %v", file, err)
            continue
        }

        // Only show individual successes in verbose mode
        if ui.Context().Verbose {
            ui.Success("Processed %s", file)
        }
    }

    // Always show final result
    ui.Success("Completed processing %d files", len(files))
    return nil
}
```

### Verbose Mode Details

```go
func analyzeProject(ui *ui.Output, projectPath string) error {
    ui.Info("Analyzing project at %s", projectPath)

    // Basic analysis (always shown)
    files, err := findPerlFiles(projectPath)
    if err != nil {
        ui.Error("Failed to scan project: %v", err)
        return err
    }

    ui.Info("Found %d Perl files", len(files))

    // Detailed analysis (verbose only)
    if ui.Context().Verbose {
        ui.SubHeader("Project Structure")

        structure := analyzeStructure(files)
        for dir, count := range structure {
            ui.Debug("%s: %d files", dir, count)
        }

        ui.SubHeader("File Types")
        types := analyzeFileTypes(files)
        for ext, count := range types {
            ui.Debug("%s files: %d", ext, count)
        }

        ui.SubHeader("Complexity Analysis")
        complexity := analyzeComplexity(files)
        ui.Debug("Average lines per file: %.1f", complexity.AvgLines)
        ui.Debug("Total functions: %d", complexity.Functions)
        ui.Debug("Total packages: %d", complexity.Packages)
    }

    ui.Success("Project analysis complete")
    return nil
}
```

## Error Handling Examples

### Graceful Error Recovery

```go
func processMultipleFiles(ui *ui.Output, files []string) error {
    var failures []string
    successCount := 0

    ui.Info("Processing %d files", len(files))

    for _, file := range files {
        if err := processFile(file); err != nil {
            ui.Error("Failed to process %s: %v", file, err)
            failures = append(failures, file)
            continue
        }

        successCount++
        if ui.Context().Verbose {
            ui.Success("Processed %s", file)
        }
    }

    // Summary
    if len(failures) > 0 {
        ui.Warning("Processing completed with %d failures:", len(failures))
        ui.List(failures)
        ui.Info("Successfully processed: %d/%d files", successCount, len(files))
    } else {
        ui.Success("Successfully processed all %d files", len(files))
    }

    return nil
}
```

### User Input Validation

```go
func validateAndProcess(ui *ui.Output, args []string) error {
    if len(args) == 0 {
        ui.Error("No files specified")
        ui.Info("Usage: pvm command <file1> [file2...]")
        return fmt.Errorf("missing arguments")
    }

    var validFiles []string

    for _, arg := range args {
        if !strings.HasSuffix(arg, ".pl") && !strings.HasSuffix(arg, ".pm") {
            ui.Warning("Skipping non-Perl file: %s", arg)
            continue
        }

        if _, err := os.Stat(arg); err != nil {
            ui.Error("File not found: %s", arg)
            continue
        }

        validFiles = append(validFiles, arg)
    }

    if len(validFiles) == 0 {
        ui.Error("No valid Perl files found")
        return fmt.Errorf("no valid files")
    }

    ui.Info("Processing %d valid files", len(validFiles))
    return processFiles(ui, validFiles)
}
```

## Interactive Examples

### Confirmation Prompts

```go
func dangerousOperation(ui *ui.Output, target string) error {
    ui.Warning("This will permanently delete %s", target)

    if ui.Context().Interactive {
        ui.Printf("Are you sure? (y/N): ")

        var response string
        fmt.Scanln(&response)

        if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
            ui.Info("Operation cancelled")
            return nil
        }
    } else {
        ui.Info("Running in non-interactive mode, proceeding...")
    }

    ui.Status("Performing operation...")

    if err := performDeletion(target); err != nil {
        ui.Error("Operation failed: %v", err)
        return err
    }

    ui.Success("Operation completed successfully")
    return nil
}
```

### Status Updates

```go
func longRunningOperation(ui *ui.Output) error {
    ui.Header("Long Running Operation")

    phases := []struct {
        name string
        fn   func() error
    }{
        {"Initializing", initialize},
        {"Processing data", processData},
        {"Generating output", generateOutput},
        {"Cleaning up", cleanup},
    }

    for i, phase := range phases {
        ui.Status("Phase %d/%d: %s", i+1, len(phases), phase.name)

        start := time.Now()
        if err := phase.fn(); err != nil {
            ui.Error("Phase failed: %v", err)
            return err
        }
        duration := time.Since(start)

        if ui.Context().Verbose {
            ui.Success("Completed %s in %v", phase.name, duration)
        }
    }

    ui.Success("Operation completed successfully")
    return nil
}
```

## Testing Examples

### Unit Test with UI Capture

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

    // Test success case
    err := processFile(output, "test.pl")
    assert.NoError(t, err)

    result := buf.String()
    assert.Contains(t, result, "✓")
    assert.Contains(t, result, "test.pl")

    // Test error case
    buf.Reset()
    err = processFile(output, "nonexistent.pl")
    assert.Error(t, err)

    result = buf.String()
    assert.Contains(t, result, "✗")
    assert.Contains(t, result, "nonexistent.pl")
}
```

### Integration Test

```go
func TestFullWorkflow(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    // Create test files
    testFile := filepath.Join(env.RootDir, "test.pl")
    require.NoError(t, os.WriteFile(testFile, []byte("print 'hello';"), 0644))

    // Test command execution
    stdout := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "check", testFile},
        "Check command should succeed")

    // Verify styled output
    assert.Contains(t, stdout, "✓")
    assert.Contains(t, stdout, "test.pl")
    assert.NotContains(t, stdout, "Error")

    // Test verbose mode
    verboseOut := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "--verbose", "check", testFile},
        "Verbose check should succeed")

    // Verbose should contain more detail
    assert.Greater(t, len(verboseOut), len(stdout))
}
```

These examples demonstrate the full range of UI framework capabilities and show how to create consistent, beautiful CLI experiences across all PVM components.
