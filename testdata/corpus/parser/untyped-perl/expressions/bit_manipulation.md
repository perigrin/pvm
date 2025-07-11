---
category: untyped-perl
subcategory: expressions
tags:
    - bit_manipulation
    - bitwise
    - flag_setting
    - left_shift
---

# Bit Manipulation

Setting a specific bit using shift and OR

```perl
$bit_set = $flags | (1 << $bit_number);
```

# Expected Compilation Outcomes

## Bit Manipulation

### Clean Perl Output

```perl
$bit_set = $flags | (1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_set = $flags | (1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
