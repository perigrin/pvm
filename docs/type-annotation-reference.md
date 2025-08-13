# PVM Type Annotation Reference

## Overview

PVM (Perl Version Manager) now includes comprehensive support for type annotations in Perl code. This document provides a complete reference for all supported type annotation features.

## Basic Type Annotations

### Simple Types

Type annotations can be added to variable declarations using built-in type names:

```perl
my Int $count = 42;
my Str $name = "example";
my Bool $flag = 1;
my Num $pi = 3.14159;
```

**Supported Built-in Types:**
- `Int` - Integer values
- `Str` - String values
- `Bool` - Boolean values (accepts only: `1`, `0`, `""`, `undef` per Types::Standard)
- `Num` - Numeric values (including floats)
- `ArrayRef` - Array references
- `HashRef` - Hash references (value-type only, use `Map` for key-value constraints)
- `Map` - Hash references with key-value type constraints
- `CodeRef` - Code references
- `ScalarRef` - Scalar references
- `GlobRef` - Glob references
- `Undef` - Undefined values

### Array and Hash Types

Arrays and hashes can be typed with parameterized types:

```perl
# Array references with element types
my ArrayRef[Int] @numbers = (1, 2, 3);
my ArrayRef[Str] $array_ref = ["a", "b", "c"];

# Hash references with value types only
my HashRef[Str] %config = (key => 'value');
my HashRef[Int] $hash_ref = {x => 1, y => 2};

# Map for both key and value type constraints (per Types::Standard)
my Map[Str, Int] %scores = (alice => 95, bob => 87);
my Map[UserID, UserData] $user_map = { 1 => $user_data };
```

### Custom Types

You can use custom type names and package-qualified types:

```perl
my MyType $custom;
my Package::CustomType $qualified;
my Local::User $user = Local::User->new();
```

### Different Scoping Keywords

Type annotations work with all variable scoping keywords:

```perl
my Int $lexical = 42;
our Int $global_counter = 0;
state Str $persistent_cache = "";
local Int $temporarily_modified = $original;
```

## Advanced Type Expressions

### Union Types

Union types allow a variable to accept multiple types using the `|` operator:

```perl
# Simple union types
my Int|Str $flexible = 42;
my Bool|Undef $maybe_flag;

# Multi-way unions
my Int|Str|Bool $multi = "text";
my Num|ArrayRef|HashRef $complex;

# Union types with custom types
my MyClass|OtherClass $object;
my Package::Type1|Package::Type2 $qualified;
```

### Parameterized Types

Parameterized types specify type parameters using bracket notation:

```perl
# Basic parameterized types
my ArrayRef[Int] @numbers;
my HashRef[Str] %strings;        # Single parameter: value type only
my CodeRef[Int, Str] $function;

# Key-value constraints using Map (per Types::Standard)
my Map[Str, Int] %mapping;       # Key type: Str, Value type: Int
my Tuple[Int, Str, Bool] $triple;

# Nested parameterization
my ArrayRef[ArrayRef[Int]] @matrix;
my HashRef[ArrayRef[Str]] %grouped_strings;  # Values are ArrayRef[Str]
my ArrayRef[Map[Str, Int]] @array_of_maps;   # Array of string-to-int maps

# Parameterized with unions
my ArrayRef[Int|Str] @mixed;
my HashRef[Bool|Undef] %flags;   # Hash with Bool or Undef values
my Map[UserID|Str, UserData] %users;  # Keys can be UserID or Str
```

### Intersection Types

Intersection types require values to satisfy multiple type constraints using the `&` operator:

```perl
my Object&Serializable $serializable_object;
my Readable&Writable $file_handle;
my (Int|Str)&Defined $defined_value;
my ArrayRef[Object&Clonable] @clonable_objects;
```

### Negation Types

Negation types exclude specific types using the `!` operator:

```perl
my !Undef $definitely_defined;
my ArrayRef[!Str] @non_strings;
my HashRef[!Undef] %defined_values;
```

### Complex Type Combinations

Complex combinations use operator precedence: negation > intersection > union

```perl
my Int|Str&Defined $complex;  # Equivalent to (Int|(Str&Defined))
my (Int|Str)&Defined $grouped;
my !Undef&Serializable $constrained;
```

## Method Signatures

### Basic Method Types

Methods can have typed parameters and return types using the **only supported syntax**:

```perl
method Int calculate(Int $a, Int $b) {
    return $a + $b;
}

method Str greet(Str $name) {
    return "Hello, $name!";
}
```

**Syntax Standards**:
- ✅ **Supported**: `method ReturnType name(params)` - Return type after keyword
- ❌ **Not Supported**: `method name(params) -> ReturnType` - Arrow syntax is not implemented
- ⚠️ **Deprecated**: `method name(params) returns ReturnType` - Legacy syntax, prefer above

### Optional Parameters

Parameters can have default values:

```perl
method ArrayRef[Str] process(Str $input, Bool $validate = 1) {
    my @result = split /,/, $input;
    return \@result;
}
```

### Named Parameters

Methods can use named parameter syntax:

```perl
method ConnectionConfig configure(
    :$host as Str,
    :$port as Int = 8080,
    :$ssl as Bool = 0,
    :$timeout as Optional[Num]
) {
    return ConnectionConfig->new(
        host => $host,
        port => $port,
        ssl => $ssl,
        timeout => $timeout
    );
}
```

### Variadic Parameters

Methods can accept variable numbers of arguments:

```perl
method Int sum(Int *@numbers) {
    my $total = 0;
    $total += $_ for @numbers;
    return $total;
}
```

### Complex Method Signatures

Methods can combine all features:

```perl
method Result[T, ProcessingError] complex_method<T>(
    Required[T] $input,
    Optional[CodeRef[T, Bool]] $validator = undef,
    :$timeout as Num = 30.0,
    Slurpy[Any] *@rest
) where T: Serializable {
    return success($input);
}
```

## Classes and Roles

### Typed Fields

Class fields can have type annotations:

```perl
class User {
    field Int $id;
    field Str $name;
    field Optional[Email] $email;
    field ArrayRef[Role] $roles = [];

    method User new(Int $id, Str $name, Optional[Email] $email = undef) {
        return bless {
            id => $id,
            name => $name,
            email => $email,
            roles => []
        }, __PACKAGE__;
    }
}
```

### Generic Classes

Classes can be parameterized with type parameters:

```perl
class Container<T> where T: Serializable {
    field ArrayRef[T] $items = [];

    method Void add(T $item) {
        push @{$items}, $item;
    }

    method ArrayRef[T] get_all() {
        return $items;
    }
}
```

### Role Definitions

Roles can specify typed method signatures:

```perl
role Serializable {
    method Str serialize();
    method Self deserialize(Str $data);
}

role Cacheable<K> where K: Serializable {
    field Optional[DateTime] $cached_at;
    method K cache_key();
    method Bool is_stale();
}
```

### Inheritance with Types

Classes can inherit and implement roles with type constraints:

```perl
class Document : BaseDocument does Serializable, Cacheable<DocumentId> {
    field Str $content;
    field DateTime $created;
    field Optional[UserRef] $author;

    method Str serialize() {
        return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        });
    }

    method DocumentId cache_key() {
        return $self->id;
    }
}
```

## Type Assertions and Constraints

### Runtime Type Assertions

Type assertions validate types at runtime using the `as` keyword:

```perl
# Basic type assertions
my $number = $input as Int;
my $text = $data as Str;
my $ref = $object as MyClass;

# Type assertions in expressions
my $result = ($calculation + $offset) as Num;
my $item = $array->[$index] as ItemType;

# Conditional type assertions
my $value = $maybe_number as Int // 0;
my $obj = $input as MyClass or die "Wrong type";
```

### Type Constraints

Type constraints add runtime validation using `where` clauses:

```perl
my $validated = $input as (Int where { $_ > 0 });
my $range = $number as (Num where { $_ >= 0 && $_ <= 100 });

method ArrayRef[T] process<T>(ArrayRef[T] $data)
    where T: Serializable {
    return $data;
}
```

### Complex Constraints

Constraints can include multiple conditions:

```perl
method U transform<T, U>(T $input)
    where T: Serializable&Defined,
          U: Deserializable&!Undef {
    return deserialize($input->serialize());
}

method ProcessResult handle<T>(T $object)
    where T does EventHandler,
          T can 'process',
          T->VERSION >= 1.5 {
    return $object->process();
}
```

## Type Definitions

### Type Aliases

Create reusable type definitions:

```perl
type UserId = Int where { $_ > 0 };
type Email = Str where { $_ =~ /\@/ };
type Result<T, E> = Success<T> | Failure<E>;
```

### Complex Type Definitions

Define sophisticated type relationships:

```perl
type EventHandler<T> = CodeRef[T, Bool|Str];
type DataStore<K, V> = HashRef[ArrayRef[Tuple[K, V]]];
type Tree<T> = T|ArrayRef[Tree[T]];  # Recursive types

my DataStore[Str, MyClass] %store;
my EventHandler[ClickEvent] $click_handler;
my Tree[Int] $number_tree;
```

## Best Practices

### Progressive Adoption

Start with basic type annotations and gradually add more sophisticated features:

```perl
# Phase 1: Basic types
my Int $count = 0;
my Str $name = "example";

# Phase 2: Method signatures
method Int calculate(Int $a, Int $b) {
    return $a + $b;
}

# Phase 3: Advanced features
method ArrayRef[T] process<T>(ArrayRef[T] $input)
    where T: Serializable {
    return $input->map(sub { $_->clone() });
}

# Phase 4: Full class architecture
class DataProcessor<T> does Cacheable
    where T: Serializable&Defined {
    # Implementation...
}
```

### Backward Compatibility

Type annotations are completely optional. Existing untyped Perl code continues to work unchanged:

```perl
# This untyped code works exactly as before
my $var = "value";
sub function { return shift() * 2; }
for my $item (@list) { process($item); }

# Mixed with typed code
my Int $typed = 42;
my $untyped = "still works";
```

### Performance Considerations

- Type annotations have minimal runtime overhead
- Complex type expressions may increase parse time slightly
- Type assertions add runtime validation cost
- Use type constraints judiciously in performance-critical code

## Error Handling

### Type Annotation Errors

Common type annotation syntax errors and their solutions:

```perl
# Error: Missing closing bracket
my ArrayRef[Int $var;  # Wrong
my ArrayRef[Int] $var; # Correct

# Error: Invalid union syntax
my Int||Str $bad;      # Wrong
my Int|Str $good;      # Correct

# Error: Malformed constraint
my Int where $invalid; # Wrong
my Int where { $_ > 0 } $valid; # Correct
```

### Type Assertion Failures

Type assertions can fail at runtime:

```perl
my $value = $input as Int;  # Dies if $input is not an Int

# Safe type assertion with fallback
my $value = eval { $input as Int } // 0;

# Conditional type assertion
if ($input ~~ Int) {
    my $typed = $input as Int;
    # Use $typed safely
}
```

## Integration with Tools

### LSP Support

The Language Server Protocol integration provides:

- Type annotation syntax highlighting
- Type error detection and reporting
- Auto-completion for type names
- Hover information for typed variables
- Go-to-definition for custom types

### Type Checker Integration

The PSC type checker uses type annotations for:

- Static type checking
- Type inference
- Error detection before runtime
- Optimization opportunities

### Editor Integration

Popular editors support PVM type annotations through:

- Syntax highlighting extensions
- Type-aware code completion
- Inline type error reporting
- Refactoring with type preservation

## Migration Guide

### Syntax Migration

**From deprecated `returns` syntax**:
```perl
# Old (deprecated)
method process(ArrayRef[Str] $data) returns ProcessResult {
    # Implementation
}

# New (preferred)
method ProcessResult process(ArrayRef[Str] $data) {
    # Implementation
}
```

**Arrow syntax is not supported**:
```perl
# Not supported - will cause parse errors
method process(ArrayRef[Str] $data) -> ProcessResult {  # ❌ Parse error
    # Implementation
}

# Use this instead
method ProcessResult process(ArrayRef[Str] $data) {     # ✅ Correct
    # Implementation
}
```

### Existing Projects

To add type annotations to existing projects:

1. **Start with critical functions**: Add types to your most important methods first
2. **Use gradual typing**: Mix typed and untyped code as needed
3. **Focus on interfaces**: Type public APIs before internal implementation
4. **Add constraints incrementally**: Start with basic types, add constraints later
5. **Update syntax**: Convert any legacy `returns` syntax to preferred format

### Testing Strategy

When adding type annotations:

1. **Maintain existing tests**: All current tests should continue to pass
2. **Add type-specific tests**: Test type constraints and assertions
3. **Test mixed scenarios**: Ensure typed and untyped code interact correctly
4. **Performance testing**: Verify type annotations don't degrade performance
5. **Syntax validation**: Ensure all method signatures use supported syntax

## Types::Standard Alignment

PVM's type system aligns with Perl's **Types::Standard** module for consistency and interoperability:

### Bool Type Constraints

The `Bool` type follows Types::Standard semantics, accepting only these values:

```perl
my Bool $valid_true = 1;        # ✅ Numeric one
my Bool $valid_false = 0;       # ✅ Numeric zero
my Bool $empty_string = "";     # ✅ Empty string
my Bool $undefined = undef;     # ✅ Undefined value

# These are NOT valid Bool values:
my Bool $invalid = 2;           # ❌ Other numbers
my Bool $invalid = "true";      # ❌ String literals
my Bool $invalid = "false";     # ❌ String literals
my Bool $invalid = -1;          # ❌ Negative numbers
```

**Important**: Unlike traditional Perl boolean context, `Int` values are **not** automatically compatible with `Bool` types. Use explicit conversion or appropriate type annotations.

### HashRef vs Map Distinction

Following Types::Standard conventions:

```perl
# HashRef: constrains VALUES only (keys are always strings)
my HashRef[Int] %scores = (alice => 95, bob => 87);     # ✅ Value constraint only
my HashRef[UserData] %users = (id1 => $user1);         # ✅ Any string keys

# Map: constrains BOTH keys and values
my Map[UserID, UserData] %user_map = (123 => $user1);  # ✅ Key and value constraints
my Map[Str, Int] %string_to_int = (key => 42);         # ✅ Explicit key type

# This syntax is no longer supported:
my HashRef[UserID, UserData] %old_syntax;              # ❌ Use Map instead
```

### Migration from Legacy Syntax

**Old HashRef syntax**:
```perl
# Before (no longer supported)
my HashRef[Str, Int] %mapping = (key => 42);           # ❌ Error

# After (Types::Standard compliant)
my Map[Str, Int] %mapping = (key => 42);               # ✅ Correct
```

**Type compatibility changes**:
```perl
# Before: Int could be assigned to Bool
my Bool $flag = 42;                                     # ❌ No longer valid

# After: Explicit Bool values required
my Bool $flag = 1;                                      # ✅ Valid Bool value
my Bool $flag = ($value > 0) ? 1 : 0;                 # ✅ Explicit conversion
```

This alignment ensures compatibility with existing Perl type checking tools and provides more precise type semantics.

## Conclusion

PVM's type annotation system provides powerful tools for improving Perl code quality while maintaining full backward compatibility. Start with simple type annotations and gradually adopt more advanced features as your team becomes comfortable with the system.

For more information, see:
- [Typed Perl Specification](typed-perl-specification.md)
- [Development Guide](developer-guide.md)
- [Migration Guide](migration-guide.md)
