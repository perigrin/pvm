# PVX Isolation Levels

PVX (Perl Version eXecutor) provides configurable isolation levels to control the execution environment for Perl scripts. This document describes the available isolation levels and how to use them.

## Overview

PVX supports four isolation levels, offering increasing degrees of isolation from the host environment:

- **none**: No isolation, runs the script directly in the current environment.
- **low**: Minimal isolation that allows local module installation.
- **medium**: Stronger isolation with a clean PERL5LIB environment.
- **high**: Maximum isolation with minimal environment variables and controlled filesystem access.

## Isolation Levels in Detail

### none

The `none` isolation level runs the script directly in the current environment with no isolation. This is useful for scripts that need full access to the host environment.

- Inherits all environment variables
- Uses the current working directory
- Has full access to the filesystem
- Uses the system's Perl environment

Example:
```bash
pvm pvx --isolation none script.pl
```

### low

The `low` isolation level creates a minimal isolation layer that behaves like a `local::lib` environment. This is useful for scripts that need to install modules locally without affecting the host environment.

- Uses existing Perl environment, but allows installing modules locally
- Inherits all environment variables
- Has full access to the filesystem
- Sets up a local module installation directory
- Preserves the existing PERL5LIB value while adding isolation paths

Example:
```bash
pvm pvx --isolation low script.pl
```

### medium

The `medium` isolation level provides stronger isolation by restricting module access. This is useful for scripts that need a clean environment but still need access to the host filesystem.

- Creates a clean PERL5LIB environment with only isolation paths
- Still inherits most environment variables
- Has full access to the filesystem but restricts module installation
- Creates a more isolated environment for module installation

Example:
```bash
pvm pvx --isolation medium script.pl
```

### high

The `high` isolation level creates the strongest isolation possible without using containers. This is useful for running untrusted scripts or ensuring reproducible environments.

- Creates a clean environment with only essential environment variables
- Restricts module access to only the isolation directory
- Restricts filesystem access to specified paths where possible
- Sets additional PVM_* environment variables for script-based isolation
- Provides the most isolated environment possible without containers

Example:
```bash
pvm pvx --isolation high script.pl
```

## Advanced Isolation Features

### Isolated Output Directory

For high isolation level, you can create a separate output directory for script-generated files with the `--isolated-output` flag:

```bash
pvm pvx --isolation high --isolated-output script.pl
```

This creates a dedicated `/output` directory within the isolation environment where all generated files will be placed. When using the `--verbose` flag, PVX will list all files created in this directory.

Benefits of isolated output:
- Prevents scripts from modifying files outside the designated area
- Makes it easy to collect and process generated files
- Improves security when running untrusted code
- Simplifies cleanup after execution

### File System Access Control

With high isolation, you can control filesystem access with the following flags:

- `--ro-path`: Specify paths that should be accessible for reading but not writing
- `--rw-path`: Specify paths that should be accessible for both reading and writing

```bash
pvm pvx --isolation high --ro-path /usr/share/perl --rw-path /tmp/output script.pl
```

Multiple paths can be specified by repeating the flags:

```bash
pvm pvx --isolation high \
  --ro-path /usr/share/perl \
  --ro-path /etc/perl \
  --rw-path /tmp/output \
  --rw-path /var/log/myapp \
  script.pl
```

Note: These paths are passed to the script as environment variables (`PVM_READONLY_PATHS` and `PVM_READWRITE_PATHS`). The script needs to respect these paths using a module like `pvm::isolation` (automatically loaded for high isolation) to enforce the restrictions.

### Saving Output Files

You can automatically save files generated in the isolated output directory to a permanent location:

```bash
pvm pvx --isolation high --isolated-output --save-output-dir /path/to/save script.pl
```

The saved files will preserve their permissions and will be copied to the specified directory after script execution. When using `--verbose`, detailed information about saved files will be displayed.

The save directory can use environment variables (e.g., `$HOME`, `$PWD`):

```bash
pvm pvx --isolation high --isolated-output --save-output-dir $PWD/outputs script.pl
```

### Environment Variable Control

You can control which environment variables are preserved in isolation:

- `--preserve-env`: Specify environment variables to preserve in high isolation
- `--clear-env`: Specify environment variables to remove in all isolation levels

```bash
# Preserve specific variables in high isolation
pvm pvx --isolation high --preserve-env MY_APP_CONFIG --preserve-env LICENSE_KEY script.pl

# Clear specific variables in any isolation level
pvm pvx --isolation medium --clear-env PERL_MB_OPT --clear-env PERL_LOCAL_LIB_ROOT script.pl
```

Note that `--clear-env` takes precedence over `--preserve-env` if both are specified for the same variable.

### Custom Environment Variables

You can set custom environment variables directly:

```bash
pvm pvx --isolation high -e MY_VAR=value script.pl
```

### Custom Module Paths

Add additional directories to the PERL5LIB environment variable:

```bash
pvm pvx --include-path /path/to/modules script.pl
```

You can specify multiple additional module paths:

```bash
pvm pvx --include-path /path/to/modules1 --include-path /path/to/modules2 script.pl
```

You can also specify a custom directory for module installation:

```bash
pvm pvx --module-path /path/to/modules script.pl
```

## Cleanup Behavior

By default, PVX cleans up temporary isolation directories after script execution. You can disable this behavior with the `--no-cleanup` flag:

```bash
pvm pvx --isolation medium --no-cleanup script.pl
```

The location of the isolation directory will be displayed in the output when using the `--verbose` flag.

Benefits of keeping the isolation directory:
- Allows inspecting the environment after execution
- Enables reuse of installed modules for subsequent runs
- Helps with debugging by preserving the execution context

## Custom Isolation Directory

You can specify a custom directory to use for isolation with the `--isolation-dir` flag:

```bash
pvm pvx --isolation medium --isolation-dir /path/to/isolation script.pl
```

Note that user-specified isolation directories are not automatically cleaned up, regardless of the `--no-cleanup` flag setting. This is to prevent accidental deletion of user-specified directories.

## Combining Isolation Features

Isolation features can be combined for advanced use cases:

```bash
# Full example with multiple features
pvm pvx --isolation high \
  --isolated-output \
  --save-output-dir $PWD/results \
  --ro-path /usr/share/perl \
  --ro-path /etc/perl \
  --rw-path $PWD/data \
  --preserve-env API_KEY \
  --preserve-env CONFIG_PATH \
  --include-path $PWD/lib \
  --module-path $HOME/.perl_modules \
  script.pl
```

## Legacy Support

For backward compatibility, the `--isolated` flag is still supported and is equivalent to `--isolation low`:

```bash
# These are equivalent
pvm pvx --isolated script.pl
pvm pvx --isolation low script.pl
```

## Environment Variables

All isolation levels preserve custom environment variables set in the command line. However, they differ in how they handle existing environment variables:

- **none**: All environment variables are inherited.
- **low**: All environment variables are inherited, plus local::lib variables.
- **medium**: Most environment variables are inherited, but PERL5LIB is cleaned.
- **high**: Only essential environment variables are inherited (PATH, HOME, USER, SHELL, TMPDIR, TERM, etc.).

### Special Environment Variables

When using high isolation, PVX sets additional environment variables that can be used by the script:

- `PVM_ISOLATION_LEVEL`: The current isolation level (e.g., "high")
- `PVM_ISOLATION_DIR`: The path to the isolation directory
- `PVM_OUTPUT_DIR`: The path to the isolated output directory (when using `--isolated-output`)
- `PVM_ISOLATED_OUTPUT`: Set to "1" when `--isolated-output` is used
- `PVM_READONLY_PATHS`: Colon-separated list of read-only paths
- `PVM_READWRITE_PATHS`: Colon-separated list of read-write paths

These variables can be used by the script to respect isolation boundaries.

## File System Access

All isolation levels still have access to the host filesystem by default. The difference is in where new files are created and what paths are controlled:

- **none**: Files are created in the current directory.
- **low/medium**: Files are created in the isolation directory by default.
- **high**: Files are created in the isolated output directory when using `--isolated-output`.

When using high isolation with `--ro-path` and `--rw-path`, the filesystem access is controlled through environment variables. For full filesystem isolation, consider using container technologies like Docker.

## Configuration

Isolation settings can be specified in the configuration file (`pvm.toml`) to provide default values for all PVX executions:

```toml
[pvx]
isolation_level = "medium"
cleanup_after = true
isolation_ro_paths = ["/usr", "/etc"]
isolation_rw_paths = ["/tmp"]
isolated_output = true
save_output_dir = "$PWD/output"
preserve_env_vars = ["LICENSE_KEY", "API_TOKEN"]
additional_module_paths = ["$HOME/perl/lib"]
custom_module_path = "$HOME/.pvm/modules"
```

Command-line options take precedence over configuration file settings.

## Best Practices

1. **Start with low isolation** and increase as needed
2. **Use `--verbose`** to see what PVX is doing with isolation
3. **Specify explicit paths** instead of relying on default locations
4. **Use isolated output** when running untrusted code
5. **Consider filesystem access** requirements when choosing isolation level
6. **Preserve only necessary environment variables** in high isolation
7. **Use configuration files** for consistent isolation settings
8. **Set up custom module paths** to avoid reinstalling modules
