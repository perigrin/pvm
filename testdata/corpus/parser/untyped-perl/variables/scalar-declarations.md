---
category: untyped-perl
subcategory: variables
tags:
    - scalars
    - declarations
    - scoping
    - package-qualification
    - variables
---

# Scalar Declarations

Test scalar variable declarations with different scoping keywords

```perl
my $simple;
my $assigned = 42;
our $package_var;
state $persistent = "initial";
local $localized;
$Package::qualified = "value";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $simple;
my $assigned = 42;
our $package_var;
state $persistent = "initial";
local $localized;
$Package::qualified = "value";
```

## Typed Perl Output

```perl
my $simple;
my $assigned = 42;
our $package_var;
state $persistent = "initial";
local $localized;
$Package::qualified = "value";
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
    var_decl
      variable
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
    var_decl
      variable
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
```

## JSON AST

```json
{
  "path": "/tmp/scalar_declarations_test.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 6,
      "Column": 31,
      "Offset": 128
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
          "Column": 11,
          "Offset": 10
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
              "Column": 11,
              "Offset": 10
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
                  "Column": 11,
                  "Offset": 10
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
                      "Column": 11,
                      "Offset": 10
                    },
                    "text": "simple"
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
          "Column": 11,
          "Offset": 10
        },
        "end": {
          "Line": 1,
          "Column": 12,
          "Offset": 11
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 12
        },
        "end": {
          "Line": 2,
          "Column": 18,
          "Offset": 29
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 12
            },
            "end": {
              "Line": 2,
              "Column": 18,
              "Offset": 29
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 12
                },
                "end": {
                  "Line": 2,
                  "Column": 18,
                  "Offset": 29
                },
                "name": "assigned",
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
          "Column": 18,
          "Offset": 29
        },
        "end": {
          "Line": 2,
          "Column": 19,
          "Offset": 30
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 31
        },
        "end": {
          "Line": 3,
          "Column": 17,
          "Offset": 47
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 31
            },
            "end": {
              "Line": 3,
              "Column": 17,
              "Offset": 47
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 31
                },
                "end": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 34
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 1,
                      "Offset": 31
                    },
                    "end": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 34
                    },
                    "value": "our",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 35
                },
                "end": {
                  "Line": 3,
                  "Column": 17,
                  "Offset": 47
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 35
                    },
                    "end": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 36
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 36
                    },
                    "end": {
                      "Line": 3,
                      "Column": 17,
                      "Offset": 47
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
          "Column": 17,
          "Offset": 47
        },
        "end": {
          "Line": 3,
          "Column": 18,
          "Offset": 48
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 49
        },
        "end": {
          "Line": 4,
          "Column": 30,
          "Offset": 78
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 49
            },
            "end": {
              "Line": 4,
              "Column": 30,
              "Offset": 78
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 49
                },
                "end": {
                  "Line": 4,
                  "Column": 30,
                  "Offset": 78
                },
                "name": "persistent",
                "sigil": "$"
              }
            ],
            "decl_type": "state"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 4,
          "Column": 30,
          "Offset": 78
        },
        "end": {
          "Line": 4,
          "Column": 31,
          "Offset": 79
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 5,
          "Column": 1,
          "Offset": 80
        },
        "end": {
          "Line": 5,
          "Column": 17,
          "Offset": 96
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 5,
              "Column": 1,
              "Offset": 80
            },
            "end": {
              "Line": 5,
              "Column": 17,
              "Offset": 96
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 5,
                  "Column": 1,
                  "Offset": 80
                },
                "end": {
                  "Line": 5,
                  "Column": 6,
                  "Offset": 85
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 5,
                      "Column": 1,
                      "Offset": 80
                    },
                    "end": {
                      "Line": 5,
                      "Column": 6,
                      "Offset": 85
                    },
                    "value": "local",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 5,
                  "Column": 7,
                  "Offset": 86
                },
                "end": {
                  "Line": 5,
                  "Column": 17,
                  "Offset": 96
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 5,
                      "Column": 7,
                      "Offset": 86
                    },
                    "end": {
                      "Line": 5,
                      "Column": 8,
                      "Offset": 87
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 5,
                      "Column": 8,
                      "Offset": 87
                    },
                    "end": {
                      "Line": 5,
                      "Column": 17,
                      "Offset": 96
                    },
                    "text": "localized"
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
          "Column": 17,
          "Offset": 96
        },
        "end": {
          "Line": 5,
          "Column": 18,
          "Offset": 97
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 6,
          "Column": 1,
          "Offset": 98
        },
        "end": {
          "Line": 6,
          "Column": 30,
          "Offset": 127
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 6,
              "Column": 1,
              "Offset": 98
            },
            "end": {
              "Line": 6,
              "Column": 30,
              "Offset": 127
            },
            "children": [
              {
                "type": "scalar",
                "start": {
                  "Line": 6,
                  "Column": 1,
                  "Offset": 98
                },
                "end": {
                  "Line": 6,
                  "Column": 20,
                  "Offset": 117
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 1,
                      "Offset": 98
                    },
                    "end": {
                      "Line": 6,
                      "Column": 2,
                      "Offset": 99
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 2,
                      "Offset": 99
                    },
                    "end": {
                      "Line": 6,
                      "Column": 20,
                      "Offset": 117
                    },
                    "text": "Package::qualified"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 6,
                  "Column": 21,
                  "Offset": 118
                },
                "end": {
                  "Line": 6,
                  "Column": 22,
                  "Offset": 119
                },
                "text": "="
              },
              {
                "type": "interpolated_string_literal",
                "start": {
                  "Line": 6,
                  "Column": 23,
                  "Offset": 120
                },
                "end": {
                  "Line": 6,
                  "Column": 30,
                  "Offset": 127
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 6,
                      "Column": 23,
                      "Offset": 120
                    },
                    "end": {
                      "Line": 6,
                      "Column": 24,
                      "Offset": 121
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 6,
                          "Column": 23,
                          "Offset": 120
                        },
                        "end": {
                          "Line": 6,
                          "Column": 24,
                          "Offset": 121
                        },
                        "value": "\"",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 6,
                      "Column": 24,
                      "Offset": 121
                    },
                    "end": {
                      "Line": 6,
                      "Column": 29,
                      "Offset": 126
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 6,
                          "Column": 24,
                          "Offset": 121
                        },
                        "end": {
                          "Line": 6,
                          "Column": 29,
                          "Offset": 126
                        },
                        "value": "value",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 6,
                      "Column": 29,
                      "Offset": 126
                    },
                    "end": {
                      "Line": 6,
                      "Column": 30,
                      "Offset": 127
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 6,
                          "Column": 29,
                          "Offset": 126
                        },
                        "end": {
                          "Line": 6,
                          "Column": 30,
                          "Offset": 127
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
          "Line": 6,
          "Column": 30,
          "Offset": 127
        },
        "end": {
          "Line": 6,
          "Column": 31,
          "Offset": 128
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 128
}
```
