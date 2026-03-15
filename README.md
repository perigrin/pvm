# PVM - Perl Version Manager

A fast, cross-platform Perl development toolkit. Manages Perl installations,
modules, and execution environments from a single Go binary.

## Components

- **PVM** - Install, switch, and manage Perl versions
- **PSC** - Parse and analyze Perl source code
- **PM** - Install and manage CPAN modules
- **PVX** - Run Perl scripts in isolated environments

## Installation

### From Release

```bash
# Linux (AMD64)
curl -LO https://github.com/perigrin/pvm/releases/latest/download/pvm-linux-amd64.tar.gz
tar -xzf pvm-linux-amd64.tar.gz
sudo mv pvm-linux-amd64 /usr/local/bin/pvm

# macOS (Apple Silicon)
curl -LO https://github.com/perigrin/pvm/releases/latest/download/pvm-darwin-arm64.tar.gz
tar -xzf pvm-darwin-arm64.tar.gz
chmod +x pvm-darwin-arm64
sudo mv pvm-darwin-arm64 /usr/local/bin/pvm
```

Add shell integration to your profile:

```bash
# bash (~/.bashrc) or zsh (~/.zshrc)
eval "$(pvm init)"

# fish (~/.config/fish/config.fish)
pvm init | source

# PowerShell ($PROFILE)
pvm init | Invoke-Expression
```

### From Source

Requires Go 1.24+. No C compiler or external tools needed.

```bash
git clone https://github.com/perigrin/pvm.git
cd pvm
make
```

## Quick Start

```bash
# Install a Perl version
pvm install 5.40.2

# Switch to it
pvm use 5.40.2

# Set it as default
pvm global 5.40.2

# List installed versions
pvm versions

# Install a CPAN module
pm install Mojolicious

# Run a script
pvx script.pl

# Parse a Perl file
psc parse lib/MyModule.pm
```

## Version Resolution

PVM resolves the active Perl version in this order:

1. `PVM_PERL_VERSION` environment variable
2. `.perl-version` file in current or parent directory
3. Global version set via `pvm global`
4. System Perl

This is compatible with plenv's `.perl-version` files. PVM can also
detect and use Perl versions installed by plenv or perlbrew.

## PSC - Perl Structural Checker

PSC provides tools for inspecting Perl source code using a pure-Go
parser (no external dependencies).

```bash
# Display the AST of a Perl file
psc parse lib/MyModule.pm

# Show as S-expression
psc parse --format sexpr lib/MyModule.pm

# Analyze dependencies
psc analyze lib/
```

## Configuration

Configuration uses TOML files following the XDG Base Directory spec:

- User: `~/.config/pvm/pvm.toml`
- Project: `.pvm/pvm.toml`

All configuration is optional. PVM works out of the box with sensible
defaults.

```bash
pvm config show       # Show current config
pvm config init       # Create default config
```

## Architecture

Single Go binary with multiple entry points via symlink detection.
Invoke as `pvm`, `pvx`, `pm`, or `psc` for component-specific
behavior.

```bash
pvm symlinks create   # Create symlinks for all components
pvm symlinks verify   # Check symlinks exist
```

### Building

```bash
make            # Build all four binaries
make test       # Run test suite
make clean      # Clean build artifacts
make cross-compile  # Build for all platforms
```

Pure Go with no CGO. Cross-compiles to Linux, macOS, and Windows
without a C compiler.

## Integration

PVM integrates with existing Perl version managers:

```bash
pvm import-from plenv      # Register plenv-installed Perls
pvm import-from perlbrew   # Register perlbrew-installed Perls
```

## License

MIT
