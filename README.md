# PVM Ecosystem

The PVM Ecosystem provides a comprehensive suite of tools for Perl development environment management. Built as a single Go binary with multiple entry points, the system consists of four main components:

- **PVM (Perl Version Manager)** - Manages Perl installations and versions
- **PVX (Perl Version eXecutor)** - Executes modules/scripts in isolated environments
- **PVI (Perl Version Installer)** - Manages CPAN modules for installed Perl versions
- **PSC (Perl Script Compiler)** - Provides static type checking for Perl code

## Project Structure

- `cmd/` - Command entry points for each component
  - `pvm/` - Main Perl Version Manager command
  - `pvx/` - Perl Version eXecutor command
  - `pvi/` - Perl Version Installer command
  - `psc/` - Perl Script Compiler command
- `internal/` - Internal packages not meant for external use
  - `cli/` - Shared CLI framework used by all components
  - `pvm/` - PVM-specific commands
  - `pvx/` - PVX-specific commands
  - `pvi/` - PVI-specific commands
  - `psc/` - PSC-specific commands
  - `version/` - Shared version information
- `pkg/` - Public packages that may be used by other tools
- `docs/` - Documentation
- `test/` - Test fixtures and helpers

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
- Environment isolation for script execution
- Type checking with flow-sensitive analysis
- CPAN module management

## License

This project is licensed under the MIT License - see the LICENSE file for details.
