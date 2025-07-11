---
category: untyped-perl
subcategory: expressions
tags:
    - prefix
    - decrement
    - increment
    - side_effects
---

# Prefix Decrement

Prefix increment and decrement in expression

```perl
$result = --$counter * ++$multiplier;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = --$counter * ++$multiplier;
```

### Typed Perl Output

```perl
$result = --$counter * ++$multiplier;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
