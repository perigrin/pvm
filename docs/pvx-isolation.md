# PVX Isolation Levels

PVX (Perl Version eXecutor) provides configurable isolation levels to control the execution environment for Perl scripts. This document describes the available isolation levels and how to use them.

## Overview

PVX supports four isolation levels, offering increasing degrees of isolation from the host environment:

- **none**: No isolation, runs the script directly in the current environment.
- **low**: Minimal isolation that allows local module installation.
- **medium**: Stronger isolation with a clean PERL5LIB environment.
- **high**: Maximum isolation with minimal environment variables.

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

Example:
```bash
pvm pvx --isolation low script.pl
```

### medium

The `medium` isolation level provides stronger isolation by restricting module access. This is useful for scripts that need a clean environment but still need access to the host filesystem.

- Creates a clean PERL5LIB environment
- Still inherits most environment variables
- Has full access to the filesystem but restricts module installation
- Creates a more isolated environment for module installation

Example:
```bash
pvm pvx --isolation medium script.pl
```

### high

The `high` isolation level creates the strongest isolation possible without using containers. This is useful for running untrusted scripts.

- Creates a clean environment with minimal environment variables
- Restricts module access to only the isolation directory
- Restricts filesystem access to specified paths where possible
- Creates a more robust isolation for running untrusted code
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

### File System Access Control

With high isolation, you can control filesystem access with the following flags:

- `--read-only-path`: Specify paths that should be accessible for reading but not writing
- `--read-write-path`: Specify paths that should be accessible for both reading and writing

```bash
pvm pvx --isolation high --read-only-path /usr/share/perl --read-write-path /tmp/output script.pl
```

Multiple paths can be specified by repeating the flags.

### Saving Output Files

You can automatically save files generated in the isolated output directory to a permanent location:

```bash
pvm pvx --isolation high --isolated-output --save-output-dir /path/to/save script.pl
```

### Custom Module Paths

Add additional directories to the PERL5LIB environment variable:

```bash
pvm pvx --additional-module-path /path/to/modules script.pl
```

You can also specify a custom directory for module installation:

```bash
pvm pvx --custom-module-path /path/to/modules script.pl
```

## Cleanup Behavior

By default, PVX cleans up temporary isolation directories after script execution. You can disable this behavior with the `--no-cleanup` flag:

```bash
pvm pvx --isolation medium --no-cleanup script.pl
```

The location of the isolation directory will be displayed in the output when using the `--verbose` flag.

## Custom Isolation Directory

You can specify a custom directory to use for isolation with the `--isolation-dir` flag:

```bash
pvm pvx --isolation medium --isolation-dir /path/to/isolation script.pl
```

Note that user-specified isolation directories are not automatically cleaned up.

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

## File System Access

All isolation levels still have access to the host filesystem. The difference is in where new files are created:

- **none**: Files are created in the current directory.
- **low/medium**: Files are created in the isolation directory by default.
- **high**: Files are created in the isolated output directory when using `--isolated-output`.

When using high isolation with `--read-only-path` and `--read-write-path`, the filesystem access is further controlled, but not fully sandboxed. If you need complete filesystem isolation, consider using container technologies like Docker.
