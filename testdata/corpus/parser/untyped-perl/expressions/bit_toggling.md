---
category: untyped-perl
subcategory: expressions
tags:
    - bit_manipulation
    - bitwise
    - flag_toggling
    - xor
---

# Bit Toggling

Toggling a specific bit using XOR

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

# Expected Compilation Outcomes

## Bit Toggling

### Clean Perl Output

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
AST {
  Path: /tmp/bit_toggling.pl
  Source length: 43 characters
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
  "path": "/tmp/bit_toggling.pl",
  "root": {
    "type": "source_file",
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
        "type": "expression_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 1,
          "Column": 43,
          "Offset": 42
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
              "Column": 43,
              "Offset": 42
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
                    "text": "bit_toggled"
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
                  "Column": 43,
                  "Offset": 42
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
                        "value": "^",
                        "kind": "string"
                      }
                    ]
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
                    "text": "("
                  },
                  {
                    "type": "binary_expression",
                    "start": {
                      "Line": 1,
                      "Column": 26,
                      "Offset": 25
                    },
                    "end": {
                      "Line": 1,
                      "Column": 42,
                      "Offset": 41
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
                        "text": "1"
                      },
                      {
                        "type": "expression_stmt",
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
                            "type": "literal",
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
                            "value": "<<",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 31,
                          "Offset": 30
                        },
                        "end": {
                          "Line": 1,
                          "Column": 42,
                          "Offset": 41
                        },
                        "children": [
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
                            "text": "$"
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
                              "Column": 42,
                              "Offset": 41
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
                      "Column": 42,
                      "Offset": 41
                    },
                    "end": {
                      "Line": 1,
                      "Column": 43,
                      "Offset": 42
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
          "Column": 43,
          "Offset": 42
        },
        "end": {
          "Line": 1,
          "Column": 44,
          "Offset": 43
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 43
}
```
