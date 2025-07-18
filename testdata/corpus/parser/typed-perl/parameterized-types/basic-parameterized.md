---
category: typed-perl
subcategory: parameterized-types
tags:
    - basic-parameters
    - ArrayRef
    - HashRef
    - CodeRef
    - parameterized-types
type_check: true
---

# Basic Parameterized

Basic parameterized types with single type parameters

```perl
my ArrayRef[Int] @numbers;
my HashRef[Str] %strings;
my CodeRef[Int, Str] $function;
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 84 characters
  Type Annotations:
    VarAnnotation: @numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: %strings :: HashRef[Str] at 2:1
    VarAnnotation: $function :: CodeRef[Int, Str] at 3:1
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
                expression_stmt
                  literal
            expression_stmt
              literal
        hash
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

### JSON Format

```json
{
  "path": "/tmp/basic_parameterized.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 32,
      "Offset": 84
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
                          "Column": 16,
                          "Offset": 15
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
                "type": "array",
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
                        "value": "@",
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
                    "text": "numbers"
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
          "Column": 25,
          "Offset": 51
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
              "Column": 25,
              "Offset": 51
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
                  "Column": 16,
                  "Offset": 42
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
                      "Column": 16,
                      "Offset": 42
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
                          "Column": 11,
                          "Offset": 37
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
                              "Column": 11,
                              "Offset": 37
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
                          "Column": 11,
                          "Offset": 37
                        },
                        "end": {
                          "Line": 2,
                          "Column": 12,
                          "Offset": 38
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 11,
                              "Offset": 37
                            },
                            "end": {
                              "Line": 2,
                              "Column": 12,
                              "Offset": 38
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
                          "Column": 12,
                          "Offset": 38
                        },
                        "end": {
                          "Line": 2,
                          "Column": 15,
                          "Offset": 41
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 12,
                              "Offset": 38
                            },
                            "end": {
                              "Line": 2,
                              "Column": 15,
                              "Offset": 41
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 2,
                                  "Column": 12,
                                  "Offset": 38
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 15,
                                  "Offset": 41
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 2,
                                      "Column": 12,
                                      "Offset": 38
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 15,
                                      "Offset": 41
                                    },
                                    "value": "Str",
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
                          "Column": 15,
                          "Offset": 41
                        },
                        "end": {
                          "Line": 2,
                          "Column": 16,
                          "Offset": 42
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
                              "Column": 16,
                              "Offset": 42
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
                  "Line": 2,
                  "Column": 17,
                  "Offset": 43
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
                      "Column": 17,
                      "Offset": 43
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
                          "Column": 17,
                          "Offset": 43
                        },
                        "end": {
                          "Line": 2,
                          "Column": 18,
                          "Offset": 44
                        },
                        "value": "%",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 18,
                      "Offset": 44
                    },
                    "end": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 51
                    },
                    "text": "strings"
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
          "Column": 25,
          "Offset": 51
        },
        "end": {
          "Line": 2,
          "Column": 26,
          "Offset": 52
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 53
        },
        "end": {
          "Line": 3,
          "Column": 31,
          "Offset": 83
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 53
            },
            "end": {
              "Line": 3,
              "Column": 31,
              "Offset": 83
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 53
                },
                "end": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 55
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 56
                },
                "end": {
                  "Line": 3,
                  "Column": 21,
                  "Offset": 73
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 56
                    },
                    "end": {
                      "Line": 3,
                      "Column": 21,
                      "Offset": 73
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 4,
                          "Offset": 56
                        },
                        "end": {
                          "Line": 3,
                          "Column": 11,
                          "Offset": 63
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 4,
                              "Offset": 56
                            },
                            "end": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 63
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
                          "Offset": 63
                        },
                        "end": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 64
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 63
                            },
                            "end": {
                              "Line": 3,
                              "Column": 12,
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
                          "Line": 3,
                          "Column": 12,
                          "Offset": 64
                        },
                        "end": {
                          "Line": 3,
                          "Column": 20,
                          "Offset": 72
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 12,
                              "Offset": 64
                            },
                            "end": {
                              "Line": 3,
                              "Column": 15,
                              "Offset": 67
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 3,
                                  "Column": 12,
                                  "Offset": 64
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 15,
                                  "Offset": 67
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 3,
                                      "Column": 12,
                                      "Offset": 64
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 15,
                                      "Offset": 67
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
                              "Offset": 67
                            },
                            "end": {
                              "Line": 3,
                              "Column": 16,
                              "Offset": 68
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 3,
                                  "Column": 15,
                                  "Offset": 67
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 16,
                                  "Offset": 68
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
                              "Column": 17,
                              "Offset": 69
                            },
                            "end": {
                              "Line": 3,
                              "Column": 20,
                              "Offset": 72
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 3,
                                  "Column": 17,
                                  "Offset": 69
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 20,
                                  "Offset": 72
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 3,
                                      "Column": 17,
                                      "Offset": 69
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 20,
                                      "Offset": 72
                                    },
                                    "value": "Str",
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
                          "Column": 20,
                          "Offset": 72
                        },
                        "end": {
                          "Line": 3,
                          "Column": 21,
                          "Offset": 73
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 20,
                              "Offset": 72
                            },
                            "end": {
                              "Line": 3,
                              "Column": 21,
                              "Offset": 73
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
                  "Column": 22,
                  "Offset": 74
                },
                "end": {
                  "Line": 3,
                  "Column": 31,
                  "Offset": 83
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 22,
                      "Offset": 74
                    },
                    "end": {
                      "Line": 3,
                      "Column": 23,
                      "Offset": 75
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 23,
                      "Offset": 75
                    },
                    "end": {
                      "Line": 3,
                      "Column": 31,
                      "Offset": 83
                    },
                    "text": "function"
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
          "Column": 31,
          "Offset": 83
        },
        "end": {
          "Line": 3,
          "Column": 32,
          "Offset": 84
        },
        "text": ";"
      }
    ]
  }
}
```

## After Type Inference

### Text Format

```
AST {
  Path:
  Source length: 84 characters
  Type Annotations:
    VarAnnotation: @numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: %strings :: HashRef[Str] at 2:1
    VarAnnotation: $function :: CodeRef[Int, Str] at 3:1
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
                expression_stmt
                  literal
            expression_stmt
              literal
        hash
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

### JSON Format

```json
{
  "path": "/tmp/basic_parameterized.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 32,
      "Offset": 84
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
                          "Column": 16,
                          "Offset": 15
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
                "type": "array",
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
                        "value": "@",
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
                    "text": "numbers"
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
          "Column": 25,
          "Offset": 51
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
              "Column": 25,
              "Offset": 51
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
                  "Column": 16,
                  "Offset": 42
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
                      "Column": 16,
                      "Offset": 42
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
                          "Column": 11,
                          "Offset": 37
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
                              "Column": 11,
                              "Offset": 37
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
                          "Column": 11,
                          "Offset": 37
                        },
                        "end": {
                          "Line": 2,
                          "Column": 12,
                          "Offset": 38
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 11,
                              "Offset": 37
                            },
                            "end": {
                              "Line": 2,
                              "Column": 12,
                              "Offset": 38
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
                          "Column": 12,
                          "Offset": 38
                        },
                        "end": {
                          "Line": 2,
                          "Column": 15,
                          "Offset": 41
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 12,
                              "Offset": 38
                            },
                            "end": {
                              "Line": 2,
                              "Column": 15,
                              "Offset": 41
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 2,
                                  "Column": 12,
                                  "Offset": 38
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 15,
                                  "Offset": 41
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 2,
                                      "Column": 12,
                                      "Offset": 38
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 15,
                                      "Offset": 41
                                    },
                                    "value": "Str",
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
                          "Column": 15,
                          "Offset": 41
                        },
                        "end": {
                          "Line": 2,
                          "Column": 16,
                          "Offset": 42
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
                              "Column": 16,
                              "Offset": 42
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
                  "Line": 2,
                  "Column": 17,
                  "Offset": 43
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
                      "Column": 17,
                      "Offset": 43
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
                          "Column": 17,
                          "Offset": 43
                        },
                        "end": {
                          "Line": 2,
                          "Column": 18,
                          "Offset": 44
                        },
                        "value": "%",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 18,
                      "Offset": 44
                    },
                    "end": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 51
                    },
                    "text": "strings"
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
          "Column": 25,
          "Offset": 51
        },
        "end": {
          "Line": 2,
          "Column": 26,
          "Offset": 52
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 53
        },
        "end": {
          "Line": 3,
          "Column": 31,
          "Offset": 83
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 53
            },
            "end": {
              "Line": 3,
              "Column": 31,
              "Offset": 83
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 53
                },
                "end": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 55
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 56
                },
                "end": {
                  "Line": 3,
                  "Column": 21,
                  "Offset": 73
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 56
                    },
                    "end": {
                      "Line": 3,
                      "Column": 21,
                      "Offset": 73
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 4,
                          "Offset": 56
                        },
                        "end": {
                          "Line": 3,
                          "Column": 11,
                          "Offset": 63
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 4,
                              "Offset": 56
                            },
                            "end": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 63
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
                          "Offset": 63
                        },
                        "end": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 64
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 63
                            },
                            "end": {
                              "Line": 3,
                              "Column": 12,
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
                          "Line": 3,
                          "Column": 12,
                          "Offset": 64
                        },
                        "end": {
                          "Line": 3,
                          "Column": 20,
                          "Offset": 72
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 12,
                              "Offset": 64
                            },
                            "end": {
                              "Line": 3,
                              "Column": 15,
                              "Offset": 67
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 3,
                                  "Column": 12,
                                  "Offset": 64
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 15,
                                  "Offset": 67
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 3,
                                      "Column": 12,
                                      "Offset": 64
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 15,
                                      "Offset": 67
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
                              "Offset": 67
                            },
                            "end": {
                              "Line": 3,
                              "Column": 16,
                              "Offset": 68
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 3,
                                  "Column": 15,
                                  "Offset": 67
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 16,
                                  "Offset": 68
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
                              "Column": 17,
                              "Offset": 69
                            },
                            "end": {
                              "Line": 3,
                              "Column": 20,
                              "Offset": 72
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 3,
                                  "Column": 17,
                                  "Offset": 69
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 20,
                                  "Offset": 72
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 3,
                                      "Column": 17,
                                      "Offset": 69
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 20,
                                      "Offset": 72
                                    },
                                    "value": "Str",
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
                          "Column": 20,
                          "Offset": 72
                        },
                        "end": {
                          "Line": 3,
                          "Column": 21,
                          "Offset": 73
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 20,
                              "Offset": 72
                            },
                            "end": {
                              "Line": 3,
                              "Column": 21,
                              "Offset": 73
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
                  "Column": 22,
                  "Offset": 74
                },
                "end": {
                  "Line": 3,
                  "Column": 31,
                  "Offset": 83
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 22,
                      "Offset": 74
                    },
                    "end": {
                      "Line": 3,
                      "Column": 23,
                      "Offset": 75
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 23,
                      "Offset": 75
                    },
                    "end": {
                      "Line": 3,
                      "Column": 31,
                      "Offset": 83
                    },
                    "text": "function"
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
          "Column": 31,
          "Offset": 83
        },
        "end": {
          "Line": 3,
          "Column": 32,
          "Offset": 84
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
my @numbers;
my %strings;
my $function;
```

## Typed Perl Output

```perl
my ArrayRef[Int] @numbers;
my HashRef[Str] %strings;
my CodeRef[Int, Str] $function;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
