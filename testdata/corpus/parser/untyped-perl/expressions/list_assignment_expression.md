---
category: untyped-perl
subcategory: expressions
tags:
    - list
    - assignment
    - function_calls
    - multiple
---

# List Assignment Expression

List assignment with function call

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

## Text AST

```
AST {
  Path: /tmp/list_assignment_expression.pl
  Source length: 48 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
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
          expression_stmt
            literal
          array
            expression_stmt
              literal
            token
        token
        token
        ambiguous_function_call_expression
          expression_stmt
            literal
          list_expression
            match_regexp
              expression_stmt
                literal
              expression_stmt
                literal
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

## JSON AST

```json
{
  "path": "/tmp/list_assignment_expression.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 49,
      "Offset": 48
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
          "Column": 48,
          "Offset": 47
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
              "Column": 48,
              "Offset": 47
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
                "text": "("
              },
              {
                "type": "list_expression",
                "start": {
                  "Line": 1,
                  "Column": 2,
                  "Offset": 1
                },
                "end": {
                  "Line": 1,
                  "Column": 24,
                  "Offset": 23
                },
                "children": [
                  {
                    "type": "scalar",
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
                    "children": [
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 2,
                          "Offset": 1
                        },
                        "end": {
                          "Line": 1,
                          "Column": 3,
                          "Offset": 2
                        },
                        "text": "$"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 3,
                          "Offset": 2
                        },
                        "end": {
                          "Line": 1,
                          "Column": 8,
                          "Offset": 7
                        },
                        "text": "first"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
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
                    "children": [
                      {
                        "type": "literal",
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
                        "value": ",",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 10,
                      "Offset": 9
                    },
                    "end": {
                      "Line": 1,
                      "Column": 17,
                      "Offset": 16
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
                          "Column": 17,
                          "Offset": 16
                        },
                        "text": "second"
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
                        "value": ",",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "array",
                    "start": {
                      "Line": 1,
                      "Column": 19,
                      "Offset": 18
                    },
                    "end": {
                      "Line": 1,
                      "Column": 24,
                      "Offset": 23
                    },
                    "children": [
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
                            "value": "@",
                            "kind": "string"
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
                          "Column": 24,
                          "Offset": 23
                        },
                        "text": "rest"
                      }
                    ]
                  }
                ]
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
                "text": ")"
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
                "text": "="
              },
              {
                "type": "ambiguous_function_call_expression",
                "start": {
                  "Line": 1,
                  "Column": 28,
                  "Offset": 27
                },
                "end": {
                  "Line": 1,
                  "Column": 48,
                  "Offset": 47
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 28,
                      "Offset": 27
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
                          "Column": 28,
                          "Offset": 27
                        },
                        "end": {
                          "Line": 1,
                          "Column": 33,
                          "Offset": 32
                        },
                        "value": "split",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "list_expression",
                    "start": {
                      "Line": 1,
                      "Column": 34,
                      "Offset": 33
                    },
                    "end": {
                      "Line": 1,
                      "Column": 48,
                      "Offset": 47
                    },
                    "children": [
                      {
                        "type": "match_regexp",
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
                        "children": [
                          {
                            "type": "expression_stmt",
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
                                  "Column": 35,
                                  "Offset": 34
                                },
                                "value": "/",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
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
                            "children": [
                              {
                                "type": "literal",
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
                                "value": ",",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 36,
                              "Offset": 35
                            },
                            "end": {
                              "Line": 1,
                              "Column": 37,
                              "Offset": 36
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 1,
                                  "Column": 36,
                                  "Offset": 35
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 37,
                                  "Offset": 36
                                },
                                "value": "/",
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
                            "value": ",",
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
                          "Column": 48,
                          "Offset": 47
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
                              "Column": 48,
                              "Offset": 47
                            },
                            "text": "csv_line"
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
          "Column": 48,
          "Offset": 47
        },
        "end": {
          "Line": 1,
          "Column": 49,
          "Offset": 48
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 48
}
```

# Expected Compilation Outcomes

## List Assignment Expression

### Clean Perl Output

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

### Typed Perl Output

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
