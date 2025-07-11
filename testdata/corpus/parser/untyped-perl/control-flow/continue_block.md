---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - while
    - continue
    - cleanup
---

# Continue Block

While loop with continue block

```perl
while ($condition) {
    process();
} continue {
    cleanup();
    update_condition();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while ($condition) {
    process();
} continue {
    cleanup();
    update_condition();
}
```

## Typed Perl Output

```perl
while ($condition) {
    process();
} continue {
    cleanup();
    update_condition();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
