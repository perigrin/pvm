---
category: untyped-perl
subcategory: expressions
tags:
    - bit_manipulation
    - bitwise
    - flag_clearing
    - complex
---

# Bit Clearing

Clearing a specific bit using NOT and AND

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

# Expected Compilation Outcomes

## Bit Clearing

### Clean Perl Output

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
AST {
  Path: /tmp/bit_clearing.pl
  Source length: 44 characters
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
          scalar
            token
            token
          expression_stmt
            literal
          unary_expression
            expression_stmt
              literal
            token
            binary_expression
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
  "path": "/tmp/bit_clearing.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 45,
      "Offset": 44
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
          "Column": 44,
          "Offset": 43
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
              "Column": 44,
              "Offset": 43
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
                  "Column": 13,
                  "Offset": 12
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
                      "Column": 13,
                      "Offset": 12
                    },
                    "text": "bit_cleared"
                  }
                ]
              },
              {
                "type": "token",
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
                "text": "="
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
                  "Column": 44,
                  "Offset": 43
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
                      "Column": 22,
                      "Offset": 21
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
                          "Column": 22,
                          "Offset": 21
                        },
                        "text": "flags"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 23,
                      "Offset": 22
                    },
                    "end": {
                      "Line": 1,
                      "Column": 24,
                      "Offset": 23
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 23,
                          "Offset": 22
                        },
                        "end": {
                          "Line": 1,
                          "Column": 24,
                          "Offset": 23
                        },
                        "value": "&",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "unary_expression",
                    "start": {
                      "Line": 1,
                      "Column": 25,
                      "Offset": 24
                    },
                    "end": {
                      "Line": 1,
                      "Column": 44,
                      "Offset": 43
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 25,
                          "Offset": 24
                        },
                        "end": {
                          "Line": 1,
                          "Column": 26,
                          "Offset": 25
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 25,
                              "Offset": 24
                            },
                            "end": {
                              "Line": 1,
                              "Column": 26,
                              "Offset": 25
                            },
                            "value": "~",
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
                        "text": "("
                      },
                      {
                        "type": "binary_expression",
                        "start": {
                          "Line": 1,
                          "Column": 27,
                          "Offset": 26
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
                              "Column": 27,
                              "Offset": 26
                            },
                            "end": {
                              "Line": 1,
                              "Column": 28,
                              "Offset": 27
                            },
                            "text": "1"
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
                              "Column": 31,
                              "Offset": 30
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
                                  "Column": 31,
                                  "Offset": 30
                                },
                                "value": "<<",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 32,
                              "Offset": 31
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
                                  "Column": 32,
                                  "Offset": 31
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 33,
                                  "Offset": 32
                                },
                                "text": "$"
                              },
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 33,
                                  "Offset": 32
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 43,
                                  "Offset": 42
                                },
                                "text": "bit_number"
                              }
                            ]
                          }
                        ]
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
                          "Column": 44,
                          "Offset": 43
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
          "Column": 44,
          "Offset": 43
        },
        "end": {
          "Line": 1,
          "Column": 45,
          "Offset": 44
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 44
}
```
