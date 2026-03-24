# Windows Module Toolchain Support

## Problem

PVM's module build pipeline hardcodes Unix assumptions: `make` as the build
command and `./Build` as a directly executable script. On Windows with
Strawberry Perl, the correct make command is `gmake` (GNU Make bundled with
Strawberry), and `./Build` must be invoked as `perl Build` since Windows
lacks shebang handling.

The dependency checker also assumes MSVC tools (`nmake`, `cl`) rather than
Strawberry Perl's GNU toolchain (`gmake`, `gcc`).

## Approach: Perl's Config.pm Detection

Ask the active Perl what make command it expects:

```
perl -MConfig -e "print $Config{make}"
```

This returns the correct value for any Perl distribution:
- Strawberry Perl → `gmake`
- MSVC-built Perl → `nmake`
- Unix Perl → `make`

This avoids hardcoding tool names or maintaining a mapping table. Perl
already knows its own build requirements.

## Changes

### builder.go: detectMakeCommand

Add a function that queries the active Perl for its make command:

```go
func detectMakeCommand(perlPath string) (string, error) {
    out, err := exec.Command(perlPath, "-MConfig", "-e", `print $Config{make}`).Output()
    if err != nil {
        return "make", err
    }
    return strings.TrimSpace(string(out)), nil
}
```

Call this once at the top of `buildAndInstallModule`, before any commands run.

### builder.go: Replace hardcoded make

Every place that currently uses `"make"` (build, test, install steps for
Makefile.PL modules) uses the detected make command instead.

### builder.go: Fix ./Build on Windows

For Module::Build modules, replace `./Build` with `perl Build` on Windows:

```go
buildCmd := []string{"./Build"}
if runtime.GOOS == "windows" {
    buildCmd = []string{options.PerlPath, "Build"}
}
```

Apply to all three invocations: build, test, install.

### dependencies.go: Update Windows dependencies

Replace the MSVC-only check with Strawberry-aware detection:

```go
case "windows":
    deps = []dependency{
        {name: "gmake", command: "gmake"},
        {name: "gcc",   command: "gcc"},
    }
```

Update install hints to reference Strawberry Perl.

## What stays unchanged

- Environment variable setup (already uses os.PathListSeparator)
- Module download/extraction (already cross-platform)
- CPAN provider (platform-agnostic)
- local::lib isolation (already uses filepath.Join)

## Scope

Two files, ~45 lines of production code changes. This is the first phase of
Windows parity — focused on making `pvm module install` work with Strawberry
Perl. Shell integration (CMD template) and `pvm install` (Strawberry
download) are separate follow-up work.
