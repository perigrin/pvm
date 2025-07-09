---
category: untyped-perl
subcategory: control-flow
tags:
    - postfix
    - if
    - conditional
---

# Postfix If

Postfix if conditional statement

```perl
print "Hello" if $debug;
$count++ if $item;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
print "Hello" if $debug;
$count++ if $item;
```

## Typed Perl Output

```perl
print "Hello" if $debug;
$count++ if $item;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
