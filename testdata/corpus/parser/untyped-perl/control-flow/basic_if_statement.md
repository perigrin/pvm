---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - if
    - conditional
    - block
---

# Basic If Statement

Basic if statement with block

```perl
if ($condition) {
    do_something();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($condition) {
    do_something();
}
```

## Typed Perl Output

```perl
if ($condition) {
    do_something();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text Format

```
source_file
  conditional_statement
    expression_stmt
      literal
    token
    scalar
      token
      token
    token
    block_stmt
      token
      expression_stmt
        literal
      token
      token
```

## JSON Format

```json
{
  "path": "/tmp/basic_if_statement.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 1,
      "Offset": 40
    },
    "children": [
      {
        "type": "conditional_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 3,
          "Column": 2,
          "Offset": 39
        },
        "children": [
          {
            "type": "expression_stmt",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 3,
              "Offset": 2
            },
            "children": [
              {
                "type": "literal",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 3,
                  "Offset": 2
                },
                "value": "if",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 1,
              "Column": 4,
              "Offset": 3
            },
            "end": {
              "Line": 1,
              "Column": 5,
              "Offset": 4
            },
            "text": "("
          },
          {
            "type": "scalar",
            "start": {
              "Line": 1,
              "Column": 5,
              "Offset": 4
            },
            "end": {
              "Line": 1,
              "Column": 15,
              "Offset": 14
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 5,
                  "Offset": 4
                },
                "end": {
                  "Line": 1,
                  "Column": 6,
                  "Offset": 5
                },
                "text": "$"
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
                  "Column": 15,
                  "Offset": 14
                },
                "text": "condition"
              }
            ]
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
              "Column": 16,
              "Offset": 15
            },
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 17,
              "Offset": 16
            },
            "end": {
              "Line": 3,
              "Column": 2,
              "Offset": 39
            },
            "children": [
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
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 22
                },
                "end": {
                  "Line": 2,
                  "Column": 19,
                  "Offset": 36
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 22
                    },
                    "end": {
                      "Line": 2,
                      "Column": 19,
                      "Offset": 36
                    },
                    "value": "do_something()",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 19,
                  "Offset": 36
                },
                "end": {
                  "Line": 2,
                  "Column": 20,
                  "Offset": 37
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 38
                },
                "end": {
                  "Line": 3,
                  "Column": 2,
                  "Offset": 39
                },
                "text": "}"
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 40
}
```
