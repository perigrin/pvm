---
category: untyped-perl
subcategory: variables
tags:
    - hashes
    - declarations
    - package-qualification
    - scoping
    - variables
---

# Hash Declarations

Test hash variable declarations with different scoping and assignment patterns

```perl
my %simple_hash;
my %assigned = (key => 'value');
our %package_hash;
%Package::qualified = (a => 1, b => 2);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my %simple_hash;
my %assigned = (key => 'value');
our %package_hash;
%Package::qualified = (a => 1, b => 2);
```

## Typed Perl Output

```perl
my %simple_hash;
my %assigned = (key => 'value');
our %package_hash;
%Package::qualified = (a => 1, b => 2);
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
      hash
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
      hash
        expression_stmt
          literal
        token
  token
  expression_statement
    assignment_expression
      hash
        expression_stmt
          literal
        token
      token
      token
      list_expression
        expression_stmt
          literal
        expression_stmt
          literal
        token
        expression_stmt
          literal
        expression_stmt
          literal
        expression_stmt
          literal
        token
      token
  token
```

## JSON AST

```json
{
  "path": "/tmp/hash_declarations_test.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 40,
      "Offset": 108
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
          "Column": 16,
          "Offset": 15
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
              "Column": 16,
              "Offset": 15
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
                "type": "hash",
                "start": {
                  "Line": 1,
                  "Column": 4,
                  "Offset": 3
                },
                "end": {
                  "Line": 1,
                  "Column": 16,
                  "Offset": 15
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
                        "value": "%",
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
                      "Column": 16,
                      "Offset": 15
                    },
                    "text": "simple_hash"
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
          "Column": 16,
          "Offset": 15
        },
        "end": {
          "Line": 1,
          "Column": 17,
          "Offset": 16
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 17
        },
        "end": {
          "Line": 2,
          "Column": 32,
          "Offset": 48
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 17
            },
            "end": {
              "Line": 2,
              "Column": 32,
              "Offset": 48
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 17
                },
                "end": {
                  "Line": 2,
                  "Column": 32,
                  "Offset": 48
                },
                "name": "assigned",
                "sigil": "%"
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
          "Column": 32,
          "Offset": 48
        },
        "end": {
          "Line": 2,
          "Column": 33,
          "Offset": 49
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 50
        },
        "end": {
          "Line": 3,
          "Column": 18,
          "Offset": 67
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 50
            },
            "end": {
              "Line": 3,
              "Column": 18,
              "Offset": 67
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 50
                },
                "end": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 53
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 1,
                      "Offset": 50
                    },
                    "end": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 53
                    },
                    "value": "our",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "hash",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 54
                },
                "end": {
                  "Line": 3,
                  "Column": 18,
                  "Offset": 67
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 54
                    },
                    "end": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 55
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
                          "Column": 6,
                          "Offset": 55
                        },
                        "value": "%",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 55
                    },
                    "end": {
                      "Line": 3,
                      "Column": 18,
                      "Offset": 67
                    },
                    "text": "package_hash"
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
          "Column": 18,
          "Offset": 67
        },
        "end": {
          "Line": 3,
          "Column": 19,
          "Offset": 68
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 69
        },
        "end": {
          "Line": 4,
          "Column": 39,
          "Offset": 107
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 69
            },
            "end": {
              "Line": 4,
              "Column": 39,
              "Offset": 107
            },
            "children": [
              {
                "type": "hash",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 69
                },
                "end": {
                  "Line": 4,
                  "Column": 20,
                  "Offset": 88
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 1,
                      "Offset": 69
                    },
                    "end": {
                      "Line": 4,
                      "Column": 2,
                      "Offset": 70
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 1,
                          "Offset": 69
                        },
                        "end": {
                          "Line": 4,
                          "Column": 2,
                          "Offset": 70
                        },
                        "value": "%",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 2,
                      "Offset": 70
                    },
                    "end": {
                      "Line": 4,
                      "Column": 20,
                      "Offset": 88
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
                  "Offset": 89
                },
                "end": {
                  "Line": 4,
                  "Column": 22,
                  "Offset": 90
                },
                "text": "="
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 23,
                  "Offset": 91
                },
                "end": {
                  "Line": 4,
                  "Column": 24,
                  "Offset": 92
                },
                "text": "("
              },
              {
                "type": "list_expression",
                "start": {
                  "Line": 4,
                  "Column": 24,
                  "Offset": 92
                },
                "end": {
                  "Line": 4,
                  "Column": 38,
                  "Offset": 106
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 24,
                      "Offset": 92
                    },
                    "end": {
                      "Line": 4,
                      "Column": 25,
                      "Offset": 93
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 24,
                          "Offset": 92
                        },
                        "end": {
                          "Line": 4,
                          "Column": 25,
                          "Offset": 93
                        },
                        "value": "a",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 26,
                      "Offset": 94
                    },
                    "end": {
                      "Line": 4,
                      "Column": 28,
                      "Offset": 96
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 26,
                          "Offset": 94
                        },
                        "end": {
                          "Line": 4,
                          "Column": 28,
                          "Offset": 96
                        },
                        "value": "=>",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 29,
                      "Offset": 97
                    },
                    "end": {
                      "Line": 4,
                      "Column": 30,
                      "Offset": 98
                    },
                    "text": "1"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 30,
                      "Offset": 98
                    },
                    "end": {
                      "Line": 4,
                      "Column": 31,
                      "Offset": 99
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 30,
                          "Offset": 98
                        },
                        "end": {
                          "Line": 4,
                          "Column": 31,
                          "Offset": 99
                        },
                        "value": ",",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 32,
                      "Offset": 100
                    },
                    "end": {
                      "Line": 4,
                      "Column": 33,
                      "Offset": 101
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 32,
                          "Offset": 100
                        },
                        "end": {
                          "Line": 4,
                          "Column": 33,
                          "Offset": 101
                        },
                        "value": "b",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 34,
                      "Offset": 102
                    },
                    "end": {
                      "Line": 4,
                      "Column": 36,
                      "Offset": 104
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 34,
                          "Offset": 102
                        },
                        "end": {
                          "Line": 4,
                          "Column": 36,
                          "Offset": 104
                        },
                        "value": "=>",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 37,
                      "Offset": 105
                    },
                    "end": {
                      "Line": 4,
                      "Column": 38,
                      "Offset": 106
                    },
                    "text": "2"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 38,
                  "Offset": 106
                },
                "end": {
                  "Line": 4,
                  "Column": 39,
                  "Offset": 107
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
          "Line": 4,
          "Column": 39,
          "Offset": 107
        },
        "end": {
          "Line": 4,
          "Column": 40,
          "Offset": 108
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 108
}
```
