---
category: typed-perl
subcategory: methods-fields
tags:
    - field-declarations
    - field-initialization
    - typed-fields
type_check: true
---

# Basic Field Declarations

Basic typed field declarations with and without initializers

```perl
field Int $count = 0;
field Str $name;
field Bool $is_active = 1;
field Num $rate = 3.14;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 89 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
    VarAnnotation: $is_active :: Bool at 3:1
    VarAnnotation: $rate :: Num at 4:1
    FieldAnnotation: $count :: Int at 1:1
    FieldAnnotation: $name :: Str at 2:1
    FieldAnnotation: $is_active :: Bool at 3:1
    FieldAnnotation: $rate :: Num at 4:1
  Root: source_file
  Tree Structure:
  source_file
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
        token
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
        token
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 89 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
    VarAnnotation: $is_active :: Bool at 3:1
    VarAnnotation: $rate :: Num at 4:1
    FieldAnnotation: $count :: Int at 1:1
    FieldAnnotation: $name :: Str at 2:1
    FieldAnnotation: $is_active :: Bool at 3:1
    FieldAnnotation: $rate :: Num at 4:1
  Root: source_file
  Tree Structure:
  source_file
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
        token
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
        token
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
field $count = 0;
field $name;
field $is_active = 1;
field $rate = 3.14;
```

## Typed Perl Output

```perl
field Int $count = 0;
field Str $name;
field Bool $is_active = 1;
field Num $rate = 3.14;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
