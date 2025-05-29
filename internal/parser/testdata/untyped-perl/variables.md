---
category: untyped-perl
subcategory: variables
tags:
    - arrays
    - context
    - declarations
    - edge-cases
    - expressions
    - hashes
    - naming
    - package-qualification
    - scalars
    - scoping
    - special-vars
    - variables
---

# Array Declarations

Test array variable declarations with different scoping and assignment patterns

```perl
my @simple_array;
my @assigned = (1, 2, 3);
our @package_array;
@Package::qualified = qw(a b c);
```

## Complex Variable Names

Test variables with underscores, numbers, and package qualifiers

```perl
my $var_with_underscores;
my $var123;
my $package_var;
$Some::Package::qualified_var = "test";
our $_private_var;
my $CamelCase_Var;
```

## Hash Declarations

Test hash variable declarations with different scoping and assignment patterns

```perl
my %simple_hash;
my %assigned = (key => 'value');
our %package_hash;
%Package::qualified = (a => 1, b => 2);
```

## Scalar Declarations

Test scalar variable declarations with different scoping keywords

```perl
my $simple;
my $assigned = 42;
our $package_var;
state $persistent = "initial";
local $localized;
$Package::qualified = "value";
```

## Variable Context Variations

Test variable declarations in different contexts (statement, expression)

```perl
if (my $scoped = get_value()) { print $scoped; }
for my $iterator (1..10) { print $iterator; }
while (my $line = <DATA>) { chomp $line; }
```

## Variable Edge Cases

Test edge cases like empty declarations and special variable names

```perl
my $;
my @;
my %;
my $$;
my $0;
my $var = undef;
my @empty = ();
my %empty = ();
```
