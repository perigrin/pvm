# Windows PowerShell Support

## Status: Partially Complete

PowerShell shell integration code is implemented and working.
Windows CI and test porting is deferred — see "Lessons Learned" below.

## What Was Implemented

### PowerShell Template (`internal/perl/shell_templates/pvm.ps1`)

Embedded template with full bash parity:
- PATH cleanup and shim directory prepend
- `pvm` shell function that intercepts version-switching commands
  and rehashes PATH afterward
- `_pvm_update_perl_path` function for PATH rebuilding
- Conflict detection (plenv/perlbrew in PATH)
- Fortune quote on init

### Shell Detection (`internal/perl/shell.go`)

Environment-first detection on all platforms:
1. Check `PSModulePath` env var (PowerShell on any OS, including
   pwsh on Linux/macOS)
2. Check `SHELL` env var (bash/zsh/fish)
3. OS-based fallback (CMD on Windows, bash on Unix)

### Shell Config (`internal/shell/config.go`)

PowerShell case with profile paths:
- `Documents/PowerShell/Microsoft.PowerShell_profile.ps1` (PS 7)
- `Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1` (PS 5.1)
- Init command: `pvm init | Invoke-Expression`

### Version Switching (`internal/perl/shell.go`)

PowerShell case in `GenerateShellUse()`:
- `$env:VAR = 'value'` syntax
- `Remove-Item Env:VAR` for unsetting
- `Write-Host` for output
- Uses `DetectShell()` instead of raw `SHELL` env var

### Path Normalization (`internal/config/accessors.go`)

`expandEnvironmentVariables` applies `filepath.Clean` after variable
expansion to normalize mixed path separators on Windows.

## What Is Deferred: Windows CI and Test Porting

Windows was added to the CI matrix but removed after discovering
extensive test failures. Windows CI is deferred until the test
suite is properly ported.

## Lessons Learned

Running the Windows CI revealed these categories of issues that
need to be addressed before Windows CI can be enabled:

### 1. Path separator assumptions

Many tests construct expected paths with hardcoded `/`. Production
code using `filepath.Clean` or `filepath.Join` produces `\` on
Windows, causing assertion mismatches. Fix: use `filepath.FromSlash`
in test expectations.

### 2. Binary naming

Tests create mock executables named `pvm`, `perl`, etc. Windows
requires `.exe` suffix. The production code (`shim.go`) already
handles this, but test code does not.

### 3. Unix-only test patterns

Several tests depend on Unix-specific features:
- Bash scripts as mock executables (won't run on Windows)
- `chmod` for permission testing (Windows ACLs work differently)
- `/tmp` as a path prefix (not valid on Windows)
- Hardcoded `:` as PATH separator

### 4. PSModulePath on CI runners

GitHub Actions Linux runners have PowerShell installed, which sets
`PSModulePath`. This causes `DetectShell()` to return PowerShell
instead of bash. Tests that check bash-specific `pvm init` output
need `ForceBashDetection` to clear `PSModulePath`.

### 5. Windows short path names (8.3)

`os.TempDir()` on Windows can return 8.3 short names (e.g.,
`RUNNER~1` instead of `runneradmin`). Path comparisons fail unless
both sides are normalized with `filepath.EvalSymlinks`.

### 6. Archive extraction

The updater's archive validation looks for platform-specific binary
names (`pvm.exe` on Windows). Test archives need to include the
correct filename.

## Recommended Approach for Future Windows CI

1. Create a `helpers.SkipOnWindows(t, "reason")` helper
2. Tag each failing test with `SkipOnWindows` and a specific reason
3. Fix tests in batches by category (paths, binary names, permissions)
4. Enable Windows CI when the skip count is low enough
5. Consider a separate `windows_test.go` for Windows-specific tests
