---
category: untyped-perl
subcategory: variables
tags:
    - arrays
    - declarations
    - package-qualification
    - scoping
    - variables
---

# Array Declarations

Test array variable declarations with different scoping and assignment patterns

```perl
my @simple_array;
my @assigned = (1, 2, 3);
our @package_array;
@Package::qualified = qw(a b c);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @simple_array;
my @assigned = (1, 2, 3);
our @package_array;
@Package::qualified = qw(a b c);
```

## Typed Perl Output

```perl
my @simple_array;
my @assigned = (1, 2, 3);
our @package_array;
@Package::qualified = qw(a b c);
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
source_file
  expression_statement
    variable_declaration
      token
      array
        expression_stmt
          literal
        token
  token
  expression_statement
    var_decl
      variable
  token
  expression_statement
    variable_declaration
      expression_stmt
        literal
      array
        expression_stmt
          literal
        token
  token
  expression_statement
    assignment_expression
      array
        expression_stmt
          literal
        token
      token
      quoted_word_list
        expression_stmt
          literal
        expression_stmt
          literal
        expression_stmt
          literal
        expression_stmt
          literal
  token
```

## JSON AST

```json
{
  "path": "/tmp/array_declarations_test.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 33,
      "Offset": 96
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
          "Column": 17,
          "Offset": 16
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 17,
              "Offset": 16
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
                  "Column": 3,
                  "Offset": 2
                },
                "text": "my"
              },
              {
                "type": "array",
                "start": {
                  "Line": 1,
                  "Column": 4,
                  "Offset": 3
                },
                "end": {
                  "Line": 1,
                  "Column": 17,
                  "Offset": 16
                },
                "children": [
                  {
                    "type": "expression_stmt",
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
                    "children": [
                      {
                        "type": "literal",
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
                        "value": "@",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 5,
                      "Offset": 4
                    },
                    "end": {
                      "Line": 1,
                      "Column": 17,
                      "Offset": 16
                    },
                    "text": "simple_array"
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
          "Column": 17,
          "Offset": 16
        },
        "end": {
          "Line": 1,
          "Column": 18,
          "Offset": 17
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 18
        },
        "end": {
          "Line": 2,
          "Column": 25,
          "Offset": 42
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 18
            },
            "end": {
              "Line": 2,
              "Column": 25,
              "Offset": 42
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 18
                },
                "end": {
                  "Line": 2,
                  "Column": 25,
                  "Offset": 42
                },
                "name": "assigned",
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
          "Line": 2,
          "Column": 25,
          "Offset": 42
        },
        "end": {
          "Line": 2,
          "Column": 26,
          "Offset": 43
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 44
        },
        "end": {
          "Line": 3,
          "Column": 19,
          "Offset": 62
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 44
            },
            "end": {
              "Line": 3,
              "Column": 19,
              "Offset": 62
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 44
                },
                "end": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 47
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 1,
                      "Offset": 44
                    },
                    "end": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 47
                    },
                    "value": "our",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "array",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 48
                },
                "end": {
                  "Line": 3,
                  "Column": 19,
                  "Offset": 62
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 48
                    },
                    "end": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 49
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 3,
                          "Column": 5,
                          "Offset": 48
                        },
                        "end": {
                          "Line": 3,
                          "Column": 6,
                          "Offset": 49
                        },
                        "value": "@",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 49
                    },
                    "end": {
                      "Line": 3,
                      "Column": 19,
                      "Offset": 62
                    },
                    "text": "package_array"
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
          "Line": 3,
          "Column": 19,
          "Offset": 62
        },
        "end": {
          "Line": 3,
          "Column": 20,
          "Offset": 63
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 64
        },
        "end": {
          "Line": 4,
          "Column": 32,
          "Offset": 95
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 64
            },
            "end": {
              "Line": 4,
              "Column": 32,
              "Offset": 95
            },
            "children": [
              {
                "type": "array",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 64
                },
                "end": {
                  "Line": 4,
                  "Column": 20,
                  "Offset": 83
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 1,
                      "Offset": 64
                    },
                    "end": {
                      "Line": 4,
                      "Column": 2,
                      "Offset": 65
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 1,
                          "Offset": 64
                        },
                        "end": {
                          "Line": 4,
                          "Column": 2,
                          "Offset": 65
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
                      "Column": 2,
                      "Offset": 65
                    },
                    "end": {
                      "Line": 4,
                      "Column": 20,
                      "Offset": 83
                    },
                    "text": "Package::qualified"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 21,
                  "Offset": 84
                },
                "end": {
                  "Line": 4,
                  "Column": 22,
                  "Offset": 85
                },
                "text": "="
              },
              {
                "type": "quoted_word_list",
                "start": {
                  "Line": 4,
                  "Column": 23,
                  "Offset": 86
                },
                "end": {
                  "Line": 4,
                  "Column": 32,
                  "Offset": 95
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 23,
                      "Offset": 86
                    },
                    "end": {
                      "Line": 4,
                      "Column": 25,
                      "Offset": 88
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 23,
                          "Offset": 86
                        },
                        "end": {
                          "Line": 4,
                          "Column": 25,
                          "Offset": 88
                        },
                        "value": "qw",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 25,
                      "Offset": 88
                    },
                    "end": {
                      "Line": 4,
                      "Column": 26,
                      "Offset": 89
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 25,
                          "Offset": 88
                        },
                        "end": {
                          "Line": 4,
                          "Column": 26,
                          "Offset": 89
                        },
                        "value": "(",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 26,
                      "Offset": 89
                    },
                    "end": {
                      "Line": 4,
                      "Column": 31,
                      "Offset": 94
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 26,
                          "Offset": 89
                        },
                        "end": {
                          "Line": 4,
                          "Column": 31,
                          "Offset": 94
                        },
                        "value": "a b c",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 31,
                      "Offset": 94
                    },
                    "end": {
                      "Line": 4,
                      "Column": 32,
                      "Offset": 95
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 31,
                          "Offset": 94
                        },
                        "end": {
                          "Line": 4,
                          "Column": 32,
                          "Offset": 95
                        },
                        "value": ")",
                        "kind": "string"
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
          "Line": 4,
          "Column": 32,
          "Offset": 95
        },
        "end": {
          "Line": 4,
          "Column": 33,
          "Offset": 96
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 96
}
```
