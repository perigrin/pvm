---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - for
    - c_style
    - increment
    - initialization
---

# For Loop C Style

C-style for loop with initialization, condition, and increment

```perl
for (my $i = 0; $i < $max; $i++) {
    handle($i);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for (my $i = 0; $i < $max; $i++) {
    handle($i);
}
```

## Typed Perl Output

```perl
for (my $i = 0; $i < $max; $i++) {
    handle($i);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
