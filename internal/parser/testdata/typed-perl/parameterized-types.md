---
category: typed-perl
subcategory: parameterized-types
tags:
    - ArrayRef
    - CodeRef
    - Function
    - HashRef
    - Map
    - Tuple
    - basic-parameters
    - class-fields
    - complex-combinations
    - complex-nesting
    - custom-types
    - deep-nesting
    - field-declarations
    - formatting
    - generics
    - method-signatures
    - multiple-parameters
    - nested-parameters
    - package-qualified
    - parameterized-types
    - parameters
    - return-types
    - unions
    - whitespace
---

# Basic Parameterized

Basic parameterized types with single type parameters

```perl
my ArrayRef[Int] @numbers;
my HashRef[Str] %strings;
my CodeRef[Int, Str] $function;
```

## Complex Combinations

Complex combinations of parameterized types with unions and deep nesting

```perl
my Map[Str, ArrayRef[HashRef[Int|Bool]]] $complex;
my Container[ArrayRef[MyType]|HashRef[OtherType]] $flexible;
my Result[Data[UserInfo], Error[ValidationFailure]] $nested_result;
```

## Custom Parameterized

Custom parameterized types and package-qualified generics

```perl
my Container[MyType] $custom_container;
my Package::Generic[Int] $qualified;
my Result[UserData, ErrorCode] $result;
```

## Field Declarations

Parameterized types in field declarations

```perl
field ArrayRef[MyClass] $objects;
field HashRef[ArrayRef[Str]] $nested_data;
field Optional[ArrayRef[Int]] $maybe_numbers;
```

## Method Signatures

Parameterized types in method signatures

```perl
method process(ArrayRef[Str] $input) returns HashRef[Int] {
  return {};
}

method transform(Map[Str, Int] $data) returns ArrayRef[Result[Str, Error]] {
  return [];
}
```

## Multiple Parameters

Parameterized types with multiple type parameters

```perl
my Map[Str, Int] %mapping;
my Tuple[Int, Str, Bool] $triple;
my Function[Int, Str, Bool] $complex_func;
```

## Nested Parameterized

Nested parameterized types with multiple levels

```perl
my ArrayRef[ArrayRef[Int]] @matrix;
my HashRef[ArrayRef[Str]] %grouped_strings;
my ArrayRef[HashRef[Int]] @array_of_hashes;
```

## Whitespace Variations

Parameterized types with various whitespace patterns

```perl
my ArrayRef[ Int ] @spaced;
my HashRef[Str] %compact;
my Map[ Str , Int ] %loose;
my ArrayRef[
  Int
] @multiline;
```
