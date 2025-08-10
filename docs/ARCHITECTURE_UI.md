# PVM UI Framework Architecture

## Overview

The PVM UI Framework provides a unified, beautiful interface across all PVM components (pvm, pvx, pm, psc). Built on Charm's Fang library, it replaces direct output calls with styled, consistent CLI experiences that adapt to user preferences and terminal capabilities.

## Design Goals

1. **Consistency**: Identical styling and behavior across all components
2. **Performance**: Minimal overhead with efficient rendering
3. **Accessibility**: Proper contrast, fallbacks, and screen reader support
4. **Flexibility**: Supports multiple output modes and customization
5. **Maintainability**: Clean separation of concerns and testable architecture

## Core Architecture

### Component Structure

```
internal/cli/ui/
├── output.go          # Main Output interface and implementation
├── styles.go          # Fang styling definitions and themes
├── types.go           # Type definitions and color modes
├── context.go         # UI context management
├── table.go           # Table rendering functionality
├── progress.go        # Progress indicators and status
└── output_test.go     # Comprehensive test suite
```

### Key Types

```go
// Core output interface
type Output struct {
    ctx *UIContext
}

// Context carries UI state and preferences
type UIContext struct {
    Writer      io.Writer
    ColorMode   ColorMode
    Quiet       bool
    Verbose     bool
    Interactive bool
}

// Color mode handling
type ColorMode int
const (
    ColorAuto   ColorMode = iota
    ColorAlways
    ColorNever
)
```

### Integration Points

```go
// CLI integration through context
func GetUI(cmd *cobra.Command) *ui.Output
func WithUI(ctx context.Context, output *ui.Output) context.Context

// Component registration
cli.RegisterComponent("pvm", pvm.NewCommand)
cli.RegisterComponent("pvx", pvx.NewCommand)
cli.RegisterComponent("pm", pm.NewCommand)
cli.RegisterComponent("psc", psc.NewCommand)
```

## Information Flow

### Command Execution Flow

```
User Command
    ↓
CLI Router (internal/cli/router.go)
    ↓
Component Detection
    ↓
UI Context Creation (internal/cli/root.go)
    ↓
Command Execution with UI Access
    ↓
Styled Output via Fang
    ↓
Terminal Display
```

### Context Propagation

```go
// Context flows through command execution
rootCmd.SetContext(cli.WithUI(ctx, ui.NewOutput(uiCtx)))

// Commands access UI through context
func runCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)  // Extracts UI from context
    ui.Success("Operation completed")
    return nil
}
```

## Styling System

### Theme Definition

```go
// styles.go - Fang style definitions
var (
    SuccessStyle = fang.NewStyle().
        Foreground(fang.Color("#22AA22")).
        Bold(true)

    ErrorStyle = fang.NewStyle().
        Foreground(fang.Color("#CC2222")).
        Bold(true)

    // ... additional styles
)
```

### Adaptive Rendering

```go
// Styles adapt to color mode and terminal capabilities
func (o *Output) renderWithStyle(style fang.Style, format string, args ...interface{}) {
    text := fmt.Sprintf(format, args...)

    switch o.ctx.ColorMode {
    case ColorNever:
        fmt.Fprint(o.ctx.Writer, text)
    case ColorAlways:
        fmt.Fprint(o.ctx.Writer, style.Render(text))
    case ColorAuto:
        if fang.HasDarkBackground() {
            fmt.Fprint(o.ctx.Writer, style.Render(text))
        } else {
            fmt.Fprint(o.ctx.Writer, text)
        }
    }
}
```

## Output Methods

### Message Types

```go
// Core output methods with consistent styling
func (o *Output) Success(format string, args ...interface{})   // ✓ Green
func (o *Output) Error(format string, args ...interface{})    // ✗ Red
func (o *Output) Warning(format string, args ...interface{})  // ⚠ Yellow
func (o *Output) Info(format string, args ...interface{})     // ℹ Blue
func (o *Output) Debug(format string, args ...interface{})    // Gray (verbose only)
```

### Structured Output

```go
// Headers and organization
func (o *Output) Header(text string)         // Large header with underline
func (o *Output) SubHeader(text string)      // Medium header with accent

// Data presentation
func (o *Output) Table(headers []string, rows [][]string)  // Aligned table
func (o *Output) List(items []string)                      // Bulleted list

// Progress and status
func (o *Output) Progress(current, total int, message string)  // Progress bar
func (o *Output) Status(message string)                        // Status update
```

### Context-Aware Methods

```go
// Methods that respect user preferences
func (o *Output) VerboseInfo(format string, args ...interface{}) {
    if o.ctx.Verbose {
        o.Info(format, args...)
    }
}

func (o *Output) QuietSuccess(format string, args ...interface{}) {
    if !o.ctx.Quiet {
        o.Success(format, args...)
    }
}
```

## Performance Optimizations

### Efficient Rendering

1. **Buffer Reuse**: Internal buffers are reused across calls
2. **Lazy Formatting**: Format strings evaluated only when needed
3. **Style Caching**: Fang styles computed once and cached
4. **Context Checks**: Expensive operations skipped in quiet mode

```go
// Efficient pattern - check context before expensive operations
if ui.Context().Verbose {
    expensiveDebugInfo := computeExpensiveInfo()
    ui.Debug("Debug info: %s", expensiveDebugInfo)
}
```

### Memory Management

```go
// Output reuses buffers and streams large content
type Output struct {
    ctx    *UIContext
    buffer *bytes.Buffer  // Reused for formatting
    writer io.Writer      // Direct streaming for large output
}
```

## Testing Architecture

### Unit Testing

```go
// Testable UI with captured output
func TestUIOutput(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever,
        Quiet:     false,
        Verbose:   true,
    }
    output := ui.NewOutput(ctx)

    output.Success("Test message")
    result := buf.String()

    assert.Contains(t, result, "✓")
    assert.Contains(t, result, "Test message")
}
```

### Integration Testing

```go
// End-to-end testing with real commands
func TestCommandIntegration(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    stdout := helpers.AssertPVMSucceeds(t, env,
        []string{"pvm", "help"}, "Help should work")

    // Verify styled output
    assert.Contains(t, stdout, "Usage:")
    assert.NotContains(t, stdout, "Error")
}
```

### Quality Assurance Tests

```go
// Prevent ugly output patterns
func TestUglyOutputPrevention(t *testing.T) {
    // Test against known ugly patterns like repetitive progress bars
    output := captureOutput(runLongOperation)

    progressBarCount := countProgressBars(output)
    assert.Less(t, progressBarCount, 5, "Too many progress indicators")

    pathCount := countLongPaths(output)
    assert.Less(t, pathCount, 3, "Too many file paths shown")
}
```

## Error Handling Architecture

### Separation of Concerns

```go
// Internal packages return structured errors
func processFile(path string) error {
    if err := validate(path); err != nil {
        return errors.NewUserInputError("PVM", "001",
            "Invalid file", err).WithLocation(path)
    }
    return nil
}

// CLI layer handles display formatting
func runCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)

    for _, file := range args {
        if err := processFile(file); err != nil {
            ui.Error("Processing failed: %v", err)  // UI formats error
            continue
        }
        ui.Success("Processed: %s", file)
    }
    return nil
}
```

### Error Display Consistency

```go
// All errors get consistent styling and context
func (o *Output) Error(format string, args ...interface{}) {
    icon := "✗"
    text := fmt.Sprintf(format, args...)

    styled := o.errorStyle.Render(fmt.Sprintf("%s %s", icon, text))
    fmt.Fprintln(o.ctx.Writer, styled)
}
```

## Extension Points

### Custom Output Types

```go
// Framework supports extension with new output types
func (o *Output) CustomMessage(icon string, style fang.Style, format string, args ...interface{}) {
    text := fmt.Sprintf(format, args...)
    rendered := style.Render(fmt.Sprintf("%s %s", icon, text))
    fmt.Fprintln(o.ctx.Writer, rendered)
}
```

### Theme Customization

```go
// Themes can be swapped or customized
type Theme struct {
    Success fang.Style
    Error   fang.Style
    Warning fang.Style
    Info    fang.Style
}

func (o *Output) SetTheme(theme Theme) {
    o.successStyle = theme.Success
    o.errorStyle = theme.Error
    // ... etc
}
```

### Format Plugins

```go
// Future: Support for different output formats
type OutputFormatter interface {
    FormatSuccess(message string) string
    FormatError(message string) string
    FormatTable(headers []string, rows [][]string) string
}

// JSON formatter, XML formatter, etc.
```

## Migration Support

### Backward Compatibility

The UI framework maintains 100% backward compatibility with existing functionality while providing enhanced visual experience:

```go
// Old: Direct output
fmt.Printf("Processing %s\n", filename)

// New: Styled output with same functionality
ui.Info("Processing %s", filename)
```

### Gradual Migration

Components can be migrated incrementally without breaking functionality:

1. **Phase 1**: Add UI framework alongside existing output
2. **Phase 2**: Replace direct output calls with UI methods
3. **Phase 3**: Remove legacy output patterns
4. **Phase 4**: Add enhanced features (tables, progress, etc.)

## Security Considerations

### Output Sanitization

```go
// All output is properly escaped and sanitized
func (o *Output) sanitizeOutput(text string) string {
    // Remove potential ANSI injection
    text = stripANSICodes(text)

    // Limit output length to prevent DOS
    if len(text) > maxOutputLength {
        text = text[:maxOutputLength] + "..."
    }

    return text
}
```

### Terminal Security

```go
// Framework protects against terminal manipulation
func (o *Output) secureRender(text string) string {
    // Only allow safe ANSI codes
    return allowedANSIOnly(text)
}
```

## Future Architecture

### Planned Enhancements

1. **Interactive Elements**: Progress bars, confirmations, selections
2. **Custom Themes**: User-defined color schemes and styling
3. **Export Formats**: JSON, XML, CSV output modes
4. **Plugin System**: Third-party output format extensions
5. **Advanced Layouts**: Multi-column, paned interfaces
6. **Terminal Features**: Mouse support, advanced positioning

### Extensibility Framework

```go
// Plugin architecture for future enhancements
type UIPlugin interface {
    Name() string
    Render(context UIContext, data interface{}) string
    Supports(outputType OutputType) bool
}

func RegisterUIPlugin(plugin UIPlugin) {
    uiPlugins[plugin.Name()] = plugin
}
```

This architecture provides a solid foundation for beautiful, consistent CLI experiences across the entire PVM ecosystem while maintaining performance, testability, and extensibility.
