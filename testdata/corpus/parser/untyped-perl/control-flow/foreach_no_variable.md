---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - default_variable
    - underscore
---

# Foreach No Variable

Foreach loop using default variable $_

```perl
foreach (@items) {
    process($_);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach (@items) {
    process($_);
}
```

## Typed Perl Output

```perl
foreach (@items) {
    process($_);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
