---
category: untyped-perl
subcategory: expressions
tags:
    - comparison
    - literals
    - mixed
    - logical
---

# Comparison With Literals

Comparisons with literal values

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

# Expected Compilation Outcomes

## Comparison With Literals

### Clean Perl Output

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

### Typed Perl Output

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/comparison_with_literals.pl
  Source length: 53 characters
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
          binary_expression
            relational_expression
              scalar
                token
                token
              expression_stmt
                literal
              token
            expression_stmt
              literal
            relational_expression
              scalar
                token
                token
              expression_stmt
                literal
              token
          expression_stmt
            literal
          equality_expression
            scalar
              token
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
  "path": "/tmp/comparison_with_literals.pl",
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
      "Offset": 53
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
          "Column": 52,
          "Offset": 51
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
              "Column": 52,
              "Offset": 51
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
                    "text": "check"
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
                  "Column": 52,
                  "Offset": 51
                },
                "children": [
                  {
                    "type": "binary_expression",
                    "start": {
                      "Line": 1,
                      "Column": 10,
                      "Offset": 9
                    },
                    "end": {
                      "Line": 1,
                      "Column": 37,
                      "Offset": 36
                    },
                    "children": [
                      {
                        "type": "relational_expression",
                        "start": {
                          "Line": 1,
                          "Column": 10,
                          "Offset": 9
                        },
                        "end": {
                          "Line": 1,
                          "Column": 20,
                          "Offset": 19
                        },
                        "children": [
                          {
                            "type": "scalar",
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
                                "text": "value"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
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
                                  "Column": 18,
                                  "Offset": 17
                                },
                                "value": ">",
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
                              "Column": 20,
                              "Offset": 19
                            },
                            "text": "0"
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
                          "Column": 23,
                          "Offset": 22
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
                              "Column": 23,
                              "Offset": 22
                            },
                            "value": "&&",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "relational_expression",
                        "start": {
                          "Line": 1,
                          "Column": 24,
                          "Offset": 23
                        },
                        "end": {
                          "Line": 1,
                          "Column": 37,
                          "Offset": 36
                        },
                        "children": [
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 24,
                              "Offset": 23
                            },
                            "end": {
                              "Line": 1,
                              "Column": 30,
                              "Offset": 29
                            },
                            "children": [
                              {
                                "type": "token",
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
                                "text": "$"
                              },
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 25,
                                  "Offset": 24
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 30,
                                  "Offset": 29
                                },
                                "text": "count"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 31,
                              "Offset": 30
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
                                  "Column": 31,
                                  "Offset": 30
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 33,
                                  "Offset": 32
                                },
                                "value": "<=",
                                "kind": "string"
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
                              "Column": 37,
                              "Offset": 36
                            },
                            "text": "100"
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
                      "Column": 40,
                      "Offset": 39
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
                          "Column": 40,
                          "Offset": 39
                        },
                        "value": "&&",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "equality_expression",
                    "start": {
                      "Line": 1,
                      "Column": 41,
                      "Offset": 40
                    },
                    "end": {
                      "Line": 1,
                      "Column": 52,
                      "Offset": 51
                    },
                    "children": [
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 41,
                          "Offset": 40
                        },
                        "end": {
                          "Line": 1,
                          "Column": 46,
                          "Offset": 45
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
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 42,
                              "Offset": 41
                            },
                            "end": {
                              "Line": 1,
                              "Column": 46,
                              "Offset": 45
                            },
                            "text": "name"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 47,
                          "Offset": 46
                        },
                        "end": {
                          "Line": 1,
                          "Column": 49,
                          "Offset": 48
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 47,
                              "Offset": 46
                            },
                            "end": {
                              "Line": 1,
                              "Column": 49,
                              "Offset": 48
                            },
                            "value": "ne",
                            "kind": "string"
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
                          "Column": 52,
                          "Offset": 51
                        },
                        "text": "''"
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
          "Column": 52,
          "Offset": 51
        },
        "end": {
          "Line": 1,
          "Column": 53,
          "Offset": 52
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 53
}
```
