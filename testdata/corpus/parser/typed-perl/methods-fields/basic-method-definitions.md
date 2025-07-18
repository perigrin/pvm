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
method Int calculate(Int $a, Int $b) {
    return $a + $b;
}

method Str greet(Str $name) {
    return "Hello, $name!";
}

method Bool is_valid(Bool $flag) {
    return $flag;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 177 characters
  Type Annotations:
    MethodReturnAnnotation: calculate :: Int at 1:8
    MethodParamAnnotation: $a :: Int at 1:22
    MethodParamAnnotation: $b :: Int at 1:30
    MethodReturnAnnotation: greet :: Str at 5:8
    MethodParamAnnotation: $name :: Str at 5:18
    MethodReturnAnnotation: is_valid :: Bool at 9:8
    MethodParamAnnotation: $flag :: Bool at 9:22
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

## JSON AST

```json
{
  "path": "/tmp/test_code_2.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 11,
      "Column": 2,
      "Offset": 177
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
              "Column": 38,
              "Offset": 37
            },
            "end": {
              "Line": 3,
              "Column": 2,
              "Offset": 60
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 38,
                  "Offset": 37
                },
                "end": {
                  "Line": 1,
                  "Column": 39,
                  "Offset": 38
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 43
                },
                "end": {
                  "Line": 2,
                  "Column": 19,
                  "Offset": 57
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 43
                    },
                    "end": {
                      "Line": 2,
                      "Column": 19,
                      "Offset": 57
                    },
                    "value": "return $a + $b",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 19,
                  "Offset": 57
                },
                "end": {
                  "Line": 2,
                  "Column": 20,
                  "Offset": 58
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 59
                },
                "end": {
                  "Line": 3,
                  "Column": 2,
                  "Offset": 60
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "calculate"
      },
      {
        "type": "method_decl",
        "start": {
          "Line": 5,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 7,
          "Column": 2,
          "Offset": 0
        },
        "children": [
          {
            "type": "block_stmt",
            "start": {
              "Line": 5,
              "Column": 29,
              "Offset": 90
            },
            "end": {
              "Line": 7,
              "Column": 2,
              "Offset": 121
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 5,
                  "Column": 29,
                  "Offset": 90
                },
                "end": {
                  "Line": 5,
                  "Column": 30,
                  "Offset": 91
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 6,
                  "Column": 5,
                  "Offset": 96
                },
                "end": {
                  "Line": 6,
                  "Column": 27,
                  "Offset": 118
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 6,
                      "Column": 5,
                      "Offset": 96
                    },
                    "end": {
                      "Line": 6,
                      "Column": 27,
                      "Offset": 118
                    },
                    "value": "return \"Hello, $name!\"",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 6,
                  "Column": 27,
                  "Offset": 118
                },
                "end": {
                  "Line": 6,
                  "Column": 28,
                  "Offset": 119
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 7,
                  "Column": 1,
                  "Offset": 120
                },
                "end": {
                  "Line": 7,
                  "Column": 2,
                  "Offset": 121
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "greet"
      },
      {
        "type": "method_decl",
        "start": {
          "Line": 9,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 11,
          "Column": 2,
          "Offset": 0
        },
        "children": [
          {
            "type": "block_stmt",
            "start": {
              "Line": 9,
              "Column": 34,
              "Offset": 156
            },
            "end": {
              "Line": 11,
              "Column": 2,
              "Offset": 177
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 9,
                  "Column": 34,
                  "Offset": 156
                },
                "end": {
                  "Line": 9,
                  "Column": 35,
                  "Offset": 157
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 10,
                  "Column": 5,
                  "Offset": 162
                },
                "end": {
                  "Line": 10,
                  "Column": 17,
                  "Offset": 174
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 10,
                      "Column": 5,
                      "Offset": 162
                    },
                    "end": {
                      "Line": 10,
                      "Column": 17,
                      "Offset": 174
                    },
                    "value": "return $flag",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 10,
                  "Column": 17,
                  "Offset": 174
                },
                "end": {
                  "Line": 10,
                  "Column": 18,
                  "Offset": 175
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 11,
                  "Column": 1,
                  "Offset": 176
                },
                "end": {
                  "Line": 11,
                  "Column": 2,
                  "Offset": 177
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "is_valid"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "calculate",
      "type_expression": {
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
      "position": {
        "Line": 1,
        "Column": 8,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$a",
      "type_expression": {
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
      "position": {
        "Line": 1,
        "Column": 22,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "$b",
      "type_expression": {
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
      "position": {
        "Line": 1,
        "Column": 30,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "greet",
      "type_expression": {
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
      "position": {
        "Line": 5,
        "Column": 8,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$name",
      "type_expression": {
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
      "position": {
        "Line": 5,
        "Column": 18,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "is_valid",
      "type_expression": {
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
      "position": {
        "Line": 9,
        "Column": 8,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$flag",
      "type_expression": {
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
      "position": {
        "Line": 9,
        "Column": 22,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    }
  ],
  "errors": [],
  "source_length": 177
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 177 characters
  Type Annotations:
    MethodReturnAnnotation: calculate :: Int at 1:8
    MethodParamAnnotation: $a :: Int at 1:22
    MethodParamAnnotation: $b :: Int at 1:30
    MethodReturnAnnotation: greet :: Str at 5:8
    MethodParamAnnotation: $name :: Str at 5:18
    MethodReturnAnnotation: is_valid :: Bool at 9:8
    MethodParamAnnotation: $flag :: Bool at 9:22
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
method calculate($a, $b) {
    return $a + $b;
}

method greet($name) {
    return "Hello, $name!";
}

method is_valid($flag) {
    return $flag;
}
```

## Typed Perl Output

```perl
method Int calculate(Int $a, Int $b) {
    return $a + $b;
}

method Str greet(Str $name) {
    return "Hello, $name!";
}

method Bool is_valid(Bool $flag) {
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
