---
category: untyped-perl
subcategory: variables
tags:
    - arrays
    - variables
    - assignments
---

# Array Operations

Basic array operations and assignments

```perl
my @numbers = (1, 2, 3, 4, 5);
my $first = $;
my $count = @numbers;
push @numbers, 6;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @numbers = (1, 2, 3, 4, 5);
my $first = $;
my $count = @numbers;
push @numbers, 6;
```

## Typed Perl Output

```perl
my @numbers = (1, 2, 3, 4, 5);
my $first = $;
my $count = @numbers;
push @numbers, 6;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
