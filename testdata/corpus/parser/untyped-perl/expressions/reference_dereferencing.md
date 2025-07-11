---
category: untyped-perl
subcategory: expressions
tags:
    - references
    - dereferencing
    - arithmetic
---

# Reference Dereferencing

Reference dereferencing in expressions

```perl
$sum = ${$scalar_ref} + @{$array_ref} + keys %{$hash_ref};
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$sum = ${$scalar_ref} + @{$array_ref} + keys %{$hash_ref};
```

### Typed Perl Output

```perl
$sum = ${$scalar_ref} + @{$array_ref} + keys %{$hash_ref};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
