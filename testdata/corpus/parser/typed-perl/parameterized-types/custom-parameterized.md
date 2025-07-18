---
category: typed-perl
subcategory: parameterized-types
tags:
    - custom-types
    - generics
    - package-qualified
    - parameterized-types
type_check: true
---

# Custom Parameterized

Custom parameterized types and package-qualified generics

```perl
my Container[MyType] $custom_container;
my Package::Generic[Int] $qualified;
my Result[UserData, ErrorCode] $result;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 116 characters
  Type Annotations:
    VarAnnotation: $custom_container :: Container[MyType] at 1:1
    VarAnnotation: $qualified :: Package::Generic[Int] at 2:1
    VarAnnotation: $result :: Result[UserData, ErrorCode] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                expression_stmt
                  literal
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
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                expression_stmt
                  literal
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
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
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
  Source length: 116 characters
  Type Annotations:
    VarAnnotation: $custom_container :: Container[MyType] at 1:1
    VarAnnotation: $qualified :: Package::Generic[Int] at 2:1
    VarAnnotation: $result :: Result[UserData, ErrorCode] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                expression_stmt
                  literal
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
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                expression_stmt
                  literal
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
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
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
        scalar
          token
          token
    token
}
```

## JSON AST

```json
{
  "path": "custom-parameterized.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 1, "Column": 30, "Offset": 29},
    "children": [
      {
        "type": "expression_statement",
        "start": {"Line": 1, "Column": 1, "Offset": 0},
        "end": {"Line": 1, "Column": 29, "Offset": 28},
        "children": [
          {
            "type": "variable_declaration",
            "start": {"Line": 1, "Column": 1, "Offset": 0},
            "end": {"Line": 1, "Column": 29, "Offset": 28},
            "children": [
              {
                "type": "type_expression",
                "start": {"Line": 1, "Column": 4, "Offset": 3},
                "end": {"Line": 1, "Column": 21, "Offset": 20},
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {"Line": 1, "Column": 4, "Offset": 3},
                    "end": {"Line": 1, "Column": 21, "Offset": 20}
                  }
                ]
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$custom",
      "type_expression": {
        "Kind": 4,
        "Name": "Container[MyType]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "MyType",
            "OriginalString": "MyType"
          }
        ],
        "OriginalString": "Container[MyType]"
      },
      "position": {"Line": 1, "Column": 1, "Offset": 0},
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 29
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $custom_container;
my $qualified;
my $result;
```

## Typed Perl Output

```perl
my Container[MyType] $custom_container;
my Package::Generic[Int] $qualified;
my Result[UserData, ErrorCode] $result;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
