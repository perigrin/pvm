---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - complex
    - precedence
    - parentheses
---

# Complex Logical

Complex logical expression with precedence

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

# Expected Compilation Outcomes

## Complex Logical

### Clean Perl Output

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

### Typed Perl Output

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
