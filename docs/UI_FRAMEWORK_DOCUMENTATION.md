# PVM Fang UI Framework - Complete Documentation

This comprehensive documentation covers the Fang UI integration across all PVM components, providing complete API reference, usage guidelines, and examples for future development.

## Table of Contents

1. [UI Framework Architecture Overview](#ui-framework-architecture-overview)
2. [API Reference for internal/cli/ui](#api-reference-for-internal-cli-ui)
3. [Styling Guidelines and Patterns](#styling-guidelines-and-patterns)
4. [Integration Examples and Best Practices](#integration-examples-and-best-practices)
5. [Troubleshooting and Common Issues](#troubleshooting-and-common-issues)
6. [Future Enhancement Guidelines](#future-enhancement-guidelines)

---

## UI Framework Architecture Overview

### Design Philosophy

The PVM Fang UI Framework replaces all direct output calls (`fmt.Print*`, `cmd.Print*`) with beautiful, consistent Fang-powered styling. The architecture follows these principles:

1. **Separation of Concerns**: Internal packages return structured data, CLI layer handles formatting
2. **Global Consistency**: All components (pvm, pvx, pvi, psc) use identical styling patterns
3. **Context Awareness**: UI adapts to quiet/verbose modes, color preferences, and terminal capabilities
4. **Performance First**: Minimal overhead with efficient rendering
5. **Clean Architecture**: No UI dependencies in internal business logic packages

### Component Architecture

```
PVM CLI Ecosystem
├── internal/cli/ui/           # Core UI Framework
│   ├── output.go             # Main output interface and methods
│   ├── styles.go             # Fang styling definitions and themes
│   ├── types.go              # Type definitions and interfaces
│   ├── output_test.go        # Comprehensive test suite
│   └── styles_test.go        # Style validation tests
├── internal/cli/root.go       # CLI integration and UI context injection
├── internal/cli/help.go       # Enhanced help system with Fang styling
└── Components Integration
    ├── internal/pvm/          # PVM commands using UI framework
    ├── internal/pvx/          # PVX commands using UI framework
    ├── internal/pvi/          # PVI commands using UI framework
    └── internal/psc/          # PSC commands using UI framework
```

### Integration Flow

1. **Command Execution**: Cobra command receives user input
2. **UI Context Setup**: `cli.setupUI()` creates UI context based on flags and environment
3. **Command Logic**: Commands get UI instance via `cli.GetUI(cmd)`
4. **Styled Output**: All output flows through UI framework methods
5. **Context Awareness**: Output respects quiet/verbose modes and color preferences

### Key Dependencies

- **github.com/charmbracelet/fang**: Core styling engine
- **github.com/charmbracelet/lipgloss/v2**: Style definitions and rendering
- **github.com/spf13/cobra**: CLI framework integration

---

## API Reference for internal/cli/ui

### Core Types

#### UIContext
```go
type UIContext struct {
    Writer      io.Writer  // Output destination (usually os.Stdout)
    ErrorWriter io.Writer  // Error output destination (usually os.Stderr)
    ColorMode   ColorMode  // Color output mode (Auto/Always/Never)
    Quiet       bool       // Suppress non-essential output
    Verbose     bool       // Show detailed information
    Interactive bool       // Enable interactive features
}
```

#### ColorMode
```go
type ColorMode int

const (
    ColorAuto   ColorMode = iota // Auto-detect terminal capabilities
    ColorAlways                  // Always use colors
    ColorNever                   // Never use colors
)
```

#### OutputLevel
```go
type OutputLevel int

const (
    LevelDebug   OutputLevel = iota
    LevelInfo
    LevelSuccess
    LevelWarning
    LevelError
)
```

### Output Interface

#### Basic Output Methods
```go
type Output struct {
    context *UIContext
    styles  Styles
}

// Core output methods with automatic styling
func (o *Output) Success(message string, args ...interface{})
func (o *Output) Error(message string, args ...interface{})
func (o *Output) Warning(message string, args ...interface{})
func (o *Output) Info(message string, args ...interface{})
func (o *Output) Debug(message string, args ...interface{})

// Formatted output (respects quiet mode)
func (o *Output) Printf(format string, args ...interface{})
func (o *Output) Println(args ...interface{})
```

#### Structured Output Methods
```go
// Headers and sections
func (o *Output) Header(title string)
func (o *Output) SubHeader(title string)
func (o *Output) Section(title, content string)

// Tabular data display
func (o *Output) Table(headers []string, rows [][]string)
func (o *Output) TableWithOptions(opts TableOptions)

// List data display
func (o *Output) List(items []string)
func (o *Output) ListWithOptions(opts ListOptions)

// Key-value pairs
func (o *Output) KeyValue(pairs map[string]string)

// Progress and status
func (o *Output) Status(message string)
func (o *Output) Progress(current, total int, message string)
```

#### Advanced Output Methods
```go
// Boxed content
func (o *Output) Box(content string)

// Basic markdown rendering
func (o *Output) Markdown(content string)
```

#### Context Management
```go
// Context access and modification
func (o *Output) Context() *UIContext
func (o *Output) SetWriter(w io.Writer)
func (o *Output) SetQuiet(quiet bool)
func (o *Output) SetVerbose(verbose bool)
func (o *Output) SetColorMode(mode ColorMode)
```

### Configuration Options

#### TableOptions
```go
type TableOptions struct {
    Headers     []string
    Rows        [][]string
    Title       string
    ShowBorders bool
    Compact     bool
}
```

#### ListOptions
```go
type ListOptions struct {
    Items      []string
    Title      string
    Numbered   bool
    BulletChar string
}
```

#### ProgressOptions
```go
type ProgressOptions struct {
    Current int
    Total   int
    Message string
    Width   int
}
```

### Factory Functions

#### Constructors
```go
// Create new output instance with custom context
func NewOutput(ctx *UIContext) *Output

// Create output instance with default settings
func NewDefaultOutput() *Output
```

#### CLI Integration
```go
// Get UI instance for command (creates if needed)
func GetUI(cmd *cobra.Command) *Output

// Setup UI context for command execution
func setupUI(cmd *cobra.Command)
```

---

## Styling Guidelines and Patterns

### Color Scheme

#### Default Theme
```go
var DefaultTheme = Theme{
    Primary:   "#7C3AED", // Purple - primary actions and headers
    Secondary: "#3B82F6", // Blue - secondary information
    Accent:    "#10B981", // Green - success states

    Success: "#10B981", // Green - successful operations
    Warning: "#F59E0B", // Amber - warnings and cautions
    Error:   "#EF4444", // Red - errors and failures
    Info:    "#3B82F6", // Blue - informational messages
    Debug:   "#6B7280", // Gray - debug output

    Border:    "#E5E7EB", // Light gray - borders and separators
    Highlight: "#F3F4F6", // Very light gray - backgrounds
    Muted:     "#9CA3AF", // Medium gray - secondary text
}
```

### Message Types and Icons

#### Status Messages
- **Success**: Green checkmark (✓) with success message
- **Error**: Red X (✗) with error details
- **Warning**: Yellow warning (⚠) with caution message
- **Info**: Blue info (ℹ) with informational text
- **Debug**: Gray debug symbol (🐛) with detailed output (verbose mode only)
- **Status**: Blue arrow (→) for ongoing operations
- **Progress**: Lightning bolt (⚡) with progress indicators

#### Typography Hierarchy
- **Header**: Large, bold text with primary color and padding
- **SubHeader**: Medium text with secondary color and bold weight
- **Body Text**: Standard text with default styling
- **Code**: Monospace text with accent color and background
- **Muted**: Secondary information in gray

#### Interactive Elements
- **Buttons**: Rounded borders with primary background
- **Links**: Primary color with underline
- **Tables**: Headers in primary color, aligned columns with proper spacing
- **Lists**: Colored bullets with consistent indentation

### Style Application Rules

1. **Consistency**: All similar operations use identical styling patterns
2. **Hierarchy**: Visual importance matches functional importance
3. **Accessibility**: Minimum contrast ratios met for all color combinations
4. **Context**: Styles adapt based on quiet/verbose modes
5. **Terminal Support**: Graceful degradation for limited terminals

---

## Integration Examples and Best Practices

### Basic Command Integration

#### Simple Command Pattern
```go
func myCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    // Show what we're doing
    ui.Info("Starting operation with %d arguments", len(args))

    // Process each argument
    for _, arg := range args {
        ui.Status("Processing: %s", arg)

        if err := processArgument(arg); err != nil {
            ui.Error("Failed to process %s: %v", arg, err)
            continue // Continue with other arguments
        }

        ui.Success("Processed: %s", arg)
    }

    ui.Success("Operation completed successfully")
    return nil
}
```

#### Command with Structured Output
```go
func listCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    items, err := getItems()
    if err != nil {
        ui.Error("Failed to retrieve items: %v", err)
        return err
    }

    if len(items) == 0 {
        ui.Warning("No items found")
        ui.Info("Use 'command create' to add items")
        return nil
    }

    ui.Header(fmt.Sprintf("Found %d items", len(items)))

    headers := []string{"Name", "Type", "Status", "Created"}
    rows := make([][]string, len(items))

    for i, item := range items {
        rows[i] = []string{
            item.Name,
            item.Type,
            item.Status,
            item.Created.Format("2006-01-02"),
        }
    }

    ui.Table(headers, rows)

    if ui.Context().Verbose {
        ui.SubHeader("Summary")
        ui.KeyValue(map[string]string{
            "Total Items":   fmt.Sprintf("%d", len(items)),
            "Active Items":  fmt.Sprintf("%d", countActive(items)),
            "Last Updated":  time.Now().Format("2006-01-02 15:04:05"),
        })
    }

    return nil
}
```

### Error Handling Patterns

#### Graceful Error Recovery
```go
func batchProcess(ui *ui.Output, files []string) error {
    var failures []string
    successCount := 0

    ui.Info("Processing %d files", len(files))

    for i, file := range files {
        ui.Progress(i+1, len(files), "Processing files")

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

    // Summary with appropriate message type
    if len(failures) > 0 {
        ui.Warning("Processing completed with errors:")
        ui.List(failures)
        ui.Info("Successfully processed: %d/%d files", successCount, len(files))
    } else {
        ui.Success("Successfully processed all %d files", successCount)
    }

    return nil
}
```

#### User Input Validation
```go
func validateInput(ui *ui.Output, args []string) ([]string, error) {
    if len(args) == 0 {
        ui.Error("No arguments provided")
        ui.Info("Usage: command <arg1> [arg2...]")
        ui.Info("Use 'command --help' for more information")
        return nil, fmt.Errorf("missing arguments")
    }

    var validArgs []string

    for _, arg := range args {
        if !isValid(arg) {
            ui.Warning("Skipping invalid argument: %s", arg)
            continue
        }

        validArgs = append(validArgs, arg)
    }

    if len(validArgs) == 0 {
        ui.Error("No valid arguments found")
        ui.Info("Valid arguments must match pattern: %s", getValidPattern())
        return nil, fmt.Errorf("no valid arguments")
    }

    ui.Info("Processing %d valid arguments", len(validArgs))
    return validArgs, nil
}
```

### Context-Aware Output

#### Verbose Mode Handling
```go
func contextAwareOperation(ui *ui.Output) error {
    ui.Info("Starting complex operation")

    // Basic information (always shown)
    config := loadConfiguration()
    ui.Info("Using configuration: %s", config.Name)

    // Detailed information (verbose only)
    if ui.Context().Verbose {
        ui.SubHeader("Configuration Details")
        ui.KeyValue(map[string]string{
            "Config File":    config.Path,
            "Last Modified": config.Modified.Format("2006-01-02 15:04:05"),
            "Format":        config.Format,
            "Size":          fmt.Sprintf("%d bytes", config.Size),
        })

        ui.SubHeader("Environment")
        env := getEnvironmentInfo()
        for key, value := range env {
            ui.Debug("%s: %s", key, value)
        }
    }

    // Processing (with appropriate detail level)
    phases := getProcessingPhases()
    for i, phase := range phases {
        if ui.Context().Verbose {
            ui.Status("Phase %d/%d: %s", i+1, len(phases), phase.Name)
        } else {
            ui.Progress(i+1, len(phases), "Processing")
        }

        if err := executePhase(phase); err != nil {
            ui.Error("Phase %s failed: %v", phase.Name, err)
            return err
        }

        if ui.Context().Verbose {
            ui.Success("Completed %s", phase.Name)
        }
    }

    ui.Success("Operation completed successfully")
    return nil
}
```

#### Quiet Mode Respect
```go
func quietModeAware(ui *ui.Output, critical bool) {
    // Critical information (shown even in quiet mode)
    if critical {
        ui.Error("Critical error occurred")
        ui.Info("System requires immediate attention")
    }

    // Non-critical information (respects quiet mode)
    if !ui.Context().Quiet {
        ui.Info("Processing background tasks...")
        ui.Status("Cleaning temporary files...")
    }

    // Important results (always shown)
    ui.Success("Operation completed")
}
```

### Component-Specific Integration

#### PVM Component Example
```go
func pvmListVersions(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    ui.Header("Available Perl Versions")

    versions, err := pvm.GetInstalledVersions()
    if err != nil {
        ui.Error("Failed to list versions: %v", err)
        return err
    }

    if len(versions) == 0 {
        ui.Warning("No Perl versions installed")
        ui.Info("Install a version with: pvm install <version>")
        ui.Info("See available versions: pvm available")
        return nil
    }

    headers := []string{"Version", "Status", "Location"}
    rows := make([][]string, len(versions))

    for i, v := range versions {
        status := "Available"
        if v.IsCurrent() {
            status = "Current"
        }
        if v.IsSystem() {
            status = "System"
        }

        rows[i] = []string{v.Version, status, v.Path}
    }

    ui.Table(headers, rows)

    if current, err := pvm.GetCurrentVersion(); err == nil {
        ui.Info("Current version: %s", current.Version)
    }

    ui.Info("Switch versions with: pvm use <version>")

    return nil
}
```

#### PVX Component Example
```go
func pvxRun(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    scriptPath := args[0]

    ui.Info("Executing script: %s", scriptPath)

    // Dependency analysis
    ui.Status("Analyzing dependencies...")
    deps, err := pvx.AnalyzeDependencies(scriptPath)
    if err != nil {
        ui.Error("Dependency analysis failed: %v", err)
        return err
    }

    if len(deps.Missing) > 0 {
        ui.Warning("Missing dependencies detected:")
        ui.List(deps.Missing)
        ui.Info("Install missing dependencies with:")
        ui.Printf("  pvi install %s\n", strings.Join(deps.Missing, " "))

        continueFlag, _ := cmd.Flags().GetBool("continue")
        if !continueFlag {
            ui.Error("Cannot execute without dependencies")
            ui.Info("Use --continue to execute anyway")
            return fmt.Errorf("missing dependencies")
        }
    }

    // Execution
    ui.Status("Executing script...")
    start := time.Now()

    result, err := pvx.ExecuteScript(scriptPath, args[1:])
    duration := time.Since(start)

    if err != nil {
        ui.Error("Script execution failed: %v", err)
        if ui.Context().Verbose && result.Stderr != "" {
            ui.SubHeader("Error Output")
            ui.Printf("%s", result.Stderr)
        }
        return err
    }

    ui.Success("Script completed in %v", duration)

    if ui.Context().Verbose {
        ui.SubHeader("Execution Details")
        ui.KeyValue(map[string]string{
            "Exit Code":    fmt.Sprintf("%d", result.ExitCode),
            "Duration":     duration.String(),
            "Memory Used":  fmt.Sprintf("%.1f MB", result.MemoryMB),
        })

        if result.Stdout != "" {
            ui.SubHeader("Output")
            ui.Printf("%s", result.Stdout)
        }
    }

    return nil
}
```

### Testing Integration

#### Unit Test Example
```go
func TestUIOutput(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever,
        Quiet:     false,
        Verbose:   false,
    }
    output := ui.NewOutput(ctx)

    // Test basic output
    output.Success("Test success message")
    output.Error("Test error message")
    output.Info("Test info message")

    result := buf.String()

    assert.Contains(t, result, "✓ Test success message")
    assert.Contains(t, result, "✗ Test error message")
    assert.Contains(t, result, "ℹ Test info message")

    // Test structured output
    buf.Reset()
    headers := []string{"Col1", "Col2"}
    rows := [][]string{{"Val1", "Val2"}, {"Val3", "Val4"}}
    output.Table(headers, rows)

    tableOutput := buf.String()
    assert.Contains(t, tableOutput, "Col1")
    assert.Contains(t, tableOutput, "Col2")
    assert.Contains(t, tableOutput, "Val1")
    assert.Contains(t, tableOutput, "Val2")
}
```

#### Integration Test Example
```go
func TestCommandIntegration(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    // Test normal output
    stdout := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "version"}, "Version command should work")

    assert.Contains(t, stdout, "PVM")
    assert.Regexp(t, `\d+\.\d+\.\d+`, stdout)

    // Test verbose output
    verboseOut := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "--verbose", "version"}, "Verbose version should work")

    assert.Contains(t, verboseOut, "Version Information")
    assert.Contains(t, verboseOut, "Build Time")
    assert.Contains(t, verboseOut, "Commit")
    assert.Greater(t, len(verboseOut), len(stdout))

    // Test error handling
    stderr := helpers.AssertPVMFails(t, env,
        []string{"pvm", "nonexistent"}, "Unknown command should fail")

    assert.Contains(t, stderr, "✗ unknown command")
    assert.Contains(t, stderr, "Did you mean?")
}
```

---

## Troubleshooting and Common Issues

### Common Problems and Solutions

#### Problem: Output Not Appearing
**Symptoms**: Commands run but produce no visible output
**Causes**:
- Quiet mode enabled (`--quiet` flag or context setting)
- UI context not properly initialized
- Output method called but context is nil

**Solutions**:
```go
// Check if UI context is properly set
ui := cli.GetUI(cmd)
if ui.Context().Quiet {
    ui.Info("Quiet mode is enabled")
}

// Force output regardless of quiet mode (use sparingly)
ui.Error("Critical error - always shown")

// Debug UI context
if ui.Context().Verbose {
    ui.Debug("UI Context: Quiet=%v, Verbose=%v, ColorMode=%v",
        ui.Context().Quiet, ui.Context().Verbose, ui.Context().ColorMode)
}
```

#### Problem: Colors Not Working
**Symptoms**: Output appears but without colors or styling
**Causes**:
- Terminal doesn't support colors
- Color mode set to `ColorNever`
- Environment variable forcing no colors

**Solutions**:
```bash
# Test with forced colors
pvm --color=always version

# Check environment variables
unset NO_COLOR
unset FORCE_COLOR

# Debug color detection
pvm --verbose --debug version
```

**Code fixes**:
```go
// Force color mode for testing
ui := cli.GetUI(cmd)
ui.SetColorMode(ui.ColorAlways)

// Check color support
if ui.Context().ColorMode == ui.ColorNever {
    ui.Debug("Color output disabled")
}
```

#### Problem: Inconsistent Formatting
**Symptoms**: Mixed styled and unstyled output
**Causes**:
- Mixing UI framework with direct `fmt.Print*` calls
- Some commands not converted to use UI framework
- Third-party libraries producing direct output

**Solutions**:
```bash
# Find remaining direct output calls
grep -r "fmt\.Print" internal/
grep -r "cmd\.Print" cmd/

# Look for mixed output patterns
grep -r "os\.Stdout" internal/
```

**Code fixes**:
```go
// WRONG - mixing output methods
fmt.Printf("Processing...\n")
ui.Success("Done")

// CORRECT - use UI framework consistently
ui.Status("Processing...")
ui.Success("Done")
```

#### Problem: Performance Issues
**Symptoms**: Slow command execution, UI lag
**Causes**:
- Expensive operations in UI calls
- Excessive output in tight loops
- Inefficient string formatting

**Solutions**:
```go
// WRONG - expensive operation in UI call
ui.Info("Processing: %s", expensiveOperation())

// CORRECT - evaluate first, then display
result := expensiveOperation()
ui.Info("Processing: %s", result)

// WRONG - UI calls in tight loop
for _, item := range hugelist {
    ui.Status("Processing %s", item)
    process(item)
}

// CORRECT - batch or periodic updates
for i, item := range hugelist {
    if i%100 == 0 {
        ui.Progress(i, len(hugelist), "Processing items")
    }
    process(item)
}
```

#### Problem: Test Failures After Migration
**Symptoms**: Tests fail after converting to UI framework
**Causes**:
- Tests expect specific output format
- Tests capture output incorrectly
- Missing UI context in tests

**Solutions**:
```go
// Update test to use UI framework
func TestCommandWithUI(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever, // Disable colors for testing
        Quiet:     false,
        Verbose:   false,
    }

    cmd := &cobra.Command{}
    cmd.SetContext(cli.WithUI(context.Background(), ui.NewOutput(ctx)))

    err := runCommand(cmd, []string{"test"})
    assert.NoError(t, err)

    output := buf.String()
    assert.Contains(t, output, "expected content")
}
```

### Debug Mode Usage

#### Enabling Debug Output
```bash
# Enable verbose and debug mode
pvm --verbose --debug command

# Check UI framework status
PVM_DEBUG=ui pvm command

# Full debugging
PVM_DEBUG=all pvm --verbose command
```

#### Debug Information Available
- UI context configuration
- Color mode detection
- Style application decisions
- Output routing information
- Performance metrics

#### Debug Code Patterns
```go
func debugUIState(ui *ui.Output) {
    if ui.Context().Verbose {
        ui.Debug("=== UI Framework Debug ===")
        ui.Debug("Quiet: %v", ui.Context().Quiet)
        ui.Debug("Verbose: %v", ui.Context().Verbose)
        ui.Debug("ColorMode: %v", ui.Context().ColorMode)
        ui.Debug("Interactive: %v", ui.Context().Interactive)
        ui.Debug("Writer: %T", ui.Context().Writer)
        ui.Debug("===========================")
    }
}
```

### Performance Optimization

#### Profiling UI Performance
```bash
# Profile command execution
go tool pprof -http=:8080 pvm profile.prof

# Time command execution
time pvm command args

# Memory usage analysis
pvm --verbose command 2>&1 | grep -i memory
```

#### Optimization Techniques
```go
// Lazy evaluation for expensive debug info
if ui.Context().Verbose {
    ui.Debug("Expensive debug info: %s", func() string {
        return generateExpensiveDebugInfo()
    }())
}

// Batch output operations
var outputBuffer strings.Builder
for _, item := range items {
    outputBuffer.WriteString(fmt.Sprintf("  %s\n", item))
}
ui.Printf(outputBuffer.String())

// Early return for quiet mode
if ui.Context().Quiet && !critical {
    return
}
ui.Info("Non-critical information")
```

---

## Future Enhancement Guidelines

### Planned Enhancements

#### 1. Custom Themes
```go
// User-defined color schemes
type CustomTheme struct {
    Name   string
    Colors map[string]string
}

func LoadCustomTheme(path string) (*CustomTheme, error)
func (ui *Output) SetTheme(theme *CustomTheme)
```

#### 2. Plugin System
```go
// Allow extensions to add custom output formats
type OutputPlugin interface {
    Name() string
    Render(content string, context *UIContext) string
}

func RegisterOutputPlugin(plugin OutputPlugin)
func (ui *Output) UsePlugin(name string) error
```

#### 3. Interactive Elements
```go
// Progress bars, confirmations, selections
func (ui *Output) ProgressBar(current, total int, message string)
func (ui *Output) Confirm(message string) bool
func (ui *Output) Select(options []string) (int, error)
```

#### 4. Export Formats
```go
// JSON, XML, CSV output modes
func (ui *Output) SetFormat(format OutputFormat)
func (ui *Output) ExportTable(headers []string, rows [][]string, format OutputFormat)
```

#### 5. Internationalization
```go
// Multi-language support
func (ui *Output) SetLanguage(lang string)
func (ui *Output) T(key string, args ...interface{}) string
```

### Extension Points

#### Adding New Output Methods
1. Define the method signature in the `OutputRenderer` interface
2. Implement the method in the `Output` struct
3. Add appropriate styling using the `Styles` system
4. Write comprehensive tests
5. Update documentation

#### Creating Custom Styles
```go
// Extend the Styles struct
type ExtendedStyles struct {
    Styles
    CustomStyle lipgloss.Style
}

// Override GetDefaultStyles
func GetCustomStyles() ExtendedStyles {
    base := GetDefaultStyles()
    return ExtendedStyles{
        Styles: base,
        CustomStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5722")),
    }
}
```

#### Adding Terminal Features
- Mouse support for interactive elements
- Advanced layouts using lipgloss containers
- Real-time updates and streaming output
- Terminal size detection and responsive layouts

### Contribution Guidelines

#### When Contributing to UI Framework

1. **Follow Existing Patterns**: Use established styling and naming conventions
2. **Add Comprehensive Tests**: Include unit tests, integration tests, and visual tests
3. **Update Documentation**: Modify this documentation and add examples
4. **Test Across Terminals**: Verify functionality in different terminal environments
5. **Verify Accessibility**: Ensure color contrast and screen reader compatibility
6. **Profile Performance**: Measure impact of new features
7. **Maintain Backward Compatibility**: Don't break existing functionality

#### Code Review Checklist

- [ ] No direct `fmt.Print*` or `cmd.Print*` calls
- [ ] Proper context awareness (quiet/verbose modes)
- [ ] Consistent styling patterns
- [ ] Comprehensive test coverage
- [ ] Documentation updates
- [ ] Performance impact assessed
- [ ] Accessibility compliance verified
- [ ] Cross-platform compatibility tested

#### Testing Requirements

- Unit tests with >95% coverage
- Integration tests for command workflows
- Visual regression tests for styling changes
- Performance benchmarks for new features
- Accessibility compliance validation

### Migration Path for Future Changes

#### Adding New Components
1. Follow the established component integration pattern
2. Use `cli.GetUI(cmd)` for UI access
3. Implement all output through UI framework methods
4. Add comprehensive tests
5. Update component documentation

#### Deprecating Legacy Patterns
1. Identify deprecated patterns
2. Provide migration guides
3. Add deprecation warnings
4. Support transition period
5. Remove after appropriate notice

This comprehensive documentation provides everything needed to understand, use, and extend the PVM Fang UI Framework. The system transforms the CLI experience while maintaining architectural integrity and providing a foundation for future enhancements.
