---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - for
    - list
    - qw
---

# For Loop List

For loop over literal list

```perl
for my $item (qw(apple banana cherry)) {
    print "$item\n";
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for my $item (qw(apple banana cherry)) {
    print "$item\n";
}
```

## Typed Perl Output

```perl
for my $item (qw(apple banana cherry)) {
    print "$item\n";
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
