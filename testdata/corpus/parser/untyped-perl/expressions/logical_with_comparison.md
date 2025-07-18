---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - comparison
    - mixed_operators
    - complex
---

# Logical With Comparison

Logical operators combined with comparison operators

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

# Expected Compilation Outcomes

## Logical With Comparison

### Clean Perl Output

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

### Typed Perl Output

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text Format

```
source_file
  expression_statement
    assignment_expression
      scalar
        token
        token
      token
      binary_expression
        binary_expression
          token
          relational_expression
            scalar
              token
              token
            expression_stmt
              literal
            token
          token
          expression_stmt
            literal
          token
          relational_expression
            scalar
              token
              token
            expression_stmt
              literal
            token
          token
        expression_stmt
          literal
        token
        equality_expression
          scalar
            token
            token
          expression_stmt
            literal
          token
        token
  token
```

## JSON Format

```json
{
  "path": "/tmp/logical_with_comparison.pl",
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
      "Offset": 47
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
          "Column": 46,
          "Offset": 45
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
              "Column": 46,
              "Offset": 45
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
                  "Column": 46,
                  "Offset": 45
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
                      "Column": 33,
                      "Offset": 32
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
                        "text": "("
                      },
                      {
                        "type": "relational_expression",
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
                        "children": [
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 12,
                              "Offset": 11
                            },
                            "end": {
                              "Line": 1,
                              "Column": 14,
                              "Offset": 13
                            },
                            "children": [
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
                                "text": "$"
                              },
                              {
                                "type": "token",
                                "start": {
                                  "Line": 1,
                                  "Column": 13,
                                  "Offset": 12
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 14,
                                  "Offset": 13
                                },
                                "text": "a"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 15,
                              "Offset": 14
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
                                  "Column": 15,
                                  "Offset": 14
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 16,
                                  "Offset": 15
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
                              "Column": 17,
                              "Offset": 16
                            },
                            "end": {
                              "Line": 1,
                              "Column": 18,
                              "Offset": 17
                            },
                            "text": "0"
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
                          "Column": 19,
                          "Offset": 18
                        },
                        "text": ")"
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
                          "Column": 22,
                          "Offset": 21
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
                              "Column": 22,
                              "Offset": 21
                            },
                            "value": "&&",
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
                          "Column": 24,
                          "Offset": 23
                        },
                        "text": "("
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
                          "Column": 32,
                          "Offset": 31
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
                              "Column": 26,
                              "Offset": 25
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
                                  "Column": 26,
                                  "Offset": 25
                                },
                                "text": "b"
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
                                "value": "<",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 29,
                              "Offset": 28
                            },
                            "end": {
                              "Line": 1,
                              "Column": 32,
                              "Offset": 31
                            },
                            "text": "100"
                          }
                        ]
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
                          "Column": 33,
                          "Offset": 32
                        },
                        "text": ")"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 34,
                      "Offset": 33
                    },
                    "end": {
                      "Line": 1,
                      "Column": 36,
                      "Offset": 35
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
                          "Column": 36,
                          "Offset": 35
                        },
                        "value": "||",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 37,
                      "Offset": 36
                    },
                    "end": {
                      "Line": 1,
                      "Column": 38,
                      "Offset": 37
                    },
                    "text": "("
                  },
                  {
                    "type": "equality_expression",
                    "start": {
                      "Line": 1,
                      "Column": 38,
                      "Offset": 37
                    },
                    "end": {
                      "Line": 1,
                      "Column": 45,
                      "Offset": 44
                    },
                    "children": [
                      {
                        "type": "scalar",
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
                            "type": "token",
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
                            "text": "$"
                          },
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
                            "text": "c"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
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
                        "children": [
                          {
                            "type": "literal",
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
                            "value": "==",
                            "kind": "string"
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
                        "text": "0"
                      }
                    ]
                  },
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
                    "text": ")"
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
          "Column": 46,
          "Offset": 45
        },
        "end": {
          "Line": 1,
          "Column": 47,
          "Offset": 46
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 47
}
```
