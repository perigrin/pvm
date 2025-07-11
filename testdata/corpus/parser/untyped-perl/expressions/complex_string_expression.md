---
category: untyped-perl
subcategory: expressions
tags:
    - string
    - complex
    - concatenation
    - repetition
---

# Complex String Expression

Complex string expression with multiple operators

```perl
$result = ($prefix . $name) x 3 . $suffix;
```

# Expected Compilation Outcomes

## Complex String Expression

### Clean Perl Output

```perl
$result = ($prefix . $name) x 3 . $suffix;
```

### Typed Perl Output

```perl
$result = ($prefix . $name) x 3 . $suffix;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
