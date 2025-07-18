---
category: untyped-perl
subcategory: expressions
tags:
    - method_calls
    - arithmetic
    - objects
    - function_calls
---

# Method Call Expression

Method calls in arithmetic expression

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

## Text AST

```
AST {
  Path: /tmp/method_call_expression.pl
  Source length: 60 characters
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
          method_call_expression
            scalar
              token
              token
            token
            token
            token
            scalar
              token
              token
            token
          expression_stmt
            literal
          method_call_expression
            scalar
              token
              token
            token
            token
            token
            list_expression
              scalar
                token
                token
              expression_stmt
                literal
              scalar
                token
                token
            token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/method_call_expression.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 61,
      "Offset": 60
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
          "Column": 60,
          "Offset": 59
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
              "Column": 60,
              "Offset": 59
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
                  "Column": 8,
                  "Offset": 7
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
                      "Column": 8,
                      "Offset": 7
                    },
                    "text": "result"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 9,
                  "Offset": 8
                },
                "end": {
                  "Line": 1,
                  "Column": 10,
                  "Offset": 9
                },
                "text": "="
              },
              {
                "type": "binary_expression",
                "start": {
                  "Line": 1,
                  "Column": 11,
                  "Offset": 10
                },
                "end": {
                  "Line": 1,
                  "Column": 60,
                  "Offset": 59
                },
                "children": [
                  {
                    "type": "method_call_expression",
                    "start": {
                      "Line": 1,
                      "Column": 11,
                      "Offset": 10
                    },
                    "end": {
                      "Line": 1,
                      "Column": 32,
                      "Offset": 31
                    },
                    "children": [
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 11,
                          "Offset": 10
                        },
                        "end": {
                          "Line": 1,
                          "Column": 18,
                          "Offset": 17
                        },
                        "children": [
                          {
                            "type": "token",
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
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 12,
                              "Offset": 11
                            },
                            "end": {
                              "Line": 1,
                              "Column": 18,
                              "Offset": 17
                            },
                            "text": "object"
                          }
                        ]
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
                          "Column": 20,
                          "Offset": 19
                        },
                        "text": "->"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 20,
                          "Offset": 19
                        },
                        "end": {
                          "Line": 1,
                          "Column": 26,
                          "Offset": 25
                        },
                        "text": "method"
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
                        "text": "("
                      },
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 27,
                          "Offset": 26
                        },
                        "end": {
                          "Line": 1,
                          "Column": 31,
                          "Offset": 30
                        },
                        "children": [
                          {
                            "type": "token",
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
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 28,
                              "Offset": 27
                            },
                            "end": {
                              "Line": 1,
                              "Column": 31,
                              "Offset": 30
                            },
                            "text": "arg"
                          }
                        ]
                      },
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
                        "text": ")"
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
                        "value": "+",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "method_call_expression",
                    "start": {
                      "Line": 1,
                      "Column": 35,
                      "Offset": 34
                    },
                    "end": {
                      "Line": 1,
                      "Column": 60,
                      "Offset": 59
                    },
                    "children": [
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 35,
                          "Offset": 34
                        },
                        "end": {
                          "Line": 1,
                          "Column": 41,
                          "Offset": 40
                        },
                        "children": [
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 35,
                              "Offset": 34
                            },
                            "end": {
                              "Line": 1,
                              "Column": 36,
                              "Offset": 35
                            },
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 36,
                              "Offset": 35
                            },
                            "end": {
                              "Line": 1,
                              "Column": 41,
                              "Offset": 40
                            },
                            "text": "other"
                          }
                        ]
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 41,
                          "Offset": 40
                        },
                        "end": {
                          "Line": 1,
                          "Column": 43,
                          "Offset": 42
                        },
                        "text": "->"
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
                          "Column": 52,
                          "Offset": 51
                        },
                        "text": "calculate"
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
                        "text": "("
                      },
                      {
                        "type": "list_expression",
                        "start": {
                          "Line": 1,
                          "Column": 53,
                          "Offset": 52
                        },
                        "end": {
                          "Line": 1,
                          "Column": 59,
                          "Offset": 58
                        },
                        "children": [
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 53,
                              "Offset": 52
                            },
                            "end": {
                              "Line": 1,
                              "Column": 55,
                              "Offset": 54
                            },
                            "children": [
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 53,
                                  "Offset": 52
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 54,
                                  "Offset": 53
                                },
                                "text": "$"
                              },
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 54,
                                  "Offset": 53
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 55,
                                  "Offset": 54
                                },
                                "text": "x"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 55,
                              "Offset": 54
                            },
                            "end": {
                              "Line": 1,
                              "Column": 56,
                              "Offset": 55
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 1,
                                  "Column": 55,
                                  "Offset": 54
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 56,
                                  "Offset": 55
                                },
                                "value": ",",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 57,
                              "Offset": 56
                            },
                            "end": {
                              "Line": 1,
                              "Column": 59,
                              "Offset": 58
                            },
                            "children": [
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 57,
                                  "Offset": 56
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 58,
                                  "Offset": 57
                                },
                                "text": "$"
                              },
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 58,
                                  "Offset": 57
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 59,
                                  "Offset": 58
                                },
                                "text": "y"
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 59,
                          "Offset": 58
                        },
                        "end": {
                          "Line": 1,
                          "Column": 60,
                          "Offset": 59
                        },
                        "text": ")"
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
          "Column": 60,
          "Offset": 59
        },
        "end": {
          "Line": 1,
          "Column": 61,
          "Offset": 60
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 60
}
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

### Typed Perl Output

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
