# PVM Ecosystem Configuration Guide

The PVM Ecosystem uses [TOML](https://toml.io) files for configuration, following the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) for configuration file locations.

## Configuration File Locations

Configuration files are loaded from the following locations, in order of precedence (later files override earlier ones):

1. **System-wide configuration**: `/etc/pvm/pvm.toml`
2. **User configuration**: `$XDG_CONFIG_HOME/pvm/pvm.toml` (defaults to `~/.config/pvm/pvm.toml`)
3. **Project configuration**: `.pvm/pvm.toml` in the project directory

All configuration files are **optional**. The system works without any configuration files present, using sensible defaults.

## Configuration Format

The configuration file is divided into sections, each corresponding to one of the main components:

- `[pvm]` - Configuration for Perl Version Manager
- `[pvx]` - Configuration for Perl Version eXecutor
- `[pm]` - Configuration for Perl Version Installer
- `[psc]` - Configuration for Perl Script Compiler

Each section contains specific options for that component. See the [example configuration file](config-example.toml) for a complete list of available options.

## Environment Variable Expansion

Configuration values that begin with `$` will be expanded using environment variables. For example:

```toml
[psc]
type_definitions_path = "$XDG_DATA_HOME/pvm/type_definitions"
```

This will be expanded to use the value of the `XDG_DATA_HOME` environment variable.

## Common Configuration Options

### PVM Configuration

```toml
[pvm]
# Default Perl version to use when none is specified
default_perl = "5.38.0"

# Number of parallel jobs to use during build
build_jobs = 4

# Mirror to use for downloading Perl source
download_mirror = "https://www.cpan.org/src/5.0"

# Version aliases map alias names to actual version strings
version_aliases = { latest = "5.38.0", stable = "5.36.0" }
```

### PVX Configuration

```toml
[pvx]
# Level of isolation for script execution
# Valid values: "none", "low", "medium", "high"
isolation_level = "medium"

# Whether to automatically install missing dependencies
always_install_deps = true

# Maximum execution time in seconds
timeout = 300
```

### PM Configuration

```toml
[pm]
# Preferred installation method
# Valid values: "auto", "cpanm", "cpan", "cpm"
preferred_installer = "auto"

# CPAN mirror to use
default_mirror = "https://cpan.metacpan.org"

# Whether to run tests during module installation
test_during_install = false
```

### PSC Configuration

```toml
[psc]
# Path to type definitions
type_definitions_path = "$XDG_DATA_HOME/pvm/type_definitions"

# Whether to enable more rigorous type checking
strict_mode = false

# Patterns to exclude from watch mode
watch_exclude = ["**/node_modules/**", "**/.git/**"]
```

## Creating Configuration Files

You can create configuration files manually, or use the following command to generate a default configuration file:

```bash
pvm config init [--system|--user|--project]
```

This will create a configuration file with default values and comments explaining each option. The `--system`, `--user`, and `--project` flags determine where the configuration file will be created.

## Viewing Configuration

To view the current effective configuration (merged from all sources), use:

```bash
pvm config show
```

To view a specific configuration value:

```bash
pvm config get pvm.default_perl
```

## Setting Configuration Values

To set a configuration value:

```bash
pvm config set pvm.default_perl 5.36.0
```

This will update the user configuration file (`$XDG_CONFIG_HOME/pvm/pvm.toml`). To update a different configuration file, use the `--system`, `--user`, or `--project` flags.

## Default Values

If a configuration option is not specified in any configuration file, a sensible default value will be used. The default values are:

See the [example configuration file](config-example.toml) for a complete list of default values.
