---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - given
    - when
    - regex
    - pattern
---

# Given When Regex

Given-when with regex matching

```perl
given ($input) {
    when (/^\d+$/) { print "number"; }
    when (/^[a-zA-Z]+$/) { print "letters"; }
    when (/^\s*$/) { print "empty"; }
    default { print "mixed"; }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
given ($input) {
    when (/^\d+$/) { print "number"; }
    when (/^[a-zA-Z]+$/) { print "letters"; }
    when (/^\s*$/) { print "empty"; }
    default { print "mixed"; }
}
```

## Typed Perl Output

```perl
given ($input) {
    when (/^\d+$/) { print "number"; }
    when (/^[a-zA-Z]+$/) { print "letters"; }
    when (/^\s*$/) { print "empty"; }
    default { print "mixed"; }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
