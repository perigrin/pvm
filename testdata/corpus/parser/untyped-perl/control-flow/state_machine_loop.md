---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - state_machine
    - given
    - when
    - loop
---

# State Machine Loop

State machine implementation with loop and switch

```perl
my $state = 'start';
while (1) {
    given ($state) {
        when ('start') {
            initialize();
            $state = 'process';
        }
        when ('process') {
            if (process_item()) {
                $state = 'finish';
            } else {
                $state = 'error';
            }
        }
        when ('error') {
            handle_error();
            $state = 'start';
        }
        when ('finish') {
            finalize();
            last;
        }
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $state = 'start';
while (1) {
    given ($state) {
        when ('start') {
            initialize();
            $state = 'process';
        }
        when ('process') {
            if (process_item()) {
                $state = 'finish';
            } else {
                $state = 'error';
            }
        }
        when ('error') {
            handle_error();
            $state = 'start';
        }
        when ('finish') {
            finalize();
            last;
        }
    }
}
```

## Typed Perl Output

```perl
my $state = 'start';
while (1) {
    given ($state) {
        when ('start') {
            initialize();
            $state = 'process';
        }
        when ('process') {
            if (process_item()) {
                $state = 'finish';
            } else {
                $state = 'error';
            }
        }
        when ('error') {
            handle_error();
            $state = 'start';
        }
        when ('finish') {
            finalize();
            last;
        }
    }
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
│       └── variable("state", "$")
├── token(";")
└── loop_statement
    ├── expression_stmt
    │   └── literal("while")
    ├── token("(")
    ├── token("1")
    ├── token(")")
    └── block_stmt
        ├── token("{")
        ├── expression_stmt
        │   └── literal("given ($state) {\n        when (\"start\") {\n            initialize();\n            $state = \"process\";\n        }\n        when (\"process\") {\n            if (process_item()) {\n                $state = \"finish\";\n            } else {\n                $state = \"error\";\n            }\n        }\n        when (\"error\") {\n            handle_error();\n            $state = \"start\";\n        }\n        when (\"finish\") {\n            finalize();\n            last;\n        }\n    }")
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
      "Line": 25,
      "Column": 1,
      "Offset": 501
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
          "Column": 20,
          "Offset": 19
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
              "Line": 1,
              "Column": 20,
              "Offset": 19
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
                  "Line": 1,
                  "Column": 20,
                  "Offset": 19
                },
                "name": "state",
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
          "Line": 1,
          "Column": 20,
          "Offset": 19
        },
        "end": {
          "Line": 1,
          "Column": 21,
          "Offset": 20
        },
        "text": ";"
      },
      {
        "type": "loop_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 21
        },
        "end": {
          "Line": 24,
          "Column": 2,
          "Offset": 500
        },
        "children": [
          {
            "type": "expression_stmt",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 21
            },
            "end": {
              "Line": 2,
              "Column": 6,
              "Offset": 26
            },
            "children": [
              {
                "type": "literal",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 21
                },
                "end": {
                  "Line": 2,
                  "Column": 6,
                  "Offset": 26
                },
                "value": "while",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 2,
              "Column": 7,
              "Offset": 27
            },
            "end": {
              "Line": 2,
              "Column": 8,
              "Offset": 28
            },
            "text": "("
          },
          {
            "type": "token",
            "start": {
              "Line": 2,
              "Column": 8,
              "Offset": 28
            },
            "end": {
              "Line": 2,
              "Column": 9,
              "Offset": 29
            },
            "text": "1"
          },
          {
            "type": "token",
            "start": {
              "Line": 2,
              "Column": 9,
              "Offset": 29
            },
            "end": {
              "Line": 2,
              "Column": 10,
              "Offset": 30
            },
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 2,
              "Column": 11,
              "Offset": 31
            },
            "end": {
              "Line": 24,
              "Column": 2,
              "Offset": 500
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 11,
                  "Offset": 31
                },
                "end": {
                  "Line": 2,
                  "Column": 12,
                  "Offset": 32
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 37
                },
                "end": {
                  "Line": 23,
                  "Column": 6,
                  "Offset": 498
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 37
                    },
                    "end": {
                      "Line": 23,
                      "Column": 6,
                      "Offset": 498
                    },
                    "value": "given ($state) {\n        when (\"start\") {\n            initialize();\n            $state = \"process\";\n        }\n        when (\"process\") {\n            if (process_item()) {\n                $state = \"finish\";\n            } else {\n                $state = \"error\";\n            }\n        }\n        when (\"error\") {\n            handle_error();\n            $state = \"start\";\n        }\n        when (\"finish\") {\n            finalize();\n            last;\n        }\n    }",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 24,
                  "Column": 1,
                  "Offset": 499
                },
                "end": {
                  "Line": 24,
                  "Column": 2,
                  "Offset": 500
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
  "source_length": 501
}
```
