# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test Commands
- Build: `go build -o pvm ./cmd/pvm`
- Test all: `go test ./...`
- Test single package: `go test ./path/to/package`
- Test with coverage: `go test -cover ./...`
- Lint: `golangci-lint run`
- Run locally: `go run ./cmd/pvm [commands]`

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
