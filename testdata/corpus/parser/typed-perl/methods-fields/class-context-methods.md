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

    method Num add(Num $a, Num $b) {
        return $a + $b;
    }

    method Num get_precision() {
        return $precision;
    }

    method Void set_precision(Num $new_precision) {
        $precision = $new_precision;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 285 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Num at 4:12
    MethodReturnAnnotation: get_precision :: Num at 8:12
    MethodReturnAnnotation: set_precision :: Void at 12:12
    VarAnnotation: Calculator :: class at 1:1
    VarAnnotation: $precision :: Num at 2:5
    MethodReturnAnnotation: add :: Num at 4:12
    MethodReturnAnnotation: get_precision :: Num at 8:12
    MethodReturnAnnotation: set_precision :: Void at 12:12
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
  Source length: 285 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Num at 4:12
    MethodReturnAnnotation: get_precision :: Num at 8:12
    MethodReturnAnnotation: set_precision :: Void at 12:12
    VarAnnotation: Calculator :: class at 1:1
    VarAnnotation: $precision :: Num at 2:5
    MethodReturnAnnotation: add :: Num at 4:12
    MethodReturnAnnotation: get_precision :: Num at 8:12
    MethodReturnAnnotation: set_precision :: Void at 12:12
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


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Calculator {
    field $precision = 0.001;

    method add($a, $b) {
        return $a + $b;
    }

    method get_precision() {
        return $precision;
    }

    method set_precision($new_precision) {
        $precision = $new_precision;
    }
}
```

## Typed Perl Output

```perl
class Calculator {
    field Num $precision = 0.001;

    method Num add(Num $a, Num $b) {
        return $a + $b;
    }

    method Num get_precision() {
        return $precision;
    }

    method Void set_precision(Num $new_precision) {
        $precision = $new_precision;
    }
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
