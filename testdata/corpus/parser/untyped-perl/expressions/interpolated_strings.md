---
category: untyped-perl
subcategory: expressions
tags:
    - string
    - interpolation
    - variables
    - quoting
---

# Interpolated Strings

String interpolation in double quotes

```perl
$message = "Hello $name, your score is $score";
```

# Expected Compilation Outcomes

## Interpolated Strings

### Clean Perl Output

```perl
$message = "Hello $name, your score is $score";
```

### Typed Perl Output

```perl
$message = "Hello $name, your score is $score";
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
