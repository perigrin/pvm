---
category: untyped-perl
subcategory: expressions
tags:
    - string
    - literals
    - quoting
---

# String Literals

String operations with different literal types

```perl
$result = 'single' . "double" . `backtick` . qq{quoted};
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = 'single' . "double" . `backtick` . qq{quoted};
```

### Typed Perl Output

```perl
$result = 'single' . "double" . `backtick` . qq{quoted};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
