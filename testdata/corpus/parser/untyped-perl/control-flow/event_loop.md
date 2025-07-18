---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - event_loop
    - given
    - when
    - continue
    - validation
---

# Event Loop

Event loop with various event types and validation

```perl
while (my $event = get_next_event()) {
    given ($event->{type}) {
        when ('timer') {
            handle_timer($event);
            continue;
        }
        when ('input') {
            unless (validate_input($event)) {
                log_invalid_input($event);
                next;
            }
            process_input($event);
        }
        when ('shutdown') {
            cleanup_and_exit();
            last;
        }
        default {
            log_unknown_event($event);
        }
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while (my $event = get_next_event()) {
    given ($event->{type}) {
        when ('timer') {
            handle_timer($event);
            continue;
        }
        when ('input') {
            unless (validate_input($event)) {
                log_invalid_input($event);
                next;
            }
            process_input($event);
        }
        when ('shutdown') {
            cleanup_and_exit();
            last;
        }
        default {
            log_unknown_event($event);
        }
    }
}
```

## Typed Perl Output

```perl
while (my $event = get_next_event()) {
    given ($event->{type}) {
        when ('timer') {
            handle_timer($event);
            continue;
        }
        when ('input') {
            unless (validate_input($event)) {
                log_invalid_input($event);
                next;
            }
            process_input($event);
        }
        when ('shutdown') {
            cleanup_and_exit();
            last;
        }
        default {
            log_unknown_event($event);
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
├── loop_statement
│   ├── expression_stmt
│   │   └── literal("while")
│   ├── token("(")
│   ├── var_decl
│   │   └── variable("event", "$")
│   ├── token(")")
│   └── block_stmt
│       ├── token("{")
│       ├── expression_stmt
│       │   └── literal("given ($event->{type}) {\n        when (\"timer\") {\n            handle_timer($event);\n            continue;\n        }\n        when (\"input\") {\n            unless (validate_input($event)) {\n                log_invalid_input($event);\n                next;\n            }\n            process_input($event);\n        }\n        when (\"shutdown\") {\n            cleanup_and_exit();\n            last;\n        }\n        default {\n            log_unknown_event($event);\n        }\n    }")
│       └── token("}")
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
      "Line": 23,
      "Column": 1,
      "Offset": 517
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
          "Line": 22,
          "Column": 2,
          "Offset": 516
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
            "type": "var_decl",
            "start": {
              "Line": 1,
              "Column": 8,
              "Offset": 7
            },
            "end": {
              "Line": 1,
              "Column": 36,
              "Offset": 35
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 1,
                  "Column": 8,
                  "Offset": 7
                },
                "end": {
                  "Line": 1,
                  "Column": 36,
                  "Offset": 35
                },
                "name": "event",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          },
          {
            "type": "token",
            "start": {
              "Line": 1,
              "Column": 36,
              "Offset": 35
            },
            "end": {
              "Line": 1,
              "Column": 37,
              "Offset": 36
            },
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 38,
              "Offset": 37
            },
            "end": {
              "Line": 22,
              "Column": 2,
              "Offset": 516
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 38,
                  "Offset": 37
                },
                "end": {
                  "Line": 1,
                  "Column": 39,
                  "Offset": 38
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 43
                },
                "end": {
                  "Line": 21,
                  "Column": 6,
                  "Offset": 514
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 43
                    },
                    "end": {
                      "Line": 21,
                      "Column": 6,
                      "Offset": 514
                    },
                    "value": "given ($event->{type}) {\n        when (\"timer\") {\n            handle_timer($event);\n            continue;\n        }\n        when (\"input\") {\n            unless (validate_input($event)) {\n                log_invalid_input($event);\n                next;\n            }\n            process_input($event);\n        }\n        when (\"shutdown\") {\n            cleanup_and_exit();\n            last;\n        }\n        default {\n            log_unknown_event($event);\n        }\n    }",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 22,
                  "Column": 1,
                  "Offset": 515
                },
                "end": {
                  "Line": 22,
                  "Column": 2,
                  "Offset": 516
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
  "source_length": 517
}
```
