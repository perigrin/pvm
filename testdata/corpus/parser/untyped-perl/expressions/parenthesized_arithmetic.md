---
category: untyped-perl
subcategory: expressions
tags:
    - parentheses
    - arithmetic
    - precedence
---

# Parenthesized Arithmetic

Arithmetic with parentheses for precedence control

```perl
$result = ($a + $b) * ($c - $d) / ($e || 1);
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = ($a + $b) * ($c - $d) / ($e || 1);
```

### Typed Perl Output

```perl
$result = ($a + $b) * ($c - $d) / ($e || 1);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
