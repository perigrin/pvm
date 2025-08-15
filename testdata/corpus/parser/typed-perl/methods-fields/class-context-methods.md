---
category: typed-perl
subcategory: methods-fields
tags:
    - class-context
    - class-definitions
    - typed-methods
    - typed-fields
type_check: true
---

# Class Context Methods

Methods and fields within class context with type annotations

```perl
class Calculator {
    field Num $precision = 0.001;

    method Num add(Num $a, Num $b) {
        return $a + $b;
    }

    method Num get_precision() {
        return $precision;
    }

    method Void set_precision(Num $new_precision) {
        $precision = $new_precision;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 285 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Num at 4:12
    MethodParamAnnotation: $a :: Num at 4:20
    MethodParamAnnotation: $b :: Num at 4:28
    MethodReturnAnnotation: get_precision :: Num at 8:12
    MethodReturnAnnotation: set_precision :: Void at 12:12
    MethodParamAnnotation: $new_precision :: Num at 12:31
    VarAnnotation: Calculator :: class at 1:1
    VarAnnotation: $precision :: Num at 2:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            variable
          token
          token
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
  "path": "/tmp/test_code_3.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 15,
      "Column": 2,
      "Offset": 285
    },
    "children": [
      {
        "type": "class_decl",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 15,
          "Column": 2,
          "Offset": 285
        },
        "children": [
          {
            "type": "method_decl",
            "start": {
              "Line": 4,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 6,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 4,
                  "Column": 36,
                  "Offset": 89
                },
                "end": {
                  "Line": 6,
                  "Column": 6,
                  "Offset": 120
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 36,
                      "Offset": 89
                    },
                    "end": {
                      "Line": 4,
                      "Column": 37,
                      "Offset": 90
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 5,
                      "Column": 9,
                      "Offset": 99
                    },
                    "end": {
                      "Line": 5,
                      "Column": 23,
                      "Offset": 113
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 5,
                          "Column": 9,
                          "Offset": 99
                        },
                        "end": {
                          "Line": 5,
                          "Column": 23,
                          "Offset": 113
                        },
                        "value": "return $a + $b",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 5,
                      "Column": 23,
                      "Offset": 113
                    },
                    "end": {
                      "Line": 5,
                      "Column": 24,
                      "Offset": 114
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 5,
                      "Offset": 119
                    },
                    "end": {
                      "Line": 6,
                      "Column": 6,
                      "Offset": 120
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "add"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 8,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 10,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 8,
                  "Column": 32,
                  "Offset": 153
                },
                "end": {
                  "Line": 10,
                  "Column": 6,
                  "Offset": 187
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 8,
                      "Column": 32,
                      "Offset": 153
                    },
                    "end": {
                      "Line": 8,
                      "Column": 33,
                      "Offset": 154
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 9,
                      "Column": 9,
                      "Offset": 163
                    },
                    "end": {
                      "Line": 9,
                      "Column": 26,
                      "Offset": 180
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 9,
                          "Column": 9,
                          "Offset": 163
                        },
                        "end": {
                          "Line": 9,
                          "Column": 26,
                          "Offset": 180
                        },
                        "value": "return $precision",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 9,
                      "Column": 26,
                      "Offset": 180
                    },
                    "end": {
                      "Line": 9,
                      "Column": 27,
                      "Offset": 181
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 10,
                      "Column": 5,
                      "Offset": 186
                    },
                    "end": {
                      "Line": 10,
                      "Column": 6,
                      "Offset": 187
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "get_precision"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 12,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 14,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 12,
                  "Column": 51,
                  "Offset": 239
                },
                "end": {
                  "Line": 14,
                  "Column": 6,
                  "Offset": 283
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 12,
                      "Column": 51,
                      "Offset": 239
                    },
                    "end": {
                      "Line": 12,
                      "Column": 52,
                      "Offset": 240
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 13,
                      "Column": 9,
                      "Offset": 249
                    },
                    "end": {
                      "Line": 13,
                      "Column": 36,
                      "Offset": 276
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 13,
                          "Column": 9,
                          "Offset": 249
                        },
                        "end": {
                          "Line": 13,
                          "Column": 36,
                          "Offset": 276
                        },
                        "value": "$precision = $new_precision",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 13,
                      "Column": 36,
                      "Offset": 276
                    },
                    "end": {
                      "Line": 13,
                      "Column": 37,
                      "Offset": 277
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 14,
                      "Column": 5,
                      "Offset": 282
                    },
                    "end": {
                      "Line": 14,
                      "Column": 6,
                      "Offset": 283
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "set_precision"
          }
        ],
        "name": "Calculator"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "add",
      "type_expression": {
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
      "position": {
        "Line": 4,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$a",
      "type_expression": {
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
      "position": {
        "Line": 4,
        "Column": 20,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "$b",
      "type_expression": {
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
      "position": {
        "Line": 4,
        "Column": 28,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "get_precision",
      "type_expression": {
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
      "position": {
        "Line": 8,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "set_precision",
      "type_expression": {
        "Kind": 0,
        "Name": "Void",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Void"
      },
      "position": {
        "Line": 12,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$new_precision",
      "type_expression": {
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
      "position": {
        "Line": 12,
        "Column": 31,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "Calculator",
      "type_expression": {
        "Kind": 0,
        "Name": "class",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "class"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$precision",
      "type_expression": {
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
      "position": {
        "Line": 2,
        "Column": 5,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 285
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 285 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Num at 4:12
    MethodParamAnnotation: $a :: Num at 4:20
    MethodParamAnnotation: $b :: Num at 4:28
    MethodReturnAnnotation: get_precision :: Num at 8:12
    MethodReturnAnnotation: set_precision :: Void at 12:12
    MethodParamAnnotation: $new_precision :: Num at 12:31
    VarAnnotation: Calculator :: class at 1:1
    VarAnnotation: $precision :: Num at 2:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            variable
          token
          token
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
class Calculator {
    field $precision = 0.001;

    method add($a, $b) {
        return $a + $b;
    }

    method get_precision() {
        return $precision;
    }

    method set_precision($new_precision) {
        $precision = $new_precision;
    }
}
```

## Typed Perl Output

```perl
class Calculator {
    field Num $precision = 0.001;

    method Num add(Num $a, Num $b) {
        return $a + $b;
    }

    method Num get_precision() {
        return $precision;
    }

    method Void set_precision(Num $new_precision) {
        $precision = $new_precision;
    }
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
