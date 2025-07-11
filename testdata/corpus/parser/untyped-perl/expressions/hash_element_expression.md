---
category: untyped-perl
subcategory: expressions
tags:
    - hash
    - key_access
    - arithmetic
    - indexing
---

# Hash Element Expression

Hash element access in arithmetic expression

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

# Expected Compilation Outcomes

## Hash Element Expression

### Clean Perl Output

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

### Typed Perl Output

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
