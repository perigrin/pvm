---
category: untyped-perl
subcategory: packages
tags:
    - variables
    - qualification
    - namespaces
    - references
    - global
---

# Package Variables

Test package-qualified variable declarations and usage

```perl
$MyPackage::variable = "value";
@Package::array = (1, 2, 3);
%Other::Package::hash = (key => 'value');
our $Package::qualified;
my $ref = \&MyPackage::function;
$main::global = "in main";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
$MyPackage::variable = "value";
@Package::array = (1, 2, 3);
%Other::Package::hash = (key => 'value');
our $Package::qualified;
my $ref = \&MyPackage::function;
$main::global = "in main";
```

## Typed Perl Output

```perl
$MyPackage::variable = "value";
@Package::array = (1, 2, 3);
%Other::Package::hash = (key => 'value');
our $Package::qualified;
my $ref = \&MyPackage::function;
$main::global = "in main";
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
