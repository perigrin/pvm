---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - iterator
    - closure
    - state
---

# Iterator Pattern

Iterator pattern with closure and state variables

```perl
my $iterator = sub {
    state @items = (1..10);
    state $index = 0;
    return if $index >= @items;
    return $items[$index++];
};

while (defined(my $item = $iterator->())) {
    process($item);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $iterator = sub {
    state @items = (1..10);
    state $index = 0;
    return if $index >= @items;
    return $items[$index++];
};

while (defined(my $item = $iterator->())) {
    process($item);
}
```

## Typed Perl Output

```perl
my $iterator = sub {
    state @items = (1..10);
    state $index = 0;
    return if $index >= @items;
    return $items[$index++];
};

while (defined(my $item = $iterator->())) {
    process($item);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
├── expression_statement
│   └── var_decl
│       └── variable("iterator", "$")
├── token(";")
└── loop_statement
    ├── expression_stmt
    │   └── literal("while")
    ├── token("(")
    ├── func1op_call_expression
    │   ├── expression_stmt
    │   │   └── literal("defined")
    │   ├── token("(")
    │   ├── var_decl
    │   │   └── variable("item", "$")
    │   └── token(")")
    ├── token(")")
    └── block_stmt
        ├── token("{")
        ├── expression_stmt
        │   └── literal("process($item)")
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
      "Line": 11,
      "Column": 1,
      "Offset": 202
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
          "Line": 6,
          "Column": 2,
          "Offset": 133
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 6,
              "Column": 2,
              "Offset": 133
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 6,
                  "Column": 2,
                  "Offset": 133
                },
                "name": "iterator",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 6,
          "Column": 2,
          "Offset": 133
        },
        "end": {
          "Line": 6,
          "Column": 3,
          "Offset": 134
        },
        "text": ";"
      },
      {
        "type": "loop_statement",
        "start": {
          "Line": 8,
          "Column": 1,
          "Offset": 136
        },
        "end": {
          "Line": 10,
          "Column": 2,
          "Offset": 201
        },
        "children": [
          {
            "type": "expression_stmt",
            "start": {
              "Line": 8,
              "Column": 1,
              "Offset": 136
            },
            "end": {
              "Line": 8,
              "Column": 6,
              "Offset": 141
            },
            "children": [
              {
                "type": "literal",
                "start": {
                  "Line": 8,
                  "Column": 1,
                  "Offset": 136
                },
                "end": {
                  "Line": 8,
                  "Column": 6,
                  "Offset": 141
                },
                "value": "while",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 8,
              "Column": 7,
              "Offset": 142
            },
            "end": {
              "Line": 8,
              "Column": 8,
              "Offset": 143
            },
            "text": "("
          },
          {
            "type": "func1op_call_expression",
            "start": {
              "Line": 8,
              "Column": 8,
              "Offset": 143
            },
            "end": {
              "Line": 8,
              "Column": 41,
              "Offset": 176
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 8,
                  "Column": 8,
                  "Offset": 143
                },
                "end": {
                  "Line": 8,
                  "Column": 15,
                  "Offset": 150
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 8,
                      "Column": 8,
                      "Offset": 143
                    },
                    "end": {
                      "Line": 8,
                      "Column": 15,
                      "Offset": 150
                    },
                    "value": "defined",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 8,
                  "Column": 15,
                  "Offset": 150
                },
                "end": {
                  "Line": 8,
                  "Column": 16,
                  "Offset": 151
                },
                "text": "("
              },
              {
                "type": "var_decl",
                "start": {
                  "Line": 8,
                  "Column": 16,
                  "Offset": 151
                },
                "end": {
                  "Line": 8,
                  "Column": 40,
                  "Offset": 175
                },
                "children": [
                  {
                    "type": "variable",
                    "start": {
                      "Line": 8,
                      "Column": 16,
                      "Offset": 151
                    },
                    "end": {
                      "Line": 8,
                      "Column": 40,
                      "Offset": 175
                    },
                    "name": "item",
                    "sigil": "$"
                  }
                ],
                "decl_type": "my"
              },
              {
                "type": "token",
                "start": {
                  "Line": 8,
                  "Column": 40,
                  "Offset": 175
                },
                "end": {
                  "Line": 8,
                  "Column": 41,
                  "Offset": 176
                },
                "text": ")"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 8,
              "Column": 41,
              "Offset": 176
            },
            "end": {
              "Line": 8,
              "Column": 42,
              "Offset": 177
            },
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 8,
              "Column": 43,
              "Offset": 178
            },
            "end": {
              "Line": 10,
              "Column": 2,
              "Offset": 201
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 8,
                  "Column": 43,
                  "Offset": 178
                },
                "end": {
                  "Line": 8,
                  "Column": 44,
                  "Offset": 179
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 9,
                  "Column": 5,
                  "Offset": 184
                },
                "end": {
                  "Line": 9,
                  "Column": 19,
                  "Offset": 198
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 9,
                      "Column": 5,
                      "Offset": 184
                    },
                    "end": {
                      "Line": 9,
                      "Column": 19,
                      "Offset": 198
                    },
                    "value": "process($item)",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 9,
                  "Column": 19,
                  "Offset": 198
                },
                "end": {
                  "Line": 9,
                  "Column": 20,
                  "Offset": 199
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 10,
                  "Column": 1,
                  "Offset": 200
                },
                "end": {
                  "Line": 10,
                  "Column": 2,
                  "Offset": 201
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
  "source_length": 202
}
```
