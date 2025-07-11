---
category: untyped-perl
subcategory: expressions
tags:
    - method_calls
    - arithmetic
    - objects
    - function_calls
---

# Method Call Expression

Method calls in arithmetic expression

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

### Typed Perl Output

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
