---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - hash
    - keys
---

# Foreach Loop Hash

Foreach loop over hash keys

```perl
foreach my $key (keys %hash) {
    process($key, $hash{$key});
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $key (keys %hash) {
    process($key, $hash{$key});
}
```

## Typed Perl Output

```perl
foreach my $key (keys %hash) {
    process($key, $hash{$key});
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
