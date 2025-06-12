---
category: typed-perl
subcategory: simple-annotations
tags:
    - arrays
    - backward-compatibility
    - built-in-types
    - complex-assignments
    - custom-types
    - expressions
    - formatting
    - hashes
    - local
    - maybe-types
    - mixed-code
    - optional-types
    - our
    - package-qualified
    - parameterized-types
    - scoping
    - state
    - typed-variables
    - undefined
    - untyped-variables
    - variable-declarations
    - whitespace
---

# Basic Typed Variables
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Basic typed variable declarations with built-in types

```perl
my Int $count = 42;
my Str $name = "example";
my Bool $flag = 1;
my Num $pi = 3.14159;
```

## Complex Assignments
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Type annotations with complex assignment expressions

```perl
my Int $calculated = $base + $increment;
my Str $interpolated = "Value: $count";
my Bool $comparison = $a > $b;
my Num $result = $x * $y + $z;
```

## Custom Types

Variable declarations with custom and package-qualified types

```perl
my MyType $custom;
my Package::CustomType $qualified;
my UserClass $user = UserClass->new();
```

## Mixed Typed Untyped
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Mixed typed and untyped variable declarations in the same code

```perl
my Int $typed_var = 42;
my $untyped_var = "hello";
my Str $another_typed = "world";
my @untyped_array = (1, 2, 3);
my ArrayRef[Int] @typed_array = (4, 5, 6);
```

## Optional Annotations

Type annotations including optional/maybe types and undefined values

```perl
my Int $with_type;
my Str $also_typed = undef;
my Optional[Int] $maybe_int;
my Maybe[Str] $maybe_str;
```

## Scoping Keywords
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Type annotations with different scoping keywords (our, state, local)

```perl
our Int $global_counter = 0;
state Str $persistent_cache = "";
local Num $localized = 1.0;
```

## Typed Arrays Hashes
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Typed array and hash declarations with basic parameterized types

```perl
my ArrayRef[Int] @numbers = (1, 2, 3);
my HashRef[Str] %config = (key => 'value');
my ArrayRef[Str] @strings = ("a", "b", "c");
```

## Whitespace Variations
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Type annotations with various whitespace patterns

```perl
my  Int  $spaced = 42;
my	Str	$tabbed="test";
my Int$compact=100;
my   ArrayRef[Str]   @loose_array   =   ();
```
