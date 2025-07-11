---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - unless
---

# Unless Statement

Unless conditional statement

```perl
unless ($negative_condition) {
    execute();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
unless ($negative_condition) {
    execute();
}
```

## Typed Perl Output

```perl
unless ($negative_condition) {
    execute();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
