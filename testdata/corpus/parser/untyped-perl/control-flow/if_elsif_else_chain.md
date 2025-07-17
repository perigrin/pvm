---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - if
    - elsif
    - else
    - chain
---

# If Elsif Else Chain

Complete if-elsif-else conditional chain

```perl
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
} else {
    do_default();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
} else {
    do_default();
}
```

## Typed Perl Output

```perl
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
} else {
    do_default();
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
    elsif
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
      else
        expression_stmt
          literal
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
  "path": "/tmp/if_elsif_else_chain.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 8,
      "Column": 1,
      "Offset": 102
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
          "Line": 7,
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
          },
          {
            "type": "elsif",
            "start": {
              "Line": 3,
              "Column": 3,
              "Offset": 40
            },
            "end": {
              "Line": 7,
              "Column": 2,
              "Offset": 101
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 40
                },
                "end": {
                  "Line": 3,
                  "Column": 8,
                  "Offset": 45
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 3,
                      "Offset": 40
                    },
                    "end": {
                      "Line": 3,
                      "Column": 8,
                      "Offset": 45
                    },
                    "value": "elsif",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 9,
                  "Offset": 46
                },
                "end": {
                  "Line": 3,
                  "Column": 10,
                  "Offset": 47
                },
                "text": "("
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 3,
                  "Column": 10,
                  "Offset": 47
                },
                "end": {
                  "Line": 3,
                  "Column": 16,
                  "Offset": 53
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 10,
                      "Offset": 47
                    },
                    "end": {
                      "Line": 3,
                      "Column": 11,
                      "Offset": 48
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 11,
                      "Offset": 48
                    },
                    "end": {
                      "Line": 3,
                      "Column": 16,
                      "Offset": 53
                    },
                    "text": "other"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 16,
                  "Offset": 53
                },
                "end": {
                  "Line": 3,
                  "Column": 17,
                  "Offset": 54
                },
                "text": ")"
              },
              {
                "type": "block_stmt",
                "start": {
                  "Line": 3,
                  "Column": 18,
                  "Offset": 55
                },
                "end": {
                  "Line": 5,
                  "Column": 2,
                  "Offset": 74
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 18,
                      "Offset": 55
                    },
                    "end": {
                      "Line": 3,
                      "Column": 19,
                      "Offset": 56
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 5,
                      "Offset": 61
                    },
                    "end": {
                      "Line": 4,
                      "Column": 15,
                      "Offset": 71
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 5,
                          "Offset": 61
                        },
                        "end": {
                          "Line": 4,
                          "Column": 15,
                          "Offset": 71
                        },
                        "value": "do_other()",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 15,
                      "Offset": 71
                    },
                    "end": {
                      "Line": 4,
                      "Column": 16,
                      "Offset": 72
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 5,
                      "Column": 1,
                      "Offset": 73
                    },
                    "end": {
                      "Line": 5,
                      "Column": 2,
                      "Offset": 74
                    },
                    "text": "}"
                  }
                ]
              },
              {
                "type": "else",
                "start": {
                  "Line": 5,
                  "Column": 3,
                  "Offset": 75
                },
                "end": {
                  "Line": 7,
                  "Column": 2,
                  "Offset": 101
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 5,
                      "Column": 3,
                      "Offset": 75
                    },
                    "end": {
                      "Line": 5,
                      "Column": 7,
                      "Offset": 79
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 5,
                          "Column": 3,
                          "Offset": 75
                        },
                        "end": {
                          "Line": 5,
                          "Column": 7,
                          "Offset": 79
                        },
                        "value": "else",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "block_stmt",
                    "start": {
                      "Line": 5,
                      "Column": 8,
                      "Offset": 80
                    },
                    "end": {
                      "Line": 7,
                      "Column": 2,
                      "Offset": 101
                    },
                    "children": [
                      {
                        "type": "token",
                        "start": {
                          "Line": 5,
                          "Column": 8,
                          "Offset": 80
                        },
                        "end": {
                          "Line": 5,
                          "Column": 9,
                          "Offset": 81
                        },
                        "text": "{"
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 6,
                          "Column": 5,
                          "Offset": 86
                        },
                        "end": {
                          "Line": 6,
                          "Column": 17,
                          "Offset": 98
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 6,
                              "Column": 5,
                              "Offset": 86
                            },
                            "end": {
                              "Line": 6,
                              "Column": 17,
                              "Offset": 98
                            },
                            "value": "do_default()",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 6,
                          "Column": 17,
                          "Offset": 98
                        },
                        "end": {
                          "Line": 6,
                          "Column": 18,
                          "Offset": 99
                        },
                        "text": ";"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 7,
                          "Column": 1,
                          "Offset": 100
                        },
                        "end": {
                          "Line": 7,
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
