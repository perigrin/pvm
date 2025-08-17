---
category: typed-perl
subcategory: parameterized-types
tags:
    - complex-combinations
    - complex-nesting
    - deep-nesting
    - unions
    - Map
    - parameterized-types
type_check: true
---

# Complex Combinations

Complex combinations of parameterized types with unions and deep nesting

```perl
my Map[Str, ArrayRef[HashRef[Int|Bool]]] $complex;
my Container[ArrayRef[MyType]|HashRef[OtherType]] $flexible;
my Result[Data[UserInfo], Error[ValidationFailure]] $nested_result;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 179 characters
  Type Annotations:
    VarAnnotation: $complex :: Map[Str, ArrayRef[HashRef[Int|Bool]]] at 1:1
    VarAnnotation: $flexible :: Container[ArrayRef[MyType]|HashRef[OtherType]] at 2:1
    VarAnnotation: $nested_result :: Result[Data[UserInfo], Error[ValidationFailure]] at 3:1
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
  Source length: 179 characters
  Type Annotations:
    VarAnnotation: $complex :: Map[Str, ArrayRef[HashRef[Int|Bool]]] at 1:1
    VarAnnotation: $flexible :: Container[ArrayRef[MyType]|HashRef[OtherType]] at 2:1
    VarAnnotation: $nested_result :: Result[Data[UserInfo], Error[ValidationFailure]] at 3:1
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
              parameterized_type
                expression_stmt
                  literal (ArrayRef)
                expression_stmt
                  literal ([)
                type_parameter_list
                  type_expression
                    parameterized_type
                      expression_stmt
                        literal (HashRef)
                      expression_stmt
                        literal ([)
                      type_parameter_list
                        type_expression
                          union_type
                            type_expression
                              expression_stmt
                                literal (Int)
                            expression_stmt
                              literal (|)
                            type_expression
                              expression_stmt
                                literal (Bool)
                      expression_stmt
                        literal (])
                expression_stmt
                  literal (])
          expression_stmt
            literal (])
      scalar
        token ($)
        token (complex)
  token (;)
  expression_statement
    variable_declaration
      token (my)
      type_expression
        parameterized_type
          expression_stmt
            literal (Container)
          expression_stmt
            literal ([)
          type_parameter_list
            type_expression
              union_type
                type_expression
                  parameterized_type
                    expression_stmt
                      literal (ArrayRef)
                    expression_stmt
                      literal ([)
                    type_parameter_list
                      type_expression
                        expression_stmt
                          literal (MyType)
                    expression_stmt
                      literal (])
                expression_stmt
                  literal (|)
                type_expression
                  parameterized_type
                    expression_stmt
                      literal (HashRef)
                    expression_stmt
                      literal ([)
                    type_parameter_list
                      type_expression
                        expression_stmt
                          literal (OtherType)
                    expression_stmt
                      literal (])
          expression_stmt
            literal (])
      scalar
        token ($)
        token (flexible)
  token (;)
  expression_statement
    variable_declaration
      token (my)
      type_expression
        parameterized_type
          expression_stmt
            literal (Result)
          expression_stmt
            literal ([)
          type_parameter_list
            type_expression
              parameterized_type
                expression_stmt
                  literal (Data)
                expression_stmt
                  literal ([)
                type_parameter_list
                  type_expression
                    expression_stmt
                      literal (UserInfo)
                expression_stmt
                  literal (])
            expression_stmt
              literal (,)
            type_expression
              parameterized_type
                expression_stmt
                  literal (Error)
                expression_stmt
                  literal ([)
                type_parameter_list
                  type_expression
                    expression_stmt
                      literal (ValidationFailure)
                expression_stmt
                  literal (])
          expression_stmt
            literal (])
      scalar
        token ($)
        token (nested_result)
  token (;)
```

## JSON AST

```json
{
  "path": "/tmp/complex_combinations.pl",
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
      "Offset": 180
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
          "Column": 50,
          "Offset": 49
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
              "Column": 50,
              "Offset": 49
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
                  "Column": 41,
                  "Offset": 40
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
                      "Column": 41,
                      "Offset": 40
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
                          "Column": 40,
                          "Offset": 39
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
                              "Column": 40,
                              "Offset": 39
                            },
                            "children": [
                              {
                                "type": "parameterized_type",
                                "start": {
                                  "Line": 1,
                                  "Column": 13,
                                  "Offset": 12
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 40,
                                  "Offset": 39
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
                                      "Column": 21,
                                      "Offset": 20
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
                                          "Column": 21,
                                          "Offset": 20
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
                                      "Column": 21,
                                      "Offset": 20
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 22,
                                      "Offset": 21
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 1,
                                          "Column": 21,
                                          "Offset": 20
                                        },
                                        "end": {
                                          "Line": 1,
                                          "Column": 22,
                                          "Offset": 21
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
                                      "Column": 22,
                                      "Offset": 21
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 39,
                                      "Offset": 38
                                    },
                                    "children": [
                                      {
                                        "type": "type_expression",
                                        "start": {
                                          "Line": 1,
                                          "Column": 22,
                                          "Offset": 21
                                        },
                                        "end": {
                                          "Line": 1,
                                          "Column": 39,
                                          "Offset": 38
                                        },
                                        "children": [
                                          {
                                            "type": "parameterized_type",
                                            "start": {
                                              "Line": 1,
                                              "Column": 22,
                                              "Offset": 21
                                            },
                                            "end": {
                                              "Line": 1,
                                              "Column": 39,
                                              "Offset": 38
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
                                                  "Column": 29,
                                                  "Offset": 28
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
                                                      "Column": 29,
                                                      "Offset": 28
                                                    },
                                                    "value": "HashRef",
                                                    "kind": "string"
                                                  }
                                                ]
                                              },
                                              {
                                                "type": "expression_stmt",
                                                "start": {
                                                  "Line": 1,
                                                  "Column": 29,
                                                  "Offset": 28
                                                },
                                                "end": {
                                                  "Line": 1,
                                                  "Column": 30,
                                                  "Offset": 29
                                                },
                                                "children": [
                                                  {
                                                    "type": "literal",
                                                    "start": {
                                                      "Line": 1,
                                                      "Column": 29,
                                                      "Offset": 28
                                                    },
                                                    "end": {
                                                      "Line": 1,
                                                      "Column": 30,
                                                      "Offset": 29
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
                                                  "Column": 30,
                                                  "Offset": 29
                                                },
                                                "end": {
                                                  "Line": 1,
                                                  "Column": 38,
                                                  "Offset": 37
                                                },
                                                "children": [
                                                  {
                                                    "type": "type_expression",
                                                    "start": {
                                                      "Line": 1,
                                                      "Column": 30,
                                                      "Offset": 29
                                                    },
                                                    "end": {
                                                      "Line": 1,
                                                      "Column": 38,
                                                      "Offset": 37
                                                    },
                                                    "children": [
                                                      {
                                                        "type": "union_type",
                                                        "start": {
                                                          "Line": 1,
                                                          "Column": 30,
                                                          "Offset": 29
                                                        },
                                                        "end": {
                                                          "Line": 1,
                                                          "Column": 38,
                                                          "Offset": 37
                                                        },
                                                        "children": [
                                                          {
                                                            "type": "type_expression",
                                                            "start": {
                                                              "Line": 1,
                                                              "Column": 30,
                                                              "Offset": 29
                                                            },
                                                            "end": {
                                                              "Line": 1,
                                                              "Column": 33,
                                                              "Offset": 32
                                                            },
                                                            "children": [
                                                              {
                                                                "type": "expression_stmt",
                                                                "start": {
                                                                  "Line": 1,
                                                                  "Column": 30,
                                                                  "Offset": 29
                                                                },
                                                                "end": {
                                                                  "Line": 1,
                                                                  "Column": 33,
                                                                  "Offset": 32
                                                                },
                                                                "children": [
                                                                  {
                                                                    "type": "literal",
                                                                    "start": {
                                                                      "Line": 1,
                                                                      "Column": 30,
                                                                      "Offset": 29
                                                                    },
                                                                    "end": {
                                                                      "Line": 1,
                                                                      "Column": 33,
                                                                      "Offset": 32
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
                                                              "Column": 33,
                                                              "Offset": 32
                                                            },
                                                            "end": {
                                                              "Line": 1,
                                                              "Column": 34,
                                                              "Offset": 33
                                                            },
                                                            "children": [
                                                              {
                                                                "type": "literal",
                                                                "start": {
                                                                  "Line": 1,
                                                                  "Column": 33,
                                                                  "Offset": 32
                                                                },
                                                                "end": {
                                                                  "Line": 1,
                                                                  "Column": 34,
                                                                  "Offset": 33
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
                                                              "Column": 34,
                                                              "Offset": 33
                                                            },
                                                            "end": {
                                                              "Line": 1,
                                                              "Column": 38,
                                                              "Offset": 37
                                                            },
                                                            "children": [
                                                              {
                                                                "type": "expression_stmt",
                                                                "start": {
                                                                  "Line": 1,
                                                                  "Column": 34,
                                                                  "Offset": 33
                                                                },
                                                                "end": {
                                                                  "Line": 1,
                                                                  "Column": 38,
                                                                  "Offset": 37
                                                                },
                                                                "children": [
                                                                  {
                                                                    "type": "literal",
                                                                    "start": {
                                                                      "Line": 1,
                                                                      "Column": 34,
                                                                      "Offset": 33
                                                                    },
                                                                    "end": {
                                                                      "Line": 1,
                                                                      "Column": 38,
                                                                      "Offset": 37
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
                                                  "Line": 1,
                                                  "Column": 38,
                                                  "Offset": 37
                                                },
                                                "end": {
                                                  "Line": 1,
                                                  "Column": 39,
                                                  "Offset": 38
                                                },
                                                "children": [
                                                  {
                                                    "type": "literal",
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
                                                    "value": "]",
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
                                      "Line": 1,
                                      "Column": 39,
                                      "Offset": 38
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 40,
                                      "Offset": 39
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 1,
                                          "Column": 39,
                                          "Offset": 38
                                        },
                                        "end": {
                                          "Line": 1,
                                          "Column": 40,
                                          "Offset": 39
                                        },
                                        "value": "]",
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
                          "Line": 1,
                          "Column": 40,
                          "Offset": 39
                        },
                        "end": {
                          "Line": 1,
                          "Column": 41,
                          "Offset": 40
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 40,
                              "Offset": 39
                            },
                            "end": {
                              "Line": 1,
                              "Column": 41,
                              "Offset": 40
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
                  "Line": 1,
                  "Column": 42,
                  "Offset": 41
                },
                "end": {
                  "Line": 1,
                  "Column": 50,
                  "Offset": 49
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 42,
                      "Offset": 41
                    },
                    "end": {
                      "Line": 1,
                      "Column": 43,
                      "Offset": 42
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 43,
                      "Offset": 42
                    },
                    "end": {
                      "Line": 1,
                      "Column": 50,
                      "Offset": 49
                    },
                    "text": "complex"
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
          "Column": 50,
          "Offset": 49
        },
        "end": {
          "Line": 1,
          "Column": 51,
          "Offset": 50
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 51
        },
        "end": {
          "Line": 2,
          "Column": 53,
          "Offset": 103
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 51
            },
            "end": {
              "Line": 2,
              "Column": 53,
              "Offset": 103
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 51
                },
                "end": {
                  "Line": 2,
                  "Column": 3,
                  "Offset": 53
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 54
                },
                "end": {
                  "Line": 2,
                  "Column": 44,
                  "Offset": 94
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 2,
                      "Column": 4,
                      "Offset": 54
                    },
                    "end": {
                      "Line": 2,
                      "Column": 44,
                      "Offset": 94
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 4,
                          "Offset": 54
                        },
                        "end": {
                          "Line": 2,
                          "Column": 13,
                          "Offset": 63
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 4,
                              "Offset": 54
                            },
                            "end": {
                              "Line": 2,
                              "Column": 13,
                              "Offset": 63
                            },
                            "value": "Container",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 13,
                          "Offset": 63
                        },
                        "end": {
                          "Line": 2,
                          "Column": 14,
                          "Offset": 64
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 13,
                              "Offset": 63
                            },
                            "end": {
                              "Line": 2,
                              "Column": 14,
                              "Offset": 64
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
                          "Column": 14,
                          "Offset": 64
                        },
                        "end": {
                          "Line": 2,
                          "Column": 43,
                          "Offset": 93
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 14,
                              "Offset": 64
                            },
                            "end": {
                              "Line": 2,
                              "Column": 43,
                              "Offset": 93
                            },
                            "children": [
                              {
                                "type": "union_type",
                                "start": {
                                  "Line": 2,
                                  "Column": 14,
                                  "Offset": 64
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 43,
                                  "Offset": 93
                                },
                                "children": [
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 2,
                                      "Column": 14,
                                      "Offset": 64
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 30,
                                      "Offset": 80
                                    },
                                    "children": [
                                      {
                                        "type": "parameterized_type",
                                        "start": {
                                          "Line": 2,
                                          "Column": 14,
                                          "Offset": 64
                                        },
                                        "end": {
                                          "Line": 2,
                                          "Column": 30,
                                          "Offset": 80
                                        },
                                        "children": [
                                          {
                                            "type": "expression_stmt",
                                            "start": {
                                              "Line": 2,
                                              "Column": 14,
                                              "Offset": 64
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 22,
                                              "Offset": 72
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 2,
                                                  "Column": 14,
                                                  "Offset": 64
                                                },
                                                "end": {
                                                  "Line": 2,
                                                  "Column": 22,
                                                  "Offset": 72
                                                },
                                                "value": "ArrayRef",
                                                "kind": "string"
                                              }
                                            ]
                                          },
                                          {
                                            "type": "expression_stmt",
                                            "start": {
                                              "Line": 2,
                                              "Column": 22,
                                              "Offset": 72
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 23,
                                              "Offset": 73
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 2,
                                                  "Column": 22,
                                                  "Offset": 72
                                                },
                                                "end": {
                                                  "Line": 2,
                                                  "Column": 23,
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
                                              "Line": 2,
                                              "Column": 23,
                                              "Offset": 73
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 29,
                                              "Offset": 79
                                            },
                                            "children": [
                                              {
                                                "type": "type_expression",
                                                "start": {
                                                  "Line": 2,
                                                  "Column": 23,
                                                  "Offset": 73
                                                },
                                                "end": {
                                                  "Line": 2,
                                                  "Column": 29,
                                                  "Offset": 79
                                                },
                                                "children": [
                                                  {
                                                    "type": "expression_stmt",
                                                    "start": {
                                                      "Line": 2,
                                                      "Column": 23,
                                                      "Offset": 73
                                                    },
                                                    "end": {
                                                      "Line": 2,
                                                      "Column": 29,
                                                      "Offset": 79
                                                    },
                                                    "children": [
                                                      {
                                                        "type": "literal",
                                                        "start": {
                                                          "Line": 2,
                                                          "Column": 23,
                                                          "Offset": 73
                                                        },
                                                        "end": {
                                                          "Line": 2,
                                                          "Column": 29,
                                                          "Offset": 79
                                                        },
                                                        "value": "MyType",
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
                                              "Column": 29,
                                              "Offset": 79
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 30,
                                              "Offset": 80
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 2,
                                                  "Column": 29,
                                                  "Offset": 79
                                                },
                                                "end": {
                                                  "Line": 2,
                                                  "Column": 30,
                                                  "Offset": 80
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
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 2,
                                      "Column": 30,
                                      "Offset": 80
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 31,
                                      "Offset": 81
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 2,
                                          "Column": 30,
                                          "Offset": 80
                                        },
                                        "end": {
                                          "Line": 2,
                                          "Column": 31,
                                          "Offset": 81
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
                                      "Column": 31,
                                      "Offset": 81
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 43,
                                      "Offset": 93
                                    },
                                    "children": [
                                      {
                                        "type": "parameterized_type",
                                        "start": {
                                          "Line": 2,
                                          "Column": 31,
                                          "Offset": 81
                                        },
                                        "end": {
                                          "Line": 2,
                                          "Column": 43,
                                          "Offset": 93
                                        },
                                        "children": [
                                          {
                                            "type": "expression_stmt",
                                            "start": {
                                              "Line": 2,
                                              "Column": 31,
                                              "Offset": 81
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 38,
                                              "Offset": 88
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 2,
                                                  "Column": 31,
                                                  "Offset": 81
                                                },
                                                "end": {
                                                  "Line": 2,
                                                  "Column": 38,
                                                  "Offset": 88
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
                                              "Column": 38,
                                              "Offset": 88
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 39,
                                              "Offset": 89
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 2,
                                                  "Column": 38,
                                                  "Offset": 88
                                                },
                                                "end": {
                                                  "Line": 2,
                                                  "Column": 39,
                                                  "Offset": 89
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
                                              "Column": 39,
                                              "Offset": 89
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 42,
                                              "Offset": 92
                                            },
                                            "children": [
                                              {
                                                "type": "type_expression",
                                                "start": {
                                                  "Line": 2,
                                                  "Column": 39,
                                                  "Offset": 89
                                                },
                                                "end": {
                                                  "Line": 2,
                                                  "Column": 42,
                                                  "Offset": 92
                                                },
                                                "children": [
                                                  {
                                                    "type": "expression_stmt",
                                                    "start": {
                                                      "Line": 2,
                                                      "Column": 39,
                                                      "Offset": 89
                                                    },
                                                    "end": {
                                                      "Line": 2,
                                                      "Column": 42,
                                                      "Offset": 92
                                                    },
                                                    "children": [
                                                      {
                                                        "type": "literal",
                                                        "start": {
                                                          "Line": 2,
                                                          "Column": 39,
                                                          "Offset": 89
                                                        },
                                                        "end": {
                                                          "Line": 2,
                                                          "Column": 42,
                                                          "Offset": 92
                                                        },
                                                        "value": "OtherType",
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
                                              "Column": 42,
                                              "Offset": 92
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 43,
                                              "Offset": 93
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 2,
                                                  "Column": 42,
                                                  "Offset": 92
                                                },
                                                "end": {
                                                  "Line": 2,
                                                  "Column": 43,
                                                  "Offset": 93
                                                },
                                                "value": "]",
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
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 43,
                          "Offset": 93
                        },
                        "end": {
                          "Line": 2,
                          "Column": 44,
                          "Offset": 94
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 43,
                              "Offset": 93
                            },
                            "end": {
                              "Line": 2,
                              "Column": 44,
                              "Offset": 94
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
                  "Column": 45,
                  "Offset": 95
                },
                "end": {
                  "Line": 2,
                  "Column": 53,
                  "Offset": 103
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 45,
                      "Offset": 95
                    },
                    "end": {
                      "Line": 2,
                      "Column": 46,
                      "Offset": 96
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 46,
                      "Offset": 96
                    },
                    "end": {
                      "Line": 2,
                      "Column": 53,
                      "Offset": 103
                    },
                    "text": "flexible"
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
          "Column": 53,
          "Offset": 103
        },
        "end": {
          "Line": 2,
          "Column": 54,
          "Offset": 104
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 105
        },
        "end": {
          "Line": 3,
          "Column": 65,
          "Offset": 169
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 105
            },
            "end": {
              "Line": 3,
              "Column": 65,
              "Offset": 169
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 105
                },
                "end": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 107
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 108
                },
                "end": {
                  "Line": 3,
                  "Column": 51,
                  "Offset": 155
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 108
                    },
                    "end": {
                      "Line": 3,
                      "Column": 51,
                      "Offset": 155
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 4,
                          "Offset": 108
                        },
                        "end": {
                          "Line": 3,
                          "Column": 10,
                          "Offset": 114
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 4,
                              "Offset": 108
                            },
                            "end": {
                              "Line": 3,
                              "Column": 10,
                              "Offset": 114
                            },
                            "value": "Result",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 10,
                          "Offset": 114
                        },
                        "end": {
                          "Line": 3,
                          "Column": 11,
                          "Offset": 115
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 10,
                              "Offset": 114
                            },
                            "end": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 115
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
                          "Column": 11,
                          "Offset": 115
                        },
                        "end": {
                          "Line": 3,
                          "Column": 50,
                          "Offset": 154
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 115
                            },
                            "end": {
                              "Line": 3,
                              "Column": 23,
                              "Offset": 127
                            },
                            "children": [
                              {
                                "type": "parameterized_type",
                                "start": {
                                  "Line": 3,
                                  "Column": 11,
                                  "Offset": 115
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 23,
                                  "Offset": 127
                                },
                                "children": [
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 3,
                                      "Column": 11,
                                      "Offset": 115
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 15,
                                      "Offset": 119
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 11,
                                          "Offset": 115
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 15,
                                          "Offset": 119
                                        },
                                        "value": "Data",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 3,
                                      "Column": 15,
                                      "Offset": 119
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 16,
                                      "Offset": 120
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 15,
                                          "Offset": 119
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 16,
                                          "Offset": 120
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
                                      "Column": 16,
                                      "Offset": 120
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 22,
                                      "Offset": 126
                                    },
                                    "children": [
                                      {
                                        "type": "type_expression",
                                        "start": {
                                          "Line": 3,
                                          "Column": 16,
                                          "Offset": 120
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 22,
                                          "Offset": 126
                                        },
                                        "children": [
                                          {
                                            "type": "expression_stmt",
                                            "start": {
                                              "Line": 3,
                                              "Column": 16,
                                              "Offset": 120
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 22,
                                              "Offset": 126
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 3,
                                                  "Column": 16,
                                                  "Offset": 120
                                                },
                                                "end": {
                                                  "Line": 3,
                                                  "Column": 22,
                                                  "Offset": 126
                                                },
                                                "value": "UserInfo",
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
                                      "Column": 22,
                                      "Offset": 126
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 23,
                                      "Offset": 127
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 22,
                                          "Offset": 126
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 23,
                                          "Offset": 127
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
                            "type": "expression_stmt",
                            "start": {
                              "Line": 3,
                              "Column": 23,
                              "Offset": 127
                            },
                            "end": {
                              "Line": 3,
                              "Column": 24,
                              "Offset": 128
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 3,
                                  "Column": 23,
                                  "Offset": 127
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 24,
                                  "Offset": 128
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
                              "Column": 25,
                              "Offset": 129
                            },
                            "end": {
                              "Line": 3,
                              "Column": 50,
                              "Offset": 154
                            },
                            "children": [
                              {
                                "type": "parameterized_type",
                                "start": {
                                  "Line": 3,
                                  "Column": 25,
                                  "Offset": 129
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 50,
                                  "Offset": 154
                                },
                                "children": [
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 3,
                                      "Column": 25,
                                      "Offset": 129
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 30,
                                      "Offset": 134
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 25,
                                          "Offset": 129
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 30,
                                          "Offset": 134
                                        },
                                        "value": "Error",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 3,
                                      "Column": 30,
                                      "Offset": 134
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 31,
                                      "Offset": 135
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 30,
                                          "Offset": 134
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 31,
                                          "Offset": 135
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
                                      "Column": 31,
                                      "Offset": 135
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 49,
                                      "Offset": 153
                                    },
                                    "children": [
                                      {
                                        "type": "type_expression",
                                        "start": {
                                          "Line": 3,
                                          "Column": 31,
                                          "Offset": 135
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 49,
                                          "Offset": 153
                                        },
                                        "children": [
                                          {
                                            "type": "expression_stmt",
                                            "start": {
                                              "Line": 3,
                                              "Column": 31,
                                              "Offset": 135
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 49,
                                              "Offset": 153
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 3,
                                                  "Column": 31,
                                                  "Offset": 135
                                                },
                                                "end": {
                                                  "Line": 3,
                                                  "Column": 49,
                                                  "Offset": 153
                                                },
                                                "value": "ValidationFailure",
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
                                      "Column": 49,
                                      "Offset": 153
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 50,
                                      "Offset": 154
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 49,
                                          "Offset": 153
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 50,
                                          "Offset": 154
                                        },
                                        "value": "]",
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
                          "Column": 50,
                          "Offset": 154
                        },
                        "end": {
                          "Line": 3,
                          "Column": 51,
                          "Offset": 155
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 50,
                              "Offset": 154
                            },
                            "end": {
                              "Line": 3,
                              "Column": 51,
                              "Offset": 155
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
                  "Column": 52,
                  "Offset": 156
                },
                "end": {
                  "Line": 3,
                  "Column": 65,
                  "Offset": 169
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 52,
                      "Offset": 156
                    },
                    "end": {
                      "Line": 3,
                      "Column": 53,
                      "Offset": 157
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 53,
                      "Offset": 157
                    },
                    "end": {
                      "Line": 3,
                      "Column": 65,
                      "Offset": 169
                    },
                    "text": "nested_result"
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
          "Column": 65,
          "Offset": 169
        },
        "end": {
          "Line": 3,
          "Column": 66,
          "Offset": 170
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
my $complex;
my $flexible;
my $nested_result;
```

## Typed Perl Output

```perl
my Map[Str, ArrayRef[HashRef[Int|Bool]]] $complex;
my Container[ArrayRef[MyType]|HashRef[OtherType]] $flexible;
my Result[Data[UserInfo], Error[ValidationFailure]] $nested_result;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
