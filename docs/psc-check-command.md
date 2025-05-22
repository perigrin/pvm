# PSC Check Command Documentation

## Overview

The `psc check` command performs static type checking on Perl files using type annotations. It analyzes Perl code for type compatibility issues, validates type annotations against the type hierarchy, and provides detailed error reporting with line numbers and suggestions.

## Usage

```bash
psc check [options] <file|directory>...
```

## Options

- `--strict, -s`: Exit with non-zero status if type errors are found
- `--verbose, -v`: Enable verbose output showing type annotations and detailed information
- `--recursive, -r`: Recursively check all Perl files in directories

## Examples

### Basic File Checking

```bash
# Check a single Perl file
psc check script.pl

# Check multiple files
psc check script.pl module.pm test.t
```

### Directory Checking

```bash
# Check all Perl files in a directory (recursively)
psc check --recursive lib/

# Check current directory recursively
psc check --recursive .
```

### Verbose Output

```bash
# Get detailed information about type annotations
psc check --verbose script.pl
```

### Strict Mode

```bash
# Exit with error code if type errors are found (useful for CI/CD)
psc check --strict script.pl
```

## Supported File Types

The check command processes files with the following extensions:
- `.pl` - Perl scripts
- `.pm` - Perl modules
- `.t` - Perl test files

Other file types are automatically skipped.

## Type Checking Features

### Supported Types

- Basic types: `Int`, `Str`, `Num`, `Bool`
- Collection types: `ArrayRef`, `HashRef`
- Custom type definitions
- Union types: `Int|Str`
- Intersection types: `Object&Serializable`
- Negation types: `!Undef`
- Parameterized types: `ArrayRef[Int]`

### Type Annotations

```perl
# Typed variable declarations
my Int $count = 42;
my Str $name = "hello";
my Bool $flag = 1;

# Typed field declarations (in classes)
field Int $id;
field Str $title;

# Type assertions
my $value = $input as Int;
```

### Type Inference

The type checker includes intelligent type inference:

```perl
# Inferred as Int
my $count = 42;

# Inferred as Str
my $name = "hello";

# Inferred as Num
my $pi = 3.14159;

# Unknown type (fallback for complex expressions)
my $result = some_function();
```

## Error Output Format

Errors are displayed in compiler-style format with helpful context:

```
script.pl:5:17: error: Type mismatch: expected Int, got Str
   3: use warnings;
   4:
>> 5: my Int $count = "hello";
   6: my Str $name = 42;
   7: print "Done\n";
                   ^
   help: Consider using numeric conversion: int($value) or 0 + $value
```

### Error Components

1. **Location**: `filename:line:column:`
2. **Severity**: `error:`, `warning:`, or `info:`
3. **Message**: Descriptive error message
4. **Context**: Source code lines around the error
5. **Marker**: Visual indicator pointing to the problem
6. **Suggestion**: Helpful advice for fixing the error

## Error Types and Suggestions

### Type Mismatch Errors

**Int/Str Mismatches:**
```perl
my Int $count = "hello";  # Error: Type mismatch: expected Int, got Str
# Suggestion: Consider using numeric conversion: int($value) or 0 + $value

my Str $name = 42;        # Error: Type mismatch: expected Str, got Int
# Suggestion: Consider using string interpolation: "$value" or explicit conversion
```

**General Type Mismatches:**
```perl
my Bool $flag = [];       # Error: Type mismatch: expected Bool, got ArrayRef
# Suggestion: Check that the assigned value matches the declared type
```

### Undefined Variable Errors

```perl
print $undefined_var;     # Error: Variable undefined or not found
# Suggestion: Make sure the variable is declared before use, or check for typos
```

### Type Annotation Errors

```perl
my InvalidType $var = 42; # Error: Invalid type annotation syntax
# Suggestion: Review the type annotation syntax: my TypeName $variable = value;
```

## Exit Codes

- `0`: Success (no type errors found, or errors found but not in strict mode)
- `1`: Type errors found in strict mode, or system error occurred

## Integration with PVM Ecosystem

### Configuration

Type checking behavior can be configured through PVM configuration files:

```toml
[psc]
strict_mode = true
show_context_lines = 3
enable_colors = true
```

### CI/CD Integration

Use strict mode for continuous integration:

```bash
# In your CI script
psc check --strict --recursive src/
if [ $? -ne 0 ]; then
    echo "Type checking failed!"
    exit 1
fi
```

### Editor Integration

The check command provides structured output suitable for editor integration:

```bash
# JSON output for programmatic processing (future feature)
psc check --format json script.pl
```

## Performance Considerations

### Large Codebases

For large codebases, consider:

- Using `--recursive` to check entire directory trees efficiently
- Running checks in parallel on different directories
- Integrating with build systems for incremental checking

### Memory Usage

The type checker caches parsed files and type information for better performance:

- Source code is cached per file for context display
- Type definitions are cached globally
- AST information is cached during analysis

## Troubleshooting

### Common Issues

**Build Errors:**
```bash
# Ensure tree-sitter is built
make tree-sitter

# Full build if needed
make
```

**No Files Processed:**
- Check file extensions (.pl, .pm, .t)
- Use `--verbose` to see which files are being processed
- Verify file paths are correct

**Parser Errors:**
- Check Perl syntax is valid
- Ensure type annotations follow Typed Perl syntax
- Use `--verbose` for detailed parsing information

### Debug Mode

Enable verbose output for debugging:

```bash
psc check --verbose script.pl
```

This shows:
- Files being processed
- Type annotations found
- Parsing progress
- Type inference results

## Type System Limitations

### Current Limitations

1. **Limited Flow Analysis**: Basic control flow analysis is implemented
2. **Function Signatures**: Limited support for complex function type signatures
3. **Dynamic Features**: Dynamic Perl features may not be fully analyzed
4. **CPAN Modules**: External module types may not be available

### Future Enhancements

- Enhanced flow-sensitive analysis
- Better CPAN module integration
- Generic type support
- Advanced type inference

## Examples

### Simple Type Checking

**Input (script.pl):**
```perl
use strict;
use warnings;

my Int $count = 42;
my Str $message = "Hello, World!";
my Bool $debug = 1;

print "$message (count: $count)\n" if $debug;
```

**Command:**
```bash
psc check script.pl
```

**Output:**
```
✓ script.pl: No type errors found
```

### Error Detection

**Input (errors.pl):**
```perl
use strict;
use warnings;

my Int $count = "not a number";
my Str $name = 42;
```

**Command:**
```bash
psc check errors.pl
```

**Output:**
```
errors.pl:4:17: error: Type mismatch: expected Int, got Str
   2: use warnings;
   3:
>> 4: my Int $count = "not a number";
   5: my Str $name = 42;
                   ^
   help: Consider using numeric conversion: int($value) or 0 + $value

errors.pl:5:16: error: Type mismatch: expected Str, got Int
   3:
   4: my Int $count = "not a number";
>> 5: my Str $name = 42;
                  ^
   help: Consider using string interpolation: "$value" or explicit conversion
```

## See Also

- [Typed Perl Specification](../typed-perl-spec.md)
- [PSC Commands Overview](psc-commands.md)
- [Type System Documentation](type-checking.md)
- [PVM Configuration](configuration.md)
