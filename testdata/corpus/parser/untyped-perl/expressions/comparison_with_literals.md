---
category: untyped-perl
subcategory: expressions
tags:
    - comparison
    - literals
    - mixed
    - logical
---

# Comparison With Literals

Comparisons with literal values

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

# Expected Compilation Outcomes

## Comparison With Literals

### Clean Perl Output

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

### Typed Perl Output

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
