---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - until
    - loop
---

# Basic Until Loop

Basic until loop

```perl
until ($done) {
    continue_processing();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
until ($done) {
    continue_processing();
}
```

## Typed Perl Output

```perl
until ($done) {
    continue_processing();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
