---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - conditional
    - return
---

# Conditional With Return

Conditional with return statements

```perl
if ($error) {
    return $error_value;
}
return $success_value;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($error) {
    return $error_value;
}
return $success_value;
```

## Typed Perl Output

```perl
if ($error) {
    return $error_value;
}
return $success_value;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
