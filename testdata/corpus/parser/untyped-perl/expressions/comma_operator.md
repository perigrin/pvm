---
category: untyped-perl
subcategory: expressions
tags:
    - comma
    - sequence
    - evaluation
    - side_effects
---

# Comma Operator

Comma operator for sequential evaluation

```perl
$result = ($operation1, $operation2, $final_value);
```

# Expected Compilation Outcomes

## Comma Operator

### Clean Perl Output

```perl
$result = ($operation1, $operation2, $final_value);
```

### Typed Perl Output

```perl
$result = ($operation1, $operation2, $final_value);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/comma_operator.pl
  Source length: 52 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
        scalar
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
  "path": "/tmp/comma_operator.pl",
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
      "Offset": 52
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
          "Column": 51,
          "Offset": 50
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
              "Column": 51,
              "Offset": 50
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
                "type": "list_expression",
                "start": {
                  "Line": 1,
                  "Column": 12,
                  "Offset": 11
                },
                "end": {
                  "Line": 1,
                  "Column": 50,
                  "Offset": 49
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
                      "Column": 23,
                      "Offset": 22
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
                          "Column": 23,
                          "Offset": 22
                        },
                        "text": "operation1"
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
                        "value": ",",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 25,
                      "Offset": 24
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
                          "Column": 25,
                          "Offset": 24
                        },
                        "end": {
                          "Line": 1,
                          "Column": 26,
                          "Offset": 25
                        },
                        "text": "$"
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
                          "Column": 36,
                          "Offset": 35
                        },
                        "text": "operation2"
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
                        "value": ",",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 38,
                      "Offset": 37
                    },
                    "end": {
                      "Line": 1,
                      "Column": 50,
                      "Offset": 49
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
                          "Column": 50,
                          "Offset": 49
                        },
                        "text": "final_value"
                      }
                    ]
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
                  "Column": 51,
                  "Offset": 50
                },
                "text": ")"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 51,
          "Offset": 50
        },
        "end": {
          "Line": 1,
          "Column": 52,
          "Offset": 51
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 52
}
```
