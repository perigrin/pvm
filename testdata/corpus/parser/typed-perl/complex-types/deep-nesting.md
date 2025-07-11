---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - deep-nesting
    - parameterized-types
    - complex-combinations
---

# Deep Nesting

Deeply nested parameterized types with complex combinations

```perl
my ArrayRef[HashRef[ArrayRef[Int|Str]]] @deep_nested;
my Map[Str, ArrayRef[Tuple[Int, Bool|Str]]] %complex_map;
my Container[Wrapper[Inner[Data[Value]]]] $deeply_nested;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @deep_nested;
my %complex_map;
my $deeply_nested;
```

## Typed Perl Output

```perl
my ArrayRef[HashRef[ArrayRef[Int|Str]]] @deep_nested;
my Map[Str, ArrayRef[Tuple[Int, Bool|Str]]] %complex_map;
my Container[Wrapper[Inner[Data[Value]]]] $deeply_nested;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
