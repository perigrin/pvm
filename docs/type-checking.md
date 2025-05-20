# Type Checking in PSC

This document describes the type checking system implemented in PSC (Perl Script Compiler) as part of the PVM Ecosystem.

## Overview

PSC provides gradual type checking for Perl scripts and modules through the following features:

- Type annotation parsing for Perl code
- Type hierarchy and compatibility checking
- Type error reporting
- Command-line tools for checking and stripping type annotations

The type checking system is design to be:

- Unobtrusive: Type annotations are optional and can be stripped for compatibility with regular Perl
- Gradual: Typing can be added incrementally to existing code
- Flexible: Supports a variety of Perl-specific types and type combinations

## Type Annotation Syntax

PSC supports several types of annotations:

### Variable Declarations

```perl
my Int $count = 42;            # Scalar with type annotation
my Str @names = ('Alice', 'Bob');  # Array with type annotation
my HashRef[Str, Int] %ages;    # Hash with parameterized type annotation
```

### Function Parameters and Returns

```perl
sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

method calculate(Num $value) -> HashRef[Str, Num] {
    return { result => $value * 2 };
}
```

### Attribute/Field Declarations

```perl
class Person {
    field Str $name;
    field Int $age;
}
```

### Type Definitions

```perl
type UserId = Int;
type UserRecord = HashRef[Str, Str|Int|ArrayRef];
```

## Type Hierarchy

PSC implements a comprehensive type hierarchy with the following characteristics:

- All types are subtypes of `Any`
- Scalar types include `Str`, `Num`, `Int`, `Float`, `Bool`, etc.
- Container types include `Array`, `Hash`, `List`, etc.
- Reference types include `ArrayRef`, `HashRef`, `CodeRef`, etc.
- Special types include `Maybe[T]` for optional values

The primary subtype relationships are:

- `Int` is a subtype of `Num`
- `Num` is a subtype of `Scalar`
- `Scalar` is a subtype of `Any`
- `Ref` is a subtype of `Any`
- `ArrayRef` is a subtype of `Ref`
- `HashRef` is a subtype of `Ref`
- etc.

## Parameterized Types

PSC supports parameterized types for containers:

- `ArrayRef[T]` - Array reference containing elements of type T
- `HashRef[K,V]` - Hash reference with keys of type K and values of type V
- `Maybe[T]` - Type that can be either T or undefined
- `Optional[T]` - Similar to Maybe, but specifically for optional hash keys

## Type Checking Rules

The type checking engine enforces the following rules:

1. Values can be assigned to variables of the same type or a supertype
2. Function arguments must be subtypes of the declared parameter types
3. Function return values must be subtypes of the declared return type
4. Field/attribute assignments must be compatible with the declared type

## PSC Commands

### Type Checking

The `psc check` command analyzes one or more Perl files for type errors:

```bash
# Check a single file
psc check myfile.pl

# Check all Perl files in a directory
psc check lib/

# Check with verbose output
psc check --verbose myfile.pl

# Check with custom report format
psc check --format json myfile.pl

# Exclude certain files
psc check --exclude "test_*.pl" lib/
```

### Stripping Annotations

The `psc strip` command removes type annotations for compatibility:

```bash
# Strip annotations and print to stdout
psc strip myfile.pl

# Strip annotations and write to a new file
psc strip myfile.pl clean.pl
```

### Type Definitions

PSC provides commands for managing type definitions:

```bash
# List available type definitions
psc def list

# Generate a type definition for a module
psc def generate Module::Name --save

# Import a type definition from a file
psc def import module_types.json

# Export a type definition to a file
psc def export Module::Name output.json

# Install a type definition for a CPAN module
psc def install Moose
```

### Running Type-Checked Code

The `psc run` command type-checks code before executing it:

```bash
# Type-check and run a script
psc run script.pl

# Type-check and run with arguments
psc run script.pl arg1 arg2

# Run with a specific Perl version
psc run --perl 5.36.0 script.pl
```

### Watching for Changes

The `psc watch` command continuously monitors files for changes:

```bash
# Watch a file for changes
psc watch script.pl

# Watch a directory
psc watch lib/
```

## Integration with PVX

PSC integrates with PVX (Perl Version eXecutor) to run type-checked code in isolated environments:

1. PSC checks the code for type errors
2. If no errors are found, PSC strips the annotations
3. PVX executes the code in an isolated environment with the specified Perl version

## Future Enhancements

Future enhancements to the type checking system may include:

- Flow-sensitive type analysis
- Type inference for variables without annotations
- More sophisticated type compatibility rules
- Enhanced editor integrations (LSP)
- Custom type declarations
- Module-level type definitions