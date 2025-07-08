---
category: untyped-perl
subcategory: variables
tags:
    - hashes
    - variables
    - assignments
---

# Hash Operations

Basic hash operations and assignments

```perl
my %person = (
    name => "John",
    age => 30,
    city => "Boston"
);

my $name = $person{name};
$person{age} = 31;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my %person = (
    name => "John",
    age => 30,
    city => "Boston"
);

my $name = $person{name};
$person{age} = 31;
```

## Typed Perl Output

```perl
my %person = (
    name => "John",
    age => 30,
    city => "Boston"
);

my $name = $person{name};
$person{age} = 31;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
