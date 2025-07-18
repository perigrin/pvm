---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - redo
    - restart
---

# Redo Statement

Redo statement to restart current iteration

```perl
while ($condition) {
    my $input = get_input();
    redo if invalid($input);
    process($input);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while ($condition) {
    my $input = get_input();
    redo if invalid($input);
    process($input);
}
```

## Typed Perl Output

```perl
while ($condition) {
    my $input = get_input();
    redo if invalid($input);
    process($input);
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
    ├── scalar
    │   ├── token("$")
    │   └── token("condition")
    ├── token(")")
    └── block_stmt
        ├── token("{")
        ├── expression_stmt
        │   └── literal("my $input = get_input()")
        ├── token(";")
        ├── expression_stmt
        │   └── literal("redo if invalid($input)")
        ├── token(";")
        ├── expression_stmt
        │   └── literal("process($input)")
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
      "Line": 6,
      "Column": 1,
      "Offset": 102
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
          "Line": 5,
          "Column": 2,
          "Offset": 101
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
              "Line": 5,
              "Column": 2,
              "Offset": 101
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
                  "Column": 28,
                  "Offset": 48
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
                      "Column": 28,
                      "Offset": 48
                    },
                    "value": "my $input = get_input()",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 28,
                  "Offset": 48
                },
                "end": {
                  "Line": 2,
                  "Column": 29,
                  "Offset": 49
                },
                "text": ";"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 54
                },
                "end": {
                  "Line": 3,
                  "Column": 28,
                  "Offset": 77
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 54
                    },
                    "end": {
                      "Line": 3,
                      "Column": 28,
                      "Offset": 77
                    },
                    "value": "redo if invalid($input)",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 28,
                  "Offset": 77
                },
                "end": {
                  "Line": 3,
                  "Column": 29,
                  "Offset": 78
                },
                "text": ";"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 4,
                  "Column": 5,
                  "Offset": 83
                },
                "end": {
                  "Line": 4,
                  "Column": 20,
                  "Offset": 98
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 4,
                      "Column": 5,
                      "Offset": 83
                    },
                    "end": {
                      "Line": 4,
                      "Column": 20,
                      "Offset": 98
                    },
                    "value": "process($input)",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 20,
                  "Offset": 98
                },
                "end": {
                  "Line": 4,
                  "Column": 21,
                  "Offset": 99
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 5,
                  "Column": 1,
                  "Offset": 100
                },
                "end": {
                  "Line": 5,
                  "Column": 2,
                  "Offset": 101
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
  "source_length": 102
}
```
