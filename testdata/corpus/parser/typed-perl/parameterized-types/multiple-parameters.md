---
category: typed-perl
subcategory: parameterized-types
tags:
    - multiple-parameters
    - Map
    - Tuple
    - Function
    - parameterized-types
type_check: true
---

# Multiple Parameters

Parameterized types with multiple type parameters

```perl
my Map[Str, Int] %mapping;
my Tuple[Int, Str, Bool] $triple;
my Function[Int, Str, Bool] $complex_func;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 103 characters
  Type Annotations:
    VarAnnotation: %mapping :: Map[Str, Int] at 1:1
    VarAnnotation: $triple :: Tuple[Int, Str, Bool] at 2:1
    VarAnnotation: $complex_func :: Function[Int, Str, Bool] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 103 characters
  Type Annotations:
    VarAnnotation: %mapping :: Map[Str, Int] at 1:1
    VarAnnotation: $triple :: Tuple[Int, Str, Bool] at 2:1
    VarAnnotation: $complex_func :: Function[Int, Str, Bool] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
}
```

## Text AST

```
source_file
  expression_statement
    variable_declaration
      token (my)
      type_expression
        parameterized_type
          expression_stmt
            literal (Map)
          expression_stmt
            literal ([)
          type_parameter_list
            type_expression
              expression_stmt
                literal (Str)
            expression_stmt
              literal (,)
            type_expression
              expression_stmt
                literal (Int)
          expression_stmt
            literal (])
      hash
        expression_stmt
          literal (%)
        token (mapping)
  token (;)
  expression_statement
    variable_declaration
      token (my)
      type_expression
        parameterized_type
          expression_stmt
            literal (Tuple)
          expression_stmt
            literal ([)
          type_parameter_list
            type_expression
              expression_stmt
                literal (Int)
            expression_stmt
              literal (,)
            type_expression
              expression_stmt
                literal (Str)
            expression_stmt
              literal (,)
            type_expression
              expression_stmt
                literal (Bool)
          expression_stmt
            literal (])
      scalar
        token ($)
        token (triple)
  token (;)
  expression_statement
    variable_declaration
      token (my)
      type_expression
        parameterized_type
          expression_stmt
            literal (Function)
          expression_stmt
            literal ([)
          type_parameter_list
            type_expression
              expression_stmt
                literal (Int)
            expression_stmt
              literal (,)
            type_expression
              expression_stmt
                literal (Str)
            expression_stmt
              literal (,)
            type_expression
              expression_stmt
                literal (Bool)
          expression_stmt
            literal (])
      scalar
        token ($)
        token (complex_func)
  token (;)
```

## JSON AST

```json
{
  "path": "/tmp/multiple_parameters.pl",
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
      "Offset": 104
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
          "Column": 26,
          "Offset": 25
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
              "Column": 26,
              "Offset": 25
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
                  "Column": 17,
                  "Offset": 16
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
                      "Column": 17,
                      "Offset": 16
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
                          "Column": 7,
                          "Offset": 6
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
                              "Column": 7,
                              "Offset": 6
                            },
                            "value": "Map",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 7,
                          "Offset": 6
                        },
                        "end": {
                          "Line": 1,
                          "Column": 8,
                          "Offset": 7
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 7,
                              "Offset": 6
                            },
                            "end": {
                              "Line": 1,
                              "Column": 8,
                              "Offset": 7
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
                          "Column": 8,
                          "Offset": 7
                        },
                        "end": {
                          "Line": 1,
                          "Column": 16,
                          "Offset": 15
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 1,
                              "Column": 8,
                              "Offset": 7
                            },
                            "end": {
                              "Line": 1,
                              "Column": 11,
                              "Offset": 10
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 1,
                                  "Column": 8,
                                  "Offset": 7
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 11,
                                  "Offset": 10
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 1,
                                      "Column": 8,
                                      "Offset": 7
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 11,
                                      "Offset": 10
                                    },
                                    "value": "Str",
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
                              "Column": 11,
                              "Offset": 10
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
                                  "Column": 11,
                                  "Offset": 10
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 12,
                                  "Offset": 11
                                },
                                "value": ",",
                                "kind": "string"
                              }
                            ]
                          },
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
                "type": "hash",
                "start": {
                  "Line": 1,
                  "Column": 18,
                  "Offset": 17
                },
                "end": {
                  "Line": 1,
                  "Column": 26,
                  "Offset": 25
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 18,
                      "Offset": 17
                    },
                    "end": {
                      "Line": 1,
                      "Column": 19,
                      "Offset": 18
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 18,
                          "Offset": 17
                        },
                        "end": {
                          "Line": 1,
                          "Column": 19,
                          "Offset": 18
                        },
                        "value": "%",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 19,
                      "Offset": 18
                    },
                    "end": {
                      "Line": 1,
                      "Column": 26,
                      "Offset": 25
                    },
                    "text": "mapping"
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
          "Column": 26,
          "Offset": 25
        },
        "end": {
          "Line": 1,
          "Column": 27,
          "Offset": 26
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 27
        },
        "end": {
          "Line": 2,
          "Column": 33,
          "Offset": 59
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 27
            },
            "end": {
              "Line": 2,
              "Column": 33,
              "Offset": 59
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 27
                },
                "end": {
                  "Line": 2,
                  "Column": 3,
                  "Offset": 29
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 30
                },
                "end": {
                  "Line": 2,
                  "Column": 25,
                  "Offset": 51
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 2,
                      "Column": 4,
                      "Offset": 30
                    },
                    "end": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 51
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 4,
                          "Offset": 30
                        },
                        "end": {
                          "Line": 2,
                          "Column": 9,
                          "Offset": 35
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 4,
                              "Offset": 30
                            },
                            "end": {
                              "Line": 2,
                              "Column": 9,
                              "Offset": 35
                            },
                            "value": "Tuple",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 9,
                          "Offset": 35
                        },
                        "end": {
                          "Line": 2,
                          "Column": 10,
                          "Offset": 36
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 9,
                              "Offset": 35
                            },
                            "end": {
                              "Line": 2,
                              "Column": 10,
                              "Offset": 36
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
                          "Column": 10,
                          "Offset": 36
                        },
                        "end": {
                          "Line": 2,
                          "Column": 24,
                          "Offset": 50
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 10,
                              "Offset": 36
                            },
                            "end": {
                              "Line": 2,
                              "Column": 13,
                              "Offset": 39
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 2,
                                  "Column": 10,
                                  "Offset": 36
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 13,
                                  "Offset": 39
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 2,
                                      "Column": 10,
                                      "Offset": 36
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 13,
                                      "Offset": 39
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
                              "Column": 13,
                              "Offset": 39
                            },
                            "end": {
                              "Line": 2,
                              "Column": 14,
                              "Offset": 40
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 2,
                                  "Column": 13,
                                  "Offset": 39
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 14,
                                  "Offset": 40
                                },
                                "value": ",",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 15,
                              "Offset": 41
                            },
                            "end": {
                              "Line": 2,
                              "Column": 18,
                              "Offset": 44
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 2,
                                  "Column": 15,
                                  "Offset": 41
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 18,
                                  "Offset": 44
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 2,
                                      "Column": 15,
                                      "Offset": 41
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 18,
                                      "Offset": 44
                                    },
                                    "value": "Str",
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
                              "Offset": 44
                            },
                            "end": {
                              "Line": 2,
                              "Column": 19,
                              "Offset": 45
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 2,
                                  "Column": 18,
                                  "Offset": 44
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 19,
                                  "Offset": 45
                                },
                                "value": ",",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 20,
                              "Offset": 46
                            },
                            "end": {
                              "Line": 2,
                              "Column": 24,
                              "Offset": 50
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 2,
                                  "Column": 20,
                                  "Offset": 46
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 24,
                                  "Offset": 50
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 2,
                                      "Column": 20,
                                      "Offset": 46
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 24,
                                      "Offset": 50
                                    },
                                    "value": "Bool",
                                    "kind": "string"
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
                          "Column": 24,
                          "Offset": 50
                        },
                        "end": {
                          "Line": 2,
                          "Column": 25,
                          "Offset": 51
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 24,
                              "Offset": 50
                            },
                            "end": {
                              "Line": 2,
                              "Column": 25,
                              "Offset": 51
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
                  "Column": 26,
                  "Offset": 52
                },
                "end": {
                  "Line": 2,
                  "Column": 33,
                  "Offset": 59
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 26,
                      "Offset": 52
                    },
                    "end": {
                      "Line": 2,
                      "Column": 27,
                      "Offset": 53
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 27,
                      "Offset": 53
                    },
                    "end": {
                      "Line": 2,
                      "Column": 33,
                      "Offset": 59
                    },
                    "text": "triple"
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
          "Column": 33,
          "Offset": 59
        },
        "end": {
          "Line": 2,
          "Column": 34,
          "Offset": 60
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 61
        },
        "end": {
          "Line": 3,
          "Column": 42,
          "Offset": 102
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 61
            },
            "end": {
              "Line": 3,
              "Column": 42,
              "Offset": 102
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 61
                },
                "end": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 63
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 64
                },
                "end": {
                  "Line": 3,
                  "Column": 28,
                  "Offset": 88
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 64
                    },
                    "end": {
                      "Line": 3,
                      "Column": 28,
                      "Offset": 88
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 4,
                          "Offset": 64
                        },
                        "end": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 72
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 4,
                              "Offset": 64
                            },
                            "end": {
                              "Line": 3,
                              "Column": 12,
                              "Offset": 72
                            },
                            "value": "Function",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 72
                        },
                        "end": {
                          "Line": 3,
                          "Column": 13,
                          "Offset": 73
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 12,
                              "Offset": 72
                            },
                            "end": {
                              "Line": 3,
                              "Column": 13,
                              "Offset": 73
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
                          "Column": 13,
                          "Offset": 73
                        },
                        "end": {
                          "Line": 3,
                          "Column": 27,
                          "Offset": 87
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 13,
                              "Offset": 73
                            },
                            "end": {
                              "Line": 3,
                              "Column": 16,
                              "Offset": 76
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 3,
                                  "Column": 13,
                                  "Offset": 73
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 16,
                                  "Offset": 76
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 3,
                                      "Column": 13,
                                      "Offset": 73
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 16,
                                      "Offset": 76
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
                              "Column": 16,
                              "Offset": 76
                            },
                            "end": {
                              "Line": 3,
                              "Column": 17,
                              "Offset": 77
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 3,
                                  "Column": 16,
                                  "Offset": 76
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 17,
                                  "Offset": 77
                                },
                                "value": ",",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 18,
                              "Offset": 78
                            },
                            "end": {
                              "Line": 3,
                              "Column": 21,
                              "Offset": 81
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 3,
                                  "Column": 18,
                                  "Offset": 78
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 21,
                                  "Offset": 81
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 3,
                                      "Column": 18,
                                      "Offset": 78
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 21,
                                      "Offset": 81
                                    },
                                    "value": "Str",
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
                              "Column": 21,
                              "Offset": 81
                            },
                            "end": {
                              "Line": 3,
                              "Column": 22,
                              "Offset": 82
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 3,
                                  "Column": 21,
                                  "Offset": 81
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 22,
                                  "Offset": 82
                                },
                                "value": ",",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 23,
                              "Offset": 83
                            },
                            "end": {
                              "Line": 3,
                              "Column": 27,
                              "Offset": 87
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 3,
                                  "Column": 23,
                                  "Offset": 83
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 27,
                                  "Offset": 87
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 3,
                                      "Column": 23,
                                      "Offset": 83
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 27,
                                      "Offset": 87
                                    },
                                    "value": "Bool",
                                    "kind": "string"
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
                          "Column": 27,
                          "Offset": 87
                        },
                        "end": {
                          "Line": 3,
                          "Column": 28,
                          "Offset": 88
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 27,
                              "Offset": 87
                            },
                            "end": {
                              "Line": 3,
                              "Column": 28,
                              "Offset": 88
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
                  "Column": 29,
                  "Offset": 89
                },
                "end": {
                  "Line": 3,
                  "Column": 42,
                  "Offset": 102
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 29,
                      "Offset": 89
                    },
                    "end": {
                      "Line": 3,
                      "Column": 30,
                      "Offset": 90
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 30,
                      "Offset": 90
                    },
                    "end": {
                      "Line": 3,
                      "Column": 42,
                      "Offset": 102
                    },
                    "text": "complex_func"
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
          "Column": 42,
          "Offset": 102
        },
        "end": {
          "Line": 3,
          "Column": 43,
          "Offset": 103
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
my %mapping;
my $triple;
my $complex_func;
```

## Typed Perl Output

```perl
my Map[Str, Int] %mapping;
my Tuple[Int, Str, Bool] $triple;
my Function[Int, Str, Bool] $complex_func;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
