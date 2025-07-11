---
category: untyped-perl
subcategory: expressions
tags:
    - mixed
    - precedence
    - logical
    - low_precedence
    - word_form
---

# Mixed Precedence

Mixed logical operators with different precedence levels

```perl
$result = $a and $b || $c && $d or $e;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = $a and $b || $c && $d or $e;
```

### Typed Perl Output

```perl
$result = $a and $b || $c && $d or $e;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
