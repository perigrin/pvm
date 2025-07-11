---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - negation-types
    - complex-combinations
    - intersection-types
---

# Negation Combinations

Negation types combined with parameterized and intersection types

```perl
my ArrayRef[!Undef] @non_undef_array;
my HashRef[Str, !Empty] %non_empty_values;
my Optional[!Null&!Undef] $definitely_defined;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @non_undef_array;
my %non_empty_values;
my $definitely_defined;
```

## Typed Perl Output

```perl
my ArrayRef[!Undef] @non_undef_array;
my HashRef[Str, !Empty] %non_empty_values;
my Optional[!Null&!Undef] $definitely_defined;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
