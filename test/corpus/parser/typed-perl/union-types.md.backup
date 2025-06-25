---
category: typed-perl
subcategory: union-types
tags:
    - complex-expressions
    - complex-parameters
    - custom-types
    - fields
    - formatting
    - method-signatures
    - multi-way-unions
    - nested-contexts
    - package-qualified
    - parameter-types
    - parameterized-types
    - return-types
    - simple-unions
    - union-types
    - variable-declarations
    - whitespace
---

# Complex Expressions

Union types within complex type expressions like parameterized types

```perl
my ArrayRef[Int|Str] @mixed_array;
field HashRef[Int|Bool] $mixed_hash;
my CodeRef[Int|Str, Bool|Undef] $flexible_function;
```

## Custom Types Unions

Union types with custom and package-qualified type names

```perl
my MyClass|OtherClass $object;
my Package::Type1|Package::Type2 $qualified;
my UserType|SystemType|DefaultType $flexible;
```

## Method Signatures Unions

Union types in method parameter and return type signatures

```perl
method process(Int|Str $input) returns Bool|Str {
    if (ref $input) {
        return "Invalid";
    }
    return 1;
}
```

## Multi Way Unions

Multi-way union types with three or more alternatives

```perl
my Int|Str|Bool $multi = "text";
my Num|ArrayRef|HashRef $complex;
my Int|Str|Bool|Undef $nullable = undef;
```

## Nested Contexts

Union types in nested method signature contexts with complex parameters

```perl
method complex(
    ArrayRef[Int|Str] $data,
    CodeRef|Undef $callback
) returns HashRef[Bool|Str] {
    return {};
}
```

## Simple Union Types

Simple union type expressions with two types

```perl
my Int|Str $flexible = 42;
my Bool|Undef $maybe_flag;
my Num|Str $mixed_value = "text";
```

## Whitespace Variations

Union types with different whitespace formatting variations

```perl
my Int | Str $spaced;
my Int|Str|Bool $compact;
my  Num  |  Str  |  Bool  $extra_spaced;
my Int|
    Str|
    Bool $multiline;
```
