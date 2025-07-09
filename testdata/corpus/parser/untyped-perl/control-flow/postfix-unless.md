---
category: untyped-perl
subcategory: control-flow
tags:
    - postfix
    - unless
    - conditional
---

# Postfix Unless

Postfix unless conditional statement

```perl
print "Error" unless $success;
$retry++ unless $done;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
print "Error" unless $success;
$retry++ unless $done;
```

## Typed Perl Output

```perl
print "Error" unless $success;
$retry++ unless $done;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
