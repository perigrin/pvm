---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - iterator
    - closure
    - state
---

# Iterator Pattern

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

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
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

## Typed Perl Output

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

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
