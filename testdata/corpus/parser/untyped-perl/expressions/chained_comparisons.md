---
category: untyped-perl
subcategory: expressions
tags:
    - comparison
    - chained
    - ordering
    - logical
---

# Chained Comparisons

Chained comparisons for ordering check

```perl
$ordered = $a < $b && $b < $c && $c < $d;
```

# Expected Compilation Outcomes

## Chained Comparisons

### Clean Perl Output

```perl
$ordered = $a < $b && $b < $c && $c < $d;
```

### Typed Perl Output

```perl
$ordered = $a < $b && $b < $c && $c < $d;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
