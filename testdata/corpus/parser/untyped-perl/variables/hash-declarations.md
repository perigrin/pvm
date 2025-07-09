---
category: untyped-perl
subcategory: variables
tags:
    - hash
    - declarations
    - my
    - variables
---

# Hash Declarations

Basic hash variable declarations

```perl
my %person = (name => "John", age => 30);
my %config = ("debug" => 1, "verbose" => 0);
my %empty = ();
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my %person = (name => "John", age => 30);
my %config = ("debug" => 1, "verbose" => 0);
my %empty = ();
```

## Typed Perl Output

```perl
my %person = (name => "John", age => 30);
my %config = ("debug" => 1, "verbose" => 0);
my %empty = ();
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
