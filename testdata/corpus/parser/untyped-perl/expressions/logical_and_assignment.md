---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - and
    - assignment
    - conditional
---

# Logical And Assignment

Logical AND assignment operator

```perl
$flag &&= $condition;
```

# Expected Compilation Outcomes

## Logical And Assignment

### Clean Perl Output

```perl
$flag &&= $condition;
```

### Typed Perl Output

```perl
$flag &&= $condition;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
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
  "path": "/tmp/logical_and_assignment.pl",
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
      "Offset": 22
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
          "Column": 21,
          "Offset": 20
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
              "Column": 21,
              "Offset": 20
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
                  "Column": 6,
                  "Offset": 5
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
                      "Column": 6,
                      "Offset": 5
                    },
                    "text": "flag"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 1,
                  "Column": 7,
                  "Offset": 6
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
                      "Column": 7,
                      "Offset": 6
                    },
                    "end": {
                      "Line": 1,
                      "Column": 10,
                      "Offset": 9
                    },
                    "value": "&&=",
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
                  "Column": 21,
                  "Offset": 20
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
                      "Column": 21,
                      "Offset": 20
                    },
                    "text": "condition"
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
          "Column": 21,
          "Offset": 20
        },
        "end": {
          "Line": 1,
          "Column": 22,
          "Offset": 21
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 22
}
```
