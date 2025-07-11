---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - state_machine
    - given
    - when
    - loop
---

# State Machine Loop

State machine implementation with loop and switch

```perl
my $state = 'start';
while (1) {
    given ($state) {
        when ('start') {
            initialize();
            $state = 'process';
        }
        when ('process') {
            if (process_item()) {
                $state = 'finish';
            } else {
                $state = 'error';
            }
        }
        when ('error') {
            handle_error();
            $state = 'start';
        }
        when ('finish') {
            finalize();
            last;
        }
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $state = 'start';
while (1) {
    given ($state) {
        when ('start') {
            initialize();
            $state = 'process';
        }
        when ('process') {
            if (process_item()) {
                $state = 'finish';
            } else {
                $state = 'error';
            }
        }
        when ('error') {
            handle_error();
            $state = 'start';
        }
        when ('finish') {
            finalize();
            last;
        }
    }
}
```

## Typed Perl Output

```perl
my $state = 'start';
while (1) {
    given ($state) {
        when ('start') {
            initialize();
            $state = 'process';
        }
        when ('process') {
            if (process_item()) {
                $state = 'finish';
            } else {
                $state = 'error';
            }
        }
        when ('error') {
            handle_error();
            $state = 'start';
        }
        when ('finish') {
            finalize();
            last;
        }
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
