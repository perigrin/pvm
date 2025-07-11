---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - smartmatch
    - pattern
---

# Smartmatch Operator

Smartmatch operator usage

```perl
if ($value ~~ @valid_values) {
    process($value);
} elsif ($value ~~ /pattern/) {
    handle_pattern($value);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($value ~~ @valid_values) {
    process($value);
} elsif ($value ~~ /pattern/) {
    handle_pattern($value);
}
```

## Typed Perl Output

```perl
if ($value ~~ @valid_values) {
    process($value);
} elsif ($value ~~ /pattern/) {
    handle_pattern($value);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
