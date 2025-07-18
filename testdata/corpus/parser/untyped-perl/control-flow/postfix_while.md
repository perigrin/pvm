---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - postfix
    - do_while
---

# Postfix While

Do-while loop (postfix while)

```perl
do {
    process();
} while ($continue);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
do {
    process();
} while ($continue);
```

## Typed Perl Output

```perl
do {
    process();
} while ($continue);
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
├── expression_statement
│   └── postfix_loop_expression
│       ├── do_expression
│       │   ├── expression_stmt
│       │   │   └── literal("do")
│       │   └── block_stmt
│       │       ├── token("{")
│       │       ├── expression_stmt
│       │       │   └── literal("process()")
│       │       ├── token(";")
│       │       └── token("}")
│       ├── expression_stmt
│       │   └── literal("while")
│       ├── token("(")
│       ├── scalar
│       │   ├── token("$")
│       │   └── token("continue")
│       └── token(")")
└── token(";")
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
      "Line": 4,
      "Column": 1,
      "Offset": 41
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
          "Line": 3,
          "Column": 20,
          "Offset": 39
        },
        "children": [
          {
            "type": "postfix_loop_expression",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 3,
              "Column": 20,
              "Offset": 39
            },
            "children": [
              {
                "type": "do_expression",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 3,
                  "Column": 2,
                  "Offset": 21
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
                        "value": "do",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "block_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 4,
                      "Offset": 3
                    },
                    "end": {
                      "Line": 3,
                      "Column": 2,
                      "Offset": 21
                    },
                    "children": [
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
                        "text": "{"
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 5,
                          "Offset": 9
                        },
                        "end": {
                          "Line": 2,
                          "Column": 14,
                          "Offset": 18
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 5,
                              "Offset": 9
                            },
                            "end": {
                              "Line": 2,
                              "Column": 14,
                              "Offset": 18
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
                          "Offset": 18
                        },
                        "end": {
                          "Line": 2,
                          "Column": 15,
                          "Offset": 19
                        },
                        "text": ";"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 3,
                          "Column": 1,
                          "Offset": 20
                        },
                        "end": {
                          "Line": 3,
                          "Column": 2,
                          "Offset": 21
                        },
                        "text": "}"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 22
                },
                "end": {
                  "Line": 3,
                  "Column": 8,
                  "Offset": 27
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 3,
                      "Offset": 22
                    },
                    "end": {
                      "Line": 3,
                      "Column": 8,
                      "Offset": 27
                    },
                    "value": "while",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 9,
                  "Offset": 28
                },
                "end": {
                  "Line": 3,
                  "Column": 10,
                  "Offset": 29
                },
                "text": "("
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 3,
                  "Column": 10,
                  "Offset": 29
                },
                "end": {
                  "Line": 3,
                  "Column": 19,
                  "Offset": 38
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 10,
                      "Offset": 29
                    },
                    "end": {
                      "Line": 3,
                      "Column": 11,
                      "Offset": 30
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 11,
                      "Offset": 30
                    },
                    "end": {
                      "Line": 3,
                      "Column": 19,
                      "Offset": 38
                    },
                    "text": "continue"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 19,
                  "Offset": 38
                },
                "end": {
                  "Line": 3,
                  "Column": 20,
                  "Offset": 39
                },
                "text": ")"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 3,
          "Column": 20,
          "Offset": 39
        },
        "end": {
          "Line": 3,
          "Column": 21,
          "Offset": 40
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 41
}
```
