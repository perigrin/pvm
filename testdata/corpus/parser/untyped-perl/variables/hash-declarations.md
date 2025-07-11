---
category: untyped-perl
subcategory: variables
tags:
    - hashes
    - declarations
    - package-qualification
    - scoping
    - variables
---

# Hash Declarations

Test hash variable declarations with different scoping and assignment patterns

```perl
my %simple_hash;
my %assigned = (key => 'value');
our %package_hash;
%Package::qualified = (a => 1, b => 2);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my %simple_hash;
my %assigned = (key => 'value');
our %package_hash;
%Package::qualified = (a => 1, b => 2);
```

## Typed Perl Output

```perl
my %simple_hash;
my %assigned = (key => 'value');
our %package_hash;
%Package::qualified = (a => 1, b => 2);
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
