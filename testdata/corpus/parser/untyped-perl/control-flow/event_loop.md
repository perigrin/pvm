---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - event_loop
    - given
    - when
    - continue
    - validation
---

# Event Loop

Event loop with various event types and validation

```perl
while (my $event = get_next_event()) {
    given ($event->{type}) {
        when ('timer') {
            handle_timer($event);
            continue;
        }
        when ('input') {
            unless (validate_input($event)) {
                log_invalid_input($event);
                next;
            }
            process_input($event);
        }
        when ('shutdown') {
            cleanup_and_exit();
            last;
        }
        default {
            log_unknown_event($event);
        }
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while (my $event = get_next_event()) {
    given ($event->{type}) {
        when ('timer') {
            handle_timer($event);
            continue;
        }
        when ('input') {
            unless (validate_input($event)) {
                log_invalid_input($event);
                next;
            }
            process_input($event);
        }
        when ('shutdown') {
            cleanup_and_exit();
            last;
        }
        default {
            log_unknown_event($event);
        }
    }
}
```

## Typed Perl Output

```perl
while (my $event = get_next_event()) {
    given ($event->{type}) {
        when ('timer') {
            handle_timer($event);
            continue;
        }
        when ('input') {
            unless (validate_input($event)) {
                log_invalid_input($event);
                next;
            }
            process_input($event);
        }
        when ('shutdown') {
            cleanup_and_exit();
            last;
        }
        default {
            log_unknown_event($event);
        }
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
