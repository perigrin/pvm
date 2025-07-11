---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - union-types
    - parameterized-types
    - complex-combinations
---

# Nested Unions In Parameterized

Union types nested within parameterized types

```perl
my ArrayRef[Int|Str|Bool] @complex_array;
my HashRef[ArrayRef[Int]|HashRef[Str]] %nested_complex;
my Map[Str, Int|Undef] %optional_values;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @complex_array;
my %nested_complex;
my %optional_values;
```

## Typed Perl Output

```perl
my ArrayRef[Int|Str|Bool] @complex_array;
my HashRef[ArrayRef[Int]|HashRef[Str]] %nested_complex;
my Map[Str, Int|Undef] %optional_values;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
