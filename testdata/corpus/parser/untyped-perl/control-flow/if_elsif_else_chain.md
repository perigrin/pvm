---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - if
    - elsif
    - else
    - chain
---

# If Elsif Else Chain

Complete if-elsif-else conditional chain

```perl
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
} else {
    do_default();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
} else {
    do_default();
}
```

## Typed Perl Output

```perl
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
} else {
    do_default();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
