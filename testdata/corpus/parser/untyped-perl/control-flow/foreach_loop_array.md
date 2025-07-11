---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - array
---

# Foreach Loop Array

Foreach loop over array

```perl
foreach my $item (@list) {
    process($item);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $item (@list) {
    process($item);
}
```

## Typed Perl Output

```perl
foreach my $item (@list) {
    process($item);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
