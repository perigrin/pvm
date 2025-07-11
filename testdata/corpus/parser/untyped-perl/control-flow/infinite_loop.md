---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - while
    - infinite
    - last
---

# Infinite Loop

Infinite loop with break condition

```perl
while (1) {
    handle_request();
    last if $shutdown;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while (1) {
    handle_request();
    last if $shutdown;
}
```

## Typed Perl Output

```perl
while (1) {
    handle_request();
    last if $shutdown;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
