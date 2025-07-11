---
category: untyped-perl
subcategory: variables
tags:
    - context
    - expressions
    - scoping
    - variables
---

# Variable Context Variations

Test variable declarations in different contexts (statement, expression)

```perl
if (my $scoped = get_value()) { print $scoped; }
for my $iterator (1..10) { print $iterator; }
while (my $line = <DATA>) { chomp $line; }
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if (my $scoped = get_value()) { print $scoped; }
for my $iterator (1..10) { print $iterator; }
while (my $line = <DATA>) { chomp $line; }
```

## Typed Perl Output

```perl
if (my $scoped = get_value()) { print $scoped; }
for my $iterator (1..10) { print $iterator; }
while (my $line = <DATA>) { chomp $line; }
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
