---
category: untyped-perl
subcategory: subroutines
tags:
    - complex
    - nested
    - recursive
    - parameters
    - context
    - edge_cases
---

# Complex Subroutines

Test complex subroutine patterns and edge cases

```perl
# Nested subroutine definitions
sub outer_function {
    my $param = shift;

    my $inner = sub {
        return $param * 2;
    };

    return $inner;
}

# Recursive subroutines
sub factorial {
    my $n = shift;
    return $n <= 1 ? 1 : $n * factorial($n - 1);
}

# Subroutines with complex parameter handling
sub complex_params {
    my ($first, @rest) = @_;
    my %options = @rest;
    return process($first, \%options);
}

# Subroutines returning different types
sub context_sensitive {
    return wantarray ? ('list', 'result') : 'scalar result';
}

# Forward declaration usage
sub uses_forward {
    return forward_declared(42);
}

sub forward_declared {
    my $value = shift;
    return $value * 3;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
# Nested subroutine definitions
sub outer_function {
    my $param = shift;

    my $inner = sub {
        return $param * 2;
    };

    return $inner;
}

# Recursive subroutines
sub factorial {
    my $n = shift;
    return $n <= 1 ? 1 : $n * factorial($n - 1);
}

# Subroutines with complex parameter handling
sub complex_params {
    my ($first, @rest) = @_;
    my %options = @rest;
    return process($first, \%options);
}

# Subroutines returning different types
sub context_sensitive {
    return wantarray ? ('list', 'result') : 'scalar result';
}

# Forward declaration usage
sub uses_forward {
    return forward_declared(42);
}

sub forward_declared {
    my $value = shift;
    return $value * 3;
}
```

## Typed Perl Output

```perl
# Nested subroutine definitions
sub outer_function {
    my $param = shift;

    my $inner = sub {
        return $param * 2;
    };

    return $inner;
}

# Recursive subroutines
sub factorial {
    my $n = shift;
    return $n <= 1 ? 1 : $n * factorial($n - 1);
}

# Subroutines with complex parameter handling
sub complex_params {
    my ($first, @rest) = @_;
    my %options = @rest;
    return process($first, \%options);
}

# Subroutines returning different types
sub context_sensitive {
    return wantarray ? ('list', 'result') : 'scalar result';
}

# Forward declaration usage
sub uses_forward {
    return forward_declared(42);
}

sub forward_declared {
    my $value = shift;
    return $value * 3;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
