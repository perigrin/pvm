---
category: untyped-perl
subcategory: expressions
tags:
    - bit_manipulation
    - bitwise
    - flag_clearing
    - complex
---

# Bit Clearing

Clearing a specific bit using NOT and AND

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

# Expected Compilation Outcomes

## Bit Clearing

### Clean Perl Output

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
