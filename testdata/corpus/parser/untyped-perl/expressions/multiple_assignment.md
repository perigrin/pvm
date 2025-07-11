---
category: untyped-perl
subcategory: expressions
tags:
    - multiple
    - assignment
    - list
---

# Multiple Assignment

Multiple assignment with list context

```perl
($a, $b, $c) = ($x, $y, $z);
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
($a, $b, $c) = ($x, $y, $z);
```

### Typed Perl Output

```perl
($a, $b, $c) = ($x, $y, $z);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
