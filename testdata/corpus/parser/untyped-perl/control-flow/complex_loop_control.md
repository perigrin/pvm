---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - nested
    - loop_control
    - next
    - last
---

# Complex Loop Control

Complex loop control with multiple conditions

```perl
foreach my $batch (@batches) {
    foreach my $item (@{$batch}) {
        next if $item->{skip};
        my $result = process($item);
        if ($result->{error}) {
            log_error($result->{error});
            next;
        }
        last if $result->{complete};
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $batch (@batches) {
    foreach my $item (@{$batch}) {
        next if $item->{skip};
        my $result = process($item);
        if ($result->{error}) {
            log_error($result->{error});
            next;
        }
        last if $result->{complete};
    }
}
```

## Typed Perl Output

```perl
foreach my $batch (@batches) {
    foreach my $item (@{$batch}) {
        next if $item->{skip};
        my $result = process($item);
        if ($result->{error}) {
            log_error($result->{error});
            next;
        }
        last if $result->{complete};
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
