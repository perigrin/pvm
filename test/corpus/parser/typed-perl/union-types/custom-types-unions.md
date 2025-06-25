---
category: typed-perl
subcategory: union-types
tags:
    - custom-types
    - package-qualified
    - union-types
type_check: true
---

# Custom Types Unions

Union types with custom and package-qualified type names

```perl
my MyClass|OtherClass $object;
my Package::Type1|Package::Type2 $qualified;
my UserType|SystemType|DefaultType $flexible;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 121 characters
  Type Annotations:
    VarAnnotation: $object :: MyClass|OtherClass at 1:1
    VarAnnotation: $qualified :: Package::Type1|Package::Type2 at 2:1
    VarAnnotation: $flexible :: UserType|SystemType|DefaultType at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          union_type
            type_expression
              expression_stmt
                literal
            expression_stmt
              literal
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
          union_type
            type_expression
              expression_stmt
                literal
            expression_stmt
              literal
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
          union_type
            type_expression
              union_type
                type_expression
                  expression_stmt
                    literal
                expression_stmt
                  literal
                type_expression
                  expression_stmt
                    literal
            expression_stmt
              literal
            type_expression
              expression_stmt
                literal
        scalar
          token
          token
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 121 characters
  Type Annotations:
    VarAnnotation: $object :: MyClass|OtherClass at 1:1
    VarAnnotation: $qualified :: Package::Type1|Package::Type2 at 2:1
    VarAnnotation: $flexible :: UserType|SystemType|DefaultType at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          union_type
            type_expression
              expression_stmt
                literal
            expression_stmt
              literal
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
          union_type
            type_expression
              expression_stmt
                literal
            expression_stmt
              literal
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
          union_type
            type_expression
              union_type
                type_expression
                  expression_stmt
                    literal
                expression_stmt
                  literal
                type_expression
                  expression_stmt
                    literal
            expression_stmt
              literal
            type_expression
              expression_stmt
                literal
        scalar
          token
          token
    token
}
```

# Expected Type Errors

```
(none)
```
