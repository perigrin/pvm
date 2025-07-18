---
category: untyped-perl
subcategory: variables
tags:
    - naming
    - package-qualification
    - variables
---

# Complex Variable Names

Test variables with underscores, numbers, and package qualifiers

```perl
my $var_with_underscores;
my $var123;
my $package_var;
$Some::Package::qualified_var = "test";
our $_private_var;
my $CamelCase_Var;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $var_with_underscores;
my $var123;
my $package_var;
$Some::Package::qualified_var = "test";
our $_private_var;
my $CamelCase_Var;
```

## Typed Perl Output

```perl
my $var_with_underscores;
my $var123;
my $package_var;
$Some::Package::qualified_var = "test";
our $_private_var;
my $CamelCase_Var;
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
      scalar
        token
        token
  token
  expression_statement
    variable_declaration
      token
      scalar
        token
        token
  token
  expression_statement
    variable_declaration
      token
      scalar
        token
        token
  token
  expression_statement
    assignment_expression
      scalar
        token
        token
      token
      interpolated_string_literal
        expression_stmt
          literal
        expression_stmt
          literal
        expression_stmt
          literal
  token
  expression_statement
    variable_declaration
      expression_stmt
        literal
      scalar
        token
        token
  token
  expression_statement
    variable_declaration
      token
      scalar
        token
        token
  token
```

## JSON AST

```json
{
  "path": "/tmp/complex_variable_names_test.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 6,
      "Column": 19,
      "Offset": 132
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
          "Column": 25,
          "Offset": 24
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
              "Column": 25,
              "Offset": 24
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
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 4,
                  "Offset": 3
                },
                "end": {
                  "Line": 1,
                  "Column": 25,
                  "Offset": 24
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
                    "text": "$"
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
                      "Column": 25,
                      "Offset": 24
                    },
                    "text": "var_with_underscores"
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
          "Column": 25,
          "Offset": 24
        },
        "end": {
          "Line": 1,
          "Column": 26,
          "Offset": 25
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 26
        },
        "end": {
          "Line": 2,
          "Column": 11,
          "Offset": 36
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 26
            },
            "end": {
              "Line": 2,
              "Column": 11,
              "Offset": 36
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 26
                },
                "end": {
                  "Line": 2,
                  "Column": 3,
                  "Offset": 28
                },
                "text": "my"
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 29
                },
                "end": {
                  "Line": 2,
                  "Column": 11,
                  "Offset": 36
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 4,
                      "Offset": 29
                    },
                    "end": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 30
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 30
                    },
                    "end": {
                      "Line": 2,
                      "Column": 11,
                      "Offset": 36
                    },
                    "text": "var123"
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
          "Line": 2,
          "Column": 11,
          "Offset": 36
        },
        "end": {
          "Line": 2,
          "Column": 12,
          "Offset": 37
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 38
        },
        "end": {
          "Line": 3,
          "Column": 16,
          "Offset": 53
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 38
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
                  "Column": 1,
                  "Offset": 38
                },
                "end": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 40
                },
                "text": "my"
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 41
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
                      "Column": 4,
                      "Offset": 41
                    },
                    "end": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 42
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 42
                    },
                    "end": {
                      "Line": 3,
                      "Column": 16,
                      "Offset": 53
                    },
                    "text": "package_var"
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
          "Column": 16,
          "Offset": 53
        },
        "end": {
          "Line": 3,
          "Column": 17,
          "Offset": 54
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 55
        },
        "end": {
          "Line": 4,
          "Column": 39,
          "Offset": 93
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 55
            },
            "end": {
              "Line": 4,
              "Column": 39,
              "Offset": 93
            },
            "children": [
              {
                "type": "scalar",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 55
                },
                "end": {
                  "Line": 4,
                  "Column": 30,
                  "Offset": 84
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 1,
                      "Offset": 55
                    },
                    "end": {
                      "Line": 4,
                      "Column": 2,
                      "Offset": 56
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 2,
                      "Offset": 56
                    },
                    "end": {
                      "Line": 4,
                      "Column": 30,
                      "Offset": 84
                    },
                    "text": "Some::Package::qualified_var"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 31,
                  "Offset": 85
                },
                "end": {
                  "Line": 4,
                  "Column": 32,
                  "Offset": 86
                },
                "text": "="
              },
              {
                "type": "interpolated_string_literal",
                "start": {
                  "Line": 4,
                  "Column": 33,
                  "Offset": 87
                },
                "end": {
                  "Line": 4,
                  "Column": 39,
                  "Offset": 93
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 33,
                      "Offset": 87
                    },
                    "end": {
                      "Line": 4,
                      "Column": 34,
                      "Offset": 88
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 33,
                          "Offset": 87
                        },
                        "end": {
                          "Line": 4,
                          "Column": 34,
                          "Offset": 88
                        },
                        "value": "\"",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 34,
                      "Offset": 88
                    },
                    "end": {
                      "Line": 4,
                      "Column": 38,
                      "Offset": 92
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 34,
                          "Offset": 88
                        },
                        "end": {
                          "Line": 4,
                          "Column": 38,
                          "Offset": 92
                        },
                        "value": "test",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 4,
                      "Column": 38,
                      "Offset": 92
                    },
                    "end": {
                      "Line": 4,
                      "Column": 39,
                      "Offset": 93
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 4,
                          "Column": 38,
                          "Offset": 92
                        },
                        "end": {
                          "Line": 4,
                          "Column": 39,
                          "Offset": 93
                        },
                        "value": "\"",
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
          "Column": 39,
          "Offset": 93
        },
        "end": {
          "Line": 4,
          "Column": 40,
          "Offset": 94
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 5,
          "Column": 1,
          "Offset": 95
        },
        "end": {
          "Line": 5,
          "Column": 18,
          "Offset": 112
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 5,
              "Column": 1,
              "Offset": 95
            },
            "end": {
              "Line": 5,
              "Column": 18,
              "Offset": 112
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 5,
                  "Column": 1,
                  "Offset": 95
                },
                "end": {
                  "Line": 5,
                  "Column": 4,
                  "Offset": 98
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 5,
                      "Column": 1,
                      "Offset": 95
                    },
                    "end": {
                      "Line": 5,
                      "Column": 4,
                      "Offset": 98
                    },
                    "value": "our",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 5,
                  "Column": 5,
                  "Offset": 99
                },
                "end": {
                  "Line": 5,
                  "Column": 18,
                  "Offset": 112
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 5,
                      "Column": 5,
                      "Offset": 99
                    },
                    "end": {
                      "Line": 5,
                      "Column": 6,
                      "Offset": 100
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 5,
                      "Column": 6,
                      "Offset": 100
                    },
                    "end": {
                      "Line": 5,
                      "Column": 18,
                      "Offset": 112
                    },
                    "text": "_private_var"
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
          "Line": 5,
          "Column": 18,
          "Offset": 112
        },
        "end": {
          "Line": 5,
          "Column": 19,
          "Offset": 113
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 6,
          "Column": 1,
          "Offset": 114
        },
        "end": {
          "Line": 6,
          "Column": 18,
          "Offset": 131
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 6,
              "Column": 1,
              "Offset": 114
            },
            "end": {
              "Line": 6,
              "Column": 18,
              "Offset": 131
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 6,
                  "Column": 1,
                  "Offset": 114
                },
                "end": {
                  "Line": 6,
                  "Column": 3,
                  "Offset": 116
                },
                "text": "my"
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 6,
                  "Column": 4,
                  "Offset": 117
                },
                "end": {
                  "Line": 6,
                  "Column": 18,
                  "Offset": 131
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 4,
                      "Offset": 117
                    },
                    "end": {
                      "Line": 6,
                      "Column": 5,
                      "Offset": 118
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 5,
                      "Offset": 118
                    },
                    "end": {
                      "Line": 6,
                      "Column": 18,
                      "Offset": 131
                    },
                    "text": "CamelCase_Var"
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
          "Line": 6,
          "Column": 18,
          "Offset": 131
        },
        "end": {
          "Line": 6,
          "Column": 19,
          "Offset": 132
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 132
}
```
