---
category: typed-perl
subcategory: union-types
tags:
    - multi-way-unions
    - union-types
    - variable-declarations
type_check: true
---

# Multi Way Unions

Multi-way union types with three or more alternatives

```perl
my Int|Str|Bool $multi = "text";
my Num|ArrayRef|HashRef $complex;
my Int|Str|Bool|Undef $nullable = undef;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 107 characters
  Type Annotations:
    VarAnnotation: $multi :: Int|Str|Bool at 1:1
    VarAnnotation: $complex :: Num|ArrayRef|HashRef at 2:1
    VarAnnotation: $nullable :: Int|Str|Bool|Undef at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
        literal
    token
    expression_stmt
      literal
    token
    expression_statement
      var_decl
        variable
        literal
    token
}
```

## Text AST

```
AST {
  Path: /tmp/multi-way-unions.pl
  Source length: 108 characters
  Type Annotations:
    VarAnnotation: $multi :: Int|Str|Bool at 1:1
    VarAnnotation: $complex :: Num|ArrayRef|HashRef at 2:1
    VarAnnotation: $nullable :: Int|Str|Bool|Undef at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
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
    expression_statement
      var_decl
        variable
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/multi-way-unions.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 1,
      "Offset": 108
    },
    "children": [
      {
        "type": "expression_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 1,
          "Column": 32,
          "Offset": 31
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 32,
              "Offset": 31
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 32,
                  "Offset": 31
                },
                "name": "Int|Str|Bool",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$multi",
      "type_expression": {
        "Kind": 0,
        "Name": "Int|Str|Bool",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
            "Kind": 0,
            "Name": "Int",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Int"
          },
          {
            "Kind": 0,
            "Name": "Str",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Str"
          },
          {
            "Kind": 0,
            "Name": "Bool",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Bool"
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Int|Str|Bool"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$complex",
      "type_expression": {
        "Kind": 0,
        "Name": "Num|ArrayRef|HashRef",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
            "Kind": 0,
            "Name": "Num",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Num"
          },
          {
            "Kind": 0,
            "Name": "ArrayRef",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "ArrayRef"
          },
          {
            "Kind": 0,
            "Name": "HashRef",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "HashRef"
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Num|ArrayRef|HashRef"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$nullable",
      "type_expression": {
        "Kind": 0,
        "Name": "Int|Str|Bool|Undef",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
            "Kind": 0,
            "Name": "Int",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Int"
          },
          {
            "Kind": 0,
            "Name": "Str",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Str"
          },
          {
            "Kind": 0,
            "Name": "Bool",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Bool"
          },
          {
            "Kind": 0,
            "Name": "Undef",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Undef"
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Int|Str|Bool|Undef"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "source_length": 108
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 107 characters
  Type Annotations:
    VarAnnotation: $multi :: Int|Str|Bool at 1:1
    VarAnnotation: $complex :: Num|ArrayRef|HashRef at 2:1
    VarAnnotation: $nullable :: Int|Str|Bool|Undef at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
        literal
    token
    expression_stmt
      literal
    token
    expression_statement
      var_decl
        variable
        literal
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $multi = "text";
my $complex;
my $nullable = undef;
```

## Typed Perl Output

```perl
my Int|Str|Bool $multi = "text";
my Num|ArrayRef|HashRef $complex;
my Int|Str|Bool|Undef $nullable = undef;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
