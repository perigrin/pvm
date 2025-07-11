---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - continue
    - progress
---

# Foreach Continue

Foreach loop with continue block

```perl
foreach my $item (@items) {
    process($item);
} continue {
    log_progress();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $item (@items) {
    process($item);
} continue {
    log_progress();
}
```

## Typed Perl Output

```perl
foreach my $item (@items) {
    process($item);
} continue {
    log_progress();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
