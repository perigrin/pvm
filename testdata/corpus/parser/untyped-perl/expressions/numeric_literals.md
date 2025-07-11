---
category: untyped-perl
subcategory: expressions
tags:
    - numeric
    - literals
    - arithmetic
    - hexadecimal
---

# Numeric Literals

Arithmetic with various numeric literal formats

```perl
$result = 42 + 3.14 + 1e5 + 0xFF;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = 42 + 3.14 + 1e5 + 0xFF;
```

### Typed Perl Output

```perl
$result = 42 + 3.14 + 1e5 + 0xFF;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
