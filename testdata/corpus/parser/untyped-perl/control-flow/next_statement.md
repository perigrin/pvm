---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - next
    - loop_control
---

# Next Statement

Next statement to skip to next iteration

```perl
foreach my $item (@list) {
    next if skip_condition($item);
    process($item);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $item (@list) {
    next if skip_condition($item);
    process($item);
}
```

## Typed Perl Output

```perl
foreach my $item (@list) {
    next if skip_condition($item);
    process($item);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
