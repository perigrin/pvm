---
category: typed-perl
subcategory: methods-fields
tags:
    - class-context
    - class-definitions
    - typed-methods
    - typed-fields
type_check: true
---

# Class Context Methods

Methods and fields within class context with type annotations

```perl
class Calculator {
    field Num $precision = 0.001;

    method add(Num $a, Num $b) returns Num {
        return $a + $b;
    }

    method get_precision() returns Num {
        return $precision;
    }

    method set_precision(Num $new_precision) returns Void {
        $precision = $new_precision;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 309 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Num at 4:40
    MethodReturnAnnotation: get_precision :: Num at 8:36
    MethodReturnAnnotation: set_precision :: Void at 12:54
    VarAnnotation: Calculator :: class at 1:1
    VarAnnotation: $precision :: Num at 2:5
    MethodReturnAnnotation: add :: Num at 4:40
    MethodReturnAnnotation: get_precision :: Num at 8:36
    MethodReturnAnnotation: set_precision :: Void at 12:54
    FieldAnnotation: $precision :: Num at 2:1
    MethodParamAnnotation: $a :: Num at 4:1
    MethodParamAnnotation: $b :: Num at 4:1
    MethodParamAnnotation: $new_precision :: Num at 12:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
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
  Source length: 309 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Num at 4:40
    MethodReturnAnnotation: get_precision :: Num at 8:36
    MethodReturnAnnotation: set_precision :: Void at 12:54
    VarAnnotation: Calculator :: class at 1:1
    VarAnnotation: $precision :: Num at 2:5
    MethodReturnAnnotation: add :: Num at 4:40
    MethodReturnAnnotation: get_precision :: Num at 8:36
    MethodReturnAnnotation: set_precision :: Void at 12:54
    FieldAnnotation: $precision :: Num at 2:1
    MethodParamAnnotation: $a :: Num at 4:1
    MethodParamAnnotation: $b :: Num at 4:1
    MethodParamAnnotation: $new_precision :: Num at 12:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
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
