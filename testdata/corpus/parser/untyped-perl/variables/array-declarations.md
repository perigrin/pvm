---
category: untyped-perl
subcategory: variables
tags:
    - array
    - declarations
    - my
    - variables
---

# Array Declarations

Basic array variable declarations

```perl
my @items = ("apple", "banana", "cherry");
my @numbers = (1, 2, 3, 4, 5);
my @empty = ();
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @items = ("apple", "banana", "cherry");
my @numbers = (1, 2, 3, 4, 5);
my @empty = ();
```

## Typed Perl Output

```perl
my @items = ("apple", "banana", "cherry");
my @numbers = (1, 2, 3, 4, 5);
my @empty = ();
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
