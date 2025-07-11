---
category: untyped-perl
subcategory: expressions
tags:
    - bit_manipulation
    - bitwise
    - flag_toggling
    - xor
---

# Bit Toggling

Toggling a specific bit using XOR

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

# Expected Compilation Outcomes

## Bit Toggling

### Clean Perl Output

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
