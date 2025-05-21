# PSC Commands Documentation

PSC (Perl Script Compiler) provides static type checking for Perl code. It allows developers to add type annotations to their Perl code and verify type correctness without execution.

## Command Overview

PSC includes the following commands:

| Command | Description |
|---------|-------------|
| `psc check` | Check a file or directory for type errors |
| `psc strip` | Strip type annotations from a file |
| `psc run` | Type-check and execute a file |
| `psc watch` | Watch files and report type errors on change |
| `psc def` | Manage type definitions |

## Command Details

### psc check

Check a file or directory for type errors.

```
psc check [file|dir]
```

#### Options

| Option | Description |
|--------|-------------|
| `-v, --verbose` | Enable verbose output |
| `-w, --warnings` | Show warnings as well as errors |
| `-f, --format` | Report format (text, json) |
| `-e, --exclude` | Patterns to exclude (e.g., 'test_*.pl') |
| `--flow-sensitive` | Enable flow-sensitive analysis (default: true) |
| `--no-flow-sensitive` | Disable flow-sensitive analysis |
| `--show-refinements` | Show type refinements from flow-sensitive analysis |
| `--skip-flow-checks` | Skip flow-sensitive type checks but still perform refinements |
| `-p, --flow-pattern` | Additional flow-sensitive patterns to recognize (e.g., 'isa_check') |

#### Description

The `check` command analyzes Perl code for type errors without executing it. It can check a single file or recursively check all Perl files in a directory.

**Flow-Sensitive Analysis**

Flow-sensitive analysis refines types based on control flow and validation patterns. For example:

```perl
my Maybe[Int] $x = get_value();
if (defined($x)) {
    # In this branch, $x is treated as Int, not Maybe[Int]
    return $x + 1;
}
```

Flow-sensitive analysis recognizes common validation patterns like:
- `defined($var)` checks for Maybe types
- `ref($var) eq 'ARRAY'` checks for array references
- `ref($var) eq 'HASH'` checks for hash references
- `$var->isa('ClassName')` checks for class instances

Use the `--show-refinements` flag to see how types are refined by the analysis.

### psc strip

Strip type annotations from a file.

```
psc strip [file] [output]
```

#### Description

The `strip` command removes type annotations from Perl code, producing valid Perl code that can be executed by any Perl interpreter. If an output file is not specified, the result is printed to stdout.

### psc run

Type-check and execute a file.

```
psc run [file] [args...]
```

#### Options

| Option | Description |
|--------|-------------|
| `-v, --verbose` | Enable verbose output |
| `--skip-check` | Skip type checking |
| `-p, --perl` | Use a specific Perl version |

#### Description

The `run` command performs type checking on a Perl file and then executes it if no errors are found. It uses PVX (Perl Version eXecutor) to run the code in an isolated environment.

### psc watch

Watch files and report type errors on change.

```
psc watch [file|dir]
```

#### Options

| Option | Description |
|--------|-------------|
| `--exclude` | Patterns to exclude from watching |

#### Description

The `watch` command continuously monitors files for changes and performs type checking whenever a file is modified. It provides real-time feedback about type errors as you write code.

### psc def

Manage type definitions.

```
psc def [subcommand]
```

#### Subcommands

| Subcommand | Description |
|------------|-------------|
| `list` | List available type definitions |
| `show` | Show details of a specific type |
| `add` | Add a new type definition |
| `remove` | Remove a type definition |

#### Description

The `def` command allows you to manage type definitions that can be used in type annotations. Type definitions are stored in a central repository and are available to all PSC operations.

## Examples

### Basic Type Checking

```bash
# Check a single file
psc check myfile.pl

# Check all Perl files in a directory
psc check lib/

# Verbose output with warnings
psc check --verbose --warnings myfile.pl

# Exclude test files
psc check --exclude='*_test.pl' lib/
```

### Flow-Sensitive Analysis

```bash
# Show type refinements
psc check --verbose --show-refinements myfile.pl

# Disable flow-sensitive analysis
psc check --no-flow-sensitive myfile.pl

# Skip flow-sensitive checks but perform refinements
psc check --skip-flow-checks myfile.pl
```

### Stripping Type Annotations

```bash
# Remove annotations and write to a new file
psc strip typed.pl untyped.pl

# Print untyped code to stdout
psc strip typed.pl
```

### Running Type-Checked Code

```bash
# Type-check and run a script
psc run script.pl arg1 arg2

# Skip type checking
psc run --skip-check script.pl

# Use a specific Perl version
psc run --perl 5.36.0 script.pl
```

### Watching for Changes

```bash
# Watch a directory for changes
psc watch lib/

# Watch a file
psc watch myfile.pl
```

### Managing Type Definitions

```bash
# List all available types
psc def list

# Show details of a specific type
psc def show ArrayRef

# Add a new type
psc def add MyType

# Remove a type
psc def remove MyType
```

## Integration with PVX

PSC integrates with PVX (Perl Version eXecutor) to run type-checked code in isolated environments. This ensures that the execution environment matches the type checking environment.

When using `psc run`, PVX creates a clean environment with the correct Perl version and required dependencies. This provides consistent behavior between type checking and execution.