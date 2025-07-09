---
category: untyped-perl
subcategory: variables
tags:
    - scalar
    - declarations
    - my
    - variables
---

# Scalar Declarations

Basic scalar variable declarations

```perl
my $name = "example";
my $count = 42;
my $flag = 1;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $name = "example";
my $count = 42;
my $flag = 1;
```

## Typed Perl Output

```perl
my $name = "example";
my $count = 42;
my $flag = 1;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
