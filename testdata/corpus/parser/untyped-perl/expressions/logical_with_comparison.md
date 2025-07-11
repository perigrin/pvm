---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - comparison
    - mixed_operators
    - complex
---

# Logical With Comparison

Logical operators combined with comparison operators

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

# Expected Compilation Outcomes

## Logical With Comparison

### Clean Perl Output

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

### Typed Perl Output

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
