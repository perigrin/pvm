---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - chained
    - and
    - multiple
---

# Chained Logical

Chained logical AND operations

```perl
$valid = $a && $b && $c && $d;
```

# Expected Compilation Outcomes

## Chained Logical

### Clean Perl Output

```perl
$valid = $a && $b && $c && $d;
```

### Typed Perl Output

```perl
$valid = $a && $b && $c && $d;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
