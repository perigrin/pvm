---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - method-signatures
    - parameterized-returns
    - complex-types
    - parameterized-types
---

# Complex Method Signatures

Complex method signatures with advanced parameter and return types

```perl
method transform(
    ArrayRef[HashRef[Int|Str]] $input,
    CodeRef[Str, Bool] $validator
) returns HashRef[ArrayRef[Int]|Str] {
    return {};
}

method process(
    Map[Str, ArrayRef[Data|Error]] $complex_input,
    Optional[Handler[Request|Response]] $handler
) returns Result[Array[ProcessedData], ProcessingError] {
    return success([]);
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
      token
  method_decl
    block_stmt
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
      token
  method_decl
    block_stmt
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
{ return {} }{ return success([]); }
```

## Typed Perl Output

```perl
method transform(
    ArrayRef[HashRef[Int|Str]] $input,
    CodeRef[Str, Bool] $validator
) returns HashRef[ArrayRef[Int]|Str] {
    return {};
}

method process(
    Map[Str, ArrayRef[Data|Error]] $complex_input,
    Optional[Handler[Request|Response]] $handler
) returns Result[Array[ProcessedData], ProcessingError] {
    return success([]);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
