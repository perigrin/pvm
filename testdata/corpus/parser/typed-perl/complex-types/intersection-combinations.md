---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - intersection-types
    - complex-combinations
    - parameterized-types
---

# Intersection Combinations

Intersection types combined with parameterized and union types

```perl
my ArrayRef[Object&Serializable] @serializable_list;
my HashRef[ArrayRef[Int|Str]&Defined] %defined_arrays;
my Container[Data&Validated&Cached] $safe_container;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @serializable_list;
my %defined_arrays;
my $safe_container;
```

## Typed Perl Output

```perl
my ArrayRef[Object&Serializable] @serializable_list;
my HashRef[ArrayRef[Int|Str]&Defined] %defined_arrays;
my Container[Data&Validated&Cached] $safe_container;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
