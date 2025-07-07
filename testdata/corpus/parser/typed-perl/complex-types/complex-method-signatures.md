---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - method-signatures
    - parameterized-return-types
    - complex-types
    - parameterized-types
---

# Complex Method Signatures

Complex method signatures with advanced parameter and return types

```perl
method HashRef[ArrayRef[Int]|Str] transform(ArrayRef[HashRef[Int|Str]] $input, CodeRef[Str, Bool] $validator) {
    return {};
}

method Result[Array[ProcessedData], ProcessingError] process(Map[Str, ArrayRef[Data|Error]] $complex_input, Optional[Handler[Request|Response]] $handler) {
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
use v5.36;
method transform($input, $validator) {
    return {};
}

method process($complex_input, $handler) {
    return success([]);
}
```

## Typed Perl Output

```perl
method HashRef[ArrayRef[Int]|Str] transform(ArrayRef[HashRef[Int|Str]] $input, CodeRef[Str, Bool] $validator) {
    return {};
}

method Result[Array[ProcessedData], ProcessingError] process(Map[Str, ArrayRef[Data|Error]] $complex_input, Optional[Handler[Request|Response]] $handler) {
    return success([]);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
