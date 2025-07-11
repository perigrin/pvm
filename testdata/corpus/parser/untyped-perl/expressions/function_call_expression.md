---
category: untyped-perl
subcategory: expressions
tags:
    - function_calls
    - arithmetic
    - math
    - complex
---

# Function Call Expression

Function calls in arithmetic expression

```perl
$result = sqrt($value) + sin($angle * 3.14159 / 180);
```

# Expected Compilation Outcomes

## Function Call Expression

### Clean Perl Output

```perl
$result = sqrt($value) + sin($angle * 3.14159 / 180);
```

### Typed Perl Output

```perl
$result = sqrt($value) + sin($angle * 3.14159 / 180);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
