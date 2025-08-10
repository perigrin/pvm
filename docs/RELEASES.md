# PVM Release Process

This document explains how to create releases for the PVM project using the automated GitHub Actions workflow.

## Release Types

The release system supports two types of releases:

### 1. Automatic Releases (Tag-triggered)

When you push a Git tag, the release workflow automatically triggers:

```bash
# Create and push a release tag
git tag v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

**Pre-release Detection**: Tags matching these patterns are automatically marked as pre-releases:
- `v*-rc*` (e.g., `v1.0.0-rc1`, `v1.0.0-rc2`)
- `v*-alpha*` (e.g., `v1.0.0-alpha1`)
- `v*-beta*` (e.g., `v1.0.0-beta1`)

All other tags are treated as stable releases.

### 2. Manual Releases (Workflow Dispatch)

You can manually trigger a release from the GitHub Actions tab:

1. Go to **Actions** → **Release** workflow
2. Click **Run workflow**
3. Enter the tag name (e.g., `v1.0.0-rc3`)
4. Choose whether to mark as pre-release
5. Click **Run workflow**

## Supported Platforms

The release workflow builds binaries for:

| Platform | Architecture | PSC Support | Archive Format |
|----------|-------------|-------------|----------------|
| Linux | AMD64 | ✅ Yes | `.tar.gz` |
| Linux | ARM64 | ❌ No* | `.tar.gz` |
| macOS | AMD64 (Intel) | ✅ Yes | `.tar.gz` |
| macOS | ARM64 (Apple Silicon) | ✅ Yes | `.tar.gz` |
| Windows | AMD64 | ❌ No* | `.zip` |

*PSC (Perl Script Compiler) is not available on these platforms due to tree-sitter CGO compilation complexity.

## Components Included

Each release includes these components where supported:

- **pvm**: Main Perl version manager
- **pm**: Package installer with dependency management
- **pvx**: Perl script executor with isolation
- **psc**: Perl script compiler with type checking (Linux/macOS only)

## Release Workflow Details

### Build Matrix Strategy

The workflow uses a build matrix to compile for all platforms in parallel:

```yaml
strategy:
  fail-fast: false
  matrix:
    include:
      - goos: linux
        goarch: amd64
        runner: ubuntu-latest
        build_psc: true
      # ... other platforms
```

### Build Process

1. **Setup**: Install Go, Node.js, and tree-sitter CLI
2. **Dependencies**: Install platform-specific build tools
3. **Tree-sitter**: Build the typed Perl grammar (if PSC supported)
4. **Compilation**: Build all components with version metadata
5. **Testing**: Test binaries on native platforms
6. **Archiving**: Create platform-specific archives
7. **Release**: Upload to GitHub Releases with generated notes

### Version Metadata

Each binary includes build metadata:

```bash
./pvm version
# Output: pvm 1.0.0-rc2
# Built: 2025-06-15T23:31:00Z
# Commit: 80b6e82...
```

This is injected during build via ldflags:

```bash
-X 'tamarou.com/pvm/internal/version.Version=${VERSION}'
-X 'tamarou.com/pvm/internal/version.BuildTime=${BUILD_TIME}'
-X 'tamarou.com/pvm/internal/version.CommitHash=${COMMIT}'
```

## Release Notes Generation

The workflow automatically generates comprehensive release notes including:

- **Changelog**: Git commit messages since previous tag
- **Installation instructions**: Platform-specific download/setup
- **Component descriptions**: What each binary does
- **Platform notes**: PSC availability and limitations
- **Support links**: Documentation and issue reporting

## Testing a Release

After the workflow completes, you can test the release:

```bash
# Download the appropriate archive for your platform
wget https://github.com/perigrin/pvm/releases/download/v1.0.0-rc2/pvm-1.0.0-rc2-linux-amd64.tar.gz

# Extract and test
tar -xzf pvm-1.0.0-rc2-linux-amd64.tar.gz
chmod +x pvm-* pm-* pvx-*
./pvm-linux-amd64 version
```

## Troubleshooting

### Common Issues

1. **Tree-sitter build failures**: Usually platform/architecture specific
   - Check the build matrix `build_psc` flag
   - PSC is disabled on problematic platforms

2. **Cross-compilation errors**: ARM64 builds can be complex
   - Linux ARM64 builds on Ubuntu with cross-compilation tools
   - May need to adjust CGO settings for specific platforms

3. **Archive creation failures**: Different tools on different platforms
   - Unix: `tar -czf`
   - Windows: `Compress-Archive` PowerShell cmdlet

### Manual Release Recovery

If the automated release fails, you can create a manual release:

1. **Build locally**: Use `make cross-compile` for your platform
2. **Create archives**: Package binaries manually
3. **Upload manually**: Use GitHub web interface to create release

### Workflow Debugging

To debug the release workflow:

1. **Check Actions logs**: GitHub Actions tab shows detailed build logs
2. **Download artifacts**: Failed builds still upload artifacts for inspection
3. **Test locally**: Use the same commands from the workflow locally

## Best Practices

### Tagging Strategy

- **Stable releases**: `v1.0.0`, `v1.1.0`, `v2.0.0`
- **Release candidates**: `v1.0.0-rc1`, `v1.0.0-rc2`
- **Development builds**: `v1.0.0-alpha1`, `v1.0.0-beta1`

### Release Timing

- **RC releases**: For testing new features
- **Stable releases**: After RC testing completes
- **Patch releases**: For critical bug fixes

### Version Bumping

Update the version in relevant files before tagging:

```bash
# The workflow automatically sets version from tag
# No manual version file updates needed
git tag v1.1.0 -m "Release v1.1.0"
```

## Security Considerations

- **Build reproducibility**: Builds are deterministic with version metadata
- **Artifact signing**: Future enhancement for binary signing
- **Dependency scanning**: Included in CI workflow
- **Token security**: Uses `GITHUB_TOKEN` with minimal required permissions

The release system provides a robust, automated way to create cross-platform releases with minimal manual intervention while maintaining quality and security standards.
