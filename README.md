# PVM Ecosystem

The PVM Ecosystem provides a comprehensive suite of tools for Perl development with modern, TypeScript-quality tooling. Built with a modernized architecture inspired by TypeScript-Go patterns, the system delivers enhanced performance, symbol-aware language server features, and production-quality development experience.

## Core Components

- **PVM (Perl Version Manager)** - Manages Perl installations and versions
- **PSC (Perl Script Compiler)** - Static type checker with enhanced LSP and symbol-aware features
- **PVI (Perl Version Installer)** - Module installer with type-aware dependency management
- **PVX (Perl Version eXecutor)** - Isolated script execution with dependency resolution

## 🔧 Installation

### Linux
```bash
# Download for your platform
wget https://github.com/perigrin/pvm/releases/download/v1.0.0-rc32/pvm-1.0.0-rc32-linux-amd64.tar.gz
tar -xzf pvm-1.0.0-rc32-linux-amd64.tar.gz

# Make binaries executable and test
chmod +x pvm-* pvi-* pvx-*
./pvm-linux-amd64 version
```

### macOS
```bash
# Download for your platform (Intel or Apple Silicon)
wget https://github.com/perigrin/pvm/releases/download/v1.0.0-rc32/pvm-1.0.0-rc32-darwin-amd64.tar.gz
tar -xzf pvm-1.0.0-rc32-darwin-amd64.tar.gz

# Make executable and remove quarantine flag
chmod +x pvm-darwin-amd64
xattr -d com.apple.quarantine pvm-darwin-amd64

# Test installation
./pvm-darwin-amd64 version
```

**⚠️ macOS Security Notice**: These binaries are unsigned. If you get a security warning:
1. **Right-click the binary** → **"Open"** → **"Open"** (bypasses Gatekeeper)
2. **Or use Terminal**: `xattr -d com.apple.quarantine pvm-darwin-*`
3. **Or temporarily disable Gatekeeper**: `sudo spctl --master-disable` (re-enable with `--master-enable`)

## New to PVM? 

Check out the [Getting Started Guide](docs/quickstart.md) for immediate hands-on experience.

**Enhanced Features:**
- 🐪 Manage multiple Perl versions seamlessly
- 📦 CPAN module management with type-aware dependency resolution
- 🔍 Static type checking with symbol-aware analysis
- 🏃 Isolated script execution with performance optimization
- 🔗 Seamless integration with enhanced language server features
- ⚡ Real-time diagnostics with actionable error messages
- 🎯 Accurate goto definition, find references, and code completion

## Modernized Architecture

### Compiler Pipeline
- `internal/scanner/` - Lexical analysis and tokenization
- `internal/parser/` - AST generation with tree-sitter integration
- `internal/ast/` - Consolidated AST node types and navigation
- `internal/binder/` - Symbol resolution and scope management
- `internal/typechecker/` - Type analysis using symbol information
- `internal/compiler/` - Code generation to multiple targets

### Language Server
- `internal/ls/` - Language service business logic
- `internal/lsp/` - LSP protocol handling and communication
- `internal/diagnostics/` - Enhanced error reporting with symbol context

### Core Components
- `cmd/` - Command entry points for each component
- `internal/cli/` - Unified CLI framework and error handling
- `internal/pvm/`, `internal/psc/`, `internal/pvi/`, `internal/pvx/` - Component-specific implementations
- `docs/` - Comprehensive documentation including architecture guides
- `test/` - Testing infrastructure with baseline and performance tests

## CLI Framework

The PVM Ecosystem uses a unified CLI framework based on [Cobra](https://github.com/spf13/cobra) with the following features:

- **Single Binary, Multiple Entry Points**: The same binary can be invoked as `pvm`, `pvx`, `pvi`, or `psc` (using symlinks) and will provide the appropriate functionality based on how it was invoked.
- **Consistent Command Structure**: All components share the same command structure, error handling, and help text formatting.
- **Command Registration**: Components register their commands with a central registry.
- **Structured Error Handling**: Errors include categories, context, and helpful hints.
- **Global Flags**: Common flags like `--verbose` and `--debug` are available across all components.

### Error Handling and Logging

The PVM Ecosystem includes a comprehensive error handling and logging framework:

- **Structured Errors**: All errors include component prefix, error code, category, description, and optional context.
- **Error Categories**: Errors are categorized as Configuration, Version, Module, Execution, Type, System, or User Input errors.
- **Contextual Information**: Errors can include detailed information about where they occurred and how to fix them.
- **Error Propagation**: Errors maintain their structure when passed between components.
- **Multi-level Logging**: Support for Debug, Info, Warning, Error, and Fatal log levels.
- **Configurable Verbosity**: Logging level can be controlled via the `--verbose` flag.
- **Component Tagging**: Log messages are tagged with the component that generated them.

See [Error Handling and Logging Guide](docs/error-handling.md) for more details.

### Command Router

The command router enables the single binary to function as multiple tools based on how it's invoked:

- **Binary Name Detection**: Automatically detects which component to run based on the filename used to invoke the binary.
- **Symlink Support**: Works with symlinks on Unix-like systems and file copies on Windows.
- **Fallback Behavior**: If invoked with an unknown name, defaults to PVM functionality.
- **Debug Mode**: Run with `--debug` flag to see detailed information about the invocation.
- **Symlink Management**: Includes commands to create and verify the necessary symlinks:
  ```
  pvm symlinks create  # Create symlinks for all components
  pvm symlinks verify  # Check if all symlinks exist
  ```

## Building the Project

```bash
# Build all components using the Makefile
make

# Build a specific component
make pvm
make pvx
make pvi
make psc

# Run tests
go test ./...
```

## Features

- Zero-configuration operation
- Integration with existing Perl version managers (plenv, perlbrew)
- Multi-level environment isolation for script execution (none, low, medium, high)
- Type checking with flow-sensitive analysis
- CPAN module management
- Flexible TOML-based configuration system

## Configuration

The PVM Ecosystem uses TOML configuration files located according to the XDG Base Directory Specification:

- System-wide configuration: `/etc/pvm/pvm.toml`
- User configuration: `$XDG_CONFIG_HOME/pvm/pvm.toml` (defaults to `~/.config/pvm/pvm.toml`)
- Project configuration: `.pvm/pvm.toml` in the project directory

All configuration files are optional. The system works without any configuration files present, using sensible defaults.

Basic configuration commands:

```bash
# Show the current configuration
pvm config show

# Get a specific configuration value
pvm config get pvm.default_perl

# Set a configuration value
pvm config set pvm.default_perl 5.36.0

# Create a new configuration file with default values
pvm config init
```

See [Configuration Guide](docs/configuration.md) for more details.

## Documentation

### UI Framework
- **[UI Framework Documentation](docs/ui_framework.md)** - Complete API reference and architecture for beautiful CLI output
- **[UI Examples](docs/ui_examples.md)** - Practical usage examples for all UI components
- **[UI Migration Guide](docs/ui_migration_guide.md)** - Step-by-step migration from direct output to styled UI

### Architecture and Development
- **[Architecture Overview](docs/architecture-overview.md)** - Complete system architecture with TypeScript-Go patterns
- **[Developer Guide](docs/developer-guide.md)** - Working with the modernized compiler pipeline
- **[Contributor Guide](docs/contributor-guide.md)** - Contributing to PVM development
- **[Build System Guide](docs/build-system-guide.md)** - Modern build system capabilities

### Language Server Protocol
- **[LSP Guide](docs/lsp-guide.md)** - Complete LSP features, setup, and performance
- **[Migration Guide](docs/migration-guide.md)** - Upgrading from legacy PVM versions

### User Guides
- **[Quick Start](docs/quickstart.md)** - 15-minute hands-on experience
- **[Typed Perl Specification](docs/typed-perl-specification.md)** - Complete type system reference
- **[Workflow Guides](docs/)** - Development, migration, and CI/CD workflows

## PVX Isolation Levels

PVX provides multiple isolation levels for script execution:

- **none** - No isolation, uses system environment
- **low** - Minimal isolation with local module installation
- **medium** - Clean PERL5LIB environment with restricted module access
- **high** - Strongest possible isolation without containers

Example usage:
```bash
# Run with low isolation level
pvx --isolation=low script.pl

# Run with high isolation level and custom isolation directory
pvx --isolation=high --isolation-dir=/path/to/isolation script.pl
```

See [PVX Isolation Guide](docs/pvx-isolation.md) for more details.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
