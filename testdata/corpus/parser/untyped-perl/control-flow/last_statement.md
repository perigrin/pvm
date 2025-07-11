---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - last
    - loop_control
---

# Last Statement

Last statement to break out of loop

```perl
foreach my $item (@list) {
    process($item);
    last if stop_condition($item);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $item (@list) {
    process($item);
    last if stop_condition($item);
}
```

## Typed Perl Output

```perl
foreach my $item (@list) {
    process($item);
    last if stop_condition($item);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
