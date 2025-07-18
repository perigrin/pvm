---
category: untyped-perl
subcategory: expressions
tags:
    - assignment
    - chained
    - multiple
    - sequential
---

# Chained Assignment

Chained assignment operations

```perl
$a = $b = $c = $value;
```

# Expected Compilation Outcomes

## Chained Assignment

### Clean Perl Output

```perl
$a = $b = $c = $value;
```

### Typed Perl Output

```perl
$a = $b = $c = $value;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  expression_statement
    assignment_expression
      scalar
        token
        token
      token
      assignment_expression
        scalar
          token
          token
        token
        assignment_expression
          scalar
            token
            token
          token
          scalar
            token
            token
  token
```

## JSON AST

```json
{
  "path": "/tmp/test_chained_assignment.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 23,
      "Offset": 22
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
          "Column": 22,
          "Offset": 21
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 22,
              "Offset": 21
            },
            "children": [
              {
                "type": "scalar",
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
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 1,
                      "Offset": 0
                    },
                    "end": {
                      "Line": 1,
                      "Column": 2,
                      "Offset": 1
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 2,
                      "Offset": 1
                    },
                    "end": {
                      "Line": 1,
                      "Column": 3,
                      "Offset": 2
                    },
                    "text": "a"
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
                "text": "="
              },
              {
                "type": "assignment_expression",
                "start": {
                  "Line": 1,
                  "Column": 6,
                  "Offset": 5
                },
                "end": {
                  "Line": 1,
                  "Column": 22,
                  "Offset": 21
                },
                "children": [
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 1,
                      "Column": 6,
                      "Offset": 5
                    },
                    "end": {
                      "Line": 1,
                      "Column": 8,
                      "Offset": 7
                    },
                    "children": [
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 6,
                          "Offset": 5
                        },
                        "end": {
                          "Line": 1,
                          "Column": 7,
                          "Offset": 6
                        },
                        "text": "$"
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
                        "text": "b"
                      }
                    ]
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
                    "text": "="
                  },
                  {
                    "type": "assignment_expression",
                    "start": {
                      "Line": 1,
                      "Column": 11,
                      "Offset": 10
                    },
                    "end": {
                      "Line": 1,
                      "Column": 22,
                      "Offset": 21
                    },
                    "children": [
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 11,
                          "Offset": 10
                        },
                        "end": {
                          "Line": 1,
                          "Column": 13,
                          "Offset": 12
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
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 12,
                              "Offset": 11
                            },
                            "end": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
                            },
                            "text": "c"
                          }
                        ]
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 1,
                          "Column": 14,
                          "Offset": 13
                        },
                        "end": {
                          "Line": 1,
                          "Column": 15,
                          "Offset": 14
                        },
                        "text": "="
                      },
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 16,
                          "Offset": 15
                        },
                        "end": {
                          "Line": 1,
                          "Column": 22,
                          "Offset": 21
                        },
                        "children": [
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 16,
                              "Offset": 15
                            },
                            "end": {
                              "Line": 1,
                              "Column": 17,
                              "Offset": 16
                            },
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 17,
                              "Offset": 16
                            },
                            "end": {
                              "Line": 1,
                              "Column": 22,
                              "Offset": 21
                            },
                            "text": "value"
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
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 22,
          "Offset": 21
        },
        "end": {
          "Line": 1,
          "Column": 23,
          "Offset": 22
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 22
}
```
