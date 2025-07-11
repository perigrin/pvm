---
category: integration
subcategory: basic-programs
tags:
    - typed-variables
    - functions
    - basic-types
type_check: false
should_error: false
---

# Basic Typed Program

A simple program using basic typed Perl features that should parse successfully

```perl
use v5.36;
use strict;
use warnings;

# Basic typed variables
my Int $count = 42;
my Str $name = "example";
my Bool $active = 1;

# Simple function with types
sub Str process_user(Int $id, Str $name) {
    return "User $id: $name";
}

# Simple usage
my Str $result = process_user($count, $name);
print $result . "\n";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
use strict;
use warnings;

# Basic typed variables
my $count = 42;
my $name = "example";
my $active = 1;

# Simple function with types
sub process_user($id, $name) {
    return "User $id: $name";
}

# Simple usage
my $result = process_user($count, $name);
print $result . "\n";
```

## Typed Perl Output

```perl
use v5.36;
use strict;
use warnings;

# Basic typed variables
my Int $count = 42;
my Str $name = "example";
my Bool $active = 1;

# Simple function with types
sub Str process_user(Int$id, Str$name) {
    return "User $id: $name";
}

# Simple usage
my Str $result = process_user($count, $name);
print $result . "\n";
```

## Inferred Perl Output

```perl
# Type inference not implemented - placeholder
use v5.36;
use strict;
use warnings;

# Basic typed variables
my Int $count = 42;
my Str $name = "example";
my Bool $active = 1;

# Simple function with types
sub Str process_user(Int$id, Str$name) {
    return "User $id: $name";
}

# Simple usage
my Str $result = process_user($count, $name);
print $result . "\n";
```

# Expected Type Errors

```
(none)
```
