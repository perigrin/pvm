# PVM UI Framework Troubleshooting Guide

This guide provides comprehensive troubleshooting information for the PVM Fang UI Framework, covering common issues, diagnostic procedures, and solutions.

## Quick Diagnostic Commands

Before diving into specific issues, run these diagnostic commands to gather information:

```bash
# Check UI framework status
pvm --verbose --debug version

# Test output modes
pvm --help                    # Normal mode
pvm --quiet --help           # Quiet mode
pvm --verbose --help         # Verbose mode

# Check for direct output calls
grep -r "fmt\.Print" internal/
grep -r "cmd\.Print" cmd/

# Verify test suite
make test
```

## Common Issues and Solutions

### 1. No Output Appearing

#### Symptoms
- Commands execute but produce no visible output
- Success/error messages missing
- Progress indicators not showing

#### Diagnosis
```bash
# Check if quiet mode is enabled
pvm --verbose version

# Verify UI context initialization
PVM_DEBUG=ui pvm version
```

#### Causes and Solutions

**Cause**: Quiet mode enabled
```go
// Check quiet mode status
ui := cli.GetUI(cmd)
if ui.Context().Quiet {
    ui.Debug("Quiet mode is active")
}

// Solution: Use appropriate output methods
ui.Error("Critical errors always show")  // Shows even in quiet mode
ui.Success("Important results show")     // Shows even in quiet mode
ui.Info("Regular info respects quiet")   // Hidden in quiet mode
```

**Cause**: UI context not initialized
```go
// WRONG - UI context missing
func badCommand(cmd *cobra.Command, args []string) error {
    fmt.Println("This bypasses UI framework")
    return nil
}

// CORRECT - Proper UI initialization
func goodCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    ui.Info("This uses UI framework")
    return nil
}
```

**Cause**: Output redirected or captured
```bash
# Test direct output
pvm version

# Test with explicit output
pvm version 2>&1 | cat

# Check if output is going to stderr
pvm version 2>/dev/null
```

### 2. Colors and Styling Not Working

#### Symptoms
- Text appears without colors
- Icons/symbols not displaying correctly
- Formatting appears broken

#### Diagnosis
```bash
# Check terminal capabilities
echo $TERM
echo $COLORTERM

# Test color support
pvm --color=always version
pvm --color=never version

# Check environment variables
env | grep -E "(COLOR|TERM)"
```

#### Causes and Solutions

**Cause**: Terminal doesn't support colors
```bash
# Force color mode for testing
export FORCE_COLOR=1
pvm version

# Or use command flag
pvm --color=always version
```

**Cause**: Color mode incorrectly set
```go
// Check current color mode
ui := cli.GetUI(cmd)
ui.Debug("Color mode: %v", ui.Context().ColorMode)

// Force specific color mode
ui.SetColorMode(ui.ColorAlways)  // Force colors
ui.SetColorMode(ui.ColorNever)   // Disable colors
ui.SetColorMode(ui.ColorAuto)    // Auto-detect
```

**Cause**: Environment variables disabling colors
```bash
# Common environment variables that disable colors
unset NO_COLOR
unset MONOCHROME
unset TERM_PROGRAM

# Reset terminal type
export TERM=xterm-256color
```

**Cause**: Font/character encoding issues
```bash
# Test character support
echo "✓ ✗ ⚠ ℹ 🐛 → ⚡"

# Use ASCII-only mode if needed
PVM_ASCII_ONLY=1 pvm version
```

### 3. Inconsistent Output Formatting

#### Symptoms
- Mix of styled and unstyled output
- Inconsistent message formatting
- Some commands appear differently than others

#### Diagnosis
```bash
# Find mixed output patterns
grep -r "fmt\.Print\|cmd\.Print" internal/ cmd/

# Check for direct stdout/stderr usage
grep -r "os\.Stdout\|os\.Stderr" internal/

# Compare command outputs
pvm --help
pvx --help
pm --help
psc --help
```

#### Causes and Solutions

**Cause**: Mixed output methods
```go
// WRONG - mixing output methods
func inconsistentCommand(cmd *cobra.Command, args []string) error {
    fmt.Printf("Starting process...\n")      // Direct output
    ui := cli.GetUI(cmd)
    ui.Success("Process completed")          // UI framework
    return nil
}

// CORRECT - consistent UI framework usage
func consistentCommand(cmd *cobra.Command, args []string) error {
    ui := cli.GetUI(cmd)
    ui.Status("Starting process...")
    ui.Success("Process completed")
    return nil
}
```

**Cause**: Third-party library output
```go
// Capture third-party output
func captureThirdPartyOutput(ui *ui.Output) error {
    // Redirect stdout temporarily
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    // Run third-party code
    thirdPartyFunction()

    // Restore stdout
    w.Close()
    os.Stdout = oldStdout

    // Read captured output
    output, _ := io.ReadAll(r)

    // Display through UI framework
    ui.Info("Third-party output: %s", string(output))
    return nil
}
```

**Cause**: Incomplete migration
```bash
# Find commands not using UI framework
grep -l "fmt\.Print" cmd/*.go

# Check for cobra command output methods
grep -r "cmd\.Print\|cmd\.Printf" cmd/
```

### 4. Performance Issues

#### Symptoms
- Slow command execution
- UI lag or delays
- High memory usage
- Stuttering progress indicators

#### Diagnosis
```bash
# Profile command execution
time pvm --verbose command args

# Check memory usage
pvm --verbose command args 2>&1 | grep -i memory

# Profile with Go tools
go tool pprof -http=:8080 profile.prof
```

#### Causes and Solutions

**Cause**: Expensive operations in UI calls
```go
// WRONG - expensive operation in UI call
for _, file := range files {
    ui.Info("Processing: %s", analyzeFile(file)) // Expensive
}

// CORRECT - separate computation from display
for _, file := range files {
    analysis := analyzeFile(file)  // Expensive operation
    ui.Info("Processing: %s", analysis)  // Quick display
}
```

**Cause**: Excessive UI calls in loops
```go
// WRONG - UI call per iteration
for i, item := range hugeList {
    ui.Status("Processing item %d: %s", i, item)
    processItem(item)
}

// CORRECT - periodic UI updates
for i, item := range hugeList {
    if i%100 == 0 || i == len(hugeList)-1 {
        ui.Progress(i+1, len(hugeList), "Processing items")
    }
    processItem(item)
}
```

**Cause**: Inefficient string operations
```go
// WRONG - string concatenation in loop
var output string
for _, item := range items {
    output += fmt.Sprintf("  %s\n", item)
}
ui.Printf(output)

// CORRECT - use string builder
var builder strings.Builder
for _, item := range items {
    builder.WriteString(fmt.Sprintf("  %s\n", item))
}
ui.Printf(builder.String())
```

### 5. Test Failures After Migration

#### Symptoms
- Unit tests failing after UI framework conversion
- Integration tests not capturing output correctly
- Expected output format changed

#### Diagnosis
```bash
# Run specific failing tests
go test -v ./internal/pvm -run TestSpecificCommand

# Check test output capture
go test -v ./cmd -run TestCommandOutput

# Compare old vs new test expectations
git diff HEAD~1 -- "**/*_test.go"
```

#### Causes and Solutions

**Cause**: Test expects old output format
```go
// OLD test expecting direct output
func TestOldCommand(t *testing.T) {
    // Capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    runCommand(nil, []string{"test"})

    w.Close()
    os.Stdout = old
    output, _ := io.ReadAll(r)

    assert.Contains(t, string(output), "Expected message")
}

// NEW test using UI framework
func TestNewCommand(t *testing.T) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:    &buf,
        ColorMode: ui.ColorNever,
        Quiet:     false,
        Verbose:   false,
    }

    cmd := &cobra.Command{}
    cmd.SetContext(cli.WithUI(context.Background(), ui.NewOutput(ctx)))

    runCommand(cmd, []string{"test"})

    output := buf.String()
    assert.Contains(t, output, "Expected message")
}
```

**Cause**: UI context not properly set in tests
```go
// Ensure UI context in test helpers
func setupTestCommand(t *testing.T) (*cobra.Command, *bytes.Buffer) {
    var buf bytes.Buffer
    ctx := &ui.UIContext{
        Writer:      &buf,
        ErrorWriter: &buf,
        ColorMode:   ui.ColorNever,
        Quiet:       false,
        Verbose:     false,
        Interactive: false,
    }

    cmd := &cobra.Command{}
    cmd.SetContext(cli.WithUI(context.Background(), ui.NewOutput(ctx)))

    return cmd, &buf
}
```

**Cause**: Output format changed with styling
```go
// Update test expectations for styled output
assert.Contains(t, output, "✓")  // Success symbol
assert.Contains(t, output, "✗")  // Error symbol
assert.Contains(t, output, "ℹ")  // Info symbol
assert.Contains(t, output, "⚠")  // Warning symbol

// Or test without colors
ctx.ColorMode = ui.ColorNever
// Then test plain text expectations
```

## Debugging Procedures

### Enable Debug Mode

#### Environment Variables
```bash
# Enable UI-specific debugging
export PVM_DEBUG=ui

# Enable comprehensive debugging
export PVM_DEBUG=all

# Increase verbosity
export PVM_VERBOSE=1
```

#### Command Line Flags
```bash
# Verbose output
pvm --verbose command

# Debug mode
pvm --debug command

# Combined verbose and debug
pvm --verbose --debug command
```

### Debug Output Analysis

#### UI Context Information
```go
func debugUIContext(ui *ui.Output) {
    if ui.Context().Verbose {
        ui.Debug("=== UI Framework Debug Info ===")
        ui.Debug("Quiet Mode: %v", ui.Context().Quiet)
        ui.Debug("Verbose Mode: %v", ui.Context().Verbose)
        ui.Debug("Color Mode: %v", ui.Context().ColorMode)
        ui.Debug("Interactive: %v", ui.Context().Interactive)
        ui.Debug("Writer Type: %T", ui.Context().Writer)
        ui.Debug("Error Writer Type: %T", ui.Context().ErrorWriter)
        ui.Debug("===============================")
    }
}
```

#### Style Application Debugging
```go
func debugStyles(ui *ui.Output) {
    if ui.Context().Verbose {
        ui.Debug("Testing style rendering:")
        ui.Success("Success style test")
        ui.Error("Error style test")
        ui.Warning("Warning style test")
        ui.Info("Info style test")
        ui.Debug("Debug style test")
    }
}
```

### Performance Profiling

#### CPU Profiling
```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof -bench=. ./internal/cli/ui

# Analyze profile
go tool pprof cpu.prof
```

#### Memory Profiling
```bash
# Generate memory profile
go test -memprofile=mem.prof -bench=. ./internal/cli/ui

# Analyze memory usage
go tool pprof mem.prof
```

#### Command Timing
```bash
# Time command execution
time pvm command args

# Detailed timing with verbose output
time pvm --verbose command args

# Compare with previous versions
time pvm-old command args
time pvm-new command args
```

## Environment-Specific Issues

### Windows-Specific Issues

#### PowerShell Color Support
```powershell
# Enable ANSI color support
$env:FORCE_COLOR = "1"

# Check PowerShell version
$PSVersionTable.PSVersion

# Test Unicode support
Write-Host "✓ ✗ ⚠ ℹ"
```

#### Command Prompt Issues
```cmd
# Enable ANSI escape sequences
reg add HKCU\Console /v VirtualTerminalLevel /t REG_DWORD /d 1

# Use Windows Terminal for better support
wt.exe pvm command
```

### macOS-Specific Issues

#### Terminal App Issues
```bash
# Check terminal capabilities
echo $TERM_PROGRAM

# Test with different terminals
# Terminal.app
pvm version

# iTerm2
pvm version

# VS Code terminal
pvm version
```

#### Font Issues
```bash
# Verify font supports Unicode
echo "✓ ✗ ⚠ ℹ 🐛 → ⚡"

# Recommended fonts:
# - SF Mono
# - Menlo
# - Monaco
# - Fira Code
```

### Linux-Specific Issues

#### Terminal Compatibility
```bash
# Test different terminals
gnome-terminal -e "pvm version"
xterm -e "pvm version"
konsole -e "pvm version"

# Check terminfo
infocmp $TERM
```

#### SSH/Remote Issues
```bash
# Test over SSH
ssh user@host "pvm version"

# Force terminal type
ssh -t user@host "TERM=xterm-256color pvm version"

# Disable colors for scripts
ssh user@host "pvm --color=never version"
```

## Recovery Procedures

### Restore Default Settings

#### Reset UI Configuration
```bash
# Remove any custom configurations
rm -f ~/.pvm/ui-config.toml

# Reset environment variables
unset PVM_DEBUG
unset PVM_VERBOSE
unset FORCE_COLOR
unset NO_COLOR
```

#### Factory Reset UI Framework
```go
// In code, reset to defaults
ui := cli.GetUI(cmd)
ui.SetColorMode(ui.ColorAuto)
ui.SetQuiet(false)
ui.SetVerbose(false)
```

### Emergency Bypass

#### Disable UI Framework Temporarily
```bash
# Use plain output mode
PVM_PLAIN_OUTPUT=1 pvm command

# Or modify code temporarily
export PVM_DEBUG_BYPASS_UI=1
```

#### Fallback to Legacy Output
```go
// Emergency fallback (temporary only)
func emergencyOutput(message string) {
    if os.Getenv("PVM_EMERGENCY_FALLBACK") == "1" {
        fmt.Fprintln(os.Stderr, message)
        return
    }
    // Normal UI framework code
}
```

## Prevention Strategies

### Code Review Checklist

- [ ] No direct `fmt.Print*` or `cmd.Print*` calls
- [ ] Proper UI context initialization
- [ ] Context awareness (quiet/verbose modes)
- [ ] Performance considerations for UI calls
- [ ] Test coverage for UI interactions
- [ ] Error handling uses UI framework

### Automated Testing

#### CI/CD Integration
```yaml
# GitHub Actions example
- name: Test UI Framework
  run: |
    make test
    # Test different output modes
    ./pvm --quiet --help >/dev/null
    ./pvm --verbose --help >/dev/null
    ./pvm --color=never --help >/dev/null
    ./pvm --color=always --help >/dev/null
```

#### Regression Testing
```bash
# Automated UI output testing
./test/ui/test-all-outputs.sh

# Compare with baseline
diff baseline-output.txt current-output.txt
```

### Monitoring

#### Performance Monitoring
```bash
# Monitor command execution times
time pvm command > /dev/null

# Check for performance regressions
benchmark-pvm-commands.sh
```

#### Output Quality Monitoring
```bash
# Verify consistent styling
check-output-consistency.sh

# Test across terminals
test-terminal-compatibility.sh
```

This troubleshooting guide provides comprehensive solutions for common UI framework issues. When encountering problems, start with the quick diagnostic commands and work through the relevant sections based on symptoms observed.
