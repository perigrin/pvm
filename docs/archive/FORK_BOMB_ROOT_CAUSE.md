# Fork Bomb Root Cause Analysis and Fix

## Issue Summary
`pvm tool add` command was causing a fork bomb that created hundreds of recursive pvm processes, requiring system restarts to recover.

## Root Cause Analysis

### The Problem Chain
1. **Recent Change**: Commit `db1880b` (Aug 5, 2025) changed tool installation from direct installer to PVX isolated environments
2. **PVX Integration**: Tool installation now calls `pvx.ExecuteScript()` with a Perl script that runs `system('cpanm', $module)`
3. **Environment Shims**: PVX creates isolated environments with shims in `binDir` for common tools including `cpanm`
4. **PATH Pollution**: The isolated `binDir` is prepended to PATH, causing `system('cpanm')` to find the PVX shim instead of real cpanm
5. **Recursive Loop**: The cpanm shim calls `pvm pvx` which creates new isolated environments with more shims
6. **Fork Bomb**: Each recursive call spawns more pvm processes exponentially

### Technical Details

**File**: `internal/pvx/executor.go`
**Function**: `generateEnvironmentShims()`
**Line**: 1433

The function creates shims for:
```go
commands := []string{"perl", "cpan", "prove", "perldoc", "h2ph", "h2xs", "enc2xs", "xsubpp", "corelist", "cpanm"}
```

**Shim Content** (line 1462):
```bash
exec "$PVM_EXEC" pvx --name "$env_name" --isolation=local --no-cleanup -e 'exec { "cpanm" } "cpanm", @ARGV;' -- "$@"
```

**The Recursion**:
1. `pvm tool add ack` → PVX execution with EnvName: "tool-ack"  
2. PVX creates cpanm shim in isolated binDir
3. Installation script calls `system('cpanm', 'App::Ack')`
4. PATH finds cpanm shim → calls `pvm pvx` recursively
5. Infinite recursion = fork bomb

## The Fix

**File**: `internal/pvx/executor.go`
**Lines**: 1437-1449

Added logic to exclude cpanm shim creation for tool installation environments:

```go
// Skip cpanm shim for tool installation environments to prevent fork bomb
// Tool installations use cpanm directly and creating a shim causes recursive pvm calls
if strings.HasPrefix(options.EnvName, "tool-") {
    // Filter out cpanm from commands for tool installation environments
    filteredCommands := []string{}
    for _, cmd := range commands {
        if cmd != "cpanm" {
            filteredCommands = append(filteredCommands, cmd)
        }
    }
    commands = filteredCommands
    if options.Verbose {
        log.Infof("Excluded cpanm shim for tool installation environment '%s' to prevent fork bomb", options.EnvName)
    }
}
```

## Fix Validation

### Why This Fix Works
1. **Targeted**: Only affects tool installation (EnvName prefix "tool-")
2. **Minimal**: Only excludes cpanm shim, preserves all other functionality
3. **Safe**: Tool installations can still use real cpanm from system PATH
4. **Preserves Features**: Other environments still get cpanm shims when appropriate

### What Still Works
- All other shims (perl, cpan, prove, etc.) are still created
- Non-tool environments still get cpanm shims
- Tool installation functionality preserved
- PVX isolation functionality preserved

### Tests
- Existing PVX tests pass
- Command help functionality works
- No regression in related functionality

## Prevention
This issue was introduced by a significant architectural change (direct installer → PVX isolation) without adequate testing of the recursive scenarios. Future changes to tool installation should:

1. Test with actual tool installations, not just unit tests
2. Consider environment pollution and PATH manipulation effects  
3. Test for recursive subprocess calls
4. Use process monitoring during testing

## Files Modified
1. `internal/pvx/executor.go` - Added cpanm shim exclusion logic
2. `internal/pvm/command.go` - Re-enabled tool add command (was temporarily disabled)

---
*Analysis completed: August 6, 2025*
*Fix applied and tested successfully*