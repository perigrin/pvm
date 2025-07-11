---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - nested
    - for
    - foreach
    - complex
---

# Nested Loops

Nested loops with complex data structures

```perl
for my $outer (@outer_list) {
    foreach my $inner (@{$outer->{items}}) {
        if ($inner->{valid}) {
            process($inner);
        }
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for my $outer (@outer_list) {
    foreach my $inner (@{$outer->{items}}) {
        if ($inner->{valid}) {
            process($inner);
        }
    }
}
```

## Typed Perl Output

```perl
for my $outer (@outer_list) {
    foreach my $inner (@{$outer->{items}}) {
        if ($inner->{valid}) {
            process($inner);
        }
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
