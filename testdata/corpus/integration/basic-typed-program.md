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
