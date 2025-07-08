---
category: typed-perl
subcategory: methods-fields
tags:
    - method-definitions
    - method-signatures
    - parameter-types
    - return-types
type_check: true
---

# Basic Method Definitions

Basic typed method definitions with parameter and return types

```perl
method Int calculate(Int $a, Int $b) {
    return $a + $b;
}

method Str greet(Str $name) {
    return "Hello, $name!";
}

method Bool is_valid(Bool $flag) {
    return $flag;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 177 characters
  Type Annotations:
    MethodReturnAnnotation: calculate :: Int at 1:8
    MethodParamAnnotation: $a :: Int at 1:22
    MethodParamAnnotation: $b :: Int at 1:30
    MethodReturnAnnotation: greet :: Str at 5:8
    MethodParamAnnotation: $name :: Str at 5:18
    MethodReturnAnnotation: is_valid :: Bool at 9:8
    MethodParamAnnotation: $flag :: Bool at 9:22
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
  Source length: 177 characters
  Type Annotations:
    MethodReturnAnnotation: calculate :: Int at 1:8
    MethodParamAnnotation: $a :: Int at 1:22
    MethodParamAnnotation: $b :: Int at 1:30
    MethodReturnAnnotation: greet :: Str at 5:8
    MethodParamAnnotation: $name :: Str at 5:18
    MethodReturnAnnotation: is_valid :: Bool at 9:8
    MethodParamAnnotation: $flag :: Bool at 9:22
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
method calculate($a, $b) {
    return $a + $b;
}

method greet($name) {
    return "Hello, $name!";
}

method is_valid($flag) {
    return $flag;
}
```

## Typed Perl Output

```perl
method Int calculate(Int $a, Int $b) {
    return $a + $b;
}

method Str greet(Str $name) {
    return "Hello, $name!";
}

method Bool is_valid(Bool $flag) {
    return $flag;
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
