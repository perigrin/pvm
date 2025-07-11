---
category: untyped-perl
subcategory: subroutines
tags:
    - prototypes
    - attributes
    - lvalue
    - const
    - method
---

# Prototypes And Attributes

Test subroutine prototypes and attributes

```perl
# Prototypes
sub prototype_sub ($$) {
    my ($a, $b) = @_;
    return $a + $b;
}

sub arrayref_proto (\@) {
    my $arrayref = shift;
    return @$arrayref;
}

sub optional_proto ($;$) {
    my ($required, $optional) = @_;
    return defined $optional ? $required + $optional : $required;
}

# Attributes
sub attributed_sub : lvalue {
    return $global_var;
}

sub multi_attr : method lvalue {
    return $_[0]->{value};
}

sub const_sub : const {
    return 42;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
# Prototypes
sub prototype_sub ($$) {
    my ($a, $b) = @_;
    return $a + $b;
}

sub arrayref_proto (\@) {
    my $arrayref = shift;
    return @$arrayref;
}

sub optional_proto ($;$) {
    my ($required, $optional) = @_;
    return defined $optional ? $required + $optional : $required;
}

# Attributes
sub attributed_sub : lvalue {
    return $global_var;
}

sub multi_attr : method lvalue {
    return $_[0]->{value};
}

sub const_sub : const {
    return 42;
}
```

## Typed Perl Output

```perl
# Prototypes
sub prototype_sub ($$) {
    my ($a, $b) = @_;
    return $a + $b;
}

sub arrayref_proto (\@) {
    my $arrayref = shift;
    return @$arrayref;
}

sub optional_proto ($;$) {
    my ($required, $optional) = @_;
    return defined $optional ? $required + $optional : $required;
}

# Attributes
sub attributed_sub : lvalue {
    return $global_var;
}

sub multi_attr : method lvalue {
    return $_[0]->{value};
}

sub const_sub : const {
    return 42;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
