# Release Workflow Documentation

This document describes the automated release process for PVM (Perl Version Manager).

## Overview

The release workflow automatically builds cross-platform binaries for all major platforms (Linux, macOS, Windows) and uploads them as GitHub releases. It also manages version bumping and generates release notes.

## Triggering Releases

### Method 1: Manual Workflow Dispatch (Recommended)

1. Go to the GitHub Actions tab in the repository
2. Select the "Release" workflow
3. Click "Run workflow"
4. Choose the version bump type:
   - **patch**: Bug fixes (0.1.0 → 0.1.1)
   - **minor**: New features (0.1.0 → 0.2.0)
   - **major**: Breaking changes (0.1.0 → 1.0.0)
5. Optionally specify an exact version number (e.g., "1.0.0")
6. Click "Run workflow"

The workflow will:
- Automatically bump the version in `internal/version/version.go`
- Commit the version change
- Create and push a new git tag
- Build binaries for all platforms
- Create a GitHub release with assets

### Method 2: Git Tag Push

You can also trigger a release by pushing a git tag:

```bash
# Manually update the version in internal/version/version.go
git add internal/version/version.go
git commit -m "RELEASE: Bump version to 1.0.0"

# Create and push tag
git tag v1.0.0
git push origin v1.0.0
```

This will trigger the build and release process without the version bumping step.

## Supported Platforms

The workflow builds binaries for:

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

## Components Included

Each release includes:

- **pvm**: Perl Version Manager
- **pvx**: Perl Version Executor
- **pvi**: Perl Version Installer
- **psc**: Perl Static Checker (when CGO build succeeds)

## Build Features

### Version Information

All binaries include build-time version information:
- Version number
- Build timestamp
- Git commit hash
- Go version used
- Target OS/architecture

Access this information with:
```bash
./pvm --version
```

### PSC Build Handling

The PSC component requires CGO and tree-sitter integration, which may fail on some cross-compilation targets. The workflow handles this gracefully:
- Attempts to build PSC for all platforms
- Includes PSC in the release if build succeeds
- Continues with other components if PSC build fails
- Logs clear messages about PSC availability

### Testing

Each build includes basic smoke tests:
- Verifies binaries can execute
- Checks version information display
- Reports any failures clearly

## Release Assets

Each release creates platform-specific archives:

- `pvm-v1.0.0-linux-amd64.tar.gz`
- `pvm-v1.0.0-linux-arm64.tar.gz`
- `pvm-v1.0.0-darwin-amd64.tar.gz`
- `pvm-v1.0.0-darwin-arm64.tar.gz`
- `pvm-v1.0.0-windows-amd64.zip`

Each archive contains:
- All available binaries
- README.md and LICENSE
- BUILD_INFO.txt with build details

## Release Notes

The workflow automatically generates release notes including:
- Changes since the previous release (git log)
- List of included components
- Build information

## Installation Script

The workflow maintains an installation script at `scripts/install.sh` that:
- Detects the user's OS and architecture
- Downloads the appropriate release archive
- Installs PVM components

## Troubleshooting

### Build Failures

If builds fail:
1. Check the Actions logs for specific error messages
2. PSC build failures are often related to CGO/tree-sitter issues
3. Cross-compilation issues may require platform-specific fixes

### Version Conflicts

If version bumping fails:
1. Ensure the current branch is up to date
2. Check for uncommitted changes in `internal/version/version.go`
3. Verify the version format is semantic (major.minor.patch)

### Permission Issues

If the workflow fails to create releases:
1. Verify the repository has the necessary GitHub Actions permissions
2. Check that the GITHUB_TOKEN has release creation permissions

## Manual Release Process

If automation fails, you can create releases manually:

1. Update version in `internal/version/version.go`
2. Run local builds: `make cross-compile`
3. Create archives: `make release`
4. Create GitHub release manually
5. Upload the archives from `build/release/`

## Future Improvements

Planned enhancements:
- Automated changelog generation
- Release candidate builds
- Homebrew formula updates
- Package manager integration
- Code signing for releases
