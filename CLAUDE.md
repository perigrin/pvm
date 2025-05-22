# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test Commands

### Basic Commands
- Build all: `make all` (builds all components including tree-sitter)
- Build individual components:
  - PVM: `make pvm` or `go build -o build/pvm ./cmd/pvm`
  - PVX: `make pvx` or `go build -o build/pvx ./cmd/pvx`
  - PVI: `make pvi` or `go build -o build/pvi ./cmd/pvi`
  - PSC: `make psc` (requires tree-sitter build first)
- Test all: `go test ./...`
- Test single package: `go test ./path/to/package`
- Test with coverage: `go test -cover ./...`
- Lint: `golangci-lint run`
- Clean: `make clean`

### Cross-Platform Build
- Cross-compile all platforms: `make cross-compile`
- Create release archives: `make release`

Supported platforms:
- Linux (AMD64, ARM64)
- macOS (AMD64, ARM64)
- Windows (AMD64)

### Tree-sitter Build (Required for PSC)
PSC uses tree-sitter-perl with custom type annotation extensions. The build process:

1. **Prerequisites**: Node.js and npm must be installed for tree-sitter-cli
2. **Build tree-sitter**: `make tree-sitter` or `./bin/build_tree_sitter.sh`

This process:
- Clones the official tree-sitter-perl grammar from GitHub
- Integrates our custom type annotation grammar extensions
- Generates the extended parser using tree-sitter-cli
- Builds the shared library for go-tree-sitter integration

### PSC-Specific Build Issues
PSC requires tree-sitter integration which has additional dependencies:
- Tree-sitter C library headers
- Extended perl grammar with type annotations
- CGO build flags for header includes

**Current Status**: Tree-sitter build completes successfully but Go build integration still needs work on CGO path resolution.

## GitHub Workflows

The project includes automated CI/CD workflows:

### Build Workflow (`.github/workflows/build.yml`)
- Triggers on pushes to main/develop and pull requests
- Tests across Ubuntu, Windows, and macOS
- Builds tree-sitter library for each platform
- Runs comprehensive tests and linting
- Uploads build artifacts for each platform

### Quick Test Workflow (`.github/workflows/test.yml`)
- Triggers on feature branch pushes and PRs
- Fast feedback with basic tests and linting
- Validates documentation and markdown

### Release Workflow (`.github/workflows/release.yml`)
- Triggers on version tags (v*.*.*)
- Cross-compiles for all supported platforms
- Creates GitHub releases with downloadable archives
- Includes both native builds and shared libraries

## Project Architecture

The PVM Ecosystem is a comprehensive suite of tools for Perl development environment management built as a single Go binary with multiple entry points:

- **PVM (Perl Version Manager)** - Manages Perl installations and versions
- **PVX (Perl Version eXecutor)** - Executes modules/scripts in isolated environments
- **PVI (Perl Version Installer)** - Manages CPAN modules for installed Perl versions
- **PSC (Perl Script Compiler)** - Provides static type checking for Perl code

### Core Components
- **Configuration System**: TOML-based with XDG Base Directory layering (project, user, system)
- **Version Manager**: Handles Perl installation, detection, and versioning
- **Environment Manager**: Manages execution environments with isolation
- **Module Manager**: Handles CPAN integration and dependency resolution
- **Type System**: Implements gradual typing for Perl with flow-sensitive analysis

### Implementation Phases
1. **Foundation Layer** - Core infrastructure and basic functionality
2. **Environment Management** - Perl installation and execution environment
3. **Module Management** - CPAN integration and module operations
4. **Type System** - Type checking and advanced integrations

## Code Organization

- `cmd/` - Entry points for each component (pvm, pvx, pvi, psc)
- `internal/` - Internal packages not meant for external use
- `pkg/` - Public packages that may be used by other tools
- `test/` - Test fixtures and helpers

## Integration Points

The components work together with well-defined integration points:

- PVM provides version resolution for PVX execution environments
- PVX uses PVI to install required modules in isolated environments
- PSC can pass type-checked code to PVX for execution
- PVI and PSC work together for type definition management

## Configuration System

Configuration uses TOML format with layered files:

- Project configuration: `.pvm/pvm.toml`
- User configuration: `$XDG_CONFIG_HOME/pvm/pvm.toml`
- System configuration: `/etc/pvm/pvm.toml`

All files are optional with the system working with sensible defaults.

## Development Guidelines

- Follow Test-Driven Development approach - write tests before implementation
- Ensure cross-platform compatibility (Linux, macOS, Windows)
- Handle errors with structured error codes and detailed messages
- Maintain backward compatibility with existing Perl version managers
- Optimize for both performance and usability

## Golang Best Practices

- use Testify for testing
- use gofumpt for formatting
- use golangci-lint for linting

## Go Formatting and Linting Issues

### Common Linting Issues

1. **Error Handling (errcheck)**
   - Always check errors from:
     - File operations: `os.Remove`, `os.RemoveAll`, `file.Close`
     - Environment modifications: `os.Chdir`, `os.Setenv`, `os.Unsetenv`
     - I/O operations: `fmt.Fprintf`, `resp.Body.Close`, `w.Write`

2. **Variable Usage (ineffassign)**
   - Avoid unused variable assignments
   - Either use the variable or remove the assignment

3. **Code Structure (go-critic)**
   - Convert if-else chains to switch statements where appropriate

### Best Practices for Error Handling

1. **Deferred Functions**
   ```go
   // Correct way to handle errors in deferred calls
   defer func() {
       _ = file.Close()
   }()
   ```

2. **Direct Function Calls**
   ```go
   // When error handling isn't necessary but linter requires check
   _ = os.Remove(tmpPath)
   ```

3. **Test Files**
   - In test code, ensure environment cleanup with proper error handling:
   ```go
   defer func() { _ = os.RemoveAll(tempDir) }()
   ```

### Pre-commit Hooks

The project uses the following pre-commit hooks:

1. **go vet** - Analyzes code for potential errors
   - Runs `go vet ./...` on the entire module
   - Ensures code correctness before committing

2. **Automated Fixes**
   - `fix_error_checks.sh` - Corrects common error handling issues
   - `fix_linting.sh` - Addresses switch statement formatting
   - `fix_indentation.sh` - Fixes indentation in specific files

### Running Linters

```bash
# Run all linting checks
golangci-lint run

# Run specific checks
golangci-lint run --disable-all --enable=errcheck
golangci-lint run --disable-all --enable=ineffassign
golangci-lint run --disable-all --enable=staticcheck
golangci-lint run --disable-all --enable=go-critic
```

For more details on linting issues and their resolutions, see `docs/linting-issues.md`.
