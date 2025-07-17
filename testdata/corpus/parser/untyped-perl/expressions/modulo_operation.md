---
category: untyped-perl
subcategory: expressions
tags:
    - modulo
    - math
    - arithmetic
---

# Modulo Operation

Modulo operator for remainder calculation

```perl
$remainder = $dividend % $divisor;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$remainder = $dividend % $divisor;
```

### Typed Perl Output

```perl
$remainder = $dividend % $divisor;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
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

## JSON Format

```json
{
  "path": "/tmp/modulo_operation.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 35,
      "Offset": 34
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
          "Column": 34,
          "Offset": 33
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
              "Column": 34,
              "Offset": 33
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
                  "Column": 11,
                  "Offset": 10
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
                      "Column": 11,
                      "Offset": 10
                    },
                    "text": "remainder"
                  }
                ]
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
                "text": "="
              },
              {
                "type": "binary_expression",
                "start": {
                  "Line": 1,
                  "Column": 14,
                  "Offset": 13
                },
                "end": {
                  "Line": 1,
                  "Column": 34,
                  "Offset": 33
                },
                "children": [
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 14,
                      "Offset": 13
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
                          "Column": 14,
                          "Offset": 13
                        },
                        "end": {
                          "Line": 1,
                          "Column": 15,
                          "Offset": 14
                        },
                        "text": "$"
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
                          "Column": 23,
                          "Offset": 22
                        },
                        "text": "dividend"
                      }
                    ]
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
                      "Column": 25,
                      "Offset": 24
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
                          "Column": 25,
                          "Offset": 24
                        },
                        "value": "%",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 26,
                      "Offset": 25
                    },
                    "end": {
                      "Line": 1,
                      "Column": 34,
                      "Offset": 33
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
                        "text": "$"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 27,
                          "Offset": 26
                        },
                        "end": {
                          "Line": 1,
                          "Column": 34,
                          "Offset": 33
                        },
                        "text": "divisor"
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
          "Column": 34,
          "Offset": 33
        },
        "end": {
          "Line": 1,
          "Column": 35,
          "Offset": 34
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 34
}
```
