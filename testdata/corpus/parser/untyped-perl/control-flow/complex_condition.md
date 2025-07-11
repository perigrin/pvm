---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - conditional
    - complex_expression
    - logical
---

# Complex Condition

Conditional with complex boolean expression

```perl
if (($a > 0) && ($b < 100) || ($c == $d)) {
    complex_action();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if (($a > 0) && ($b < 100) || ($c == $d)) {
    complex_action();
}
```

## Typed Perl Output

```perl
if (($a > 0) && ($b < 100) || ($c == $d)) {
    complex_action();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
