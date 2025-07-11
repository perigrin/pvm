---
category: untyped-perl
subcategory: expressions
tags:
    - postfix
    - increment
    - side_effects
    - indexing
---

# Postfix Increment

Postfix increment in expression

```perl
$result = $array[$counter++] + $hash{$key++};
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = $array[$counter++] + $hash{$key++};
```

### Typed Perl Output

```perl
$result = $array[$counter++] + $hash{$key++};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
