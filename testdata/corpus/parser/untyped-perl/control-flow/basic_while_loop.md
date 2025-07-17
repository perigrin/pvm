---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - while
    - loop
---

# Basic While Loop

Basic while loop

```perl
while ($condition) {
    process();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while ($condition) {
    process();
}
```

## Typed Perl Output

```perl
while ($condition) {
    process();
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
  loop_statement
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
  "path": "/tmp/basic_while_loop.pl",
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
      "Offset": 38
    },
    "children": [
      {
        "type": "loop_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 3,
          "Column": 2,
          "Offset": 37
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
              "Column": 6,
              "Offset": 5
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
                  "Column": 6,
                  "Offset": 5
                },
                "value": "while",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 1,
              "Column": 7,
              "Offset": 6
            },
            "end": {
              "Line": 1,
              "Column": 8,
              "Offset": 7
            },
            "text": "("
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
              "Column": 18,
              "Offset": 17
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
                  "Column": 18,
                  "Offset": 17
                },
                "text": "condition"
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
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 20,
              "Offset": 19
            },
            "end": {
              "Line": 3,
              "Column": 2,
              "Offset": 37
            },
            "children": [
              {
                "type": "token",
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
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 25
                },
                "end": {
                  "Line": 2,
                  "Column": 14,
                  "Offset": 34
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 25
                    },
                    "end": {
                      "Line": 2,
                      "Column": 14,
                      "Offset": 34
                    },
                    "value": "process()",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 14,
                  "Offset": 34
                },
                "end": {
                  "Line": 2,
                  "Column": 15,
                  "Offset": 35
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 36
                },
                "end": {
                  "Line": 3,
                  "Column": 2,
                  "Offset": 37
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
  "source_length": 38
}
```
