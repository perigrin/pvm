---
category: untyped-perl
subcategory: variables
tags:
    - arrays
    - declarations
    - package-qualification
    - scoping
    - variables
---

# Array Declarations

Test array variable declarations with different scoping and assignment patterns

```perl
my @simple_array;
my @assigned = (1, 2, 3);
our @package_array;
@Package::qualified = qw(a b c);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @simple_array;
my @assigned = (1, 2, 3);
our @package_array;
@Package::qualified = qw(a b c);
```

## Typed Perl Output

```perl
my @simple_array;
my @assigned = (1, 2, 3);
our @package_array;
@Package::qualified = qw(a b c);
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
