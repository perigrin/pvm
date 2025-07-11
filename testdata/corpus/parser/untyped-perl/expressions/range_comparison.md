---
category: untyped-perl
subcategory: expressions
tags:
    - range
    - comparison
    - multiple
    - ordering
---

# Range Comparison

Range check using multiple comparisons

```perl
$in_range = ($min <= $value) && ($value <= $max);
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$in_range = ($min <= $value) && ($value <= $max);
```

### Typed Perl Output

```perl
$in_range = ($min <= $value) && ($value <= $max);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
