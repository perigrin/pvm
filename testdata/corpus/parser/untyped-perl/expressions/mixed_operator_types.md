---
category: untyped-perl
subcategory: expressions
tags:
    - mixed_operators
    - string
    - arithmetic
    - comparison
---

# Mixed Operator Types

Expression mixing string, arithmetic, and comparison operators

```perl
$result = ($str . '_' . $num) eq ($prefix . ($count + 1));
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = ($str . '_' . $num) eq ($prefix . ($count + 1));
```

### Typed Perl Output

```perl
$result = ($str . '_' . $num) eq ($prefix . ($count + 1));
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
