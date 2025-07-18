---
category: untyped-perl
subcategory: expressions
tags:
    - bitwise
    - precedence
    - multiple_operators
    - complex
---

# Bitwise Precedence

Bitwise operations with operator precedence

```perl
$result = $a | $b & $c ^ $d;
```

# Expected Compilation Outcomes

## Bitwise Precedence

### Clean Perl Output

```perl
$result = $a | $b & $c ^ $d;
```

### Typed Perl Output

```perl
$result = $a | $b & $c ^ $d;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
AST {
  Path: /tmp/bitwise_precedence.pl
  Source length: 28 characters
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
            scalar
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
              scalar
                token
                token
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
  "path": "/tmp/bitwise_precedence.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 29,
      "Offset": 28
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
          "Column": 28,
          "Offset": 27
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
              "Column": 28,
              "Offset": 27
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
                  "Column": 28,
                  "Offset": 27
                },
                "children": [
                  {
                    "type": "binary_expression",
                    "start": {
                      "Line": 1,
                      "Column": 11,
                      "Offset": 10
                    },
                    "end": {
                      "Line": 1,
                      "Column": 23,
                      "Offset": 22
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
                          "Column": 13,
                          "Offset": 12
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
                              "Column": 13,
                              "Offset": 12
                            },
                            "text": "a"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 14,
                          "Offset": 13
                        },
                        "end": {
                          "Line": 1,
                          "Column": 15,
                          "Offset": 14
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 14,
                              "Offset": 13
                            },
                            "end": {
                              "Line": 1,
                              "Column": 15,
                              "Offset": 14
                            },
                            "value": "|",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "binary_expression",
                        "start": {
                          "Line": 1,
                          "Column": 16,
                          "Offset": 15
                        },
                        "end": {
                          "Line": 1,
                          "Column": 23,
                          "Offset": 22
                        },
                        "children": [
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 16,
                              "Offset": 15
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
                                  "Column": 16,
                                  "Offset": 15
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 17,
                                  "Offset": 16
                                },
                                "text": "$"
                              },
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
                                "text": "b"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
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
                            "children": [
                              {
                                "type": "literal",
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
                                "value": "&",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "scalar",
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
                                "type": "token",
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
                                "text": "$"
                              },
                              {
                                "type": "token",
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
                                "text": "c"
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
                        "value": "^",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 26,
                      "Offset": 25
                    },
                    "end": {
                      "Line": 1,
                      "Column": 28,
                      "Offset": 27
                    },
                    "children": [
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
                        "text": "$"
                      },
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
                        "text": "d"
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
          "Column": 28,
          "Offset": 27
        },
        "end": {
          "Line": 1,
          "Column": 29,
          "Offset": 28
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 28
}
```
