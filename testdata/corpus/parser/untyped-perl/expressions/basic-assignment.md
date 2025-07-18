---
category: untyped-perl
subcategory: expressions
tags:
    - basic
    - assignment
    - variables
---

# Basic Assignment

Basic assignment operator

```perl
$var = $value;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
$var = $value;
```

## Typed Perl Output

```perl
$var = $value;
```

## Inferred Perl Output

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
      scalar
        token
        token
  token
```

## JSON Format

```json
{
  "path": "/tmp/basic_assignment.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 15,
      "Offset": 14
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
          "Column": 14,
          "Offset": 13
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
              "Column": 14,
              "Offset": 13
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
                  "Column": 5,
                  "Offset": 4
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
                      "Column": 5,
                      "Offset": 4
                    },
                    "text": "var"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 6,
                  "Offset": 5
                },
                "end": {
                  "Line": 1,
                  "Column": 7,
                  "Offset": 6
                },
                "text": "="
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 8,
                  "Offset": 7
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
                      "Column": 8,
                      "Offset": 7
                    },
                    "end": {
                      "Line": 1,
                      "Column": 9,
                      "Offset": 8
                    },
                    "text": "$"
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
                      "Column": 14,
                      "Offset": 13
                    },
                    "text": "value"
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
          "Column": 14,
          "Offset": 13
        },
        "end": {
          "Line": 1,
          "Column": 15,
          "Offset": 14
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 14
}
```
