---
category: untyped-perl
subcategory: expressions
tags:
    - bit_manipulation
    - bitwise
    - flag_setting
    - left_shift
---

# Bit Manipulation

Setting a specific bit using shift and OR

```perl
$bit_set = $flags | (1 << $bit_number);
```

# Expected Compilation Outcomes

## Bit Manipulation

### Clean Perl Output

```perl
$bit_set = $flags | (1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_set = $flags | (1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
AST {
  Path: /tmp/bit_manipulation.pl
  Source length: 39 characters
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
  "path": "/tmp/bit_manipulation.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 40,
      "Offset": 39
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
          "Column": 39,
          "Offset": 38
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
              "Column": 39,
              "Offset": 38
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
                    "text": "bit_set"
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
                  "Column": 39,
                  "Offset": 38
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
                      "Column": 18,
                      "Offset": 17
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
                          "Column": 18,
                          "Offset": 17
                        },
                        "text": "flags"
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
                        "value": "|",
                        "kind": "string"
                      }
                    ]
                  },
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
                    "text": "("
                  },
                  {
                    "type": "binary_expression",
                    "start": {
                      "Line": 1,
                      "Column": 22,
                      "Offset": 21
                    },
                    "end": {
                      "Line": 1,
                      "Column": 38,
                      "Offset": 37
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
                        "text": "1"
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
                          "Column": 26,
                          "Offset": 25
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
                              "Column": 26,
                              "Offset": 25
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
                          "Column": 27,
                          "Offset": 26
                        },
                        "end": {
                          "Line": 1,
                          "Column": 38,
                          "Offset": 37
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
                              "Column": 38,
                              "Offset": 37
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
                      "Column": 38,
                      "Offset": 37
                    },
                    "end": {
                      "Line": 1,
                      "Column": 39,
                      "Offset": 38
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
          "Column": 39,
          "Offset": 38
        },
        "end": {
          "Line": 1,
          "Column": 40,
          "Offset": 39
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 39
}
```
