---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - while
    - assignment
    - file_handle
---

# While With Assignment

While loop with assignment in condition

```perl
while (my $line = <$fh>) {
    chomp $line;
    process($line);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while (my $line = <$fh>) {
    chomp $line;
    process($line);
}
```

## Typed Perl Output

```perl
while (my $line = <$fh>) {
    chomp $line;
    process($line);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
