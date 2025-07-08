---
category: untyped-perl
subcategory: subroutines
tags:
    - subroutines
    - functions
    - basic
---

# Simple Subroutine

Basic subroutine definition and call

```perl
sub greet {
    my $name = shift;
    return "Hello, $name!";
}

my $message = greet("World");
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub greet {
    my $name = shift;
    return "Hello, $name!";
}

my $message = greet("World");
```

## Typed Perl Output

```perl
sub greet {
    my $name = shift;
    return "Hello, $name!";
}

my $message = greet("World");
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
