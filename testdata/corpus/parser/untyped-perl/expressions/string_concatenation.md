---
category: untyped-perl
subcategory: expressions
tags:
    - string
    - concatenation
    - basic
---

# String Concatenation

Basic string concatenation operator

```perl
$combined = $first . $second;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$combined = $first . $second;
```

### Typed Perl Output

```perl
$combined = $first . $second;
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
  "path": "/tmp/string_concatenation.pl",
  "root": {
    "type": "source_file",
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
        "type": "expression_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 1,
          "Column": 29,
          "Offset": 28
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
              "Column": 29,
              "Offset": 28
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
                  "Column": 10,
                  "Offset": 9
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
                      "Column": 10,
                      "Offset": 9
                    },
                    "text": "combined"
                  }
                ]
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
                "text": "="
              },
              {
                "type": "binary_expression",
                "start": {
                  "Line": 1,
                  "Column": 13,
                  "Offset": 12
                },
                "end": {
                  "Line": 1,
                  "Column": 29,
                  "Offset": 28
                },
                "children": [
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 13,
                      "Offset": 12
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
                          "Column": 13,
                          "Offset": 12
                        },
                        "end": {
                          "Line": 1,
                          "Column": 14,
                          "Offset": 13
                        },
                        "text": "$"
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
                          "Column": 19,
                          "Offset": 18
                        },
                        "text": "first"
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
                      "Column": 21,
                      "Offset": 20
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
                          "Column": 21,
                          "Offset": 20
                        },
                        "value": ".",
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
                      "Column": 29,
                      "Offset": 28
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
                          "Column": 29,
                          "Offset": 28
                        },
                        "text": "second"
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
          "Column": 29,
          "Offset": 28
        },
        "end": {
          "Line": 1,
          "Column": 30,
          "Offset": 29
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 29
}
```
