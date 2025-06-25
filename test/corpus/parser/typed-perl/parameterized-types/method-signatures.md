---
category: typed-perl
subcategory: parameterized-types
tags:
    - method-signatures
    - return-types
    - parameters
    - parameterized-types
type_check: true
---

# Method Signatures

Parameterized types in method signatures

```perl
method process(ArrayRef[Str] $input) returns HashRef[Int] {
  return {};
}

method transform(Map[Str, Int] $data) returns ArrayRef[Result[Str, Error]] {
  return [];
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 167 characters
  Type Annotations:
    MethodReturnAnnotation: process :: HashRef[Int] at 1:46
    MethodReturnAnnotation: transform :: ArrayRef[Result[Str, Error]] at 5:47
    MethodParamAnnotation: $input :: ArrayRef[Str] at 1:1
    MethodParamAnnotation: $data :: Int] at 5:1
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
  Source length: 167 characters
  Type Annotations:
    MethodReturnAnnotation: process :: HashRef[Int] at 1:46
    MethodReturnAnnotation: transform :: ArrayRef[Result[Str, Error]] at 5:47
    MethodParamAnnotation: $input :: ArrayRef[Str] at 1:1
    MethodParamAnnotation: $data :: Int] at 5:1
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
