---
category: untyped-perl
subcategory: subroutines
tags:
    - basic
    - definitions
    - parameters
    - qualified
---

# Basic Subroutine Definitions

Test basic subroutine definitions and simple subroutines

```perl
sub simple_sub {
    return "result";
}

sub with_params {
    my ($first, $second) = @_;
    return $first + $second;
}

sub Package::qualified_sub {
    return 42;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub simple_sub {
    return "result";
}

sub with_params {
    my ($first, $second) = @_;
    return $first + $second;
}

sub Package::qualified_sub {
    return 42;
}
```

## Typed Perl Output

```perl
sub simple_sub {
    return "result";
}

sub with_params {
    my ($first, $second) = @_;
    return $first + $second;
}

sub Package::qualified_sub {
    return 42;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
