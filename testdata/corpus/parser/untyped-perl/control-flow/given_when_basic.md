---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - given
    - when
    - switch
    - default
---

# Given When Basic

Basic given-when switch statement

```perl
given ($value) {
    when (1) { print "one"; }
    when (2) { print "two"; }
    default { print "other"; }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
given ($value) {
    when (1) { print "one"; }
    when (2) { print "two"; }
    default { print "other"; }
}
```

## Typed Perl Output

```perl
given ($value) {
    when (1) { print "one"; }
    when (2) { print "two"; }
    default { print "other"; }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
