---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - labeled
    - last
    - search
    - nested
---

# Labeled Last

Labeled last to break out of outer loop

```perl
SEARCH: foreach my $item (@items) {
    foreach my $prop (@properties) {
        if ($item->{$prop} eq $target) {
            $found = $item;
            last SEARCH;
        }
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
SEARCH: foreach my $item (@items) {
    foreach my $prop (@properties) {
        if ($item->{$prop} eq $target) {
            $found = $item;
            last SEARCH;
        }
    }
}
```

## Typed Perl Output

```perl
SEARCH: foreach my $item (@items) {
    foreach my $prop (@properties) {
        if ($item->{$prop} eq $target) {
            $found = $item;
            last SEARCH;
        }
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
