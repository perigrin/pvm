---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - redo
    - restart
---

# Redo Statement

Redo statement to restart current iteration

```perl
while ($condition) {
    my $input = get_input();
    redo if invalid($input);
    process($input);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while ($condition) {
    my $input = get_input();
    redo if invalid($input);
    process($input);
}
```

## Typed Perl Output

```perl
while ($condition) {
    my $input = get_input();
    redo if invalid($input);
    process($input);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
