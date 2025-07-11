---
category: untyped-perl
subcategory: variables
tags:
    - naming
    - package-qualification
    - variables
---

# Complex Variable Names

Test variables with underscores, numbers, and package qualifiers

```perl
my $var_with_underscores;
my $var123;
my $package_var;
$Some::Package::qualified_var = "test";
our $_private_var;
my $CamelCase_Var;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $var_with_underscores;
my $var123;
my $package_var;
$Some::Package::qualified_var = "test";
our $_private_var;
my $CamelCase_Var;
```

## Typed Perl Output

```perl
my $var_with_underscores;
my $var123;
my $package_var;
$Some::Package::qualified_var = "test";
our $_private_var;
my $CamelCase_Var;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
