---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - while
    - loop
---

# Basic While Loop

Basic while loop

```perl
while ($condition) {
    process();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while ($condition) {
    process();
}
```

## Typed Perl Output

```perl
while ($condition) {
    process();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
