---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - given
    - when
    - array
    - qw
    - smartmatch
---

# Given When Arrays

Given-when with array reference matching

```perl
given ($day) {
    when ([qw(sat sun)]) { print "weekend"; }
    when ([qw(mon tue wed thu fri)]) { print "weekday"; }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
given ($day) {
    when ([qw(sat sun)]) { print "weekend"; }
    when ([qw(mon tue wed thu fri)]) { print "weekday"; }
}
```

## Typed Perl Output

```perl
given ($day) {
    when ([qw(sat sun)]) { print "weekend"; }
    when ([qw(mon tue wed thu fri)]) { print "weekday"; }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
