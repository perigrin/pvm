---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - exception
    - error_handling
    - eval
    - do_block
---

# Exception Handling Eval

Exception handling with eval block

```perl
eval {
    risky_operation();
    1;
} or do {
    my $error = $@ || 'Unknown error';
    handle_error($error);
};
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
eval {
    risky_operation();
    1;
} or do {
    my $error = $@ || 'Unknown error';
    handle_error($error);
};
```

## Typed Perl Output

```perl
eval {
    risky_operation();
    1;
} or do {
    my $error = $@ || 'Unknown error';
    handle_error($error);
};
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
