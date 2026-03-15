# Windows PowerShell Support

## Decision

Add PowerShell shell integration to PVM, enabling Windows as a
first-class platform. The existing shell infrastructure already
supports PowerShell types and detection — this fills in the template
and adapts the E2E test harness.

## Scope

- PowerShell only (no CMD shell integration)
- `.cmd` batch shim wrappers already exist in the codebase
- Full parity with bash template (PATH management, shell function,
  version switching with rehash, conflict detection, fortune quote)
- E2E test harness adapted for Windows (not blanket-skipped)

## Changes

### PowerShell Template (`internal/perl/shell_templates/pvm.ps1`)

Embedded template providing:
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

Add PowerShell case with profile paths:
- `Documents/PowerShell/Microsoft.PowerShell_profile.ps1` (PS 7)
- `Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1` (PS 5.1)
- Init command: `pvm init | Invoke-Expression`

### Version Switching (`internal/perl/shell.go`)

Add PowerShell case to `GenerateShellUse()`:
- `$env:VAR = 'value'` syntax
- `Remove-Item Env:VAR` for unsetting
- `Write-Host` for output
- Update function to use `DetectShell()` instead of raw `SHELL`
  env var

### E2E Test Harness (`test/e2e/helpers/`)

- Use `os.PathListSeparator` instead of hardcoded `:`
- Use `.exe` suffix on Windows for binary names
- Use PowerShell for shell integration setup on Windows
- Remove blanket `TestMain` Windows skip
- Add per-test `skipIfNotUnix` for bash/zsh/fish-specific tests
