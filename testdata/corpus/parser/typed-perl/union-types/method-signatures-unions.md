---
category: typed-perl
subcategory: union-types
tags:
    - method-signatures
    - parameter-types
    - return-types
    - union-types
type_check: true
---

# Method Signatures Unions

Union types in method parameter and return type signatures

```perl
method Bool|Str process(Int|Str $input) {
    if (ref $input) {
        return "Invalid";
    }
    return 1;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 111 characters
  Type Annotations:
    MethodReturnAnnotation: process :: Bool|Str at 1:8
    MethodParamAnnotation: $input :: Int|Str at 1:25
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        expression_stmt
          literal
        token
        token
}
```

## Text AST

```
AST {
  Path: /tmp/method-signatures-unions.pl
  Source length: 112 characters
  Type Annotations:
    MethodReturnAnnotation: process :: Bool|Str at 1:8
    MethodParamAnnotation: $input :: Int|Str at 1:25
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        expression_stmt
          literal
        token
        token
}
```

## JSON AST

```json
{
  "path": "/tmp/method-signatures-unions.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 7,
      "Column": 1,
      "Offset": 112
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
          "Line": 6,
          "Column": 2,
          "Offset": 0
        },
        "children": [
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 41,
              "Offset": 40
            },
            "end": {
              "Line": 6,
              "Column": 2,
              "Offset": 111
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 41,
                  "Offset": 40
                },
                "end": {
                  "Line": 1,
                  "Column": 42,
                  "Offset": 41
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 46
                },
                "end": {
                  "Line": 4,
                  "Column": 6,
                  "Offset": 95
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 46
                    },
                    "end": {
                      "Line": 4,
                      "Column": 6,
                      "Offset": 95
                    },
                    "value": "if (ref $input) {\n        return \"Invalid\";\n    }",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 5,
                  "Column": 5,
                  "Offset": 100
                },
                "end": {
                  "Line": 5,
                  "Column": 13,
                  "Offset": 108
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 5,
                      "Column": 5,
                      "Offset": 100
                    },
                    "end": {
                      "Line": 5,
                      "Column": 13,
                      "Offset": 108
                    },
                    "value": "return 1",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 5,
                  "Column": 13,
                  "Offset": 108
                },
                "end": {
                  "Line": 5,
                  "Column": 14,
                  "Offset": 109
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 6,
                  "Column": 1,
                  "Offset": 110
                },
                "end": {
                  "Line": 6,
                  "Column": 2,
                  "Offset": 111
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "process"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "process",
      "type_expression": {
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
      },
      "position": {
        "Line": 1,
        "Column": 8,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$input",
      "type_expression": {
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
      },
      "position": {
        "Line": 1,
        "Column": 25,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    }
  ],
  "source_length": 112
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 111 characters
  Type Annotations:
    MethodReturnAnnotation: process :: Bool|Str at 1:8
    MethodParamAnnotation: $input :: Int|Str at 1:25
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
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
method process($input) {
    if (ref $input) {
        return "Invalid";
    }
    return 1;
}
```

## Typed Perl Output

```perl
method Bool|Str process(Int|Str $input) {
    if (ref $input) {
        return "Invalid";
    }
    return 1;
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
