---
category: untyped-perl
subcategory: expressions
tags:
    - mixed
    - numeric
    - string
    - comparison
    - equality
---

# Mixed Numeric String

Mixed numeric and string comparisons

```perl
$result = ($num == 42) && ($str eq 'hello');
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = ($num == 42) && ($str eq 'hello');
```

### Typed Perl Output

```perl
$result = ($num == 42) && ($str eq 'hello');
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
