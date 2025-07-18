---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - chained
    - and
    - multiple
---

# Chained Logical

Chained logical AND operations

```perl
$valid = $a && $b && $c && $d;
```

# Expected Compilation Outcomes

## Chained Logical

### Clean Perl Output

```perl
$valid = $a && $b && $c && $d;
```

### Typed Perl Output

```perl
$valid = $a && $b && $c && $d;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

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
        expression_stmt
          literal
        scalar
          token
          token
  token
```

## JSON AST

```json
{
  "path": "/tmp/test_chained_logical.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 31,
      "Offset": 30
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
          "Column": 30,
          "Offset": 29
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
              "Column": 30,
              "Offset": 29
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
                    "text": "valid"
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
                  "Column": 30,
                  "Offset": 29
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
                      "Column": 24,
                      "Offset": 23
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
                          "Column": 18,
                          "Offset": 17
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
                              "Column": 12,
                              "Offset": 11
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
                                  "Column": 12,
                                  "Offset": 11
                                },
                                "text": "a"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
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
                                  "Column": 13,
                                  "Offset": 12
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 15,
                                  "Offset": 14
                                },
                                "value": "&&",
                                "kind": "string"
                              }
                            ]
                          },
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
                          "Column": 21,
                          "Offset": 20
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
                              "Column": 21,
                              "Offset": 20
                            },
                            "value": "&&",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 22,
                          "Offset": 21
                        },
                        "end": {
                          "Line": 1,
                          "Column": 24,
                          "Offset": 23
                        },
                        "children": [
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
                            "text": "$"
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
                            "text": "c"
                          }
                        ]
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 25,
                      "Offset": 24
                    },
                    "end": {
                      "Line": 1,
                      "Column": 27,
                      "Offset": 26
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
                          "Column": 27,
                          "Offset": 26
                        },
                        "value": "&&",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 28,
                      "Offset": 27
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
                          "Column": 28,
                          "Offset": 27
                        },
                        "end": {
                          "Line": 1,
                          "Column": 29,
                          "Offset": 28
                        },
                        "text": "$"
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
                          "Column": 30,
                          "Offset": 29
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
          "Column": 30,
          "Offset": 29
        },
        "end": {
          "Line": 1,
          "Column": 31,
          "Offset": 30
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 30
}
```
