---
category: error-cases
subcategory: complex-types
tags:
    - complex-expressions
    - error-recovery
    - malformed-types
---

# Complex Malformed Types

Malformed complex type expressions that should produce helpful error messages

<!-- should_error: true -->
<!-- expected_error: error[TSP001] -->

```perl
my ArrayRef[Int| $incomplete_union;
my HashRef[Str,, Int] $double_comma;
my Container[Wrapper[Missing $unclosed_nested;
my (Int|Str& $mixed_operators_without_close;
my Map[Str ArrayRef[Int]] $missing_comma_in_params;
```
