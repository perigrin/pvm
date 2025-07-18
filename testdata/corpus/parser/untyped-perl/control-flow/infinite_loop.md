---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - while
    - infinite
    - last
---

# Infinite Loop

Infinite loop with break condition

```perl
while (1) {
    handle_request();
    last if $shutdown;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while (1) {
    handle_request();
    last if $shutdown;
}
```

## Typed Perl Output

```perl
while (1) {
    handle_request();
    last if $shutdown;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
└── loop_statement
    ├── expression_stmt
    │   └── literal("while")
    ├── token("(")
    ├── token("1")
    ├── token(")")
    └── block_stmt
        ├── token("{")
        ├── expression_stmt
        │   └── literal("handle_request()")
        ├── token(";")
        ├── expression_stmt
        │   └── literal("last if $shutdown")
        ├── token(";")
        └── token("}")
```

## JSON AST

```json
{
  "path": "temp_test.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 5,
      "Column": 1,
      "Offset": 59
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
          "Line": 4,
          "Column": 2,
          "Offset": 58
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
            "text": "1"
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
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 11,
              "Offset": 10
            },
            "end": {
              "Line": 4,
              "Column": 2,
              "Offset": 58
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
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 16
                },
                "end": {
                  "Line": 2,
                  "Column": 21,
                  "Offset": 32
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 16
                    },
                    "end": {
                      "Line": 2,
                      "Column": 21,
                      "Offset": 32
                    },
                    "value": "handle_request()",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 21,
                  "Offset": 32
                },
                "end": {
                  "Line": 2,
                  "Column": 22,
                  "Offset": 33
                },
                "text": ";"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 38
                },
                "end": {
                  "Line": 3,
                  "Column": 22,
                  "Offset": 55
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 38
                    },
                    "end": {
                      "Line": 3,
                      "Column": 22,
                      "Offset": 55
                    },
                    "value": "last if $shutdown",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 22,
                  "Offset": 55
                },
                "end": {
                  "Line": 3,
                  "Column": 23,
                  "Offset": 56
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 57
                },
                "end": {
                  "Line": 4,
                  "Column": 2,
                  "Offset": 58
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
  "source_length": 59
}
```
