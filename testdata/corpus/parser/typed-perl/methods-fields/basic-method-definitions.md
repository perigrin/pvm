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
method calculate(Int $a, Int $b) returns Int {
    return $a + $b;
}

method greet(Str $name) returns Str {
    return "Hello, $name!";
}

method is_valid(Bool $flag) returns Bool {
    return $flag;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 201 characters
  Type Annotations:
    MethodReturnAnnotation: calculate :: Int at 1:42
    MethodReturnAnnotation: greet :: Str at 5:33
    MethodReturnAnnotation: is_valid :: Bool at 9:37
    MethodParamAnnotation: $a :: Int at 1:1
    MethodParamAnnotation: $b :: Int at 1:1
    MethodParamAnnotation: $name :: Str at 5:1
    MethodParamAnnotation: $flag :: Bool at 9:1
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
  Source length: 201 characters
  Type Annotations:
    MethodReturnAnnotation: calculate :: Int at 1:42
    MethodReturnAnnotation: greet :: Str at 5:33
    MethodReturnAnnotation: is_valid :: Bool at 9:37
    MethodParamAnnotation: $a :: Int at 1:1
    MethodParamAnnotation: $b :: Int at 1:1
    MethodParamAnnotation: $name :: Str at 5:1
    MethodParamAnnotation: $flag :: Bool at 9:1
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
{ return $a + $b; }{ return "Hello, $name!"; }{ return $flag; }
```

## Typed Perl Output

```perl
method calculate(Int $a, Int $b) returns Int {
    return $a + $b;
}

method greet(Str $name) returns Str {
    return "Hello, $name!";
}

method is_valid(Bool $flag) returns Bool {
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
