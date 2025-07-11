---
category: untyped-perl
subcategory: expressions
tags:
    - parentheses
    - nested
    - complex
    - precedence
---

# Deeply Nested Parentheses

Deeply nested parenthesized expression

```perl
$result = ((($a + $b) * ($c - $d)) / (($e + $f) || 1));
```

# Expected Compilation Outcomes

## Deeply Nested Parentheses

### Clean Perl Output

```perl
$result = ((($a + $b) * ($c - $d)) / (($e + $f) || 1));
```

### Typed Perl Output

```perl
$result = ((($a + $b) * ($c - $d)) / (($e + $f) || 1));
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
