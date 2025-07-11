---
category: untyped-perl
subcategory: expressions
tags:
    - slice
    - array
    - range
    - indexing
---

# Slice Expression

Array slice with range operator

```perl
@subset = @array[$start .. $end];
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
@subset = @array[$start .. $end];
```

### Typed Perl Output

```perl
@subset = @array[$start .. $end];
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
