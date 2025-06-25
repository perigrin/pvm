---
category: typed-perl
subcategory: simple-annotations
tags:
    - mixed-code
    - typed-variables
    - untyped-variables
    - backward-compatibility
type_check: true
---

# Mixed Typed Untyped

Mixed typed and untyped variable declarations in the same code

```perl
my Int $typed_var = 42;
my $untyped_var = "hello";
my Str $another_typed = "world";
my @untyped_array = (1, 2, 3);
my ArrayRef[Int] @typed_array = (4, 5, 6);
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 157 characters
  Type Annotations:
    VarAnnotation: $typed_var :: Int at 1:1
    VarAnnotation: $another_typed :: Str at 3:1
    VarAnnotation: @typed_array :: ArrayRef[Int] at 5:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 157 characters
  Type Annotations:
    VarAnnotation: $typed_var :: Int at 1:1
    VarAnnotation: $another_typed :: Str at 3:1
    VarAnnotation: @typed_array :: ArrayRef[Int] at 5:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
}
```

# Expected Type Errors

```
(none)
```
