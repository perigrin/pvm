---
category: untyped-perl
subcategory: expressions
tags:
    - bitwise
    - complex
    - multiple_operators
    - parentheses
---

# Complex Bitwise

Complex bitwise expression with multiple operations

```perl
$flags = ($a & $mask) | ($b << 4) | (~$c & 0xFF);
```

# Expected Compilation Outcomes

## Complex Bitwise

### Clean Perl Output

```perl
$flags = ($a & $mask) | ($b << 4) | (~$c & 0xFF);
```

### Typed Perl Output

```perl
$flags = ($a & $mask) | ($b << 4) | (~$c & 0xFF);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
