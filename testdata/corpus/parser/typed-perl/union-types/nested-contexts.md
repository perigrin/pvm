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
method HashRef[Bool|Str] complex(ArrayRef[Int|Str] $data, CodeRef|Undef $callback) {
    return {};
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 111 characters
  Type Annotations:
    MethodReturnAnnotation: complex :: HashRef[Bool|Str] at 1:8
    MethodParamAnnotation: $data :: ArrayRef[Int|Str] at 1:1
    MethodParamAnnotation: $callback :: CodeRef|Undef at 1:1
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
  Source length: 111 characters
  Type Annotations:
    MethodReturnAnnotation: complex :: HashRef[Bool|Str] at 1:8
    MethodParamAnnotation: $data :: ArrayRef[Int|Str] at 1:1
    MethodParamAnnotation: $callback :: CodeRef|Undef at 1:1
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


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
method complex($data, $callback) {
    return {};
}
```

## Typed Perl Output

```perl
method HashRef[Bool|Str] complex(ArrayRef[Int|Str] $data, CodeRef|Undef $callback) {
    return {};
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
