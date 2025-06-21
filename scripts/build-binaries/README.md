# Binary Build Scripts

This directory contains scripts and configuration files for building Perl binaries across multiple platforms for PVM distribution.

## Overview

The binary build system automates the process of:
- Downloading Perl source code
- Cross-platform compilation with optimized settings
- Creating distribution-ready packages
- Generating checksums for verification
- Supporting multiple platforms and architectures

## Files

- **`build-perl.sh`** - Main build script for automated Perl binary compilation
- **`platforms.json`** - Platform configuration defining supported targets and build settings
- **`build-config.yaml`** - Build configuration with optimization flags and package settings
- **`README.md`** - This documentation file

## Quick Start

### Prerequisites

Ensure these tools are installed:
- `curl` - For downloading Perl source
- `tar` - For archive extraction
- `make` - For building Perl
- `gcc` or `clang` - C compiler
- `jq` - For JSON configuration parsing

### Basic Usage

Build Perl for the current platform:
```bash
./build-perl.sh --version 5.38.2 --platform linux-amd64
```

Build for all supported platforms:
```bash
./build-perl.sh --version 5.38.2 --platform all
```

Build with verbose output and clean build directory:
```bash
./build-perl.sh --version 5.38.2 --platform darwin-arm64 --verbose --clean
```

## Supported Platforms

| Platform | OS | Architecture | Archive Format | Status |
|----------|----|--------------|----- ----------|--------|
| `linux-amd64` | Linux | x86_64 | tar.gz | ✅ Ready |
| `linux-arm64` | Linux | ARM64 | tar.gz | ✅ Ready |
| `darwin-amd64` | macOS | x86_64 | tar.gz | ✅ Ready |
| `darwin-arm64` | macOS | ARM64 | tar.gz | ✅ Ready |
| `windows-amd64` | Windows | x86_64 | zip | ⚠️ Experimental |

## Configuration

### Platform Configuration (`platforms.json`)

Defines platform-specific build settings:
- Compiler and compilation flags
- Archive format (tar.gz vs zip)
- Binary extension (.exe for Windows)
- Cross-compilation toolchain requirements

### Build Configuration (`build-config.yaml`)

Controls build behavior:
- Perl Configure script options
- Optimization levels
- Package contents and exclusions
- Quality assurance tests
- Cache settings

## Build Process

1. **Dependency Check** - Verifies required tools are available
2. **Source Download** - Downloads Perl source from CPAN
3. **Platform Configuration** - Loads platform-specific settings
4. **Configure** - Runs Perl's Configure script with optimized options
5. **Build** - Compiles Perl with parallel jobs
6. **Package** - Creates distribution archive with metadata
7. **Verification** - Generates checksums and runs smoke tests

## Output Structure

Built binaries are placed in `build/binaries/`:
```
build/binaries/
├── perl-5.38.2-linux-amd64.tar.gz
├── perl-5.38.2-linux-amd64.tar.gz.sha256
├── perl-5.38.2-darwin-arm64.tar.gz
├── perl-5.38.2-darwin-arm64.tar.gz.sha256
└── build/
    ├── linux-amd64/
    └── darwin-arm64/
```

Each package contains:
- `bin/` - Perl executables
- `lib/` - Perl libraries and modules
- `man/` - Manual pages
- `PERL_BUILD_INFO` - Build metadata
- `README.txt` - Installation instructions

## Cross-Platform Building

### Linux to Linux ARM64
```bash
# Install cross-compilation toolchain
sudo apt-get install gcc-aarch64-linux-gnu

# Build for ARM64
./build-perl.sh --version 5.38.2 --platform linux-arm64
```

### macOS Universal Builds
The script automatically handles architecture-specific builds on macOS using the appropriate compiler flags.

### Windows Cross-Compilation
Windows builds use MinGW-w64 for cross-compilation from Linux:
```bash
# Install MinGW-w64
sudo apt-get install mingw-w64

# Build for Windows
./build-perl.sh --version 5.38.2 --platform windows-amd64
```

## Testing Builds

### Local Testing
Test a built binary package:
```bash
# Extract and test
tar -xzf build/binaries/perl-5.38.2-linux-amd64.tar.gz
cd perl-5.38.2-linux-amd64
./bin/perl -v
./bin/perl -e 'print "Hello from Perl!\n"'
```

### Smoke Tests
The build script includes built-in smoke tests:
- Version verification
- Basic script execution
- Core module loading
- Standard pragma functionality

## Troubleshooting

### Build Failures

**Configure Fails**
- Check compiler availability (`gcc --version`)
- Verify development headers are installed
- Review platform-specific requirements

**Compilation Errors**
- Ensure sufficient disk space (>2GB recommended)
- Check build dependencies in `platforms.json`
- Try building with `--verbose` for detailed output

**Missing Dependencies**
```bash
# Ubuntu/Debian
sudo apt-get install build-essential curl jq

# macOS
xcode-select --install
brew install jq

# RHEL/CentOS
sudo yum groupinstall "Development Tools"
sudo yum install curl jq
```

### Platform-Specific Issues

**Linux ARM64**
- Requires `gcc-aarch64-linux-gnu` for cross-compilation
- May need `qemu-user-static` for testing built binaries

**macOS**
- Requires Xcode command line tools
- Universal binary support requires macOS 11+ for ARM64

**Windows**
- Cross-compilation from Linux only
- Requires MinGW-w64 toolchain
- May need Wine for testing built binaries

## Performance Optimization

### Build Speed
- Use `parallel_jobs: auto` in build-config.yaml
- Set `run_tests: false` to skip test suite
- Enable build caching in configuration

### Package Size
- Adjust exclusion patterns in build-config.yaml
- Remove unnecessary man pages and documentation
- Use higher compression levels for smaller archives

## Security Considerations

- All downloads use HTTPS with certificate validation
- SHA-256 checksums generated for all packages
- Build environment isolation recommended
- Source code integrity verified against CPAN

## Integration with CI/CD

This build system is designed for integration with GitHub Actions and other CI/CD platforms. See the GitHub Actions workflow configuration for automated binary building.

## Contributing

When adding new platforms or modifying build configurations:

1. Update `platforms.json` with new platform definitions
2. Add platform-specific build requirements
3. Test builds locally before submitting changes
4. Update documentation with new platform support
5. Verify generated packages work on target systems

## Support

For issues with the build system:
- Check this README for troubleshooting guidance
- Review build logs for specific error messages
- Verify platform requirements are met
- Test with verbose output enabled

The build system is designed to be robust and provide clear error messages to help diagnose issues quickly.
