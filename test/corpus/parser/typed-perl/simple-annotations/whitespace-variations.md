---
category: typed-perl
subcategory: simple-annotations
tags:
    - whitespace
    - formatting
    - typed-variables
type_check: true
---

# Whitespace Variations

Type annotations with various whitespace patterns

```perl
my  Int  $spaced = 42;
my	Str	$tabbed="test";
my Int$compact=100;
my   ArrayRef[Str]   @loose_array   =   ();
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 109 characters
  Type Annotations:
    VarAnnotation: $spaced :: Int at 1:1
    VarAnnotation: $tabbed :: Str at 2:1
    VarAnnotation: $compact :: Int at 3:1
    VarAnnotation: @loose_array :: ArrayRef[Str] at 4:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          scalar
            token
            token
        token
        interpolated_string_literal
          expression_stmt
            literal
          expression_stmt
            literal
          expression_stmt
            literal
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
  Source length: 109 characters
  Type Annotations:
    VarAnnotation: $spaced :: Int at 1:1
    VarAnnotation: $tabbed :: Str at 2:1
    VarAnnotation: $compact :: Int at 3:1
    VarAnnotation: @loose_array :: ArrayRef[Str] at 4:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          scalar
            token
            token
        token
        interpolated_string_literal
          expression_stmt
            literal
          expression_stmt
            literal
          expression_stmt
            literal
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
