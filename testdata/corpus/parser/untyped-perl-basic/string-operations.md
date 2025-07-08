---
category: untyped-perl
subcategory: expressions
tags:
    - strings
    - concatenation
    - interpolation
---

# String Operations

Basic string operations and interpolation

```perl
my $first = "Hello";
my $second = "World";
my $greeting = $first . ", " . $second . "!";
my $interpolated = "Message: $greeting";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $first = "Hello";
my $second = "World";
my $greeting = $first . ", " . $second . "!";
my $interpolated = "Message: $greeting";
```

## Typed Perl Output

```perl
my $first = "Hello";
my $second = "World";
my $greeting = $first . ", " . $second . "!";
my $interpolated = "Message: $greeting";
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
