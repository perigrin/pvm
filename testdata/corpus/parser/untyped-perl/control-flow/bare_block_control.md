---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - bare_block
    - loop_control
    - last
---

# Bare Block Control

Loop control in bare block

```perl
{
    my $x = calculate();
    last if $x > 100;
    process($x);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
{
    my $x = calculate();
    last if $x > 100;
    process($x);
}
```

## Typed Perl Output

```perl
{
    my $x = calculate();
    last if $x > 100;
    process($x);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
