---
category: untyped-perl
subcategory: variables
tags:
    - arrays
    - variables
    - assignments
---

# Array Operations

Basic array operations and assignments

```perl
my @numbers = (1, 2, 3, 4, 5);
my $first = $numbers[0];
my $count = @numbers;
push @numbers, 6;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @numbers = (1, 2, 3, 4, 5);
my $first = $numbers[0];
my $count = @numbers;
push @numbers, 6;
```

## Typed Perl Output

```perl
my @numbers = (1, 2, 3, 4, 5);
my $first = $numbers[0];
my $count = @numbers;
push @numbers, 6;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text Format

```
source_file
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
  expression_statement
    ambiguous_function_call_expression
      expression_stmt
        literal
      list_expression
        array
          expression_stmt
            literal
          token
        expression_stmt
          literal
        token
  token
```

## JSON Format

```json
{
  "path": "/tmp/array-operations.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 18,
      "Offset": 95
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
          "Column": 30,
          "Offset": 29
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
              "Column": 30,
              "Offset": 29
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
                  "Column": 30,
                  "Offset": 29
                },
                "name": "numbers",
                "sigil": "@"
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
          "Column": 30,
          "Offset": 29
        },
        "end": {
          "Line": 1,
          "Column": 31,
          "Offset": 30
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 31
        },
        "end": {
          "Line": 2,
          "Column": 24,
          "Offset": 54
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 31
            },
            "end": {
              "Line": 2,
              "Column": 24,
              "Offset": 54
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 31
                },
                "end": {
                  "Line": 2,
                  "Column": 24,
                  "Offset": 54
                },
                "name": "first",
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
          "Line": 2,
          "Column": 24,
          "Offset": 54
        },
        "end": {
          "Line": 2,
          "Column": 25,
          "Offset": 55
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 56
        },
        "end": {
          "Line": 3,
          "Column": 21,
          "Offset": 76
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 56
            },
            "end": {
              "Line": 3,
              "Column": 21,
              "Offset": 76
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 56
                },
                "end": {
                  "Line": 3,
                  "Column": 21,
                  "Offset": 76
                },
                "name": "count",
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
          "Line": 3,
          "Column": 21,
          "Offset": 76
        },
        "end": {
          "Line": 3,
          "Column": 22,
          "Offset": 77
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 78
        },
        "end": {
          "Line": 4,
          "Column": 17,
          "Offset": 94
        },
        "children": [
          {
            "type": "ambiguous_function_call_expression",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 78
            },
            "end": {
              "Line": 4,
              "Column": 17,
              "Offset": 94
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 78
                },
                "end": {
                  "Line": 4,
                  "Column": 5,
                  "Offset": 82
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 4,
                      "Column": 1,
                      "Offset": 78
                    },
                    "end": {
                      "Line": 4,
                      "Column": 5,
                      "Offset": 82
                    },
                    "value": "push",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "list_expression",
                "start": {
                  "Line": 4,
                  "Column": 6,
                  "Offset": 83
                },
                "end": {
                  "Line": 4,
                  "Column": 17,
                  "Offset": 94
                },
                "children": [
                  {
                    "type": "array",
                    "start": {
                      "Line": 4,
                      "Column": 6,
                      "Offset": 83
                    },
                    "end": {
                      "Line": 4,
                      "Column": 14,
                      "Offset": 91
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 4,
                          "Column": 6,
                          "Offset": 83
                        },
                        "end": {
                          "Line": 4,
                          "Column": 7,
                          "Offset": 84
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 4,
                              "Column": 6,
                              "Offset": 83
                            },
                            "end": {
                              "Line": 4,
                              "Column": 7,
                              "Offset": 84
                            },
                            "value": "@",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 4,
                          "Column": 7,
                          "Offset": 84
                        },
                        "end": {
                          "Line": 4,
                          "Column": 14,
                          "Offset": 91
                        },
                        "text": "numbers"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 14,
                      "Offset": 91
                    },
                    "end": {
                      "Line": 4,
                      "Column": 15,
                      "Offset": 92
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 14,
                          "Offset": 91
                        },
                        "end": {
                          "Line": 4,
                          "Column": 15,
                          "Offset": 92
                        },
                        "value": ",",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 16,
                      "Offset": 93
                    },
                    "end": {
                      "Line": 4,
                      "Column": 17,
                      "Offset": 94
                    },
                    "text": "6"
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
          "Line": 4,
          "Column": 17,
          "Offset": 94
        },
        "end": {
          "Line": 4,
          "Column": 18,
          "Offset": 95
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 95
}
```
