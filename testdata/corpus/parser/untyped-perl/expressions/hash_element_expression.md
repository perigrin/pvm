---
category: untyped-perl
subcategory: expressions
tags:
    - hash
    - key_access
    - arithmetic
    - indexing
---

# Hash Element Expression

Hash element access in arithmetic expression

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

## Text AST

```
AST {
  Path: /tmp/hash_element_expression.pl
  Source length: 64 characters
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
            hash_element_expression
              container_variable
                token
                token
              token
              scalar
                token
                token
              token
            expression_stmt
              literal
            hash_element_expression
              container_variable
                token
                token
              token
              token
              token
          expression_stmt
            literal
          hash_element_expression
            container_variable
              token
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
  "path": "/tmp/hash_element_expression.pl",
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
      "Offset": 64
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
          "Column": 63,
          "Offset": 62
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
              "Column": 63,
              "Offset": 62
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
                    "text": "total"
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
                  "Column": 63,
                  "Offset": 62
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
                      "Column": 45,
                      "Offset": 44
                    },
                    "children": [
                      {
                        "type": "hash_element_expression",
                        "start": {
                          "Line": 1,
                          "Column": 10,
                          "Offset": 9
                        },
                        "end": {
                          "Line": 1,
                          "Column": 21,
                          "Offset": 20
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
                              "Column": 15,
                              "Offset": 14
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
                                  "Column": 15,
                                  "Offset": 14
                                },
                                "text": "hash"
                              }
                            ]
                          },
                          {
                            "type": "token",
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
                            "text": "{"
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
                              "Column": 20,
                              "Offset": 19
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
                                  "Column": 20,
                                  "Offset": 19
                                },
                                "text": "key"
                              }
                            ]
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
                              "Column": 21,
                              "Offset": 20
                            },
                            "text": "}"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
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
                              "Column": 23,
                              "Offset": 22
                            },
                            "value": "*",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "hash_element_expression",
                        "start": {
                          "Line": 1,
                          "Column": 24,
                          "Offset": 23
                        },
                        "end": {
                          "Line": 1,
                          "Column": 45,
                          "Offset": 44
                        },
                        "children": [
                          {
                            "type": "container_variable",
                            "start": {
                              "Line": 1,
                              "Column": 24,
                              "Offset": 23
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
                                  "Column": 31,
                                  "Offset": 30
                                },
                                "text": "config"
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
                            "text": "{"
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
                              "Column": 44,
                              "Offset": 43
                            },
                            "text": "'multiplier'"
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
                            "text": "}"
                          }
                        ]
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
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
                    "children": [
                      {
                        "type": "literal",
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
                        "value": "+",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "hash_element_expression",
                    "start": {
                      "Line": 1,
                      "Column": 48,
                      "Offset": 47
                    },
                    "end": {
                      "Line": 1,
                      "Column": 63,
                      "Offset": 62
                    },
                    "children": [
                      {
                        "type": "container_variable",
                        "start": {
                          "Line": 1,
                          "Column": 48,
                          "Offset": 47
                        },
                        "end": {
                          "Line": 1,
                          "Column": 57,
                          "Offset": 56
                        },
                        "children": [
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 48,
                              "Offset": 47
                            },
                            "end": {
                              "Line": 1,
                              "Column": 49,
                              "Offset": 48
                            },
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 49,
                              "Offset": 48
                            },
                            "end": {
                              "Line": 1,
                              "Column": 57,
                              "Offset": 56
                            },
                            "text": "defaults"
                          }
                        ]
                      },
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
                        "text": "{"
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 58,
                          "Offset": 57
                        },
                        "end": {
                          "Line": 1,
                          "Column": 62,
                          "Offset": 61
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 58,
                              "Offset": 57
                            },
                            "end": {
                              "Line": 1,
                              "Column": 62,
                              "Offset": 61
                            },
                            "value": "rate",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 62,
                          "Offset": 61
                        },
                        "end": {
                          "Line": 1,
                          "Column": 63,
                          "Offset": 62
                        },
                        "text": "}"
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
          "Column": 63,
          "Offset": 62
        },
        "end": {
          "Line": 1,
          "Column": 64,
          "Offset": 63
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 64
}
```

# Expected Compilation Outcomes

## Hash Element Expression

### Clean Perl Output

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

### Typed Perl Output

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
