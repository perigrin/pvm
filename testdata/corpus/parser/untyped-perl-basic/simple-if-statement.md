---
category: untyped-perl
subcategory: control-flow
tags:
    - control-flow
    - conditionals
    - if-statements
---

# Simple If Statement

Basic if statement with conditional logic

```perl
if ($age >= 18) {
    print "You are an adult\n";
} else {
    print "You are a minor\n";
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($age >= 18) {
    print "You are an adult\n";
} else {
    print "You are a minor\n";
}
```

## Typed Perl Output

```perl
if ($age >= 18) {
    print "You are an adult\n";
} else {
    print "You are a minor\n";
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
    relational_expression
      scalar
        token
        token
      expression_stmt
        literal
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
  "path": "/tmp/simple-if-statement.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 5,
      "Column": 2,
      "Offset": 91
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
          "Line": 5,
          "Column": 2,
          "Offset": 91
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
            "type": "relational_expression",
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
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 5,
                  "Offset": 4
                },
                "end": {
                  "Line": 1,
                  "Column": 9,
                  "Offset": 8
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
                      "Column": 9,
                      "Offset": 8
                    },
                    "text": "age"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 1,
                  "Column": 10,
                  "Offset": 9
                },
                "end": {
                  "Line": 1,
                  "Column": 12,
                  "Offset": 11
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 1,
                      "Column": 10,
                      "Offset": 9
                    },
                    "end": {
                      "Line": 1,
                      "Column": 12,
                      "Offset": 11
                    },
                    "value": ">=",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 13,
                  "Offset": 12
                },
                "end": {
                  "Line": 1,
                  "Column": 15,
                  "Offset": 14
                },
                "text": "18"
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
              "Offset": 51
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
                  "Column": 31,
                  "Offset": 48
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
                      "Column": 31,
                      "Offset": 48
                    },
                    "value": "print \"You are an adult\\n\"",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 31,
                  "Offset": 48
                },
                "end": {
                  "Line": 2,
                  "Column": 32,
                  "Offset": 49
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 50
                },
                "end": {
                  "Line": 3,
                  "Column": 2,
                  "Offset": 51
                },
                "text": "}"
              }
            ]
          },
          {
            "type": "else",
            "start": {
              "Line": 3,
              "Column": 3,
              "Offset": 52
            },
            "end": {
              "Line": 5,
              "Column": 2,
              "Offset": 91
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 52
                },
                "end": {
                  "Line": 3,
                  "Column": 7,
                  "Offset": 56
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 3,
                      "Offset": 52
                    },
                    "end": {
                      "Line": 3,
                      "Column": 7,
                      "Offset": 56
                    },
                    "value": "else",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "block_stmt",
                "start": {
                  "Line": 3,
                  "Column": 8,
                  "Offset": 57
                },
                "end": {
                  "Line": 5,
                  "Column": 2,
                  "Offset": 91
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 8,
                      "Offset": 57
                    },
                    "end": {
                      "Line": 3,
                      "Column": 9,
                      "Offset": 58
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 5,
                      "Offset": 63
                    },
                    "end": {
                      "Line": 4,
                      "Column": 30,
                      "Offset": 88
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 5,
                          "Offset": 63
                        },
                        "end": {
                          "Line": 4,
                          "Column": 30,
                          "Offset": 88
                        },
                        "value": "print \"You are a minor\\n\"",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 30,
                      "Offset": 88
                    },
                    "end": {
                      "Line": 4,
                      "Column": 31,
                      "Offset": 89
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 5,
                      "Column": 1,
                      "Offset": 90
                    },
                    "end": {
                      "Line": 5,
                      "Column": 2,
                      "Offset": 91
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
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 91
}
```
