# Build-Install Workflow Guide

## Overview

PVM now supports a flexible build-install workflow that separates Perl building from installation, enabling binary distribution, relocatable builds, and custom mirror configurations. This system maintains full backward compatibility while adding powerful new capabilities.

## Key Features

- **Relocatable Builds**: Perl installations that can be moved between systems
- **Build-Only Mode**: Build Perl without installing it into PVM's version management
- **Install from Build**: Install Perl from build directories
- **Archive Support**: Create and install from portable tar.gz archives
- **Custom Mirrors**: Configure custom binary mirrors with authentication
- **Automated Upload**: Integrate with CI/CD for automated binary publishing

## Commands Overview

### build-perl (Enhanced)

The `build-perl` command now supports additional flags for the build-install workflow:

```bash
# Traditional behavior (build AND install) - unchanged
pvm build-perl 5.38.0

# Build-only mode (no installation)
pvm build-perl 5.38.0 --build-only

# Build to custom output directory
pvm build-perl 5.38.0 --build-only --output-dir /path/to/output

# Build and upload to configured mirror
pvm build-perl 5.38.0 --upload
```

#### Flags

- `--build-only`: Build Perl without installing it into PVM's version management
- `--output-dir DIR`: Specify custom output directory for the build
- `--upload`: Upload the built binary to configured mirrors (requires --build-only)

### install-perl (New)

The new `install-perl` command installs Perl from various sources:

```bash
# Install from build directory
pvm install-perl --from-build /path/to/build --version 5.38.0-custom

# Install from archive
pvm install-perl /path/to/perl-5.38.0.tar.gz --version 5.38.0-archive

# Install from URL
pvm install-perl https://example.com/perl-5.38.0.tar.gz --version 5.38.0-remote

# Auto-detect version from build directory
pvm install-perl --from-build /path/to/build
```

#### Flags

- `--from-build DIR`: Install from a build directory
- `--version VERSION`: Specify version name for PVM's version management
- `--force`: Force installation even if version already exists
- `--mirror MIRROR`: Override default mirror for URL downloads

### upload-binary (New)

Upload built binaries to GitHub releases or custom mirrors:

```bash
# Upload to GitHub releases
pvm upload-binary /path/to/build --version 5.38.0 --platform linux-amd64

# Upload archive
pvm upload-binary /path/to/perl-5.38.0.tar.gz --version 5.38.0 --platform linux-amd64

# Upload to custom mirror
pvm upload-binary /path/to/build --version 5.38.0 --platform linux-amd64 --mirror custom-mirror
```

#### Flags

- `--version VERSION`: Specify version for upload
- `--platform PLATFORM`: Specify platform (e.g., linux-amd64, darwin-arm64)
- `--archive`: Create and upload archive instead of directory
- `--mirror MIRROR`: Upload to specific mirror
- `--github-token TOKEN`: GitHub authentication token
- `--force`: Force upload even if version exists

## Workflow Examples

### Binary Distribution Workflow

1. **Build relocatable Perl**:
   ```bash
   pvm build-perl 5.38.0 --build-only --output-dir ./perl-5.38.0-linux-amd64
   ```

2. **Create distribution archive**:
   ```bash
   tar -czf perl-5.38.0-linux-amd64.tar.gz -C ./perl-5.38.0-linux-amd64 .
   ```

3. **Distribute and install on target system**:
   ```bash
   pvm install-perl perl-5.38.0-linux-amd64.tar.gz --version 5.38.0
   ```

### CI/CD Integration

1. **Automated build and upload**:
   ```bash
   pvm build-perl 5.38.0 --upload
   ```

2. **Platform matrix build**:
   ```bash
   # In GitHub Actions or similar CI system
   for platform in linux-amd64 darwin-amd64 darwin-arm64; do
     pvm build-perl 5.38.0 --build-only --output-dir perl-5.38.0-$platform
     pvm upload-binary perl-5.38.0-$platform --version 5.38.0 --platform $platform
   done
   ```

### Development Workflow

1. **Build for testing**:
   ```bash
   pvm build-perl 5.38.0 --build-only --output-dir ./test-build
   ```

2. **Test the build**:
   ```bash
   ./test-build/bin/perl -v
   ./test-build/bin/perl -e "use strict; use warnings; print 'OK\\n'"
   ```

3. **Install if satisfied**:
   ```bash
   pvm install-perl --from-build ./test-build --version 5.38.0-test
   ```

## Configuration

### Custom Mirrors

Configure custom binary mirrors in your PVM configuration file:

```toml
[pvm.binary]
  # Enable custom mirrors
  enable_custom_mirrors = true

  # Default mirror priority
  mirror_priority = ["custom-primary", "github-releases", "cpan-mirror"]

[[pvm.binary.custom_mirrors]]
  name = "custom-primary"
  url_template = "https://binaries.example.com/perl/{version}/{platform}/perl-{version}-{platform}.tar.gz"
  priority = 1

  # Authentication configuration
  [pvm.binary.custom_mirrors.auth]
    type = "bearer"
    token = "${CUSTOM_MIRROR_TOKEN}"

[[pvm.binary.custom_mirrors]]
  name = "enterprise-mirror"
  url_template = "https://enterprise.internal/perl/{version}/perl-{platform}.tar.gz"
  priority = 2

  [pvm.binary.custom_mirrors.auth]
    type = "basic"
    username = "${ENTERPRISE_USER}"
    password = "${ENTERPRISE_PASS}"

# Version-specific mappings
[pvm.binary.version_mappings]
  "5.38.0" = "5.38.0-custom-build-1"
  "5.36.0" = "5.36.0-patched"
```

### Authentication Types

Custom mirrors support various authentication methods:

- **None**: No authentication required
- **Basic**: Username/password authentication
- **Bearer**: Bearer token authentication
- **API Key**: Custom API key in headers
- **OAuth2**: OAuth2 client credentials flow

### URL Templates

URL templates support variable substitution:

- `{version}`: Perl version (e.g., "5.38.0")
- `{platform}`: Platform identifier (e.g., "linux-amd64")
- `{arch}`: Architecture (e.g., "amd64", "arm64")
- `{os}`: Operating system (e.g., "linux", "darwin")

## Relocatable Builds

PVM builds are relocatable by default, meaning they can be moved to different directories or systems and continue to function correctly. This is achieved through:

1. **Relocatable Inc**: Built with `-Duserelocatableinc` Configure option
2. **Relative Paths**: @INC paths are calculated relative to the Perl binary location
3. **Self-Contained**: All dependencies are included in the build directory

### Testing Relocatability

```bash
# Build Perl
pvm build-perl 5.38.0 --build-only --output-dir ./perl-original

# Test original location
./perl-original/bin/perl -v

# Move to different location
mv ./perl-original ./perl-moved

# Test moved location
./perl-moved/bin/perl -v
./perl-moved/bin/perl -e "print join('\\n', @INC)"
```

## Troubleshooting

### Common Issues

1. **Permission Denied**:
   ```
   Error: failed to create output directory: permission denied
   ```
   Solution: Ensure you have write permissions to the output directory.

2. **Invalid Archive**:
   ```
   Error: failed to extract archive: not a valid tar.gz file
   ```
   Solution: Verify the archive file is a valid tar.gz archive.

3. **Missing Perl Binary**:
   ```
   Error: build directory does not contain a valid Perl installation
   ```
   Solution: Ensure the build directory contains a complete Perl installation with bin/perl.

4. **Version Already Exists**:
   ```
   Error: version 5.38.0 already exists
   ```
   Solution: Use `--force` flag or choose a different version name.

### Debug Information

Use the `--debug` flag for detailed information:

```bash
pvm --debug build-perl 5.38.0 --build-only
pvm --debug install-perl --from-build ./build-dir
```

### Log Files

Build and installation logs are stored in:
- `$XDG_CACHE_HOME/pvm/logs/build-perl.log`
- `$XDG_CACHE_HOME/pvm/logs/install-perl.log`

## Migration from Previous Versions

The new build-install workflow is fully backward compatible:

### Existing Workflows

All existing commands continue to work unchanged:

```bash
# These commands work exactly as before
pvm build-perl 5.38.0        # Still builds AND installs
pvm install 5.38.0           # Still installs from binary/source
pvm list                     # Still shows installed versions
```

### New Capabilities

The new workflow adds capabilities without breaking existing functionality:

```bash
# New: Build without installing
pvm build-perl 5.38.0 --build-only

# New: Install from build directory
pvm install-perl --from-build /path/to/build

# New: Install from archive
pvm install-perl archive.tar.gz
```

## Performance Considerations

### Build Performance

- **Parallel Builds**: Use `-j$(nproc)` for parallel compilation
- **Cache Directories**: Reuse build caches when possible
- **Output Directory**: Use fast storage (SSD) for build output

### Distribution Performance

- **Archive Compression**: tar.gz provides good compression for distribution
- **Mirror Selection**: Choose geographically close mirrors
- **Authentication Caching**: OAuth2 tokens are cached to reduce overhead

### Installation Performance

- **Local Sources**: Installing from local directories is fastest
- **Archive Extraction**: Parallel extraction when supported
- **Version Registration**: Minimal overhead for PVM integration

## GitHub Actions Integration

A complete GitHub Actions workflow is available at `.github/workflows/build-perl-binaries.yml`:

```yaml
name: Build Perl Binaries

on:
  workflow_dispatch:
    inputs:
      perl_version:
        description: 'Perl version to build'
        required: true
        default: 'latest2'

jobs:
  build:
    strategy:
      matrix:
        platform: [linux-amd64, darwin-amd64, darwin-arm64]

    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build PVM
        run: make

      - name: Build and Upload Perl
        run: ./pvm build-perl ${{ github.event.inputs.perl_version }} --upload
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

This workflow builds Perl binaries for multiple platforms and automatically uploads them to GitHub releases, making them available for community use.
