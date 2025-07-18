---
category: untyped-perl
subcategory: expressions
tags:
    - regex
    - matching
    - logical
---

# Regex Match Expression

Regular expression matching in logical expression

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

## Text AST

```
AST {
  Path: /tmp/actual_regex.pl
  Source length: 58 characters
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
          token
          binary_expression
            scalar
              token
              token
            expression_stmt
              literal
            match_regexp
              expression_stmt
                literal
              regexp_content
                expression_stmt
                  literal
              expression_stmt
                literal
          token
          expression_stmt
            literal
          token
          binary_expression
            scalar
              token
              token
            expression_stmt
              literal
            match_regexp
              expression_stmt
                literal
              regexp_content
                expression_stmt
                  literal
                expression_stmt
                  literal
                expression_stmt
                  literal
              expression_stmt
                literal
          token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/actual_regex.pl",
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
      "Offset": 58
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
          "Column": 57,
          "Offset": 56
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
              "Column": 57,
              "Offset": 56
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
                  "Column": 57,
                  "Offset": 56
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
                    "text": "("
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
                        "type": "scalar",
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
                              "Column": 17,
                              "Offset": 16
                            },
                            "text": "input"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
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
                        "children": [
                          {
                            "type": "literal",
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
                            "value": "=~",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "match_regexp",
                        "start": {
                          "Line": 1,
                          "Column": 21,
                          "Offset": 20
                        },
                        "end": {
                          "Line": 1,
                          "Column": 28,
                          "Offset": 27
                        },
                        "children": [
                          {
                            "type": "expression_stmt",
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
                                  "Column": 22,
                                  "Offset": 21
                                },
                                "value": "/",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "regexp_content",
                            "start": {
                              "Line": 1,
                              "Column": 22,
                              "Offset": 21
                            },
                            "end": {
                              "Line": 1,
                              "Column": 27,
                              "Offset": 26
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
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
                                    "type": "literal",
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
                                    "value": "\\d",
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
                                "value": "/",
                                "kind": "string"
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
                    "text": ")"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 30,
                      "Offset": 29
                    },
                    "end": {
                      "Line": 1,
                      "Column": 32,
                      "Offset": 31
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 30,
                          "Offset": 29
                        },
                        "end": {
                          "Line": 1,
                          "Column": 32,
                          "Offset": 31
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
                      "Column": 33,
                      "Offset": 32
                    },
                    "end": {
                      "Line": 1,
                      "Column": 34,
                      "Offset": 33
                    },
                    "text": "("
                  },
                  {
                    "type": "binary_expression",
                    "start": {
                      "Line": 1,
                      "Column": 34,
                      "Offset": 33
                    },
                    "end": {
                      "Line": 1,
                      "Column": 56,
                      "Offset": 55
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
                          "Column": 40,
                          "Offset": 39
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
                              "Column": 40,
                              "Offset": 39
                            },
                            "text": "email"
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
                            "value": "=~",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "match_regexp",
                        "start": {
                          "Line": 1,
                          "Column": 44,
                          "Offset": 43
                        },
                        "end": {
                          "Line": 1,
                          "Column": 56,
                          "Offset": 55
                        },
                        "children": [
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
                                "value": "/",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "regexp_content",
                            "start": {
                              "Line": 1,
                              "Column": 45,
                              "Offset": 44
                            },
                            "end": {
                              "Line": 1,
                              "Column": 55,
                              "Offset": 54
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
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
                                "children": [
                                  {
                                    "type": "literal",
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
                                    "value": "@",
                                    "kind": "string"
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
                                  "Column": 51,
                                  "Offset": 50
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
                                      "Column": 51,
                                      "Offset": 50
                                    },
                                    "value": "\\.",
                                    "kind": "string"
                                  }
                                ]
                              },
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 1,
                                  "Column": 51,
                                  "Offset": 50
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 53,
                                  "Offset": 52
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 1,
                                      "Column": 51,
                                      "Offset": 50
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 53,
                                      "Offset": 52
                                    },
                                    "value": "\\w",
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
                                "value": "/",
                                "kind": "string"
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
                      "Column": 56,
                      "Offset": 55
                    },
                    "end": {
                      "Line": 1,
                      "Column": 57,
                      "Offset": 56
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
          "Column": 57,
          "Offset": 56
        },
        "end": {
          "Line": 1,
          "Column": 58,
          "Offset": 57
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 58
}
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

### Typed Perl Output

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
