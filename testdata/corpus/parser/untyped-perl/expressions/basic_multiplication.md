---
category: untyped-perl
subcategory: expressions
tags:
    - multiplication
    - arithmetic
    - basic
    - binary_operator
---

# Basic Multiplication

Basic multiplication operator

```perl
$result = $a * $b;
```

# Expected Compilation Outcomes

## Basic Multiplication

### Clean Perl Output

```perl
$result = $a * $b;
```

### Typed Perl Output

```perl
$result = $a * $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
AST {
  Path: /tmp/basic_multiplication.pl
  Source length: 19 characters
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
          scalar
            token
            token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/basic_multiplication.pl",
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
      "Offset": 19
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
          "Column": 18,
          "Offset": 17
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
              "Column": 18,
              "Offset": 17
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
                  "Column": 18,
                  "Offset": 17
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
                      "Column": 13,
                      "Offset": 12
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
                          "Column": 13,
                          "Offset": 12
                        },
                        "text": "a"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
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
                    "children": [
                      {
                        "type": "literal",
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
                        "value": "*",
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
              }
            ]
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
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 19
}
```
