# Windows Strawberry Perl Install Flow

## Problem

`pvm install 5.x.x` does not work on Windows. The PVM binary mirror
(`github.com/perigrin/pvm/releases`) hosts pre-built Perl binaries for
Linux and macOS but not Windows. Windows users can only import existing
Perl installations.

## Approach: Strawberry Perl as Binary Provider

On Windows, when no PVM binary is available, automatically download
Strawberry Perl's portable .zip from their GitHub releases. This is
transparent to the user — `pvm install 5.38.0` works the same on all
platforms. The `--mirror` flag overrides this behavior.

## Components

### 1. GitHub Release Discovery (strawberry.go)

Query the GitHub API for Strawberry Perl releases:

```
GET https://api.github.com/repos/StrawberryPerl/Perl-Dist-Strawberry/releases
```

Find the release asset matching the requested Perl version. Strawberry
uses 4-part versions (5.38.2.2) while Perl uses 3-part (5.38.2). Match
on the first three components and pick the latest 4th component.

Asset filename pattern: `strawberry-perl-{version}-64bit-portable.zip`

Cache the release list for the session to avoid repeated API calls.

### 2. Install Flow Hook (command.go)

After `CheckBinaryAvailability` returns false on Windows:

```go
if !available && runtime.GOOS == "windows" && mirror == "" {
    strawberryURL, err := perl.FindStrawberryRelease(version)
    if err == nil {
        available = true
        binaryOptions.StrawberryURL = strawberryURL
    }
}
```

### 3. Download and Extraction (install_binary.go)

Reuse existing download infrastructure (caching, progress, retry, ZIP
extraction). When `StrawberryURL` is set, download from that URL instead
of the PVM mirror.

### 4. Layout Relocation (strawberry.go)

After extraction, Strawberry's layout differs from PVM's expectation:

```
Strawberry:              PVM expects:
├── perl/                ├── bin/perl.exe
│   ├── bin/perl.exe     ├── lib/
│   ├── lib/             └── c/  (toolchain)
│   └── site/
├── c/
│   ├── bin/gcc.exe
│   └── bin/gmake.exe
└── ...
```

Relocate after extraction:
- Move `perl/bin/` → `bin/`
- Move `perl/lib/` → `lib/`
- Move `perl/site/` → `site/` (if exists)
- Keep `c/` at root (toolchain for XS modules)
- Remove empty `perl/` directory

### 5. Registration

Register with `Source: "strawberry"` to distinguish from PVM-built or
system-imported Perls. Existing `pvm list`, `pvm use`, `pvm uninstall`
work unchanged.

## User Experience

```
$ pvm install 5.38.0
Checking for pre-built binary...
No PVM binary available for windows-amd64
Downloading Strawberry Perl 5.38.0.1 portable...
████████████████████████████████ 100% (85.2 MB)
Extracting...
Relocating Strawberry Perl layout...
Verifying installation...
✓ Perl 5.38.0 installed successfully (source: Strawberry Perl)
```

## What stays unchanged

- `pvm list` / `pvm use` / `pvm current` — registry-based, unaffected
- `pvm uninstall` — removes version directory, unaffected
- `--mirror` flag — skips Strawberry, uses custom mirror
- Linux/macOS install flow — completely unaffected
- Manual URL import — follow-up work, not in this phase

## Scope

One new file (`strawberry.go` + tests), minor additions to `command.go`
and `install_binary.go`. The download, caching, and extraction
infrastructure already exists and handles Windows ZIP files.
