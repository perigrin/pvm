# Getting Started with PVM Ecosystem

The PVM Ecosystem is a comprehensive set of tools for Perl development environment management. This guide will help you get started with the four main components:

- **PVM** - Perl Version Manager
- **PVX** - Perl Version eXecutor
- **PM** - Perl Version Installer (CPAN module manager)
- **PSC** - Perl Script Compiler (static type checker)

## Installation

### Option 1: Download Binary (Recommended)

See the [Quickstart Guide](../quickstart.md#installation) for platform-specific binary installation instructions.

### Option 2: Build from Source

For development or unsupported platforms, see [BUILD.md](../../BUILD.md).

## Basic Usage

### 1. Managing Perl Versions (PVM)

Install and switch between different Perl versions:

```bash
# List available Perl versions
pvm list

# Install a specific version
pvm install 5.36.0

# Use a specific version globally
pvm global 5.36.0

# Use a specific version for current project
pvm local 5.34.0

# Show currently active version
pvm current
```

### 2. Running Scripts in Isolated Environments (PVX)

Execute Perl scripts with specific versions and isolation:

```bash
# Run script with default Perl
pvx run script.pl

# Run with specific version
pvx run --perl 5.36.0 script.pl

# Run with high isolation (for testing)
pvx run --isolation high script.pl

# Run with custom environment variables
pvx run --env "DEBUG=1" script.pl
```

### 3. Managing CPAN Modules (PM)

Install and manage CPAN modules for different Perl versions:

```bash
# Install a module
pm install Moose

# Install for specific Perl version
pm install --perl 5.36.0 DBI

# List installed modules
pm list

# Update all modules
pm update

# Search for modules
pm search JSON

# Remove a module
pm remove Test::More
```

### 4. Type Checking Perl Code (PSC)

Add static type checking to your Perl code:

#### Adding Type Annotations

Create a file `example.pl`:

```perl
#!/usr/bin/perl
use v5.36;

# Variable type annotations
my Int $count = 42;
my Str $name = "Alice";
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];

# Function with type annotations
sub Str greet(Str $name, Int $age) {
    return "Hello $name, you are $age years old!";
}

# Usage
my Str $greeting = greet($name, $count);
say $greeting;
```

#### Type Checking

```bash
# Check for type errors
psc check example.pl

# Check all files in a directory
psc check --recursive lib/

# Watch files for changes and check continuously
psc watch lib/

# Show detailed information
psc check --verbose example.pl
```

#### Stripping Type Annotations

```bash
# Remove types for compatibility
psc strip example.pl clean.pl

# Or output to stdout
psc strip example.pl
```

#### Running Type-Checked Code

```bash
# Type-check and run if no errors
psc run example.pl
```

## Advanced Features

### Flow-Sensitive Type Analysis

PSC can refine types based on code flow:

```perl
my Maybe[Str] $input = get_user_input();

if (defined($input)) {
    # Here $input is refined from Maybe[Str] to Str
    my Int $length = length($input);  # This is safe
}
```

### Type Definitions for Modules

Generate type definitions for existing modules:

```bash
# Generate type definitions
psc def generate Moose --save

# List available type definitions
psc def list

# Import custom type definitions
psc def import my_types.json
```

### Integration Between Components

The components work together seamlessly:

```bash
# Install modules, type-check, and run
pm install --perl 5.36.0 JSON::XS
psc check --perl 5.36.0 my_script.pl
psc run --perl 5.36.0 my_script.pl
```

## Configuration

Configuration uses TOML files in these locations (in order of precedence):

1. Project: `.pvm/pvm.toml`
2. User: `$XDG_CONFIG_HOME/pvm/pvm.toml`
3. System: `/etc/pvm/pvm.toml`

Example configuration:

```toml
[pvm]
default_version = "5.36.0"
auto_switch = true

[psc]
enable_flow_sensitive = true
strict_mode = false

[pm]
auto_install_deps = true
skip_tests = false

[pvx]
default_isolation = "medium"
```

## Common Workflows

### Starting a New Project

```bash
# Create project directory
mkdir my-project && cd my-project

# Set Perl version for this project
pvm local 5.36.0

# Install dependencies
pm install Moose DBI JSON::XS

# Create typed Perl code
cat > lib/MyApp.pm << 'EOF'
package MyApp;
use v5.36;

field Str $name;
field Int $version = 1;

method Str greet() {
    return "Hello from $name v$version";
}

1;
EOF

# Type check the code
psc check --recursive lib/

# Run your application
psc run bin/app.pl
```

### Migrating Existing Code

```bash
# Check existing code (will show type inference)
psc check legacy_script.pl

# Add gradual typing
# Edit your code to add type annotations

# Verify the types work
psc check legacy_script.pl

# Keep a clean version for deployment
psc strip legacy_script.pl production_script.pl
```

### Continuous Development

```bash
# In one terminal: watch for changes
psc watch --recursive lib/

# In another terminal: develop your code
# The watcher will automatically check your changes
```

## Type System Basics

### Core Types

- `Any` - Any value
- `Str` - String values
- `Int` - Integer numbers
- `Num` - Numeric values (includes Int)
- `Bool` - Boolean values
- `ArrayRef[T]` - Array reference containing type T
- `HashRef[K,V]` - Hash reference with key type K and value type V
- `Maybe[T]` - Value of type T or undef

### Using Types

```perl
# Simple types
my Str $message = "Hello";
my Int $count = 42;
my Bool $flag = 1;

# Container types
my ArrayRef[Str] $names = ["Alice", "Bob"];
my HashRef[Str, Int] $ages = { alice => 30, bob => 25 };

# Optional values
my Maybe[Str] $optional = undef;
$optional = "now has a value";

# Function signatures
sub Int process_data(ArrayRef[HashRef[Str, Any]] $records) {
    return scalar @$records;
}
```

## Getting Help

Each command provides help via the `--help` flag:

```bash
pvm --help           # General help
psc check --help     # Help for specific command
pm install --help   # Module installation help
pvx run --help       # Script execution help
```

For more information, see the documentation in the `docs/` directory or visit the project website.

## Troubleshooting

### Common Issues

1. **Tree-sitter errors**: Ensure you have the Tree-sitter library installed for PSC functionality
2. **Permission errors**: Make sure the PVM installation directory is writable
3. **Path issues**: Verify that PVM's bin directory is in your PATH
4. **Module conflicts**: Use `pm list` to check installed modules and versions

### Getting Support

- Check the documentation in `docs/`
- Look for similar issues in the GitHub issues
- Create a new issue with detailed information about your problem

## What's Next?

- Read the [Type Checking Guide](type-checking.md) for advanced PSC usage
- Learn about [Configuration](configuration.md) options
- Explore [Integration Patterns](integration.md) between components
- Check out [Performance Tips](performance.md) for large projects
