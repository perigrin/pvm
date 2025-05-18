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
- `pkg/` - Public packages that may be used by other tools
- `docs/` - Documentation
- `test/` - Test fixtures and helpers

## Building the Project

```bash
# Build all components
go build ./...

# Build a specific component
go build -o pvm ./cmd/pvm
go build -o pvx ./cmd/pvx
go build -o pvi ./cmd/pvi
go build -o psc ./cmd/psc

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
