---
category: untyped-perl
subcategory: expressions
tags:
    - list
    - assignment
    - function_calls
    - multiple
---

# List Assignment Expression

List assignment with function call

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

# Expected Compilation Outcomes

## List Assignment Expression

### Clean Perl Output

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

### Typed Perl Output

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
