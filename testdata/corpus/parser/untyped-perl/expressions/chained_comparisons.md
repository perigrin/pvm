---
category: untyped-perl
subcategory: expressions
tags:
    - comparison
    - chained
    - ordering
    - logical
---

# Chained Comparisons

Chained comparisons for ordering check

```perl
$ordered = $a < $b && $b < $c && $c < $d;
```

# Expected Compilation Outcomes

## Chained Comparisons

### Clean Perl Output

```perl
$ordered = $a < $b && $b < $c && $c < $d;
```

### Typed Perl Output

```perl
$ordered = $a < $b && $b < $c && $c < $d;
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
          relational_expression
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
          relational_expression
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
        relational_expression
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
  "path": "/tmp/test_chained_comparisons.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 42,
      "Offset": 41
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
          "Column": 41,
          "Offset": 40
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
              "Column": 41,
              "Offset": 40
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
                  "Column": 9,
                  "Offset": 8
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
                      "Column": 9,
                      "Offset": 8
                    },
                    "text": "ordered"
                  }
                ]
              },
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
                "text": "="
              },
              {
                "type": "binary_expression",
                "start": {
                  "Line": 1,
                  "Column": 12,
                  "Offset": 11
                },
                "end": {
                  "Line": 1,
                  "Column": 41,
                  "Offset": 40
                },
                "children": [
                  {
                    "type": "binary_expression",
                    "start": {
                      "Line": 1,
                      "Column": 12,
                      "Offset": 11
                    },
                    "end": {
                      "Line": 1,
                      "Column": 30,
                      "Offset": 29
                    },
                    "children": [
                      {
                        "type": "relational_expression",
                        "start": {
                          "Line": 1,
                          "Column": 12,
                          "Offset": 11
                        },
                        "end": {
                          "Line": 1,
                          "Column": 19,
                          "Offset": 18
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
                                "value": "<",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "scalar",
                            "start": {
                              "Line": 1,
                              "Column": 17,
                              "Offset": 16
                            },
                            "end": {
                              "Line": 1,
                              "Column": 19,
                              "Offset": 18
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
                                  "Column": 19,
                                  "Offset": 18
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
                        "type": "relational_expression",
                        "start": {
                          "Line": 1,
                          "Column": 23,
                          "Offset": 22
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
                              "Column": 23,
                              "Offset": 22
                            },
                            "end": {
                              "Line": 1,
                              "Column": 25,
                              "Offset": 24
                            },
                            "children": [
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
                                "text": "$"
                              },
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
                                "text": "b"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
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
                            "children": [
                              {
                                "type": "literal",
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
                                "value": "<",
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
                        "value": "&&",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "relational_expression",
                    "start": {
                      "Line": 1,
                      "Column": 34,
                      "Offset": 33
                    },
                    "end": {
                      "Line": 1,
                      "Column": 41,
                      "Offset": 40
                    },
                    "children": [
                      {
                        "type": "scalar",
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
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 34,
                              "Offset": 33
                            },
                            "end": {
                              "Line": 1,
                              "Column": 35,
                              "Offset": 34
                            },
                            "text": "$"
                          },
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
                            "text": "c"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
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
                        "children": [
                          {
                            "type": "literal",
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
                            "value": "<",
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
                          "Column": 41,
                          "Offset": 40
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
                              "Column": 41,
                              "Offset": 40
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
          "Column": 42,
          "Offset": 41
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 41
}
```
