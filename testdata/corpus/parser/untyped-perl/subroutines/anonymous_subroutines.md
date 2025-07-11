---
category: untyped-perl
subcategory: subroutines
tags:
    - anonymous
    - code_references
    - closures
    - parameters
---

# Anonymous Subroutines

Test anonymous subroutine definitions and code references

```perl
my $code_ref = sub {
    my $param = shift;
    return $param * 2;
};

my $closure = sub {
    my $x = shift;
    return sub { return $x + shift };
};

my $anon_with_params = sub {
    my ($a, $b) = @_;
    return $a . $b;
};
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $code_ref = sub {
    my $param = shift;
    return $param * 2;
};

my $closure = sub {
    my $x = shift;
    return sub { return $x + shift };
};

my $anon_with_params = sub {
    my ($a, $b) = @_;
    return $a . $b;
};
```

## Typed Perl Output

```perl
my $code_ref = sub {
    my $param = shift;
    return $param * 2;
};

my $closure = sub {
    my $x = shift;
    return sub { return $x + shift };
};

my $anon_with_params = sub {
    my ($a, $b) = @_;
    return $a . $b;
};
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
