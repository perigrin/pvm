---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - try_catch
    - eval
    - error_handling
    - local
---

# Try Catch Finally

Try-catch-finally pattern with eval

```perl
{
    local $@;
    eval {
        dangerous_operation();
    };
    if ($@) {
        handle_exception($@);
    } else {
        success_handler();
    }
    # cleanup always runs
    cleanup();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
{
    local $@;
    eval {
        dangerous_operation();
    };
    if ($@) {
        handle_exception($@);
    } else {
        success_handler();
    }
    # cleanup always runs
    cleanup();
}
```

## Typed Perl Output

```perl
{
    local $@;
    eval {
        dangerous_operation();
    };
    if ($@) {
        handle_exception($@);
    } else {
        success_handler();
    }
    # cleanup always runs
    cleanup();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
