---
category: untyped-perl
subcategory: control-flow
tags:
    - control-flow
    - conditionals
    - if-statements
---

# Simple If Statement

Basic if statement with conditional logic

```perl
if ($age >= 18) {
    print "You are an adult\n";
} else {
    print "You are a minor\n";
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($age >= 18) {
    print "You are an adult\n";
} else {
    print "You are a minor\n";
}
```

## Typed Perl Output

```perl
if ($age >= 18) {
    print "You are an adult\n";
} else {
    print "You are a minor\n";
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
