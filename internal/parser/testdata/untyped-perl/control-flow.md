---
category: untyped-perl
subcategory: control-flow
tags:
    - array
    - assignment
    - bare_block
    - block
    - break
    - c_style
    - chain
    - cleanup
    - closure
    - complex
    - complex_expression
    - conditional
    - continue
    - control_flow
    - coroutine
    - default
    - default_variable
    - depth
    - dispatch
    - do_block
    - do_until
    - do_while
    - else
    - elsif
    - error_handling
    - eval
    - event_loop
    - exception
    - explicit
    - fallthrough
    - file_handle
    - file_test
    - for
    - foreach
    - function_call
    - given
    - glob
    - hash
    - if
    - increment
    - infinite
    - initialization
    - iterator
    - keys
    - labeled
    - last
    - list
    - local
    - logical
    - loop
    - loop_control
    - mixed
    - multiple
    - nested
    - next
    - no_default
    - optional
    - options
    - outer
    - parallel
    - pattern
    - pipeline
    - postfix
    - progress
    - qw
    - range
    - recursive
    - redo
    - regex
    - restart
    - retry
    - return
    - scope
    - search
    - simulation
    - single_line
    - smartmatch
    - state
    - state_machine
    - statement_modifier
    - switch
    - tree
    - try_catch
    - underscore
    - unless
    - until
    - validation
    - value
    - variable_declaration
    - when
    - while
    - workers
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

## Basic If Statement

Basic if statement with block

```perl
if ($condition) {
    do_something();
}
```

## Basic Until Loop

Basic until loop

```perl
until ($done) {
    continue_processing();
}
```

## Basic While Loop

Basic while loop

```perl
while ($condition) {
    process();
}
```

## Complex Condition

Conditional with complex boolean expression

```perl
if (($a > 0) && ($b < 100) || ($c == $d)) {
    complex_action();
}
```

## Complex Loop Control

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

## Conditional With Assignment

Conditional with variable assignment in condition

```perl
if (my $result = function_call()) {
    process($result);
}
```

## Conditional With Return

Conditional with return statements

```perl
if ($error) {
    return $error_value;
}
return $success_value;
```

## Continue Block

While loop with continue block

```perl
while ($condition) {
    process();
} continue {
    cleanup();
    update_condition();
}
```

## Coroutine Simulation

Coroutine simulation using dispatch table and redo

```perl
sub coroutine {
    my $state = shift;

    DISPATCH: {
        $state eq 'init' and do {
            initialize_data();
            $state = 'process';
            redo DISPATCH;
        };

        $state eq 'process' and do {
            return process_next() || 'done';
        };

        $state eq 'done' and do {
            cleanup();
            return undef;
        };
    }
}
```

## Event Loop
<!-- expected_error: ParseError - given/when not implemented -->

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

## Exception Handling Eval

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

## For Loop C Style

C-style for loop with initialization, condition, and increment

```perl
for (my $i = 0; $i < $max; $i++) {
    handle($i);
}
```

## For Loop List

For loop over literal list

```perl
for my $item (qw(apple banana cherry)) {
    print "$item\n";
}
```

## For Loop Range

For loop with range operator

```perl
for my $i (0..$max) {
    handle($i);
}
```

## Foreach Continue

Foreach loop with continue block

```perl
foreach my $item (@items) {
    process($item);
} continue {
    log_progress();
}
```

## Foreach Loop Array

Foreach loop over array

```perl
foreach my $item (@list) {
    process($item);
}
```

## Foreach Loop Hash

Foreach loop over hash keys

```perl
foreach my $key (keys %hash) {
    process($key, $hash{$key});
}
```

## Foreach No Variable

Foreach loop using default variable $_

```perl
foreach (@items) {
    process($_);
}
```

## Given No Default
<!-- expected_error: ParseError - given/when not implemented -->

Given-when without default clause

```perl
given ($option) {
    when ('verbose') { $verbose = 1; }
    when ('quiet') { $quiet = 1; }
    when ('debug') { $debug = 1; }
}
```

## Given When Arrays
<!-- expected_error: ParseError - given/when not implemented -->

Given-when with array reference matching

```perl
given ($day) {
    when ([qw(sat sun)]) { print "weekend"; }
    when ([qw(mon tue wed thu fri)]) { print "weekday"; }
}
```

## Given When Basic
<!-- expected_error: ParseError - given/when not implemented -->

Basic given-when switch statement

```perl
given ($value) {
    when (1) { print "one"; }
    when (2) { print "two"; }
    default { print "other"; }
}
```

## Given When Complex Condition
<!-- expected_error: ParseError - given/when not implemented -->

Given-when with complex conditions

```perl
given ($record) {
    when ($_->{type} eq 'user' && $_->{active}) {
        process_active_user($_);
    }
    when ($_->{type} eq 'admin') {
        process_admin($_);
    }
    default { process_unknown($_); }
}
```

## Given When Nested
<!-- expected_error: ParseError - given/when not implemented -->

Nested given-when statements

```perl
given ($type) {
    when ('user') {
        given ($action) {
            when ('create') { create_user(); }
            when ('delete') { delete_user(); }
            default { invalid_action(); }
        }
    }
    when ('admin') { admin_action(); }
}
```

## Given When Ranges
<!-- expected_error: ParseError - given/when not implemented -->

Given-when with range matching

```perl
given ($score) {
    when (90..100) { print "A"; }
    when (80..89)  { print "B"; }
    when (70..79)  { print "C"; }
    default        { print "F"; }
}
```

## Given When Regex
<!-- expected_error: ParseError - given/when not implemented -->

Given-when with regex matching

```perl
given ($input) {
    when (/^\d+$/) { print "number"; }
    when (/^[a-zA-Z]+$/) { print "letters"; }
    when (/^\s*$/) { print "empty"; }
    default { print "mixed"; }
}
```

## If Elsif Else Chain

Complete if-elsif-else conditional chain

```perl
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
} else {
    do_default();
}
```

## Infinite Loop

Infinite loop with break condition

```perl
while (1) {
    handle_request();
    last if $shutdown;
}
```

## Iterator Pattern

Iterator pattern with closure and state variables

```perl
my $iterator = sub {
    state @items = (1..10);
    state $index = 0;
    return if $index >= @items;
    return $items[$index++];
};

while (defined(my $item = $iterator->())) {
    process($item);
}
```

## Labeled Last

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

## Labeled Next

Labeled next to skip outer loop iteration

```perl
OUTER: for my $i (1..10) {
    for my $j (1..10) {
        next OUTER if ($i * $j > 50);
        print "$i x $j = ", $i * $j, "\n";
    }
}
```

## Labeled Redo

Labeled redo to restart specific loop

```perl
RETRY: while ($attempts < $max_attempts) {
    my $result = try_operation();
    if ($result eq 'retry') {
        $attempts++;
        redo RETRY;
    }
    return $result;
}
```

## Last Statement

Last statement to break out of loop

```perl
foreach my $item (@list) {
    process($item);
    last if stop_condition($item);
}
```

## Last With Value

Last statement with return value

```perl
my $result = do {
    for my $item (@items) {
        if ($item->{target}) {
            last $item->{value};
        }
    }
    'default';
};
```

## Loop With Complex Iterator

Loop with function call as iterator

```perl
foreach my $file (glob("*.txt")) {
    process_file($file);
}
```

## Multiple Elsif

Multiple elsif clauses in conditional

```perl
if ($a) {
    first();
} elsif ($b) {
    second();
} elsif ($c) {
    third();
} else {
    default();
}
```

## Nested Conditionals

Nested conditional statements

```perl
if ($outer) {
    if ($inner) {
        nested_action();
    } else {
        inner_else();
    }
} else {
    outer_else();
}
```

## Nested Loops

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

## Nested Mixed Structures

Deeply nested mixed control structures

```perl
if ($enabled) {
    foreach my $item (@items) {
        if ($item->{process}) {
            while (my $data = $item->next()) {
                last if process_data($data);
            }
        }
    }
}
```

## Next Statement

Next statement to skip to next iteration

```perl
foreach my $item (@list) {
    next if skip_condition($item);
    process($item);
}
```

## Next With Postfix

Next with postfix unless condition

```perl
for my $file (@files) {
    next unless -f $file;
    process_file($file);
}
```

## Parallel Processing Sim

Parallel processing simulation with worker management

```perl
my @workers = ();
for my $i (1..$num_workers) {
    push @workers, {
        id => $i,
        status => 'idle',
        task => undef
    };
}

while (@tasks || grep { $_->{status} eq 'busy' } @workers) {
    foreach my $worker (@workers) {
        if ($worker->{status} eq 'idle' && @tasks) {
            $worker->{task} = shift @tasks;
            $worker->{status} = 'busy';
        } elsif ($worker->{status} eq 'busy') {
            if (task_complete($worker->{task})) {
                finish_task($worker->{task});
                $worker->{status} = 'idle';
                $worker->{task} = undef;
            }
        }
    }
    sleep(0.1); # simulate time passage
}
```

## Pipeline Processing

Pipeline processing with error handling and optional stages

```perl
foreach my $input (@inputs) {
    my $result = $input;

    STAGE: foreach my $stage (@pipeline) {
        eval {
            $result = $stage->process($result);
        };
        if ($@) {
            log_error("Stage failed: $@");
            next STAGE if $stage->{optional};
            last STAGE;
        }
    }

    push @outputs, $result if defined $result;
}
```

## Postfix If

Postfix if conditional

```perl
do_something() if $condition;
```

## Postfix Unless

Postfix unless conditional

```perl
execute() unless $skip;
```

## Postfix Until

Do-until loop (postfix until)

```perl
do {
    attempt();
} until ($success);
```

## Postfix While

Do-while loop (postfix while)

```perl
do {
    process();
} while ($continue);
```

## Recursive Descent

Recursive tree processing with depth control

```perl
sub process_tree {
    my ($node, $depth) = @_;
    return if $depth > $max_depth;

    if ($node->{type} eq 'leaf') {
        process_leaf($node);
    } else {
        foreach my $child (@{$node->{children}}) {
            process_tree($child, $depth + 1);
        }
    }
}
```

## Redo Statement

Redo statement to restart current iteration

```perl
while ($condition) {
    my $input = get_input();
    redo if invalid($input);
    process($input);
}
```

## Single Line If Block

Single line if statement with braces

```perl
if ($condition) { single_statement(); }
```

## Smartmatch Operator

Smartmatch operator usage

```perl
if ($value ~~ @valid_values) {
    process($value);
} elsif ($value ~~ /pattern/) {
    handle_pattern($value);
}
```

## State Machine Loop
<!-- expected_error: ParseError - given/when not implemented -->

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

## Try Catch Finally

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

## Unless Else

Unless statement with else clause

```perl
unless ($condition) {
    when_false();
} else {
    when_true();
}
```

## Unless Statement

Unless conditional statement

```perl
unless ($negative_condition) {
    execute();
}
```

## When Break
<!-- expected_error: ParseError - given/when not implemented -->

When clause with break statement

```perl
given ($status) {
    when ('active') {
        process_active();
        break;
    }
    when ('inactive') { process_inactive(); }
    default { handle_unknown(); }
}
```

## When Continue
<!-- expected_error: ParseError - given/when not implemented -->

When clause with continue statement

```perl
given ($value) {
    when ($_ > 0) {
        print "positive";
        continue;
    }
    when ($_ % 2 == 0) { print "even"; }
    when ($_ % 2 == 1) { print "odd"; }
}
```

## While With Assignment

While loop with assignment in condition

```perl
while (my $line = <$fh>) {
    chomp $line;
    process($line);
}
```
