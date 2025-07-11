---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - if
    - conditional
    - block
---

# Basic If Statement

Basic if statement with block

```perl
if ($condition) {
    do_something();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($condition) {
    do_something();
}
```

## Typed Perl Output

```perl
if ($condition) {
    do_something();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
