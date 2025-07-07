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
method HashRef[Int] process(ArrayRef[Str] $input) {
  return {};
}

method ArrayRef[Result[Str, Error]] transform(Map[Str, Int] $data) {
  return [];
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 147 characters
  Type Annotations:
    MethodReturnAnnotation: process :: HashRef[Int] at 1:8
    MethodReturnAnnotation: transform :: ArrayRef[Result[Str, Error]] at 1:8
    MethodParamAnnotation: ArrayRef[Result[Str :: ArrayRef[Str] at 1:1
    MethodParamAnnotation: $data :: Int] at 1:1
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
  Source length: 147 characters
  Type Annotations:
    MethodReturnAnnotation: process :: HashRef[Int] at 1:8
    MethodReturnAnnotation: transform :: ArrayRef[Result[Str, Error]] at 1:8
    MethodParamAnnotation: ArrayRef[Result[Str :: ArrayRef[Str] at 1:1
    MethodParamAnnotation: $data :: Int] at 1:1
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


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
method process($input) {
  return {};
}

method transform($data) {
  return [];
}
```

## Typed Perl Output

```perl
method HashRef[Int] process(ArrayRef[Str] $input) {
  return {};
}

method ArrayRef[Result[Str, Error]] transform(Map[Str, Int] $data) {
  return [];
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
