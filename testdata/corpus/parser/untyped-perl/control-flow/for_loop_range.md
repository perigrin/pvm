---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - for
    - range
---

# For Loop Range

For loop with range operator

```perl
for my $i (0..$max) {
    handle($i);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for my $i (0..$max) {
    handle($i);
}
```

## Typed Perl Output

```perl
for my $i (0..$max) {
    handle($i);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
