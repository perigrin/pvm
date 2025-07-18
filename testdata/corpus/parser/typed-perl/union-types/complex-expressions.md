---
category: typed-perl
subcategory: union-types
tags:
    - complex-expressions
    - parameterized-types
    - union-types
    - fields
type_check: true
---

# Complex Expressions

Union types within complex type expressions like parameterized types

```perl
my ArrayRef[Int|Str] @mixed_array;
field HashRef[Int|Bool] $mixed_hash;
my CodeRef[Int|Str, Bool|Undef] $flexible_function;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 123 characters
  Type Annotations:
    VarAnnotation: @mixed_array :: ArrayRef[Int|Str] at 1:1
    VarAnnotation: $mixed_hash :: HashRef[Int|Bool] at 2:1
    VarAnnotation: $flexible_function :: CodeRef[Int|Str, Bool|Undef] at 3:1
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
        array
          expression_stmt
            literal
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
  Source length: 123 characters
  Type Annotations:
    VarAnnotation: @mixed_array :: ArrayRef[Int|Str] at 1:1
    VarAnnotation: $mixed_hash :: HashRef[Int|Bool] at 2:1
    VarAnnotation: $flexible_function :: CodeRef[Int|Str, Bool|Undef] at 3:1
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
        array
          expression_stmt
            literal
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
        scalar
          token
          token
    token
}
```

## Text AST

```
AST {
  Path: /tmp/complex-expressions.pl
  Source length: 123 characters
  Type Annotations:
    VarAnnotation: @mixed_array :: ArrayRef[Int|Str] at 1:1
    VarAnnotation: $mixed_hash :: HashRef[Int|Bool] at 2:1
    VarAnnotation: $flexible_function :: CodeRef[Int|Str, Bool|Undef] at 3:1
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
        array
          expression_stmt
            literal
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
        scalar
          token
          token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/complex-expressions.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 52,
      "Offset": 123
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
          "Column": 34,
          "Offset": 33
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 34,
              "Offset": 33
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 3,
                  "Offset": 2
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 1,
                  "Column": 4,
                  "Offset": 3
                },
                "end": {
                  "Line": 1,
                  "Column": 21,
                  "Offset": 20
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 1,
                      "Column": 4,
                      "Offset": 3
                    },
                    "end": {
                      "Line": 1,
                      "Column": 21,
                      "Offset": 20
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 4,
                          "Offset": 3
                        },
                        "end": {
                          "Line": 1,
                          "Column": 12,
                          "Offset": 11
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 4,
                              "Offset": 3
                            },
                            "end": {
                              "Line": 1,
                              "Column": 12,
                              "Offset": 11
                            },
                            "value": "ArrayRef",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 12,
                          "Offset": 11
                        },
                        "end": {
                          "Line": 1,
                          "Column": 13,
                          "Offset": 12
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 12,
                              "Offset": 11
                            },
                            "end": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
                            },
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_parameter_list",
                        "start": {
                          "Line": 1,
                          "Column": 13,
                          "Offset": 12
                        },
                        "end": {
                          "Line": 1,
                          "Column": 20,
                          "Offset": 19
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
                            },
                            "end": {
                              "Line": 1,
                              "Column": 20,
                              "Offset": 19
                            },
                            "children": [
                              {
                                "type": "union_type",
                                "start": {
                                  "Line": 1,
                                  "Column": 13,
                                  "Offset": 12
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 20,
                                  "Offset": 19
                                },
                                "children": [
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 1,
                                      "Column": 13,
                                      "Offset": 12
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 16,
                                      "Offset": 15
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 1,
                                          "Column": 13,
                                          "Offset": 12
                                        },
                                        "end": {
                                          "Line": 1,
                                          "Column": 16,
                                          "Offset": 15
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 1,
                                              "Column": 13,
                                              "Offset": 12
                                            },
                                            "end": {
                                              "Line": 1,
                                              "Column": 16,
                                              "Offset": 15
                                            },
                                            "value": "Int",
                                            "kind": "string"
                                          }
                                        ]
                                      }
                                    ]
                                  },
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 1,
                                      "Column": 16,
                                      "Offset": 15
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 17,
                                      "Offset": 16
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 1,
                                          "Column": 16,
                                          "Offset": 15
                                        },
                                        "end": {
                                          "Line": 1,
                                          "Column": 17,
                                          "Offset": 16
                                        },
                                        "value": "|",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 1,
                                      "Column": 17,
                                      "Offset": 16
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 20,
                                      "Offset": 19
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 1,
                                          "Column": 17,
                                          "Offset": 16
                                        },
                                        "end": {
                                          "Line": 1,
                                          "Column": 20,
                                          "Offset": 19
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 1,
                                              "Column": 17,
                                              "Offset": 16
                                            },
                                            "end": {
                                              "Line": 1,
                                              "Column": 20,
                                              "Offset": 19
                                            },
                                            "value": "Str",
                                            "kind": "string"
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 20,
                          "Offset": 19
                        },
                        "end": {
                          "Line": 1,
                          "Column": 21,
                          "Offset": 20
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 20,
                              "Offset": 19
                            },
                            "end": {
                              "Line": 1,
                              "Column": 21,
                              "Offset": 20
                            },
                            "value": "]",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "array",
                "start": {
                  "Line": 1,
                  "Column": 22,
                  "Offset": 21
                },
                "end": {
                  "Line": 1,
                  "Column": 34,
                  "Offset": 33
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 22,
                      "Offset": 21
                    },
                    "end": {
                      "Line": 1,
                      "Column": 23,
                      "Offset": 22
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 22,
                          "Offset": 21
                        },
                        "end": {
                          "Line": 1,
                          "Column": 23,
                          "Offset": 22
                        },
                        "value": "@",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 23,
                      "Offset": 22
                    },
                    "end": {
                      "Line": 1,
                      "Column": 34,
                      "Offset": 33
                    },
                    "text": "mixed_array"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 34,
          "Offset": 33
        },
        "end": {
          "Line": 1,
          "Column": 35,
          "Offset": 34
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 35
        },
        "end": {
          "Line": 2,
          "Column": 36,
          "Offset": 70
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 35
            },
            "end": {
              "Line": 2,
              "Column": 36,
              "Offset": 70
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 35
                },
                "end": {
                  "Line": 2,
                  "Column": 6,
                  "Offset": 40
                },
                "text": "field"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 7,
                  "Offset": 41
                },
                "end": {
                  "Line": 2,
                  "Column": 24,
                  "Offset": 58
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 2,
                      "Column": 7,
                      "Offset": 41
                    },
                    "end": {
                      "Line": 2,
                      "Column": 24,
                      "Offset": 58
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 7,
                          "Offset": 41
                        },
                        "end": {
                          "Line": 2,
                          "Column": 14,
                          "Offset": 48
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 7,
                              "Offset": 41
                            },
                            "end": {
                              "Line": 2,
                              "Column": 14,
                              "Offset": 48
                            },
                            "value": "HashRef",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 14,
                          "Offset": 48
                        },
                        "end": {
                          "Line": 2,
                          "Column": 15,
                          "Offset": 49
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 14,
                              "Offset": 48
                            },
                            "end": {
                              "Line": 2,
                              "Column": 15,
                              "Offset": 49
                            },
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_parameter_list",
                        "start": {
                          "Line": 2,
                          "Column": 15,
                          "Offset": 49
                        },
                        "end": {
                          "Line": 2,
                          "Column": 23,
                          "Offset": 57
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 15,
                              "Offset": 49
                            },
                            "end": {
                              "Line": 2,
                              "Column": 23,
                              "Offset": 57
                            },
                            "children": [
                              {
                                "type": "union_type",
                                "start": {
                                  "Line": 2,
                                  "Column": 15,
                                  "Offset": 49
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 23,
                                  "Offset": 57
                                },
                                "children": [
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 2,
                                      "Column": 15,
                                      "Offset": 49
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 18,
                                      "Offset": 52
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 2,
                                          "Column": 15,
                                          "Offset": 49
                                        },
                                        "end": {
                                          "Line": 2,
                                          "Column": 18,
                                          "Offset": 52
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 2,
                                              "Column": 15,
                                              "Offset": 49
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 18,
                                              "Offset": 52
                                            },
                                            "value": "Int",
                                            "kind": "string"
                                          }
                                        ]
                                      }
                                    ]
                                  },
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 2,
                                      "Column": 18,
                                      "Offset": 52
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 19,
                                      "Offset": 53
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 2,
                                          "Column": 18,
                                          "Offset": 52
                                        },
                                        "end": {
                                          "Line": 2,
                                          "Column": 19,
                                          "Offset": 53
                                        },
                                        "value": "|",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 2,
                                      "Column": 19,
                                      "Offset": 53
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 23,
                                      "Offset": 57
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 2,
                                          "Column": 19,
                                          "Offset": 53
                                        },
                                        "end": {
                                          "Line": 2,
                                          "Column": 23,
                                          "Offset": 57
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 2,
                                              "Column": 19,
                                              "Offset": 53
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 23,
                                              "Offset": 57
                                            },
                                            "value": "Bool",
                                            "kind": "string"
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 23,
                          "Offset": 57
                        },
                        "end": {
                          "Line": 2,
                          "Column": 24,
                          "Offset": 58
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 23,
                              "Offset": 57
                            },
                            "end": {
                              "Line": 2,
                              "Column": 24,
                              "Offset": 58
                            },
                            "value": "]",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 2,
                  "Column": 25,
                  "Offset": 59
                },
                "end": {
                  "Line": 2,
                  "Column": 36,
                  "Offset": 70
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 59
                    },
                    "end": {
                      "Line": 2,
                      "Column": 26,
                      "Offset": 60
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 26,
                      "Offset": 60
                    },
                    "end": {
                      "Line": 2,
                      "Column": 36,
                      "Offset": 70
                    },
                    "text": "mixed_hash"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 2,
          "Column": 36,
          "Offset": 70
        },
        "end": {
          "Line": 2,
          "Column": 37,
          "Offset": 71
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 72
        },
        "end": {
          "Line": 3,
          "Column": 51,
          "Offset": 122
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 72
            },
            "end": {
              "Line": 3,
              "Column": 51,
              "Offset": 122
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 72
                },
                "end": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 74
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 75
                },
                "end": {
                  "Line": 3,
                  "Column": 34,
                  "Offset": 105
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 75
                    },
                    "end": {
                      "Line": 3,
                      "Column": 34,
                      "Offset": 105
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 4,
                          "Offset": 75
                        },
                        "end": {
                          "Line": 3,
                          "Column": 11,
                          "Offset": 82
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 4,
                              "Offset": 75
                            },
                            "end": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 82
                            },
                            "value": "CodeRef",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 11,
                          "Offset": 82
                        },
                        "end": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 83
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 82
                            },
                            "end": {
                              "Line": 3,
                              "Column": 12,
                              "Offset": 83
                            },
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_parameter_list",
                        "start": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 83
                        },
                        "end": {
                          "Line": 3,
                          "Column": 33,
                          "Offset": 104
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 12,
                              "Offset": 83
                            },
                            "end": {
                              "Line": 3,
                              "Column": 19,
                              "Offset": 90
                            },
                            "children": [
                              {
                                "type": "union_type",
                                "start": {
                                  "Line": 3,
                                  "Column": 12,
                                  "Offset": 83
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 19,
                                  "Offset": 90
                                },
                                "children": [
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 3,
                                      "Column": 12,
                                      "Offset": 83
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 15,
                                      "Offset": 86
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 3,
                                          "Column": 12,
                                          "Offset": 83
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 15,
                                          "Offset": 86
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 3,
                                              "Column": 12,
                                              "Offset": 83
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 15,
                                              "Offset": 86
                                            },
                                            "value": "Int",
                                            "kind": "string"
                                          }
                                        ]
                                      }
                                    ]
                                  },
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 3,
                                      "Column": 15,
                                      "Offset": 86
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 16,
                                      "Offset": 87
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 15,
                                          "Offset": 86
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 16,
                                          "Offset": 87
                                        },
                                        "value": "|",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 3,
                                      "Column": 16,
                                      "Offset": 87
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 19,
                                      "Offset": 90
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 3,
                                          "Column": 16,
                                          "Offset": 87
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 19,
                                          "Offset": 90
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 3,
                                              "Column": 16,
                                              "Offset": 87
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 19,
                                              "Offset": 90
                                            },
                                            "value": "Str",
                                            "kind": "string"
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 3,
                              "Column": 19,
                              "Offset": 90
                            },
                            "end": {
                              "Line": 3,
                              "Column": 21,
                              "Offset": 92
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 3,
                                  "Column": 19,
                                  "Offset": 90
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 21,
                                  "Offset": 92
                                },
                                "value": ", ",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 21,
                              "Offset": 92
                            },
                            "end": {
                              "Line": 3,
                              "Column": 33,
                              "Offset": 104
                            },
                            "children": [
                              {
                                "type": "union_type",
                                "start": {
                                  "Line": 3,
                                  "Column": 21,
                                  "Offset": 92
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 33,
                                  "Offset": 104
                                },
                                "children": [
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 3,
                                      "Column": 21,
                                      "Offset": 92
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 25,
                                      "Offset": 96
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 3,
                                          "Column": 21,
                                          "Offset": 92
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 25,
                                          "Offset": 96
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 3,
                                              "Column": 21,
                                              "Offset": 92
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 25,
                                              "Offset": 96
                                            },
                                            "value": "Bool",
                                            "kind": "string"
                                          }
                                        ]
                                      }
                                    ]
                                  },
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 3,
                                      "Column": 25,
                                      "Offset": 96
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 26,
                                      "Offset": 97
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 25,
                                          "Offset": 96
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 26,
                                          "Offset": 97
                                        },
                                        "value": "|",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 3,
                                      "Column": 26,
                                      "Offset": 97
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 33,
                                      "Offset": 104
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 3,
                                          "Column": 26,
                                          "Offset": 97
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 33,
                                          "Offset": 104
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 3,
                                              "Column": 26,
                                              "Offset": 97
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 33,
                                              "Offset": 104
                                            },
                                            "value": "Undef",
                                            "kind": "string"
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 33,
                          "Offset": 104
                        },
                        "end": {
                          "Line": 3,
                          "Column": 34,
                          "Offset": 105
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 33,
                              "Offset": 104
                            },
                            "end": {
                              "Line": 3,
                              "Column": 34,
                              "Offset": 105
                            },
                            "value": "]",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 3,
                  "Column": 35,
                  "Offset": 106
                },
                "end": {
                  "Line": 3,
                  "Column": 51,
                  "Offset": 122
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 35,
                      "Offset": 106
                    },
                    "end": {
                      "Line": 3,
                      "Column": 36,
                      "Offset": 107
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 36,
                      "Offset": 107
                    },
                    "end": {
                      "Line": 3,
                      "Column": 51,
                      "Offset": 122
                    },
                    "text": "flexible_function"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 3,
          "Column": 51,
          "Offset": 122
        },
        "end": {
          "Line": 3,
          "Column": 52,
          "Offset": 123
        },
        "text": ";"
      }
    ]
  }
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @mixed_array;
field $mixed_hash;
my $flexible_function;
```

## Typed Perl Output

```perl
my ArrayRef[Int|Str] @mixed_array;
field HashRef[Int|Bool] $mixed_hash;
my CodeRef[Int|Str, Bool|Undef] $flexible_function;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
