# PVM Fang UI Framework Documentation

## Overview

The PVM Fang UI Framework provides beautiful, consistent styling across all PVM components (pvm, pvx, pm, psc). Built on top of Charm's Fang library, it replaces all direct output calls with styled, user-friendly displays that enhance the CLI experience.

## Architecture

### Core Components

- **internal/cli/ui/output.go** - Main output interface and methods
- **internal/cli/ui/styles.go** - Fang styling definitions and themes
- **internal/cli/ui/types.go** - Type definitions and interfaces
- **internal/cli/ui/context.go** - UI context management
- **internal/cli/root.go** - CLI integration and context injection

### Design Principles

1. **Separation of Concerns**: Internal packages return structured data, CLI layer handles formatting
2. **Global Consistency**: All components use identical styling patterns
3. **Context Awareness**: UI adapts to quiet/verbose modes, color preferences, and terminal capabilities
4. **Performance First**: Minimal overhead with efficient rendering
5. **Accessibility**: Proper color contrast and fallback modes

## API Reference

### Core Output Methods

```go
type Output struct {
    ctx *UIContext
}

// Basic output methods
func (o *Output) Success(format string, args ...interface{})
func (o *Output) Error(format string, args ...interface{})
func (o *Output) Warning(format string, args ...interface{})
func (o *Output) Info(format string, args ...interface{})
func (o *Output) Debug(format string, args ...interface{})

// Formatted output
func (o *Output) Printf(format string, args ...interface{})
func (o *Output) Print(a ...interface{})
func (o *Output) Println(a ...interface{})

// Structured output
func (o *Output) Header(text string)
func (o *Output) SubHeader(text string)
func (o *Output) Table(headers []string, rows [][]string)
func (o *Output) List(items []string)
func (o *Output) Progress(current, total int, message string)
func (o *Output) Status(message string)
```

### Context Management

```go
type UIContext struct {
    Writer      io.Writer
    ColorMode   ColorMode
    Quiet       bool
    Verbose     bool
    Interactive bool
}

// Context methods
func (o *Output) Context() *UIContext
func (o *Output) SetQuiet(quiet bool)
func (o *Output) SetVerbose(verbose bool)
func (o *Output) SetColorMode(mode ColorMode)
```

### Color Modes

```go
type ColorMode int

const (
    ColorAuto   ColorMode = iota // Auto-detect terminal capabilities
    ColorAlways                  // Always use colors
    ColorNever                   // Never use colors
)
```

## Usage Guidelines

### Getting UI Access in Commands

```go
func runCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    // Use UI methods instead of fmt.Print* or cmd.Print*
    ui.Info("Processing file: %s", filename)

    if err != nil {
        ui.Error("Failed to process: %v", err)
        return err
    }

    ui.Success("File processed successfully")
    return nil
}
```

### Error Handling Pattern

```go
// Internal packages return structured errors
func processFile(path string) error {
    if _, err := os.Stat(path); err != nil {
        return errors.NewUserInputError("PVM", "001",
            "File not found", err).WithLocation(path)
    }
    return nil
}

// CLI layer handles UI formatting
func runCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    for _, file := range args {
        if err := processFile(file); err != nil {
            // UI framework handles error display formatting
            ui.Error("Processing failed: %v", err)
            continue
        }
        ui.Success("Processed: %s", file)
    }
    return nil
}
```

### Structured Output Examples

#### Table Output
```go
ui := cli.GetUI(cmd)

headers := []string{"Component", "Status", "Version"}
rows := [][]string{
    {"PVM", "Active", "1.0.0"},
    {"PVX", "Active", "1.0.0"},
    {"PM", "Active", "1.0.0"},
    {"PSC", "Active", "1.0.0"},
}

ui.Table(headers, rows)
```

#### List Output
```go
ui := cli.GetUI(cmd)

steps := []string{
    "Initialize project structure",
    "Configure Perl version",
    "Install dependencies",
    "Run tests",
}

ui.Header("Setup Steps")
ui.List(steps)
```

#### Progress Indicators
```go
ui := cli.GetUI(cmd)

for i, file := range files {
    ui.Progress(i+1, len(files), "Processing files")
    processFile(file)
}
```

## Styling Patterns

### Message Types

- **Success**: Green checkmark (✓) with success message
- **Error**: Red X (✗) with error details
- **Warning**: Yellow warning (⚠) with caution message
- **Info**: Blue info (ℹ) with informational text
- **Debug**: Gray debug symbol with detailed output (verbose mode only)

### Headers and Structure

- **Header**: Large, bold text with underline
- **SubHeader**: Medium text with subtle styling
- **Table**: Aligned columns with proper spacing
- **List**: Bulleted items with consistent indentation

### Color Scheme

- **Primary**: Blue (#0066CC) for informational elements
- **Success**: Green (#22AA22) for successful operations
- **Warning**: Yellow (#FFAA00) for warnings and cautions
- **Error**: Red (#CC2222) for errors and failures
- **Muted**: Gray (#666666) for secondary information

## Integration Patterns

### Adding UI to New Commands

1. **Create Command Function**
```go
func newMyCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mycommand [args]",
        Short: "Description of my command",
        RunE: func(cmd *cobra.Command, args []string) error {
            ui := cli.GetUI(cmd)
            return runMyCommand(ui, args)
        },
    }
    return cmd
}
```

2. **Implement Command Logic**
```go
func runMyCommand(ui *ui.Output, args []string) error {
    ui.Info("Starting command execution")

    for _, arg := range args {
        if err := processArg(arg); err != nil {
            ui.Error("Failed to process %s: %v", arg, err)
            continue
        }
        ui.Success("Processed: %s", arg)
    }

    return nil
}
```

### Converting Existing Commands

**Before (direct output):**
```go
func oldCommand(cmd *cobra.Command, args []string) error {
    fmt.Printf("Processing %d files...\n", len(args))

    for _, file := range args {
        fmt.Printf("Processing: %s\n", file)
        if err := process(file); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            return err
        }
    }

    fmt.Println("All files processed successfully")
    return nil
}
```

**After (UI framework):**
```go
func newCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    ui.Info("Processing %d files...", len(args))

    for _, file := range args {
        ui.Status("Processing: %s", file)
        if err := process(file); err != nil {
            ui.Error("Processing failed: %v", err)
            return err
        }
    }

    ui.Success("All files processed successfully")
    return nil
}
```

## Context Handling

### Respecting User Preferences

```go
ui := cli.GetUI(cmd)

// Conditional output based on verbosity
if ui.Context().Verbose {
    ui.Debug("Detailed processing information")
    ui.Info("Processing file with advanced options")
}

// Respect quiet mode
if !ui.Context().Quiet {
    ui.Status("Background processing...")
}

// Always show important information
ui.Success("Operation completed")
```

### Dynamic Context Modification

```go
ui := cli.GetUI(cmd)

// Temporarily increase verbosity for debugging
if debugFlag {
    ui.SetVerbose(true)
    defer ui.SetVerbose(false)
}

// Disable colors for scripting
if scriptMode {
    ui.SetColorMode(ui.ColorNever)
}
```

## Performance Considerations

### Efficient Output

- **Batch Operations**: Group related output calls
- **Lazy Formatting**: Use format strings, avoid string concatenation
- **Context Checks**: Check quiet/verbose before expensive operations

```go
// Efficient pattern
if ui.Context().Verbose {
    ui.Debug("Expensive debug info: %s", expensiveOperation())
}

// Inefficient pattern
debugInfo := expensiveOperation() // Always runs
ui.Debug("Debug info: %s", debugInfo)
```

### Memory Management

- **Buffer Reuse**: UI framework reuses internal buffers
- **Stream Output**: Large outputs are streamed, not buffered
- **Context Sharing**: UI contexts are shared across command execution

## Testing UI Output

### Unit Testing

```go
func TestCommandOutput(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever,
        Quiet:     false,
        Verbose:   true,
    }
    output := ui.NewOutput(ctx)

    // Test command logic
    err := runCommand(output, []string{"test.pl"})
    assert.NoError(t, err)

    result := buf.String()
    assert.Contains(t, result, "Success")
    assert.Contains(t, result, "test.pl")
}
```

### Integration Testing

```go
func TestCommandIntegration(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    stdout := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "help"}, "Help should work")

    // Verify styled output
    assert.Contains(t, stdout, "Usage:")
    assert.Contains(t, stdout, "Commands:")

    // Verify no direct fmt.Print output
    lines := strings.Split(stdout, "\n")
    for _, line := range lines {
        if strings.Contains(line, "Usage:") {
            // Should be styled
            break
        }
    }
}
```

## Migration Checklist

When converting a command to use the UI framework:

- [ ] Replace all `fmt.Print*` calls with `ui.*` methods
- [ ] Replace all `cmd.Print*` calls with `ui.*` methods
- [ ] Update error handling to use UI error formatting
- [ ] Add appropriate success/status messages
- [ ] Test in verbose and quiet modes
- [ ] Verify color and no-color output
- [ ] Update associated tests
- [ ] Document any new output patterns

## Troubleshooting

### Common Issues

**Issue**: Output not appearing
- **Cause**: Quiet mode enabled
- **Solution**: Check `ui.Context().Quiet` or use methods that always output

**Issue**: Colors not working
- **Cause**: Terminal doesn't support colors or ColorNever mode
- **Solution**: Test with `--color=always` flag

**Issue**: Inconsistent formatting
- **Cause**: Mixing UI framework with direct output
- **Solution**: Replace all direct output calls with UI methods

**Issue**: Performance problems
- **Cause**: Expensive operations in UI calls
- **Solution**: Move expensive operations outside UI calls, use lazy evaluation

### Debug Mode

Enable debug output to troubleshoot UI issues:

```bash
pvm --verbose --debug command args
```

This will show:
- UI context information
- Styling decisions
- Output routing
- Performance metrics

## Best Practices

1. **Always use UI framework** - Never mix with direct output calls
2. **Respect context** - Check quiet/verbose modes appropriately
3. **Consistent messaging** - Use similar patterns across components
4. **Error clarity** - Provide actionable error messages
5. **Progress feedback** - Show progress for long-running operations
6. **Test all modes** - Verify quiet, verbose, and color modes
7. **Performance awareness** - Keep UI calls efficient
8. **Accessibility** - Ensure output works with screen readers

## Future Enhancements

The UI framework is designed to be extensible:

- **Custom Themes**: Support for user-defined color schemes
- **Plugin System**: Allow extensions to add custom output formats
- **Interactive Elements**: Progress bars, confirmations, selections
- **Export Formats**: JSON, XML, CSV output modes
- **Internationalization**: Multi-language support
- **Terminal Features**: Mouse support, advanced layouts

## Contributing

When contributing to the UI framework:

1. Follow existing styling patterns
2. Add comprehensive tests
3. Update documentation
4. Test across different terminals
5. Verify accessibility compliance
6. Profile performance impact
7. Maintain backward compatibility

The Fang UI Framework transforms PVM from a functional CLI tool into a beautiful, user-friendly experience while maintaining all existing functionality and performance characteristics.
