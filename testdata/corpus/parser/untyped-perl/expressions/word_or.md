---
category: untyped-perl
subcategory: expressions
tags:
    - word_form
    - or
    - logical
    - low_precedence
---

# Word Or

Word form of logical OR with lower precedence

```perl
$or_result = $x or $y;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$or_result = $x or $y;
```

### Typed Perl Output

```perl
$or_result = $x or $y;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
