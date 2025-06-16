# PVM Build Instructions

This document provides comprehensive instructions for building PVM and its components.

## Prerequisites

### Required Tools
- **Go 1.24.4+**: Download from [golang.org](https://golang.org/dl/)
- **Node.js 20+**: For tree-sitter CLI
- **Git**: For source control
- **Make**: For build automation

### Platform-Specific Requirements

#### Linux
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y build-essential gcc

# For ARM64 cross-compilation
sudo apt-get install -y gcc-aarch64-linux-gnu
```

#### macOS
```bash
# Install Xcode command line tools
xcode-select --install
```

#### Windows
```bash
# Install mingw for CGO support
choco install mingw -y
```

### Tree-sitter Dependencies
```bash
# Install tree-sitter CLI globally
npm install -g tree-sitter-cli
```

## Build Process

### Quick Start
```bash
# Clone the repository
git clone https://github.com/perigrin/pvm.git
cd pvm

# Build everything (development mode)
make

# Run tests
make test
```

### Step-by-Step Build

#### 1. Install Development Tools
```bash
make install-tools
```

This installs Go development tools:
- `stringer`: Code generation
- `moq`: Mock generation
- `gotestsum`: Enhanced test runner
- `staticcheck`: Static analysis
- `govulncheck`: Vulnerability scanning

#### 2. Build Tree-sitter Parser
```bash
make tree-sitter
```

This generates the tree-sitter parser for typed Perl:
- Runs `tree-sitter generate` to create parser.c
- Builds the tree-sitter library with CGO support
- Creates necessary header files for compilation

#### 3. Build PVM Components
```bash
# Development build (with debug symbols)
make build-dev

# Or production build (optimized)
make build-release
```

This creates:
- `build/pvm`: Main binary (includes all components via symlink detection)
- Symlinks: `build/pvi`, `build/pvx`, `build/psc` (created automatically)

## Component Architecture

PVM uses a unified binary architecture:

- **pvm**: Main Perl version manager
- **pvi**: Package installer with dependency management
- **pvx**: Perl script executor with isolation
- **psc**: Perl script compiler with type checking

All components are built into a single `pvm` binary. The other commands are symlinks that are automatically created.

## CGO Configuration

PVM requires CGO for tree-sitter integration. The build system automatically sets:

```bash
export CGO_ENABLED=1
export CGO_CFLAGS=-I$(PWD)/tree-sitter-typed-perl/include -I$(PWD)/tree-sitter-typed-perl -I$(PWD)/tree-sitter-typed-perl/src
```

## Cross-Platform Builds

### Local Cross-Compilation
```bash
# Build for all platforms
make cross-compile

# Create release archives
make release
```

This creates binaries for:
- Linux: AMD64, ARM64
- macOS: ARM64 (Apple Silicon only)
- Windows: AMD64

### Platform-Specific Notes

#### Linux ARM64
- Uses native ARM64 runners for compilation
- Full PSC support with tree-sitter on native ARM64

#### macOS
- Native builds on Apple Silicon only (Intel Macs no longer supported)
- Full PSC support on Apple Silicon

#### Windows
- PSC disabled due to tree-sitter CGO complexity
- Uses MinGW for CGO compilation

## Testing

### Test Suites
```bash
# Quick test (short mode, 10% sampling)
make test

# Full comprehensive test suite
make test-full

# Baseline regression tests
make test-baselines

# Performance benchmarks
make test-performance

# Coverage analysis
make test-coverage
```

### Test Categories
- **Unit tests**: Component-specific functionality
- **Integration tests**: End-to-end workflows
- **Performance tests**: Benchmarking and optimization
- **Baseline tests**: Regression prevention

## Troubleshooting

### Common Issues

#### Tree-sitter Build Failures
**Error**: `tsp_unicode.h: No such file or directory`

**Solution**: The required header file is committed to the repository. Ensure you have the latest code:
```bash
git pull origin pu
make clean
make tree-sitter
```

#### CGO Compilation Issues
**Error**: `C compiler "gcc" not found`

**Solution**: Install build tools for your platform:
```bash
# Linux
sudo apt-get install build-essential

# macOS
xcode-select --install

# Windows
choco install mingw
```

#### Missing Go Tools
**Error**: Tool not found (stringer, moq, etc.)

**Solution**: Install development tools:
```bash
make install-tools
```

### Build Performance

#### Parallel Builds
The test system automatically detects CPU count for optimal parallelization:
```bash
# Uses all available CPU cores
make test
```

#### Build Monitoring
```bash
# Monitor build performance
make build-monitor

# View performance report
make performance-report
```

## Development Workflow

### Setting Up Development Environment
```bash
# Complete development setup
make setup

# Verify everything works
make test-baselines
```

### Making Changes
```bash
# Before making changes, ensure clean state
make clean
make test

# Make your changes...

# Test changes
make test
make lint

# Build for release testing
make build-release
```

### Code Quality
```bash
# Format code
make fmt

# Static analysis
make lint

# Security scanning
make security

# Full quality check
make check
```

## CI/CD Integration

The project includes GitHub Actions workflows for:

- **Continuous Integration**: Testing on multiple platforms
- **Release Automation**: Cross-platform binary builds
- **Security Scanning**: Vulnerability and security analysis

See `.github/workflows/` for workflow definitions.

## Advanced Features

### Performance Optimization
```bash
# Run performance analysis
make profile

# Optimization testing
make optimize

# Complete performance analysis
make performance-analysis
```

### Memory Management
The system includes comprehensive memory profiling and optimization tools in the `internal/performance` package.

## Support

- **Documentation**: See project README and inline code documentation
- **Issues**: Report bugs via [GitHub Issues](https://github.com/perigrin/pvm/issues)
- **Discussions**: Join the conversation in [GitHub Discussions](https://github.com/perigrin/pvm/discussions)

## Version Information

Built binaries include version metadata:
```bash
./build/pvm version
# Output includes: version, build time, commit hash
```

This information is automatically injected during the build process.
