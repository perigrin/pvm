---
category: untyped-perl
subcategory: variables
tags:
    - scalars
    - declarations
    - scoping
    - package-qualification
    - variables
---

# Scalar Declarations

Test scalar variable declarations with different scoping keywords

```perl
my $simple;
my $assigned = 42;
our $package_var;
state $persistent = "initial";
local $localized;
$Package::qualified = "value";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $simple;
my $assigned = 42;
our $package_var;
state $persistent = "initial";
local $localized;
$Package::qualified = "value";
```

## Typed Perl Output

```perl
my $simple;
my $assigned = 42;
our $package_var;
state $persistent = "initial";
local $localized;
$Package::qualified = "value";
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
