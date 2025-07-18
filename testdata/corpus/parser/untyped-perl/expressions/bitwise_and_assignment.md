---
category: untyped-perl
subcategory: expressions
tags:
    - bitwise
    - and
    - assignment
    - mask
---

# Bitwise And Assignment

Bitwise AND assignment operator

```perl
$flags &= $mask;
```

# Expected Compilation Outcomes

## Bitwise And Assignment

### Clean Perl Output

```perl
$flags &= $mask;
```

### Typed Perl Output

```perl
$flags &= $mask;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
AST {
  Path: /tmp/bitwise_and_assignment.pl
  Source length: 16 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
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
  "path": "/tmp/bitwise_and_assignment.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 17,
      "Offset": 16
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
          "Column": 16,
          "Offset": 15
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
              "Column": 16,
              "Offset": 15
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
                    "text": "flags"
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
                  "Column": 10,
                  "Offset": 9
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
                      "Column": 10,
                      "Offset": 9
                    },
                    "value": "&=",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 11,
                  "Offset": 10
                },
                "end": {
                  "Line": 1,
                  "Column": 16,
                  "Offset": 15
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
                      "Column": 16,
                      "Offset": 15
                    },
                    "text": "mask"
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
          "Column": 16,
          "Offset": 15
        },
        "end": {
          "Line": 1,
          "Column": 17,
          "Offset": 16
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 16
}
```
