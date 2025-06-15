# PVM Command Reference

**Comprehensive reference for all PVM commands and options.**

This document provides complete coverage of all PVM commands, subcommands, and their options. Commands are organized by functional area.

## Table of Contents

1. [Version Management](#version-management)
2. [Configuration Management](#configuration-management)
3. [Project Management](#project-management)
4. [Build System](#build-system)
5. [Development Environment](#development-environment)
6. [Module Management](#module-management)
7. [Test Execution](#test-execution)
8. [Environment Management](#environment-management)
9. [Tool Management](#tool-management)
10. [Shell Integration](#shell-integration)
11. [Help System](#help-system)
12. [Advanced Features](#advanced-features)

---

## Version Management

### pvm install
Install a Perl version from source.

```bash
pvm install [version]
```

**Options:**
- `--source <file>` - Source archive file path (default: download or use cached)
- `--prefix <dir>` - Installation directory (default: XDG_DATA_HOME/pvm/versions/<version>)
- `--jobs <n>` - Number of parallel build jobs (default: number of CPU cores)
- `--test` - Run Perl tests after building
- `--skip-build` - Skip build and import from existing installation

**Examples:**
```bash
pvm install 5.38.0
pvm install 5.36.0 --test --jobs 8
pvm install 5.40.0 --source perl-5.40.0.tar.gz
```

### pvm use
Use a specific version in the current shell.

```bash
pvm use [version]
```

**Examples:**
```bash
pvm use 5.38.0
pvm use latest
```

### pvm global
Set the global default Perl version.

```bash
pvm global [version]
```

### pvm local
Set the local version for a directory.

```bash
pvm local [version]
```

### pvm versions
List installed Perl versions.

```bash
pvm versions
```

**Options:**
- `--paths` - Show installation paths
- `--source` - Show source and installation time

### pvm available
List available Perl versions for installation.

```bash
pvm available
```

### pvm uninstall
Remove a Perl version.

```bash
pvm uninstall [version]
```

**Options:**
- `--force` - Skip confirmation prompt

---

## Configuration Management

The configuration system supports system-wide, user-level, and project-level configurations with proper precedence.

### pvm config show
Display the effective configuration after merging all sources.

```bash
pvm config show
```

**Options:**
- `--format <format>` - Output format: toml, json, yaml (default: toml)

### pvm config get
Get a specific configuration value.

```bash
pvm config get [section.key]
```

**Examples:**
```bash
pvm config get pvm.default_perl
pvm config get pvi.preferred_installer
pvm config get build.mode
```

### pvm config set
Set a configuration value.

```bash
pvm config set [section.key] [value]
```

**Options:**
- `--system` - Update system-wide configuration
- `--user` - Update user configuration (default)
- `--project` - Update project configuration

**Examples:**
```bash
pvm config set pvm.default_perl 5.38.0
pvm config set build.mode distribution --project
pvm config set pvi.test_during_install true --user
```

### pvm config unset
Remove a configuration value (revert to default).

```bash
pvm config unset [section.key]
```

**Options:**
- `--system` - Update system-wide configuration
- `--user` - Update user configuration (default)
- `--project` - Update project configuration

### pvm config init
Initialize a configuration file with default values.

```bash
pvm config init
```

**Options:**
- `--system` - Create system-wide configuration
- `--user` - Create user configuration (default)
- `--project` - Create project configuration

### pvm config validate
Validate the current configuration for errors and warnings.

```bash
pvm config validate
```

### pvm config diff
Show configuration differences.

```bash
pvm config diff [config-file]
```

If no file is specified, compares with default configuration.

### pvm config sources
Show configuration sources and their precedence order.

```bash
pvm config sources
```

### pvm config backup
Create a backup of current configuration files.

```bash
pvm config backup [backup-directory]
```

Default backup directory: `~/.config/pvm/backups`

### pvm config restore
Restore configuration from a backup file.

```bash
pvm config restore [backup-file]
```

### pvm config list-backups
List available configuration backups.

```bash
pvm config list-backups [backup-directory]
```

### pvm config generate
Generate configuration from templates or profiles.

```bash
pvm config generate
```

**Options:**
- `--template <name>` - Template name to use for generation
- `--profile <name>` - Profile name to use for generation
- `--output <file>` - Output file path (default: pvm.toml)
- `--var <key=value>` - Template variables (can be specified multiple times)
- `--list` - List available templates and profiles
- `--force` - Overwrite existing output file

**Examples:**
```bash
pvm config generate --list
pvm config generate --template basic --output pvm.toml
pvm config generate --profile development --var perl_version=5.38.0
```

---

## Project Management

### pvm project init
Initialize a new project with templates and scaffolding.

```bash
pvm project init [project-name]
```

**Options:**
- `--template <name>` - Project template to use
- `--force` - Overwrite existing project files

**Examples:**
```bash
pvm project init my-app
pvm project init web-service --template web
pvm project init . --force  # Initialize in current directory
```

### pvm project status
Show comprehensive project status and health information.

```bash
pvm project status
```

**Options:**
- `--json` - Output in JSON format for scripting

**Shows:**
- Project detection results
- Perl version consistency
- Dependencies status
- Build status
- Configuration validation

### pvm project doctor
Run comprehensive health checks with optional auto-fixing.

```bash
pvm project doctor
```

**Options:**
- `--fix` - Attempt automatic fixes for detected issues
- `--verbose` - Show detailed diagnostic information
- `--json` - Output results in JSON format

**Health checks include:**
- Perl version matches .perl-version
- Required modules are installed
- Configuration issues
- Build artifacts currency
- Project structure validation

### pvm project templates
List available project templates.

```bash
pvm project templates
```

---

## Build System

### pvm build
Unified build command with multiple modes and targets.

```bash
pvm build [target]
```

**Options:**
- `--mode <mode>` - Build mode: distribution, inline, both (default: distribution)
- `--output <dir>` - Output directory (default: build)
- `--clean` - Clean output before building
- `--watch` - Continuous building with file monitoring
- `--check-only` - Type checking without compilation
- `--inline` - Development build (.pmc file generation)
- `--strict` - Enable strict type checking
- `--skip-typecheck` - Skip type checking phase
- `--skip-metadata` - Skip metadata generation
- `--include-tests` - Include tests in distribution
- `--include-scripts` - Include scripts in distribution

**Build Modes:**
- **distribution** - CPAN-ready package with metadata
- **inline** - Development .pmc files for fast iteration
- **both** - Generate both distribution and inline builds

**Examples:**
```bash
pvm build                           # Default distribution build
pvm build --inline                  # Development build
pvm build --watch                   # Continuous building
pvm build --clean --strict         # Clean strict build
pvm build lib --mode inline        # Build specific target
```

---

## Development Environment

### pvm dev
Start an integrated development environment with multiple coordinated services.

```bash
pvm dev
```

**Options:**
- `--build` - Enable build watching service (default: true)
- `--test` - Enable test execution service (default: true)
- `--typecheck` - Enable type checking service (default: true)
- `--test-interval <seconds>` - Test execution interval (default: 5)
- `--typecheck-interval <seconds>` - Type checking interval (default: 2)
- `--verbose` - Show detailed service output

**Services provided:**
- **Build Watcher** - Continuous .pmc file generation
- **Test Runner** - Execute tests on file changes
- **Type Checker** - Continuous validation
- **Status Dashboard** - Real-time development status

**Example:**
```bash
pvm dev                                      # Start all services
pvm dev --test-interval 10 --verbose       # Custom intervals with verbose output
```

---

## Module Management

### pvm module install
Install modules with project-aware behavior.

```bash
pvm module install [modules...]
```

**Options:**
- `--dev` - Include development dependencies (when installing from cpanfile)
- `--force-global` - Force global installation even in project context
- `--parallel <n>` - Number of parallel installations

**Behavior:**
- **In project**: Installs to `./lib/` directory automatically
- **Outside project**: Installs to user or system location
- **No modules specified**: Installs from cpanfile

**Examples:**
```bash
pvm module install                    # Install from cpanfile (runtime only)
pvm module install --dev            # Install from cpanfile (runtime + dev)
pvm module install DBI DBD::mysql   # Install specific modules
```

### pvm module add
Add a module to cpanfile and install it.

```bash
pvm module add [module] [version-spec]
```

**Options:**
- `--dev` - Add as development dependency

**Examples:**
```bash
pvm module add DBI                    # Add runtime dependency
pvm module add Test::More --dev      # Add development dependency
pvm module add DBI ">=1.643"         # Add with version constraint
```

### pvm module sync
Generate or update cpanfile.snapshot lockfile.

```bash
pvm module sync
```

**Options:**
- `--from-snapshot` - Install exact versions from lockfile

**Examples:**
```bash
pvm module sync                       # Generate/update lockfile
pvm module install --from-snapshot   # Install from exact lockfile
```

---

## Test Execution

### pvm test
Run tests with project-aware environment setup.

```bash
pvm test [pattern]
```

**Options:**
- `--verbose` - Show detailed test output
- `--parallel <n>` - Number of parallel test processes
- `--coverage` - Enable coverage reporting (if available)
- `--perl <version>` - Use specific Perl version
- `--test-dir <dir>` - Test directory (default: t/)

**Features:**
- Automatic test discovery in `t/` directory
- Project-aware `@INC` setup
- TAP output parsing
- Environment setup for project dependencies

**Examples:**
```bash
pvm test                             # Run all tests
pvm test t/basic.t                   # Run specific test
pvm test --verbose --parallel 4     # Verbose parallel execution
```

---

## Environment Management

Named isolation environments for development workflow.

### pvm env list
List all named environments.

```bash
pvm env list
```

### pvm env activate
Generate activation command for an environment.

```bash
pvm env activate [name]
```

**Example:**
```bash
eval $(pvm env activate my-env)
```

### pvm env remove
Remove a named environment.

```bash
pvm env remove [name]
```

### pvm env info
Show detailed information about an environment.

```bash
pvm env info [name]
```

---

## Tool Management

Global tool installation and management.

### pvm tool install
Install a tool and make it available as a command.

```bash
pvm tool install [tool[@version]]
```

**Examples:**
```bash
pvm tool install cpanm
pvm tool install Perl::Tidy@20230912
```

### pvm tool run
Run a tool without installing it permanently.

```bash
pvm tool run [tool] [args...]
```

### pvm tool list
List all installed tools.

```bash
pvm tool list
```

### pvm tool upgrade
Upgrade an installed tool to the latest version.

```bash
pvm tool upgrade [tool]
```

### pvm tool uninstall
Remove an installed tool.

```bash
pvm tool uninstall [tool]
```

---

## Shell Integration

### pvm init
Generate shell integration script.

```bash
pvm init
```

**Options:**
- `--generate` - Generate shell initialization scripts instead of printing

**Usage:**
```bash
eval "$(pvm init)"
```

### pvm shell init
Initialize shell integration.

```bash
pvm shell init
```

### pvm shell setup
Show instructions for setting up shell integration.

```bash
pvm shell setup
```

---

## Help System

Context-aware help with workflow suggestions.

### pvm help
Show context-aware help based on current project state.

```bash
pvm help [topic]
```

**Topics:**
- `workflows` - Common development workflows
- `getting-started` - New user onboarding
- `troubleshooting` - Diagnostic commands and solutions
- `next` - Suggested next steps based on project state

**Examples:**
```bash
pvm help                        # Context-aware help
pvm help workflows             # Show common workflows
pvm help getting-started       # New user guide
```

---

## Advanced Features

### pvm mcp-server
Start Model Context Protocol server for AI/LLM integration.

```bash
pvm mcp-server
```

**Options:**
- `--host <host>` - Server host (default: localhost)
- `--port <port>` - Server port (default: 3000)
- `--no-auto-discover` - Disable automatic project discovery
- `--config <file>` - Configuration file path

### pvm version-util
Version manipulation utilities.

```bash
pvm version-util [command]
```

**Commands:**
- `parse <version>` - Parse version strings
- `compare <v1> <v2>` - Compare versions
- `satisfies <version> <constraints>` - Check constraints
- `alias <alias>` - Resolve version aliases

---

## Configuration Reference

### Configuration Hierarchy

1. **Command line flags** (highest priority)
2. **Project configuration** (`.pvm/pvm.toml` or `pvm.toml`)
3. **User configuration** (`~/.config/pvm/pvm.toml`)
4. **System configuration** (`/etc/pvm/pvm.toml`)
5. **Built-in defaults** (lowest priority)

### Configuration Sections

#### [pvm]
```toml
default_perl = "5.38.0"
build_jobs = 4
download_mirror = "https://www.cpan.org/src/5.0"
run_tests = true
```

#### [build]
```toml
mode = "distribution"              # "distribution", "inline", "both"
output_dir = "build"
clean_before_build = true

[build.typecheck]
strict = false
experimental = false
target_perl = "5.36"

[build.files]
include = ["lib/**/*.pm"]
exclude = ["local/**", "build/**"]
watch_dirs = ["lib", "script", "t"]

[build.distribution]
include_tests = true
include_scripts = true
installer = "ExtUtils::MakeMaker"
```

#### [pvi]
```toml
preferred_installer = "auto"       # "auto", "cpanm", "cpan", "cpm"
default_mirror = "https://cpan.metacpan.org"
test_during_install = false
cache_modules = true
check_signatures = true
```

#### [pvx]
```toml
cache_modules = true
cleanup_after = true
isolation_level = "medium"         # "none", "low", "medium", "high"
always_install_deps = true
timeout = 300
max_memory = "512MB"
```

#### [psc]
```toml
strict_mode = false
generate_missing_types = true
check_before_run = true
watch_exclude = ["**/node_modules/**", "**/.git/**"]
```

#### [project]
```toml
name = "MyApp"
version = "1.0.0"
perl_version = "5.36"
description = "My Perl application"
license = "perl_5"
```

---

## Exit Codes

- **0** - Success
- **1** - General error
- **2** - Misuse of shell command
- **64** - Usage error (invalid arguments)
- **65** - Data format error
- **66** - Cannot open input
- **69** - Service unavailable
- **70** - Internal software error
- **78** - Configuration error

---

## Environment Variables

- `PVM_ROOT` - Override PVM root directory
- `PVM_CACHE_DIR` - Override cache directory
- `PVM_CONFIG_DIR` - Override configuration directory
- `PVM_DATA_DIR` - Override data directory
- `PERL5LIB` - Perl library paths (automatically managed)

---

*This command reference covers all available PVM commands and options. For workflow-specific guidance, see the workflow documentation in the main docs directory.*
