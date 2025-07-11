---
category: untyped-perl
subcategory: expressions
tags:
    - bitwise
    - hexadecimal
    - mask
    - literals
---

# Bitwise With Hex

Bitwise operation with hexadecimal literal

```perl
$masked = $value & 0xFF00;
```

# Expected Compilation Outcomes

## Bitwise With Hex

### Clean Perl Output

```perl
$masked = $value & 0xFF00;
```

### Typed Perl Output

```perl
$masked = $value & 0xFF00;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
