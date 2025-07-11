---
category: untyped-perl
subcategory: expressions
tags:
    - array
    - indexing
    - arithmetic
    - multiple_operators
---

# Array Element Expression

Array element access in arithmetic expression

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

# Expected Compilation Outcomes

## Array Element Expression

### Clean Perl Output

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

### Typed Perl Output

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
