---
category: untyped-perl
subcategory: expressions
tags:
    - expressions
    - arithmetic
    - assignment
---

# Basic Expressions

Basic arithmetic and assignment expressions

```perl
my $a = 10;
my $b = 5;
my $sum = $a + $b;
my $product = $a * $b;
my $difference = $a - $b;
my $quotient = $a / $b;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $a = 10;
my $b = 5;
my $sum = $a + $b;
my $product = $a * $b;
my $difference = $a - $b;
my $quotient = $a / $b;
```

## Typed Perl Output

```perl
my $a = 10;
my $b = 5;
my $sum = $a + $b;
my $product = $a * $b;
my $difference = $a - $b;
my $quotient = $a / $b;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
