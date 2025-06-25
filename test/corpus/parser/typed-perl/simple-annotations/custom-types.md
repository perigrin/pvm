---
category: typed-perl
subcategory: simple-annotations
tags:
    - custom-types
    - package-qualified
    - typed-variables
type_check: true
---

# Custom Types

Variable declarations with custom and package-qualified types

```perl
my MyType $custom;
my Package::CustomType $qualified;
my UserClass $user = UserClass->new();
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 92 characters
  Type Annotations:
    VarAnnotation: $custom :: MyType at 1:1
    VarAnnotation: $qualified :: Package::CustomType at 2:1
    VarAnnotation: $user :: UserClass at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          expression_stmt
            literal
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          expression_stmt
            literal
        scalar
          token
          token
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
  Source length: 92 characters
  Type Annotations:
    VarAnnotation: $custom :: MyType at 1:1
    VarAnnotation: $qualified :: Package::CustomType at 2:1
    VarAnnotation: $user :: UserClass at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          expression_stmt
            literal
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          expression_stmt
            literal
        scalar
          token
          token
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
