---
category: typed-perl
subcategory: union-types
tags:
    - nested-contexts
    - complex-parameters
    - method-signatures
    - union-types
type_check: true
---

# Nested Contexts

Union types in nested method signature contexts with complex parameters

```perl
method complex(
    ArrayRef[Int|Str] $data,
    CodeRef|Undef $callback
) returns HashRef[Bool|Str] {
    return {};
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 119 characters
  Type Annotations:
    MethodReturnAnnotation: complex :: HashRef[Bool|Str] at 4:11
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 119 characters
  Type Annotations:
    MethodReturnAnnotation: complex :: HashRef[Bool|Str] at 4:11
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

# Expected Type Errors

```
(none)
```
