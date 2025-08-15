---
category: typed-perl
subcategory: union-types
tags:
    - nested-contexts
    - complex-parameters
    - method-signatures
    - union-types
type_check: true
---

# Nested Contexts

Union types in nested method signature contexts with complex parameters

```perl
method HashRef[Bool|Str] complex(ArrayRef[Int|Str] $data, CodeRef|Undef $callback) {
    return {};
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 101 characters
  Type Annotations:
    MethodReturnAnnotation: complex :: HashRef[Bool|Str] at 1:8
    MethodParamAnnotation: $data :: ArrayRef[Int|Str] at 1:34
    MethodParamAnnotation: $callback :: CodeRef|Undef at 1:59
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

## Text AST

```
AST {
  Path: /tmp/nested-contexts.pl
  Source length: 102 characters
  Type Annotations:
    MethodReturnAnnotation: complex :: HashRef[Bool|Str] at 1:8
    MethodParamAnnotation: $data :: ArrayRef[Int|Str] at 1:34
    MethodParamAnnotation: $callback :: CodeRef|Undef at 1:59
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

## JSON AST

```json
{
  "path": "/tmp/nested-contexts.pl",
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
      "Offset": 102
    },
    "children": [
      {
        "type": "method_decl",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 3,
          "Column": 2,
          "Offset": 0
        },
        "children": [
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 84,
              "Offset": 83
            },
            "end": {
              "Line": 3,
              "Column": 2,
              "Offset": 101
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 84,
                  "Offset": 83
                },
                "end": {
                  "Line": 1,
                  "Column": 85,
                  "Offset": 84
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 89
                },
                "end": {
                  "Line": 2,
                  "Column": 14,
                  "Offset": 98
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 89
                    },
                    "end": {
                      "Line": 2,
                      "Column": 14,
                      "Offset": 98
                    },
                    "value": "return {}",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 14,
                  "Offset": 98
                },
                "end": {
                  "Line": 2,
                  "Column": 15,
                  "Offset": 99
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 100
                },
                "end": {
                  "Line": 3,
                  "Column": 2,
                  "Offset": 101
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "complex"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "complex",
      "type_expression": {
        "Kind": 4,
        "Name": "HashRef[Bool|Str]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "Bool|Str",
            "Parameters": null,
            "IsUnion": true,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": [
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
              }
            ],
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Bool|Str"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "HashRef[Bool|Str]"
      },
      "position": {
        "Line": 1,
        "Column": 8,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$data",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[Int|Str]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "Int|Str",
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
              }
            ],
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Int|Str"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[Int|Str]"
      },
      "position": {
        "Line": 1,
        "Column": 34,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "$callback",
      "type_expression": {
        "Kind": 0,
        "Name": "CodeRef|Undef",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
            "Kind": 0,
            "Name": "CodeRef",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "CodeRef"
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
        "OriginalString": "CodeRef|Undef"
      },
      "position": {
        "Line": 1,
        "Column": 59,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    }
  ],
  "source_length": 102
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 101 characters
  Type Annotations:
    MethodReturnAnnotation: complex :: HashRef[Bool|Str] at 1:8
    MethodParamAnnotation: $data :: ArrayRef[Int|Str] at 1:34
    MethodParamAnnotation: $callback :: CodeRef|Undef at 1:59
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
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
method complex($data, $callback) {
    return {};
}
```

## Typed Perl Output

```perl
method HashRef[Bool|Str] complex(ArrayRef[Int|Str] $data, CodeRef|Undef $callback) {
    return {};
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
