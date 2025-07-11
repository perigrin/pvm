---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - conditional
    - assignment
    - variable_declaration
---

# Conditional With Assignment

Conditional with variable assignment in condition

```perl
if (my $result = function_call()) {
    process($result);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if (my $result = function_call()) {
    process($result);
}
```

## Typed Perl Output

```perl
if (my $result = function_call()) {
    process($result);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
