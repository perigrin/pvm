---
category: untyped-perl
subcategory: expressions
tags:
    - array
    - indexing
    - arithmetic
    - multiple_operators
---

# Array Element Expression

Array element access in arithmetic expression

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

# Expected Compilation Outcomes

## Array Element Expression

### Clean Perl Output

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

### Typed Perl Output

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/array_element_expr.pl
  Source length: 51 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
        scalar
          token
          token
        token
        binary_expression
          array_element_expression
            container_variable
              token
              token
            expression_stmt
              literal
            binary_expression
              scalar
                token
                token
              expression_stmt
                literal
              token
            expression_stmt
              literal
          expression_stmt
            literal
          array_element_expression
            array_element_expression
              container_variable
                token
                token
              expression_stmt
                literal
              scalar
                token
                token
              expression_stmt
                literal
            expression_stmt
              literal
            scalar
              token
              token
            expression_stmt
              literal
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/array_element_expr.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 2,
      "Column": 1,
      "Offset": 51
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
            "type": "assignment_expression",
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
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 7,
                  "Offset": 6
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
                      "Column": 2,
                      "Offset": 1
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 2,
                      "Offset": 1
                    },
                    "end": {
                      "Line": 1,
                      "Column": 7,
                      "Offset": 6
                    },
                    "text": "value"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 8,
                  "Offset": 7
                },
                "end": {
                  "Line": 1,
                  "Column": 9,
                  "Offset": 8
                },
                "text": "="
              },
              {
                "type": "binary_expression",
                "start": {
                  "Line": 1,
                  "Column": 10,
                  "Offset": 9
                },
                "end": {
                  "Line": 1,
                  "Column": 50,
                  "Offset": 49
                },
                "children": [
                  {
                    "type": "array_element_expression",
                    "start": {
                      "Line": 1,
                      "Column": 10,
                      "Offset": 9
                    },
                    "end": {
                      "Line": 1,
                      "Column": 28,
                      "Offset": 27
                    },
                    "children": [
                      {
                        "type": "container_variable",
                        "start": {
                          "Line": 1,
                          "Column": 10,
                          "Offset": 9
                        },
                        "end": {
                          "Line": 1,
                          "Column": 16,
                          "Offset": 15
                        },
                        "children": [
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 10,
                              "Offset": 9
                            },
                            "end": {
                              "Line": 1,
                              "Column": 11,
                              "Offset": 10
                            },
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 11,
                              "Offset": 10
                            },
                            "end": {
                              "Line": 1,
                              "Column": 16,
                              "Offset": 15
                            },
                            "text": "array"
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
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "binary_expression",
                        "start": {
                          "Line": 1,
                          "Column": 17,
                          "Offset": 16
                        },
                        "end": {
                          "Line": 1,
                          "Column": 27,
                          "Offset": 26
                        },
                        "children": [
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 17,
                              "Offset": 16
                            },
                            "end": {
                              "Line": 1,
                              "Column": 23,
                              "Offset": 22
                            },
                            "children": [
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 17,
                                  "Offset": 16
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 18,
                                  "Offset": 17
                                },
                                "text": "$"
                              },
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 18,
                                  "Offset": 17
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 23,
                                  "Offset": 22
                                },
                                "text": "index"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 24,
                              "Offset": 23
                            },
                            "end": {
                              "Line": 1,
                              "Column": 25,
                              "Offset": 24
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 1,
                                  "Column": 24,
                                  "Offset": 23
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 25,
                                  "Offset": 24
                                },
                                "value": "+",
                                "kind": "string"
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
                            "text": "1"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 27,
                          "Offset": 26
                        },
                        "end": {
                          "Line": 1,
                          "Column": 28,
                          "Offset": 27
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 27,
                              "Offset": 26
                            },
                            "end": {
                              "Line": 1,
                              "Column": 28,
                              "Offset": 27
                            },
                            "value": "]",
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
                        "value": "+",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "array_element_expression",
                    "start": {
                      "Line": 1,
                      "Column": 31,
                      "Offset": 30
                    },
                    "end": {
                      "Line": 1,
                      "Column": 50,
                      "Offset": 49
                    },
                    "children": [
                      {
                        "type": "array_element_expression",
                        "start": {
                          "Line": 1,
                          "Column": 31,
                          "Offset": 30
                        },
                        "end": {
                          "Line": 1,
                          "Column": 44,
                          "Offset": 43
                        },
                        "children": [
                          {
                            "type": "container_variable",
                            "start": {
                              "Line": 1,
                              "Column": 31,
                              "Offset": 30
                            },
                            "end": {
                              "Line": 1,
                              "Column": 38,
                              "Offset": 37
                            },
                            "children": [
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 31,
                                  "Offset": 30
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 32,
                                  "Offset": 31
                                },
                                "text": "$"
                              },
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 32,
                                  "Offset": 31
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 38,
                                  "Offset": 37
                                },
                                "text": "matrix"
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
                                "value": "[",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 39,
                              "Offset": 38
                            },
                            "end": {
                              "Line": 1,
                              "Column": 43,
                              "Offset": 42
                            },
                            "children": [
                              {
                                "type": "token",
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
                                "text": "$"
                              },
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 40,
                                  "Offset": 39
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 43,
                                  "Offset": 42
                                },
                                "text": "row"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 43,
                              "Offset": 42
                            },
                            "end": {
                              "Line": 1,
                              "Column": 44,
                              "Offset": 43
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 1,
                                  "Column": 43,
                                  "Offset": 42
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 44,
                                  "Offset": 43
                                },
                                "value": "]",
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
                          "Column": 44,
                          "Offset": 43
                        },
                        "end": {
                          "Line": 1,
                          "Column": 45,
                          "Offset": 44
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 44,
                              "Offset": 43
                            },
                            "end": {
                              "Line": 1,
                              "Column": 45,
                              "Offset": 44
                            },
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 45,
                          "Offset": 44
                        },
                        "end": {
                          "Line": 1,
                          "Column": 49,
                          "Offset": 48
                        },
                        "children": [
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 45,
                              "Offset": 44
                            },
                            "end": {
                              "Line": 1,
                              "Column": 46,
                              "Offset": 45
                            },
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 46,
                              "Offset": 45
                            },
                            "end": {
                              "Line": 1,
                              "Column": 49,
                              "Offset": 48
                            },
                            "text": "col"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 49,
                          "Offset": 48
                        },
                        "end": {
                          "Line": 1,
                          "Column": 50,
                          "Offset": 49
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 49,
                              "Offset": 48
                            },
                            "end": {
                              "Line": 1,
                              "Column": 50,
                              "Offset": 49
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
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 51
}
```
