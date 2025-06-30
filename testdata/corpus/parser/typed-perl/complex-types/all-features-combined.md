---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - all-features
    - complex-combinations
    - intersection-types
    - negation-types
    - parameterized-types
    - union-types
    - type-assertions
---

# All Features Combined

Complex combination of all type features: unions, intersections, negations, parameterized types, and assertions

```perl
method complex_processing(
    ArrayRef[HashRef[Int|Str]&!Undef] $validated_data,
    (Processor[Request]|Handler[Response])&Configured $handler,
    Optional[Logger[Info|Error]] $logger
) returns Result[Map[Str, Array[ProcessedItem|FailureReason]], ProcessingError&Detailed] {
    my $transformed = $validated_data as ArrayRef[Data&Processed];
    return success($transformed->map(sub { process($_) }));
}
```

## Expected AST

### Before Type Inference

```
source_file
  method_decl
    block_stmt
      token
      expression_stmt
        literal
      token
      expression_stmt
        literal
      token
      token
```

### After Type Inference

```
source_file
  method_decl
    block_stmt
      token
      expression_stmt
        literal
      token
      expression_stmt
        literal
      token
      token
```

#
# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
{ my $transformed = $validated_data as ArrayRef[Data&Processed]; return success($transformed->map(sub { process($_) })); }
```

## Typed Perl Output

```perl
method complex_processing(
    ArrayRef[HashRef[Int|Str]&!Undef] $validated_data,
    (Processor[Request]|Handler[Response])&Configured $handler,
    Optional[Logger[Info|Error]] $logger
) returns Result[Map[Str, Array[ProcessedItem|FailureReason]], ProcessingError&Detailed] {
    my $transformed = $validated_data as ArrayRef[Data&Processed];
    return success($transformed->map(sub { process($_) }));
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
