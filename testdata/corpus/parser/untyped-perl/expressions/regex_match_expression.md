---
category: untyped-perl
subcategory: expressions
tags:
    - regex
    - matching
    - logical
---

# Regex Match Expression

Regular expression matching in logical expression

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

### Typed Perl Output

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
