---
category: typed-perl
subcategory: complex-types
tags:
    - all-features
    - complex-combinations
    - complex-types
    - deep-nesting
    - intersection-types
    - many-alternatives
    - method-calls
    - method-signatures
    - negation-types
    - parameterized-returns
    - parameterized-types
    - parenthesized-unions
    - performance
    - stress-testing
    - type-assertions
    - union-types
---

# All Features Combined
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Complex combination of all type features: unions, intersections, negations, parameterized types, and assertions

```perl
method complex_processing(
    ArrayRef[HashRef[Int|Str]&!Undef] $validated_data,
    (Processor[Request]|Handler[Response])&Configured $handler,
    Optional[Logger[Info|Error]] $logger
) -> Result[Map[Str, Array[ProcessedItem|FailureReason]], ProcessingError&Detailed] {
    my $transformed = $validated_data as ArrayRef[Data&Processed];
    return success($transformed->map(sub { process($_) }));
}
```

## Complex Method Signatures

Complex method signatures with advanced parameter and return types

```perl
method transform(
    ArrayRef[HashRef[Int|Str]] $input,
    CodeRef[Str, Bool] $validator
) -> HashRef[ArrayRef[Int]|Str] {
    return {};
}

method process(
    Map[Str, ArrayRef[Data|Error]] $complex_input,
    Optional[Handler[Request|Response]] $handler
) -> Result[Array[ProcessedData], ProcessingError] {
    return success([]);
}
```

## Complex Type Assertions

Type assertions with complex type expressions

```perl
my $result = $data as ArrayRef[HashRef[Int|Bool]];
my $complex = ($input->process()) as Map[Str, ArrayRef[MyType]];
my $transformed = $obj->convert() as Result[Data[User], Error[String]];
```

## Deep Nesting

Deeply nested parameterized types with complex combinations

```perl
my ArrayRef[HashRef[ArrayRef[Int|Str]]] @deep_nested;
my Map[Str, ArrayRef[Tuple[Int, Bool|Str]]] %complex_map;
my Container[Wrapper[Inner[Data[Value]]]] $deeply_nested;
```

## Intersection Combinations

Intersection types combined with parameterized and union types

```perl
my ArrayRef[Object&Serializable] @serializable_list;
my HashRef[ArrayRef[Int|Str]&Defined] %defined_arrays;
my Container[Data&Validated&Cached] $safe_container;
```

## Negation Combinations

Negation types combined with parameterized and intersection types

```perl
my ArrayRef[!Undef] @non_undef_array;
my HashRef[Str, !Empty] %non_empty_values;
my Optional[!Null&!Undef] $definitely_defined;
```

## Nested Unions In Parameterized

Union types nested within parameterized types

```perl
my ArrayRef[Int|Str|Bool] @complex_array;
my HashRef[ArrayRef[Int]|HashRef[Str]] %nested_complex;
my Map[Str, Int|Undef] %optional_values;
```

## Parameterized Unions

Parameterized types within union expressions

```perl
my (ArrayRef[Int]|HashRef[Str]) $param_union;
my (Container[MyType]|Wrapper[OtherType]) $flexible;
my (Result[Data, Error]|Maybe[Value]) $outcome;
```

## Stress Testing

Stress testing with very deep nesting and many union alternatives

```perl
my Map[Str, ArrayRef[HashRef[Tuple[Int, Result[Data[UserInfo[Profile[Settings[Theme]]]], Error[ValidationFailure[FieldError[TypeError]]]]]]] %extremely_nested;
my (A|B|C|D|E|F|G|H|I|J|K|L|M|N|O|P|Q|R|S|T) $many_union_alternatives;
my Container[Wrapper[Inner[Core[Data[Value[Item[Element[Node[Leaf]]]]]]]]] $very_deep;
```
