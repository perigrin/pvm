---
category: untyped-perl
subcategory: control-flow
tags:
    - control-flow
    - loops
    - foreach
---

# Simple Loop

Basic foreach loop over an array

```perl
my @items = ("apple", "banana", "cherry");
for my $item (@items) {
    print "Item: $item\n";
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @items = ("apple", "banana", "cherry");
for my $item (@items) {
    print "Item: $item\n";
}
```

## Typed Perl Output

```perl
my @items = ("apple", "banana", "cherry");
for my $item (@items) {
    print "Item: $item\n";
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
