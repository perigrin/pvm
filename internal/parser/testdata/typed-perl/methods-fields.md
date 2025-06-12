---
category: typed-perl
subcategory: methods-fields
tags:
    - access-control
    - backward-compatibility
    - class-context
    - class-definitions
    - complex-fields
    - complex-methods
    - complex-return-types
    - custom-types
    - field-declarations
    - field-initialization
    - field-modifiers
    - field-visibility
    - gradual-typing
    - method-definitions
    - method-signatures
    - mixed-typing
    - optional-parameters
    - parameter-types
    - parameterized-types
    - return-types
    - typed-fields
    - typed-methods
---

# Basic Field Declarations

Basic typed field declarations with and without initializers

```perl
field Int $count = 0;
field Str $name;
field Bool $is_active = 1;
field Num $rate = 3.14;
```

## Basic Method Definitions
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Basic typed method definitions with parameter and return types

```perl
method calculate(Int $a, Int $b) -> Int {
    return $a + $b;
}

method greet(Str $name) -> Str {
    return "Hello, $name!";
}

method is_valid(Bool $flag) -> Bool {
    return $flag;
}
```

## Class Context Methods

Methods and fields within class context with type annotations

```perl
class Calculator {
    field Num $precision = 0.001;

    method add(Num $a, Num $b) -> Num {
        return $a + $b;
    }

    method get_precision() -> Num {
        return $precision;
    }

    method set_precision(Num $new_precision) -> Void {
        $precision = $new_precision;
    }
}
```

## Complex Field Types

Field declarations with complex parameterized types and custom types

```perl
field ArrayRef[Int] $numbers = [];
field HashRef[Str] $config = {};
field CodeRef[Int, Str] $formatter;
field ArrayRef[MyType] $items;
field HashRef[ArrayRef[Str]] $grouped_data = {};
```

## Complex Method Signatures
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Complex method signatures with parameterized types, optional parameters, and multiple parameter types

```perl
method process(ArrayRef[Str] $data, Bool $validate = 1) -> ArrayRef[Str] {
    my @result = @{$data};
    return \@result;
}

method transform(HashRef[Int] $input, CodeRef $callback) -> HashRef[Int] {
    my %result;
    for my $key (keys %{$input}) {
        $result{$key} = $callback->($input->{$key});
    }
    return \%result;
}

method complex_method(
    ArrayRef[HashRef[Int]] $data,
    Optional[CodeRef] $processor,
    Slurpy[Str] @extra_args
) -> Bool {
    return 1;
}
```

## Field Access Modifiers
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Field declarations with access modifiers and visibility keywords

```perl
# Different field visibility patterns
field Int $public_field = 1;
field private Str $private_field = "secret";
field protected Bool $protected_field = 0;
field readonly ArrayRef[Int] $readonly_field = [1, 2, 3];
field static HashRef[Str] $class_field = {};
```

## Method Return Types
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Methods with various return type annotations including complex types

```perl
method get_number() -> Int {
    return 42;
}

method get_array() -> ArrayRef[Str] {
    return ["a", "b", "c"];
}

method get_hash() -> HashRef[Int] {
    return { count => 5, total => 100 };
}

method get_nothing() -> Void {
    # Side effects only
    print "Done\n";
}

method get_optional() -> Optional[Str] {
    return undef;
}
```

## Mixed Typed Untyped
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Mixed typed and untyped methods and fields in the same context

```perl
# Mixed typed and untyped methods and fields
field Int $typed_field = 42;
field $untyped_field = "hello";

method typed_method(Str $input) -> Str {
    return uc($input);
}

sub untyped_sub {
    my ($param) = @_;
    return $param * 2;
}

method partially_typed($untyped, Int $typed) -> Str {
    return "$untyped: $typed";
}
```
